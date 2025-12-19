package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisClient "xiuxian/server-go/internal/redis"
)

// GetItemDetails 获取物品详情
// 对应 GET /api/player/inventory/item/:itemId
func GetItemDetails(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	itemID := c.Param("itemId")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "物品未找到",
		})
		return
	}

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "物品未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"item":    equipment,
	})
}

// GetPlayerEquipment 获取玩家装备列表
// 对应 GET /api/player/equipment
// 支持查询参数: type/equip_type(装备类型), quality(品质), equipped(是否已装备)
func GetPlayerEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 获取zap logger并记录入参日志
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// �记录入参信息
	zapLogger.Info("查询装备列表入参",
		zap.Uint("userID", userID),
		zap.String("type", c.Query("type")),
		zap.String("equip_type", c.Query("equip_type")),
		zap.String("quality", c.Query("quality")),
		zap.String("equipped", c.Query("equipped")))

	// 构建查询条件
	query := db.DB.Model(&models.Equipment{}).Where("user_id = ?", userID)

	// 挔装备类型过滤
	typeParam := c.Query("type")
	if typeParam != "" {
		query = query.Where("equip_type = ?", typeParam)
	} else if equip_type := c.Query("equip_type"); equip_type != "" {
		query = query.Where("equip_type = ?", equip_type)
	}

	// 按品质过滤
	if quality := c.Query("quality"); quality != "" {
		query = query.Where("quality = ?", quality)
	}

	// 按装备状态过滤
	equipped := c.Query("equipped")
	if equipped != "" {
		query = query.Where("equipped = ?", equipped == "true")
	}

	// 执行查询
	var equipment []models.Equipment
	if err := query.Find(&equipment).Error; err != nil {
		zapLogger.Error("装备列表查询数据库失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	// 如果指定了 equip_type 参数，进一步过滤结果
	// 确保返回的装备确实是指定类型的（处理数据库中可能存在的NULL值情况）
	equipTypeFilter := c.Query("equip_type")
	if equipTypeFilter != "" {
		filtered := make([]models.Equipment, 0)
		for _, equip := range equipment {
			// 确保 EquipType 不为 nil 并且等于请求的类型
			if equip.EquipType != nil && *equip.EquipType == equipTypeFilter {
				filtered = append(filtered, equip)
			}
		}
		equipment = filtered
	} else if typeParam != "" {
		// 如果使用的是 type 参数，也进行同样的过滤
		filtered := make([]models.Equipment, 0)
		for _, equip := range equipment {
			// 确保 EquipType 不为 nil 并且等于请求的类型
			if equip.EquipType != nil && *equip.EquipType == typeParam {
				filtered = append(filtered, equip)
			}
		}
		equipment = filtered
	}

	// 记录出参信息和详细装备属性
	zapLogger.Info("GetPlayerEquipment响应",
		zap.Uint("userID", userID),
		zap.Int("equipmentCount", len(equipment)),
		zap.String("query", c.Request.URL.RawQuery))

	// 记录每个装备的详细属性用于调试
	for _, equip := range equipment {
		zapLogger.Debug("装备详情",
			zap.Uint("userID", userID),
			zap.String("equipmentID", equip.ID),
			zap.String("equipmentName", equip.Name),
			zap.String("equipType", func() string {
				if equip.EquipType != nil {
					return *equip.EquipType
				}
				return "nil"
			}()),
			zap.String("slot", func() string {
				if equip.Slot != nil {
					return *equip.Slot
				}
				return "nil"
			}()),
			zap.Bool("equipped", equip.Equipped),
			zap.Int("enhanceLevel", equip.EnhanceLevel),
			zap.String("quality", equip.Quality),
			zap.Any("stats", string(equip.Stats)))
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"equipment": equipment,
	})
}

// GetEquipmentDetails 获取装备详情
// 对应 GET /api/player/equipment/details/:id
func GetEquipmentDetails(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "装备未找到",
		})
		return
	}

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "装备未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	// 记录返回给前端的数据
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	zapLogger.Info("GetEquipmentDetails 出参",
		zap.String("equipmentId", id),
		zap.Any("responseData", equipment))

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"equipment": equipment,
	})
}

