package player

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/datatypes"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
	"xiuxian/server-go/internal/spirit"
)

// helper 获取当前用户 ID
func currentUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := val.(uint)
	return id, ok
}

// assembleFullPlayerData 组合用户+物品+宠物+草药+丹药数据
func assembleFullPlayerData(userID uint) (gin.H, error) {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	// 获取用户物品
	var items []models.Item
	if err := db.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	// 获取用户装备
	var equipments []models.Equipment
	if err := db.DB.Where("user_id = ?", userID).Find(&equipments).Error; err != nil {
		return nil, err
	}
	// 获取用户宠物
	var pets []models.Pet
	if err := db.DB.Where("user_id = ?", userID).Find(&pets).Error; err != nil {
		return nil, err
	}
	// 获取用户草药
	var herbs []models.Herb
	if err := db.DB.Where("user_id = ?", userID).Find(&herbs).Error; err != nil {
		return nil, err
	}
	// 获取用户丹药
	var pills []models.Pill
	if err := db.DB.Where("user_id = ?", userID).Find(&pills).Error; err != nil {
		return nil, err
	}

	// 将装备、草药和丹药合并到物品列表中
	for _, equipment := range equipments {
		items = append(items, models.Item{
			ID:        equipment.ID,
			UserID:    equipment.UserID,
			ItemID:    equipment.EquipmentID,
			Name:      equipment.Name,
			Type:      equipment.Type,
			Slot:      equipment.Slot,    // 添加 Slot 字段，装备槽位
			Stats:     equipment.Stats,   // 添加 Stats 字段，装备属性
			Quality:   equipment.Quality, // 添加 Quality 字段，装备品质
			Equipped:  equipment.Equipped,
			EquipType: equipment.EquipType, // 添加 EquipType 字段，装备类型
		})
	}

	for _, herb := range herbs {
		items = append(items, models.Item{
			ID:      fmt.Sprintf("herb_%d", herb.ID),
			UserID:  herb.UserID,
			ItemID:  herb.HerbID,
			Name:    herb.Name,
			Type:    "herb",
			Details: datatypes.JSON(fmt.Sprintf("{\"count\": %d}", herb.Count)),
		})
	}

	for _, pill := range pills {
		items = append(items, models.Item{
			ID:      fmt.Sprintf("pill_%d", pill.ID),
			UserID:  pill.UserID,
			ItemID:  pill.PillID,
			Name:    pill.Name,
			Type:    "pill",
			Details: pill.Effect,
			// Description: pill.Description, // 如果需要可以添加额外字段
		})
	}

	// artifacts 目前 Node 版固定为空数组
	artifacts := []interface{}{}

	data := gin.H{
		"user":      user,
		"items":     items,
		"pets":      pets,
		"artifacts": artifacts,
	}

	return data, nil
}

