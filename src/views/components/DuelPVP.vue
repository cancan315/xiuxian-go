<template>
  <div class="pvp-section">
    <!-- 道友对战说明 -->
    <n-alert title="道友对战" type="info" style="margin-bottom: 16px;">
      与其他道友斗法，印证修炼所得！每天20次，消耗灵力，胜利可获得修为和灵石。00:00重置
    </n-alert>
    
    <!-- 道友列表 -->
    <n-card title="可挑战的道友" size="small">
      <n-spin :show="isLoadingOpponents">
        <n-space vertical>
          <n-list bordered>
            <!-- 道友列表项 -->
            <n-list-item v-for="opponent in opponents" :key="opponent.id">
              <n-thing>
                <template #header>
                  <n-space align="center">
                    <span>{{ opponent.name }}</span>
                    <!-- 境界标签 -->
                    <n-tag :type="getRealmTagType(opponent.level)">
                      {{ getRealmName(opponent.level).name }}
                    </n-tag>
                  </n-space>
                </template>
                <template #description>
                  <n-space>
                    <span>修为: {{ opponent.cultivation }}/{{ opponent.maxCultivation }}</span>
                    <span>灵石: {{ opponent.spiritStones }}</span>
                  </n-space>
                </template>
                <template #footer>
                  <n-space justify="end">
                    <!-- 切磋按钮 -->
                    <n-button type="primary" size="small" @click="handleChallengeClick(opponent)" :disabled="isBattleInProgress">
                      切磋
                    </n-button>
                    <!-- 查看信息按钮 -->
                    <n-button size="small" @click="$emit('view-player-info', opponent)">
                      查看信息
                    </n-button>
                  </n-space>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-space>
      </n-spin>
      
      <!-- 刷新道友按钮 -->
      <n-space justify="center" style="margin-top: 16px;">
        <n-button @click="refreshOpponents">刷新道友</n-button>
      </n-space>
    </n-card>

    <!-- 对战记录 -->
    <n-card title="对战记录" size="small" style="margin-top: 16px;">
      <n-data-table 
        :columns="pvpRecordColumns" 
        :data="pvpRecords" 
        :pagination="{ pageSize: 5 }"
        striped
      />
    </n-card>

    <!-- 战斗结果弹窗组件 -->
    <BattleResultModal
      ref="battleResultModalRef"
      v-model:show="showBattleResult"
      :battle-result-data="battleResultData"
      @close="closeBattleResultModal"
      @continue-battle="handleContinueBattle"
    />
  </div>
</template>

<script setup>
import { ref, computed, h, nextTick } from 'vue'
import { NCard, NAlert, NSpace, NButton, NList, NListItem, NThing, NTag, NDataTable, NSpin, useMessage } from 'naive-ui'
import { getRealmName } from '../../plugins/realm'
import APIService from '../../services/api'
import { getAuthToken } from '../../stores/db'
import { usePlayerInfoStore } from '../../stores/playerInfo'
import { getRealmTagType } from '../utils/duelHelper'
import BattleResultModal from './BattleResultModal.vue'

// 获取消息提示实例
const message = useMessage()

// 定义 emit 事件
const emit = defineEmits(['challenge-player', 'view-player-info', 'battle-start', 'battle-end'])

// 获取玩家信息 store
const playerInfoStore = usePlayerInfoStore()

// 状态管理
const isLoadingOpponents = ref(false)
const opponents = ref([])
const pvpRecords = ref([])
const isBattleInProgress = ref(false)
const currentBattleOpponent = ref(null)
const showBattleResult = ref(false)
const battleResultData = ref(null)
const battleResultModalRef = ref(null)

// PVP对战记录表格列定义
const pvpRecordColumns = [
  {
    title: '道友',
    key: 'opponentName',
    width: 150
  },
  {
    title: '结果',
    key: 'result',
    render(row) {
      // 根据结果类型渲染不同颜色的标签
      return h(NTag, { type: row.result === '胜利' ? 'success' : 'error' }, { default: () => row.result })
    }
  },
  {
    title: '获得声望',
    key: 'prestigeGained',
    align: 'center'
  },
  {
    title: '获得灵石',
    key: 'spiritStonesGained',
    align: 'center'
  },
  {
    title: '战斗时间',
    key: 'battleTime',
    width: 180
  }
]

