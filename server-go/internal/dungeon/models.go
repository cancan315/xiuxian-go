package dungeon

// 秘境请求
type DungeonRequest struct {
	Difficulty string `json:"difficulty"` // 难度: easy, normal, hard, expert
	Action     string `json:"action"`     // 操作: start, get_options, select_buff, fight, end
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
	Difficulty   string `json:"difficulty"`  // 难度
	RefreshCount int    `json:"refreshCount"` // 刷新次数
}

// 增益选项
type BuffOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // common, rare, epic
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
	Success      bool   `json:"success"`
	Round        int    `json:"round"`
	MaxRounds    int    `json:"maxRounds"`
	PlayerHealth float64 `json:"playerHealth"`
	EnemyHealth  float64 `json:"enemyHealth"`
	Log          string `json:"log"`
}

// 战斗结果
type FightResult struct {
	Success     bool        `json:"success"`
	Victory     bool        `json:"victory"`
	Floor       int         `json:"floor"`
	Message     string      `json:"message"`
	Rewards     []interface{} `json:"rewards"`
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
	UserID       uint
	Floor        int
	Difficulty   string
	RefreshCount int
	SelectedBuffs []string // 已选增益ID
	PlayerBuffs  map[string]interface{} // 玩家已获得的增益效果
}
