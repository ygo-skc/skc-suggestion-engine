package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"

	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

func getBatchCardInfo(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "batch card info")
	logger.Info("Getting batch card info")

	batchCardInfo := model.BatchCardData[model.CardIDs]{CardInfo: model.CardDataMap{}, UnknownResources: model.CardIDs{}}
	var err *model.APIError
	if reqBody := batchRequestValidator(ctx, res, req, batchCardInfo, "card info"); reqBody == nil {
		return
	} else if batchCardInfo, err = skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
	} else {
		if len(batchCardInfo.UnknownResources) > 0 {
			logger.Warn(fmt.Sprintf("Following card IDs are not valid (no card data found in DB). IDs: %v", batchCardInfo.UnknownResources))
		}
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(batchCardInfo)
	}
}

func batchRequestValidator(ctx context.Context, res http.ResponseWriter, req *http.Request, nothingToProcessBody interface{},
	op string) *model.BatchCardIDs {
	logger := util.LoggerFromContext(ctx)
	var reqBody model.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		logger.Error(fmt.Sprintf("Error occurred while reading batch %s request body: Error %v", op, err))
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized", StatusCode: http.StatusBadRequest}, res)
		return nil
	}

	// validate body
	if err := validation.ValidateBatchCardIDs(reqBody); err != nil {
		err.HandleServerResponse(res)
		return nil
	}

	if len(reqBody.CardIDs) == 0 {
		logger.Warn("Nothing to process - missing cardID data")
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(nothingToProcessBody)
		return nil
	}

	return &reqBody
}

func getBatchSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "batch card suggestions")
	logger.Info("Batch card suggestions requested")

	if reqBody := batchRequestValidator(ctx, res, req, noBatchSuggestions, "suggestion"); reqBody == nil {
		return
	} else if suggestionSubjectsCardData, err := skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs
		suggestions := getBatchSuggestions(ctx, suggestionSubjectsCardData, ccIDs)

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(suggestions)
	}

}

func getBatchSuggestions(ctx context.Context, suggestionSubjectsCardData model.BatchCardData[model.CardIDs],
	ccIDs map[string]int) model.BatchCardSuggestions[model.CardIDs] {
	suggestionChan := make(chan model.CardSuggestions, 20)
	go fetchBatchSuggestions(ctx, suggestionSubjectsCardData, suggestionChan, func(cardInfo model.Card) model.CardSuggestions {
		return getCardSuggestions(ctx, cardInfo, ccIDs)
	})

	uniqueNamedMaterialsByCardID, uniqueNamedReferencesByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)
	uniqueMaterialArchetypes, uniqueReferencedArchetypes := make(map[string]struct{}), make(map[string]struct{})

	suggestions := model.BatchCardSuggestions[model.CardIDs]{
		UnknownResources:     suggestionSubjectsCardData.UnknownResources,
		FalsePositives:       make(model.CardIDs, 0, 5),
		NamedMaterials:       make([]model.CardReference, 0, 5),
		NamedReferences:      make([]model.CardReference, 0, 5),
		MaterialArchetypes:   make([]string, 0),
		ReferencedArchetypes: make([]string, 0)}
	for s := range suggestionChan {
		parseSuggestionReferences(s.NamedMaterials, uniqueNamedMaterialsByCardID,
			suggestionSubjectsCardData.CardInfo, &suggestions.FalsePositives)
		parseSuggestionReferences(s.NamedReferences, uniqueNamedReferencesByCardIDs,
			suggestionSubjectsCardData.CardInfo, &suggestions.FalsePositives)
		groupArchetypes(s.MaterialArchetypes, uniqueMaterialArchetypes, &suggestions.MaterialArchetypes)
		groupArchetypes(s.ReferencedArchetypes, uniqueReferencedArchetypes, &suggestions.ReferencedArchetypes)
	}

	suggestions.NamedMaterials = getUniqueReferences(uniqueNamedMaterialsByCardID)
	suggestions.NamedReferences = getUniqueReferences(uniqueNamedReferencesByCardIDs)

	sort.SliceStable(suggestions.NamedMaterials, sortBatchReferences(suggestions.NamedMaterials, ccIDs))
	sort.SliceStable(suggestions.NamedReferences, sortBatchReferences(suggestions.NamedReferences, ccIDs))

	return suggestions
}

func sortBatchReferences(refs []model.CardReference, ccIDs map[string]int) func(i, j int) bool {
	return func(i, j int) bool {
		iv, jv := refs[i], refs[j]
		switch {
		case iv.Occurrences != jv.Occurrences:
			return iv.Occurrences > jv.Occurrences
		case iv.Card.CardColor != jv.Card.CardColor:
			return ccIDs[iv.Card.CardColor] < ccIDs[jv.Card.CardColor]
		default:
			return iv.Card.CardName < jv.Card.CardName
		}
	}
}

