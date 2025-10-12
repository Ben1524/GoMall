package repository

import (
	"user/domain/model"

	"gorm.io/gorm"
)

type IUserRepository interface {
	InitTable() error
	Create(user *model.User) (int64, error)
	FindByID(id int64) (*model.User, error)
	FindByEmail(identifier string) (*model.User, error)
	UpdateProfile(id int64, phone, email, avatar string) (*model.User, error)
	UpdatePassword(id int64, newHash string) error
}

type UserRepository struct {
	mysqlDb *gorm.DB
}

// 创建orderRepository
func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{mysqlDb: db}
}

// 初始化表
func (u *UserRepository) InitTable() error {
	return u.mysqlDb.AutoMigrate(&model.User{})
}

func (u *UserRepository) Create(user *model.User) (int64, error) {
	if err := u.mysqlDb.Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (u *UserRepository) FindByID(id int64) (*model.User, error) {
	var m model.User
	if err := u.mysqlDb.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (u *UserRepository) FindByEmail(identifier string) (*model.User, error) {
	var m model.User
	if err := u.mysqlDb.Where("email = ?", identifier).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (u *UserRepository) UpdateProfile(id int64, phone, email, avatar string) (*model.User, error) {
	updates := map[string]interface{}{}
	if phone != "" {
		updates["phone"] = phone
	}
	if email != "" {
		updates["email"] = email
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return u.FindByID(id)
	}
	if err := u.mysqlDb.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return u.FindByID(id)
}

func (u *UserRepository) UpdatePassword(id int64, newHash string) error {
	return u.mysqlDb.Model(&model.User{}).Where("id = ?", id).Update("password_hash", newHash).Error
}
