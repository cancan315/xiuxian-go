package duel

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// MonsterBaseAttributes 妖兽基础属性
type MonsterBaseAttributes struct {
	Attack  int `json:"attack"`  // 攻击力
	Health  int `json:"health"`  // 血量
	Defense int `json:"defense"` // 防御力
	Speed   int `json:"speed"`   // 速度
}

// MonsterCombatAttributes 妖兽战斗属性
type MonsterCombatAttributes struct {
	CritRate    float64 `json:"critRate"`    // 暴击率
	ComboRate   float64 `json:"comboRate"`   // 连击率
	CounterRate float64 `json:"counterRate"` // 反击率
	StunRate    float64 `json:"stunRate"`    // 眩晕率
	DodgeRate   float64 `json:"dodgeRate"`   // 闪避率
	VampireRate float64 `json:"vampireRate"` // 吸血率
}

// MonsterRewards 妖兽奖励
type MonsterRewards struct {
	DropItems string `json:"dropItems"` // 掉落物品描述
}

// Monster 妖兽配置
type Monster struct {
	ID               int            `json:"id"`               // 妖兽ID
	Name             string         `json:"name"`             // 妖兽名称
	Difficulty       string         `json:"difficulty"`       // 难度: normal, hard, boss
	Level            int            `json:"level"`            // 等级
	Description      string         `json:"description"`      // 妖兽描述
	BaseAttributes   datatypes.JSON `json:"baseAttributes"`   // 基础属性 (JSON)
	CombatAttributes datatypes.JSON `json:"combatAttributes"` // 战斗属性 (JSON)
	Rewards          datatypes.JSON `json:"rewards"`          // 奖励信息 (JSON)
}

// GetAllMonsters 获取所有妖兽配置
func GetAllMonsters() []Monster {
	return monsterConfigs
}

// GetMonsterByID 根据ID获取妖兽配置（支持普通妖兽和除魔卫道）
func GetMonsterByID(id int) *Monster {
	// 先在普通妖兽中查找
	for i := range monsterConfigs {
		if monsterConfigs[i].ID == id {
			return &monsterConfigs[i]
		}
	}
	// 如果没找到，在除魔卫道中查找
	for i := range demonSlayingConfigs {
		if demonSlayingConfigs[i].ID == id {
			return &demonSlayingConfigs[i]
		}
	}
	return nil
}

// 妖兽配置数据
var monsterConfigs = []Monster{
	{
		ID:          1,
		Name:        "赤焰虎",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "生活在火焰山脉的猛虎，浑身赤红如火，喜好吞吃灵精草",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          2,
		Name:        "青木狼",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "奔跑于青木林海的狼王，速度敏捷，守护着云雾花",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          3,
		Name:        "雷击树妖",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "经历雷击而不死，复而成妖。守护着雷击根",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          4,
		Name:        "黑水玄蛇",
		Difficulty:  "zhuji",
		Level:       2,
		Description: "潜伏在深潭中的巨蛇，毒性猛烈，洞穴常伴有龙息草",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":1500,\"health\":15000,\"defense\":1000,\"speed\":1500}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.15,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0.1,\"dodgeRate\":0,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          5,
		Name:        "裂风螳螂",
		Difficulty:  "zhuji",
		Level:       2,
		Description: "双臂如镰，快若疾风，刀气能撕裂护体灵光，是筑基修士极难应付的对手。守护着玄阴草",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":1500,\"health\":15000,\"defense\":1000,\"speed\":1500}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.15,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0.1,\"dodgeRate\":0,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          6,
		Name:        "寒晶毒蝎",
		Difficulty:  "zhuji",
		Level:       2,
		Description: "栖息于极寒毒沼，尾钩蕴含奇寒剧毒，甲壳如冰晶般坚固。守护着寒霜莲",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":1500,\"health\":15000,\"defense\":1000,\"speed\":1500}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.15,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0.1,\"dodgeRate\":0,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          7,
		Name:        "金翅大鹏",
		Difficulty:  "jindan",
		Level:       3,
		Description: "翔翔天际的神鸟，速度极快，族地常有九叶灵芝",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":3000,\"health\":30000,\"defense\":2000,\"speed\":3000}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.2,\"comboRate\":0.1,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.15,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          8,
		Name:        "八荒幻蝶",
		Difficulty:  "jindan",
		Level:       3,
		Description: "鳞粉能构造覆盖山林的庞大幻境，其本体脆弱但极难寻觅，考验修士的心性与洞察力。守护着紫金参",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":3000,\"health\":30000,\"defense\":2000,\"speed\":3000}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.2,\"comboRate\":0.1,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.15,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          9,
		Name:        "太阴玉蟾",
		Difficulty:  "jindan",
		Level:       3,
		Description: "栖息于至阴月华汇聚的寒潭，其鸣叫能引动心魔。腹中养殖着各种灵草",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":3000,\"health\":30000,\"defense\":2000,\"speed\":3000}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.2,\"comboRate\":0.1,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.15,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
}

// 除魔卫道配置数据（新增）
var demonSlayingConfigs = []Monster{
	{
		ID:          101,
		Name:        "合欢宗弟子",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "修炼合欢魔功的邪道弟子，擅长魅惑之术",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵石,修为,丹方残页\"}",
		)),
	},
	{
		ID:          102,
		Name:        "百炼宗叛徒",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "背叛百炼宗的叛徒，擅长练器之术，传闻因偷盗传宗仙器而背叛宗门",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵石,修为,装备\"}",
		)),
	},
	{
		ID:          103,
		Name:        "兽王宗叛徒",
		Difficulty:  "lianqi",
		Level:       1,
		Description: "背叛兽王宗的叛徒，擅长御兽之术，传闻因偷盗传宗灵宠袋而背叛宗门",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":15,\"health\":300,\"defense\":5,\"speed\":20}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.1,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.05,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵石,修为,灵宠\"}",
		)),
	},
	{
		ID:          104,
		Name:        "魔焰门弟子",
		Difficulty:  "zhuji",
		Level:       2,
		Description: "修炼魔焰之力的邪道弟子，攻击凶猛",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":500,\"health\":6000,\"defense\":300,\"speed\":500}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.15,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0.1,\"dodgeRate\":0,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵石,修为,丹方残页\"}",
		)),
	},
	{
		ID:          105,
		Name:        "鬼灵门弟子",
		Difficulty:  "jindan",
		Level:       3,
		Description: "修炼鬼道之法的邪道高手，诡异莫测",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":3000,\"health\":30000,\"defense\":2000,\"speed\":3000}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.2,\"comboRate\":0.1,\"counterRate\":0,\"stunRate\":0,\"dodgeRate\":0.15,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵石,修为,丹方残页\"}",
		)),
	},
}

