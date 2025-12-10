package player

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
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

	var items []models.Item
	if err := db.DB.Where("userId = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}

	var pets []models.Pet
	if err := db.DB.Where("userId = ?", userID).Find(&pets).Error; err != nil {
		return nil, err
	}

	var herbs []models.Herb
	if err := db.DB.Where("userId = ?", userID).Find(&herbs).Error; err != nil {
		return nil, err
	}

	var pills []models.Pill
	if err := db.DB.Where("userId = ?", userID).Find(&pills).Error; err != nil {
		return nil, err
	}

	// artifacts 目前 Node 版固定为空数组
	artifacts := []interface{}{}

	return gin.H{
		"user":      user,
		"items":     items,
		"pets":      pets,
		"herbs":     herbs,
		"pills":     pills,
		"artifacts": artifacts,
	}, nil
}

// GetPlayerData 对应 GET /api/player/data （已弃用，但前端仍在使用）
func GetPlayerData(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	data, err := assembleFullPlayerData(userID)
	if err != nil {
		log.Printf("get player data failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// InitializePlayer 对应 GET /api/player/init
func InitializePlayer(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	data, err := assembleFullPlayerData(userID)
	if err != nil {
		log.Printf("initialize player failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetPlayerSpirit 对应 GET /api/player/spirit
func GetPlayerSpirit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	var user models.User
	if err := db.DB.Select("spirit").First(&user, userID).Error; err != nil {
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
	var list []struct {
		ID           uint   `json:"id"`
		PlayerName   string `json:"playerName"`
		Level        int    `json:"level"`
		Realm        string `json:"realm"`
		SpiritStones int    `json:"spiritStones"`
	}

	if err := db.DB.Model(&models.User{}).
		Select("id, \"playerName\", level, realm, \"spiritStones\"").
		Order("level DESC").
		Order("\"spiritStones\" DESC").
		Limit(100).
		Scan(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
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

	if err := db.DB.Where("userId = ? AND (equipmentId IN ? OR id IN ?)", userID, req.ItemIDs, req.ItemIDs).
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

	if err := db.DB.Where("userId = ? AND (petId IN ? OR id IN ?)", userID, req.PetIDs, req.PetIDs).
		Delete(&models.Pet{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	log.Printf("成功删除 %d 个灵宠, user=%d", len(req.PetIDs), userID)
	c.JSON(http.StatusOK, gin.H{"message": "灵宠删除成功"})
}

// DeployPet 对应 POST /api/player/pets/:id/deploy
func DeployPet(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	petID := c.Param("id")
	var pet models.Pet
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 检查是否已有出战灵宠
	var activePet models.Pet
	if err := db.DB.Where("userId = ? AND isActive = ?", userID, true).First(&activePet).Error; err == nil {
		if activePet.ID == pet.ID {
			// 已经出战
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "灵宠已经出战"})
			return
		}
		// 召回当前出战灵宠
		if err := db.DB.Model(&models.Pet{}).Where("userId = ? AND isActive = ?", userID, true).
			Update("isActive", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}
	}

	// 标记目标灵宠为出战
	if err := db.DB.Model(&models.Pet{}).
		Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).
		Update("isActive", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	// 重新加载最新状态
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

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

	// 解析灵宠 combatAttributes
	petCombat := jsonToFloatMap(pet.CombatAttributes)

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

	// 写回用户属性
	updates := map[string]interface{}{
		"baseAttributes":    toJSON(baseAttrs),
		"combatAttributes":  toJSON(combatAttrs),
		"combatResistance":  toJSON(combatRes),
		"specialAttributes": toJSON(specialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "出战成功",
		"pet": gin.H{
			"id":       pet.ID,
			"name":     pet.Name,
			"isActive": pet.IsActive,
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
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
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
		Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).
		Update("isActive", false).Error; err != nil {
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

	// 先移除基础属性百分比加成
	if pet.AttackBonus != 0 {
		baseAttrs["attack"] = baseAttrs["attack"] / (1 + pet.AttackBonus)
	}
	if pet.DefenseBonus != 0 {
		baseAttrs["defense"] = baseAttrs["defense"] / (1 + pet.DefenseBonus)
	}
	if pet.HealthBonus != 0 {
		baseAttrs["health"] = baseAttrs["health"] / (1 + pet.HealthBonus)
	}

	// 再移除 combatAttributes 数值
	if v, ok := petCombat["attack"]; ok {
		baseAttrs["attack"] = baseAttrs["attack"] - v
	}
	if v, ok := petCombat["defense"]; ok {
		baseAttrs["defense"] = baseAttrs["defense"] - v
	}
	if v, ok := petCombat["health"]; ok {
		baseAttrs["health"] = baseAttrs["health"] - v
	}
	if v, ok := petCombat["speed"]; ok {
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
	clampZero(combatAttrs, "critRate", petCombat["critRate"])
	clampZero(combatAttrs, "comboRate", petCombat["comboRate"])
	clampZero(combatAttrs, "counterRate", petCombat["counterRate"])
	clampZero(combatAttrs, "stunRate", petCombat["stunRate"])
	clampZero(combatAttrs, "dodgeRate", petCombat["dodgeRate"])
	clampZero(combatAttrs, "vampireRate", petCombat["vampireRate"])

	clampZero(combatRes, "critResist", petCombat["critResist"])
	clampZero(combatRes, "comboResist", petCombat["comboResist"])
	clampZero(combatRes, "counterResist", petCombat["counterResist"])
	clampZero(combatRes, "stunResist", petCombat["stunResist"])
	clampZero(combatRes, "dodgeResist", petCombat["dodgeResist"])
	clampZero(combatRes, "vampireResist", petCombat["vampireResist"])

	for _, key := range []string{"healBoost", "critDamageBoost", "critDamageReduce", "finalDamageBoost", "finalDamageReduce", "combatBoost", "resistanceBoost"} {
		if v, ok := petCombat[key]; ok {
			specialAttrs[key] = specialAttrs[key] - v
			if specialAttrs[key] < 0 {
				specialAttrs[key] = 0
			}
		}
	}

	updates := map[string]interface{}{
		"baseAttributes":    toJSON(baseAttrs),
		"combatAttributes":  toJSON(combatAttrs),
		"combatResistance":  toJSON(combatRes),
		"specialAttributes": toJSON(specialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "召回成功",
		"pet": gin.H{
			"id":       pet.ID,
			"name":     pet.Name,
			"isActive": false,
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
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).First(&pet).Error; err != nil {
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
		Update("petEssence", gorm.Expr("\"petEssence\" - ?", req.EssenceCount)).Error; err != nil {
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
		Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).
		Updates(map[string]interface{}{
			"level":        newLevel,
			"attackBonus":  newBonus,
			"defenseBonus": newBonus,
			"healthBonus":  newBonus,
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
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).First(&targetPet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "目标灵宠未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
		return
	}

	var foodPet models.Pet
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).First(&foodPet).Error; err != nil {
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
			Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, petID).
			Updates(map[string]interface{}{
				"star":         newStar,
				"attackBonus":  newBonus,
				"defenseBonus": newBonus,
				"healthBonus":  newBonus,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "服务器错误", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "升星成功", "newStar": newStar})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "升星失败，材料已消耗"})
	}

	// 删除材料灵宠
	if err := db.DB.Where("userId = ? AND (petId = ? OR id = ?)", userID, req.FoodPetID, req.FoodPetID).
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

	query := db.DB.Where("userId = ? AND isActive = ?", userID, false)
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

// GetPlayerInventory 对应 GET /api/player/inventory/:userId?
// 为保持与 Node 版兼容，这里仍然支持可选 userId，但在当前 Go 实现中
// 实际使用的是 JWT 中的用户 ID（即当前登录用户）。
func GetPlayerInventory(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 构建查询条件
	query := db.DB.Model(&models.Equipment{}).Where("userId = ?", userID)

	// 处理类型过滤：优先 type，其次 equipType
	typeParam := c.Query("type")
	if typeParam != "" {
		query = query.Where("equipType = ?", typeParam)
	} else if equipType := c.Query("equipType"); equipType != "" {
		query = query.Where("equipType = ?", equipType)
	}

	// 品质过滤
	if quality := c.Query("quality"); quality != "" {
		query = query.Where("quality = ?", quality)
	}

	// 装备状态过滤
	if equipped := c.Query("equipped"); equipped != "" {
		query = query.Where("equipped = ?", strings.ToLower(equipped) == "true")
	}

	var equipment []models.Equipment
	if err := query.Order("createdAt DESC").Find(&equipment).Error; err != nil {
		log.Printf("获取玩家装备数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	// Node 版返回 { success: true, items: [...] }
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   equipment,
	})
}

// GetItemDetails 对应 GET /api/player/inventory/item/:itemId
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
	if err := db.DB.Where("id = ? AND userId = ?", itemID, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "物品未找到",
			})
			return
		}
		log.Printf("获取物品详情失败: %v", err)
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

// GetPlayerEquipment 对应 GET /api/player/equipment/:userId?
func GetPlayerEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	query := db.DB.Model(&models.Equipment{}).Where("userId = ?", userID)

	// 类型过滤：使用 type 或 equipType
	typeParam := c.Query("type")
	if typeParam != "" {
		query = query.Where("equipType = ?", typeParam)
	} else if equipType := c.Query("equipType"); equipType != "" {
		query = query.Where("equipType = ?", equipType)
	}

	if quality := c.Query("quality"); quality != "" {
		query = query.Where("quality = ?", quality)
	}

	if equipped := c.Query("equipped"); equipped != "" {
		query = query.Where("equipped = ?", strings.ToLower(equipped) == "true")
	}

	var equipment []models.Equipment
	if err := query.Order("createdAt DESC").Find(&equipment).Error; err != nil {
		log.Printf("获取玩家装备数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"equipment": equipment,
	})
}

// GetEquipmentDetails 对应 GET /api/player/equipment/details/:id
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
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "装备未找到",
			})
			return
		}
		log.Printf("获取装备详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "服务器错误",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"equipment": equipment,
	})
}

// EnhanceEquipment 对应 POST /api/player/equipment/:id/enhance
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

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 解析 stats
	stats := map[string]float64{}
	if len(equipment.Stats) > 0 {
		_ = json.Unmarshal(equipment.Stats, &stats)
	}

	// 简化版强化逻辑
	currentLevel := equipment.EnhanceLevel
	cost := 10 * (currentLevel + 1)
	if req.ReinforceStones < cost || user.ReinforceStones < cost {
		c.JSON(http.StatusOK, gin.H{
			"success":  false,
			"message":  "强化石不足",
			"cost":     cost,
			"oldStats": stats,
			"newStats": stats,
		})
		return
	}

	oldStats := map[string]float64{}
	for k, v := range stats {
		oldStats[k] = v
	}

	for k, v := range stats {
		stats[k] = v * 1.1
	}
	equipment.EnhanceLevel = currentLevel + 1
	equipment.Stats = toJSON(stats)

	// 简单 requiredRealm：每 10 级+1
	newRequiredRealm := (equipment.EnhanceLevel / 10) + 1
	equipment.RequiredRealm = newRequiredRealm

	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforceStones", gorm.Expr("reinforceStones - ?", cost)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "强化成功",
		"cost":             cost,
		"oldStats":         oldStats,
		"newStats":         stats,
		"newLevel":         equipment.EnhanceLevel,
		"newRequiredRealm": newRequiredRealm,
	})
}

// ReforgeEquipment 对应 POST /api/player/equipment/:id/reforge
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

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	stats := map[string]float64{}
	if len(equipment.Stats) > 0 {
		_ = json.Unmarshal(equipment.Stats, &stats)
	}

	if len(stats) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的装备"})
		return
	}

	oldStats := map[string]float64{}
	for k, v := range stats {
		oldStats[k] = v
	}

	rand.Seed(time.Now().UnixNano())
	newStats := map[string]float64{}
	for k, v := range stats {
		delta := rand.Float64()*1.0 - 0.5
		newVal := v * (1 + delta)
		newStats[k] = newVal
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "洗练成功",
		"cost":     10,
		"oldStats": oldStats,
		"newStats": newStats,
	})
}

