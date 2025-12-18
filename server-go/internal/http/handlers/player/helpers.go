package player

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/models"
)

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
	// ✅ 改进：先移除数值加成，再移除百分比加成（顺序与应用相反）

	// 第一步：移除 combatAttributes 数值加成（从当前值减去）
	if v, ok := petCombat["attack"]; ok && v != 0 {
		baseAttrs["attack"] = baseAttrs["attack"] - v
		if baseAttrs["attack"] < 0 {
			baseAttrs["attack"] = 0
		}
	}
	if v, ok := petCombat["defense"]; ok && v != 0 {
		baseAttrs["defense"] = baseAttrs["defense"] - v
		if baseAttrs["defense"] < 0 {
			baseAttrs["defense"] = 0
		}
	}
	if v, ok := petCombat["health"]; ok && v != 0 {
		baseAttrs["health"] = baseAttrs["health"] - v
		if baseAttrs["health"] < 0 {
			baseAttrs["health"] = 0
		}
	}
	if v, ok := petCombat["speed"]; ok && v != 0 {
		baseAttrs["speed"] = baseAttrs["speed"] - v
		if baseAttrs["speed"] < 0 {
			baseAttrs["speed"] = 0
		}
	}

	// 第二步：移除百分比基础属性加成
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
		if v, ok := petCombat[it.key]; ok && v != 0 {
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
		return t == "true"
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
				"user_id":      userID,
				"equipment_id": itemID,
				"name":         toString(item["name"]),
				"type":         typeVal,
				"slot":         slot,
				"details":      toJSON(item["details"]),
				"stats":        toJSON(item["stats"]),
				"quality":      toString(item["quality"]),
				"equipped":     boolFrom(item["equipped"]),
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
			"user_id":  userID,
			"item_id":  itemID,
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
			"user_id":          userID,
			"pet_id":           petID,
			"name":             toString(pet["name"]),
			"type":             toString(pet["type"]),
			"rarity":           rarity,
			"level":            pet["level"],
			"star":             pet["star"],
			"quality":          toJSON(quality),
			"combatAttributes": toJSON(combatAttrs),
			"is_active":        boolFrom(pet["is_active"]),
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
	model.IsActive = boolFrom(pet["is_active"])
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
			"user_id": userID,
			"herb_id": herbID,
			"name":    toString(herb["name"]),
			"count":   herb["count"],
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
			"user_id":     userID,
			"pill_id":     pillID,
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
