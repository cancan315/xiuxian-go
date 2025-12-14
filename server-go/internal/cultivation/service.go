package cultivation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"gorm.io/datatypes"
)

// CultivationService 修炼服务
type CultivationService struct {
	userID uint
}

// NewCultivationService 创建修炼服务
func NewCultivationService(userID uint) *CultivationService {
	return &CultivationService{
		userID: userID,
	}
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
func calculateCultivationGain(level int, luck float64, cultivationRate float64) float64 {
	gain := getCurrentCultivationGain(level) * cultivationRate
	// 幸运暴击：30% × luck 概率触发双倍修为
	if rand.Float64() < ExtraCultivationChance*luck {
		gain *= 2
	}
	return gain
}

// SingleCultivate 单次打坐修炼
func (s *CultivationService) SingleCultivate() (*CultivationResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取玩家属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)
	luck := 1.0
	if v, ok := attrs["luck"]; ok {
		if l, ok := v.(float64); ok && l > 0 {
			luck = l
		}
	}

	// 计算当前等级的修炼消耗
	cultivationCost := getCurrentCultivationCost(user.Level)

	// 检查灵力是否充足
	if user.Spirit < cultivationCost {
		return &CultivationResponse{
			Success: false,
			Error:   fmt.Sprintf("灵力不足，需要%.0f，当前%.0f", cultivationCost, user.Spirit),
		}, nil
	}

	// 消耗灵力
	user.Spirit -= cultivationCost

	// 计算修为获得（包含幸运暴击）
	cultivationGain := calculateCultivationGain(user.Level, luck, cultivationRate)
	user.Cultivation += cultivationGain

	// 检查是否需要突破
	var breakthroughResult *BreakthroughResponse
	if user.Cultivation >= user.MaxCultivation {
		breakthroughResult = s.performBreakthrough(&user, &attrs)
	}

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	// 保存数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":          user.Spirit,
		"cultivation":     user.Cultivation,
		"level":           user.Level,
		"realm":           user.Realm,
		"max_cultivation": user.MaxCultivation,
		"base_attributes": user.BaseAttributes,
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

// CultivateUntilBreakthrough 一键突破（快速修炼至突破）
func (s *CultivationService) CultivateUntilBreakthrough() (*CultivationResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查是否已达到最高等级
	if user.Level >= GetMaxLevel() {
		return &CultivationResponse{
			Success: false,
			Error:   "已达到最高境界，无法继续突破",
		}, nil
	}

	// 获取玩家属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)

	// 计算需要的修为和灵力
	remainingCultivation := user.MaxCultivation - user.Cultivation
	gainPerCycle := getCurrentCultivationGain(user.Level) * cultivationRate
	costPerCycle := getCurrentCultivationCost(user.Level)

	if gainPerCycle <= 0 {
		return &CultivationResponse{
			Success: false,
			Error:   "修炼效率异常",
		}, nil
	}

	// 计算需要的修炼次数和总灵力消耗
	times := int(math.Ceil(remainingCultivation / gainPerCycle))
	spiritCost := float64(times) * costPerCycle

	// 检查灵力是否充足
	if user.Spirit < spiritCost {
		return &CultivationResponse{
			Success: false,
			Error:   fmt.Sprintf("灵力不足，突破需要%.0f灵力，当前灵力：%.1f", spiritCost, user.Spirit),
		}, nil
	}

	// 消耗灵力并达到突破条件
	user.Spirit -= spiritCost
	user.Cultivation = user.MaxCultivation

	// 执行突破
	breakthroughResult := s.performBreakthrough(&user, &attrs)
	if breakthroughResult == nil {
		return &CultivationResponse{
			Success: false,
			Error:   "突破失败",
		}, nil
	}

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	// 保存数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":          user.Spirit,
		"cultivation":     user.Cultivation,
		"level":           user.Level,
		"realm":           user.Realm,
		"max_cultivation": user.MaxCultivation,
		"base_attributes": user.BaseAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &CultivationResponse{
		Success:            true,
		CultivationGain:    remainingCultivation,
		SpiritCost:         spiritCost,
		CurrentCultivation: user.Cultivation,
		Breakthrough: map[string]interface{}{
			"newLevel":          breakthroughResult.NewLevel,
			"newRealm":          breakthroughResult.NewRealm,
			"newMaxCultivation": breakthroughResult.NewMaxCultivation,
			"spiritReward":      breakthroughResult.SpiritReward,
			"newSpiritRate":     breakthroughResult.NewSpiritRate,
			"message":           breakthroughResult.Message,
		},
		Message: breakthroughResult.Message,
	}, nil
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
	user.Cultivation = 0 // 重置修为

	// 突破奖励：灵力奖励
	spiritReward := float64(BreakthroughReward * user.Level)
	user.Spirit += spiritReward

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

// GetCultivationData 获取修炼数据
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

	return &CultivationData{
		Level:           user.Level,
		Realm:           user.Realm,
		Cultivation:     user.Cultivation,
		MaxCultivation:  user.MaxCultivation,
		Spirit:          user.Spirit,
		SpiritCost:      spiritCost,
		CultivationGain: cultivationGain,
		CultivationRate: cultivationRate,
		SpiritRate:      spiritRate,
		UnlockedRealms:  unlockedRealms,
	}, nil
}

// AutoCultivate 自动修炼（模拟一段时间的修炼）
func (s *CultivationService) AutoCultivate(durationMs int) (*AutoCultivationResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取玩家属性
	attrs := s.getPlayerAttributes(&user)
	cultivationRate := attrs["cultivationRate"].(float64)
	luck := 1.0
	if v, ok := attrs["luck"]; ok {
		if l, ok := v.(float64); ok && l > 0 {
			luck = l
		}
	}

	// 模拟修炼过程（每秒修炼一次）
	iterations := int(math.Max(1, float64(durationMs)/1000))
	totalCultivationGain := 0.0
	totalSpiritCost := 0.0
	breakthroughCount := 0
	breakthroughDetails := []interface{}{}

	for i := 0; i < iterations; i++ {
		// 计算当前等级的修炼消耗
		cultivationCost := getCurrentCultivationCost(user.Level)

		// 检查灵力是否充足
		if user.Spirit < cultivationCost {
			break
		}

		// 消耗灵力和获得修为（包含幸运暴击）
		user.Spirit -= cultivationCost
		cultivationGain := calculateCultivationGain(user.Level, luck, cultivationRate)
		user.Cultivation += cultivationGain

		totalCultivationGain += cultivationGain
		totalSpiritCost += cultivationCost

		// 检查突破
		if user.Cultivation >= user.MaxCultivation {
			oldRealm := user.Realm
			breakthroughResult := s.performBreakthrough(&user, &attrs)
			if breakthroughResult != nil {
				breakthroughCount++
				breakthroughDetails = append(breakthroughDetails, map[string]interface{}{
					"from":         oldRealm,
					"to":           user.Realm,
					"spiritReward": breakthroughResult.SpiritReward,
				})
			}
		}
	}

	// 保存属性
	s.setPlayerAttributes(&user, attrs)

	// 保存数据
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":          user.Spirit,
		"cultivation":     user.Cultivation,
		"level":           user.Level,
		"realm":           user.Realm,
		"max_cultivation": user.MaxCultivation,
		"base_attributes": user.BaseAttributes,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &AutoCultivationResponse{
		Success:              true,
		TotalCultivationGain: totalCultivationGain,
		TotalSpiritCost:      totalSpiritCost,
		Breakthroughs:        breakthroughCount,
		FinalCultivation:     user.Cultivation,
		BreakthroughDetails:  breakthroughDetails,
		Message:              fmt.Sprintf("自动修炼完成，获得%.0f修为，突破%d次", totalCultivationGain, breakthroughCount),
	}, nil
}