// ConfirmReforge 对应 POST /api/player/equipment/:id/reforge-confirm
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

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if req.Confirmed {
		equipment.Stats = toJSON(req.NewStats)
		if err := db.DB.Save(&equipment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "洗练属性已应用", "stats": req.NewStats})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已保留原有属性"})
}

// EquipEquipment 对应 POST /api/player/equipment/:id/equip
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

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 解析装备 stats
	equipStats := jsonToFloatMap(equipment.Stats)

	// 同类型只允许一个装备为 equipped
	if err := db.DB.Model(&models.Equipment{}).
		Where("userId = ? AND equipType = ?", userID, equipment.EquipType).
		Update("equipped", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	equipment.Equipped = true
	if req.Slot != "" {
		equipment.Slot = &req.Slot
	} else if equipment.EquipType != nil {
		s := *equipment.EquipType
		equipment.Slot = &s
	}

	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	// 加载用户与出战灵宠
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	var activePet models.Pet
	hasActivePet := db.DB.Where("userId = ? AND isActive = ?", userID, true).First(&activePet).Error == nil
	petCombat := jsonToFloatMap(activePet.CombatAttributes)

	// 先移除灵宠加成（百分比和基础/战斗/抗性/特殊）
	if hasActivePet {
		removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	// 应用装备属性
	applyEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

	// 重新应用灵宠加成
	if hasActivePet {
		applyPetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	// 写回用户属性
	updates := map[string]interface{}{
		"baseAttributes":    toJSON(baseAttrs),
		"combatAttributes":  toJSON(combatAttrs),
		"combatResistance":  toJSON(combatRes),
		"specialAttributes": toJSON(specialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "装备穿戴成功",
		"equipment": equipment,
		"user": gin.H{
			"baseAttributes":    baseAttrs,
			"combatAttributes":  combatAttrs,
			"combatResistance":  combatRes,
			"specialAttributes": specialAttrs,
			"reinforceStones":   user.ReinforceStones,
			"refinementStones":  user.RefinementStones,
		},
	})
}

// UnequipEquipment 对应 POST /api/player/equipment/:id/unequip
func UnequipEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if !equipment.Equipped {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "装备未处于装备状态，无法卸下"})
		return
	}

	// 解析装备 stats
	equipStats := jsonToFloatMap(equipment.Stats)

	equipment.Equipped = false
	equipment.Slot = nil
	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	baseAttrs := jsonToFloatMap(user.BaseAttributes)
	combatAttrs := jsonToFloatMap(user.CombatAttributes)
	combatRes := jsonToFloatMap(user.CombatResistance)
	specialAttrs := jsonToFloatMap(user.SpecialAttributes)

	var activePet models.Pet
	hasActivePet := db.DB.Where("userId = ? AND isActive = ?", userID, true).First(&activePet).Error == nil
	petCombat := jsonToFloatMap(activePet.CombatAttributes)

	// 先移除灵宠加成
	if hasActivePet {
		removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	// 移除装备属性（与穿戴相反）
	removeEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

	// 重新应用灵宠加成
	if hasActivePet {
		applyPetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	updates := map[string]interface{}{
		"baseAttributes":    toJSON(baseAttrs),
		"combatAttributes":  toJSON(combatAttrs),
		"combatResistance":  toJSON(combatRes),
		"specialAttributes": toJSON(specialAttrs),
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "装备卸下成功",
		"equipment": equipment,
		"user": gin.H{
			"baseAttributes":    baseAttrs,
			"combatAttributes":  combatAttrs,
			"combatResistance":  combatRes,
			"specialAttributes": specialAttrs,
			"reinforceStones":   user.ReinforceStones,
			"refinementStones":  user.RefinementStones,
		},
	})
}

