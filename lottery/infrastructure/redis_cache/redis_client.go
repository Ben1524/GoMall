package redis_cache

import (
	"fmt"
	"log/slog"

	"github.com/Ben1524/GoMall/common/config"
	"github.com/redis/go-redis/v9"
)

// redis客户端

//Password     string        `json:"password" yaml:"password" mapstructure:"password"`
//Database     int           `json:"database" yaml:"database" mapstructure:"database"`
//PoolSize     int           `json:"pool_size" yaml:"pool_size" mapstructure:"pool_size"`
//MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
//DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout" mapstructure:"dial_timeout"`
//ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout" mapstructure:"read_timeout"`
//WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" mapstructure:"write_timeout"`

func NewRedisClient(cfg *config.Config) *redis.Client {
	if cfg == nil {
		panic("config is nil")
	}

	slog.Info(fmt.Sprintf("redis client config:%v", cfg.Redis))

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%v", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		MinIdleConns: cfg.Redis.MinIdleConns,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	return rdb
}
