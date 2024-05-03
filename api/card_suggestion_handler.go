package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/util"
	"github.com/ygo-skc/skc-suggestion-engine/validation"
)

var (
	quotedStringRegex = regexp.MustCompile("^(\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')|[\\W](\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')")
)

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getCardSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Printf("Getting suggestions for card w/ ID: %s", cardID)

	if cardToGetSuggestionsFor, err := skcDBInterface.GetDesiredCardInDBUsingID(cardID); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		ccIds, _ := skcDBInterface.GetCardColorIDs() // retrieve card color IDs
		suggestions := getSuggestions(*cardToGetSuggestionsFor, ccIds)

		log.Printf("Found %d unique material references", len(*suggestions.NamedMaterials))
		log.Printf("Found %d unique named references", len(*suggestions.NamedReferences))
		log.Printf("Has self reference: %t", suggestions.HasSelfReference)

		json.NewEncoder(res).Encode(suggestions)
	}
}

func getSuggestions(cardToGetSuggestionsFor model.Card, ccIds map[string]int) *model.CardSuggestions {
	suggestions := model.CardSuggestions{Card: &cardToGetSuggestionsFor}
	materialString := cardToGetSuggestionsFor.GetPotentialMaterialsAsString()

	// setup channels
	materialChannel, referenceChannel := make(chan bool), make(chan bool)

	// get materials if card is from extra deck
	if cardToGetSuggestionsFor.IsExtraDeckMonster() {
		go getMaterialRefs(&suggestions, materialString, ccIds, materialChannel)
	} else {
		materialChannel = nil
		suggestions.NamedMaterials = &[]model.CardReference{}
		suggestions.MaterialArchetypes = &[]string{}

		log.Printf("%s is not an ED monster", cardToGetSuggestionsFor.CardID)
	}

	go getNonMaterialRefs(&suggestions, cardToGetSuggestionsFor, materialString, ccIds, referenceChannel)

	// join channels
	if materialChannel != nil {
		<-materialChannel
	}
	<-referenceChannel

	return &suggestions
}

func getMaterialRefs(s *model.CardSuggestions, materialString string, ccIds map[string]int, c chan bool) {
	s.NamedMaterials, s.MaterialArchetypes = getReferences(materialString)
	sortCardReferences(s.NamedMaterials, ccIds)

	c <- true
}

// get named references - excludes materials
// will also check and remove self references
func getNonMaterialRefs(s *model.CardSuggestions, cardToGetSuggestionsFor model.Card, materialString string, ccIds map[string]int, c chan bool) {
	s.NamedReferences, s.ReferencedArchetypes = getReferences(strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, materialString, ""))
	s.HasSelfReference = util.RemoveSelfReference(cardToGetSuggestionsFor.CardName, s.NamedReferences)
	sortCardReferences(s.NamedReferences, ccIds)

	c <- true
}

