package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/redis/go-redis/v9"

	"user_service/internal/pkg/conf"
)

var (
	// Redis客户端
	Client *redis.Client
)

// 初始化Redis
func InitRedis(config *conf.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:        config.Password,
		DB:              config.DB,
		MaxIdleConns:    config.MaxIdle,
		PoolSize:        config.MaxActive,
		ConnMaxIdleTime: time.Duration(config.IdleTimeout) * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		hlog.Errorf("Redis连接失败: %v", err)
		return nil, err
	}

	hlog.Infof("Redis连接成功: %s:%d", config.Host, config.Port)
	Client = client
	return client, nil
}

// 设置键值
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// 获取值
func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// 删除键
func Del(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}

// 设置过期时间
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// 检查键是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	val, err := Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