// jsonToFloatMap 将 JSON 转换为 map[string]float64
func jsonToFloatMap(j datatypes.JSON) map[string]float64 {
	if len(j) == 0 {
		return map[string]float64{}
	}
	var m map[string]float64
	if err := json.Unmarshal(j, &m); err != nil {
		return map[string]float64{}
	}
	return m
}

func SellEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND userId = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if err := db.DB.Delete(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "装备出售成功",
		"stonesReceived": stones,
	})
}

// BatchSellEquipment 对应 POST /api/player/equipment/batch-sell
func BatchSellEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var req struct {
		Quality string `json:"quality"`
		Type    string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	query := db.DB.Model(&models.Equipment{}).Where("userId = ?", userID)
	if req.Quality != "" {
		query = query.Where("quality = ?", req.Quality)
	}
	if req.Type != "" {
		query = query.Where("equipType = ?", req.Type)
	}

	var list []models.Equipment
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if len(list) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "没有找到符合条件的装备"})
		return
	}

	qualityStoneMap := map[string]int{
		"mythic":    6,
		"legendary": 5,
		"epic":      4,
		"rare":      3,
		"uncommon":  2,
		"common":    1,
	}

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

	if err := db.DB.Where("id IN ?", ids).Delete(&models.Equipment{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "成功出售装备",
		"equipmentSold":  len(list),
		"stonesReceived": totalStones,
	})
}

