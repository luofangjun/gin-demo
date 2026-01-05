package logic

import "time"

const (
	UserCacheKey = "user:%d"        // 用户缓存键格式
	UserCacheTTL = 30 * time.Minute // 用户缓存过期时间
)

