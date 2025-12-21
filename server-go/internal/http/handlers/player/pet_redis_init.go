package player

import (
	"context"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisClient "xiuxian/server-go/internal/redis"
)

// InitPetResourcesCache 在用户登录时初始化灵宠资源缓存
// 这个函数应该在用户认证后、返回响应前调用
func InitPetResourcesCache(ctx context.Context, userID uint) error {
	// 从数据库获取用户的灵宠精华数量
	var user models.User
	if err := db.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}

	// 同步到 Redis
	return redisClient.SyncPetResourcesToRedis(
		ctx,
		userID,
		int64(user.PetEssence),
	)
}

// SyncPetResourcesToDB 从 Redis 同步灵宠资源到数据库
// 这个函数应该在用户登出或定期任务中调用，确保 Redis 缓存与数据库同步
func SyncPetResourcesToDB(ctx context.Context, userID uint) error {
	// 从 Redis 获取灵宠资源
	resources, err := redisClient.GetPetResources(ctx, userID)
	if err != nil {
		// 如果 Redis 中没有，说明没有任何操作，无需同步
		return nil
	}

	// 更新数据库
	return db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"pet_essence": resources.PetEssence,
		}).Error
}
