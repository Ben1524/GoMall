package handler

import (
	"context"
	"fmt"
	"product/domain/model"
	"product/domain/service"
	. "product/proto/product"

	common "github.com/Ben1524/GoMall/common/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Product struct {
	ProductDataService service.IProductDataService
	tracer             trace.Tracer // 新增：用于创建span的tracer
}

// 初始化handler时，创建唯一的tracer
func NewProductHandler(service service.IProductDataService) *Product {
	return &Product{
		ProductDataService: service,
		// 定义tracer名称（建议包含服务名和组件名，确保唯一）
		tracer: otel.Tracer("product/handler", trace.WithInstrumentationVersion("v1.0.0")),
	}
}

// 根据ID查找商品
func (h *Product) FindProductByID(ctx context.Context, request *RequestID, response *ProductInfo) error {
	ctx, span := h.tracer.Start(ctx, "FindProductByID",
		trace.WithAttributes(
			attribute.Int64("product.id", request.ProductId),
		),
	)
	defer span.End()

	productData, err := h.ProductDataService.FindProductByID(request.ProductId)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if err := common.SwapTo(productData, response); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// 添加商品
func (h *Product) AddProduct(ctx context.Context, request *ProductInfo, response *ResponseProduct) error {
	productAdd := &model.Product{}
	if err := common.SwapTo(request, productAdd); err != nil {
		return err
	}
	fmt.Println(productAdd)
	productID, err := h.ProductDataService.AddProduct(productAdd)
	if err != nil {
		return err
	}
	response.ProductId = productID
	return nil
}

// 商品更新
func (h *Product) UpdateProduct(ctx context.Context, request *ProductInfo, response *Response) error {
	productAdd := &model.Product{}
	if err := common.SwapTo(request, productAdd); err != nil {
		return err
	}
	err := h.ProductDataService.UpdateProduct(productAdd)
	if err != nil {
		return err
	}
	response.Msg = "更新成功"
	return nil
}

// 根据ID删除对应商品
func (h *Product) DeleteProductByID(ctx context.Context, request *RequestID, response *Response) error {
	if err := h.ProductDataService.DeleteProduct(request.ProductId); err != nil {
		return err
	}
	response.Msg = "删除成功"
	return nil
}

// 查找所有商品
func (h *Product) FindAllProduct(ctx context.Context, request *RequestAll, response *AllProduct) error {
	productAll, err := h.ProductDataService.FindAllProduct()
	if err != nil {
		return err
	}

	for _, v := range productAll {
		productInfo := &ProductInfo{}
		err := common.SwapTo(v, productInfo)
		if err != nil {
			return err
		}
		response.ProductInfo = append(response.ProductInfo, productInfo)
	}
	return nil
}
