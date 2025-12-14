# 心跳检测与自动下线功能实现指南

## 功能概述

实现了完整的WebSocket心跳检测机制，前端每秒发送心跳信号，后端10秒未收到心跳后自动调用下线逻辑。

## 架构设计

### 前端部分

**文件**: `src/services/websocket.js`

#### 心跳发送机制
```javascript
// 每秒发送一次心跳信号
startHeartbeat() {
  this.heartbeatInterval = setInterval(() => {
    this.sendHeartbeat();  // 发送 { type: 'ping', timestamp: ... }
  }, 1000);  // 改为1000ms（1秒）
}
```

**特点**:
- ✅ 前端每秒发送一次 `ping` 消息
- ✅ 自动建立连接时启动（第56行）
- ✅ 断开连接时停止

### 后端部分

**文件**: `internal/websocket/manager.go`

#### 1. ClientConnection结构扩展
```go
type ClientConnection struct {
    UserID            uint
    Username          string
    conn              *websocket.Conn
    send              chan *Message
    done              chan struct{}
    manager           *ConnectionManager
    lastHeartbeat     time.Time      // ✨ 记录最后心跳时间
    heartbeatTimeout  time.Duration  // ✨ 心跳超时阈值（10秒）
}
```

#### 2. 客户端连接注册
```go
// 初始化时设置心跳超时为10秒
client.lastHeartbeat = time.Now()
client.heartbeatTimeout = 10 * time.Second
```

#### 3. 读取心跳消息
在 `readLoop()` 中处理心跳：
```go
if msg.Type == "ping" {
    c.lastHeartbeat = time.Now()  // 更新最后心跳时间
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    continue
}
```

#### 4. 心跳超时检测
在 `writeLoop()` 中每秒检查一次：
```go
heartbeatCheckTicker := time.NewTicker(1 * time.Second)

select {
case <-heartbeatCheckTicker.C:
    // 检查是否超时（10秒未收到心跳）
    if time.Since(c.lastHeartbeat) > c.heartbeatTimeout {
        // 自动下线处理
        go c.performLogout()
        return
    }
}
```

#### 5. 自动下线逻辑
```go
func (c *ClientConnection) performLogout() {
    // 1. 从Redis中移除玩家ID
    rdb.SRem(ctx, "server:online:players", playerIDStr)
    
    // 2. 更新玩家状态为离线
    rdb.HSet(ctx, "player:online:"+playerIDStr, map[string]interface{}{
        "status": "offline",
        "logoutTime": timestamp,
    })
    
    // 3. 记录日志
    logger.Info("心跳超时自动下线", userID, username)
}
```

## 工作流程

```
┌─────────────────────────────────────────────────────────┐
│                     前端 (WebSocket)                     │
├─────────────────────────────────────────────────────────┤
│  每秒发送 ping 信号:                                    │
│  { type: "ping", timestamp: Date.now() }               │
└──────────────────────┬──────────────────────────────────┘
                       │
                       │ WebSocket Message
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   后端 (Golang)                         │
├─────────────────────────────────────────────────────────┤
│  readLoop():                                            │
│  - 接收 ping 消息                                       │
│  - 更新 lastHeartbeat = now()                          │
│  - 继续处理下一个消息                                   │
│                                                         │
│  writeLoop():                                           │
│  - 每秒检查一次: time.Since(lastHeartbeat)             │
│  - 如果超过10秒，执行 performLogout()                  │
│    ├── 从Redis集合移除玩家ID                           │
│    ├── 标记状态为offline                               │
│    └── 记录日志并断开连接                               │
└─────────────────────────────────────────────────────────┘
                       │
                       ▼
              灵力增长任务不再处理该玩家
              (Redis中没有此玩家ID)
```

## 时间线

