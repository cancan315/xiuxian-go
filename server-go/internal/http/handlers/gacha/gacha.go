package gacha

import (
	"github.com/gin-gonic/gin"

	"xiuxian/server-go/internal/gacha"
)

// DrawGacha 对应 POST /api/gacha/draw
func DrawGacha(c *gin.Context) {
	gacha.DrawGacha(c)
}

// ProcessAutoActions 对应 POST /api/gacha/auto-actions
func ProcessAutoActions(c *gin.Context) {
	gacha.ProcessAutoActions(c)
}