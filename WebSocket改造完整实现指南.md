# WebSocket 改造实现完整指南

## 一、项目结构总结

### 后端 WebSocket 模块（Go）

```
server-go/internal/websocket/
├── manager.go              # 连接管理器 - 处理WebSocket连接生命周期
├── handler.go              # 处理器 - HTTP升级为WebSocket
├── message.go             # 消息定义（可选）
├── router.go              # 路由注册
├── spirit_handler.go      # 灵力增长事件处理
├── dungeon_handler.go     # 战斗事件处理
├── leaderboard_handler.go # 排行榜更新处理
└── exploration_handler.go # 探索事件处理
```

### 前端 WebSocket 模块（Vue3）

```
src/
├── services/
│   └── websocket.js       # WebSocket管理器
├── composables/
│   └── useWebSocket.js    # Vue3组合式API
└── components/
    └── WebSocketDebug.vue # 调试面板
```

---

## 二、后端集成步骤

### Step 1: 依赖检查

```bash
cd server-go
go get github.com/gorilla/websocket
```

### Step 2: 修改 main.go

已在 `cmd/server/main.go` 中添加：

```go
import (
    "xiuxian/server-go/internal/websocket"
    "context"
)

// 在初始化logger之后添加
wsManager := websocket.NewConnectionManager(logger)
ctx := context.Background()
wsManager.Start(ctx)

// 初始化事件处理器
wsHandlers := websocket.InitializeHandlers(wsManager, logger)

// 注册WebSocket路由
websocket.RegisterWebSocketRoutes(r, wsManager, logger)

// 注入到上下文
r.Use(func(c *gin.Context) {
    c.Set("ws_manager", wsManager)
    c.Set("ws_handlers", wsHandlers)
    c.Next()
})
```

### Step 3: 灵力增长后台任务与WebSocket集成

在 `internal/spirit/spirit_grow.go` 的 `calculateAndUpdateSpiritGain()` 方法中添加：

```go
// 获取WebSocket处理器并发送实时更新
if handlers, exists := c.Get("ws_handlers"); exists {
    if h, ok := handlers.(*websocket.Handlers); ok {
        h.Spirit.NotifySpiritUpdate(
            user.ID, 
            oldSpirit, 
            user.Spirit, 
            spiritRate, 
            elapsedSeconds,
        )
    }
}
```

### Step 4: 编译验证

```bash
cd server-go
go build -o bin/server ./cmd/server

# 或使用 air 热重载
air
```

---

## 三、前端集成步骤

### Step 1: 安装dependencies

```bash
# WebSocket 库已在 package.json 中
npm install
```

### Step 2: 在主应用中初始化WebSocket

修改 `src/App.vue` 的 setup 函数：

```javascript
import { useWebSocket } from '@/composables/useWebSocket'

export default {
  async setup() {
    const ws = useWebSocket()
    const userStore = useUserStore()
    
    // 登录成功后初始化WebSocket
    onMounted(async () => {
      if (userStore.token && userStore.userId) {
        try {
          await ws.initWebSocket(userStore.token, userStore.userId)
          
          // 订阅灵力增长
          ws.subscribeSpiritGrowthData((data) => {
            userStore.spirit = data.newSpirit
            console.log(`灵力增长: +${data.gainAmount}`)
          })
          
          // 订阅战斗事件
          ws.subscribeDungeonEventData((data) => {
            console.log(`战斗: ${data.message}`)
          })
          
          // 订阅排行榜更新
          ws.subscribeLeaderboardUpdateData((data) => {
            console.log(`排行榜更新:`, data)
          })
          
          // 订阅探索事件
          ws.subscribeExplorationEventData((data) => {
            console.log(`探索: ${data.message}`)
          })
        } catch (error) {
          console.error('WebSocket初始化失败:', error)
        }
      }
    })
    
    // 离线时断开连接
    onUnmounted(() => {
      ws.disconnect()
    })
    
    return { ws }
  }
}
```

