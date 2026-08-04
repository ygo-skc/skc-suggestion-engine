package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	cModel "github.com/ygo-skc/skc-go/common/v3/model"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
	"github.com/ygo-skc/skc-suggestion-engine/downstream"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	similarCardsOp   = "Similar Cards"
	semanticSearchOp = "Semantic Card Search"
)

func getSimilarCardsHandler(res http.ResponseWriter, req *http.Request) {
	cardID := chi.URLParam(req, "cardID")

	logger, ctx := cUtil.InitRequest(req.Context(), apiName, similarCardsOp, slog.String("card_id", cardID))
	logger.Info("Finding similar cards")

	cardProto, err := downstream.YGO.CardService.GetCardByIDProto(ctx, cardID)
	if err != nil {
		err.HandleServerResponse(res)
		return
	}
	subject := cModel.YGOCardRESTFromProto(cardProto)

	embeddedQuery, err := embedQuery(ctx, subject.GetEffect())
	if err != nil {
		logger.Error("Could not embed card text", slog.Any("err", err))
		err.HandleServerResponse(res)
		return
	}

	similarCards := model.SimilarCards{Card: subject}
	fetchResults := func() ([]model.VectorSearchResult, *cModel.APIError) {
		return skcSuggestionEngineDBInterface.SearchSimilarCards(ctx, subject, embeddedQuery)
	}
	if matches, err := getSimilarCards(ctx, subject.GetEffect(), fetchResults); err != nil {
		logger.Error("Could not retrieve similar cards", slog.Any("err", err))
		err.HandleServerResponse(res)
		return
	} else {
		similarCards.Matches = matches
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(similarCards); err != nil {
		logger.Error("Could not encode card similarity response", slog.Any("err", err), slog.String("card_id", cardID))
	}
}

func getSemanticCardResultsHandler(res http.ResponseWriter, req *http.Request) {
	query := strings.TrimSpace(req.URL.Query().Get("q"))

	logger, ctx := cUtil.InitRequest(req.Context(), apiName, semanticSearchOp, slog.String("query", query))
	logger.Info("Performing semantic keyword search")

	if query == "" {
		logger.Error("Rejecting semantic search with empty query")
		badRequest := &cModel.APIError{Message: "Query parameter 'q' is required.", StatusCode: http.StatusBadRequest}
		badRequest.HandleServerResponse(res)
		return
	}

	embeddedQuery, err := embedQuery(ctx, query)
	if err != nil {
		logger.Error("Could not embed search query", slog.Any("err", err))
		err.HandleServerResponse(res)
		return
	}

	searchResults := model.SemanticSearchResults{Query: query}
	fetchResults := func() ([]model.VectorSearchResult, *cModel.APIError) {
		return skcSuggestionEngineDBInterface.SemanticKeywordSearch(ctx, query, embeddedQuery)
	}
	if matches, err := getSimilarCards(ctx, query, fetchResults); err != nil {
		logger.Error("Could not retrieve semantic search results", slog.Any("err", err))
		err.HandleServerResponse(res)
		return
	} else {
		searchResults.Matches = matches
	}

	res.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(res).Encode(searchResults); err != nil {
		logger.Error("Could not encode semantic search response", slog.Any("err", err))
	}
}

func embedQuery(ctx context.Context, query string) ([]float32, *cModel.APIError) {
	voyageRes, err := downstream.EmbedText(ctx, []string{query}, model.VoyageQueryInput)
	if err != nil {
		return nil, err
	}

	return voyageRes.Data[0].Embedding, nil
}

func getSimilarCards(ctx context.Context, query string,
	fetchSearchResults func() ([]model.VectorSearchResult, *cModel.APIError)) ([]cModel.YGOCard, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)

	vectorSearchResults, err := fetchSearchResults()
	if err != nil {
		return nil, err
	}

	vectorSearchResults, err = rerank(ctx, vectorSearchResults, query, 20)
	if err != nil {
		logger.Error("Error during re-ranking", slog.Any("err", err))
		return nil, err
	}

	similarCardIDs := make(cModel.CardIDs, 0, len(vectorSearchResults))
	for _, vectorSearchResult := range vectorSearchResults {
		similarCardIDs = append(similarCardIDs, vectorSearchResult.ID)
	}

	cardsProto, err := downstream.YGO.CardService.GetCardsByIDProto(ctx, similarCardIDs)
	if err != nil {
		logger.Error("Could not retrieve information about cards from search results", slog.Any("err", err))
		return nil, err
	}
	similarCardData := cModel.BatchCardDataFromProto[cModel.CardIDs](cardsProto, cModel.CardIDAsKey)

	if len(similarCardData.UnknownResources) > 0 {
		logger.Warn("Some vector search IDs had no matching metadata", slog.Any("unknown_card_ids", similarCardData.UnknownResources))
	}

	similarCards := make([]cModel.YGOCard, 0, len(similarCardIDs))
	for _, id := range similarCardIDs {
		if card, isPresent := similarCardData.CardInfo[id]; isPresent {
			similarCards = append(similarCards, card)
		}
	}

	return similarCards, nil
}

func rerank(ctx context.Context, vectorSearchResults []model.VectorSearchResult, query string, topK uint8) ([]model.VectorSearchResult, *cModel.APIError) {
	docs := make([]string, len(vectorSearchResults))
	for i, vectorSearchResult := range vectorSearchResults {
		docs[i] = vectorSearchResult.Text
	}

	voyageRes, err := downstream.RerankVectorResults(ctx, docs, query, topK)
	if err != nil {
		return nil, err
	}

	rankedResults := make([]model.VectorSearchResult, 0, topK)
	for _, rerankResult := range voyageRes.Data {
		rankedResults = append(rankedResults, vectorSearchResults[rerankResult.Index])
	}

	return rankedResults, nil
}
