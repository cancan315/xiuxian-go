<template>
  <n-layout>
    <n-layout-header bordered>
      <n-page-header>
        <template #title>排行榜</template>
        <template #extra>
          <n-button @click="fetchLeaderboard">刷新</n-button>
        </template>
      </n-page-header>
    </n-layout-header>
    <n-layout-content class="leaderboard-content">
      <n-card :bordered="false">
        <n-spin :show="loading">
          <n-empty v-if="leaderboard.length === 0 && !loading" description="暂无排行榜数据">
            <template #extra>
              <n-button @click="fetchLeaderboard">刷新</n-button>
            </template>
          </n-empty>
          
          <n-data-table
            v-else
            :columns="columns"
            :data="leaderboard"
            :pagination="false"
            :bordered="false"
            :single-line="false"
          />
        </n-spin>
      </n-card>
    </n-layout-content>
  </n-layout>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import APIService from '../services/api'

const message = useMessage()
const loading = ref(false)
const leaderboard = ref([])

// 表格列定义
const columns = [
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
    width: 200
  },
  {
    title: '境界',
    key: 'realm',
    width: 200
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

// 获取排行榜数据
const fetchLeaderboard = async () => {
  try {
    loading.value = true
    const data = await APIService.getLeaderboard()
    leaderboard.value = data
  } catch (error) {
    console.error('获取排行榜失败:', error)
    message.error('获取排行榜失败')
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取数据
onMounted(() => {
  fetchLeaderboard()
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