package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) aiSecurityApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/aisecurity/rules", []string{"context"}, []string{"rules"}, p.aiSecurityController.GetRules, access.SystemSettingsApiGatewayView),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/aisecurity/rules", []string{"context", "body"}, []string{"rules"}, p.aiSecurityController.SaveRules, access.SystemSettingsApiGatewayManager),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/aisecurity/scan", []string{"context", "body"}, []string{"result"}, p.aiSecurityController.Scan, access.SystemSettingsApiGatewayManager),
	}
}
