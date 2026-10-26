package payment

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/module/license"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/payment/provider"
	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/payment"

	"github.com/google/uuid"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

// demoPrivateKey signs delivered licenses in this open-source/demo build. For
// production, override payment.license_private_key with the vendor's real
// Ed25519 private key (must match the public key in module/license/signer.go).
const demoPrivateKey = "c3UgqulxPN7d53wrWY9btsv9foHi7xhJ76BwXOA7XpifqMSbtjF+z/DoMXzSeDlQLqgzFMKdG+rirfAKq9xlzw=="

// CreateInput is the checkout request from the UI.
type CreateInput struct {
	Provider   string               `json:"provider"`
	Customer   string               `json:"customer"`
	Currency   string               `json:"currency"`
	Items      []provider.OrderItem `json:"items"`
	SuccessURL string               `json:"success_url"`
	CancelURL  string               `json:"cancel_url"`
	NotifyURL  string               `json:"notify_url"`
}

// OrderView is an order as exposed to the UI.
type OrderView struct {
	UUID             string  `json:"uuid"`
	Customer         string  `json:"customer"`
	Provider         string  `json:"provider"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	LicenseModuleKey string  `json:"license_module_key"`
	CheckoutURL      string  `json:"checkout_url,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

// IPaymentModule manages orders and delivers licenses on successful payment.
type IPaymentModule interface {
	CreateOrder(ctx context.Context, input *CreateInput) (*OrderView, error)
	GetOrder(ctx context.Context, uuid string) (*OrderView, error)
	ListOrders(ctx context.Context) ([]*OrderView, error)
	Webhook(ctx context.Context, providerType string, body []byte, headers map[string]string) error
}

func init() {
	autowire.Auto[IPaymentModule](func() reflect.Value {
		return reflect.ValueOf(new(imlPaymentModule))
	})
}

var _ IPaymentModule = (*imlPaymentModule)(nil)

type imlPaymentModule struct {
	orderStore     payment.IOrderStore     `autowired:""`
	paymentStore   payment.IPaymentStore   `autowired:""`
	settingService setting.ISettingService `autowired:""`
	licenseService license.ILicenseService `autowired:""`
}

// providerConfig reads the provider credentials from KV (transparently
// decrypted for any sensitive keys in the setting whitelist).
func (i *imlPaymentModule) providerConfig(ctx context.Context, providerType string) provider.ProviderConfig {
	prefix := "payment." + providerType + "."
	apiKey, _ := i.settingService.Get(ctx, prefix+"api_key")
	secret, _ := i.settingService.Get(ctx, prefix+"secret")
	wsecret, _ := i.settingService.Get(ctx, prefix+"webhook_secret")
	appID, _ := i.settingService.Get(ctx, prefix+"app_id")
	endpoint, _ := i.settingService.Get(ctx, prefix+"endpoint")
	return provider.ProviderConfig{
		APIKey:        apiKey,
		Secret:        secret,
		WebhookSecret: wsecret,
		AppID:         appID,
		Endpoint:      endpoint,
	}
}

func (i *imlPaymentModule) CreateOrder(ctx context.Context, input *CreateInput) (*OrderView, error) {
	if input.Provider == "" {
		return nil, errors.New("provider is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}
	prov := provider.New(input.Provider, i.providerConfig(ctx, input.Provider))
	if prov == nil {
		return nil, errors.New("unsupported payment provider")
	}
	req := &provider.OrderReq{
		Customer: input.Customer,
		Currency: input.Currency,
		Items:    input.Items,
		Metadata: map[string]string{
			"success_url": input.SuccessURL,
			"cancel_url":  input.CancelURL,
			"notify_url":  input.NotifyURL,
			"order_ref":   uuid.NewString(),
		},
	}
	result, err := prov.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}
	itemsBytes, _ := json.Marshal(input.Items)
	now := time.Now()
	order := &payment.Order{
		UUID:             req.Metadata["order_ref"],
		Customer:         input.Customer,
		Provider:         input.Provider,
		Amount:           sumItems(input.Items),
		Currency:         input.Currency,
		Items:            string(itemsBytes),
		LicenseModuleKey: input.Items[0].LicenseModuleKey,
		Status:           "created",
		Creator:          utils.UserId(ctx),
		Updater:          utils.UserId(ctx),
		CreateAt:         now,
		UpdateAt:         now,
	}
	if err := i.orderStore.Insert(ctx, order); err != nil {
		return nil, err
	}
	view := toOrderView(order)
	view.CheckoutURL = result.CheckoutURL
	return view, nil
}

func (i *imlPaymentModule) GetOrder(ctx context.Context, id string) (*OrderView, error) {
	orders, err := i.orderStore.Search(ctx, "", map[string]interface{}{"uuid": id}, "")
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, errors.New("order not found")
	}
	return toOrderView(orders[0]), nil
}

func (i *imlPaymentModule) ListOrders(ctx context.Context) ([]*OrderView, error) {
	orders, err := i.orderStore.Search(ctx, "", nil, "create_at desc")
	if err != nil {
		return nil, err
	}
	list := make([]*OrderView, 0, len(orders))
	for _, o := range orders {
		list = append(list, toOrderView(o))
	}
	return list, nil
}

func (i *imlPaymentModule) Webhook(ctx context.Context, providerType string, body []byte, headers map[string]string) error {
	prov := provider.New(providerType, i.providerConfig(ctx, providerType))
	if prov == nil {
		return errors.New("unsupported payment provider")
	}
	result, err := prov.ParseWebhook(ctx, body, headers)
	if err != nil {
		return err
	}
	if result.Status != "paid" {
		return nil
	}
	orders, err := i.orderStore.Search(ctx, "", map[string]interface{}{"uuid": result.OrderID}, "")
	if err != nil || len(orders) == 0 {
		return errors.New("order not found for webhook")
	}
	order := orders[0]
	if order.Status == "paid" {
		return nil // idempotent
	}
	now := time.Now()
	order.Status = "paid"
	order.Updater = utils.UserId(ctx)
	order.UpdateAt = now

	// Deliver the license for the purchased module.
	key := i.licensePrivateKey(ctx)
	signed, serr := license.Sign(key, &license.License{
		Edition:  license.EditionEnterprise,
		Customer: order.Customer,
		Modules:  []string{order.LicenseModuleKey},
		MaxNodes: 0,
		IssuedAt: now.Unix(),
		Expiry:   0,
	})
	if serr == nil {
		if payload, jerr := json.Marshal(signed); jerr == nil {
			order.LicensePayload = string(payload)
			// Activate so the module is unlocked immediately in this build.
			_, _ = i.licenseService.Activate(ctx, string(payload))
		}
	}

	txn := &payment.Payment{
		UUID:        uuid.NewString(),
		OrderUUID:   order.UUID,
		Provider:    order.Provider,
		ProviderTxn: result.PaymentID,
		Amount:      order.Amount,
		Status:      "paid",
		RawPayload:  string(body),
		Creator:     utils.UserId(ctx),
		CreateAt:    now,
	}
	if _, err := i.orderStore.Update(ctx, order); err != nil {
		return err
	}
	return i.paymentStore.Insert(ctx, txn)
}

func (i *imlPaymentModule) licensePrivateKey(ctx context.Context) string {
	if v, has := i.settingService.Get(ctx, "payment.license_private_key"); has && v != "" {
		return v
	}
	return demoPrivateKey
}

func sumItems(items []provider.OrderItem) float64 {
	t := 0.0
	for _, it := range items {
		t += it.Amount
	}
	return t
}

func toOrderView(o *payment.Order) *OrderView {
	return &OrderView{
		UUID:             o.UUID,
		Customer:         o.Customer,
		Provider:         o.Provider,
		Amount:           o.Amount,
		Currency:         o.Currency,
		Status:           o.Status,
		LicenseModuleKey: o.LicenseModuleKey,
		CreatedAt:        o.CreateAt.Format("2006-01-02 15:04:05"),
	}
}
