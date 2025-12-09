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
  import LogPanel from '../components/LogPanel.vue'
  import { triggerRandomEvent, eventTypes } from '../plugins/events'
  import { getRealmName } from '../plugins/realm'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  const persistenceStore = usePersistenceStore()
  
  const message = useMessage()
  const explorationLogs = ref([])
  const isExploring = ref(false)
  const explorationWorker = ref(null)
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
  const startExploration = () => {
    if (isExploring.value) return
    
    isExploring.value = true
    addExplorationLog('开始探索...')
    
    if (explorationWorker.value) {
      explorationWorker.value.terminate()
    }
    
    explorationWorker.value = new Worker(new URL('../workers/exploration.js', import.meta.url))
    explorationWorker.value.onmessage = e => {
      const { type, data } = e.data
      if (type === 'exploration_update') {
        // 更新探索状态
        addExplorationLog(data.log)
      } else if (type === 'exploration_event') {
        // 触发事件
        triggerEvent(data.event)
      } else if (type === 'exploration_end') {
        // 探索结束
        finishExploration()
      }
    }
    
    explorationWorker.value.postMessage({
      type: 'start',
      explorationTime: 10000, // 10秒探索时间
      luck: playerInfoStore.luck
    })
  }

  // 停止探索
  const stopExploration = () => {
    if (explorationWorker.value) {
      explorationWorker.value.terminate()
      explorationWorker.value = null
    }
    
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
      case eventTypes.ITEM_FOUND:
        addExplorationLog(`发现了${event.item.name}！`)
        break
      case eventTypes.SPIRIT_STONE_FOUND:
        addExplorationLog(`发现了${event.amount}灵石！`)
        break
      case eventTypes.HERB_FOUND:
        addExplorationLog(`发现了${event.herb.name}！`)
        break
      case eventTypes.PILL_RECIPE_FRAGMENT_FOUND:
        addExplorationLog(`发现了丹方残页！`)
        break
      case eventTypes.BATTLE_ENCOUNTER:
        addExplorationLog(`遭遇了${event.enemy.name}！`)
        break
    }
  }

  // 处理事件选择
  const handleEventChoice = (choice) => {
    if (!currentEvent.value) return
    
    const event = currentEvent.value
    showEventModal.value = false
    
    // 根据选择处理事件
    switch (event.type) {
      case eventTypes.ITEM_FOUND:
        // 获得物品
        inventoryStore.items.push(event.item)
        statsStore.itemsFound += 1
        addExplorationLog(`获得了${event.item.name}`)
        break
      case eventTypes.SPIRIT_STONE_FOUND:
        // 获得灵石
        inventoryStore.spiritStones += event.amount
        addExplorationLog(`获得了${event.amount}灵石`)
        break
      case eventTypes.HERB_FOUND:
        // 获得灵草
        const existingHerb = inventoryStore.herbs.find(h => h.id === event.herb.id)
        if (existingHerb) {
          existingHerb.count += event.amount
        } else {
          inventoryStore.herbs.push({
            ...event.herb,
            count: event.amount
          })
        }
        addExplorationLog(`获得了${event.amount}个${event.herb.name}`)
        break
      case eventTypes.PILL_RECIPE_FRAGMENT_FOUND:
        // 获得丹方残页
        if (!pillsStore.pillFragments[event.recipeId]) {
          pillsStore.pillFragments[event.recipeId] = 0
        }
        pillsStore.pillFragments[event.recipeId] += event.fragments
        
        // 检查是否可以合成完整丹方
        const recipe = pillsStore.pillRecipes.find(r => r.id === event.recipeId)
        if (recipe && pillsStore.pillFragments[event.recipeId] >= recipe.fragmentsNeeded) {
          pillsStore.pillFragments[event.recipeId] -= recipe.fragmentsNeeded
          if (!pillsStore.pillRecipes.includes(event.recipeId)) {
            pillsStore.pillRecipes.push(event.recipeId)
            statsStore.unlockedPillRecipes += 1
          }
          addExplorationLog(`获得了完整的${recipe.name}丹方！`)
        } else {
          addExplorationLog(`获得了${event.fragments}片丹方残页`)
        }
        break
      case eventTypes.BATTLE_ENCOUNTER:
        // 战斗事件 - 这里简单处理，实际应该进入战斗界面
        const win = Math.random() > 0.5 // 简单的胜负判定
        if (win) {
          addExplorationLog(`战胜了${event.enemy.name}，获得了奖励！`)
          // 简单奖励
          const reward = Math.floor(Math.random() * 100) + 50
          inventoryStore.spiritStones += reward
        } else {
          addExplorationLog(`败给了${event.enemy.name}`)
        }
        break
    }
    
    currentEvent.value = null
    statsStore.explorationCount += 1
  }

  // 结束探索
  const finishExploration = () => {
    if (explorationWorker.value) {
      explorationWorker.value.terminate()
      explorationWorker.value = null
    }
    
    isExploring.value = false
    addExplorationLog('探索结束')
  }

  onUnmounted(() => {
    if (explorationWorker.value) {
      explorationWorker.value.terminate()
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