package billing

import (
	"context"
	"encoding/json"
	"reflect"

	cluster_service "github.com/RiemannChow/nantianmen-ai-gateway/service/cluster"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	monitor_module "github.com/RiemannChow/nantianmen-ai-gateway/module/monitor"
	monitor_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/monitor/dto"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

const providerTopN = 200

// IBillingService computes bills and channel rebates from usage data and the
// operator-configured pricing/rebate configuration.
type IBillingService interface {
	// GetPricing returns the current pricing configuration (zero-valued if unset).
	GetPricing(ctx context.Context) *PricingConfig
	// SavePricing persists the pricing configuration.
	SavePricing(ctx context.Context, cfg *PricingConfig) (*PricingConfig, error)
	// GetRebate returns the current channel-rebate configuration (empty if unset).
	GetRebate(ctx context.Context) RebateConfig
	// SaveRebate persists the channel-rebate configuration.
	SaveRebate(ctx context.Context, cfg RebateConfig) (RebateConfig, error)
	// Overview computes the bill + rebate for the [start, end] window.
	Overview(ctx context.Context, start int64, end int64) (*BillingOverview, error)
}

func init() {
	autowire.Auto[IBillingService](func() reflect.Value {
		return reflect.ValueOf(new(imlBillingService))
	})
}

type imlBillingService struct {
	settingService setting.ISettingService                `autowired:""`
	monitorModule  monitor_module.IMonitorStatisticModule `autowired:""`
	clusterService cluster_service.IClusterService        `autowired:""`
}

func operator(ctx context.Context) string {
	if id := utils.UserId(ctx); id != "" {
		return id
	}
	return "system"
}

func (i *imlBillingService) GetPricing(ctx context.Context) *PricingConfig {
	raw, has := i.settingService.Get(ctx, KeyPricing)
	if !has || raw == "" {
		return DefaultPricing()
	}
	var cfg PricingConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return DefaultPricing()
	}
	return &cfg
}

func (i *imlBillingService) SavePricing(ctx context.Context, cfg *PricingConfig) (*PricingConfig, error) {
	if cfg == nil {
		cfg = DefaultPricing()
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyPricing, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlBillingService) GetRebate(ctx context.Context) RebateConfig {
	raw, has := i.settingService.Get(ctx, KeyRebate)
	if !has || raw == "" {
		return RebateConfig{}
	}
	var cfg RebateConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return RebateConfig{}
	}
	return cfg
}

func (i *imlBillingService) SaveRebate(ctx context.Context, cfg RebateConfig) (RebateConfig, error) {
	if cfg == nil {
		cfg = RebateConfig{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyRebate, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Overview aggregates usage (calls / tokens / instances) for the window and
// applies the configured pricing + rebate rules. Monitor/cluster failures are
// tolerated: the affected metric simply contributes zero rather than failing
// the whole report (useful when InfluxDB or data-plane is not yet connected).
func (i *imlBillingService) Overview(ctx context.Context, start int64, end int64) (*BillingOverview, error) {
	pricing := i.GetPricing(ctx)
	rebate := i.GetRebate(ctx)

	out := &BillingOverview{
		Start:   start,
		End:     end,
		Pricing: pricing,
	}

	common := &monitor_dto.CommonInput{Start: start, End: end}

	// 1) total calls (request summary)
	summary, err := i.monitorModule.RequestSummary(ctx, common)
	if err == nil && summary != nil {
		out.Calls = summary.Total
	}

	// 2) aggregate tokens (AI chart overview)
	ai, err := i.monitorModule.AIChartOverview(ctx, "", start, end)
	if err == nil && ai != nil {
		out.InputTokens = ai.InputTokenTotal
		out.OutputTokens = ai.OutputTokenTotal
	}

	// 3) per-provider call counts (drives rebate + token distribution)
	providers, perr := i.monitorModule.TopProviderStatistics(ctx, providerTopN, common)
	if perr == nil && len(providers) > 0 {
		var sumCalls int64
		for _, p := range providers {
			if p.MonCommonData != nil {
				sumCalls += p.RequestTotal
			}
		}
		for _, p := range providers {
			var calls int64
			if p.MonCommonData != nil {
				calls = p.RequestTotal
			}
			pb := &ProviderBill{Provider: p.Name, Calls: calls}

			// distribute aggregate tokens to providers by call share
			if sumCalls > 0 {
				share := float64(calls) / float64(sumCalls)
				pb.InputTokens = int64(float64(out.InputTokens) * share)
				pb.OutputTokens = int64(float64(out.OutputTokens) * share)
			}
			pb.CallCost = round2(float64(calls) * pricing.PerCall)
			pb.TokenCost = round2(
				float64(pb.InputTokens)/1000*pricing.Per1KInputToken +
					float64(pb.OutputTokens)/1000*pricing.Per1KOutputToken)
			pb.Bill = round2(pb.CallCost + pb.TokenCost)
			pb.Rebate = round2(float64(calls) * rebate[p.Name])
			out.Providers = append(out.Providers, pb)
		}
	}

	// 4) instance count (enabled data-plane clusters)
	clusters, cerr := i.clusterService.List(ctx)
	if cerr == nil {
		for _, c := range clusters {
			if c.Enabled {
				out.Instances++
			}
		}
	}

	// 5) aggregate costs
	out.CallCost = round2(float64(out.Calls) * pricing.PerCall)
	out.TokenCost = round2(
		float64(out.InputTokens)/1000*pricing.Per1KInputToken +
			float64(out.OutputTokens)/1000*pricing.Per1KOutputToken)
	out.InstanceCost = round2(float64(out.Instances) * pricing.PerInstanceMonth)
	out.BillTotal = round2(out.CallCost + out.TokenCost + out.InstanceCost)

	var rebateTotal float64
	for _, pb := range out.Providers {
		rebateTotal += pb.Rebate
	}
	out.RebateTotal = round2(rebateTotal)
	out.NetTotal = round2(out.BillTotal - out.RebateTotal)

	return out, nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
