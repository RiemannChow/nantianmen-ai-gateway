package costopt

// Setting keys persisted via service/setting (KV store).
const (
	// KeyBudget holds the JSON-encoded BudgetConfig.
	KeyBudget = "costopt.budget"
	// KeyProviders holds the JSON-encoded map[provider]ProviderCost.
	KeyProviders = "costopt.providers"
	// KeySubstitutes holds the JSON-encoded map[fromProvider]toProvider.
	KeySubstitutes = "costopt.substitutes"
	// KeyJudge holds the JSON-encoded JudgeConfig (OpenAI-compatible endpoint).
	KeyJudge = "costopt.judge"
	// KeyNotify holds the JSON-encoded NotifyConfig (alert delivery channels).
	KeyNotify = "costopt.notify"
	// KeyNotifyState holds the JSON-encoded map of per-signal alert state used
	// for de-duplication / cooldown of outgoing notifications.
	KeyNotifyState = "costopt.notify_state"
	// KeyConsumerBudgets holds the JSON-encoded ConsumerBudgetConfig (per
	// subscriber/key spend caps + optional limiting enforcement).
	KeyConsumerBudgets = "costopt.consumer_budgets"
)

// BudgetConfig is the monthly spend budget and alerting threshold.
type BudgetConfig struct {
	MonthlyBudget  float64 `json:"monthly_budget"`  // 月度预算（结算币种）
	AlertThreshold float64 `json:"alert_threshold"` // 预警阈值（占预算比例 0-1）
}

// ProviderCost is the per-provider unit cost used to compare models and to
// project savings when shifting traffic between providers.
type ProviderCost struct {
	PerCall          float64 `json:"per_call"`
	Per1KInputToken  float64 `json:"per_1k_input_token"`
	Per1KOutputToken float64 `json:"per_1k_output_token"`
}

// JudgeConfig points at an OpenAI-compatible chat-completions endpoint used by
// the LLM-as-judge evaluator (both to generate candidate responses and to judge).
type JudgeConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// Window is a [start, end] time range in unix seconds.
type Window struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// RecommendationItem is one concrete cost-saving suggestion.
type RecommendationItem struct {
	Type             string  `json:"type"` // model_substitution | caching | right_sizing | generic
	Title            string  `json:"title"`
	Detail           string  `json:"detail"`
	FromProvider     string  `json:"from_provider,omitempty"`
	ToProvider       string  `json:"to_provider,omitempty"`
	ShiftedCalls     int64   `json:"shifted_calls,omitempty"`
	ProjectedSavings float64 `json:"projected_savings"`
}

// RecommendationResult is the output of the cost-optimization advisor.
type RecommendationResult struct {
	Window                Window                  `json:"window"`
	Pricing               map[string]float64      `json:"pricing"` // global pricing used as fallback
	ProviderCosts         map[string]ProviderCost `json:"provider_costs"`
	Recommendations       []*RecommendationItem   `json:"recommendations"`
	TotalProjectedSavings float64                 `json:"total_projected_savings"`
}

