package cultivation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"

	"gorm.io/datatypes"
)

// CultivationService 修炼服务
type CultivationService struct {
	userID        uint
	spiritGrowMgr interface{ GetPlayerSpiritFromCache(uint) float64 }
}

// NewCultivationService 创建修炼服务
func NewCultivationService(userID uint) *CultivationService {
	return &CultivationService{
		userID: userID,
	}
}

// SetSpiritGrowManager 设置灵力增长管理器（用于读取缓存灵力）
func (s *CultivationService) SetSpiritGrowManager(mgr interface{ GetPlayerSpiritFromCache(uint) float64 }) {
	s.spiritGrowMgr = mgr
}

// getSpiritValue 获取灵力值，优先使用缓存
func (s *CultivationService) getSpiritValue() float64 {
	if s.spiritGrowMgr != nil {
		return s.spiritGrowMgr.GetPlayerSpiritFromCache(s.userID)
	}
	// 降级到数据库查询
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return 0
	}
	return user.Spirit
}

// getPlayerAttributes 获取玩家属性（从BaseAttributes JSON中）
func (s *CultivationService) getPlayerAttributes(user *models.User) map[string]interface{} {
	attrs := make(map[string]interface{})
	if user.BaseAttributes != nil {
		if err := json.Unmarshal(user.BaseAttributes, &attrs); err != nil {
			return map[string]interface{}{"cultivationRate": 1.0, "spiritRate": 1.0}
		}
	}
	// 确保有默认值
	if _, ok := attrs["cultivationRate"]; !ok {
		attrs["cultivationRate"] = 1.0
	}
	if _, ok := attrs["spiritRate"]; !ok {
		attrs["spiritRate"] = 1.0
	}
	return attrs
}

// setPlayerAttributes 保存玩家属性到BaseAttributes JSON
func (s *CultivationService) setPlayerAttributes(user *models.User, attrs map[string]interface{}) error {
	attrJSON, err := json.Marshal(attrs)
	if err != nil {
		return err
	}
	user.BaseAttributes = datatypes.JSON(attrJSON)
	return nil
}

// getCurrentCultivationCost 计算当前等级的修炼消耗
func getCurrentCultivationCost(level int) float64 {
	return float64(BaseCultivationCost) * math.Pow(CultivationCostMultiplier, float64(level-1))
}

// getCurrentCultivationGain 计算当前等级的修炼获得
func getCurrentCultivationGain(level int) float64 {
	return float64(BaseCultivationGain) * math.Pow(CultivationGainMultiplier, float64(level-1))
}

// calculateCultivationGain 计算实际获得的修为（包含幸运暴击）
func calculateCultivationGain(level int, cultivationRate float64) float64 {
	gain := getCurrentCultivationGain(level) * cultivationRate
	r := rand.Float64() // [0,1)

	switch {
	case r < 0.01: // 1%
		gain *= 10
	case r < 0.06: // 5%
		gain *= 5
	case r < 0.16: // 10%
		gain *= 4
	case r < 0.36: // 20%
		gain *= 3
	case r < 0.66: // 30%
		gain *= 2
	}

	// 保留一位小数
	gain = math.Round(gain*10) / 10
	return math.Round(gain*10) / 10
}

