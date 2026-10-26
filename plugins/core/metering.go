package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) meteringApis() []pm3.Api {
	return []pm3.Api{
		// local usage collection (Apinto / any collector pushes access-log events)
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/metering/events", []string{"context", "body"}, []string{"result"}, p.meteringController.Ingest, access.SystemSettingsApiGatewayManager),
		// generate representative sample usage so the FinOps suite is non-empty on a fresh install
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/metering/seed", []string{"context", "query:days"}, []string{"result"}, p.meteringController.Seed, access.SystemSettingsApiGatewayManager),
	}
}
