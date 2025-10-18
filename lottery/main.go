package main

import (
	"context"
	"log/slog"
	"os"

	pb "lottery/proto/lottery"

	cfgpkg "github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/db"

	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"

	m "lottery/domain/model"
	"lottery/domain/repository"
	"lottery/domain/service"
	"lottery/handler"
	cachepkg "lottery/infrastructure/redis_cache"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "lottery/config.example.yaml"
	}
	cfg, err := cfgpkg.Load(cfgPath)
	if err != nil {
		slog.Error("load config", "err", err)
		panic(err)
	}

	// 初始化数据库
	gdb, err := db.NewMysqlGorm(cfg)
	if err != nil {
		panic(err)
	}

	// 自动迁移（根据模型）
	if err := gdb.AutoMigrate(&m.LotteryActivity{}, &m.LotteryAward{}); err != nil {
		slog.Error("auto migrate", "err", err)
		panic(err)
	}

	// 组装仓库与服务
	actRepo := repository.NewActivityRepository(gdb)
	awRepo := repository.NewAwardRepository(gdb)
	// 初始化 Redis 客户端与缓存
	rdb := cachepkg.NewRedisClient(cfg)
	cache := cachepkg.NewRedisCache(rdb)
	svc := service.NewLotteryService(actRepo, awRepo, cache)
	svc.InitActivityCache(context.Background(), 20)

	// 微服务
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		consulAddr = "consul:8500"
	}
	consulReg := consul.NewConsulRegistry(registry.Addrs(consulAddr))
	srv := micro.NewService(
		micro.Name(cfg.Server.ServiceName),
		micro.Registry(consulReg),
	)
	srv.Init()

	// 注册handler
	if err := pb.RegisterLotteryHandler(srv.Server(), handler.New(svc)); err != nil {
		panic(err)
	}

	// 启动
	if err := srv.Run(); err != nil {
		panic(err)
	}
	_ = context.Background()
}
