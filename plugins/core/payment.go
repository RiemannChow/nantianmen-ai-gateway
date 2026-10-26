package core

import (
	"net/http"

	"github.com/RiemannChow/nantianmen-ai-gateway/resources/access"

	"github.com/eolinker/go-common/ignore"
	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) paymentAPIs() []pm3.Api {
	// The webhook callback is public (no session); ignore the auth gate so the
	// provider can POST to it directly.
	ignore.IgnorePath("login", http.MethodPost, "/api/v1/payment/webhook/:provider")
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/payment/order", []string{"context", "body"}, []string{"order"}, p.paymentController.CreateOrder, access.SystemSettingsPaymentManager),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/payment/orders", []string{"context"}, []string{"list"}, p.paymentController.ListOrders, access.SystemSettingsPaymentView),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/payment/order", []string{"context", "query:id"}, []string{"order"}, p.paymentController.GetOrder, access.SystemSettingsPaymentView),
		pm3.CreateApiSimple(http.MethodPost, "/api/v1/payment/webhook/:provider", p.paymentController.Webhook),
	}
}
