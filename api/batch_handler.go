package api

import (
	"context"

	"fmt"
	"net/http"
	"slices"
	"sort"

	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

const (
	batchCardInfoOp        = "Batch Card Info"
	batchCardSuggestionsOp = "Batch Card Suggestions"
	batchCardSupportOp     = "Batch Card Support"
)

func getBatchCardInfo(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, batchCardInfoOp)
	logger.Info("Getting batch card info")

	if reqBody := batchRequestValidator[cModel.CardIDs, cModel.BatchCardData[cModel.CardIDs]](ctx, res, req); reqBody == nil {
		return
	} else if batchCardInfo, err := downstream.YGO.CardService.GetCardsByID(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
	} else {
		if len(batchCardInfo.UnknownResources) > 0 {
			logger.Warn(fmt.Sprintf("Following card IDs are not valid (no card data found in DB). IDs: %v", batchCardInfo.UnknownResources))
		}
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(batchCardInfo)
	}
}

func batchRequestValidator[RK cModel.YGOResourceKey, T cModel.BatchCardData[RK] | model.BatchCardSuggestions[RK] | model.BatchCardSupport[RK]](
	ctx context.Context, res http.ResponseWriter, req *http.Request) *cModel.BatchCardIDs {
	logger := cUtil.RetrieveLogger(ctx)
	var reqBody cModel.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		logger.Error(fmt.Sprintf("Error occurred while reading batch request body: Error %v", err))
		cModel.HandleServerResponse(cModel.APIError{Message: "Body could not be deserialized", StatusCode: http.StatusBadRequest}, res)
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

		var empty T
		switch any(empty).(type) {
		case cModel.BatchCardData[RK]:
			json.NewEncoder(res).Encode(
				cModel.BatchCardData[RK]{CardInfo: make(cModel.CardDataMap, 0), UnknownResources: make(RK, 0)},
			)
		case model.BatchCardSuggestions[RK]:
			json.NewEncoder(res).Encode(
				model.BatchCardSuggestions[RK]{
					NamedMaterials:       make([]model.CardReference, 0),
					NamedReferences:      make([]model.CardReference, 0),
					MaterialArchetypes:   make([]string, 0),
					ReferencedArchetypes: make([]string, 0),
					UnknownResources:     make(RK, 0),
					FalsePositives:       make(RK, 0),
				},
			)
		case model.BatchCardSupport[RK]:
			json.NewEncoder(res).Encode(
				model.BatchCardSupport[RK]{
					ReferencedBy:     make([]model.CardReference, 0),
					MaterialFor:      make([]model.CardReference, 0),
					UnknownResources: make(RK, 0),
					FalsePositives:   make(RK, 0),
				},
			)
		}
		return nil
	}

	return &reqBody
}

func getBatchSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, batchCardSuggestionsOp)
	logger.Info("Batch card suggestions requested")

	if reqBody := batchRequestValidator[cModel.CardIDs, model.BatchCardSuggestions[cModel.CardIDs]](ctx, res, req); reqBody == nil {
		return
	} else if suggestionSubjectsCardData, err := downstream.YGO.CardService.GetCardsByID(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		ccIDs, _ := downstream.YGO.CardService.GetCardColorsProto(ctx) // retrieve card color IDs
		suggestions := getBatchSuggestions(ctx, *suggestionSubjectsCardData, ccIDs.Values)

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(suggestions)
	}
}

func getBatchSuggestions(ctx context.Context, suggestionSubjectsCardData cModel.BatchCardData[cModel.CardIDs],
	ccIDs map[string]uint32) model.BatchCardSuggestions[cModel.CardIDs] {
	suggestionChan := make(chan model.CardSuggestions, 20)
	go fetchBatchSuggestions(ctx, suggestionSubjectsCardData, suggestionChan, func(cardInfo cModel.YGOCard) model.CardSuggestions {
		return getCardSuggestions(ctx, cardInfo, ccIDs)
	})

	uniqueNamedMaterialsByCardID, uniqueNamedReferencesByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)
	uniqueMaterialArchetypes, uniqueReferencedArchetypes := make(map[string]struct{}), make(map[string]struct{})

	suggestions := model.BatchCardSuggestions[cModel.CardIDs]{
		UnknownResources:     suggestionSubjectsCardData.UnknownResources,
		FalsePositives:       make(cModel.CardIDs, 0, 5),
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

	// sort output
	sort.SliceStable(suggestions.NamedMaterials, sortBatchReferences(suggestions.NamedMaterials, ccIDs))
	sort.SliceStable(suggestions.NamedReferences, sortBatchReferences(suggestions.NamedReferences, ccIDs))
	sort.Strings(suggestions.MaterialArchetypes)
	sort.Strings(suggestions.ReferencedArchetypes)
	sort.Strings(suggestions.FalsePositives)
	sort.Strings(suggestions.UnknownResources)

	return suggestions
}

