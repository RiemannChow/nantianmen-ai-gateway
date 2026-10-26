package multitenant

import (
	"reflect"
	"strconv"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/multitenant"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type TenantsInput struct {
	Items map[string]multitenant.TenantItem `json:"items"`
}

type IMultiTenantController interface {
	GetTenants(ctx *gin.Context) (*TenantsInput, error)
	SaveTenants(ctx *gin.Context, input *TenantsInput) (*TenantsInput, error)
	TenantUsage(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantUsage, error)
	TenantBilling(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantBilling, error)
	TenantStatus(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantStatus, error)
	TenantsBilling(ctx *gin.Context, start string, end string) ([]*multitenant.TenantBilling, error)
}

func init() {
	autowire.Auto[IMultiTenantController](func() reflect.Value {
		return reflect.ValueOf(new(imlMultiTenantController))
	})
}

type imlMultiTenantController struct {
	service multitenant.IMultiTenantService `autowired:""`
}

func gate(ctx *gin.Context) error {
	return license.Check(ctx, "multi_tenant")
}

func (i *imlMultiTenantController) GetTenants(ctx *gin.Context) (*TenantsInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	cfg, err := i.service.GetTenants(ctx)
	if err != nil {
		return nil, err
	}
	return &TenantsInput{Items: cfg.Items}, nil
}

func (i *imlMultiTenantController) SaveTenants(ctx *gin.Context, input *TenantsInput) (*TenantsInput, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if input == nil {
		input = &TenantsInput{Items: map[string]multitenant.TenantItem{}}
	}
	cfg, err := i.service.SaveTenants(ctx, &multitenant.TenantConfig{Items: input.Items})
	if err != nil {
		return nil, err
	}
	return &TenantsInput{Items: cfg.Items}, nil
}

func (i *imlMultiTenantController) TenantUsage(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantUsage, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	return i.service.TenantUsage(ctx, id, s, e)
}

func (i *imlMultiTenantController) TenantBilling(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantBilling, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	return i.service.TenantBilling(ctx, id, s, e)
}

func (i *imlMultiTenantController) TenantStatus(ctx *gin.Context, id string, start string, end string) (*multitenant.TenantStatus, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	return i.service.TenantStatus(ctx, id, s, e)
}

func (i *imlMultiTenantController) TenantsBilling(ctx *gin.Context, start string, end string) ([]*multitenant.TenantBilling, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	return i.service.TenantsBilling(ctx, s, e)
}

func window(start string, end string) (int64, int64) {
	now := time.Now().Unix()
	s, e := now-(30*24*3600), now
	if v, err := strconv.ParseInt(start, 10, 64); err == nil {
		s = v
	}
	if v, err := strconv.ParseInt(end, 10, 64); err == nil {
		e = v
	}
	if e < s {
		e = s
	}
	return s, e
}
