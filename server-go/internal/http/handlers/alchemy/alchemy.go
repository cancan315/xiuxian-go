package alchemy

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	alchemySvc "xiuxian/server-go/internal/alchemy"
)

// 请求/响应类型别名
type (
	AlchemyRequest      = alchemySvc.AlchemyRequest
	CraftRequest        = alchemySvc.CraftRequest
	BuyFragmentRequest  = alchemySvc.BuyFragmentRequest
	CraftResult         = alchemySvc.CraftResult
	BuyFragmentResult   = alchemySvc.BuyFragmentResult
	AllRecipesResponse  = alchemySvc.AllRecipesResponse
	RecipeDetailResponse = alchemySvc.RecipeDetailResponse
)

// GetAllRecipes 获取所有丹方列表
// GET /api/alchemy/recipes
func GetAllRecipes(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	playerLevelStr := c.Query("playerLevel")
	playerLevel := 1
	if levelVal, err := strconv.Atoi(playerLevelStr); err == nil {
		playerLevel = levelVal
	}

	zapLogger.Info("GetAllRecipes 入参",
		zap.Uint("userID", uid),
		zap.Int("playerLevel", playerLevel))

	service := alchemySvc.NewAlchemyService(uid)

	// TODO: 从前端状态获取以下数据
	unlockedRecipes := make(map[string]bool)
	fragments := make(map[string]int)

	// 为了演示，假设一个丹方已解锁
	unlockedRecipes["spirit_gathering"] = true
	fragments["spirit_gathering"] = 5

	allRecipes, err := service.GetAllRecipes(playerLevel)
	if err != nil {
		zapLogger.Error("获取丹方列表失败",
			zap.Uint("userID", uid),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取丹方列表失败", "error": err.Error()})
		return
	}

	zapLogger.Info("GetAllRecipes 出参",
		zap.Uint("userID", uid),
		zap.Int("recipeCount", len(allRecipes.Recipes)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    allRecipes,
		"message": "获取丹方列表成功",
	})
}

// GetRecipeDetail 获取单个丹方详细信息
// GET /api/alchemy/recipes/:recipeId
func GetRecipeDetail(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	recipeID := c.Param("recipeId")
	playerLevelStr := c.Query("playerLevel")
	playerLevel := 1
	if levelVal, err := strconv.Atoi(playerLevelStr); err == nil {
		playerLevel = levelVal
	}

	zapLogger.Info("GetRecipeDetail 入参",
		zap.Uint("userID", uid),
		zap.String("recipeID", recipeID),
		zap.Int("playerLevel", playerLevel))

	service := alchemySvc.NewAlchemyService(uid)

	// TODO: 从前端状态获取以下数据
	unlockedRecipes := make(map[string]bool)
	fragments := make(map[string]int)
	unlockedRecipes["spirit_gathering"] = true
	fragments["spirit_gathering"] = 5

	detail, err := service.GetRecipeDetail(recipeID, playerLevel, unlockedRecipes, fragments)
	if err != nil {
		zapLogger.Error("获取丹方详情失败",
			zap.Uint("userID", uid),
			zap.String("recipeID", recipeID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取丹方详情失败", "error": err.Error()})
		return
	}

	zapLogger.Info("GetRecipeDetail 出参",
		zap.Uint("userID", uid),
		zap.String("recipeID", recipeID))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    detail,
		"message": "获取丹方详情成功",
	})
}

// CraftPill 炼制丹药
// POST /api/alchemy/craft
func CraftPill(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req CraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	zapLogger.Info("CraftPill 入参",
		zap.Uint("userID", uid),
		zap.String("recipeID", req.RecipeID),
		zap.Int("playerLevel", req.PlayerLevel))

	service := alchemySvc.NewAlchemyService(uid)

	// 从请求体获取用户状态数据
	unlockedRecipes := make(map[string]bool)
	for _, recipeID := range req.UnlockedRecipes {
		unlockedRecipes[recipeID] = true
	}
	
	// 默认值处理
	playerLevel := req.PlayerLevel
	if playerLevel <= 0 {
		playerLevel = 1
	}
	luck := req.Luck
	if luck <= 0 {
		luck = 1.0
	}
	alchemyRate := req.AlchemyRate
	if alchemyRate <= 0 {
		alchemyRate = 1.0
	}

	result, err := service.CraftPill(req.RecipeID, playerLevel, unlockedRecipes, req.InventoryHerbs, luck, alchemyRate)
	if err != nil {
		zapLogger.Error("炼制丹药失败",
			zap.Uint("userID", uid),
			zap.String("recipeID", req.RecipeID),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	zapLogger.Info("CraftPill 出参",
		zap.Uint("userID", uid),
		zap.String("recipeID", req.RecipeID),
		zap.Bool("success", result.Success))

	c.JSON(http.StatusOK, gin.H{
		"success": result.Success,
		"data":    result,
		"message": result.Message,
	})
}

// BuyFragment 购买丹方残页
// POST /api/alchemy/buy-fragment
func BuyFragment(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	var req BuyFragmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数错误", "error": err.Error()})
		return
	}

	zapLogger.Info("BuyFragment 入参",
		zap.Uint("userID", uid),
		zap.String("recipeID", req.RecipeID),
		zap.Int("quantity", req.Quantity),
		zap.Int("currentFragments", req.CurrentFragments))

	service := alchemySvc.NewAlchemyService(uid)

	// 从请求体获取用户状态数据
	unlockedRecipes := make(map[string]bool)
	for _, recipeID := range req.UnlockedRecipes {
		unlockedRecipes[recipeID] = true
	}

	result, err := service.BuyFragment(req.RecipeID, req.Quantity, req.CurrentFragments, unlockedRecipes)
	if err != nil {
		zapLogger.Error("购买残页失败",
			zap.Uint("userID", uid),
			zap.String("recipeID", req.RecipeID),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	zapLogger.Info("BuyFragment 出参",
		zap.Uint("userID", uid),
		zap.String("recipeID", req.RecipeID),
		zap.Bool("recipeUnlocked", result.RecipeUnlocked))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"message": result.Message,
	})
}

// GetConfigs 获取炼丹系统配置 (品阶、类型、灵草等)
// GET /api/alchemy/configs
func GetConfigs(c *gin.Context) {
	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户未授权"})
		return
	}

	uid := userID.(uint)
	logger, _ := c.Get("zap_logger")
	zapLogger := logger.(*zap.Logger)

	zapLogger.Info("GetConfigs 入参",
		zap.Uint("userID", uid))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"grades":       alchemySvc.GetAllGrades(),
			"types":        alchemySvc.GetAllTypes(),
			"recipes":      alchemySvc.GetAllRecipeConfigs(),
			"herbs":        alchemySvc.GetAllHerbs(),
		},
		"message": "获取配置成功",
	})
}
