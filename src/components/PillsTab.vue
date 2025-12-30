<template>
  <div class="pills-container">
    <!-- 加载状态 -->
    <n-spin :show="loading" description="加载中...">
      <!-- 丹药网格 -->
      <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="paginatedPills.length > 0">
        <n-grid-item v-for="pill in paginatedPills" :key="pill.pillId">
          <n-card hoverable>
            <template #header>
              <n-space justify="space-between">
                <span>{{ pill.name }}</span>
                <n-tag type="info" size="small">{{ pill.count }}</n-tag>
              </n-space>
            </template>
            <n-space vertical :size="8">
              <div class="pill-detail">
                <span class="label">ID:</span>
                <span class="value">{{ pill.pillId }}</span>
              </div>
              <div class="pill-detail">
                <span class="label">描述:</span>
                <span class="value">{{ pill.description || '暂无描述' }}</span>
              </div>
              
              <!-- 效果显示 -->
              <div v-if="pill.effect && pill.effect.type" class="pill-effect">
                <div style="font-size: 12px; color: var(--text-color-secondary);">
                  <span>效果: {{ getStatName(pill.effect.type) }}</span>
                </div>
                <div style="font-size: 12px; margin-top: 4px; color: #1890ff;">
                  <span>+{{ formatStatValue(pill.effect.type, pill.effect.value) }}</span>
                </div>
              </div>
              
              <!-- 服用按钮 -->
              <n-button size="small" type="primary" block @click="usePill(pill)">服用</n-button>
            </n-space>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 空状态 -->
      <n-empty v-else-if="!loading" description="暂无丹药" />

      <!-- 分页控件 -->
      <div class="pagination-container" v-if="pagination.total > 0">
        <n-space justify="space-between" align="center">
          <span class="pagination-info">
            总计: {{ pagination.total }} 种丹药 | 当前页: {{ pagination.page }}/{{ pagination.totalPages }}
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
  import { ref, onMounted, computed } from 'vue'
  import { useMessage, NSpin, NGrid, NGridItem, NCard, NSpace, NButton, NEmpty, NTag } from 'naive-ui'
  import APIService from '../services/api'
  import { getAuthToken } from '../stores/db'
  import { getStatName, formatStatValue } from '../plugins/stats'

  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  const message = useMessage()
  const loading = ref(false)
  const allPills = ref([])  // 存储所有丹药数据（不分页）
  const currentPage = ref(1)  // 当前页码
  const pageSize = ref(20)  // 每页显示的丹药种类数

  // 计算属性：按pillId聚合所有丹药
  const groupedPills = computed(() => {
    const pillMap = {}
    
    allPills.value.forEach(pill => {
      const key = pill.pillId
      if (pillMap[key]) {
        // 相同丹药，数量累加
        pillMap[key].count += 1
      } else {
        // 新丹药类型
        pillMap[key] = {
          id: pill.id,
          pillId: pill.pillId,
          name: pill.name,
          count: 1,
          description: pill.description,
          effect: pill.effect
        }
      }
    })
    
    // 按count降序排列，数量多的排在前面
    return Object.values(pillMap).sort((a, b) => b.count - a.count)
  })

  // 计算属性：当前页显示的丹药
  const paginatedPills = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return groupedPills.value.slice(start, end)
  })

  // 计算属性：分页信息
  const pagination = computed(() => {
    const total = groupedPills.value.length
    const totalPages = Math.ceil(total / pageSize.value)
    return {
      page: currentPage.value,
      pageSize: pageSize.value,
      total: total,
      totalPages: totalPages
    }
  })

  // 加载所有丹药数据
  const loadAllPills = async () => {
    try {
      loading.value = true
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        return
      }

      console.log('[PillsTab] 开始加载所有丹药数据')
      
      // 加载第一页获取总数
      let allData = []
      let page = 1
      let totalPages = 1

      while (page <= totalPages) {
        console.log(`[PillsTab] 加载第 ${page} 页丹药数据`)
        const response = await APIService.getPillsList(token, {
          page: page,
          pageSize: 100,  // 每次请求100条以减少请求次数
          sort: 'id',
          order: 'asc'
        })

        if (response && response.pills) {
          // 转换字段映射
          const processedPills = response.pills.map(pill => ({
            id: pill.id || pill.ID,
            userId: pill.userId || pill.UserID,
            pillId: pill.pillId || pill.PillID,
            name: pill.name || pill.Name,
            description: pill.description || pill.Description || '',
            effect: pill.effect || pill.Effect || null
          }))

          allData = allData.concat(processedPills)
          totalPages = response.pagination.totalPages
          page++
        } else {
          break
        }
      }

      allPills.value = allData
      currentPage.value = 1  // 重置到第一页

      console.log(`[PillsTab] 成功加载所有丹药数据，总数: ${allData.length}`)
      console.log('[PillsTab] 聚合后的丹药数据:', groupedPills.value)
    } catch (error) {
      console.error('[PillsTab] 加载丹药数据失败:', error)
      message.error('加载丹药数据失败: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  // 上一页
  const goToPreviousPage = () => {
    if (currentPage.value > 1) {
      currentPage.value--
    }
  }

  // 下一页
  const goToNextPage = () => {
    if (currentPage.value < pagination.value.totalPages) {
      currentPage.value++
    }
  }

  // 使用丹药
  const usePill = async (pill) => {
    try {
      loading.value = true
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        return
      }

      console.log('[PillsTab] 开始服用丹药:', pill.name)
      
      // 使用聚合后的 pill 的 ID 进行服用（这里取第一个相同类型的丹药 ID）
      // 需要找到第一个对应这个 pillId 的丹药
      const firstPill = allPills.value.find(p => p.pillId === pill.pillId)
      if (!firstPill) {
        message.error('未找到该丹药')
        return
      }

      const response = await APIService.post(`/player/pills/${firstPill.id}/consume`, {}, token)

      if (response && response.success) {
        message.success(`成功服用${pill.name}丹药`)
        console.log('[PillsTab] 服用丹药成功:', response.data)
        
        // 刷新丹药列表
        await loadAllPills()
      } else {
        message.error(response.message || '服用丹药失败')
      }
    } catch (error) {
      console.error('[PillsTab] 服用丹药失败:', error)
      message.error('服用丹药失败: ' + error.message)
    } finally {
      loading.value = false
    }
  }

  // 组件挂载时加载数据
  onMounted(() => {
    console.log('[PillsTab] 组件挂载，开始加载丹药数据')
    loadAllPills()
  })
</script>

<style scoped>
.pills-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.pill-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.pill-detail .label {
  color: var(--text-color-secondary);
  min-width: 50px;
}

.pill-detail .value {
  color: var(--text-color);
  flex: 1;
  word-break: break-all;
}

.pill-effect {
  padding: 8px 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  border-bottom: 1px solid rgba(255,255,255,0.1);
  margin: 8px 0;
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