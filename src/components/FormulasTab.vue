<template>
  <div class="formulas-container">
    <!-- 加载状态 -->
    <n-spin :show="loading" description="加载中...">
      <!-- 丹方标签页 -->
      <n-tabs type="segment" v-if="formulas.length > 0 || loading">
        <n-tab-pane name="complete" tab="完整丹方">
          <div class="formulas-grid">
            <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="completeFormulas.length > 0">
              <n-grid-item v-for="formula in completeFormulas" :key="formula.recipeId">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ getFormulaName(formula.recipeId) }}</span>
                      <n-tag type="success" size="small">完整</n-tag>
                    </n-space>
                  </template>
                  <p style="margin: 0; color: var(--text-color-secondary); font-size: 13px;">{{ getFormulaDescription(formula.recipeId) }}</p>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-empty v-else description="暂无完整丹方" />
          </div>
        </n-tab-pane>

        <n-tab-pane name="incomplete" tab="残缺丹方">
          <div class="formulas-grid">
            <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="incompleteFormulas.length > 0">
              <n-grid-item v-for="formula in incompleteFormulas" :key="formula.recipeId">
                <n-card hoverable>
                  <template #header>
                    <n-space justify="space-between">
                      <span>{{ getFormulaName(formula.recipeId) }}</span>
                      <n-tag type="warning" size="small">残缺</n-tag>
                    </n-space>
                  </template>
                  <n-space vertical>
                    <p style="margin: 0 0 12px 0; color: var(--text-color-secondary); font-size: 13px;">{{ getFormulaDescription(formula.recipeId) }}</p>
                    <n-progress
                      type="line"
                      :percentage="Number(((formula.count / getFragmentsNeeded(formula.recipeId)) * 100).toFixed(2))"
                      :show-indicator="true"
                      indicator-placement="inside"
                    >
                      收集进度: {{ formula.count }}/{{ getFragmentsNeeded(formula.recipeId) }}
                    </n-progress>
                    <!-- 购买残页按钮 -->
                    <n-button 
                      block
                      type="primary"
                      size="small"
                      :loading="buyingFragmentFor === formula.recipeId"
                      @click="buyFragment(formula.recipeId)"
                    >
                      购买残页 (50000灵石/片)
                    </n-button>
                  </n-space>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-empty v-else description="暂无残缺丹方" />
          </div>
        </n-tab-pane>
      </n-tabs>

      <!-- 空状态 -->
      <n-empty v-else-if="!loading" description="暂无丹方数据" />

      <!-- 分页控件 -->
      <div class="pagination-container" v-if="pagination.total > 0">
        <n-space justify="space-between" align="center">
          <span class="pagination-info">
            总计: {{ pagination.total }} 个丹方 | 当前页: {{ pagination.page }}/{{ pagination.totalPages }}
          </span>
          <n-space>
            <n-button 
              quaternary 
              circle
              :disabled="pagination.page <= 1 || loading"
              @click="goToPreviousPage"
            >
              <template #icon>
                <span>❮</span>
              </template>
            </n-button>
            
            <span class="page-indicator">{{ pagination.page }}/{{ pagination.totalPages }}</span>
            
            <n-button 
              quaternary 
              circle
              :disabled="pagination.page >= pagination.totalPages || loading"
              @click="goToNextPage"
            >
              <template #icon>
                <span>❯</span>
              </template>
            </n-button>
          </n-space>
        </n-space>
      </div>
    </n-spin>
  </div>
</template>

