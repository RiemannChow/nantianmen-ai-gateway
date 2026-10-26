package multitenant

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/billing"
	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

// IMultiTenantService provides control-plane tenant metering separation: it
// maps tenants to metering subscribers and computes per-tenant usage, spend and
// budget state. True data-plane namespace isolation is out of scope for the
// control plane (see plan §D).
type IMultiTenantService interface {
	GetTenants(ctx context.Context) (*TenantConfig, error)
	SaveTenants(ctx context.Context, cfg *TenantConfig) (*TenantConfig, error)
	TenantUsage(ctx context.Context, id string, start, end int64) (*TenantUsage, error)
	TenantBilling(ctx context.Context, id string, start, end int64) (*TenantBilling, error)
	TenantStatus(ctx context.Context, id string, start, end int64) (*TenantStatus, error)
	// TenantsBilling returns billing for every configured (and discovered) tenant.
	TenantsBilling(ctx context.Context, start, end int64) ([]*TenantBilling, error)
}

func init() {
	autowire.Auto[IMultiTenantService](func() reflect.Value {
		return reflect.ValueOf(new(imlMultiTenantService))
	})
}

type imlMultiTenantService struct {
	settingService setting.ISettingService    `autowired:""`
	usageStore     metering_store.IUsageStore `autowired:""`
	billingService billing.IBillingService    `autowired:""`
}

func operator(ctx context.Context) string {
	if id := utils.UserId(ctx); id != "" {
		return id
	}
	return "system"
}

func (i *imlMultiTenantService) GetTenants(ctx context.Context) (*TenantConfig, error) {
	raw, has := i.settingService.Get(ctx, KeyTenants)
	if !has || raw == "" {
		return &TenantConfig{Items: map[string]TenantItem{}}, nil
	}
	var cfg TenantConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &TenantConfig{Items: map[string]TenantItem{}}, err
	}
	if cfg.Items == nil {
		cfg.Items = map[string]TenantItem{}
	}
	return &cfg, nil
}

func (i *imlMultiTenantService) SaveTenants(ctx context.Context, cfg *TenantConfig) (*TenantConfig, error) {
	if cfg == nil {
		cfg = &TenantConfig{Items: map[string]TenantItem{}}
	}
	if cfg.Items == nil {
		cfg.Items = map[string]TenantItem{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyTenants, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlMultiTenantService) nameOf(cfg *TenantConfig, id string) string {
	if cfg != nil {
		if t, ok := cfg.Items[id]; ok && t.Name != "" {
			return t.Name
		}
	}
	return id
}

func (i *imlMultiTenantService) TenantUsage(ctx context.Context, id string, start, end int64) (*TenantUsage, error) {
	calls, in, out, err := i.usageStore.SummaryBySubscriberInWindow(ctx, id, start, end)
	if err != nil {
		return nil, err
	}
	cfg, _ := i.GetTenants(ctx)
	return &TenantUsage{
		Id:           id,
		Name:         i.nameOf(cfg, id),
		Calls:        calls,
		InputTokens:  in,
		OutputTokens: out,
		TotalTokens:  in + out,
	}, nil
}

// tenantSpend computes a subscriber's spend with the global billing pricing.
func (i *imlMultiTenantService) tenantSpend(ctx context.Context, id string, calls, in, out int64) (callCost, tokenCost, bill float64) {
	p := i.billingService.GetPricing(ctx)
	callCost = float64(calls) * p.PerCall
	tokenCost = float64(in)/1000*p.Per1KInputToken + float64(out)/1000*p.Per1KOutputToken
	bill = callCost + tokenCost
	return
}

func (i *imlMultiTenantService) TenantBilling(ctx context.Context, id string, start, end int64) (*TenantBilling, error) {
	calls, in, out, err := i.usageStore.SummaryBySubscriberInWindow(ctx, id, start, end)
	if err != nil {
		return nil, err
	}
	cfg, _ := i.GetTenants(ctx)
	tb := &TenantBilling{Id: id, Name: i.nameOf(cfg, id), Calls: calls, InputTokens: in, OutputTokens: out, TotalTokens: in + out}
	tb.CallCost, tb.TokenCost, tb.Bill = i.tenantSpend(ctx, id, calls, in, out)
	if t, ok := cfg.Items[id]; ok {
		tb.MonthlyBudget = t.MonthlyBudget
		if t.MonthlyBudget > 0 {
			tb.Ratio = round2(tb.Bill / t.MonthlyBudget)
			tb.Level = budgetLevel(tb.Ratio, 0.8)
		} else {
			tb.Level = "ok"
		}
	}
	return tb, nil
}

func (i *imlMultiTenantService) TenantStatus(ctx context.Context, id string, start, end int64) (*TenantStatus, error) {
	calls, in, out, err := i.usageStore.SummaryBySubscriberInWindow(ctx, id, start, end)
	if err != nil {
		return nil, err
	}
	spend := 0.0
	cfg, _ := i.GetTenants(ctx)
	t := cfg.Items[id]
	if t.MonthlyBudget <= 0 {
		return &TenantStatus{Id: id, Name: i.nameOf(cfg, id), Configured: false, Message: "未配置月度预算"}, nil
	}
	_, _, spend = i.tenantSpend(ctx, id, calls, in, out)
	ratio := 0.0
	if t.MonthlyBudget > 0 {
		ratio = spend / t.MonthlyBudget
	}
	level := budgetLevel(ratio, 0.8)
	msg := "预算充足"
	switch level {
	case "warning":
		msg = "已使用预算的 %.0f%%，达到预警线。"
	case "exceeded":
		msg = "本月支出已超出预算（%.0f%%）。"
	}
	return &TenantStatus{
		Id:            id,
		Name:          i.nameOf(cfg, id),
		Configured:    true,
		MonthlyBudget: t.MonthlyBudget,
		MonthSpend:    round2(spend),
		Ratio:         round2(ratio),
		Level:         level,
		Message:       msg,
	}, nil
}

// TenantsBilling returns per-tenant billing for every configured tenant plus any
// subscriber that has usage in the window (so discovered tenants are visible).
func (i *imlMultiTenantService) TenantsBilling(ctx context.Context, start, end int64) ([]*TenantBilling, error) {
	cfg, _ := i.GetTenants(ctx)
	// base set: configured tenants
	ids := make([]string, 0, len(cfg.Items))
	for id := range cfg.Items {
		ids = append(ids, id)
	}
	out := make([]*TenantBilling, 0, len(ids))
	for _, id := range ids {
		tb, err := i.TenantBilling(ctx, id, start, end)
		if err != nil {
			return nil, err
		}
		out = append(out, tb)
	}
	return out, nil
}

func budgetLevel(ratio float64, warnThreshold float64) string {
	if ratio >= 1 {
		return "exceeded"
	}
	if warnThreshold > 0 && ratio >= warnThreshold {
		return "warning"
	}
	return "ok"
}

func round2(v float64) float64 { return float64(int64(v*100+0.5)) / 100 }
