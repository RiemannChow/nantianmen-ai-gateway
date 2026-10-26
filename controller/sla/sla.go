package sla

import (
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

// ISLAController exposes the enterprise SLA entitlements. It is gated by the
// sla license module so the endpoint returns a license-required error unless the
// active license grants SLA support.
type ISLAController interface {
	Info(ctx *gin.Context) (*license.SLADetail, error)
}

func init() {
	autowire.Auto[ISLAController](func() reflect.Value {
		return reflect.ValueOf(new(imlSLAController))
	})
}

var _ ISLAController = (*imlSLAController)(nil)

type imlSLAController struct {
	licenseService license.ILicenseService `autowired:""`
}

func (i *imlSLAController) Info(ctx *gin.Context) (*license.SLADetail, error) {
	if err := license.Check(ctx, "sla"); err != nil {
		return nil, err
	}
	view := i.licenseService.Get(ctx)
	return license.ResolveSLA(view.SLA), nil
}
