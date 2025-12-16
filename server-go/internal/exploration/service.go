package exploration

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"xiuxian/server-go/internal/cultivation"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"gorm.io/datatypes"
)

// ExplorationService 探索服务
// 负责处理玩家探索相关的核心逻辑：
//   - 灵力检查
//   - 随机事件触发
//   - 事件结果结算
//   - 玩家属性更新
//
// 每个 ExplorationService 实例只服务于一个玩家（userID）
type ExplorationService struct {
	userID uint // 当前探索的玩家ID
}

// NewExplorationService 创建探索服务实例
// @param userID 玩家ID
func NewExplorationService(userID uint) *ExplorationService {
	return &ExplorationService{
		userID: userID,
	}
}

// CheckSpiritCost 检查灵力是否满足一次探索消耗
// 当前规则：每次探索固定消耗 100 点灵力
func (s *ExplorationService) CheckSpiritCost() (bool, error) {
	var user models.User
	// 查询玩家数据
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// 灵力不足时不可探索
	if user.Spirit < 100 {
		return false, nil
	}

	return true, nil
}

// cultivationRewardByRealm 根据境界计算修为奖励
func cultivationRewardByRealm(user *models.User, rate float64) int {
	realm := cultivation.GetRealmByLevel(user.Level)
	if realm == nil {
		return 0
	}
	return int(math.Floor(float64(realm.MaxCultivation) * rate))
}

// StartExploration 开始探索
// @param duration 探索持续时间（毫秒）
// @return 触发的事件列表、日志字符串、错误
func (s *ExplorationService) StartExploration(duration int) ([]ExplorationEvent, string, error) {
	// 获取玩家数据
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, "", fmt.Errorf("failed to get user: %w", err)
	}

	// 创建随机数生成器（避免使用全局 rand）
	r := rand.New(rand.NewSource(rand.Int63()))

	var events []ExplorationEvent // 触发的事件集合
	var logs []string             // 文本日志集合

	// 根据幸运值计算事件触发概率
	// 基础触发概率为 30%，幸运值作为倍率
	luck := s.calculateLuck(&user)
	eventChance := 0.3 * luck

	// 根据探索时间计算事件检查次数
	// 每 3000ms（约 3 秒）检查一次事件
	checkCount := int(math.Ceil(float64(duration) / 3000.0))
	for i := 0; i < checkCount; i++ {
		// 命中事件概率时触发随机事件
		if r.Float64() < eventChance {
			event := s.triggerRandomEvent(&user, r)
			if event != nil {
				events = append(events, *event)
				logs = append(logs, fmt.Sprintf("[事件]%s", event.Description))
			}
		}
	}

	// 若未触发任何事件，记录一条默认日志
	if len(logs) == 0 {
		logs = append(logs, "探索了一段时间，未发生特殊事件")
	}

	// 探索结算：消耗灵力、（可扩展基础奖励）
	user.Spirit -= 100 // 固定消耗 100 灵力

	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit": user.Spirit,
	}).Error; err != nil {
		return nil, "", fmt.Errorf("failed to update user: %w", err)
	}

	// 拼接日志文本
	logStr := ""
	for _, log := range logs {
		logStr += log + "\n"
	}

	return events, logStr, nil
}

