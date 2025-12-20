package dungeon

// 秘境请求
type DungeonRequest struct {
	Floor          int    `json:"floor,omitempty"`          // 当前层数
	Difficulty     string `json:"difficulty"`               // 难度: easy, normal, hard, expert
	Action         string `json:"action,omitempty"`         // 操作: start, get_options, select_buff, fight, end
	SelectedBuffID string `json:"selectedBuffId,omitempty"` // 选择的增益ID
}

// 秘境响应
type DungeonResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// 秘境数据
type DungeonData struct {
	Floor        int    `json:"floor"`        // 当前层数
	Difficulty   string `json:"difficulty"`   // 难度
	RefreshCount int    `json:"refreshCount"` // 刷新次数
}

// 增益选项
type BuffOption struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`   // common, rare, epic
	Effect      map[string]interface{} `json:"effect"` // 属性修改
}

// 增益选项列表响应
type BuffOptionsResponse struct {
	Floor   int          `json:"floor"`
	Options []BuffOption `json:"options"`
}

// 战斗请求
type FightRequest struct {
	Floor int `json:"floor"`
}

// 战斗响应
type FightResponse struct {
	Success      bool    `json:"success"`
	Round        int     `json:"round"`
	MaxRounds    int     `json:"maxRounds"`
	PlayerHealth float64 `json:"playerHealth"`
	EnemyHealth  float64 `json:"enemyHealth"`
	Log          string  `json:"log"`
}

// 战斗结果
type FightResult struct {
	Success bool          `json:"success"`
	Victory bool          `json:"victory"`
	Floor   int           `json:"floor"`
	Message string        `json:"message"`
	Rewards []interface{} `json:"rewards"`
}

// 难度修饰符
type DifficultyModifier struct {
	HealthMod float64
	DamageMod float64
	RewardMod float64
}

// 增益配置
type BuffConfig struct {
	ID          string
	Name        string
	Description string
	Type        string
	Effects     map[string]interface{}
}

// 秘境会话 (用于跟踪玩家当前秘境状态)
type DungeonSession struct {
	UserID        uint
	Floor         int
	Difficulty    string
	RefreshCount  int
	SelectedBuffs []string               // 已选增益ID
	PlayerBuffs   map[string]interface{} // 玩家已获得的增益效果
}

// 战斗属性 (与前端CombatStats对应)
type CombatStats struct {
	Health            float64 // 当前生命值
	MaxHealth         float64 // 最大生命值
	Damage            float64 // 攻击力
	Defense           float64 // 防御力
	Speed             float64 // 速度
	CritRate          float64 // 暴击率
	ComboRate         float64 // 连击率
	CounterRate       float64 // 反击率
	StunRate          float64 // 眩晕率
	DodgeRate         float64 // 闪避率
	VampireRate       float64 // 吸血率
	CritResist        float64 // 抗暴击
	ComboResist       float64 // 抗连击
	CounterResist     float64 // 抗反击
	StunResist        float64 // 抗眩晕
	DodgeResist       float64 // 抗闪避
	VampireResist     float64 // 抗吸血
	HealBoost         float64 // 强化治疗
	CritDamageBoost   float64 // 强化爆伤
	CritDamageReduce  float64 // 弱化爆伤
	FinalDamageBoost  float64 // 最终增伤
	FinalDamageReduce float64 // 最终减伤
	CombatBoost       float64 // 战斗属性提升
	ResistanceBoost   float64 // 战斗抗性提升
}

// 伤害结果
type DamageResult struct {
	Damage    float64 // 计算后的伤害
	IsCrit    bool    // 是否暴击
	IsCombo   bool    // 是否连击
	IsVampire bool    // 是否吸血
	IsStun    bool    // 是否眩晕
}

// 被伤害结果
type TakeDamageResult struct {
	Dodged        bool    // 是否闪避
	Damage        float64 // 实际伤害
	CurrentHealth float64 // 当前生命值
	IsDead        bool    // 是否死亡
	IsCounter     bool    // 是否反击
}

// 回合战斗结果 - 单个回合的完整结果
type RoundData struct {
	Round        int           `json:"round"`             // 回合数
	PlayerHealth float64       `json:"playerHealth"`      // 玩家血量
	EnemyHealth  float64       `json:"enemyHealth"`       // 敌人血量
	Logs         []string      `json:"logs"`              // 本回合日志
	BattleEnded  bool          `json:"battleEnded"`       // 战斗是否结束
	Victory      bool          `json:"victory"`           // 是否胜利 (战斗结束时有效)
	Rewards      []interface{} `json:"rewards,omitempty"` // 奖励 (战斗结束时有效)
}

// 战斗状态 - 保存在Redis中的战斗状态信息
type BattleStatus struct {
	UserID          uint                   `json:"userID"`          // 用户ID
	Floor           int                    `json:"floor"`           // 当前层数
	Difficulty      string                 `json:"difficulty"`      // 难度
	Round           int                    `json:"round"`           // 当前回合
	PlayerHealth    float64                `json:"playerHealth"`    // 玩家当前血量
	PlayerMaxHealth float64                `json:"playerMaxHealth"` // 玩家最大血量
	EnemyHealth     float64                `json:"enemyHealth"`     // 敌人当前血量
	EnemyMaxHealth  float64                `json:"enemyMaxHealth"`  // 敌人最大血量
	PlayerStats     *CombatStats           `json:"playerStats"`     // 玩家战斗属性
	EnemyStats      *CombatStats           `json:"enemyStats"`      // 敌人战斗属性
	BuffEffects     map[string]interface{} `json:"buffEffects"`     // 玩家已选增益
	BattleLog       []string               `json:"battleLog"`       // 完整战斗日志
}