func groupArchetypes(archetypesToParse []string, uniqueArchetypeSet map[string]struct{}, uniqueArchetypes *[]string) {
	for _, archetype := range archetypesToParse {
		if _, exists := uniqueArchetypeSet[archetype]; !exists {
			uniqueArchetypeSet[archetype] = struct{}{}
			*uniqueArchetypes = append(*uniqueArchetypes, archetype)
		}
	}
}

// uses references for a card and builds upon uniqueReferencesByCardID and uniqueReferences
func parseSuggestionReferences(referencesToParse []model.CardReference, uniqueReferencesByCardID map[string]*model.CardReference,
	subjects model.CardDataMap, falsePositives *model.CardIDs) {
	for _, suggestion := range referencesToParse {
		suggestionID := suggestion.Card.CardID
		if _, refPreviouslyAdded := uniqueReferencesByCardID[suggestionID]; refPreviouslyAdded {
			uniqueReferencesByCardID[suggestionID].Occurrences += suggestion.Occurrences
		} else if _, isFalsePositive := subjects[suggestionID]; isFalsePositive && !slices.Contains(*falsePositives, suggestionID) {
			*falsePositives = append(*falsePositives, suggestionID)
		} else if !refPreviouslyAdded && !isFalsePositive {
			uniqueReferencesByCardID[suggestionID] = &model.CardReference{Card: suggestion.Card, Occurrences: suggestion.Occurrences}
		}
	}
}

func getUniqueReferences(uniqueReferences map[string]*model.CardReference) []model.CardReference {
	references := []model.CardReference{}
	for _, ref := range uniqueReferences {
		references = append(references, *ref)
	}

	return references
}

func getBatchSupportHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := util.NewRequestSetup(context.Background(), "batch card support")
	logger.Info("Batch card support requested")

	if reqBody := batchRequestValidator(ctx, res, req, noBatchSuggestions, "support"); reqBody == nil {
		return
	} else if suggestionSubjectsCardData, err := skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(getBatchSupport(ctx, suggestionSubjectsCardData))
	}
}

func getBatchSupport(ctx context.Context, suggestionSubjectsCardData model.BatchCardData[model.CardIDs]) model.BatchCardSupport[model.CardIDs] {
	supportChan := make(chan model.CardSupport, 20)
	go fetchBatchSuggestions(ctx, suggestionSubjectsCardData, supportChan, func(cardInfo model.Card) model.CardSupport {
		cardSupport, _ := getCardSupport(ctx, cardInfo)
		return cardSupport
	})

	support := model.BatchCardSupport[model.CardIDs]{
		FalsePositives:   make(model.CardIDs, 0, 5),
		UnknownResources: suggestionSubjectsCardData.UnknownResources}
	uniqueReferenceByCardID, uniqueMaterialByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)

	ccIDs, _ := skcDBInterface.GetCardColorIDs(ctx) // retrieve card color IDs

	for s := range supportChan {
		parseSuggestionReferences(s.ReferencedBy, uniqueReferenceByCardID,
			suggestionSubjectsCardData.CardInfo, &support.FalsePositives)
		parseSuggestionReferences(s.MaterialFor, uniqueMaterialByCardIDs,
			suggestionSubjectsCardData.CardInfo, &support.FalsePositives)
	}

	support.ReferencedBy = getUniqueReferences(uniqueReferenceByCardID)
	support.MaterialFor = getUniqueReferences(uniqueMaterialByCardIDs)

	sort.SliceStable(support.ReferencedBy, sortBatchReferences(support.ReferencedBy, ccIDs))
	sort.SliceStable(support.MaterialFor, sortBatchReferences(support.MaterialFor, ccIDs))

	return support
}

type batchSuggestionTask[T model.CardSupport | model.CardSuggestions] struct {
	card       model.Card
	resultChan chan<- T
	process    func(card model.Card) T
}

func (t batchSuggestionTask[T]) Process() {
	t.resultChan <- t.process(t.card)
}

func fetchBatchSuggestions[T model.CardSupport | model.CardSuggestions](ctx context.Context, suggestionSubjectsCardData model.BatchCardData[model.CardIDs],
	resultChan chan<- T, process func(card model.Card) T) {
	tasks := []util.Task{}
	for _, cardInfo := range suggestionSubjectsCardData.CardInfo {
		// card ID is invalid
		if slices.Contains(suggestionSubjectsCardData.UnknownResources, cardInfo.CardID) {
			continue
		}

		tasks = append(tasks, batchSuggestionTask[T]{card: cardInfo, resultChan: resultChan, process: process})
	}

	pool := *util.NewWorkerPool(tasks, util.WithContext(ctx), util.WithWorkers(10))
	pool.Run()
	close(resultChan)
}
