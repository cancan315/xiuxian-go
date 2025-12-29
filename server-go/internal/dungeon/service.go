package dungeon

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/dungeon/battle"
	"xiuxian/server-go/internal/dungeon/battle/resolver"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
)

const (
	RequireSpirit       = 10000
	RequireSpiritStones = 1000
)

// ========== 秘境灵力消耗计算函数 ==========

// calculateDungeonSpritCost 计算秘境灵力消耗
// 公式: dungeonCost = 2880 * 1.2^(Level-1)
func calculateDungeonSpiritCost(level int) float64 {
	const baseCost = 2880.0
	const costMultiplier = 1.2
	return baseCost * math.Pow(costMultiplier, float64(level-1))
}

// checkDungeonEntryCost 检查进入秘境的消耗灵力
// 调整后：根据等级动态计算灵力消耗（原石消耗暂时保持不变）
func checkDungeonEntryCost(user *models.User, floor int) error {
	// 根据等级计算灵力消耗
	spiritCost := calculateDungeonSpiritCost(user.Level)
	// 灵石消耗保持原逻辑（基于层数）
	stoneCost := int(float64(floor) * 10)

	if user.Spirit < spiritCost || user.SpiritStones < stoneCost {
		requiredSpirit := spiritCost - user.Spirit
		if requiredSpirit < 0 {
			requiredSpirit = 0
		}
		return fmt.Errorf("进入秘境需要灵力%.0f、灵石%d，灵力还差%.0f", spiritCost, stoneCost, requiredSpirit)
	}
	return nil
}

func applyDifficultyToEnemy(stats *CombatStats, modifier DifficultyModifier) {
	stats.Health *= modifier.HealthMod
	stats.MaxHealth *= modifier.HealthMod
	stats.Damage *= modifier.DamageMod
	stats.Defense *= modifier.DefenseMod
}

// DungeonService 秘境服务
type DungeonService struct {
	userID        uint
	selectedBuffs []string               // 已选增益ID
	buffEffects   map[string]interface{} // 已选增益的效果应用
	rewardAmount  int                    // 战斗奖励金额
}

// NewDungeonService 创建秘境服务实例
func NewDungeonService(userID uint) *DungeonService {
	return &DungeonService{
		userID:        userID,
		selectedBuffs: []string{},
		buffEffects:   make(map[string]interface{}),
	}
}

// toCombatStats 将 dungeon.CombatStats 转换为 battle.CombatStats
func toCombatStats(stats *CombatStats) *battle.CombatStats {
	return &battle.CombatStats{
		Health:            stats.Health,
		MaxHealth:         stats.MaxHealth,
		Damage:            stats.Damage,
		Defense:           stats.Defense,
		Speed:             stats.Speed,
		CritRate:          stats.CritRate,
		ComboRate:         stats.ComboRate,
		CounterRate:       stats.CounterRate,
		StunRate:          stats.StunRate,
		DodgeRate:         stats.DodgeRate,
		VampireRate:       stats.VampireRate,
		CritResist:        stats.CritResist,
		ComboResist:       stats.ComboResist,
		CounterResist:     stats.CounterResist,
		StunResist:        stats.StunResist,
		DodgeResist:       stats.DodgeResist,
		VampireResist:     stats.VampireResist,
		HealBoost:         stats.HealBoost,
		CritDamageBoost:   stats.CritDamageBoost,
		CritDamageReduce:  stats.CritDamageReduce,
		FinalDamageBoost:  stats.FinalDamageBoost,
		FinalDamageReduce: stats.FinalDamageReduce,
		CombatBoost:       stats.CombatBoost,
		ResistanceBoost:   stats.ResistanceBoost,
	}
}

