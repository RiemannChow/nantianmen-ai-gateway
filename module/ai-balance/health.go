package ai_balance

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	ai_provider_local "github.com/RiemannChow/nantianmen-ai-gateway/ai-provider/local"

	ai_balance "github.com/RiemannChow/nantianmen-ai-gateway/service/ai-balance"

	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/cluster"

	"github.com/eolinker/eosc/log"
)

// autoFailoverKey gates the scheduled automatic failover. When set to "true" the
// periodic probe loop runs and flips abnormal models offline automatically.
const autoFailoverKey = "ai_balance.auto_failover"

// probeInterval is how often the scheduled health probe runs.
const probeInterval = 1 * time.Minute

func (i *imlBalanceModule) startHealthProbe(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(probeInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				enabled, _ := i.settingService.Get(ctx, autoFailoverKey)
				if enabled != "true" {
					continue
				}
				i.probeAll(ctx)
			}
		}
	}()
}

// probeAll probes every model, updates its health record, and flips the gateway
// state when a model's health changes.
func (i *imlBalanceModule) probeAll(ctx context.Context) {
	list, err := i.balanceService.Search(ctx, "", nil, "priority asc")
	if err != nil {
		log.Errorf("health probe search error: %v", err)
		return
	}
	for _, item := range list {
		healthy := i.probeOne(ctx, item)
		i.applyHealth(ctx, item, healthy)
	}
}

// applyHealth persists the probe result and, when the state flips, re-syncs the
// gateway (online/offline) for that model.
func (i *imlBalanceModule) applyHealth(ctx context.Context, item *ai_balance.Balance, healthy bool) {
	now := time.Now().Unix()
	newState := 1
	failCount := item.FailCount
	lastSuccessAt := item.LastSuccessAt
	if healthy {
		lastSuccessAt = now
	} else {
		newState = 0
		failCount = item.FailCount + 1
	}
	if err := i.balanceService.UpdateHealth(ctx, item.Id, newState, now, lastSuccessAt, int64(failCount)); err != nil {
		log.Errorf("health probe update error: %v", err)
		return
	}
	if newState != item.State {
		if base := i.baseURL(ctx, item); base != "" {
			if err := i.syncGateway(ctx, cluster.DefaultClusterID, []*gateway.DynamicRelease{newRelease(item, base)}, newState == 1); err != nil {
				log.Errorf("health probe sync gateway error: %v", err)
			}
		}
	}
}

// probeOne does a best-effort reachability check against the model's base URL.
// Any HTTP response (including 4xx/5xx) counts as reachable/healthy; only
// transport errors (timeout, connection refused, DNS failure) count as unhealthy.
func (i *imlBalanceModule) probeOne(ctx context.Context, item *ai_balance.Balance) bool {
	base := i.baseURL(ctx, item)
	if base == "" {
		return false
	}
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base, nil)
	if err != nil {
		return false
	}
	if item.Provider != ai_provider_local.ProviderLocal {
		if key, kerr := i.aiKeyService.DefaultKey(ctx, item.Provider); kerr == nil && key != nil {
			if ak := extractAPIKey(key.Config); ak != "" {
				req.Header.Set("Authorization", "Bearer "+ak)
			}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return true
}

// extractAPIKey best-effort extracts an API key from a key config JSON blob.
func extractAPIKey(config string) string {
	if config == "" {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(config), &m); err != nil {
		return ""
	}
	for _, k := range []string{"api_key", "apiKey", "key", "token", "access_key"} {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
