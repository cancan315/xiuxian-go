# WebSocket客户端集成

<cite>
**本文档引用文件**  
- [useWebSocket.js](file://src/composables/useWebSocket.js)
- [websocket.js](file://src/services/websocket.js)
- [playerInfo.js](file://src/stores/playerInfo.js)
- [LogPanel.vue](file://src/components/LogPanel.vue)
- [Cultivation.vue](file://src/views/Cultivation.vue)
- [Exploration.vue](file://src/views/Exploration.vue)
- [handler.go](file://server-go/internal/websocket/handler.go)
- [manager.go](file://server-go/internal/websocket/manager.go)
- [spirit_handler.go](file://server-go/internal/websocket/spirit_handler.go)
- [dungeon_handler.go](file://server-go/internal/websocket/dungeon_handler.go)
- [exploration_handler.go](file://server-go/internal/websocket/exploration_handler.go)
</cite>

## 目录
1. [引言](#引言)
2. [WebSocket连接管理](#websocket连接管理)
3. [消息处理与事件订阅](#消息处理与事件订阅)
4. [与Pinia状态管理器的集成](#与pinia状态管理器的集成)
5. [消息序列化格式](#消息序列化格式)
6. [错误处理机制](#错误处理机制)
7. [实际组件集成示例](#实际组件集成示例)
8. [心跳与连接保活](#心跳与连接保活)
9. [重连机制](#重连机制)
10. [总结](#总结)

## 引言
本项目通过WebSocket实现实时通信，为修仙游戏提供灵力增长、战斗事件、探索进度等实时更新功能。前端通过`useWebSocket.js`组合式函数封装WebSocket连接的创建、重连、消息监听与发送，与Pinia状态管理器（如`playerInfo.js`）集成，将实时消息自动更新到全局状态并触发UI响应。本文档详细说明WebSocket客户端的实现机制。

## WebSocket连接管理
前端WebSocket连接由`useWebSocket.js`组合式函数管理，该函数封装了连接的创建、状态监控和断开逻辑。连接通过`wsManager.connect()`方法建立，需要提供用户认证token和用户ID。

```mermaid
flowchart TD
A[初始化WebSocket] --> B{是否已连接}
B --> |是| C[跳过连接]
B --> |否| D[构建WebSocket URL]
D --> E[创建WebSocket连接]
E --> F[设置事件监听器]
F --> G[连接成功]
G --> H[启动心跳]
H --> I[更新连接状态]
```

**Diagram sources**  
- [useWebSocket.js](file://src/composables/useWebSocket.js#L21-L32)
- [websocket.js](file://src/services/websocket.js#L36-L67)

**Section sources**  
- [useWebSocket.js](file://src/composables/useWebSocket.js#L10-L135)
- [websocket.js](file://src/services/websocket.js#L4-L267)

## 消息处理与事件订阅
WebSocket客户端支持多种事件类型的订阅，包括灵力增长、战斗事件、排行榜更新和探索事件。消息处理采用发布-订阅模式，通过`wsManager.on()`方法注册事件监听器。

```mermaid
classDiagram
class WebSocketManager {
+listeners : Map[string, Function[]]
+on(type, callback)
+off(type, callback)
+handleMessage(message)
+emit(type, data)
}
class useWebSocket {
+subscribeSpiritGrowthData(callback)
+subscribeDungeonEventData(callback)
+subscribeLeaderboardUpdateData(callback)
+subscribeExplorationEventData(callback)
}
WebSocketManager --> useWebSocket : "提供事件订阅接口"
```

**Diagram sources**  
- [websocket.js](file://src/services/websocket.js#L15-L252)
- [useWebSocket.js](file://src/composables/useWebSocket.js#L38-L73)

**Section sources**  
- [websocket.js](file://src/services/websocket.js#L102-L127)
- [useWebSocket.js](file://src/composables/useWebSocket.js#L38-L73)

## 与Pinia状态管理器的集成
WebSocket客户端与Pinia状态管理器深度集成，通过订阅实时消息自动更新全局状态。以`playerInfo.js`为例，灵力增长事件会自动更新玩家的灵力值。

```mermaid
sequenceDiagram
participant Server as "服务器"
participant WebSocket as "WebSocket客户端"
participant Pinia as "Pinia状态管理器"
participant UI as "用户界面"
Server->>WebSocket : 发送灵力增长消息
WebSocket->>WebSocket : handleMessage()
WebSocket->>Pinia : 更新playerInfoStore.spirit
Pinia->>UI : 触发响应式更新
UI->>UI : 重新渲染灵力显示
```

**Diagram sources**  
- [useWebSocket.js](file://src/composables/useWebSocket.js#L38-L42)
- [playerInfo.js](file://src/stores/playerInfo.js#L16-L17)

**Section sources**  
- [useWebSocket.js](file://src/composables/useWebSocket.js#L146-L158)
- [playerInfo.js](file://src/stores/playerInfo.js#L6-L65)

## 消息序列化格式
WebSocket消息采用JSON格式进行序列化，包含消息类型、用户ID、时间戳和具体数据。后端定义了多种消息类型，每种类型有特定的数据结构。

```mermaid
erDiagram
MESSAGE {
string type PK
uint userId
int64 timestamp
json data
}
SPIRIT_GROWTH {
float64 oldSpirit
float64 newSpirit
float64 gainAmount
float64 spiritRate
float64 elapsedSeconds
}
DUNGEON_EVENT {
string eventType
string dungeon
string message
int roundNum
float64 playerHp
float64 enemyHp
float64 damageDealt
float64 damageTaken
json loot
}
EXPLORATION_EVENT {
string eventType
string exploreName
string message
int progress
int durationSecs
int elapsedSecs
json discovery
json reward
string errorMsg
}
MESSAGE ||--o{ SPIRIT_GROWTH : "spirit:grow"
MESSAGE ||--o{ DUNGEON_EVENT : "dungeon:event"
MESSAGE ||--o{ EXPLORATION_EVENT : "exploration:event"
```

**Diagram sources**  
- [manager.go](file://server-go/internal/websocket/manager.go#L41-L47)
- [spirit_handler.go](file://server-go/internal/websocket/spirit_handler.go#L18-L26)
- [dungeon_handler.go](file://server-go/internal/websocket/dungeon_handler.go#L10-L21)
- [exploration_handler.go](file://server-go/internal/websocket/exploration_handler.go#L10-L21)

**Section sources**  
- [manager.go](file://server-go/internal/websocket/manager.go#L41-L47)
- [spirit_handler.go](file://server-go/internal/websocket/spirit_handler.go#L18-L26)

## 错误处理机制
WebSocket客户端实现了完善的错误处理机制，包括连接失败、消息处理异常和网络中断等情况。错误通过事件系统通知上层组件。

```mermaid
flowchart TD
A[WebSocket错误] --> B{错误类型}
B --> |连接错误| C[触发connection:error事件]
B --> |消息处理错误| D[记录错误日志]
B --> |发送失败| E[返回false状态]
C --> F[上层组件处理错误]
D --> G[继续正常运行]
E --> H[上层组件重试或提示]
```

**Diagram sources**  
- [websocket.js](file://src/services/websocket.js#L73-L76)
- [websocket.js](file://src/services/websocket.js#L121-L123)
- [websocket.js](file://src/services/websocket.js#L168-L171)

**Section sources**  
- [websocket.js](file://src/services/websocket.js#L73-L76)
- [websocket.js](file://src/services/websocket.js#L121-L123)

## 实际组件集成示例
`LogPanel.vue`组件展示了如何订阅特定事件类型并渲染实时日志。该组件使用Web Worker处理日志，避免阻塞主线程。

```mermaid
sequenceDiagram
participant LogPanel as "LogPanel.vue"
participant Worker as "log.js Worker"
participant WebSocket as "WebSocket客户端"
WebSocket->>LogPanel : 接收消息
LogPanel->>Worker : postMessage(ADD_LOG)
Worker->>Worker : 添加日志到数组
Worker->>Worker : 限制日志数量
Worker->>LogPanel : postMessage(LOGS_UPDATED)
LogPanel->>LogPanel : 更新UI显示
```

**Diagram sources**  
- [LogPanel.vue](file://src/components/LogPanel.vue#L47-L62)
- [log.js](file://src/workers/log.js#L7-L21)

**Section sources**  
- [LogPanel.vue](file://src/components/LogPanel.vue#L25-L104)
- [log.js](file://src/workers/log.js#L1-L56)

## 心跳与连接保活
WebSocket客户端实现了心跳机制，每秒向服务器发送心跳包，确保连接活跃。服务器端会检查心跳超时，自动下线长时间无响应的客户端。

```mermaid
sequenceDiagram
participant Client as "客户端"
participant Server as "服务器"
loop 每秒
Client->>Server : ping消息
Server->>Client : pong响应
Server->>Server : 更新lastHeartbeat
end
Server->>Server : 每秒检查心跳超时
Server->>Server : 超时则执行下线逻辑
```

**Diagram sources**  
- [websocket.js](file://src/services/websocket.js#L177-L190)
- [manager.go](file://server-go/internal/websocket/manager.go#L196-L212)
- [manager.go](file://server-go/internal/websocket/manager.go#L241-L250)

**Section sources**  
- [websocket.js](file://src/services/websocket.js#L177-L205)
- [manager.go](file://server-go/internal/websocket/manager.go#L196-L212)

## 重连机制
当WebSocket连接意外断开时，客户端会自动尝试重连，最多尝试5次，每次重试间隔逐渐增加。主动断开不会触发自动重连。

```mermaid
flowchart TD
A[连接关闭] --> B{是否主动断开}
B --> |是| C[不重连]
B --> |否| D{重连次数 < 最大次数}
D --> |是| E[计算重连延迟]
E --> F[设置定时器重连]
F --> G[连接成功?]
G --> |是| H[停止重连]
G --> |否| I[增加重连次数]
I --> D
D --> |否| J[停止重连]
```

**Diagram sources**  
- [websocket.js](file://src/services/websocket.js#L84-L87)
- [websocket.js](file://src/services/websocket.js#L210-L223)

**Section sources**  
- [websocket.js](file://src/services/websocket.js#L84-L87)
- [websocket.js](file://src/services/websocket.js#L210-L223)

## 总结
本项目的WebSocket客户端实现了一个健壮的实时通信系统，通过组合式函数封装了连接管理、消息处理和状态更新。与Pinia状态管理器的深度集成确保了实时数据能够自动更新到全局状态，触发UI响应。心跳和重连机制保证了连接的稳定性，错误处理机制提高了系统的可靠性。通过Web Worker处理日志等耗时操作，避免了主线程阻塞，提升了用户体验。