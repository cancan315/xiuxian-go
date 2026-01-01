package router

import (
	"github.com/gin-gonic/gin"

	"xiuxian/server-go/internal/http/handlers/alchemy"
	"xiuxian/server-go/internal/http/handlers/auth"
	"xiuxian/server-go/internal/http/handlers/cultivation"
	"xiuxian/server-go/internal/http/handlers/duel"
	"xiuxian/server-go/internal/http/handlers/exploration"
	"xiuxian/server-go/internal/http/handlers/gacha"
	"xiuxian/server-go/internal/http/handlers/online"
	"xiuxian/server-go/internal/http/handlers/player"
	"xiuxian/server-go/internal/http/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// /api/auth 路由，对齐 Node 版
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", auth.Register)
		authGroup.POST("/login", auth.Login)
		authGroup.POST("/logout", middleware.Protect(), auth.Logout) // ✅ 新增：登出路由
		authGroup.GET("/user", middleware.Protect(), auth.GetUser)
	}

	// /api/player 路由
	playerGroup := r.Group("/api/player")
	{
		// leaderboard 公开访问（支持四种排行榜）
		playerGroup.GET("/leaderboard", player.GetLeaderboard)
		playerGroup.GET("/leaderboard/:type", player.GetLeaderboard)

		// 其余接口需要认证
		playerGroup.Use(middleware.Protect())
		playerGroup.GET("/data", player.GetPlayerData)
		playerGroup.GET("/spirit", player.GetPlayerSpirit)
		playerGroup.PUT("/spirit", player.UpdateSpirit)
		playerGroup.GET("/spirit/gain", player.GetPlayerSpiritGain)
		playerGroup.POST("/spirit/apply-gain", player.ApplySpiritGain)
		playerGroup.PATCH("/data", player.UpdatePlayerData)
		playerGroup.DELETE("/items", player.DeleteItems)
		playerGroup.DELETE("/pets", player.DeletePets)
		playerGroup.POST("/change-name", player.ChangePlayerName)

		// 库存/装备相关查询接口

		playerGroup.GET("/equipment", player.GetPlayerEquipment)
		playerGroup.GET("/equipment/details/:id", player.GetEquipmentDetails)
		playerGroup.GET("/herbs", player.GetHerbsPaginated)
		playerGroup.GET("/pills", player.GetPillsPaginated)
		playerGroup.GET("/formulas", player.GetFormulasPaginated)

		// 装备系统写操作
		playerGroup.POST("/equipment/:id/enhance", player.EnhanceEquipment)
		playerGroup.POST("/equipment/:id/reforge", player.ReforgeEquipment)
		playerGroup.POST("/equipment/:id/reforge-confirm", player.ConfirmReforge)
		playerGroup.POST("/equipment/:id/equip", player.EquipEquipment)
		playerGroup.POST("/equipment/:id/unequip", player.UnequipEquipment)
		playerGroup.DELETE("/equipment/:id", player.SellEquipment)
		playerGroup.POST("/equipment/batch-sell", player.BatchSellEquipment)

		// 丹药系统
		playerGroup.POST("/pills/:id/consume", player.ConsumePill)

		// 灵宠系统
		playerGroup.POST("/pets/:id/deploy", player.DeployPet)
		playerGroup.POST("/pets/:id/recall", player.RecallPet)
		playerGroup.POST("/pets/:id/upgrade", player.UpgradePet)
		playerGroup.POST("/pets/:id/evolve", player.EvolvePet)
		playerGroup.POST("/pets/batch-release", player.BatchReleasePets)
	}

	// /api/online 路由（无需鉴权）
	onlineGroup := r.Group("/api/online")
	{
		onlineGroup.POST("/login", online.Login)
		onlineGroup.POST("/heartbeat", online.Heartbeat)
		onlineGroup.POST("/logout", online.Logout)
		onlineGroup.GET("/players", online.GetOnlinePlayers)
		onlineGroup.GET("/player/:playerId", online.GetPlayerOnlineStatus)
	}

	// /api/gacha 路由
	gachaGroup := r.Group("/api/gacha")
	{
		gachaGroup.Use(middleware.Protect())
		gachaGroup.POST("/draw", gacha.DrawGacha)
		gachaGroup.POST("/auto-actions", gacha.ProcessAutoActions)
	}

	// /api/exploration 路由
	explorationGroup := r.Group("/api/exploration")
	{
		explorationGroup.Use(middleware.Protect())
		explorationGroup.POST("/start", exploration.StartExploration)
		explorationGroup.POST("/event-choice", exploration.HandleEventChoice)
	}

	// /api/cultivation 路由
	cultivationGroup := r.Group("/api/cultivation")
	{
		cultivationGroup.Use(middleware.Protect())
		cultivationGroup.POST("/single", cultivation.SingleCultivate)
		// cultivationGroup.POST("/auto", cultivation.AutoCultivate)
		cultivationGroup.GET("/data", cultivation.GetCultivationData)
		// ✅ 新增：聚灵阵相关路由
		cultivationGroup.POST("/formation", cultivation.UseFormation)
		// ✅ 新增：结婴突破路由
		cultivationGroup.POST("/breakthrough-jieying", cultivation.BreakthroughJieYing)
	}

	// /api/alchemy 路由
	alchemyGroup := r.Group("/api/alchemy")
	{
		alchemyGroup.Use(middleware.Protect())
		alchemyGroup.GET("/configs", alchemy.GetConfigs)
		alchemyGroup.GET("/recipes", alchemy.GetAllRecipes)
		alchemyGroup.GET("/recipes/:recipeId", alchemy.GetRecipeDetail)
		alchemyGroup.POST("/craft", alchemy.CraftPill)
		alchemyGroup.POST("/buy-fragment", alchemy.BuyFragment)
	}

	// /api/duel 路由
	duelGroup := r.Group("/api/duel")
	{
		duelGroup.Use(middleware.Protect())
		// 斗法状态
		duelGroup.GET("/status", duel.GetDuelStatus) // 获取斗法状态（每日挑战次数、灵力消耗）
		// PvP对手相关
		duelGroup.GET("/opponents", duel.GetDuelOpponents)
		duelGroup.GET("/player/:playerId/battle-data", duel.GetPlayerBattleData)
		duelGroup.POST("/battle-attributes", duel.GetBattleAttributes) // 获取双方完整战斗属性
		duelGroup.GET("/records", duel.GetDuelRecords)
		duelGroup.POST("/record-result", duel.RecordBattleResult)
		duelGroup.POST("/claim-rewards", duel.ClaimBattleRewards)
		// PvP战斗相关端点
		duelGroup.POST("/start-pvp", duel.StartPvPBattle)
		duelGroup.POST("/execute-pvp-round", duel.ExecutePvPRound)
		duelGroup.POST("/end-pvp", duel.EndPvPBattle)
		// PvE妖2B挑战相关
		duelGroup.GET("/monster-challenges", duel.GetMonsterChallenges) // 获取妖2B挑战列表（支持分页和难度过滤）
		duelGroup.GET("/monster/:id", duel.GetMonsterByIDAPI)           // 获取妖2B详细信息
		// PvE战斗相关端点
		duelGroup.POST("/start-pve", duel.StartPvEBattle)
		duelGroup.POST("/execute-pve-round", duel.ExecutePvERound)
		duelGroup.POST("/end-pve", duel.EndPvEBattle)
		// 除魔卫道相关（新增）
		duelGroup.GET("/demon-slaying-challenges", duel.GetDemonSlayingChallenges) // 获取除魔卫道挑战列表
	}

	// 谊测端点 (有效期内不需要认证)
	testGroup := r.Group("/api/test")
	{
		testGroup.GET("/monster-challenges", duel.GetMonsterChallenges) // 此测试端点不需要认证
	}

	// /api/admin 路由（管理员操作）
	adminGroup := r.Group("/api/admin")
	{
		// 清除排行榜缓存（用于修复接口）
		adminGroup.POST("/leaderboard/clear-cache", player.ClearLeaderboardCache)
	}
}
