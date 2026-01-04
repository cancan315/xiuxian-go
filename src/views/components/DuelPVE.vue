<template>
  <div class="pve-section">
    <!-- 妖兽挑战说明 -->
    <n-alert title="妖兽挑战" type="info" style="margin-bottom: 16px;">
      <n-space vertical size="small">
        <div>消耗灵力，降服不同等级的妖兽，有概率获得灵草，灵草用于炼制丹药。</div>
        <n-space>
          <n-tag type="warning">已挑战：{{ pveCount }}/100</n-tag>
          <n-tag type="info">灵力消耗：{{ spiritCost }}</n-tag>
          <n-tag type="success">当前灵力：{{ currentSpirit }}</n-tag>
        </n-space>
        <div style="font-size: 12px; color: #999;">每日00:00重置挑战次数</div>
      </n-space>
    </n-alert>
    
    <!-- 妖兽难度选择 -->
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
        
        <!-- 妖兽列表 -->
        <n-spin :show="isLoadingMonsters">
          <n-list bordered>
            <n-list-item v-for="monster in monsters" :key="monster.id">
              <n-thing>
                <template #header>
                  <n-space align="center">
                    <span>{{ monster.name }}</span>
                    <!-- 难度标签 -->
                    <n-tag :type="getDifficultyTagType(monster.difficulty)">
                      {{ getDifficultyName(monster.difficulty) }}
                    </n-tag>
                  </n-space>
                </template>
                <template #description>
                  <!-- 妖兽属性描述 -->
                  <n-descriptions label-placement="left" :column="2" size="small">
                    <n-descriptions-item label="血量">{{ monster.baseAttributes?.health || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="攻击">{{ monster.baseAttributes?.attack || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="防御">{{ monster.baseAttributes?.defense || 'N/A' }}</n-descriptions-item>
                    <n-descriptions-item label="速度">{{ monster.baseAttributes?.speed || 'N/A' }}</n-descriptions-item>
                  </n-descriptions>
                </template>
                <template #footer>
                  <n-space justify="end">
                    <!-- 挑战妖兽按钮 -->
                    <n-button 
                      type="primary" 
                      size="small" 
                      :loading="isChallengingMonster === monster.id"
                      @click="handleChallengeMonster(monster)"
                    >
                      降服
                    </n-button>
                    <!-- 自动降伏按钮 -->
                    <n-button 
                      :type="isAutoFighting === monster.id ? 'warning' : 'success'" 
                      size="small"
                      @click="toggleAutoFight(monster)"
                    >
                      {{ isAutoFighting === monster.id ? '停止自动降伏' : '开始自动降伏' }}
                    </n-button>
                    <!-- 查看妖兽详细信息按钮 -->
                    <n-button size="small" @click="handleViewMonsterInfo(monster)">
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
          <span>共 {{ totalMonsters }} 只妖兽（第 {{ currentPage }}/{{ totalPages }} 页）</span>
          <n-space>
            <n-button 
              :disabled="currentPage <= 1 || isLoadingMonsters" 
              @click="() => { currentPage = Math.max(1, currentPage - 1); loadMonsters(); }"
            >
              上一页
            </n-button>
            <n-button 
              :disabled="currentPage >= totalPages || isLoadingMonsters" 
              @click="() => { currentPage = Math.min(totalPages, currentPage + 1); loadMonsters(); }"
            >
              下一页
            </n-button>
          </n-space>
        </n-space>
      </n-space>
    </n-card>

    <!-- 妖兽详细信息弹窗 -->
    <MonsterInfoModal 
      :show="showMonsterInfoModal" 
      :monster="selectedMonster"
      @update:show="showMonsterInfoModal = $event"
    />

    <!-- 战斗结果弹窗 -->
    <BattleResultModal 
      :show="showBattleResultModal" 
      :battle-result-data="battleResultData"
      @update:show="showBattleResultModal = $event"
      @close="handleCloseBattleResultModal"
    />

    <!-- 自动降伏日志面板 -->
    <n-card style="margin-top: 16px;" v-if="showAutoFightLog">
  <LogPanel ref="autoFightLogRef" title="自动降伏妖兽日志" />
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
const selectedDifficulty = ref('lianqi')
const monsters = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const totalMonsters = ref(0)
const totalPages = ref(0)
const isLoadingMonsters = ref(false)
const isChallengingMonster = ref(null) // 正在挑战的妖兽ID

// 灵力状态
const spiritCost = ref(0)
const currentSpirit = ref(0)
const pveCount = ref(0) // 已挑战次数

// 妖兽信息弹窗
const showMonsterInfoModal = ref(false)
const selectedMonster = ref(null)

// 战斗结果弹窗
const showBattleResultModal = ref(false)
const battleResultData = ref(null)
const currentBattleMonster = ref(null) // 当前战斗的妖兽
const isBattleInProgress = ref(false) // 战斗是否进行中
// 是否正在自动降伏（逻辑状态）
const isAutoFighting = ref(null) // monster.id | null
const autoFightMonsterId = ref(null) // 自动降伏锁定的 monster.id
// 是否显示日志面板（UI 状态）
const showAutoFightLog = ref(true)
// 日志组件引用
const autoFightLogRef = ref(null)

// 难度选项
const difficulties = [
  { label: '练气', value: 'lianqi' },
  { label: '筑基', value: 'zhuji' },
  { label: '金丹', value: 'jindan' }
]
 // 开始下一场自动战斗
const startNextAutoBattle = async () => {
  const token = getAuthToken()
  if (!token || !autoFightMonsterId.value) return false

  const monster = monsters.value.find(
    m => m.id === autoFightMonsterId.value
  )
  if (!monster) {
    autoFightLogRef.value?.addLog('❌ 未找到妖兽，自动降伏终止')
    return false
  }

  autoFightLogRef.value?.addLog('🔄 开始下一场自动降伏')

  const playerBattleDataRes = await APIService.getPlayerBattleData(
    playerInfoStore.id,
    token
  )
  if (!playerBattleDataRes.success) {
    autoFightLogRef.value?.addLog('❌ 获取玩家数据失败')
    return false
  }

  const startBattleRes = await APIService.startPvEBattle(
    monster.id,
    playerBattleDataRes.data,
    monster,
    token
  )
  if (!startBattleRes.success) {
    autoFightLogRef.value?.addLog('❌ 开始新战斗失败')
    return false
  }

  currentBattleMonster.value = monster
  autoFightLogRef.value?.addLog(
    `⚔️ 新战斗开始（回合 ${startBattleRes.data.round || 1}）`
  )

  return true
}

const autoFightLoop = async () => {
  while (isAutoFighting.value === autoFightMonsterId.value) {
    const token = getAuthToken()
    if (!token) {
      autoFightLogRef.value?.addLog('❌ 登录失效，自动降伏停止')
      break
    }

    try {
      const res = await APIService.executePvERound(
        autoFightMonsterId.value,
        token
      )

      if (!res.success) {
        autoFightLogRef.value?.addLog('❌ 战斗异常，自动降伏停止')
        break
      }

      const data = res.data

      // 👉 这里你可以继续补充详细回合日志
      // ✅ 打印每回合日志（关键）
      if (Array.isArray(data.logs)) {
        data.logs.forEach(log => {
          autoFightLogRef.value?.addLog(log)
        })
      }

      if (data.battle_ended) {
        if (data.victory) {
          autoFightLogRef.value?.addLog('🎉 战斗胜利')

          // ✅ 奖励日志（关键）
          if (Array.isArray(data.rewards) && data.rewards.length > 0) {
            autoFightLogRef.value?.addLog('🎁 获得奖励：')
            data.rewards.forEach(reward => {
              autoFightLogRef.value?.addLog(
                `- ${reward.name} ×${reward.count}`
              )
            })
          }

          await APIService.endPvEBattle(
            autoFightMonsterId.value,
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
          autoFightLogRef.value?.addLog('❌ 战斗失败，自动降伏停止')
          break
        }
      }

      await new Promise(r => setTimeout(r, 1000))
    } catch (e) {
      autoFightLogRef.value?.addLog('❌ 自动降伏异常')
      break
    }
  }

  // ✅ 统一收尾
  isAutoFighting.value = null
  autoFightMonsterId.value = null
  currentBattleMonster.value = null
  autoFightLogRef.value?.addLog('自动降伏结束')
}

/**
 * 加载妖兽列表
 */
const loadMonsters = async () => {
  try {
    isLoadingMonsters.value = true
    const token = getAuthToken()
    
    if (!token) {
      message.error('请先登录')
      return
    }
    
    const response = await APIService.getMonsters(
      token,
      currentPage.value,
      pageSize.value,
      selectedDifficulty.value === 'all' ? '' : selectedDifficulty.value
    )
    
    if (response.success) {
      monsters.value = response.data.monsters
      currentPage.value = response.data.page
      pageSize.value = response.data.pageSize
      totalMonsters.value = response.data.total
      totalPages.value = response.data.totalPages
    } else {
      message.error(response.message || '加载妖兽列表失败')
    }

    // 加载灵力状态
    await loadSpiritStatus()
  } catch (error) {
    console.error('[DuelPVE] 加载妖兽列表异常:', error)
    message.error('加载妖兽列表失败')
  } finally {
    isLoadingMonsters.value = false
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
      pveCount.value = response.data.pveCount || 0
    }
  } catch (error) {
    console.error('[DuelPVE] 获取灵力状态失败:', error)
  }
}

/**
 * 查看妖兽详细信息
 */
const handleViewMonsterInfo = async (monster) => {
  try {
    const token = getAuthToken()
    if (!token) {
      message.error('请先登录')
      return
    }

    // 获取妖兽详细信息
    const response = await APIService.getMonsterInfo(monster.id, token)
    if (response.success) {
      selectedMonster.value = response.data
      showMonsterInfoModal.value = true
    } else {
      message.error(response.message || '获取妖兽信息失败')
    }
  } catch (error) {
    console.error('[DuelPVE] 获取妖兽信息异常:', error)
    message.error('获取妖兽信息失败')
  }
}

/**
 * 挑战妖兽
 */
const handleChallengeMonster = async (monster) => {
  try {
    isChallengingMonster.value = monster.id
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
      monster.id,
      playerBattleDataRes.data,
      monster,
      token
    )

    if (!startBattleRes.success) {
      message.error(startBattleRes.message || '开始战斗失败')
      return
    }

    // 初始化战斗数据
    currentBattleMonster.value = monster
    isBattleInProgress.value = true
    battleResultData.value = startBattleRes.data
    showBattleResultModal.value = true

    // 如果初始化后还没结束，继续执行回合
    if (!startBattleRes.data.battle_ended) {
      await executeBattleRound(token)
    }
  } catch (error) {
    console.error('[DuelPVE] 挑战妖兽异常:', error)
    message.error('挑战妖兽失败')
  } finally {
    isChallengingMonster.value = null
  }
}

/**
 * 执行战斗回合
 */
const executeBattleRound = async (token) => {
  if (!isBattleInProgress.value || !currentBattleMonster.value) return

  try {
    const response = await APIService.executePvERound(
      currentBattleMonster.value.id,
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
    console.error('[DuelPVE] 执行战斗回合异常:', error)
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
  if (currentBattleMonster.value) {
    const token = getAuthToken()
    if (token) {
      await APIService.endPvEBattle(currentBattleMonster.value.id, token)
    }
    currentBattleMonster.value = null
  }

  showBattleResultModal.value = false
  battleResultData.value = null
}

/**
 * 切换自动降伏
 */
const toggleAutoFight = async (monster) => {
  // 🛑 停止
  if (isAutoFighting.value === monster.id) {
    isAutoFighting.value = null
    autoFightLogRef.value?.addLog('🛑 玩家手动停止自动降伏')
    return
  }

  // ▶ 开始
  const token = getAuthToken()
  if (!token) {
    message.error('请先登录')
    return
  }

  isAutoFighting.value = monster.id
  autoFightMonsterId.value = monster.id
  currentBattleMonster.value = monster
  showAutoFightLog.value = true

  await nextTick()

  autoFightLogRef.value?.addLog(`开始自动降伏 ${monster.name}`)

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
    monster.id,
    playerBattleDataRes.data,
    monster,
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
  loadMonsters()
})

// 监听难度变化
watch(selectedDifficulty, () => {
  currentPage.value = 1 // 不同难度时重置到第一页
  loadMonsters()
})
</script>

<style scoped>
.pve-section {
  padding: 8px;
}
</style>
