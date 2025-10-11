package model

import "time"

type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(50);not null" json:"username"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string    `gorm:"type:varchar(20);uniqueIndex" json:"phone,omitempty"`
	Avatar       string    `gorm:"type:varchar(255);not null" json:"avatar,omitempty"`
	Status       int       `gorm:"type:tinyint(1);not null" json:"status,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