// SingleCultivate 单次打坐修炼
func (s *CultivationService) SingleCultivate() (*CultivationResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查打坐间隔（3秒一次）
	lastCultivateKey := fmt.Sprintf("cultivation:lasttime:%d", s.userID)
	lastCultivateTime, err := redis.Client.Get(redis.Ctx, lastCultivateKey).Int64()
	if err == nil && lastCultivateTime > 0 {
		elapsed := time.Now().UnixMilli() - lastCultivateTime
		const cultivateInterval = 1000 // 1秒（毫秒）
		if elapsed < cultivateInterval {
			waitTime := time.Duration(cultivateInterval-elapsed) * time.Millisecond
			return &CultivationResponse{
				Success: false,
				Error:   fmt.Sprintf("丹田正在提炼灵力，请等待%.1f秒再次修炼", float64(waitTime.Milliseconds())/1000),
			}, nil
		}
	}

	// 记录本次打坐的时间
	redis.Client.Set(redis.Ctx, lastCultivateKey, time.Now().UnixMilli(), 24*time.Hour)

	// 获取玩家属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)

	// 计算当前等级的修炼消耗
	cultivationCost := getCurrentCultivationCost(user.Level)

	// ✅ 检查灵力时，优先使用缓存值
	currentSpirit := s.getSpiritValue()
	if currentSpirit < cultivationCost {
		return &CultivationResponse{
			Success: false,
			Error:   fmt.Sprintf("灵力不足，需要%.0f，当前%.0f", cultivationCost, currentSpirit),
		}, nil
	}

	// 消耗灵力
	user.Spirit -= cultivationCost

	// 计算修为获得（包含幸运暴击）
	cultivationGain := calculateCultivationGain(user.Level, cultivationRate)
	user.Cultivation = math.Round((user.Cultivation+cultivationGain)*10) / 10
	// 检查是否需要突破
	// Level 27（金丹期九层）不自动突破，需要手动突破。
	var breakthroughResult *BreakthroughResponse
	if user.Cultivation >= user.MaxCultivation && user.Level != 27 {
		breakthroughResult = s.performBreakthrough(&user, &attrs)
	}

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	// 保存数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":             user.Spirit,
		"cultivation":        user.Cultivation,
		"level":              user.Level,
		"realm":              user.Realm,
		"max_cultivation":    user.MaxCultivation,
		"base_attributes":    user.BaseAttributes,
		"combat_attributes":  user.CombatAttributes,
		"combat_resistance":  user.CombatResistance,
		"special_attributes": user.SpecialAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	resp := &CultivationResponse{
		Success:            true,
		CultivationGain:    cultivationGain,
		SpiritCost:         cultivationCost,
		CurrentCultivation: user.Cultivation,
	}

	if breakthroughResult != nil {
		resp.Breakthrough = map[string]interface{}{
			"newLevel":          breakthroughResult.NewLevel,
			"newRealm":          breakthroughResult.NewRealm,
			"newMaxCultivation": breakthroughResult.NewMaxCultivation,
			"spiritReward":      breakthroughResult.SpiritReward,
			"newSpiritRate":     breakthroughResult.NewSpiritRate,
			"message":           breakthroughResult.Message,
		}
		resp.Message = breakthroughResult.Message
	}

	return resp, nil
}

// performBreakthrough 执行突破逻辑
func (s *CultivationService) performBreakthrough(user *models.User, attrs *map[string]interface{}) *BreakthroughResponse {
	// 获取下一个境界
	nextRealm := GetNextRealm(user.Level)
	if nextRealm == nil || user.Level >= GetMaxLevel() {
		return nil
	}

	// 更新用户信息
	user.Level = nextRealm.Level
	user.Realm = nextRealm.Name
	user.MaxCultivation = nextRealm.MaxCultivation
	user.Cultivation = 0.0 // 重置修为

	// 突破奖励：灵力奖励
	spiritReward := float64(BreakthroughReward * user.Level)
	user.Spirit += spiritReward

	// ✅ 新增：升级时重新计算玩家属性
	s.reinitializePlayerAttributes(user, attrs)

	// 突破奖励：提升灵力获取速率
	newSpiritRate := 1.0
	if spiritRate, ok := (*attrs)["spiritRate"].(float64); ok {
		newSpiritRate = spiritRate * BreakthroughBonus
		(*attrs)["spiritRate"] = newSpiritRate
	} else {
		newSpiritRate = BreakthroughBonus
		(*attrs)["spiritRate"] = newSpiritRate
	}

	// 更新解锁的境界
	s.unlockRealm(user, attrs)

	return &BreakthroughResponse{
		Success:           true,
		NewLevel:          user.Level,
		NewRealm:          user.Realm,
		NewMaxCultivation: user.MaxCultivation,
		SpiritReward:      spiritReward,
		NewSpiritRate:     newSpiritRate,
		Message:           fmt.Sprintf("恭喜突破至 %s！", user.Realm),
	}
}

