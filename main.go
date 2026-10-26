package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/metrics"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/cftool"
	"github.com/eolinker/go-common/permit"
	"github.com/eolinker/go-common/server"
)

var (
	version  string
	confPath string
)

func init() {
	flag.StringVar(&confPath, "c", "config.yml", "`config` file path for server ")
}

type ServerConfig struct {
	Port    int `yaml:"port"`
	Metrics struct {
		// admin 端点地址（/metrics、/debug/pprof、/healthz）。
		// 默认仅监听 127.0.0.1:9464；需要远程采集时显式配置（注意访问控制）。
		// 设为 "off" 可完全关闭。
		Addr string `yaml:"addr"`
	} `yaml:"metrics"`
}

// Server hardening parameters.
//
// ReadHeaderTimeout/IdleTimeout protect against slowloris-style attacks that
// hold connections open without sending a complete request.
//
// WriteTimeout is deliberately left unset: the gateway proxies LLM traffic
// including long-lived streaming (SSE) responses whose total duration cannot
// be bounded up front. Per-route proxy timeouts are configured on the
// router/proxy layer instead.
const (
	readHeaderTimeout = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 30 * time.Second

	// defaultMetricsAddr 是 admin 端点默认地址：仅本机，避免指标与 pprof
	// 随主端口暴露。
	defaultMetricsAddr = "127.0.0.1:9464"
	// metricsDisabled 是关闭 admin 端点的配置值。
	metricsDisabled = "off"
)

func main() {
	doCheck()

	flag.Parse()

	cf := new(ServerConfig)
	cftool.Register(fmt.Sprintf("root:%s", confPath), cf)
	cftool.ReadFile(confPath)

	ser := server.CreateServer()

	err := autowire.CheckComplete()
	if err != nil {
		log.Fatal("check autowired:", err)
		return
	}

	if cf.Port == 0 {
		log.Fatal("need port")
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cf.Port))
	if err != nil {
		log.Fatal(err)
		return
	}
	handler := ser.Build()
	for access, paths := range handler.Permits() {
		permit.AddPermitRule(access, paths...)
	}

	// Root context is cancelled on SIGINT/SIGTERM so the server can drain
	// in-flight requests (including long LLM streams) instead of being
	// killed mid-request by the process manager.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// admin 端点（/metrics + pprof + healthz），独立端口，默认仅本机监听
	var adminSrv *http.Server
	metricsAddr := cf.Metrics.Addr
	if metricsAddr == "" {
		metricsAddr = defaultMetricsAddr
	}
	if metricsAddr != metricsDisabled {
		adminSrv = metrics.ServeAdmin(metricsAddr)
		go func() {
			if err := adminSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Errorf("admin server error: %v", err)
			}
		}()
	}

	srv := &http.Server{
		// 请求指标采集包装（透传 Hijack/Flush，不影响 SSE 与 WebSocket）
		Handler:           metrics.Instrument(handler),
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
			// Unblock the main goroutine; a graceful drain is no longer possible.
			stop()
		}
	}()

	v := version
	if v == "" {
		v = "dev"
	}
	if adminSrv != nil {
		log.Infof("nantianmen-ai-gateway %s listening on :%d (metrics on %s)", v, cf.Port, metricsAddr)
	} else {
		log.Infof("nantianmen-ai-gateway %s listening on :%d (metrics disabled)", v, cf.Port)
	}

	<-ctx.Done()
	log.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Errorf("graceful shutdown failed, forcing close: %v", err)
		_ = srv.Close()
	}
	if adminSrv != nil {
		if err := adminSrv.Shutdown(shutdownCtx); err != nil {
			log.Errorf("admin server shutdown failed: %v", err)
			_ = adminSrv.Close()
		}
	}
	log.Info("server stopped")
}
