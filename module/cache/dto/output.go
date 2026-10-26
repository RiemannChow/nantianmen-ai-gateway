package dto

// Stats summarizes semantic cache usage.
type Stats struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hit_rate"`
	Entries int64   `json:"entries"`
}

// Entry is a single cached response.
type Entry struct {
	Key        string `json:"key"`
	Model      string `json:"model"`
	Tenant     string `json:"tenant"`
	PromptHash string `json:"prompt_hash"`
	Response   string `json:"response"`
	TokenCount int64  `json:"token_count"`
	CreatedAt  int64  `json:"created_at"`
}
