package websocket

import (
	"encoding/json"
	"math"
	"time"
	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"

	"go.uber.org/zap"
)

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// SpiritGrowthEvent 灵力增长事件数据
type SpiritGrowthEvent struct {
	UserID         uint    `json:"userId"`
	OldSpirit      float64 `json:"oldSpirit"`
	NewSpirit      float64 `json:"newSpirit"`
	GainAmount     float64 `json:"gainAmount"`
	SpiritRate     float64 `json:"spiritRate"`
	ElapsedSeconds float64 `json:"elapsedSeconds"`
	Timestamp      int64   `json:"timestamp"`
}

// SpiritHandler 灵力更新处理器
type SpiritHandler struct {
	manager *ConnectionManager
	logger  *zap.Logger
}

// NewSpiritHandler 创建灵力处理器
func NewSpiritHandler(manager *ConnectionManager, logger *zap.Logger) *SpiritHandler {
	return &SpiritHandler{
		manager: manager,
		logger:  logger,
	}
}

// BroadcastSpiritGrowth 广播灵力增长事件
func (sh *SpiritHandler) BroadcastSpiritGrowth(userID uint, event SpiritGrowthEvent) error {
	event.UserID = userID
	event.Timestamp = time.Now().Unix()

	err := sh.manager.SendToUser(userID, "spirit:grow", event)
	if err != nil {
		sh.logger.Error("广播灵力增长事件失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	//sh.logger.Debug("灵力增长事件已发送",
	//	zap.Uint("userID", userID),
	//	zap.Float64("gainAmount", event.GainAmount),
	//	zap.Float64("newSpirit", event.NewSpirit))

	return nil
}

// NotifySpiritUpdate 通知灵力更新（来自后台任务）
func (sh *SpiritHandler) NotifySpiritUpdate(userID uint, oldSpirit, newSpirit, spiritRate, elapsedSeconds float64) error {
	if newSpirit <= oldSpirit {
		return nil // 没有增长，不发送消息
	}
	// 保留一位小数

	event := SpiritGrowthEvent{
		UserID:         userID,
		OldSpirit:      round1(oldSpirit),
		NewSpirit:      round1(newSpirit),
		GainAmount:     round1(newSpirit - oldSpirit),
		SpiritRate:     spiritRate,
		ElapsedSeconds: elapsedSeconds,
		Timestamp:      time.Now().Unix(),
	}

	return sh.BroadcastSpiritGrowth(userID, event)
}

// GetSpiritUpdateFromDB 从数据库获取灵力信息并发送更新
func (sh *SpiritHandler) GetSpiritUpdateFromDB(userID uint) error {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		sh.logger.Error("获取用户数据失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	// 构建灵力更新事件
	spiritRate := 1.0 // 从baseAttributes中获取
	// TODO: 解析BaseAttributes JSON获取spiritRate

	event := SpiritGrowthEvent{
		UserID:     userID,
		NewSpirit:  round1(user.Spirit),
		SpiritRate: spiritRate,
		Timestamp:  time.Now().Unix(),
	}

	dataBytes, _ := json.Marshal(event)
	msg := &Message{
		Type:      "spirit:grow",
		UserID:    userID,
		Timestamp: time.Now().Unix(),
		Data:      dataBytes,
	}

	sh.manager.broadcast <- msg
	return nil
}

// SubscribeSpiritUpdates 订阅灵力更新（后台任务用）
// 这个方法可以被灵力增长后台任务调用
func (sh *SpiritHandler) SubscribeSpiritUpdates() chan *SpiritGrowthEvent {
	eventChan := make(chan *SpiritGrowthEvent, 10)
	return eventChan
}
