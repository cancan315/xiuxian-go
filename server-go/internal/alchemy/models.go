package alchemy

import "gorm.io/datatypes"

// 丹方请求
type AlchemyRequest struct {
	Action        string `json:"action"`         // 操作: list_recipes, get_recipe, calculate_effect, craft_pill, buy_fragment
	RecipeID      string `json:"recipeId,omitempty"`
	Grade         string `json:"grade,omitempty"`
	Type          string `json:"type,omitempty"`
	PlayerLevel   int    `json:"playerLevel,omitempty"`
}

// 丹方响应
type AlchemyResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// 丹方品阶配置
type PillGrade struct {
	ID           string  `json:"id"`           // grade1-grade9
	Name         string  `json:"name"`         // 一品-九品
	Difficulty   float64 `json:"difficulty"`   // 难度系数
	SuccessRate  float64 `json:"successRate"`  // 基础成功率
	FragmentsNeeded int  `json:"fragmentsNeeded"` // 所需残页数
}

// 丹药类型倍数
type PillType struct {
	ID              string  `json:"id"`             // spirit, cultivation, attribute, special
	Name            string  `json:"name"`           // 灵力类、修炼类、属性类、特殊类
	EffectMultiplier float64 `json:"effectMultiplier"` // 效果倍数
}

// 灵草品质等级
type HerbQuality struct {
	ID       string  `json:"id"`       // common, fine, rare, epic, immortal
	Name     string  `json:"name"`     // 普通、精品、珍品、仙品
	Modifier float64 `json:"modifier"` // 品质倍数 (1.0, 1.5, 2.0, 3.0, 5.0)
}

// 灵草配置
type HerbConfig struct {
	ID          string `json:"id"`          // herb_id
	Name        string `json:"name"`        // 灵草名称
	Description string `json:"description"`
	Price       int    `json:"price"`       // 灵石价格
}

// 丹方配置
type RecipeConfig struct {
	ID              string            `json:"id"`               // recipe_id
	Name            string            `json:"name"`             // 丹药名称
	Description     string            `json:"description"`      // 描述
	Grade           string            `json:"grade"`            // grade1-grade9
	Type            string            `json:"type"`             // spirit, cultivation, attribute, special
	Materials       []MaterialRequire  `json:"materials"`        // 材料需求
	BaseEffect      PillEffect         `json:"baseEffect"`       // 基础效果
	FragmentsNeeded int               `json:"fragmentsNeeded"`  // 所需残页数量
}

// 材料需求
type MaterialRequire struct {
	HerbID string `json:"herbId"`
	Count  int    `json:"count"`
}

// 丹药效果
type PillEffect struct {
	Type     string  `json:"type"`     // 效果类型
	Value    float64 `json:"value"`    // 效果数值
	Duration int     `json:"duration"` // 持续时间(秒)
}

// 丹方学习响应
type RecipeInfo struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Grade           string            `json:"grade"`
	GradeName       string            `json:"gradeName"`
	Type            string            `json:"type"`
	TypeName        string            `json:"typeName"`
	Materials       []MaterialInfo     `json:"materials"`
	FragmentsNeeded int               `json:"fragmentsNeeded"`
	CurrentEffect   PillEffectResult   `json:"currentEffect"`
}

// 材料信息
type MaterialInfo struct {
	HerbID   string `json:"herbId"`
	HerbName string `json:"herbName"`
	Count    int    `json:"count"`
	Owned    int    `json:"owned"`
}

// 计算后的丹药效果
type PillEffectResult struct {
	Type        string  `json:"type"`
	Value       float64 `json:"value"`
	Duration    int     `json:"duration"`
	SuccessRate float64 `json:"successRate"`
}