### Step 3: 在页面中使用 composables

```javascript
// 在任何组件中使用
import { useWebSocket, useSpiritGrowth, useDungeonCombat, useLeaderboard, useExploration } from '@/composables/useWebSocket'

export default {
  setup() {
    const ws = useWebSocket()
    const spirit = useSpiritGrowth()
    const combat = useDungeonCombat()
    const leaderboard = useLeaderboard()
    const exploration = useExploration()
    
    // 订阅灵力增长
    ws.subscribeSpiritGrowthData(spirit.handleSpiritGrowth)
    
    // 订阅战斗事件
    ws.subscribeDungeonEventData(combat.handleDungeonEvent)
    
    // 订阅排行榜更新
    ws.subscribeLeaderboardUpdateData(leaderboard.handleLeaderboardUpdate)
    
    // 订阅探索事件
    ws.subscribeExplorationEventData(exploration.handleExplorationEvent)
    
    return {
      spirit,
      combat,
      leaderboard,
      exploration
    }
  }
}
```

---

## 四、API 端点

### WebSocket 端点

```
ws://localhost:3000/ws?userId={userId}&token={token}
```

**连接参数：**
- `userId` (必需): 用户ID
- `token` (必需): 认证token

### 消息格式

#### 灵力增长（spirit:grow）

服务器 → 客户端：
```json
{
  "type": "spirit:grow",
  "userId": 1,
  "timestamp": 1702489200,
  "data": {
    "userId": 1,
    "oldSpirit": 100.0,
    "newSpirit": 115.03,
    "gainAmount": 15.03,
    "spiritRate": 1.5,
    "elapsedSeconds": 10.0,
    "timestamp": 1702489200
  }
}
```

#### 战斗事件（dungeon:event）

服务器 → 客户端：
```json
{
  "type": "dungeon:event",
  "userId": 1,
  "timestamp": 1702489200,
  "data": {
    "userId": 1,
    "eventType": "combat_round",
    "dungeon": "魔渊秘境",
    "message": "攻击成功！",
    "roundNum": 3,
    "playerHp": 250.0,
    "enemyHp": 150.0,
    "damageDealt": 50.0,
    "damageTaken": 30.0,
    "timestamp": 1702489200
  }
}
```

#### 排行榜更新（leaderboard:update）

服务器 → 客户端：
```json
{
  "type": "leaderboard:update",
  "userId": 1,
  "timestamp": 1702489200,
  "data": {
    "type": "full_refresh",
    "category": "spirit",
    "updateTime": 1702489200,
    "top10": [
      {
        "rank": 1,
        "userId": 2,
        "username": "玩家2",
        "spirit": 5000.0,
        "power": 3500.0,
        "level": 10
      }
    ],
    "userRank": {
      "rank": 5,
      "value": 1500.0,
      "percent": 50.0
    },
    "timestamp": 1702489200
  }
}
```

#### 探索事件（exploration:event）

服务器 → 客户端：
```json
{
  "type": "exploration:event",
  "userId": 1,
  "timestamp": 1702489200,
  "data": {
    "userId": 1,
    "eventType": "progress",
    "exploreName": "古老遗迹",
    "message": "正在探索中...",
    "progress": 50,
    "durationSecs": 300,
    "elapsedSecs": 150,
    "timestamp": 1702489200
  }
}
```

### 心跳机制

客户端 → 服务器（每20秒）：
```json
{
  "type": "ping",
  "timestamp": 1702489200
}
```

---

## 五、使用示例

### 后端发送灵力增长事件

```go
// 在控制器或服务中获取handlers
handlers := c.MustGet("ws_handlers").(*websocket.Handlers)

// 发送灵力增长事件
handlers.Spirit.NotifySpiritUpdate(
    userId,
    oldSpirit,
    newSpirit,
    spiritRate,
    elapsedSeconds,
)
```

### 后端发送战斗事件

