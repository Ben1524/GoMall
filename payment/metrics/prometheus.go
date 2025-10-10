package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/server"
)

var (
	registerOnce sync.Once

	serverRequestTotal    *prometheus.CounterVec   // 单增计数器
	serverRequestDuration *prometheus.HistogramVec // 直方图,可以以用来统计请求延迟等

	clientRequestTotal    *prometheus.CounterVec
	clientRequestDuration *prometheus.HistogramVec
)

// Prometheus 负责暴露 Prometheus 相关能力（HTTP 服务 + 指标包装器）。
type Prometheus struct {
	serviceName string
	enabled     bool
}

// New 创建 Prometheus 集成实例。
func New(serviceName string, enabled bool) *Prometheus {
	if enabled {
		initCollectors()
	}
	return &Prometheus{serviceName: serviceName, enabled: enabled}
}

// 注册指标
func initCollectors() {
	registerOnce.Do(func() {
		serverRequestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gomall",
				Subsystem: "payment",
				Name:      "rpc_server_requests_total",
				Help:      "Total number of RPC requests handled by the service.",
			},
			[]string{"service", "endpoint", "code"},
		)

		serverRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gomall",
				Subsystem: "payment",
				Name:      "rpc_server_request_duration_seconds",
				Help:      "RPC handler latency in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"service", "endpoint"},
		)

		clientRequestTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "gomall",
				Subsystem: "payment",
				Name:      "rpc_client_requests_total",
				Help:      "Total number of outgoing RPC requests initiated by the service.",
			},
			[]string{"caller", "target", "endpoint", "code"}, // caller 是本服务名, target 是被调用服务名
		)

		clientRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "gomall",
				Subsystem: "payment",
				Name:      "rpc_client_request_duration_seconds",
				Help:      "Outgoing RPC latency in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"caller", "target", "endpoint"},
		)

		prometheus.MustRegister(
			serverRequestTotal,
			serverRequestDuration,
			clientRequestTotal,
			clientRequestDuration,
		)
	})
}

// StartHTTPServer 启动 Prometheus metrics HTTP 服务。
func (p *Prometheus) StartHTTPServer(host, port, path string) {
	if !p.enabled {
		return
	}

	// 确保 path 以 / 开头
	if path == "" {
		path = "/metrics"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle(path, promhttp.Handler())

		addr := fmt.Sprintf("%s:%s", host, port)
		slog.Info("Prometheus metrics server started", "addr", addr, "path", path)

		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			slog.Error("Prometheus metrics server stopped unexpectedly", "error", err)
		}
	}()
}

// ServerWrapper 返回 server.HandlerWrapper，对进入本服务的 RPC 请求收集指标。
func (p *Prometheus) ServerWrapper() server.HandlerWrapper {
	if !p.enabled {
		return func(h server.HandlerFunc) server.HandlerFunc {
			return h
		}
	}

	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			start := time.Now()
			err := h(ctx, req, rsp) // 调用实际的处理器

			// 记录指标

			endpoint := sanitizeEndpoint(req.Endpoint())
			code := "success"
			if err != nil {
				code = "error"
			}

			serverRequestTotal.WithLabelValues(p.serviceName, endpoint, code).Inc()
			serverRequestDuration.WithLabelValues(p.serviceName, endpoint).Observe(time.Since(start).Seconds())

			return err
		}
	}
}

// ClientWrapper 返回 client.Wrapper，对本服务发起的外部 RPC 请求收集指标。
func (p *Prometheus) ClientWrapper() client.Wrapper {
	if !p.enabled {
		return func(c client.Client) client.Client { return c }
	}

	return func(c client.Client) client.Client {
		return &promClient{
			Client:      c,
			serviceName: p.serviceName,
		}
	}
}

type promClient struct {
	client.Client
	serviceName string
}

func (p *promClient) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	start := time.Now()
	err := p.Client.Call(ctx, req, rsp, opts...)

	target := req.Service()
	endpoint := sanitizeEndpoint(req.Endpoint())
	code := "success"
	if err != nil {
		code = "error"
	}

	clientRequestTotal.WithLabelValues(p.serviceName, target, endpoint, code).Inc()
	clientRequestDuration.WithLabelValues(p.serviceName, target, endpoint).Observe(time.Since(start).Seconds())

	return err
}

func sanitizeEndpoint(endpoint string) string {
	if endpoint == "" {
		return "unknown"
	}
	return strings.ReplaceAll(endpoint, " ", "_")
}
