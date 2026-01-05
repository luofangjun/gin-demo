package logic

import (
	"context"
	"fmt"

	"gin-project/database"
	"gin-project/model"
)

// CreateUser 创建用户
func CreateUser(ctx context.Context, user *model.User) error {
	// 验证数据合法性
	if user.Name == "" || user.Email == "" {
		return fmt.Errorf("用户姓名和邮箱不能为空")
	}

	// 检查邮箱是否已存在（使用带追踪的数据库客户端，自动追踪）
	var existingUser model.User
	err := database.DB.WithContext(ctx).Where("email = ?", user.Email).First(&existingUser).Error
	if err == nil {
		// 用户已存在
		return fmt.Errorf("邮箱 %s 已存在", user.Email)
	}

	// 插入数据库（使用带追踪的数据库客户端，自动追踪）
	err = database.DB.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}

	// 清除相关的缓存（使用带追踪的 Redis 客户端，自动追踪）
	cacheKey := fmt.Sprintf(UserCacheKey, user.ID)
	database.RedisClient.Del(ctx, cacheKey)

	return nil
}

// UpdateUser 更新用户信息
func UpdateUser(ctx context.Context, user *model.User) error {
	// 验证数据合法性
	if user.ID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	// 更新数据库，但不更新CreatedAt字段（使用带追踪的数据库客户端，自动追踪）
	err := database.DB.WithContext(ctx).Model(&model.User{}).Select("name", "email", "age", "status", "updated_at").Where("id = ?", user.ID).Updates(user).Error
	if err != nil {
		return err
	}

	// 清除相关的缓存（使用带追踪的 Redis 客户端，自动追踪）
	cacheKey := fmt.Sprintf(UserCacheKey, user.ID)
	database.RedisClient.Del(ctx, cacheKey)

	return nil
}
