package marketplace

import (
	"reflect"

	marketplace_module "github.com/RiemannChow/nantianmen-ai-gateway/module/marketplace"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/marketplace/dto"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IMarketplaceController interface {
	Catalog(ctx *gin.Context) ([]*marketplace_module.InstalledView, error)
	Install(ctx *gin.Context, id string) error
	Enable(ctx *gin.Context, id string, enabled bool) error
	GetRevenueShare(ctx *gin.Context) (map[string]float64, error)
	SetRevenueShare(ctx *gin.Context, input *dto.RevenueShare) error
}

func init() {
	autowire.Auto[IMarketplaceController](func() reflect.Value {
		return reflect.ValueOf(new(imlMarketplaceController))
	})
}
