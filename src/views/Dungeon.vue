<template>
  <div class="dungeon-container">
    <n-card title="秘境演法">
      <n-alert v-if="!isAutoExploring" type="info">
          说明：开启秘境需要消耗灵力、10灵石，获取灵石、强化石、洗炼石、灵兽精华。请道友前往秘境中追寻仙缘。
      </n-alert>
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="playerInfoStore.dungeonDifficulty"
            placeholder="请选择难度"
            :options="dungeonOptions"
            style="width: 120px"
            :disabled="dungeonState.inCombat || dungeonState.showingOptions"
            clearable
          />
          <n-button
            type="primary"
            @click="startDungeon"
            :disabled="dungeonState.inCombat || dungeonState.showingOptions"
          >
            开启秘境
          </n-button>
        </n-space>
      </template>
      <n-space vertical>
        <!-- 层数显示 -->
        <n-statistic label="当前层数" :value="dungeonState.floor" />
        
        <!-- 战斗日志面板 - 始终显示 -->
        <log-panel ref="logRef" :logs="combatLog" style="margin-top: 16px" />
        
        <!-- 选项界面 -->
        <n-card v-if="dungeonState.showingOptions" title="选择增益">
          <template #header-extra>
            <n-space>
              <n-button type="primary" @click="handleRefreshOptions" :disabled="refreshNumber === 0">
                刷新增益({{ refreshNumber }})
              </n-button>
            </n-space>
          </template>
          <div class="option-cards">
            <div
              v-for="option in dungeonState.currentOptions"
              :key="option.id"
              class="option-card"
              :style="{ borderColor: getOptionColor(option.type).color }"
              @click="selectOption(option)"
            >
              <div class="option-name">{{ option.name }}</div>
              <div class="option-description">{{ option.description }}</div>
              <div class="option-quality" :style="{ color: getOptionColor(option.type).color }">
                {{ getOptionColor(option.type).name }}
              </div>
            </div>
          </div>
        </n-card>
        <!-- 战斗界面 -->
        <template v-if="dungeonState.inCombat && dungeonState.combatManager">
          <n-card :bordered="false">
            <n-divider>
              {{ dungeonState.combatManager.round }} / {{ dungeonState.combatManager.maxRounds }}回合
            </n-divider>
            <!-- 添加战斗场景 -->
            <div class="combat-scene">
              <div class="character player" :class="{ attack: playerAttacking, hurt: playerHurt }">
                <div v-if="playerAttacking" class="attack-effect player-effect"></div>
                <n-button class="character-name" type="info" dashed @click="infoCliclk('player')">
                  {{ dungeonState.combatManager.player.name }}
                </n-button>
                <div class="character-avatar player-avatar">
                  {{ dungeonState.combatManager.player.name[0] }}
                </div>
                <div class="health-bar">
                  <div
                    class="health-fill"
                    :style="{
                      width: `${
                        (dungeonState.combatManager.player.currentHealth /
                          dungeonState.combatManager.player.stats.maxHealth) *
                        100
                      }%`
                    }"
                  ></div>
                </div>
              </div>
              <div class="character enemy" :class="{ attack: enemyAttacking, hurt: enemyHurt }">
                <div v-if="enemyAttacking" class="attack-effect enemy-effect"></div>
                <n-button class="character-name" type="error" dashed @click="infoCliclk('enemy')">
                  {{ dungeonState.combatManager.enemy.name }}
                </n-button>
                <div class="character-avatar enemy-avatar">
                  {{ dungeonState.combatManager.enemy.name[0] }}
                </div>
                <div class="health-bar">
                  <div
                    class="health-fill"
                    :style="{
                      width: `${
                        (dungeonState.combatManager.enemy.currentHealth /
                          dungeonState.combatManager.enemy.stats.maxHealth) *
                        100
                      }%`
                    }"
                  ></div>
                </div>
              </div>
            </div>
            <n-modal
              v-model:show="infoShow"
              preset="dialog"
              :title="`${
                infoType == 'player' ? dungeonState.combatManager.player.name : dungeonState.combatManager.enemy.name
              }的属性`"
            >
              <n-card :bordered="false">
                <!-- 玩家属性 -->
                <template v-if="infoType == 'player'">
                  <n-divider>基础属性</n-divider>
                  <n-descriptions bordered :column="2">
                    <n-descriptions-item label="生命值">
                      {{ dungeonState.combatManager.player.currentHealth.toFixed(1) }} /
                      {{ dungeonState.combatManager.player.stats.maxHealth.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="攻击力">
                      {{ dungeonState.combatManager.player.stats.damage.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="防御力">
                      {{ dungeonState.combatManager.player.stats.defense.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="速度">
                      {{ dungeonState.combatManager.player.stats.speed.toFixed(1) }}
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>战斗属性</n-divider>
                  <n-descriptions bordered :column="3">
                    <n-descriptions-item label="暴击率">
                      {{ (dungeonState.combatManager.player.stats.critRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="连击率">
                      {{ (dungeonState.combatManager.player.stats.comboRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="反击率">
                      {{ (dungeonState.combatManager.player.stats.counterRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="眩晕率">
                      {{ (dungeonState.combatManager.player.stats.stunRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="闪避率">
                      {{ (dungeonState.combatManager.player.stats.dodgeRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="吸血率">
                      {{ (dungeonState.combatManager.player.stats.vampireRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>战斗抗性</n-divider>
                  <n-descriptions bordered :column="3">
                    <n-descriptions-item label="抗暴击">
                      {{ (dungeonState.combatManager.player.stats.critResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗连击">
                      {{ (dungeonState.combatManager.player.stats.comboResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗反击">
                      {{ (dungeonState.combatManager.player.stats.counterResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗眩晕">
                      {{ (dungeonState.combatManager.player.stats.stunResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗闪避">
                      {{ (dungeonState.combatManager.player.stats.dodgeResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗吸血">
                      {{ (dungeonState.combatManager.player.stats.vampireResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>特殊属性</n-divider>
                  <n-descriptions bordered :column="4">
                    <n-descriptions-item label="强化治疗">
                      {{ (dungeonState.combatManager.player.stats.healBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="强化爆伤">
                      {{ (dungeonState.combatManager.player.stats.critDamageBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="弱化爆伤">
                      {{ (dungeonState.combatManager.player.stats.critDamageReduce * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="最终增伤">
                      {{ (dungeonState.combatManager.player.stats.finalDamageBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="最终减伤">
                      {{ (dungeonState.combatManager.player.stats.finalDamageReduce * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="战斗属性提升">
                      {{ (dungeonState.combatManager.player.stats.combatBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="战斗抗性提升">
                      {{ (dungeonState.combatManager.player.stats.resistanceBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                </template>
                <!-- 敌人属性 -->
                <template v-else>
                  <n-divider>基础属性</n-divider>
                  <n-descriptions bordered :column="2">
                    <n-descriptions-item label="生命值">
                      {{ dungeonState.combatManager.enemy.currentHealth.toFixed(1) }} /
                      {{ dungeonState.combatManager.enemy.stats.maxHealth.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="攻击力">
                      {{ dungeonState.combatManager.enemy.stats.damage.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="防御力">
                      {{ dungeonState.combatManager.enemy.stats.defense.toFixed(1) }}
                    </n-descriptions-item>
                    <n-descriptions-item label="速度">
                      {{ dungeonState.combatManager.enemy.stats.speed.toFixed(1) }}
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>战斗属性</n-divider>
                  <n-descriptions bordered :column="3">
                    <n-descriptions-item label="暴击率">
                      {{ (dungeonState.combatManager.enemy.stats.critRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="连击率">
                      {{ (dungeonState.combatManager.enemy.stats.comboRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="反击率">
                      {{ (dungeonState.combatManager.enemy.stats.counterRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="眩晕率">
                      {{ (dungeonState.combatManager.enemy.stats.stunRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="闪避率">
                      {{ (dungeonState.combatManager.enemy.stats.dodgeRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="吸血率">
                      {{ (dungeonState.combatManager.enemy.stats.vampireRate * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>战斗抗性</n-divider>
                  <n-descriptions bordered :column="3">
                    <n-descriptions-item label="抗暴击">
                      {{ (dungeonState.combatManager.enemy.stats.critResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗连击">
                      {{ (dungeonState.combatManager.enemy.stats.comboResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗反击">
                      {{ (dungeonState.combatManager.enemy.stats.counterResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗眩晕">
                      {{ (dungeonState.combatManager.enemy.stats.stunResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗闪避">
                      {{ (dungeonState.combatManager.enemy.stats.dodgeResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="抗吸血">
                      {{ (dungeonState.combatManager.enemy.stats.vampireResist * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                  <n-divider>特殊属性</n-divider>
                  <n-descriptions bordered :column="3">
                    <n-descriptions-item label="强化治疗">
                      {{ (dungeonState.combatManager.enemy.stats.healBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="强化爆伤">
                      {{ (dungeonState.combatManager.enemy.stats.critDamageBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="弱化爆伤">
                      {{ (dungeonState.combatManager.enemy.stats.critDamageReduce * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="最终增伤">
                      {{ (dungeonState.combatManager.enemy.stats.finalDamageBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="最终减伤">
                      {{ (dungeonState.combatManager.enemy.stats.finalDamageReduce * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="战斗属性提升">
                      {{ (dungeonState.combatManager.enemy.stats.combatBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                    <n-descriptions-item label="战斗抗性提升">
                      {{ (dungeonState.combatManager.enemy.stats.resistanceBoost * 100 || 0).toFixed(1) }}%
                    </n-descriptions-item>
                  </n-descriptions>
                </template>
              </n-card>
            </n-modal>
          </n-card>
        </template>
      </n-space>
    </n-card>
  </div>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { getAuthToken } from '../stores/db'
  import APIService from '../services/api'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import LogPanel from '../components/LogPanel.vue'

  const playerInfoStore = usePlayerInfoStore()
  
  // 初始化难度为 'easy'（凡界）
  if (!playerInfoStore.dungeonDifficulty) {
    playerInfoStore.dungeonDifficulty = 'easy'
  }
  const message = useMessage()
  const combatLog = ref([])
  const logRef = ref(null)
  const refreshNumber = ref(3)
  const infoShow = ref(false)
  const infoType = ref('')
  const playerAttacking = ref(false)
  const enemyAttacking = ref(false)
  const playerHurt = ref(false)
  const enemyHurt = ref(false)

  // 秘境难度选项
  const dungeonOptions = [
    { label: '凡界', value: 'easy' },
    { label: '小世界', value: 'normal' },
    { label: '灵界', value: 'hard' },
    { label: '仙界', value: 'expert' }
  ]

  // 秘境状态
  const dungeonState = ref({
    floor: 1,
    inCombat: false,
    showingOptions: false,
    currentOptions: [],
    combatManager: null
  })

  // 玩家和敌人属性现在从后端API获取，无需本地计算

  // 添加战斗日志
  const addBattleLog = (message) => {
    combatLog.value.push({
      time: new Date().toLocaleTimeString(),
      content: message
    })
    // 限制日志数量
    if (combatLog.value.length > 100) {
      combatLog.value.shift()
    }
  }

  // 开始秘境探索
  const startDungeon = async () => {
    try {
      const token = getAuthToken()
      const data = await APIService.post('/dungeon/start', {
        difficulty: playerInfoStore.dungeonDifficulty
      }, token)
      if (data.success) {
        dungeonState.value.floor = data.data.floor
        dungeonState.value.showingOptions = true
        refreshNumber.value = data.data.refreshCount
        generateOptions()
        message.success('秘境已开启')
      } else {
        message.error(data.message || '开启秘境失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 生成选项（调用后端 API）
  const generateOptions = async () => {
    try {
      const token = getAuthToken()
      const data = await APIService.get(`/dungeon/buffs/${dungeonState.value.floor}`, {}, token)
      if (data.success) {
        dungeonState.value.currentOptions = data.data.options.map(opt => ({
          ...opt,
          type: opt.type || 'common'
        }))
      } else {
        message.error('获取增益失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 刷新选项
  const handleRefreshOptions = async () => {
    if (refreshNumber.value > 0) {
      refreshNumber.value--
      await generateOptions()
    }
  }

  // 选择选项
  const selectOption = async (option) => {
    try {
      const token = getAuthToken()
      const data = await APIService.post('/dungeon/save-buff', {
        floor: dungeonState.value.floor,
        difficulty: playerInfoStore.dungeonDifficulty,
        buffId: option.id
      }, token)
      if (data.success) {
        message.success(`已选择增益：${option.name}`)
        
        // 由后端管理流程，前端不再维护 floor++
        dungeonState.value.showingOptions = false
        dungeonState.value.floor = data.data.floor // 使用后端返回的下一层整数
        
        // 自动开始战斗
        setTimeout(() => {
          startFight()
        }, 1000)
      } else {
        message.error(data.message || '选择增益失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 开始战斗（调用后端 API）
  const startFight = async () => {
    try {
      const token = getAuthToken()
      const data = await APIService.post('/dungeon/fight', {
        floor: dungeonState.value.floor,
        difficulty: playerInfoStore.dungeonDifficulty
      }, token)
      if (data.success) {
        const result = data.data
        dungeonState.value.inCombat = true
        // 从后端API获取玩家和敌人属性，而不是本地计算
        dungeonState.value.combatManager = {
          round: 1,
          maxRounds: result.maxRounds,
          player: {
            name: playerInfoStore.name,
            currentHealth: result.playerHealth,
            stats: result.playerStats
          },
          enemy: {
            name: '秘境守卫',
            currentHealth: result.enemyHealth,
            stats: result.enemyStats
          }
        }
        
        // 启动轮询执行回合
        addBattleLog('战斗已初始化，开始逐回合执行战斗...')
        // 添加消耗信息到日志
        if (result.spiritCost && result.stoneCost) {
          addBattleLog(`消耗: 灵力 -${result.spiritCost}, 灵石 -${result.stoneCost}`)
        }
        await pollAndExecuteRounds(token)
      } else {
        message.error(data.message || '开始战斗失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 轮询执行回合
  const pollAndExecuteRounds = async (token) => {
    let isVictory = false
    let isDefeat = false
    
    while (dungeonState.value.inCombat && !isVictory && !isDefeat) {
      try {
        // 1. 调用后端执行当前回合
        const executeRes = await APIService.post('/dungeon/execute-round', {
          floor: dungeonState.value.floor,
          difficulty: playerInfoStore.dungeonDifficulty
        }, token)
        
        if (!executeRes.success) {
          message.error('执行回合失败：' + executeRes.message)
          break
        }
        
        // 2. 轮询获取回合结果（每3秒查询一次）
        let roundData = null
        let pollCount = 0
        const maxPolls = 5 // 最多轮询5次（15秒超时）
        
        while (pollCount < maxPolls && !roundData) {
          await new Promise(resolve => setTimeout(resolve, 3000)) // 等待3秒
          
          const pollRes = await APIService.get('/dungeon/round-data', {}, token)
          if (pollRes.success && pollRes.data) {
            roundData = pollRes.data
            break
          }
          
          pollCount++
        }
        
        if (!roundData) {
          message.error('获取回合数据超时')
          break
        }
        
        // 3. 更新战斗界面
        dungeonState.value.combatManager.round = roundData.round
        dungeonState.value.combatManager.player.currentHealth = roundData.playerHealth
        dungeonState.value.combatManager.enemy.currentHealth = roundData.enemyHealth
        
        // 4. 显示日志
        if (roundData.logs && roundData.logs.length > 0) {
          roundData.logs.forEach(log => {
            addBattleLog(log)
          })
        }
        
        // 5. 检查战斗是否结束
        if (roundData.battleEnded) {
          if (roundData.victory) {
            addBattleLog('战斗胜利！')
            isVictory = true
            if (roundData.rewards) {
              gainBattleReward(roundData.rewards)
            }
          } else {
            addBattleLog('战斗失败！')
            isDefeat = true
          }
        }
        
        // 为了可视化效果，短暂延迟后继续下一回合
        await new Promise(resolve => setTimeout(resolve, 1000))
        
      } catch (error) {
        message.error('轮询过程中出错：' + error.message)
        break
      }
    }
    
    // 战斗结束处理
    dungeonState.value.inCombat = false
    
    if (isVictory) {
      message.success('战斗胜利')
      // 一段时间后会自动生成新的增益选项
      setTimeout(() => {
        dungeonState.value.showingOptions = true
        generateOptions()
      }, 2000)
    } else if (isDefeat) {
      message.error('战斗失败，秘境探索结束')
      // 结束秘境
      endDungeon(false)
    }
  }

  // 结束秘境
  const endDungeon = async (victory) => {
    try {
      const token = getAuthToken()
      const data = await APIService.post('/dungeon/end', {
        floor: dungeonState.value.floor,
        victory: victory
      }, token)
      if (data.success) {
        // 重置秘境状态
        dungeonState.value.floor = 1
        dungeonState.value.inCombat = false
        dungeonState.value.showingOptions = false
        dungeonState.value.currentOptions = []
        refreshNumber.value = 3
        
        if (victory) {
          message.success(`成功探索到第${data.data.floor}层，获得${data.data.totalReward}个灵石`)
          // 更新上u云数据
          playerInfoStore.spiritStones = data.data.spiritStones
        }
      } else {
        message.error(data.message || '结束秘境失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 获取选项颜色
  const getOptionColor = (type) => {
    switch (type) {
      case 'common':
        return { name: '普通', color: '#cccccc' }
      case 'rare':
        return { name: '稀有', color: '#0066ff' }
      case 'epic':
        return { name: '史诗', color: '#cc00ff' }
      default:
        return { name: '普通', color: '#cccccc' }
    }
  }

  // 信息点击
  const infoCliclk = (type) => {
    infoType.value = type
    infoShow.value = true
  }

  onMounted(() => {
    // 组件挂载时的初始化
  })

  onUnmounted(() => {
    // Worker 已移除
  })
</script>

<style scoped>
  .dungeon-container {
    margin: 0 auto;
  }

  .option-cards {
    display: flex;
    gap: 16px;
    padding: 16px;
    margin: 0 auto;
  }

  .option-card {
    position: relative;
    padding: 20px;
    border: 2px solid;
    border-radius: 12px;
    background: var(--n-color);
    cursor: pointer;
    transition: all 0.3s ease;
    display: flex;
    flex-direction: column;
    min-height: 100px;
    width: 33%;
  }

  .option-card:hover {
    transform: translateX(5px);
    box-shadow: 4px 4px 12px rgba(0, 0, 0, 0.1);
  }

  .option-name {
    font-size: 1.3em;
    font-weight: bold;
    margin-bottom: 12px;
    padding-right: 80px;
  }

  .option-description {
    flex-grow: 1;
    font-size: 1em;
    color: var(--n-text-color);
    line-height: 1.6;
    margin-bottom: 8px;
  }

  .option-quality {
    position: absolute;
    top: 20px;
    right: 20px;
    font-size: 0.9em;
    font-weight: bold;
    padding: 4px 12px;
    border-radius: 20px;
    background: var(--n-color);
  }

  .combat-scene {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px;
    margin-bottom: 20px;
    min-height: 200px;
    background: rgba(0, 0, 0, 0.05);
    border-radius: 8px;
  }

  .character {
    display: flex;
    flex-direction: column;
    align-items: center;
    transition: transform 0.3s ease;
  }

  .character-avatar {
    font-size: 48px;
    margin: 10px 0;
  }

  .character-name {
    font-weight: bold;
    margin-bottom: 8px;
  }

  .health-bar {
    width: 100px;
    height: 10px;
    background: #ff000033;
    border-radius: 5px;
    overflow: hidden;
  }

  .health-fill {
    height: 100%;
    background: #ff0000;
    transition: width 0.3s ease;
  }

  .character.attack {
    animation: attack 0.5s ease;
  }

  .character.hurt {
    animation: hurt 0.5s ease;
  }

  .character-avatar {
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    font-weight: bold;
    margin: 10px 0;
    color: #fff;
  }

  .player-avatar {
    background: linear-gradient(135deg, #4caf50, #2196f3);
    border-radius: 12px;
  }

  .enemy-avatar {
    background: linear-gradient(135deg, #ff5722, #e91e63);
    clip-path: polygon(50% 0%, 100% 38%, 100% 100%, 0 100%, 0% 38%);
  }

  .attack-effect {
    position: absolute;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    pointer-events: none;
  }

  .player-effect {
    background: radial-gradient(circle, #4caf50, #2196f3);
    animation: player-attack-effect 0.5s ease-out;
    right: -10px;
  }

  .enemy-effect {
    background: radial-gradient(circle, #ff5722, #e91e63);
    animation: enemy-attack-effect 0.5s ease-out;
    left: -10px;
  }

  .enemy.attack {
    animation: enemy-attack 0.5s ease;
  }

  @keyframes player-attack-effect {
    0% {
      transform: scale(0.5) translateX(0);
      opacity: 1;
    }
    100% {
      transform: scale(1.5) translateX(200px);
      opacity: 0;
    }
  }

  @keyframes enemy-attack-effect {
    0% {
      transform: scale(0.5) translateX(0);
      opacity: 1;
    }
    100% {
      transform: scale(1.5) translateX(-200px);
      opacity: 0;
    }
  }

  @keyframes attack {
    0% {
      transform: translateX(0) rotate(0deg);
    }
    25% {
      transform: translateX(20px) rotate(5deg);
    }
    50% {
      transform: translateX(40px) rotate(0deg);
    }
    75% {
      transform: translateX(20px) rotate(-5deg);
    }
    100% {
      transform: translateX(0) rotate(0deg);
    }
  }

  @keyframes hurt {
    0% {
      transform: translateX(0);
    }
    25% {
      transform: translateX(-10px);
    }
    75% {
      transform: translateX(10px);
    }
    100% {
      transform: translateX(0);
    }
  }

  @keyframes enemy-attack {
    0% {
      transform: translateX(0) rotate(0deg);
    }
    25% {
      transform: translateX(-20px) rotate(-5deg);
    }
    50% {
      transform: translateX(-40px) rotate(0deg);
    }
    75% {
      transform: translateX(-20px) rotate(5deg);
    }
    100% {
      transform: translateX(0) rotate(0deg);
    }
  }
</style>