func sortBatchReferences(refs []model.CardReference, ccIDs map[string]uint32) func(i, j int) bool {
	return func(i, j int) bool {
		iv, jv := refs[i], refs[j]
		switch {
		case iv.Occurrences != jv.Occurrences:
			return iv.Occurrences > jv.Occurrences
		case iv.Card.GetColor() != jv.Card.GetColor():
			return ccIDs[iv.Card.GetColor()] < ccIDs[jv.Card.GetColor()]
		default:
			return iv.Card.GetColor() < jv.Card.GetColor()
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
	subjects cModel.CardDataMap, falsePositives *cModel.CardIDs) {
	for _, suggestion := range referencesToParse {
		suggestionID := suggestion.Card.GetID()
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
	logger, ctx := cUtil.InitRequest(context.Background(), apiName, batchCardSupportOp)
	logger.Info("Batch card support requested")

	if reqBody := batchRequestValidator[cModel.CardIDs, model.BatchCardSupport[cModel.CardIDs]](ctx, res, req); reqBody == nil {
		return
	} else if suggestionSubjectsCardData, err := downstream.YGO.CardService.GetCardsByID(ctx, reqBody.CardIDs); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(getBatchSupport(ctx, *suggestionSubjectsCardData))
	}
}

func getBatchSupport(ctx context.Context, suggestionSubjectsCardData cModel.BatchCardData[cModel.CardIDs]) model.BatchCardSupport[cModel.CardIDs] {
	supportChan := make(chan model.CardSupport, 20)
	go fetchBatchSuggestions(ctx, suggestionSubjectsCardData, supportChan, func(cardInfo cModel.YGOCard) model.CardSupport {
		cardSupport, _ := getCardSupport(ctx, cardInfo)
		return cardSupport
	})

	support := model.BatchCardSupport[cModel.CardIDs]{
		FalsePositives:   make(cModel.CardIDs, 0, 5),
		UnknownResources: suggestionSubjectsCardData.UnknownResources}
	uniqueReferenceByCardID, uniqueMaterialByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)

	ccIDs, _ := downstream.YGO.CardService.GetCardColorsProto(ctx) // retrieve card color IDs

	for s := range supportChan {
		parseSuggestionReferences(s.ReferencedBy, uniqueReferenceByCardID,
			suggestionSubjectsCardData.CardInfo, &support.FalsePositives)
		parseSuggestionReferences(s.MaterialFor, uniqueMaterialByCardIDs,
			suggestionSubjectsCardData.CardInfo, &support.FalsePositives)
	}

	support.ReferencedBy = getUniqueReferences(uniqueReferenceByCardID)
	support.MaterialFor = getUniqueReferences(uniqueMaterialByCardIDs)

	sort.SliceStable(support.ReferencedBy, sortBatchReferences(support.ReferencedBy, ccIDs.Values))
	sort.SliceStable(support.MaterialFor, sortBatchReferences(support.MaterialFor, ccIDs.Values))

	return support
}

type batchSuggestionTask[T model.CardSupport | model.CardSuggestions] struct {
	card       cModel.YGOCard
	resultChan chan<- T
	process    func(card cModel.YGOCard) T
}

func (t batchSuggestionTask[T]) Process() {
	t.resultChan <- t.process(t.card)
}

func fetchBatchSuggestions[T model.CardSupport | model.CardSuggestions](ctx context.Context, suggestionSubjectsCardData cModel.BatchCardData[cModel.CardIDs],
	resultChan chan<- T, process func(card cModel.YGOCard) T) {
	tasks := []cUtil.Task{}
	for _, cardInfo := range suggestionSubjectsCardData.CardInfo {
		// card ID is invalid
		if slices.Contains(suggestionSubjectsCardData.UnknownResources, cardInfo.GetID()) {
			continue
		}

		tasks = append(tasks, batchSuggestionTask[T]{card: cardInfo, resultChan: resultChan, process: process})
	}

	pool := *cUtil.NewWorkerPool(tasks, cUtil.WithContext(ctx), cUtil.WithWorkers(10))
	pool.Run()
	close(resultChan)
}
