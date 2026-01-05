package database

import (
	"context"
	"log"
	"time"

	"gin-project/config"

	"github.com/redis/go-redis/v9"
	redisotel "github.com/redis/go-redis/extra/redisotel/v9"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.Config) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 【最佳实践】使用 redisotel 自动追踪所有 Redis 操作（零代码入侵）
	// 仅在追踪启用时注册追踪，避免不必要的性能开销
	if cfg.Tracing.Enabled {
		if err := redisotel.InstrumentTracing(RedisClient); err != nil {
			panic("failed to instrument redis tracing: " + err.Error())
		}
		log.Println("Redis 追踪已启用")
	} else {
		log.Println("Redis 追踪未启用（性能优化模式）")
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		panic("failed to connect to redis: " + err.Error())
	}
}
