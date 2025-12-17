package spirit

import (
	"encoding/json"
	"fmt"
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
	syncTicker    *time.Ticker
	stopChan      chan struct{}
	checkInterval time.Duration
	syncInterval  time.Duration
	wsHandlers    *websocket.Handlers
}

// NewSpiritGrowManager 创建灵力增长管理器
func NewSpiritGrowManager(logger *zap.Logger, wsHandlers *websocket.Handlers) *SpiritGrowManager {
	return &SpiritGrowManager{
		logger:        logger,
		checkInterval: 1 * time.Second,
		syncInterval:  30 * time.Second, // 每30秒同步一次数据库
		stopChan:      make(chan struct{}),
		wsHandlers:    wsHandlers,
	}
}

// Start 启动灵力增长后台任务
func (m *SpiritGrowManager) Start() {
	m.ticker = time.NewTicker(m.checkInterval)
	m.syncTicker = time.NewTicker(m.syncInterval)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.processOnlinePlayersSpiritGrow()
			case <-m.syncTicker.C:
				m.syncAllPlayersSpirit()
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
	if m.syncTicker != nil {
		m.syncTicker.Stop()
	}
	// 停止前同步一次所有玩家数据
	m.syncAllPlayersSpirit()
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
		return
	}

	// 处理每个在线玩家的灵力增长（仅在Redis中累积）
	for _, playerID := range onlinePlayerIDs {
		m.calculateSpiritGainInRedis(playerID)
	}
}

