package billing

// Setting keys persisted via service/setting (KV store).
const (
	// KeyPricing holds the JSON-encoded PricingConfig.
	KeyPricing = "billing.pricing"
	// KeyRebate holds the JSON-encoded RebateConfig.
	KeyRebate = "billing.rebate"
)

// PricingConfig holds the unit pricing used for monthly bill computation.
// All monetary amounts are expressed in the gateway operator's settlement
// currency (e.g. CNY). Empty/zero fields mean "not charged".
type PricingConfig struct {
	PerCall          float64 `json:"per_call"`            // 每调用次单价
	Per1KInputToken  float64 `json:"per_1k_input_token"`  // 每 1K 输入 Token 单价
	Per1KOutputToken float64 `json:"per_1k_output_token"` // 每 1K 输出 Token 单价
	PerInstanceMonth float64 `json:"per_instance_month"`  // 每网关实例·月 单价
}

// DefaultPricing returns a zero-valued pricing config (everything free) used
// when no pricing has been configured yet.
func DefaultPricing() *PricingConfig {
	return &PricingConfig{}
}

// RebateConfig maps a model-provider identifier (e.g. "openai", "azure",
// "通义千问") to the channel commission the vendor pays the operator per call
// routed to that provider (返佣单价/次). Providers absent from the map earn no
// rebate.
type RebateConfig map[string]float64

// ProviderBill is one model provider's contribution to the bill and rebate.
type ProviderBill struct {
	Provider     string  `json:"provider"`
	Calls        int64   `json:"calls"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	CallCost     float64 `json:"call_cost"`
	TokenCost    float64 `json:"token_cost"`
	Bill         float64 `json:"bill"`
	Rebate       float64 `json:"rebate"`
}

// BillingOverview is the computed bill + channel rebate for a time window.
type BillingOverview struct {
	Start        int64           `json:"start"`
	End          int64           `json:"end"`
	Calls        int64           `json:"calls"`
	InputTokens  int64           `json:"input_tokens"`
	OutputTokens int64           `json:"output_tokens"`
	Instances    int             `json:"instances"`
	Pricing      *PricingConfig  `json:"pricing"`
	CallCost     float64         `json:"call_cost"`
	TokenCost    float64         `json:"token_cost"`
	InstanceCost float64         `json:"instance_cost"`
	BillTotal    float64         `json:"bill_total"`
	RebateTotal  float64         `json:"rebate_total"`
	NetTotal     float64         `json:"net_total"`
	Providers    []*ProviderBill `json:"providers"`
}
