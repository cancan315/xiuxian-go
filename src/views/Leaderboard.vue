<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>排行榜</template>
        <template #extra>
          <n-button @click="fetchAllLeaderboards">刷新</n-button>
        </template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content class="leaderboard-content">
      <n-card :bordered="false">
        <n-tabs type="line" v-model:value="activeTab" @update:value="onTabChange">
          <!-- 境界排行榜 -->
          <n-tab-pane name="realm" tab="境界排行">
            <n-spin :show="loading.realm">
              <n-empty v-if="leaderboards.realm.length === 0 && !loading.realm" description="暂无排行榜数据">
                <template #extra>
                  <n-button @click="fetchLeaderboardByType('realm')">刷新</n-button>
                </template>
              </n-empty>
              <n-data-table
                v-else
                :columns="realmColumns"
                :data="leaderboards.realm"
                :pagination="pagination.realm"
                :bordered="false"
                :single-line="false"
              />
            </n-spin>
          </n-tab-pane>
          
          <!-- 灵石排行榜 -->
          <n-tab-pane name="spiritStones" tab="灵石排行">
            <n-spin :show="loading.spiritStones">
              <n-empty v-if="leaderboards.spiritStones.length === 0 && !loading.spiritStones" description="暂无排行榜数据">
                <template #extra>
                  <n-button @click="fetchLeaderboardByType('spiritStones')">刷新</n-button>
                </template>
              </n-empty>
              <n-data-table
                v-else
                :columns="spiritStonesColumns"
                :data="leaderboards.spiritStones"
                :pagination="pagination.spiritStones"
                :bordered="false"
                :single-line="false"
              />
            </n-spin>
          </n-tab-pane>
          
          <!-- 装备排行榜 -->
          <n-tab-pane name="equipment" tab="装备排行">
            <n-spin :show="loading.equipment">
              <n-empty v-if="leaderboards.equipment.length === 0 && !loading.equipment" description="暂无排行榜数据">
                <template #extra>
                  <n-button @click="fetchLeaderboardByType('equipment')">刷新</n-button>
                </template>
              </n-empty>
              <n-data-table
                v-else
                :columns="equipmentColumns"
                :data="leaderboards.equipment"
                :pagination="pagination.equipment"
                :bordered="false"
                :single-line="false"
              />
            </n-spin>
          </n-tab-pane>
          
          <!-- 灵宠排行榜 -->
          <n-tab-pane name="pets" tab="灵宠排行">
            <n-spin :show="loading.pets">
              <n-empty v-if="leaderboards.pets.length === 0 && !loading.pets" description="暂无排行榜数据">
                <template #extra>
                  <n-button @click="fetchLeaderboardByType('pets')">刷新</n-button>
                </template>
              </n-empty>
              <n-data-table
                v-else
                :columns="petsColumns"
                :data="leaderboards.pets"
                :pagination="pagination.pets"
                :bordered="false"
                :single-line="false"
              />
            </n-spin>
          </n-tab-pane>
        </n-tabs>
      </n-card>
    </n-layout-content>
  </n-layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import APIService from '../services/api'

// 调试模式：设置为 true 可以在控制台查看详细日志
const DEBUG_MODE = true

const message = useMessage()
const activeTab = ref('realm')

if (DEBUG_MODE) {
  console.log('[排行榜] 调试模式已启用，将在控制台显示详细日志')
}

// 加载状态（分别控制每个分榜）
const loading = ref({
  realm: false,
  spiritStones: false,
  equipment: false,
  pets: false
})

// 排行榜数据（分别存储四个分榜的数据）
const leaderboards = ref({
  realm: [],
  spiritStones: [],
  equipment: [],
  pets: []
})

// 分页配置（为每个分榜单独配置）
const pagination = ref({
  realm: {
    pageSize: 10,
    page: 1,
    pageCount: 1,
    itemCount: 0,
    prefix: (info) => `第 ${info.page} 页`
  },
  spiritStones: {
    pageSize: 10,
    page: 1,
    pageCount: 1,
    itemCount: 0,
    prefix: (info) => `第 ${info.page} 页`
  },
  equipment: {
    pageSize: 10,
    page: 1,
    pageCount: 1,
    itemCount: 0,
    prefix: (info) => `第 ${info.page} 页`
  },
  pets: {
    pageSize: 10,
    page: 1,
    pageCount: 1,
    itemCount: 0,
    prefix: (info) => `第 ${info.page} 页`
  }
})

// 境界排行榜列定义
const realmColumns = [
  {
    title: '排名',
    key: 'rank',
    width: 80,
    render(row, index) {
      const rank = index + 1
      let medal = ''
      if (rank === 1) {
        medal = '🥇'
      } else if (rank === 2) {
        medal = '🥈'
      } else if (rank === 3) {
        medal = '🥉'
      }
      return `${medal} ${rank}`
    }
  },
  {
    title: '道号',
    key: 'playerName',
    width: 120
  },
  {
    title: '境界',
    key: 'realm',
    width: 150
  }
]

// 灵石排行榜列定义
const spiritStonesColumns = [
  {
    title: '排名',
    key: 'rank',
    width: 80,
    render(row, index) {
      const rank = index + 1
      let medal = ''
      if (rank === 1) {
        medal = '🥇'
      } else if (rank === 2) {
        medal = '🥈'
      } else if (rank === 3) {
        medal = '🥉'
      }
      return `${medal} ${rank}`
    }
  },
  {
    title: '道号',
    key: 'playerName',
    width: 120
  },
  {
    title: '灵石',
    key: 'spiritStones',
    width: 150,
    render(row) {
      return `${row.spiritStones} 💠`
    }
  }
]

