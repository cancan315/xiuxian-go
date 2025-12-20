# 秘境逐回合战斗实现总结

## 概述

本次实现将秘境战斗系统从一次性执行所有回合改为"暂停-获取-继续"的流式模式。前端可以通过轮询获取每个回合的战斗数据，后端在Redis中维护战斗状态。

## 实现内容

### 1. 数据模型 (models.go)

#### RoundData - 单回合战斗结果
```go
type RoundData struct {
    Round           int           // 回合数
    PlayerHealth    float64       // 玩家血量
    EnemyHealth     float64       // 敌人血量
    Logs            []string      // 本回合日志
    BattleEnded     bool          // 战斗是否结束
    Victory         bool          // 是否胜利 (战斗结束时有效)
    Rewards         []interface{} // 奖励 (战斗结束时有效)
}
```

#### BattleStatus - 战斗状态 (保存在Redis)
```go
type BattleStatus struct {
    UserID           uint                    // 用户ID
    Floor            int                     // 当前层数
    Difficulty       string                  // 难度
    Round            int                     // 当前回合
    PlayerHealth     float64                 // 玩家当前血量
    PlayerMaxHealth  float64                 // 玩家最大血量
    EnemyHealth      float64                 // 敌人当前血量
    EnemyMaxHealth   float64                 // 敌人最大血量
    PlayerStats      *CombatStats            // 玩家战斗属性
    EnemyStats       *CombatStats            // 敌人战斗属性
    BuffEffects      map[string]interface{}  // 玩家已选增益
    BattleLog        []string                // 完整战斗日志
}
```

### 2. 核心业务逻辑 (service.go & execute_round.go)

#### StartFight() - 改造
- **原逻辑**: 一次性执行所有回合并返回完整结果
- **新逻辑**: 
  - 初始化战斗状态 (BattleStatus)
  - 保存到Redis (key: `dungeon:battle:status:{userID}`)
  - 返回初始化成功消息
  - **不执行任何回合**

#### ExecuteRound() - 新增 (execute_round.go)
- 从Redis加载当前的BattleStatus
- 执行**完整的一个回合**（包括双方可能的多个攻击）
- 更新血量、日志、状态等
- 检查战斗是否结束：
  - 玩家或敌人血量≤0 → 战斗结束
  - 回合数≥100 → 战斗失败
- 计算奖励（如果胜利）
- 保存更新的BattleStatus到Redis
- 返回RoundData（包含本回合的所有数据）

#### Redis操作方法
- `SaveBattleStatusToRedis()` - 保存战斗状态
- `LoadBattleStatusFromRedis()` - 加载战斗状态
- `SaveRoundDataToRedis()` - 保存回合数据
- `GetRoundDataFromRedis()` - 获取回合数据
- `ClearBattleStatusFromRedis()` - 清除战斗状态
- `ClearRoundDataFromRedis()` - 清除回合数据

#### EndDungeon() - 改造
- 更新玩家数据（奖励）
- **新增**: 清理Redis中的BattleStatus和RoundData
- 防止过期数据积累

### 3. HTTP 处理器 (dungeon.go)

#### GetRoundData - GET /api/dungeon/round-data
- 从Redis获取最新的RoundData
- 返回本回合的完整战斗信息
- 日志、血量、战斗状态等

#### ExecuteRound - POST /api/dungeon/execute-round
- 调用服务的ExecuteRound()执行一回合
- 自动保存RoundData到Redis
- 返回本回合的结果

### 4. 路由注册 (router.go)
```
GET  /api/dungeon/round-data   → GetRoundData
POST /api/dungeon/execute-round → ExecuteRound
```

## Redis 数据结构

### 战斗状态存储
- **Key**: `dungeon:battle:status:{userID}`
- **格式**: JSON (Marshal/Unmarshal)
- **TTL**: 60分钟
- **内容**: 完整的BattleStatus，包含玩家/敌人属性、已选增益、战斗日志等

