package router

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	error "web/error"
	"web/handler"
	"web/handler/auth"

	config "github.com/Ben1524/GoMall/common/config"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v5"
)

var tokenRegexp = regexp.MustCompile("token=[^&]+")

func logFormatter(param gin.LogFormatterParams) string {
	if (param.ClientIP == "127.0.0.1" || param.ClientIP == "::1") && param.Path == "/health" {
		return ""
	}

	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency - param.Latency%time.Second
	}
	path := tokenRegexp.ReplaceAllString(param.Path, "token=[masked]")
	return fmt.Sprintf("%v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format(time.RFC3339),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		path,
		param.ErrorMessage,
	)
}

func NotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, gin.H{
		"error":      fmt.Sprintf("Route %s not found", filepath.Clean(ctx.Request.URL.Path)),
		"error_code": http.StatusNotFound,
	})
}

func NewRouter(ctx context.Context, cfg *config.Config, srv micro.Service) (g *gin.Engine, exit func()) {
	g = gin.New()

	authenticator := auth.NewAuthenticator(handler.NewUserClient(srv),auth.WithTimeout(24*time.Hour),auth.WithMaxRefresh(24*time.Hour))
	jwtMiddleware,err := auth.NewJWTMiddleware(authenticator,cfg.JWT.Secret)
	
	if err != nil {
		panic(err)
	}
	err = jwtMiddleware.MiddlewareInit()
	if err != nil {
		panic(err)
	}

	// nginx相关配置
	g.RemoteIPHeaders = []string{"X-Forwarded-For"}
	g.SetTrustedProxies(nil) // 信任所有代理
	g.ForwardedByClientIP = true
	g.Use(gin.LoggerWithFormatter(logFormatter), gin.Recovery(), error.GinErrorHandler())
	g.NoRoute(NotFound)

	// 注册服务路由增加中间件验证
	userGroup := g.Group("/user/v1/")
	handler.RegisterUserRouter(userGroup,jwtMiddleware,cfg, srv)

	lotteryGroup := g.Group("/lottery/v1/")
	lotteryGroup.Use(jwtMiddleware.MiddlewareFunc())
	handler.RegisterLotteryRouter(lotteryGroup,cfg, srv)

	exit = func() {
		srv.Server().Stop()
	}
	return g, exit

}
