package player

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)



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

func GetPlayerEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	query := db.DB.Model(&models.Equipment{}).Where("user_id = ?", userID)

	typeParam := c.Query("type")
	if typeParam != "" {
		query = query.Where("equip_type = ?", typeParam)
	} else if equip_type := c.Query("equip_type"); equip_type != "" {
		query = query.Where("equip_type = ?", equip_type)
	}

	if quality := c.Query("quality"); quality != "" {
		query = query.Where("quality = ?", quality)
	}

	equipped := c.Query("equipped")
	if equipped != "" {
		query = query.Where("equipped = ?", equipped == "true")
	}

	var equipment []models.Equipment
	if err := query.Find(&equipment).Error; err != nil {
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

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"equipment": equipment,
	})
}

func EnhanceEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var req struct {
		ReinforceStones int `json:"reinforce_stones"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
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

	stats := map[string]float64{}
	if len(equipment.Stats) > 0 {
		_ = json.Unmarshal(equipment.Stats, &stats)
	}

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

	newRequiredRealm := (equipment.EnhanceLevel / 10) + 1
	equipment.RequiredRealm = newRequiredRealm

	if err := db.DB.Save(&equipment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones - ?", cost)).Error; err != nil {
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
		RefinementStones int `json:"refinement_stones"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
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
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
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
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "装备未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	equipStats := jsonToFloatMap(equipment.Stats)

	if err := db.DB.Model(&models.Equipment{}).
		Where("user_id = ? AND equip_type = ?", userID, equipment.EquipType).
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
	hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil
	petCombat := jsonToFloatMap(activePet.CombatAttributes)

	if hasActivePet {
		removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	applyEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

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
		"message":   "装备穿戴成功",
		"equipment": equipment,
		"user": gin.H{
			"baseAttributes":    baseAttrs,
			"combatAttributes":  combatAttrs,
			"combatResistance":  combatRes,
			"specialAttributes": specialAttrs,
			"reinforce_stones":   user.ReinforceStones,
			"refinement_stones":  user.RefinementStones,
		},
	})
}

func UnequipEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
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
	hasActivePet := db.DB.Where("user_id = ? AND is_active = ?", userID, true).First(&activePet).Error == nil
	petCombat := jsonToFloatMap(activePet.CombatAttributes)

	if hasActivePet {
		removePetBonuses(baseAttrs, combatAttrs, combatRes, specialAttrs, &activePet, petCombat)
	}

	removeEquipmentStats(baseAttrs, combatAttrs, combatRes, specialAttrs, equipStats)

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
			"reinforce_stones":   user.ReinforceStones,
			"refinement_stones":  user.RefinementStones,
		},
	})
}

func SellEquipment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	id := c.Param("id")
	var equipment models.Equipment
	if err := db.DB.Where("id = ? AND user_id = ?", id, userID).First(&equipment).Error; err != nil {
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

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones + ?", stones)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "装备出售成功",
		"stonesReceived": stones,
	})
}

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

	query := db.DB.Model(&models.Equipment{}).Where("user_id = ?", userID)
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

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("reinforce_stones", gorm.Expr("reinforce_stones + ?", totalStones)).Error; err != nil {
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