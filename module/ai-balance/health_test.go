package ai_balance

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ai_provider_local "github.com/RiemannChow/nantianmen-ai-gateway/ai-provider/local"

	ai_balance "github.com/RiemannChow/nantianmen-ai-gateway/service/ai-balance"

	"github.com/RiemannChow/nantianmen-ai-gateway/service/setting"
)

// stubSetting is a minimal ISettingService implementation for tests.
type stubSetting struct {
	val string
	has bool
}

func (s *stubSetting) Get(ctx context.Context, name string) (string, bool) { return s.val, s.has }
func (s *stubSetting) Set(ctx context.Context, name, value, operator string) error {
	return nil
}
func (s *stubSetting) All(ctx context.Context) (map[string]string, error) { return nil, nil }

func TestProbeOneReachable(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer up.Close()

	m := &imlBalanceModule{settingService: &stubSetting{val: up.URL, has: true}}
	item := &ai_balance.Balance{Provider: ai_provider_local.ProviderLocal, Type: 1}
	if !m.probeOne(context.Background(), item) {
		t.Fatalf("expected healthy for reachable server %s", up.URL)
	}
}

func TestProbeOneUnreachable(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	downURL := down.URL
	down.Close() // connection refused from now on

	m := &imlBalanceModule{settingService: &stubSetting{val: downURL, has: true}}
	item := &ai_balance.Balance{Provider: ai_provider_local.ProviderLocal, Type: 1}
	if m.probeOne(context.Background(), item) {
		t.Fatalf("expected unhealthy for unreachable server %s", downURL)
	}
}

func TestProbeOneNoAddress(t *testing.T) {
	// no ollama address configured -> nothing to probe -> unhealthy
	m := &imlBalanceModule{settingService: &stubSetting{val: "", has: false}}
	item := &ai_balance.Balance{Provider: ai_provider_local.ProviderLocal, Type: 1}
	if m.probeOne(context.Background(), item) {
		t.Fatal("expected unhealthy when base URL is unresolvable")
	}
}

func TestExtractAPIKey(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"api_key field", `{"api_key":"sk-abc"}`, "sk-abc"},
		{"key field", `{"key":"secret"}`, "secret"},
		{"nested none", `{"x":1}`, ""},
		{"empty", ``, ""},
		{"malformed", `{not json`, ""},
	}
	for _, c := range cases {
		if got := extractAPIKey(c.in); got != c.want {
			t.Fatalf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

var _ setting.ISettingService = (*stubSetting)(nil)