// ProviderSpend is one provider's usage + spend in a window.
type ProviderSpend struct {
	Provider     string  `json:"provider"`
	Calls        int64   `json:"calls"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	Spend        float64 `json:"spend"`
	Share        float64 `json:"share"` // fraction of total spend
}

// AnalyticsResult is the advanced spend-analytics breakdown.
type AnalyticsResult struct {
	Window      Window           `json:"window"`
	TotalSpend  float64          `json:"total_spend"`
	TotalCalls  int64            `json:"total_calls"`
	TotalTokens int64            `json:"total_tokens"`
	Providers   []*ProviderSpend `json:"providers"`
}

// AnomalyItem is one detected anomaly / risk.
type AnomalyItem struct {
	Type     string `json:"type"`     // budget_at_risk | concentration | no_budget
	Severity string `json:"severity"` // info | warning | critical
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

// AnomalyResult is the anomaly-detection output.
type AnomalyResult struct {
	Window    Window         `json:"window"`
	Anomalies []*AnomalyItem `json:"anomalies"`
}

// BudgetStatus compares month-to-date spend with the configured budget.
type BudgetStatus struct {
	Configured     bool    `json:"configured"`
	MonthlyBudget  float64 `json:"monthly_budget"`
	AlertThreshold float64 `json:"alert_threshold"`
	MonthStart     int64   `json:"month_start"`
	Now            int64   `json:"now"`
	MonthSpend     float64 `json:"month_spend"`
	Ratio          float64 `json:"ratio"` // month_spend / monthly_budget
	Level          string  `json:"level"` // ok | warning | exceeded
	Message        string  `json:"message"`
}

// EvaluateInput is the request body for the LLM-as-judge evaluation.
type EvaluateInput struct {
	Model      string   `json:"model"`       // model under test (label only)
	Prompts    []string `json:"prompts"`     // test prompts
	Responses  []string `json:"responses"`   // optional pre-supplied candidate responses
	Criteria   string   `json:"criteria"`    // judging criteria
	JudgeModel string   `json:"judge_model"` // optional override of JudgeConfig.Model
	ShiftRatio float64  `json:"shift_ratio"` // fraction (0-1) used by substitution recs
}

// EvaluationCase is one judged prompt.
type EvaluationCase struct {
	Prompt    string  `json:"prompt"`
	Response  string  `json:"response"`
	Score     float64 `json:"score"` // 0-10
	Rationale string  `json:"rationale"`
	Error     string  `json:"error,omitempty"`
}

// EvaluateResult aggregates the judge scores.
type EvaluateResult struct {
	Model        string            `json:"model"`
	JudgeModel   string            `json:"judge_model"`
	AverageScore float64           `json:"average_score"`
	Cases        []*EvaluationCase `json:"cases"`
}

// ---- alert delivery (notification) ----

// WebhookChannelConfig is a generic HTTP webhook (covers Slack / 飞书 / 钉钉 / any
// HTTP endpoint). Payload is rendered from Template via text/template.
type WebhookChannelConfig struct {
	Enabled  bool              `json:"enabled"`
	URL      string            `json:"url"`
	Method   string            `json:"method"`   // default POST
	Headers  map[string]string `json:"headers"`  // extra headers, e.g. Authorization
	Template string            `json:"template"` // body template; {{.Title}} {{.Body}} {{.Severity}} etc.
}

// EmailChannelConfig is an SMTP-delivered alert.
type EmailChannelConfig struct {
	Enabled         bool     `json:"enabled"`
	To              []string `json:"to"`
	SubjectTemplate string   `json:"subject_template"`
	BodyTemplate    string   `json:"body_template"`
}

// NotifyConfig holds the enabled alert delivery channels.
type NotifyConfig struct {
	Webhook WebhookChannelConfig `json:"webhook"`
	Email   EmailChannelConfig   `json:"email"`
}

// AlertPayload is the rendered alert handed to each channel.
type AlertPayload struct {
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Severity string                 `json:"severity"` // info | warning | critical
	Source   string                 `json:"source"`   // signal key, e.g. "budget"
	Time     int64                  `json:"time"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// AlertState persists the last-sent state of a signal for de-dup / cooldown.
type AlertState struct {
	LastRank      int    `json:"last_rank"`      // last severity rank sent (0=none/ok)
	LastSeverity  string `json:"last_severity"`  // last severity bucket sent
	LastFiredAt   int64  `json:"last_fired_at"`  // unix sec of last dispatch
	CooldownUntil int64  `json:"cooldown_until"` // unix sec; do not fire before this
}

// ---- per-consumer / per-key budget & quota ----

// ConsumerBudgetItem is a per-subscriber spend cap.
type ConsumerBudgetItem struct {
	Consumer       string  `json:"consumer"`        // == metering subscriber id
	MonthlyBudget  float64 `json:"monthly_budget"`  // monthly spend cap (settlement currency)
	AlertThreshold float64 `json:"alert_threshold"` // 0-1, ratio at which a warning fires
	Qps            float64 `json:"qps"`             // rate-limit (requests/sec) pushed as limiting policy when Enforce
	Enabled        bool    `json:"enabled"`         // include in eval / anomalies
	Enforce        bool    `json:"enforce"`         // push a limiting strategy to the data plane
}

// ConsumerBudgetConfig holds the per-consumer budget map.
type ConsumerBudgetConfig struct {
	Items map[string]ConsumerBudgetItem `json:"items"` // keyed by consumer id
}

// ConsumerBudgetStatus is the live evaluation of one consumer's month-to-date spend.
type ConsumerBudgetStatus struct {
	Configured     bool    `json:"configured"`
	Consumer       string  `json:"consumer"`
	MonthlyBudget  float64 `json:"monthly_budget"`
	AlertThreshold float64 `json:"alert_threshold"`
	MonthSpend     float64 `json:"month_spend"`
	Ratio          float64 `json:"ratio"`
	Level          string  `json:"level"` // ok | warning | exceeded
	Message        string  `json:"message"`
}
