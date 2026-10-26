package marketplace

import (
	marketplace_module "github.com/RiemannChow/nantianmen-ai-gateway/module/marketplace"
	marketplace_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/marketplace/dto"
	"github.com/gin-gonic/gin"
)

var _ IMarketplaceController = (*imlMarketplaceController)(nil)

type imlMarketplaceController struct {
	module marketplace_module.IMarketplaceModule `autowired:""`
}

func (i *imlMarketplaceController) Catalog(ctx *gin.Context) ([]*marketplace_module.InstalledView, error) {
	return i.module.Catalog(ctx)
}

func (i *imlMarketplaceController) Install(ctx *gin.Context, id string) error {
	return i.module.Install(ctx, id)
}

func (i *imlMarketplaceController) Enable(ctx *gin.Context, id string, enabled bool) error {
	return i.module.Enable(ctx, id, enabled)
}

func (i *imlMarketplaceController) GetRevenueShare(ctx *gin.Context) (map[string]float64, error) {
	return i.module.GetRevenueShare(ctx)
}

func (i *imlMarketplaceController) SetRevenueShare(ctx *gin.Context, input *marketplace_dto.RevenueShare) error {
	return i.module.SetRevenueShare(ctx, input.Shares)
}
