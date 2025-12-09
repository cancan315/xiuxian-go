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
  import { getRealmName } from '../plugins/realm'
  import LogPanel from '../components/LogPanel.vue'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  
  const message = useMessage()
  const cultivationWorker = ref(null)
  const lastSaveTime = ref(Date.now())
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

  // 境界名称
  const realmName = computed(() => {
    return getRealmName(playerInfoStore.level - 1)?.name || playerInfoStore.realm
  })

  // 计算修炼效率加成
  const cultivationBonus = computed(() => {
    let bonus = playerInfoStore.cultivationRate
    
    // 装备加成（由于getArtifactBonus方法不存在，这里暂时不考虑装备加成）
    // bonus *= combatStore.getArtifactBonus ? combatStore.getArtifactBonus('cultivationRate') : 1
    
    return bonus
  })

  // 计算灵力获取加成
  const spiritBonus = computed(() => {
    let bonus = playerInfoStore.spiritRate
    
    // 装备加成（由于getArtifactBonus方法不存在，这里暂时不考虑装备加成）
    // bonus *= combatStore.getArtifactBonus ? combatStore.getArtifactBonus('spiritRate') : 1
    
    return bonus
  })

  // 计算突破所需成本
  const calculateBreakthroughCost = () => {
    return playerInfoStore.maxCultivation - playerInfoStore.cultivation
  }

  // 打坐修炼方法
  const cultivate = () => {
    if (playerInfoStore.spirit >= cultivationCost.value) {
      // 消耗灵力
      playerInfoStore.spirit -= cultivationCost.value
      
      // 增加修为
      playerInfoStore.cultivation += cultivationGain.value * cultivationBonus.value
      
      // 检查是否可以突破
      if (playerInfoStore.cultivation >= playerInfoStore.maxCultivation) {
        tryBreakthrough()
      }
      
      // 记录日志
      if (logRef.value) {
        logRef.value.addLog(`修炼获得 ${cultivationGain.value * cultivationBonus.value} 点修为`)
      }
      
      // 增加修炼时间统计
      statsStore.totalCultivationTime += 1
      
      const now = Date.now()
      if (now - lastSaveTime.value > 30000) { // 30秒保存一次
        lastSaveTime.value = now
      }
    } else {
      message.warning('灵力不足，无法修炼')
    }
  }

  // 切换自动修炼
  const toggleAutoCultivation = () => {
    isAutoCultivating.value = !isAutoCultivating.value
    if (isAutoCultivating.value) {
      startCultivation()
      message.success('开始自动修炼')
    } else {
      stopCultivation()
      message.info('停止自动修炼')
    }
  }

  // 一键突破
  const cultivateUntilBreakthrough = () => {
    const cost = calculateBreakthroughCost()
    if (playerInfoStore.spirit >= cost) {
      playerInfoStore.spirit -= cost
      playerInfoStore.cultivation = playerInfoStore.maxCultivation
      tryBreakthrough()
      
      // 记录日志
      if (logRef.value) {
        logRef.value.addLog(`消耗 ${cost} 灵力一键突破`)
      }
    } else {
      message.warning('灵力不足，无法一键突破')
    }
  }

  // 开始修炼
  const startCultivation = () => {
    if (cultivationWorker.value) return
    
    cultivationWorker.value = new Worker(new URL('../workers/cultivation.js', import.meta.url))
    cultivationWorker.value.onmessage = e => {
      const { type, data } = e.data
      if (type === 'cultivate') {
        // 增加修为
        playerInfoStore.cultivation += data.amount * cultivationBonus.value
        
        // 增加灵力
        playerInfoStore.spirit += data.spirit * spiritBonus.value
        
        // 增加修炼时间统计
        statsStore.totalCultivationTime += 1
        
        // 检查是否可以突破
        if (playerInfoStore.cultivation >= playerInfoStore.maxCultivation) {
          tryBreakthrough()
        }
        
        const now = Date.now()
        if (now - lastSaveTime.value > 30000) { // 30秒保存一次
          lastSaveTime.value = now
        }
      }
    }
    cultivationWorker.value.postMessage({ 
      type: 'start',
      baseCultivationRate: 1,
      baseSpiritRate: 1
    })
  }

  // 停止修炼
  const stopCultivation = () => {
    if (cultivationWorker.value) {
      cultivationWorker.value.terminate()
      cultivationWorker.value = null
    }
  }

  // 尝试突破
  const tryBreakthrough = () => {
    // 境界等级对应的境界名称和修为上限
    const realmsLength = getRealmName().length
    // 检查是否可以突破到下一个境界
    if (playerInfoStore.level < realmsLength) {
      const nextRealm = getRealmName(playerInfoStore.level)
      // 更新境界信息
      playerInfoStore.level += 1
      playerInfoStore.realm = nextRealm.name // 使用完整的境界名称（如：练气一层）
      playerInfoStore.maxCultivation = nextRealm.maxCultivation
      playerInfoStore.cultivation = 0 // 重置修为值
      statsStore.breakthroughCount += 1 // 增加突破次数
      // 解锁新境界
      if (!playerInfoStore.unlockedRealms.includes(nextRealm.name)) {
        playerInfoStore.unlockedRealms.push(nextRealm.name)
      }
      // 突破奖励
      playerInfoStore.spirit += 100 * playerInfoStore.level // 获得灵力奖励
      playerInfoStore.spiritRate *= 1.2 // 提升灵力获取倍率
      message.success(`恭喜突破至${nextRealm.name}！`)
      
      // 记录日志
      if (logRef.value) {
        logRef.value.addLog(`恭喜突破至${nextRealm.name}！`)
      }
      
      return true
    }
    return false
  }

  onMounted(() => {
    startCultivation()
  })

  onUnmounted(() => {
    stopCultivation()
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