// 以下是内部帮助函数，用于将前端传来的 map 数据增量写入数据库

func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(b)
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

// removePetBonuses 在修改装备前移除灵宠提供的所有加成
func removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs map[string]float64, pet *models.Pet, petCombat map[string]float64) {
	// 百分比基础属性加成
	if pet.AttackBonus != 0 {
		if cur := baseAttrs["attack"]; cur != 0 {
			baseAttrs["attack"] = cur / (1 + pet.AttackBonus)
		}
	}
	if pet.DefenseBonus != 0 {
		if cur := baseAttrs["defense"]; cur != 0 {
			baseAttrs["defense"] = cur / (1 + pet.DefenseBonus)
		}
	}
	if pet.HealthBonus != 0 {
		if cur := baseAttrs["health"]; cur != 0 {
			baseAttrs["health"] = cur / (1 + pet.HealthBonus)
		}
	}

	// 基础属性与战斗属性、抗性、特殊属性的数值加成（从当前值减去）
	if v, ok := petCombat["attack"]; ok {
		baseAttrs["attack"] = (baseAttrs["attack"] - v)
	}
	if v, ok := petCombat["defense"]; ok {
		baseAttrs["defense"] = (baseAttrs["defense"] - v)
	}
	if v, ok := petCombat["health"]; ok {
		baseAttrs["health"] = (baseAttrs["health"] - v)
	}
	if v, ok := petCombat["speed"]; ok {
		baseAttrs["speed"] = (baseAttrs["speed"] - v)
	}

	floatKeys := []struct {
		key string
		m   map[string]float64
	}{
		{"critRate", combatAttrs},
		{"comboRate", combatAttrs},
		{"counterRate", combatAttrs},
		{"stunRate", combatAttrs},
		{"dodgeRate", combatAttrs},
		{"vampireRate", combatAttrs},
		{"critResist", combatRes},
		{"comboResist", combatRes},
		{"counterResist", combatRes},
		{"stunResist", combatRes},
		{"dodgeResist", combatRes},
		{"vampireResist", combatRes},
		{"healBoost", specialAttrs},
		{"critDamageBoost", specialAttrs},
		{"critDamageReduce", specialAttrs},
		{"finalDamageBoost", specialAttrs},
		{"finalDamageReduce", specialAttrs},
		{"combatBoost", specialAttrs},
		{"resistanceBoost", specialAttrs},
	}

	for _, it := range floatKeys {
		if v, ok := petCombat[it.key]; ok {
			it.m[it.key] = it.m[it.key] - v
			if it.m[it.key] < 0 {
				it.m[it.key] = 0
			}
		}
	}
}

