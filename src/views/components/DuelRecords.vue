<template>
  <div class="records-section">
    <n-space vertical>
      <!-- 战绩统计 -->
      <n-card title="战绩统计" size="small">
        <n-space>
          <n-statistic label="总战斗次数" :value="battleStats.totalBattles" />
          <n-statistic label="胜率" :value="`${battleStats.winRate}%`" />
          <n-statistic label="当前连胜" :value="battleStats.currentWinStreak" />
          <n-statistic label="最高连胜" :value="battleStats.maxWinStreak" />
        </n-space>
      </n-card>

      <!-- 详细记录 -->
      <n-card title="战斗记录" size="small">
        <n-data-table 
          :columns="recordColumns" 
          :data="battleRecords" 
          :pagination="{ pageSize: 10 }"
          :loading="isLoadingRecords"
          striped
        />
      </n-card>
    </n-space>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NCard, NSpace, NDataTable, NStatistic, NTag } from 'naive-ui'
import APIService from '../../services/api'
import { getAuthToken } from '../../stores/db'

// 状态管理
const isLoadingRecords = ref(false)
const battleRecords = ref([])

/**
 * 战绩统计
 */
const battleStats = ref({
  totalBattles: 0,  // 总战斗次数
  wins: 0,          // 胜利次数
  losses: 0,        // 失败次数
  winRate: 0,       // 胜率
  currentWinStreak: 0, // 当前连胜
  maxWinStreak: 0   // 最高连胜
})

/**
 * 战斗记录表格列定义
 */
const recordColumns = [
  {
    title: '战斗类型',
    key: 'battleType',
    render(row) {
      // 将战斗类型代码转换为中文描述
      const typeMap = { pvp: '道友斗法', pve: '降服妖兽' }
      return typeMap[row.battleType] || row.battleType
    }
  },
  {
    title: '对手',
    key: 'opponent'
  },
  {
    title: '结果',
    key: 'result',
    render(row) {
      // 根据结果渲染不同颜色的标签
      return h(NTag, { 
        type: row.result === '胜利' ? 'success' : 'error',
        size: 'small'
      }, { default: () => row.result })
    }
  },
  {
    title: '奖励',
    key: 'rewards',
    render(row) {
      return row.rewards || '无'
    }
  },
  {
    title: '战斗时间',
    key: 'time',
    width: 180
  }
]

/**
 * 加载战斗记录
 */
const loadBattleRecords = async () => {
  isLoadingRecords.value = true
  try {
    const token = getAuthToken()
    const response = await APIService.getBattleRecords(token)
    if (response.success) {
      battleRecords.value = response.data.records
      battleStats.value = response.data.stats
    }
  } catch (error) {
    console.error('获取战斗记录失败:', error)
  } finally {
    isLoadingRecords.value = false
  }
}

// 初始化加载
onMounted(() => {
  loadBattleRecords()
})
</script>

<style scoped>
.records-section {
  padding: 8px;
}
</style>
