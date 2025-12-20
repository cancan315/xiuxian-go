package router

import (
	"github.com/gin-gonic/gin"

	"xiuxian/server-go/internal/http/handlers/alchemy"
	"xiuxian/server-go/internal/http/handlers/auth"
	"xiuxian/server-go/internal/http/handlers/cultivation"
	"xiuxian/server-go/internal/http/handlers/dungeon"
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
		// leaderboard 公开访问
		playerGroup.GET("/leaderboard", player.GetLeaderboard)

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

		// 库存/装备相关查询接口

		playerGroup.GET("/equipment", player.GetPlayerEquipment)
		playerGroup.GET("/equipment/details/:id", player.GetEquipmentDetails)

		// 装备系统写操作
		playerGroup.POST("/equipment/:id/enhance", player.EnhanceEquipment)
		playerGroup.POST("/equipment/:id/reforge", player.ReforgeEquipment)
		playerGroup.POST("/equipment/:id/reforge-confirm", player.ConfirmReforge)
		playerGroup.POST("/equipment/:id/equip", player.EquipEquipment)
		playerGroup.POST("/equipment/:id/unequip", player.UnequipEquipment)
		playerGroup.DELETE("/equipment/:id", player.SellEquipment)
		playerGroup.POST("/equipment/batch-sell", player.BatchSellEquipment)

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
	}

	// /api/dungeon 路由
	dungeonGroup := r.Group("/api/dungeon")
	{
		dungeonGroup.Use(middleware.Protect())
		dungeonGroup.POST("/start", dungeon.StartDungeon)
		dungeonGroup.GET("/buffs/:floor", dungeon.GetBuffOptions)
		dungeonGroup.POST("/select-buff", dungeon.SelectBuff)
		dungeonGroup.POST("/save-buff", dungeon.SaveBuff)      // 新增: 保存已选增益
		dungeonGroup.GET("/load-session", dungeon.LoadSession) // 新增: 加载断线重连的会话
		dungeonGroup.POST("/fight", dungeon.StartFight)
		dungeonGroup.GET("/round-data", dungeon.GetRoundData)     // 新增: 获取回合数据
		dungeonGroup.POST("/execute-round", dungeon.ExecuteRound) // 新增: 执行回合
		dungeonGroup.POST("/end", dungeon.EndDungeon)
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
}
