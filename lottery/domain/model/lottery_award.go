package model

import (
	"lottery/domain/model/product"
	"time"

	"gorm.io/gorm"
)

// LotteryAward 抽奖活动奖品关联表（多对多关系）
type LotteryAward struct {
	ID int64 `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	// 关键修正：通过 uniqueIndex 标签定义联合索引，所有字段使用相同的索引名
	ActivityID        int64          `gorm:"column:activity_id;type:bigint;not null;comment:'关联的抽奖活动ID';uniqueIndex:uix_activity_product_size" json:"activity_id"`
	ProductRanking    int64          `gorm:"column:product_ranking;type:bigint;not null;comment:'关联的奖品产品排名';uniqueIndex:uix_activity_product_size" json:"product_ranking"`
	ProductID         int64          `gorm:"column:product_id;not null;comment:'关联的奖品产品ID';uniqueIndex:uix_activity_product_size" json:"product_id"`
	ProductImageID    *int64         `gorm:"column:product_image_id;null;comment:'奖品图片ID（可选）'" json:"product_image_id"`
	ProductSizeID     *int64         `gorm:"column:product_size_id;null;comment:'奖品规格ID（可选）';uniqueIndex:uix_activity_product_size" json:"product_size_id"`
	TotalQuantity     int64          `gorm:"column:total_quantity;type:bigint;not null;comment:'该奖品总数量'" json:"total_quantity"`
	RemainingQuantity int64          `gorm:"column:remaining_quantity;type:bigint;not null;comment:'该奖品剩余数量'" json:"remaining_quantity"`
	Probability       float64        `gorm:"column:probability;type:decimal(10,6);not null;comment:'中奖概率（0-100的百分比）'" json:"probability"`
	Sort              int            `gorm:"column:sort;type:int;default:0;comment:'展示排序'" json:"sort"`
	CreateTime        time.Time      `gorm:"column:create_time;type:datetime;autoCreateTime;comment:'创建时间'" json:"create_time"`
	UpdateTime        time.Time      `gorm:"column:update_time;type:datetime;autoUpdateTime;comment:'更新时间'" json:"update_time"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index" json:"-"`

	// 关联的产品信息
	Product     *product.Product     `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
	ProductSize *product.ProductSize `gorm:"foreignKey:ProductSizeID;references:ID" json:"product_size,omitempty"`
}

// 表名指定
func (LotteryAward) TableName() string {
	return "lottery_awards"
}
