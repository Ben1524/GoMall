package handler

import (
	"net/http"
	"sync"
	user "web/proto/user"

	"github.com/Ben1524/GoMall/common/config"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v5"
	_ "golang.org/x/time/rate"
)

// 1. 全局变量优化：添加限流器（按用户ID限流，防止刷接口）
var (
	userClient user.UserService
	onceUser          sync.Once
)


func NewUserClient(srv micro.Service) user.UserService {
	onceUser.Do(func() {
		userClient = user.NewUserService("go.micro.service.user", srv.Client())
	})
	return userClient
}

func RegisterUserRouter(r *gin.RouterGroup,jwtMiddleware *jwt.GinJWTMiddleware,cfg *config.Config, srv micro.Service) {
	_ = NewUserClient(srv)
	r.POST("/login", jwtMiddleware.LoginHandler)
	r.POST("/refresh_token", jwtMiddleware.RefreshHandler)
	r.POST("/register", RegisterHandler)
}

func RegisterHandler(c *gin.Context) {
	// 获取注册信息

	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Code: 400, Message: err.Error()})
		return
	}
	// 调用用户服务注册
	user, err := userClient.Register(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Code: 0, Message: "注册成功", Data: user})
}


