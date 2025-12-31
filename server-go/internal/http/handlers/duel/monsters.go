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

// GetMonsterByID 根据ID获取妖兽配置
func GetMonsterByID(id int) *Monster {
	for i := range monsterConfigs {
		if monsterConfigs[i].ID == id {
			return &monsterConfigs[i]
		}
	}
	return nil
}

// 妖兽配置数据
var monsterConfigs = []Monster{
	{
		ID:          1,
		Name:        "赤焰虎",
		Difficulty:  "normal",
		Level:       1,
		Description: "生活在火焰山脉的猛虎，浑身赤红如火",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":90,\"health\":900,\"defense\":45,\"speed\":90}",
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
		Name:        "黑水玄蛇",
		Difficulty:  "hard",
		Level:       2,
		Description: "潜伏在深潭中的巨蛇，毒性猛烈",
		BaseAttributes: datatypes.JSON([]byte(
			"{\"attack\":180,\"health\":1800,\"defense\":90,\"speed\":180}",
		)),
		CombatAttributes: datatypes.JSON([]byte(
			"{\"critRate\":0.15,\"comboRate\":0,\"counterRate\":0,\"stunRate\":0.1,\"dodgeRate\":0,\"vampireRate\":0}",
		)),
		Rewards: datatypes.JSON([]byte(
			"{\"dropItems\":\"灵草\"}",
		)),
	},
	{
		ID:          3,
		Name:        "金翅大鹏",
		Difficulty:  "boss",
		Level:       3,
		Description: "翱翔天际的神鸟，速度极快",
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
