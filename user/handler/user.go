package handler

import (
	"context"
	"fmt"
	"time"
	"user/domain/service"
	pb "user/proto/user"
)

type User struct {
	svc service.IUserService
}

const (
	LoginSuccess = 0
	LoginFail    = 1
)

func NewUserHandler(svc service.IUserService) *User {
	return &User{svc: svc}
}

// 简单占位 token 生成（建议后续使用 common/config 中 jwt.secret 生成标准 JWT）
func makeToken(userID int64) string {
	return fmt.Sprintf("token-%d-%d", userID, time.Now().Unix())
}

func (h *User) Register(ctx context.Context, req *pb.RegisterRequest, rsp *pb.RegisterResponse) error {
	u, err := h.svc.Register(req)
	if err != nil {
		rsp.Code = 1
		rsp.Message = err.Error()
		rsp.User = nil
		return nil
	}
	rsp.Code = 0
	rsp.Message = "ok"
	rsp.User = &pb.UserInfo{
		Id:         u.ID,
		Username:   u.Username,
		Phone:      u.Phone,
		Email:      u.Email,
		Avatar:     u.Avatar,
		Status:     int32(u.Status),
		CreateTime: u.CreatedAt.Format(time.RFC3339),
		UpdateTime: u.UpdatedAt.Format(time.RFC3339),
	}
	return nil
}

func (h *User) Login(ctx context.Context, req *pb.LoginRequest, rsp *pb.LoginResponse) error {
	u, err := h.svc.Login(req)
	if err != nil {
		rsp.Code = LoginFail
		rsp.Message = err.Error()
		rsp.User = nil
		return nil
	}
	rsp.Code = LoginSuccess
	rsp.Message = "ok"
	rsp.User = &pb.UserInfo{
		Id:         u.ID,
		Username:   u.Username,
		Phone:      u.Phone,
		Email:      u.Email,
		Avatar:     u.Avatar,
		Status:     int32(u.Status),
		CreateTime: u.CreatedAt.Format(time.RFC3339),
		UpdateTime: u.UpdatedAt.Format(time.RFC3339),
	}
	return nil
}

func (h *User) GetUser(ctx context.Context, req *pb.GetUserRequest, rsp *pb.GetUserResponse) error {
	if req.UserId == 0 {
		rsp.Code = 1
		rsp.Message = "user_id required"
		return nil
	}
	u, err := h.svc.GetUser(req.UserId)
	if err != nil {
		rsp.Code = 1
		rsp.Message = err.Error()
		rsp.User = nil
		return nil
	}
	rsp.Code = 0
	rsp.Message = "ok"
	rsp.User = &pb.UserInfo{
		Id:         u.ID,
		Username:   u.Username,
		Phone:      u.Phone,
		Email:      u.Email,
		Avatar:     u.Avatar,
		Status:     int32(u.Status),
		CreateTime: u.CreatedAt.Format(time.RFC3339),
		UpdateTime: u.UpdatedAt.Format(time.RFC3339),
	}
	return nil
}

func (h *User) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest, rsp *pb.UpdateUserResponse) error {
	return nil
}

func (h *User) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest, rsp *pb.ChangePasswordResponse) error {
	return nil
}
