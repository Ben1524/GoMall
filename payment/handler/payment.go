package handler

import (
	"context"
	common "github.com/Ben1524/GoMall/common/utils"
	"go.opentelemetry.io/otel/trace"

	"payment/domain/model"
	"payment/domain/service"
	payment "payment/proto/payment"
)

type Payment struct {
	PaymentDataService service.IPaymentDataService
	tracer             trace.Tracer // 新增：用于创建span的trace
}

func NewPaymentHandler(service service.IPaymentDataService) *Payment {
	return &Payment{
		PaymentDataService: service,
		// 定义tracer名称（建议包含服务名和组件名，确保唯一）
		tracer: trace.NewNoopTracerProvider().Tracer("payment/handler", trace.WithInstrumentationVersion("v1.0.0")),
	}
}

func (e *Payment) AddPayment(ctx context.Context, request *payment.PaymentInfo, response *payment.PaymentID) error {
	payment := &model.Payment{}
	if err := common.SwapTo(request, payment); err != nil {
		ErrorHandle(err)
	}
	paymentID, err := e.PaymentDataService.AddPayment(payment)
	if err != nil {
		ErrorHandle(err)
	}
	response.PaymentId = paymentID
	return nil
}

func (e *Payment) UpdatePayment(ctx context.Context, request *payment.PaymentInfo, response *payment.Response) error {
	payment := &model.Payment{}
	if err := common.SwapTo(request, payment); err != nil {
		ErrorHandle(err)
	}
	return e.PaymentDataService.UpdatePayment(payment)
}

func (e *Payment) DeletePaymentByID(ctx context.Context, request *payment.PaymentID, response *payment.Response) error {
	return e.PaymentDataService.DeletePayment(request.PaymentId)
}

func (e *Payment) FindPaymentByID(ctx context.Context, request *payment.PaymentID, response *payment.PaymentInfo) error {
	payment, err := e.PaymentDataService.FindPaymentByID(request.PaymentId)
	if err != nil {
		ErrorHandle(err)
	}
	return common.SwapTo(payment, response)
}

func (e *Payment) FindAllPayment(ctx context.Context, request *payment.All, response *payment.PaymentAll) error {
	allPayment, err := e.PaymentDataService.FindAllPayment()
	if err != nil {
		ErrorHandle(err)
	}

	for _, v := range allPayment {
		paymentInfo := &payment.PaymentInfo{}
		if err := common.SwapTo(v, paymentInfo); err != nil {
			ErrorHandle(err)
		}
		response.PaymentInfo = append(response.PaymentInfo, paymentInfo)
	}
	return nil
}
