// go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"order/domain/repository"
	srv "order/domain/service"
	"order/handler"
	"order/metrics"
	"os"
	"os/signal"
	"syscall"

	config "github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/db"
	"github.com/Ben1524/GoMall/common/otel"
	"go-micro.dev/v5"
	"go-micro.dev/v5/client"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/server"
	"go-micro.dev/v5/wrapper/trace/opentelemetry"
	ratelimit3 "go.uber.org/ratelimit"
	"golang.org/x/time/rate"

	pb "order/proto/order"

	// 限流器（Uber 令牌桶）
	ratelimit "github.com/micro/plugins/v5/wrapper/ratelimiter/uber"
)

// QPS 限流阈值（按需调整）
const qps = 500 // QPS是指每秒的访问量

// 自定义限流器中间件（非阻塞模式，超过阈值直接拒绝）
func NewNonBlockingLimiter(qps int) server.HandlerWrapper {
	limiter := rate.NewLimiter(rate.Limit(qps), qps*2) // 桶容量=QPS
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			// 尝试获取令牌，不等待
			if !limiter.Allow() {
				// 没有令牌，直接返回错误（可自定义错误类型）
				return fmt.Errorf("请求过于频繁，请稍后再试")
			}
			// 有令牌，继续处理请求
			return h(ctx, req, rsp)
		}
	}
}

func main() {
	cfg, err := config.Load("D:\\GolandProjects\\GoMall\\order\\config.example.yaml")
	if err != nil {
		slog.Error("config加载失败", "error", err)
		panic(err)
	}
	slog.Info("config加载成功", "path", "order/config.example.yaml")

	promMetrics := metrics.New(cfg.Server.ServiceName, cfg.Metrics.Enabled)
	if cfg.Metrics.Enabled {
		promMetrics.StartHTTPServer(cfg.Metrics.Host, cfg.Metrics.Port, cfg.Metrics.Path)
	}

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

	// MySQL
	mysqlDB, err := db.NewMysqlGorm(cfg)
	if err != nil {
		slog.Error("初始化MySQL连接失败", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := mysqlDB.Close(); err != nil {
			slog.Warn("关闭MySQL连接失败", "error", err)
		} else {
			slog.Info("MySQL连接已关闭")
		}
	}()

	orderRepository := repository.NewOrderRepository(mysqlDB)

	if err := orderRepository.InitTable(); err != nil {
		slog.Error("init table error")
		panic(err)
	}

	orderService := srv.NewOrderDataService(orderRepository)

	consulRegistry := consul.NewConsulRegistry(registry.Addrs("127.0.0.1:8500"))

	slog.Info(cfg.Metrics.Host + ":" + cfg.Metrics.Port)

	consulRegistry.Register(&registry.Service{
		Name: cfg.Server.ServiceName,
		Nodes: []*registry.Node{
			{
				Metadata: map[string]string{
					"prometheus-monitor": "true",     // 必须和Prometheus的过滤规则匹配
					"service-type":       "go-micro", // 可选，自定义标签
				},
				Address: cfg.Metrics.Host + ":" + cfg.Metrics.Port,
			}},
	})

	handlerWrappers := []server.HandlerWrapper{
		ratelimit.NewHandlerWrapper(qps, ratelimit3.WithSlack(3*qps)),
		opentelemetry.NewHandlerWrapper(),
	}
	if cfg.Metrics.Enabled {
		handlerWrappers = append([]server.HandlerWrapper{promMetrics.ServerWrapper()}, handlerWrappers...)
	}

	clientWrappers := []client.Wrapper{
		ratelimit.NewClientWrapper(qps, ratelimit3.WithSlack(3*qps)),
		opentelemetry.NewClientWrapper(),
	}
	if cfg.Metrics.Enabled {
		// 注意promMetrics.ClientWrapper()必须放在最前面,作为最外层的Wrapper
		clientWrappers = append([]client.Wrapper{promMetrics.ClientWrapper()}, clientWrappers...)
	}

	serviceMetadata := map[string]string{
		"service-type":       "go-micro",
		"prometheus-monitor": fmt.Sprintf("%t", cfg.Metrics.Enabled),
		"address":            fmt.Sprintf("%s:%s", cfg.Metrics.Host, cfg.Metrics.Port),
	}
	if cfg.Metrics.Enabled {
		serviceMetadata["metrics_host"] = cfg.Metrics.Host
		serviceMetadata["metrics_port"] = cfg.Metrics.Port
		serviceMetadata["metrics_path"] = cfg.Metrics.Path
	}

	serviceOptions := []micro.Option{
		micro.Name(cfg.Server.ServiceName),
		micro.Version("latest"),
		micro.Registry(consulRegistry),
		micro.Metadata(serviceMetadata),
		micro.WrapHandler(handlerWrappers...),
		micro.WrapClient(clientWrappers...),
	}

	service := micro.NewService(serviceOptions...)
	service.Init()
	if err := pb.RegisterOrderHandler(service.Server(), handler.NewOrderHandler(orderService)); err != nil {
		slog.Error("注册Cart处理器失败", "error", err)
		os.Exit(1)
	}

	if err := service.Run(); err != nil {
		slog.Error("服务运行失败", "error", err)
	}
}
