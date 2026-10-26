package prompt_dto

// Save is the request body for creating or updating a prompt template.
// When Id is empty a new template is created with a generated id.
type Save struct {
	Id        string   `json:"id"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Content   string   `json:"content"`
	Variables []string `json:"variables"`
}

// Delete is the request body for removing a template by id.
type Delete struct {
	Id string `json:"id"`
}