// applyPetBonuses 在装备属性调整后重新应用灵宠加成
func applyPetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs map[string]float64, pet *models.Pet, petCombat map[string]float64) {
	// 百分比基础属性加成
	if pet.AttackBonus != 0 {
		baseAttrs["attack"] = baseAttrs["attack"] * (1 + pet.AttackBonus)
	}
	if pet.DefenseBonus != 0 {
		baseAttrs["defense"] = baseAttrs["defense"] * (1 + pet.DefenseBonus)
	}
	if pet.HealthBonus != 0 {
		baseAttrs["health"] = baseAttrs["health"] * (1 + pet.HealthBonus)
	}

	// 基础属性与战斗属性、抗性、特殊属性的数值加成（累加）
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

	floatKeys := []struct {
		key string
		m   map[string]float64
		cap float64
	}{
		{"critRate", combatAttrs, 1},
		{"comboRate", combatAttrs, 1},
		{"counterRate", combatAttrs, 1},
		{"stunRate", combatAttrs, 1},
		{"dodgeRate", combatAttrs, 1},
		{"vampireRate", combatAttrs, 1},
		{"critResist", combatRes, 1},
		{"comboResist", combatRes, 1},
		{"counterResist", combatRes, 1},
		{"stunResist", combatRes, 1},
		{"dodgeResist", combatRes, 1},
		{"vampireResist", combatRes, 1},
		{"healBoost", specialAttrs, -1},
		{"critDamageBoost", specialAttrs, -1},
		{"critDamageReduce", specialAttrs, -1},
		{"finalDamageBoost", specialAttrs, -1},
		{"finalDamageReduce", specialAttrs, -1},
		{"combatBoost", specialAttrs, -1},
		{"resistanceBoost", specialAttrs, -1},
	}

	for _, it := range floatKeys {
		if v, ok := petCombat[it.key]; ok {
			it.m[it.key] = it.m[it.key] + v
			if it.cap > 0 && it.m[it.key] > it.cap {
				it.m[it.key] = it.cap
			}
		}
	}
}

