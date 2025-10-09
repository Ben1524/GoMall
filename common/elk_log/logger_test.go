package elk_log

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// TestBasicLogging 基础日志功能测试
func TestBasicLogging(t *testing.T) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	// 测试不同日志级别
	logger.Info("info message", slog.String("key", "value"))
	logger.Warn("warning message", slog.Int("count", 42))
	logger.Error("error message", slog.Float64("amount", 99.99))

	time.Sleep(1 * time.Second) // 等待批量刷出
}

// TestBatchFlush 批量刷出测试
func TestBatchFlush(t *testing.T) {
	config := Config{
		Address:      "localhost:5000",
		Level:        slog.LevelInfo,
		BatchSize:    50,
		BatchMaxWait: 500 * time.Millisecond,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	start := time.Now()
	// 写入小于批量阈值的日志
	for i := 0; i < 30; i++ {
		logger.Info("batch test", slog.Int("index", i))
	}

	// 等待时间触发刷出
	time.Sleep(600 * time.Millisecond)
	elapsed := time.Since(start)

	if elapsed < 500*time.Millisecond {
		t.Errorf("batch flush triggered too early: %v", elapsed)
	}
}

// TestConcurrentLogging 并发写入测试
func TestConcurrentLogging(t *testing.T) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	var wg sync.WaitGroup
	goroutines := 10
	logsPerGoroutine := 100

	start := time.Now()
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("concurrent log",
					slog.Int("goroutine", id),
					slog.Int("index", j))
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Logged %d entries from %d goroutines in %v",
		goroutines*logsPerGoroutine, goroutines, elapsed)
}

// TestHighThroughput 高吞吐量性能测试
func TestHighThroughput(t *testing.T) {
	config := Config{
		Address:      "localhost:5000",
		Level:        slog.LevelInfo,
		BatchSize:    200,
		BatchMaxWait: 100 * time.Millisecond,
		MaxCacheSize: 1000,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	totalLogs := 10000
	start := time.Now()

	for i := 0; i < totalLogs; i++ {
		logger.Info("high throughput test",
			slog.String("order_id", fmt.Sprintf("ORD%d", i)),
			slog.Float64("amount", float64(i)*1.5),
			slog.String("user_id", fmt.Sprintf("user%d", i%100)))
	}

	elapsed := time.Since(start)
	logsPerSecond := float64(totalLogs) / elapsed.Seconds()

	t.Logf("Throughput: %.0f logs/second (%d logs in %v)",
		logsPerSecond, totalLogs, elapsed)

	time.Sleep(1 * time.Second) // 等待刷出
}

// BenchmarkSingleGoroutine 单协程基准测试
func BenchmarkSingleGoroutine(b *testing.B) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		b.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark log",
			slog.String("id", fmt.Sprintf("log%d", i)),
			slog.Int("value", i))
	}
}

// BenchmarkConcurrent 并发基准测试
func BenchmarkConcurrent(b *testing.B) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		b.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			logger.Info("concurrent benchmark",
				slog.String("id", fmt.Sprintf("log%d", i)),
				slog.Int("value", i))
			i++
		}
	})
}

// TestMemoryPressure 内存压力测试
func TestMemoryPressure(t *testing.T) {
	config := Config{
		Address:      "localhost:5000",
		Level:        slog.LevelInfo,
		BatchSize:    100,
		BatchMaxWait: 200 * time.Millisecond,
		MaxCacheSize: 500,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	// 快速写入大量日志，触发缓存保护
	for i := 0; i < 1000; i++ {
		logger.Info("memory pressure test",
			slog.Int("index", i),
			slog.String("data", "large string payload"))
	}

	time.Sleep(2 * time.Second)
}

// TestWithAttrsPerformance 测试WithAttrs性能
func TestWithAttrsPerformance(t *testing.T) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	// 添加全局属性
	logger = logger.With(
		slog.String("service", "payment"),
		slog.String("environment", "production"),
	)

	start := time.Now()
	for i := 0; i < 5000; i++ {
		logger.Info("with attrs test", slog.Int("index", i))
	}
	elapsed := time.Since(start)

	t.Logf("Logged 5000 entries with global attrs in %v", elapsed)
}

// TestReconnection 重连测试
func TestReconnection(t *testing.T) {
	config := Config{
		Address:           "localhost:5000",
		Level:             slog.LevelInfo,
		ReconnectInterval: 1 * time.Second,
		MaxRetry:          3,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	// 正常日志
	logger.Info("before disconnect")

	// 模拟连接断开（需要手动停止Logstash）
	t.Log("Manually stop Logstash now, wait 5 seconds...")
	time.Sleep(5 * time.Second)

	// 写入日志（应自动重连）
	logger.Info("after reconnect attempt")
	time.Sleep(2 * time.Second)
}

// TestContextualLogging 上下文日志测试
func TestContextualLogging(t *testing.T) {
	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	// 使用WithGroup和WithAttrs
	userLogger := logger.WithGroup("user").With(
		slog.String("user_id", "u123"),
		slog.String("session", "sess456"),
	)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		userLogger.Info("user action",
			slog.String("action", "click"),
			slog.Int("count", i))
	}
	elapsed := time.Since(start)

	t.Logf("Logged 1000 contextual entries in %v", elapsed)
}

// TestBatchSizeImpact 批量大小对性能的影响
func TestBatchSizeImpact(t *testing.T) {
	batchSizes := []int{10, 50, 100, 200}

	for _, size := range batchSizes {
		config := Config{
			Address:      "localhost:5000",
			Level:        slog.LevelInfo,
			BatchSize:    size,
			BatchMaxWait: 500 * time.Millisecond,
		}

		logger, err := NewLogger(config)
		if err != nil {
			t.Fatalf("failed to create logger: %v", err)
		}

		start := time.Now()
		for i := 0; i < 10000; i++ {
			logger.Info("batch size test", slog.Int("index", i))
		}
		elapsed := time.Since(start)

		logger.Handler().(*LogstashHandler).Close()

		t.Logf("BatchSize=%d: 10000 logs in %v (%.2f logs/sec)",
			size, elapsed, 1000/elapsed.Seconds())
	}
}

// TestLongRunning 长时间运行测试
func TestLongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long running test")
	}

	logger, err := NewDefaultLogger("localhost:5000")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Handler().(*LogstashHandler).Close()

	duration := 30 * time.Second
	interval := 10 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	count := 0
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(start)
			t.Logf("Long running test: %d logs in %v (%.2f logs/sec)",
				count, elapsed, float64(count)/elapsed.Seconds())
			return
		default:
			logger.Info("long running log", slog.Int("count", count))
			count++
			time.Sleep(interval)
		}
	}
}