// fromCombatStats 将 battle.CombatStats 转换为 dungeon.CombatStats
func fromCombatStats(stats *battle.CombatStats) *CombatStats {
	return &CombatStats{
		Health:            stats.Health,
		MaxHealth:         stats.MaxHealth,
		Damage:            stats.Damage,
		Defense:           stats.Defense,
		Speed:             stats.Speed,
		CritRate:          stats.CritRate,
		ComboRate:         stats.ComboRate,
		CounterRate:       stats.CounterRate,
		StunRate:          stats.StunRate,
		DodgeRate:         stats.DodgeRate,
		VampireRate:       stats.VampireRate,
		CritResist:        stats.CritResist,
		ComboResist:       stats.ComboResist,
		CounterResist:     stats.CounterResist,
		StunResist:        stats.StunResist,
		DodgeResist:       stats.DodgeResist,
		VampireResist:     stats.VampireResist,
		HealBoost:         stats.HealBoost,
		CritDamageBoost:   stats.CritDamageBoost,
		CritDamageReduce:  stats.CritDamageReduce,
		FinalDamageBoost:  stats.FinalDamageBoost,
		FinalDamageReduce: stats.FinalDamageReduce,
		CombatBoost:       stats.CombatBoost,
		ResistanceBoost:   stats.ResistanceBoost,
	}
}

// GetRandomBuffs 获取随机增益选项
func (s *DungeonService) GetRandomBuffs(floor int) ([]BuffOption, error) {
	// 计算概率 (common 通过 else 分支隐式处理)
	rareChance := 0.25
	epicChance := 0.05

	if floor%10 == 0 {
		rareChance = 0.3
		epicChance = 0.2
	} else if floor%5 == 0 {
		rareChance = 0.35
		epicChance = 0.15
	}

	var options []BuffOption
	usedIds := make(map[string]bool)

	for len(options) < 3 {
		randVal := rand.Float64()
		var pool string

		if randVal < epicChance {
			pool = "epic"
		} else if randVal < epicChance+rareChance {
			pool = "rare"
		} else {
			pool = "common"
		}

		// 从池中随机选择
		configs := BuffConfigPool[pool]
		var available []BuffConfig

		for _, cfg := range configs {
			if !usedIds[cfg.ID] {
				available = append(available, cfg)
			}
		}

		if len(available) > 0 {
			idx := rand.Intn(len(available))
			cfg := available[idx]
			usedIds[cfg.ID] = true

			option := BuffOption{
				ID:          cfg.ID,
				Name:        cfg.Name,
				Description: cfg.Description,
				Type:        cfg.Type,
				Effect:      cfg.Effects,
			}
			options = append(options, option)
		} else {
			// 从所有池中选择
			var allAvailable []BuffConfig
			for _, poolConfigs := range BuffConfigPool {
				for _, cfg := range poolConfigs {
					if !usedIds[cfg.ID] {
						allAvailable = append(allAvailable, cfg)
					}
				}
			}

			if len(allAvailable) > 0 {
				idx := rand.Intn(len(allAvailable))
				cfg := allAvailable[idx]
				usedIds[cfg.ID] = true

				option := BuffOption{
					ID:          cfg.ID,
					Name:        cfg.Name,
					Description: cfg.Description,
					Type:        cfg.Type,
					Effect:      cfg.Effects,
				}
				options = append(options, option)
			} else {
				break
			}
		}
	}

	return options, nil
}

