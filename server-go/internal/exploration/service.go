package exploration

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"gorm.io/datatypes"
)

// ExplorationService 探索服务
type ExplorationService struct {
	userID uint
}

// NewExplorationService 创建探索服务实例
func NewExplorationService(userID uint) *ExplorationService {
	return &ExplorationService{
		userID: userID,
	}
}

// StartExploration 开始探索
func (s *ExplorationService) StartExploration(duration int) ([]ExplorationEvent, string, error) {
	// 获取玩家数据
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	// 使用随机种子
	r := rand.New(rand.NewSource(rand.Int63()))

	var events []ExplorationEvent
	var logs []string

	// 根据幸运值计算事件触发概率
	// 从 base_attributes 中获取 luck
	luck := s.calculateLuck(&user)
	eventChance := 0.3 * luck // 基础触发概率为30%，受幸运值影响

	// 模拟若干次事件检查
	checkCount := int(math.Ceil(float64(duration) / 1000.0)) // 每1秒检查一次
	for i := 0; i < checkCount; i++ {
		if r.Float64() < eventChance {
			event := s.triggerRandomEvent(&user, r)
			if event != nil {
				events = append(events, *event)
				logs = append(logs, fmt.Sprintf("[事件]%s", event.Description))
			}
		}
	}

	// 如果没有触发事件，至少生成一条基础日志
	if len(logs) == 0 {
		logs = append(logs, "探索了一段时间，未发生特殊事件")
	}

	// 更新探索统计
	user.SpiritStones += 10 // 每次探索基础奖励
	if err := db.DB.Model(&user).Update("spirit_stones", user.SpiritStones).Error; err != nil {
		return nil, "", fmt.Errorf("failed to update user: %w", err)
	}

	logStr := ""
	for _, log := range logs {
		logStr += log + "\n"
	}

	return events, logStr, nil
}

