package handler

import (
	"cartApi/proto/cart"
	pb "cartApi/proto/cartApi"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CartApiHandler struct {
	cli cart.CartService
}

const defaultRequestTimeout = 5 * time.Second

func NewCartApiHandler(cli cart.CartService) *CartApiHandler {
	return &CartApiHandler{
		cli: cli,
	}
}

func (c *CartApiHandler) FindAll(ctx context.Context, req *pb.Request, res *pb.Response) error {

	return nil
}

// RegisterRoutes 将购物车相关路由注册到给定路由组。
func (c *CartApiHandler) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/carts", c.handleAddCart)
	group.DELETE("/carts/user/:userID", c.handleCleanCart)
	group.PATCH("/carts/:id/increase", c.handleIncreaseItem) //
	group.PATCH("/carts/:id/decrease", c.handleDecreaseItem)
	group.DELETE("/carts/:id", c.handleDeleteItem)
	group.GET("/carts/user/:userID", c.handleGetAll)
}

func (c *CartApiHandler) handleAddCart(ctx *gin.Context) {
	var payload cart.CartInfo
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		respondBadRequest(ctx, "invalid request payload", err)
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), defaultRequestTimeout)
	defer cancel()

	resp, err := c.cli.AddCart(requestCtx, &payload)
	if err != nil {
		respondServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"cart_id": resp.GetCartId(),
		"message": resp.GetMsg(),
	})
}

func (c *CartApiHandler) handleCleanCart(ctx *gin.Context) {
	userID, ok := parseIDParam(ctx, "userID")
	if !ok {
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), defaultRequestTimeout)
	defer cancel()

	resp, err := c.cli.CleanCart(requestCtx, &cart.Clean{UserId: userID})
	if err != nil {
		respondServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resp.GetMeg()})
}

func (c *CartApiHandler) handleIncreaseItem(ctx *gin.Context) {
	c.handleChangeItem(ctx, true)
}

func (c *CartApiHandler) handleDecreaseItem(ctx *gin.Context) {
	c.handleChangeItem(ctx, false)
}

func (c *CartApiHandler) handleDeleteItem(ctx *gin.Context) {
	id, ok := parseIDParam(ctx, "id")
	if !ok {
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), defaultRequestTimeout)
	defer cancel()

	resp, err := c.cli.DeleteItemByID(requestCtx, &cart.CartID{Id: id})
	if err != nil {
		respondServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resp.GetMeg()})
}

func (c *CartApiHandler) handleGetAll(ctx *gin.Context) {
	userID, ok := parseIDParam(ctx, "userID")
	if !ok {
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), defaultRequestTimeout)
	defer cancel()

	resp, err := c.cli.GetAll(requestCtx, &cart.CartFindAll{UserId: userID})
	if err != nil {
		respondServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"items": resp.GetCartInfo()})
}

func (c *CartApiHandler) handleChangeItem(ctx *gin.Context, increase bool) {
	id, ok := parseIDParam(ctx, "id")
	if !ok {
		return
	}

	var body struct {
		Change int64 `json:"change_num" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		respondBadRequest(ctx, "invalid request payload", err)
		return
	}

	if body.Change <= 0 {
		respondBadRequest(ctx, "change_num must be greater than 0", nil)
		return
	}

	requestCtx, cancel := context.WithTimeout(ctx.Request.Context(), defaultRequestTimeout)
	defer cancel()

	item := &cart.Item{Id: id, ChangeNum: body.Change}
	var (
		resp *cart.Response
		err  error
	)

	if increase {
		resp, err = c.cli.Incr(requestCtx, item)
	} else {
		resp, err = c.cli.Decr(requestCtx, item)
	}

	if err != nil {
		respondServiceError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resp.GetMeg()})
}

func parseIDParam(ctx *gin.Context, key string) (int64, bool) {
	raw := ctx.Param(key)
	if raw == "" {
		respondBadRequest(ctx, fmt.Sprintf("missing path parameter: %s", key), nil)
		return 0, false
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		respondBadRequest(ctx, fmt.Sprintf("invalid %s", key), err)
		return 0, false
	}
	return value, true
}

func respondBadRequest(ctx *gin.Context, message string, err error) {
	payload := gin.H{"error": message}
	if err != nil {
		payload["details"] = err.Error()
	}
	ctx.JSON(http.StatusBadRequest, payload)
}

func respondServiceError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}