| 时间点 | 事件 | 说明 |
|--------|------|------|
| T=0s | 玩家登录 | WebSocket连接建立，lastHeartbeat=now |
| T=1s | 前端发送ping | 后端接收，lastHeartbeat更新 |
| T=2s | 前端发送ping | 后端接收，lastHeartbeat更新 |
| ... | ... | ... |
| T=10s | 最后一次收到ping | lastHeartbeat=T=10s |
| T=11-20s | 未收到ping | 后端检查: Since(lastHeartbeat)=0-10s ✓ |
| T=20s | 检查超时 | Since(lastHeartbeat)=10s，触发下线 ✗ |
| T=20s+ | 自动下线 | performLogout()执行 |

## 自动下线后的影响

✅ **Redis操作**：
- 从 `server:online:players` 集合中移除玩家ID
- 标记 `player:online:{id}` 的状态为 `offline`
- 灵力增长任务无法再找到该玩家

✅ **前端表现**：
- WebSocket连接断开
- 前端触发 `connection:close` 事件
- 自动重连（重连5次，每次延迟递增）

✅ **数据保留**：
- 玩家数据库记录保留
- 灵力值保留（不再增长）
- 可以再次登录恢复

## 测试步骤

### 1. 启动后端
```bash
cd server-go
go run ./cmd/server
```

### 2. 创建前端测试脚本
```javascript
// 10秒钟无心跳测试
const ws = new WebSocket('ws://localhost:3000/ws?userId=1&token=...');

ws.onopen = () => {
  console.log('连接成功');
  // 不发送任何ping，等待10秒后被自动下线
};

ws.onclose = () => {
  console.log('连接关闭（10秒后）');
};
```

### 3. 观察后端日志
```
"心跳超时，正在下线" userID=1
"从在线集合中移除失败" 或 "心跳超时自动下线" userID=1
```

## 配置参数

### 可调整的参数

| 参数 | 文件位置 | 当前值 | 说明 |
|------|---------|--------|------|
| 心跳间隔 | `src/services/websocket.js:165` | 1000ms | 前端发送心跳的频率 |
| 超时阈值 | `internal/websocket/manager.go:117` | 10s | 后端认为超时的时间 |
| 检查频率 | `internal/websocket/manager.go:233` | 1s | 后端检查心跳的频率 |

## 日志示例

```
// 正常运行（有心跳）
{"level":"debug","caller":"websocket/manager.go:193","msg":"接收心跳","userID":1}

// 心跳超时（无ping信号）
{"level":"warn","caller":"websocket/manager.go:235","msg":"心跳超时，正在下线","userID":1,"超时时间":"10.5s"}
{"level":"info","caller":"websocket/manager.go:263","msg":"心跳超时自动下线","userID":1,"username":"玩家1"}

// Redis操作日志
{"level":"info","caller":"websocket/manager.go:254","msg":"从在线集合中移除失败","userID":1}
```

## 故障排查

### 问题1：客户端连接立即断开
**原因**: 心跳超时时间设置过短
**解决**: 增加 `heartbeatTimeout` 值或减小前端心跳间隔

### 问题2：后端没有自动下线日志
**原因**: 
- 前端正常发送心跳
- 或者WebSocket连接已断开

**检查**:
```
1. 查看前端控制台是否有 "接收心跳" 日志
2. 检查WebSocket连接状态
```

### 问题3：灵力继续增长
**原因**: 玩家未被从Redis中正确移除
**解决**:
```bash
# 检查Redis在线玩家列表
redis-cli
> SMEMBERS server:online:players
> DEL player:online:1  # 手动清除
```

## 相关文件修改记录

1. **前端修改**: `src/services/websocket.js`
   - 修改心跳间隔从20s改为1s

2. **后端修改**: `internal/websocket/manager.go`
   - 添加心跳时间戳字段
   - 添加心跳超时检测
   - 添加自动下线逻辑

3. **无需修改的文件**:
   - `internal/http/handlers/online/online.go` - 已有下线接口
   - `internal/spirit/spirit_grow.go` - 自动读取Redis在线列表
