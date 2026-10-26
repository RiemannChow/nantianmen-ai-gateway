package prompt_dto

// Item is a single prompt template returned by the API.
type Item struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Content   string   `json:"content"`
	Variables []string `json:"variables"`
	UpdatedAt int64    `json:"updated_at"`
}
