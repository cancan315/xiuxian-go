package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"xiuxian/server-go/internal/db"
	"xiuxian/server-go/internal/models"
	redisc "xiuxian/server-go/internal/redis"
)

// ConnectionManager 管理所有WebSocket连接
type ConnectionManager struct {
	clients    map[uint]*ClientConnection // userID -> connection
	broadcast  chan *Message              // 广播消息
	register   chan *ClientConnection     // 注册连接
	unregister chan *ClientConnection     // 注销连接
	mu         sync.RWMutex               // 读写锁
	logger     *zap.Logger
}

// ClientConnection 代表一个客户端连接
type ClientConnection struct {
	UserID           uint
	Username         string
	conn             *websocket.Conn
	send             chan *Message
	done             chan struct{}
	manager          *ConnectionManager
	lastHeartbeat    time.Time     // 最后一次心跳时间
	heartbeatTimeout time.Duration // 心跳超时时间
}

// Message WebSocket消息格式
type Message struct {
	Type      string          `json:"type"`      // 消息类型: spirit:grow, dungeon:event等
	UserID    uint            `json:"userId"`    // 用户ID
	Timestamp int64           `json:"timestamp"` // 时间戳
	Data      json.RawMessage `json:"data"`      // 具体数据
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(logger *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		clients:    make(map[uint]*ClientConnection),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *ClientConnection),
		unregister: make(chan *ClientConnection),
		logger:     logger,
	}
}

// Start 启动连接管理器
func (cm *ConnectionManager) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case client := <-cm.register:
				cm.mu.Lock()
				cm.clients[client.UserID] = client
				cm.mu.Unlock()
				cm.logger.Info("客户端连接",
					zap.Uint("userID", client.UserID),
					zap.String("username", client.Username),
					zap.Int("totalClients", len(cm.clients)))

			case client := <-cm.unregister:
				cm.mu.Lock()
				if _, exists := cm.clients[client.UserID]; exists {
					delete(cm.clients, client.UserID)
					close(client.send)
				}
				cm.mu.Unlock()
				cm.logger.Info("客户端断开连接",
					zap.Uint("userID", client.UserID),
					zap.Int("totalClients", len(cm.clients)))

			case msg := <-cm.broadcast:
				cm.mu.RLock()
				if client, exists := cm.clients[msg.UserID]; exists {
					select {
					case client.send <- msg:
					default:
						// 发送队列满，记录警告
						cm.logger.Warn("消息队列已满，丢弃消息",
							zap.Uint("userID", msg.UserID),
							zap.String("type", msg.Type))
					}
				}
				cm.mu.RUnlock()

			case <-ctx.Done():
				cm.logger.Info("连接管理器关闭")
				return
			}
		}
	}()
}

// RegisterClient 注册客户端连接
func (cm *ConnectionManager) RegisterClient(userID uint, username string, conn *websocket.Conn) *ClientConnection {
	client := &ClientConnection{
		UserID:           userID,
		Username:         username,
		conn:             conn,
		send:             make(chan *Message, 64),
		done:             make(chan struct{}),
		manager:          cm,
		lastHeartbeat:    time.Now(),       // 初始化为当前时间
		heartbeatTimeout: 10 * time.Second, // 设置心跳10秒超时
	}

	cm.register <- client

	// 启动读写goroutine
	go client.readLoop()
	go client.writeLoop()

	return client
}

// UnregisterClient 注销客户端连接
func (cm *ConnectionManager) UnregisterClient(client *ClientConnection) {
	select {
	case cm.unregister <- client:
	default:
	}
}

// SendToUser 发送消息给指定用户
func (cm *ConnectionManager) SendToUser(userID uint, messageType string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("消息序列化失败: %w", err)
	}

	msg := &Message{
		Type:      messageType,
		UserID:    userID,
		Timestamp: time.Now().Unix(),
		Data:      dataBytes,
	}

	cm.broadcast <- msg
	return nil
}

// GetOnlineCount 获取在线玩家数
func (cm *ConnectionManager) GetOnlineCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.clients)
}

// IsUserOnline 检查用户是否在线
func (cm *ConnectionManager) IsUserOnline(userID uint) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.clients[userID]
	return exists
}

