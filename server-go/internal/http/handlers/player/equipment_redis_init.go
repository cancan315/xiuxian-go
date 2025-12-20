package player

import (
	"context"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisClient "xiuxian/server-go/internal/redis"
)

// InitEquipmentResourcesCache 在用户登录时初始化装备资源缓存
// 这个函数应该在用户认证后、返回响应前调用
func InitEquipmentResourcesCache(ctx context.Context, userID uint) error {
	// 从数据库获取用户的强化石和洗练石数量
	var user models.User
	if err := db.DB.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}

	// 同步到 Redis
	return redisClient.SyncEquipmentResourcesToRedis(
		ctx,
		userID,
		int64(user.ReinforceStones),
		int64(user.RefinementStones),
	)
}

// SyncEquipmentResourcesToDB 从 Redis 同步装备资源到数据库
// 这个函数应该在用户登出或定期任务中调用，确保 Redis 缓存与数据库同步
func SyncEquipmentResourcesToDB(ctx context.Context, userID uint) error {
	// 从 Redis 获取装备资源
	resources, err := redisClient.GetEquipmentResources(ctx, userID)
	if err != nil {
		// 如果 Redis 中没有，说明没有任何操作，无需同步
		return nil
	}

	// 更新数据库
	return db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"reinforce_stones":  resources.ReinforceStones,
			"refinement_stones": resources.RefinementStones,
		}).Error
}
