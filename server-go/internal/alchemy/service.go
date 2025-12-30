package alchemy

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// AlchemyService 炼丹服务
type AlchemyService struct {
	userID uint
}

// NewAlchemyService 创建炼丹服务实例
func NewAlchemyService(userID uint) *AlchemyService {
	return &AlchemyService{
		userID: userID,
	}
}

// 品阶配置
var pillGrades = map[string]PillGrade{
	"grade1": {ID: "grade1", Name: "一品", Difficulty: 1, SuccessRate: 0.6, FragmentsNeeded: 10},
	"grade2": {ID: "grade2", Name: "二品", Difficulty: 1.2, SuccessRate: 0.5, FragmentsNeeded: 15},
	"grade3": {ID: "grade3", Name: "三品", Difficulty: 1.5, SuccessRate: 0.3, FragmentsNeeded: 20},
	"grade4": {ID: "grade4", Name: "四品", Difficulty: 2, SuccessRate: 0.1, FragmentsNeeded: 25},
	"grade5": {ID: "grade5", Name: "五品", Difficulty: 2.5, SuccessRate: 0.1, FragmentsNeeded: 30},
	"grade6": {ID: "grade6", Name: "六品", Difficulty: 3, SuccessRate: 0.1, FragmentsNeeded: 35},
	"grade7": {ID: "grade7", Name: "七品", Difficulty: 4, SuccessRate: 0.1, FragmentsNeeded: 40},
	"grade8": {ID: "grade8", Name: "八品", Difficulty: 5, SuccessRate: 0.1, FragmentsNeeded: 45},
	"grade9": {ID: "grade9", Name: "九品", Difficulty: 6, SuccessRate: 0.1, FragmentsNeeded: 50},
}

// 丹药类型配置
var pillTypes = map[string]PillType{
	"spirit":      {ID: "spirit", Name: "灵力类", EffectMultiplier: 1.0},
	"cultivation": {ID: "cultivation", Name: "修炼类", EffectMultiplier: 1.2},
	"attribute":   {ID: "attribute", Name: "属性类", EffectMultiplier: 1.5},
	"special":     {ID: "special", Name: "特殊类", EffectMultiplier: 2.0},
}

// 灵草配置
var herbConfigs = map[string]HerbConfig{
	"spirit_grass":         {ID: "spirit_grass", Name: "灵精草"},
	"cloud_flower":         {ID: "cloud_flower", Name: "云雾花"},
	"thunder_root":         {ID: "thunder_root", Name: "雷击根"},
	"dragon_breath_herb":   {ID: "dragon_breath_herb", Name: "龙息草"},
	"immortal_jade_grass":  {ID: "immortal_jade_grass", Name: "仙玉草"},
	"dark_yin_grass":       {ID: "dark_yin_grass", Name: "玄阴草"},
	"nine_leaf_lingzhi":    {ID: "nine_leaf_lingzhi", Name: "九叶灵芝"},
	"purple_ginseng":       {ID: "purple_ginseng", Name: "紫金参"},
	"frost_lotus":          {ID: "frost_lotus", Name: "寒霜莲"},
	"fire_heart_flower":    {ID: "fire_heart_flower", Name: "火心花"},
	"moonlight_orchid":     {ID: "moonlight_orchid", Name: "月华兰"},
	"sun_essence_flower":   {ID: "sun_essence_flower", Name: "日精花"},
	"five_elements_grass":  {ID: "five_elements_grass", Name: "五行草"},
	"phoenix_feather_herb": {ID: "phoenix_feather_herb", Name: "凤羽草"},
	"celestial_dew_grass":  {ID: "celestial_dew_grass", Name: "天露草"},
}

