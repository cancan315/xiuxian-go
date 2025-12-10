package gacha

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
)

// currentUserID 与 player handler 中逻辑保持一致
func currentUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// ---------- 概率与配置，与 Node gachaController.js 保持一致 ----------

var equipmentQualityProbabilities = map[string]float64{
	"mythic":    0.001,
	"legendary": 0.003,
	"epic":      0.016,
	"rare":      0.03,
	"uncommon":  0.15,
	"common":    0.80,
}

var petRarityProbabilities = map[string]float64{
	"mythic":    0.001,
	"legendary": 0.003,
	"epic":      0.016,
	"rare":      0.03,
	"uncommon":  0.15,
	"common":    0.80,
}

var attributeTypes = struct {
	BaseAttributes    []string
	CombatAttributes  []string
	CombatResistance  []string
	SpecialAttributes []string
}{
	BaseAttributes:   []string{"attack", "health", "defense", "speed"},
	CombatAttributes: []string{"critRate", "comboRate", "counterRate", "stunRate", "dodgeRate", "vampireRate"},
	CombatResistance: []string{"critResist", "comboResist", "counterResist", "stunResist", "dodgeResist", "vampireResist"},
	SpecialAttributes: []string{
		"healBoost", "critDamageBoost", "critDamageReduce",
		"finalDamageBoost", "finalDamageReduce",
		"combatBoost", "resistanceBoost",
	},
}

var qualityAttributeRules = map[string]map[string]struct {
	Base       int
	Combat     int
	Resistance int
	Special    int
}{
	"equipment": {
		"common":    {Base: 1},
		"uncommon":  {Base: 2},
		"rare":      {Base: 3},
		"epic":      {Base: 4, Combat: 1},
		"legendary": {Base: 4, Combat: 6, Resistance: 6},
		"mythic":    {Base: 4, Combat: 6, Resistance: 6, Special: 7},
	},
	"pet": {
		"common":    {Base: 1},
		"uncommon":  {Base: 2},
		"rare":      {Base: 3},
		"epic":      {Base: 4, Combat: 1},
		"legendary": {Base: 4, Combat: 6, Resistance: 6},
		"mythic":    {Base: 4, Combat: 6, Resistance: 6, Special: 7},
	},
}

// 简化：只实现与随机生成直接相关的名称/描述表，细节与 Node 相同意义即可
var petNamesByRarity = map[string][]string{
	"common":    {"白狐", "灰狼", "黄牛", "黑虎", "赤马", "棕熊", "青蛇", "紫貂", "银鼠", "金蝉", "彩雀", "田园犬"},
	"uncommon":  {"妖猴", "妖驴", "妖豹", "妖蛇", "妖虎", "妖熊", "妖鹰", "妖蝎", "妖蛛", "妖蝠", "妖蟾", "妖蜈"},
	"rare":      {"灵狐", "灵鹿", "灵龟", "灵蛇", "灵猿", "灵鹤", "灵鲤", "灵雀", "灵虎", "灵豹", "灵猫", "灵犬"},
	"epic":      {"饕餮", "穷奇", "梼杌", "混沌", "九婴", "相柳", "凿齿", "修蛇", "封豨", "大风", "巴蛇", "朱厌"},
	"legendary": {"麒麟", "凤凰", "龙龟", "白泽", "重明鸟", "当康", "乘黄", "英招", "夫诸", "天马", "青牛", "玄龟"},
	"mythic":    {"应龙", "夔牛", "毕方", "饕餮", "九尾狐", "玉兔", "金蟾", "青鸾", "火凤", "水麒麟", "土蝼", "陆吾"},
}

var petDescriptionsByRarity = map[string][]string{
	"common":    {"一只普通的小动物，刚刚开启灵智"},
	"uncommon":  {"已具初步妖力的妖兽，有一定培养价值"},
	"rare":      {"天生蕴含灵气的珍稀兽类，颇具潜力"},
	"epic":      {"来自上古时代的神秘异兽，力量强大"},
	"legendary": {"祥瑞降临，世间少有的瑞兽"},
	"mythic":    {"超越凡俗的仙界仙兽，几近传说"},
}

