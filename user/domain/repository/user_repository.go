package repository

import (
	"user/domain/model"

	"gorm.io/gorm"
)

type IUserRepository interface {
	InitTable() error
}

// 创建orderRepository
func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{mysqlDb: db}
}

type UserRepository struct {
	mysqlDb *gorm.DB
}

// 初始化表
func (u *UserRepository) InitTable() error {
	return u.mysqlDb.AutoMigrate(&model.User{}).Error
}
