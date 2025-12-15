<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import LogPanel from '../components/LogPanel.vue'
  import APIService from '../services/api'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  
  const message = useMessage()
  const explorationLogs = ref([])
  const isExploring = ref(false)
  const explorationWorker = ref(null) // 保留但不使用，与后端API集成
  const currentEvent = ref(null)
  const showEventModal = ref(false)

  // 添加探索日志
  const addExplorationLog = (message) => {
    explorationLogs.value.push({
      time: new Date().toLocaleTimeString(),
      message: message
    })
    // 限制日志数量
    if (explorationLogs.value.length > 100) {
      explorationLogs.value.shift()
    }
  }

  // 开始探索
  const startExploration = async () => {
    if (isExploring.value) return
    
    isExploring.value = true
    addExplorationLog('开始探索...')
    
    try {
      const response = await apiClient.post('/api/exploration/start', {
        duration: 10000 // 10秒探索时间
      })
      
      if (response.success) {
        // 处理返回的事件
        if (response.events && response.events.length > 0) {
          for (const event of response.events) {
            triggerEvent(event)
          }
        } else {
          addExplorationLog('探索了一段时间，未发生特殊事件')
        }
        
        // 添加后端日志
        if (response.log) {
          addExplorationLog(response.log)
        }
      }
    } catch (error) {
      addExplorationLog(`探索失败: ${error.message}`)
      message.error('探索失败')
    } finally {
      finishExploration()
    }
  }

  // 停止探索
  const stopExploration = () => {
    isExploring.value = false
    addExplorationLog('停止探索')
  }

  // 触发事件
  const triggerEvent = (event) => {
    currentEvent.value = event
    showEventModal.value = true
    statsStore.eventTriggered += 1
    
    // 根据事件类型添加日志
    switch (event.type) {
      case 'item_found':
        addExplorationLog('发现了物品！')
        break
      case 'spirit_stone_found':
        addExplorationLog(`发现了${event.amount}灵石！`)
        break
      case 'herb_found':
        addExplorationLog('发现了灵草！')
        break
      case 'pill_recipe_fragment_found':
        addExplorationLog('发现了丹方残页！')
        break
      case 'battle_encounter':
        addExplorationLog('遭遇了妖兽！')
        break
    }
  }

  // 处理事件选择
  const handleEventChoice = async (choice) => {
    if (!currentEvent.value) return
    
    const event = currentEvent.value
    showEventModal.value = false
    
    try {
      // 调用后端API处理事件选择
      const response = await apiClient.post('/api/exploration/event-choice', {
        eventType: event.type,
        choice: choice
      })
      
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
    }
    
    currentEvent.value = null
    statsStore.explorationCount += 1
  }

  // 结束探索
  const finishExploration = () => {
    isExploring.value = false
    addExplorationLog('探索结束')
  }

  onUnmounted(() => {
    // Worker已移除，后端处理所有逻辑
  })

  // 从后端获取最新的玩家数据
  onMounted(async () => {
    try {
      const response = await apiClient.get('/api/player/data')
      // 同步玩家数据到store
    } catch (error) {
      console.error('Failed to load player data:', error)
    }
  })
</script>

<template>
  <div class="exploration-container">
    <n-card title="探索">
      <n-space vertical>
        <n-alert v-if="!isExploring" type="info">
          点击开始探索按钮，在修仙世界中展开冒险之旅。
        </n-alert>
        
        <n-space>
          <n-button @click="startExploration" :disabled="isExploring" type="primary">
            {{ isExploring ? '探索中...' : '开始探索' }}
          </n-button>
          <n-button @click="stopExploration" :disabled="!isExploring" type="error">
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