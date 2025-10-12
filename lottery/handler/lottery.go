package handler

import (
	"context"
	"errors"
	"lottery/domain/service"
	pb "lottery/proto/lottery"
)

type Lottery struct{ svc *service.LotteryService }

func New(svc *service.LotteryService) *Lottery { return &Lottery{svc: svc} }

func (h *Lottery) Participate(ctx context.Context, req *pb.ParticipateRequest, rsp *pb.ParticipateResponse) error {
	aw, err := h.svc.Participate(ctx, req.ActivityId, req.UserId)
	if err != nil {
		rsp.Success = false
		rsp.Message = err.Error()
		return nil
	}
	rsp.Success = true
	rsp.Message = "ok"

	productInfo, err := h.svc.GetProductInfo(ctx, aw)
	if err != nil {
		return errors.New("failed to get product info: " + err.Error())
	}

	rsp.Award = &pb.Award{
		Id:            aw.ID,
		ActivityId:    aw.ActivityID,
		ProductId:     aw.ProductID,
		ProductSizeId: valueOrZero(aw.ProductSizeID),
		Probability:   aw.Probability,
		Remaining:     aw.RemainingQuantity,
		ProductName:   productInfo.ProductName,
		ProductImage:  productInfo.ProductImage,
		ProductSize:   productInfo.ProductSize,
		ProductPrice:  productInfo.ProductPrice,
	}
	return nil
}

func (h *Lottery) ListActivities(ctx context.Context, req *pb.ListActivitiesRequest, rsp *pb.ListActivitiesResponse) error {
	acts, err := h.svc.ListActivities(ctx, req.Status)
	if err != nil {
		return err
	}
	rsp.Activities = make([]*pb.Activity, 0, len(acts))
	for _, a := range acts {
		rsp.Activities = append(rsp.Activities, &pb.Activity{
			Id:          a.ID,
			Name:        a.ActivityName,
			Description: a.Description,
			StartTime:   a.StartTime.Unix(),
			EndTime:     a.EndTime.Unix(),
			Status:      int32(a.Status),
		})
	}
	return nil
}

func (h *Lottery) ListAwards(ctx context.Context, req *pb.ListAwardsRequest, rsp *pb.ListAwardsResponse) error {
	if req.ActivityId == 0 {
		return errors.New("activity_id required")
	}
	aws, err := h.svc.ListAwards(ctx, req.ActivityId)
	if err != nil {
		return err
	}
	rsp.Awards = make([]*pb.Award, 0, len(aws))
	for _, aw := range aws {
		productInfo, err := h.svc.GetProductInfo(ctx, aw)
		if err != nil {
			return errors.New("failed to get product info: " + err.Error())
		}
		rsp.Awards = append(rsp.Awards, &pb.Award{
			Id:            aw.ID,
			ActivityId:    aw.ActivityID,
			ProductId:     aw.ProductID,
			ProductSizeId: valueOrZero(aw.ProductSizeID),
			Probability:   aw.Probability,
			Remaining:     aw.RemainingQuantity,
			ProductName:   productInfo.ProductName,
			ProductImage:  productInfo.ProductImage,
			ProductSize:   productInfo.ProductSize,
			ProductPrice:  productInfo.ProductPrice,
		})
	}
	return nil
}

func valueOrZero(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
