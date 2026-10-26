package provider

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/payment/provider/xml"
)

const wechatUnifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"

// WechatProvider implements PaymentProvider for WeChat Pay (APIv2, NATIVE/MWEB).
type WechatProvider struct {
	cfg ProviderConfig
}

func (w *WechatProvider) Name() string { return "wechat" }

func (w *WechatProvider) CreateOrder(ctx context.Context, req *OrderReq) (*OrderResult, error) {
	if w.cfg.AppID == "" || w.cfg.WebhookSecret == "" {
		return nil, errors.New("wechat: app id and api key (mch key) are required")
	}
	params := map[string]string{
		"appid":            w.cfg.AppID,
		"mch_id":           w.cfg.Secret, // merchant id carried in Secret for parity with config shape
		"nonce_str":        randomNonce(),
		"body":             "南天门 AI 网关订单",
		"out_trade_no":     req.Metadata["order_ref"],
		"total_fee":        fmt.Sprintf("%d", int64(sumItems(req)*100)),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       req.Metadata["notify_url"],
		"trade_type":       "NATIVE",
	}
	params["sign"] = signWechat(params, w.cfg.WebhookSecret)
	body := xml.MarshalMap("xml", params)
	resp, err := postXML(ctx, wechatUnifiedOrderURL, body)
	if err != nil {
		return nil, err
	}
	res, err := xml.ParseMap(resp)
	if err != nil {
		return nil, err
	}
	result := &OrderResult{Raw: map[string]interface{}{"return_code": res["return_code"]}}
	if res["result_code"] == "SUCCESS" {
		result.PaymentID = res["prepay_id"]
		result.CheckoutURL = res["code_url"]
	}
	return result, nil
}

// VerifyWebhook validates the WeChat notify signature (HMAC-SHA256) when the API
// key is configured; tolerant when absent (demo mode).
func (w *WechatProvider) VerifyWebhook(ctx context.Context, body []byte, headers map[string]string) error {
	if w.cfg.WebhookSecret == "" {
		return nil
	}
	res, err := xml.ParseMap(body)
	if err != nil {
		return err
	}
	sign := res["sign"]
	delete(res, "sign")
	computed := signWechat(res, w.cfg.WebhookSecret)
	if !hmac.Equal([]byte(strings.ToUpper(computed)), []byte(strings.ToUpper(sign))) {
		return errors.New("wechat: signature mismatch")
	}
	return nil
}

func (w *WechatProvider) ParseWebhook(ctx context.Context, body []byte, headers map[string]string) (*WebhookResult, error) {
	if err := w.VerifyWebhook(ctx, body, headers); err != nil {
		return nil, err
	}
	res, err := xml.ParseMap(body)
	if err != nil {
		return nil, err
	}
	r := &WebhookResult{Raw: map[string]interface{}{}}
	r.OrderID = res["out_trade_no"]
	r.PaymentID = res["transaction_id"]
	if res["return_code"] == "SUCCESS" && res["result_code"] == "SUCCESS" {
		r.Status = "paid"
	} else {
		r.Status = "pending"
	}
	return r, nil
}

var _ PaymentProvider = (*WechatProvider)(nil)

func signWechat(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] == "" {
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
		buf.WriteString(params[k])
	}
	buf.WriteString("&key=")
	buf.WriteString(key)
	h := sha256.Sum256([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

// randomNonce uses crypto/rand for an unpredictable nonce.
func randomNonce() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "nonce"
	}
	return hex.EncodeToString(b)
}
