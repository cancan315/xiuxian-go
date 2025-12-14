package websocket

import (
	"time"

	"go.uber.org/zap"
)

// LeaderboardUpdate 排行榜更新数据
type LeaderboardUpdate struct {
	Type       string             `json:"type"`     // 更新类型: update, full_refresh
	Category   string             `json:"category"` // 排行榜类别: spirit, power, level
	UpdateTime int64              `json:"updateTime"`
	Top10      []LeaderboardEntry `json:"top10"`
	UserRank   *UserRankInfo      `json:"userRank"` // 用户当前排名信息
	Timestamp  int64              `json:"timestamp"`
}

// LeaderboardEntry 排行榜条目
type LeaderboardEntry struct {
	Rank      int     `json:"rank"`
	UserID    uint    `json:"userId"`
	Username  string  `json:"username"`
	Spirit    float64 `json:"spirit"`
	Power     float64 `json:"power"`
	Level     int     `json:"level"`
	AvatarURL string  `json:"avatarUrl"`
}

// UserRankInfo 用户排名信息
type UserRankInfo struct {
	Rank    int     `json:"rank"`
	Value   float64 `json:"value"`
	Percent float64 `json:"percent"` // 超越百分比
}

// LeaderboardHandler 排行榜处理器
type LeaderboardHandler struct {
	manager *ConnectionManager
	logger  *zap.Logger
}

// NewLeaderboardHandler 创建排行榜处理器
func NewLeaderboardHandler(manager *ConnectionManager, logger *zap.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{
		manager: manager,
		logger:  logger,
	}
}

// BroadcastUpdate 广播排行榜更新
func (lh *LeaderboardHandler) BroadcastUpdate(userID uint, update LeaderboardUpdate) error {
	update.Timestamp = time.Now().Unix()

	err := lh.manager.SendToUser(userID, "leaderboard:update", update)
	if err != nil {
		lh.logger.Error("广播排行榜更新失败",
			zap.Uint("userID", userID),
			zap.Error(err))
		return err
	}

	lh.logger.Debug("排行榜更新已发送",
		zap.Uint("userID", userID),
		zap.String("category", update.Category))

	return nil
}

// NotifyRankChange 通知排名变化
func (lh *LeaderboardHandler) NotifyRankChange(userID uint, category string, newRank, oldRank int, value float64) error {
	changeType := "update"
	if newRank < oldRank {
		changeType = "rank_up"
	} else if newRank > oldRank {
		changeType = "rank_down"
	}

	update := LeaderboardUpdate{
		Type:       changeType,
		Category:   category,
		UpdateTime: time.Now().Unix(),
		UserRank: &UserRankInfo{
			Rank:  newRank,
			Value: value,
		},
		Timestamp: time.Now().Unix(),
	}

	return lh.BroadcastUpdate(userID, update)
}

// NotifyFullRefresh 通知完整刷新（用户上线或首次加载）
func (lh *LeaderboardHandler) NotifyFullRefresh(userID uint, category string, top10 []LeaderboardEntry, userRank *UserRankInfo) error {
	update := LeaderboardUpdate{
		Type:       "full_refresh",
		Category:   category,
		UpdateTime: time.Now().Unix(),
		Top10:      top10,
		UserRank:   userRank,
		Timestamp:  time.Now().Unix(),
	}

	return lh.BroadcastUpdate(userID, update)
}

// NotifySpiritLeaderboardUpdate 通知灵力排行榜更新
func (lh *LeaderboardHandler) NotifySpiritLeaderboardUpdate(userID uint, top10 []LeaderboardEntry, userRank *UserRankInfo) error {
	return lh.NotifyFullRefresh(userID, "spirit", top10, userRank)
}

// NotifyPowerLeaderboardUpdate 通知战力排行榜更新
func (lh *LeaderboardHandler) NotifyPowerLeaderboardUpdate(userID uint, top10 []LeaderboardEntry, userRank *UserRankInfo) error {
	return lh.NotifyFullRefresh(userID, "power", top10, userRank)
}

// NotifyLevelLeaderboardUpdate 通知等级排行榜更新
func (lh *LeaderboardHandler) NotifyLevelLeaderboardUpdate(userID uint, top10 []LeaderboardEntry, userRank *UserRankInfo) error {
	return lh.NotifyFullRefresh(userID, "level", top10, userRank)
}

// GetTopPlayers 获取排行榜前N名（辅助方法，从缓存或数据库获取）
func (lh *LeaderboardHandler) GetTopPlayers(category string, limit int) ([]LeaderboardEntry, error) {
	// TODO: 从缓存或数据库获取排行榜数据
	// 这里是占位符实现
	return []LeaderboardEntry{}, nil
}

// BroadcastToAllOnline 广播到所有在线玩家（系统级排行榜更新）
func (lh *LeaderboardHandler) BroadcastToAllOnline(category string, top10 []LeaderboardEntry) error {
	// TODO: 遍历所有在线玩家，向每个玩家发送其对应的排行榜信息
	lh.logger.Info("广播排行榜更新到所有在线玩家",
		zap.String("category", category),
		zap.Int("topCount", len(top10)))
	return nil
}
