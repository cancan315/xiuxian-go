# 秘境逐回合战斗 - 快速参考

## 核心API

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/dungeon/fight` | 初始化战斗，保存到Redis |
| GET | `/api/dungeon/round-data` | 获取最新回合数据 |
| POST | `/api/dungeon/execute-round` | 执行一个回合 |
| POST | `/api/dungeon/end` | 结束战斗，更新数据，清理Redis |

## 前端轮询流程（伪代码）

```javascript
async function battleLoop() {
  // 1. 初始化
  await POST('/api/dungeon/fight', { floor, difficulty })
  
  // 2. 轮询循环
  while (true) {
    // 获取当前数据
    const data = await GET('/api/dungeon/round-data')
    
    // 更新UI
    updateLogs(data.logs)
    updateHealth(data.playerHealth, data.enemyHealth)
    
    // 检查结束
    if (data.battleEnded) break
    
    // 等待3秒后执行下一回合
    await sleep(3000)
    await POST('/api/dungeon/execute-round')
  }
  
  // 3. 结束
  await POST('/api/dungeon/end', { floor, victory: data.victory })
}
```

## Redis 键结构

```
dungeon:battle:status:{userID}  → BattleStatus (完整战斗状态)
dungeon:battle:round:{userID}   → RoundData (本回合数据)
```

## 数据响应示例

### StartFight 成功
```json
{
  "success": true,
  "data": {
    "message": "战斗已初始化，请调用执行回合接口开始战斗"
  }
}
```

### GetRoundData 成功
```json
{
  "success": true,
  "data": {
    "round": 1,
    "playerHealth": 95,
    "enemyHealth": 90,
    "logs": ["第1回合：玩家对敌人造成伤害10"],
    "battleEnded": false
  }
}
```

### ExecuteRound 成功（继续）
```json
{
  "success": true,
  "data": {
    "round": 2,
    "playerHealth": 85,
    "enemyHealth": 75,
    "logs": ["第2回合：..."],
    "battleEnded": false
  }
}
```

### ExecuteRound 成功（结束）
```json
{
  "success": true,
  "data": {
    "round": 10,
    "playerHealth": 50,
    "enemyHealth": 0,
    "logs": ["第10回合：敌人已被击败！"],
    "battleEnded": true,
    "victory": true,
    "rewards": [{"type": "spirit_stone", "amount": 150}]
  }
}
```

### EndDungeon 成功
```json
{
  "success": true,
  "data": {
    "floor": 1,
    "totalReward": 150,
    "victory": true,
    "spiritStones": 1250
  }
}
```

## 关键参数说明

| 参数 | 说明 | 范围 |
|------|------|------|
| floor | 秘境层数 | 1-∞ |
| difficulty | 难度 | easy, normal, hard, expert |
| round | 当前回合数 | 1-100 |
| playerHealth | 玩家当前血量 | 0-maxHealth |
| battleEnded | 战斗是否结束 | true/false |
| victory | 是否胜利 | true/false |
| logs | 本回合日志 | 字符串数组 |
| rewards | 奖励信息 | 对象数组 |

## 错误处理

### 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|---------|
| "战斗状态不存在" | 未调用StartFight或Redis过期 | 重新调用StartFight |
| "回合数据不存在" | Redis中无数据 | 调用ExecuteRound生成数据 |
| "执行回合失败" | 战斗已结束或状态异常 | 检查battleEnded标志 |

### 错误响应示例

```json
{
  "success": false,
  "message": "战斗状态不存在",
  "error": "加载战斗状态失败: redis: nil"
}
```

## 时间流程

```
时刻      操作              响应时间
T+0s     StartFight       <100ms (初始化)
T+0s     GetRoundData     <100ms (获取第一回合)
T+3s     ExecuteRound     200-500ms (执行一回合)
T+3s     GetRoundData     <100ms (获取更新)
T+6s     ExecuteRound     200-500ms (执行下一回合)
...      ...              ...
T+30s    EndDungeon       <100ms (结束并清理)
```

## 优化建议

### 前端优化
1. **缓存日志** - 避免重复显示
2. **优化重绘** - 只更新变化的UI元素
3. **错误重试** - 网络失败时自动重试（最多3次）
4. **超时控制** - 30秒无响应则中止战斗

### 后端优化
1. **批量操作** - 减少Redis往返
2. **异步保存** - 不阻塞主线程
3. **数据压缩** - 减少网络传输

## 常见问题

**Q: 战斗中网络中断怎么办？**
A: 再次调用GetRoundData获取当前状态，然后继续。BattleStatus保存在Redis中，60分钟内有效。

**Q: 可以加快回合速度吗？**
A: 可以，调整前端轮询间隔（目前是3秒）。建议不要低于1秒，避免服务器压力过大。

**Q: 100回合后战斗为什么自动失败？**
A: 这是设计的防护机制，防止无限战斗。如需修改，编辑execute_round.go中的maxRounds常量。

**Q: EndDungeon是必须调用的吗？**
A: 必须！否则Redis中的战斗数据会在60分钟后过期，浪费存储空间。

**Q: 同一用户可以同时进行多个战斗吗？**
A: 不可以。一个userID只能有一个活跃的BattleStatus。如果需要同时多个战斗，需要修改Redis键的设计。

