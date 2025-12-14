// src/components/WebSocketDebug.vue
// WebSocket调试面板组件

<template>
  <div class="websocket-debug">
    <n-card title="WebSocket 调试面板" size="small">
      <!-- 连接状态 -->
      <n-space vertical>
        <div class="status-row">
          <span>连接状态:</span>
          <n-tag :type="isConnected ? 'success' : 'error'">
            {{ isConnected ? '已连接' : '未连接' }}
          </n-tag>
        </div>

        <!-- 连接控制 -->
        <n-space>
          <n-button 
            type="primary" 
            @click="connect"
            :disabled="isConnected"
            size="small"
          >
            连接
          </n-button>
          <n-button 
            type="error" 
            @click="disconnect"
            :disabled="!isConnected"
            size="small"
          >
            断开
          </n-button>
        </n-space>

        <!-- 接收的消息 -->
        <div class="messages-section">
          <h4>最近消息</h4>
          <n-scrollbar style="height: 300px">
            <n-space vertical style="padding: 10px">
              <div 
                v-for="(msg, idx) in recentMessages" 
                :key="idx"
                class="message-item"
              >
                <div class="msg-time">{{ formatTime(msg.time) }}</div>
                <div class="msg-type">{{ msg.type }}</div>
                <div class="msg-data">{{ formatData(msg.data) }}</div>
              </div>
            </n-space>
          </n-scrollbar>
        </div>

        <!-- 统计信息 -->
        <n-divider />
        <div class="stats">
          <p>灵力增长事件: {{ stats.spiritGrowth }}</p>
          <p>战斗事件: {{ stats.dungeonEvent }}</p>
          <p>排行榜更新: {{ stats.leaderboardUpdate }}</p>
          <p>探索事件: {{ stats.explorationEvent }}</p>
        </div>
      </n-space>
    </n-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NCard, NSpace, NButton, NTag, NDivider, NScrollbar } from 'naive-ui'
import { wsManager } from '@/services/websocket'

const isConnected = ref(false)
const recentMessages = ref([])
const stats = ref({
  spiritGrowth: 0,
  dungeonEvent: 0,
  leaderboardUpdate: 0,
  explorationEvent: 0
})

const userStore = useUserStore()

async function connect() {
  try {
    const token = userStore.token
    const userId = userStore.userId
    await wsManager.connect(token, userId)
    isConnected.value = true
  } catch (error) {
    console.error('连接失败:', error)
  }
}

function disconnect() {
  wsManager.disconnect()
  isConnected.value = false
}

function formatTime(time) {
  return new Date(time).toLocaleTimeString()
}

function formatData(data) {
  return JSON.stringify(data, null, 2).substring(0, 100)
}

function addMessage(type, data) {
  recentMessages.value.unshift({
    type,
    data,
    time: Date.now()
  })
  
  if (recentMessages.value.length > 20) {
    recentMessages.value.pop()
  }

  // 更新统计
  switch(type) {
    case 'spirit:grow':
      stats.value.spiritGrowth++
      break
    case 'dungeon:event':
      stats.value.dungeonEvent++
      break
    case 'leaderboard:update':
      stats.value.leaderboardUpdate++
      break
    case 'exploration:event':
      stats.value.explorationEvent++
      break
  }
}

onMounted(() => {
  wsManager.on('spirit:grow', (data) => addMessage('spirit:grow', data))
  wsManager.on('dungeon:event', (data) => addMessage('dungeon:event', data))
  wsManager.on('leaderboard:update', (data) => addMessage('leaderboard:update', data))
  wsManager.on('exploration:event', (data) => addMessage('exploration:event', data))
  
  wsManager.on('connection:open', () => {
    isConnected.value = true
  })
  
  wsManager.on('connection:close', () => {
    isConnected.value = false
  })
})
</script>

<style scoped>
.websocket-debug {
  margin: 20px 0;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.messages-section {
  margin-top: 15px;
}

.messages-section h4 {
  margin-bottom: 10px;
}

.message-item {
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
  margin-bottom: 5px;
}

.msg-time {
  font-size: 12px;
  color: #999;
}

.msg-type {
  font-weight: bold;
  color: #333;
  margin: 2px 0;
}

.msg-data {
  font-size: 12px;
  color: #666;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
}

.stats {
  padding: 10px;
  background: #f9f9f9;
  border-radius: 4px;
}

.stats p {
  margin: 5px 0;
  font-size: 14px;
}
</style>
