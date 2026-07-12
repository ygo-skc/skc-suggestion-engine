package model

type EmbeddingRequest struct {
	Input           []string `json:"input"`
	Model           string   `json:"model"`
	InputType       string   `json:"input_type,omitempty"`
	OutputDimension int      `json:"output_dimension,omitempty"`
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
