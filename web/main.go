package main

import (
	"context"
	"log/slog"
	"web/router"

	cfgpkg "github.com/Ben1524/GoMall/common/config"

	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
)

func main() {

	// 初始化微服务网关
	// 加载配置
	cfg, err := cfgpkg.Load("D:\\GolandProjects\\GoMall\\web\\config.example.yaml")
	if err != nil {
		slog.Error("load config", "err", err)
		panic(err)
	}

	// 微服务
	consulReg := consul.NewConsulRegistry(registry.Addrs("127.0.0.1:8500"))
	srv := micro.NewService(
		micro.Name(cfg.Server.ServiceName),
		micro.Registry(consulReg),
	)

	engine, exit := router.NewRouter(context.Background(), cfg, srv)
	defer exit()
	if engine.Run(":8080"); err != nil {
		slog.Error("start http server", "err", err)
		panic(err)
	}

}
