package elk_log

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	json "github.com/bytedance/sonic"
)

// Config 配置保持不变（批量参数沿用）
type Config struct {
	Address           string        // Logstash的TCP地址（必填）
	Level             slog.Level    // 日志级别（默认Info）
	ConnectTimeout    time.Duration // 连接超时（默认5s）
	WriteTimeout      time.Duration // 写入超时（默认5s）
	ReconnectInterval time.Duration // 重连间隔（默认3s）
	MaxRetry          int           // 连接重试次数（默认3次）
	BatchSize         int           // 批量触发阈值（默认100条）
	BatchMaxWait      time.Duration // 最大等待时间（默认500ms）
	MaxCacheSize      int           // 缓存队列最大容量（默认500条）
}

// 关键修复1：定义可导出的日志结构体（首字母大写），并通过json标签匹配Logstash格式
// logEntry 日志条目（字段全导出，标签指定Logstash期望的键名）
type logEntry struct {
	Timestamp string   `json:"@timestamp"`       // 匹配Logstash的@timestamp
	Level     string   `json:"level"`            // 日志级别（如INFO/ERROR）
	Message   string   `json:"message"`          // 日志内容
	Source    string   `json:"source"`           // 固定标识：应用来源
	Groups    []string `json:"groups,omitempty"` // 分组（可选，空则不序列化）
	// 业务属性：直接作为顶层字段，而非嵌套map（修复嵌套问题）
	Attrs map[string]interface{} `json:"-"` // 用`json:"-"`忽略该字段，后续手动合并到顶层
}

// LogstashHandler 修复后的批量处理器
type LogstashHandler struct {
	config       Config
	conn         net.Conn
	muLock       sync.Locker            // 保护连接读写
	level        slog.Level             // 日志级别
	globalAttrs  map[string]interface{} // 全局属性（平级合并）
	groups       []string               // 日志分组
	cache        []*logEntry            // 缓存队列（用修复后的logEntry）
	muCache      sync.Mutex             // 保护缓存读写
	workerCtx    context.Context        // Worker协程上下文
	workerCancel context.CancelFunc     // Worker取消函数
	wg           sync.WaitGroup         // 等待Worker退出
	isClosed     atomic.Bool            // 是否已关闭（原子操作）
}

// newLogstashHandler 创建修复后的处理器
func newLogstashHandler(config Config) (*LogstashHandler, error) {
	// 1. 设置默认配置
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 5 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 5 * time.Second
	}
	if config.ReconnectInterval == 0 {
		config.ReconnectInterval = 3 * time.Second
	}
	if config.MaxRetry <= 0 {
		config.MaxRetry = 3
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchMaxWait <= 0 {
		config.BatchMaxWait = 500 * time.Millisecond
	}
	if config.MaxCacheSize <= 0 {
		config.MaxCacheSize = 500
	}
	if config.Address == "" {
		return nil, errors.New("logstash address cannot be empty")
	}

	// 2. 初始化处理器（全局属性默认包含source）
	handler := &LogstashHandler{
		config: config,
		level:  config.Level,
		globalAttrs: map[string]interface{}{
			"source": "go-application", // 固定应用来源，避免遗漏
		},
		cache:  make([]*logEntry, 0, config.BatchSize), // 预分配缓存
		muLock: NewSpinLock(),                          // 使用自定义自旋锁保护连接
	}

	// 3. 建立初始连接
	if err := handler.connect(); err != nil {
		return nil, err
	}

	// 4. 启动批量Worker
	handler.workerCtx, handler.workerCancel = context.WithCancel(context.Background())
	handler.wg.Add(1)
	go handler.batchWorker()

	return handler, nil
}

