<template>
  <div class="duel-container">
    <!-- 斗法主卡片 -->
    <n-card title="斗法">
      <!-- 主标签页导航 -->
      <n-tabs type="line" animated>
        <!-- 玩家对战标签页 -->
        <n-tab-pane name="pvp" tab="道友斗法">
          <!-- PVP组件 -->
          <DuelPVP 
            @challenge-player="handleChallengePlayer" 
            @view-player-info="handleViewPlayerInfo"
          />
        </n-tab-pane>

        <!-- 妖兽挑战标签页 -->
        <n-tab-pane name="pve" tab="降服妖兽">
          <!-- PVE组件 -->
          <DuelPVE 
            @challenge-monster="handleChallengeMonster" 
            @view-monster-info="handleViewMonsterInfo"
          />
        </n-tab-pane>

        <!-- 斗法战绩标签页 -->
        <n-tab-pane name="records" tab="战绩">
          <!-- 战绩组件 -->
          <DuelRecords />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 战斗模拟弹窗 -->
    <BattleModal
      :show="showBattleModal"
      :title="battleTitle"
      :battle-logs="battleLogs"
      :battle-result="battleResult"
      @close="handleCloseBattleModal"
      @claim-rewards="handleClaimRewards"
    />

    <!-- 玩家信息弹窗 -->
    <PlayerInfoModal
      :show="showPlayerInfoModal"
      :player="selectedPlayer"
      @update:show="showPlayerInfoModal = $event"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { NCard, NTabs, NTabPane } from 'naive-ui'
import { usePlayerInfoStore } from '../stores/playerInfo'
import APIService from '../services/api'
import { getAuthToken } from '../stores/db'
import { simulateBattle } from '../plugins/battleSimulator'
import { useDuel } from './composables/useDuel'

// 导入子组件
import DuelPVP from './components/DuelPVP.vue'
import DuelPVE from './components/DuelPVE.vue'
import DuelRecords from './components/DuelRecords.vue'
import BattleModal from './components/BattleModal.vue'
import PlayerInfoModal from './components/PlayerInfoModal.vue'

// 获取玩家信息存储
const playerInfoStore = usePlayerInfoStore()

// 使用斗法业务逻辑组合式函数
const {
  showBattleModal,
  battleLogs,
  battleResult,
  battleTitle,
  selectedPlayer,
  showPlayerInfoModal,
  challengePlayer,
  challengeMonster,
  viewPlayerInfo,
  closeBattleModal,
  claimRewards
} = useDuel()

// 监听 token 变化，当 token 可用时初始化子组件
const duelPVPRef = ref(null)
watch(
  () => getAuthToken(),
  (token) => {
    if (token && duelPVPRef.value) {
      duelPVPRef.value.refreshOpponents?.()
    }
  }
)

/**
 * 处理挑战玩家事件
 */
const handleChallengePlayer = async (opponent) => {
  await challengePlayer(opponent)
}

/**
 * 处理查看玩家信息事件
 */
const handleViewPlayerInfo = (player) => {
  viewPlayerInfo(player)
}

/**
 * 处理挑战妖兽事件
 */
const handleChallengeMonster = async (monster) => {
  await challengeMonster(monster)
}

/**
 * 处理查看妖兽信息事件
 */
const handleViewMonsterInfo = (monster) => {
  // 可以显示更详细的妖兽信息
  window.$message.info(`${monster.name}：${monster.description || '暂无详细描述'}`)
}

/**
 * 处理关闭战斗弹窗事件
 */
const handleCloseBattleModal = () => {
  closeBattleModal()
}

/**
 * 处理领取奖励事件
 */
const handleClaimRewards = async () => {
  await claimRewards()
}

// 初始化
onMounted(() => {
  // 初始化完成
})
</script>

<style scoped>
.duel-container {
  max-width: 1200px;
  margin: 0 auto;
}
</style>