<script setup>
  import { ref, computed, onMounted } from 'vue'
  import { useMessage, NSpin, NTabs, NTabPane, NGrid, NGridItem, NCard, NSpace, NButton, NEmpty, NProgress, NTag } from 'naive-ui'
  import { pillRecipes } from '../plugins/pills'
  import APIService from '../services/api'
  import { getAuthToken } from '../stores/db'

  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  const message = useMessage()
  const loading = ref(false)
  const buyingFragmentFor = ref(null) // 跟踪正在购买的丹方ID
  const formulas = ref([])
  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0
  })

  // 完整丹方（已收集满）
  const completeFormulas = computed(() => {
    return formulas.value.filter(formula => {
      const needed = getFragmentsNeeded(formula.recipeId)
      return formula.count >= needed
    })
  })

  // 残缺丹方（未收集满）
  const incompleteFormulas = computed(() => {
    return formulas.value.filter(formula => {
      const needed = getFragmentsNeeded(formula.recipeId)
      return formula.count < needed
    })
  })

  // 获取丹方名称
  const getFormulaName = (recipeId) => {
    const recipe = pillRecipes.find(r => r.id === recipeId)
    return recipe ? recipe.name : recipeId
  }

  // 获取丹方描述
  const getFormulaDescription = (recipeId) => {
    const recipe = pillRecipes.find(r => r.id === recipeId)
    return recipe ? recipe.description : '未知丹方'
  }

  // 获取所需残页数
  const getFragmentsNeeded = (recipeId) => {
    const recipe = pillRecipes.find(r => r.id === recipeId)
    return recipe ? recipe.fragmentsNeeded : 0
  }

  // 加载丹方数据
  const loadFormulas = async (page = 1) => {
    try {
      loading.value = true
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        return
      }

      console.log('[FormulasTab] 开始加载丹方数据，页码:', page)
      const response = await APIService.getFormulasList(token, {
        page: page,
        pageSize: pagination.value.pageSize,
        sort: 'id',
        order: 'asc'
      })

      if (response && response.formulas) {
        // ✅ API 已自动转换字段为小驼峰，直接使用
        formulas.value = response.formulas
        pagination.value = response.pagination
        
        console.log(`[FormulasTab] 成功加载丹方数据，数量: ${response.formulas.length}, 总数: ${response.pagination.total}`)
        console.log('[FormulasTab] 处理后的丹方数据:', formulas.value)
      } else {
        message.error('获取丹方列表失败')
      }
    } catch (error) {
      console.error('[FormulasTab] 加载丹方数据失败:', error)
      message.error('加载丹方数据失败: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  // 上一页
  const goToPreviousPage = () => {
    if (pagination.value.page > 1) {
      loadFormulas(pagination.value.page - 1)
    }
  }

  // 下一页
  const goToNextPage = () => {
    if (pagination.value.page < pagination.value.totalPages) {
      loadFormulas(pagination.value.page + 1)
    }
  }

  // 组件挂载时加载数据
  onMounted(() => {
    console.log('[FormulasTab] 组件挂载，开始加载丹方数据')
    loadFormulas(1)
  })

  // 购买残页
  const buyFragment = async (recipeId) => {
    try {
      buyingFragmentFor.value = recipeId
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        return
      }

      console.log('[FormulasTab] 开始购买残页:', recipeId)
      const response = await APIService.post('/alchemy/buy-fragment', {
        recipeId: recipeId,
        quantity: 1,
        currentFragments: props.playerInfoStore.pillFragments[recipeId] || 0,
        unlockedRecipes: props.playerInfoStore.pillRecipes || []
      }, token)

      if (response.success && response.data && response.data.success) {
        // 更新前端信息
        props.playerInfoStore.pillFragments[recipeId] = response.data.fragmentsOwned
        
        if (response.data.recipeUnlocked) {
          message.success(`成功合成${getFormulaName(recipeId)}丹方！`)
          // 更新已解锁丹方列表
          if (!props.playerInfoStore.pillRecipes.includes(recipeId)) {
            props.playerInfoStore.pillRecipes.push(recipeId)
          }
          props.playerInfoStore.unlockedPillRecipes += 1
        } else {
          message.success(`购买成功，当前拥有${response.data.fragmentsOwned}片残页`)
        }
        
        // 刷新丹方列表
        await loadFormulas(pagination.value.page)
      } else {
        message.error(response.data?.message || '购买失败')
      }
    } catch (error) {
      console.error('[FormulasTab] 购买残页失败:', error)
      message.error('购买残页失败: ' + error.message)
    } finally {
      buyingFragmentFor.value = null
    }
  }
</script>

<style scoped>
.formulas-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.formulas-grid {
  min-height: 200px;
}

.pagination-container {
  padding: 16px;
  background-color: var(--n-bg-color);
  border-radius: 4px;
  border: 1px solid var(--n-border-color);
}

.pagination-info {
  font-size: 13px;
  color: var(--text-color-secondary);
}

.page-indicator {
  font-size: 12px;
  min-width: 40px;
  text-align: center;
  color: var(--text-color-secondary);
}
</style>