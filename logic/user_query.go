package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"gin-project/database"
	"gin-project/model"

	"github.com/redis/go-redis/v9"
)

// GetUserByID 根据ID查询用户，优先从缓存获取
// 使用带追踪的数据库和缓存客户端，自动追踪所有操作
func GetUserByID(ctx context.Context, id uint) (*model.User, error) {

	// 先从缓存获取
	cacheKey := fmt.Sprintf(UserCacheKey, id)
	user := &model.User{}

	// 从Redis获取数据（使用带追踪的客户端，自动追踪）
	jsonData, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中，解析数据
		if err := json.Unmarshal([]byte(jsonData), user); err == nil {
			return user, nil
		}
		// 反序列化错误已由 Redis 追踪自动记录
	} else if err != redis.Nil {
		// Redis 错误已由追踪自动记录
	}

	// 缓存未命中，从数据库查询（使用带追踪的客户端，自动追踪）
	err = database.DB.WithContext(ctx).First(user, id).Error
	if err != nil {
		return nil, err
	}

	// 将查询结果存入缓存（异步执行，使用带追踪的客户端，自动追踪）
	go func() {
		jsonData, _ := json.Marshal(user)
		database.RedisClient.Set(ctx, cacheKey, string(jsonData), UserCacheTTL)
	}()

	return user, nil
}

// GetAllUsers 查询所有用户
func GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User

	// 使用带追踪的数据库客户端（自动追踪）
	err := database.DB.WithContext(ctx).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}
