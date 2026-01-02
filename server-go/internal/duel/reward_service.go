package duel

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/exploration"
	"xiuxian/server-go/internal/models"

	"gorm.io/gorm"
)

// PvPRewards PvP战斗奖励结果
type PvPRewards struct {
	SpiritStones int64 `json:"spirit_stones"`
	Cultivation  int64 `json:"cultivation"`
}

// PvERewards PvE战斗奖励结果（仅灵草）
type PvERewards struct {
	HerbID  string `json:"herb_id"` // 灵草ID
	Name    string `json:"name"`    // 灵草名称
	Count   int    `json:"count"`   // 灵草数量
	Quality string `json:"quality"` // 灵草品质
}

// DemonSlayingRewards 除魔卫道战斗奖励结果（灵石、修为、丹方残页）
type DemonSlayingRewards struct {
	SpiritStones     int64  `json:"spirit_stones"`      // 灵石奖励
	Cultivation      int64  `json:"cultivation"`        // 修为奖励
	PillFragmentID   string `json:"pill_fragment_id"`   // 丹方残页ID（可为空）
	PillFragmentName string `json:"pill_fragment_name"` // 丹方名称
}

// RewardService 奖励服务
type RewardService struct {
	config *RewardConfig
}

// NewRewardService 创建奖励服务实例
func NewRewardService(config *RewardConfig) *RewardService {
	if config == nil {
		config = DefaultRewardConfig()
	}
	return &RewardService{config: config}
}

// CalculateRewards 计算PvP战斗奖励
// 需要传入玩家等级以计算基于等级的灵石和修为奖励
func (rs *RewardService) CalculateRewards(status *PvPBattleStatus, playerLevel int) *PvPRewards {
	spiritStones := rs.calculateSpiritStoneReward(status.Round, playerLevel)
	cultivation := rs.calculateCultivationReward(status.Round, playerLevel)

	return &PvPRewards{
		SpiritStones: spiritStones,
		Cultivation:  cultivation,
	}
}

// CalculateRewardsForPvE 计算降伏妖兽战斗奖励（仅灵草）
// difficulty: normal(普通), hard(困难), boss(噩梦)
func (rs *RewardService) CalculateRewardsForPvE(status *PvEBattleStatus, playerLevel int, difficulty string) *PvERewards {
	// 70% 概率不奖励任何物品
	if rand.Float64() < 0.7 {
		return nil
	}

	// 从配置中随机选择一种灵草
	if len(exploration.HerbConfigs) == 0 {
		log.Printf("[Reward] 灵草配置为空")
		return nil
	}

	// 随机选择一种灵草
	randomIndex := rand.Intn(len(exploration.HerbConfigs))
	herbConfig := exploration.HerbConfigs[randomIndex]

	// 随机生成灵草品质
	quality := exploration.GetRandomQuality(rand.Float64())

	// 使用降伏妖兽专用的难度倍数计算
	multiplier := rs.getMonsterDifficultyMultiplier(difficulty)
	herbCount := int(math.Round(multiplier))

	log.Printf("[Reward] 计算降伏妖兽奖励 - 难度: %s, 倍数: %.2f, 灵草: %s, 数量: %d, 品质: %s",
		difficulty, multiplier, herbConfig.Name, herbCount, quality)

	// 每次战斗奖励的灵草数量根据难度倍增（四舍五入取整）
	return &PvERewards{
		HerbID:  herbConfig.ID,
		Name:    herbConfig.Name,
		Count:   herbCount,
		Quality: quality,
	}
}