// parsePlayerAttributes 从数据库用户数据中解析玩家属性
func (s *DungeonService) parsePlayerAttributes(user *models.User) (*CombatStats, error) {
	// 解析基础属性JSON
	var baseAttrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &baseAttrs); err != nil {
		return nil, fmt.Errorf("解析基础属性失败: %w", err)
	}

	// 解析战斗属性JSON
	var combatAttrs map[string]interface{}
	if err := json.Unmarshal(user.CombatAttributes, &combatAttrs); err != nil {
		return nil, fmt.Errorf("解析战斗属性失败: %w", err)
	}

	// 解析战斗抗性JSON
	var combatResist map[string]interface{}
	if err := json.Unmarshal(user.CombatResistance, &combatResist); err != nil {
		return nil, fmt.Errorf("解析战斗抗性失败: %w", err)
	}

	// 解析特殊属性JSON
	var specialAttrs map[string]interface{}
	if err := json.Unmarshal(user.SpecialAttributes, &specialAttrs); err != nil {
		return nil, fmt.Errorf("解析特殊属性失败: %w", err)
	}

	// 从数据库中直接读取属性值
	var health, attack, defense, speed float64
	var critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate float64
	var critResist, comboResist, counterResist, stunResist, dodgeResist, vampireResist float64
	var healBoost, critDamageBoost, critDamageReduce, finalDamageBoost, finalDamageReduce, combatBoost, resistanceBoost float64

	// 从 baseAttributes 读取基础属性
	if v, ok := baseAttrs["health"].(float64); ok {
		health = v
	}
	if v, ok := baseAttrs["attack"].(float64); ok {
		attack = v
	}
	if v, ok := baseAttrs["defense"].(float64); ok {
		defense = v
	}
	if v, ok := baseAttrs["speed"].(float64); ok {
		speed = v
	}

	// 从 combatAttributes 读取战斗属性
	if v, ok := combatAttrs["critRate"].(float64); ok {
		critRate = v
	}
	if v, ok := combatAttrs["comboRate"].(float64); ok {
		comboRate = v
	}
	if v, ok := combatAttrs["counterRate"].(float64); ok {
		counterRate = v
	}
	if v, ok := combatAttrs["stunRate"].(float64); ok {
		stunRate = v
	}
	if v, ok := combatAttrs["dodgeRate"].(float64); ok {
		dodgeRate = v
	}
	if v, ok := combatAttrs["vampireRate"].(float64); ok {
		vampireRate = v
	}

	// 从 combatResistance 读取战斗抗性
	if v, ok := combatResist["critResist"].(float64); ok {
		critResist = v
	}
	if v, ok := combatResist["comboResist"].(float64); ok {
		comboResist = v
	}
	if v, ok := combatResist["counterResist"].(float64); ok {
		counterResist = v
	}
	if v, ok := combatResist["stunResist"].(float64); ok {
		stunResist = v
	}
	if v, ok := combatResist["dodgeResist"].(float64); ok {
		dodgeResist = v
	}
	if v, ok := combatResist["vampireResist"].(float64); ok {
		vampireResist = v
	}

	// 从 specialAttributes 读取特殊属性
	if v, ok := specialAttrs["healBoost"].(float64); ok {
		healBoost = v
	}
	if v, ok := specialAttrs["critDamageBoost"].(float64); ok {
		critDamageBoost = v
	}
	if v, ok := specialAttrs["critDamageReduce"].(float64); ok {
		critDamageReduce = v
	}
	if v, ok := specialAttrs["finalDamageBoost"].(float64); ok {
		finalDamageBoost = v
	}
	if v, ok := specialAttrs["finalDamageReduce"].(float64); ok {
		finalDamageReduce = v
	}
	if v, ok := specialAttrs["combatBoost"].(float64); ok {
		combatBoost = v
	}
	if v, ok := specialAttrs["resistanceBoost"].(float64); ok {
		resistanceBoost = v
	}

	// 初始化玩家战斗属性
	return &CombatStats{
		Health:            health,
		MaxHealth:         health,
		Damage:            attack,
		Defense:           defense,
		Speed:             speed,
		CritRate:          critRate,
		ComboRate:         comboRate,
		CounterRate:       counterRate,
		StunRate:          stunRate,
		DodgeRate:         dodgeRate,
		VampireRate:       vampireRate,
		CritResist:        critResist,
		ComboResist:       comboResist,
		CounterResist:     counterResist,
		StunResist:        stunResist,
		DodgeResist:       dodgeResist,
		VampireResist:     vampireResist,
		HealBoost:         healBoost,
		CritDamageBoost:   critDamageBoost,
		CritDamageReduce:  critDamageReduce,
		FinalDamageBoost:  finalDamageBoost,
		FinalDamageReduce: finalDamageReduce,
		CombatBoost:       combatBoost,
		ResistanceBoost:   resistanceBoost,
	}, nil
}