// HandleEventChoice 处理事件选择
func (s *ExplorationService) HandleEventChoice(eventType string, choice interface{}) (interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	switch eventType {
	case EventTypeItemFound:
		// 获得物品
		return s.handleItemFound(&user, choice)
	case EventTypeSpiritStoneFound:
		// 获得灵石
		return s.handleSpiritStoneFound(&user, choice)
	case EventTypeHerbFound:
		// 获得灵草
		return s.handleHerbFound(&user, choice)
	case EventTypePillRecipeFragment:
		// 获得丹方残页
		return s.handlePillRecipeFragment(&user, choice)
	case EventTypeBattleEncounter:
		// 战斗遭遇
		return s.handleBattleEncounter(&user, choice)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

// triggerRandomEvent 触发随机事件
func (s *ExplorationService) triggerRandomEvent(user *models.User, r *rand.Rand) *ExplorationEvent {
	events := []struct {
		name        string
		description string
		chance      float64
		handler     func(*models.User, *rand.Rand) *ExplorationEvent
	}{
		{
			name:        "古老石碑",
			description: "发现一块刻有上古功法的石碑。",
			chance:      0.08,
			handler:     s.eventAncientTablet,
		},
		{
			name:        "灵泉",
			description: "偶遇一处天然灵泉。",
			chance:      0.12,
			handler:     s.eventSpiritSpring,
		},
		{
			name:        "古修遗府",
			description: "意外发现一位上古大能的洞府。",
			chance:      0.03,
			handler:     s.eventAncientMaster,
		},
		{
			name:        "妖兽袭击",
			description: "遭遇一只实力强大的妖兽。",
			chance:      0.15,
			handler:     s.eventMonsterAttack,
		},
		{
			name:        "走火入魔",
			description: "修炼出现偏差，走火入魔。",
			chance:      0.12,
			handler:     s.eventCultivationDeviation,
		},
		{
			name:        "秘境宝藏",
			description: "发现一处上古修士遗留的宝藏。",
			chance:      0.05,
			handler:     s.eventTreasureTrove,
		},
		{
			name:        "顿悟",
			description: "修炼中突然顿悟。",
			chance:      0.08,
			handler:     s.eventEnlightenment,
		},
		{
			name:        "心魔侵扰",
			description: "遭受心魔侵扰，修为受损。",
			chance:      0.15,
			handler:     s.eventQiDeviation,
		},
		{
			name:        "灵草发现",
			description: "在一处隐秘山谷中发现了珍贵的灵草。",
			chance:      0.10,
			handler:     s.eventHerbDiscovery,
		},
		{
			name:        "丹方残页",
			description: "在古籍中找到了一张丹方残页。",
			chance:      0.07,
			handler:     s.eventPillRecipeFragment,
		},
	}

	for _, evt := range events {
		if r.Float64() <= evt.chance {
			event := evt.handler(user, r)
			if event != nil {
				event.Description = fmt.Sprintf("[%s]%s", evt.name, evt.description)
			}
			return event
		}
	}

	return nil
}

// Event handlers
func (s *ExplorationService) eventAncientTablet(user *models.User, r *rand.Rand) *ExplorationEvent {
	bonus := int(math.Floor(30 * (float64(user.Level)/5 + 1)))
	user.Cultivation = math.Round((user.Cultivation+float64(bonus))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("领悟石碑上的功法，获得%d点修为", bonus),
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventSpiritSpring(user *models.User, r *rand.Rand) *ExplorationEvent {
	bonus := int(math.Floor(60 * (float64(user.Level)/3 + 1)))
	user.Spirit += float64(bonus)
	db.DB.Model(user).Update("spirit", user.Spirit)

	return &ExplorationEvent{
		Type:        EventTypeSpiritStoneFound,
		Description: fmt.Sprintf("饮用灵泉，灵力增加%d点", bonus),
		Amount:      bonus,
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventAncientMaster(user *models.User, r *rand.Rand) *ExplorationEvent {
	cultivationBonus := int(math.Floor(120 * (float64(user.Level)/2 + 1)))
	spiritBonus := int(math.Floor(180 * (float64(user.Level)/2 + 1)))
	user.Cultivation = math.Round((user.Cultivation+float64(cultivationBonus))*10) / 10
	user.Spirit += float64(spiritBonus)
	db.DB.Model(user).Updates(map[string]interface{}{
		"cultivation": user.Cultivation,
		"spirit":      user.Spirit,
	})

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("获得上古大能传承，修为增加%d点，灵力增加%d点", cultivationBonus, spiritBonus),
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventMonsterAttack(user *models.User, r *rand.Rand) *ExplorationEvent {
	damage := int(math.Floor(80 * (float64(user.Level)/4 + 1)))
	user.Spirit = math.Max(0, user.Spirit-float64(damage))
	db.DB.Model(user).Update("spirit", user.Spirit)

	return &ExplorationEvent{
		Type:        EventTypeBattleEncounter,
		Description: fmt.Sprintf("与妖兽激战，损失%d点灵力", damage),
		Amount:      damage,
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventCultivationDeviation(user *models.User, r *rand.Rand) *ExplorationEvent {
	damage := int(math.Floor(50 * (float64(user.Level)/3 + 1)))
	user.Cultivation = math.Round(math.Max(0, user.Cultivation-float64(damage))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("走火入魔，损失%d点修为", damage),
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventTreasureTrove(user *models.User, r *rand.Rand) *ExplorationEvent {
	stoneBonus := int(math.Floor(30 * (float64(user.Level)/2 + 1)))
	user.SpiritStones += stoneBonus
	db.DB.Model(user).Update("spirit_stones", user.SpiritStones)

	return &ExplorationEvent{
		Type:        EventTypeSpiritStoneFound,
		Description: fmt.Sprintf("发现宝藏，获得%d颗灵石", stoneBonus),
		Amount:      stoneBonus,
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventEnlightenment(user *models.User, r *rand.Rand) *ExplorationEvent {
	bonus := int(math.Floor(50 * (float64(user.Level)/4 + 1)))
	user.Cultivation = math.Round((user.Cultivation+float64(bonus))*10) / 10

	// 获取玩家属性并更新 spiritRate
	attrs := s.getPlayerAttributes(user)
	if spiritRate, ok := attrs["spiritRate"].(float64); ok {
		attrs["spiritRate"] = spiritRate * 1.05 // 提升5%
	} else {
		attrs["spiritRate"] = 1.05
	}
	s.setPlayerAttributes(user, attrs)

	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("突然顿悟，获得%d点修为，灵力获取速率提升5%%", bonus),
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

func (s *ExplorationService) eventQiDeviation(user *models.User, r *rand.Rand) *ExplorationEvent {
	damage := int(math.Floor(60 * (float64(user.Level)/3 + 1)))
	user.Spirit = math.Max(0, user.Spirit-float64(damage))
	user.Cultivation = math.Max(0, user.Cultivation-float64(damage))
	db.DB.Model(user).Updates(map[string]interface{}{
		"spirit":      user.Spirit,
		"cultivation": user.Cultivation,
	})

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("遭受心魔侵扰，损失%d点灵力和修为", damage),
		Choices: []EventChoice{
			{Text: "继续探索", Value: "continue"},
		},
	}
}

// 新增事件处理函数
func (s *ExplorationService) eventHerbDiscovery(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 随机选择一种灵草
	if len(HerbConfigs) == 0 {
		return nil
	}
	herbIndex := r.Intn(len(HerbConfigs))
	herbConfig := HerbConfigs[herbIndex]

	// 随机品质
	qualityRand := r.Float64()
	quality := GetRandomQuality(qualityRand)
	value := GetHerbValue(herbConfig.BaseValue, quality)

	// 创建灵草记录
	herb := models.Herb{
		UserID: s.userID,
		HerbID: herbConfig.ID,
		Name:   herbConfig.Name,
		Count:  1,
	}
	db.DB.Create(&herb)

	return &ExplorationEvent{
		Type:   EventTypeHerbFound,
		Herb:   herb,
		Amount: value,
		Choices: []EventChoice{
			{Text: "收起灵草", Value: "collect"},
		},
	}
}

func (s *ExplorationService) eventPillRecipeFragment(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 随机选择一个丹方
	if len(PillRecipes) == 0 {
		return nil
	}
	recipeIndex := r.Intn(len(PillRecipes))
	recipe := PillRecipes[recipeIndex]

	// 创建或更新丹方残页记录
	var fragment models.PillFragment
	result := db.DB.Where("user_id = ? AND recipe_id = ?", s.userID, recipe.ID).First(&fragment)

	if result.Error == nil {
		// 已存在，增加数量
		fragment.Count++
		db.DB.Save(&fragment)
	} else {
		// 新建记录
		fragment = models.PillFragment{
			UserID:   s.userID,
			RecipeID: recipe.ID,
			Count:    1,
		}
		db.DB.Create(&fragment)
	}

	return &ExplorationEvent{
		Type:      EventTypePillRecipeFragment,
		RecipeID:  recipe.ID,
		Fragments: fragment.Count,
		Choices: []EventChoice{
			{Text: "收起残页", Value: "collect"},
		},
	}
}

// Event choice handlers
func (s *ExplorationService) handleItemFound(user *models.User, choice interface{}) (interface{}, error) {
	// 实现物品获取逻辑
	return map[string]interface{}{
		"type":    "item",
		"message": "获得了物品",
	}, nil
}

func (s *ExplorationService) handleSpiritStoneFound(user *models.User, choice interface{}) (interface{}, error) {
	// 灵石已在事件中处理
	return map[string]interface{}{
		"type":    "spirit_stone",
		"message": "获得了灵石",
	}, nil
}

func (s *ExplorationService) handleHerbFound(user *models.User, choice interface{}) (interface{}, error) {
	herbID := "spirit_grass" // 示例
	amount := 1

	// 检查是否已有该灵草
	var herb models.Herb
	result := db.DB.Where("user_id = ? AND herb_id = ?", s.userID, herbID).First(&herb)

	if result.Error == nil {
		// 已存在，更新数量
		herb.Count += amount
		db.DB.Save(&herb)
	} else {
		// 新增灵草
		herb = models.Herb{
			UserID: s.userID,
			HerbID: herbID,
			Name:   "灵精草",
			Count:  amount,
		}
		db.DB.Create(&herb)
	}

	return map[string]interface{}{
		"type":    "herb",
		"message": fmt.Sprintf("获得了%d个灵精草", amount),
	}, nil
}

func (s *ExplorationService) handlePillRecipeFragment(user *models.User, choice interface{}) (interface{}, error) {
	// 实现丹方残页获取逻辑
	return map[string]interface{}{
		"type":    "pill_fragment",
		"message": "获得了丹方残页",
	}, nil
}

func (s *ExplorationService) handleBattleEncounter(user *models.User, choice interface{}) (interface{}, error) {
	// 简单的胜负判定
	win := rand.Float64() > 0.5

	var message string
	reward := 0

	if win {
		reward = int(rand.Float64()*50 + 50)
		user.SpiritStones += reward
		db.DB.Model(user).Update("spirit_stones", user.SpiritStones)
		message = fmt.Sprintf("战胜了妖兽，获得了%d颗灵石", reward)
	} else {
		message = "败给了妖兽"
	}

	return map[string]interface{}{
		"type":    "battle",
		"message": message,
		"reward":  reward,
	}, nil
}

// calculateLuck 计算幸运值
func (s *ExplorationService) calculateLuck(user *models.User) float64 {
	if user.BaseAttributes == nil {
		return 1.0
	}

	var attrs map[string]interface{}
	if err := json.Unmarshal(user.BaseAttributes, &attrs); err != nil {
		return 1.0
	}

	luck, ok := attrs["luck"].(float64)
	if !ok {
		return 1.0
	}

	return luck
}

// getPlayerAttributes 获取玩家属性（从BaseAttributes JSON中）
func (s *ExplorationService) getPlayerAttributes(user *models.User) map[string]interface{} {
	attrs := make(map[string]interface{})
	if user.BaseAttributes != nil {
		if err := json.Unmarshal(user.BaseAttributes, &attrs); err != nil {
			return map[string]interface{}{"cultivationRate": 1.0, "spiritRate": 1.0}
		}
	}
	// 确保有默认值
	if _, ok := attrs["cultivationRate"]; !ok {
		attrs["cultivationRate"] = 1.0
	}
	if _, ok := attrs["spiritRate"]; !ok {
		attrs["spiritRate"] = 1.0
	}
	return attrs
}

// setPlayerAttributes 保存玩家属性到BaseAttributes JSON
func (s *ExplorationService) setPlayerAttributes(user *models.User, attrs map[string]interface{}) error {
	attrJSON, err := json.Marshal(attrs)
	if err != nil {
		return err
	}
	user.BaseAttributes = datatypes.JSON(attrJSON)
	return nil
}