// CalculateRewardsForDemonSlaying 计算除魔卫道战斗奖励（灵石、修为、丹方残页）
// difficulty: normal(普通), hard(困难), boss(噩梦)
func (rs *RewardService) CalculateRewardsForDemonSlaying(status *PvEBattleStatus, playerLevel int, difficulty string) *DemonSlayingRewards {
	// 基础灵石奖励（比PvP少一些）
	baseSpiritStones := rs.calculateSpiritStoneReward(status.Round, playerLevel) / 2
	// 基础修为奖励（比PvP少一些）
	baseCultivation := rs.calculateCultivationReward(status.Round, playerLevel) / 2

	// 使用除魔卫道专用的难度倍数计算
	multiplier := rs.getDemonSlayingDifficultyMultiplier(difficulty)

	rewards := &DemonSlayingRewards{
		SpiritStones: int64(math.Round(float64(baseSpiritStones) * multiplier)),
		Cultivation:  int64(math.Round(float64(baseCultivation) * multiplier)),
	}

	log.Printf("[Reward] 计算除魔卫道奖励 - 玩家等级: %d, 难度: %s, 倍数: %.2f, 灵石: %d, 修为: %d",
		playerLevel, difficulty, multiplier, rewards.SpiritStones, rewards.Cultivation)

	// 20% 概率获得丹方残页
	if rand.Float64() < 0.2 {
		// 从探索配置中随机选择一个丹方（按权重）
		var activeRecipes []exploration.PillRecipe
		for _, recipe := range exploration.PillRecipes {
			if recipe.Weight > 0 {
				activeRecipes = append(activeRecipes, recipe)
			}
		}

		if len(activeRecipes) > 0 {
			// 计算权重总和
			totalWeight := 0.0
			for _, recipe := range activeRecipes {
				totalWeight += recipe.Weight
			}

			// 在 [0, totalWeight) 区间内取随机数
			rnd := rand.Float64() * totalWeight
			sum := 0.0
			for _, recipe := range activeRecipes {
				sum += recipe.Weight
				if rnd < sum {
					rewards.PillFragmentID = recipe.ID
					rewards.PillFragmentName = recipe.Name
					break
				}
			}
		}
	}

	return rewards
}

// calculateSpiritStoneReward 计算灵石奖励
// 根据玩家等级和战斗回合数计算灵石奖励
// 基础奖励公式：BaseFormula.Base * (BaseFormula.Multiplier ^ (level-1))
func (rs *RewardService) calculateSpiritStoneReward(round int, playerLevel int) int64 {
	cfg := rs.config.SpiritStones

	// 使用公式计算基础奖励：50 * 1.05^(level-1)
	baseReward := int64(cfg.BaseFormula.Base * math.Pow(cfg.BaseFormula.Multiplier, float64(playerLevel-1)))

	// 根据回合数调整奖励(回合越少奖励越多)
	roundBonus := int64(math.Max(0, cfg.RoundBonus-float64(round)/20.0))
	if roundBonus > cfg.MaxRoundBonus {
		roundBonus = cfg.MaxRoundBonus
	}

	return baseReward + roundBonus
}

// calculateCultivationReward 计算修为奖励
// 根据玩家等级和战斗回合数计算修为奖励
// 基础奖励公式：BaseFormula.Base * (BaseFormula.Multiplier ^ (level-1))
func (rs *RewardService) calculateCultivationReward(round int, playerLevel int) int64 {
	cfg := rs.config.Cultivation

	// 使用公式计算基础奖励：504 * 1.05^(level-1)
	baseReward := int64(cfg.BaseFormula.Base * math.Pow(cfg.BaseFormula.Multiplier, float64(playerLevel-1)))

	// 根据回合数调整奖励
	roundBonus := int64(math.Max(0, cfg.RoundBonus-float64(round)/20.0))
	if roundBonus > cfg.MaxRoundBonus {
		roundBonus = cfg.MaxRoundBonus
	}

	return baseReward + roundBonus
}

// ApplyRewardMultiplier 应用奖励倍率
func (rs *RewardService) ApplyRewardMultiplier(rewards *PvPRewards) *PvPRewards {
	multiplier := rs.calculateRewardMultiplier()

	return &PvPRewards{
		SpiritStones: int64(float64(rewards.SpiritStones) * multiplier),
		Cultivation:  int64(float64(rewards.Cultivation) * multiplier),
	}
}

// calculateRewardMultiplier 计算奖励倍率
func (rs *RewardService) calculateRewardMultiplier() float64 {
	r := rand.Float64() // [0.0, 1.0)

	// 遍历概率配置，找到对应的倍率
	for threshold, multiplier := range rs.config.Multiplier.Probabilities {
		if r < threshold {
			return multiplier
		}
	}

	// 默认返回1.0倍率
	return 1.0
}

