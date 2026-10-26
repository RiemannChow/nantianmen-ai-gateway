package aisecurity

import (
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/aisecurity"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type IAIsecurityController interface {
	GetRules(ctx *gin.Context) (*aisecurity.AISecurityRules, error)
	SaveRules(ctx *gin.Context, cfg *aisecurity.AISecurityRules) (*aisecurity.AISecurityRules, error)
	Scan(ctx *gin.Context, input *aisecurity.ScanInput) (*aisecurity.ScanResult, error)
}

func init() {
	autowire.Auto[IAIsecurityController](func() reflect.Value {
		return reflect.ValueOf(new(imlAIsecurityController))
	})
}

var _ IAIsecurityController = (*imlAIsecurityController)(nil)

type imlAIsecurityController struct {
	service aisecurity.IAIsecurityService `autowired:""`
}

func gate(ctx *gin.Context) error {
	return license.Check(ctx, "ai_security")
}

func (i *imlAIsecurityController) GetRules(ctx *gin.Context) (*aisecurity.AISecurityRules, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.GetRules(ctx), nil
}

func (i *imlAIsecurityController) SaveRules(ctx *gin.Context, cfg *aisecurity.AISecurityRules) (*aisecurity.AISecurityRules, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &aisecurity.AISecurityRules{}
	}
	return i.service.SaveRules(ctx, cfg)
}

func (i *imlAIsecurityController) Scan(ctx *gin.Context, input *aisecurity.ScanInput) (*aisecurity.ScanResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	return i.service.Scan(ctx, input)
}