// GetPlayerData 对应 GET /api/player/data 获取玩家完整数据
func GetPlayerData(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	// �记录入参
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("GetPlayerData 入参",
		zap.Uint("userID", userID))

	data, err := assembleFullPlayerData(userID)
	if err != nil {
		// log.Printf("get player data failed: %v", err)
		zapLogger.Error("get player data failed",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 记录出参
	items, _ := data["items"].([]models.Item)
	pets, _ := data["pets"].([]models.Pet)
	zapLogger.Info("GetPlayerData 出参",
		zap.Uint("userID", userID),
		zap.Int("itemsCount", len(items)),
		zap.Int("petsCount", len(pets)),
		zap.Any("responseData", data))

	c.JSON(http.StatusOK, data)
}

// GetPlayerSpirit 对应 GET /api/player/spirit
func GetPlayerSpirit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 从数据库查询灵力值
	var user models.User
	if err := db.DB.Select("spirit").First(&user, userID).Error; err != nil {
		zapLogger.Error("获取灵力值失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"spirit": user.Spirit})
}

// UpdateSpirit 对应 PUT /api/player/spirit
func UpdateSpirit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		Spirit float64 `json:"spirit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("spirit", req.Spirit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "灵力值更新成功"})
}

// GetLeaderboard 对应 GET /api/player/leaderboard
// 与 Node playerController.getLeaderboard 对齐
func GetLeaderboard(c *gin.Context) {
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	cacheKey := "leaderboard:top100"
	cacheTTL := 120 * time.Second

	// 尝试从Redis缓存读取
	cachedData, err := redis.Client.Get(redis.Ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中，解析并返回
		var list []struct {
			ID           uint   `gorm:"column:id" json:"id"`
			PlayerName   string `gorm:"column:player_name" json:"playerName"`
			Level        int    `gorm:"column:level" json:"level"`
			Realm        string `gorm:"column:realm" json:"realm"`
			SpiritStones int    `gorm:"column:spirit_stones" json:"spiritStones"`
		}

		if err := json.Unmarshal([]byte(cachedData), &list); err == nil {
			zapLogger.Info("排行榜缓存命中", zap.String("cacheKey", cacheKey))
			c.JSON(http.StatusOK, list)
			return
		}
	}

	// 缓存未命中或解析失败，从数据库查询排行榜数据
	var list []struct {
		ID           uint   `gorm:"column:id" json:"id"`
		PlayerName   string `gorm:"column:player_name" json:"playerName"`
		Level        int    `gorm:"column:level" json:"level"`
		Realm        string `gorm:"column:realm" json:"realm"`
		SpiritStones int    `gorm:"column:spirit_stones" json:"spiritStones"`
	}

	if err := db.DB.Model(&models.User{}).
		Select("id, player_name, level, realm, spirit_stones").
		Order("level DESC").
		Order("spirit_stones DESC").
		Limit(100).
		Scan(&list).Error; err != nil {
		zapLogger.Error("获取排行榜失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 将结果写入Redis缓存
	if data, err := json.Marshal(list); err == nil {
		if err := redis.Client.Set(redis.Ctx, cacheKey, string(data), cacheTTL).Err(); err != nil {
			zapLogger.Warn("排行榜缓存写入失败", zap.Error(err))
			// 忽略缓存失败，继续返回数据
		}
	}

	c.JSON(http.StatusOK, list)
}

// UpdatePlayerData 对应 PATCH /api/player/data
// 目前先实现一个记录日志的 stub，返回与 Node 版相同的成功消息，
// 后续再逐步迁移完整的增量更新逻辑。
func UpdatePlayerData(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		User  map[string]interface{}   `json:"user"`
		Items []map[string]interface{} `json:"items"`
		Pets  []map[string]interface{} `json:"pets"`
		Herbs []map[string]interface{} `json:"herbs"`
		Pills []map[string]interface{} `json:"pills"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	log.Printf("接收到玩家数据增量更新请求: userId=%d, itemCount=%d, petCount=%d, herbCount=%d, pillCount=%d",
		userID, len(req.Items), len(req.Pets), len(req.Herbs), len(req.Pills))

	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": tx.Error.Error()})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新用户基础数据（只更新提供的字段）
	if req.User != nil {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(req.User).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// 处理物品/装备
	for _, item := range req.Items {
		if err := upsertItemOrEquipment(tx, userID, item); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// 处理宠物
	for _, pet := range req.Pets {
		if err := upsertPet(tx, userID, pet); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// 处理灵草
	for _, herb := range req.Herbs {
		if err := upsertHerb(tx, userID, herb); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// 处理丹药
	for _, pill := range req.Pills {
		if err := upsertPill(tx, userID, pill); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	log.Printf("玩家数据增量更新完成: %d", userID)
	c.JSON(http.StatusOK, gin.H{"message": "数据增量更新成功"})
}

// DeleteItems 对应 DELETE /api/player/items
func DeleteItems(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		ItemIDs []string `json:"itemIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的物品ID列表"})
		return
	}

	if err := db.DB.Where("user_id = ? AND (equipment_id IN ? OR id IN ?)", userID, req.ItemIDs, req.ItemIDs).
		Delete(&models.Equipment{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	log.Printf("成功删除 %d 个物品, user=%d", len(req.ItemIDs), userID)
	c.JSON(http.StatusOK, gin.H{"message": "物品删除成功"})
}

// DeletePets 对应 DELETE /api/player/pets
func DeletePets(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		PetIDs []string `json:"petIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || len(req.PetIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的宠物ID列表"})
		return
	}

	if err := db.DB.Where("user_id = ? AND (pet_id IN ? OR id IN ?)", userID, req.PetIDs, req.PetIDs).
		Delete(&models.Pet{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	log.Printf("成功删除 %d 个灵宠, user=%d", len(req.PetIDs), userID)
	c.JSON(http.StatusOK, gin.H{"message": "灵宠删除成功"})
}

// GetPlayerSpiritGain 对应 GET /api/player/spirit/gain
// 获取玩家累积的灵力增长量（从Redis中读取）
func GetPlayerSpiritGain(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 从上下文获取灵力增长管理器
	spiritManagerVal, exists := c.Get("spirit_manager")
	if !exists {
		zapLogger.Error("灵力增长管理器未初始化")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	spiritManager := spiritManagerVal.(*spirit.SpiritGrowManager)

	// 获取玩家在Redis中的灵力增长量
	spiritGain, err := spiritManager.GetPlayerSpiritGain(userID)
	if err != nil {
		zapLogger.Error("获取灵力增长量失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	zapLogger.Info("获取玩家灵力增长量",
		zap.Uint("userID", userID),
		zap.Float64("spiritGain", spiritGain))

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"spiritGain": spiritGain,
		"message":    "获取灵力增长量成功",
	})
}

// ApplySpiritGain 对应 POST /api/player/spirit/apply-gain
// 将灵力增长量应用到玩家（写入数据库并清空Redis）
func ApplySpiritGain(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		SpiritGain float64 `json:"spiritGain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 从上下文获取灵力增长管理器
	spiritManagerVal, exists := c.Get("spirit_manager")
	if !exists {
		zapLogger.Error("灵力增长管理器未初始化")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	spiritManager := spiritManagerVal.(*spirit.SpiritGrowManager)

	// 获取玩家当前灵力
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		zapLogger.Error("获取玩家信息失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 计算新的灵力值
	newSpirit := user.Spirit + req.SpiritGain

	newSpirit = math.Round(newSpirit*10) / 10

	// 更新数据库
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("spirit", newSpirit).Error; err != nil {
		zapLogger.Error("更新灵力失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 清空Redis中的灵力增长缓存
	if err := spiritManager.ClearPlayerSpiritGain(userID); err != nil {
		zapLogger.Error("清空灵力增长缓存失败", zap.Uint("userID", userID), zap.Error(err))
		// 不中断流程，已经成功写入数据库了
	}

	zapLogger.Info("应用玩家灵力增长",
		zap.Uint("userID", userID),
		zap.Float64("spiritGain", req.SpiritGain),
		zap.Float64("oldSpirit", user.Spirit),
		zap.Float64("newSpirit", newSpirit))

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "灵力增长应用成功",
		"oldSpirit":  user.Spirit,
		"newSpirit":  newSpirit,
		"spiritGain": req.SpiritGain,
	})
}
