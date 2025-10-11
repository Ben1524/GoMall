package model

import (
	"time"

	"gorm.io/gorm"
)

// 抽奖活动状态常量
const (
	LotteryStatusNotStarted = 0 // 未开始
	LotteryStatusActive     = 1 // 进行中
	LotteryStatusEnded      = 2 // 已结束
)

// LotteryActivity 抽奖活动主表
type LotteryActivity struct {
	ID           int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	ActivityName string         `gorm:"column:activity_name;type:varchar(100);not null;comment:'抽奖活动名称'" json:"activity_name"`
	Description  string         `gorm:"column:description;type:varchar(500);null;comment:'活动描述'" json:"description"`
	StartTime    time.Time      `gorm:"column:start_time;type:datetime;not null;comment:'活动开始时间'" json:"start_time"`
	EndTime      time.Time      `gorm:"column:end_time;type:datetime;not null;comment:'活动结束时间'" json:"end_time"`
	Status       int8           `gorm:"column:status;type:tinyint(1);default:0;not null;comment:'活动状态'" json:"status"`
	CreateTime   time.Time      `gorm:"column:create_time;type:datetime;autoCreateTime;comment:'创建时间'" json:"create_time"`
	UpdateTime   time.Time      `gorm:"column:update_time;type:datetime;autoUpdateTime;comment:'更新时间'" json:"update_time"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index" json:"-"`

	// 关联的奖品列表
	Awards []*LotteryAward `gorm:"foreignKey:ActivityID" json:"awards,omitempty"`
}

func (LotteryActivity) TableName() string {
	return "lottery_activities"
}
