package service

// service/lottery_service.go
import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	m "lottery/domain/model"
	"lottery/domain/model/product"
	"lottery/domain/repository"
	cache "lottery/infrastructure/redis_cache"
	"lottery/lottery_handler"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// 新增：活动状态常量（需与model.LotteryActivity对齐）
const (
	LotteryStatusNotStarted = 0
	LotteryStatusActive     = 1
	LotteryStatusEnded      = 2
	// 用户抽奖记录缓存Key前缀
	CacheKeyUserLotteryPrefix = "lottery:user:activity:%d:%d" // CacheKeyUserLotteryPrefix + userID + activityID
	// 活动缓存Key前缀
	CacheKeyActivityPrefix = "lottery:activity:"
	// 奖品缓存Key前缀
	CacheKeyAwardPrefix = "lottery:award:"
)

var (
	ErrActivityNotFound  = errors.New("activity not found")
	ErrActivityNotActive = errors.New("activity not active (status or time)")
	ErrNoAwardConfigured = errors.New("no awards configured")
	ErrAllOutOfStock     = errors.New("all awards out of stock")
	ErrUserAlreadyJoined = errors.New("user has already participated in this activity") // 新增：重复抽奖错误
)

// 优化：用sync.Mutex确保随机数并发安全
type LotteryService struct {
	activities repository.ActivityRepository
	awards     repository.AwardRepository
	cache      *cache.RedisCache
}

// 获取产品信息

// 优化：初始化随机数（支持并发安全）
func NewLotteryService(a repository.ActivityRepository, w repository.AwardRepository, c *cache.RedisCache) *LotteryService {
	return &LotteryService{
		activities: a,
		awards:     w,
		cache:      c,
	}
}

func (s *LotteryService) GetProductInfo(ctx context.Context, award *m.LotteryAward) (*product.ProductInfo, error) {
	// 获取奖品信息
	return s.awards.GetProductInfo(ctx, award)
}

// 优化1：修复缓存Key、库存初始化、Context问题
func (s *LotteryService) InitActivityCache(ctx context.Context, activityID int64) error {
	// 1. 从DB获取奖品（剩余库存）
	awards, err := s.awards.ListByActivity(ctx, activityID)
	if err != nil {
		return fmt.Errorf("list awards failed: %w", err)
	}
	if len(awards) == 0 {
		return ErrNoAwardConfigured
	}

	// 2. 从DB获取活动信息
	act, err := s.activities.Get(ctx, activityID)
	if err != nil {
		return fmt.Errorf("get activity failed: %w", err)
	}
	if act == nil {
		return ErrActivityNotFound
	}

	// 3. 计算缓存有效期（活动结束后10分钟）
	now := time.Now()
	var duration time.Duration
	if now.Before(act.EndTime) {
		duration = act.EndTime.Sub(now) + 10*time.Minute
	} else {
		duration = 10 * time.Minute // 已结束活动缓存10分钟
	}

	// 4. 缓存活动信息（用JSON序列化）
	actKey := CacheKeyActivityPrefix + strconv.FormatInt(activityID, 10)
	if err := s.cache.Set(ctx, actKey, act, duration); err != nil {
		return fmt.Errorf("set activity cache failed: %w", err)
	}

	awardKeyPrefix := CacheKeyAwardPrefix + strconv.FormatInt(activityID, 10) + ":"
	sort.Slice(awards, func(i, j int) bool {
		return awards[i].ProductRanking < awards[j].ProductRanking
	})

	multiKeyVals := make(map[string]interface{})
	for _, aw := range awards {
		// 缓存剩余库存（而非总库存）
		awardKey := awardKeyPrefix + strconv.FormatInt(aw.ProductRanking, 10)
		multiKeyVals[awardKey] = aw.RemainingQuantity
		multiKeyVals[awardKey+":prob"] = aw.Probability // 缓存概率
	}
	if err := s.cache.MSet(ctx, multiKeyVals, duration); err != nil {
		return fmt.Errorf("mset awards cache failed: %w", err)
	}

	return nil
}

// 扣减库存
func (s *LotteryService) DecrementAwardStock(ctx context.Context, activityID int64, ranking int64) error {
	awardKey := CacheKeyAwardPrefix + strconv.FormatInt(activityID, 10) + ":" + strconv.FormatInt(ranking, 10)
	res, err := s.cache.Decr(ctx, awardKey).Result()
	if err != nil {
		return fmt.Errorf("decrement award stock failed: %w", err)
	}
	if res < 0 {
		// 库存不足，回滚
		s.cache.Incr(ctx, awardKey)
		slog.Warn("award out of stock", "activityID", activityID, "ranking", ranking)
		return ErrAllOutOfStock
	}
	return nil
}