// HandleEventChoice 处理探索事件的玩家选择
// eventType 用于区分事件类型
// choice 为前端传入的选择结果（当前多为占位）
func (s *ExplorationService) HandleEventChoice(eventType string, choice interface{}) (interface{}, error) {
	var user models.User
	if err := db.DB.First(&user, s.userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	switch eventType {
	case EventTypeItemFound:
		return s.handleItemFound(&user, choice)
	case EventTypeSpiritStoneFound:
		return s.handleSpiritStoneFound(&user, choice)
	case EventTypeHerbFound:
		return s.handleHerbFound(&user, choice)
	case EventTypePillRecipeFragment:
		return s.handlePillRecipeFragment(&user, choice)
	case EventTypeBattleEncounter:
		return s.handleBattleEncounter(&user, choice)
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

// triggerRandomEvent 根据“权重随机”机制触发一个事件
//
// 【改动说明】
// 1. 旧逻辑：顺序遍历 + 单独概率判断（前面的事件更容易被命中）
// 2. 新逻辑：所有事件概率视为“权重”，先求权重总和，再进行一次随机
// 3. 优点：
//   - 概率语义更清晰（真实权重）
//   - 方便后续数值策划调整
//   - 避免事件顺序导致的隐性偏差
func (s *ExplorationService) triggerRandomEvent(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 事件配置（chance 作为权重使用）
	events := []struct {
		name    string
		weight  float64
		handler func(*models.User, *rand.Rand) *ExplorationEvent
	}{
		{"古老石碑", 8, s.eventAncientTablet},
		{"灵泉", 12, s.eventSpiritSpring},
		{"古修遗府", 3, s.eventAncientMaster},
		{"妖兽袭击", 15, s.eventMonsterAttack},
		{"走火入魔", 10, s.eventCultivationDeviation},
		{"秘境宝藏", 5, s.eventTreasureTrove},
		{"顿悟", 7, s.eventEnlightenment},
		{"心魔侵扰", 12, s.eventQiDeviation},
		{"灵草发现", 18, s.eventHerbDiscovery},
		{"丹方残页", 10, s.eventPillRecipeFragment},
	}

	// 计算总权重
	totalWeight := 0.0
	for _, e := range events {
		totalWeight += e.weight
	}

	// 在 [0, totalWeight) 区间内取随机数
	rnd := r.Float64() * totalWeight

	// 命中区间判断
	acc := 0.0
	for _, e := range events {
		acc += e.weight
		if rnd <= acc {
			return e.handler(user, r)
		}
	}

	return nil
}

// ======================= 具体事件处理函数 =======================

// eventAncientTablet 古老石碑：增加修为
func (s *ExplorationService) eventAncientTablet(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 普通事件：1%最大修为值MaxCultivation
	bonus := cultivationRewardByRealm(user, 0.01)
	user.Cultivation = math.Round((user.Cultivation+float64(bonus))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        "cultivation_boost",
		Description: fmt.Sprintf("[参与散修论道大会]交流修炼心得，领悟道法，修为增加%d点", bonus),
		Amount:      bonus,
	}
}

// eventSpiritSpring 魔道杂役弟子修为和灵力提升
func (s *ExplorationService) eventSpiritSpring(user *models.User, r *rand.Rand) *ExplorationEvent {
	cultivationBonus := cultivationRewardByRealm(user, 0.05)
	spiritBonus := int(math.Floor(10 * (float64(user.Level)/2 + 1)))
	user.Cultivation = math.Round((user.Cultivation+float64(cultivationBonus))*10) / 10
	user.Spirit += float64(spiritBonus)

	db.DB.Model(user).Updates(map[string]interface{}{
		"cultivation": user.Cultivation,
		"spirit":      user.Spirit,
	})

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("[出门游历]偶遇百鬼门杂役弟子，果断出手，大战300回合，就地斩杀。修为+%d，灵力+%d", cultivationBonus, spiritBonus),
	}
}

// eventAncientMaster 古修遗府：修为 + 灵力 大幅提升
func (s *ExplorationService) eventAncientMaster(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 稀有事件：5%最大修为值MaxCultivation
	cultivationBonus := cultivationRewardByRealm(user, 0.05)
	spiritBonus := int(math.Floor(18 * (float64(user.Level)/2 + 1)))

	user.Cultivation = math.Round((user.Cultivation+float64(cultivationBonus))*10) / 10
	user.Spirit += float64(spiritBonus)

	db.DB.Model(user).Updates(map[string]interface{}{
		"cultivation": user.Cultivation,
		"spirit":      user.Spirit,
	})

	return &ExplorationEvent{
		Type:        EventTypeItemFound,
		Description: fmt.Sprintf("[探索古修遗府]经过一番搜刮，寻到上古修士炼制的废丹，服用后，修为+%d，灵力+%d", cultivationBonus, spiritBonus),
	}
}

// eventMonsterAttack 妖兽袭击：损失修为当前境界的5%
func (s *ExplorationService) eventMonsterAttack(user *models.User, r *rand.Rand) *ExplorationEvent {
	damage := cultivationRewardByRealm(user, 0.05)
	user.Cultivation = math.Round(math.Max(0, user.Cultivation-float64(damage))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        EventTypeBattleEncounter,
		Description: fmt.Sprintf("[偶遇妖兽]探索妖兽山脉，惊扰沉睡的鬼鬼大王，被妖王生吞，损失%d修为", damage),
		Amount:      damage,
	}
}

// eventCultivationDeviation 走火入魔：修为受损当前境界的20%
func (s *ExplorationService) eventCultivationDeviation(user *models.User, r *rand.Rand) *ExplorationEvent {
	damage := cultivationRewardByRealm(user, 0.2)
	user.Cultivation = math.Round(math.Max(0, user.Cultivation-float64(damage))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        "cultivation_damage",
		Description: fmt.Sprintf("[修仙家族招婿]参与招婿，被家主幼女一剑扫飞，擂台惨败，修为损失%d点", damage),
		Amount:      damage,
	}
}

// eventTreasureTrove 秘境宝藏：获得灵石
func (s *ExplorationService) eventTreasureTrove(user *models.User, r *rand.Rand) *ExplorationEvent {
	stoneBonus := 1
	user.SpiritStones += stoneBonus
	db.DB.Model(user).Update("spirit_stones", user.SpiritStones)

	return &ExplorationEvent{
		Type:        EventTypeSpiritStoneFound,
		Description: fmt.Sprintf("[深入废矿]你到废弃矿洞挖掘三天三夜，挖到精铁伴生碎石，到坊市出售后，获得%d颗灵石", stoneBonus),
		Amount:      stoneBonus,
	}
}

// eventEnlightenment 顿悟：增加修为
func (s *ExplorationService) eventEnlightenment(user *models.User, r *rand.Rand) *ExplorationEvent {
	bonus := cultivationRewardByRealm(user, 0.2)
	user.Cultivation = math.Round((user.Cultivation+float64(bonus))*10) / 10
	db.DB.Model(user).Update("cultivation", user.Cultivation)

	return &ExplorationEvent{
		Type:        "cultivation_boost",
		Description: fmt.Sprintf("[灵山游历]寻觅到天然阵法，枯坐百日，顿悟自然道法，修为增加%d点", bonus),
		Amount:      bonus,
	}
}

// eventQiDeviation 合欢女修：修为和灵力双重受损
func (s *ExplorationService) eventQiDeviation(user *models.User, r *rand.Rand) *ExplorationEvent {
	// 修为损失30%
	cultivationDamage := cultivationRewardByRealm(user, 0.3)
	user.Cultivation = math.Max(0, user.Cultivation-float64(cultivationDamage))
	// 灵力损失50%
	spiritDamage := int(user.Spirit / 2)
	user.Spirit = math.Max(0, user.Spirit-float64(spiritDamage))
	db.DB.Model(user).Updates(map[string]interface{}{
		"spirit":      user.Spirit,
		"cultivation": user.Cultivation,
	})

	return &ExplorationEvent{
		Type:        "spirit_and_cultivation_damage",
		Description: fmt.Sprintf("[偶遇女修]被合欢外门女修撞见，经过奋力挣扎，事后，被吸取修为%d灵力各损失%d点", cultivationDamage, spiritDamage),
		Amount:      cultivationDamage,
	}
}

// eventHerbDiscovery 灵草发现：获得灵草
func (s *ExplorationService) eventHerbDiscovery(user *models.User, r *rand.Rand) *ExplorationEvent {
	if len(HerbConfigs) == 0 {
		return nil
	}
	herbIndex := r.Intn(len(HerbConfigs))
	herbConfig := HerbConfigs[herbIndex]

	// 创建灵草记录
	herb := models.Herb{
		UserID: s.userID,
		HerbID: herbConfig.ID,
		Name:   herbConfig.Name,
		Count:  1,
	}
	if err := db.DB.Create(&herb).Error; err != nil {
		return nil
	}

	return &ExplorationEvent{
		Type:        EventTypeHerbFound,
		Description: fmt.Sprintf("[灵草发现]获得%s", herbConfig.Name),
		Amount:      int(herbConfig.BaseValue),
	}
}

// eventPillRecipeFragment 丹方残页：获得丹方残页
func (s *ExplorationService) eventPillRecipeFragment(user *models.User, r *rand.Rand) *ExplorationEvent {
	if len(PillRecipes) == 0 {
		return nil
	}
	recipeIndex := r.Intn(len(PillRecipes))
	recipe := PillRecipes[recipeIndex]

	// 创建或更新丹方残页记录
	var fragment models.PillFragment
	result := db.DB.Where("user_id = ? AND recipe_id = ?", s.userID, recipe.ID).First(&fragment)

	if result.Error == nil {
		fragment.Count++
		db.DB.Save(&fragment)
	} else {
		fragment = models.PillFragment{
			UserID:   s.userID,
			RecipeID: recipe.ID,
			Count:    1,
		}
		db.DB.Create(&fragment)
	}

	return &ExplorationEvent{
		Type:        EventTypePillRecipeFragment,
		Description: fmt.Sprintf("[丹方残页]获得%s的残页", recipe.Name),
		Fragments:   fragment.Count,
	}
}

// ======================= 事件选择处理函数 =======================

// handleItemFound 处理获得物品事件
func (s *ExplorationService) handleItemFound(user *models.User, choice interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":    "item",
		"message": "已确认获得物品",
		"status":  "success",
	}, nil
}

// handleSpiritStoneFound 处理获得灵石事件
func (s *ExplorationService) handleSpiritStoneFound(user *models.User, choice interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":    "spirit_stone",
		"message": "已确认获得灵石",
		"status":  "success",
	}, nil
}

