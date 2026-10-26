package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) tracingApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/tracing/events", []string{"context", "body"}, []string{"result"}, p.tracingController.Ingest, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/tracing/seed", []string{"context", "query:days"}, []string{"result"}, p.tracingController.Seed, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/tracing/overview", []string{"context", "query:start", "query:end"}, []string{"overview"}, p.tracingController.Overview, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/tracing/list", []string{"context", "query:start", "query:end", "query:page", "query:page_size"}, []string{"list"}, p.tracingController.List, access.SystemSettingsApiGatewayView),
	}
}