/**
 * 切磋按钮点击处理 - 调用后端战斗器API
 */
const handleChallengeClick = async (opponent) => {
  const token = getAuthToken()
  if (!token) {
    console.warn('[DuelPVP] 未获取到认证令牌')
    return
  }

  try {
    isBattleInProgress.value = true
    currentBattleOpponent.value = opponent

    console.log('[DuelPVP] 开始切磋，对手:', opponent.name)

    // 从 store 获取当前用户ID
    const playerID = playerInfoStore.id
    console.log('[DuelPVP] 从 playerInfoStore 获取用户ID:', {
      playerID,
      storeID: playerInfoStore.id,
      type: typeof playerID
    })

    if (!playerID || isNaN(playerID)) {
      console.error('[DuelPVP] 无效的玩家ID:', playerID)
      return
    }

    // 第一步：从后端数据库获取双方批量战斗属性
    // 调用新的API：/api/duel/battle-attributes
    const battleAttributesResponse = await APIService.getBattleAttributes(
      token,
      playerID, // 玩家ID
      opponent.id  // 对手ID
    )

    if (!battleAttributesResponse.success) {
      console.error('[DuelPVP] 获取战斗属性失败', battleAttributesResponse.message)
      return
    }

    const playerData = battleAttributesResponse.data?.playerData
    const opponentData = battleAttributesResponse.data?.opponentData

    if (!playerData || !opponentData) {
      console.error('[DuelPVP] 战斗属性数据不完整')
      return
    }

    console.log('[DuelPVP] 玩家战斗数据:', playerData)
    console.log('[DuelPVP] 对手战斗数据:', opponentData)

    // 第二步：调用后端批郸丰吐API开始战斗
    const startResponse = await APIService.startPvPBattle(
      token,
      opponent.id,
      playerData,
      opponentData
    )

    if (!startResponse.success) {
      // 检查是否是每日限制错误（HTTP 429）
      if (startResponse.message && startResponse.message.includes('今日斗法次数已达上限')) {
        message.error(startResponse.message)
        console.error('[DuelPVP] 今日斗法次数已达上限')
      } else {
        console.error('[DuelPVP] 开始战斗失败:', startResponse.message)
        message.error(startResponse.message || '开始战斗失败')
      }
      return
    }

    console.log('[DuelPVP] 战斗已初始化，准备执行回合')
    
    // 立即打开弹窗，显示初始状态
    battleResultData.value = {
      round: 0,
      player_health: playerData.baseAttributes?.health || 100,
      player_max_health: playerData.baseAttributes?.health || 100,
      opponent_health: opponentData.baseAttributes?.health || 100,
      opponent_max_health: opponentData.baseAttributes?.health || 100,
      logs: ['战斗已初始化，准备开始！'],
      battle_ended: false,
      victory: false,
      rewards: []
    }
    showBattleResult.value = true

    emit('battle-start', { opponent, battleData: startResponse.data })

    // 第三步：循环执行回合直到战斗结束
    await executeBattleRounds(token, opponent.id)
  } catch (error) {
    console.error('[DuelPVP] 切磋过程中出错:', error)
  } finally {
    isBattleInProgress.value = false
  }
}

/**
 * 执行战斗回合 - 更新时自动滚动到底部
 */
