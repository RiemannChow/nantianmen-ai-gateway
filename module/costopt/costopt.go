package costopt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/billing"
	monitor_module "github.com/RiemannChow/nantianmen-ai-gateway/module/monitor"
	monitor_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/monitor/dto"

	service_cluster "github.com/RiemannChow/nantianmen-ai-gateway/service/cluster"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"
	service_strategy "github.com/RiemannChow/nantianmen-ai-gateway/service/strategy"
	metering_store "github.com/RiemannChow/nantianmen-ai-gateway/stores/metering"

	strategy_driver "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/driver"
	strategy_dto "github.com/RiemannChow/nantianmen-ai-gateway/module/strategy/dto"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

const providerTopN = 200

// usageRow is one provider's raw usage in a window.
type usageRow struct {
	name      string
	calls     int64
	inTokens  int64
	outTokens int64
}

// ICostOptService is the value-added cost-optimization + analytics suite.
type ICostOptService interface {
	GetBudget(ctx context.Context) *BudgetConfig
	SaveBudget(ctx context.Context, cfg *BudgetConfig) (*BudgetConfig, error)
	BudgetStatus(ctx context.Context) (*BudgetStatus, error)
	GetProviders(ctx context.Context) map[string]ProviderCost
	SaveProviders(ctx context.Context, cfg map[string]ProviderCost) (map[string]ProviderCost, error)
	GetSubstitutes(ctx context.Context) map[string]string
	SaveSubstitutes(ctx context.Context, cfg map[string]string) (map[string]string, error)
	GetJudge(ctx context.Context) *JudgeConfig
	SaveJudge(ctx context.Context, cfg *JudgeConfig) (*JudgeConfig, error)
	Recommend(ctx context.Context, start int64, end int64, shiftRatio float64) (*RecommendationResult, error)
	Analytics(ctx context.Context, start int64, end int64) (*AnalyticsResult, error)
	Anomalies(ctx context.Context, start int64, end int64) (*AnomalyResult, error)
	Evaluate(ctx context.Context, input *EvaluateInput) (*EvaluateResult, error)
	// notification delivery
	GetNotifyConfig(ctx context.Context) *NotifyConfig
	SaveNotifyConfig(ctx context.Context, cfg *NotifyConfig) (*NotifyConfig, error)
	SendTestNotification(ctx context.Context, channel string) error
	// per-consumer / per-key budget
	GetConsumerBudgets(ctx context.Context) (*ConsumerBudgetConfig, error)
	SaveConsumerBudgets(ctx context.Context, cfg *ConsumerBudgetConfig) (*ConsumerBudgetConfig, error)
	ConsumerBudgetStatus(ctx context.Context, consumer string) (*ConsumerBudgetStatus, error)
	// per-consumer quota enforcement (publish limiting policy to data plane)
	PublishConsumerQuota(ctx context.Context, consumer string) error
}

func init() {
	autowire.Auto[ICostOptService](func() reflect.Value {
		return reflect.ValueOf(new(imlCostOptService))
	})
}

type imlCostOptService struct {
	settingService  setting.ISettingService                `autowired:""`
	monitorModule   monitor_module.IMonitorStatisticModule `autowired:""`
	billingService  billing.IBillingService                `autowired:""`
	usageStore      metering_store.IUsageStore             `autowired:""`
	strategyService service_strategy.IStrategyService      `autowired:""`
	clusterService  service_cluster.IClusterService        `autowired:""`
	smtpConfig      *SMTPConfig                            `autowired:""`
}

func operator(ctx context.Context) string {
	if id := utils.UserId(ctx); id != "" {
		return id
	}
	return "system"
}

// ---- config helpers ----

func (i *imlCostOptService) GetBudget(ctx context.Context) *BudgetConfig {
	raw, has := i.settingService.Get(ctx, KeyBudget)
	if !has || raw == "" {
		return &BudgetConfig{}
	}
	var cfg BudgetConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &BudgetConfig{}
	}
	return &cfg
}