// 获取所有奖品（缓存降级）
func (s *LotteryService) GetAllAwards(ctx context.Context, activityID int64) ([]*m.LotteryAward, error) {
	// 1. 尝试从redis获取奖品列表
	var awards []*m.LotteryAward
	var cnt int64
	var prob float64
	for i := 1; i < math.MaxInt32; i++ {
		key := fmt.Sprintf("%s%d:%d", CacheKeyAwardPrefix, activityID, i)
		// 如果key不存在直接返回
		var aw m.LotteryAward
		err := s.cache.Get(ctx, key, &cnt)
		if errors.Is(err, redis.Nil) {
			break
		}
		if err == nil {
			aw.RemainingQuantity = cnt
		}
		if err := s.cache.Get(ctx, key+":prob", &prob); err == nil {
			aw.Probability = prob
		}
		// 从key中获取ranking
		rankingStsr := strings.Split(key, ":")
		aw.ProductRanking, _ = strconv.ParseInt(rankingStsr[len(rankingStsr)-1], 10, 64)
		awards = append(awards, &aw)
	}
	if len(awards) > 0 {
		return awards, nil
	}

	// 2. 缓存缺失则从DB加载 ,但是可能会导致缓存和DB不一致
	return nil, ErrNoAwardConfigured
}

// 优化2：增加幂等性、缓存降级、状态校验
func (s *LotteryService) Participate(ctx context.Context, activityID, userID int64) (*m.LotteryAward, error) {
	// 1. 幂等性校验：用户是否已参与
	userLotteryKey := fmt.Sprintf(CacheKeyUserLotteryPrefix, userID, activityID)
	exists, err := s.cache.Exists(ctx, userLotteryKey).Result()
	if err != nil {
		return nil, fmt.Errorf("check user participation failed: %w", err)
	}
	if exists == 1 {
		return nil, ErrUserAlreadyJoined
	}

	const tryCount = 10
	for i := 0; i < tryCount; i++ {
		awards, err := s.GetAllAwards(ctx, activityID)
		if err != nil {
			return nil, err
		}
		if len(awards) == 0 {
			return nil, ErrNoAwardConfigured
		}

		// 计算总概率
		var totalProb float64
		probs := make([]float64, 0, len(awards))
		rankList := make([]int64, 0, len(awards))
		for _, aw := range awards {
			if aw.RemainingQuantity > 0 {
				totalProb += aw.Probability
				rankList = append(rankList, aw.ProductRanking)
				probs = append(probs, aw.Probability)
			}
		}
		if totalProb <= 0 {
			return nil, ErrNoAwardConfigured
		}
		if len(rankList) == 0 {
			return nil, ErrAllOutOfStock
		}
		// 抽奖
		idx := lottery_handler.Lottery(probs)
		if idx < 0 || idx >= len(rankList) {
			return nil, ErrNoAwardConfigured
		}

		fmt.Printf("%v:%v\n", rankList, probs[idx])

		ranking := rankList[idx]

		// 先去更新缓存
		if err := s.DecrementAwardStock(ctx, activityID, int64(ranking)); err != nil {
			if errors.Is(err, ErrAllOutOfStock) {
				// 继续重试
				continue
			}
			return nil, err
		}
		// 获取中奖奖品
		award, err := s.awards.FindAwardByRanking(ctx, activityID, int64(ranking))
		if err != nil || award == nil {
			return nil, fmt.Errorf("find award failed: %w", err)
		}
		// 标记用户已参与（设置与活动同步的有效期）
		if err := s.cache.Set(ctx, userLotteryKey, 1, time.Hour*24); err != nil {
			return nil, fmt.Errorf("mark user participation failed: %w", err)
		}
		// （可选）异步同步库存到DB（用消息队列或定时任务）
		// 这里简化：直接调用DB更新（实际建议异步，避免阻塞抽奖流程）
		if _, err := s.awards.DecrementRemaining(ctx, award.ID); err != nil {
			// 日志记录：库存同步失败，后续重试
			return award, fmt.Errorf("sync remaining to db failed: %w", err)
		}
		return award, nil
	}
	return nil, fmt.Errorf("failed to participate after %d attempts", tryCount)
}

func (s *LotteryService) ListActivities(ctx context.Context, status int32) ([]*m.LotteryActivity, error) {
	return s.activities.List(ctx, status)
}

func (s *LotteryService) ListAwards(ctx context.Context, activityID int64) ([]*m.LotteryAward, error) {
	return s.awards.ListByActivity(ctx, activityID)
}