// EnhanceEquipment 装备强化
// 对应 POST /api/player/equipment/:id/enhance
func EnhanceEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var req struct {
		ReinforceStones int `json:"reinforceStones"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	// ✅ 新增：获取logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	userIDStr := fmt.Sprintf("%d", userID)
	lockKey := fmt.Sprintf("user:%s:enhance_lock", userIDStr)

	// ✅ 新增：检查是否已有强化在进行中（防并发）
	lockVal, err := redisClient.Client.Get(c, lockKey).Result()
	if err == nil && lockVal != "" {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "装备强化正在进行中，请稍候",
		})
		return
	}

	// ✅ 新增：设置强化锁
	redisClient.Client.Set(c, lockKey, time.Now().Unix(), 10*time.Second)
	defer func() {
		// ✅ 新增：释放锁
		redisClient.Client.Del(c, lockKey)
	}()

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 获取用户信息
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：记录强化前状态
	zapLogger.Info("装备强化开始",
		zap.Uint("userID", userID),
		zap.String("equipmentID", equipment.ID),
		zap.String("equipmentName", equipment.Name),
		zap.Int("currentEnhanceLevel", equipment.EnhanceLevel),
		zap.Int("userReinforceStones", user.ReinforceStones))

	// 解析装备属性
	stats := map[string]float64{}
	if len(equipment.Stats) > 0 {
		_ = json.Unmarshal(equipment.Stats, &stats)
	}

	// 计算强化成本
	currentLevel := equipment.EnhanceLevel
	cost := 10 * (currentLevel + 1)

	// 检查强化石是否足够
	if user.ReinforceStones < cost {
		c.JSON(http.StatusOK, gin.H{
			"success":  false,
			"message":  "强化石不足",
			"cost":     cost,
			"oldStats": stats,
			"newStats": stats,
		})
		return
	}

	// 保存原始属性用于返回
	oldStats := map[string]float64{}
	for k, v := range stats {
		oldStats[k] = v
	}

	// ✅ 新增：第一步 - 如果装备已装备，先卸下
	if equipment.Equipped {
		zapLogger.Info("装备强化前先卸下装备",
			zap.String("equipmentID", equipment.ID),
			zap.Bool("equipped", equipment.Equipped))

		// 获取装备属性信息用于卸下
		equipStats := jsonToFloatMap(equipment.Stats)

		// 解析用户属性
		baseAttrs := jsonToFloatMap(user.BaseAttributes)
		combatAttrs := jsonToFloatMap(user.CombatAttributes)
		combatRes := jsonToFloatMap(user.CombatResistance)
		specialAttrs := jsonToFloatMap(user.SpecialAttributes)

		// 处理灵宠属性（如果有出战灵宠）
		var activePet models.Pet
		hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil
		if hasActivePet {
			petCombat := jsonToFloatMap(activePet.CombatAttributes)
			removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
		}

		// 移除装备属性
		removeEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

		// 重新应用灵宠加成
		if hasActivePet {
			activePetFresh := models.Pet{}
			db.DB.First(&activePetFresh, activePet.ID)
			petCombat := jsonToFloatMap(activePetFresh.CombatAttributes)
			applyPetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePetFresh, petCombat)
		}

		// 卸下装备
		equipment.Equipped = false
		equipment.Slot = nil
		if err := db.DB.Save(&equipment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}

		// 更新用户属性
		updates := map[string]interface{}{
			"base_attributes":    toJSON(baseAttrs),
			"combat_attributes":  toJSON(combatAttrs),
			"combat_resistance":  toJSON(combatRes),
			"special_attributes": toJSON(specialAttrs),
		}
		if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// ✅ 新增：第二步 - 执行强化逻辑
	// 强化属性（每个属性增加10%）
	for k, v := range stats {
		stats[k] = v * 1.1
	}

	// 更新装备等级和境界要求
	equipment.EnhanceLevel = currentLevel + 1
	equipment.Stats = toJSON(stats)

	newRequiredRealm := (equipment.EnhanceLevel / 10) + 1
	equipment.RequiredRealm = newRequiredRealm

	// 保存装备变更
	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 扣除用户强化石
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones - ?", cost)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：第三步 - 强化后自动穿戴装备
	zapLogger.Info("装备强化完成，准备重新穿戴",
		zap.String("equipmentID", equipment.ID),
		zap.Int("newEnhanceLevel", equipment.EnhanceLevel))

	// 刷新用户数据
	var userFresh models.User
	if err := db.DB.First(&userFresh, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 重新穿戴装备
	equipStats := jsonToFloatMap(equipment.Stats)
	baseAttrs := jsonToFloatMap(userFresh.BaseAttributes)
	combatAttrs := jsonToFloatMap(userFresh.CombatAttributes)
	combatRes := jsonToFloatMap(userFresh.CombatResistance)
	specialAttrs := jsonToFloatMap(userFresh.SpecialAttributes)

	// 处理灵宠属性
	var activePet models.Pet
	hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil
	if hasActivePet {
		petCombat := jsonToFloatMap(activePet.CombatAttributes)
		removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	// 应用新的装备属性
	applyEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

	// 重新应用灵宠加成
	if hasActivePet {
		activePetFresh := models.Pet{}
		db.DB.First(&activePetFresh, activePet.ID)
		petCombat := jsonToFloatMap(activePetFresh.CombatAttributes)
		applyPetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePetFresh, petCombat)
	}

	// 更新装备状态为已穿戴
	equipment.Equipped = true
	equipment.Slot = equipment.EquipType
	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 更新用户属性
	updates := map[string]interface{}{
		"base_attributes":    toJSON(baseAttrs),
		"combat_attributes":  toJSON(combatAttrs),
		"combat_resistance":  toJSON(combatRes),
		"special_attributes": toJSON(specialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：缓存更新后的装备数据到Redis（可选，用于快速查询）
	equipmentCacheKey := fmt.Sprintf("user:%s:equipment:%s", userIDStr, equipment.ID)
	equipmentJSON, _ := json.Marshal(equipment)
	redisClient.Client.Set(c, equipmentCacheKey, string(equipmentJSON), 5*time.Second)

	// ✅ 新增：记录强化后状态
	zapLogger.Info("装备强化完成",
		zap.String("equipmentID", equipment.ID),
		zap.Int("newEnhanceLevel", equipment.EnhanceLevel),
		zap.Int("costReinforceStones", cost),
		zap.Any("oldStats", oldStats),
		zap.Any("newStats", stats))

	// 返回强化结果
	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "强化成功",
		"cost":             cost,
		"oldStats":         oldStats,
		"newStats":         stats,
		"newLevel":         equipment.EnhanceLevel,
		"newRequiredRealm": newRequiredRealm,
		"user": gin.H{
			"baseAttributes":    baseAttrs,
			"combatAttributes":  combatAttrs,
			"combatResistance":  combatRes,
			"specialAttributes": specialAttrs,
			"reinforce_stones":  userFresh.ReinforceStones - cost,
		},
	})
}

// ReforgeEquipment 装备洗练
// 对应 POST /api/player/equipment/:id/reforge
func ReforgeEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var req struct {
		RefinementStones int `json:"refinementStones"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 获取用户信息检查洗练石
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 洗练成本定义
	const refinementCost = 10

	// 检查洗练石是否足够
	if user.RefinementStones < refinementCost {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "洗练石不足",
			"cost":    refinementCost,
		})
		return
	}

	// 解析装备属性
	stats := map[string]float64{}
	if len(equipment.Stats) > 0 {
		_ = json.Unmarshal(equipment.Stats, &stats)
	}

	// 检查装备是否有属性
	if len(stats) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的装备"})
		return
	}

	// 保存原始属性
	oldStats := map[string]float64{}
	for k, v := range stats {
		oldStats[k] = v
	}

	// 生成新属性（随机浮动±50%）
	newStats := map[string]float64{}
	for k, v := range stats {
		delta := rand.Float64()*1.0 - 0.5 // 随机值 [-0.5, 0.5]
		newVal := v * (1 + delta)
		newStats[k] = newVal
	}

	// 返回洗练结果供用户确认
	// 注意：此时还没有扣除洗练石，需要用户确认后再扣除
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "洗练成功",
		"cost":     refinementCost,
		"oldStats": oldStats,
		"newStats": newStats,
	})
}