// getDifficultyMultiplier 根据难度获取随机奖励倍数（通用）
// normal(普通)=1倍(固定), hard(困难)=1-2個随机), boss(噩梦)=1-3倍(随机)
func (rs *RewardService) getDifficultyMultiplier(difficulty string) float64 {
	switch difficulty {
	case "normal":
		// 普通难度：固定1倍
		return 1.0
	case "hard":
		// 困难难度：1-2倍随机
		return 1.0 + rand.Float64() // [1.0, 2.0)
	case "boss":
		// 噩梦难度：1-3倍随机
		return 1.0 + rand.Float64()*2.0 // [1.0, 3.0)
	default:
		// 默认返回1倍
		return 1.0
	}
}

// getMonsterDifficultyMultiplier 降伏妖兽难度倍数（仅影响灵草数量）
// normal(普通)=1倍, hard(困难)=1-2倍随机, boss(噩梦)=1-3倍随机
func (rs *RewardService) getMonsterDifficultyMultiplier(difficulty string) float64 {
	switch difficulty {
	case "normal":
		return 1.0
	case "hard":
		return 1.0 + rand.Float64()
	case "boss":
		return 1.0 + rand.Float64()*2.0
	default:
		return 1.0
	}
}

// getDemonSlayingDifficultyMultiplier 除魔卫道难度倍数（影响灵石和修为）
// normal(普通)=1倍, hard(困难)=1-2倍随机, boss(噩梦)=1-3倍随机
func (rs *RewardService) getDemonSlayingDifficultyMultiplier(difficulty string) float64 {
	switch difficulty {
	case "normal":
		return 1.0
	case "hard":
		return 1.0 + rand.Float64()
	case "boss":
		return 1.0 + rand.Float64()*2.0
	default:
		return 1.0
	}
}

// GrantRewardsToPlayer 将奖励发放给玩家
func (rs *RewardService) GrantRewardsToPlayer(playerID int64, rewards *PvPRewards) error {
	var user models.User
	if err := db.DB.First(&user, playerID).Error; err != nil {
		return fmt.Errorf("获取玩家信息失败: %w", err)
	}

	// 更新灵石
	user.SpiritStones += int(rewards.SpiritStones)

	// 更新修为
	user.Cultivation += float64(rewards.Cultivation)
	if user.Cultivation > user.MaxCultivation {
		user.Cultivation = user.MaxCultivation
	}

	// 保存到数据库
	if err := db.DB.Model(&user).
		Update("spirit_stones", user.SpiritStones).
		Update("cultivation", user.Cultivation).Error; err != nil {
		log.Printf("[Reward] 更新玩家资源失败: %v", err)
		return fmt.Errorf("更新玩家资源失败: %w", err)
	}

	log.Printf("[Reward] 玩家 %d 获得奖励 - 灵石: %d, 修为: %d",
		playerID, rewards.SpiritStones, rewards.Cultivation)

	return nil
}

// GrantRewardsToPlayerWithMultiplier 计算并发放包含倍率的奖励
func (rs *RewardService) GrantRewardsToPlayerWithMultiplier(playerID int64, rewards *PvPRewards) error {
	// 应用倍率
	finalRewards := rs.ApplyRewardMultiplier(rewards)
	// 发放奖励
	return rs.GrantRewardsToPlayer(playerID, finalRewards)
}

