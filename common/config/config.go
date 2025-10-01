package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	envConfigFile = "CONFIG_FILE"
)

var defaultConfigCandidates = []string{
	"config/config.yaml",
	"config/config.yml",
	"config.yaml",
	"config.yml",
}


// Config 定义应用的所有配置项，支持 YAML 与环境变量加载。
type Config struct {
	Server   ServerConfig   `json:"server" yaml:"server" mapstructure:"server"`
	Database DatabaseConfig `json:"database" yaml:"database" mapstructure:"database"`
	Redis    RedisConfig    `json:"redis" yaml:"redis" mapstructure:"redis"`
	JWT      JWTConfig      `json:"jwt" yaml:"jwt" mapstructure:"jwt"`
	Log      LogConfig      `json:"log" yaml:"log" mapstructure:"log"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq" yaml:"rabbitmq" mapstructure:"rabbitmq"`
	Etcd     EtcdConfig     `json:"etcd" yaml:"etcd" mapstructure:"etcd"`
	Jaeger   JaegerConfig   `json:"jaeger" yaml:"jaeger" mapstructure:"jaeger"`
	Metrics  MetricsConfig  `json:"metrics" yaml:"metrics" mapstructure:"metrics"`
	Security SecurityConfig `json:"security" yaml:"security" mapstructure:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `json:"host" yaml:"host" mapstructure:"host"`
	Port         string        `json:"port" yaml:"port" mapstructure:"port"`
	Mode         string        `json:"mode" yaml:"mode" mapstructure:"mode"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout" mapstructure:"idle_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `json:"host" yaml:"host" mapstructure:"host"`
	Port            string        `json:"port" yaml:"port" mapstructure:"port"`
	Username        string        `json:"username" yaml:"username" mapstructure:"username"`
	Password        string        `json:"password" yaml:"password" mapstructure:"password"`
	Database        string        `json:"database" yaml:"database" mapstructure:"database"`
	SSLMode         string        `json:"ssl_mode" yaml:"ssl_mode" mapstructure:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `json:"host" yaml:"host" mapstructure:"host"`
	Port         string        `json:"port" yaml:"port" mapstructure:"port"`
	Password     string        `json:"password" yaml:"password" mapstructure:"password"`
	Database     int           `json:"database" yaml:"database" mapstructure:"database"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns" mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout" mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" mapstructure:"write_timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret            string        `json:"secret" yaml:"secret" mapstructure:"secret"`
	AccessExpiration  time.Duration `json:"access_expiration" yaml:"access_expiration" mapstructure:"access_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration" yaml:"refresh_expiration" mapstructure:"refresh_expiration"`
	Issuer            string        `json:"issuer" yaml:"issuer" mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level" yaml:"level" mapstructure:"level"`
	Format     string `json:"format" yaml:"format" mapstructure:"format"`
	Output     string `json:"output" yaml:"output" mapstructure:"output"`
	MaxSize    int    `json:"max_size" yaml:"max_size" mapstructure:"max_size"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	Compress   bool   `json:"compress" yaml:"compress" mapstructure:"compress"`
}

// RabbitMQConfig 消息队列配置
type RabbitMQConfig struct {
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     string `json:"port" yaml:"port" mapstructure:"port"`
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	VHost    string `json:"vhost" yaml:"vhost" mapstructure:"vhost"`
}

// EtcdConfig 服务发现配置
type EtcdConfig struct {
	Endpoints   []string      `json:"endpoints" yaml:"endpoints" mapstructure:"endpoints"`
	Username    string        `json:"username" yaml:"username" mapstructure:"username"`
	Password    string        `json:"password" yaml:"password" mapstructure:"password"`
	DialTimeout time.Duration `json:"dial_timeout" yaml:"dial_timeout" mapstructure:"dial_timeout"`
	TTL         int           `json:"ttl" yaml:"ttl" mapstructure:"ttl"`
}

// JaegerConfig 链路追踪配置
type JaegerConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Host        string  `json:"host" yaml:"host" mapstructure:"host"`
	Port        string  `json:"port" yaml:"port" mapstructure:"port"`
	ServiceName string  `json:"service_name" yaml:"service_name" mapstructure:"service_name"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate" mapstructure:"sample_rate"`
}

// MetricsConfig 监控指标配置
type MetricsConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Host    string `json:"host" yaml:"host" mapstructure:"host"`
	Port    string `json:"port" yaml:"port" mapstructure:"port"`
	Path    string `json:"path" yaml:"path" mapstructure:"path"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimitRPS     int      `json:"rate_limit_rps" yaml:"rate_limit_rps" mapstructure:"rate_limit_rps"`
	RateLimitBurst   int      `json:"rate_limit_burst" yaml:"rate_limit_burst" mapstructure:"rate_limit_burst"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers" mapstructure:"allowed_headers"`
	ExposeHeaders    []string `json:"expose_headers" yaml:"expose_headers" mapstructure:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials" mapstructure:"allow_credentials"`
}