// readLoop 读取客户端消息
func (c *ClientConnection) readLoop() {
	defer func() {
		c.manager.UnregisterClient(c)
		c.conn.Close()
		close(c.done)
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.manager.logger.Error("WebSocket错误",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
			}
			return
		}

		// 处理心跳
		if msg.Type == "ping" {
			c.lastHeartbeat = time.Now() // 更新内存中的心跳时间

			// ✅ 同时更新Redis中的lastHeartbeat
			playerIDStr := strconv.FormatUint(uint64(c.UserID), 10)
			key := "player:online:" + playerIDStr
			lastHeartbeatMs := time.Now().UnixMilli()
			if err := redisc.Client.HSet(redisc.Ctx, key, "lastHeartbeat", strconv.FormatInt(lastHeartbeatMs, 10)).Err(); err != nil {
				c.manager.logger.Error("更新Redis心跳时间失败",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
			}

			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			//c.manager.logger.Debug("接收心跳",
			//	zap.Uint("userID", c.UserID))
			continue
		}

		// 记录接收的消息
		c.manager.logger.Debug("收到消息",
			zap.Uint("userID", c.UserID),
			zap.String("type", msg.Type))
	}
}

// writeLoop 向客户端发送消恫
func (c *ClientConnection) writeLoop() {
	heartbeatCheckTicker := time.NewTicker(1 * time.Second) // 每秒检查一次心跳
	defer func() {
		heartbeatCheckTicker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteJSON(msg); err != nil {
				c.manager.logger.Error("发送消恫失败",
					zap.Uint("userID", c.UserID),
					zap.Error(err))
				return
			}

		case <-heartbeatCheckTicker.C:
			// 检查是否超时
			if time.Since(c.lastHeartbeat) > c.heartbeatTimeout {
				c.manager.logger.Warn("心跳超时，正在下线玩家",
					zap.Uint("userID", c.UserID),
					zap.Duration("超时时间", time.Since(c.lastHeartbeat)))
				// 接下来自动调用/api/online/logout
				go c.performLogout()
				return
			}

		case <-c.done:
			return
		}
	}
}

// performLogout 对客户端进行下线处理
// 当心跳超时时，自动调用此方法，清下Redis中的在线数据
// 并执行类似/api/online/logout的逻辑
func (c *ClientConnection) performLogout() {
	playerIDStr := strconv.FormatUint(uint64(c.UserID), 10)

	// ✅ 先更新数据库中的LastSpiritGainTime（与Logout API逻辑一致）
	now := time.Now()
	if err := db.DB.Model(&models.User{}).Where("id = ?", c.UserID).Update("last_spirit_gain_time", now).Error; err != nil {
		c.manager.logger.Error("更新玩家灵力增长时间戳失败",
			zap.Uint("userID", c.UserID),
			zap.Error(err))
		// 不中断，继续执行下线
	}

	// 从 Redis 中移除玩家ID
	rdb := redisc.Client
	key := "player:online:" + playerIDStr

	// 更新玩家状态为offline
	if err := rdb.HSet(redisc.Ctx, key, map[string]interface{}{
		"status":     "offline",
		"logoutTime": strconv.FormatInt(time.Now().UnixMilli(), 10),
	}).Err(); err != nil {
		c.manager.logger.Error("设置Redis数据失败",
			zap.Uint("userID", c.UserID),
			zap.Error(err))
	}

	// 从在线玩家集合中移除
	// ✅ 此步不容失败，必须一定于离线前移除
	c.manager.logger.Info("此步将从在线集合中移除玩家",
		zap.Uint("userID", c.UserID),
		zap.String("playerIDStr", playerIDStr))

	if err := rdb.SRem(redisc.Ctx, "server:online:players", playerIDStr).Err(); err != nil {
		c.manager.logger.Error("从在线集合中移除失败",
			zap.Uint("userID", c.UserID),
			zap.String("playerIDStr", playerIDStr),
			zap.Error(err))
	} else {
		// ✅ 成功离线
		c.manager.logger.Info("成功从在线集合中移除",
			zap.Uint("userID", c.UserID),
			zap.String("playerIDStr", playerIDStr))
	}

	c.manager.logger.Info("心跳超时自动下线",
		zap.Uint("userID", c.UserID),
		zap.String("username", c.Username))
}
