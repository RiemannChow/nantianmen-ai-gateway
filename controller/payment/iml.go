package payment

import (
	"io"
	"strings"

	payment_module "github.com/RiemannChow/nantianmen-ai-gateway/module/payment"
	"github.com/gin-gonic/gin"
)

var _ IPaymentController = (*imlPaymentController)(nil)

type imlPaymentController struct {
	module payment_module.IPaymentModule `autowired:""`
}

func (i *imlPaymentController) CreateOrder(ctx *gin.Context, input *payment_module.CreateInput) (*payment_module.OrderView, error) {
	return i.module.CreateOrder(ctx, input)
}

func (i *imlPaymentController) GetOrder(ctx *gin.Context, id string) (*payment_module.OrderView, error) {
	return i.module.GetOrder(ctx, id)
}

func (i *imlPaymentController) ListOrders(ctx *gin.Context) ([]*payment_module.OrderView, error) {
	return i.module.ListOrders(ctx)
}

// Webhook is a raw gin handler so it can read the unparsed body and headers
// needed for signature verification. The path carries :provider.
func (i *imlPaymentController) Webhook(ctx *gin.Context) {
	providerType := ctx.Param("provider")
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "read body failed"})
		return
	}
	headers := map[string]string{}
	for k, v := range ctx.Request.Header {
		if len(v) > 0 {
			headers[k] = strings.Join(v, ",")
		}
	}
	if err := i.module.Webhook(ctx, providerType, body, headers); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, gin.H{"status": "ok"})
}