// Load 从 YAML 配置文件加载配置，并允许环境变量覆盖。paths 可以显式指定配置文件，若为空则按顺序尝试默认路径。
func Load(paths ...string) (*Config, error) {
	v := viper.New()
	v.SetTypeByDefaultValue(true)
	applyDefaults(v)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	found, lookedUp, err := attachConfigFile(v, paths...)
	if err != nil {
		return nil, err
	}

	explicitProvided := hasExplicitPath(paths...)
	if !found {
		if envPath := strings.TrimSpace(os.Getenv(envConfigFile)); envPath != "" {
			explicitProvided = true
		}
	}

	if found {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else if explicitProvided {
		return nil, fmt.Errorf("未找到配置文件，已尝试: %s", strings.Join(lookedUp, ", "))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		stringToDurationHook(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

// MustLoad 是 Load 的便捷封装，发生错误时直接 panic。
func MustLoad(paths ...string) *Config {
	cfg, err := Load(paths...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// GetRabbitMQURL 获取RabbitMQ连接URL
func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.VHost,
	)
}

func applyDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 120*time.Second)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "ecommerce")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", time.Hour)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 10)
	v.SetDefault("redis.dial_timeout", 5*time.Second)
	v.SetDefault("redis.read_timeout", 3*time.Second)
	v.SetDefault("redis.write_timeout", 3*time.Second)

	v.SetDefault("jwt.secret", "your-jwt-secret-key")
	v.SetDefault("jwt.access_expiration", 15*time.Minute)
	v.SetDefault("jwt.refresh_expiration", 24*time.Hour)
	v.SetDefault("jwt.issuer", "ecommerce-service")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 3)
	v.SetDefault("log.max_age", 28)
	v.SetDefault("log.compress", true)

	v.SetDefault("rabbitmq.host", "localhost")
	v.SetDefault("rabbitmq.port", "5672")
	v.SetDefault("rabbitmq.username", "guest")
	v.SetDefault("rabbitmq.password", "guest")
	v.SetDefault("rabbitmq.vhost", "/")

	v.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	v.SetDefault("etcd.username", "")
	v.SetDefault("etcd.password", "")
	v.SetDefault("etcd.dial_timeout", 5*time.Second)
	v.SetDefault("etcd.ttl", 30)

	v.SetDefault("jaeger.enabled", false)
	v.SetDefault("jaeger.host", "localhost")
	v.SetDefault("jaeger.port", "14268")
	v.SetDefault("jaeger.service_name", "ecommerce-service")
	v.SetDefault("jaeger.sample_rate", 0.1)

	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.host", "0.0.0.0")
	v.SetDefault("metrics.port", "9090")
	v.SetDefault("metrics.path", "/metrics")

	v.SetDefault("security.rate_limit_rps", 100)
	v.SetDefault("security.rate_limit_burst", 200)
	v.SetDefault("security.allowed_origins", []string{"*"})
	v.SetDefault("security.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("security.allowed_headers", []string{"*"})
	v.SetDefault("security.expose_headers", []string{})
	v.SetDefault("security.allow_credentials", true)
}

func attachConfigFile(v *viper.Viper, explicitPaths ...string) (bool, []string, error) {
	candidates := make([]string, 0, len(explicitPaths)+len(defaultConfigCandidates)+1)
	candidates = append(candidates, explicitPaths...)
	if envPath := strings.TrimSpace(os.Getenv(envConfigFile)); envPath != "" {
		candidates = append(candidates, envPath)
	}
	candidates = append(candidates, defaultConfigCandidates...)

	lookedUp := make([]string, 0, len(candidates))
	seen := map[string]struct{}{}

	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		paths := expandCandidate(candidate)
		for _, p := range paths {
			if _, ok := seen[p]; !ok {
				seen[p] = struct{}{}
			}
			lookedUp = append(lookedUp, p)
			abs, ok, err := ensureFile(p)
			if err != nil {
				return false, lookedUp, fmt.Errorf("检查配置文件 %s 失败: %w", p, err)
			}
			if ok {
				v.SetConfigFile(abs)
				return true, lookedUp, nil
			}
		}
	}

	return false, lookedUp, nil
}

func hasExplicitPath(paths ...string) bool {
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			return true
		}
	}
	return false
}

func expandCandidate(candidate string) []string {
	info, err := os.Stat(candidate)
	if err == nil && info.IsDir() {
		return []string{
			filepath.Join(candidate, "config.yaml"),
			filepath.Join(candidate, "config.yml"),
		}
	}
	return []string{candidate}
}

func ensureFile(path string) (string, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	if info.IsDir() {
		return "", false, nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, true, nil
	}
	return abs, true, nil
}

func stringToDurationHook() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if to != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}

		switch v := data.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return time.Duration(0), nil
			}
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("无法解析 duration '%s': %w", v, err)
			}
			return d, nil
		case int:
			return time.Duration(v) * time.Second, nil
		case int64:
			return time.Duration(v) * time.Second, nil
		case float64:
			return time.Duration(v * float64(time.Second)), nil
		default:
			return data, nil
		}
	}
}
