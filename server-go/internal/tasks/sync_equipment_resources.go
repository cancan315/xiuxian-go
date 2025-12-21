package tasks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisc "xiuxian/server-go/internal/redis"

	"go.uber.org/zap"
)

var logger *zap.Logger

// InitTasks 初始化所有后台任务
func InitTasks(zapLogger *zap.Logger) {
	logger = zapLogger

	// ✅ 优化：改为每1秒检查一次，根据操作时间戳判断是否需要同步
	// 检查频率：1秒
	// 同步条件：用户5秒钟未操作
	StartEquipmentResourcesSyncTask(1 * time.Second)

	// 灵宠资源定期同步任务（同样改为1秒检查频率）
	StartPetResourcesSyncTask(1 * time.Second)

	logger.Info("后台同步任务已启动", zap.String("checkInterval", "1秒"), zap.String("syncCondition", "5秒未操作"))
}

// ============================================
// 装备资源定期同步任务
// ============================================

// StartEquipmentResourcesSyncTask 启动装备资源定期同步任务
// ✅ 优化：改为每1秒检查一次时间戳，根据是否超过5秒未操作来决定是否同步
func StartEquipmentResourcesSyncTask(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logger.Info("启动装备资源定期同步任务",
			zap.Duration("checkInterval", interval),
			zap.String("strategy", "基于操作时间戳的智能同步"),
			zap.Int("syncTimeoutSeconds", 5))

		for range ticker.C {
			syncAllEquipmentResources()
		}
	}()
}

// syncAllEquipmentResources 同步所有用户的装备资源缓存到数据库
// ✅ 优化：5秒未操作后，同步到DB并清理缓存
func syncAllEquipmentResources() {
	ctx := context.Background()

	// 扫描所有 Redis 中的装备资源缓存键
	// 键格式：user:USER_ID:equipment:resources
	pattern := "user:*:equipment:resources"

	var cursor uint64
	var keys []string

	// 使用 SCAN 遍历所有匹配的键
	// SCAN 是非阻塞的，不会锁定整个 Redis
	for {
		scanResult, nextCursor, err := redisc.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.Error("扫描 Redis 装备资源键失败", zap.Error(err))
			break
		}

		keys = append(keys, scanResult...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return
	}

	// ✅ 优化：只同步5秒内未操作的用户
	// 对每个键进行处理
	successCount := 0
	failedCount := 0
	skippedCount := 0
	const syncTimeoutSeconds int64 = 5 // 5秒内没有操作了，才执行同步

	for _, key := range keys {
		// 解析键获取 userID
		// 键格式：user:USER_ID:equipment:resources
		userID, err := parseUserIDFromKey(key)
		if err != nil {
			failedCount++
			continue
		}

		// ✅ 关键：检查该用户的最后操作时间戳 (user:{id}:equipment:last_operation_time)
		// 只有超过5秒未操作的用户才会被同步
		if !redisc.ShouldSyncEquipmentToDb(ctx, userID, syncTimeoutSeconds) {
			// 最近5秒内有操作，暂时不同步
			skippedCount++
			continue
		}

		// 获取 Redis 中的资源数据
		resources, err := redisc.GetEquipmentResources(ctx, userID)
		if err != nil {
			failedCount++
			continue
		}

		// 同步到数据库
		if err := db.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"reinforce_stones":  resources.ReinforceStones,
				"refinement_stones": resources.RefinementStones,
			}).Error; err != nil {

			failedCount++
			continue
		}

		// ✅ 优化：同步后清理装备资源缓存和时间戳
		if err := redisc.Client.Del(ctx, fmt.Sprintf(redisc.EquipmentResourceKeyFormat, userID)).Err(); err != nil {
			logger.Warn("清理装备资源缓存失败", zap.Uint("userID", userID), zap.Error(err))
		}

		// 清理时间戳键
		if err := redisc.Client.Del(ctx, fmt.Sprintf(redisc.EquipmentLastOperationTimeKeyFormat, userID)).Err(); err != nil {
			logger.Warn("清理时间戳键失败", zap.Uint("userID", userID), zap.Error(err))
		}

		successCount++
	}

	// 只在有实际操作时才输出日志
	if successCount > 0 || failedCount > 0 {
		logger.Info("装备资源同步完成",
			zap.Int("total", len(keys)),
			zap.Int("success", successCount),
			zap.Int("failed", failedCount),
			zap.Int("skipped", skippedCount))
	}
}

