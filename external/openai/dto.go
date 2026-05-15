package openai

// embeddingRequest is the request body for the OpenAI Embeddings API.
type embeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format"`
}

// embeddingResponse is the response body from the OpenAI Embeddings API.
type embeddingResponse struct {
	Data  []embeddingData `json:"data"`
	Model string          `json:"model"`
	Usage embeddingUsage  `json:"usage"`
}

type embeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type embeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// apiError is the error response body from the OpenAI API.
type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