// 丹方配置 (12个丹方)
var recipes = []RecipeConfig{
	{
		ID:          "spirit_gathering",
		Name:        "聚灵丹",
		Description: "十年灵草炼制，服用后恢复少量灵力",
		Grade:       "grade1",
		Type:        "spirit",
		Materials: []MaterialRequire{
			{HerbID: "spirit_grass", Count: 2},
			{HerbID: "cloud_flower", Count: 1},
		},
		FragmentsNeeded: 10,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "cultivation_boost",
		Name:        "聚气丹",
		Description: "十年灵草炼制，服用后增加少量修为",
		Grade:       "grade2",
		Type:        "cultivation",
		Materials: []MaterialRequire{
			{HerbID: "cloud_flower", Count: 2},
			{HerbID: "thunder_root", Count: 1},
		},
		FragmentsNeeded: 15,
		BaseEffect:      PillEffect{Type: "cultivation", Value: 20, Duration: 0},
	},
	{
		ID:          "spirit_recovery",
		Name:        "回灵丹",
		Description: "百年灵草炼制，服用后恢复大量灵力",
		Grade:       "grade2",
		Type:        "spirit",
		Materials: []MaterialRequire{
			{HerbID: "dark_yin_grass", Count: 2},
			{HerbID: "frost_lotus", Count: 1},
		},
		FragmentsNeeded: 15,
		BaseEffect:      PillEffect{Type: "spirit", Value: 1880, Duration: 0},
	},
	{
		ID:          "thunder_power",
		Name:        "雷灵丹",
		Description: "千年灵草炼制，蕴含狂暴的天雷能量服用后增加攻击",
		Grade:       "grade3",
		Type:        "attribute",
		Materials: []MaterialRequire{
			{HerbID: "thunder_root", Count: 2},
			{HerbID: "dragon_breath_herb", Count: 1},
		},
		FragmentsNeeded: 20,
		BaseEffect:      PillEffect{Type: "attributeAttack", Value: 10, Duration: 0},
	},
	{
		ID:          "essence_condensation",
		Name:        "凝元丹",
		Description: "千年灵草炼制，服用后增加大量修为",
		Grade:       "grade3",
		Type:        "cultivation",
		Materials: []MaterialRequire{
			{HerbID: "nine_leaf_lingzhi", Count: 2},
			{HerbID: "purple_ginseng", Count: 1},
		},
		FragmentsNeeded: 20,
		BaseEffect:      PillEffect{Type: "cultivation", Value: 50, Duration: 0},
	},
	{
		ID:          "mind_clarity",
		Name:        "清心丹",
		Description: "暂未开放",
		Grade:       "grade3",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 20,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "immortal_essence",
		Name:        "仙灵丹",
		Description: "暂未开放",
		Grade:       "grade4",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 25,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "fire_essence",
		Name:        "火元丹",
		Description: "暂未开放",
		Grade:       "grade4",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 25,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "five_elements_pill",
		Name:        "五行丹",
		Description: "暂未开放",
		Grade:       "grade5",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 30,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "celestial_essence_pill",
		Name:        "天元丹",
		Description: "暂未开放",
		Grade:       "grade6",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 35,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "sun_moon_pill",
		Name:        "日月丹",
		Description: "暂未开放",
		Grade:       "grade7",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 40,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
	{
		ID:          "phoenix_rebirth_pill",
		Name:        "涅槃丹",
		Description: "暂未开放",
		Grade:       "grade8",
		Type:        "spirit",
		Materials:   []MaterialRequire{
			// 暂未开放，可以留空或设置特殊材料
		},
		FragmentsNeeded: 45,
		BaseEffect:      PillEffect{Type: "spirit", Value: 880, Duration: 0},
	},
}

// 根据品阶计算所需残页数量
func getFragmentsNeededForGrade(grade string) int {
	if g, ok := pillGrades[grade]; ok {
		return g.FragmentsNeeded
	}
	return 0
}

// GetAllRecipes 获取所有丹方列表
func (s *AlchemyService) GetAllRecipes(playerLevel int) (*AllRecipesResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// ✅ 从数据库获取用户真实的已解锁丹方和残页数据
	var userAlchemyData models.UserAlchemyDataDB
	if err := db.DB.Where("user_id = ?", s.userID).First(&userAlchemyData).Error; err != nil {
		fmt.Printf("[DEBUG] 数据库中找不到该用户的炼丹数据: %v\n", err)
		// 如果记录不存在，创建新记录
		userAlchemyData = models.UserAlchemyDataDB{
			UserID:          s.userID,
			RecipesUnlocked: "{}",
			PillsCrafted:    0,
			PillsConsumed:   0,
			AlchemyLevel:    1,
			AlchemyRate:     1.0,
		}
		db.DB.Create(&userAlchemyData)
		fmt.Printf("[DEBUG] 已创建新的炼丹记录\n")
	} else {
		fmt.Printf("[DEBUG] 从数据库中查到用户炼丹数据: ID=%d, RecipesUnlocked=%s\n", userAlchemyData.ID, userAlchemyData.RecipesUnlocked)
	}

	// 将JSON数据转换为map
	unlockedRecipes := make(map[string]bool)
	if userAlchemyData.RecipesUnlocked != "" {
		var recipesMap map[string]bool
		if err := json.Unmarshal([]byte(userAlchemyData.RecipesUnlocked), &recipesMap); err == nil {
			unlockedRecipes = recipesMap
			fmt.Printf("[DEBUG] JSON解析成功: %v\n", unlockedRecipes)
		} else {
			fmt.Printf("[DEBUG] JSON解析失败: %v\n", err)
		}
	}

	fmt.Printf("[DEBUG] GetAllRecipes: userID=%d, RecipesUnlocked=%s, unlockedCount=%d\n", s.userID, userAlchemyData.RecipesUnlocked, len(unlockedRecipes))
	fmt.Printf("[DEBUG] unlockedRecipes map content: %+v\n", unlockedRecipes)

	fragments := make(map[string]int)
	var pillFragments []models.PillFragment
	if err := db.DB.Where("user_id = ?", s.userID).Find(&pillFragments).Error; err == nil {
		for _, fragment := range pillFragments {
			fragments[fragment.RecipeID] = fragment.Count
		}
	}

	// 构建丹方详情列表
	var recipeDetails []RecipeDetailResponse
	for _, recipe := range recipes {
		isUnlocked := unlockedRecipes[recipe.ID]
		fmt.Printf("[DEBUG] Recipe: %s (ID=%s), isUnlocked=%v\n", recipe.Name, recipe.ID, isUnlocked)
		detail := s.buildRecipeDetail(&recipe, &user, isUnlocked, fragments[recipe.ID], playerLevel)
		recipeDetails = append(recipeDetails, detail)
	}

	return &AllRecipesResponse{
		Grades:  pillGrades,
		Types:   pillTypes,
		Recipes: recipeDetails,
		PlayerStats: UserAlchemyData{
			RecipesUnlocked: unlockedRecipes,
			Fragments:       fragments,
			AlchemyLevel:    userAlchemyData.AlchemyLevel,
			AlchemyRate:     userAlchemyData.AlchemyRate,
		},
	}, nil
}

// GetRecipeDetail 获取单个丹方的详细信息
func (s *AlchemyService) GetRecipeDetail(recipeID string, playerLevel int, unlockedRecipes map[string]bool, fragments map[string]int) (*RecipeDetailResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 查找丹方
	var recipe *RecipeConfig
	for i := range recipes {
		if recipes[i].ID == recipeID {
			recipe = &recipes[i]
			break
		}
	}

	if recipe == nil {
		return nil, fmt.Errorf("丹方不存在: %s", recipeID)
	}

	isUnlocked := unlockedRecipes[recipeID]
	detail := s.buildRecipeDetail(recipe, &user, isUnlocked, fragments[recipeID], playerLevel)
	return &detail, nil
}

// buildRecipeDetail 构建丹方详情
func (s *AlchemyService) buildRecipeDetail(recipe *RecipeConfig, user *models.User, isUnlocked bool, fragmentsOwned int, playerLevel int) RecipeDetailResponse {
	// 获取品阶和类型信息
	grade := pillGrades[recipe.Grade]
	recipeType := pillTypes[recipe.Type]

	// 计算材料信息
	var materials []MaterialInfo
	for _, req := range recipe.Materials {
		herbName := "未知灵草"
		if herb, ok := herbConfigs[req.HerbID]; ok {
			herbName = herb.Name
		}
		materials = append(materials, MaterialInfo{
			HerbID:   req.HerbID,
			HerbName: herbName,
			Count:    req.Count,
			Owned:    0, // 从前端库存获取（实际应从inventory store获取）
		})
	}

	// 计算效果
	effect := s.calculatePillEffect(recipe, playerLevel)

	return RecipeDetailResponse{
		ID:               recipe.ID,
		Name:             recipe.Name,
		Description:      recipe.Description,
		Grade:            recipe.Grade,
		GradeName:        grade.Name,
		Type:             recipe.Type,
		TypeName:         recipeType.Name,
		SuccessRate:      grade.SuccessRate,
		Materials:        materials,
		FragmentsNeeded:  recipe.FragmentsNeeded,
		CurrentFragments: fragmentsOwned,
		IsUnlocked:       isUnlocked,
		CurrentEffect:    effect,
		PillsInInventory: 0, // 从前端库存获取
	}
}

// CalculatePillEffect 计算丹药实际效果
func (s *AlchemyService) calculatePillEffect(recipe *RecipeConfig, playerLevel int) PillEffectResult {
	grade := pillGrades[recipe.Grade]
	recipeType := pillTypes[recipe.Type]

	// 基础效果随境界提升
	levelMultiplier := 1.0 + float64(playerLevel-1)*0.1
	effectValue := recipe.BaseEffect.Value * recipeType.EffectMultiplier * levelMultiplier

	return PillEffectResult{
		Type:        recipe.BaseEffect.Type,
		Value:       effectValue,
		Duration:    recipe.BaseEffect.Duration,
		SuccessRate: grade.SuccessRate,
	}
}

// CraftPill 炼制丹药
func (s *AlchemyService) CraftPill(recipeID string, playerLevel int, unlockedRecipes map[string]bool, inventoryHerbs map[string]int, luck float64, alchemyRate float64) (*CraftResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 查找丹方
	var recipe *RecipeConfig
	for i := range recipes {
		if recipes[i].ID == recipeID {
			recipe = &recipes[i]
			break
		}
	}

	if recipe == nil {
		return nil, fmt.Errorf("丹方不存在: %s", recipeID)
	}

	// 检查是否掌握该丹方
	if !unlockedRecipes[recipeID] {
		return nil, fmt.Errorf("未掌握该丹方")
	}

	// 检查材料是否足够
	for _, material := range recipe.Materials {
		if inventoryHerbs[material.HerbID] < material.Count {
			return nil, fmt.Errorf("材料不足: %s", material.HerbID)
		}
	}

	// 计算成功率
	grade := pillGrades[recipe.Grade]

	// 确保 luck 和 alchemyRate 有效
	if luck <= 0 {
		luck = 1.0
	}
	if alchemyRate <= 0 {
		alchemyRate = 1.0
	}

	successRate := grade.SuccessRate * luck * alchemyRate

	// 尝试炼制
	if rand.Float64() > successRate {
		return &CraftResult{
			Success:     false,
			Message:     "炼制失败",
			SuccessRate: successRate,
		}, nil
	}

	// 计算消耗的材料
	consumedHerbs := make(map[string]int)
	for _, material := range recipe.Materials {
		consumedHerbs[material.HerbID] = material.Count
	}

	// 从数据库扣除灵草材料
	// ✅ 改进逻辑：先聚合相同herbId的记录，然后再扣除
	for herbID, count := range consumedHerbs {
		// 查询该herbId的所有记录
		var herbs []models.Herb
		if err := db.DB.Where("user_id = ? AND herb_id = ?", s.userID, herbID).Find(&herbs).Error; err != nil {
			return nil, fmt.Errorf("查询灵草失败: %w", err)
		}

		// 计算总数
		totalCount := 0
		for _, h := range herbs {
			totalCount += h.Count
		}

		// 如果总数小于需要扣除的数量，则报错
		if totalCount < count {
			return nil, fmt.Errorf("材料不足: %s, 拥有: %d, 需要: %d", herbID, totalCount, count)
		}

		// 逐个扣除，直到扣除足够的数量
		remaining := count
		for _, herb := range herbs {
			if remaining <= 0 {
				break
			}

			if herb.Count >= remaining {
				// 当前记录足以扣除剩余数量
				if err := db.DB.Model(&herb).Update("count", herb.Count-remaining).Error; err != nil {
					return nil, fmt.Errorf("扣除材料失败: %w", err)
				}
				remaining = 0
			} else {
				// 当前记录不足，全部扣除
				remaining -= herb.Count
				if err := db.DB.Model(&herb).Update("count", 0).Error; err != nil {
					return nil, fmt.Errorf("扣除材料失败: %w", err)
				}
			}
		}
	}

	// 创建丹药记录 (保存到数据库)
	pill := models.Pill{
		UserID:      s.userID,
		PillID:      recipeID,
		Name:        recipe.Name,
		Description: recipe.Description,
	}

	// 计算并保存丹药效果到JSON字段
	effect := s.calculatePillEffect(recipe, playerLevel)
	effectJSON, _ := json.Marshal(effect)
	pill.Effect = effectJSON

	if err := db.DB.Create(&pill).Error; err != nil {
		return nil, fmt.Errorf("创建丹药失败: %w", err)
	}

	// 更新用户炼丹统计数据
	var userAlchemyData models.UserAlchemyDataDB
	if err := db.DB.Where("user_id = ?", s.userID).First(&userAlchemyData).Error; err != nil {
		return nil, fmt.Errorf("获取用户炼丹数据失败: %w", err)
	}

	userAlchemyData.PillsCrafted++
	if err := db.DB.Save(&userAlchemyData).Error; err != nil {
		return nil, fmt.Errorf("更新用户炼丹数据失败: %w", err)
	}

	return &CraftResult{
		Success:       true,
		Message:       "炼制成功",
		PillID:        fmt.Sprintf("%d", pill.ID),
		PillName:      recipe.Name,
		SuccessRate:   successRate,
		ConsumedHerbs: consumedHerbs,
		PillEffect:    &effect,
	}, nil
}

// BuyFragment 购买丹方残页
func (s *AlchemyService) BuyFragment(recipeID string, quantity int, currentFragments int, unlockedRecipes map[string]bool) (*BuyFragmentResult, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 查找丹方
	var recipe *RecipeConfig
	for i := range recipes {
		if recipes[i].ID == recipeID {
			recipe = &recipes[i]
			break
		}
	}

	if recipe == nil {
		return nil, fmt.Errorf("丹方不存在: %s", recipeID)
	}

	// 计算花费灵石
	// 根据品阶计算价格：grade1=100灵石/份, grade2=150, etc.
	gradeNum := 0
	for i := 1; i <= 9; i++ {
		if fmt.Sprintf("grade%d", i) == recipe.Grade {
			gradeNum = i
			break
		}
	}
	price := 50000 + (gradeNum-1)*50 // grade1=100, grade2=150, ...

	totalCost := price * quantity
	if user.SpiritStones < totalCost {
		return nil, fmt.Errorf("灵石不足")
	}

	// 扣除灵石
	user.SpiritStones -= totalCost
	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, fmt.Errorf("更新灵石失败: %w", err)
	}

	// 更新残页数量到数据库
	var fragment models.PillFragment
	if err := db.DB.Where("user_id = ? AND recipe_id = ?", s.userID, recipeID).First(&fragment).Error; err != nil {
		// 如果不存在，则创建新记录
		fragment = models.PillFragment{
			UserID:   s.userID,
			RecipeID: recipeID,
			Count:    quantity,
		}
		if err := db.DB.Create(&fragment).Error; err != nil {
			return nil, fmt.Errorf("创建残页记录失败: %w", err)
		}
	} else {
		// 更新现有记录
		fragment.Count += quantity
		if err := db.DB.Save(&fragment).Error; err != nil {
			return nil, fmt.Errorf("更新残页记录失败: %w", err)
		}
	}

	// 检查是否可以合成完整丹方
	newFragmentCount := fragment.Count
	recipeUnlocked := false
	newFragmentsAfterCraft := newFragmentCount

	fmt.Printf("[DEBUG BuyFragment] recipeID=%s, currentFragments=%d, needed=%d\n", recipeID, newFragmentCount, recipe.FragmentsNeeded)

	if newFragmentCount >= recipe.FragmentsNeeded && !unlockedRecipes[recipeID] {
		fmt.Printf("[DEBUG BuyFragment] 合成条件满足，开始合成\n")
		newFragmentsAfterCraft = newFragmentCount - recipe.FragmentsNeeded
		recipeUnlocked = true

		// 更新残页数量（扣除用于合成的数量）
		fragment.Count = newFragmentsAfterCraft
		if err := db.DB.Save(&fragment).Error; err != nil {
			return nil, fmt.Errorf("更新残页记录失败: %w", err)
		}

		// ✅ 更新 UserAlchemyDataDB 中的 RecipesUnlocked 字段
		var userAlchemyData models.UserAlchemyDataDB
		if err := db.DB.Where("user_id = ?", s.userID).First(&userAlchemyData).Error; err == nil {
			unlockedRecipes[recipeID] = true
			recipesJSON, _ := json.Marshal(unlockedRecipes)
			userAlchemyData.RecipesUnlocked = string(recipesJSON)
			fmt.Printf("[DEBUG BuyFragment] 更新RecipesUnlocked: %s\n", userAlchemyData.RecipesUnlocked)
			if err := db.DB.Save(&userAlchemyData).Error; err != nil {
				return nil, fmt.Errorf("更新已解锁丹方失败: %w", err)
			}
			fmt.Printf("[DEBUG BuyFragment] 合成成功\n")
		}
	} else {
		fmt.Printf("[DEBUG BuyFragment] 合成条件不满足 (fragCount=%d >= needed=%d: %v, unlocked=%v)\n", newFragmentCount, recipe.FragmentsNeeded, newFragmentCount >= recipe.FragmentsNeeded, unlockedRecipes[recipeID])
	}

	return &BuyFragmentResult{
		Success:        true,
		Message:        "购买成功",
		RecipeID:       recipeID,
		FragmentsOwned: newFragmentsAfterCraft,
		RecipeUnlocked: recipeUnlocked,
	}, nil
}