// initEnemyStats 根据玩家属性和层数初始化敌人属性
// initEnemyStats 生成指定楼层敌人的战斗属性
// 该函数基于地牢层数(floor)和难度修饰符(modifier)动态计算敌人的所有战斗属性
// 返回指向CombatStats结构体的指针，包含完整的敌人属性集
func (s *DungeonService) initEnemyStats(floor int) *CombatStats {
	// 基础属性基准值 - 这些值定义了1层敌人(无层数加成)的基本强度
	// 这些数值可根据游戏平衡性进行全局调整:
	// - baseHealth: 基础生命值，决定了敌人的耐久度
	// - baseDamage: 基础攻击力，决定了敌人的输出能力
	// - baseDefense: 基础防御力，结合难度系数影响伤害减免
	// - baseSpeed: 基础速度，影响行动顺序和频率
	baseHealth := 100.0 // 1层敌人基础生命值
	baseDamage := 5.0   // 1层敌人基础伤害值
	baseDefense := 3.0  // 1层敌人基础防御值
	baseSpeed := 1.0    // 1层敌人基础速度值(通常作为相对值基准)

	return &CombatStats{
		// 核心战斗属性 - 随层数线性增长
		// 生命值计算: 基础值 * (1 + 层数*10%)
		// 示例: 5层敌人生命值 = 100 * (1 + 5*0.1) = 150
		Health:    10 * float64(floor) * baseHealth * (1 + float64(floor)*0.1), // 10倍层数乘以基础生命值，加上10%的层数加成
		MaxHealth: 10 * float64(floor) * baseHealth * (1 + float64(floor)*0.1), // 与当前生命值相同，表示满血状态

		// 伤害计算: 基础值 * (1 + 层数*8%)
		// 增长率略低于生命值，保持战斗回合数相对稳定
		Damage: 10 * float64(floor) * baseDamage * (1 + float64(floor)*0.08), // 10倍层数乘以基础伤害值，加上8%的层数加成

		// 防御计算: 基础值 * (1 + 层数*7%)
		Defense: 10 * float64(floor) * baseDefense * (1 + float64(floor)*0.07), // 10倍层数乘以基础防御值，加上7%的层数加成

		// 速度计算: 基础值 * (1 + 层数*3%)
		// 速度增长率较低，避免高层数敌人行动过于频繁
		Speed: 10 * float64(floor) * baseSpeed * (1 + float64(floor)*0.03), // 10倍层数乘以基础速度值，加上3%的层数加成

		// 战斗概率属性 - 从基础值开始，每层增加固定百分比
		// 注意: 这些属性理论上可能超过1.0(100%)，实际使用时应在战斗系统中限制最大值
		CritRate:    0.05 + float64(floor)*0.02, // 基础5%暴击率，每层+2%
		ComboRate:   0.03 + float64(floor)*0.02, // 基础3%连击率，每层+2%
		CounterRate: 0.03 + float64(floor)*0.02, // 基础3%反击率，每层+2%
		StunRate:    0.02 + float64(floor)*0.01, // 基础2%眩晕率，每层+1%
		DodgeRate:   0.05 + float64(floor)*0.02, // 基础5%闪避率，每层+2%
		VampireRate: 0.02 + float64(floor)*0.01, // 基础2%吸血率，每层+1%

		// 抵抗类属性 - 减少受到负面效果的概率
		CritResist:    0.02 + float64(floor)*0.01, // 基础2%暴击抵抗，每层+1%
		ComboResist:   0.02 + float64(floor)*0.01, // 基础2%连击抵抗，每层+1%
		CounterResist: 0.02 + float64(floor)*0.01, // 基础2%反击抵抗，每层+1%
		StunResist:    0.02 + float64(floor)*0.01, // 基础2%眩晕抵抗，每层+1%
		DodgeResist:   0.02 + float64(floor)*0.01, // 基础2%闪避抵抗，每层+1%
		VampireResist: 0.02 + float64(floor)*0.01, // 基础2%吸血抵抗，每层+1%

		// 伤害增强/减免属性 - 影响最终伤害计算
		HealBoost:         0.05 + float64(floor)*0.02, // 基础5%治疗增强，每层+2%
		CritDamageBoost:   0.2 + float64(floor)*0.1,   // 基础20%暴伤增强，每层+10%(高增长确保暴击威胁)
		CritDamageReduce:  0.1 + float64(floor)*0.05,  // 基础10%暴伤减免，每层+5%
		FinalDamageBoost:  0.05 + float64(floor)*0.02, // 基础5%最终伤害增强，每层+2%
		FinalDamageReduce: 0.05 + float64(floor)*0.02, // 基础5%最终伤害减免，每层+2%

		// 综合战斗属性
		CombatBoost:     0.03 + float64(floor)*0.02, // 基础3%战斗增强(综合进攻能力)，每层+2%
		ResistanceBoost: 0.03 + float64(floor)*0.02, // 基础3%抵抗增强(综合防御能力)，每层+2%
	}
}

