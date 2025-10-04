package handler

import (
	"cartApi/proto/cart"
	pb "cartApi/proto/cartApi"
	"context"
)

type CartApiHandler struct {
	cli cart.CartService
}

func NewCartApiHandler(cli cart.CartService) *CartApiHandler {
	return &CartApiHandler{
		cli: cli,
	}
}

func (c *CartApiHandler) FindAll(ctx context.Context, req *pb.Request, res *pb.Response) error {

	return nil
}