// 装备类型配置（名称/slot 与 Node 对齐）
var equipmentTypes = map[string]struct {
	Name     string
	Slot     string
	Prefixes map[string][]string
}{
	"faqi": {
		Name: "法宝",
		Slot: "faqi",
		Prefixes: map[string][]string{
			"common":    {"粗制", "劣质", "破损", "锈蚀"},
			"uncommon":  {"精制", "优质", "改良", "强化"},
			"rare":      {"上品", "精品", "珍品", "灵韵"},
			"epic":      {"异宝", "奇珍", "灵光", "宝华"},
			"legendary": {"法宝", "神兵", "灵宝", "仙锋"},
			"mythic":    {"仙器", "神器", "天器", "圣器"},
		},
	},
	"guanjin": {
		Name: "冠巾",
		Slot: "guanjin",
		Prefixes: map[string][]string{
			"common":    {"布制", "麻织", "粗布", "素巾"},
			"uncommon":  {"丝织", "锦制", "绣冠", "轻纱"},
			"rare":      {"灵丝", "云锦", "霓裳", "羽冠"},
			"epic":      {"宝冠", "灵冠", "星辰", "月华"},
			"legendary": {"仙冠", "神冠", "紫金", "龙纹"},
			"mythic":    {"天冠", "圣冠", "混沌", "太虚"},
		},
	},
	"daopao": {
		Name: "道袍",
		Slot: "daopao",
		Prefixes: map[string][]string{
			"common":    {"粗布", "麻衣", "布衣", "素袍"},
			"uncommon":  {"丝绸", "锦袍", "绣衣", "轻衫"},
			"rare":      {"灵绸", "云裳", "霞衣", "霓裳"},
			"epic":      {"宝衣", "灵衣", "星辰", "月华"},
			"legendary": {"仙衣", "神袍", "紫绶", "龙袍"},
			"mythic":    {"天衣", "圣袍", "混沌", "太虚"},
		},
	},
	"yunlv": {
		Name: "云履",
		Slot: "yunlv",
		Prefixes: map[string][]string{
			"common":    {"布鞋", "草鞋", "麻鞋", "木屐"},
			"uncommon":  {"皮靴", "丝履", "锦履", "绣鞋"},
			"rare":      {"灵靴", "云履", "霞履", "霓履"},
			"epic":      {"宝履", "灵履", "星辰", "月华"},
			"legendary": {"仙履", "神履", "紫霞", "龙履"},
			"mythic":    {"天履", "圣履", "混沌", "太虚"},
		},
	},
	"fabao": {
		Name: "本命法宝",
		Slot: "fabao",
		Prefixes: map[string][]string{
			"common":    {"粗制", "劣质", "仿制", "赝品"},
			"uncommon":  {"精制", "良品", "上品", "优质"},
			"rare":      {"灵宝", "珍宝", "异宝", "奇宝"},
			"epic":      {"重宝", "至宝", "灵光", "宝华"},
			"legendary": {"仙宝", "神宝", "天宝", "圣宝"},
			"mythic":    {"至尊", "无上", "混沌", "太虚"},
		},
	},
}

// ---------- 工具函数 ----------

func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON("null")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON("null")
	}
	return datatypes.JSON(b)
}

func weightedRandom(weights []float64) int {
	var total float64
	for _, w := range weights {
		total += w
	}
	if total <= 0 {
		return 0
	}
	r := rand.Float64() * total
	for i, w := range weights {
		if r < w {
			return i
		}
		r -= w
	}
	return len(weights) - 1
}

