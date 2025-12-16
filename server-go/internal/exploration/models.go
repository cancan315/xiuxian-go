package exploration

// 事件类型常量
const (
	EventTypeItemFound          = "item_found"
	EventTypeSpiritStoneFound   = "spirit_stone_found"
	EventTypeHerbFound          = "herb_found"
	EventTypePillRecipeFragment = "pill_recipe_fragment_found"
	EventTypeBattleEncounter    = "battle_encounter"
)

// 随机事件类型
const (
	RandomEventAncientTablet        = "ancient_tablet"
	RandomEventSpiritSpring         = "spirit_spring"
	RandomEventAncientMaster        = "ancient_master"
	RandomEventMonsterAttack        = "monster_attack"
	RandomEventCultivationDeviation = "cultivation_deviation"
	RandomEventTreasureTrove        = "treasure_trove"
	RandomEventEnlightenment        = "enlightenment"
	RandomEventQiDeviation          = "qi_deviation"
)

// ExplorationEvent 探索事件结构
type ExplorationEvent struct {
	Type        string        `json:"type"`        // 事件类型
	Description string        `json:"description"` // 事件描述
	Item        interface{}   `json:"item,omitempty"`
	Amount      int           `json:"amount,omitempty"`
	Herb        interface{}   `json:"herb,omitempty"`
	Enemy       interface{}   `json:"enemy,omitempty"`
	RecipeID    string        `json:"recipeId,omitempty"`
	Fragments   int           `json:"fragments,omitempty"`
	Choices     []EventChoice `json:"choices"`
}

// EventChoice 事件选择项
type EventChoice struct {
	Text  string      `json:"text"`
	Value interface{} `json:"value,omitempty"`
}

// ExplorationResult 探索结果
type ExplorationResult struct {
	Events []ExplorationEvent `json:"events"`
	Log    string             `json:"log"`
}

// Herb 灵草信息
type Herb struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Quality   string  `json:"quality"`
	Value     int     `json:"value"`
	BaseValue int     `json:"-"`
	Category  string  `json:"-"`
	Chance    float64 `json:"-"`
}

// HerbConfig 灵草配置
type HerbConfig struct {
	ID          string
	Name        string
	Description string
	BaseValue   int
	Category    string
	Chance      float64
}

// PillRecipe 丹药配方
type PillRecipe struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Grade           string `json:"grade"`
	Type            string `json:"type"`
	FragmentsNeeded int    `json:"fragmentsNeeded"`
}

// RandomEventConfig 随机事件配置
type RandomEventConfig struct {
	ID          string
	Name        string
	Description string
	Chance      float64
	EventType   string
	Effect      func(playerStats *PlayerStats) error
}

// PlayerStats 玩家统计属性
type PlayerStats struct {
	Level               int     `json:"level"`
	Cultivation         float64 `json:"cultivation"`
	Spirit              float64 `json:"spirit"`
	SpiritStones        int     `json:"spiritStones"`
	Luck                float64 `json:"luck"`
	EventTriggered      int     `json:"eventTriggered"`
	ExplorationCount    int     `json:"explorationCount"`
	ItemsFound          int     `json:"itemsFound"`
	UnlockedPillRecipes int     `json:"unlockedPillRecipes"`
}

// ExplorationRequest 探索请求
type ExplorationRequest struct{}

// ExplorationResponse 探索响应
type ExplorationResponse struct {
	Success bool               `json:"success"`
	Events  []ExplorationEvent `json:"events"`
	Log     string             `json:"log"`
	Error   string             `json:"error,omitempty"`
}

// EventChoiceRequest 事件选择请求
type EventChoiceRequest struct {
	EventType string      `json:"eventType"`
	Choice    interface{} `json:"choice"`
}

// EventChoiceResponse 事件选择响应
type EventChoiceResponse struct {
	Success bool        `json:"success"`
	Rewards interface{} `json:"rewards"`
	Error   string      `json:"error,omitempty"`
}
