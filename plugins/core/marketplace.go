package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"

	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) marketplaceAPIs() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/marketplace/catalog", []string{"context"}, []string{"list"}, p.marketplaceController.Catalog, access.SystemSettingsMarketplaceView),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/marketplace/install", []string{"context", "query:id"}, nil, p.marketplaceController.Install, access.SystemSettingsMarketplaceManager),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/marketplace/enable", []string{"context", "query:id", "query:enabled"}, nil, p.marketplaceController.Enable, access.SystemSettingsMarketplaceManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/marketplace/revenue-share", []string{"context"}, []string{"shares"}, p.marketplaceController.GetRevenueShare, access.SystemSettingsMarketplaceManager),
		pm3.CreateApiWidthDoc(http.MethodPut, "/api/v1/marketplace/revenue-share", []string{"context", "body"}, nil, p.marketplaceController.SetRevenueShare, access.SystemSettingsMarketplaceManager),
	}
}
