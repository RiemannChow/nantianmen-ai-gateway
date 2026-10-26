package metering

import (
	"reflect"
	"strconv"

	metering_module "github.com/RiemannChow/nantianmen-ai-gateway/module/metering"
	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

// IngestInput is the request body for batch usage ingestion.
type IngestInput struct {
	Events []*metering_store.UsageEvent `json:"events"`
}

type IMeteringController interface {
	Ingest(ctx *gin.Context, input *IngestInput) (*MeteringResult, error)
	Seed(ctx *gin.Context, days string) (*MeteringResult, error)
}

func init() {
	autowire.Auto[IMeteringController](func() reflect.Value {
		return reflect.ValueOf(new(imlMeteringController))
	})
}

// MeteringResult is a small generic acknowledgement payload.
type MeteringResult struct {
	Count  int    `json:"count"`
	Days   int    `json:"days,omitempty"`
	Status string `json:"status,omitempty"`
}

type imlMeteringController struct {
	service metering_module.IMeteringService `autowired:""`
}

func (i *imlMeteringController) Ingest(ctx *gin.Context, input *IngestInput) (*MeteringResult, error) {
	if input == nil {
		input = &IngestInput{}
	}
	if err := i.service.Ingest(ctx, input.Events); err != nil {
		return nil, err
	}
	return &MeteringResult{Count: len(input.Events)}, nil
}

func (i *imlMeteringController) Seed(ctx *gin.Context, days string) (*MeteringResult, error) {
	d := 30
	if v, err := strconv.Atoi(days); err == nil && v > 0 {
		d = v
	}
	n, err := i.service.Seed(ctx, d)
	if err != nil {
		return nil, err
	}
	return &MeteringResult{Count: n, Status: "seeded", Days: d}, nil
}
