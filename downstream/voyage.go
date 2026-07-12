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

const voyageEmbeddingsPath = "embeddings"

var voyageEmbeddingErr = &cModel.APIError{Message: "Error occurred while generating embeddings", StatusCode: http.StatusInternalServerError}

var voyageBaseURL = &url.URL{Scheme: "https", Host: "api.voyageai.com", Path: "/v1"}

var voyageHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          30,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	},
}

func newVoyageRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)

	req, err := http.NewRequestWithContext(ctx, method, voyageBaseURL.JoinPath(path).String(), body)
	if err != nil {
		logger.Error("Error building Voyage request", "err", err, "path", path)
		return nil, voyageEmbeddingErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cUtil.EnvMap["VOYAGE_API_KEY"])
	return req, nil
}

func GetEmbeddings(ctx context.Context, input []string, inputType string) (*model.EmbeddingResponse, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	logger.Info("Fetching embeddings from Voyage AI")

	reqBody := model.EmbeddingRequest{
		Input:           input,
		Model:           "voyage-4",
		InputType:       inputType,
		OutputDimension: 512,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("Error marshalling Voyage embedding request", "err", err)
		return nil, voyageEmbeddingErr
	}

	req, apiErr := newVoyageRequest(ctx, http.MethodPost, voyageEmbeddingsPath, bytes.NewReader(payload))
	if apiErr != nil {
		return nil, apiErr
	}

	voyageRes, err := voyageHTTPClient.Do(req)
	if err != nil {
		logger.Error("Error calling Voyage AI embeddings API", "err", err)
		return nil, voyageEmbeddingErr
	}

	body, apiErr := parseResponseBody(ctx, voyageRes)
	if apiErr != nil {
		return nil, apiErr
	}

	var result model.EmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error("Error unmarshalling Voyage AI embeddings response", "err", err)
		return nil, voyageEmbeddingErr
	}

	return &result, nil
}
