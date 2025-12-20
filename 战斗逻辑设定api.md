# 秘境逐回合战斗 API 文档

## 概述

本文档描述了后端实现的逐回合战斗系统。战斗流程已从一次性执行所有回合改为暂停-获取-继续的模式，前端可以通过轮询获取每回合的战斗结果。

## 战斗流程

```
1. 前端调用 StartFight → 后端初始化战斗状态到Redis → 返回成功
2. 前端启动3秒轮询：
   a. 调用 GetRoundData → 获取当前回合的数据
   b. 更新UI显示日志和血量
   c. 如果 battleEnded == true → 战斗结束，停止轮询
   d. 否则 → 调用 ExecuteRound 执行下一回合
   e. 等待3秒后重复
3. 战斗结束后调用 EndDungeon → 更新玩家数据并清理Redis
```

## API 端点

### 1. 开始战斗 (初始化)
**POST** `/api/dungeon/fight`

**请求体:**
```json
{
  "floor": 1,
  "difficulty": "normal"
}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "victory": false,
    "floor": 1,
    "message": "战斗已初始化，请调用执行回合接口开始战斗"
  }
}
```

**说明:** 初始化战斗状态，保存到Redis。此时还未执行任何回合。

---

### 2. 获取回合数据
**GET** `/api/dungeon/round-data`

**响应 (成功):**
```json
{
  "success": true,
  "data": {
    "round": 1,
    "playerHealth": 95.0,
    "enemyHealth": 90.0,
    "logs": [
      "第1回合：玩家对敌人造成伤害10，暴击伤害5",
      "第1回合：敌人对玩家造成伤害5"
    ],
    "battleEnded": false,
    "victory": false,
    "rewards": null
  }
}
```

**响应 (无数据):**
```json
{
  "success": false,
  "message": "回合数据不存在"
}
```

**说明:** 
- 获取Redis中最新的回合数据
- 如果battleEnded为true，表示战斗已结束，包含victory和rewards字段
- logs数组包含本回合的所有战斗日志

---

### 3. 执行回合
**POST** `/api/dungeon/execute-round`

**请求体:** (无)

**响应 (继续战斗):**
```json
{
  "success": true,
  "data": {
    "round": 2,
    "playerHealth": 85.0,
    "enemyHealth": 75.0,
    "logs": [
      "第2回合：玩家对敌人造成伤害15，吸血回复3",
      "第2回合：敌人对玩家造成伤害8"
    ],
    "battleEnded": false,
    "victory": false,
    "rewards": null
  }
}
```

**响应 (战斗结束-胜利):**
```json
{
  "success": true,
  "data": {
    "round": 10,
    "playerHealth": 50.0,
    "enemyHealth": 0.0,
    "logs": ["第10回合：敌人已被击败！"],
    "battleEnded": true,
    "victory": true,
    "rewards": [
      {
        "type": "spirit_stone",
        "amount": 150
      }
    ]
  }
}
```

**响应 (错误):**
```json
{
  "success": false,
  "message": "执行回合失败",
  "error": "战斗状态不存在"
}
```

**说明:**
- 执行一个完整的回合战斗
- 回合中可能包含两个角色的多个攻击（玩家先手/敌人先手）
- 返回的logs包含本回合所有发生的战斗事件
- 自动保存更新的战斗状态到Redis

---

### 4. 结束秘境
**POST** `/api/dungeon/end`

**请求体:**
```json
{
  "floor": 1,
  "victory": true
}
```

**响应:**
```json
{
  "success": true,
  "data": {
    "floor": 1,
    "totalReward": 150,
    "victory": true,
    "spiritStones": 1250
  },
  "message": "秘境已结束"
}
```

**说明:** 
- 更新玩家数据（如果胜利，加入奖励）
- 清理Redis中的战斗状态和回合数据
- 必须在战斗结束后调用

---

## 前端实现示例

### Vue 3 + Composition API

```javascript
import { ref, onMounted, onUnmounted } from 'vue'
import api from '@/services/api'

export function useDungeonRound() {
  const roundData = ref(null)
  const logs = ref([])
  const playerHealth = ref(0)
  const enemyHealth = ref(0)
  const battleEnded = ref(false)
  const pollingInterval = ref(null)
  
  // 获取回合数据
  const getRoundData = async () => {
    try {
      const response = await api.get('/api/dungeon/round-data')
      if (response.data.success) {
        roundData.value = response.data.data
        logs.value = [...logs.value, ...roundData.value.logs]
        playerHealth.value = roundData.value.playerHealth
        enemyHealth.value = roundData.value.enemyHealth
        battleEnded.value = roundData.value.battleEnded
        
        if (roundData.value.victory) {
          console.log('战斗胜利！', roundData.value.rewards)
        }
      }
    } catch (error) {
      console.error('获取回合数据失败:', error)
    }
  }
  
  // 执行回合
  const executeRound = async () => {
    try {
      const response = await api.post('/api/dungeon/execute-round')
      if (response.data.success) {
        // 立即获取新的回合数据
        await getRoundData()
      }
    } catch (error) {
      console.error('执行回合失败:', error)
    }
  }
  
  // 启动战斗轮询
  const startPolling = async () => {
    // 第一次立即获取
    await getRoundData()
    
    // 如果战斗未结束，每3秒轮询一次
    pollingInterval.value = setInterval(async () => {
      if (!battleEnded.value) {
        // 先获取当前数据
        await getRoundData()
        
        // 如果还未结束，执行下一回合
        if (!battleEnded.value) {
          await executeRound()
        }
      } else {
        // 战斗结束，停止轮询
        stopPolling()
      }
    }, 3000)
  }
  
  // 停止轮询
  const stopPolling = () => {
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value)
      pollingInterval.value = null
    }
  }
  
  onUnmounted(() => {
    stopPolling()
  })
  
  return {
    roundData,
    logs,
    playerHealth,
    enemyHealth,
    battleEnded,
    getRoundData,
    executeRound,
    startPolling,
    stopPolling
  }
}
```

