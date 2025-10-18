package main

import (
	"context"
	"log/slog"
	"os"
	"web/router"

	cfgpkg "github.com/Ben1524/GoMall/common/config"

	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
)

func main() {
	// load config from CONFIG_PATH or default
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "web/config.example.yaml"
	}
	cfg, err := cfgpkg.Load(cfgPath)
	if err != nil {
		slog.Error("load config", "err", err)
		panic(err)
	}

	// consul address from env CONSUL_ADDR or default to consul:8500 (docker-compose service name)
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		consulAddr = "consul:8500"
	}
	consulReg := consul.NewConsulRegistry(registry.Addrs(consulAddr))

	// 微服务
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
