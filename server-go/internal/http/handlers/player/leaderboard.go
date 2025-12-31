package player

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
)

// getQualityName 根据品质代码获取中文名称
func getQualityName(qualityCode string) string {
	if config, exists := models.EquipmentQualityConfigs[qualityCode]; exists {
		return config.Name
	}
	return qualityCode
}

// getRarityName 根据稀有度代码获取中文名称
func getRarityName(rarityCode string) string {
	if config, exists := models.PetQualityConfigs[rarityCode]; exists {
		return config.Name
	}
	return rarityCode
}

// GetLeaderboard 对应 GET /api/player/leaderboard - 境界排行
func GetLeaderboard(c *gin.Context) {
	leaderboardType := c.Param("type")
	if leaderboardType == "" {
		leaderboardType = c.Query("type")
	}
	if leaderboardType == "" {
		leaderboardType = "realm"
	}

	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("[排行榜] 获取排行榜请求", zap.String("排行榜类型", leaderboardType))

	switch leaderboardType {
	case "realm":
		getRealmLeaderboard(c, zapLogger)
	case "spiritStones":
		getSpiritStonesLeaderboard(c, zapLogger)
	case "equipment":
		getEquipmentLeaderboard(c, zapLogger)
	case "pets":
		getPetsLeaderboard(c, zapLogger)
	case "duel": // ✅ 新增：斗法排行榜
		getDuelLeaderboard(c, zapLogger)
	default:
		getRealmLeaderboard(c, zapLogger)
	}
}

