package main

import (
	"product/handler"
	pb "product/proto"

	"go-micro.dev/v5"
)

func main() {
	// Create service
	service := micro.New("product")

	// Initialize service
	service.Init()

	// Register handler
	pb.RegisterProductHandler(service.Server(), handler.New())

	// Run service
	service.Run()
}
