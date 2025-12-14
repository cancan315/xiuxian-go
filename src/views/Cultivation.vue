<template>
  <n-card title="修炼">
    <n-space vertical>
      <n-alert type="info" show-icon>
        <template #icon>
          <n-icon>
            <BookOutline />
          </n-icon>
        </template>
        通过打坐修炼来提升修为，积累足够的修为后可以尝试突破境界。
      </n-alert>
      <n-space vertical>
        <n-button type="primary" size="large" block @click="cultivate" :disabled="playerInfoStore.spirit < cultivationCost">
          打坐修炼 (消耗 {{ cultivationCost }} 灵力)
        </n-button>
        <n-button :type="isAutoCultivating ? 'warning' : 'success'" size="large" block @click="toggleAutoCultivation">
          {{ isAutoCultivating ? '停止自动修炼' : '开始自动修炼' }}
        </n-button>
        <n-button
          type="info"
          size="large"
          block
          @click="cultivateUntilBreakthrough"
          :disabled="playerInfoStore.spirit < calculateBreakthroughCost()"
        >
          一键突破
        </n-button>
      </n-space>
      <n-divider>修炼详情</n-divider>
      <n-descriptions bordered>
        <n-descriptions-item label="灵力获取速率">{{ baseGainRate * playerInfoStore.spiritRate }} / 秒</n-descriptions-item>
        <n-descriptions-item label="修炼效率">{{ cultivationGain }} 修为 / 次</n-descriptions-item>
        <n-descriptions-item label="突破所需修为">
          {{ playerInfoStore.maxCultivation }}
        </n-descriptions-item>
      </n-descriptions>
      <log-panel ref="logRef" title="修炼日志" />
    </n-space>
  </n-card>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { usePersistenceStore } from '../stores/persistence'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { BookOutline } from '@vicons/ionicons5' // 导入图标组件
  import LogPanel from '../components/LogPanel.vue'
  import { apiClient } from '../services/api'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  
  const message = useMessage()
  const isAutoCultivating = ref(false)
  const logRef = ref(null)

  // 修炼消耗
  const cultivationCost = computed(() => {
    return 1 // 每次修炼消耗1点灵力
  })

  // 修炼获得的修为
  const cultivationGain = computed(() => {
    return 1 // 每次修炼获得1点修为
  })

  // 基础灵力获取速率
  const baseGainRate = 1

  // 计算突破所需成本
  const calculateBreakthroughCost = () => {
    return playerInfoStore.maxCultivation - playerInfoStore.cultivation
  }

  // 打坐修炼方法
  const cultivate = async () => {
    try {
      const response = await apiClient.post('/api/cultivation/single', {})
      
      if (response.success) {
        // 后端返回的数据字段名
        playerInfoStore.spirit -= response.spiritCost
        playerInfoStore.cultivation = response.currentCultivation
        
        // 记录日志
        if (logRef.value) {
          logRef.value.addLog(`修炼获得 ${response.cultivationGain.toFixed(0)} 点修为`)
        }
        
        // 检查是否有突破
        if (response.breakthrough) {
          const bt = response.breakthrough
          playerInfoStore.level = bt.newLevel
          playerInfoStore.realm = bt.newRealm
          playerInfoStore.maxCultivation = bt.newMaxCultivation
          playerInfoStore.spirit += bt.spiritReward
          playerInfoStore.spiritRate = bt.newSpiritRate
          statsStore.breakthroughCount += 1
          message.success(bt.message)
          if (logRef.value) {
            logRef.value.addLog(bt.message)
          }
        }
        
        // 增加修炼时间统计
        statsStore.totalCultivationTime += 1
      } else {
        message.warning(response.error || '修炼失败')
      }
    } catch (error) {
      message.error('修炼请求失败：' + error.message)
      if (logRef.value) {
        logRef.value.addLog(`修炼失败：${error.message}`)
      }
    }
  }

  // 切换自动修炼
  const toggleAutoCultivation = async () => {
    if (isAutoCultivating.value) {
      // 停止自动修炼
      isAutoCultivating.value = false
      message.info('停止自动修炼')
      return
    }
    
    // 开始自动修炼
    isAutoCultivating.value = true
    message.success('开始自动修炼（10秒）')
    
    try {
      const response = await apiClient.post('/api/cultivation/auto', {
        duration: 10000
      })
      
      if (response.success) {
        // 更新玩家数据 - 同步最新数据
        await syncCultivationData()
        
        // 记录日志
        if (logRef.value) {
          if (response.message) {
            logRef.value.addLog(response.message)
          }
          if (response.breakthroughDetails && Array.isArray(response.breakthroughDetails)) {
            for (const detail of response.breakthroughDetails) {
              logRef.value.addLog(`突破：${detail}`)
            }
          }
        }
        
        // 更新突破次数
        statsStore.breakthroughCount += response.breakthroughs
        statsStore.totalCultivationTime += 10
        
        message.success(`自动修炼完成！获得 ${response.totalCultivationGain.toFixed(0)} 点修为，突破 ${response.breakthroughs} 次`)
      } else {
        message.warning(response.error || '自动修炼失败')
      }
    } catch (error) {
      message.error('自动修炼请求失败：' + error.message)
    } finally {
      isAutoCultivating.value = false
    }
  }

  // 一键突破
  const cultivateUntilBreakthrough = async () => {
    try {
      const response = await apiClient.post('/api/cultivation/breakthrough', {})
      
      if (response.success) {
        // 同步最新数据
        await syncCultivationData()
        
        // 记录日志
        if (logRef.value) {
          logRef.value.addLog(`一键突破成功`)
          if (response.message) {
            logRef.value.addLog(response.message)
          }
        }
        
        // 更新统计
        statsStore.breakthroughCount += 1
        message.success(response.message || '突破成功！')
      } else {
        message.warning(response.error || '突破失败')
      }
    } catch (error) {
      message.error('突破请求失败：' + error.message)
    }
  }

  // 从后端获取最新修炼数据
  const syncCultivationData = async () => {
    try {
      const response = await apiClient.get('/api/cultivation/data')
      if (response.success && response.data) {
        const data = response.data
        playerInfoStore.level = data.level
        playerInfoStore.realm = data.realm
        playerInfoStore.cultivation = data.cultivation
        playerInfoStore.maxCultivation = data.maxCultivation
        playerInfoStore.spirit = data.spirit
        playerInfoStore.cultivationRate = data.cultivationRate
        playerInfoStore.spiritRate = data.spiritRate
      }
    } catch (error) {
      console.error('Failed to sync cultivation data:', error)
    }
  }

  // 绑定自动修炼停止快捷键
  const handleKeyboard = (event) => {
    if (event.code === 'Space') {
      event.preventDefault()
      cultivate()
    }
  }

  onMounted(() => {
    // 页面加载时同步修炼数据
    syncCultivationData()
    // 每10秒同步一次数据
    const interval = setInterval(syncCultivationData, 10000)
    // 绑定键盘快捷键
    window.addEventListener('keydown', handleKeyboard)
    
    onUnmounted(() => {
      clearInterval(interval)
      window.removeEventListener('keydown', handleKeyboard)
    })
  })
</script>

<style scoped>
  .n-space {
    width: 100%;
  }

  .n-button {
    margin-bottom: 12px;
  }

  .n-collapse {
    margin-top: 12px;
  }
</style>