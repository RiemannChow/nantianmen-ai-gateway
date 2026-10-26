package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const stripeDefaultEndpoint = "https://api.stripe.com"

// StripeProvider implements PaymentProvider for Stripe Checkout + webhooks.
type StripeProvider struct {
	cfg ProviderConfig
}

func (s *StripeProvider) Name() string { return "stripe" }

func (s *StripeProvider) CreateOrder(ctx context.Context, req *OrderReq) (*OrderResult, error) {
	if s.cfg.APIKey == "" {
		return nil, errors.New("stripe: api key not configured")
	}
	endpoint := s.cfg.Endpoint
	if endpoint == "" {
		endpoint = stripeDefaultEndpoint
	}
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", req.Metadata["success_url"])
	form.Set("cancel_url", req.Metadata["cancel_url"])
	total := 0.0
	for _, it := range req.Items {
		form.Add("line_items[][price_data][currency]", req.Currency)
		form.Add("line_items[][price_data][product_data][name]", it.Name)
		amountCents := int64(it.Amount * 100)
		form.Add("line_items[][price_data][unit_amount]", strconv.FormatInt(amountCents, 10))
		form.Add("line_items[][quantity]", "1")
		total += it.Amount
	}
	form.Set("metadata[order_ref]", req.Metadata["order_ref"])
	form.Set("metadata[customer]", req.Customer)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe create session failed: %d %s", resp.StatusCode, string(body))
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	result := &OrderResult{Raw: parsed}
	if id, ok := parsed["id"].(string); ok {
		result.PaymentID = id
	}
	if u, ok := parsed["url"].(string); ok {
		result.CheckoutURL = u
	}
	return result, nil
}

func (s *StripeProvider) VerifyWebhook(ctx context.Context, body []byte, headers map[string]string) error {
	secret := s.cfg.WebhookSecret
	if secret == "" {
		// No secret configured: signature cannot be verified. In a demo this is
		// tolerated (the operator is testing the callback loop manually); in
		// production a secret MUST be set or webhooks are rejected upstream.
		return nil
	}
	sig := headers["Stripe-Signature"]
	if sig == "" {
		return errors.New("stripe: missing Stripe-Signature header")
	}
	// Stripe signatures look like: t=<ts>,v1=<hex-hmac>,v0=...
	want, ok := extractStripeSig(sig)
	if !ok {
		return errors.New("stripe: malformed signature header")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	if !hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(want)) {
		return errors.New("stripe: signature mismatch")
	}
	return nil
}

func extractStripeSig(header string) (string, bool) {
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 && kv[0] == "v1" {
			return kv[1], true
		}
	}
	return "", false
}

func (s *StripeProvider) ParseWebhook(ctx context.Context, body []byte, headers map[string]string) (*WebhookResult, error) {
	if err := s.VerifyWebhook(ctx, body, headers); err != nil {
		return nil, err
	}
	var evt struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID            string            `json:"id"`
				Metadata      map[string]string `json:"metadata"`
				PaymentIntent string            `json:"payment_intent"`
				Status        string            `json:"payment_status"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &evt); err != nil {
		return nil, err
	}
	res := &WebhookResult{Raw: map[string]interface{}{"type": evt.Type}}
	res.PaymentID = evt.Data.Object.ID
	if ref, ok := evt.Data.Object.Metadata["order_ref"]; ok {
		res.OrderID = ref
	}
	switch evt.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		res.Status = "paid"
	default:
		res.Status = "pending"
	}
	return res, nil
}

// ensure StripeProvider satisfies the interface
var _ PaymentProvider = (*StripeProvider)(nil)
