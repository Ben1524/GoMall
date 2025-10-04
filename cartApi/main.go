package main

import (
	"cartApi/handler"
	"cartApi/proto/cart"
	"cartApi/router"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/otel"
	gobreaker2 "github.com/micro/plugins/v5/wrapper/breaker/gobreaker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
	"go-micro.dev/v5"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/wrapper/trace/opentelemetry"
)

// 定义Prometheus指标
var (
	// 熔断器状态：0=关闭, 1=打开, 2=半打开
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "当前熔断器状态 (0=closed, 1=open, 2=half-open)",
		},
		[]string{"breaker_name"}, // 按熔断器名称标签区分
	)

	// 总请求数
	circuitBreakerTotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_total_requests",
			Help: "熔断器处理的总请求数",
		},
		[]string{"breaker_name"},
	)

	// 失败请求数
	circuitBreakerFailedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failed_requests",
			Help: "熔断器处理的失败请求数",
		},
		[]string{"breaker_name"},
	)

	// 状态切换次数
	circuitBreakerStateChanges = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_state_changes",
			Help: "熔断器状态切换总次数",
		},
		[]string{"breaker_name", "from_state", "to_state"},
	)
)

func init() {
	// 注册指标到Prometheus默认注册表
	prometheus.MustRegister(
		circuitBreakerState,
		circuitBreakerTotalRequests,
		circuitBreakerFailedRequests,
		circuitBreakerStateChanges,
	)
}

func startMetricsServer(host, port string) {

	// 启动Prometheus metrics HTTP服务
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		url := host + ":" + port
		slog.Info("Prometheus metrics server started at " + url)
		if err := http.ListenAndServe(url, nil); err != nil {
			slog.Error("Failed to start metrics server", "error", err)
		}
	}()
}

// 将gobreaker.State转换为数值（用于指标存储）
func stateToValue(state gobreaker.State) float64 {
	switch state {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateOpen:
		return 1
	case gobreaker.StateHalfOpen:
		return 2
	default:
		return -1
	}
}

func NewCircuitBreaker(name string) client.Wrapper {
	bs := gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// 记录状态切换指标
			circuitBreakerStateChanges.WithLabelValues(
				name,
				from.String(),
				to.String(),
			).Inc()

			// 更新当前状态指标
			circuitBreakerState.WithLabelValues(name).Set(stateToValue(to))

			slog.Info("熔断器状态变化",
				"name", name,
				"from", from,
				"to", to)
		},
	}
	return gobreaker2.NewCustomClientWrapper(bs, gobreaker2.BreakService)

}

func main() {
	breakerName := "cartApi"

	cfg, err := config.Load("config.example.yaml")
	if err != nil {
		slog.Error("config加载失败", "error", err)
		panic(err)
	}
	slog.Info("config加载成功", "path", "cart/config.example.yaml")

	startMetricsServer(cfg.Metrics.Host, cfg.Metrics.Port)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// OTel/Jaeger
	tp, err := otel.InitTracerProvider(ctx, cfg)
	if err != nil {
		slog.Error("初始化Jaeger追踪失败", "error", err)
		os.Exit(1)
	}
	defer func() {
		if tp != nil {
			if err := tp.Shutdown(ctx); err != nil {
				slog.Warn("关闭Jaeger追踪器失败", "error", err)
			} else {
				slog.Info("Jaeger追踪器已关闭")
			}
		}
	}()

	// 创建服务
	consulRegistry := consul.NewConsulRegistry(registry.Addrs("127.0.0.1:8500"))
	service := micro.NewService(
		micro.Name(cfg.Server.ServiceName),
		micro.WrapClient(NewCircuitBreaker(breakerName)),   // 自定义熔断器中间件
		micro.WrapClient(opentelemetry.NewClientWrapper()), // 只作为调用方，不会触发被调用逻辑
		micro.Registry(consulRegistry),
	)
	service.Init()

	cartSrv := cart.NewCartService("go.micro.service.cart", service.Client())

	h := handler.NewCartApiHandler(cartSrv)
	engine := router.New(cfg, h)

	if engine == nil {
		slog.Error("router初始化失败", "error", err)
		panic(errors.New("router初始化失败"))
	}
	if err := engine.Run(cfg.Server.Host + ":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
		slog.Error("启动HTTP服务失败", "error", err)
		panic(err)
	}

	slog.Info("HTTP服务已启动", "host", cfg.Server.Host, "port", cfg.Server.Port)

	<-ctx.Done()
	stop()
	slog.Info("正在关闭HTTP服务...")

	// service.Run()
}
