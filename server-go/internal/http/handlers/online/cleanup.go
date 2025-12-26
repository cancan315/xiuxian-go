package online

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisc "xiuxian/server-go/internal/redis"
)

// StartHeartbeatMonitor 启动心跳监控，定期清理超时的玩家数据
// 应该在服务启动时调用，运行一个后台goroutine
func StartHeartbeatMonitor(logger *zap.Logger) {
	go func() {
		ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
		defer ticker.Stop()

		for range ticker.C {
			checkAndCleanupTimeoutPlayers(logger)
		}
	}()

	logger.Info("心跳监控已启动，将每5秒检查一次超时玩家")
}

// checkAndCleanupTimeoutPlayers 检查并清理心跳超时的玩家
func checkAndCleanupTimeoutPlayers(logger *zap.Logger) {
	rdb := redisc.Client
	ctx := context.Background()

	// 获取所有在线玩家
	playerIDs, err := rdb.SMembers(ctx, "server:online:players").Result()
	if err != nil {
		logger.Error("获取在线玩家列表失败", zap.Error(err))
		return
	}

	currentTime := time.Now().UnixMilli()
	heartbeatTimeoutMs := int64(600 * 1000) // 10分钟超时

	for _, playerIDStr := range playerIDs {
		key := "player:online:" + playerIDStr

		// 获取玩家的最后心跳时间
		data, err := rdb.HGetAll(ctx, key).Result()
		if err != nil {
			logger.Warn("获取玩家数据失败", zap.String("playerID", playerIDStr), zap.Error(err))
			continue
		}

		if len(data) == 0 {
			continue
		}

		lastHeartbeatStr, ok := data["lastHeartbeat"]
		if !ok {
			continue
		}

		lastHeartbeat, err := strconv.ParseInt(lastHeartbeatStr, 10, 64)
		if err != nil {
			logger.Warn("解析心跳时间失败", zap.String("playerID", playerIDStr), zap.Error(err))
			continue
		}

		// 检查是否超时
		if currentTime-lastHeartbeat > heartbeatTimeoutMs {
			// 转换playerID为uint
			parsedID, err := strconv.ParseUint(playerIDStr, 10, 32)
			if err != nil {
				logger.Warn("解析玩家ID失败", zap.String("playerID", playerIDStr), zap.Error(err))
				continue
			}
			playerID := uint(parsedID)

			logger.Warn("玩家心跳超时，进行清理",
				zap.String("playerID", playerIDStr),
				zap.Int64("lastHeartbeat", lastHeartbeat),
				zap.Int64("currentTime", currentTime),
				zap.Int64("timeDiff", currentTime-lastHeartbeat))

			// ✅ 新增：同步装备资源缓存到数据库
			if err := syncEquipmentResourcesOnTimeout(ctx, playerID, logger); err != nil {
				logger.Warn("同步装备资源到数据库失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// ✅ 新增：清理装备强化和洗练缓存
			if err := cleanupEquipmentCache(playerID, logger); err != nil {
				logger.Warn("清理装备缓存失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// ✅ 新增：同步灵宠资源缓存到数据库
			if err := syncPetResourcesOnTimeout(ctx, playerID, logger); err != nil {
				logger.Warn("同步灵宠资源到数据库失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// ✅ 新增：清理灵宠升级升星缓存
			if err := cleanupPetCache(playerID, logger); err != nil {
				logger.Warn("清理灵宠缓存失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// ✅ 新增：同步灵宠资源缓存到数据库
			if err := syncPetResourcesOnTimeout(ctx, playerID, logger); err != nil {
				logger.Warn("同步灵宠资源到数据库失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// ✅ 新增：清理灵宠升级升星缓存
			if err := cleanupPetCache(playerID, logger); err != nil {
				logger.Warn("清理灵宠缓存失败",
					zap.Uint("userID", playerID),
					zap.Error(err))
				// 不中断清理流程
			}

			// 删除在线状态
			if err := rdb.Del(ctx, key).Err(); err != nil {
				logger.Error("删除超时玩家在线状态失败",
					zap.String("playerID", playerIDStr),
					zap.Error(err))
			}

			// 从在线集合中移除
			if err := rdb.SRem(ctx, "server:online:players", playerIDStr).Err(); err != nil {
				logger.Error("从在线集合中移除超时玩家失败",
					zap.String("playerID", playerIDStr),
					zap.Error(err))
			}

			// 清理战斗数据
			cleanupDungeonData(playerID, rdb, logger)
		}
	}
}

// CleanupDungeonDataByRedis 通过Redis直接清理战斗数据（用于其他服务调用）
func CleanupDungeonDataByRedis(playerID uint, logger *zap.Logger) error {
	rdb := redisc.Client
	ctx := context.Background()

	battleStatusKey := fmt.Sprintf("dungeon:battle:status:%d", playerID)
	roundDataKey := fmt.Sprintf("dungeon:battle:round:%d", playerID)

	if err := rdb.Del(ctx, battleStatusKey, roundDataKey).Err(); err != nil {
		logger.Error("清理战斗数据失败",
			zap.Uint("userID", playerID),
			zap.Error(err))
		return err
	}

	logger.Info("已清理玩家战斗数据",
		zap.Uint("userID", playerID),
		zap.String("battleStatusKey", battleStatusKey),
		zap.String("roundDataKey", roundDataKey))

	return nil
}

// syncEquipmentResourcesOnTimeout 心跳超时时同步装备资源缓存到数据库
// 将 Redis 中的强化石和洗练石数量同步回数据库
func syncEquipmentResourcesOnTimeout(ctx context.Context, playerID uint, logger *zap.Logger) error {
	// 从 Redis 获取装备资源缓存
	resources, err := redisc.GetEquipmentResources(ctx, playerID)
	if err != nil {
		// 缓存不存在，说明没有进行过装备操作，无需同步
		return nil
	}

	// 更新数据库
	if err := db.DB.Model(&models.User{}).
		Where("id = ?", playerID).
		Updates(map[string]interface{}{
			"reinforce_stones":  resources.ReinforceStones,
			"refinement_stones": resources.RefinementStones,
		}).Error; err != nil {
		logger.Error("同步装备资源到数据库失败",
			zap.Uint("userID", playerID),
			zap.Error(err))
		return err
	}

	logger.Info("已同步装备资源到数据库",
		zap.Uint("userID", playerID),
		zap.Int64("reinforceStones", resources.ReinforceStones),
		zap.Int64("refinementStones", resources.RefinementStones))

	return nil
}

// cleanupEquipmentCache 清理玩家的装备缓存（强化和洗练）
// 清理以下缓存键：
// - user:{userID}:equipment:resources (装备资源)
// - user:{userID}:equipment:{equipID}:enhance:lock (强化锁)
// - user:{userID}:equipment:{equipID}:reforge:lock (洗练锁)
// - user:{userID}:equipment:list (装备列表缓存)
func cleanupEquipmentCache(playerID uint, logger *zap.Logger) error {
	ctx := context.Background()
	rdb := redisc.Client

	// 清理装备资源缓存
	equipmentResourceKey := fmt.Sprintf("user:%d:equipment:resources", playerID)
	if err := rdb.Del(ctx, equipmentResourceKey).Err(); err != nil {
		logger.Warn("删除装备资源缓存失败",
			zap.Uint("userID", playerID),
			zap.String("key", equipmentResourceKey),
			zap.Error(err))
	}

	// 清理装备列表缓存
	equipmentListKey := fmt.Sprintf("user:%d:equipment:list", playerID)
	if err := rdb.Del(ctx, equipmentListKey).Err(); err != nil {
		logger.Warn("删除装备列表缓存失败",
			zap.Uint("userID", playerID),
			zap.String("key", equipmentListKey),
			zap.Error(err))
	}

	// 使用 SCAN 扫描并清理所有该用户的装备相关缓存键
	// 包括：user:{userID}:equipment:*
	pattern := fmt.Sprintf("user:%d:equipment:*", playerID)
	var cursor uint64
	keysToDelete := []string{}

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.Warn("扫描装备缓存键失败",
				zap.Uint("userID", playerID),
				zap.Error(err))
			break
		}

		keysToDelete = append(keysToDelete, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	// 批量删除所有扫描到的键
	if len(keysToDelete) > 0 {
		if err := rdb.Del(ctx, keysToDelete...).Err(); err != nil {
			logger.Warn("批量删除装备缓存失败",
				zap.Uint("userID", playerID),
				zap.Int("keysCount", len(keysToDelete)),
				zap.Error(err))
			return err
		}

		logger.Info("已清理装备缓存",
			zap.Uint("userID", playerID),
			zap.Int("keysCount", len(keysToDelete)))
	}

	return nil
}

// syncPetResourcesOnTimeout 心跳超时时同步灵宠资源缓存到数据库
// 将 Redis 中的灵宠精华数量同步回数据库
func syncPetResourcesOnTimeout(ctx context.Context, playerID uint, logger *zap.Logger) error {
	// 从 Redis 获取灵宠资源缓存
	resources, err := redisc.GetPetResources(ctx, playerID)
	if err != nil {
		// 缓存不存在，说明没有进行过灵宠操作，无需同步
		return nil
	}

	// 更新数据库
	if err := db.DB.Model(&models.User{}).
		Where("id = ?", playerID).
		Updates(map[string]interface{}{
			"pet_essence": resources.PetEssence,
		}).Error; err != nil {
		logger.Error("同步灵宠资源到数据库失败",
			zap.Uint("userID", playerID),
			zap.Error(err))
		return err
	}

	logger.Info("已同步灵宠资源到数据库",
		zap.Uint("userID", playerID),
		zap.Int64("petEssence", resources.PetEssence))

	return nil
}

// cleanupPetCache 清理玩家的灵宠缓存（升级和升星）
// 清理以下缓存键：
// - user:{userID}:pet:resources (灵宠精华资源)
// - user:{userID}:pet:{petID}:upgrade:lock (升级锁)
// - user:{userID}:pet:{petID}:evolve:lock (升星锁)
// - user:{userID}:pet:{petID} (灵宠数据缓存)
// - user:{userID}:pet:list (灵宠列表缓存)
func cleanupPetCache(playerID uint, logger *zap.Logger) error {
	ctx := context.Background()
	rdb := redisc.Client

	// 清理灵宠资源缓存
	petResourceKey := fmt.Sprintf("user:%d:pet:resources", playerID)
	if err := rdb.Del(ctx, petResourceKey).Err(); err != nil {
		logger.Warn("删除灵宠资源缓存失败",
			zap.Uint("userID", playerID),
			zap.String("key", petResourceKey),
			zap.Error(err))
	}

	// 清理灵宠列表缓存
	petListKey := fmt.Sprintf("user:%d:pet:list", playerID)
	if err := rdb.Del(ctx, petListKey).Err(); err != nil {
		logger.Warn("删除灵宠列表缓存失败",
			zap.Uint("userID", playerID),
			zap.String("key", petListKey),
			zap.Error(err))
	}

	// 使用 SCAN 扫描并清理所有该用户的灵宠相关缓存键
	// 包括：user:{userID}:pet:*
	pattern := fmt.Sprintf("user:%d:pet:*", playerID)
	var cursor uint64
	keysToDelete := []string{}

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.Warn("扫描灵宠缓存键失败",
				zap.Uint("userID", playerID),
				zap.Error(err))
			break
		}

		keysToDelete = append(keysToDelete, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	// 批量删除所有扫描到的键
	if len(keysToDelete) > 0 {
		if err := rdb.Del(ctx, keysToDelete...).Err(); err != nil {
			logger.Warn("批量删除灵宠缓存失败",
				zap.Uint("userID", playerID),
				zap.Int("keysCount", len(keysToDelete)),
				zap.Error(err))
			return err
		}

		logger.Info("已清理灵宠缓存",
			zap.Uint("userID", playerID),
			zap.Int("keysCount", len(keysToDelete)))
	}

	return nil
}
