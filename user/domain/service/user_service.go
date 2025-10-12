package service

import (
	"errors"
	"time"

	"user/domain/model"
	"user/domain/repository"
	pb "user/proto/user"

	"golang.org/x/crypto/bcrypt"
)

type IUserService interface {
	Register(req *pb.RegisterRequest) (*model.User, error)
	Login(req *pb.LoginRequest) (*model.User, error)
	GetUser(id int64) (*model.User, error)
	UpdateUser(id int64, req *pb.UpdateUserRequest) (*model.User, error)
	ChangePassword(id int64, oldPwd, newPwd string) error
}

type UserService struct {
	repo repository.IUserRepository
}

func NewUserService(repo repository.IUserRepository) IUserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(req *pb.RegisterRequest) (*model.User, error) {
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, errors.New("username/password/email required")
	}
	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Phone:        req.Phone,
		Avatar:       "",
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	_, err = s.repo.Create(u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Login(req *pb.LoginRequest) (*model.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	// if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
	// 	return nil, errors.New("invalid credentials")
	// }
	return u, nil
}

func (s *UserService) GetUser(id int64) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) UpdateUser(id int64, req *pb.UpdateUserRequest) (*model.User, error) {
	return s.repo.UpdateProfile(id, req.Phone, req.Email, req.Avatar)
}

func (s *UserService) ChangePassword(id int64, oldPwd, newPwd string) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPwd)); err != nil {
		return errors.New("old password not match")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, string(hash))
}