// reinitializePlayerAttributes 升级时重新初始化玩家属性
// 步骤：
// 1. 重新计算基础属性（基于新等级）
// 2. 重新应用装备加成
// 3. 重新应用灵宠加成
func (s *CultivationService) reinitializePlayerAttributes(user *models.User, attrs *map[string]interface{}) {
	level := user.Level

	// 步骤1：重新计算基础属性
	baseAttrs := map[string]interface{}{
		"speed":   float64(10 * level),
		"attack":  float64(10 * level),
		"health":  float64(100 * level),
		"defense": float64(5 * level),
	}

	// 计算灵力速率
	spiritRate := 1.0 * math.Pow(1.2, float64(level-1))
	spiritRate = math.Round(spiritRate*100) / 100
	baseAttrs["spiritRate"] = spiritRate

	// 保留cultivationRate
	if cr, ok := (*attrs)["cultivationRate"]; ok {
		baseAttrs["cultivationRate"] = cr
	} else {
		baseAttrs["cultivationRate"] = 1.0
	}

	// 保留unlockedRealms
	if ur, ok := (*attrs)["unlockedRealms"]; ok {
		baseAttrs["unlockedRealms"] = ur
	}

	// 初始化战斗属性
	combatAttrs := map[string]float64{
		"critRate":     0,
		"critDamage":   0,
		"comboRate":    0,
		"stunRate":     0,
		"lifesteal":    0,
		"attackBonus":  0,
		"defenseBonus": 0,
	}

	// 初始化抗性属性
	combatRes := map[string]float64{
		"critResist":  0,
		"comboResist": 0,
		"stunResist":  0,
	}

	// 初始化特殊属性
	specialAttrs := map[string]float64{
		"luck":        0,
		"spiritBonus": 0,
	}

	// 步骤2：重新应用装备加成
	var equipments []models.Equipment
	if err := db.DB.Where("user_id = ? AND equipped = ?", user.ID, true).Find(&equipments).Error; err == nil {
		for _, equipment := range equipments {
			s.applyEquipmentStats(&baseAttrs, &combatAttrs, &combatRes, &specialAttrs, &equipment)
		}
	}

	// 步骣3：重新应用灵宠加成
	var pets []models.Pet
	if err := db.DB.Where("user_id = ? AND is_active = ?", user.ID, true).Find(&pets).Error; err == nil {
		for _, pet := range pets {
			s.applyPetStats(&baseAttrs, &combatAttrs, &combatRes, &specialAttrs, &pet)
		}
	}

	// 更新用户属性JSON
	baseJSON, _ := json.Marshal(baseAttrs)
	combatJSON, _ := json.Marshal(combatAttrs)
	resJSON, _ := json.Marshal(combatRes)
	specialJSON, _ := json.Marshal(specialAttrs)

	user.BaseAttributes = datatypes.JSON(baseJSON)
	user.CombatAttributes = datatypes.JSON(combatJSON)
	user.CombatResistance = datatypes.JSON(resJSON)
	user.SpecialAttributes = datatypes.JSON(specialJSON)

	// 更新attrs以便之后使用
	*attrs = baseAttrs
}

// applyEquipmentStats 应用装备属性加成
func (s *CultivationService) applyEquipmentStats(baseAttrs *map[string]interface{}, combatAttrs, combatRes, specialAttrs *map[string]float64, equipment *models.Equipment) {
	var stats map[string]interface{}
	if equipment.Stats != nil {
		json.Unmarshal(equipment.Stats, &stats)
	}

	for key, value := range stats {
		if v, ok := value.(float64); ok {
			switch key {
			case "speed", "attack", "health", "defense":
				if current, ok := (*baseAttrs)[key].(float64); ok {
					(*baseAttrs)[key] = current + v
				}
			case "critRate", "critDamage", "comboRate", "stunRate", "lifesteal", "attackBonus", "defenseBonus":
				(*combatAttrs)[key] += v
			case "critResist", "comboResist", "stunResist":
				(*combatRes)[key] += v
			case "luck", "spiritBonus":
				(*specialAttrs)[key] += v
			}
		}
	}
}

