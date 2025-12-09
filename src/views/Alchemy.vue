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
                  <n-tag type="info">{{ pillGrades[recipe.grade].name }}</n-tag>
                  <n-tag type="warning">{{ pillTypes[recipe.type].name }}</n-tag>
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
          <n-list-item v-for="material in selectedRecipe.materials" :key="material.herb">
            <n-space justify="space-between">
              <n-space>
                <span>{{ getHerbName(material.herb) }}</span>
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
        :disabled="!selectedRecipe || !checkMaterials(selectedRecipe)"
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
  import { usePersistenceStore } from '../stores/persistence'
  import { ref, computed, onMounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import { pillRecipes, tryCreatePill, calculatePillEffect, pillGrades, pillTypes } from '../plugins/pills'
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
  const selectedRecipe = ref(null)
  const showRecipeDetail = ref(false)
  const recipeDetail = ref(null)
  const logRef = ref(null)

  // 解锁的丹方列表
  const unlockedRecipes = computed(() => {
    return pillRecipes.filter(recipe => pillsStore.pillRecipes.includes(recipe.id))
  })

  // 当前选中丹方的效果
  const currentEffect = computed(() => {
    if (!selectedRecipe.value) return { value: 0, duration: 0, successRate: 0 }
    const recipe = selectedRecipe.value
    const grade = pillGrades[recipe.grade]
    const effect = calculatePillEffect(recipe, playerInfoStore.level)
    return {
      ...effect,
      successRate: grade.successRate
    }
  })

  // 选择丹方
  const selectRecipe = (recipe) => {
    selectedRecipe.value = recipe
  }

  // 获取灵草名称
  const getHerbName = (herbId) => {
    const herbs = [
      { id: 'spirit_grass', name: '灵精草' },
      { id: 'cloud_flower', name: '云雾花' },
      { id: 'thunder_root', name: '雷击根' },
      { id: 'dragon_breath_herb', name: '龙息草' },
      { id: 'immortal_jade_grass', name: '仙玉草' },
      { id: 'dark_yin_grass', name: '玄阴草' },
      { id: 'nine_leaf_lingzhi', name: '九叶灵芝' },
      { id: 'purple_ginseng', name: '紫金参' },
      { id: 'frost_lotus', name: '寒霜莲' },
      { id: 'fire_heart_flower', name: '火心花' },
      { id: 'moonlight_orchid', name: '月华兰' },
      { id: 'sun_essence_flower', name: '日精花' },
      { id: 'five_elements_grass', name: '五行草' },
      { id: 'phoenix_feather_herb', name: '凤羽草' },
      { id: 'celestial_dew_grass', name: '天露草' }
    ]
    
    const herb = herbs.find(h => h.id === herbId)
    return herb ? herb.name : herbId
  }

  // 获取材料状态（拥有数量/需要数量）
  const getMaterialStatus = (material) => {
    const ownedCount = inventoryStore.herbs.filter(h => h.id === material.herb).length
    return `${ownedCount}/${material.count}`
  }

  // 检查材料是否充足
  const checkMaterials = (recipe) => {
    for (const material of recipe.materials) {
      const ownedCount = inventoryStore.herbs.filter(h => h.id === material.herb).length
      if (ownedCount < material.count) {
        return false
      }
    }
    return true
  }

  // 分组的丹方列表
  const groupedRecipes = computed(() => {
    const complete = []
    const incomplete = []
    
    pillRecipes.forEach(recipe => {
      const fragments = pillsStore.pillFragments[recipe.id] || 0
      if (pillsStore.pillRecipes.includes(recipe.id)) {
        complete.push({ ...recipe, fragments, fragmentsNeeded: recipe.fragmentsNeeded })
      } else {
        incomplete.push({ ...recipe, fragments, fragmentsNeeded: recipe.fragmentsNeeded })
      }
    })
    
    return { complete, incomplete }
  })

  // 显示丹方详情
  const showRecipeDetails = (recipe) => {
    recipeDetail.value = recipe
    showRecipeDetail.value = true
  }

  // 炼制丹药
  const craftPill = () => {
    const recipe = selectedRecipe.value
    if (!recipe || !pillsStore.pillRecipes.includes(recipe.id)) {
      message.error('未掌握该丹方')
      return
    }
    
    // 检查材料是否足够
    if (!checkMaterials(recipe)) {
      message.error('材料不足')
      return
    }
    
    // 计算成功率（受幸运值影响）
    const grade = pillGrades[recipe.grade]
    const successRate = grade.successRate * playerInfoStore.luck * playerInfoStore.alchemyRate
    
    // 尝试炼制
    if (Math.random() > successRate) {
      message.error('炼制失败')
      if (logRef.value) {
        logRef.value.addLog(`炼制${recipe.name}失败`)
      }
      return
    }
    
    // 消耗材料
    recipe.materials.forEach(material => {
      for (let i = 0; i < material.count; i++) {
        const index = inventoryStore.herbs.findIndex(h => h.id === material.herb)
        if (index > -1) {
          inventoryStore.herbs.splice(index, 1)
        }
      }
    })
    
    // 创建丹药
    const effect = calculatePillEffect(recipe, playerInfoStore.level)
    const pill = {
      id: `${recipe.id}_${Date.now()}`,
      name: recipe.name,
      description: recipe.description,
      type: 'pill',
      effect
    }
    
    inventoryStore.items.push(pill)
    pillsStore.pillsCrafted++
    message.success('炼制成功')
    if (logRef.value) {
      logRef.value.addLog(`成功炼制${recipe.name}`)
    }
  }

  // 购买丹方残页
  const buyFragment = (recipeId) => {
    const recipe = pillRecipes.find(r => r.id === recipeId)
    if (!recipe) return
    
    const fragmentPrice = recipe.grade * 100 // 根据丹方等级定价
    if (inventoryStore.spiritStones < fragmentPrice) {
      message.error('灵石不足')
      return
    }
    
    // 扣除灵石
    inventoryStore.spiritStones -= fragmentPrice
    
    // 获得残页
    if (!pillsStore.pillFragments[recipeId]) {
      pillsStore.pillFragments[recipeId] = 0
    }
    pillsStore.pillFragments[recipeId]++
    
    // 检查是否可以合成完整丹方
    if (pillsStore.pillFragments[recipeId] >= recipe.fragmentsNeeded) {
      pillsStore.pillFragments[recipeId] -= recipe.fragmentsNeeded
      if (!pillsStore.pillRecipes.includes(recipeId)) {
        pillsStore.pillRecipes.push(recipeId)
        statsStore.unlockedPillRecipes += 1
      }
      message.success(`成功合成${recipe.name}丹方！`)
    } else {
      message.success(`购买成功，当前拥有${pillsStore.pillFragments[recipeId]}片残页`)
    }

  }
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