// applyEquipmentStats 将装备属性加到玩家属性上
func applyEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs map[string]float64, equipStats map[string]float64) {
	for stat, v := range equipStats {
		if v == 0 {
			continue
		}
		switch stat {
		case "attack", "health", "defense", "speed":
			baseAttrs[stat] = baseAttrs[stat] + v
		case "critRate", "comboRate", "counterRate", "stunRate", "dodgeRate", "vampireRate":
			combatAttrs[stat] = combatAttrs[stat] + v
		case "critResist", "comboResist", "counterResist", "stunResist", "dodgeResist", "vampireResist":
			combatRes[stat] = combatRes[stat] + v
		default:
			specialAttrs[stat] = specialAttrs[stat] + v
		}
	}
}

// removeEquipmentStats 将装备属性从玩家属性中移除
func removeEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs map[string]float64, equipStats map[string]float64) {
	for stat, v := range equipStats {
		if v == 0 {
			continue
		}
		var m map[string]float64
		switch stat {
		case "attack", "health", "defense", "speed":
			m = baseAttrs
		case "critRate", "comboRate", "counterRate", "stunRate", "dodgeRate", "vampireRate":
			m = combatAttrs
		case "critResist", "comboResist", "counterResist", "stunResist", "dodgeResist", "vampireResist":
			m = combatRes
		default:
			m = specialAttrs
		}
		m[stat] = m[stat] - v
		if m[stat] < 0 {
			m[stat] = 0
		}
	}
}

func boolFrom(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.ToLower(t) == "true"
	default:
		return false
	}
}

