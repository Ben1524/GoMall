package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	common "github.com/Ben1524/GoMall/common/config"
	db "github.com/Ben1524/GoMall/common/db"
	"github.com/Ben1524/GoMall/common/otel"
	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/wrapper/trace/opentelemetry"

	"product/domain/repository"
	productDataService "product/domain/service"
	"product/handler"
	pb "product/proto/product"
)

func main() {
	// 加载配置文件
	config, err := common.Load("product/config.example.yaml")
	if err != nil {
		slog.Error("加载配置文件失败", "error", err)
		os.Exit(1)
	}
	slog.Info("配置文件加载成功", "path", "product/config.example.yaml")

	// 初始化上下文，支持优雅退出
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// 初始化Jaeger追踪
	tracerProvider, err := otel.InitTracerProvider(ctx, config)
	if err != nil {
		slog.Error("初始化Jaeger追踪失败", "error", err)
		os.Exit(1)
	}
	defer func() {
		if tracerProvider != nil {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				slog.Warn("关闭Jaeger追踪器失败", "error", err)
			} else {
				slog.Info("Jaeger追踪器已关闭")
			}
		}
	}()

	// 初始化Consul注册中心
	consulRegistry := consul.NewConsulRegistry(
		registry.Addrs("127.0.0.1:8500"), // 简化注册中心地址配置
	)

	// 创建微服务实例
	service := micro.NewService(
		micro.Name(config.Server.ServiceName),
		micro.Version("latest"),
		micro.Registry(consulRegistry),
		// 集成OpenTelemetry追踪中间件
		micro.WrapHandler(opentelemetry.NewHandlerWrapper()),
		micro.WrapClient(opentelemetry.NewClientWrapper()),
	)

	// 初始化MySQL连接
	mysqlDB, err := db.NewMysqlGorm(config)
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

	// 初始化领域层组件
	productRepo := repository.NewProductRepository(mysqlDB)
	if err := productRepo.InitTable(); err != nil {
		slog.Error("初始化产品表失败", "error", err)
		os.Exit(1)
	}
	productSvc := productDataService.NewProductDataService(productRepo)

	// 打印服务配置信息
	slog.Info("服务配置信息",
		"mode", config.Server.Mode,
		"host", config.Server.Host,
		"port", config.Server.Port,
	)

	// 根据运行模式初始化服务地址
	if config.Server.Mode == "dev" {
		service.Init(micro.Address(config.Server.Host + ":" + config.Server.Port))
	} else {
		service.Init()
	}
	slog.Info("服务初始化完成")

	// 注册处理器
	if err := pb.RegisterProductHandler(service.Server(), handler.NewProductHandler(productSvc)); err != nil {
		slog.Error("注册产品处理器失败", "error", err)
		os.Exit(1)
	}

	// 启动服务
	slog.Info("开始启动产品服务...")
	if err := service.Run(); err != nil {
		slog.Error("产品服务运行失败", "error", err)
		os.Exit(1)
	}
	slog.Info("产品服务已正常退出")
}
