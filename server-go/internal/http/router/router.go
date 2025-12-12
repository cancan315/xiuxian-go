package router

import (
	"github.com/gin-gonic/gin"

	"xiuxian/server-go/internal/http/handlers/auth"
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
}
