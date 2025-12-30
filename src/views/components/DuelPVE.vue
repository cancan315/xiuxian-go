<template>
  <div class="pve-section">
    <!-- 妖兽挑战说明 -->
    <n-alert title="妖兽挑战" type="info" style="margin-bottom: 16px;">
      消耗灵力，降服不同等级的妖兽，有概率获得灵草，灵草用于炼制丹药。
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
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
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
const monsters = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const totalMonsters = ref(0)
const totalPages = ref(0)
const isLoadingMonsters = ref(false)
const isChallengingMonster = ref(null) // 正在挑战的妖兽ID

// 妖兽信息弹窗
const showMonsterInfoModal = ref(false)
const selectedMonster = ref(null)

// 战斗结果弹窗
const showBattleResultModal = ref(false)
const battleResultData = ref(null)
const currentBattleMonster = ref(null) // 当前战斗的妖兽
const isBattleInProgress = ref(false) // 战斗是否进行中

// 难度选项
const difficulties = [
  { label: '普通', value: 'normal' },
  { label: '困难', value: 'hard' },
  { label: '噩梦', value: 'boss' }
]

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
  } catch (error) {
    console.error('[DuelPVE] 加载妖兽列表异常:', error)
    message.error('加载妖兽列表失败')
  } finally {
    isLoadingMonsters.value = false
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