// ============================================
// 灵宠资源定期同步任务
// ============================================

// StartPetResourcesSyncTask 启动灵宠资源定期同步任务
// ✅ 优化：改为每1秒检查一次时间戳，根据是否超过5秒未操作来决定是否同步
func StartPetResourcesSyncTask(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logger.Info("启动灵宠资源定期同步任务",
			zap.Duration("checkInterval", interval),
			zap.String("strategy", "基于操作时间戳的智能同步"),
			zap.Int("syncTimeoutSeconds", 5))

		for range ticker.C {
			syncAllPetResources()
		}
	}()
}

// syncAllPetResources 同步所有用户的灵宠资源缓存到数据库
// ✅ 策略：5秒未操作后，同步到DB并清理缓存
func syncAllPetResources() {
	ctx := context.Background()

	// 扫描所有 Redis 中的灵宠资源缓存键
	// 键格式：user:USER_ID:pet:resources
	pattern := "user:*:pet:resources"

	var cursor uint64
	var keys []string

	// 使用 SCAN 遍历所有匹配的键
	for {
		scanResult, nextCursor, err := redisc.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.Error("扫描 Redis 灵宠资源键失败", zap.Error(err))
			break
		}

		keys = append(keys, scanResult...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return
	}

	// ✅ 优化：只同步5秒内未操作的用户
	// 对每个键进行处理
	successCount := 0
	failedCount := 0
	skippedCount := 0
	const syncTimeoutSeconds int64 = 5 // 5秒内没有操作了，才执行同步

	for _, key := range keys {
		// 解析键获取 userID
		// 键格式：user:USER_ID:pet:resources
		userID, err := parseUserIDFromKey(key)
		if err != nil {
			failedCount++
			continue
		}

		// ✅ 关键：检查该用户的最后操作时间戳 (user:{id}:pet:last_operation_time)
		// 只有超过5秒未操作的用户才会被同步
		if !redisc.ShouldSyncPetToDb(ctx, userID, syncTimeoutSeconds) {
			// 最近5秒内有操作，暂时不同步
			skippedCount++
			continue
		}

		// 获取 Redis 中的资源数据
		resources, err := redisc.GetPetResources(ctx, userID)
		if err != nil {
			failedCount++
			continue
		}

		// 同步到数据库
		if err := db.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{
				"pet_essence": resources.PetEssence,
			}).Error; err != nil {

			failedCount++
			continue
		}

		// ✅ 优化：同步后清理灵宠资源缓存和时间戳
		if err := redisc.Client.Del(ctx, fmt.Sprintf(redisc.PetResourceKeyFormat, userID)).Err(); err != nil {
			logger.Warn("清理灵宠资源缓存失败", zap.Uint("userID", userID), zap.Error(err))
		}

		// 清理时间戳键
		if err := redisc.Client.Del(ctx, fmt.Sprintf(redisc.PetLastOperationTimeKeyFormat, userID)).Err(); err != nil {
			logger.Warn("清理时间戳键失败", zap.Uint("userID", userID), zap.Error(err))
		}

		successCount++
	}

	// 只在有实际操作时才输出日志
	if successCount > 0 || failedCount > 0 {
		logger.Info("灵宠资源同步完成",
			zap.Int("total", len(keys)),
			zap.Int("success", successCount),
			zap.Int("failed", failedCount),
			zap.Int("skipped", skippedCount))
	}
}

// ============================================
// 辅助函数
// ============================================

// parseUserIDFromKey 从 Redis 键中解析用户 ID
// 支持的键格式：
//   - user:USER_ID:equipment:resources
//   - user:USER_ID:pet:resources
//   - user:USER_ID:...
func parseUserIDFromKey(key string) (uint, error) {
	// 使用 strings.Split 分割键
	// 键格式：user:USER_ID:xxx:...
	parts := strings.Split(key, ":")

	if len(parts) < 2 {
		return 0, fmt.Errorf("无效的键格式: %s", key)
	}

	// 第一个部分应该是 "user"
	if parts[0] != "user" {
		return 0, fmt.Errorf("键不是 user 类型: %s", key)
	}

	// 第二个部分是 userID
	id64, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("解析 userID 失败: %s, error: %w", parts[1], err)
	}

	return uint(id64), nil
}
