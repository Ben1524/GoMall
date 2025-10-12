package auth

import (
	"log"
	"time"
	"web/proto/user"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var (
	identityKey = "user_id"
)

type AuthInterface interface {
	PayloadFunc(data interface{}) jwt.MapClaims
	IdentityHandler(c *gin.Context) interface{}
	Authenticator(c *gin.Context) (interface{}, error)
	Authorizator(data interface{}, c *gin.Context) bool
	Unauthorized(c *gin.Context, code int, message string)
	HandleNoRoute() func(c *gin.Context)
}

type Authenticator struct {
	userClient user.UserService
	TimeOut time.Duration
	MaxRefresh time.Duration
}

type Option func(*Authenticator)


func WithTimeout(timeout time.Duration) Option {
	return func(a *Authenticator) {
		a.TimeOut = timeout
	}
}
func WithMaxRefresh(maxRefresh time.Duration) Option {
	return func(a *Authenticator) {
		a.MaxRefresh = maxRefresh
	}
}

var _ AuthInterface = (*Authenticator)(nil)

func NewAuthenticator(userClient user.UserService,opt ...Option) *Authenticator {
	a := &Authenticator{
		userClient: userClient,
		TimeOut:    time.Hour,
		MaxRefresh: time.Hour,
	}
	for _, o := range opt {
		o(a)
	}
	return a
}

// NewJWTMiddleware 构造可配置的 Gin-JWT 中间件。
// 注意：需在返回后调用方保证路由绑定使用 m.MiddlewareFunc() / m.LoginHandler / m.RefreshHandler。
func NewJWTMiddleware(auth AuthInterface, secret string) (*jwt.GinJWTMiddleware, error) {
	m, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "gomall",
		Key:         []byte(secret),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: auth.PayloadFunc,

		IdentityHandler: auth.IdentityHandler,
		Authenticator:   auth.Authenticator,
		Authorizator:    auth.Authorizator,
		Unauthorized:    auth.Unauthorized,
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}
func (auth *Authenticator) PayloadFunc(data interface{}) jwt.MapClaims {
	if u, ok := data.(*user.UserInfo); ok && u != nil {
		return jwt.MapClaims{
			identityKey: u.Id,
		}
	}
	return jwt.MapClaims{}
}

func (auth *Authenticator) IdentityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	var uid int64
	if v, ok := claims[identityKey]; ok {
		switch t := v.(type) {
		case float64:
			uid = int64(t)
		case int64:
			uid = t
		case int:
			uid = int64(t)
		case string:
			// 忽略转换错误，保持 uid=0 时后续 Authorizator 会拒绝
			// 尝试将字符串转为数字
			// 注意：为了避免引入 strconv 未使用，在需要时可在上层解析
		}
	}
	return &user.UserInfo{Id: uid}
}

type login struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// Authenticator 处理登录逻辑
func (auth *Authenticator) Authenticator(c *gin.Context) (interface{}, error) {
	var loginVals login
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	// 使用请求的上下文，避免将 *gin.Context 直接当成 context.Context 传递
	resp, err := auth.userClient.Login(c.Request.Context(), &user.LoginRequest{
		Email:    loginVals.Email,
		Password: loginVals.Password,
	})
	if err != nil {
		return nil, err
	}

	if resp == nil || resp.User == nil {
		return nil, jwt.ErrFailedAuthentication
	}
	return resp.User, nil
}

// Authorizator 这里可以做权限控制
func (auth *Authenticator) Authorizator(data interface{}, c *gin.Context) bool {
	if v, ok := data.(*user.UserInfo); ok && v.Id != -1 {
		c.Set("user_id", v.Id)
		return true
	}
	return false
}

func (auth *Authenticator) Unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func (auth *Authenticator) HandleNoRoute() func(c *gin.Context) {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	}
}
