package downstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	cModel "github.com/ygo-skc/skc-go/common/v2/model"
	cUtil "github.com/ygo-skc/skc-go/common/v2/util"
	"github.com/ygo-skc/skc-suggestion-engine/model"
)

const (
	voyageEmbeddingsPath = "embeddings"
	voyageRerankPath     = "rerank"

	voyageEmbeddingModel = "voyage-4"
	voyageRerankModel    = "rerank-2.5"
)

var (
	voyageBaseURL = &url.URL{Scheme: "https", Host: "api.voyageai.com", Path: "/v1"}

	voyageHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 1 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
		},
	}
)

func newVoyageEmbeddingErr() *cModel.APIError {
	return &cModel.APIError{Message: "Error occurred while generating embeddings", StatusCode: http.StatusInternalServerError}
}

func newVoyageRerankErr() *cModel.APIError {
	return &cModel.APIError{Message: "Error occurred while re-ranking", StatusCode: http.StatusInternalServerError}
}

func newVoyageRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)

	req, err := http.NewRequestWithContext(ctx, method, voyageBaseURL.JoinPath(path).String(), body)
	if err != nil {
		logger.Error("Error building Voyage request", "err", err, "path", path)
		return nil, &cModel.APIError{Message: "Error calling downstream service", StatusCode: http.StatusInternalServerError}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cUtil.EnvMap["VOYAGE_API_KEY"])
	return req, nil
}

func doVoyageRequest[T any](ctx context.Context, path string, reqBody any, newErr func() *cModel.APIError) (*T, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)

	payload, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Error marshalling Voyage request", "err", err, "path", path)
		return nil, newErr()
	}

	req, apiErr := newVoyageRequest(ctx, http.MethodPost, path, bytes.NewReader(payload))
	if apiErr != nil {
		return nil, apiErr
	}

	voyageRes, err := voyageHTTPClient.Do(req)
	if err != nil {
		logger.Error("Error calling Voyage API", "err", err, "path", path)
		return nil, newErr()
	}

	body, apiErr := parseResponseBody(ctx, voyageRes)
	if apiErr != nil {
		return nil, apiErr
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error unmarshalling Voyage response", "err", err, "path", path)
		return nil, newErr()
	}

	return &result, nil
}

func EmbedText(ctx context.Context, input []string, inputType model.VoyageInputType) (*model.EmbeddingResponse, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("Calling Voyage API to embed text")

	reqBody := model.EmbeddingRequest{
		Input:           input,
		Model:           voyageEmbeddingModel,
		InputType:       inputType,
		OutputDimension: 512,
	}

	result, apiErr := doVoyageRequest[model.EmbeddingResponse](ctx, voyageEmbeddingsPath, reqBody, newVoyageEmbeddingErr)
	if apiErr != nil {
		return nil, apiErr
	}

	if len(result.Data) != len(input) {
		logger.Error("Voyage API returned incorrect number of embeddings", "num_input", len(input), "num_embeddings", len(result.Data))
		return nil, newVoyageEmbeddingErr()
	}

	return result, nil
}

func RerankVectorResults(ctx context.Context, input []string, query string, topK uint8) (*model.RerankResponse, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("Calling Voyage API to rerank vector results")

	reqBody := model.RerankRequest{
		Query:     query,
		Documents: input,
		Model:     voyageRerankModel,
		TopK:      topK,
	}

	result, apiErr := doVoyageRequest[model.RerankResponse](ctx, voyageRerankPath, reqBody, newVoyageRerankErr)
	if apiErr != nil {
		return nil, apiErr
	}

	if expectedSize := min(int(topK), len(input)); len(result.Data) != expectedSize {
		logger.Error("Voyage API returned incorrect number of re-ranked elements", "expected_size", topK, "actual", len(result.Data))
		return nil, newVoyageRerankErr()
	}

	return result, nil
}
