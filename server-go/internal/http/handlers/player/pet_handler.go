package player

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisc "xiuxian/server-go/internal/redis"
)

// DeployPet 对应 POST /api/player/pets/:id/deploy
func DeployPet(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 获取zap logger
	logger, exists := c.Get("zap_logger")
	if !exists {
		// 如果无法获取logger，创建一个默认的
		logger, _ = zap.NewProduction()
	}
	zapLogger := logger.(*zap.Logger)

	petID := c.Param("id")
	var pet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 打印出战前灵宠的属性
	petCombat := jsonToFloatMap(pet.CombatAttributes)
	zapLogger.Info("灵宠出战前的灵宠属性",
		zap.Uint("userID", userID),
		zap.String("petID", pet.ID),
		zap.String("petName", pet.Name),
		zap.String("rarity", pet.Rarity),
		zap.Int("level", pet.Level),
		zap.Int("star", pet.Star),
		zap.Float64("attackBonus", pet.AttackBonus),
		zap.Float64("defenseBonus", pet.DefenseBonus),
		zap.Float64("healthBonus", pet.HealthBonus),
		zap.Any("combatAttributes", petCombat))

	// 读取玩家属性
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	// 打印出战前玩家属性
	zapLogger.Info("灵宠出战前的玩家属性",
		zap.Uint("userID", userID),
		zap.Any("baseAttributes", baseAttrs),
		zap.Any("combatAttributes", combatAttrs),
		zap.Any("combatResistance", combatRes),
		zap.Any("specialAttributes", specialAttrs))

	// ✅ 改进：使用属性管理器来统一处理属性变更
	attrMgr := NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)

	// 检查是否已有出战灵宠
	var activePet models.Pet
	if err := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error; err == nil {
		if activePet.ID == pet.ID {
			// 已经出战
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "灵宠已经出战"})
			return
		}

		// ✅ 改进：统一使用 removePetBonuses 移除旧灵宠加成
		activePetCombat := jsonToFloatMap(activePet.CombatAttributes)
		attrMgr.RemovePetBonuses(&activePet, activePetCombat)

		// 召回当前出战灵宠
		if err := db.DB.Model(&models.Pet{}).Where("user_id = ? AND is_active = ?", userID, true).
			Update("is_active", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// ✅ 改进：不在此处重新应用装备属性
	// 原因：玩家当前保存到DB的属性已经包含了已装备的装备属性
	// 直接应用灵宠加成即可，装备属性已在用户属性中

	// 标记目标灵宠为出战
	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).
		Update("is_active", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 重新加载最新状态
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 改进：直接使用 AttributeManager 的统一应用方法
	// 玩家属性中已包含装备属性，直接应用灵宠加成
	petCombat = jsonToFloatMap(pet.CombatAttributes)
	attrMgr.ApplyPetBonuses(&pet, petCombat)

	// ✅ 改进：打印属性应用顺序
	zapLogger.Info("灵宠出战属性应用顺序",
		zap.String("说明", "玩家属性已包含装备加成"),
		zap.String("step1", "从数据库读取当前用户属性（包含装备属性）"),
		zap.String("step2", "移除旧灵宠的属性加成"),
		zap.String("step3", "应用新灵宠的属性加成（按数值→战斗→特殊→百分比顺序）"))

	// 打印出战后玩家属性
	zapLogger.Info("灵宠出战后的玩家属性",
		zap.Uint("userID", userID),
		zap.Any("baseAttributes", attrMgr.BaseAttrs),
		zap.Any("combatAttributes", attrMgr.CombatAttrs),
		zap.Any("combatResistance", attrMgr.CombatRes),
		zap.Any("specialAttributes", attrMgr.SpecialAttrs))

	// ✅ 改进：使用属性管理器中的属性值
	updates := map[string]interface{}{
		"base_attributes":    toJSON(attrMgr.BaseAttrs),
		"combat_attributes":  toJSON(attrMgr.CombatAttrs),
		"combat_resistance":  toJSON(attrMgr.CombatRes),
		"special_attributes": toJSON(attrMgr.SpecialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "出战成功",
		"pet": gin.H{
			"id":        pet.ID,
			"name":      pet.Name,
			"is_active": pet.IsActive,
		},
	})
}