func sortCardReferences(cr *[]model.CardReference, ccIds map[string]int) {
	// sorting alphabetically from a-z
	sort.SliceStable(*cr, func(i, j int) bool {
		return (*cr)[i].Card.CardName < (*cr)[j].Card.CardName
	})

	// sorting by card color
	sort.SliceStable(*cr, func(i, j int) bool {
		return ccIds[(*cr)[i].Card.CardColor] < ccIds[(*cr)[j].Card.CardColor]
	})
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func getReferences(s string) (*[]model.CardReference, *[]string) {
	namedReferences, referenceOccurrence, archetypalReferences := isolateReferences(s)

	uniqueReferences := make([]model.CardReference, 0, len(namedReferences))
	for _, card := range namedReferences {
		uniqueReferences = append(uniqueReferences, model.CardReference{Card: card, Occurrences: referenceOccurrence[card.CardID]})
	}

	return &uniqueReferences, &archetypalReferences
}

func isolateReferences(s string) (map[string]model.Card, map[string]int, []string) {
	tokens := quotedStringRegex.FindAllString(s, -1)

	namedReferences, referenceOccurrence, archetypalReferences := buildReferenceObjects(tokens)

	// get unique archetypes
	uniqueArchetypalReferences := make([]string, 0, len(archetypalReferences))
	for ref := range archetypalReferences {
		uniqueArchetypalReferences = append(uniqueArchetypalReferences, ref)
	}
	sort.Strings(uniqueArchetypalReferences) // needed as source of this array was a map and maps don't have predictable sorting - tests will fail randomly without sort

	if len(archetypalReferences) > 0 {
		log.Printf("Could not find the following in DB: %v. Potentially archetypes?", archetypalReferences)
	}

	return namedReferences, referenceOccurrence, uniqueArchetypalReferences
}

// cycles through tokens - makes DB calls where necessary and attempts to build objects containing direct references (and their occurrences), archetype references
func buildReferenceObjects(tokens []string) (map[string]model.Card, map[string]int, map[string]bool) {
	namedReferences := map[string]model.Card{}
	referenceOccurrence := map[string]int{}
	archetypalReferences := map[string]bool{}
	tokenToCardId := map[string]string{} // maps token to its cardID - token will only have cardID if token is found in DB
	totalTokens := len(tokens)

	if totalTokens != 0 {
		for i := 0; i < totalTokens; i++ {
			model.CleanupToken(&tokens[i])
		}

		batchCardData, _ := skcDBInterface.GetDesiredCardsFromDBUsingMultipleCardNames(tokens)

		for _, token := range tokens {
			// if we already searched the token before we don't need to waste time re-searching it in DB

			// if token is present in archetype slice, skip token
			if _, isPresent := archetypalReferences[token]; isPresent {
				continue
			}

			// if token mapped to a cardId in previous loop, increase number of occurrences by 1 and skip any other processing this iteration as we already did the processing before
			if _, isPresent := tokenToCardId[token]; isPresent {
				referenceOccurrence[tokenToCardId[token]] += 1
				continue
			}

			if card, isPresent := batchCardData.CardInfo[token]; !isPresent {
				// add occurrence of archetype to map
				archetypalReferences[token] = true
			} else {
				// add occurrence of referenced card to maps
				namedReferences[card.CardID] = card
				referenceOccurrence[card.CardID] = 1
				tokenToCardId[token] = card.CardID
			}
		}
	}

	return namedReferences, referenceOccurrence, archetypalReferences
}

func getBatchSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Getting batch suggestions")

	// deserialize body
	var reqBody model.BatchCardIDs
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		log.Printf("Error occurred while reading batch suggestions request body. Error %s", err)
		model.HandleServerResponse(model.APIError{Message: "Body could not be deserialized.", StatusCode: http.StatusBadRequest}, res)
		return
	}

	// validate body
	if err := validation.ValidateBatchCardIDs(reqBody); err != nil {
		res.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(res).Encode(err)
		return
	}

	if len(reqBody.CardIDs) == 0 {
		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode("Empty") //TODO: return appropriate body
		return
	}

	if suggestionSubjectsCardData, err := skcDBInterface.GetDesiredCardInDBUsingMultipleCardIDs(reqBody.CardIDs); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		ccIds, _ := skcDBInterface.GetCardColorIDs() // retrieve card color IDs

		suggestionChan := make(chan *model.CardSuggestions)
		unknownIDs := suggestionSubjectsCardData.CardInfo.FindMissingIDs(reqBody.CardIDs)
		for _, cardInfo := range suggestionSubjectsCardData.CardInfo {
			if slices.Contains(unknownIDs, cardInfo.CardID) {
				continue
			}
			go func(card model.Card) {
				suggestionChan <- getSuggestions(card, ccIds)
			}(cardInfo)
		}

		uniqueNamedMaterialsByCardID, uniqueNamedReferencesByCardIDs := make(map[string]*model.CardReference), make(map[string]*model.CardReference)
		suggestions := model.BatchCardSuggestions[model.CardIDs]{UnknownResources: unknownIDs,
			NamedMaterials: make([]model.CardReference, 0, 5), NamedReferences: make([]model.CardReference, 0, 5)}
		for i := 0; i < len(suggestionSubjectsCardData.CardInfo)-len(unknownIDs); i++ {
			s := <-suggestionChan
			groupSuggestions(*s.NamedMaterials, uniqueNamedMaterialsByCardID, &suggestions.NamedMaterials)
			groupSuggestions(*s.NamedReferences, uniqueNamedReferencesByCardIDs, &suggestions.NamedReferences)
		}
		sort.SliceStable(suggestions.NamedMaterials, sortBatchReferences(suggestions.NamedMaterials))
		sort.SliceStable(suggestions.NamedReferences, sortBatchReferences(suggestions.NamedReferences))

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(suggestions)
	}
}

func sortBatchReferences(refs []model.CardReference) func(i, j int) bool {
	return func(i, j int) bool {
		return refs[i].Occurrences > refs[j].Occurrences
	}
}

// uses references for a card and builds upon uniqueReferencesByCardID and uniqueReferences
func groupSuggestions(cardReferences []model.CardReference, uniqueReferencesByCardID map[string]*model.CardReference, uniqueReferences *[]model.CardReference) {
	for _, suggestion := range cardReferences {
		if batchSuggestion, exists := uniqueReferencesByCardID[suggestion.Card.CardID]; exists {
			batchSuggestion.Occurrences += suggestion.Occurrences
			uniqueReferencesByCardID[suggestion.Card.CardID] = batchSuggestion
		} else {
			*uniqueReferences = append(*uniqueReferences, model.CardReference{Card: suggestion.Card, Occurrences: suggestion.Occurrences})
			uniqueReferencesByCardID[suggestion.Card.CardID] = &(*uniqueReferences)[len(*uniqueReferences)-1]
		}
	}
}
