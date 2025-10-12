package repository

import (
	"context"

	m "lottery/domain/model"
	"lottery/domain/model/product"

	"gorm.io/gorm"
)

type AwardRepository interface {
	ListByActivity(ctx context.Context, activityID int64) ([]*m.LotteryAward, error)
	ListByActivityV2(ctx context.Context, activityID int64) (chan *m.LotteryAward, error) // 新增：异步版本，返回奖品通道
	// DecrementRemaining 原子扣减库存，成功返回影响行数1，失败（无库存）为0
	DecrementRemaining(ctx context.Context, awardID int64) (int64, error)
	FindAwardByRanking(ctx context.Context, activityID int64, ranking int64) (*m.LotteryAward, error)
	GetProductInfo(ctx context.Context, award *m.LotteryAward) (*product.ProductInfo, error) // 新增：获取奖品及关联产品信息
}

type awardRepository struct {
	db *gorm.DB
}

func NewAwardRepository(db *gorm.DB) AwardRepository {
	return &awardRepository{db: db}
}

func (r *awardRepository) GetProductInfo(ctx context.Context, award *m.LotteryAward) (*product.ProductInfo, error) {
	// 多表连接查询，获取奖品及其关联的产品信息
	// 说明：
	// - product_sizes 与 products 通过 size_product_id 关联
	// - product_images 与 products 通过 image_product_id 关联
	// - 使用表别名避免歧义
	var info product.ProductInfo
	err := r.db.WithContext(ctx).
		Table("products p").
		Select("p.product_name, p.product_price, pi.image_url AS product_image, ps.size_name AS product_size").
		Joins("LEFT JOIN product_sizes ps ON ps.size_product_id = p.id").
		Joins("LEFT JOIN product_images pi ON pi.image_product_id = p.id").
		Where("p.id = ? AND ps.id = ? AND pi.is_primary = true", award.ProductID, award.ProductSizeID).
		Limit(1).
		Scan(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (r *awardRepository) ListByActivityV2(ctx context.Context, activityID int64) (chan *m.LotteryAward, error) {
	outChan := make(chan *m.LotteryAward, 500) // 带缓冲通道，防止阻塞
	go func() {
		defer close(outChan)
		var awards []*m.LotteryAward
		if err := r.db.WithContext(ctx).
			Where("activity_id = ?", activityID).
			Find(&awards).Error; err != nil {
			return
		}
		for _, award := range awards {
			outChan <- award
		}
	}()
	return outChan, nil
}

func (r *awardRepository) FindAwardByRanking(ctx context.Context, activityID int64, ranking int64) (*m.LotteryAward, error) {
	var award m.LotteryAward
	if err := r.db.WithContext(ctx).
		Where("activity_id = ? AND product_ranking = ?", activityID, ranking).
		First(&award).Error; err != nil {
		return nil, err
	}
	return &award, nil
}

func (r *awardRepository) ListByActivity(ctx context.Context, activityID int64) ([]*m.LotteryAward, error) {
	var list []*m.LotteryAward
	if err := r.db.WithContext(ctx).
		Where("activity_id = ?", activityID).
		Order("sort ASC, id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *awardRepository) DecrementRemaining(ctx context.Context, awardID int64) (int64, error) {
	tx := r.db.WithContext(ctx).Model(&m.LotteryAward{}).
		Where("id = ? AND remaining_quantity > 0", awardID).
		UpdateColumn("remaining_quantity", gorm.Expr("remaining_quantity - 1"))
	return tx.RowsAffected, tx.Error
}
