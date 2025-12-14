package websocket

import (
	"time"

	"go.uber.org/zap"
)

// ExplorationEvent 探索事件数据
type ExplorationEvent struct {
	UserID       uint        `json:"userId"`
	EventType    string      `json:"eventType"`    // 事件类型: start, progress, discovery, complete, failure
	ExploreName  string      `json:"exploreName"`  // 探索地点名称
	Message      string      `json:"message"`      // 事件消息
	Progress     int         `json:"progress"`     // 进度百分比
	DurationSecs int         `json:"durationSecs"` // 持续时间（秒）
	ElapsedSecs  int         `json:"elapsedSecs"`  // 已用时间（秒）
	Discovery    interface{} `json:"discovery"`    // 发现内容
	Reward       interface{} `json:"reward"`       // 奖励
	ErrorMsg     string      `json:"errorMsg"`     // 失败信息
	Timestamp    int64       `json:"timestamp"`
}

// ExplorationHandler 探索事件处理器
type ExplorationHandler struct {
	manager *ConnectionManager
	logger  *zap.Logger
}

// NewExplorationHandler 创建探索处理器
func NewExplorationHandler(manager *ConnectionManager, logger *zap.Logger) *ExplorationHandler {
	return &ExplorationHandler{
		manager: manager,
		logger:  logger,
	}
}

// BroadcastExplorationEvent 广播探索事件
func (eh *ExplorationHandler) BroadcastExplorationEvent(userID uint, event ExplorationEvent) error {
	event.UserID = userID
	event.Timestamp = time.Now().Unix()

	err := eh.manager.SendToUser(userID, "exploration:event", event)
	if err != nil {
		eh.logger.Error("广播探索事件失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	eh.logger.Debug("探索事件已发送",
		zap.Uint("userID", userID),
		zap.String("eventType", event.EventType),
		zap.String("exploreName", event.ExploreName))

	return nil
}

// NotifyExplorationStart 通知探索开始
func (eh *ExplorationHandler) NotifyExplorationStart(userID uint, exploreName string, durationSecs int) error {
	event := ExplorationEvent{
		UserID:       userID,
		EventType:    "start",
		ExploreName:  exploreName,
		Message:      "探索开始！",
		DurationSecs: durationSecs,
		Progress:     0,
		Timestamp:    time.Now().Unix(),
	}
	return eh.BroadcastExplorationEvent(userID, event)
}

// NotifyExplorationProgress 通知探索进度
func (eh *ExplorationHandler) NotifyExplorationProgress(userID uint, exploreName string,
	elapsedSecs, durationSecs int) error {

	progress := (elapsedSecs * 100) / durationSecs
	if progress > 100 {
		progress = 100
	}

	event := ExplorationEvent{
		UserID:       userID,
		EventType:    "progress",
		ExploreName:  exploreName,
		Message:      "正在探索中...",
		Progress:     progress,
		ElapsedSecs:  elapsedSecs,
		DurationSecs: durationSecs,
		Timestamp:    time.Now().Unix(),
	}
	return eh.BroadcastExplorationEvent(userID, event)
}

// NotifyDiscovery 通知发现
func (eh *ExplorationHandler) NotifyDiscovery(userID uint, exploreName string, discovery interface{}) error {
	event := ExplorationEvent{
		UserID:      userID,
		EventType:   "discovery",
		ExploreName: exploreName,
		Message:     "发现了什么！",
		Discovery:   discovery,
		Timestamp:   time.Now().Unix(),
	}
	return eh.BroadcastExplorationEvent(userID, event)
}

// NotifyExplorationComplete 通知探索完成
func (eh *ExplorationHandler) NotifyExplorationComplete(userID uint, exploreName string, reward interface{}) error {
	event := ExplorationEvent{
		UserID:      userID,
		EventType:   "complete",
		ExploreName: exploreName,
		Message:     "探索完成！",
		Progress:    100,
		Reward:      reward,
		Timestamp:   time.Now().Unix(),
	}
	return eh.BroadcastExplorationEvent(userID, event)
}

// NotifyExplorationFailure 通知探索失败
func (eh *ExplorationHandler) NotifyExplorationFailure(userID uint, exploreName string, errorMsg string) error {
	event := ExplorationEvent{
		UserID:      userID,
		EventType:   "failure",
		ExploreName: exploreName,
		Message:     "探索失败！",
		ErrorMsg:    errorMsg,
		Timestamp:   time.Now().Unix(),
	}
	return eh.BroadcastExplorationEvent(userID, event)
}

// NotifyMultipleDiscoveries 通知多个发现（如探索期间发现多个物品）
func (eh *ExplorationHandler) NotifyMultipleDiscoveries(userID uint, exploreName string, discoveries []interface{}) error {
	for i, discovery := range discoveries {
		if err := eh.NotifyDiscovery(userID, exploreName, discovery); err != nil {
			eh.logger.Error("广播发现事件失败",
				zap.Uint("userID", userID),
				zap.Int("discoveryIndex", i))
			return err
		}
		// 避免事件过于密集
		time.Sleep(50 * time.Millisecond)
	}
	return nil
}
