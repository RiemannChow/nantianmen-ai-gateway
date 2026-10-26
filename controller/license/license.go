package license

import (
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

// ActivateInput is the request body for license activation.
type ActivateInput struct {
	Payload string `json:"payload"`
}

type ILicenseController interface {
	// Get returns the current license overview.
	Get(ctx *gin.Context) (*license.LicenseView, error)
	// Activate validates and stores a signed license payload.
	Activate(ctx *gin.Context, input *ActivateInput) (*license.LicenseView, error)
	// Modules returns entitlement status for all enterprise modules.
	Modules(ctx *gin.Context) (*license.ModulesView, error)
}

func init() {
	autowire.Auto[ILicenseController](func() reflect.Value {
		return reflect.ValueOf(new(imlLicenseController))
	})
}

var _ ILicenseController = (*imlLicenseController)(nil)

type imlLicenseController struct {
	licenseService license.ILicenseService `autowired:""`
}

func (i *imlLicenseController) Get(ctx *gin.Context) (*license.LicenseView, error) {
	return i.licenseService.Get(ctx), nil
}

func (i *imlLicenseController) Activate(ctx *gin.Context, input *ActivateInput) (*license.LicenseView, error) {
	if input == nil || input.Payload == "" {
		return nil, license.NewLicenseError("License 内容为空")
	}
	return i.licenseService.Activate(ctx, input.Payload)
}

func (i *imlLicenseController) Modules(ctx *gin.Context) (*license.ModulesView, error) {
	return i.licenseService.Modules(ctx), nil
}
