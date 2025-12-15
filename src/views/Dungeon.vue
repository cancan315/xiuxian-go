<template>
  <div class="dungeon-container">
    <n-card title="秘境探索">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="statsStore.dungeonDifficulty"
            placeholder="请选择难度"
            :options="dungeonOptions"
            style="width: 120px"
            :disabled="dungeonState.inCombat || dungeonState.showingOptions"
          />
          <n-button
            type="primary"
            @click="startDungeon"
            :disabled="dungeonState.inCombat || dungeonState.showingOptions"
          >
            开始探索
          </n-button>
        </n-space>
      </template>
      <n-space vertical>
        <!-- 层数显示 -->
        <n-statistic label="当前层数" :value="dungeonState.floor" />
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
            <!-- 战斗日志 -->
            <log-panel ref="logRef" :messages="combatLog" style="margin-top: 16px" />
          </n-card>
        </template>
      </n-space>
    </n-card>
  </div>
</template>

<script setup>
  // 修改为使用模块化store
  import { usePlayerInfoStore } from '../stores/playerInfo'
  import { useInventoryStore } from '../stores/inventory'
  import { useEquipmentStore } from '../stores/equipment'
  import { usePetsStore } from '../stores/pets'
  import { usePillsStore } from '../stores/pills'
  import { useSettingsStore } from '../stores/settings'
  import { useStatsStore } from '../stores/stats'
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useMessage } from 'naive-ui'
  import LogPanel from '../components/LogPanel.vue'

  const playerInfoStore = usePlayerInfoStore()
  const inventoryStore = useInventoryStore()
  const equipmentStore = useEquipmentStore()
  const petsStore = usePetsStore()
  const pillsStore = usePillsStore()
  const settingsStore = useSettingsStore()
  const statsStore = useStatsStore()
  
  const message = useMessage()
  const showBattleLog = ref(false)
  const battleLogs = ref([])
  const currentEnemy = ref(null)
  const isInBattle = ref(false)
  const battleResult = ref(null)
  const dungeonWorker = ref(null)
  const infoShow = ref(false)
  const infoType = ref('')
  const playerAttacking = ref(false)
  const enemyAttacking = ref(false)
  const playerHurt = ref(false)
  const enemyHurt = ref(false)
  const refreshNumber = ref(3)
  const combatLog = ref([])
  const logRef = ref(null)

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

  // 获取玩家总属性（基础属性+装备加成+灵宠加成）
  const playerStats = computed(() => {
    // 基础属性
    const base = { ...playerInfoStore.baseAttributes }
    
    // 计算装备加成（已废弃，装备属性加成由后端计算并同步到playerInfoStore）
    const equipmentBonus = {
      attack: 0,
      defense: 0,
      health: 0,
      speed: 0,
      critRate: 0,
      comboRate: 0,
      counterRate: 0,
      stunRate: 0,
      dodgeRate: 0,
      vampireRate: 0,
      critResist: 0,
      comboResist: 0,
      counterResist: 0,
      stunResist: 0,
      dodgeResist: 0,
      vampireResist: 0,
      healBoost: 0,
      critDamageBoost: 0,
      critDamageReduce: 0,
      finalDamageBoost: 0,
      finalDamageReduce: 0,
      combatBoost: 0,
      resistanceBoost: 0
    }
    
    // 计算灵宠加成（已移除，因为灵宠属性加成逻辑已转移到后端API处理）
    const petBonus = {
      attack: 0,
      defense: 0,
      health: 0,
      critRate: 0,
      comboRate: 0,
      counterRate: 0,
      stunRate: 0,
      dodgeRate: 0,
      vampireRate: 0,
      critResist: 0,
      comboResist: 0,
      counterResist: 0,
      stunResist: 0,
      dodgeResist: 0,
      vampireResist: 0,
      healBoost: 0,
      critDamageBoost: 0,
      critDamageReduce: 0,
      finalDamageBoost: 0,
      finalDamageReduce: 0,
      combatBoost: 0,
      resistanceBoost: 0
    }
    
    // 合并所有属性
    return {
      attack: base.attack + equipmentBonus.attack, // 移除了 petBonus.attack
      defense: base.defense + equipmentBonus.defense, // 移除了 petBonus.defense
      health: base.health + equipmentBonus.health, // 移除了 petBonus.health
      speed: base.speed + equipmentBonus.speed, // 移除了 petBonus.speed
      critRate: Math.min(1, playerInfoStore.combatAttributes.critRate + equipmentBonus.critRate), // 移除了 petBonus.critRate
      comboRate: Math.min(1, playerInfoStore.combatAttributes.comboRate + equipmentBonus.comboRate), // 移除了 petBonus.comboRate
      counterRate: Math.min(1, playerInfoStore.combatAttributes.counterRate + equipmentBonus.counterRate), // 移除了 petBonus.counterRate
      stunRate: Math.min(1, playerInfoStore.combatAttributes.stunRate + equipmentBonus.stunRate), // 移除了 petBonus.stunRate
      dodgeRate: Math.min(1, playerInfoStore.combatAttributes.dodgeRate + equipmentBonus.dodgeRate), // 移除了 petBonus.dodgeRate
      vampireRate: Math.min(1, playerInfoStore.combatAttributes.vampireRate + equipmentBonus.vampireRate), // 移除了 petBonus.vampireRate
      critResist: Math.min(1, playerInfoStore.combatResistance.critResist + equipmentBonus.critResist), // 移除了 petBonus.critResist
      comboResist: Math.min(1, playerInfoStore.combatResistance.comboResist + equipmentBonus.comboResist), // 移除了 petBonus.comboResist
      counterResist: Math.min(1, playerInfoStore.combatResistance.counterResist + equipmentBonus.counterResist), // 移除了 petBonus.counterResist
      stunResist: Math.min(1, playerInfoStore.combatResistance.stunResist + equipmentBonus.stunResist), // 移除了 petBonus.stunResist
      dodgeResist: Math.min(1, playerInfoStore.combatResistance.dodgeResist + equipmentBonus.dodgeResist), // 移除了 petBonus.dodgeResist
      vampireResist: Math.min(1, playerInfoStore.combatResistance.vampireResist + equipmentBonus.vampireResist), // 移除了 petBonus.vampireResist
      healBoost: playerInfoStore.specialAttributes.healBoost + equipmentBonus.healBoost, // 移除了 petBonus.healBoost
      critDamageBoost: playerInfoStore.specialAttributes.critDamageBoost + equipmentBonus.critDamageBoost, // 移除了 petBonus.critDamageBoost
      critDamageReduce: playerInfoStore.specialAttributes.critDamageReduce + equipmentBonus.critDamageReduce, // 移除了 petBonus.critDamageReduce
      finalDamageBoost: playerInfoStore.specialAttributes.finalDamageBoost + equipmentBonus.finalDamageBoost, // 移除了 petBonus.finalDamageBoost
      finalDamageReduce: playerInfoStore.specialAttributes.finalDamageReduce + equipmentBonus.finalDamageReduce, // 移除了 petBonus.finalDamageReduce
      combatBoost: playerInfoStore.specialAttributes.combatBoost + equipmentBonus.combatBoost, // 移除了 petBonus.combatBoost
      resistanceBoost: playerInfoStore.specialAttributes.resistanceBoost + equipmentBonus.resistanceBoost // 移除了 petBonus.resistanceBoost
    }
  })

  // 敌人属性计算
  const enemyStats = computed(() => {
    if (!currentEnemy.value) return null
    
    // 基础属性基于难度和楼层
    const baseAttack = Math.floor(5 * playerStats.value.attack * (0.5 + 0.1 * statsStore.dungeonDifficulty) * (1 + currentEnemy.value.floor * 0.05))
    const baseDefense = Math.floor(2 * playerStats.value.defense * (0.5 + 0.1 * statsStore.dungeonDifficulty) * (1 + currentEnemy.value.floor * 0.05))
    const baseHealth = Math.floor(20 * playerStats.value.health * (0.5 + 0.1 * statsStore.dungeonDifficulty) * (1 + currentEnemy.value.floor * 0.05))
    const baseSpeed = Math.floor(3 * playerStats.value.speed * (0.5 + 0.1 * statsStore.dungeonDifficulty) * (1 + currentEnemy.value.floor * 0.05))
    
    return {
      attack: baseAttack,
      defense: baseDefense,
      health: baseHealth,
      maxHealth: baseHealth,
      speed: baseSpeed,
      critRate: 0.05 * statsStore.dungeonDifficulty,
      comboRate: 0.05 * statsStore.dungeonDifficulty,
      counterRate: 0.05 * statsStore.dungeonDifficulty,
      stunRate: 0.05 * statsStore.dungeonDifficulty,
      dodgeRate: 0.05 * statsStore.dungeonDifficulty,
      vampireRate: 0.05 * statsStore.dungeonDifficulty,
      critResist: 0.05 * statsStore.dungeonDifficulty,
      comboResist: 0.05 * statsStore.dungeonDifficulty,
      counterResist: 0.05 * statsStore.dungeonDifficulty,
      stunResist: 0.05 * statsStore.dungeonDifficulty,
      dodgeResist: 0.05 * statsStore.dungeonDifficulty,
      vampireResist: 0.05 * statsStore.dungeonDifficulty,
      healBoost: 0,
      critDamageBoost: 0,
      critDamageReduce: 0,
      finalDamageBoost: 0,
      finalDamageReduce: 0,
      combatBoost: 0,
      resistanceBoost: 0
    }
  })

  // 添加战斗日志
  const addBattleLog = (message) => {
    battleLogs.value.push({
      time: new Date().toLocaleTimeString(),
      message: message
    })
    // 限制日志数量
    if (battleLogs.value.length > 100) {
      battleLogs.value.shift()
    }
  }

  // 开始战斗
  const startBattle = (enemy) => {
    currentEnemy.value = enemy
    isInBattle.value = true
    battleLogs.value = []
    battleResult.value = null
    addBattleLog(`遭遇${enemy.name}！`)
    
    // 初始化战斗
    if (dungeonWorker.value) {
      dungeonWorker.value.terminate()
    }
    
    dungeonWorker.value = new Worker(new URL('../workers/exploration.js', import.meta.url))
    dungeonWorker.value.onmessage = e => {
      const { type, data } = e.data
      if (type === 'battle_update') {
        // 更新战斗状态
        Object.assign(currentEnemy.value, data.enemy)
        addBattleLog(data.log)
      } else if (type === 'battle_end') {
        // 战斗结束
        finishBattle(data.result)
      }
    }
    
    dungeonWorker.value.postMessage({
      type: 'start_battle',
      player: playerStats.value,
      enemy: enemyStats.value,
      enemyInfo: enemy
    })
  }

  // 结束战斗
  const finishBattle = (result) => {
    if (dungeonWorker.value) {
      dungeonWorker.value.terminate()
      dungeonWorker.value = null
    }
    
    isInBattle.value = false
    battleResult.value = result
    
    if (result.victory) {
      // 战斗胜利
      addBattleLog('战斗胜利！')
      
      // 增加统计数据
      if (currentEnemy.value.isBoss) {
        statsStore.dungeonBossKills += 1
      } else if (currentEnemy.value.isElite) {
        statsStore.dungeonEliteKills += 1
      } else {
        statsStore.dungeonTotalKills += 1
      }
      
      // 获得奖励
      gainBattleReward(result.rewards)
    } else {
      // 战斗失败
      addBattleLog('战斗失败！')
      statsStore.dungeonDeathCount += 1
    }
    
    statsStore.dungeonTotalRuns += 1
  }

  // 获得战斗奖励
  const gainBattleReward = (rewards) => {
    rewards.forEach(reward => {
      switch (reward.type) {
        case 'spirit_stone':
          inventoryStore.spiritStones += reward.amount
          addBattleLog(`获得${reward.amount}灵石`)
          break
        case 'herb':
          // 添加灵草到背包
          const existingHerb = inventoryStore.herbs.find(h => h.id === reward.herb.id)
          if (existingHerb) {
            existingHerb.count += reward.amount
          } else {
            inventoryStore.herbs.push({
              ...reward.herb,
              count: reward.amount
            })
          }
          addBattleLog(`获得${reward.amount}个${reward.herb.name}`)
          break
        case 'equipment':
          inventoryStore.items.push(reward.equipment)
          addBattleLog(`获得装备${reward.equipment.name}`)
          break
        case 'pet':
          petsStore.pets.push(reward.pet)
          addBattleLog(`获得灵宠${reward.pet.name}`)
          break
      }
    })
    statsStore.dungeonTotalRewards += 1
  }

  // 逃跑
  const flee = () => {
    if (dungeonWorker.value) {
      dungeonWorker.value.terminate()
      dungeonWorker.value = null
    }
    
    isInBattle.value = false
    currentEnemy.value = null
    addBattleLog('你逃跑了！')
  }

  // 开始秘境探索
  const startDungeon = async () => {
    try {
      const response = await fetch('/api/dungeon/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          difficulty: statsStore.dungeonDifficulty
        })
      })
      const data = await response.json()
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

  // 生成选项（调用后端API）
  const generateOptions = async () => {
    try {
      const response = await fetch(`/api/dungeon/buffs/${dungeonState.value.floor}`, {
        method: 'GET'
      })
      const data = await response.json()
      if (data.success) {
        dungeonState.value.currentOptions = data.data.options.map(opt => ({
          ...opt,
          type: opt.Type || 'common'
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
      const response = await fetch('/api/dungeon/select-buff', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          buffID: option.id
        })
      })
      const data = await response.json()
      if (data.success) {
        message.success(`已选择增益：${option.name}`)
        
        // 进入下一层或者开始战斗
        dungeonState.value.showingOptions = false
        dungeonState.value.floor++
        
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

  // 开始战斗（调用后端API）
  const startFight = async () => {
    try {
      const response = await fetch('/api/dungeon/fight', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          floor: dungeonState.value.floor,
          difficulty: statsStore.dungeonDifficulty
        })
      })
      const data = await response.json()
      if (data.success) {
        const result = data.data
        dungeonState.value.inCombat = true
        dungeonState.value.combatManager = {
          round: 1,
          maxRounds: 10,
          player: {
            name: playerInfoStore.name,
            currentHealth: playerStats.value.health,
            stats: playerStats.value
          },
          enemy: {
            name: '秘境守卫',
            currentHealth: 100,
            stats: {
              maxHealth: 100,
              damage: 20,
              defense: 10,
              speed: 15,
              critRate: 0.1,
              comboRate: 0.1,
              counterRate: 0.1,
              stunRate: 0.1,
              dodgeRate: 0.1,
              vampireRate: 0.1,
              critResist: 0.1,
              comboResist: 0.1,
              counterResist: 0.1,
              stunResist: 0.1,
              dodgeResist: 0.1,
              vampireResist: 0.1,
              healBoost: 0,
              critDamageBoost: 0,
              critDamageReduce: 0,
              finalDamageBoost: 0,
              finalDamageReduce: 0,
              combatBoost: 0,
              resistanceBoost: 0
            }
          }
        }
        
        // 处理战斗结果
        if (result.Victory) {
          addBattleLog('战斗胜利！')
          if (result.Rewards) {
            gainBattleReward(result.Rewards)
          }
        } else {
          addBattleLog('战斗失败！')
        }
        
        // 结束战斗
        setTimeout(() => {
          dungeonState.value.inCombat = false
          if (result.Victory) {
            message.success('战斗胜利')
            // 一段时间后会自动生成新的增益选项
            setTimeout(() => {
              dungeonState.value.showingOptions = true
              generateOptions()
            }, 2000)
          } else {
            message.error('战斗失败。懂悲，秘境探索结束')
            // 结束秘境
            endDungeon(false)
          }
        }, 3000)
      } else {
        message.error(data.message || '开始战斗失败')
      }
    } catch (error) {
      message.error('网络错误：' + error.message)
    }
  }

  // 结束秘境
  const endDungeon = async (victory) => {
    try {
      const response = await fetch('/api/dungeon/end', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          floor: dungeonState.value.floor,
          victory: victory
        })
      })
      const data = await response.json()
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

  onUnmounted(() => {
    if (dungeonWorker.value) {
      dungeonWorker.value.terminate()
    }
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