// GetUnlockedRecipes 获取已解锁的丹方列表
func (s *AlchemyService) GetUnlockedRecipes(unlockedRecipes map[string]bool, playerLevel int) ([]RecipeDetailResponse, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	var result []RecipeDetailResponse
	for _, recipe := range recipes {
		if unlockedRecipes[recipe.ID] {
			detail := s.buildRecipeDetail(&recipe, &user, true, 0, playerLevel)
			result = append(result, detail)
		}
	}
	return result, nil
}

// GetRecipeByID 根据ID获取丹方配置
func GetRecipeByID(recipeID string) *RecipeConfig {
	for i := range recipes {
		if recipes[i].ID == recipeID {
			return &recipes[i]
		}
	}
	return nil
}

// GetHerbNameByID 获取灵草名称
func GetHerbNameByID(herbID string) string {
	if herb, ok := herbConfigs[herbID]; ok {
		return herb.Name
	}
	return "未知灵草"
}

// GetPillGradeName 获取品阶名称
func GetPillGradeName(grade string) string {
	if g, ok := pillGrades[grade]; ok {
		return g.Name
	}
	return "未知品阶"
}

// GetPillTypeName 获取类型名称
func GetPillTypeName(pillType string) string {
	if t, ok := pillTypes[pillType]; ok {
		return t.Name
	}
	return "未知类型"
}

