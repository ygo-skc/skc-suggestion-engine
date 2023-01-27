package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	quotedStringRegex            = regexp.MustCompile("^(\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')|[\\W](\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')")
	deckListCardAndQuantityRegex = regexp.MustCompile("[1-3][xX][0-9]{8}")
)

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Println("Getting suggestions for card w/ ID:", cardID)

	if cardToGetSuggestionsFor, err := skcDBInterface.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.WriteHeader(err.StatusCode)
		json.NewEncoder(res).Encode(err)
	} else {
		suggestions := getSuggestions(cardToGetSuggestionsFor)

		log.Printf("Found %d unique material references", len(*suggestions.NamedMaterials))
		log.Printf("Found %d unique named references", len(*suggestions.NamedReferences))
		log.Printf("Has self reference: %t", suggestions.HasSelfReference)

		json.NewEncoder(res).Encode(suggestions)
	}
}

func getSuggestions(cardToGetSuggestionsFor *model.Card) *model.CardSuggestions {
	suggestions := model.CardSuggestions{Card: cardToGetSuggestionsFor}
	materialString := cardToGetSuggestionsFor.GetPotentialMaterialsAsString()

	// setup channels
	materialChannel, referenceChannel := make(chan bool), make(chan bool)

	// get materials if card is from extra deck
	if cardToGetSuggestionsFor.IsExtraDeckMonster() {
		go getMaterialRefs(&suggestions, materialString, materialChannel)
	} else {
		materialChannel = nil
		suggestions.NamedMaterials = &[]model.CardReference{}
		suggestions.MaterialArchetypes = &[]string{}

		log.Printf("%s is not an ED monster", cardToGetSuggestionsFor.CardID)
	}

	go getNonMaterialRefs(&suggestions, *cardToGetSuggestionsFor, materialString, referenceChannel)

	// get decks that feature card
	suggestions.Decks, _ = skcSuggestionEngineDBInterface.GetDecksThatFeatureCards([]string{cardToGetSuggestionsFor.CardID})

	// join channels
	if materialChannel != nil {
		<-materialChannel
	}
	<-referenceChannel

	return &suggestions
}

func getMaterialRefs(s *model.CardSuggestions, materialString string, c chan bool) {
	s.NamedMaterials, s.MaterialArchetypes = getReferences(materialString)
	c <- true
}

// get named references - excludes materials
// will also check and remove self references
func getNonMaterialRefs(s *model.CardSuggestions, cardToGetSuggestionsFor model.Card, materialString string, c chan bool) {
	s.NamedReferences, s.ReferencedArchetypes = getReferences(strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, materialString, ""))
	s.HasSelfReference = removeSelfReference(cardToGetSuggestionsFor.CardName, s.NamedReferences)

	c <- true
}

// looks for a self reference, if a self reference is found it is removed from original slice
// this method returns true if a self reference is found
func removeSelfReference(self string, cr *[]model.CardReference) bool {
	hasSelfRef := false

	if cr != nil {
		x := 0
		for _, ref := range *cr {
			if ref.Card.CardName != self {
				(*cr)[x] = ref
				x++
			} else {
				hasSelfRef = true
			}
		}

		*cr = (*cr)[:x]
		return hasSelfRef
	} else {
		return hasSelfRef
	}
}

// Uses regex to find all direct references to cards (or potentially archetypes) and searches it in the DB.
// If a direct name reference is found in the DB, then it is returned as a suggestion.
func getReferences(s string) (*[]model.CardReference, *[]string) {
	namedReferences, referenceOccurrence, archetypalReferences := isolateReferences(s)

	uniqueReferences := make([]model.CardReference, 0, len(namedReferences))
	for _, card := range namedReferences {
		uniqueReferences = append(uniqueReferences, model.CardReference{Card: card, Occurrences: referenceOccurrence[card.CardID]})
	}

	sort.SliceStable(uniqueReferences, func(i, j int) bool {
		return uniqueReferences[i].Card.CardName < uniqueReferences[j].Card.CardName // sorting alphabetically from a-z
	})

	return &uniqueReferences, &archetypalReferences
}

func isolateReferences(s string) (map[string]model.Card, map[string]int, []string) {
	tokens := quotedStringRegex.FindAllString(s, -1)

	namedReferences := map[string]model.Card{}
	referenceOccurrence := map[string]int{}
	archetypalReferences := map[string]bool{}
	tokenToCardId := map[string]string{} // maps token to its cardID - token will only have cardID if token is found in DB

	for _, token := range tokens {
		cleanupToken(&token)

		// if we already searched the token before we don't need to waste time re-searching it in DB
		if _, isPresent := archetypalReferences[token]; isPresent {
			continue
		}
		if _, isPresent := tokenToCardId[token]; isPresent {
			referenceOccurrence[tokenToCardId[token]] += 1
			continue
		}

		if card, err := skcDBInterface.FindDesiredCardInDBUsingName(token); err != nil {
			// add occurrence of archetype to map
			archetypalReferences[token] = true
		} else {
			// add occurrence of referenced card to maps
			namedReferences[card.CardID] = card
			referenceOccurrence[card.CardID] = 1
			tokenToCardId[token] = card.CardID
		}
	}

	// get unique archetypes
	uniqueArchetypalReferences := make([]string, len(archetypalReferences))
	ind := 0
	for ref := range archetypalReferences {
		uniqueArchetypalReferences[ind] = ref
		ind++
	}
	sort.Strings(uniqueArchetypalReferences) // needed as source of this array was a map and maps don't have predictable sorting - tests will fail randomly without sort

	if len(archetypalReferences) > 0 {
		log.Printf("Could not find the following in DB: %v. Potentially archetypes?", archetypalReferences)
	}

	return namedReferences, referenceOccurrence, uniqueArchetypalReferences
}

func cleanupToken(token *string) {
	*token = strings.TrimSpace(*token)
	*token = strings.ReplaceAll(*token, "\".", "")
	*token = strings.ReplaceAll(*token, "\".", "")
	*token = strings.ReplaceAll(*token, "'.", "")
	*token = strings.ReplaceAll(*token, "',", "")

	*token = strings.Trim(*token, "'")
	*token = strings.Trim(*token, "\"")
}
