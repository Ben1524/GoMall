package otel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ben1524/GoMall/common/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

// 初始化OpenTelemetry TracerProvider，配置Jaeger作为后端
func InitTracerProviderWithName(ctx context.Context, cfg *config.Config, name string) (*sdktrace.TracerProvider, error) {
	jaegerCfg := cfg.Jaeger
	if jaegerCfg.Host == "" || jaegerCfg.Port == "" {
		return nil, fmt.Errorf("jaeger configuration is incomplete")
	}

	slog.Info(fmt.Sprintf("%s:%s", jaegerCfg.Host, jaegerCfg.Port))
	// 配置OTLP导出器选项
	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", jaegerCfg.Host, jaegerCfg.Port)),
	}

	// 根据配置决定是否使用TLS
	if !jaegerCfg.UseTLS {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	// 创建OTLP导出器
	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 定义服务资源（包含默认资源+自定义属性）
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
			semconv.DeploymentName(jaegerCfg.Deployment),
			attribute.String("team", "backend"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 根据环境选择采样器
	var sampler sdktrace.Sampler
	switch jaegerCfg.Deployment {
	case "production":
		// 生产环境使用概率采样，例如10%
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))
	default:
		// 开发/测试环境全采样
		sampler = sdktrace.AlwaysSample()
	}

	// 创建TracerProvider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局组件
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

// 初始化OpenTelemetry TracerProvider，配置Jaeger作为后端
func InitTracerProvider(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	jaegerCfg := cfg.Jaeger
	if jaegerCfg.Host == "" || jaegerCfg.Port == "" {
		return nil, fmt.Errorf("jaeger configuration is incomplete")
	}

	slog.Info(fmt.Sprintf("%s:%s", jaegerCfg.Host, jaegerCfg.Port))
	// 配置OTLP导出器选项
	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%s", jaegerCfg.Host, jaegerCfg.Port)),
	}

	// 根据配置决定是否使用TLS
	if !jaegerCfg.UseTLS {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	// 创建OTLP导出器
	exporter, err := otlptracegrpc.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// 定义服务资源（包含默认资源+自定义属性）
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.Server.ServiceName),
			semconv.DeploymentName(jaegerCfg.Deployment),
			attribute.String("team", "backend"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 根据环境选择采样器
	var sampler sdktrace.Sampler
	switch jaegerCfg.Deployment {
	case "production":
		// 生产环境使用概率采样，例如10%
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))
	default:
		// 开发/测试环境全采样
		sampler = sdktrace.AlwaysSample()
	}

	// 创建TracerProvider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局组件
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

// 优雅关闭TracerProvider，确保追踪数据被正确导出
func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp != nil {
		return tp.Shutdown(ctx)
	}
	return nil
}
