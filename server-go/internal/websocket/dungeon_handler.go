package websocket

import (
	"time"

	"go.uber.org/zap"
)

// DungeonEvent 战斗事件数据
type DungeonEvent struct {
	UserID      uint        `json:"userId"`
	EventType   string      `json:"eventType"`   // 事件类型: start, combat_round, victory, defeat, treasure
	Dungeon     string      `json:"dungeon"`     // 秘境名称
	Message     string      `json:"message"`     // 事件消息
	RoundNum    int         `json:"roundNum"`    // 战斗轮数
	PlayerHP    float64     `json:"playerHp"`    // 玩家HP
	EnemyHP     float64     `json:"enemyHp"`     // 敌人HP
	DamageDealt float64     `json:"damageDealt"` // 伤害造成
	DamageTaken float64     `json:"damageTaken"` // 伤害承受
	Loot        interface{} `json:"loot"`        // 战利品
	Timestamp   int64       `json:"timestamp"`
}

// DungeonHandler 战斗事件处理器
type DungeonHandler struct {
	manager *ConnectionManager
	logger  *zap.Logger
}

// NewDungeonHandler 创建战斗处理器
func NewDungeonHandler(manager *ConnectionManager, logger *zap.Logger) *DungeonHandler {
	return &DungeonHandler{
		manager: manager,
		logger:  logger,
	}
}

// BroadcastDungeonEvent 广播战斗事件
func (dh *DungeonHandler) BroadcastDungeonEvent(userID uint, event DungeonEvent) error {
	event.UserID = userID
	event.Timestamp = time.Now().Unix()

	err := dh.manager.SendToUser(userID, "dungeon:event", event)
	if err != nil {
		dh.logger.Error("广播战斗事件失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	dh.logger.Debug("战斗事件已发送",
		zap.Uint("userID", userID),
		zap.String("eventType", event.EventType),
		zap.String("dungeon", event.Dungeon))

	return nil
}

// NotifyDungeonStart 通知秘境开始
func (dh *DungeonHandler) NotifyDungeonStart(userID uint, dungeonName string) error {
	event := DungeonEvent{
		UserID:    userID,
		EventType: "start",
		Dungeon:   dungeonName,
		Message:   "秘境冒险开始！",
		Timestamp: time.Now().Unix(),
	}
	return dh.BroadcastDungeonEvent(userID, event)
}

// NotifyCombatRound 通知战斗轮次
func (dh *DungeonHandler) NotifyCombatRound(userID uint, dungeonName string, roundNum int,
	playerHP, enemyHP, damageDealt, damageTaken float64) error {

	message := "战斗进行中..."
	if damageDealt > 0 {
		message = "攻击成功！"
	}

	event := DungeonEvent{
		UserID:      userID,
		EventType:   "combat_round",
		Dungeon:     dungeonName,
		Message:     message,
		RoundNum:    roundNum,
		PlayerHP:    playerHP,
		EnemyHP:     enemyHP,
		DamageDealt: damageDealt,
		DamageTaken: damageTaken,
		Timestamp:   time.Now().Unix(),
	}
	return dh.BroadcastDungeonEvent(userID, event)
}

// NotifyVictory 通知战胜
func (dh *DungeonHandler) NotifyVictory(userID uint, dungeonName string, loot interface{}) error {
	event := DungeonEvent{
		UserID:    userID,
		EventType: "victory",
		Dungeon:   dungeonName,
		Message:   "恭喜战胜！",
		Loot:      loot,
		Timestamp: time.Now().Unix(),
	}
	return dh.BroadcastDungeonEvent(userID, event)
}

// NotifyDefeat 通知战败
func (dh *DungeonHandler) NotifyDefeat(userID uint, dungeonName string) error {
	event := DungeonEvent{
		UserID:    userID,
		EventType: "defeat",
		Dungeon:   dungeonName,
		Message:   "战斗失败，请继续修炼！",
		Timestamp: time.Now().Unix(),
	}
	return dh.BroadcastDungeonEvent(userID, event)
}

// NotifyTreasure 通知宝藏发现
func (dh *DungeonHandler) NotifyTreasure(userID uint, dungeonName string, treasure interface{}) error {
	event := DungeonEvent{
		UserID:    userID,
		EventType: "treasure",
		Dungeon:   dungeonName,
		Message:   "发现宝藏！",
		Loot:      treasure,
		Timestamp: time.Now().Unix(),
	}
	return dh.BroadcastDungeonEvent(userID, event)
}

// BroadcastMultipleEvents 广播多个事件（如连续战斗轮次）
func (dh *DungeonHandler) BroadcastMultipleEvents(userID uint, events []DungeonEvent) error {
	for _, event := range events {
		if err := dh.BroadcastDungeonEvent(userID, event); err != nil {
			return err
		}
		// 避免事件过于密集
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