// RoundResult 单回合战斗结果
type RoundResult struct {
	Round        int      `json:"round"`
	PlayerHealth float64  `json:"playerHealth"`
	EnemyHealth  float64  `json:"enemyHealth"`
	Log          []string `json:"log"`
	BattleEnded  bool     `json:"battleEnded"`
	Victory      bool     `json:"victory"`
}

// StartFight 开始战斗 - 初始化战斗状态到Redis，准备第一回合
func (s *DungeonService) StartFight(floor int, difficulty string) (*FightResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	// 检查进入秘境的消耗灵力10000，灵石1000
	if err := checkDungeonEntryCost(&user, floor); err != nil {
		return nil, err
	}
	// 立即扣除消耗 (动态计算)
	// spiritCost := int(float64(floor) * 1000) // 灵力消耗
	// stoneCost := int(float64(floor) * 10)    // 灵石消耗
	spiritCostFloat := calculateDungeonSpiritCost(user.Level)
	spiritCost := int(spiritCostFloat)
	// 灵石消耗基于层数
	stoneCost := int(float64(floor) * 10)
	user.Spirit = math.Max(0, user.Spirit-spiritCostFloat)
	user.SpiritStones -= stoneCost
	if user.SpiritStones < 0 {
		user.SpiritStones = 0
	}
	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":        user.Spirit,
		"spirit_stones": user.SpiritStones,
	}).Error; err != nil {
		return nil, fmt.Errorf("扣除消耗失败: %w", err)
	}
	modifier := GetDifficultyModifier(difficulty)

	// 解析玩家属性
	playerStats, err := s.parsePlayerAttributes(&user)
	if err != nil {
		return nil, err
	}

	// 应用已选择的增益效果
	playerStats = fromCombatStats(resolver.ApplyBuffEffectsToStats(toCombatStats(playerStats), s.buffEffects))

	// 应用难度修饰符,玩家不需要
	// playerStats.Health *= modifier.HealthMod
	// playerStats.MaxHealth *= modifier.HealthMod
	// playerStats.Damage *= modifier.DamageMod
	// 初始化敌人属性
	enemyStats := s.initEnemyStats(floor)
	applyDifficultyToEnemy(enemyStats, modifier)

	// 创建战斗状态并保存到Redis
	battleStatus := &BattleStatus{
		UserID:          s.userID,
		Floor:           floor,
		Difficulty:      difficulty,
		Round:           0,
		PlayerHealth:    playerStats.MaxHealth,
		PlayerMaxHealth: playerStats.MaxHealth,
		EnemyHealth:     enemyStats.MaxHealth,
		EnemyMaxHealth:  enemyStats.MaxHealth,
		PlayerStats:     playerStats,
		EnemyStats:      enemyStats,
		BuffEffects:     s.buffEffects,
		BattleLog:       []string{},
	}

	if err := s.SaveBattleStatusToRedis(battleStatus); err != nil {
		// Redis保存失败不影响战斗流程，仅记录日志
		fmt.Printf("保存战斗状态到Redis失败: %v\n", err)
	}

	// 返回战斗初始化成功，包含完整的玩家和敌人属性
	result := &FightResult{
		Success:         true,
		Victory:         false,
		Floor:           floor,
		Message:         fmt.Sprintf("战斗已初始化，消耗%.0f灵力、%d灵石，请调用执行回合接口开始战斗", spiritCostFloat, stoneCost),
		SpiritCost:      spiritCost,
		StoneCost:       stoneCost,
		PlayerStats:     playerStats,
		EnemyStats:      enemyStats,
		PlayerHealth:    playerStats.MaxHealth,
		EnemyHealth:     enemyStats.MaxHealth,
		PlayerMaxHealth: playerStats.MaxHealth,
		EnemyMaxHealth:  enemyStats.MaxHealth,
		MaxRounds:       100, // 最大100回合
	}

	return result, nil
}