// applyPetStats 应用灵宠属性加成
func (s *CultivationService) applyPetStats(baseAttrs *map[string]interface{}, combatAttrs, combatRes, specialAttrs *map[string]float64, pet *models.Pet) {
	var petCombat map[string]interface{}
	if pet.CombatAttributes != nil {
		json.Unmarshal(pet.CombatAttributes, &petCombat)
	}

	for key, value := range petCombat {
		if v, ok := value.(float64); ok {
			switch key {
			case "speed", "attack", "health", "defense":
				if current, ok := (*baseAttrs)[key].(float64); ok {
					(*baseAttrs)[key] = current + v
				}
			case "critRate", "critDamage", "comboRate", "stunRate", "lifesteal", "attackBonus", "defenseBonus":
				(*combatAttrs)[key] += v
			case "critResist", "comboResist", "stunResist":
				(*combatRes)[key] += v
			case "luck", "spiritBonus":
				(*specialAttrs)[key] += v
			}
		}
	}
}

// unlockRealm 解锁境界（更新数据库）
func (s *CultivationService) unlockRealm(user *models.User, attrs *map[string]interface{}) error {
	// 获取已解锁的境界列表
	var unlockedRealms []string
	if v, ok := (*attrs)["unlockedRealms"]; ok {
		if realms, ok := v.([]interface{}); ok {
			for _, r := range realms {
				if realm, ok := r.(string); ok {
					unlockedRealms = append(unlockedRealms, realm)
				}
			}
		}
	}

	// 添加新的境界
	found := false
	for _, r := range unlockedRealms {
		if r == user.Realm {
			found = true
			break
		}
	}
	if !found {
		unlockedRealms = append(unlockedRealms, user.Realm)
	}

	// 保存回属性
	(*attrs)["unlockedRealms"] = unlockedRealms

	return nil
}

// getCurrentFormationCost 计算当前等级的聚灵阵消耗
func getCurrentFormationCost(level int) int {
	return int(float64(BaseFormationCost) * math.Pow(FormationCostMultiplier, float64(level-1)))
}

// getCurrentFormationGain 计算当前等级的聚灵阵获得
func getCurrentFormationGain(level int) float64 {
	return float64(BaseFormationGain) * math.Pow(FormationGainMultiplier, float64(level-1))
}

