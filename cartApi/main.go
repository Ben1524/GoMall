package main

import (
	"cartApi/proto/cart"
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	gobreaker2 "github.com/micro/plugins/v5/wrapper/breaker/gobreaker"
	"github.com/sony/gobreaker"
	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
)

func main() {
	bs := gobreaker.Settings{
		Name:        "cartApi",        // 熔断器名称（用于标识，建议与服务名关联）
		MaxRequests: 5,                // 熔断器处于半打开状态时，允许通过的最大请求数
		Interval:    10 * time.Second, // 统计失败率的时间窗口（如10秒内的请求会被计入统计）
		Timeout:     30 * time.Second, // 熔断器从"打开"状态切换到"半打开"状态的等待时间
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 触发熔断器打开的条件：10秒内请求数≥10，且失败率≥50%
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// 状态变化时的回调（如记录日志）
			// 状态包括：StateClosed（关闭）、StateOpen（打开）、StateHalfOpen（半打开）
			slog.Info("熔断器 %s 状态变化: %s -> %s\n", name, from, to)
		},
	}
	goBreaker := gobreaker2.NewCustomClientWrapper(bs, gobreaker2.BreakService)

	// Create service
	consulRegistry := consul.NewConsulRegistry(registry.Addrs("127.0.0.1:8500"))

	service := micro.NewService(
		micro.Name("go.micro.api.cart"),
		micro.WrapClient(goBreaker), // 允许你在客户端的调用链中「插入」中间件，这些中间件会拦截客户端的请求和响应
		micro.Registry(consulRegistry),
	)
	// Initialize service
	service.Init()

	cartSrv := cart.NewCartService("go.micro.service.cart", service.Client())

	for i := 0; i <= math.MaxInt32; i++ {
		timeOutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(50)*time.Millisecond)
		defer cancel()
		_, err := cartSrv.GetAll(timeOutCtx, &cart.CartFindAll{UserId: 1})
		if errors.As(err, &gobreaker.ErrOpenState) {
			slog.Error(err.Error())
			time.Sleep(10 * time.Millisecond)
		}
		//slog.Info(rep.String())
	}

	// Run service
	service.Run()
}