// SelectBuffAndApplyEffects 选择增益并记录效果
func (s *DungeonService) SelectBuffAndApplyEffects(buffID string) (map[string]interface{}, error) {
	buff, err := s.SelectBuff(buffID)
	if err != nil {
		return nil, err
	}

	// 记录已选增益
	s.selectedBuffs = append(s.selectedBuffs, buffID)

	// 提取效果并应用
	if effects, ok := buff["effects"].(map[string]interface{}); ok {
		for k, v := range effects {
			s.buffEffects[k] = v
		}
	}

	return buff, nil
}

// SelectBuff 选择增益
func (s *DungeonService) SelectBuff(buffID string) (map[string]interface{}, error) {
	// 查找增益配置
	var selectedConfig *BuffConfig

	for _, poolConfigs := range BuffConfigPool {
		for _, cfg := range poolConfigs {
			if cfg.ID == buffID {
				selectedConfig = &cfg
				break
			}
		}
		if selectedConfig != nil {
			break
		}
	}

	if selectedConfig == nil {
		return nil, fmt.Errorf("增益不存在: %s", buffID)
	}

	return map[string]interface{}{
		"id":          selectedConfig.ID,
		"name":        selectedConfig.Name,
		"description": selectedConfig.Description,
		"effects":     selectedConfig.Effects,
	}, nil
}

// EndDungeon 结束秘境
// 次季：引入了rewardAmount字段，计算宁亭是她StartFight中计算的，这里翰撩直接返回
func (s *DungeonService) EndDungeon(floor int, victory bool) (map[string]interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 简接使用StartFight中计算的rewardAmount
	// 次程：EndDungeon仅达辂引入深深新邮存并欧说
	if victory && s.rewardAmount > 0 {
		user.SpiritStones += s.rewardAmount
	}

	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, fmt.Errorf("更新用户数据失败: %w", err)
	}

	// 清除Redis中的战斗状态和回合数据
	if err := s.ClearBattleStatusFromRedis(); err != nil {
		fmt.Printf("清除战斗状态失败: %v\n", err)
	}
	if err := s.ClearRoundDataFromRedis(); err != nil {
		fmt.Printf("清除回合数据失败: %v\n", err)
	}

	return map[string]interface{}{
		"floor":        floor,
		"totalReward":  s.rewardAmount,
		"victory":      victory,
		"spiritStones": user.SpiritStones,
	}, nil
}

// SaveSessionToRedis 将秘境会话保存到Redis
func (s *DungeonService) SaveSessionToRedis(floor int, difficulty string) error {
	session := &DungeonSession{
		UserID:        s.userID,
		Floor:         floor,
		Difficulty:    difficulty,
		RefreshCount:  3,
		SelectedBuffs: s.selectedBuffs,
		PlayerBuffs:   s.buffEffects,
	}

	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// 清除成秘境会话，有效期60分钟
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// LoadSessionFromRedis 从 Redis 加载秘境会话
func (s *DungeonService) LoadSessionFromRedis() (*DungeonSession, error) {
	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err // 会话不存在
	}

	var session DungeonSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	// 恢复中断不丢失的增益效果
	s.selectedBuffs = session.SelectedBuffs
	s.buffEffects = session.PlayerBuffs

	return &session, nil
}