```go
handlers := c.MustGet("ws_handlers").(*websocket.Handlers)

// 战斗开始
handlers.Dungeon.NotifyDungeonStart(userId, "魔渊秘境")

// 战斗轮次
handlers.Dungeon.NotifyCombatRound(
    userId, 
    "魔渊秘境", 
    roundNum,
    playerHP,
    enemyHP,
    damageDealt,
    damageTaken,
)

// 战斗胜利
handlers.Dungeon.NotifyVictory(userId, "魔渊秘境", loot)
```

### 前端处理灵力增长

```javascript
const ws = useWebSocket()
const spirit = useSpiritGrowth()

// 订阅灵力增长
ws.subscribeSpiritGrowthData((data, meta) => {
  console.log(`灵力增长: ${data.gainAmount}`)
  spirit.handleSpiritGrowth(data)
  
  // 更新UI
  userStore.spirit = data.newSpirit
})

// 获取事件历史
console.log(spirit.spiritGrowthEvents) // 最后100条记录
console.log(spirit.totalGainedSpirit)   // 总获得灵力
```

### 前端处理战斗事件

```javascript
const ws = useWebSocket()
const combat = useDungeonCombat()

ws.subscribeDungeonEventData((data) => {
  combat.handleDungeonEvent(data)
  
  // 根据事件类型更新UI
  switch(data.eventType) {
    case 'start':
      showDungeonBattle()
      break
    case 'combat_round':
      updateCombatUI(data)
      break
    case 'victory':
      showVictoryModal(data.loot)
      break
    case 'defeat':
      showDefeatModal()
      break
  }
})
```

---

## 六、部署说明

### 开发环境

```bash
# 后端
cd server-go
air  # 热重载开发

# 前端
npm run dev  # Vite开发服务器
```

### 生产环境

```bash
# 后端编译
cd server-go
go build -o bin/server ./cmd/server

# 配置环境变量
export LOG_LEVEL=info
export PORT=3000
export DATABASE_URL="..."
export REDIS_URL="..."

# 启动服务
./bin/server

# 前端构建
npm run build

# Nginx配置（参考现有的nginx.conf）
# 需要配置WebSocket代理
location /ws {
    proxy_pass http://backend:3000;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

### Nginx WebSocket 配置

```nginx
upstream backend {
    server localhost:3000;
}

server {
    listen 80;
    server_name _;

    # 静态文件
    location / {
        root /app/dist;
        try_files $uri $uri/ /index.html;
    }

    # API代理
    location /api {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket代理
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # WebSocket超时设置
        proxy_read_timeout 86400;
        proxy_send_timeout 86400;
    }
}
```

---

## 七、监控和调试

### 启用WebSocket调试面板

在需要调试的页面中添加：

```vue
<template>
  <div>
    <!-- 主要内容 -->
    <MainContent />
    
    <!-- 调试面板（仅在开发环境显示） -->
    <WebSocketDebug v-if="isDevelopment" />
  </div>
</template>

<script setup>
import WebSocketDebug from '@/components/WebSocketDebug.vue'
import { useWebSocket } from '@/composables/useWebSocket'

const isDevelopment = import.meta.env.DEV
const ws = useWebSocket()
</script>
```

### 查看日志

后端日志：
```bash
# 实时日志
tail -f logs/server.log | grep "WebSocket\|spirit\|dungeon"

# 查看特定事件
grep "灵力增长事件已发送" logs/server.log
```

前端日志（浏览器控制台）：
```javascript
// 查看所有WebSocket消息
localStorage.debug = 'websocket:*'

// 或者在应用中启用debug
import { wsManager } from '@/services/websocket'
window.wsManager = wsManager // 暴露到全局作用域
```

---

## 八、常见问题

### Q1: 连接失败 "WebSocket握手失败"

**原因：** 
- 后端WebSocket路由未正确注册
- 用户ID或token参数缺失或格式错误

**解决：**
```javascript
// 确保连接URL格式正确
const ws = new WebSocket(
  `ws://localhost:3000/ws?userId=1&token=xxx_valid_token_xxx`
)
```

### Q2: 消息无法接收

**原因：**
- 连接未建立（检查isConnected状态）
- 消息处理器未订阅
- 服务器未正确发送消息

**解决：**
```javascript
// 确保连接成功
console.log(wsManager.isConnected)

