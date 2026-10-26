package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestBucketPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"/", "/"},
		{"", "/"},
		{"/api/v1/service/abc-123/api/xyz", "/api/v1/service"},
		{"/api/v1/ai-api", "/api/v1/ai-api"},
		{"/api", "/api"},
		{"/assets/index-abc123.js", "/assets"},
		{"/service/detail", "/service"},
		{"/login", "/login"},
	}
	for _, c := range cases {
		if got := bucketPath(c.in); got != c.want {
			t.Errorf("bucketPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestInstrumentRecordsStatusAndPath(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/boom" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	h := Instrument(next)

	before := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/api/v1/ok", "200"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ok", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	after := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/api/v1/ok", "200"))
	if after != before+1 {
		t.Errorf("requests_total delta = %v, want 1", after-before)
	}

	// 500 状态码应被捕获
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/boom", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := testutil.ToFloat64(requestsTotal.WithLabelValues("GET", "/api/v1/boom", "500")); got < 1 {
		t.Error("500 request was not counted")
	}
}

func TestInstrumentPassesThroughFlusher(t *testing.T) {
	// 下游若支持 Flush，包装器必须透传，否则 SSE 流式会被破坏
	var flushed bool
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
			flushed = true
		}
	})
	h := Instrument(next)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/v1/stream", nil))
	if !flushed {
		t.Error("Flush was not passed through the instrumented handler")
	}
}

func TestAdminHandler(t *testing.T) {
	srv := httptest.NewServer(AdminHandler())
	defer srv.Close()

	for _, path := range []string{"/metrics", "/healthz", "/debug/pprof/"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", path, resp.StatusCode)
		}
	}

	// /metrics 应包含自定义指标与默认 Go 运行时指标
	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	body := make([]byte, 64*1024)
	n, _ := resp.Body.Read(body)
	got := string(body[:n])
	for _, want := range []string{"nantianmen_http_requests_total", "go_goroutines"} {
		if !containsStr(got, want) {
			t.Errorf("/metrics output missing %q", want)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
