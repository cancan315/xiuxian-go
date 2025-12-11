package gacha

import (
	"net/http"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	
	"gorm.io/gorm"

	"xiuxian/server-go/internal/db"
)

// currentUserID 与 player handler 中逻辑保持一致
func CurrentUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// ---------- 处理函数 ----------

// DrawGacha 对应 POST /api/gacha/draw
func DrawGacha(c *gin.Context) {
	userID, ok := CurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)
	
	// 记录认证信息用于调试
	authHeader := c.GetHeader("Authorization")
	zapLogger.Info("抽奖请求",
		zap.Uint("用户ID", userID),
		zap.String("认证头", authHeader))

	var req struct {
		PoolType    string `json:"poolType"`
		Count       int    `json:"count"`
		UseWishlist bool   `json:"useWishlist"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		zapLogger.Error("抽奖请求绑定错误",
			zap.Error(err))
		return
	}
	
	zapLogger.Info("抽奖请求参数",
		zap.String("抽奖池类型", req.PoolType),
		zap.Int("数量", req.Count),
		zap.Bool("使用心愿单", req.UseWishlist))
		
	if req.Count <= 0 {
		req.Count = 1
	}

	user, err := GetUser(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}
	// 打印玩家信息
	zapLogger.Info("抽奖前玩家信息",
		zap.Uint("用户ID", user.ID),
		zap.Int("等级", user.Level),
		zap.Int("灵石", user.SpiritStones),
		zap.Int("强化石", user.ReinforceStones))

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
		zapLogger.Info("抽奖失败 - 灵石不足",
			zap.Uint("用户", userID),
			zap.Int("拥有", user.SpiritStones),
			zap.Int("需要", required))
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
			equipment, err := GenerateEquipment(userID, user.Level, zapLogger)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
			zapLogger.Info("获得装备",
				zap.Uint("用户", userID),
				zap.String("装备名称", equipment.Name),
				zap.String("品质", equipment.Quality),
				zap.Any("属性", equipment.Stats))
			items = append(items, equipment)
		} else if req.PoolType == "pet" {
			pet, err := GeneratePet(userID, user.Level, zapLogger)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
				return
			}
			zapLogger.Info("获得灵宠",
				zap.Uint("用户", userID),
				zap.String("灵宠名称", pet.Name),
				zap.String("稀有度", pet.Rarity),
				zap.Any("属性", pet.CombatAttributes))
			items = append(items, pet)
		}
	}

	// 扣除灵石
	if err := DeductSpiritStones(userID, required); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
		return
	}

	newStones := user.SpiritStones - required
	zapLogger.Info("抽奖成功",
		zap.Uint("用户", userID),
		zap.String("抽奖池类型", req.PoolType),
		zap.Int("数量", req.Count),
		zap.Int("剩余灵石", newStones))
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"items":        items,
		"message":      "抽奖成功",
		"spirit_stones": newStones,
	})
}

// ProcessAutoActions 对应 POST /api/gacha/auto-actions
func ProcessAutoActions(c *gin.Context) {
	userID, ok := CurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "用户未授权"})
		return
	}

	// 获取zap logger
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	// 记录认证信息用于调试
	authHeader := c.GetHeader("Authorization")
	zapLogger.Info("自动处理请求",
		zap.Uint("用户ID", userID),
		zap.String("认证头", authHeader))

	var req struct {
		Items               []map[string]interface{} `json:"items"`
		AutoSellQualities   []string                 `json:"autoSellQualities"`
		AutoReleaseRarities []string                 `json:"autoReleaseRarities"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "请求参数错误"})
		zapLogger.Error("自动处理请求绑定错误",
			zap.Error(err))
		return
	}
	
	zapLogger.Info("自动处理请求参数",
		zap.Int("物品数", len(req.Items)),
		zap.Strings("自动出售品质", req.AutoSellQualities),
		zap.Strings("自动释放稀有度", req.AutoReleaseRarities))

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
			if err := DeleteEquipment(userID, idStr); err != nil {
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
			if err := DeletePet(userID, idStr); err != nil {
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
		if err := AddReinforceStones(userID, stonesGained); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "服务器错误", "error": err.Error()})
			return
		}
	}

	zapLogger.Info("自动处理成功",
		zap.Uint("用户", userID),
		zap.Int("出售物品数", len(soldItems)),
		zap.Int("释放灵宽数", len(releasedPets)),
		zap.Int("获得强化石", stonesGained))
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"soldItems":    soldItems,
		"releasedPets": releasedPets,
		"stonesGained": stonesGained,
		"message":      "自动处理完成",
	})
}