// GetAllRecipeConfigs 获取所有丹方配置 (用于前端初始化)
func GetAllRecipeConfigs() []RecipeConfig {
	return recipes
}

// GetAllGrades 获取所有品阶配置
func GetAllGrades() map[string]PillGrade {
	return pillGrades
}

// GetAllTypes 获取所有类型配置
func GetAllTypes() map[string]PillType {
	return pillTypes
}

// GetAllHerbs 获取所有灵草配置
func GetAllHerbs() map[string]HerbConfig {
	return herbConfigs
}

// GetUserAlchemyData 获取用户炼丹数据
func (s *AlchemyService) GetUserAlchemyData() (*UserAlchemyData, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}

	// 获取用户的丹方残页数据
	var fragments []models.PillFragment
	if err := db.DB.Where("user_id = ?", s.userID).Find(&fragments).Error; err != nil {
		return nil, fmt.Errorf("获取残页数据失败: %w", err)
	}

	// 获取用户的炼丹统计数据
	var userAlchemyData models.UserAlchemyDataDB
	if err := db.DB.Where("user_id = ?", s.userID).First(&userAlchemyData).Error; err != nil {
		// 如果不存在，则创建默认数据
		userAlchemyData = models.UserAlchemyDataDB{
			UserID:        s.userID,
			PillsCrafted:  0,
			PillsConsumed: 0,
			AlchemyLevel:  1,
			AlchemyRate:   1.0,
		}
		db.DB.Create(&userAlchemyData)
	}

	// 构建fragments map
	fragmentsMap := make(map[string]int)
	for _, fragment := range fragments {
		fragmentsMap[fragment.RecipeID] = fragment.Count
	}

	// 构建已解锁的丹方map
	unlockedRecipes := make(map[string]bool)
	// 优先从数据库中的 recipes_unlocked 字段获取
	if userAlchemyData.RecipesUnlocked != "" {
		var recipesMap map[string]bool
		if err := json.Unmarshal([]byte(userAlchemyData.RecipesUnlocked), &recipesMap); err == nil {
			unlockedRecipes = recipesMap
		}
	}
	// 如果 recipes_unlocked 为空或解析失败，根据残页数量判断是否解锁
	if len(unlockedRecipes) == 0 {
		for recipeID, count := range fragmentsMap {
			// 查找对应的丹方配置
			for _, recipe := range recipes {
				if recipe.ID == recipeID && count >= recipe.FragmentsNeeded {
					unlockedRecipes[recipeID] = true
					break
				}
			}
		}
	}

	return &UserAlchemyData{
		RecipesUnlocked: unlockedRecipes,
		Fragments:       fragmentsMap,
		PillsCrafted:    userAlchemyData.PillsCrafted,
		PillsConsumed:   userAlchemyData.PillsConsumed,
		AlchemyLevel:    userAlchemyData.AlchemyLevel,
		AlchemyRate:     userAlchemyData.AlchemyRate,
	}, nil
}
