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

	"go.uber.org/zap"
)

// SpiritGrowManager 灵力增长管理器
type SpiritGrowManager struct {
	logger           *zap.Logger
	ticker           *time.Ticker
	syncTicker       *time.Ticker
	autoGrowTicker   *time.Ticker
	stopChan         chan struct{}
	checkInterval    time.Duration
	syncInterval     time.Duration
	autoGrowInterval time.Duration
}

// NewSpiritGrowManager 创建灵力增长管理器
func NewSpiritGrowManager(logger *zap.Logger) *SpiritGrowManager {
	return &SpiritGrowManager{
		logger:           logger,
		checkInterval:    5 * time.Second,
		syncInterval:     30 * time.Second,
		autoGrowInterval: 1 * time.Minute,
		stopChan:         make(chan struct{}),
	}
}

// Start 启动灵力增长后台任务
func (m *SpiritGrowManager) Start() {
	m.ticker = time.NewTicker(m.checkInterval)
	m.syncTicker = time.NewTicker(m.syncInterval)
	m.autoGrowTicker = time.NewTicker(m.autoGrowInterval)

	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.processOnlinePlayersSpiritGrow()
			case <-m.syncTicker.C:
				m.syncAllPlayersSpirit()
			case <-m.autoGrowTicker.C:
				m.processAutoSpiritGrowth()
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
	if m.autoGrowTicker != nil {
		m.autoGrowTicker.Stop()
	}
	// 停止前同步一次所有玩家数据
	m.syncAllPlayersSpirit()
	close(m.stopChan)
}

// processAutoSpiritGrowth 每分钟处理一次自动灵力增长
func (m *SpiritGrowManager) processAutoSpiritGrowth() {
	// 从 Redis 获取所有需要自动增长的玩家ID列表
	playerIDStrs, err := redis.Client.SMembers(redis.Ctx, "spirit:auto:grow:players").Result()
	if err != nil {
		m.logger.Error("[灵力自动增长] 获取玩家列表失败", zap.Error(err))
		return
	}

	if len(playerIDStrs) == 0 {
		return
	}

	m.logger.Debug("[灵力自动增长] 开始处理玩家灵力增长",
		zap.Int("playerCount", len(playerIDStrs)))

	// 处理每个玩家的灵力增长
	for _, playerIDStr := range playerIDStrs {
		userID, err := strconv.ParseUint(playerIDStr, 10, 32)
		if err != nil {
			m.logger.Warn("[灵力自动增长] 无效的玩家ID",
				zap.String("playerID", playerIDStr),
				zap.Error(err))
			continue
		}

		m.addAutoSpiritToDatabase(uint(userID))
	}
}

// addAutoSpiritToDatabase 计算60秒灵力增长并直接写入数据库
func (m *SpiritGrowManager) addAutoSpiritToDatabase(userID uint) {
	// 从数据库获取玩家信息
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		m.logger.Warn("[灵力自动增长] 玩家不存在",
			zap.Uint("userID", userID))
		return
	}

	// 获取灵力倍率
	spiritRate := m.getPlayerSpiritRate(&user)

	// 计算60秒灵力增长量（1.0是基础增长速度）
	spiritGain := 1.0 * spiritRate * 60.0

	// 保留两位小数
	spiritGain = math.Round(spiritGain*100) / 100

	// 更新数据库中的灵力
	if err := db.DB.Model(&user).
		Update("spirit", user.Spirit+spiritGain).
		Error; err != nil {
		m.logger.Error("[灵力自动增长] 更新灵力失败",
			zap.Uint("userID", userID),
			zap.Float64("spiritGain", spiritGain),
			zap.Error(err))
		return
	}

	m.logger.Debug("[灵力自动增长] 已增加玩家灵力",
		zap.Uint("userID", userID),
		zap.Float64("spiritGain", spiritGain),
		zap.Float64("currentSpirit", user.Spirit+spiritGain))
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

	// 处理每个在线玩家的灵力增长（直接写入数据库）
	for _, playerID := range onlinePlayerIDs {
		m.calculateSpiritGainInDatabase(playerID)
	}
}

