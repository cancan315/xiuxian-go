package spirit

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	"xiuxian/server-go/internal/redis"
	"xiuxian/server-go/internal/websocket"

	"go.uber.org/zap"
)

// SpiritGrowManager 灵力增长管理器
type SpiritGrowManager struct {
	logger        *zap.Logger
	ticker        *time.Ticker
	stopChan      chan struct{}
	checkInterval time.Duration
	wsHandlers    *websocket.Handlers
}

// NewSpiritGrowManager 创建灵力增长管理器
func NewSpiritGrowManager(logger *zap.Logger, wsHandlers *websocket.Handlers) *SpiritGrowManager {
	return &SpiritGrowManager{
		logger:        logger,
		checkInterval: 1 * time.Second,
		stopChan:      make(chan struct{}),
		wsHandlers:    wsHandlers,
	}
}

// Start 启动灵力增长后台任务
func (m *SpiritGrowManager) Start() {
	m.ticker = time.NewTicker(m.checkInterval)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.processOnlinePlayersSpiritGrow()
			case <-m.stopChan:
				m.logger.Info("灵力增长后台任务已停止")
				return
			}
		}
	}()

	m.logger.Info("灵力增长后台任务已启动")
}

// Stop 停止灵力增长后台任务
func (m *SpiritGrowManager) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.stopChan)
}

// processOnlinePlayersSpiritGrow 处理在线玩家的灵力增长
func (m *SpiritGrowManager) processOnlinePlayersSpiritGrow() {
	// 从 Redis 获取在线玩家列表
	onlinePlayerIDs, err := redis.Client.SMembers(redis.Ctx, "server:online:players").Result()
	if err != nil {
		m.logger.Error("获取在线玩家列表失败", zap.Error(err))
		return
	}

	if len(onlinePlayerIDs) == 0 {
		//m.logger.Debug("没有在线玩家")
		return
	}

	// ✅ 打印获取到的列表，便日志排查
	//m.logger.Debug("处理在线玩家灵力增长",
	//	zap.Strings("onlinePlayerIDs", onlinePlayerIDs),
	//	zap.Int("count", len(onlinePlayerIDs)))

	// 处理每个在线玩家的灵力增长
	for _, playerID := range onlinePlayerIDs {
		m.calculateAndUpdateSpiritGain(playerID)
	}
}

// calculateAndUpdateSpiritGain 计算并更新单个玩家的灵力增长
func (m *SpiritGrowManager) calculateAndUpdateSpiritGain(playerID string) {
	// ✅ 检查玩家是否在线（通过Redis key是否存在）
	exists, err := redis.Client.Exists(redis.Ctx, "player:online:"+playerID).Result()
	if err != nil {
		m.logger.Error("检查玩家在线状态失败",
			zap.String("playerID", playerID),
			zap.Error(err))
		return
	}
	// 如果玩家数据不存在（已离线），直接返回
	if exists == 0 {
		//	m.logger.Info("玩家已离线，跳过灵力增长计算",
		//		zap.String("playerID", playerID),
		//		zap.String("reason", "player:online:{playerID} 不存在"))
		return
	}

	// ✅ 玩家在线，继续处理
	// m.logger.Debug("玩家已确认在线，开始计算灵力增长",
	//	zap.String("playerID", playerID))

	var user models.User

	// ... existing code ...
	parsedID, err := strconv.ParseUint(playerID, 10, 32)
	if err != nil {
		m.logger.Warn("无效的玩家ID", zap.String("playerID", playerID), zap.Error(err))
		return
	}

	// 按ID查询玩家
	if err := db.DB.First(&user, uint(parsedID)).Error; err != nil {
		m.logger.Warn("玩家不存在", zap.String("playerID", playerID))
		return
	}

	now := time.Now()
	elapsedSeconds := now.Sub(user.LastSpiritGainTime).Seconds()

	// ✅ 打印详细的时间信息，便日志排查：为什么一次性增长可能过大
	//m.logger.Info("灵力增长时间计算",
	//	zap.Uint("userID", user.ID),
	//	zap.Time("now", now),
	//	zap.Time("LastSpiritGainTime", user.LastSpiritGainTime),
	//	zap.Float64("elapsedSeconds", elapsedSeconds),
	//	zap.Float64("spiritRate", m.getPlayerSpiritRate(&user)))

	if elapsedSeconds < 1 {
		//m.logger.Debug("在线时间太短，跳过增长",
		//	zap.Uint("userID", user.ID),
		//	zap.Float64("elapsedSeconds", elapsedSeconds))
		return
	}

	oldSpirit := user.Spirit
	spiritRate := m.getPlayerSpiritRate(&user)
	spiritGain := 1.0 * spiritRate * elapsedSeconds

	// ✅ 保留两位小数
	oldSpirit = math.Round(oldSpirit*100) / 100
	user.Spirit = math.Round((user.Spirit+spiritGain)*100) / 100
	user.LastSpiritGainTime = now

	if err := db.DB.Model(&user).Updates(map[string]interface{}{
		"spirit":                user.Spirit,
		"last_spirit_gain_time": now,
	}).Error; err != nil {
		m.logger.Error("更新玩家灵力失败",
			zap.Uint("userID", user.ID),
			zap.Error(err))
		return
	}

	//m.logger.Info("灵力增长",
	//	zap.Uint("userID", user.ID),
	//	zap.Float64("elapsedSeconds", elapsedSeconds),
	//	zap.Float64("spiritRate", spiritRate),
	//	zap.Float64("gain", spiritGain),
	//	zap.Float64("oldSpirit", oldSpirit),
	//	zap.Float64("newSpirit", user.Spirit))

	// 推送灵力增长事件给客户端
	if m.wsHandlers != nil && m.wsHandlers.Spirit != nil {
		m.wsHandlers.Spirit.NotifySpiritUpdate(user.ID, oldSpirit, user.Spirit, spiritRate, elapsedSeconds)
	}
}

// getPlayerSpiritRate 获取玩家的灵力倍率
func (m *SpiritGrowManager) getPlayerSpiritRate(user *models.User) float64 {
	spiritRate := 1.0

	if user.BaseAttributes != nil && len(user.BaseAttributes) > 0 {
		var attrs map[string]interface{}
		if err := json.Unmarshal(user.BaseAttributes, &attrs); err == nil {
			if rate, ok := attrs["spiritRate"]; ok {
				if f, ok := rate.(float64); ok {
					spiritRate = f
				}
			}
		}
	}

	return spiritRate
}
