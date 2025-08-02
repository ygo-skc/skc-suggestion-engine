package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

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

func getCardSuggestions(ctx context.Context, cardToGetSuggestionsFor cModel.YGOCard,
	ccIDs map[string]uint32) model.CardSuggestions {
	suggestions := model.CardSuggestions{Card: cardToGetSuggestionsFor}
	materialString := cModel.GetPotentialMaterialsAsString(cardToGetSuggestionsFor)

	wg := sync.WaitGroup{}

	// get materials if card is from extra deck
	if cModel.IsExtraDeckMonster(cardToGetSuggestionsFor) {
		wg.Add(2)
		go getMaterialRefs(ctx, &suggestions, materialString, ccIDs, &wg)
	} else {
		wg.Add(1)
		suggestions.NamedMaterials = []model.CardReference{}
		suggestions.MaterialArchetypes = []string{}

		cUtil.RetrieveLogger(ctx).Debug("Not and extra deck monster")
	}
	go getNonMaterialRefs(ctx, &suggestions, cardToGetSuggestionsFor, materialString, ccIDs, &wg)

	wg.Wait()
	return suggestions
}

type supportData struct {
	namedReferences     cModel.CardDataMap
	referenceOccurrence map[string]int
	archetypeSet        map[string]struct{}
	cardIdByToken       map[string]string
}

// get unique archetypes
func (sd supportData) uniqueArchetypes() []string {
	uniqueArchetypes := make([]string, 0, len(sd.archetypeSet))
	for ref := range sd.archetypeSet {
		uniqueArchetypes = append(uniqueArchetypes, ref)
	}
	sort.Strings(uniqueArchetypes) // needed as source of this array was a map and maps don't have predictable sorting - tests will fail randomly without sort
	return uniqueArchetypes
}

func (sd supportData) cardReferences() []model.CardReference {
	uniqueReferences := make([]model.CardReference, 0, len(sd.namedReferences))
	for _, card := range sd.namedReferences {
		uniqueReferences = append(uniqueReferences, model.CardReference{Card: card, Occurrences: sd.referenceOccurrence[card.GetID()]})
	}
	return uniqueReferences
}

func getMaterialRefs(ctx context.Context, s *model.CardSuggestions, materialString string, ccIDs map[string]uint32, wg *sync.WaitGroup) {
	defer wg.Done()
	sd := generateSupportData(ctx, quotedStringRegex.FindAllString(materialString, -1))
	s.NamedMaterials, s.MaterialArchetypes = sd.cardReferences(), sd.uniqueArchetypes()
	sortCardReferences(&s.NamedMaterials, ccIDs)
}

// get named references - excludes materials
// will also check and remove self references
func getNonMaterialRefs(ctx context.Context, s *model.CardSuggestions, cardToGetSuggestionsFor cModel.YGOCard, materialString string,
	ccIDs map[string]uint32, wg *sync.WaitGroup) {
	defer wg.Done()
	sd := generateSupportData(ctx, quotedStringRegex.FindAllString(strings.ReplaceAll(cardToGetSuggestionsFor.GetEffect(), materialString, ""), -1))
	s.NamedReferences, s.ReferencedArchetypes = sd.cardReferences(), sd.uniqueArchetypes()
	s.HasSelfReference = model.RemoveSelfReference(cardToGetSuggestionsFor.GetName(), &s.NamedReferences)
	sortCardReferences(&s.NamedReferences, ccIDs)
}

// cycles through tokens - makes DB calls where necessary and attempts to build objects containing direct references (and their occurrences), archetype references
func generateSupportData(ctx context.Context, tokens []string) supportData {
	sd := supportData{namedReferences: cModel.CardDataMap{}, referenceOccurrence: map[string]int{}, archetypeSet: make(map[string]struct{})}

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

			// if token mapped to a cardId in previous loop, increase number of occurrences by 1 and skip any other processing this iteration as we already did the processing before
			if _, isPresent := tokenToCardId[token]; isPresent {
				sd.referenceOccurrence[tokenToCardId[token]] += 1
				continue
			}

			if card, isPresent := batchCardData.CardInfo[token]; !isPresent {
				// add occurrence of archetype to set
				sd.archetypeSet[token] = struct{}{}
			} else {
				// add occurrence of referenced card to maps
				sd.namedReferences[card.GetID()] = card
				sd.referenceOccurrence[card.GetID()] = 1
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
