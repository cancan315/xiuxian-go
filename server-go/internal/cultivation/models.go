package cultivation

// 修炼相关常量
const (
	BaseCultivationCost       = 10  // 基础修炼消耗灵力
	BaseCultivationGain       = 1   // 基础修炼获得修为
	CultivationCostMultiplier = 1.5 // 等级消耗倍率
	CultivationGainMultiplier = 1.2 // 等级获得倍率
	BaseGainRate              = 1   // 基础灵力获取速率
	BreakthroughReward        = 100 // 突破奖励倍数
	BreakthroughBonus         = 1.2 // 突破灵力获取倍率提升
	ExtraCultivationChance    = 0.3 // 额外修为概率基数
)

// 修炼请求
type CultivationRequest struct {
	CultivationType string `json:"cultivationType"` // single, auto, breakthrough
	Duration        int    `json:"duration"`        // 持续时长(ms) - 用于自动修炼
}

// 修炼响应
type CultivationResponse struct {
	Success            bool        `json:"success"`
	CultivationGain    float64     `json:"cultivationGain"`    // 获得的修为
	SpiritCost         float64     `json:"spiritCost"`         // 消耗的灵力
	CurrentCultivation float64     `json:"currentCultivation"` // 当前修为
	Breakthrough       interface{} `json:"breakthrough,omitempty"`
	Message            string      `json:"message,omitempty"`
	Error              string      `json:"error,omitempty"`
}

// 突破响应
type BreakthroughResponse struct {
	Success           bool    `json:"success"`
	NewLevel          int     `json:"newLevel"`
	NewRealm          string  `json:"newRealm"`
	NewMaxCultivation float64 `json:"newMaxCultivation"`
	SpiritReward      float64 `json:"spiritReward"`
	NewSpiritRate     float64 `json:"newSpiritRate"`
	Message           string  `json:"message"`
}

// 自动修炼响应
type AutoCultivationResponse struct {
	Success              bool          `json:"success"`
	TotalCultivationGain float64       `json:"totalCultivationGain"` // 总获得修为
	TotalSpiritCost      float64       `json:"totalSpiritCost"`      // 总消耗灵力
	Breakthroughs        int           `json:"breakthroughs"`        // 突破次数
	FinalCultivation     float64       `json:"finalCultivation"`     // 最终修为
	BreakthroughDetails  []interface{} `json:"breakthroughDetails"`  // 突破详情
	Message              string        `json:"message,omitempty"`
	Error                string        `json:"error,omitempty"`
}

// 修炼数据
type CultivationData struct {
	Level           int      `json:"level"`
	Realm           string   `json:"realm"`
	Cultivation     float64  `json:"cultivation"`
	MaxCultivation  float64  `json:"maxCultivation"`
	Spirit          float64  `json:"spirit"`
	SpiritCost      float64  `json:"spiritCost"`      // 灵力消耗
	CultivationGain float64  `json:"cultivationGain"` // 修炼获得修为
	CultivationRate float64  `json:"cultivationRate"`
	SpiritRate      float64  `json:"spiritRate"`
	UnlockedRealms  []string `json:"unlockedRealms"`
}

// 境界信息
type RealmInfo struct {
	Level          int     `json:"level"`
	Name           string  `json:"name"`
	MaxCultivation float64 `json:"maxCultivation"`
}

// 玩家修炼统计
type PlayerStats struct {
	Level                int      `json:"level"`
	Realm                string   `json:"realm"`
	Cultivation          float64  `json:"cultivation"`
	MaxCultivation       float64  `json:"maxCultivation"`
	Spirit               float64  `json:"spirit"`
	CultivationRate      float64  `json:"cultivationRate"`
	SpiritRate           float64  `json:"spiritRate"`
	UnlockedRealms       []string `json:"unlockedRealms"`
	TotalCultivationTime int      `json:"totalCultivationTime"`
	BreakthroughCount    int      `json:"breakthroughCount"`
}
