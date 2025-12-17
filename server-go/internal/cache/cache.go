package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"xiuxian/server-go/internal/redis"

	"go.uber.org/zap"
)

// 缓存键常量
const (
	KeyLeaderboardTop100 = "leaderboard:top100"
	KeyPlayerSpirit      = "player:spirit:%d"
	KeyPlayerCultivation = "player:cultivation:%d"
)

// 缓存 TTL 常量
const (
	TTLLeaderboard  = 60 * time.Second  // 排行榜 60 秒
	TTLCultivation  = 120 * time.Second // 修炼数据 120 秒
	TTLPlayerSpirit = 0                 // 灵力值由后台任务管理，不设置 TTL
)

// GetLeaderboardFromCache 从缓存获取排行榜数据
func GetLeaderboardFromCache(logger *zap.Logger) ([]map[string]interface{}, error) {
	val, err := redis.Client.Get(redis.Ctx, KeyLeaderboardTop100).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // 缓存不存在
		}
		logger.Warn("获取排行榜缓存失败", zap.Error(err))
		return nil, err
	}

	var list []map[string]interface{}
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		logger.Warn("解析排行榜缓存失败", zap.Error(err))
		return nil, err
	}

	logger.Debug("从缓存获取排行榜成功", zap.Int("count", len(list)))
	return list, nil
}

// SetLeaderboardCache 将排行榜数据存储到缓存
func SetLeaderboardCache(data []map[string]interface{}, logger *zap.Logger) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("序列化排行榜失败", zap.Error(err))
		return err
	}

	err = redis.Client.Set(redis.Ctx, KeyLeaderboardTop100, string(jsonData), TTLLeaderboard).Err()
	if err != nil {
		logger.Error("存储排行榜缓存失败", zap.Error(err))
		return err
	}

	logger.Debug("排行榜缓存已更新", zap.Int("count", len(data)))
	return nil
}

// DeleteLeaderboardCache 删除排行榜缓存
func DeleteLeaderboardCache(logger *zap.Logger) error {
	err := redis.Client.Del(redis.Ctx, KeyLeaderboardTop100).Err()
	if err != nil {
		logger.Warn("删除排行榜缓存失败", zap.Error(err))
		return err
	}
	logger.Debug("排行榜缓存已删除")
	return nil
}

// GetPlayerSpiritFromCache 从缓存获取玩家灵力值
func GetPlayerSpiritFromCache(userID uint, logger *zap.Logger) (float64, error) {
	key := fmt.Sprintf(KeyPlayerSpirit, userID)
	val, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil // 缓存不存在
		}
		logger.Warn("获取灵力缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return 0, err
	}

	var spirit float64
	if err := json.Unmarshal([]byte(val), &spirit); err != nil {
		logger.Warn("解析灵力缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return 0, err
	}

	logger.Debug("从缓存获取灵力值成功", zap.Uint("userID", userID), zap.Float64("spirit", spirit))
	return spirit, nil
}

// GetCultivationDataFromCache 从缓存获取修炼数据
func GetCultivationDataFromCache(userID uint, logger *zap.Logger) (map[string]interface{}, error) {
	key := fmt.Sprintf(KeyPlayerCultivation, userID)
	val, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil // 缓存不存在
		}
		logger.Warn("获取修炼数据缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		logger.Warn("解析修炼数据缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, err
	}

	logger.Debug("从缓存获取修炼数据成功", zap.Uint("userID", userID))
	return data, nil
}

// SetCultivationDataCache 将修炼数据存储到缓存
func SetCultivationDataCache(userID uint, data interface{}, logger *zap.Logger) error {
	key := fmt.Sprintf(KeyPlayerCultivation, userID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("序列化修炼数据失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}

	err = redis.Client.Set(redis.Ctx, key, string(jsonData), TTLCultivation).Err()
	if err != nil {
		logger.Error("存储修炼数据缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}

	logger.Debug("修炼数据缓存已更新", zap.Uint("userID", userID))
	return nil
}

// DeleteCultivationDataCache 删除修炼数据缓存
func DeleteCultivationDataCache(userID uint, logger *zap.Logger) error {
	key := fmt.Sprintf(KeyPlayerCultivation, userID)
	err := redis.Client.Del(redis.Ctx, key).Err()
	if err != nil {
		logger.Warn("删除修炼数据缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}
	logger.Debug("修炼数据缓存已删除", zap.Uint("userID", userID))
	return nil
}

// DeletePlayerSpiritCache 删除玩家灵力缓存
func DeletePlayerSpiritCache(userID uint, logger *zap.Logger) error {
	key := fmt.Sprintf(KeyPlayerSpirit, userID)
	err := redis.Client.Del(redis.Ctx, key).Err()
	if err != nil {
		logger.Warn("删除灵力缓存失败", zap.Uint("userID", userID), zap.Error(err))
		return err
	}
	logger.Debug("灵力缓存已删除", zap.Uint("userID", userID))
	return nil
}
