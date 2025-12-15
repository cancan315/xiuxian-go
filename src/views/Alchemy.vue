<template>
  <n-card title="丹药炼制">
    <n-space vertical>
      <template v-if="unlockedRecipes.length > 0">
        <n-divider>丹方选择</n-divider>
        <!-- 丹方选择 -->
        <n-grid :cols="2" :x-gap="12">
          <n-grid-item v-for="recipe in unlockedRecipes" :key="recipe.id">
            <n-card :title="recipe.name" size="small">
              <n-space vertical>
                <n-text depth="3">{{ recipe.description }}</n-text>
                <n-space>
                  <n-tag type="info">{{ recipe.gradeName }}</n-tag>
                  <n-tag type="warning">{{ recipe.typeName }}</n-tag>
                </n-space>
                <n-button
                  @click="selectRecipe(recipe)"
                  block
                  :type="selectedRecipe?.id === recipe.id ? 'primary' : 'default'"
                >
                  {{ selectedRecipe?.id === recipe.id ? '已选择' : '选择' }}
                </n-button>
              </n-space>
            </n-card>
          </n-grid-item>
        </n-grid>
      </template>
      <n-space vertical v-else>
        <n-empty description="暂未掌握任何丹方" />
      </n-space>
      <!-- 材料需求 -->
      <template v-if="selectedRecipe">
        <n-divider>材料需求</n-divider>
        <n-list>
          <n-list-item v-for="material in selectedRecipe.materials" :key="material.herbId">
            <n-space justify="space-between">
              <n-space>
                <span>{{ material.herbName }}</span>
                <n-tag size="small">需要数量: {{ material.count }}</n-tag>
              </n-space>
              <n-tag
                :type="getMaterialStatus(material) === `${material.count}/${material.count}` ? 'success' : 'warning'"
              >
                拥有: {{ getMaterialStatus(material) }}
              </n-tag>
            </n-space>
          </n-list-item>
        </n-list>
      </template>
      <!-- 效果预览 -->
      <template v-if="selectedRecipe">
        <n-divider>效果预览</n-divider>
        <n-descriptions bordered :column="2">
          <n-descriptions-item label="丹药介绍">
            {{ selectedRecipe.description }}
          </n-descriptions-item>
          <n-descriptions-item label="效果数值">+{{ (currentEffect.value * 100).toFixed(1) }}%</n-descriptions-item>
          <n-descriptions-item label="持续时间">{{ Math.floor(currentEffect.duration / 60) }}分钟</n-descriptions-item>
          <n-descriptions-item label="成功率">{{ (currentEffect.successRate * 100).toFixed(1) }}%</n-descriptions-item>
        </n-descriptions>
      </template>
      <!-- 炼制按钮 -->
      <n-button
        class="craft-button"
        type="primary"
        block
        v-if="selectedRecipe"
        :disabled="!selectedRecipe || !checkMaterials(selectedRecipe) || loading"
        :loading="loading"
        @click="craftPill"
      >
        {{ !checkMaterials(selectedRecipe) ? '材料不足' : '开始炼制' }}
      </n-button>
    </n-space>
    <log-panel v-if="selectedRecipe" ref="logRef" title="炼丹日志" />
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
  import { ref, computed, onMounted } from 'vue'
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
  const selectedRecipe = ref(null)
  const showRecipeDetail = ref(false)
  const recipeDetail = ref(null)
  const logRef = ref(null)
  const loading = ref(false)
  const allRecipes = ref([])
  const configs = ref(null)

  // 初始化：获取后端配置
  const initAlchemy = async () => {
    try {
      loading.value = true
      const response = await APIService.apiCall('/api/alchemy/recipes', {
        method: 'GET',
        params: {
          playerLevel: playerInfoStore.level
        }
      })
      if (response.success) {
        allRecipes.value = response.data.recipes
      }
    } catch (error) {
      console.error('初始化炼丹系统失败:', error)
      message.error('初始化炼丹系统失败')
    } finally {
      loading.value = false
    }
  }

  // 解锁的丹方列表
  const unlockedRecipes = computed(() => {
    return allRecipes.value.filter(recipe => recipe.isUnlocked)
  })

  // 当前选中丹方的效果
  const currentEffect = computed(() => {
    if (!selectedRecipe.value) return { value: 0, duration: 0, successRate: 0 }
    return selectedRecipe.value.currentEffect || { value: 0, duration: 0, successRate: 0 }
  })

  // 选择丹方
  const selectRecipe = (recipe) => {
    selectedRecipe.value = recipe
  }

  // 获取灵草名称
  const getHerbName = (herbId) => {
    const material = selectedRecipe.value?.materials?.find(m => m.herbId === herbId)
    if (material) {
      return material.herbName
    }
    return herbId
  }

  // 获取材料状态（拥有数量/需要数量）
  const getMaterialStatus = (material) => {
    const ownedCount = inventoryStore.herbs.filter(h => h.id === material.herbId).length
    return `${ownedCount}/${material.count}`
  }

  // 检查材料是否充足
  const checkMaterials = (recipe) => {
    if (!recipe || !recipe.materials) return false
    for (const material of recipe.materials) {
      const ownedCount = inventoryStore.herbs.filter(h => h.id === material.herbId).length
      if (ownedCount < material.count) {
        return false
      }
    }
    return true
  }

  // 炼制丹药
  const craftPill = async () => {
    const recipe = selectedRecipe.value
    if (!recipe || !recipe.isUnlocked) {
      message.error('未掌握该丹方')
      return
    }
    
    // 检查材料是否足够
    if (!checkMaterials(recipe)) {
      message.error('材料不足')
      return
    }
    
    try {
      loading.value = true
      
      // 构建灵草库存数据
      const inventoryHerbs = {}
      inventoryStore.herbs.forEach(h => {
        if (inventoryHerbs[h.id]) {
          inventoryHerbs[h.id]++
        } else {
          inventoryHerbs[h.id] = 1
        }
      })
      
      const response = await APIService.apiCall('/api/alchemy/craft', {
        method: 'POST',
        data: {
          recipeId: recipe.id,
          playerLevel: playerInfoStore.level || 1,
          unlockedRecipes: pillsStore.pillRecipes || [],
          inventoryHerbs: inventoryHerbs,
          luck: playerInfoStore.luck || 1.0,
          alchemyRate: playerInfoStore.alchemyRate || 1.0
        }
      })
      
      if (response.success && response.data.success) {
        message.success(`炼制成功！成功率: ${(response.data.successRate * 100).toFixed(1)}%`)
        if (logRef.value) {
          logRef.value.addLog(`成功炼制${recipe.name}`)
        }
        
        // 根据后端返回的消耗材料从前端库存扣除
        if (response.data.consumedHerbs) {
          Object.entries(response.data.consumedHerbs).forEach(([herbId, count]) => {
            for (let i = 0; i < count; i++) {
              const index = inventoryStore.herbs.findIndex(h => h.id === herbId)
              if (index > -1) {
                inventoryStore.herbs.splice(index, 1)
              }
            }
          })
        }
        
        // 更新炼制次数统计
        pillsStore.pillsCrafted++
        
        // 刷新丹方列表
        await initAlchemy()
      } else {
        message.error(response.data?.message || '炼制失败')
        if (logRef.value) {
          logRef.value.addLog(`炼制${recipe.name}失败: ${response.data?.message || '成功率不足'}`)
        }
      }
    } catch (error) {
      console.error('炼制失败:', error)
      message.error('炼制失败')
    } finally {
      loading.value = false
    }
  }

  // 购买丹方残页
  const buyFragment = async (recipeId) => {
    const recipe = allRecipes.value.find(r => r.id === recipeId)
    if (!recipe) return
    
    try {
      loading.value = true
      const response = await APIService.apiCall('/api/alchemy/buy-fragment', {
        method: 'POST',
        data: {
          recipeId: recipeId,
          quantity: 1,
          currentFragments: pillsStore.pillFragments[recipeId] || 0,
          unlockedRecipes: pillsStore.pillRecipes || []
        }
      })
      
      if (response.success && response.data.success) {
        // 更新前端状态
        pillsStore.pillFragments[recipeId] = response.data.fragmentsOwned
        
        if (response.data.recipeUnlocked) {
          message.success(`成功合成${recipe.name}丹方！`)
          if (!pillsStore.pillRecipes.includes(recipeId)) {
            pillsStore.pillRecipes.push(recipeId)
          }
          statsStore.unlockedPillRecipes += 1
        } else {
          message.success(`购买成功，当前拥有${response.data.fragmentsOwned}片残页`)
        }
        
        // 刷新丹方列表
        await initAlchemy()
      } else {
        message.error(response.data?.message || '购买失败')
      }
    } catch (error) {
      console.error('购买残页失败:', error)
      message.error('购买残页失败')
    } finally {
      loading.value = false
    }
  }

  // 生命周期：组件挂载时初始化
  onMounted(() => {
    initAlchemy()
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

  .craft-button {
    position: relative;
    overflow: hidden;
  }

  @keyframes success-ripple {
    0% {
      transform: scale(0);
      opacity: 1;
    }
    100% {
      transform: scale(4);
      opacity: 0;
    }
  }

  @keyframes fail-shake {
    0%,
    100% {
      transform: translateX(0);
    }
    25% {
      transform: translateX(-10px);
    }
    75% {
      transform: translateX(10px);
    }
  }

  .success-animation::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 20px;
    height: 20px;
    background: rgba(0, 255, 0, 0.3);
    border-radius: 50%;
    transform: translate(-50%, -50%);
    animation: success-ripple 1s ease-out;
  }

  .fail-animation {
    animation: fail-shake 0.5s ease-in-out;
  }
</style>