func upsertItemOrEquipment(tx *gorm.DB, userID uint, item map[string]interface{}) error {
	typeVal, _ := item["type"].(string)
	isEquipment := typeVal != "pet" && typeVal != "pill" && typeVal != "herb"

	// 生成逻辑ID
	itemID := toString(item["itemId"])
	if itemID == "" {
		itemID = toString(item["id"])
	}
	if itemID == "" {
		itemID = uuid.NewString()
	}

	if isEquipment {
		// 处理为 Equipment
		// 确保 slot 字段
		slot, _ := item["slot"].(string)
		if slot == "" {
			slot = typeVal
		}

		// 查找现有装备
		var existing models.Equipment
		idStr := toString(item["id"])
		q := tx.Where("userId = ? AND (equipmentId = ? OR id = ?)", userID, itemID, idStr)
		err := q.First(&existing).Error
		if err == nil {
			// 更新
			update := map[string]interface{}{
				"userId":      userID,
				"equipmentId": itemID,
				"name":        toString(item["name"]),
				"type":        typeVal,
				"slot":        slot,
				"details":     toJSON(item["details"]),
				"stats":       toJSON(item["stats"]),
				"quality":     toString(item["quality"]),
				"equipped":    boolFrom(item["equipped"]),
			}
			return tx.Model(&existing).Updates(update).Error
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// 创建
		model := models.Equipment{
			ID:          uuid.NewString(),
			UserID:      userID,
			EquipmentID: itemID,
			Name:        toString(item["name"]),
			Type:        typeVal,
		}
		model.Slot = &slot
		model.Details = toJSON(item["details"])
		model.Stats = toJSON(item["stats"])
		model.Quality = toString(item["quality"])
		model.Equipped = boolFrom(item["equipped"])
		return tx.Create(&model).Error
	}

	// 非装备：使用 Items 表
	var existingItem models.Item
	idStr := toString(item["id"])
	err := tx.Where("userId = ? AND (itemId = ? OR id = ?)", userID, itemID, idStr).First(&existingItem).Error
	if err == nil {
		update := map[string]interface{}{
			"userId":   userID,
			"itemId":   itemID,
			"name":     toString(item["name"]),
			"type":     typeVal,
			"details":  toJSON(item["details"]),
			"stats":    toJSON(item["stats"]),
			"quality":  toString(item["quality"]),
			"slot":     item["slot"],
			"equipped": boolFrom(item["equipped"]),
		}
		return tx.Model(&existingItem).Updates(update).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := models.Item{
		ID:     uuid.NewString(),
		UserID: userID,
		ItemID: itemID,
		Name:   toString(item["name"]),
		Type:   typeVal,
	}
	model.Details = toJSON(item["details"])
	model.Stats = toJSON(item["stats"])
	model.Quality = toString(item["quality"])
	model.Equipped = boolFrom(item["equipped"])
	if slot, _ := item["slot"].(string); slot != "" {
		model.Slot = &slot
	}
	return tx.Create(&model).Error
}

func upsertPet(tx *gorm.DB, userID uint, pet map[string]interface{}) error {
	// 构造 petId
	petID := toString(pet["petId"])
	if petID == "" {
		petID = toString(pet["id"])
	}
	if petID == "" {
		petID = uuid.NewString()
	}

	// 处理 quality / combatAttributes
	quality := pet["quality"]
	combatAttrs := pet["combatAttributes"]

	// 如果没有 rarity，从 quality 中尝试提取
	rarity := toString(pet["rarity"])
	if rarity == "" && quality != nil {
		b, _ := json.Marshal(quality)
		var tmp struct {
			Rarity string `json:"rarity"`
		}
		_ = json.Unmarshal(b, &tmp)
		if tmp.Rarity != "" {
			rarity = tmp.Rarity
		}
	}
	if rarity == "" {
		rarity = "mortal"
	}

	var existing models.Pet
	idStr := toString(pet["id"])
	err := tx.Where("userId = ? AND (petId = ? OR id = ?)", userID, petID, idStr).First(&existing).Error
	if err == nil {
		update := map[string]interface{}{
			"userId":           userID,
			"petId":            petID,
			"name":             toString(pet["name"]),
			"type":             toString(pet["type"]),
			"rarity":           rarity,
			"level":            pet["level"],
			"star":             pet["star"],
			"quality":          toJSON(quality),
			"combatAttributes": toJSON(combatAttrs),
			"isActive":         boolFrom(pet["isActive"]),
		}
		return tx.Model(&existing).Updates(update).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := models.Pet{
		ID:     uuid.NewString(),
		UserID: userID,
		PetID:  petID,
		Name:   toString(pet["name"]),
		Type:   toString(pet["type"]),
		Rarity: rarity,
	}
	model.Quality = toJSON(quality)
	model.CombatAttributes = toJSON(combatAttrs)
	model.IsActive = boolFrom(pet["isActive"])
	return tx.Create(&model).Error
}

func upsertHerb(tx *gorm.DB, userID uint, herb map[string]interface{}) error {
	herbID := toString(herb["herbId"])
	if herbID == "" {
		herbID = toString(herb["id"])
	}
	if herbID == "" {
		herbID = uuid.NewString()
	}

	var existing models.Herb
	idStr := toString(herb["id"])
	err := tx.Where("userId = ? AND (herbId = ? OR id = ?)", userID, herbID, idStr).First(&existing).Error
	if err == nil {
		update := map[string]interface{}{
			"userId": userID,
			"herbId": herbID,
			"name":   toString(herb["name"]),
			"count":  herb["count"],
		}
		return tx.Model(&existing).Updates(update).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := models.Herb{
		UserID: userID,
		HerbID: herbID,
		Name:   toString(herb["name"]),
	}
	if cnt, ok := herb["count"].(float64); ok {
		model.Count = int(cnt)
	}
	return tx.Create(&model).Error
}

func upsertPill(tx *gorm.DB, userID uint, pill map[string]interface{}) error {
	pillID := toString(pill["pillId"])
	if pillID == "" {
		pillID = toString(pill["id"])
	}
	if pillID == "" {
		pillID = uuid.NewString()
	}

	var existing models.Pill
	idStr := toString(pill["id"])
	err := tx.Where("userId = ? AND (pillId = ? OR id = ?)", userID, pillID, idStr).First(&existing).Error
	if err == nil {
		update := map[string]interface{}{
			"userId":      userID,
			"pillId":      pillID,
			"name":        toString(pill["name"]),
			"description": toString(pill["description"]),
			"effect":      toJSON(pill["effect"]),
		}
		return tx.Model(&existing).Updates(update).Error
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	model := models.Pill{
		UserID:      userID,
		PillID:      pillID,
		Name:        toString(pill["name"]),
		Description: toString(pill["description"]),
		Effect:      toJSON(pill["effect"]),
	}
	return tx.Create(&model).Error
}