// calculateSpiritGainInDatabase 计算灵力增长并累加到Redis
func (m *SpiritGrowManager) calculateSpiritGainInDatabase(playerID string) {
	// ✅ 检查玩家是否在线
	exists, err := redis.Client.Exists(redis.Ctx, "player:online:"+playerID).Result()
	if err != nil || exists == 0 {
		// 玩家不在线，清理相关的Redis键
		parsedID, parseErr := strconv.ParseUint(playerID, 10, 32)
		if parseErr != nil {
			m.logger.Warn("无效的玩家ID", zap.String("playerID", playerID), zap.Error(parseErr))
			return
		}

		userID := uint(parsedID)
		lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)
		spiritGainKey := fmt.Sprintf("player:spirit:gain:%d", userID)

		// 清理Redis键
		redis.Client.Del(redis.Ctx, lastGainTimeKey, spiritGainKey)
		// m.logger.Debug("玩家不在线，已清理Redis键", zap.Uint("userID", userID))
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
	spiritGainKey := fmt.Sprintf("player:spirit:gain:%d", userID)

	// 从数据库获取玩家信息
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		m.logger.Warn("玩家不存在", zap.Uint("userID", userID))
		return
	}

	// 获取上次计算时间
	lastTimeStr, err := redis.Client.Get(redis.Ctx, lastGainTimeKey).Result()
	var lastTime time.Time

	if err != nil && err.Error() != "redis: nil" {
		m.logger.Error("获取上次灵力计算时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 若Redis中无记录，使用数据库中的上次计算时间
	if err != nil && err.Error() == "redis: nil" {
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

	// 计算灵力增长
	spiritRate := m.getPlayerSpiritRate(&user)
	spiritGain := 1.0 * spiritRate * elapsedSeconds

	// 保留两位小数
	spiritGain = math.Round(spiritGain*100) / 100

	// 将灵力增长量累加到Redis
	err = redis.Client.IncrByFloat(redis.Ctx, spiritGainKey, spiritGain).Err()
	if err != nil {
		m.logger.Error("更新Redis灵力增长缓存失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	// 更新Redis中的上次计算时间（仅用于下次计算）
	err = redis.Client.Set(redis.Ctx, lastGainTimeKey, now.Format(time.RFC3339Nano), 0).Err()
	if err != nil {
		m.logger.Error("更新Redis灵力计算时间失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	m.logger.Info("已更新玩家灵力增长到Redis",
		zap.Uint("userID", userID),
		zap.Float64("spiritGain", spiritGain))
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

// syncPlayerSpiritToDatabase 将单个玩家的灵力从Redis同步到数据库（仅在停止时使用）
func (m *SpiritGrowManager) syncPlayerSpiritToDatabase(playerID string) {
	parsedID, err := strconv.ParseUint(playerID, 10, 32)
	if err != nil {
		return
	}

	userID := uint(parsedID)
	lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)

	// 从Redis获取上次更新时间
	lastTimeStr, err := redis.Client.Get(redis.Ctx, lastGainTimeKey).Result()
	if err != nil && err.Error() == "redis: nil" {
		// Redis中无缓存，无需同步
		return
	}
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

	// 更新数据库中的上次计算时间
	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).Update("last_spirit_gain_time", lastTime).Error; err != nil {
		m.logger.Error("同步上次计算时间到数据库失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return
	}

	m.logger.Debug("已同步玩家上次计算时间到数据库",
		zap.Uint("userID", userID))
}

// GetPlayerSpiritFromCache 获取玩家的灵力值（直接从数据库查询）
func (m *SpiritGrowManager) GetPlayerSpiritFromCache(userID uint) float64 {
	// 直接从数据库查询（不再使用Redis缓存）
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return 0
	}
	return user.Spirit
}

// GetPlayerSpiritGain 获取玩家在Redis中累积的灵力增长量
func (m *SpiritGrowManager) GetPlayerSpiritGain(userID uint) (float64, error) {
	spiritGainKey := fmt.Sprintf("player:spirit:gain:%d", userID)

	result, err := redis.Client.Get(redis.Ctx, spiritGainKey).Result()
	if err != nil && err.Error() == "redis: nil" {
		// 键不存在，返回0
		return 0, nil
	}
	if err != nil {
		m.logger.Error("获取Redis灵力增长失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return 0, err
	}

	spiritGain, err := strconv.ParseFloat(result, 64)
	if err != nil {
		m.logger.Error("解析灵力增长值失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return 0, err
	}

	return spiritGain, nil
}

// ClearPlayerSpiritGain 清空玩家在Redis中的灵力增长数据
func (m *SpiritGrowManager) ClearPlayerSpiritGain(userID uint) error {
	spiritGainKey := fmt.Sprintf("player:spirit:gain:%d", userID)

	err := redis.Client.Del(redis.Ctx, spiritGainKey).Err()
	if err != nil {
		m.logger.Error("清空玩家灵力增长缓存失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	m.logger.Info("已清空玩家灵力增长缓存",
		zap.Uint("userID", userID))
	return nil
}

// SyncPlayerSpiritOnLogout 玩家离线时清理Redis缓存
func (m *SpiritGrowManager) SyncPlayerSpiritOnLogout(userID uint) {
	// 清理Redis缓存中的计算时间和灵力增长
	lastGainTimeKey := fmt.Sprintf("player:spirit:lastGainTime:%d", userID)
	spiritGainKey := fmt.Sprintf("player:spirit:gain:%d", userID)
	redis.Client.Del(redis.Ctx, lastGainTimeKey, spiritGainKey)
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
