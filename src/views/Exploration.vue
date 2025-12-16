<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import LogPanel from '../components/LogPanel.vue'
  import APIService from '../services/api'
  import { getAuthToken } from '../stores/db'

  const playerInfoStore = usePlayerInfoStore()
  
  const message = useMessage()
  const explorationLogs = ref([])
  const isExploring = ref(false)
  const currentEvent = ref(null)
  const showEventModal = ref(false)
  const isAutoExploring = ref(false)

  // 添加探索日志
  const addExplorationLog = (message) => {
    explorationLogs.value.push({
      time: new Date().toLocaleTimeString(),
      content: message,
      type: 'info'
    })
    // 限制日志数量
    if (explorationLogs.value.length > 100) {
      explorationLogs.value.shift()
    }
  }

  // 单次探索
  const explorationOnce = async () => {
    try {
      const token = getAuthToken()
      if (!token) {
        addExplorationLog('未获得授权令牌')
        return false
      }
      
      const response = await APIService.startExploration(token, 10000)
      
      if (response.success) {
        // 处理返回的事件
        if (response.events && response.events.length > 0) {
          for (const event of response.events) {
            triggerEvent(event)
          }
        } else {
          addExplorationLog('探索了一段时间，未发生特殊事件')
        }
        
        // 添加后端日志（已包含所有事件信息）
        if (response.log) {
          addExplorationLog(response.log)
        }
        return true
      } else {
        addExplorationLog(`探索失败: ${response.error || '未知错误'}`)
        message.error('探索失败')
        return false
      }
    } catch (error) {
      addExplorationLog(`探索失败: ${error.message}`)
      message.error('探索失败')
      console.error('[Exploration] 探索错误:', error)
      return false
    }
  }

  // 开始探索（将灵力检查由后端处理）
  const startExploration = async () => {
    isAutoExploring.value = true
    addExplorationLog('开始自动探索...')
    
    while (isAutoExploring.value) {
      const success = await explorationOnce()
      if (!success) {
        isAutoExploring.value = false
        break
      }
      
      // 短暂延迟，避免过快调用
      await new Promise(resolve => setTimeout(resolve, 1000))
    }
    
    // 探索完成后调用 /api/player/data 接口
    await getPlayerData()
  }

  // 停止探索
  const stopExploration = () => {
    isAutoExploring.value = false
    addExplorationLog('停止探索')
    message.info('已停止探索')
  }

  // 从后端获取最新玩家数据
  const getPlayerData = async () => {
    try {
      const token = getAuthToken()
      if (!token) return
      
      const response = await APIService.getPlayerData(token)
      if (response.success) {
        // 将返回的数据更新到 usePlayerInfoStore
        playerInfoStore.spirit = response.spirit
        playerInfoStore.spiritStones = response.spiritStones
        playerInfoStore.cultivation = response.cultivation
        playerInfoStore.level = response.level
        playerInfoStore.realm = response.realm
        playerInfoStore.maxCultivation = response.maxCultivation
      }
    } catch (error) {
      console.error('[Exploration] 获取玩家数据失败:', error)
    }
  }

  // 触发事件
  const triggerEvent = (event) => {
    // 必须有后端返回的Description，否则视为失败
    if (!event.description) {
      message.error('探索失败：事件描述缺失')
      addExplorationLog('探索失败：事件描述缺失')
      return
    }
    
    currentEvent.value = event
    showEventModal.value = true
    playerInfoStore.eventTriggered += 1
    
    // 仅执行弹窗日志，不添加到日志面板（会会重複）
    // 日志由后端 response.log 統一提供
  }

  // 处理事件选择
  const handleEventChoice = async (choice) => {
    if (!currentEvent.value) return
    
    const event = currentEvent.value
    showEventModal.value = false
    
    // 如果选择是"停止探索"，直接停止自动探索不调用后端接口
    if (choice.value === 'stop') {
      isAutoExploring.value = false
      addExplorationLog('停止探索')
      currentEvent.value = null
      playerInfoStore.explorationCount += 1
      return
    }
    
    try {
      const token = getAuthToken()
      if (!token) {
        addExplorationLog('未获得授权令牌')
        return
      }
      
      // 调用后端API处理事件选择
      const response = await APIService.handleExplorationEventChoice(token, event.type, choice)
      
      if (response.success) {
        // 根据事件类型处理奖励
        switch (event.type) {
          case 'item_found':
            addExplorationLog('获得了物品')
            break
          case 'spirit_stone_found':
            addExplorationLog(`获得了${event.amount}灵石`)
            break
          case 'herb_found':
            addExplorationLog(`获得了灵草`)
            break
          case 'pill_recipe_fragment_found':
            addExplorationLog('获得了丹方残页')
            break
          case 'battle_encounter':
            addExplorationLog('战斗结束')
            break
        }
      }
    } catch (error) {
      addExplorationLog(`处理失败: ${error.message}`)
      message.error('事件处理失败')
      console.error('[Exploration] 事件处理错误:', error)
    }
    
    currentEvent.value = null
    playerInfoStore.explorationCount += 1
  }

  onUnmounted(() => {
    // 清理：停止自动探索
    isAutoExploring.value = false
  })

  // 页面挂载时获取最新的玩家数据
  onMounted(async () => {
    await getPlayerData()
  })
</script>

<template>
  <div class="exploration-container">
    <n-card title="探索">
      <n-space vertical>
        <n-alert v-if="!isAutoExploring" type="info">
          点击开始自动探索按钮，在修仙世界中展开冬险之旅。
        </n-alert>
        
        <n-space>
          <n-button @click="startExploration" :disabled="isAutoExploring" type="primary">
            {{ isAutoExploring ? '探索中...' : '开始自动探索' }}
          </n-button>
          <n-button @click="stopExploration" :disabled="!isAutoExploring" type="error">
            停止探索
          </n-button>
        </n-space>
        
        <n-divider />
        
        <log-panel :logs="explorationLogs" title="探索日志" />
      </n-space>
    </n-card>
    
    <!-- 事件弹窗 -->
    <n-modal v-model:show="showEventModal" preset="dialog" title="探索事件">
      <template #default>
        <div v-if="currentEvent">
          <p>{{ currentEvent.description }}</p>
          <n-space>
            <n-button 
              v-for="(choice, index) in currentEvent.choices" 
              :key="index"
              @click="handleEventChoice(choice)"
              type="primary"
            >
              {{ choice.text }}
            </n-button>
          </n-space>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
.exploration-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}
</style>