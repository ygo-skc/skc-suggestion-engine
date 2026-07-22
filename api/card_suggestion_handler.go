package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
	"github.com/ygo-skc/skc-suggestion-engine/suggest"
)

const (
	cardSuggestionsOp = "Card Suggestions"
)

// Handler that will be used by suggestion endpoint.
// Will retrieve fusion, synchro, etc materials and other references if they are explicitly mentioned by name and their name exists in the DB.
func getCardSuggestionsHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(req.Context(), apiName, cardSuggestionsOp, slog.String("card_id", cardID))
	logger.Info("Card suggestions requested")

	cardProto, err := downstream.YGO.CardService.GetCardByIDProto(ctx, cardID)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}
	cardToGetSuggestionsFor := cModel.YGOCardRESTFromProto(cardProto)

	ccIDs, relevantArchetypes, err := suggest.FetchMetadata(ctx, []string{cardID}, skcSuggestionEngineDBInterface)
	if err != nil {
		logger.Error("Failed to retrieve suggestion metadata", "err", err)
		err.HandleServerResponse(res)
		return
	}
	// TODO: include exclusions?

	suggestions := getCardSuggestions(ctx, cardToGetSuggestionsFor, ccIDs.GetValues(), relevantArchetypes)

	logger.Info("Card suggestions generated",
		"card_name", (cardToGetSuggestionsFor).GetName(),
		"named_materials", len(suggestions.NamedMaterials),
		"named_references", len(suggestions.NamedReferences))

	if err := json.NewEncoder(res).Encode(suggestions); err != nil {
		logger.Error("Could not encode card suggestions response", "err", err, "card_id", cardID)
	}
}

func getCardSuggestions(ctx context.Context, subject cModel.YGOCard, ccIDs map[string]uint32, relevantArchetypes []string) model.CardSuggestions {
	usd := suggest.GenerateUnparsedSuggestionData(ctx,
		suggest.QuotedStringRegex.FindAllString(subject.GetEffect(), -1), relevantArchetypes)

	materialText := cModel.GetPotentialMaterialsAsString(subject)
	effectText := strings.ReplaceAll(subject.GetEffect(), materialText, "")

	suggestions := suggest.ParseSuggestionData(subject.GetName(), materialText, effectText, usd)
	suggestions.Card = subject
	suggestions.RelevantArchetypes = relevantArchetypes // this is all archetypes the card belongs to

	slices.SortStableFunc(suggestions.NamedMaterials, suggest.SortCardReferences(ccIDs))
	slices.SortStableFunc(suggestions.NamedReferences, suggest.SortCardReferences(ccIDs))
	slices.Sort(suggestions.RelevantArchetypes)
	slices.Sort(suggestions.ReferencedArchetypes)
	slices.Sort(suggestions.MaterialArchetypes)
	suggestions.HasSelfReference = model.RemoveSelfReference(subject.GetName(), &suggestions.NamedReferences)

	return suggestions
}
