package metering

import (
	"context"
	"math/rand"
	"reflect"
	"time"

	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"
	"github.com/eolinker/go-common/autowire"
)

// IMeteringService owns the local usage collection pipeline. It is the write path
// for Apinto (or any collector) pushed access-log events, plus a demo seeder so a
// fresh install has representative data for the FinOps suite out of the box.
type IMeteringService interface {
	// Ingest persists a batch of normalized usage events.
	Ingest(ctx context.Context, events []*metering_store.UsageEvent) error
	// Seed generates realistic sample usage spanning `days` (default 30), after
	// clearing any previously seeded data. It returns the number of events written.
	Seed(ctx context.Context, days int) (int, error)
	// HasData reports whether any usage events exist in [start, end].
	HasData(ctx context.Context, start, end int64) (bool, error)
}

func init() {
	autowire.Auto[IMeteringService](func() reflect.Value {
		return reflect.ValueOf(new(imlMeteringService))
	})
}

type imlMeteringService struct {
	usageStore metering_store.IUsageStore `autowired:""`
}

func (i *imlMeteringService) Ingest(ctx context.Context, events []*metering_store.UsageEvent) error {
	if len(events) == 0 {
		return nil
	}
	return i.usageStore.Insert(ctx, events...)
}

func (i *imlMeteringService) HasData(ctx context.Context, start, end int64) (bool, error) {
	n, err := i.usageStore.CountInWindow(ctx, start, end)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Seed generates deterministic-enough, realistic usage so the billing / analytics
// / cost-optimization suites display meaningful numbers on a fresh install that
// has no real Apinto traffic yet.
func (i *imlMeteringService) Seed(ctx context.Context, days int) (int, error) {
	if days <= 0 {
		days = 30
	}
	if err := i.usageStore.DeleteAll(ctx); err != nil {
		return 0, err
	}

	providers := []string{"openai", "azure", "通义千问", "本地模型"}
	subs := []string{"team-alpha", "team-beta", "internal-app"}
	now := time.Now()
	r := rand.New(rand.NewSource(now.UnixNano()))

	events := make([]*metering_store.UsageEvent, 0, days*6)
	for d := days; d >= 0; d-- {
		day := now.AddDate(0, 0, -d)
		for h := 0; h < 6; h++ {
			ts := day.Add(time.Duration(r.Intn(24)) * time.Hour)
			p := providers[r.Intn(len(providers))]
			calls := int64(r.Intn(200) + 10)
			inTok := calls * int64(r.Intn(1500)+200)
			outTok := calls * int64(r.Intn(800)+100)
			code := 200
			switch roll := r.Intn(100); {
			case roll > 95:
				code = 500
			case roll > 88:
				code = 429
			case roll > 80:
				code = 404
			}
			events = append(events, &metering_store.UsageEvent{
				ClusterId:    "default",
				ServiceId:    "sample",
				Provider:     p,
				Subscriber:   subs[r.Intn(len(subs))],
				Calls:        calls,
				InputTokens:  inTok,
				OutputTokens: outTok,
				StatusCode:   code,
				LatencyMs:    int64(r.Intn(800) + 20),
				CreatedAt:    ts,
			})
		}
	}
	if err := i.usageStore.Insert(ctx, events...); err != nil {
		return 0, err
	}
	return len(events), nil
}
