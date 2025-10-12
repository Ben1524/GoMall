package handler

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	lottery "web/proto/lottery"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v5"
	"golang.org/x/time/rate"
)

// 1. 全局变量优化：添加限流器（按用户ID限流，防止刷接口）
var (
	lotteryClient lottery.LotteryService
	once          sync.Once
	// 限流配置：每个用户每秒最多1次请求， burst=2（允许短时间突发2次）
	limiter = rate.NewLimiter(1, 2)
	// 存储用户专属限流器（避免全局限流影响正常用户）
	userLimiters = make(map[int64]*rate.Limiter)
	limiterMu    sync.Mutex // 保护userLimiters并发安全
)

// 2. 初始化优化：确保Micro客户端只初始化一次
func newLotteryHandler(r *gin.RouterGroup, srv micro.Service) {
	once.Do(func() {
		lotteryClient = lottery.NewLotteryService("go.micro.service.lottery", srv.Client())
	})
}

// 3. 新增：获取用户专属限流器（避免全局限流）
func getUserLimiter(userID int64) *rate.Limiter {
	limiterMu.Lock()
	defer limiterMu.Unlock()

	if limiter, exists := userLimiters[userID]; exists {
		return limiter
	}
	// 每个用户独立限流：每秒1次，突发2次
	newLimiter := rate.NewLimiter(1, 2)
	userLimiters[userID] = newLimiter
	// 定时清理过期限流器（避免内存泄漏，如1小时未使用则删除）
	go func() {
		time.Sleep(1 * time.Hour)
		limiterMu.Lock()
		delete(userLimiters, userID)
		limiterMu.Unlock()
	}()
	return newLimiter
}

// 4. 统一响应结构体（规范前端处理）
type APIResponse struct {
	Code    int         `json:"code"`    // 0=成功，非0=错误（如400=参数错，401=未登录，404=活动不存在）
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据（成功时返回）
}

// 5. 接口注册优化：加版本号、改POST方法
func RegisterLotteryRouter(r *gin.RouterGroup,cfg *config.Config, srv micro.Service) {
	newLotteryHandler(r, srv)
	{
		// 抽奖接口：POST /api/v1/lottery/participate/:activity_id
		r.POST("/participate/:activity_id", ParticipateHandler)
	}
}

// 6. 核心处理函数（拆分逻辑，提高可维护性）
func ParticipateHandler(c *gin.Context) {
	// -------------------------- 1. 统一响应 defer（避免重复写JSON）
	defer func() {
		if err := recover(); err != nil {
			// 捕获Panic（如类型断言失败），返回500错误
			c.JSON(http.StatusInternalServerError, APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "server internal error: " + err.(string),
				Data:    nil,
			})
		}
	}()

	// -------------------------- 2. 获取并校验活动ID
	activityIDStr := c.Param("activity_id")
	activityID, err := strconv.ParseInt(activityIDStr, 10, 64)
	if err != nil || activityID <= 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid activity_id (must be positive integer)",
			Data:    nil,
		})
		return
	}

	// -------------------------- 3. 获取并校验用户ID（从上下文，需登录中间件注入）
	// 必须检查exists，避免Panic
	tmpUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Code:    http.StatusUnauthorized,
			Message: "user not logged in",
			Data:    nil,
		})
		return
	}
	// 类型断言必须检查ok，避免Panic
	userID, ok := tmpUserID.(int64)
	if !ok || userID <= 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid user_id (must be positive integer)",
			Data:    nil,
		})
		return
	}

	userID = tmpUserID.(int64)
	if userID <= 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid user_id (must be positive integer)",
			Data:    nil,
		})
		return
	}

	// -------------------------- 4. 接口限流（防止恶意刷请求）
	userLimiter := getUserLimiter(userID)
	if !userLimiter.Allow() {
		c.JSON(http.StatusTooManyRequests, APIResponse{
			Code:    http.StatusTooManyRequests,
			Message: "too many requests, please try again later",
			Data:    nil,
		})
		return
	}

	// -------------------------- 5. 调用下游服务（加超时控制，避免阻塞）
	// 设置3秒超时，避免下游服务无响应导致连接耗尽
	ctx, cancel := context.WithTimeout(c, 3*time.Second)
	defer cancel()

	rep, err := lotteryClient.Participate(ctx, &lottery.ParticipateRequest{
		ActivityId: activityID,
		UserId:     userID,
	})
	if err != nil {
		// 区分下游错误类型（假设Micro服务返回的错误包含业务码，此处简化处理）
		c.JSON(http.StatusInternalServerError, APIResponse{
			Code:    http.StatusInternalServerError,
			Message: "call lottery service failed: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// -------------------------- 6. 业务逻辑处理（规范响应）
	if !rep.Success {
		// 业务失败（如活动已结束、库存不足），返回对应提示
		c.JSON(http.StatusBadRequest, APIResponse{
			Code:    4001, // 自定义业务错误码（前端可区分具体失败原因）
			Message: rep.Message,
			Data:    nil,
		})
		return
	}

	// 无中奖场景
	if rep.Award == nil {
		c.JSON(http.StatusOK, APIResponse{
			Code:    0,
			Message: "no award won",
			Data: map[string]interface{}{
				"award_id":   -1,
				"product_id": -1,
			},
		})
		return
	}

	// 中奖场景（返回完整奖品信息，处理浮点精度问题：价格转分）
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "lottery success",
		Data: map[string]interface{}{
			"award_id":      rep.Award.Id,
			"product_id":    rep.Award.ProductId,
			"product_name":  rep.Award.ProductName,
			"product_image": rep.Award.ProductImage,
			"product_size":  rep.Award.ProductSize,
			"product_price": rep.Award.ProductPrice * 100, // 假设下游返回元，转成分避免浮点精度问题
		},
	})
}