// ConfirmReforge 确认洗练结果
// 对应 POST /api/player/equipment/:id/reforge-confirm
func ConfirmReforge(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var req struct {
		Confirmed bool               `json:"confirmed"`
		NewStats  map[string]float64 `json:"newStats"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	// 洗练成本定义
	const refinementCost = 10

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 根据用户选择处理洗练结果
	if req.Confirmed {
		// 用户确认新属性，更新装备属性
		equipment.Stats = toJSON(req.NewStats)
		if err := db.DB.Save(&equipment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}

		// 扣除用户洗练石
		if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
			Update("refinement_stones", gorm.Expr("refinement_stones - ?", refinementCost)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "洗练属性已应用",
			"stats":   req.NewStats,
			"cost":    refinementCost,
		})
		return
	}

	// 用户取消，保留原属性（不扣除洗练石）
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已保留原有属性"})
}

// EquipEquipment 装备穿戴
// 对应 POST /api/player/equipment/:id/equip
func EquipEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var req struct {
		Slot string `json:"slot"`
	}
	_ = c.ShouldBindJSON(&req)

	// 获取zap logger
	logger, exists := c.Get("zap_logger")
	if !exists {
		// 如果无法获取logger，创建一个默认的
		logger, _ = zap.NewProduction()
	}
	zapLogger := logger.(*zap.Logger)

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 解析装备属性
	equipStats := jsonToFloatMap(equipment.Stats)

	// 打印穿戴前装备的属性
	zapLogger.Info("穿戴装备前的装备属性",
		zap.Uint("userID", userID),
		zap.String("equipmentID", equipment.ID),
		zap.String("equipmentName", equipment.Name),
		zap.String("equipType", func() string {
			if equipment.EquipType != nil {
				return *equipment.EquipType
			}
			return "nil"
		}()),
		zap.String("quality", equipment.Quality),
		zap.Any("equipmentStats", equipStats))

	// 检查装备类型是否存在
	var equipTypeForQuery string
	if equipment.EquipType != nil {
		equipTypeForQuery = *equipment.EquipType
	} else {
		// 如果EquipType为nil，返回错误而不是使用Type字段作为后备
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "装备类型为空，无法穿戴"})
		return
	}

	// 获取用户信息
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 解析用户属性
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	// 打印穿戴前玩家属性
	zapLogger.Info("穿戴装备前的玩家属性",
		zap.Uint("userID", userID),
		zap.Any("baseAttributes", baseAttrs),
		zap.Any("combatAttributes", combatAttrs),
		zap.Any("combatResistance", combatRes),
		zap.Any("specialAttributes", specialAttrs))

	// ✅ 改进：使用属性管理器统一处理
	attrMgr := NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)

	// 处理灵宠属性（如果有出战灵宠）
	var activePet models.Pet
	hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil
	var petCombat map[string]float64

	// ✅ 改进：先移除灵宠加成
	if hasActivePet {
		petCombat = jsonToFloatMap(activePet.CombatAttributes)
		attrMgr.RemovePetBonuses(&activePet, petCombat)
	}

	// 查找同类型已装备的装备，移除其属性
	var oldEquipment []models.Equipment
	if err := db.DB.Where("user_id = ? AND equip_type = ? AND id != ? AND equipped = ?",
		userID, equipTypeForQuery, id, true).Find(&oldEquipment).Error; err == nil {
		for _, old := range oldEquipment {
			oldEquipStats := jsonToFloatMap(old.Stats)
			attrMgr.RemoveEquipmentStats(oldEquipStats)
		}
	}

	// 卸下同类型已装备的装备
	if err := db.DB.Model(&models.Equipment{}).
		Where("user_id = ? AND equip_type = ? AND id != ?", userID, equipTypeForQuery, id).
		Update("equipped", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 设置装备状态为已装备
	equipment.Equipped = true
	if req.Slot != "" {
		equipment.Slot = &req.Slot
	} else if equipment.EquipType != nil {
		s := *equipment.EquipType
		equipment.Slot = &s
	}

	// 保存装备状态
	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 改进：使用属性管理器应用装备属性⭥灵宠加成
	attrMgr.ApplyEquipmentStats(equipStats)

	// 如果有出战灵宠，重新应用灵宠加成
	if hasActivePet {
		attrMgr.ApplyPetBonuses(&activePet, petCombat)
	}

	// 打印穿戴后玩家属性
	zapLogger.Info("穿戴装备后的玩家属性",
		zap.Uint("userID", userID),
		zap.Any("baseAttributes", attrMgr.BaseAttrs),
		zap.Any("combatAttributes", attrMgr.CombatAttrs),
		zap.Any("combatResistance", attrMgr.CombatRes),
		zap.Any("specialAttributes", attrMgr.SpecialAttrs))

	// 更新用户属性
	updates := map[string]interface{}{
		"base_attributes":    toJSON(attrMgr.BaseAttrs),
		"combat_attributes":  toJSON(attrMgr.CombatAttrs),
		"combat_resistance":  toJSON(attrMgr.CombatRes),
		"special_attributes": toJSON(attrMgr.SpecialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "装备穿戴成功",
		"equipment": equipment,
		"user": gin.H{
			"baseAttributes":    attrMgr.BaseAttrs,
			"combatAttributes":  attrMgr.CombatAttrs,
			"combatResistance":  attrMgr.CombatRes,
			"specialAttributes": attrMgr.SpecialAttrs,
			"reinforce_stones":  user.ReinforceStones,
			"refinement_stones": user.RefinementStones,
		},
	})
}

// UnequipEquipment 装备卸下
// 对应 POST /api/player/equipment/:id/unequip
func UnequipEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 检查装备是否处于装备状态
	if !equipment.Equipped {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "装备未处于装备状态，无法卸下"})
		return
	}

	// 解析装备属性
	equipStats := jsonToFloatMap(equipment.Stats)

	// 更新装备状态为未装备
	equipment.Equipped = false
	equipment.Slot = nil
	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 获取用户信息
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 解析用户属性
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	// ✅ 改进：使用属性管理器统一处理
	attrMgr := NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)

	// 处理灵宠属性（如果有出战灵宠）
	var activePet models.Pet
	hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil

	// ✅ 改进：先移除灵宠加成
	if hasActivePet {
		petCombat := jsonToFloatMap(activePet.CombatAttributes)
		attrMgr.RemovePetBonuses(&activePet, petCombat)
	}

	// ✅ 改进：使用属性管理器移除装备属性
	attrMgr.RemoveEquipmentStats(equipStats)

	// ✅ 改进：如果有出战灵宠，重新应用灵宠加成
	if hasActivePet {
		petCombat := jsonToFloatMap(activePet.CombatAttributes)
		attrMgr.ApplyPetBonuses(&activePet, petCombat)
	}

	// 更新用户属性
	updates := map[string]interface{}{
		"base_attributes":    toJSON(attrMgr.BaseAttrs),
		"combat_attributes":  toJSON(attrMgr.CombatAttrs),
		"combat_resistance":  toJSON(attrMgr.CombatRes),
		"special_attributes": toJSON(attrMgr.SpecialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "装备卸下成功",
		"equipment": equipment,
		"user": gin.H{
			"baseAttributes":    attrMgr.BaseAttrs,
			"combatAttributes":  attrMgr.CombatAttrs,
			"combatResistance":  attrMgr.CombatRes,
			"specialAttributes": attrMgr.SpecialAttrs,
			"reinforce_stones":  user.ReinforceStones,
			"refinement_stones": user.RefinementStones,
		},
	})
}

// SellEquipment 出售装备
// 对应 DELETE /api/player/equipment/:id
func SellEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")

	// 查找装备
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 删除装备记录
	if err := db.DB.Delete(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 根据装备品质计算返还的强化石数量
	qualityStoneMap := map[string]int{
		"mythic":    6,
		"legendary": 5,
		"epic":      4,
		"rare":      3,
		"uncommon":  2,
		"common":    1,
	}
	stones := qualityStoneMap[equipment.Quality]
	if stones == 0 {
		stones = 1
	}

	// 增加用户强化石数量
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones + ?", stones)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 返回出售结果
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "装备出售成功",
		"stonesReceived": stones,
	})
}

// BatchSellEquipment 批量出售装备
// 对应 POST /api/player/equipment/batch-sell
func BatchSellEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 解析请求参数
	var req struct {
		Quality string `json:"quality"`
		Type    string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	// 构建查询条件
	query := db.DB.Model(&models.Equipment{}).Where("user_id = ?", userID)
	if req.Quality != "" {
		query = query.Where("quality = ?", req.Quality)
	}
	if req.Type != "" {
		query = query.Where("equipType = ?", req.Type)
	}

	// 查询符合条件的装备列表
	var list []models.Equipment
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 检查是否有符合条件的装备
	if len(list) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "没有找到符合条件的装备"})
		return
	}

	// 定义品质与返还强化石的映射关系
	qualityStoneMap := map[string]int{
		"mythic":    6,
		"legendary": 5,
		"epic":      4,
		"rare":      3,
		"uncommon":  2,
		"common":    1,
	}

	// 计算总返还强化石数量
	totalStones := 0
	ids := make([]string, 0, len(list))
	for _, eq := range list {
		ids = append(ids, eq.ID)
		v := qualityStoneMap[eq.Quality]
		if v == 0 {
			v = 1
		}
		totalStones += v
	}

	// 删除符合条件的装备
	if err := db.DB.Where("id IN ?", ids).Delete(&models.Equipment{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 增加用户强化石数量
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones + ?", totalStones)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 返回批量出售结果
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "成功出售装备",
		"equipmentSold":  len(list),
		"stonesReceived": totalStones,
	})
}
