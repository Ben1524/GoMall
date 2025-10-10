package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"paymentApi/proto/payment"
	"paymentApi/proto/paymentApi"
	"strconv"

	"github.com/Ben1524/GoMall/common/elk_log"
	"github.com/plutov/paypal/v3"
	"go.opentelemetry.io/otel/trace"
)

type PaymentApi struct {
	PaymentService payment.PaymentService
	tracer         trace.Tracer
}

const (
	ClientID string = "Ab4q3_yda8OnhXn13HvQGfCJV9tcMBABkmhMc5PAisRps4fOXlkf_f9NqT24He67A6Vrf4A4xLau0ld4"
)

func NewPaymentApiHandler(paymentService payment.PaymentService) *PaymentApi {
	return &PaymentApi{PaymentService: paymentService,
		tracer: trace.NewNoopTracerProvider().Tracer("paymentApi/handler", trace.WithInstrumentationVersion("v1.0.0")),
	}
}

// PaymentApi.PayPalRefund 通过API向外暴露为/paymentApi/payPalRefund，接收http请求
// 即：/paymentApi/payPalRefund 请求会调用 go.micro.api.paymentApi 服务的PaymentApi.PayPalRefund
func (e *PaymentApi) PayPalRefund(ctx context.Context, req *paymentApi.Request, rsp *paymentApi.Response) error {
	//验证payment 支付通道是否赋值
	if err := isOK("payment_id", req); err != nil {
		rsp.StatusCode = 500
		return err
	}
	//验证 退款号
	if err := isOK("refund_id", req); err != nil {
		rsp.StatusCode = http.StatusBadRequest
		return err
	}
	//验证 退款金额
	if err := isOK("money", req); err != nil {
		rsp.StatusCode = http.StatusBadRequest
		return err
	}

	//获取paymentID
	payID, err := strconv.ParseInt(req.Get["payment_id"].Values[0], 10, 64)
	if err != nil {
		ErrorHandle(err)
		return err
	}
	//获取支付通道信息
	paymentInfo, err := e.PaymentService.FindPaymentByID(ctx, &payment.PaymentID{PaymentId: payID})
	if err != nil {
		ErrorHandle(err)
		return err
	}
	//SID 获取 paymentInfo.PaymentSid
	//支付模式
	status := paypal.APIBaseSandBox
	if paymentInfo.PaymentStatus == "paypal_live" {
		status = paypal.APIBaseLive
	}
	//退款例子
	payout := paypal.Payout{
		SenderBatchHeader: &paypal.SenderBatchHeader{
			EmailSubject: req.Get["refund_id"].Values[0] + " 提醒你收款！",
			EmailMessage: req.Get["refund_id"].Values[0] + " 您有一个收款信息！",
			//每笔转账都要唯一
			SenderBatchID: req.Get["refund_id"].Values[0],
		},
		Items: []paypal.PayoutItem{
			{
				RecipientType: "EMAIL",
				//RecipientWallet: "",
				Receiver: "sb-vvhq82259765@personal.example.com",
				Amount: &paypal.AmountPayout{
					//币种
					Currency: "USD",
					Value:    req.Get["money"].Values[0],
				},
				Note:         req.Get["refund_id"].Values[0],
				SenderItemID: req.Get["refund_id"].Values[0],
			},
		},
	}
	//创建支付客户端
	payPalClient, err := paypal.NewClient(ClientID, paymentInfo.PaymentSid, status)
	if err != nil {
		ErrorHandle(err)
	}
	// 获取 token
	_, err = payPalClient.GetAccessToken()
	if err != nil {
		ErrorHandle(err)
	}
	paymentResult, err := payPalClient.CreateSinglePayout(payout)
	if err != nil {
		ErrorHandle(err)
	}
	elk_log.Info(fmt.Sprintf("Payment result: %v", paymentResult))
	rsp.Body = req.Get["refund_id"].Values[0] + "支付成功！"
	return err
}

func isOK(key string, req *paymentApi.Request) error {
	if _, ok := req.Get[key]; !ok {
		err := errors.New(key + " 参数异常")
		ErrorHandle(err)
		return err
	}
	return nil
}
