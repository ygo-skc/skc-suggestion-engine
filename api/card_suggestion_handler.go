package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	json "github.com/goccy/go-json"
	cModel "github.com/ygo-skc/skc-go/common/model"
	cUtil "github.com/ygo-skc/skc-go/common/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	quotedStringRegex = regexp.MustCompile("^(\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')|[\\W](\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')")
)

const (
	cardSuggestionsOp = "Card Suggestions"
)

type suggestionData struct {
	namedReferencesByToken cModel.CardDataMap
	archetypeSet           map[string]struct{}
	cardIdByToken          map[string]string
}

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getCardSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(context.Background(), apiName, cardSuggestionsOp, slog.String("card_id", cardID))
	logger.Info("Card suggestions requested")

	if cardToGetSuggestionsFor, err := downstream.YGO.CardService.GetCardByID(ctx, cardID); err != nil {
		err.HandleServerResponse(res)
		return
	} else {
		ccIDs, _ := downstream.YGO.CardService.GetCardColorsProto(ctx) // retrieve card color IDs
		suggestions := getCardSuggestions(ctx, *cardToGetSuggestionsFor, ccIDs.Values)

		logger.Info(fmt.Sprintf("%s: %d unique material references - %d unique named references",
			(*cardToGetSuggestionsFor).GetName(),
			len(suggestions.NamedMaterials), len(suggestions.NamedReferences)))

		json.NewEncoder(res).Encode(suggestions)
	}
}

func getCardSuggestions(ctx context.Context, subject cModel.YGOCard, ccIDs map[string]uint32) model.CardSuggestions {
	suggestions := model.CardSuggestions{
		Card:                 subject,
		NamedMaterials:       make([]model.CardReference, 0, 5),
		NamedReferences:      make([]model.CardReference, 0, 5),
		MaterialArchetypes:   make([]string, 0, 5),
		ReferencedArchetypes: make([]string, 0, 5),
	}

	sd := generateSuggestionData(ctx, quotedStringRegex.FindAllString(subject.GetEffect(), -1))
	materialText := cModel.GetPotentialMaterialsAsString(subject)
	effectText := strings.ReplaceAll(subject.GetEffect(), materialText, "")
	parseSuggestionData(materialText, effectText, sd, &suggestions)

	sortCardReferences(&suggestions.NamedReferences, ccIDs)
	sortCardReferences(&suggestions.NamedMaterials, ccIDs)
	sort.Strings(suggestions.ReferencedArchetypes)
	sort.Strings(suggestions.MaterialArchetypes)
	suggestions.HasSelfReference = model.RemoveSelfReference(subject.GetName(), &suggestions.NamedReferences)
	return suggestions
}

func parseSuggestionData(materialText string, effectText string, sd suggestionData, suggestions *model.CardSuggestions) {
	var edTokens = make(map[string]int)
	for _, token := range quotedStringRegex.FindAllString(materialText, -1) {
		cModel.CleanupToken(&token)
		edTokens[token]++
	}

	var nonEDTokens = make(map[string]int)
	for _, token := range quotedStringRegex.FindAllString(effectText, -1) {
		cModel.CleanupToken(&token)
		nonEDTokens[token]++
	}

	for archetype := range sd.archetypeSet {
		if _, exists := edTokens[archetype]; exists {
			suggestions.MaterialArchetypes = append(suggestions.MaterialArchetypes, archetype)
		}
		if _, exists := nonEDTokens[archetype]; exists {
			suggestions.ReferencedArchetypes = append(suggestions.ReferencedArchetypes, archetype)
		}
	}

	for _, cardRef := range sd.namedReferencesByToken {
		if occurrence, exists := edTokens[cardRef.GetName()]; exists {
			suggestions.NamedMaterials = append(suggestions.NamedMaterials, model.CardReference{Card: cardRef, Occurrences: occurrence})
		}
		if occurrence, exists := nonEDTokens[cardRef.GetName()]; exists {
			suggestions.NamedReferences = append(suggestions.NamedReferences, model.CardReference{Card: cardRef, Occurrences: occurrence})
		}
	}
}

// cycles through tokens - makes DB calls where necessary and attempts to build objects containing direct references (and their occurrences), archetype references
func generateSuggestionData(ctx context.Context, tokens []string) suggestionData {
	sd := suggestionData{namedReferencesByToken: cModel.CardDataMap{}, archetypeSet: make(map[string]struct{})}

	tokenToCardId := map[string]string{} // maps token to its cardID - token will only have cardID if token is found in DB
	totalTokens := len(tokens)

	if totalTokens != 0 {
		for i := range totalTokens {
			cModel.CleanupToken(&tokens[i])
		}

		batchCardData, _ := downstream.YGO.CardService.GetCardsByName(ctx, tokens)

		for _, token := range tokens {
			// if token is present in archetype slice, skip token
			if _, isPresent := sd.archetypeSet[token]; isPresent {
				continue
			}

			// already processed
			if _, isPresent := tokenToCardId[token]; isPresent {
				continue
			}

			if card, isPresent := batchCardData.CardInfo[token]; !isPresent {
				// add occurrence of archetype to set
				sd.archetypeSet[token] = struct{}{}
			} else {
				// add occurrence of referenced card to maps
				sd.namedReferencesByToken[token] = card
				tokenToCardId[token] = card.GetID()
			}
		}
	}

	return sd
}

func sortCardReferences(cr *[]model.CardReference, ccIDs map[string]uint32) {
	// sorting alphabetically from a-z
	sort.SliceStable(*cr, func(i, j int) bool {
		return (*cr)[i].Card.GetName() < (*cr)[j].Card.GetName()
	})

	// sorting by card color
	sort.SliceStable(*cr, func(i, j int) bool {
		return ccIDs[(*cr)[i].Card.GetColor()] < ccIDs[(*cr)[j].Card.GetColor()]
	})
}
