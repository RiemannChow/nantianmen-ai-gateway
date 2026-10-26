// Package metrics provides HTTP 请求指标采集与独立的管理端点（/metrics、pprof）。
//
// 设计要点：
//   - 通过包装主 handler 采集请求计数/时延，不侵入 go-common 的 gin 路由注册流程；
//   - /metrics 与 pprof 暴露在独立 admin 端口（默认仅 127.0.0.1），避免随主端口
//     对外暴露；需要远程采集时在 config.yml 的 metrics.addr 中显式配置；
//   - ResponseWriter 包装器实现了 Hijacker/Flusher，SSE 流式与 WebSocket
//     升级不会被破坏。
package metrics

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "nantianmen_http_requests_total",
		Help: "Total number of HTTP requests processed, by method/path/status.",
	}, []string{"method", "path", "status"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "nantianmen_http_request_duration_seconds",
		Help:    "HTTP request latency distribution in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	requestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nantianmen_http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed.",
	})
)

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration, requestsInFlight)
}

// bucketPath 限制 path 标签基数，防止动态路由把指标打爆：
//   - /api 开头取前三段（如 /api/v1/service）
//   - 其余取第一段（如 /assets、/service）
func bucketPath(p string) string {
	if p == "" {
		return "/"
	}
	segs := strings.Split(strings.Trim(p, "/"), "/")
	if len(segs) > 0 && segs[0] == "api" {
		if len(segs) > 3 {
			segs = segs[:3]
		}
	} else if len(segs) > 1 {
		segs = segs[:1]
	}
	return "/" + strings.Join(segs, "/")
}

// statusRecorder 包装 ResponseWriter 以捕获状态码，同时透传
// Hijack/Flush 等能力（SSE、WebSocket 依赖）。
type statusRecorder struct {
	http.ResponseWriter
	status  int
	written bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if !s.written {
		s.status = code
		s.written = true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.written {
		s.status = http.StatusOK
		s.written = true
	}
	return s.ResponseWriter.Write(b)
}

func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("metrics: underlying ResponseWriter does not support hijack")
}

// Instrument 包装 http.Handler，记录请求总数、时延与在途请求数。
func Instrument(next http.Handler) http.Handler {
	if next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := bucketPath(r.URL.Path)
		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start).Seconds()

		requestDuration.WithLabelValues(r.Method, path).Observe(elapsed)
		requestsTotal.WithLabelValues(r.Method, path, strconv.Itoa(rec.status)).Inc()
	})
}

// AdminHandler 返回 admin mux：/metrics、/debug/pprof/* 与 /healthz。
// 单独抽出以便测试与复用。
func AdminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

// ServeAdmin 构造 admin 服务器（/metrics + pprof + healthz）。
// 返回的 *http.Server 由调用方 goroutine 化 ListenAndServe 并纳入优雅退出，
// 以便统一记录监听错误。
func ServeAdmin(addr string) *http.Server {
	if addr == "" {
		addr = "127.0.0.1:9464"
	}
	return &http.Server{
		Addr:              addr,
		Handler:           AdminHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
}
