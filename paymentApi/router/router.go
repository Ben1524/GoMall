package router

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"log/slog"
	"paymentApi/handler"
	pb "paymentApi/proto/paymentApi"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/gin-gonic/gin"
)

func PayPalRefundHandler(h *handler.PaymentApi) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := buildProtoRequest(c)
		rsp := &pb.Response{}
		if err := h.PayPalRefund(c.Request.Context(), req, rsp); err != nil {
			slog.Error("PayPalRefund 调用失败", "error", err)
			// 如果业务未填充 statusCode，则统一返回 500
			status := http.StatusInternalServerError
			if rsp.StatusCode != 0 {
				status = int(rsp.StatusCode)
			}
			c.String(status, rsp.Body)
			return
		}
		status := http.StatusOK
		if rsp.StatusCode != 0 {
			status = int(rsp.StatusCode)
		}
		// 透传头
		for k, v := range rsp.Header {
			if v != nil {
				for _, vv := range v.Values {
					c.Writer.Header().Add(k, vv)
				}
			}
		}
		if rsp.Body == "" {
			c.Status(status)
			return
		}
		c.String(status, rsp.Body)
	}
}

// New 创建并初始化 Gin 引擎，注册网关路由。
// - cfg: 全局配置（用于设置运行模式等）
// - h:   具体业务处理器，适配 go-micro 风格的 Request/Response
func New(cfg *config.Config, h *handler.PaymentApi) *gin.Engine {
	// 无论配置如何，默认使用 Release 模式；如需自定义可在上层传入并调整
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API 网关路由：兼容历史路径 /paymentApi/payPalRefund
	// 支持 GET/POST，参数从 query/form/json 提取，写入到 proto.Request
	handlerFunc := PayPalRefundHandler(h)

	r.Any("/paymentApi/payPalRefund", handlerFunc) // 兼容历史路径

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"service": "paymentApi-gateway",
		})
	})

	return r
}

// buildProtoRequest 将 gin 上下文转换为 go-micro 风格的 proto Request
func buildProtoRequest(c *gin.Context) *pb.Request {
	req := &pb.Request{
		Method: c.Request.Method,
		Path:   c.Request.URL.Path,
		Url:    c.Request.URL.String(),
		Header: map[string]*pb.Pair{},
		Get:    map[string]*pb.Pair{},
		Post:   map[string]*pb.Pair{},
		Body:   "",
	}

	// headers
	for k, vals := range c.Request.Header {
		req.Header[k] = &pb.Pair{Key: k, Values: vals}
	}

	// query -> Get
	q := c.Request.URL.Query()
	for k, vals := range q {
		req.Get[k] = &pb.Pair{Key: k, Values: vals}
	}

	// body 原样透传
	if c.Request.Body != nil {
		// 复制一份 Body 以免影响后续读取
		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		b, _ := io.ReadAll(tee)
		req.Body = string(b)
		c.Request.Body = io.NopCloser(&buf)
	}

	// 表单（application/x-www-form-urlencoded 或 multipart/form-data）
	if err := c.Request.ParseForm(); err == nil {
		for k, vals := range c.Request.PostForm {
			req.Post[k] = &pb.Pair{Key: k, Values: vals}
		}
	}

	// 如果是 JSON 并且 query 未包含关键字段，也尝试从 JSON 中补充到 Get，方便下游逻辑
	// 仅在常用字段缺失时填充，避免覆盖 query 的优先级
	if (req.Get["payment_id"] == nil || len(req.Get["payment_id"].Values) == 0) ||
		(req.Get["refund_id"] == nil || len(req.Get["refund_id"].Values) == 0) ||
		(req.Get["money"] == nil || len(req.Get["money"].Values) == 0) {
		var jsonMap map[string]interface{}
		if err := json.Unmarshal([]byte(req.Body), &jsonMap); err == nil && jsonMap != nil {
			// 尝试补齐几个关键字段
			setIf := func(key string) {
				if v, ok := jsonMap[key]; ok {
					switch vv := v.(type) {
					case string:
						req.Get[key] = &pb.Pair{Key: key, Values: []string{vv}}
					case float64:
						// JSON 数字
						req.Get[key] = &pb.Pair{Key: key, Values: []string{formatFloat(vv)}}
					case int64:
						req.Get[key] = &pb.Pair{Key: key, Values: []string{formatInt(vv)}}
					case int:
						req.Get[key] = &pb.Pair{Key: key, Values: []string{formatInt(int64(vv))}}
					}
				}
			}
			setIf("payment_id")
			setIf("refund_id")
			setIf("money")
		}
	}

	return req
}

func formatFloat(f float64) string { return trimTrailingZeros(strconv.FormatFloat(f, 'f', -1, 64)) }

func trimTrailingZeros(s string) string {
	// 去掉小数点后多余的 0
	i := len(s)
	for i > 0 && s[i-1] == '0' {
		i--
	}
	if i > 0 && s[i-1] == '.' {
		i--
	}
	return s[:i]
}

func formatInt(i int64) string { return strconv.FormatInt(i, 10) }
