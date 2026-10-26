package provider

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const alipayDefaultGateway = "https://openapi.alipay.com/gateway.do"

// AlipayProvider implements PaymentProvider for Alipay (page/website pay).
type AlipayProvider struct {
	cfg ProviderConfig
}

func (a *AlipayProvider) Name() string { return "alipay" }

func (a *AlipayProvider) CreateOrder(ctx context.Context, req *OrderReq) (*OrderResult, error) {
	if a.cfg.AppID == "" {
		return nil, errors.New("alipay: app id not configured")
	}
	endpoint := a.cfg.Endpoint
	if endpoint == "" {
		endpoint = alipayDefaultGateway
	}
	params := url.Values{}
	params.Set("app_id", a.cfg.AppID)
	params.Set("method", "alipay.trade.page.pay")
	params.Set("format", "JSON")
	params.Set("charset", "utf-8")
	params.Set("sign_type", "RSA2")
	params.Set("timestamp", "2024-01-01 00:00:00")
	params.Set("version", "1.0")
	params.Set("notify_url", req.Metadata["notify_url"])
	params.Set("return_url", req.Metadata["return_url"])
	biz := map[string]interface{}{
		"out_trade_no": req.Metadata["order_ref"],
		"product_code": "FAST_INSTANT_TRADE_PAY",
		"total_amount": fmt.Sprintf("%.2f", sumItems(req)),
		"subject":      "南天门 AI 网关订单",
	}
	bizBytes, _ := json.Marshal(biz)
	params.Set("biz_content", string(bizBytes))
	if a.cfg.Secret != "" {
		sign, err := a.signRSA2(params)
		if err != nil {
			return nil, err
		}
		params.Set("sign", sign)
	}
	return &OrderResult{
		CheckoutURL: endpoint + "?" + params.Encode(),
		Raw:         map[string]interface{}{"method": "alipay.trade.page.pay"},
	}, nil
}

// signRSA2 produces an Alipay RSA2 (RSA-SHA256) signature over the sorted params.
func (a *AlipayProvider) signRSA2(params url.Values) (string, error) {
	key, err := parsePrivateKey(a.cfg.Secret)
	if err != nil {
		return "", err
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		if params.Get(k) == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for idx, k := range keys {
		if idx > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params.Get(k))
	}
	h := sha256.Sum256([]byte(buf.String()))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// VerifyWebhook checks the Alipay notification signature (RSA2) when the
// Alipay public key is configured; tolerant when absent (demo mode).
func (a *AlipayProvider) VerifyWebhook(ctx context.Context, body []byte, headers map[string]string) error {
	pubPEM := a.cfg.WebhookSecret
	if pubPEM == "" {
		return nil
	}
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return err
	}
	signB64 := vals.Get("sign")
	if signB64 == "" {
		return errors.New("alipay: missing sign")
	}
	pub, err := parsePublicKey(pubPEM)
	if err != nil {
		return err
	}
	sign, err := base64.StdEncoding.DecodeString(signB64)
	if err != nil {
		return err
	}
	// Rebuild the canonical string from all non-sign fields, sorted.
	keys := make([]string, 0, len(vals))
	for k := range vals {
		if k == "sign" || k == "sign_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for idx, k := range keys {
		if idx > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(vals.Get(k))
	}
	h := sha256.Sum256([]byte(buf.String()))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, h[:], sign); err != nil {
		return errors.New("alipay: signature mismatch")
	}
	return nil
}

func (a *AlipayProvider) ParseWebhook(ctx context.Context, body []byte, headers map[string]string) (*WebhookResult, error) {
	if err := a.VerifyWebhook(ctx, body, headers); err != nil {
		return nil, err
	}
	vals, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	res := &WebhookResult{Raw: map[string]interface{}{}}
	res.OrderID = vals.Get("out_trade_no")
	res.PaymentID = vals.Get("trade_no")
	if vals.Get("trade_status") == "TRADE_SUCCESS" || vals.Get("trade_status") == "TRADE_FINISHED" {
		res.Status = "paid"
	} else {
		res.Status = "pending"
	}
	return res, nil
}

var _ PaymentProvider = (*AlipayProvider)(nil)

func sumItems(req *OrderReq) float64 {
	t := 0.0
	for _, it := range req.Items {
		t += it.Amount
	}
	return t
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	pemStr = stripPemWrapper(pemStr, "PRIVATE KEY")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("alipay: invalid private key pem")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("alipay: not an RSA private key")
	}
	return rsaKey, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	pemStr = stripPemWrapper(pemStr, "PUBLIC KEY")
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("alipay: invalid public key pem")
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("alipay: not an RSA public key")
	}
	return rsaKey, nil
}

func stripPemWrapper(s, label string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "-----BEGIN") {
		return s
	}
	return "-----BEGIN " + label + "-----\n" + s + "\n-----END " + label + "-----"
}
