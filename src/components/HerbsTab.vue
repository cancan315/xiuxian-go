<template>
  <div class="herbs-container">
    <!-- 加载状态 -->
    <n-spin :show="loading" description="加载中...">
      <!-- 灵草网格 -->
      <n-grid :cols="2" :x-gap="12" :y-gap="8" v-if="paginatedHerbs.length > 0">
        <n-grid-item v-for="herb in paginatedHerbs" :key="herb.herbId">
          <n-card hoverable>
            <template #header>
              <n-space justify="space-between">
                <span>{{ herb.name }}</span>
                <n-tag type="info" size="small">{{ herb.count }}</n-tag>
              </n-space>
            </template>
            <n-space vertical :size="8">
              <div class="herb-detail">
                <span class="label">ID:</span>
                <span class="value">{{ herb.herbId }}</span>
              </div>
              <div class="herb-detail">
                <span class="label">描述:</span>
                <span class="value">{{ herb.description || '暂无描述' }}</span>
              </div>
              <div class="herb-detail">
                <span class="label">基础价值:</span>
                <n-tag type="success" size="small">{{ herb.baseValue || 0 }}</n-tag>
              </div>
              <div class="herb-detail">
                <span class="label">品质:</span>
                <n-tag :type="getQualityTagType(herb.quality)" size="small">
                  {{ herb.qualityName || '未知' }}
                </n-tag>
              </div>
            </n-space>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 空状态 -->
      <n-empty v-else-if="!loading" description="暂无灵草" />

      <!-- 分页控件 -->
      <div class="pagination-container" v-if="pagination.total > 0">
        <n-space justify="space-between" align="center">
          <span class="pagination-info">
            总计: {{ pagination.total }} 种灵草 | 当前页: {{ pagination.page }}/{{ pagination.totalPages }}
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

  const props = defineProps({
    playerInfoStore: {
      type: Object,
      required: true
    }
  })

  const message = useMessage()
  const loading = ref(false)
  const allHerbs = ref([])  // 存储所有灵草数据（不分页）
  const currentPage = ref(1)  // 当前页码
  const pageSize = ref(20)  // 每页显示的灵草种类数

  // 品质映射（与后端config.go保持一致）
  const qualityMap = {
    common: '普通',
    uncommon: '优质',
    rare: '稀有',
    epic: '极品',
    legendary: '仙品'
  }

  // 获取品质标签类型
  const getQualityTagType = (quality) => {
    const typeMap = {
      common: 'default',
      uncommon: 'success',
      rare: 'info',
      epic: 'warning',
      legendary: 'error'
    }
    return typeMap[quality] || 'default'
  }

  // 计算属性：按herbId聚合所有灵草
  const groupedHerbs = computed(() => {
    const herbMap = {}
    
    allHerbs.value.forEach(herb => {
      const key = herb.herbId
      if (herbMap[key]) {
        // 相同灵草，数量累加
        herbMap[key].count += herb.count || 1
      } else {
        // 新灵草类型
        herbMap[key] = {
          id: herb.id,
          herbId: herb.herbId,
          name: herb.name,
          count: herb.count || 1,
          description: herb.description,
          baseValue: herb.baseValue,
          quality: herb.quality || 'common',
          qualityName: herb.qualityName || qualityMap[herb.quality] || '普通'
        }
      }
    })
    
    // 按count降序排列，数量多的排在前面
    return Object.values(herbMap).sort((a, b) => b.count - a.count)
  })

  // 计算属性：当前页显示的灵草
  const paginatedHerbs = computed(() => {
    const start = (currentPage.value - 1) * pageSize.value
    const end = start + pageSize.value
    return groupedHerbs.value.slice(start, end)
  })

  // 计算属性：分页信息
  const pagination = computed(() => {
    const total = groupedHerbs.value.length
    const totalPages = Math.ceil(total / pageSize.value)
    return {
      page: currentPage.value,
      pageSize: pageSize.value,
      total: total,
      totalPages: totalPages
    }
  })

  // 加载所有灵草数据
  const loadAllHerbs = async () => {
    try {
      loading.value = true
      const token = getAuthToken()
      if (!token) {
        message.error('未找到认证令牌')
        return
      }

      console.log('[HerbsTab] 开始加载所有灵草数据')
      
      // 加载第一页获取总数
      let allData = []
      let page = 1
      let totalPages = 1

      while (page <= totalPages) {
        console.log(`[HerbsTab] 加载第 ${page} 页灵草数据`)
        const response = await APIService.getHerbsList(token, {
          page: page,
          pageSize: 100,  // 每次请求100条以减少请求次数
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
            count: herb.count || herb.Count || 1,
            description: herb.description || herb.Description || '',
            baseValue: herb.baseValue || herb.BaseValue || 0,
            quality: herb.quality || herb.Quality || 'common',
            qualityName: herb.qualityName || herb.QualityName || qualityMap[herb.quality] || '普通'
          }))

          allData = allData.concat(processedHerbs)
          totalPages = response.pagination.totalPages
          page++
        } else {
          break
        }
      }

      allHerbs.value = allData
      currentPage.value = 1  // 重置到第一页

      // 同时更新 playerInfoStore 的 herbs 数据
      props.playerInfoStore.herbs = allData

      console.log(`[HerbsTab] 成功加载所有灵草数据，总数: ${allData.length}`)
      console.log('[HerbsTab] 聚合后的灵草数据:', groupedHerbs.value)
    } catch (error) {
      console.error('[HerbsTab] 加载灵草数据失败:', error)
      message.error('加载灵草数据失败: ' + error.message)
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

  // 组件挂载时加载数据
  onMounted(() => {
    console.log('[HerbsTab] 组件挂载，开始加载灵草数据')
    loadAllHerbs()
  })
</script>

<style scoped>
.herbs-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.herb-detail {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.herb-detail .label {
  color: var(--text-color-secondary);
  min-width: 50px;
}

.herb-detail .value {
  color: var(--text-color);
  flex: 1;
  word-break: break-all;
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