// calculateSpiritGainInRedis 计算灵力增长并在Redis中累积
func (m *SpiritGrowManager) calculateSpiritGainInRedis(playerID string) {
	// ✅ 检查玩家是否在线
	exists, err := redis.Client.Exists(redis.Ctx, "player:online:"+playerID).Result()
	if err != nil || exists == 0 {
		return
	}

	parsedID, err := strconv.ParseUint(playerID, 10, 32)
	if err != nil {
		m.logger.Warn("无效的玩家ID", zap.String("playerID", playerID), zap.Error(err))
		return
	}

	userID := uint(parsedID)
	now := time.Now()

	// Redis key 定义
	lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)
	spiritCacheKey := fmt.Sprintf("player:spirit:cache:%d", userID)

	// 获取上次计算时间
	lastTimeStr, err := redis.Client.Get(redis.Ctx, lastGainTimeKey).Result()
	var lastTime time.Time

	if err != nil && err.Error() != "redis: nil" {
		m.logger.Error("获取上次灵力计算时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 若Redis中无记录，从数据库查询
	if err != nil && err.Error() == "redis: nil" {
		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			m.logger.Warn("玩家不存在", zap.Uint("userID", userID))
			return
		}
		lastTime = user.LastSpiritGainTime
	} else {
		// 解析Redis中的时间
		parsedTime, err := time.Parse(time.RFC3339Nano, lastTimeStr)
		if err != nil {
			m.logger.Warn("解析灵力计算时间失败",
				zap.Uint("userID", userID),
				zap.Error(err))
			return
		}
		lastTime = parsedTime
	}

	elapsedSeconds := now.Sub(lastTime).Seconds()

	// 时间间隔太短，跳过
	if elapsedSeconds < 1 {
		return
	}

	// 获取玩家当前灵力（优先从Redis缓存，否则从数据库）
	currentSpiritStr, err := redis.Client.Get(redis.Ctx, spiritCacheKey).Result()
	var currentSpirit float64

	if err != nil && err.Error() == "redis: nil" {
		// Redis中无缓存，从数据库读取
		var user models.User
		if err := db.DB.First(&user, userID).Error; err != nil {
			m.logger.Warn("玩家不存在", zap.Uint("userID", userID))
			return
		}
		currentSpirit = user.Spirit
	} else if err != nil {
		m.logger.Error("获取Redis灵力缓存失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	} else {
		// 从Redis字符串转换为float64
		parsedSpirit, err := strconv.ParseFloat(currentSpiritStr, 64)
		if err != nil {
			m.logger.Warn("解析灵力值失败",
				zap.Uint("userID", userID),
				zap.Error(err))
			return
		}
		currentSpirit = parsedSpirit
	}

	// 计算灵力增长
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return
	}

	oldSpirit := currentSpirit
	spiritRate := m.getPlayerSpiritRate(&user)
	spiritGain := 1.0 * spiritRate * elapsedSeconds

	// 保留两位小数
	newSpirit := math.Round((currentSpirit+spiritGain)*100) / 100

	// 将新值存储到Redis缓存中（不立即写数据库）
	spiritStr := fmt.Sprintf("%.2f", newSpirit)
	err = redis.Client.Set(redis.Ctx, spiritCacheKey, spiritStr, 0).Err()
	if err != nil {
		m.logger.Error("更新Redis灵力缓存失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 更新Redis中的上次计算时间
	err = redis.Client.Set(redis.Ctx, lastGainTimeKey, now.Format(time.RFC3339Nano), 0).Err()
	if err != nil {
		m.logger.Error("更新Redis灵力计算时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 推送灵力增长事件给客户端（实时反馈）
	if m.wsHandlers != nil && m.wsHandlers.Spirit != nil {
		m.wsHandlers.Spirit.NotifySpiritUpdate(userID, oldSpirit, newSpirit, spiritRate, elapsedSeconds)
	}
}

// syncAllPlayersSpirit 将所有玩家的灵力从Redis同步到数据库
func (m *SpiritGrowManager) syncAllPlayersSpirit() {
	// 获取在线玩家列表
	onlinePlayerIDs, err := redis.Client.SMembers(redis.Ctx, "server:online:players").Result()
	if err != nil {
		m.logger.Error("获取在线玩家列表失败", zap.Error(err))
		return
	}

	for _, playerID := range onlinePlayerIDs {
		m.syncPlayerSpiritToDatabase(playerID)
	}
}

// syncPlayerSpiritToDatabase 将单个玩家的灵力从Redis同步到数据库
func (m *SpiritGrowManager) syncPlayerSpiritToDatabase(playerID string) {
	parsedID, err := strconv.ParseUint(playerID, 10, 32)
	if err != nil {
		return
	}

	userID := uint(parsedID)

	// Redis key 定义
	spiritCacheKey := fmt.Sprintf("player:spirit:cache:%d", userID)
	lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)

	// 从Redis获取缓存的灵力值
	spiritStr, err := redis.Client.Get(redis.Ctx, spiritCacheKey).Result()
	if err != nil && err.Error() == "redis: nil" {
		// Redis中无缓存，无需同步
		return
	}
	if err != nil {
		m.logger.Error("获取Redis灵力缓存失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	newSpirit, err := strconv.ParseFloat(spiritStr, 64)
	if err != nil {
		m.logger.Error("解析灵力值失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 从Redis获取上次更新时间
	lastTimeStr, err := redis.Client.Get(redis.Ctx, lastGainTimeKey).Result()
	if err != nil {
		m.logger.Error("获取上次更新时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	lastTime, err := time.Parse(time.RFC3339Nano, lastTimeStr)
	if err != nil {
		m.logger.Error("解析时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 同步到数据库
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"spirit":                newSpirit,
		"last_spirit_gain_time": lastTime,
	}).Error; err != nil {
		m.logger.Error("同步灵力到数据库失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	m.logger.Debug("已同步玩家灵力到数据库",
		zap.Uint("userID", userID),
		zap.Float64("spirit", newSpirit))
}

// GetPlayerSpiritFromCache 获取玩家的灵力值（先查Redis后查数据库）
func (m *SpiritGrowManager) GetPlayerSpiritFromCache(userID uint) float64 {
	spiritCacheKey := fmt.Sprintf("player:spirit:cache:%d", userID)

	// 优先从Redis获取
	spiritStr, err := redis.Client.Get(redis.Ctx, spiritCacheKey).Result()
	if err == nil {
		spirit, err := strconv.ParseFloat(spiritStr, 64)
		if err == nil {
			return spirit
		}
	}

	// 降级到数据库查询
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return 0
	}
	return user.Spirit
}

// SyncPlayerSpiritOnLogout 玩家离线时强制同步灵力
func (m *SpiritGrowManager) SyncPlayerSpiritOnLogout(userID uint) {
	m.syncPlayerSpiritToDatabase(strconv.FormatUint(uint64(userID), 10))

	// 清理Redis缓存
	spiritCacheKey := fmt.Sprintf("player:spirit:cache:%d", userID)
	lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)

	redis.Client.Del(redis.Ctx, spiritCacheKey, lastGainTimeKey)
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
