<template>
  <div class="demon-slaying-section">
    <!-- 除魔卫道说明 -->
    <n-alert title="除魔卫道" type="warning" style="margin-bottom: 16px;">
      <n-space vertical size="small">
        <div>除魔卫道，降服邪道弟子！战胜后可获得灵石、修为以及随机丹方残页（聚灵丹、聚气丹、回灵丹、雷灵丹、凝元丹、渡劫丹）。</div>
        <n-space>
          <n-tag type="warning">已挑战：{{ demonCount }}/20</n-tag>
          <n-tag type="info">灵力消耗：{{ spiritCost }}</n-tag>
          <n-tag type="success">当前灵力：{{ currentSpirit }}</n-tag>
        </n-space>
        <div style="font-size: 12px; color: #999;">每日00:00重置挑战次数</div>
      </n-space>
    </n-alert>
    
    <!-- 邪修难度选择 -->
    <n-card title="选择挑战难度" size="small">
      <n-space vertical>
        <!-- 难度选择单选组 -->
        <n-radio-group v-model:value="selectedDifficulty" name="difficulty">
          <n-space>
            <n-radio v-for="difficulty in difficulties" :key="difficulty.value" :value="difficulty.value">
              {{ difficulty.label }}
            </n-radio>
          </n-space>
        </n-radio-group>
        
        <!-- 邪修列表 -->
        <n-spin :show="isLoadingDemons">
          <n-list bordered>
            <n-list-item v-for="demon in demons" :key="demon.id">
              <n-thing>
                <template #header>
                  <n-space align="center">
                    <span>{{ demon.name }}</span>
                    <!-- 难度标签 -->
                    <n-tag :type="getDifficultyTagType(demon.difficulty)">
                      {{ getDifficultyName(demon.difficulty) }}
                    </n-tag>
                  </n-space>
                </template>
                <template #description>
                  <!-- 邪修属性描述 -->
                  <n-descriptions label-placement="left" :column="2" size="small">
                    <n-descriptions-item label="血量">{{ demon.baseAttributes?.health || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="攻击">{{ demon.baseAttributes?.attack || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="防御">{{ demon.baseAttributes?.defense || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="速度">{{ demon.baseAttributes?.speed || 'N/A' }}</n-descriptions-item>
                  </n-descriptions>
                </template>
                <template #footer>
                  <n-space justify="end">
                    <!-- 挑战邪修按钮 -->
                    <n-button 
                      type="primary" 
                      size="small" 
                      :loading="isChallengingDemon === demon.id"
                      @click="handleChallengeDemon(demon)"
                    >
                      除魔
                    </n-button>
                    <!-- 自动除魔按钮 -->
                    <n-button 
                      :type="isAutoFighting === demon.id ? 'warning' : 'success'" 
                      size="small"
                      @click="toggleAutoFight(demon)"
                    >
                      {{ isAutoFighting === demon.id ? '停止自动除魔' : '开始自动除魔' }}
                    </n-button>
                    <!-- 查看邪修详细信息按钮 -->
                    <n-button size="small" @click="handleViewDemonInfo(demon)">
                      详细信息
                    </n-button>
                  </n-space>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-spin>
        
        <!-- 分页信息和按钮 -->
        <n-space justify="between" align="center" style="margin-top: 16px;">
          <span>共 {{ totalDemons }} 个邪修（第 {{ currentPage }}/{{ totalPages }} 页）</span>
          <n-space>
            <n-button 
              :disabled="currentPage <= 1 || isLoadingDemons" 
              @click="() => { currentPage = Math.max(1, currentPage - 1); loadDemons(); }"
            >
              上一页
            </n-button>
            <n-button 
              :disabled="currentPage >= totalPages || isLoadingDemons" 
              @click="() => { currentPage = Math.min(totalPages, currentPage + 1); loadDemons(); }"
            >
              下一页
            </n-button>
          </n-space>
        </n-space>
      </n-space>
    </n-card>

    <!-- 邪修详细信息弹窗 -->
    <MonsterInfoModal 
      :show="showDemonInfoModal" 
      :monster="selectedDemon"
      @update:show="showDemonInfoModal = $event"
    />

    <!-- 战斗结果弹窗 -->
    <BattleResultModal 
      :show="showBattleResultModal" 
      :battle-result-data="battleResultData"
      @update:show="showBattleResultModal = $event"
      @close="handleCloseBattleResultModal"
    />

    <!-- 自动除魔日志面板 -->
    <n-card style="margin-top: 16px;" v-if="showAutoFightLog">
      <LogPanel ref="autoFightLogRef" title="自动除魔日志" />
    </n-card>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import LogPanel from '../../components/LogPanel.vue'
import { 
  NCard, NAlert, NSpace, NButton, NList, NListItem, NThing, NTag, 
  NDescriptions, NDescriptionsItem, NRadioGroup, NRadio, NSpin,
  useMessage
} from 'naive-ui'
import APIService from '../../services/api'
import { getAuthToken } from '../../stores/db'
import { usePlayerInfoStore } from '../../stores/playerInfo'
import { getDifficultyTagType, getDifficultyName } from '../utils/duelHelper'
import MonsterInfoModal from './MonsterInfoModal.vue'
import BattleResultModal from './BattleResultModal.vue'

const message = useMessage()
const playerInfoStore = usePlayerInfoStore()

// 状态管理
const selectedDifficulty = ref('normal')
const demons = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const totalDemons = ref(0)
const totalPages = ref(0)
const isLoadingDemons = ref(false)
const isChallengingDemon = ref(null) // 正在挑战的邪俪ID

// 灵力状态
const spiritCost = ref(0)
const currentSpirit = ref(0)
const demonCount = ref(0) // 已挑战次数

// 邪修信息弹窗
const showDemonInfoModal = ref(false)
const selectedDemon = ref(null)

// 战斗结果弹窗
const showBattleResultModal = ref(false)
const battleResultData = ref(null)
const currentBattleDemon = ref(null) // 当前战斗的邪修
const isBattleInProgress = ref(false) // 战斗是否进行中
// 是否正在自动除魔（逻辑状态）
const isAutoFighting = ref(null) // demon.id | null
const autoFightDemonId = ref(null) // 自动除魔锁定的 demon.id
// 是否显示日志面板（UI 状态）
const showAutoFightLog = ref(true)
// 日志组件引用
const autoFightLogRef = ref(null)

// 难度选项
const difficulties = [
  { label: '普通', value: 'normal' },
  { label: '困难', value: 'hard' },
  { label: '噩梦', value: 'boss' }
]

// 开始下一场自动战斗
const startNextAutoBattle = async () => {
  const token = getAuthToken()
  if (!token || !autoFightDemonId.value) return false

  const demon = demons.value.find(
    m => m.id === autoFightDemonId.value
  )
  if (!demon) {
    autoFightLogRef.value?.addLog('❌ 未找到邪修，自动除魔终止')
    return false
  }

  autoFightLogRef.value?.addLog('🔄 开始下一场自动除魔')

  const playerBattleDataRes = await APIService.getPlayerBattleData(
    playerInfoStore.id,
    token
  )
  if (!playerBattleDataRes.success) {
    autoFightLogRef.value?.addLog('❌ 获取玩家数据失败')
    return false
  }

  const startBattleRes = await APIService.startPvEBattle(
    demon.id,
    playerBattleDataRes.data,
    demon,
    token
  )
  if (!startBattleRes.success) {
    autoFightLogRef.value?.addLog('❌ 开始新战斗失败')
    return false
  }

  currentBattleDemon.value = demon
  autoFightLogRef.value?.addLog(
    `⚔️ 新战斗开始（回合 ${startBattleRes.data.round || 1}）`
  )

  return true
}

const autoFightLoop = async () => {
  while (isAutoFighting.value === autoFightDemonId.value) {
    const token = getAuthToken()
    if (!token) {
      autoFightLogRef.value?.addLog('❌ 登录失效，自动除魔停止')
      break
    }

    try {
      const res = await APIService.executePvERound(
        autoFightDemonId.value,
        token
      )

      if (!res.success) {
        autoFightLogRef.value?.addLog('❌ 战斗异常，自动除魔停止')
        break
      }

      const data = res.data

      // 打印每回合日志
      if (Array.isArray(data.logs)) {
        data.logs.forEach(log => {
          autoFightLogRef.value?.addLog(log)
        })
      }

      if (data.battle_ended) {
        if (data.victory) {
          autoFightLogRef.value?.addLog('🎉 战斗胜利')

          // 奖励日志
          if (Array.isArray(data.rewards) && data.rewards.length > 0) {
            autoFightLogRef.value?.addLog('🎁 获得奖励：')
            data.rewards.forEach(reward => {
              if (reward.type === 'spirit_stone') {
                autoFightLogRef.value?.addLog(`- 灵石 +${reward.amount}`)
              } else if (reward.type === 'cultivation') {
                autoFightLogRef.value?.addLog(`- 修为 +${reward.amount}`)
              } else if (reward.type === 'pill_fragment') {
                autoFightLogRef.value?.addLog(`- ${reward.name}残页 +${reward.count}`)
              } else if (reward.type === 'herb') {
                autoFightLogRef.value?.addLog(`- ${reward.name} +${reward.count}`)
              }
            })
          }

          await APIService.endPvEBattle(
            autoFightDemonId.value,
            token
          )

          await new Promise(r => setTimeout(r, 800))

          const started = await startNextAutoBattle()
          if (!started) {
            isAutoFighting.value = null
            break
          }

          continue
        } else {
          autoFightLogRef.value?.addLog('❌ 战斗失败，自动除魔停止')
          break
        }
      }

      await new Promise(r => setTimeout(r, 1000))
    } catch (e) {
      autoFightLogRef.value?.addLog('❌ 自动除魔异常')
      break
    }
  }

  // 统一收尾
  isAutoFighting.value = null
  autoFightDemonId.value = null
  currentBattleDemon.value = null
  autoFightLogRef.value?.addLog('自动除魔结束')
}

/**
 * 加载邪俪列表
 */
const loadDemons = async () => {
  try {
    isLoadingDemons.value = true
    const token = getAuthToken()
    
    if (!token) {
      message.error('请先登录')
      return
    }
    
    const response = await APIService.getDemonSlayingChallenges(
      token,
      currentPage.value,
      pageSize.value,
      selectedDifficulty.value === 'all' ? '' : selectedDifficulty.value
    )
    
    if (response.success) {
      demons.value = response.data.monsters
      currentPage.value = response.data.page
      pageSize.value = response.data.pageSize
      totalDemons.value = response.data.total
      totalPages.value = response.data.totalPages
    } else {
      message.error(response.message || '加载除魔卫道列表失败')
    }

    // 加载灵力状态
    await loadSpiritStatus()
  } catch (error) {
    console.error('[DuelDemonSlaying] 加载除魔卫道列表异常:', error)
    message.error('加载除魔卫道列表失败')
  } finally {
    isLoadingDemons.value = false
  }
}

// 加载灵力状态
const loadSpiritStatus = async () => {
  const token = getAuthToken()
  if (!token) return

  try {
    const response = await APIService.getDuelStatus(token)
    if (response.success && response.data) {
      spiritCost.value = response.data.pveCost || 0
      currentSpirit.value = Math.floor(playerInfoStore.spirit || 0)
      demonCount.value = response.data.demonCount || 0
    }
  } catch (error) {
    console.error('[DuelDemonSlaying] 获取灵力状态失败:', error)
  }
}

/**
 * 查看邪修详细信息
 */
const handleViewDemonInfo = async (demon) => {
  try {
    const token = getAuthToken()
    if (!token) {
      message.error('请先登录')
      return
    }

    // 获取邪修详细信息
    const response = await APIService.getMonsterInfo(demon.id, token)
    if (response.success) {
      selectedDemon.value = response.data
      showDemonInfoModal.value = true
    } else {
      message.error(response.message || '获取邪修信息失败')
    }
  } catch (error) {
    console.error('[DuelDemonSlaying] 获取邪修信息异常:', error)
    message.error('获取邪修信息失败')
  }
}

/**
 * 挑战邪修
 */
const handleChallengeDemon = async (demon) => {
  try {
    isChallengingDemon.value = demon.id
    const token = getAuthToken()
    
    if (!token) {
      message.error('请先登录')
      return
    }

    // 获取玩家战斗数据
    const playerBattleDataRes = await APIService.getPlayerBattleData(playerInfoStore.id, token)
    if (!playerBattleDataRes.success) {
      message.error('获取玩家战斗数据失败')
      return
    }

    // 开始战斗
    const startBattleRes = await APIService.startPvEBattle(
      demon.id,
      playerBattleDataRes.data,
      demon,
      token
    )

    if (!startBattleRes.success) {
      message.error(startBattleRes.message || '开始战斗失败')
      return
    }

    // 初始化战斗数据
    currentBattleDemon.value = demon
    isBattleInProgress.value = true
    battleResultData.value = startBattleRes.data
    showBattleResultModal.value = true

    // 如果初始化后还没结束，继续执行回合
    if (!startBattleRes.data.battle_ended) {
      await executeBattleRound(token)
    }
  } catch (error) {
    console.error('[DuelDemonSlaying] 挑战邪修异常:', error)
    message.error('挑战邪修失败')
  } finally {
    isChallengingDemon.value = null
  }
}

/**
 * 执行战斗回合
 */
const executeBattleRound = async (token) => {
  if (!isBattleInProgress.value || !currentBattleDemon.value) return

  try {
    const response = await APIService.executePvERound(
      currentBattleDemon.value.id,
      token
    )

    if (response.success) {
      battleResultData.value = response.data

      // 如果战斗未结束，继续执行下一回合
      if (!response.data.battle_ended) {
        // 1秒后自动执行下一回合
        setTimeout(() => {
          executeBattleRound(token)
        }, 1000)
      } else {
        isBattleInProgress.value = false
      }
    } else {
      message.error(response.message || '执行战斗回合失败')
      isBattleInProgress.value = false
    }
  } catch (error) {
    console.error('[DuelDemonSlaying] 执行战斗回合异常:', error)
    message.error('执行战斗回合失败')
    isBattleInProgress.value = false
  }
}

/**
 * 处理战斗结果弹窗关闭
 */
const handleCloseBattleResultModal = async () => {
  if (isBattleInProgress.value) {
    isBattleInProgress.value = false
  }

  // 结束战斗
  if (currentBattleDemon.value) {
    const token = getAuthToken()
    if (token) {
      await APIService.endPvEBattle(currentBattleDemon.value.id, token)
    }
    currentBattleDemon.value = null
  }

  showBattleResultModal.value = false
  battleResultData.value = null
}

/**
 * 切换自动除魔
 */
const toggleAutoFight = async (demon) => {
  // 停止
  if (isAutoFighting.value === demon.id) {
    isAutoFighting.value = null
    autoFightLogRef.value?.addLog('🛑 玩家手动停止自动除魔')
    return
  }

  // 开始
  const token = getAuthToken()
  if (!token) {
    message.error('请先登录')
    return
  }

  isAutoFighting.value = demon.id
  autoFightDemonId.value = demon.id
  currentBattleDemon.value = demon
  showAutoFightLog.value = true

  await nextTick()

  autoFightLogRef.value?.addLog(`开始自动除魔 ${demon.name}`)

  const playerBattleDataRes = await APIService.getPlayerBattleData(
    playerInfoStore.id,
    token
  )
  if (!playerBattleDataRes.success) {
    message.error('获取玩家战斗数据失败')
    isAutoFighting.value = null
    return
  }

  const startBattleRes = await APIService.startPvEBattle(
    demon.id,
    playerBattleDataRes.data,
    demon,
    token
  )
  if (!startBattleRes.success) {
    message.error('开始战斗失败')
    isAutoFighting.value = null
    return
  }

  autoFightLogRef.value?.addLog(
    `初始化战斗数据，回合 ${startBattleRes.data.round || 1}`
  )

  await autoFightLoop()
}

// 初始化加载
onMounted(() => {
  loadDemons()
})

// 监听难度变化
watch(selectedDifficulty, () => {
  currentPage.value = 1 // 不同难度时重置到第一页
  loadDemons()
})
</script>

<style scoped>
.demon-slaying-section {
  padding: 8px;
}
</style>