// UseFormation 使用聚灵阵
// ✅ 新增逻辑：
// 1. 前10次消耗灵石数量降低10倍
// 2. 等级>27时，固定消耗10000灵石，概率增加MaxCultivation百分比
func (s *CultivationService) UseFormation() (*FormationResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查聚灵阵间隔（1秒一次）
	lastFormationKey := fmt.Sprintf("formation:lasttime:%d", s.userID)
	lastFormationTime, err := redis.Client.Get(redis.Ctx, lastFormationKey).Int64()
	if err == nil && lastFormationTime > 0 {
		elapsed := time.Now().UnixMilli() - lastFormationTime
		const formationInterval = FormationInterval // 1秒（毫秒）
		if elapsed < formationInterval {
			waitTime := time.Duration(formationInterval-elapsed) * time.Millisecond
			return &FormationResponse{
				Success: false,
				Error:   fmt.Sprintf("聚灵阵正在运转，请等待%.1f秒再次使用", float64(waitTime.Milliseconds())/1000),
			}, nil
		}
	}

	// 记录本次聚灵阵的时间
	redis.Client.Set(redis.Ctx, lastFormationKey, time.Now().UnixMilli(), 24*time.Hour)

	// 获取玩家属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)

	// ✅ 高等级特殊逻辑：等级>27时使用完全不同的机制
	if user.Level > 27 {
		return s.useFormationHighLevel(&user, &attrs)
	}

	// ========== 普通等级逻辑（等级<=27）==========

	// 计算当前等级的聚灵阵消耗
	formationCost := getCurrentFormationCost(user.Level)

	// ✅ 每日前10次使用聚灵阵时降低消耗10倍
	todayStr := time.Now().Format("2006-01-02")
	formationCountKey := fmt.Sprintf("formation:dailycount:%d:%s", s.userID, todayStr)
	dailyFormationCount := redis.Client.Incr(redis.Ctx, formationCountKey).Val()
	redis.Client.Expire(redis.Ctx, formationCountKey, 48*time.Hour) // 48小时后过期，确保跨日清理

	if dailyFormationCount <= 10 {
		formationCost = (formationCost + 9) / 10 // 向上取整后除以10
	}

	// 检查灵石是否足够
	if user.SpiritStones < formationCost {
		return &FormationResponse{
			Success: false,
			Error:   fmt.Sprintf("灵石不足，需要%d，当前%d", formationCost, user.SpiritStones),
		}, nil
	}

	// 消耗灵石
	user.SpiritStones -= formationCost

	// 计算修为获得（包含多档幸运暴击）
	formationGain := getCurrentFormationGain(user.Level) * cultivationRate

	r := rand.Float64() // [0,1)

	switch {
	case r < 0.01: // 1%
		formationGain *= 10
	case r < 0.06: // 5%
		formationGain *= 5
	case r < 0.16: // 10%
		formationGain *= 4
	case r < 0.36: // 20%
		formationGain *= 3
	case r < 0.66: // 30%
		formationGain *= 2
	}

	// 保留一位小数
	formationGain = math.Round(formationGain*10) / 10
	user.Cultivation = math.Round((user.Cultivation+formationGain)*10) / 10

	// 检查是否需要突破
	// Level 27（金丹期九层）不自动突破，需要手动突破。
	var breakthroughResult *BreakthroughResponse
	if user.Cultivation >= user.MaxCultivation && user.Level != 27 {
		breakthroughResult = s.performBreakthrough(&user, &attrs)
	}

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	// 保存数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit_stones":      user.SpiritStones,
		"cultivation":        user.Cultivation,
		"level":              user.Level,
		"realm":              user.Realm,
		"max_cultivation":    user.MaxCultivation,
		"base_attributes":    user.BaseAttributes,
		"combat_attributes":  user.CombatAttributes,
		"combat_resistance":  user.CombatResistance,
		"special_attributes": user.SpecialAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	resp := &FormationResponse{
		Success:            true,
		CultivationGain:    formationGain,
		StoneCost:          formationCost,
		CurrentCultivation: user.Cultivation,
	}

	if breakthroughResult != nil {
		resp.Breakthrough = map[string]interface{}{
			"newLevel":          breakthroughResult.NewLevel,
			"newRealm":          breakthroughResult.NewRealm,
			"newMaxCultivation": breakthroughResult.NewMaxCultivation,
			"spiritReward":      breakthroughResult.SpiritReward,
			"newSpiritRate":     breakthroughResult.NewSpiritRate,
			"message":           breakthroughResult.Message,
		}
		resp.Message = breakthroughResult.Message
	}

	return resp, nil
}

