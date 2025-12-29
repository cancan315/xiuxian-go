import { ref } from 'vue'
import { usePlayerInfoStore } from '../../stores/playerInfo'
import APIService from '../../services/api'
import { getAuthToken } from '../../stores/db'
import { simulateBattle } from '../../plugins/battleSimulator'

export function useDuel() {
  // 响应式状态
  const showBattleModal = ref(false)
  const battleLogs = ref([])
  const battleResult = ref(null)
  const battleTitle = ref('')
  const selectedPlayer = ref(null)
  const showPlayerInfoModal = ref(false)

  const playerInfoStore = usePlayerInfoStore()

  /**
   * 挑战玩家
   */
  const challengePlayer = async (opponent) => {
    try {
      const token = getAuthToken()
      if (!token) {
        window.$message.error('请先登录')
        return
      }

      // 获取玩家战斗数据
      const playerBattleData = await APIService.getPlayerBattleData(playerInfoStore.userId, token)
      const opponentBattleData = await APIService.getPlayerBattleData(opponent.id, token)

      if (!playerBattleData || !opponentBattleData) {
        window.$message.error('获取战斗数据失败')
        return
      }

      // 执行战斗模拟
      battleTitle.value = `PVP对战 - ${opponent.name}`
      const result = simulateBattle(playerBattleData, opponentBattleData)

      battleLogs.value = result.logs || []
      battleResult.value = result.result
      showBattleModal.value = true

      // 记录战斗结果
      const battleRecord = {
        playerID: playerInfoStore.userId,
        opponentID: opponent.id,
        opponentName: opponent.name,
        result: result.result === '胜利' ? '胜利' : '失败',
        battleType: 'pvp',
        rewards: result.rewards || ''
      }

      await APIService.recordBattleResult(token, battleRecord)
    } catch (error) {
      console.error('挑战玩家失败:', error)
      window.$message.error('挑战玩家失败')
    }
  }

  /**
   * 挑战妖兽
   */
  const challengeMonster = async (monster) => {
    try {
      const token = getAuthToken()
      if (!token) {
        window.$message.error('请先登录')
        return
      }

      // 获取玩家战斗数据
      const playerBattleData = await APIService.getPlayerBattleData(playerInfoStore.userId, token)

      if (!playerBattleData) {
        window.$message.error('获取玩家战斗数据失败')
        return
      }

      // 执行战斗模拟
      battleTitle.value = `妖兽挑战 - ${monster.name}`
      const result = simulateBattle(playerBattleData, monster)

      battleLogs.value = result.logs || []
      battleResult.value = result.result
      showBattleModal.value = true

      // 记录战斗结果
      const battleRecord = {
        playerID: playerInfoStore.userId,
        opponentID: monster.id,
        opponentName: monster.name,
        result: result.result === '胜利' ? '胜利' : '失败',
        battleType: 'pve',
        rewards: result.rewards || ''
      }

      await APIService.recordBattleResult(token, battleRecord)
    } catch (error) {
      console.error('挑战妖兽失败:', error)
      window.$message.error('挑战妖兽失败')
    }
  }

  /**
   * 查看玩家信息
   */
  const viewPlayerInfo = (player) => {
    selectedPlayer.value = player
    showPlayerInfoModal.value = true
  }

  /**
   * 关闭战斗弹窗
   */
  const closeBattleModal = () => {
    showBattleModal.value = false
    battleLogs.value = []
    battleResult.value = null
    battleTitle.value = ''
  }

  /**
   * 领取战斗奖励
   */
  const claimRewards = async () => {
    try {
      const token = getAuthToken()
      if (!token) {
        window.$message.error('请先登录')
        return
      }

      if (!battleResult.value || battleResult.value !== '胜利') {
        window.$message.warning('只有胜利的战斗才能领取奖励')
        return
      }

      // 调用领取奖励API
      const rewards = battleResult.value === '胜利' ? ['修为', '灵石'] : []
      await APIService.claimBattleRewards(token, rewards)

      window.$message.success('奖励领取成功')
      closeBattleModal()
    } catch (error) {
      console.error('领取奖励失败:', error)
      window.$message.error('领取奖励失败')
    }
  }

  return {
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
  }
}