// GrantPvERewardsToPlayer 发放PvE战斗奖励（灵草）给玩家
func (rs *RewardService) GrantPvERewardsToPlayer(playerID int64, rewards *PvERewards) error {
	if rewards == nil {
		return fmt.Errorf("PvE奖励为空")
	}

	// 查询或创建灵草记录
	var existingHerb models.Herb
	queryErr := db.DB.Where("user_id = ? AND herb_id = ?", playerID, rewards.HerbID).First(&existingHerb).Error

	if queryErr == nil {
		// 灵草已存在，更新数量
		existingHerb.Count += rewards.Count
		if err := db.DB.Model(&existingHerb).Update("count", existingHerb.Count).Error; err != nil {
			log.Printf("[Reward] 更新灵草数量失败: %v", err)
			return fmt.Errorf("更新灵草数量失败: %w", err)
		}
	} else if errors.Is(queryErr, gorm.ErrRecordNotFound) {
		// 灵草不存在，创建新记录
		newHerb := models.Herb{
			UserID:  uint(playerID),
			HerbID:  rewards.HerbID,
			Name:    rewards.Name,
			Count:   rewards.Count,
			Quality: rewards.Quality,
		}
		if err := db.DB.Create(&newHerb).Error; err != nil {
			log.Printf("[Reward] 创建灵草记录失败: %v", err)
			return fmt.Errorf("创建灵草记录失败: %w", err)
		}
	} else {
		log.Printf("[Reward] 查询灵草失败: %v", queryErr)
		return fmt.Errorf("查询灵草失败: %w", queryErr)
	}

	log.Printf("[Reward] 玩家 %d 获得PvE奖励 - 灵草: %s(%s), 数量: %d",
		playerID, rewards.Name, rewards.Quality, rewards.Count)

	return nil
}

// GrantDemonSlayingRewardsToPlayer 发放除魔卫道战斗奖励给玩家
func (rs *RewardService) GrantDemonSlayingRewardsToPlayer(playerID int64, rewards *DemonSlayingRewards) error {
	if rewards == nil {
		return fmt.Errorf("除魔卫道奖励为空")
	}

	var user models.User
	if err := db.DB.First(&user, playerID).Error; err != nil {
		return fmt.Errorf("获取玩家信息失败: %w", err)
	}

	// 更新灵石
	user.SpiritStones += int(rewards.SpiritStones)

	// 更新修为
	user.Cultivation += float64(rewards.Cultivation)
	if user.Cultivation > user.MaxCultivation {
		user.Cultivation = user.MaxCultivation
	}

	log.Printf("[Reward] 准备发放除魔卫道奖励 - 玩家ID: %d, 灵石增加: %d -> %d, 修为增加: %d -> %.2f",
		playerID, rewards.SpiritStones, user.SpiritStones, rewards.Cultivation, user.Cultivation)

	// 使用 Updates map 方式更新，避免零值字段更新失败
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit_stones": user.SpiritStones,
		"cultivation":   user.Cultivation,
	}).Error; err != nil {
		log.Printf("[Reward] 更新玩家资源失败: %v", err)
		return fmt.Errorf("更新玩家资源失败: %w", err)
	}

	log.Printf("[Reward] 成功发放灵石和修为 - 玩家ID: %d", playerID)

	// 如果有丹方残页，增加残页数量
	if rewards.PillFragmentID != "" {
		var fragment models.PillFragment
		queryErr := db.DB.Where("user_id = ? AND recipe_id = ?", playerID, rewards.PillFragmentID).First(&fragment).Error

		if queryErr == nil {
			// 残页已存在，更新数量
			fragment.Count += 1
			if err := db.DB.Model(&fragment).Update("count", fragment.Count).Error; err != nil {
				log.Printf("[Reward] 更新丹方残页数量失败: %v", err)
				return fmt.Errorf("更新丹方残页数量失败: %w", err)
			}
		} else if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			// 残页不存在，创建新记录
			newFragment := models.PillFragment{
				UserID:   uint(playerID),
				RecipeID: rewards.PillFragmentID,
				Count:    1,
			}
			if err := db.DB.Create(&newFragment).Error; err != nil {
				log.Printf("[Reward] 创建丹方残页记录失败: %v", err)
				return fmt.Errorf("创建丹方残页记录失败: %w", err)
			}
		} else {
			log.Printf("[Reward] 查询丹方残页失败: %v", queryErr)
			return fmt.Errorf("查询丹方残页失败: %w", queryErr)
		}

		log.Printf("[Reward] 玩家 %d 获得除魔卫道奖励 - 灵石: %d, 修为: %d, 丹方残页: %s",
			playerID, rewards.SpiritStones, rewards.Cultivation, rewards.PillFragmentName)
	} else {
		log.Printf("[Reward] 玩家 %d 获得除魔卫道奖励 - 灵石: %d, 修为: %d",
			playerID, rewards.SpiritStones, rewards.Cultivation)
	}

	return nil
}