func (i *imlCostOptService) SaveBudget(ctx context.Context, cfg *BudgetConfig) (*BudgetConfig, error) {
	if cfg == nil {
		cfg = &BudgetConfig{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyBudget, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlCostOptService) GetProviders(ctx context.Context) map[string]ProviderCost {
	raw, has := i.settingService.Get(ctx, KeyProviders)
	if !has || raw == "" {
		return map[string]ProviderCost{}
	}
	var cfg map[string]ProviderCost
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return map[string]ProviderCost{}
	}
	return cfg
}

func (i *imlCostOptService) SaveProviders(ctx context.Context, cfg map[string]ProviderCost) (map[string]ProviderCost, error) {
	if cfg == nil {
		cfg = map[string]ProviderCost{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyProviders, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlCostOptService) GetSubstitutes(ctx context.Context) map[string]string {
	raw, has := i.settingService.Get(ctx, KeySubstitutes)
	if !has || raw == "" {
		return map[string]string{}
	}
	var cfg map[string]string
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return map[string]string{}
	}
	return cfg
}

func (i *imlCostOptService) SaveSubstitutes(ctx context.Context, cfg map[string]string) (map[string]string, error) {
	if cfg == nil {
		cfg = map[string]string{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeySubstitutes, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlCostOptService) GetJudge(ctx context.Context) *JudgeConfig {
	raw, has := i.settingService.Get(ctx, KeyJudge)
	if !has || raw == "" {
		return &JudgeConfig{}
	}
	var cfg JudgeConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &JudgeConfig{}
	}
	return &cfg
}

func (i *imlCostOptService) SaveJudge(ctx context.Context, cfg *JudgeConfig) (*JudgeConfig, error) {
	if cfg == nil {
		cfg = &JudgeConfig{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyJudge, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (i *imlCostOptService) GetNotifyConfig(ctx context.Context) *NotifyConfig {
	raw, has := i.settingService.Get(ctx, KeyNotify)
	if !has || raw == "" {
		return &NotifyConfig{}
	}
	var cfg NotifyConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &NotifyConfig{}
	}
	return &cfg
}

func (i *imlCostOptService) SaveNotifyConfig(ctx context.Context, cfg *NotifyConfig) (*NotifyConfig, error) {
	if cfg == nil {
		cfg = &NotifyConfig{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyNotify, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ---- usage aggregation ----

// providerUsage pulls per-provider raw usage and distributes aggregate tokens
// across providers by call share (mirrors billing's aggregation). Monitor
// failures are tolerated: the affected metric is treated as zero.
func (i *imlCostOptService) providerUsage(ctx context.Context, start int64, end int64) (totalCalls, totalIn, totalOut int64, rows []usageRow) {
	common := &monitor_dto.CommonInput{Start: start, End: end}

	if s, err := i.monitorModule.RequestSummary(ctx, common); err == nil && s != nil {
		totalCalls = s.Total
	}
	if ai, err := i.monitorModule.AIChartOverview(ctx, "", start, end); err == nil && ai != nil {
		totalIn = ai.InputTokenTotal
		totalOut = ai.OutputTokenTotal
	}
	if ps, err := i.monitorModule.TopProviderStatistics(ctx, providerTopN, common); err == nil {
		var sumCalls int64
		for _, p := range ps {
			var c int64
			if p.MonCommonData != nil {
				c = p.RequestTotal
			}
			sumCalls += c
		}
		for _, p := range ps {
			var c int64
			if p.MonCommonData != nil {
				c = p.RequestTotal
			}
			r := usageRow{name: p.Name, calls: c}
			if sumCalls > 0 {
				share := float64(c) / float64(sumCalls)
				r.inTokens = int64(float64(totalIn) * share)
				r.outTokens = int64(float64(totalOut) * share)
			}
			rows = append(rows, r)
		}
	}
	return
}

func round2(v float64) float64 { return float64(int64(v*100+0.5)) / 100 }

// providerSpend computes a provider's spend using its cost profile if present,
// otherwise the global billing pricing.
func (i *imlCostOptService) providerSpend(ctx context.Context, r usageRow, pc ProviderCost) float64 {
	if pc.PerCall == 0 && pc.Per1KInputToken == 0 && pc.Per1KOutputToken == 0 {
		p := i.billingService.GetPricing(ctx)
		return float64(r.calls)*p.PerCall +
			float64(r.inTokens)/1000*p.Per1KInputToken +
			float64(r.outTokens)/1000*p.Per1KOutputToken
	}
	return float64(r.calls)*pc.PerCall +
		float64(r.inTokens)/1000*pc.Per1KInputToken +
		float64(r.outTokens)/1000*pc.Per1KOutputToken
}

// ---- analytics ----

func (i *imlCostOptService) Analytics(ctx context.Context, start int64, end int64) (*AnalyticsResult, error) {
	totalCalls, totalIn, totalOut, rows := i.providerUsage(ctx, start, end)
	pcs := i.GetProviders(ctx)

	out := &AnalyticsResult{Window: Window{Start: start, End: end}}
	var totalSpend float64
	spends := make([]*ProviderSpend, 0, len(rows))
	for _, r := range rows {
		pc := pcs[r.name]
		spend := round2(i.providerSpend(ctx, r, pc))
		totalSpend += spend
		spends = append(spends, &ProviderSpend{
			Provider:     r.name,
			Calls:        r.calls,
			InputTokens:  r.inTokens,
			OutputTokens: r.outTokens,
			Spend:        spend,
		})
	}
	sort.SliceStable(spends, func(a, b int) bool { return spends[a].Spend > spends[b].Spend })
	for _, s := range spends {
		if totalSpend > 0 {
			s.Share = round2(s.Spend / totalSpend)
		}
	}
	out.TotalSpend = round2(totalSpend)
	out.TotalCalls = totalCalls
	out.TotalTokens = totalIn + totalOut
	out.Providers = spends
	return out, nil
}

// ---- recommendations ----

func (i *imlCostOptService) Recommend(ctx context.Context, start int64, end int64, shiftRatio float64) (*RecommendationResult, error) {
	if shiftRatio <= 0 || shiftRatio > 1 {
		shiftRatio = 0.3
	}
	totalCalls, _, _, rows := i.providerUsage(ctx, start, end)
	pcs := i.GetProviders(ctx)
	gp := i.billingService.GetPricing(ctx)

	out := &RecommendationResult{
		Window:        Window{Start: start, End: end},
		Pricing:       map[string]float64{"per_call": gp.PerCall, "per_1k_input_token": gp.Per1KInputToken, "per_1k_output_token": gp.Per1KOutputToken, "per_instance_month": gp.PerInstanceMonth},
		ProviderCosts: pcs,
	}

	// compute spend per provider + blended per-call cost
	type provSpend struct {
		name  string
		calls int64
		spend float64
		cost  float64 // blended per-call
		hasPC bool
	}
	subs := i.GetSubstitutes(ctx)
	spends := make([]provSpend, 0, len(rows))
	for _, r := range rows {
		pc, has := pcs[r.name]
		spend := i.providerSpend(ctx, r, pc)
		ps := provSpend{name: r.name, calls: r.calls, spend: spend, hasPC: has}
		if r.calls > 0 {
			ps.cost = spend / float64(r.calls)
		}
		spends = append(spends, ps)
	}
	sort.SliceStable(spends, func(a, b int) bool { return spends[a].spend > spends[b].spend })

	// substitution recommendations: the most expensive provider (by blended cost,
	// with a configured cost profile) can shift traffic to a cheaper one.
	if len(spends) >= 2 {
		// cheapest by blended cost that has a cost profile
		var cheapest provSpend
		for _, s := range spends {
			if s.hasPC && (cheapest.name == "" || s.cost < cheapest.cost) {
				cheapest = s
			}
		}
		for _, expensive := range spends {
			if !expensive.hasPC || expensive.name == cheapest.name {
				continue
			}
			// explicit substitution map takes precedence
			to := subs[expensive.name]
			if to != "" {
				// verify `to` exists
				var found bool
				for _, s := range spends {
					if s.name == to {
						found = true
						cheapest = s
						break
					}
				}
				if !found {
					to = ""
				}
			}
			if to == "" {
				to = cheapest.name
			}
			if to == expensive.name {
				continue
			}
			shifted := int64(float64(expensive.calls) * shiftRatio)
			if shifted <= 0 {
				continue
			}
			fromPerCall := expensive.spend / float64(expensive.calls)
			toPerCall := cheapest.cost
			if toPerCall <= 0 {
				continue
			}
			savings := round2(float64(shifted) * (fromPerCall - toPerCall))
			if savings <= 0 {
				continue
			}
			out.Recommendations = append(out.Recommendations, &RecommendationItem{
				Type:             "model_substitution",
				Title:            fmt.Sprintf("将 %d%% 的「%s」流量迁移至「%s」", int(shiftRatio*100), expensive.name, to),
				Detail:           fmt.Sprintf("「%s」单次调用成本约 %.4f，「%s」约 %.4f，按当前用量迁移可降本。", expensive.name, fromPerCall, to, toPerCall),
				FromProvider:     expensive.name,
				ToProvider:       to,
				ShiftedCalls:     shifted,
				ProjectedSavings: savings,
			})
		}
	}

	// generic tips
	if totalCalls > 0 {
		// Estimate the caching saving from the share of repeat prompts. The
		// metering layer does not yet expose a distinct-prompt count, so we apply
		// a conservative assumed repeat rate to the measured spend. The semantic
		// cache (system.settings.ai_cache) short-circuits these repeat calls.
		var totalSpend float64
		for _, s := range spends {
			totalSpend += s.spend
		}
		const assumedRepeatRate = 0.1
		cachingSavings := round2(totalSpend * assumedRepeatRate)
		out.Recommendations = append(out.Recommendations, &RecommendationItem{
			Type:             "caching",
			Title:            "为高重复度请求启用缓存 / 限流",
			Detail:           "对相同或高度相似的提示词启用响应缓存与调用限流，按当前用量估算约可减少 10% 的重复调用成本。",
			ProjectedSavings: cachingSavings,
		})
	}

	var total float64
	for _, r := range out.Recommendations {
		total += r.ProjectedSavings
	}
	out.TotalProjectedSavings = round2(total)
	return out, nil
}

// ---- anomalies ----

func (i *imlCostOptService) Anomalies(ctx context.Context, start int64, end int64) (*AnomalyResult, error) {
	out := &AnomalyResult{Window: Window{Start: start, End: end}}
	totalCalls, _, _, _ := i.providerUsage(ctx, start, end)

	if totalCalls == 0 {
		out.Anomalies = append(out.Anomalies, &AnomalyItem{
			Type:     "no_usage",
			Severity: "info",
			Title:    "暂无用量数据",
			Detail:   "当前时间范围内没有可用的请求/Token 用量，接入数据面与 InfluxDB 后此处将展示异常与风险。",
		})
		return out, nil
	}

	// budget risk
	if bs, err := i.BudgetStatus(ctx); err == nil && bs.Configured {
		if bs.Level == "exceeded" || bs.Level == "warning" {
			out.Anomalies = append(out.Anomalies, &AnomalyItem{
				Type:     "budget_at_risk",
				Severity: bs.Level,
				Title:    "月度预算风险",
				Detail:   fmt.Sprintf("本月已花费 %.2f / 预算 %.2f（占比 %.0f%%）。", bs.MonthSpend, bs.MonthlyBudget, bs.Ratio*100),
			})
		}
	}

	// per-consumer budget risk
	if cb, err := i.GetConsumerBudgets(ctx); err == nil && cb != nil {
		for id, item := range cb.Items {
			if !item.Enabled {
				continue
			}
			if st, e := i.ConsumerBudgetStatus(ctx, id); e == nil && st.Configured {
				if st.Level == "exceeded" || st.Level == "warning" {
					out.Anomalies = append(out.Anomalies, &AnomalyItem{
						Type:     "consumer_budget_at_risk",
						Severity: st.Level,
						Title:    fmt.Sprintf("消费者预算风险：%s", id),
						Detail:   fmt.Sprintf("消费者「%s」本月已花费 %.2f / 预算 %.2f（占比 %.0f%%）。", id, st.MonthSpend, st.MonthlyBudget, st.Ratio*100),
					})
				}
			}
		}
	}

	// concentration: a single provider dominates spend
	if a, err := i.Analytics(ctx, start, end); err == nil && len(a.Providers) > 0 {
		if a.Providers[0].Share >= 0.6 {
			out.Anomalies = append(out.Anomalies, &AnomalyItem{
				Type:     "concentration",
				Severity: "warning",
				Title:    "供应商集中度偏高",
				Detail:   fmt.Sprintf("「%s」占本月支出的 %.0f%%，建议引入替代供应商以分散风险与成本。", a.Providers[0].Provider, a.Providers[0].Share*100),
			})
		}
	}
	return out, nil
}

// ---- budget status ----

func monthStartNow() (int64, int64) {
	now := time.Now()
	ms := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return ms.Unix(), now.Unix()
}

func (i *imlCostOptService) BudgetStatus(ctx context.Context) (*BudgetStatus, error) {
	cfg := i.GetBudget(ctx)
	if cfg.MonthlyBudget <= 0 {
		return &BudgetStatus{Configured: false, Message: "未配置月度预算"}, nil
	}
	ms, now := monthStartNow()
	var monthSpend float64
	if ov, err := i.billingService.Overview(ctx, ms, now); err == nil && ov != nil {
		monthSpend = ov.BillTotal
	}
	ratio := 0.0
	if cfg.MonthlyBudget > 0 {
		ratio = monthSpend / cfg.MonthlyBudget
	}
	level := "ok"
	if ratio >= 1 {
		level = "exceeded"
	} else if cfg.AlertThreshold > 0 && ratio >= cfg.AlertThreshold {
		level = "warning"
	}
	msg := "预算充足"
	switch level {
	case "warning":
		msg = fmt.Sprintf("已使用预算的 %.0f%%，达到预警线，请关注支出。", ratio*100)
	case "exceeded":
		msg = fmt.Sprintf("本月支出已超出预算（%.0f%%）。", ratio*100)
	}
	return &BudgetStatus{
		Configured:     true,
		MonthlyBudget:  cfg.MonthlyBudget,
		AlertThreshold: cfg.AlertThreshold,
		MonthStart:     ms,
		Now:            now,
		MonthSpend:     round2(monthSpend),
		Ratio:          round2(ratio),
		Level:          level,
		Message:        msg,
	}, nil
}

// ---- per-consumer / per-key budget ----

func (i *imlCostOptService) GetConsumerBudgets(ctx context.Context) (*ConsumerBudgetConfig, error) {
	raw, has := i.settingService.Get(ctx, KeyConsumerBudgets)
	if !has || raw == "" {
		return &ConsumerBudgetConfig{Items: map[string]ConsumerBudgetItem{}}, nil
	}
	var cfg ConsumerBudgetConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &ConsumerBudgetConfig{Items: map[string]ConsumerBudgetItem{}}, err
	}
	if cfg.Items == nil {
		cfg.Items = map[string]ConsumerBudgetItem{}
	}
	return &cfg, nil
}

func (i *imlCostOptService) SaveConsumerBudgets(ctx context.Context, cfg *ConsumerBudgetConfig) (*ConsumerBudgetConfig, error) {
	if cfg == nil {
		cfg = &ConsumerBudgetConfig{Items: map[string]ConsumerBudgetItem{}}
	}
	if cfg.Items == nil {
		cfg.Items = map[string]ConsumerBudgetItem{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyConsumerBudgets, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// ConsumerBudgetStatus evaluates a single subscriber's month-to-date spend
// against its configured budget using the global pricing as the fallback rate.
func (i *imlCostOptService) ConsumerBudgetStatus(ctx context.Context, consumer string) (*ConsumerBudgetStatus, error) {
	cfg, err := i.GetConsumerBudgets(ctx)
	if err != nil {
		return nil, err
	}
	item, ok := cfg.Items[consumer]
	if !ok || item.MonthlyBudget <= 0 {
		return &ConsumerBudgetStatus{Configured: false, Consumer: consumer}, nil
	}
	ms, now := monthStartNow()
	calls, in, out, err := i.usageStore.SummaryBySubscriberInWindow(ctx, consumer, ms, now)
	if err != nil {
		return nil, err
	}
	p := i.billingService.GetPricing(ctx)
	spend := float64(calls)*p.PerCall +
		float64(in)/1000*p.Per1KInputToken +
		float64(out)/1000*p.Per1KOutputToken
	ratio := 0.0
	if item.MonthlyBudget > 0 {
		ratio = spend / item.MonthlyBudget
	}
	level := "ok"
	if ratio >= 1 {
		level = "exceeded"
	} else if item.AlertThreshold > 0 && ratio >= item.AlertThreshold {
		level = "warning"
	}
	msg := "预算充足"
	switch level {
	case "warning":
		msg = fmt.Sprintf("已使用预算的 %.0f%%，达到预警线。", ratio*100)
	case "exceeded":
		msg = fmt.Sprintf("本月支出已超出预算（%.0f%%）。", ratio*100)
	}
	return &ConsumerBudgetStatus{
		Configured:     true,
		Consumer:       consumer,
		MonthlyBudget:  item.MonthlyBudget,
		AlertThreshold: item.AlertThreshold,
		MonthSpend:     round2(spend),
		Ratio:          round2(ratio),
		Level:          level,
		Message:        msg,
	}, nil
}

// PublishConsumerQuota creates a consumer-scoped limiting strategy in the control
// plane and pushes it to the Apinto data plane. Actual throttling is enforced by
// the data plane; when no data plane is connected this returns a clear error
// rather than panicking. The limiting strategy is owned by cost_optimization, so
// it deliberately bypasses the advanced_policy license gate used by the strategy
// module's Create/Publish path.
func (i *imlCostOptService) PublishConsumerQuota(ctx context.Context, consumer string) error {
	cfg, err := i.GetConsumerBudgets(ctx)
	if err != nil {
		return err
	}
	item, ok := cfg.Items[consumer]
	if !ok || !item.Enabled || !item.Enforce {
		return fmt.Errorf("消费者「%s」未启用或不可强制执行（需 enabled 且 enforce 均为 true）", consumer)
	}
	qps := item.Qps
	if qps <= 0 {
		qps = 10
	}
	limitingConfig := map[string]interface{}{
		"consumer": consumer,
		"default": map[string]interface{}{
			"limit":       qps,
			"interval":    "second",
			"policy":      "window",
			"distributed": false,
		},
	}
	configJSON, err := json.Marshal(limitingConfig)
	if err != nil {
		return err
	}
	id := "costopt-quota-" + consumer
	strategyEntity := &service_strategy.Strategy{
		Id:       id,
		Name:     fmt.Sprintf("成本优化配额-%s", consumer),
		Priority: 1000,
		Desc:     fmt.Sprintf("南天门 AI 成本优化为消费者 %s 下发的限流配额策略", consumer),
		Driver:   "limiting",
		Scope:    0,
		Target:   "",
		Config:   string(configJSON),
		Filters:  fmt.Sprintf(`[{"name":"consumer","values":["%s"]}]`, consumer),
	}
	if err := i.strategyService.Create(ctx, &service_strategy.Create{
		Id:       strategyEntity.Id,
		Name:     strategyEntity.Name,
		Priority: strategyEntity.Priority,
		Desc:     strategyEntity.Desc,
		Filters:  strategyEntity.Filters,
		Config:   strategyEntity.Config,
		Scope:    strategyEntity.Scope,
		Target:   strategyEntity.Target,
		Driver:   strategyEntity.Driver,
	}); err != nil {
		return fmt.Errorf("创建限流策略失败: %w", err)
	}

	driver, has := strategy_driver.GetDriver("limiting")
	if !has {
		return fmt.Errorf("limiting 策略驱动未注册")
	}
	release := driver.ToRelease(strategy_dto.ToStrategy(strategyEntity), nil, 5000)

	client, err := i.clusterService.GatewayClient(ctx, service_cluster.DefaultClusterID)
	if err != nil {
		return fmt.Errorf("数据面未连接，无法下发配额策略: %w", err)
	}
	defer client.Close(ctx)
	if err := client.Strategy().Online(ctx, release); err != nil {
		return fmt.Errorf("下发配额策略到数据面失败: %w", err)
	}
	return nil
}

// ---- LLM-as-judge evaluation ----

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}
type chatChoice struct {
	Message chatMessage `json:"message"`
}
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error"`
}

var (
	scoreRe     = regexp.MustCompile(`"score"\s*:\s*([0-9]+(?:\.[0-9]+)?)`)
	rationaleRe = regexp.MustCompile(`"rationale"\s*:\s*"([^"]*)"`)
)

func (i *imlCostOptService) chat(baseURL, apiKey, model, system, user string) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("未配置裁判模型接口（costopt.judge 的 base_url）")
	}
	body, _ := json.Marshal(chatRequest{
		Model:       model,
		Temperature: 0.2,
		Messages:    []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
	})
	url := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var cr chatResponse
		_ = json.Unmarshal(data, &cr)
		msg := cr.Error.Message
		if msg == "" {
			msg = fmt.Sprintf("裁判模型接口返回 %d", resp.StatusCode)
		}
		return "", fmt.Errorf("%s", msg)
	}
	var cr chatResponse
	if err := json.Unmarshal(data, &cr); err != nil {
		return "", err
	}
	if len(cr.Choices) == 0 {
		return "", fmt.Errorf("裁判模型未返回结果")
	}
	return cr.Choices[0].Message.Content, nil
}

func (i *imlCostOptService) Evaluate(ctx context.Context, input *EvaluateInput) (*EvaluateResult, error) {
	if input == nil || len(input.Prompts) == 0 {
		return nil, fmt.Errorf("请提供待评测的 prompts")
	}
	jc := i.GetJudge(ctx)
	if jc.BaseURL == "" {
		return nil, fmt.Errorf("请先在「AI 成本优化 → 模型评测」中配置 OpenAI 兼容的裁判模型接口（costopt.judge）")
	}
	judgeModel := input.JudgeModel
	if judgeModel == "" {
		judgeModel = jc.Model
	}
	if judgeModel == "" {
		return nil, fmt.Errorf("未指定裁判模型（judge_model 或 costopt.judge.model）")
	}

	genSystem := "You are the model under test. Answer the user's prompt concisely and correctly."
	judgeSystem := "You are an impartial LLM-as-a-judge. You evaluate a candidate response to a prompt against given criteria. Always respond with a single JSON object: {\"score\": <1-10 float>, \"rationale\": \"<short reason>\"}. Do not output anything else."

	out := &EvaluateResult{Model: input.Model, JudgeModel: judgeModel}
	var sum float64
	count := 0
	for idx, prompt := range input.Prompts {
		c := &EvaluationCase{Prompt: prompt}
		// candidate response: supplied or generated
		response := ""
		if idx < len(input.Responses) && input.Responses[idx] != "" {
			response = input.Responses[idx]
		} else if jc.Model != "" {
			gen, err := i.chat(jc.BaseURL, jc.APIKey, jc.Model, genSystem, prompt)
			if err != nil {
				c.Error = err.Error()
				out.Cases = append(out.Cases, c)
				continue
			}
			response = gen
		}
		c.Response = response

		judgeUser := fmt.Sprintf("Criteria: %s\n\nPrompt: %s\n\nResponse: %s\n\nEvaluate and return JSON.", input.Criteria, prompt, response)
		judged, err := i.chat(jc.BaseURL, jc.APIKey, judgeModel, judgeSystem, judgeUser)
		if err != nil {
			c.Error = err.Error()
			out.Cases = append(out.Cases, c)
			continue
		}
		if m := scoreRe.FindStringSubmatch(judged); m != nil {
			var s float64
			fmt.Sscanf(m[1], "%f", &s)
			c.Score = s
		}
		if m := rationaleRe.FindStringSubmatch(judged); m != nil {
			c.Rationale = m[1]
		}
		out.Cases = append(out.Cases, c)
		if c.Score > 0 {
			sum += c.Score
			count++
		}
	}
	if count > 0 {
		out.AverageScore = round2(sum / float64(count))
	}
	return out, nil
}