func shuffleStrings(a []string) {
	rand.Seed(time.Now().UnixNano())
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func selectRandomAttributes(list []string, count int) []string {
	if count <= 0 || len(list) == 0 {
		return nil
	}
	if count >= len(list) {
		return append([]string{}, list...)
	}
	copyList := append([]string{}, list...)
	shuffleStrings(copyList)
	return copyList[:count]
}

func generateAttributeValue(attrType, quality string, isPet bool) float64 {
	baseRanges := map[string]struct{ Min, Max float64 }{
		"attack":  {30, 100},
		"health":  {60, 200},
		"defense": {15, 50},
		"speed":   {5, 20},
	}
	combatMin, combatMax := 0.01, 0.10
	resistMin, resistMax := 0.01, 0.08
	specialMin, specialMax := 0.005, 0.05

	if contains(attributeTypes.BaseAttributes, attrType) {
		rangeVal, ok := baseRanges[attrType]
		if !ok {
			rangeVal = struct{ Min, Max float64 }{10, 50}
		}
		qualityMods := map[string]float64{
			"common":    1.0,
			"uncommon":  1.2,
			"rare":      1.5,
			"epic":      2.0,
			"legendary": 2.5,
			"mythic":    3.0,
		}
		mod := qualityMods[quality]
		if mod == 0 {
			mod = 1.0
		}
		return float64(int((rand.Float64()*(rangeVal.Max-rangeVal.Min)+rangeVal.Min)*mod + 0.5))
	}
	if contains(attributeTypes.CombatAttributes, attrType) {
		v := rand.Float64()*(combatMax-combatMin) + combatMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if contains(attributeTypes.CombatResistance, attrType) {
		v := rand.Float64()*(resistMax-resistMin) + resistMin
		return float64(int(v*1000+0.5)) / 1000
	}
	if contains(attributeTypes.SpecialAttributes, attrType) {
		v := rand.Float64()*(specialMax-specialMin) + specialMin
		return float64(int(v*1000+0.5)) / 1000
	}
	return float64(int((rand.Float64()*40 + 10) + 0.5))
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func generateQualityBasedAttributes(kind, quality string) map[string]float64 {
	rulesPerKind, ok := qualityAttributeRules[kind]
	if !ok {
		return map[string]float64{}
	}
	rules, ok := rulesPerKind[quality]
	if !ok {
		return map[string]float64{}
	}
	attrs := make(map[string]float64)

	if rules.Base > 0 {
		for _, attr := range selectRandomAttributes(attributeTypes.BaseAttributes, rules.Base) {
			attrs[attr] = generateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Combat > 0 {
		for _, attr := range selectRandomAttributes(attributeTypes.CombatAttributes, rules.Combat) {
			attrs[attr] = generateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Resistance > 0 {
		for _, attr := range selectRandomAttributes(attributeTypes.CombatResistance, rules.Resistance) {
			attrs[attr] = generateAttributeValue(attr, quality, kind == "pet")
		}
	}
	if rules.Special > 0 {
		for _, attr := range selectRandomAttributes(attributeTypes.SpecialAttributes, rules.Special) {
			attrs[attr] = generateAttributeValue(attr, quality, kind == "pet")
		}
	}
	return attrs
}

// ---------- 随机生成逻辑 ----------

type GachaEquipment struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Quality       string                 `json:"quality"`
	EquipType     string                 `json:"equipType"`
	Level         int                    `json:"level"`
	RequiredRealm int                    `json:"requiredRealm"`
	EnhanceLevel  int                    `json:"enhanceLevel"`
	Stats         map[string]float64     `json:"stats"`
	ExtraAttrs    map[string]interface{} `json:"extraAttributes"`
}

type GachaPet struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Type            string             `json:"type"`
	Rarity          string             `json:"rarity"`
	Level           int                `json:"level"`
	Star            int                `json:"star"`
	Exp             int                `json:"exp"`
	Description     string             `json:"description"`
	CombatAttrs     map[string]float64 `json:"combatAttributes"`
	AttackBonus     float64            `json:"attackBonus"`
	DefenseBonus    float64            `json:"defenseBonus"`
	HealthBonus     float64            `json:"healthBonus"`
	CreatedAtMillis int64              `json:"createdAt"`
}

func generateRandomEquipment(level int) GachaEquipment {
	// i) 选择品质
	qualities := []string{"mythic", "legendary", "epic", "rare", "uncommon", "common"}
	weights := make([]float64, len(qualities))
	for i, q := range qualities {
		weights[i] = equipmentQualityProbabilities[q] * 100
	}
	qIdx := weightedRandom(weights)
	quality := qualities[qIdx]

	// ii) 装备类型
	keys := make([]string, 0, len(equipmentTypes))
	for k := range equipmentTypes {
		keys = append(keys, k)
	}
	etypeKey := keys[rand.Intn(len(keys))]
	et := equipmentTypes[etypeKey]

	prefixes := et.Prefixes[quality]
	prefix := prefixes[rand.Intn(len(prefixes))]
	name := prefix + et.Name

	stats := generateQualityBasedAttributes("equipment", quality)

	numericLevel := level
	if numericLevel <= 0 {
		numericLevel = 1
	}

	// realm 计算逻辑按 Node 注释映射
	requiredRealm := 1
	switch {
	case numericLevel >= 1 && numericLevel <= 9:
		requiredRealm = 1
	case numericLevel >= 10 && numericLevel <= 18:
		requiredRealm = 2
	case numericLevel >= 19 && numericLevel <= 27:
		requiredRealm = 3
	case numericLevel >= 28 && numericLevel <= 36:
		requiredRealm = 4
	case numericLevel >= 37 && numericLevel <= 45:
		requiredRealm = 5
	case numericLevel >= 46 && numericLevel <= 54:
		requiredRealm = 6
	case numericLevel >= 55 && numericLevel <= 63:
		requiredRealm = 7
	case numericLevel >= 64 && numericLevel <= 72:
		requiredRealm = 8
	case numericLevel >= 73 && numericLevel <= 81:
		requiredRealm = 9
	case numericLevel >= 82 && numericLevel <= 90:
		requiredRealm = 10
	case numericLevel >= 91 && numericLevel <= 99:
		requiredRealm = 11
	case numericLevel >= 100 && numericLevel <= 108:
		requiredRealm = 12
	case numericLevel >= 109 && numericLevel <= 117:
		requiredRealm = 13
	case numericLevel >= 118 && numericLevel <= 126:
		requiredRealm = 14
	case numericLevel >= 127:
		requiredRealm = 15
	}

	return GachaEquipment{
		ID:            uuid.NewString(),
		Name:          name,
		Type:          "equipment",
		Quality:       quality,
		EquipType:     etypeKey,
		Level:         numericLevel,
		RequiredRealm: requiredRealm,
		EnhanceLevel:  0,
		Stats:         stats,
		ExtraAttrs:    map[string]interface{}{},
	}
}

func generateRandomPet(level int) GachaPet {
	rarities := []string{"mythic", "legendary", "epic", "rare", "uncommon", "common"}
	weights := make([]float64, len(rarities))
	for i, r := range rarities {
		weights[i] = petRarityProbabilities[r] * 1000
	}
	rIdx := weightedRandom(weights)
	rarity := rarities[rIdx]

	names := petNamesByRarity[rarity]
	name := names[rand.Intn(len(names))]
	descs := petDescriptionsByRarity[rarity]
	desc := descs[rand.Intn(len(descs))]

	combatAttrs := generateQualityBasedAttributes("pet", rarity)

	qualityBonusMap := map[string]float64{
		"mythic":    0.15,
		"legendary": 0.12,
		"epic":      0.09,
		"rare":      0.06,
		"uncommon":  0.03,
		"common":    0.03,
	}
	starBonusPerQuality := map[string]float64{
		"mythic":    0.02,
		"legendary": 0.01,
		"epic":      0.01,
		"rare":      0.01,
		"uncommon":  0.01,
		"common":    0.01,
	}

	baseBonus := qualityBonusMap[rarity]
	starBonus := 0 * starBonusPerQuality[rarity]
	levelBonus := 0 * (baseBonus * 0.1)
	phase := 0
	phaseBonus := float64(phase) * (baseBonus * 0.5)
	finalBonus := baseBonus + starBonus + levelBonus + phaseBonus

	return GachaPet{
		ID:              uuid.NewString(),
		Name:            name,
		Type:            "pet",
		Rarity:          rarity,
		Level:           1,
		Star:            0,
		Exp:             0,
		Description:     desc,
		CombatAttrs:     combatAttrs,
		AttackBonus:     finalBonus,
		DefenseBonus:    finalBonus,
		HealthBonus:     finalBonus,
		CreatedAtMillis: time.Now().UnixMilli(),
	}
}

// ---------- Handlers ----------

// DrawGacha 对应 POST /api/gacha/draw
func DrawGacha(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var req struct {
		PoolType    string `json:"poolType"`
		Count       int    `json:"count"`
		UseWishlist bool   `json:"useWishlist"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}
	if req.Count <= 0 {
		req.Count = 1
	}

	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	cost := map[int]int{}
	if req.UseWishlist {
		cost = map[int]int{1: 200, 10: 2000, 50: 10000, 100: 20000}
	} else {
		cost = map[int]int{1: 100, 10: 1000, 50: 5000, 100: 10000}
	}
	required := cost[req.Count]
	if required == 0 {
		if req.UseWishlist {
			required = 200 * req.Count
		} else {
			required = 100 * req.Count
		}
	}

	if user.SpiritStones < required {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "灵石不足"})
		return
	}

	items := make([]interface{}, 0, req.Count)

	tx := db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": tx.Error.Error()})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := 0; i < req.Count; i++ {
		if req.PoolType == "equipment" {
			eq := generateRandomEquipment(user.Level)
			items = append(items, eq)

			model := models.Equipment{
				ID:            uuid.NewString(),
				UserID:        userID,
				EquipmentID:   eq.ID,
				Name:          eq.Name,
				Type:          eq.Type,
				Quality:       eq.Quality,
				EnhanceLevel:  eq.EnhanceLevel,
				RequiredRealm: eq.RequiredRealm,
				Level:         eq.Level,
			}
			// 只保存 equipType，不映射到 slot
			et := eq.EquipType
			model.EquipType = &et
			model.Stats = toJSON(eq.Stats)
			model.ExtraAttributes = toJSON(eq.ExtraAttrs)
			model.Equipped = false

			if err := tx.Create(&model).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
		} else if req.PoolType == "pet" {
			p := generateRandomPet(user.Level)
			items = append(items, p)

			petModel := models.Pet{
				ID:     uuid.NewString(),
				UserID: userID,
				PetID:  p.ID,
				Name:   p.Name,
				Type:   p.Type,
				Rarity: p.Rarity,
				Level:  p.Level,
				Star:   p.Star,
			}
			petModel.CombatAttributes = toJSON(p.CombatAttrs)
			petModel.AttackBonus = p.AttackBonus
			petModel.DefenseBonus = p.DefenseBonus
			petModel.HealthBonus = p.HealthBonus
			petModel.IsActive = false

			if err := tx.Create(&petModel).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
		}
	}

	// 扣除灵石
	if err := tx.Model(&models.User{}).Where("id = ?", userID).
		Update("spiritStones", gorm.Expr("\"spiritStones\" - ?", required)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	newStones := user.SpiritStones - required
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"items":        items,
		"message":      "抽奖成功",
		"spiritStones": newStones,
	})
}

// ProcessAutoActions 对应 POST /api/gacha/auto-actions
func ProcessAutoActions(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	var req struct {
		Items               []map[string]interface{} `json:"items"`
		AutoSellQualities   []string                 `json:"autoSellQualities"`
		AutoReleaseRarities []string                 `json:"autoReleaseRarities"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		return
	}

	qualityStoneValues := map[string]int{
		"common":    1,
		"uncommon":  2,
		"rare":      5,
		"epic":      10,
		"legendary": 20,
		"mythic":    50,
	}

	soldItems := make([]map[string]interface{}, 0)
	releasedPets := make([]map[string]interface{}, 0)
	stonesGained := 0

	containsStr := func(list []string, v string) bool {
		for _, s := range list {
			if s == v {
				return true
			}
		}
		return false
	}

	for _, item := range req.Items {
		typeStr, _ := item["type"].(string)
		quality, _ := item["quality"].(string)
		rarity, _ := item["rarity"].(string)
		idAny := item["id"]
		idStr, _ := idAny.(string)

		if typeStr == "equipment" && containsStr(req.AutoSellQualities, quality) {
			// 删除装备
			if err := db.DB.Where("userId = ? AND equipmentId = ?", userID, idStr).
				Delete(&models.Equipment{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
			stonesGained += qualityStoneValues[quality]
			if qualityStoneValues[quality] == 0 {
				stonesGained++
			}
			soldItems = append(soldItems, item)
		} else if typeStr == "pet" && containsStr(req.AutoReleaseRarities, rarity) {
			// 删除灵宠
			if err := db.DB.Where("userId = ? AND petId = ?", userID, idStr).
				Delete(&models.Pet{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
			stonesGained += qualityStoneValues[rarity]
			if qualityStoneValues[rarity] == 0 {
				stonesGained++
			}
			releasedPets = append(releasedPets, item)
		}
	}

	if stonesGained > 0 {
		if err := db.DB.Model(&models.User{}).
			Where("id = ?", userID).
			Update("reinforceStones", gorm.Expr("\"reinforceStones\" + ?", stonesGained)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"soldItems":    soldItems,
		"releasedPets": releasedPets,
		"stonesGained": stonesGained,
		"message":      "自动处理完成",
	})
}
