<template>
  <n-card title="丹药炼制">
    <n-space vertical>
      <!-- 刷新按钮 -->
      <n-button @click="initAlchemy" :loading="loading">
        🔄 刷新丹方列表
      </n-button>
      <!-- 已解锁丹方 -->
      <template v-if="unlockedRecipes.length > 0">
        <n-divider>已掌握丹方</n-divider>
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
      <!-- 未获得过丹方 -->
      <template v-if="unlockedRecipes.length === 0 && incompleteRecipes.length === 0">
        <n-empty description="暂未掌握任何丹方" />
      </template>
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
          <n-descriptions-item label="效果数值">+{{ currentEffect.value  }}</n-descriptions-item>
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
          <!-- 残缺丹方 -->
      <template v-if="incompleteRecipes.length > 0">
        <n-divider>残缺丹方</n-divider>
        <n-grid :cols="2" :x-gap="12">
          <n-grid-item v-for="recipe in incompleteRecipes" :key="recipe.id">
            <n-card :title="recipe.name" size="small">
              <n-space vertical>
                <n-text depth="3">{{ recipe.description }}</n-text>
                <n-space>
                  <n-tag type="info">{{ recipe.gradeName }}</n-tag>
                  <n-tag type="warning">{{ recipe.typeName }}</n-tag>
                </n-space>
                <n-progress
                  type="line"
                  :percentage="(recipe.currentFragments / recipe.fragmentsNeeded) * 100"
                  :show-indicator="false"
                />
                <n-text depth="3" size="small">
                  残页进度: {{ recipe.currentFragments }}/{{ recipe.fragmentsNeeded }}
                </n-text>
              </n-space>
            </n-card>
          </n-grid-item>
        </n-grid>
      </template>
  </n-card>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { getAuthToken } from '../stores/db'
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import LogPanel from '../components/LogPanel.vue'
  import APIService from '../services/api'

  const playerInfoStore = usePlayerInfoStore()
  
  const message = useMessage()
  const selectedRecipe = ref(null)
  const showRecipeDetail = ref(false)
  const recipeDetail = ref(null)
  const logRef = ref(null)
  const loading = ref(false)
  const allRecipes = ref([])
  const configs = ref(null)

  // 初始化：获取后端配置和火草数据
  const initAlchemy = async () => {
    try {
      loading.value = true
      const token = getAuthToken()
        
      // 1. 加载丹方数据
      const response = await APIService.get('/alchemy/recipes', { playerLevel: playerInfoStore.level }, token)
      if (response.success) {
        allRecipes.value = response.data.recipes || []
          
        // ✅ 从后端返回的数据中更新玚家的已解锁丹方
        if (response.data.playerStats && response.data.playerStats.recipesUnlocked) {
          playerInfoStore.pillRecipes = Object.keys(response.data.playerStats.recipesUnlocked).filter(
            id => response.data.playerStats.recipesUnlocked[id] === true
          )
          playerInfoStore.pillFragments = response.data.playerStats.fragments || {}
        }
          
        console.log('[Alchemy] 成功加载丹方列表，已解锁数量:', playerInfoStore.pillRecipes.length)
      }
        
      // 2. 加载火草数据
      await loadHerbs()
    } catch (error) {
      console.error('[Alchemy] 初始化炼丹系统失败:', error)
      message.error('初始化炼丹系统失败')
    } finally {
      loading.value = false
    }
  }
  
  // 加载灵草数据
  const loadHerbs = async () => {
    try {
      const token = getAuthToken()
      if (!token) {
        console.warn('[Alchemy] 未找到认证令牌，无法加载灵草')
        return
      }
    
      console.log('[Alchemy] 开始加载灵草数据')
          
      // 一次性加载所有灵草数据（不分页）
      let allHerbs = []
      let page = 1
      let totalPages = 1
    
      while (page <= totalPages) {
        const response = await APIService.getHerbsList(token, {
          page: page,
          pageSize: 100,
          sort: 'id',
          order: 'asc'
        })
    
        if (response && response.herbs) {
          // 转换字段映射
          const processedHerbs = response.herbs.map(herb => ({
            id: herb.id || herb.ID,
            userId: herb.userId || herb.UserID,
            herbId: herb.herbId || herb.HerbID,
            name: herb.name || herb.Name,
            count: herb.count || herb.Count || 0,
            quality: herb.quality || herb.Quality || 'common'
          }))
          allHerbs = allHerbs.concat(processedHerbs)
              
          // 更新分页信息
          if (response.pagination) {
            totalPages = response.pagination.totalPages || 1
            page++
          } else {
            break
          }
        } else {
          break
        }
      }
    
      // ✅ 按 herbId 聚合灵草数据（合并相同种类）
      const groupedByHerbId = {}
      allHerbs.forEach(herb => {
        if (groupedByHerbId[herb.herbId]) {
          // 已存在该种灵草，累加数量
          groupedByHerbId[herb.herbId].count += herb.count
        } else {
          // 新灵草种类
          groupedByHerbId[herb.herbId] = { ...herb }
        }
      })
        
      // 转换为数组存储
      const aggregatedHerbs = Object.values(groupedByHerbId)
        
      // 更新 playerInfoStore 中的 herbs 数据
      playerInfoStore.herbs = aggregatedHerbs
      console.log('[Alchemy] 成功加载灵草数据，总数:', allHerbs.length)
      console.log('[Alchemy] 聚合后的灵草数据:', aggregatedHerbs)
      console.log('[Alchemy] 按herbId分组统计:', groupedByHerbId)
    } catch (error) {
      console.error('[Alchemy] 加载灵草数据失败:', error)
      // 不中断情流，继续执行
    }
  }

  // 解锁的丹方列表
  const unlockedRecipes = computed(() => {
    return allRecipes.value.filter(recipe => recipe.isUnlocked)
  })

  // 残缺丹方列表（未解锁但有残页）
  const incompleteRecipes = computed(() => {
    return allRecipes.value.filter(recipe => !recipe.isUnlocked && recipe.currentFragments > 0)
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
    const herb = playerInfoStore.herbs.find(h => h.herbId === material.herbId)
    const ownedCount = herb ? herb.count : 0
    return `${ownedCount}/${material.count}`
  }

  // 检查材料是否充足
  const checkMaterials = (recipe) => {
    if (!recipe || !recipe.materials) return false
    for (const material of recipe.materials) {
      const herb = playerInfoStore.herbs.find(h => h.herbId === material.herbId)
      const ownedCount = herb ? herb.count : 0
      console.log(`[Alchemy] 检查材料: ${material.herbId}, 拥有: ${ownedCount}, 需要: ${material.count}`)
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
      
      // 构建火草库存数据
      const inventoryHerbs = {}
      playerInfoStore.herbs.forEach(h => {
        // 每个火草对象的count是该种火草的总数量
        inventoryHerbs[h.herbId] = h.count
      })
      
      const token = getAuthToken()
      const response = await APIService.post('/alchemy/craft', {
        recipeId: recipe.id,
        playerLevel: playerInfoStore.level || 1,
        unlockedRecipes: playerInfoStore.pillRecipes || [],
        inventoryHerbs: inventoryHerbs,
        luck: playerInfoStore.luck || 1.0,
        alchemyRate: playerInfoStore.alchemyRate || 1.0
      }, token)
      
      if (response.success && response.data.success) {
        message.success(`炼制成功！成功率: ${(response.data.successRate * 100).toFixed(1)}%`)
        if (logRef.value) {
          logRef.value.addLog(`成功炼制${recipe.name}`)
        }
        
        // ✅ 移除前端手动扣除逻辑，直接刷新所有数据
        // 后端已经消耗了灵草，initAlchemy 会重新加载最新的灵草数据
        
        // 更新炼制次数统计
        playerInfoStore.pillsCrafted++
        
        // ✅ 在刷新数据之前稍作等待，确保后端事务完成
        await new Promise(resolve => setTimeout(resolve, 200))
        
        // 刷新丹方列表（包括重新加载灵草）
        await initAlchemy()
      } else {
        // ✅ 炼丹失败也要刷新灵草数据，因为后端已经消耗了灵草
        const failMsg = response.data?.message || '炼制失败'
        message.warning(failMsg)
        if (logRef.value) {
          logRef.value.addLog(`炼制${recipe.name}失败: ${failMsg}`)
        }
        
        // ✅ 刷新灵草数据（后端已消耗灵草）
        await new Promise(resolve => setTimeout(resolve, 200))
        await initAlchemy()
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
      const token = getAuthToken()
      const response = await APIService.post('/alchemy/buy-fragment', {
        recipeId: recipeId,
        quantity: 1,
        currentFragments: playerInfoStore.pillFragments[recipeId] || 0,
        unlockedRecipes: playerInfoStore.pillRecipes || []
      }, token)
      
      if (response.success && response.data.success) {
        // 更新前端状态
        playerInfoStore.pillFragments[recipeId] = response.data.fragmentsOwned
        
        if (response.data.recipeUnlocked) {
          message.success(`成功合成${recipe.name}丹方！`)
          if (!playerInfoStore.pillRecipes.includes(recipeId)) {
            playerInfoStore.pillRecipes.push(recipeId)
          }
          playerInfoStore.unlockedPillRecipes += 1
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