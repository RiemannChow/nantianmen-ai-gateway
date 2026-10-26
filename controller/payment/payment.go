package payment

import (
	"reflect"

	payment_module "github.com/RiemannChow/nantianmen-ai-gateway/module/payment"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IPaymentController interface {
	CreateOrder(ctx *gin.Context, input *payment_module.CreateInput) (*payment_module.OrderView, error)
	GetOrder(ctx *gin.Context, id string) (*payment_module.OrderView, error)
	ListOrders(ctx *gin.Context) ([]*payment_module.OrderView, error)
	Webhook(ctx *gin.Context)
}

func init() {
	autowire.Auto[IPaymentController](func() reflect.Value {
		return reflect.ValueOf(new(imlPaymentController))
	})
}
