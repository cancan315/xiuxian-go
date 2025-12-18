package player

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
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

	// 检查是否已有出战灵宠
	var activePet models.Pet
	if err := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error; err == nil {
		if activePet.ID == pet.ID {
			// 已经出战
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "灵宠已经出战"})
			return
		}

		// 召回当前出战灵宠并移除其属性
		activePetCombat := jsonToFloatMap(activePet.CombatAttributes)

		// 先移除基础属性百分比加成
		if activePet.AttackBonus != 0 {
			baseAttrs["attack"] = baseAttrs["attack"] / (1 + activePet.AttackBonus)
		}
		if activePet.DefenseBonus != 0 {
			baseAttrs["defense"] = baseAttrs["defense"] / (1 + activePet.DefenseBonus)
		}
		if activePet.HealthBonus != 0 {
			baseAttrs["health"] = baseAttrs["health"] / (1 + activePet.HealthBonus)
		}

		// 再移除 combatAttributes 数值
		if v, ok := activePetCombat["attack"]; ok {
			baseAttrs["attack"] = baseAttrs["attack"] - v
		}
		if v, ok := activePetCombat["defense"]; ok {
			baseAttrs["defense"] = baseAttrs["defense"] - v
		}
		if v, ok := activePetCombat["health"]; ok {
			baseAttrs["health"] = baseAttrs["health"] - v
		}
		if v, ok := activePetCombat["speed"]; ok {
			baseAttrs["speed"] = baseAttrs["speed"] - v
		}

		clampZero := func(m map[string]float64, key string, delta float64) {
			if delta == 0 {
				return
			}
			m[key] = m[key] - delta
			if m[key] < 0 {
				m[key] = 0
			}
		}
		clampZero(combatAttrs, "critRate", activePetCombat["critRate"])
		clampZero(combatAttrs, "comboRate", activePetCombat["comboRate"])
		clampZero(combatAttrs, "counterRate", activePetCombat["counterRate"])
		clampZero(combatAttrs, "stunRate", activePetCombat["stunRate"])
		clampZero(combatAttrs, "dodgeRate", activePetCombat["dodgeRate"])
		clampZero(combatAttrs, "vampireRate", activePetCombat["vampireRate"])

		clampZero(combatRes, "critResist", activePetCombat["critResist"])
		clampZero(combatRes, "comboResist", activePetCombat["comboResist"])
		clampZero(combatRes, "counterResist", activePetCombat["counterResist"])
		clampZero(combatRes, "stunResist", activePetCombat["stunResist"])
		clampZero(combatRes, "dodgeResist", activePetCombat["dodgeResist"])
		clampZero(combatRes, "vampireResist", activePetCombat["vampireResist"])

		for _, key := range []string{"healBoost", "critDamageBoost", "critDamageReduce", "finalDamageBoost", "finalDamageReduce", "combatBoost", "resistanceBoost"} {
			if v, ok := activePetCombat[key]; ok {
				specialAttrs[key] = specialAttrs[key] - v
				if specialAttrs[key] < 0 {
					specialAttrs[key] = 0
				}
			}
		}

		// 召回当前出战灵宠
		if err := db.DB.Model(&models.Pet{}).Where("user_id = ? AND is_active = ?", userID, true).
			Update("is_active", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

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

	// 解析灵宠 combatAttributes
	petCombat = jsonToFloatMap(pet.CombatAttributes)

	// 按 Node 逻辑：先累加 combatAttributes 中的数值
	if v, ok := petCombat["attack"]; ok {
		baseAttrs["attack"] = baseAttrs["attack"] + v
	}
	if v, ok := petCombat["defense"]; ok {
		baseAttrs["defense"] = baseAttrs["defense"] + v
	}
	if v, ok := petCombat["health"]; ok {
		baseAttrs["health"] = baseAttrs["health"] + v
	}
	if v, ok := petCombat["speed"]; ok {
		baseAttrs["speed"] = baseAttrs["speed"] + v
	}

	// 战斗属性（带 0-1 上限）
	updateRate := func(m map[string]float64, key string, delta float64) {
		if delta == 0 {
			return
		}
		m[key] = m[key] + delta
		if m[key] > 1 {
			m[key] = 1
		}
	}
	updateRate(combatAttrs, "critRate", petCombat["critRate"])
	updateRate(combatAttrs, "comboRate", petCombat["comboRate"])
	updateRate(combatAttrs, "counterRate", petCombat["counterRate"])
	updateRate(combatAttrs, "stunRate", petCombat["stunRate"])
	updateRate(combatAttrs, "dodgeRate", petCombat["dodgeRate"])
	updateRate(combatAttrs, "vampireRate", petCombat["vampireRate"])

	updateRate(combatRes, "critResist", petCombat["critResist"])
	updateRate(combatRes, "comboResist", petCombat["comboResist"])
	updateRate(combatRes, "counterResist", petCombat["counterResist"])
	updateRate(combatRes, "stunResist", petCombat["stunResist"])
	updateRate(combatRes, "dodgeResist", petCombat["dodgeResist"])
	updateRate(combatRes, "vampireResist", petCombat["vampireResist"])

	// 特殊属性：直接累加
	for _, key := range []string{"healBoost", "critDamageBoost", "critDamageReduce", "finalDamageBoost", "finalDamageReduce", "combatBoost", "resistanceBoost"} {
		if v, ok := petCombat[key]; ok {
			specialAttrs[key] = specialAttrs[key] + v
		}
	}

	// 基础属性百分比加成
	if pet.AttackBonus != 0 {
		baseAttrs["attack"] = baseAttrs["attack"] * (1 + pet.AttackBonus)
	}
	if pet.DefenseBonus != 0 {
		baseAttrs["defense"] = baseAttrs["defense"] * (1 + pet.DefenseBonus)
	}
	if pet.HealthBonus != 0 {
		baseAttrs["health"] = baseAttrs["health"] * (1 + pet.HealthBonus)
	}

	// 打印出战后玩家属性
	zapLogger.Info("灵宠出战后的玩家属性",
		zap.Uint("userID", userID),
		zap.Any("baseAttributes", baseAttrs),
		zap.Any("combatAttributes", combatAttrs),
		zap.Any("combatResistance", combatRes),
		zap.Any("specialAttributes", specialAttrs))

	// 写回用户属性
	updates := map[string]interface{}{
		"base_attributes":    toJSON(baseAttrs),
		"combat_attributes":  toJSON(combatAttrs),
		"combat_resistance":  toJSON(combatRes),
		"special_attributes": toJSON(specialAttrs),
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

	// ✅ 改进：使用统一的 removePetBonuses 函数移除灵宠加成
	removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &pet, petCombat)

	updates := map[string]interface{}{
		"base_attributes":    toJSON(baseAttrs),
		"combat_attributes":  toJSON(combatAttrs),
		"combat_resistance":  toJSON(combatRes),
		"special_attributes": toJSON(specialAttrs),
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
	if user.PetEssence < req.EssenceCount {
		c.JSON(http.StatusBadRequest, gin.H{"message": "灵宠精华不足"})
		return
	}

	// 扣除灵宠精华
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("pet_essence", gorm.Expr("pet_essence - ?", req.EssenceCount)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 计算新等级与加成
	newLevel := pet.Level + 1
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
	baseBonus := qualityBonusMap[pet.Rarity]
	starBonus := float64(pet.Star) * starBonusPerQuality[pet.Rarity]
	levelBonus := float64(newLevel-1) * (baseBonus * 0.1)
	phase := pet.Star / 5
	phaseBonus := float64(phase) * (baseBonus * 0.5)
	newBonus := baseBonus + starBonus + levelBonus + phaseBonus

	if err := db.DB.Model(&models.Pet{}).
		Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, petID, petID).
		Updates(map[string]interface{}{
			"level":         newLevel,
			"attack_bonus":  newBonus,
			"defense_bonus": newBonus,
			"health_bonus":  newBonus,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "升级成功", "newLevel": newLevel})
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
		baseBonus := qualityBonusMap[targetPet.Rarity]
		starBonus := float64(newStar) * starBonusPerQuality[targetPet.Rarity]
		levelBonus := float64(targetPet.Level-1) * (baseBonus * 0.1)
		phase := newStar / 5
		phaseBonus := float64(phase) * (baseBonus * 0.5)
		newBonus := baseBonus + starBonus + levelBonus + phaseBonus

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

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "升星成功", "newStar": newStar})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升星失败，材料已消耗"})
	}

	// 删除材料灵宠
	if err := db.DB.Where("user_id = ? AND (pet_id = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).
		Delete(&models.Pet{}).Error; err != nil {
		// 这里记录错误，但不改变前面升星结果
		log.Printf("删除材料灵宠失败: %v", err)
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

	// 增加经验值：直接更新 Users.exp 列
	if err := db.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("exp", gorm.Expr("exp + ?", totalExp)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "成功放生" + strconv.Itoa(len(pets)) + "只灵宠，获得" + strconv.Itoa(totalExp) + "点经验",
		"releasedCount": len(pets),
		"expGained":     totalExp,
	})
}
