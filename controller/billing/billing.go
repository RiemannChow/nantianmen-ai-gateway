package billing

import (
	"reflect"
	"strconv"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/billing"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

// RebateInput wraps the per-provider channel-commission map so the request body
// is explicit: { "rebate": { "openai": 0.01, "azure": 0.02 } }.
type RebateInput struct {
	Rebate map[string]float64 `json:"rebate"`
}

type IBillingController interface {
	GetPricing(ctx *gin.Context) (*billing.PricingConfig, error)
	SavePricing(ctx *gin.Context, cfg *billing.PricingConfig) (*billing.PricingConfig, error)
	GetRebate(ctx *gin.Context) (*RebateInput, error)
	SaveRebate(ctx *gin.Context, input *RebateInput) (*RebateInput, error)
	Overview(ctx *gin.Context, start string, end string) (*billing.BillingOverview, error)
}

func init() {
	autowire.Auto[IBillingController](func() reflect.Value {
		return reflect.ValueOf(new(imlBillingController))
	})
}

var _ IBillingController = (*imlBillingController)(nil)

type imlBillingController struct {
	service billing.IBillingService `autowired:""`
}

func (i *imlBillingController) GetPricing(ctx *gin.Context) (*billing.PricingConfig, error) {
	return i.service.GetPricing(ctx), nil
}

func (i *imlBillingController) SavePricing(ctx *gin.Context, cfg *billing.PricingConfig) (*billing.PricingConfig, error) {
	if cfg == nil {
		cfg = billing.DefaultPricing()
	}
	return i.service.SavePricing(ctx, cfg)
}

func (i *imlBillingController) GetRebate(ctx *gin.Context) (*RebateInput, error) {
	return &RebateInput{Rebate: i.service.GetRebate(ctx)}, nil
}

func (i *imlBillingController) SaveRebate(ctx *gin.Context, input *RebateInput) (*RebateInput, error) {
	if input == nil {
		input = &RebateInput{Rebate: map[string]float64{}}
	}
	rebate, err := i.service.SaveRebate(ctx, input.Rebate)
	if err != nil {
		return nil, err
	}
	return &RebateInput{Rebate: rebate}, nil
}

// Overview parses the [start, end] window (unix seconds). When omitted, it
// defaults to the trailing 30 days.
func (i *imlBillingController) Overview(ctx *gin.Context, start string, end string) (*billing.BillingOverview, error) {
	now := time.Now().Unix()
	s, e := now-(30*24*3600), now
	if start != "" {
		v, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return nil, err
		}
		s = v
	}
	if end != "" {
		v, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			return nil, err
		}
		e = v
	}
	if e < s {
		e = s
	}
	return i.service.Overview(ctx, s, e)
}
