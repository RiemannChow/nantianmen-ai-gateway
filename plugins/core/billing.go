package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) billingApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/billing/pricing", []string{"context"}, []string{"pricing"}, p.billingController.GetPricing, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/billing/pricing", []string{"context", "body"}, []string{"pricing"}, p.billingController.SavePricing, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/billing/rebate", []string{"context"}, []string{"rebate"}, p.billingController.GetRebate, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/billing/rebate", []string{"context", "body"}, []string{"rebate"}, p.billingController.SaveRebate, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/billing/overview", []string{"context", "query:start", "query:end"}, []string{"billing"}, p.billingController.Overview, access.SystemSettingsApiGatewayView),
	}
}
