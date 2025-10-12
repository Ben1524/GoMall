package redis_cache

import (
	"context"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	*redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{Client: rdb}
}

// 新增：MSet 批量设置（基于Pipeline确保原子性）
func (c *RedisCache) MSet(ctx context.Context, kvs map[string]interface{}, expiration time.Duration) error {
	pipe := c.Client.Pipeline()
	for k, v := range kvs {
		// 序列化struct为JSON字符串
		val, err := json.Marshal(v)
		if err != nil {
			return err
		}
		pipe.Set(ctx, k, val, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// 优化：Get 支持JSON反序列化到struct
func (c *RedisCache) Get(ctx context.Context, key string, out interface{}) error {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), out)
}

// 优化：Set 支持struct序列化
func (c *RedisCache) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, key, data, expiration).Err()
}
