package tracing

import (
	"reflect"
	"strconv"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/RiemannChow/nantianmen-ai-gateway/module/tracing"
	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"
	"github.com/eolinker/go-common/autowire"
	"github.com/gin-gonic/gin"
)

type TraceIngestInput struct {
	Events []*metering_store.TraceEvent `json:"events"`
}

type ITracingController interface {
	Ingest(ctx *gin.Context, input *TraceIngestInput) (*TracingResult, error)
	Seed(ctx *gin.Context, days string) (*TracingResult, error)
	Overview(ctx *gin.Context, start string, end string) (*tracing.TraceOverview, error)
	List(ctx *gin.Context, start string, end string, page string, pageSize string) (*tracing.TraceList, error)
}

func init() {
	autowire.Auto[ITracingController](func() reflect.Value {
		return reflect.ValueOf(new(imlTracingController))
	})
}

type TracingResult struct {
	Count  int    `json:"count"`
	Days   int    `json:"days,omitempty"`
	Status string `json:"status,omitempty"`
}

type imlTracingController struct {
	service tracing.ITracingService `autowired:""`
}

func gate(ctx *gin.Context) error {
	return license.Check(ctx, "tracing")
}

func (i *imlTracingController) Ingest(ctx *gin.Context, input *TraceIngestInput) (*TracingResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	if input == nil {
		input = &TraceIngestInput{}
	}
	if err := i.service.Ingest(ctx, input.Events); err != nil {
		return nil, err
	}
	return &TracingResult{Count: len(input.Events)}, nil
}

func (i *imlTracingController) Seed(ctx *gin.Context, days string) (*TracingResult, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	d := 30
	if v, err := strconv.Atoi(days); err == nil && v > 0 {
		d = v
	}
	n, err := i.service.Seed(ctx, d)
	if err != nil {
		return nil, err
	}
	return &TracingResult{Count: n, Status: "seeded", Days: d}, nil
}

func (i *imlTracingController) Overview(ctx *gin.Context, start string, end string) (*tracing.TraceOverview, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	return i.service.Overview(ctx, s, e)
}

func (i *imlTracingController) List(ctx *gin.Context, start string, end string, page string, pageSize string) (*tracing.TraceList, error) {
	if err := gate(ctx); err != nil {
		return nil, err
	}
	s, e := window(start, end)
	p, _ := strconv.Atoi(page)
	ps, _ := strconv.Atoi(pageSize)
	return i.service.List(ctx, s, e, p, ps)
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