// 确保订阅了事件
ws.subscribeSpiritGrowth((data) => {
  console.log('收到灵力更新:', data)
})
```

### Q3: 内存泄漏 - 连接不释放

**原因：** 未正确关闭连接或清理事件监听器

**解决：**
```javascript
// 组件卸载时清理
onUnmounted(() => {
  // 取消订阅
  unsubscribe()
  
  // 断开连接
  ws.disconnect()
})
```

### Q4: Nginx反向代理下WebSocket连接失败

**原因：** Nginx未配置WebSocket升级

**解决：**
```nginx
location /ws {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

---

## 九、性能优化

### 后端优化

```go
// 1. 连接管理器使用缓冲通道，避免阻塞
broadcast:  make(chan *Message, 256),   // 256条消息缓冲
register:   make(chan *ClientConnection),
unregister: make(chan *ClientConnection),

// 2. 消息发送使用non-blocking select
select {
case client.send <- msg:
    // 消息已发送
default:
    // 队列满，丢弃并记录日志
}

// 3. 定期清理断开的连接
// ConnectionManager中的清理逻辑已内置
```

### 前端优化

```javascript
// 1. 事件监听的及时清理
const unsubscribe = ws.subscribeSpiritGrowth((data) => {
  // 处理数据
})

// 2. 避免频繁的UI更新
// 使用防抖（debounce）处理高频更新
const debouncedUpdate = debounce((data) => {
  updateUI(data)
}, 100)

// 3. 仅保存必要的历史记录
// 默认保留最后100条记录
```

---

## 十、安全建议

1. **认证和授权**
   - 验证token的有效性
   - 确保用户只接收自己的消息

2. **消息验证**
   - 验证userID与token匹配
   - 验证消息格式和数据类型

3. **速率限制**
   - 限制单个连接的消息发送频率
   - 防止消息队列溢出

4. **加密传输**
   - 生产环境使用 wss:// (WebSocket Secure)
   - 配置SSL/TLS证书

---

## 总结

WebSocket改造完成后的架构：

```
┌─────────────────────────────────────────┐
│           前端（Vue3）                   │
│  ┌──────────────────────────────────┐   │
│  │  WebSocket 管理器 (websocket.js)  │   │
│  │  - 连接管理                      │   │
│  │  - 消息路由                      │   │
│  │  - 心跳机制                      │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  Composables (useWebSocket.js)    │   │
│  │  - useSpiritGrowth                │   │
│  │  - useDungeonCombat               │   │
│  │  - useLeaderboard                 │   │
│  │  - useExploration                 │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
                    ↕ WebSocket
┌─────────────────────────────────────────┐
│           后端（Go）                     │
│  ┌──────────────────────────────────┐   │
│  │ WebSocket 连接管理器 (manager.go) │   │
│  │ - 连接池管理                     │   │
│  │ - 消息路由                       │   │
│  │ - 心跳/ping-pong                 │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  事件处理器                       │   │
│  │ - SpiritHandler (灵力增长)       │   │
│  │ - DungeonHandler (战斗事件)      │   │
│  │ - LeaderboardHandler (排行榜)    │   │
│  │ - ExplorationHandler (探索事件)  │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  后台任务集成                     │   │
│  │ - 灵力增长任务 (spirit_grow.go)  │   │
│  │ - 排行榜更新任务                 │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
                    ↕ REST API
┌─────────────────────────────────────────┐
│         数据库 & 缓存                     │
│  - PostgreSQL (持久化)                  │
│  - Redis (缓存/在线状态)               │
└─────────────────────────────────────────┘
```

所有代码已生成完毕！按照此指南进行集成即可完成WebSocket改造。
