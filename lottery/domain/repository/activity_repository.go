package repository

import (
	"context"
	"time"

	m "lottery/domain/model"

	"gorm.io/gorm"
)

const (
	ActivityStatusNotStarted = 0 // 未开始
	ActivityStatusActive     = 1 // 进行中
	ActivityStatusEnded      = 2 // 已结束
)

type ActivityRepository interface {
	List(ctx context.Context, status int32) ([]*m.LotteryActivity, error)
	Get(ctx context.Context, id int64) (*m.LotteryActivity, error)
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) List(ctx context.Context, status int32) ([]*m.LotteryActivity, error) {
	var list []*m.LotteryActivity
	q := r.db.WithContext(ctx).Model(&m.LotteryActivity{})
	now := time.Now()
	switch status {
	case 0: // 未开始
		q = q.Where("start_time > ?", now)
	case 1: // 进行中
		q = q.Where("start_time <= ? AND end_time >= ?", now, now)
	case 2: // 已结束
		q = q.Where("end_time < ?", now)
	default:
		// all
	}
	if err := q.Order("start_time DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *activityRepository) Get(ctx context.Context, id int64) (*m.LotteryActivity, error) {
	var act m.LotteryActivity
	if err := r.db.WithContext(ctx).First(&act, id).Error; err != nil {
		return nil, err
	}
	return &act, nil
}