// ClearSessionFromRedis 清除Redis中的秘境会话
func (s *DungeonService) ClearSessionFromRedis() error {
	key := fmt.Sprintf("dungeon:session:%d", s.userID)
	return redis.Client.Del(redis.Ctx, key).Err()
}

// saveBattleStatsToRedis 将玩家和敌人属性存入Redis
func (s *DungeonService) saveBattleStatsToRedis(playerStats, enemyStats *CombatStats, floor int) error {
	// 玩家属性
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	playerData, err := json.Marshal(playerStats)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, playerKey, string(playerData), 60*time.Minute).Err(); err != nil {
		return err
	}

	// 敌人属性
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	enemyData, err := json.Marshal(enemyStats)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, enemyKey, string(enemyData), 60*time.Minute).Err(); err != nil {
		return err
	}

	return nil
}

// LoadBattleStatsFromRedis 从Redis加载玩家和敌人属性
func (s *DungeonService) LoadBattleStatsFromRedis() (*CombatStats, *CombatStats, error) {
	// 加载玩家属性
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	playerData, err := redis.Client.Get(redis.Ctx, playerKey).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("加载玩家属性失败: %w", err)
	}
	var playerStats CombatStats
	if err := json.Unmarshal([]byte(playerData), &playerStats); err != nil {
		return nil, nil, err
	}

	// 加载敌人属性
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	enemyData, err := redis.Client.Get(redis.Ctx, enemyKey).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("加载敌人属性失败: %w", err)
	}
	var enemyStats CombatStats
	if err := json.Unmarshal([]byte(enemyData), &enemyStats); err != nil {
		return nil, nil, err
	}

	return &playerStats, &enemyStats, nil
}

// ClearBattleStatsFromRedis 清除Redis中的战斗属性
func (s *DungeonService) ClearBattleStatsFromRedis() error {
	playerKey := fmt.Sprintf("dungeon:battle:%d:player", s.userID)
	enemyKey := fmt.Sprintf("dungeon:battle:%d:enemy", s.userID)
	return redis.Client.Del(redis.Ctx, playerKey, enemyKey).Err()
}

// SaveBattleStatusToRedis 将战斗状态保存到Redis
func (s *DungeonService) SaveBattleStatusToRedis(status *BattleStatus) error {
	key := fmt.Sprintf("dungeon:battle:status:%d", s.userID)
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

// LoadBattleStatusFromRedis 从Redis加载战斗状态
func (s *DungeonService) LoadBattleStatusFromRedis() (*BattleStatus, error) {
	key := fmt.Sprintf("dungeon:battle:status:%d", s.userID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var status BattleStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// SaveRoundDataToRedis 将单回合数据保存到Redis
func (s *DungeonService) SaveRoundDataToRedis(roundData *RoundData) error {
	key := fmt.Sprintf("dungeon:battle:round:%d", s.userID)
	data, err := json.Marshal(roundData)
	if err != nil {
		return err
	}
	if err := redis.Client.Set(redis.Ctx, key, string(data), 60*time.Minute).Err(); err != nil {
		return err
	}
	return nil
}

// GetRoundDataFromRedis 从Redis获取单回合数据
func (s *DungeonService) GetRoundDataFromRedis() (*RoundData, error) {
	key := fmt.Sprintf("dungeon:battle:round:%d", s.userID)
	data, err := redis.Client.Get(redis.Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var roundData RoundData
	if err := json.Unmarshal([]byte(data), &roundData); err != nil {
		return nil, err
	}

	return &roundData, nil
}

// ClearBattleStatusFromRedis 清除Redis中的战斗状态
func (s *DungeonService) ClearBattleStatusFromRedis() error {
	key := fmt.Sprintf("dungeon:battle:status:%d", s.userID)
	return redis.Client.Del(redis.Ctx, key).Err()
}

// ClearRoundDataFromRedis 清除Redis中的回合数据
func (s *DungeonService) ClearRoundDataFromRedis() error {
	key := fmt.Sprintf("dungeon:battle:round:%d", s.userID)
	return redis.Client.Del(redis.Ctx, key).Err()
}
