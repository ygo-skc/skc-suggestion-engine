package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"sync"

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
	suggestionChan := make(chan model.CardSuggestions, 5)
	go fetchBatchSuggestions(suggestionSubjectsCardData,
		func(cardInfo model.Card, wg *sync.WaitGroup, c chan<- model.CardSuggestions) {
			defer wg.Done()
			suggestionChan <- getCardSuggestions(ctx, cardInfo, ccIDs)
		}, suggestionChan)

	uniqueNamedMaterialsByCardID, uniqueNamedReferencesByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)
	uniqueMaterialArchetypes, uniqueReferencedArchetypes := make(map[string]bool), make(map[string]bool)

	suggestions := model.BatchCardSuggestions[model.CardIDs]{
		UnknownResources:     suggestionSubjectsCardData.UnknownResources,
		FalsePositives:       make(model.CardIDs, 0, 5),
		NamedMaterials:       make([]model.CardReference, 0, 5),
		NamedReferences:      make([]model.CardReference, 0, 5),
		MaterialArchetypes:   make([]string, 0),
		ReferencedArchetypes: make([]string, 0)}
	for s := range suggestionChan {
		groupSuggestionReferences(s.NamedMaterials, uniqueNamedMaterialsByCardID, &suggestions.NamedMaterials,
			suggestionSubjectsCardData.CardInfo, &suggestions.FalsePositives)
		groupSuggestionReferences(s.NamedReferences, uniqueNamedReferencesByCardIDs, &suggestions.NamedReferences,
			suggestionSubjectsCardData.CardInfo, &suggestions.FalsePositives)
		groupArchetypes(s.MaterialArchetypes, uniqueMaterialArchetypes, &suggestions.MaterialArchetypes)
		groupArchetypes(s.ReferencedArchetypes, uniqueReferencedArchetypes, &suggestions.ReferencedArchetypes)
	}

	sort.SliceStable(suggestions.NamedMaterials, sortBatchReferences(suggestions.NamedMaterials))
	sort.SliceStable(suggestions.NamedReferences, sortBatchReferences(suggestions.NamedReferences))

	return suggestions
}

func sortBatchReferences(refs []model.CardReference) func(i, j int) bool {
	return func(i, j int) bool {
		return refs[i].Occurrences > refs[j].Occurrences
	}
}

func groupArchetypes(archetypesToParse []string, uniqueArchetypeSet map[string]bool, uniqueArchetypes *[]string) {
	for _, archetype := range archetypesToParse {
		if _, exists := uniqueArchetypeSet[archetype]; !exists {
			uniqueArchetypeSet[archetype] = true
			*uniqueArchetypes = append(*uniqueArchetypes, archetype)
		}
	}
}

// uses references for a card and builds upon uniqueReferencesByCardID and uniqueReferences
func groupSuggestionReferences(referencesToParse []model.CardReference, uniqueReferencesByCardID map[string]*model.CardReference, uniqueReferences *[]model.CardReference, uniqueCardIDs model.CardDataMap, falsePositives *model.CardIDs) {
	for _, suggestion := range referencesToParse {
		if batchSuggestion, refPreviouslyAdded := uniqueReferencesByCardID[suggestion.Card.CardID]; refPreviouslyAdded {
			batchSuggestion.Occurrences += suggestion.Occurrences
			uniqueReferencesByCardID[suggestion.Card.CardID] = batchSuggestion
		} else if _, isFalsePositive := uniqueCardIDs[suggestion.Card.CardID]; isFalsePositive && !slices.Contains(*falsePositives, suggestion.Card.CardID) {
			*falsePositives = append(*falsePositives, suggestion.Card.CardID)
		} else if !refPreviouslyAdded && !isFalsePositive {
			*uniqueReferences = append(*uniqueReferences, model.CardReference{Card: suggestion.Card, Occurrences: suggestion.Occurrences})
			uniqueReferencesByCardID[suggestion.Card.CardID] = &(*uniqueReferences)[len(*uniqueReferences)-1]
		}
	}
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
		supportChan := make(chan model.CardSupport, 5)
		go fetchBatchSuggestions(suggestionSubjectsCardData,
			func(cardInfo model.Card, wg *sync.WaitGroup, c chan<- model.CardSupport) {
				defer wg.Done()
				cardSupport, _ := getCardSupport(ctx, cardInfo)
				c <- cardSupport
			}, supportChan)

		support := model.BatchCardSupport[model.CardIDs]{
			ReferencedBy:     make([]model.Card, 0, 10),
			MaterialFor:      make([]model.Card, 0, 10),
			FalsePositives:   make(model.CardIDs, 0, 5),
			UnknownResources: suggestionSubjectsCardData.UnknownResources}
		for s := range supportChan {
			support.ReferencedBy = append(support.ReferencedBy, s.ReferencedBy...)
			support.MaterialFor = append(support.MaterialFor, s.MaterialFor...)
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(support)
	}
}

func fetchBatchSuggestions[T model.CardSupport | model.CardSuggestions](suggestionSubjectsCardData model.BatchCardData[model.CardIDs],
	fetchSuggestions func(model.Card, *sync.WaitGroup, chan<- T), c chan<- T) {
	var wg sync.WaitGroup
	for _, cardInfo := range suggestionSubjectsCardData.CardInfo {
		// card ID is invalid
		if slices.Contains(suggestionSubjectsCardData.UnknownResources, cardInfo.CardID) {
			continue
		}

		wg.Add(1)
		go fetchSuggestions(cardInfo, &wg, c)
	}
	wg.Wait()
	close(c)
}
