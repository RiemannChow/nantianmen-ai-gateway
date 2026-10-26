package multitenant

// KeyTenants is the setting key holding the tenant definitions.
const KeyTenants = "multitenant.tenants"

// TenantItem defines one tenant. Its Id maps to a metering `subscriber` value,
// so tenant metering/isolation is anchored on existing usage attribution.
type TenantItem struct {
	Id            string  `json:"id"`             // == metering subscriber
	Name          string  `json:"name"`           // 展示名
	MonthlyBudget float64 `json:"monthly_budget"` // 月度支出上限（元）
	QuotaCalls    int64   `json:"quota_calls"`    // 月度调用配额（次）；0=不限
	Enabled       bool    `json:"enabled"`
}

// TenantConfig is the persisted tenant set.
type TenantConfig struct {
	Items map[string]TenantItem `json:"items"`
}

// TenantUsage is the raw per-tenant usage in a window.
type TenantUsage struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Calls        int64  `json:"calls"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	TotalTokens  int64  `json:"total_tokens"`
	LatencyMs    int64  `json:"latency_ms"`
}

// TenantBilling is the per-tenant spend derived from billing pricing.
type TenantBilling struct {
	Id            string  `json:"id"`
	Name          string  `json:"name"`
	Calls         int64   `json:"calls"`
	InputTokens   int64   `json:"input_tokens"`
	OutputTokens  int64   `json:"output_tokens"`
	TotalTokens   int64   `json:"total_tokens"`
	CallCost      float64 `json:"call_cost"`
	TokenCost     float64 `json:"token_cost"`
	Bill          float64 `json:"bill"`
	MonthlyBudget float64 `json:"monthly_budget"`
	Ratio         float64 `json:"ratio"`
	Level         string  `json:"level"` // ok | warning | exceeded
}

// TenantStatus is a lightweight budget-state projection for a single tenant.
type TenantStatus struct {
	Id            string  `json:"id"`
	Name          string  `json:"name"`
	Configured    bool    `json:"configured"`
	MonthlyBudget float64 `json:"monthly_budget"`
	MonthSpend    float64 `json:"month_spend"`
	Ratio         float64 `json:"ratio"`
	Level         string  `json:"level"`
	Message       string  `json:"message"`
}
