package provider

import "context"

// ProviderConfig holds the credentials/endpoints for a payment provider. Values
// are sourced from KV (payment.<provider>.*) and never logged.
type ProviderConfig struct {
	APIKey        string
	Secret        string
	WebhookSecret string
	AppID         string
	Endpoint      string
}

// OrderItem is a single line in a checkout order, tied to a license module.
type OrderItem struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	LicenseModuleKey string  `json:"license_module_key"`
	Amount           float64 `json:"amount"`
}

// OrderReq is the normalized checkout request passed to a provider.
type OrderReq struct {
	Customer string
	Currency string
	Items    []OrderItem
	Metadata map[string]string
}

// OrderResult is the normalized result of creating a checkout session.
type OrderResult struct {
	OrderID     string
	PaymentID   string
	CheckoutURL string
	Raw         map[string]interface{}
}

// WebhookResult is the normalized outcome of a payment callback.
type WebhookResult struct {
	OrderID   string
	PaymentID string
	Status    string // "paid" | "failed" | "pending"
	Raw       map[string]interface{}
}

// PaymentProvider abstracts a third-party payment gateway.
type PaymentProvider interface {
	// Name returns the provider identifier (stripe | alipay | wechat).
	Name() string
	// CreateOrder creates a checkout session and returns the redirect/pay URL.
	CreateOrder(ctx context.Context, req *OrderReq) (*OrderResult, error)
	// VerifyWebhook validates the callback signature. It returns a non-nil error
	// when the signature is missing/invalid and a secret is configured.
	VerifyWebhook(ctx context.Context, body []byte, headers map[string]string) error
	// ParseWebhook extracts the normalized result from a callback payload.
	ParseWebhook(ctx context.Context, body []byte, headers map[string]string) (*WebhookResult, error)
}

// New returns a PaymentProvider for the given type, or nil if unsupported.
func New(providerType string, cfg ProviderConfig) PaymentProvider {
	switch providerType {
	case "stripe":
		return &StripeProvider{cfg: cfg}
	case "alipay":
		return &AlipayProvider{cfg: cfg}
	case "wechat":
		return &WechatProvider{cfg: cfg}
	default:
		return nil
	}
}