// 炼制请求
type CraftRequest struct {
	RecipeID         string            `json:"recipeId"`
	PlayerLevel      int               `json:"playerLevel"`
	UnlockedRecipes  []string          `json:"unlockedRecipes"`  // 已掌握的丹方ID列表
	InventoryHerbs   map[string]int    `json:"inventoryHerbs"`   // 灵草库存 {herbId: count}
	Luck             float64           `json:"luck"`             // 幸运值
	AlchemyRate      float64           `json:"alchemyRate"`      // 炼丹加成率
}

// 炼制响应
type CraftResult struct {
	Success       bool              `json:"success"`
	Message       string            `json:"message"`
	PillID        string            `json:"pillId,omitempty"`         // 新创建的丹药ID
	PillName      string            `json:"pillName,omitempty"`       // 丹药名称
	SuccessRate   float64           `json:"successRate,omitempty"`    // 成功率
	ConsumedHerbs map[string]int    `json:"consumedHerbs,omitempty"` // 消耗的灵草
	PillEffect    *PillEffectResult `json:"pillEffect,omitempty"`    // 丹药效果
}

// 购买残页请求
type BuyFragmentRequest struct {
	RecipeID         string   `json:"recipeId"`
	Quantity         int      `json:"quantity"`         // 购买份数
	CurrentFragments int      `json:"currentFragments"` // 当前残页数量
	UnlockedRecipes  []string `json:"unlockedRecipes"`  // 已掌握的丹方ID列表
}

// 购买残页响应
type BuyFragmentResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RecipeID     string `json:"recipeId,omitempty"`
	FragmentsOwned int  `json:"fragmentsOwned,omitempty"`
	RecipeUnlocked bool `json:"recipeUnlocked,omitempty"` // 是否合成完整丹方
}

// 丹方列表响应
type RecipeListResponse struct {
	Unlocked   []RecipeInfo `json:"unlocked"`   // 已掌握的丹方
	Locked     []RecipeInfo `json:"locked"`     // 未掌握的丹方（显示残页进度）
}

// 用户炼丹状态
type UserAlchemyData struct {
	RecipesUnlocked map[string]bool   `json:"recipesUnlocked"`  // {recipeId: true}
	Fragments       map[string]int    `json:"fragments"`        // {recipeId: count}
	Pills           []UserPill        `json:"pills"`            // 拥有的丹药
	PillsCrafted    int               `json:"pillsCrafted"`     // 总炼制次数
	PillsConsumed   int               `json:"pillsConsumed"`    // 总服用次数
	AlchemyLevel    int               `json:"alchemyLevel"`     // 炼丹等级
	AlchemyRate     float64           `json:"alchemyRate"`      // 炼丹加成率
}

// 用户拥有的丹药
type UserPill struct {
	ID          string                 `json:"id"`          // 丹药实例ID
	RecipeID    string                 `json:"recipeId"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      datatypes.JSON         `json:"effect"`
	CreatedAt   int64                  `json:"createdAt"`
}

// 获取丹方详情响应
type RecipeDetailResponse struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Grade           string            `json:"grade"`
	GradeName       string            `json:"gradeName"`
	Type            string            `json:"type"`
	TypeName        string            `json:"typeName"`
	SuccessRate     float64           `json:"successRate"`
	Materials       []MaterialInfo     `json:"materials"`
	FragmentsNeeded int               `json:"fragmentsNeeded"`
	CurrentFragments int              `json:"currentFragments"` // 用户当前拥有的残页
	IsUnlocked      bool              `json:"isUnlocked"`
	CurrentEffect   PillEffectResult   `json:"currentEffect"`
	PillsInInventory int              `json:"pillsInInventory"` // 用户背包中该丹药数量
}

// 所有丹方列表响应
type AllRecipesResponse struct {
	Grades    map[string]PillGrade   `json:"grades"`    // 品阶配置
	Types     map[string]PillType    `json:"types"`     // 类型配置
	Recipes   []RecipeDetailResponse `json:"recipes"`   // 丹方列表
	PlayerStats UserAlchemyData      `json:"playerStats"` // 玩家统计
}