// connect 连接逻辑不变（仅保护连接）
func (h *LogstashHandler) connect() error {
	h.muLock.Lock()
	defer h.muLock.Unlock()

	if h.conn != nil {
		_ = h.conn.Close()
		h.conn = nil
	}

	var conn net.Conn
	var err error
	for i := 0; i < h.config.MaxRetry; i++ {
		conn, err = net.DialTimeout("tcp", h.config.Address, h.config.ConnectTimeout)
		if err == nil {
			h.conn = conn
			slog.Debug("logstash connection established")
			return nil
		}
		if i < h.config.MaxRetry-1 {
			time.Sleep(h.config.ReconnectInterval)
			slog.Warn("reconnecting to logstash", "retry", i+1, "err", err)
		}
	}
	return errors.New("failed to connect to logstash after " + string(rune(h.config.MaxRetry)) + " retries: " + err.Error())
}

// batchWorker 批量触发逻辑不变（时间/数量双触发）
func (h *LogstashHandler) batchWorker() {
	defer h.wg.Done()
	ticker := time.NewTicker(h.config.BatchMaxWait)
	defer ticker.Stop()

	for {
		select {
		case <-h.workerCtx.Done():
			slog.Debug("batch worker exiting, flushing remaining logs")
			if err := h.flushCache(); err != nil {
				slog.Error("failed to flush remaining logs", "err", err)
			}
			return
		case <-ticker.C:
			_ = h.flushCache()
		}
	}
}

// 关键修复2：将logEntry转为平级的map（合并Attrs到顶层），再序列化
func (h *LogstashHandler) entryToFlatMap(entry *logEntry) map[string]interface{} {
	// 1. 基础字段（@timestamp、level、message、source、groups）
	flatMap := map[string]interface{}{
		"@timestamp": entry.Timestamp,
		"level":      entry.Level,
		"message":    entry.Message,
		"source":     entry.Source,
	}
	// 2. 分组（非空才添加）
	if len(entry.Groups) > 0 {
		flatMap["groups"] = entry.Groups
	}
	// 3. 合并业务属性（平级到顶层，修复嵌套问题）
	for k, v := range entry.Attrs {
		flatMap[k] = v
	}
	return flatMap
}

// flushCache 修复：正确序列化平级日志结构
func (h *LogstashHandler) flushCache() error {
	// 1. 取出缓存日志
	h.muCache.Lock()
	if len(h.cache) == 0 {
		h.muCache.Unlock()
		return nil
	}
	batch := make([]*logEntry, len(h.cache))
	copy(batch, h.cache)
	h.cache = h.cache[:0]
	h.muCache.Unlock()

	// 2. 构建批量JSON（每条日志平级，符合Logstash期望）
	var batchData []byte
	for _, entry := range batch {
		// 转为平级map（合并Attrs）
		flatLog := h.entryToFlatMap(entry)
		// 序列化（此时所有字段都会被包含）
		data, err := json.Marshal(flatLog)
		if err != nil {
			slog.Error("failed to marshal log", "err", err, "message", entry.Message)
			continue
		}
		batchData = append(batchData, data...)
		batchData = append(batchData, '\n') // 每条日志换行（json_lines格式）
	}

	// 3. 批量写入Logstash
	h.muLock.Lock()
	defer h.muLock.Unlock()

	if h.conn == nil {
		if err := h.connect(); err != nil {
			return err
		}
	}

	// 设置写入超时
	h.conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))
	_, err := h.conn.Write(batchData)
	if err != nil {
		slog.Error("failed to write batch logs", "err", err, "batch_size", len(batch))
		// 重试逻辑
		if reconnectErr := h.connect(); reconnectErr != nil {
			return errors.New("reconnect failed: " + reconnectErr.Error())
		}
		h.conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))
		_, retryErr := h.conn.Write(batchData)
		if retryErr != nil {
			slog.Error("failed to retry batch write", "err", retryErr)
			return retryErr
		}
		slog.Debug("retry batch write success")
	}

	slog.Debug("batch write success", "batch_size", len(batch), "data_len", len(batchData))
	return nil
}

