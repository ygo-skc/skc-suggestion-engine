package suggest

import (
	"context"
	"regexp"
	"slices"
	"sync"

	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	"github.com/ygo-skc/skc-go/common/v2/parser"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-go/common/v2/ygo"
	"github.com/ygo-skc/skc-suggestion-engine/db"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

var (
	QuotedStringRegex = regexp.MustCompile("^(\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')|[\\W](\"[ \\w\\d-:@,'.]{3,}?\"|'[ \\w\\d-:@,'.]{3,}?')")
)

type UnparsedSuggestionData struct {
	namedReferencesByToken cModel.CardDataMap
	archetypeSet           map[string]struct{}
}

func FetchMetadata(ctx context.Context, subjects []string, dbInterface db.SKCSuggestionEngineDAO) (*ygo.CardColors, []string, *cModel.APIError) {
	type archetypeRes struct {
		archetypes []string
		err        *cModel.APIError
	}
	type cardColorRes struct {
		ccIDs *ygo.CardColors
		err   *cModel.APIError
	}
	var wg sync.WaitGroup

	// archetype routine
	archetypeAWG := cUtil.NewAtomicWaitGroup[archetypeRes](&wg)
	go func(awg *cUtil.AtomicWaitGroup[archetypeRes]) {
		relevantArchetypes, err := dbInterface.GetRelevantArchetypes(ctx, subjects)
		res := archetypeRes{
			archetypes: relevantArchetypes,
			err:        err,
		}
		awg.Store(&res)
	}(archetypeAWG)

	// card color routine
	cardColorAWG := cUtil.NewAtomicWaitGroup[cardColorRes](&wg)
	go func(awg *cUtil.AtomicWaitGroup[cardColorRes]) {
		ccIDs, err := downstream.YGO.CardService.GetCardColorsProto(ctx)
		res := cardColorRes{
			ccIDs: ccIDs,
			err:   err,
		}
		awg.Store(&res)
	}(cardColorAWG)

	archetypeRoutine := archetypeAWG.Load()
	if archetypeRoutine.err != nil {
		return nil, nil, archetypeRoutine.err
	}

	cardColorRoutine := cardColorAWG.Load()
	if cardColorRoutine.err != nil {
		return nil, nil, cardColorRoutine.err
	}

	return cardColorRoutine.ccIDs, archetypeRoutine.archetypes, nil
}

// parses suggestion data by transforming it into a CardSuggestion object
func ParseSuggestionData(cardName string, materialText string, effectText string, usd UnparsedSuggestionData) model.CardSuggestions {
	numReferences := len(usd.namedReferencesByToken)
	suggestions := model.CardSuggestions{
		NamedMaterials:  make([]model.CardReference, 0, numReferences),
		NamedReferences: make([]model.CardReference, 0, numReferences),
	}

	var nonArchetypeMaterialTokens, nonArchetypeReferenceTokens map[string]int
	suggestions.MaterialArchetypes, nonArchetypeMaterialTokens = partitionTokensByCardText(materialText, usd)
	suggestions.ReferencedArchetypes, nonArchetypeReferenceTokens = partitionTokensByCardText(effectText, usd)

	parseTokenAsCard(nonArchetypeMaterialTokens, usd.namedReferencesByToken, &suggestions.NamedMaterials)
	parseTokenAsCard(nonArchetypeReferenceTokens, usd.namedReferencesByToken, &suggestions.NamedReferences)

	return suggestions
}

// Uses card text and archetypes to create a list of unique archetypes and a map of non archetype tokens and their occurrence
func partitionTokensByCardText(cardText string, usd UnparsedSuggestionData) ([]string, map[string]int) {
	archetypeTokens := make([]string, 0, len(usd.archetypeSet))
	nonArchetypeTokens := make(map[string]int, len(usd.namedReferencesByToken))

	for _, token := range QuotedStringRegex.FindAllString(cardText, -1) {
		parser.CleanupToken(&token)
		if _, exists := usd.archetypeSet[token]; exists && !slices.Contains(archetypeTokens, token) {
			archetypeTokens = append(archetypeTokens, token)
		}

		if _, exists := usd.namedReferencesByToken[token]; exists {
			nonArchetypeTokens[token]++
		}
	}
	return archetypeTokens, nonArchetypeTokens
}

// creates the suggestion references and their occurrence
func parseTokenAsCard(tokenOccurrences map[string]int, namedReferencesByToken cModel.CardDataMap, references *[]model.CardReference) {
	for token, occurrence := range tokenOccurrences {
		*references = append(*references, model.CardReference{Occurrences: occurrence, Card: namedReferencesByToken[token]})
	}
}

// cycles through tokens - makes DB calls where necessary and attempts to build objects containing direct references (and their occurrences), archetype references
func GenerateUnparsedSuggestionData(ctx context.Context, tokens []string, relevantArchetypes []string) UnparsedSuggestionData {
	usd := UnparsedSuggestionData{namedReferencesByToken: cModel.CardDataMap{}, archetypeSet: make(map[string]struct{})}

	for _, archetype := range relevantArchetypes {
		usd.archetypeSet[archetype] = struct{}{}
	}

	totalTokens := len(tokens)

	if totalTokens != 0 {
		for i := range totalTokens {
			parser.CleanupToken(&tokens[i])
		}

		batchCardData, _ := downstream.YGO.CardService.GetCardsByName(ctx, tokens)

		for _, token := range tokens {
			if card, isPresent := batchCardData.CardInfo[token]; isPresent {
				usd.namedReferencesByToken[token] = card
			}
		}
	}

	return usd
}
