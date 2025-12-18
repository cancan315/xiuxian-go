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
      </n-space>
      <n-divider>修炼详情</n-divider>
      <n-descriptions bordered>
        <n-descriptions-item label="灵力获取速率">{{ playerInfoStore.spiritGainRate }} / 秒</n-descriptions-item>
        <n-descriptions-item label="修炼效率">{{ playerInfoStore.cultivationGain }} 修为 / 次</n-descriptions-item>
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
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { BookOutline } from '@vicons/ionicons5' // 导入图标组件
  import LogPanel from '../components/LogPanel.vue'
  import APIService from '../services/api'
  import { getAuthToken } from '../stores/db'

  const playerInfoStore = usePlayerInfoStore()
  
  const message = useMessage()
  const isAutoCultivating = ref(false)
  const logRef = ref(null)

  // 修炼消耗和获得（从后端获取准确数据）
  const cultivationCost = computed(() => {
    // 从后端获取准确的修炼消耗数据
    return playerInfoStore.cultivationCost || 1
  })

  // 修炼获得的修为
  const cultivationGain = computed(() => {
    // 从后端获取准确的修炼获得数据
    return playerInfoStore.cultivationGain || 1
  })

  // 基础灵力获取速率
  const baseGainRate = 1

  // 计算突破所需成本
  const calculateBreakthroughCost = () => {
    return playerInfoStore.maxCultivation - playerInfoStore.cultivation
  }

  // 打坐修练方法
  const cultivate = async () => {
    try {
      // ✅ 使用正确的URL（不需要前面的 /api）
      const token = getAuthToken();
      const response = await APIService.post('/cultivation/single', {}, token)
      
      if (response.success) {
        // 后端返回的数据字段名
        playerInfoStore.spirit -= response.spiritCost
        playerInfoStore.cultivation = response.currentCultivation
        
        // 记录日志
        if (logRef.value) {
          logRef.value.addLog(`修炼获得 ${response.cultivationGain.toFixed(1)} 点修为`)
        }
        
        // 检查是否有突破
        if (response.breakthrough) {
          const bt = response.breakthrough
          playerInfoStore.level = bt.newLevel
          playerInfoStore.realm = bt.newRealm
          playerInfoStore.maxCultivation = bt.newMaxCultivation
          playerInfoStore.spirit += bt.spiritReward
          playerInfoStore.spiritRate = bt.newSpiritRate
          playerInfoStore.breakthroughCount += 1
          message.success(bt.message)
          if (logRef.value) {
            logRef.value.addLog(bt.message)
          }
        }
        message.success('修炼成功，获得 ' + response.cultivationGain.toFixed(1) + ' 点修为')
        // 增加修炼时间统计
        playerInfoStore.totalCultivationTime += 1
        return true
      } else {
        message.warning(response.error || '修炼失败')
        return false
      }
    } catch (error) {
      message.error('修炼请求失败：' + error.message)
      if (logRef.value) {
        logRef.value.addLog(`修炼失败：${error.message}`)
      }
      return false
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
    
    // 检查灵力是否满足一次修炼消耗
    if (playerInfoStore.spirit < cultivationCost.value) {
      message.error('灵力不足，修炼失败')
      return
    }
    
    // 开始自动修练
    isAutoCultivating.value = true
    message.success('开始自动修练')
    
    // 循环调用单次修炼直到灵力不足或用户停止
    while (isAutoCultivating.value) {
      // 检查灵力是否满足一次修炼消耗
      if (playerInfoStore.spirit < cultivationCost.value) {
        isAutoCultivating.value = false
        message.warning('灵力不足，修炼停止')
        if (logRef.value) {
          logRef.value.addLog('灵力不足，修炼停止')
        }
        break
      }
      
      // 调用单次修炼
      const success = await cultivate()
      if (!success) {
        isAutoCultivating.value = false
        break
      }
      
      // 短暂延迟，避免过快调用
      await new Promise(resolve => setTimeout(resolve, 3000))
    }
  }

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
    // 绑定键盘快捷键
    window.addEventListener('keydown', handleKeyboard)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyboard)
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