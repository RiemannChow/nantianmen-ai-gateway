package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) multiTenantApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/multitenant/tenants", []string{"context"}, []string{"tenants"}, p.multiTenantController.GetTenants, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/multitenant/tenants", []string{"context", "body"}, []string{"tenants"}, p.multiTenantController.SaveTenants, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/multitenant/tenant/:id/usage", []string{"context", "path:id", "query:start", "query:end"}, []string{"usage"}, p.multiTenantController.TenantUsage, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/multitenant/tenant/:id/billing", []string{"context", "path:id", "query:start", "query:end"}, []string{"billing"}, p.multiTenantController.TenantBilling, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/multitenant/tenant/:id/status", []string{"context", "path:id", "query:start", "query:end"}, []string{"status"}, p.multiTenantController.TenantStatus, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/multitenant/billing", []string{"context", "query:start", "query:end"}, []string{"billing"}, p.multiTenantController.TenantsBilling, access.SystemSettingsApiGatewayView),
	}
}
