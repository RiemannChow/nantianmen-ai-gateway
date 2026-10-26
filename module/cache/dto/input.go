package dto

// Config is the semantic cache configuration persisted in KV.
type Config struct {
	Enabled           bool   `json:"enabled"`
	TTLSeconds        int64  `json:"ttl_seconds"`
	EmbeddingEndpoint string `json:"embedding_endpoint"`
	EmbeddingKey      string `json:"embedding_key"`
}

// Query asks whether a normalized prompt is already cached.
type Query struct {
	Tenant string `json:"tenant"`
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// Store inserts a cached response for a prompt/model pair.
type Store struct {
	Tenant     string `json:"tenant"`
	Model      string `json:"model"`
	Prompt     string `json:"prompt"`
	Response   string `json:"response"`
	TokenCount int64  `json:"token_count"`
}
