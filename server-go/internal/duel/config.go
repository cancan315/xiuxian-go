package duel

// RewardConfig 斗法奖励配置
// 用于定义PvP战斗胜利后玩家获得的各种奖励参数
type RewardConfig struct {
	// SpiritStones 灵石奖励配置
	// 控制玩家在斗法获胜后获得的基础灵石数量及奖励变化规则
	SpiritStones SpiritStoneReward `json:"spirit_stones"`
	// Cultivation 修为奖励配置
	// 控制玩家在斗法获胜后获得的基础修为数量及奖励变化规则
	Cultivation CultivationReward `json:"cultivation"`
	// Multiplier 奖励倍率配置
	// 定义奖励倍率的概率分布，提供随机奖励倍率机制
	Multiplier MultiplierConfig `json:"multiplier"`
	// MinLevelRequirement 参与斗法的最低等级要求
	// 玩家等级必须大于此值才能参与斗法
	MinLevelRequirement int `json:"min_level_requirement"`
}

// SpiritStoneReward 灵石奖励配置
// 定义斗法胜利后获得灵石奖励的具体参数
type SpiritStoneReward struct {
	// BaseFormula 基础奖励公式参数
	// 用于计算基于等级的奖励：BaseFormula.Base * (BaseFormula.Multiplier ^ (level-1))
	BaseFormula SpiritStoneFormula `json:"base_formula"`
	// RoundBonus 回合奖励系数
	// 影响根据战斗回合数调整奖励的系数，回合越少奖励越多
	RoundBonus float64 `json:"round_bonus"` // 回合奖励系数（每20回合减少的奖励）
	// MaxRoundBonus 回合奖励上限
	// 限制因回合数带来的额外奖励最大值，防止奖励过高
	MaxRoundBonus int64 `json:"max_round_bonus"` // 回合奖励上限
}

// SpiritStoneFormula 灵石奖励公式参数
type SpiritStoneFormula struct {
	// Base 基础值
	Base float64 `json:"base"`
	// Multiplier 乘数
	Multiplier float64 `json:"multiplier"`
}

// CultivationReward 修为奖励配置
// 定义斗法胜利后获得修为奖励的具体参数
type CultivationReward struct {
	// BaseFormula 基础奖励公式参数
	// 用于计算基于等级的奖励：BaseFormula.Base * (BaseFormula.Multiplier ^ (level-1))
	BaseFormula CultivationFormula `json:"base_formula"`
	// RoundBonus 回合奖励系数
	// 影响根据战斗回合数调整修为奖励的系数，回合越少奖励越多
	RoundBonus float64 `json:"round_bonus"` // 回合奖励系数
	// MaxRoundBonus 回合奖励上限
	// 限制因回合数带来的修为额外奖励最大值，防止奖励过高
	MaxRoundBonus int64 `json:"max_round_bonus"` // 回合奖励上限
}

// CultivationFormula 修为奖励公式参数
type CultivationFormula struct {
	// Base 基础值
	Base float64 `json:"base"`
	// Multiplier 乘数
	Multiplier float64 `json:"multiplier"`
}

// MultiplierConfig 奖励倍率配置
// 定义奖励倍率的概率分布，为每次战斗奖励增加随机性
type MultiplierConfig struct {
	// Probabilities 概率倍率映射
	// 键为累积概率阈值，值为对应的奖励倍率
	// 例如：0.01: 3.0 表示有1%的概率获得3倍奖励
	// 0.06: 2.5 表示接下来5%的概率(总计6%)获得2.5倍奖励
	// 依此类推...
	Probabilities map[float64]float64 `json:"probabilities"`
}

// DefaultRewardConfig 返回默认的奖励配置
// 提供游戏平衡性经过测试的初始奖励参数设置
// 可作为配置文件加载失败时的备用配置
func DefaultRewardConfig() *RewardConfig {
	return &RewardConfig{
		SpiritStones: SpiritStoneReward{
			// 使用公式 50 * 1.05^(level-1) 计算基础灵石奖励（随玩家等级增长）
			BaseFormula: SpiritStoneFormula{
				Base:       50,   // 基础值
				Multiplier: 1.05, // 乘数
			},
			// 回合奖励系数为5.0，意味着回合越短奖励越高
			RoundBonus: 5.0,
			// 最大回合奖励为5，限制额外奖励不超过此值
			MaxRoundBonus: 5,
		},
		Cultivation: CultivationReward{
			// 使用公式 50 * 1.05^(level-1) 计算基础修为奖励（随玩家等级增长）
			BaseFormula: CultivationFormula{
				Base:       50,   // 基础值
				Multiplier: 1.05, // 乘数
			},
			// 回合奖励系数为10.0，鼓励快速结束战斗
			RoundBonus: 10.0,
			// 最大回合奖励为10，控制修为增长速度
			MaxRoundBonus: 10,
		},
		Multiplier: MultiplierConfig{
			// 概率分布设计：
			// 1% 概率获得 3.0 倍奖励 (0.01)
			// 5% 概率获得 2.5 倍奖励 (0.06 - 0.01)
			// 10% 概率获得 2.0 倍奖励 (0.16 - 0.06)
			// 15% 概率获得 1.5 倍奖励 (0.31 - 0.16)
			// 69% 概率获得 1.0 倍奖励 (1.00 - 0.31)
			Probabilities: map[float64]float64{
				0.01: 3.0, // 1% 概率获得 3 倍奖励
				0.06: 2.5, // 5% 概率获得 2.5 倍奖励
				0.16: 2.0, // 10% 概率获得 2 倍奖励
				0.31: 1.5, // 15% 概率获得 1.5 倍奖励
				1.00: 1.0, // 其余 69% 概率获得 1 倍奖励
			},
		},
		// 玩家等级必须大于6才可以参与斗法
		MinLevelRequirement: 6,
	}
}