### 回合数据存储
- **Key**: `dungeon:battle:round:{userID}`
- **格式**: JSON (Marshal/Unmarshal)
- **TTL**: 60分钟
- **内容**: 本回合的RoundData，用于前端轮询获取

## 工作流程

```
前端流程：
1. 调用 StartFight(floor, difficulty)
   ↓
   后端: 初始化BattleStatus → 保存Redis → 返回
   
2. 启动3秒轮询循环:
   a. 调用 GetRoundData()
      ↓
      后端: 从Redis获取RoundData → 返回
      ↓
      前端: 更新UI (血量、日志等)
      
   b. 检查 battleEnded
      ✗ 未结束 → 等待3秒
      ✓ 已结束 → 跳转到第3步
      
   c. 如果未结束，调用 ExecuteRound()
      ↓
      后端: 从Redis加载BattleStatus
         → 执行一回合逻辑
         → 更新状态
         → 保存Redis
         → 返回RoundData
      ↓
      前端: 记录本回合数据（用于下次轮询）
      
3. 调用 EndDungeon(floor, victory)
   ↓
   后端: 更新玩家数据 → 清理Redis → 返回
```

## 关键特性

### 1. 独立的回合执行
- 每个回合的逻辑完全独立
- ExecuteRound可以多次调用，每次执行一个完整的回合
- 支持网络中断后恢复（重新GetRoundData获取状态）

### 2. Redis持久化
- 战斗状态完整保存在Redis
- 即使后端进程重启，前端可以继续战斗
- 60分钟自动过期清理

### 3. 流畅的用户体验
- 前端可以实时显示每回合的结果
- 支持3秒轮询间隔（可调）
- 日志逐步累积显示

### 4. 完整的战斗流程
- 每回合可能包含多个攻击（玩家先/敌人先）
- 正确处理特殊效果（眩晕、吸血、暴击等）
- 正确判定战斗结束（死亡、超回合）

## 测试检查清单

- [ ] StartFight 返回成功，战斗状态保存到Redis
- [ ] ExecuteRound 成功执行一回合，血量正确更新
- [ ] GetRoundData 能正确获取最新的回合数据
- [ ] 回合日志正确记录（伤害、暴击、吸血等）
- [ ] 战斗结束判断正确（死亡/超回合）
- [ ] 奖励计算正确
- [ ] EndDungeon 清理Redis数据
- [ ] 前端轮询逻辑正确（3秒间隔）
- [ ] 网络中断后能恢复战斗状态
- [ ] 多用户并发战斗不冲突

## 注意事项

1. **回合数据保存时机**: ExecuteRound自动保存RoundData，GetRoundData读取时总是获取最新数据
2. **增益效果**: StartFight时应用增益，后续回合不再应用（在playerStats中已包含）
3. **计算奖励**: 只在战斗胜利时计算，保存在s.rewardAmount，EndDungeon时使用
4. **错误处理**: Redis不可用时，仅当前请求可能失败，建议重试
5. **超时管理**: 建议前端设置30秒超时，长时间无响应则中止战斗

## 后续优化建议

1. **WebSocket替代轮询** - 实时推送回合数据而不是轮询
2. **战斗暂停功能** - 支持玩家主动暂停战斗
3. **自动战斗加速** - 支持快速自动战斗（不显示过程）
4. **战斗录像功能** - 保存完整战斗日志用于回放
5. **性能优化** - 减少Redis序列化/反序列化开销

## 相关文件

- `/DUNGEON_ROUND_POLLING_GUIDE.md` - 详细的API文档和前端实现指南
- `/server-go/internal/dungeon/execute_round.go` - ExecuteRound实现
- `/server-go/internal/dungeon/service.go` - StartFight和EndDungeon改造
- `/server-go/internal/http/handlers/dungeon/dungeon.go` - Handler实现
- `/server-go/internal/http/router/router.go` - 路由注册

