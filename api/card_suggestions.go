package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	quotedStringRegex            = regexp.MustCompile("\"[^., ].*?\"")
	deckListCardAndQuantityRegex = regexp.MustCompile("[1-3][xX][0-9]{8}")
)

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	pathVars := mux.Vars(req)
	cardID := pathVars["cardID"]
	log.Println("Getting suggestions for card:", cardID)

	if cardToGetSuggestionsFor, err := db.FindDesiredCardInDBUsingID(cardID); err != nil {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusNotFound)

		json.NewEncoder(res).Encode(err)
	} else {
		// clean up card effect
		cardToGetSuggestionsFor.CardEffect = strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, "'", "\"")

		suggestions := model.CardSuggestions{Card: cardToGetSuggestionsFor}
		var materialString string

		// get materials if card is from extra deck
		if cardToGetSuggestionsFor.IsExtraDeckMonster() {
			materialString = cardToGetSuggestionsFor.GetPotentialMaterialsAsString()
			suggestions.NamedMaterials, suggestions.MaterialArchetypes = getReferences(materialString)
			log.Printf("Found %d unique material references", len(*suggestions.NamedMaterials))
		} else {
			log.Printf("%s is not an ED monster", cardToGetSuggestionsFor.CardID)
		}

		// get named references - excludes materials
		// will also check and remove for self references
		suggestions.NamedReferences, suggestions.ReferencedArchetypes = getReferences(strings.ReplaceAll(cardToGetSuggestionsFor.CardEffect, materialString, ""))
		suggestions.HasSelfReference = removeSelfReference(cardToGetSuggestionsFor.CardName, suggestions.NamedReferences)
		log.Printf("Found %d unique named references", len(*suggestions.NamedReferences))

		// get decks that feature card
		suggestions.Decks, _ = db.GetDecksThatFeatureCards([]string{cardID})

		res.Header().Add("Content-Type", "application/json")
		json.NewEncoder(res).Encode(suggestions)
	}
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
	var archetypalReferences []string

	for _, token := range tokens {
		token = strings.ReplaceAll(token, "\"", "")

		if card, err := db.FindDesiredCardInDBUsingName(token); err != nil {
			archetypalReferences = append(archetypalReferences, token)
		} else {
			namedReferences[card.CardID] = card

			if _, isPresent := referenceOccurrence[card.CardID]; !isPresent {
				referenceOccurrence[card.CardID] = 0
			}
			referenceOccurrence[card.CardID] += 1
		}
	}

	if len(archetypalReferences) > 0 {
		log.Printf("Could not find the following in DB: %v. Potentially an archetype?", archetypalReferences)
	}

	return namedReferences, referenceOccurrence, archetypalReferences
}
