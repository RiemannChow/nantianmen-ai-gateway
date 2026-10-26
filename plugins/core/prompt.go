package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"

	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) promptAPIs() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/ai/prompts", []string{"context", "query:keyword"}, []string{"list"}, p.promptController.List, access.SystemSettingsAiPromptView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/ai/prompt", []string{"context", "body"}, []string{"template"}, p.promptController.Save, access.SystemSettingsAiPromptManager),
		pm3.CreateApiWidthDoc(http.MethodDelete, "/api/v1/ai/prompt", []string{"context", "query:id"}, nil, p.promptController.Delete, access.SystemSettingsAiPromptManager),
	}
}
