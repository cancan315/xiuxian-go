// src/services/websocket.js
// WebSocket 连接管理和消息处理

class WebSocketManager {
  constructor() {
    this.ws = null;
    this.url = null;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 3000;
    this.isIntentionallyClosed = false;  // ✅ 标记是否主动断开
    
    // 事件监听器存储
    this.listeners = {
      'spirit:grow': [],
      'dungeon:event': [],
      'leaderboard:update': [],
      'exploration:event': [],
      'connection:open': [],
      'connection:close': [],
      'connection:error': []
    };
    
    // 心跳相关
    this.heartbeatInterval = null;
    this.heartbeatTimeout = null;
  }

  /**
   * 连接到WebSocket服务器
   * @param {string} token - 认证token
   * @param {number} userId - 用户ID
   * @param {string} serverUrl - WebSocket服务器地址，默认为当前origin
   */
  connect(token, userId, serverUrl = null) {
    if (this.isConnected) {
      console.warn('[WebSocket] 已经连接，请勿重复连接');
      return Promise.resolve();
    }
    
    // ✅ 新连接时重置主动断开标记
    this.isIntentionallyClosed = false;

    return new Promise((resolve, reject) => {
      try {
        // 构建WebSocket URL
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        // ✅ 使用 hostname 而不是 host，避免端口重复
        // window.location.host 包含端口（例如 localhost:2025）
        // window.location.hostname 仅是主机名（例如 localhost）
        const hostname = window.location.hostname;
        // WebSocket 指向后端 (port 3000)
        this.url = serverUrl || `${protocol}//${hostname}:3000/ws?userId=${userId}&token=${token}`;

        console.log('[WebSocket] 正在连接到: ' + this.url);

        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
          console.log('[WebSocket] 连接成功');
          this.isConnected = true;
          this.reconnectAttempts = 0;
          this.startHeartbeat();
          this.emit('connection:open', { userId, timestamp: Date.now() });
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(JSON.parse(event.data));
        };

        this.ws.onerror = (error) => {
          console.error('[WebSocket] 连接错误:', error);
          this.emit('connection:error', error);
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('[WebSocket] 连接关闭');
          this.isConnected = false;
          this.stopHeartbeat();
          this.emit('connection:close', { timestamp: Date.now() });
          // ✅ 仅当不是主动断开时，才进行自动重连
          if (!this.isIntentionallyClosed) {
            this.attemptReconnect(token, userId, serverUrl);
          } else {
            console.log('[WebSocket] 主动断开，不进行重连');
          }
        };

      } catch (error) {
        console.error('[WebSocket] 连接失败:', error);
        reject(error);
      }
    });
  }

  /**
   * 处理来自服务器的消息
   */
  handleMessage(message) {
    const { type, userId, data, timestamp } = message;
    
    //console.log(`[WebSocket] 收到消息: ${type}`, { userId, data, timestamp });
    //console.log(`[WebSocket] 监听器状态:`, {
    //  type,
    //  hasListener: !!this.listeners[type],
     // listenerCount: this.listeners[type]?.length || 0,
    //  allListenerTypes: Object.keys(this.listeners)
    //});

    // 触发对应类型的监听器
    if (this.listeners[type]) {
      // console.log(`[WebSocket] 触发 ${this.listeners[type].length} 个监听器 (${type})`);
      this.listeners[type].forEach((callback, index) => {
        try {
          // console.log(`[WebSocket] 执行监听器 #${index + 1} (${type})`);
          callback(data, { userId, timestamp });
          // console.log(`[WebSocket] 监听器 #${index + 1} 执行完成 (${type})`);
        } catch (error) {
          console.error(`[WebSocket] 消息处理器错误 (${type}):`, error);
        }
      });
    } else {
      console.warn(`[WebSocket] 未知消息类型: ${type}`, { availableTypes: Object.keys(this.listeners) });
    }
  }

  /**
   * 订阅特定类型的消息
   * @param {string} type - 消息类型
   * @param {function} callback - 回调函数
   */
  on(type, callback) {
    if (!this.listeners[type]) {
      this.listeners[type] = [];
    }
    this.listeners[type].push(callback);

    // 返回取消订阅函数
    return () => {
      this.off(type, callback);
    };
  }

  /**
   * 取消订阅
   */
  off(type, callback) {
    if (this.listeners[type]) {
      this.listeners[type] = this.listeners[type].filter(cb => cb !== callback);
    }
  }

  /**
   * 发送消息到服务器
   */
  send(message) {
    if (!this.isConnected) {
      console.error('[WebSocket] 连接未建立，无法发送消息');
      return false;
    }

    try {
      this.ws.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error('[WebSocket] 发送消息失败:', error);
      return false;
    }
  }

  /**
   * 发送心跳包
   */
  sendHeartbeat() {
    if (this.isConnected) {
      this.send({ type: 'ping', timestamp: Date.now() });
    }
  }

  /**
   * 开始心跳
   */
  startHeartbeat() {
    // 每秒发送一次心跳
    this.heartbeatInterval = setInterval(() => {
      this.sendHeartbeat();
    }, 1000);
  }

  /**
   * 停止心跳
   */
  stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
    if (this.heartbeatTimeout) {
      clearTimeout(this.heartbeatTimeout);
      this.heartbeatTimeout = null;
    }
  }

  /**
   * 尝试重新连接
   */
  attemptReconnect(token, userId, serverUrl) {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * this.reconnectAttempts;
      console.log(`[WebSocket] 将在 ${delay}ms 后进行第 ${this.reconnectAttempts} 次重连`);
      
      setTimeout(() => {
        this.connect(token, userId, serverUrl)
          .catch(error => console.error('[WebSocket] 重连失败:', error));
      }, delay);
    } else {
      console.error('[WebSocket] 达到最大重连次数，停止重连');
    }
  }

  /**
   * 断开连接
   */
  disconnect() {
    console.log('[WebSocket] 执行主动断开');
    this.isIntentionallyClosed = true;  // ✅ 标记主动断开
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
  }

  /**
   * 触发事件
   */
  emit(type, data) {
    if (this.listeners[type]) {
      this.listeners[type].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error(`[WebSocket] 事件处理器错误 (${type}):`, error);
        }
      });
    }
  }

  /**
   * 获取连接状态
   */
  getConnectionStatus() {
    return {
      isConnected: this.isConnected,
      url: this.url,
      reconnectAttempts: this.reconnectAttempts
    };
  }
}

// 导出单例
export const wsManager = new WebSocketManager();

/**
 * 灵力增长订阅
 */
export function subscribeSpiritGrowth(callback) {
  return wsManager.on('spirit:grow', callback);
}

/**
 * 战斗事件订阅
 */
export function subscribeDungeonEvent(callback) {
  return wsManager.on('dungeon:event', callback);
}

/**
 * 排行榜更新订阅
 */
export function subscribeLeaderboardUpdate(callback) {
  return wsManager.on('leaderboard:update', callback);
}

/**
 * 探索事件订阅
 */
export function subscribeExplorationEvent(callback) {
  return wsManager.on('exploration:event', callback);
}

/**
 * 连接状态订阅
 */
export function subscribeConnectionEvent(callback) {
  wsManager.on('connection:open', (data) => {
    callback({ status: 'open', data });
  });
  wsManager.on('connection:close', (data) => {
    callback({ status: 'close', data });
  });
  wsManager.on('connection:error', (error) => {
    callback({ status: 'error', error });
  });
}

export default wsManager;
