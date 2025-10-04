package router

import (
	"net/http"
	"strings"

	"cartApi/handler"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/gin-gonic/gin"
)

const defaultAPIPrefix = "/api/v1"

// New 创建并返回配置好的 gin.Engine。
// cfg 可为 nil，此时使用默认模式和 API 前缀。
func New(cfg *config.Config, cartHandler *handler.CartApiHandler) *gin.Engine {
	if cartHandler == nil {
		panic("cart handler must not be nil")
	}

	gin.SetMode(resolveGinMode(cfg))

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	applySecurity(engine, cfg)

	apiPrefix := defaultAPIPrefix

	api := engine.Group(apiPrefix)
	cartHandler.RegisterRoutes(api)

	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	return engine
}

func resolveGinMode(cfg *config.Config) string {
	if cfg == nil {
		return gin.DebugMode
	}

	switch strings.ToLower(cfg.Server.Mode) {
	case "release", "prod", "production":
		return gin.ReleaseMode
	case "test", "testing":
		return gin.TestMode
	default:
		return gin.DebugMode
	}
}

func applySecurity(engine *gin.Engine, cfg *config.Config) {
	if cfg == nil {
		return
	}

	sec := cfg.Security
	if len(sec.AllowedOrigins) == 0 && len(sec.AllowedMethods) == 0 && len(sec.AllowedHeaders) == 0 {
		return
	}

	allowedMethods := strings.Join(fallbackSlice(sec.AllowedMethods, []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}), ", ")
	allowedHeaders := strings.Join(fallbackSlice(sec.AllowedHeaders, []string{"*"}), ", ")
	exposeHeaders := strings.Join(sec.ExposeHeaders, ", ")
	wildcardOrigin := contains(sec.AllowedOrigins, "*")

	engine.Use(func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		switch {
		case wildcardOrigin && sec.AllowCredentials && origin != "":
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		case wildcardOrigin:
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		case origin != "" && contains(sec.AllowedOrigins, origin):
			ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		if allowedMethods != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		}
		if allowedHeaders != "" {
			ctx.Writer.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		}
		if exposeHeaders != "" {
			ctx.Writer.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		}
		if sec.AllowCredentials {
			ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	})
}

func fallbackSlice(primary []string, fallback []string) []string {
	if len(primary) == 0 {
		return fallback
	}
	return primary
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
