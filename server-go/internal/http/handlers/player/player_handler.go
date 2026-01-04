package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/alchemy"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/exploration"
	"xiuxian/server-go/internal/models"
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

	// 记录入参
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

	zap.S().Infof("接收到玩家数据增量更新请求: userId=%d, itemCount=%d, petCount=%d, herbCount=%d, pillCount=%d",
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

	zap.S().Infof("玩家数据增量更新完成: %d", userID)
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

	zap.S().Infof("成功卖出 %d 个物品, user=%d", len(req.ItemIDs), userID)
	c.JSON(http.StatusOK, gin.H{"message": "灵宠卖出成功"})
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

	// ✅ 新增：先查询灵宠信息，计算灵石奖励
	var pets []models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id IN ? OR id IN ?)", userID, req.PetIDs, req.PetIDs).Find(&pets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	if len(pets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "未找到指定的灵宠"})
		return
	}

	// ✅ 新增：根据灵宠品质计算灵石奖励（和单个出售一致）
	rarity2spiritStoneMap := map[string]int{
		"mythic":    500,
		"legendary": 300,
		"epic":      150,
		"rare":      100,
		"uncommon":  80,
		"common":    50,
	}

	totalSpiritStones := 0
	for _, pet := range pets {
		reward := rarity2spiritStoneMap[pet.Rarity]
		if reward == 0 {
			reward = 100 // 默认值
		}
		totalSpiritStones += reward
	}

	// 卖出灵宠
	if err := db.DB.Where("user_id = ? AND (pet_id IN ? OR id IN ?)", userID, req.PetIDs, req.PetIDs).
		Delete(&models.Pet{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：更新玩家灵石（只返还灵石）
	if err := db.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("spirit_stones", gorm.Expr("spirit_stones + ?", totalSpiritStones)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	zap.S().Infof("成功卖出 %d 个灵宠, 获得%d灵石, user=%d", len(req.PetIDs), totalSpiritStones, userID)
	c.JSON(http.StatusOK, gin.H{
		"success":      true, // ✅ 新增：返回成功标记
		"message":      "灵宠卖出成功",
		"deletedCount": len(pets),
		"spiritStones": totalSpiritStones, // ✅ 返回灵石奖励
	})
}

// ChangePlayerName 对应 POST /api/player/change-name
func ChangePlayerName(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var req struct {
		NewName string `json:"newName" binding:"required,min=1,max=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	// 获取当前用户信息以检查道号修改次数和灵石数量
	var user models.User
	err := db.DB.First(&user, userID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取用户信息失败", "error": err.Error()})
		return
	}

	// 计算修改道号所需灵石
	var spiritStoneCost int
	if user.NameChangeCount == 0 {
		// 第一次修改免费
		spiritStoneCost = 0
	} else {
		// 后续修改需要100灵石
		spiritStoneCost = 100
	}

	// 检查灵石是否足够
	if user.SpiritStones < spiritStoneCost {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("灵石不足！修改道号需要%d颗灵石", spiritStoneCost)})
		return
	}

	// 检查新名称是否已存在
	var existingUser models.User
	err = db.DB.Where("player_name = ? AND id != ?", req.NewName, userID).First(&existingUser).Error
	if err == nil {
		// 找到了同名用户
		c.JSON(http.StatusConflict, gin.H{"message": "该道号已被其他玩家使用"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 更新玩家道号和扣除灵石
	updates := map[string]interface{}{
		"player_name": req.NewName,
	}

	// 如果不是第一次修改，则扣除灵石并增加修改次数
	if user.NameChangeCount > 0 {
		updates["spirit_stones"] = user.SpiritStones - spiritStoneCost
		updates["name_change_count"] = user.NameChangeCount + 1
	} else {
		// 第一次修改，只增加修改次数
		updates["name_change_count"] = user.NameChangeCount + 1
	}

	err = db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "修改道号失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "道号修改成功",
		"newName":         req.NewName,
		"spiritStoneCost": spiritStoneCost,
	})
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

// ApplySpiritGain 应用灵力增长（改进版）
func ApplySpiritGain(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	spiritManagerVal, exists := c.Get("spirit_manager")
	if !exists {
		zapLogger.Error("灵力增长管理器未初始化")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	spiritManager := spiritManagerVal.(*spirit.SpiritGrowManager)

	// ✅ 改进：后端从Redis直接获取灵力增长值，不依赖前端提交的值
	spiritGain, err := spiritManager.GetPlayerSpiritGain(userID)
	if err != nil {
		zapLogger.Error("获取灵力增长值失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	// 如果灵力增长为0，直接返回
	if spiritGain <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"message":    "当前无灵力增长",
			"spiritGain": 0,
		})
		return
	}

	// 获取玩家当前灵力
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		zapLogger.Error("获取玩家信息失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	// 计算新的灵力值
	newSpirit := user.Spirit + spiritGain
	newSpirit = math.Round(newSpirit*10) / 10

	// 更新数据库
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("spirit", newSpirit).Error; err != nil {
		zapLogger.Error("更新灵力失败", zap.Uint("userID", userID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		return
	}

	// 清空Redis中的灵力增长缓存
	if err := spiritManager.ClearPlayerSpiritGain(userID); err != nil {
		zapLogger.Error("清空灵力增长缓存失败", zap.Uint("userID", userID))
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "灵力增长应用成功",
		"oldSpirit":  user.Spirit,
		"newSpirit":  newSpirit,
		"spiritGain": spiritGain,
	})
}

// ✅ 修复版本：InitializePlayerAttributesOnLogin 登录时初始化玩家属性
func InitializePlayerAttributesOnLogin(user *models.User, zapLogger *zap.Logger) error {
	userID := user.ID
	level := user.Level

	zapLogger.Info("[登录初始化] 开始初始化玩家属性",
		zap.Uint("userID", userID),
		zap.Int("level", level))

	// 步骤1：获取玩家当前穿戴的所有装备（保存装备ID）
	var equippedEquipments []models.Equipment
	if err := db.DB.Where("user_id = ? AND equipped = ?", userID, true).
		Find(&equippedEquipments).Error; err != nil {
		zapLogger.Warn("[登录初始化] 查询已穿戴装备失败",
			zap.Uint("userID", userID),
			zap.Error(err))
	}

	// 提取装备ID
	var equippedEquipmentIDs []string
	for _, equipment := range equippedEquipments {
		equippedEquipmentIDs = append(equippedEquipmentIDs, equipment.ID)
	}

	zapLogger.Info("[登录初始化] 发现已穿戴装备",
		zap.Uint("userID", userID),
		zap.Int("count", len(equippedEquipmentIDs)))

	// 步骤2：卸下所有装备
	if err := db.DB.Model(&models.Equipment{}).
		Where("user_id = ? AND equipped = ?", userID, true).
		Updates(map[string]interface{}{
			"equipped": false,
			"slot":     nil,
		}).Error; err != nil {
		zapLogger.Error("[登录初始化] 卸下装备失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	if len(equippedEquipmentIDs) > 0 {
		zapLogger.Info("[登录初始化] 成功卸下所有装备",
			zap.Uint("userID", userID),
			zap.Int("count", len(equippedEquipmentIDs)))
	}

	// 步骤3：获取玩家当前出战的所有灵宠（保存灵宠ID）
	var activePets []models.Pet
	if err := db.DB.Where("user_id = ? AND is_active = ?", userID, true).
		Find(&activePets).Error; err != nil {
		zapLogger.Warn("[登录初始化] 查询出战灵宠失败",
			zap.Uint("userID", userID),
			zap.Error(err))
	}

	// 提取灵宠ID
	var activePetIDs []string
	for _, pet := range activePets {
		activePetIDs = append(activePetIDs, pet.ID)
	}

	zapLogger.Info("[登录初始化] 发现出战灵宠",
		zap.Uint("userID", userID),
		zap.Int("count", len(activePetIDs)))

	// 步骤4：召回所有灵宠
	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false).Error; err != nil {
		zapLogger.Error("[登录初始化] 召回灵宠失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	if len(activePetIDs) > 0 {
		zapLogger.Info("[登录初始化] 成功召回所有灵宠",
			zap.Uint("userID", userID),
			zap.Int("count", len(activePetIDs)))
	}

	// 步骤5：计算基础属性（不含装备和灵宠加成）
	baseAttrs := calculateBaseAttributesByLevel(level)
	spiritRate := calculateSpiritRateByLevel(level)
	baseAttrs["spiritRate"] = spiritRate
	baseAttrs["cultivationRate"] = 1.0

	// 步骤6：更新BaseAttributes到用户对象
	user.BaseAttributes = toJSON(baseAttrs)

	// 步骤7：初始化属性值（从基础属性开始）
	baseAttrsMap := jsonToFloatMap(user.BaseAttributes)
	combatAttrsMap := jsonToFloatMap(user.CombatAttributes)
	combatResMap := jsonToFloatMap(user.CombatResistance)
	specialAttrsMap := jsonToFloatMap(user.SpecialAttributes)

	// 步骤8：重新穿戴已保存的装备
	for _, equipmentID := range equippedEquipmentIDs {
		var equipment models.Equipment
		if err := db.DB.Where("id = ? AND user_id = ?", equipmentID, userID).First(&equipment).Error; err != nil {
			zapLogger.Warn("[登录初始化] 查询装备失败",
				zap.Uint("userID", userID),
				zap.String("equipmentID", equipmentID),
				zap.Error(err))
			continue
		}

		// 应用装备属性
		equipStats := jsonToFloatMap(equipment.Stats)
		attrMgr := NewAttributeManager(baseAttrsMap, combatAttrsMap, combatResMap, specialAttrsMap)
		attrMgr.ApplyEquipmentStats(equipStats)

		// 更新属性映射
		baseAttrsMap = attrMgr.BaseAttrs
		combatAttrsMap = attrMgr.CombatAttrs
		combatResMap = attrMgr.CombatRes
		specialAttrsMap = attrMgr.SpecialAttrs

		// ✅ 重新标记装备为已装备
		if err := db.DB.Model(&models.Equipment{}).
			Where("id = ?", equipmentID).
			Update("equipped", true).Error; err != nil {
			zapLogger.Warn("[登录初始化] 重新装备失败",
				zap.Uint("userID", userID),
				zap.String("equipmentID", equipmentID),
				zap.Error(err))
		} else {
			zapLogger.Debug("[登录初始化] 成功重新穿戴装备",
				zap.Uint("userID", userID),
				zap.String("equipmentID", equipmentID))
		}
	}

	// 步骤9：重新出战已保存的灵宠
	for _, petID := range activePetIDs {
		var pet models.Pet
		if err := db.DB.Where("id = ? AND user_id = ?", petID, userID).First(&pet).Error; err != nil {
			zapLogger.Warn("[登录初始化] 查询灵宠失败",
				zap.Uint("userID", userID),
				zap.String("petID", petID),
				zap.Error(err))
			continue
		}

		// 应用灵宠属性
		petCombat := jsonToFloatMap(pet.CombatAttributes)
		attrMgr := NewAttributeManager(baseAttrsMap, combatAttrsMap, combatResMap, specialAttrsMap)
		attrMgr.ApplyPetBonuses(&pet, petCombat)

		// 更新属性映射
		baseAttrsMap = attrMgr.BaseAttrs
		combatAttrsMap = attrMgr.CombatAttrs
		combatResMap = attrMgr.CombatRes
		specialAttrsMap = attrMgr.SpecialAttrs

		// ✅ 重新标记灵宠为已出战
		if err := db.DB.Model(&models.Pet{}).
			Where("id = ?", petID).
			Update("is_active", true).Error; err != nil {
			zapLogger.Warn("[登录初始化] 重新出战灵宠失败",
				zap.Uint("userID", userID),
				zap.String("petID", petID),
				zap.Error(err))
		} else {
			zapLogger.Debug("[登录初始化] 成功重新出战灵宠",
				zap.Uint("userID", userID),
				zap.String("petID", petID))
		}
	}

	// 步骤10：保存最终属性到用户对象
	user.BaseAttributes = toJSON(baseAttrsMap)
	user.CombatAttributes = toJSON(combatAttrsMap)
	user.CombatResistance = toJSON(combatResMap)
	user.SpecialAttributes = toJSON(specialAttrsMap)

	// 步骤11：更新数据库
	if err := db.DB.Model(user).Updates(map[string]interface{}{
		"base_attributes":    user.BaseAttributes,
		"combat_attributes":  user.CombatAttributes,
		"combat_resistance":  user.CombatResistance,
		"special_attributes": user.SpecialAttributes,
	}).Error; err != nil {
		zapLogger.Error("[登录初始化] 保存属性到数据库失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	zapLogger.Info("[登录初始化] 玩家属性初始化完成",
		zap.Uint("userID", userID),
		zap.Int("level", level),
		zap.Float64("spiritRate", spiritRate),
		zap.Int("reequippedCount", len(equippedEquipmentIDs)),
		zap.Int("reactivatedCount", len(activePetIDs)))

	return nil
}

// ✅ 新增：calculateSpiritRateByLevel 计算基于等级的灵力倍率
// 公式：spiritRate = 1.0 * (1.2)^(Level-1)
// 每突破一次（等级+1），灵力倍率乘以1.2
func calculateSpiritRateByLevel(level int) float64 {
	if level < 1 {
		level = 1
	}
	// 基础速度是1.0，每升一级就乘以1.2
	spiritRate := 1.0 * math.Pow(1.2, float64(level-1))
	// 保留两位小数
	return math.Round(spiritRate*100) / 100
}

// ✅ 新增：calculateBaseAttributesByLevel 计算基于等级的基础属性
// 公式：
// speed = 10 * Level
// attack = 10 * Level
// health = 100 * Level
// defense = 5 * Level
func calculateBaseAttributesByLevel(level int) map[string]interface{} {
	return map[string]interface{}{
		"speed":   float64(10 * level),
		"attack":  float64(10 * level),
		"health":  float64(100 * level),
		"defense": float64(5 * level),
	}
}

// GetHerbsPaginated 对应 GET /api/player/herbs 获取玩家灵草分页列表
func GetHerbsPaginated(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	// 解析分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "id")    // 排序字段
	order := c.DefaultQuery("order", "asc") // 排序顺序

	// 记录入参
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("GetHerbsPaginated 入参",
		zap.Uint("userID", userID),
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.String("sort", sort),
		zap.String("order", order))

	// 验证排序字段，防止SQL注入
	allowedSortFields := map[string]bool{
		"id": true, "herb_id": true, "name": true, "count": true,
	}
	if !allowedSortFields[sort] {
		sort = "id"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// 获取总数
	var total int64
	if err := db.DB.Model(&models.Herb{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		zapLogger.Error("获取灵草总数失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 计算分页
	var pageInt, pageSizeInt int
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil || pageInt < 1 {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err != nil || pageSizeInt < 1 {
		pageSizeInt = 20
	}
	if pageSizeInt > 100 {
		pageSizeInt = 100 // 最大100条
	}

	offset := (pageInt - 1) * pageSizeInt

	// 查询灵草列表
	var herbs []models.Herb
	query := db.DB.Where("user_id = ?", userID)
	if err := query.Order(fmt.Sprintf("%s %s", sort, order)).
		Offset(offset).
		Limit(pageSizeInt).
		Find(&herbs).Error; err != nil {
		zapLogger.Error("获取灵草列表失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 获取灵草配置信息，组装完整数据
	herbConfigs := exploration.HerbConfigs
	enrichedHerbs := make([]gin.H, len(herbs))
	for i, herb := range herbs {
		enrichedHerb := gin.H{
			"id":     herb.ID,
			"userId": herb.UserID,
			"herbId": herb.HerbID,
			"name":   herb.Name,
			"count":  herb.Count,
		}

		// 从配置中获取额外信息
		for _, config := range herbConfigs {
			if config.ID == herb.HerbID {
				enrichedHerb["description"] = config.Description
				enrichedHerb["baseValue"] = config.BaseValue
				break
			}
		}

		// 添加品质信息
		quality := herb.Quality
		if quality == "" {
			quality = "common" // 默认品质
		}
		enrichedHerb["quality"] = quality

		// 获取品质配置信息
		qualityConfigs := exploration.HerbQualities
		if qualityInfo, ok := qualityConfigs[quality]; ok {
			enrichedHerb["qualityName"] = qualityInfo["name"]
			enrichedHerb["qualityValue"] = qualityInfo["value"]
		} else {
			enrichedHerb["qualityName"] = "未知"
			enrichedHerb["qualityValue"] = 1
		}

		enrichedHerbs[i] = enrichedHerb
	}

	// 构建返回结构
	totalPages := (total + int64(pageSizeInt) - 1) / int64(pageSizeInt)
	data := gin.H{
		"herbs": enrichedHerbs,
		"pagination": gin.H{
			"page":       pageInt,
			"pageSize":   pageSizeInt,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	zapLogger.Info("GetHerbsPaginated 出参",
		zap.Uint("userID", userID),
		zap.Int64("total", total),
		zap.Int("herbCount", len(enrichedHerbs)))

	c.JSON(http.StatusOK, data)
}

// GetPillsPaginated 对应 GET /api/player/pills 获取玩家丹药分页列表
func GetPillsPaginated(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	// 解析分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "id")    // 排序字段
	order := c.DefaultQuery("order", "asc") // 排序顺序

	// 记录入参
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("GetPillsPaginated 入参",
		zap.Uint("userID", userID),
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.String("sort", sort),
		zap.String("order", order))

	// 验证排序字段，防止SQL注入
	allowedSortFields := map[string]bool{
		"id": true, "pill_id": true, "name": true,
	}
	if !allowedSortFields[sort] {
		sort = "id"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// 获取总数
	var total int64
	if err := db.DB.Model(&models.Pill{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		zapLogger.Error("获取丹药总数失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 计算分页
	var pageInt, pageSizeInt int
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil || pageInt < 1 {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err != nil || pageSizeInt < 1 {
		pageSizeInt = 20
	}
	if pageSizeInt > 100 {
		pageSizeInt = 100 // 最大100条
	}

	offset := (pageInt - 1) * pageSizeInt

	// 查询丹药列表
	var pills []models.Pill
	query := db.DB.Where("user_id = ?", userID)
	if err := query.Order(fmt.Sprintf("%s %s", sort, order)).
		Offset(offset).
		Limit(pageSizeInt).
		Find(&pills).Error; err != nil {
		zapLogger.Error("获取丹药列表失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	totalPages := (total + int64(pageSizeInt) - 1) / int64(pageSizeInt)
	data := gin.H{
		"pills": pills,
		"pagination": gin.H{
			"page":       pageInt,
			"pageSize":   pageSizeInt,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	zapLogger.Info("GetPillsPaginated 出参",
		zap.Uint("userID", userID),
		zap.Int64("total", total),
		zap.Int("pillCount", len(pills)))

	c.JSON(http.StatusOK, data)
}

// GetFormulasPaginated 对应 GET /api/player/formulas 获取玩家丹方分页列表
func GetFormulasPaginated(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	// 解析分页参数
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "20")
	sort := c.DefaultQuery("sort", "id")    // 排序字段
	order := c.DefaultQuery("order", "asc") // 排序顺序

	// 记录入参
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("GetFormulasPaginated 入参",
		zap.Uint("userID", userID),
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.String("sort", sort),
		zap.String("order", order))

	// 验证排序字段，防止SQL注入
	allowedSortFields := map[string]bool{
		"id": true, "recipe_id": true, "count": true,
	}
	if !allowedSortFields[sort] {
		sort = "id"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// 获取总数
	var total int64
	if err := db.DB.Model(&models.PillFragment{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		zapLogger.Error("获取丹方总数失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 计算分页
	var pageInt, pageSizeInt int
	if _, err := fmt.Sscanf(page, "%d", &pageInt); err != nil || pageInt < 1 {
		pageInt = 1
	}
	if _, err := fmt.Sscanf(pageSize, "%d", &pageSizeInt); err != nil || pageSizeInt < 1 {
		pageSizeInt = 20
	}
	if pageSizeInt > 100 {
		pageSizeInt = 100 // 最大100条
	}

	offset := (pageInt - 1) * pageSizeInt

	// 查询丹方残页列表
	var fragments []models.PillFragment
	query := db.DB.Where("user_id = ?", userID)
	if err := query.Order(fmt.Sprintf("%s %s", sort, order)).
		Offset(offset).
		Limit(pageSizeInt).
		Find(&fragments).Error; err != nil {
		zapLogger.Error("获取丹方列表失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 构建返回结构
	totalPages := (total + int64(pageSizeInt) - 1) / int64(pageSizeInt)
	data := gin.H{
		"formulas": fragments,
		"pagination": gin.H{
			"page":       pageInt,
			"pageSize":   pageSizeInt,
			"total":      total,
			"totalPages": totalPages,
		},
	}

	zapLogger.Info("GetFormulasPaginated 出参",
		zap.Uint("userID", userID),
		zap.Int64("total", total),
		zap.Int("formulaCount", len(fragments)))

	c.JSON(http.StatusOK, data)
}

// ConsumePill 对应 POST /api/player/pills/:id/consume 服用丹药
func ConsumePill(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	pillID := c.Param("id")
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("ConsumePill 入参",
		zap.Uint("userID", userID),
		zap.String("pillID", pillID))

	// 查询丹药是否存在
	var pill models.Pill
	if err := db.DB.Where("id = ? AND user_id = ?", pillID, userID).First(&pill).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "丹药不存在"})
		} else {
			zapLogger.Error("查询丹药失败",
				zap.Uint("userID", userID),
				zap.String("pillID", pillID),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误"})
		}
		return
	}

	// 查询玩家信息，用于获取等级和计算效果
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		zapLogger.Error("查询玩家信息失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询玩家信息失败"})
		return
	}

	// 解析丹药效果
	var effectData map[string]interface{}
	var effectType string
	var baseValue float64
	if err := json.Unmarshal(pill.Effect, &effectData); err == nil {
		if t, ok := effectData["type"].(string); ok {
			effectType = t
		}
		if v, ok := effectData["value"].(float64); ok {
			baseValue = v
		}
	}

	// 根据丹药 ID 查找对应的计算公式，计算效果值
	actualValue := alchemy.CalculatePillEffect(pill.PillID, baseValue, user.Level)

	zapLogger.Info("丹药效果计算",
		zap.String("pillID", pill.PillID),
		zap.String("effectType", effectType),
		zap.Float64("baseValue", baseValue),
		zap.Int("playerLevel", user.Level),
		zap.Float64("actualValue", actualValue))

	// 根据效果类型更新对应属性
	switch effectType {
	case "spirit":
		user.Spirit += actualValue
	case "cultivation":
		user.Cultivation += actualValue
	case "attributeAttack":
		// 更新基础属性中的攻击
		updateBaseAttributeAttack(&user, actualValue)
	case "duJieRate":
		// 更新渡劫成功率 (duJieRate 存储于 BaseAttributes JSON 中)
		updateBaseAttributeDuJieRate(&user, actualValue)
	}

	// 保存玩家更新
	if err := db.DB.Save(&user).Error; err != nil {
		zapLogger.Error("保存玩家属性失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "保存玩家属性失败"})
		return
	}

	// 删除丹药记录
	if err := db.DB.Delete(&pill).Error; err != nil {
		zapLogger.Error("删除丹药失败",
			zap.Uint("userID", userID),
			zap.String("pillID", pillID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除丹药失败"})
		return
	}

	// 更新用户统计数据（从 BaseAttributes JSON 中更新 pillsConsumed）
	// ✅ 使用 jsonToMap 保留所有字段类型（包括 unlockedRealms 数组和 duJieRate 等）
	baseAttrs := jsonToMap(user.BaseAttributes)
	if baseAttrs == nil {
		baseAttrs = make(map[string]interface{})
	}
	if v, ok := baseAttrs["pillsConsumed"].(float64); ok {
		baseAttrs["pillsConsumed"] = v + 1
	} else {
		baseAttrs["pillsConsumed"] = float64(1)
	}
	user.BaseAttributes = toJSON(baseAttrs)

	if err := db.DB.Model(&user).Update("base_attributes", user.BaseAttributes).Error; err != nil {
		zapLogger.Error("更新丹药统计失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新统计失败"})
		return
	}

	zapLogger.Info("ConsumePill 出参",
		zap.Uint("userID", userID),
		zap.String("pillID", pillID),
		zap.String("pillName", pill.Name),
		zap.String("effectType", effectType),
		zap.Float64("actualValue", actualValue))

	// 返回丹药信息和效果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("成功服用%s，增加%s %.0f", pill.Name, effectTypeToName(effectType), actualValue),
		"data": gin.H{
			"pillId":            pill.ID,
			"pillName":          pill.Name,
			"effect":            pill.Effect,
			"description":       pill.Description,
			"effectType":        effectType,
			"actualValue":       actualValue,
			"playerSpirit":      user.Spirit,
			"playerCultivation": user.Cultivation,
		},
	})
}