// GetMonsterByIDAPI 根据ID获取单个妖兽详细信息
// 对应 GET /api/duel/monster/:id
func GetMonsterByIDAPI(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	monster := GetMonsterByID(id)
	if monster == nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "妖兽不存在",
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    monster,
	})
}

// GetMonsterChallenges 获取妖兽挑战列表（支持分页和难度过滤）
// 对应 GET /api/duel/monster-challenges
// 查询参数: page, pageSize, difficulty
func GetMonsterChallenges(c *gin.Context) {
	logger, ok := c.Get("zap_logger")
	var zapLogger *zap.Logger
	if !ok {
		// 如果没有获取到 logger，创建一个新的
		zapLogger, _ = zap.NewProduction()
	} else {
		zapLogger = logger.(*zap.Logger)
	}

	zapLogger.Info("GetMonsterChallenges 被调用", zap.String("path", c.Request.RequestURI))

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	difficulty := c.DefaultQuery("difficulty", "") // 空表示不过滤

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	zapLogger.Debug("GetMonsterChallenges 请求",
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
		zap.String("difficulty", difficulty))

	// 过滤妖兽列表
	var filteredMonsters []Monster
	for _, monster := range monsterConfigs {
		if difficulty == "" || monster.Difficulty == difficulty {
			filteredMonsters = append(filteredMonsters, monster)
		}
	}

	// 计算分页
	offset := (page - 1) * pageSize
	total := len(filteredMonsters)
	totalPages := (total + pageSize - 1) / pageSize

	// 获取当前页的妖兽
	var pageMonsters []Monster
	if offset < total {
		end := offset + pageSize
		if end > total {
			end = total
		}
		pageMonsters = filteredMonsters[offset:end]
	}

	zapLogger.Debug("GetMonsterChallenges 响应",
		zap.Int("total", total),
		zap.Int("returned", len(pageMonsters)))

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"monsters":   pageMonsters,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}

// GetDemonSlayingChallenges 获取除魔卫道挑战列表（支持分页和难度过滤）
// 对应 GET /api/duel/demon-slaying-challenges
// 查询参数: page, pageSize, difficulty
func GetDemonSlayingChallenges(c *gin.Context) {
	logger, ok := c.Get("zap_logger")
	var zapLogger *zap.Logger
	if !ok {
		// 如果没有获取到 logger，创建一个新的
		zapLogger, _ = zap.NewProduction()
	} else {
		zapLogger = logger.(*zap.Logger)
	}

	zapLogger.Info("GetDemonSlayingChallenges 被调用", zap.String("path", c.Request.RequestURI))

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	difficulty := c.DefaultQuery("difficulty", "") // 空表示不过滤

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	zapLogger.Debug("GetDemonSlayingChallenges 请求",
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
		zap.String("difficulty", difficulty))

	// 过滤除魔卫道列表
	var filteredMonsters []Monster
	for _, monster := range demonSlayingConfigs {
		if difficulty == "" || monster.Difficulty == difficulty {
			filteredMonsters = append(filteredMonsters, monster)
		}
	}

	// 计算分页
	offset := (page - 1) * pageSize
	total := len(filteredMonsters)
	totalPages := (total + pageSize - 1) / pageSize

	// 获取当前页的怪物
	var pageMonsters []Monster
	if offset < total {
		end := offset + pageSize
		if end > total {
			end = total
		}
		pageMonsters = filteredMonsters[offset:end]
	}

	zapLogger.Debug("GetDemonSlayingChallenges 响应",
		zap.Int("total", total),
		zap.Int("returned", len(pageMonsters)))

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"monsters":   pageMonsters,
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
	})
}