// 装备排行榜列定义
const equipmentColumns = [
  {
    title: '排名',
    key: 'rank',
    width: 80,
    render(row, index) {
      const rank = index + 1
      let medal = ''
      if (rank === 1) {
        medal = '🥇'
      } else if (rank === 2) {
        medal = '🥈'
      } else if (rank === 3) {
        medal = '🥉'
      }
      return `${medal} ${rank}`
    }
  },
  {
    title: '道号',
    key: 'playerName',
    width: 100
  },
  {
    title: '装备名称',
    key: 'name',
    width: 150
  },
  {
    title: '品质',
    key: 'quality',
    width: 100,
    render(row) {
      const qualityMap = {
        '仙品': '🌠',
        '极品': '💎',
        '稀有': '🌟',
        '优质': '⭐',
        '普通': '📄'
      }
      return `${row.quality} ${qualityMap[row.quality] || ''}`
    }
  },
  {
    title: '强化等级',
    key: 'enhanceLevel',
    width: 100,
    render(row) {
      return `+${row.enhanceLevel || 0}`
    }
  }
]

// 灵宠排行榜列定义
const petsColumns = [
  {
    title: '排名',
    key: 'rank',
    width: 80,
    render(row, index) {
      const rank = index + 1
      let medal = ''
      if (rank === 1) {
        medal = '🥇'
      } else if (rank === 2) {
        medal = '🥈'
      } else if (rank === 3) {
        medal = '🥉'
      }
      return `${medal} ${rank}`
    }
  },
  {
    title: '道号',
    key: 'playerName',
    width: 100
  },
  {
    title: '灵宠名称',
    key: 'name',
    width: 120
  },
  {
    title: '稀有度',
    key: 'rarity',
    width: 100,
    render(row) {
      const rarityMap = {
        '传说': '🎆',
        '史诗': '💎',
        '稀有': '🌟',
        '精良': '⭐',
        '普通': '📄'
      }
      return `${row.rarity} ${rarityMap[row.rarity] || ''}`
    }
  },
  {
    title: '星级',
    key: 'star',
    width: 80,
    render(row) {
      return '★'.repeat(row.star || 0)
    }
  },
  {
    title: '等级',
    key: 'level',
    width: 80,
    render(row) {
      return `Lv.${row.level || 0}`
    }
  }
]

// 获取指定类型的排行榜数据
const fetchLeaderboardByType = async (type) => {
  try {
    loading.value[type] = true
    console.log(`[排行榜] 开始获取${type}排行榜数据...`)
    
    // TODO: 根据类型调用不同的API，这里假设后端有对应的接口
    // 例如: /api/leaderboard/realm, /api/leaderboard/spiritStones 等
    const data = await APIService.getLeaderboard(type)
    
    console.log(`[排行榜] ${type}排行榜数据获取成功`, {
      type,
      count: data?.length || 0,
      data: data
    })
    
    // 处理分页
    leaderboards.value[type] = data || []
    pagination.value[type].itemCount = leaderboards.value[type].length
    pagination.value[type].pageCount = Math.ceil(leaderboards.value[type].length / pagination.value[type].pageSize)
    
    console.log(`[排行榜] ${type}排行榜分页配置`, {
      itemCount: pagination.value[type].itemCount,
      pageCount: pagination.value[type].pageCount,
      pageSize: pagination.value[type].pageSize
    })
  } catch (error) {
    console.error(`[排行榜] 获取${type}排行榜失败:`, error)
    console.error(`[排行榜] 错误详情:`, {
      type,
      error: error.message,
      stack: error.stack
    })
    message.error(`获取${type}排行榜失败`)
  } finally {
    loading.value[type] = false
  }
}

// 获取所有排行榜数据
const fetchAllLeaderboards = async () => {
  console.log('[排行榜] 开始获取所有排行榜数据...')
  const startTime = Date.now()
  
  try {
    await Promise.all([
      fetchLeaderboardByType('realm'),
      fetchLeaderboardByType('spiritStones'),
      fetchLeaderboardByType('equipment'),
      fetchLeaderboardByType('pets')
    ])
    
    const duration = Date.now() - startTime
    console.log('[排行榜] 所有排行榜数据获取完成', {
      耗时: `${duration}ms`,
      数据统计: {
        境界排行: leaderboards.value.realm.length,
        灵石排行: leaderboards.value.spiritStones.length,
        装备排行: leaderboards.value.equipment.length,
        灵宠排行: leaderboards.value.pets.length
      }
    })
  } catch (error) {
    console.error('[排行榜] 获取所有排行榜数据失败:', error)
  }
}

// 标签页切换时的处理
const onTabChange = (name) => {
  console.log(`[排行榜] 切换到${name}排行榜`, {
    标签: name,
    数据条数: leaderboards.value[name]?.length || 0,
    当前页: pagination.value[name].page,
    总页数: pagination.value[name].pageCount
  })
}

// 组件挂载时获取数据
onMounted(() => {
  console.log('[排行榜] 排行榜页面已加载')
  fetchAllLeaderboards()
})
</script>

<style scoped>
.leaderboard-content {
  padding: 16px;
}

.n-card {
  border-radius: 8px;
}
</style>