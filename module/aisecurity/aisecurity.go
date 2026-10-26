package aisecurity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"
)

// IAIsecurityService is the control-plane AI-security suite: rule configuration
// and a local heuristic scan used for testing/demonstration of the same
// detections the data-plane drivers (pii-mask / prompt-injection /
// content-moderation) enforce in production.
type IAIsecurityService interface {
	GetRules(ctx context.Context) *AISecurityRules
	SaveRules(ctx context.Context, cfg *AISecurityRules) (*AISecurityRules, error)
	Scan(ctx context.Context, input *ScanInput) (*ScanResult, error)
}

func init() {
	autowire.Auto[IAIsecurityService](func() reflect.Value {
		return reflect.ValueOf(new(imlAIsecurityService))
	})
}

type imlAIsecurityService struct {
	settingService setting.ISettingService `autowired:""`
}

func operator(ctx context.Context) string {
	if id := utils.UserId(ctx); id != "" {
		return id
	}
	return "system"
}

func (i *imlAIsecurityService) GetRules(ctx context.Context) *AISecurityRules {
	raw, has := i.settingService.Get(ctx, KeyAISecurityRules)
	if !has || raw == "" {
		return &AISecurityRules{}
	}
	var cfg AISecurityRules
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return &AISecurityRules{}
	}
	return &cfg
}

func (i *imlAIsecurityService) SaveRules(ctx context.Context, cfg *AISecurityRules) (*AISecurityRules, error) {
	if cfg == nil {
		cfg = &AISecurityRules{}
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	if err := i.settingService.Set(ctx, KeyAISecurityRules, string(raw), operator(ctx)); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Scan runs the local heuristic detection. Built-in default patterns/keywords
// always apply (so the scanner is useful out of the box); configured rules are
// merged on top. When an external moderator is enabled its result is folded in.
func (i *imlAIsecurityService) Scan(ctx context.Context, input *ScanInput) (*ScanResult, error) {
	if input == nil || strings.TrimSpace(input.Text) == "" {
		return nil, fmt.Errorf("请提供待检测文本（scan.text）")
	}
	cfg := i.GetRules(ctx)
	types := map[string]bool{}
	for _, t := range input.Types {
		types[strings.ToLower(t)] = true
	}
	want := func(t string) bool { return len(types) == 0 || types[t] }

	findings := make([]Finding, 0)

	// PII
	if want("pii") {
		patterns := append([]PIIPattern{}, defaultPIIPatterns...)
		patterns = append(patterns, cfg.PIIPatterns...)
		for _, p := range patterns {
			re, err := regexp.Compile(p.Regex)
			if err != nil {
				continue
			}
			for _, loc := range re.FindAllStringIndex(input.Text, -1) {
				findings = append(findings, Finding{
					Type:     "pii",
					Rule:     p.Name,
					Match:    input.Text[loc[0]:loc[1]],
					Action:   p.Action,
					Position: loc[0],
				})
			}
		}
	}

	// Prompt injection
	if want("injection") {
		lower := strings.ToLower(input.Text)
		keywords := append([]string{}, defaultInjectionKeywords...)
		keywords = append(keywords, cfg.InjectionKeywords...)
		for _, kw := range keywords {
			if kw == "" {
				continue
			}
			if idx := strings.Index(lower, strings.ToLower(kw)); idx >= 0 {
				findings = append(findings, Finding{
					Type:     "prompt_injection",
					Rule:     "关键词",
					Match:    input.Text[idx : idx+len(kw)],
					Action:   ActionFlag,
					Position: idx,
				})
			}
		}
	}

	// Content moderation
	if want("moderation") {
		words := append([]string{}, defaultModerationWords...)
		words = append(words, cfg.ModerationWords...)
		for _, w := range words {
			if w == "" {
				continue
			}
			if idx := strings.Index(input.Text, w); idx >= 0 {
				findings = append(findings, Finding{
					Type:     "moderation",
					Rule:     "敏感词",
					Match:    input.Text[idx : idx+len(w)],
					Action:   ActionFlag,
					Position: idx,
				})
			}
		}
	}

	// External moderator (optional)
	if cfg.ExternalModerator != nil && cfg.ExternalModerator.Enabled && cfg.ExternalModerator.URL != "" {
		if ext, err := i.externalModerate(cfg.ExternalModerator, input.Text); err == nil {
			findings = append(findings, ext...)
		}
	}

	return &ScanResult{
		Safe:      len(findings) == 0,
		Findings:  findings,
		ScannedAt: time.Now().Unix(),
	}, nil
}

// externalModerate posts the text to an external moderation endpoint and folds
// any reported categories into findings. The endpoint is expected to return a
// JSON body; any field that is truthy/list is treated as a flagged category.
func (i *imlAIsecurityService) externalModerate(cfg *ModeratorConfig, text string) ([]Finding, error) {
	body, _ := json.Marshal(map[string]string{"input": text})
	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("外部审核端点返回 %d", resp.StatusCode)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	findings := make([]Finding, 0)
	for k, v := range parsed {
		flagged := false
		switch val := v.(type) {
		case bool:
			flagged = val
		case []interface{}:
			flagged = len(val) > 0
		case string:
			flagged = val != "" && val != "false"
		}
		if flagged {
			findings = append(findings, Finding{
				Type:   "moderation",
				Rule:   "外部审核:" + k,
				Match:  k,
				Action: ActionFlag,
			})
		}
	}
	return findings, nil
}
