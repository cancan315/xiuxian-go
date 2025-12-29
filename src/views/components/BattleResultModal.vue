<template>
  <!-- 战斗结果弹窗 -->
  <n-modal
    v-model:show="show"
    :title="battleResultData?.battle_ended ? '战斗结果' : '战斗进行中...'"
    preset="card"
    :type="battleResultData?.battle_ended ? (battleResultData?.victory ? 'success' : 'error') : 'default'"
    :positive-text="battleResultData?.battle_ended ? (battleResultData?.victory ? '继续战斗' : '返回') : ''"
    :negative-text="battleResultData?.battle_ended ? (battleResultData?.victory ? '返回' : '') : ''"
    :show-icon="battleResultData?.battle_ended"
    :close-on-esc="true"
    :mask-closable="true"
    style="width: 90%; max-width: 700px;"
    @positive-click="handlePositiveClick"
    @negative-click="handleNegativeClick"
    @update:show="handleModalClose"
  >
    <template v-if="battleResultData">
      <n-space vertical size="large">
        <!-- 结果标题（仅战斗结束时显示） -->
        <div v-if="battleResultData.battle_ended" style="text-align: center;">
          <n-tag 
            :type="battleResultData.victory ? 'success' : 'error'"
            size="large"
            style="font-size: 18px; padding: 10px 20px;"
          >
            {{ battleResultData.victory ? '🎉 胜利！' : '😔 失败' }}
          </n-tag>
        </div>

        <!-- 奖励信息（仅战斗结束且胜利时显示） -->
        <div v-if="battleResultData.battle_ended && battleResultData.victory && battleResultData.rewards && battleResultData.rewards.length > 0">
          <n-divider style="margin: 8px 0;">获得奖励</n-divider>
          <n-space vertical style="width: 100%;">
            <div v-for="(reward, index) in battleResultData.rewards" :key="index">
              <n-tag type="warning" size="large">
                {{ reward.type === 'spirit_stone' ? `灵石 +${reward.amount}` : reward.type }}
              </n-tag>
            </div>
          </n-space>
        </div>

        <!-- 战斗统计信息 -->
        <div class="stats-section">
          <n-divider style="margin: 8px 0;">战斗统计</n-divider>
          <n-space vertical style="width: 100%;">
            <!-- 回合数 -->
            <div style="display: flex; justify-content: space-between;">
              <span><strong>回合数：</strong></span>
              <span>{{ battleResultData.round }}</span>
            </div>

            <!-- 你的血量 -->
            <div>
              <div style="display: flex; justify-content: space-between; margin-bottom: 4px;">
                <span><strong>你的血量：</strong></span>
                <span>{{ battleResultData.player_health?.toFixed(0) || 0 }}</span>
              </div>
              <div class="health-bar-container">
                <div 
                  class="health-bar player-health"
                  :style="{ width: getPlayerHealthPercentage() + '%' }"
                >
                  <span class="health-text" v-if="getPlayerHealthPercentage() > 10">
                    {{ getPlayerHealthPercentage().toFixed(1) }}%
                  </span>
                </div>
              </div>
            </div>

            <!-- 对手血量 -->
            <div>
              <div style="display: flex; justify-content: space-between; margin-bottom: 4px;">
                <span><strong>对手血量：</strong></span>
                <span>{{ battleResultData.opponent_health?.toFixed(0) || 0 }}</span>
              </div>
              <div class="health-bar-container">
                <div 
                  class="health-bar opponent-health"
                  :style="{ width: getOpponentHealthPercentage() + '%' }"
                >
                  <span class="health-text" v-if="getOpponentHealthPercentage() > 10">
                    {{ getOpponentHealthPercentage().toFixed(1) }}%
                  </span>
                </div>
              </div>
            </div>
          </n-space>
        </div>

        <!-- 战斗日志 -->
        <div class="logs-section">
          <n-divider style="margin: 8px 0;">战斗过程</n-divider>
          <div 
            ref="battleLogsContainer"
            class="logs-container"
          >
            <div v-for="(log, index) in [...(battleResultData.logs || [])].reverse()" :key="index">
              <div 
                :class="['battle-log-item', { 'new-log': index === 0 }]"
              >
                {{ log }}
              </div>
            </div>
          </div>
        </div>
      </n-space>
    </template>
  </n-modal>
</template>

<script setup>
import { computed, ref } from 'vue'
import { NModal, NSpace, NTag, NDivider } from 'naive-ui'

// 定义 props
const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  battleResultData: {
    type: Object,
    default: null
  }
})

// 定义 emit 事件
const emit = defineEmits(['update:show', 'continue-battle', 'close'])

// 战斗日志容器 ref
const battleLogsContainer = ref(null)

// 处理 show 属性变化
const show = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

// 计算玩家血量百分比
const getPlayerHealthPercentage = () => {
  if (!props.battleResultData) return 0
  const maxHealth = props.battleResultData.player_max_health || 100
  const currentHealth = props.battleResultData.player_health || 0
  return Math.max(0, Math.min(100, (currentHealth / maxHealth) * 100))
}

// 计算对手血量百分比
const getOpponentHealthPercentage = () => {
  if (!props.battleResultData) return 0
  const maxHealth = props.battleResultData.opponent_max_health || 100
  const currentHealth = props.battleResultData.opponent_health || 0
  return Math.max(0, Math.min(100, (currentHealth / maxHealth) * 100))
}

// 处理正按钮点击
const handlePositiveClick = () => {
  if (props.battleResultData?.battle_ended && props.battleResultData?.victory) {
    emit('continue-battle')
  }
  closeBattleResultModal()
}

// 处理负按钮点击
const handleNegativeClick = () => {
  closeBattleResultModal()
}

// 处理弹窗关闭
const handleModalClose = (value) => {
  if (!value) {
    closeBattleResultModal()
  }
}

// 关闭弹窗
const closeBattleResultModal = () => {
  emit('close')
}

// 暴露 ref 给父组件（如果需要）
defineExpose({
  battleLogsContainer
})
</script>

<style scoped>
.stats-section {
  margin-top: 8px;
}

.logs-section {
  margin-top: 8px;
}

.logs-container {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 4px;
  padding: 8px;
}

/* 血条容器 */
.health-bar-container {
  width: 100%;
  height: 24px;
  background-color: #e8e8e8;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid #d0d0d0;
  position: relative;
}

/* 血条样式 */
.health-bar {
  height: 100%;
  transition: width 0.3s ease-out;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  font-size: 12px;
  font-weight: bold;
  color: white;
  text-shadow: 1px 1px 1px rgba(0, 0, 0, 0.3);
}

/* 玩家血条（绿色） */
.player-health {
  background: linear-gradient(90deg, #52c41a 0%, #73d13d 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

/* 对手血条（蓝色） */
.opponent-health {
  background: linear-gradient(90deg, #1890ff 0%, #40a9ff 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

/* 血条文字 */
.health-text {
  z-index: 1;
  font-size: 12px;
}

/* 战斗日志项动画 */
@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 战斗日志样式 */
.battle-log-item {
  padding: 8px;
  margin: 4px 0;
  border-radius: 4px;
  font-size: 14px;
  background-color: rgba(0, 0, 0, 0.05);
  animation: slideDown 0.3s ease-out;
  transition: background-color 0.3s ease;
}

/* 最新日志高亮显示 */
.battle-log-item.new-log {
  background-color: rgba(52, 211, 153, 0.2);
  border-left: 3px solid #34d399;
  padding-left: 5px;
  font-weight: 500;
}

.battle-log-item:hover {
  background-color: rgba(0, 0, 0, 0.08);
}
</style>