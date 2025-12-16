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
  const isAutoExploring = ref(false)

  // 从后端获取最新修炼数据
  const syncCultivationData = async () => {
    try {
      const token = getAuthToken();
      
      // 获取修炼消耗和获得数据
      const response = await APIService.getCultivationData(token)
      if (response.success) {
      //  console.log('[Cultivation] 修炼消耗:', response.data.spiritCost, '获得:', response.data.cultivationGain)
        playerInfoStore.level = response.data.level // 境界等级
        playerInfoStore.realm = response.data.realm // 境界
        playerInfoStore.cultivation = response.data.cultivation // 当前修为
        playerInfoStore.maxCultivation = response.data.maxCultivation // 最大修为
        playerInfoStore.spirit = response.data.spirit // 当前灵力
        playerInfoStore.cultivationCost = response.data.spiritCost       // 修炼消耗灵力
        playerInfoStore.cultivationGain = response.data.cultivationGain // 修炼获得修为
        playerInfoStore.spiritRate = response.data.spiritRate // 灵力获取倍率
        playerInfoStore.spiritStones = response.data.spiritStones // 灵石数量
        playerInfoStore.reinforceStones = response.data.reinforceStones // 强化石数量
        
      }
    } catch (error) {
      console.error('同步修为数据失败:', error)
    }
  }

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
      
      const response = await APIService.startExploration(token)
      
      if (response.success) {
        await syncCultivationData()
        // 处理返回的事件
        if (response.events && response.events.length > 0) {
          for (const event of response.events) {
            triggerEvent(event)
          }
          playerInfoStore.explorationCount += 1
        }
        
        // 添加后端日志（已包含所有事件信息）
        if (response.log) {
          addExplorationLog(response.log)
          message.success(response.log)
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
      await new Promise(resolve => setTimeout(resolve, 3000))
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

  // 处理事件（自动统计）
  const triggerEvent = (event) => {
    // 必须有后端返回的Description，否则视为失败
    if (!event.description) {
      message.error('探索失败：事件描述缺失')
      addExplorationLog('探索失败：事件描述缺失')
      return
    }
    
    playerInfoStore.eventTriggered += 1
    // 日志由后端 response.log 統一提供
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
          点击开始自动探索按钮，在修仙世界中追寻仙缘。
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
    

  </div>
</template>

<style scoped>
.exploration-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}
</style>