const executeBattleRounds = async (token, opponentId) => {
  try {
    let battleEnded = false
    let victory = false

    while (!battleEnded) {
      // 执行一回合
      const roundResponse = await APIService.executePvPRound(token, opponentId)

      if (!roundResponse.success) {
        console.error('[DuelPVP] 执行回合失败:', roundResponse.message)
        break
      }

      const roundData = roundResponse.data
      console.log('[DuelPVP] 回合数据:', roundData)

      // 实时更新弹窗中的战斗数据
      if (battleResultData.value) {
        // 兼容两种字段名（蛇形和驼峰）
        battleResultData.value.round = roundData?.round || 0
        battleResultData.value.player_health = roundData?.player_health || 0
        battleResultData.value.opponent_health = roundData?.opponent_health || 0
        
        // 追加新的日志（不覆盖旧日志）
        if (roundData?.logs && Array.isArray(roundData.logs)) {
          battleResultData.value.logs = [
            ...(battleResultData.value.logs || []),
            ...roundData.logs
          ]
          
          // 在下一个 tick 后自动滚动到顶部
          nextTick(() => {
            if (battleResultModalRef.value?.battleLogsContainer) {
              // 由于使用了 reverse，需要滚动到顶部
              battleResultModalRef.value.battleLogsContainer.scrollTop = 0
            }
          })
        }
        
        // 检查战斗是否结束
        const isBattleEnded = roundData?.battle_ended !== undefined ? roundData.battle_ended : roundData?.battleEnded
        
        if (isBattleEnded) {
          battleEnded = true
          victory = roundData?.victory !== undefined ? roundData.victory : false
          
          // 更新最终结果
          battleResultData.value.battle_ended = true
          battleResultData.value.victory = victory
          battleResultData.value.rewards = roundData?.rewards || []
          
          console.log('[DuelPVP] 战斗已结束，胜利:', victory)

          // 发出战斗结束事件
          emit('battle-end', { victory, roundData })

          // 记录战斗结果
          if (currentBattleOpponent.value) {
            await recordBattleResult(token, opponentId, victory)
          }
        }
      }
    }
  } catch (error) {
    console.error('[DuelPVP] 执行战斗回合时出错:', error)
  }
}

/**
 * 记录战斗结果
 */
const recordBattleResult = async (token, opponentId, victory) => {
  try {
    const result = victory ? '胜利' : '失败'
    await APIService.recordBattleResult(token, {
      opponentId,
      opponentName: currentBattleOpponent.value?.name || '未知对手',
      result,
      battleType: 'pvp',
      rewards: victory ? '灵石50' : ''
    })
    console.log('[DuelPVP] 战斗结果已记录')
  } catch (error) {
    console.error('[DuelPVP] 记录战斗结果失败:', error)
  }
}

/**
 * 关闭战斗结果弹窗
 */
const closeBattleResultModal = () => {
  showBattleResult.value = false
  battleResultData.value = null
}

/**
 * 继续战斗
 */
const handleContinueBattle = () => {
  console.log('[DuelPVP] 继续战斗')
  // 可以在这里添加继续战斗的逻辑
}

const refreshOpponents = async () => {
  const token = getAuthToken()
  console.log('[DuelPVP] refreshOpponents: token检查', { 
    hasToken: !!token, 
    tokenLength: token ? token.length : 0,
    tokenPreview: token ? token.substring(0, 20) + '...' : 'null'
  })
  
  if (!token) {
    console.warn('[DuelPVP] 未获取到认证令牌，无法加载道友列表')
    return
  }
  
  isLoadingOpponents.value = true
  try {
    console.log('[DuelPVP] 开始获取道友列表，发送的token:', {
      length: token.length,
      preview: token.substring(0, 20) + '...'
    })
    
    const response = await APIService.getDuelOpponents(token)
    console.log('[DuelPVP] API响应:', { success: response.success, hasData: !!response.data })
    
    if (response.success) {
      opponents.value = response.data.opponents
      console.log('[DuelPVP] 成功加载道友列表，共', opponents.value.length, '位道友')
    } else {
      console.error('[DuelPVP] 获取道友列表失败:', response.message)
    }
  } catch (error) {
    console.error('获取道友列表失败:', error)
  } finally {
    isLoadingOpponents.value = false
  }
}

// 初始化加载 - 从 store 获取用户ID
if (typeof window !== 'undefined') {
  const token = getAuthToken()
  console.log('[DuelPVP] setup阶段 - token检查', {
    hasToken: !!token,
    playerInfoStoreID: playerInfoStore.id,
    timestamp: new Date().toISOString()
  })
  if (token) {
    console.log('[DuelPVP] 从 playerInfoStore 获取玩家ID:', playerInfoStore.id)
    refreshOpponents()
  } else {
    console.log('[DuelPVP] setup时暂无令牌，等待用户登录后调用刷新')
  }
}
</script>

<style scoped>
.pvp-section {
  padding: 8px;
}

/* 战斗日志项向下滑入动画 */
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
  background-color: rgba(0, 0, 0, 0.05);
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