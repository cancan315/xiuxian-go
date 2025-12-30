package duel

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// PvPRewards PvP战斗奖励结果
type PvPRewards struct {
	SpiritStones int64 `json:"spirit_stones"`
	Cultivation  int64 `json:"cultivation"`
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

// CalculateRewardsForPvE 计算PvE战斗奖励
// 需要传入玩家等级以计算基于等级的灵石和修为奖励
func (rs *RewardService) CalculateRewardsForPvE(status *PvEBattleStatus, playerLevel int) *PvPRewards {
	spiritStones := rs.calculateSpiritStoneReward(status.Round, playerLevel)
	cultivation := rs.calculateCultivationReward(status.Round, playerLevel)

	return &PvPRewards{
		SpiritStones: spiritStones,
		Cultivation:  cultivation,
	}
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

	// 使用公式计算基础奖励：504 * 0.96^(level-1)
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