// getRealmLeaderboard 获取境界排行 - 按玩家境界排序
func getRealmLeaderboard(c *gin.Context, zapLogger *zap.Logger) {
	cacheKey := "leaderboard:realm:top100"
	cacheTTL := 2 * time.Minute

	// 尝试从Redis缓存读取
	if cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result(); err == nil {
		var result []interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			zapLogger.Info("[排行榜] 从缓存返回境界排行", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, result)
			return
		}
	}

	// 从数据库查询排行榜数据
	var list []struct {
		ID         uint   `gorm:"column:id"`
		PlayerName string `gorm:"column:player_name"`
		Realm      string `gorm:"column:realm"`
		Level      int    `gorm:"column:level"`
	}

	if err := db.DB.Model(&models.User{}).
		Select("id, player_name, realm, level").
		Where("users.player_name <> ?", "无名修士").
		Order("level DESC").
		Limit(100).
		Scan(&list).Error; err != nil {
		zapLogger.Error("获取境界排行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	result := make([]gin.H, len(list))
	for i, item := range list {
		result[i] = gin.H{
			"id":         item.ID,
			"playerName": item.PlayerName,
			"realm":      item.Realm,
			"level":      item.Level,
		}
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(result); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("[排行榜] 缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, result)
}

// getSpiritStonesLeaderboard 获取灵石排行 - 按灵石数量排序
func getSpiritStonesLeaderboard(c *gin.Context, zapLogger *zap.Logger) {
	cacheKey := "leaderboard:spiritStones:top100"
	cacheTTL := 2 * time.Minute

	// 尝试从Redis缓存读取
	if cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result(); err == nil {
		var result []interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			zapLogger.Info("[排行榜] 从缓存返回灵石排行", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, result)
			return
		}
	}

	zapLogger.Info("[排行榜] 开始获取灵石排行")
	// 从数据库查询排行榜数据
	var list []struct {
		ID           uint   `gorm:"column:id"`
		PlayerName   string `gorm:"column:player_name"`
		SpiritStones int    `gorm:"column:spirit_stones"`
		Realm        string `gorm:"column:realm"`
	}

	if err := db.DB.Model(&models.User{}).
		Select("id, player_name, spirit_stones, realm").
		Where("users.player_name <> ?", "无名修士").
		Order("spirit_stones DESC").
		Limit(100).
		Scan(&list).Error; err != nil {
		zapLogger.Error("获取灵石排行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	result := make([]gin.H, len(list))
	for i, item := range list {
		result[i] = gin.H{
			"id":           item.ID,
			"playerName":   item.PlayerName,
			"spiritStones": item.SpiritStones,
			"realm":        item.Realm,
		}
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(result); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("[排行榜] 缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	zapLogger.Info("[排行榜] 灵石排行数据获取成功", zap.Int("count", len(result)))
	c.JSON(http.StatusOK, result)
}

// getEquipmentLeaderboard 获取装备排行 - 按稀有度排序，同稀有度按强化等级排序
func getEquipmentLeaderboard(c *gin.Context, zapLogger *zap.Logger) {
	cacheKey := "leaderboard:equipment:top100"
	cacheTTL := 2 * time.Minute

	// 尝试从Redis缓存读取
	if cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result(); err == nil {
		var result []interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			zapLogger.Info("[排行榜] 从缓存返回装备排行", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, result)
			return
		}
	}

	zapLogger.Info("[排行榜] 开始获取装备排行")
	// 获取装备排行 - 显示单个装备的排行
	type EquipmentItem struct {
		ID           string `gorm:"column:id"`
		UserID       uint   `gorm:"column:user_id"`
		PlayerName   string `gorm:"column:player_name"`
		EquipmentID  string `gorm:"column:equipment_id"`
		Name         string `gorm:"column:name"`
		Quality      string `gorm:"column:quality"`
		EnhanceLevel int    `gorm:"column:enhance_level"`
	}

	var items []EquipmentItem

	// 查询已装备的装备，按稀有度和强化等级排序
	if err := db.DB.Table("equipment").
		Joins("JOIN users ON equipment.user_id = users.id").
		Select(`
			equipment.id,
			equipment.user_id,
			users.player_name,
			equipment.equipment_id,
			equipment.name,
			equipment.quality,
			equipment.enhance_level
		`).
		Where("equipment.equipped = ?", true).
		Where("users.player_name <> ?", "无名修士").
		Order("CASE equipment.quality WHEN 'mythic' THEN 1 WHEN 'legendary' THEN 2 WHEN 'epic' THEN 3 WHEN 'rare' THEN 4 WHEN 'uncommon' THEN 5 ELSE 6 END").
		Order("equipment.enhance_level DESC").
		Limit(100).
		Scan(&items).Error; err != nil {
		zapLogger.Error("获取装备排行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	result := make([]gin.H, len(items))
	for i, item := range items {
		result[i] = gin.H{
			"id":           item.ID,
			"playerName":   item.PlayerName,
			"name":         item.Name,
			"quality":      getQualityName(item.Quality),
			"enhanceLevel": item.EnhanceLevel,
		}
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(result); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("[排行榜] 缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	zapLogger.Info("[排行榜] 装备排行数据获取成功", zap.Int("count", len(result)))
	c.JSON(http.StatusOK, result)
}

// getPetsLeaderboard 获取灵宠排行 - 按稀有度排序，同稀有度按星级排序，同星级按等级排序
func getPetsLeaderboard(c *gin.Context, zapLogger *zap.Logger) {
	cacheKey := "leaderboard:pets:top100"
	cacheTTL := 2 * time.Minute

	// 尝试从Redis缓存读取
	if cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result(); err == nil {
		var result []interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			zapLogger.Info("[排行榜] 从缓存返回灵宠排行", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, result)
			return
		}
	}

	zapLogger.Info("[排行榜] 开始获取灵宠排行")
	// 获取灵宠排行 - 显示单个灵宠的排行
	type PetItem struct {
		ID         string `gorm:"column:id"`
		UserID     uint   `gorm:"column:user_id"`
		PlayerName string `gorm:"column:player_name"`
		PetID      string `gorm:"column:pet_id"`
		Name       string `gorm:"column:name"`
		Rarity     string `gorm:"column:rarity"`
		Star       int    `gorm:"column:star"`
		Level      int    `gorm:"column:level"`
	}

	var items []PetItem

	// 查询已激活的灵宠，按稀有度、星级和等级排序
	if err := db.DB.Table("pets").
		Joins("JOIN users ON pets.user_id = users.id").
		Select(`
			pets.id,
			pets.user_id,
			users.player_name,
			pets.pet_id,
			pets.name,
			pets.rarity,
			pets.star,
			pets.level
		`).
		Where("pets.is_active = ?", true).
		Where("users.player_name <> ?", "无名修士").
		Order("CASE pets.rarity WHEN 'mythic' THEN 1 WHEN 'legendary' THEN 2 WHEN 'epic' THEN 3 WHEN 'rare' THEN 4 WHEN 'uncommon' THEN 5 ELSE 6 END").
		Order("pets.star DESC").
		Order("pets.level DESC").
		Limit(100).
		Scan(&items).Error; err != nil {
		zapLogger.Error("获取灵宠排行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	result := make([]gin.H, len(items))
	for i, item := range items {
		result[i] = gin.H{
			"id":         item.ID,
			"playerName": item.PlayerName,
			"name":       item.Name,
			"rarity":     getRarityName(item.Rarity),
			"star":       item.Star,
			"level":      item.Level,
		}
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(result); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("[排行榜] 缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	zapLogger.Info("[排行榜] 灵宠排行数据获取成功", zap.Int("count", len(result)))
	c.JSON(http.StatusOK, result)
}

// getDuelLeaderboard 获取斗法排行 - 按胜场数和胜率排序
// ✅ 新增：斗法排行榜实现
func getDuelLeaderboard(c *gin.Context, zapLogger *zap.Logger) {
	cacheKey := "leaderboard:duel:top100"
	cacheTTL := 2 * time.Minute

	// 尝试从Redis缓存读取
	if cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result(); err == nil {
		var result []interface{}
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			zapLogger.Info("[排行榜] 从缓存返回斗法排行", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, result)
			return
		}
	}

	zapLogger.Info("[排行榜] 开始获取斗法排行")
	// 获取斗法排行 - 按胜场数和胜率排序
	type DuelItem struct {
		ID          uint   `gorm:"column:id"`
		PlayerName  string `gorm:"column:player_name"`
		TotalBattle int    `gorm:"column:total_battle"`
		Wins        int    `gorm:"column:wins"`
		Losses      int    `gorm:"column:losses"`
		WinRate     int    `gorm:"column:win_rate"`
	}

	var items []DuelItem

	// 查询斗法统计，按胜场数和胜率排序
	if err := db.DB.Table("users").
		Joins(`LEFT JOIN (
			SELECT 
				player_id,
				COUNT(*) as total_battle,
				SUM(CASE WHEN result = '胜利' THEN 1 ELSE 0 END) as wins,
				SUM(CASE WHEN result = '失败' THEN 1 ELSE 0 END) as losses,
				CASE 
					WHEN COUNT(*) > 0 THEN CAST(SUM(CASE WHEN result = '胜利' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) AS INT)
					ELSE 0
				END as win_rate
			FROM battle_records
			GROUP BY player_id
		) duel_stats ON users.id = duel_stats.player_id`).
		Select(`
			users.id,
			users.player_name,
			COALESCE(duel_stats.total_battle, 0) as total_battle,
			COALESCE(duel_stats.wins, 0) as wins,
			COALESCE(duel_stats.losses, 0) as losses,
			COALESCE(duel_stats.win_rate, 0) as win_rate
		`).
		Where("duel_stats.total_battle > 0").
		Where("users.player_name <> ?", "无名修士").
		Order("duel_stats.wins DESC, duel_stats.win_rate DESC, duel_stats.total_battle DESC").
		Limit(100).
		Find(&items).Error; err != nil {
		zapLogger.Error("[排行榜] 查询斗法排行失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"playerName":  item.PlayerName,
			"totalBattle": item.TotalBattle,
			"wins":        item.Wins,
			"losses":      item.Losses,
			"winRate":     item.WinRate,
		})
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(result); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("[排行榜] 缓存写入失败", zap.String("cacheKey", cacheKey), zap.Error(err))
		}
	}

	zapLogger.Info("[排行榜] 斗法排行数据获取成功", zap.Int("count", len(result)))
	c.JSON(http.StatusOK, result)
}

// ClearLeaderboardCache 对应 POST /api/admin/leaderboard/clear-cache
// 清除排行榜缓存，用于修复缓存数据格式问题
func ClearLeaderboardCache(c *gin.Context) {
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 定义要清除的缓存键
	cacheKeys := []string{
		"leaderboard:realm:top100",
		"leaderboard:spiritStones:top100",
		"leaderboard:equipment:top100",
		"leaderboard:pets:top100",
		"leaderboard:duel:top100", // ✅ 新增：斗法排行榜缓存
	}

	// 清除所有排行榜缓存
	var clearedCount int = 0
	for _, key := range cacheKeys {
		if err := redis.Client.Del(redis.Ctx, key).Err(); err != nil {
			zapLogger.Warn("清除排行榜缓存失败", zap.String("key", key), zap.Error(err))
		} else {
			clearedCount++
			zapLogger.Info("已清除排行榜缓存", zap.String("key", key))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "排行榜缓存已清除",
		"clearedCount": clearedCount,
		"totalKeys":    len(cacheKeys),
	})
}
