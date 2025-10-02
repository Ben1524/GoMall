package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	common "github.com/Ben1524/GoMall/common/config"
	"github.com/Ben1524/GoMall/common/otel"
	"go-micro.dev/v5"
	"go-micro.dev/v5/registry"
	"go-micro.dev/v5/registry/consul"
	"go-micro.dev/v5/wrapper/trace/opentelemetry"

	pb "product/proto/product"
)

// 全局客户端变量，供测试函数使用
var productClient pb.ProductService

// TestMain 用于初始化测试环境
func main() {

	// 加载配置文件
	config, err := common.Load("D:\\GolandProjects\\GoMall\\product\\config.example.yaml")
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
	tracerProvider, err := otel.InitTracerProviderWithName(ctx, config, "go.micro.service.product.client")
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
		registry.Addrs("127.0.0.1:8500"),
	)

	// 创建微服务实例
	service := micro.NewService(
		micro.Name("go.micro.service.product.client"),
		micro.Version("latest"),
		micro.Registry(consulRegistry),
		micro.WrapHandler(opentelemetry.NewHandlerWrapper()),
		micro.WrapClient(opentelemetry.NewClientWrapper()),
	)

	service.Init()

	// 初始化客户端并赋值给全局变量
	productClient = pb.NewProductService("go.micro.service.product", service.Client())

	// 运行实际的测试函数
	AddProduct()

}

// 实际测试函数：测试AddProduct接口
func AddProduct() {
	productAdd := &pb.ProductInfo{
		ProductName:        "imooc2",
		ProductSku:         "cap2",
		ProductPrice:       1.1,
		ProductDescription: "imooc-cap2",
		ProductCategoryId:  2,
		ProductImage: []*pb.ProductImage{
			{
				ImageName: "cap-image3",
				ImageCode: "capimage03",
				ImageUrl:  "capimage03",
			},
		},
		ProductSize: []*pb.ProductSize{
			{
				SizeName: "cap-size2",
				SizeCode: "cap-size-code2",
			},
		},
		ProductSeo: &pb.ProductSeo{
			SeoTitle:       "cap-seo",
			SeoKeywords:    "cap-seo",
			SeoDescription: "cap-seo",
			SeoCode:        "cap-seo",
		},
	}

	// 调用服务
	response, err := productClient.AddProduct(context.Background(), productAdd)
	if err != nil {
		// 使用t.Fatal报告错误（会终止当前测试函数）
		panic(err)
	}

	// 可以添加断言逻辑（如检查response是否符合预期）
	slog.Info("AddProduct success, response: %s", response.String())
}