// handleHerbFound 处理获得灵草事件
func (s *ExplorationService) handleHerbFound(user *models.User, choice interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":    "herb",
		"message": "已确认收起灵草",
		"status":  "success",
	}, nil
}

// handlePillRecipeFragment 处理丹方残页事件
func (s *ExplorationService) handlePillRecipeFragment(user *models.User, choice interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":    "pill_fragment",
		"message": "已确认收起残页",
		"status":  "success",
	}, nil
}

// handleBattleEncounter 处理战斗遭遇事件
func (s *ExplorationService) handleBattleEncounter(user *models.User, choice interface{}) (interface{}, error) {
	return map[string]interface{}{
		"type":    "battle",
		"message": "已确认战斗结果",
		"status":  "success",
	}, nil
}

// calculateLuck 从 BaseAttributes 中计算幸运值
// 默认幸运值为 1.0
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

// getPlayerAttributes 从 BaseAttributes 中读取玩家成长属性
func (s *ExplorationService) getPlayerAttributes(user *models.User) map[string]interface{} {
	attrs := make(map[string]interface{})
	if user.BaseAttributes != nil {
		_ = json.Unmarshal(user.BaseAttributes, &attrs)
	}

	// 设置默认值
	if _, ok := attrs["cultivationRate"]; !ok {
		attrs["cultivationRate"] = 1.0
	}
	if _, ok := attrs["spiritRate"]; !ok {
		attrs["spiritRate"] = 1.0
	}
	return attrs
}

// setPlayerAttributes 将属性写回 BaseAttributes（JSON）
func (s *ExplorationService) setPlayerAttributes(user *models.User, attrs map[string]interface{}) error {
	attrJSON, err := json.Marshal(attrs)
	if err != nil {
		return err
	}
	user.BaseAttributes = datatypes.JSON(attrJSON)
	return nil
}
