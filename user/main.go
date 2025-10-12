package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"user/domain/repository"
	"user/domain/service"
	userHandler "user/handler"
	pb "user/proto/user"

	common "github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/db"
	"github.com/Ben1524/GoMall/common/otel"
	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/wrapper/trace/opentelemetry"
)

func main() {
	// 加载配置文件
	config, err := common.Load("D:\\GolandProjects\\GoMall\\user\\config.example.yaml")
	if err != nil {
		slog.Error("加载配置文件失败", "error", err)
		os.Exit(1)
	}
	slog.Info("配置文件加载成功", "path", "user/config.example.yaml")

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
	svc := micro.NewService(
		micro.Name(config.Server.ServiceName),
		micro.Version("latest"),
		micro.Registry(consulRegistry),
		// 集成OpenTelemetry追踪中间件
		micro.WrapHandler(opentelemetry.NewHandlerWrapper()),
		micro.WrapClient(opentelemetry.NewClientWrapper()),
	)
	svc.Init()

	// 初始化MySQL连接
	mysqlDB, err := db.NewMysqlGorm(config)
	if err != nil {
		slog.Error("初始化MySQL连接失败", "error", err)
		os.Exit(1)
	}
	// 注意：gorm v2 返回 *gorm.DB，底层连接池关闭需获取 sqlDB
	// 这里不强制关闭，交由进程退出时回收，或可在此获取 sqlDB 并关闭。

	// 初始化领域层组件
	userRepo := repository.NewUserRepository(mysqlDB)
	if err := userRepo.InitTable(); err != nil {
		slog.Error("初始化用户表失败", "error", err)
		os.Exit(1)
	}

	// 初始化服务层与 handler 注册
	userSvc := service.NewUserService(userRepo)
	if err := pb.RegisterUserHandler(svc.Server(), userHandler.NewUserHandler(userSvc)); err != nil {
		slog.Error("注册用户服务处理器失败", "error", err)
		os.Exit(1)
	}

	// 启动服务
	if err := svc.Run(); err != nil {
		slog.Error("用户服务运行失败", "error", err)
		os.Exit(1)
	}

}
