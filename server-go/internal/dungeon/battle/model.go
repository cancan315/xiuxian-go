package battle

// CombatStats 战斗属性集
type CombatStats struct {
	// 基础属性
	Health    float64 // 当前生命值
	MaxHealth float64 // 最大生命值
	Damage    float64 // 伤害值
	Defense   float64 // 防御力
	Speed     float64 // 速度

	// 战斗属性（百分比）
	CritRate    float64 // 暴击率
	ComboRate   float64 // 连击率
	CounterRate float64 // 反击率
	StunRate    float64 // 眩晕率
	DodgeRate   float64 // 闪避率
	VampireRate float64 // 吸血率

	// 战斗抗性（百分比）
	CritResist    float64 // 抗暴击
	ComboResist   float64 // 抗连击
	CounterResist float64 // 抗反击
	StunResist    float64 // 抗眩晕
	DodgeResist   float64 // 抗闪避
	VampireResist float64 // 抗吸血

	// 特殊属性（百分比）
	HealBoost         float64 // 强化治疗
	CritDamageBoost   float64 // 强化爆伤
	CritDamageReduce  float64 // 弱化爆伤
	FinalDamageBoost  float64 // 最终增伤
	FinalDamageReduce float64 // 最终减伤
	CombatBoost       float64 // 战斗属性提升
	ResistanceBoost   float64 // 战斗抗性提升
}

// DamageResult 伤害计算结果
type DamageResult struct {
	BaseDamage  float64 // 基础伤害 = A.Damage - B.Defense
	CritDamage  float64 // 暴击伤害
	ComboDamage float64 // 连击伤害
	TotalDamage float64 // 总伤害
	IsCrit      bool    // 是否暴击
	IsCombo     bool    // 是否连击
	IsVampire   bool    // 是否吸血
	IsStun      bool    // 是否眩晕
	VampireHeal float64 // 吸血回复量
}

// TakeDamageResult 被伤害结果
type TakeDamageResult struct {
	Dodged        bool    // 是否闪避
	Damage        float64 // 实际伤害
	CurrentHealth float64 // 当前生命值
	IsDead        bool    // 是否死亡
	IsCounter     bool    // 是否反击
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

// FightResult 战斗结果
type FightResult struct {
	Success bool          `json:"success"`
	Victory bool          `json:"victory"`
	Floor   int           `json:"floor"`
	Message string        `json:"message"`
	Rewards []interface{} `json:"rewards"`
}
