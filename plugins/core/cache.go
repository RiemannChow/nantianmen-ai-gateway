package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"

	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) cacheAPIs() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/ai/cache/config", []string{"context"}, []string{"config"}, p.cacheController.GetConfig, access.SystemSettingsAiCacheView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/ai/cache/config", []string{"context", "body"}, nil, p.cacheController.SetConfig, access.SystemSettingsAiCacheManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/ai/cache/stats", []string{"context"}, []string{"stats"}, p.cacheController.Stats, access.SystemSettingsAiCacheView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/ai/cache/clear", []string{"context"}, nil, p.cacheController.Clear, access.SystemSettingsAiCacheManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/ai/cache/entries", []string{"context"}, []string{"list"}, p.cacheController.List, access.SystemSettingsAiCacheView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/ai/cache/query", []string{"context", "body"}, []string{"entry"}, p.cacheController.Query, access.SystemSettingsAiCacheView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/ai/cache/store", []string{"context", "body"}, nil, p.cacheController.Store, access.SystemSettingsAiCacheManager),
	}
}