### 在组件中使用

```vue
<template>
  <div class="dungeon-battle">
    <div class="battle-log">
      <div v-for="(log, index) in logs" :key="index" class="log-item">
        {{ log }}
      </div>
    </div>
    
    <div class="battle-stats">
      <div class="player-health">玩家血量: {{ playerHealth }}</div>
      <div class="enemy-health">敌人血量: {{ enemyHealth }}</div>
    </div>
    
    <button @click="startBattle" :disabled="battleInProgress">开始战斗</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useDungeonRound } from './useDungeonRound'
import api from '@/services/api'

const battleInProgress = ref(false)
const floor = ref(1)
const difficulty = ref('normal')

const {
  logs,
  playerHealth,
  enemyHealth,
  battleEnded,
  startPolling,
  stopPolling
} = useDungeonRound()

const startBattle = async () => {
  battleInProgress.value = true
  
  try {
    // 1. 初始化战斗
    await api.post('/api/dungeon/fight', {
      floor: floor.value,
      difficulty: difficulty.value
    })
    
    // 2. 启动轮询
    await startPolling()
    
    // 3. 等待战斗结束
    const checkBattle = setInterval(() => {
      if (battleEnded.value) {
        clearInterval(checkBattle)
        stopPolling()
        
        // 4. 结束秘境，更新玩家数据
        api.post('/api/dungeon/end', {
          floor: floor.value,
          victory: true // 根据实际情况设置
        })
        
        battleInProgress.value = false
      }
    }, 1000)
  } catch (error) {
    console.error('战斗失败:', error)
    battleInProgress.value = false
  }
}
</script>

<style scoped>
.dungeon-battle {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.battle-log {
  border: 1px solid #ccc;
  padding: 1rem;
  height: 300px;
  overflow-y: auto;
  background: #f5f5f5;
}

.log-item {
  padding: 0.5rem;
  border-bottom: 1px solid #e0e0e0;
}

.battle-stats {
  display: flex;
  justify-content: space-between;
  padding: 1rem;
  border: 1px solid #ccc;
}

.player-health, .enemy-health {
  font-weight: bold;
}
</style>
```

## Redis 数据结构

### 战斗状态 (dungeon:battle:status:{userID})
```json
{
  "userID": 1,
  "floor": 1,
  "difficulty": "normal",
  "round": 2,
  "playerHealth": 85.0,
  "playerMaxHealth": 100.0,
  "enemyHealth": 75.0,
  "enemyMaxHealth": 100.0,
  "playerStats": { ... },
  "enemyStats": { ... },
  "buffEffects": { ... },
  "battleLog": [...]
}
```

### 回合数据 (dungeon:battle:round:{userID})
```json
{
  "round": 2,
  "playerHealth": 85.0,
  "enemyHealth": 75.0,
  "logs": ["第2回合：...", "第2回合：..."],
  "battleEnded": false,
  "victory": false,
  "rewards": null
}
```

## 关键实现细节

### 后端处理流程

1. **StartFight**: 
   - 解析玩家属性，应用增益
   - 生成敌人属性
   - 创建BattleStatus并保存到Redis
   - 返回初始化成功

2. **ExecuteRound**:
   - 从Redis加载BattleStatus
   - 执行完整的一回合逻辑（包括双方可能的多个攻击）
   - 更新血量、日志等
   - 检查战斗是否结束（某方死亡或超过100回合）
   - 保存更新的BattleStatus到Redis
   - 返回RoundData

3. **GetRoundData**:
   - 从Redis获取最新的RoundData
   - 直接返回给前端

4. **EndDungeon**:
   - 更新玩家数据（奖励）
   - 清理Redis中的战斗数据

### 前端轮询策略

- **初始延迟**: 0秒（立即获取第一回合数据）
- **轮询间隔**: 3秒
- **轮询流程**:
  1. 获取回合数据 (GetRoundData)
  2. 更新UI
  3. 检查 battleEnded
  4. 如果未结束，执行下一回合 (ExecuteRound)
  5. 等待3秒后重复

## 注意事项

1. **Redis可用性**: 如果Redis不可用，战斗系统会降级到内存存储（仅当前请求有效）
2. **超时处理**: 建议前端设置30秒的超时，如果长时间无响应则中止战斗
3. **网络波动**: 如果轮询失败，应该重试而不是立即放弃
4. **并发限制**: 同一用户不应该同时进行多个战斗
5. **数据一致性**: EndDungeon必须在战斗结束后调用以清理Redis数据

## 故障排查

### 回合数据不存在
- 检查是否调用了StartFight
- 检查Redis连接状态
- 查看后端日志

### 战斗卡住不响应
- 检查ExecuteRound API是否正常
- 查看Redis中的BattleStatus是否被正确保存
- 检查网络连接

### 数据不一致
- 确保使用相同的userID
- 检查Redis的过期时间（60分钟）
- 清理过期的Redis数据