// useFormationHighLevel 高等级聚灵阵逻辑（等级>27）
// 固定消耗10000灵石，概率增加MaxCultivation百分比
// 55%概率增加1%, 30%概率增加2%, 10%概率增加2.5%, 5%概率增加3%
func (s *CultivationService) useFormationHighLevel(user *models.User, attrs *map[string]interface{}) (*FormationResponse, error) {
	const highLevelFormationCost = 10000 // 固定消耗10000灵石

	// 检查灵石是否足够
	if user.SpiritStones < highLevelFormationCost {
		return &FormationResponse{
			Success: false,
			Error:   fmt.Sprintf("灵石不足，需要%d，当前%d", highLevelFormationCost, user.SpiritStones),
		}, nil
	}

	// 消耗灵石
	user.SpiritStones -= highLevelFormationCost

	// 概率增加MaxCultivation百分比
	// 55% -> 1%, 30% -> 2%, 10% -> 2.5%, 5% -> 3%
	r := rand.Float64() // [0,1)
	var maxCultivationRate float64

	switch {
	case r < 0.55: // 55%
		maxCultivationRate = 0.01 // 1%
	case r < 0.85: // 30% (0.55 + 0.30)
		maxCultivationRate = 0.02 // 2%
	case r < 0.95: // 10% (0.85 + 0.10)
		maxCultivationRate = 0.025 // 2.5%
	default: // 5%
		maxCultivationRate = 0.03 // 3%
	}

	// 计算增加的MaxCultivation
	maxCultivationGain := user.MaxCultivation * maxCultivationRate
	maxCultivationGain = math.Round(maxCultivationGain*10) / 10
	user.MaxCultivation = math.Round((user.MaxCultivation+maxCultivationGain)*10) / 10

	// 保存属性
	s.setPlayerAttributes(user, *attrs)

	// 保存数据
	if err := db.DB.Model(user).Updates(map[string]interface{}{
		"spirit_stones":   user.SpiritStones,
		"max_cultivation": user.MaxCultivation,
		"base_attributes": user.BaseAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &FormationResponse{
		Success:            true,
		CultivationGain:    0, // 高等级不增加修为，只增加上限
		StoneCost:          highLevelFormationCost,
		CurrentCultivation: user.Cultivation,
		MaxCultivationGain: maxCultivationGain,
		NewMaxCultivation:  user.MaxCultivation,
		MaxCultivationRate: maxCultivationRate * 100, // 转换为百分比显示
		Message:            fmt.Sprintf("聚灵阵威能增幅！修为上限+%.1f (%.1f%%)", maxCultivationGain, maxCultivationRate*100),
	}, nil
}

// GetCultivationData 获取修炼数据
// ✅ 改进：包括并重新计算已穿戴装备和出战灵宠的属性加成
func (s *CultivationService) GetCultivationData() (*CultivationData, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 解析属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)
	spiritRate := attrs["spiritRate"].(float64)

	// 计算修炼消耗和获得
	spiritCost := getCurrentCultivationCost(user.Level)
	cultivationGain := getCurrentCultivationGain(user.Level)

	// ✅ 新增：计算聚灵阵消耗和获得，并添加到属性中
	formationCost := getCurrentFormationCost(user.Level)
	formationGain := getCurrentFormationGain(user.Level)
	// 将聚灵阵相关数据添加到基础属性中
	attrs["formationLevel"] = user.Level // 使用玩家等级作为聚灵阵等级参考
	attrs["formationGain"] = formationGain
	attrs["formationCost"] = formationCost

	// 解析已解锁的境界
	unlockedRealms := []string{}
	if v, ok := attrs["unlockedRealms"]; ok {
		if realms, ok := v.([]interface{}); ok {
			for _, r := range realms {
				if realm, ok := r.(string); ok {
					unlockedRealms = append(unlockedRealms, realm)
				}
			}
		}
	}

	// ✅ 新增：解析玩家的四大属性
	baseAttributes := make(map[string]interface{})
	combatAttributes := make(map[string]interface{})
	combatResistance := make(map[string]interface{})
	specialAttributes := make(map[string]interface{})

	if user.BaseAttributes != nil {
		json.Unmarshal(user.BaseAttributes, &baseAttributes)
	}
	if user.CombatAttributes != nil {
		json.Unmarshal(user.CombatAttributes, &combatAttributes)
	}
	if user.CombatResistance != nil {
		json.Unmarshal(user.CombatResistance, &combatResistance)
	}
	if user.SpecialAttributes != nil {
		json.Unmarshal(user.SpecialAttributes, &specialAttributes)
	}

	return &CultivationData{
		PlayerName:      user.PlayerName,
		Level:           user.Level,
		Realm:           user.Realm,
		Cultivation:     user.Cultivation,
		MaxCultivation:  user.MaxCultivation,
		Spirit:          user.Spirit,
		SpiritCost:      spiritCost,
		CultivationGain: cultivationGain,
		CultivationRate: cultivationRate,
		SpiritRate:      spiritRate,
		SpiritStones:    user.SpiritStones,
		ReinforceStones: user.ReinforceStones,
		// ✅ 新增：洗练石和灵宠精华
		RefinementStones: user.RefinementStones,
		PetEssence:       user.PetEssence,
		UnlockedRealms:   unlockedRealms,
		// ✅ 新增：返回属性数据（包含聚灵阵相关数据）
		BaseAttributes:    attrs, // 使用更新后的attrs，包含聚灵阵数据
		CombatAttributes:  combatAttributes,
		CombatResistance:  combatResistance,
		SpecialAttributes: specialAttributes,
	}, nil
}

// ✅ 新增：InitializePlayerAttributesOnLevel 根据等级初始化玩家属性
// 用于登录时重新计算属性
func InitializePlayerAttributesOnLevel(level int) map[string]interface{} {
	return map[string]interface{}{
		"speed":           float64(10 * level),
		"attack":          float64(10 * level),
		"health":          float64(100 * level),
		"defense":         float64(5 * level),
		"spiritRate":      CalculateSpiritRateByLevel(level), // 使用公共函数
		"cultivationRate": 1.0,
	}
}

// ✅ 新增：CalculateSpiritRateByLevel 计算基于等级的灵力倍率
// 导出函数供其他模块使用
// 公式：spiritRate = 1.0 * (1.2)^(Level-1)
func CalculateSpiritRateByLevel(level int) float64 {
	if level < 1 {
		level = 1
	}
	spiritRate := 1.0 * math.Pow(1.2, float64(level-1))
	// 保留两位小数
	return math.Round(spiritRate*100) / 100
}

// ✅ 新增：CalculateBaseAttributesByLevel 计算基于等级的基础属性
// 导出函数供其他模块使用
// 公式：
// speed = 10 * Level
// attack = 10 * Level
// health = 100 * Level
// defense = 5 * Level
func CalculateBaseAttributesByLevel(level int) map[string]interface{} {
	return map[string]interface{}{
		"speed":   float64(10 * level),
		"attack":  float64(10 * level),
		"health":  float64(100 * level),
		"defense": float64(5 * level),
	}
}

// BreakthroughJieYing 结婴突破处理
func (s *CultivationService) BreakthroughJieYing() (map[string]interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查是否满足结婴条件
	if user.Level != 27 || user.Cultivation < user.MaxCultivation {
		return map[string]interface{}{
			"success": false,
			"message": "不满足结婴条件",
		}, nil
	}

	// 从 BaseAttributes 中读取 duJieRate
	attrs := s.getPlayerAttributes(&user)
	duJieRate := 0.05 // 默认值
	if rate, ok := attrs["duJieRate"].(float64); ok {
		duJieRate = rate
	}

	// 判断渡劫是否成功
	randomVal := rand.Float64()
	success := randomVal < duJieRate

	if success {
		// 结婴成功，晋升为元婴期
		nextRealm := GetNextRealm(user.Level)
		if nextRealm == nil {
			return map[string]interface{}{
				"success": false,
				"message": "已是最高境界",
			}, nil
		}

		user.Level = nextRealm.Level
		user.Realm = nextRealm.Name
		user.MaxCultivation = nextRealm.MaxCultivation
		user.Cultivation = 0.0
		// 重置 duJieRate 为 5%
		attrs["duJieRate"] = 0.05

		// ✅ 升级时重新计算玩家属性
		s.reinitializePlayerAttributes(&user, &attrs)

		// 保存属性
		s.setPlayerAttributes(&user, attrs)

		// 保存数据
		if err := db.DB.Model(&user).Updates(map[string]interface{}{
			"level":              user.Level,
			"realm":              user.Realm,
			"cultivation":        user.Cultivation,
			"max_cultivation":    user.MaxCultivation,
			"base_attributes":    user.BaseAttributes,
			"combat_attributes":  user.CombatAttributes,
			"combat_resistance":  user.CombatResistance,
			"special_attributes": user.SpecialAttributes,
		}).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return map[string]interface{}{
			"success":   true,
			"message":   "突破成功，恭喜道友修炼千年，成为元婴道君",
			"newRealm":  user.Realm,
			"newLevel":  user.Level,
			"duJieRate": 0.05,
		}, nil
	}

	// 结婴失败，金丹破碎
	user.Cultivation = 0.0
	// 重置 duJieRate 为 0
	attrs["duJieRate"] = 0.0

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"cultivation":     user.Cultivation,
		"base_attributes": user.BaseAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return map[string]interface{}{
		"success": false,
		"message": "突破失败，金丹破碎，请道友夯实修为，磨炼道心后再行突破",
	}, nil
}