// RecallPet 对应 POST /api/player/pets/:id/recall
func RecallPet(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	petID := c.Param("id")
	var pet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	if !pet.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"message": "灵宠未处于出战状态，无法召回"})
		return
	}

	// 标记为未出战
	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).
		Update("is_active", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 获取玩家属性
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)
	petCombat := jsonToFloatMap(pet.CombatAttributes)

	// ✅ 改进：使用属性管理器统一移除灵宠加成
	attrMgr := NewAttributeManager(baseAttrs, combatAttrs, combatRes, specialAttrs)
	attrMgr.RemovePetBonuses(&pet, petCombat)

	updates := map[string]interface{}{
		"base_attributes":    toJSON(attrMgr.BaseAttrs),
		"combat_attributes":  toJSON(attrMgr.CombatAttrs),
		"combat_resistance":  toJSON(attrMgr.CombatRes),
		"special_attributes": toJSON(attrMgr.SpecialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "召回成功",
		"pet": gin.H{
			"id":        pet.ID,
			"name":      pet.Name,
			"is_active": false,
		},
	})
}

// UpgradePet 对应 POST /api/player/pets/:id/upgrade
func UpgradePet(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	petID := c.Param("id")
	var req struct {
		EssenceCount int `json:"essenceCount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.EssenceCount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	// ✅ 新增：尝试获取升级锁（防止并发升级）
	// 如果已有升级在进行，直接拒绝请求
	acquired, err := redisc.TryUpgradeLock(c.Request.Context(), userID, petID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	if !acquired {
		c.JSON(http.StatusBadRequest, gin.H{"message": "灵宠升级正在进行中"})
		return
	}
	defer redisc.ReleaseUpgradeLock(c.Request.Context(), userID, petID)

	// ✅ 优化：只有第一次操作时清理缓存、强制从 DB 读取最新数据
	// 检查是否存在操作时间戳（第一次操作时不存在）
	_, errTs := redisc.GetPetLastOperationTime(c.Request.Context(), userID)
	if errTs != nil {
		// 不存在 - 第一次操作，清理缓存强制从 DB 读取
		if delErr := redisc.Client.Del(c.Request.Context(), fmt.Sprintf(redisc.PetResourceKeyFormat, userID)).Err(); delErr != nil {
			log.Printf("清理上牡资源缓存失败: %v", delErr)
		}
		log.Printf("第一次操作，清理缓存从 DB 读取")
	} else {
		// 存在 - 后续操作，不清理、直接使用 Redis
		log.Printf("后续操作，不清理缓存直接用 Redis")
	}

	var pet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 优化：优先从 Redis 读灵宠精华
	// 但如果 Redis中不存在，则第一次会先通过数据库初始化缓存
	var userPetEssence int64 = int64(user.PetEssence)
	cachedResources, err := redisc.GetPetResources(c.Request.Context(), userID)
	if err == nil && cachedResources != nil {
		userPetEssence = cachedResources.PetEssence
	} else {
		// Redis 中不存在或检查失败 - 延迟初始化缓存
		if err := redisc.SetPetResources(c.Request.Context(), userID, int64(user.PetEssence)); err != nil {
			// 不中断流程，继滚重迎
			log.Printf("初始化灵宠缓存失败: %v", err)
		}
	}

	if userPetEssence < int64(req.EssenceCount) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "灵宠精华不足"})
		return
	}

	// ✅ 新增：从 Redis 扣除精华（快速更新缓存）
	newEssence, err := redisc.DecrementPetEssence(c.Request.Context(), userID, int64(req.EssenceCount))
	if err == nil {
		// Redis 操作成功，缓存已更新
		userPetEssence = newEssence
	}

	// 扣除灵宠精华（同步到数据库）
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("pet_essence", gorm.Expr("pet_essence - ?", req.EssenceCount)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 修改：升级前先保存旧属性（如果灵宠出战，需要用旧属性移除加成）
	var isActive bool = pet.IsActive
	var oldCombatAttrs map[string]float64
	if isActive {
		oldCombatAttrs = jsonToFloatMap(pet.CombatAttributes)
	}

	// 计算新等级与属性增长
	newLevel := pet.Level + 1

	// 获取灵宠当前的战斗属性
	combatAttrs := jsonToFloatMap(pet.CombatAttributes)

	// 提取基础属性字段，各自乘以2
	baseAttack := combatAttrs["attack"]
	if baseAttack == 0 {
		baseAttack = 10
	}
	newAttack := baseAttack * 1.1

	baseHealth := combatAttrs["health"]
	if baseHealth == 0 {
		baseHealth = 100
	}
	newHealth := baseHealth * 1.1

	baseDefense := combatAttrs["defense"]
	if baseDefense == 0 {
		baseDefense = 5
	}
	newDefense := baseDefense * 1.1

	baseSpeed := combatAttrs["speed"]
	if baseSpeed == 0 {
		baseSpeed = 10
	}
	newSpeed := baseSpeed * 1.1

	// 更新 CombatAttributes 中的属性值（各自乘以2）
	combatAttrs["attack"] = newAttack
	combatAttrs["health"] = newHealth
	combatAttrs["defense"] = newDefense
	combatAttrs["speed"] = newSpeed
	updatedCombatAttrs := toJSON(combatAttrs)

	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).
		Updates(map[string]interface{}{
			"level":             newLevel,
			"combat_attributes": updatedCombatAttrs,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：如果灵宠是出战状态，需要重新计算玩家属性
	if isActive {
		// 重新加载灵宠最新数据（已经被更新到DB）
		if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}

		// 重新计算玩家属性
		baseAttrs := jsonToFloatMap(user.BaseAttributes)
		combatAttrsUser := jsonToFloatMap(user.CombatAttributes)
		combatRes := jsonToFloatMap(user.CombatResistance)
		specialAttrs := jsonToFloatMap(user.SpecialAttributes)

		// 使用属性管理器重新计算
		attrMgr := NewAttributeManager(baseAttrs, combatAttrsUser, combatRes, specialAttrs)

		// 第一步：移除旧属性加成
		attrMgr.RemovePetBonuses(&pet, oldCombatAttrs)

		// 第二步：应用新的灵宠属性加成
		newCombatAttrs := jsonToFloatMap(pet.CombatAttributes)
		attrMgr.ApplyPetBonuses(&pet, newCombatAttrs)

		// 第三步：同步到数据库
		updates := map[string]interface{}{
			"base_attributes":    toJSON(attrMgr.BaseAttrs),
			"combat_attributes":  toJSON(attrMgr.CombatAttrs),
			"combat_resistance":  toJSON(attrMgr.CombatRes),
			"special_attributes": toJSON(attrMgr.SpecialAttrs),
		}
		if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}

		// 第四步：返回更新后的玩家属性
		// ✅ 新增：更新 Redis 缓存
		redisc.CachePet(c.Request.Context(), userID, petID, pet)
		redisc.InvalidatePetListCache(c.Request.Context(), userID)
		redisc.SyncPetResourcesToRedis(c.Request.Context(), userID, userPetEssence)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "升级成功",
			"pet":     pet,
			"user": gin.H{
				"baseAttributes":    attrMgr.BaseAttrs,
				"combatAttributes":  attrMgr.CombatAttrs,
				"combatResistance":  attrMgr.CombatRes,
				"specialAttributes": attrMgr.SpecialAttrs,
			},
		})
		return
	}

	// 如果灵宠不是出战状态，直接返回升级后的灵宠数据
	var updatedPet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&updatedPet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// ✅ 新增：更新 Redis 缓存
	redisc.CachePet(c.Request.Context(), userID, petID, updatedPet)
	redisc.InvalidatePetListCache(c.Request.Context(), userID)
	redisc.SyncPetResourcesToRedis(c.Request.Context(), userID, userPetEssence)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "升级成功", "pet": updatedPet})
}

// EvolvePet 对应 POST /api/player/pets/:id/evolve
func EvolvePet(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	petID := c.Param("id")
	var req struct {
		FoodPetID string `json:"foodPetId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.FoodPetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	// ✅ 新增：尝试获取升星锁（防止并发升星）
	acquired, err := redisc.TryEvolveLock(c.Request.Context(), userID, petID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	if !acquired {
		c.JSON(http.StatusBadRequest, gin.H{"message": "灵宠升星正在进行中"})
		return
	}
	defer redisc.ReleaseEvolveLock(c.Request.Context(), userID, petID)

	// ✅ 优化：只有第一次操作时清理缓存、强制从 DB 读取最新数据
	// 检查是否存在操作时间戳（第一次操作时不存在）
	_, errTs := redisc.GetPetLastOperationTime(c.Request.Context(), userID)
	if errTs != nil {
		// 不存在 - 第一次操作，清理缓存强制从 DB 读取
		if delErr := redisc.Client.Del(c.Request.Context(), fmt.Sprintf(redisc.PetResourceKeyFormat, userID)).Err(); delErr != nil {
			log.Printf("清理灵宠资源缓存失败: %v", delErr)
		}
		log.Printf("第一次操作，清理缓存从 DB 读取")
	} else {
		// 存在 - 后续操作，不清理、直接使用 Redis
		log.Printf("后续操作，不清理缓存直接用 Redis")
	}

	var targetPet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&targetPet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "目标灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	var foodPet models.Pet
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).First(&foodPet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "材料灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	if targetPet.Rarity != foodPet.Rarity {
		c.JSON(http.StatusBadRequest, gin.H{"message": "只能使用相同品质的灵宠进行升星"})
		return
	}

	successRate := 1.0
	if targetPet.Name != foodPet.Name {
		successRate = 0.3
	}
	isSuccess := rand.Float64() < successRate

	// ✅ 修改：升星前先保存旧属性（需要用旧属性移除大会岗召a）
	var isActive bool = targetPet.IsActive
	var oldAttackBonus float64 = targetPet.AttackBonus
	var oldDefenseBonus float64 = targetPet.DefenseBonus
	var oldHealthBonus float64 = targetPet.HealthBonus

	if isSuccess {
		newStar := targetPet.Star + 1
		qualityBonusMap := map[string]float64{
			"mythic":    0.15,
			"legendary": 0.12,
			"epic":      0.09,
			"rare":      0.06,
			"uncommon":  0.03,
			"common":    0.03,
		}
		starBonusPerQuality := map[string]float64{
			"mythic":    0.02,
			"legendary": 0.01,
			"epic":      0.01,
			"rare":      0.01,
			"uncommon":  0.01,
			"common":    0.01,
		}
		baseBonus := qualityBonusMap[targetPet.Rarity]                        // 品质基础加成（如mythic:0.15）
		starBonus := float64(newStar) * starBonusPerQuality[targetPet.Rarity] // 星级加成（当前星级 * 品质星级系数）
		levelBonus := float64(targetPet.Level-1) * (baseBonus * 0.1)          // 等级加成（(等级-1) * 基础加成 * 0.1）
		phase := newStar / 5                                                  // 计算阶段（每5星为1个阶段）
		phaseBonus := float64(phase) * (baseBonus * 0.5)                      // 阶段加成（阶段数 * 基础加成 * 0.5）
		newBonus := baseBonus + starBonus + levelBonus + phaseBonus           // 合并计算新的属性加成值

		if err := db.DB.Model(&models.Pet{}).
			Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).
			Updates(map[string]interface{}{
				"star":          newStar,
				"attack_bonus":  newBonus,
				"defense_bonus": newBonus,
				"health_bonus":  newBonus,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}

		// ✅ 新增：如果灵宠是出战状态，需要重新计算玩家属性
		if isActive {
			// 重新加载灵宠最新数据（已经被更新到DB）
			if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&targetPet).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
				return
			}

			// 重新加载玩家数据
			var user models.User
			if err := db.DB.First(&user, userID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
				return
			}

			// 重新计算玩家属性
			baseAttrs := jsonToFloatMap(user.BaseAttributes)
			combatAttrsUser := jsonToFloatMap(user.CombatAttributes)
			combatRes := jsonToFloatMap(user.CombatResistance)
			specialAttrs := jsonToFloatMap(user.SpecialAttributes)

			// 使用属性管理器重新计算
			attrMgr := NewAttributeManager(baseAttrs, combatAttrsUser, combatRes, specialAttrs)

			// 第一步：移除旧的属性加成
			oldPet := targetPet
			oldPet.AttackBonus = oldAttackBonus
			oldPet.DefenseBonus = oldDefenseBonus
			oldPet.HealthBonus = oldHealthBonus
			attrMgr.RemovePetBonuses(&oldPet, nil)

			// 第二步：应用新的灵宠属性加成
			petCombatNew := jsonToFloatMap(targetPet.CombatAttributes)
			attrMgr.ApplyPetBonuses(&targetPet, petCombatNew)

			// 第三步：同步到数据库
			updates := map[string]interface{}{
				"base_attributes":    toJSON(attrMgr.BaseAttrs),
				"combat_attributes":  toJSON(attrMgr.CombatAttrs),
				"combat_resistance":  toJSON(attrMgr.CombatRes),
				"special_attributes": toJSON(attrMgr.SpecialAttrs),
			}
			if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
				return
			}

			// 第四步：返回更新后的玩家属性
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "升星成功",
				"pet":     targetPet,
				"user": gin.H{
					"baseAttributes":    attrMgr.BaseAttrs,
					"combatAttributes":  attrMgr.CombatAttrs,
					"combatResistance":  attrMgr.CombatRes,
					"specialAttributes": attrMgr.SpecialAttrs,
				},
			})

			// ✅ 新增：删除材料灵宠之前，先返回给前端
			// 删除材料灵宠
			if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).
				Delete(&models.Pet{}).Error; err != nil {
				// 这里记录错误，但不改变前面升星结果
				log.Printf("删除材料灵宠失败: %v", err)
			}
			return
		}

		// 如果灵宠不是出战状态，直接返回升星后的灵宠数据
		var updatedPet models.Pet
		if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).First(&updatedPet).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}

		// ✅ 新增：更新 Redis 缓存
		redisc.CachePet(c.Request.Context(), userID, petID, updatedPet)
		redisc.InvalidatePetCache(c.Request.Context(), userID, req.FoodPetID) // 材料灵宠也清缓存
		redisc.InvalidatePetListCache(c.Request.Context(), userID)

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "升星成功", "pet": updatedPet})

		// ✅ 新增：删除材料灵宠之前，先返回给前端
		// 删除材料灵宠
		if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).
			Delete(&models.Pet{}).Error; err != nil {
			// 这里记录错误，但不改变前面升星结果
			log.Printf("删除材料灵宠失败: %v", err)
		}
		return
	} else {
		// 失败时返回成功率信息
		// ✅ 新增：升星失败也清除灵宠缓存（因为材料灵宠已消耗）
		redisc.InvalidatePetCache(c.Request.Context(), userID, req.FoodPetID)
		redisc.InvalidatePetListCache(c.Request.Context(), userID)

		successRatePercent := int(successRate * 100)
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升星失败，材料已消耗", "successRate": successRatePercent})

		// ✅ 新增：升星失败，删除材料灵宠
		if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).
			Delete(&models.Pet{}).Error; err != nil {
			log.Printf("删除材料灵宠失败: %v", err)
		}
	}
}

// BatchReleasePets 对应 POST /api/player/pets/batch-release
func BatchReleasePets(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var req struct {
		Rarity string `json:"rarity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误"})
		return
	}

	query := db.DB.Where("user_id = ? AND is_active = ?", userID, false)
	if req.Rarity != "" {
		query = query.Where("rarity = ?", req.Rarity)
	}
	var pets []models.Pet
	if err := query.Find(&pets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
	if len(pets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "没有符合条件的灵宠可放生"})
		return
	}

	totalExp := 0
	for _, p := range pets {
		lvl := p.Level
		if lvl <= 0 {
			lvl = 1
		}
		exp := int(math.Floor(float64(lvl)*float64(p.Star+1)*0.5 + 0.000001))
		if exp < 1 {
			exp = 1
		}
		totalExp += exp
	}

	// 删除这些宠物
	if err := query.Delete(&models.Pet{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}
}
