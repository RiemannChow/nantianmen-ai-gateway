package ai_balance

import (
	"reflect"

	ai_balance_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/ai-balance/dto"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IBalanceController interface {
	List(ctx *gin.Context, keyword string) ([]*ai_balance_dto.Item, error)
	Sort(ctx *gin.Context, input *ai_balance_dto.Sort) error
	Create(ctx *gin.Context, input *ai_balance_dto.Create) error
	Delete(ctx *gin.Context, id string) error
	Probe(ctx *gin.Context, id string) error
	BatchProbe(ctx *gin.Context) error
	GetAutoFailover(ctx *gin.Context) (bool, error)
	SetAutoFailover(ctx *gin.Context, enabled bool) error
}

func init() {
	autowire.Auto[IBalanceController](func() reflect.Value {
		return reflect.ValueOf(new(imlBalanceController))
	})
}