// Handle 修复：正确构建logEntry（导出字段+平级属性）
func (h *LogstashHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.isClosed.Load() {
		return errors.New("handler is closed")
	}
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	// 1. 合并全局属性和当前日志属性（平级，避免嵌套）
	attrs := make(map[string]interface{})
	// 先加全局属性
	h.muCache.Lock()
	for k, v := range h.globalAttrs {
		attrs[k] = v
	}
	h.muCache.Unlock()
	// 再加当前日志属性（覆盖同名全局属性，符合业务预期）
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	// 2. 构建logEntry（字段全导出，@timestamp格式正确）
	entry := &logEntry{
		Timestamp: time.Now().Format(time.RFC3339Nano), // 符合Logstash时间格式
		Level:     r.Level.String(),                    // 导出：INFO/ERROR等
		Message:   r.Message,                           // 导出：日志内容
		Source:    "go-application",                    // 冗余确保不丢失
		Groups:    h.groups,                            // 导出：分组（如user）
		Attrs:     attrs,                               // 平级属性（后续合并到顶层）
	}

	// 3. 加入缓存（带容量保护）
	h.muCache.Lock()
	defer h.muCache.Unlock()

	// 缓存满：紧急刷出
	if len(h.cache) >= h.config.MaxCacheSize {
		slog.Warn("log cache full, emergency flush", "max_size", h.config.MaxCacheSize)
		h.muCache.Unlock()
		_ = h.flushCache()
		h.muCache.Lock()
	}

	// 添加到缓存
	h.cache = append(h.cache, entry)

	// 达到批量阈值：触发刷出
	if len(h.cache) >= h.config.BatchSize {
		slog.Debug("batch size reached, flush", "size", len(h.cache), "threshold", h.config.BatchSize)
		h.muCache.Unlock()
		_ = h.flushCache()
		h.muCache.Lock()
	}

	return nil
}

// Enabled 不变（级别检查）
func (h *LogstashHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// WithAttrs 修复：全局属性平级合并（不嵌套）
func (h *LogstashHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.muCache.Lock()
	defer h.muCache.Unlock()

	// 复制原有全局属性
	newGlobalAttrs := make(map[string]interface{}, len(h.globalAttrs)+len(attrs))
	for k, v := range h.globalAttrs {
		newGlobalAttrs[k] = v
	}
	// 添加新属性（覆盖同名）
	for _, a := range attrs {
		newGlobalAttrs[a.Key] = a.Value.Any()
	}

	// 返回新Handler（保持不可变性）
	return &LogstashHandler{
		config:       h.config,
		conn:         h.conn,
		level:        h.level,
		globalAttrs:  newGlobalAttrs,
		groups:       append([]string{}, h.groups...), // 深拷贝分组
		cache:        h.cache,
		workerCtx:    h.workerCtx,
		workerCancel: h.workerCancel,
	}
}

// WithGroup 不变（深拷贝分组）
func (h *LogstashHandler) WithGroup(name string) slog.Handler {
	h.muCache.Lock()
	defer h.muCache.Unlock()

	newGroups := make([]string, 0, len(h.groups)+1)
	newGroups = append(newGroups, h.groups...)
	newGroups = append(newGroups, name)

	return &LogstashHandler{
		config:       h.config,
		conn:         h.conn,
		level:        h.level,
		globalAttrs:  h.globalAttrs,
		groups:       newGroups,
		cache:        h.cache,
		workerCtx:    h.workerCtx,
		workerCancel: h.workerCancel,
	}
}

func (h *LogstashHandler) Close() error {
	if h.isClosed.CompareAndSwap(false, true) {
		h.workerCancel()
		h.wg.Wait()
		h.muLock.Lock()
		err := h.flushCache()
		if err != nil {
			return err
		}
		defer h.muLock.Unlock()
		if h.conn != nil {
			_ = h.conn.Close()
			h.conn = nil
		}

		slog.Debug("handler closed")
	}
	return nil
}

// NewLogger 不变（创建Logger）
func NewLogger(config Config) (*slog.Logger, error) {
	handler, err := newLogstashHandler(config)
	if err != nil {
		return nil, err
	}
	return slog.New(handler), nil
}

// NewDefaultLogger 不变（简化默认创建）
func NewDefaultLogger(address string) (*slog.Logger, error) {
	config := Config{
		Address:      address,
		Level:        slog.LevelInfo,
		BatchSize:    100,
		BatchMaxWait: 200 * time.Millisecond,
	}
	return NewLogger(config)
}
