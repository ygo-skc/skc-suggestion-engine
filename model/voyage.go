package model

type VoyageInputType string

const (
	VoyageQueryInput    VoyageInputType = "query"
	VoyageDocumentInput VoyageInputType = "document"
)

type EmbeddingRequest struct {
	Input           []string        `json:"input"`
	Model           string          `json:"model"`
	InputType       VoyageInputType `json:"input_type,omitempty"`
	OutputDimension int             `json:"output_dimension,omitempty"`
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []Data `json:"data"`
	Model  string `json:"model"`
}

type Data struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type RerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Model     string   `json:"model"`
	TopK      uint8    `json:"top_k"`
}

type RerankResponse struct {
	Data []RerankResults `json:"data"`
}

type RerankResults struct {
	Index uint    `json:"index"`
	Score float64 `json:"relevance_score"`
}
