/**
 * 战斗模拟器
 * 负责处理战斗逻辑、伤害计算、状态效果等
 */

// 战斗日志类型
export const LogType = {
  ATTACK: 'attack',
  HEAL: 'heal',
  BUFF: 'buff',
  DEBUFF: 'debuff',
  SPECIAL: 'special',
  INFO: 'info'
}

// 战斗结果
export const BattleResult = {
  VICTORY: 'victory',
  DEFEAT: 'defeat',
  DRAW: 'draw'
}

/**
 * 计算伤害
 * @param {Object} attacker 攻击者属性
 * @param {Object} defender 防御者属性
 * @param {boolean} isCritical 是否暴击
 * @returns {number} 伤害值
 */
export function calculateDamage(attacker, defender, isCritical = false) {
  // 基础伤害公式：攻击力 * (1 - 减伤率)
  const defenseReduction = defender.defense / (defender.defense + 100)
  let baseDamage = attacker.attack * (1 - defenseReduction)
  
  // 暴击伤害
  if (isCritical) {
    const critMultiplier = 1.5 + (attacker.critDamageBoost || 0) - (defender.critDamageReduce || 0)
    baseDamage *= critMultiplier
  }
  
  // 最终伤害调整
  baseDamage *= (1 + (attacker.finalDamageBoost || 0) - (defender.finalDamageReduce || 0))
  
  return Math.max(Math.floor(baseDamage), 1) // 最少造成1点伤害
}

/**
 * 判断是否触发效果
 * @param {number} rate 触发率 (0-1)
 * @returns {boolean} 是否触发
 */
export function checkEffectTrigger(rate) {
  return Math.random() < rate
}

/**
 * 战斗参与者类
 */
export class BattleParticipant {
  constructor(data, isMonster = false) {
    this.id = data.id
    this.name = data.name
    this.isMonster = isMonster
    
    // 基础属性
    this.maxHealth = data.baseAttributes?.health || data.health || 100
    this.currentHealth = data.baseAttributes?.health || data.health || 100
    this.attack = data.baseAttributes?.attack || data.attack || 10
    this.defense = data.baseAttributes?.defense || data.defense || 5
    this.speed = data.baseAttributes?.speed || data.speed || 10
    
    // 战斗属性
    this.critRate = data.combatAttributes?.critRate || data.critRate || 0.05
    this.comboRate = data.combatAttributes?.comboRate || data.comboRate || 0
    this.counterRate = data.combatAttributes?.counterRate || data.counterRate || 0
    this.stunRate = data.combatAttributes?.stunRate || data.stunRate || 0
    this.dodgeRate = data.combatAttributes?.dodgeRate || data.dodgeRate || 0.05
    this.vampireRate = data.combatAttributes?.vampireRate || data.vampireRate || 0
    
    // 战斗抗性
    this.critResist = data.combatResistance?.critResist || data.critResist || 0
    this.comboResist = data.combatResistance?.comboResist || data.comboResist || 0
    this.counterResist = data.combatResistance?.counterResist || data.counterResist || 0
    this.stunResist = data.combatResistance?.stunResist || data.stunResist || 0
    this.dodgeResist = data.combatResistance?.dodgeResist || data.dodgeResist || 0
    this.vampireResist = data.combatResistance?.vampireResist || data.vampireResist || 0
    
    // 特殊属性
    this.healBoost = data.specialAttributes?.healBoost || data.healBoost || 0
    this.critDamageBoost = data.specialAttributes?.critDamageBoost || data.critDamageBoost || 0.5
    this.critDamageReduce = data.specialAttributes?.critDamageReduce || data.critDamageReduce || 0
    this.finalDamageBoost = data.specialAttributes?.finalDamageBoost || data.finalDamageBoost || 0
    this.finalDamageReduce = data.specialAttributes?.finalDamageReduce || data.finalDamageReduce || 0
    this.combatBoost = data.specialAttributes?.combatBoost || data.combatBoost || 0
    this.resistanceBoost = data.specialAttributes?.resistanceBoost || data.resistanceBoost || 0
    
    // 状态效果
    this.isStunned = false
    this.buffs = []
    this.debuffs = []
  }
  
  // 检查是否存活
  get isAlive() {
    return this.currentHealth > 0
  }
  
  // 添加状态效果
  addBuff(buff) {
    this.buffs.push(buff)
  }
  
  addDebuff(debuff) {
    this.debuffs.push(debuff)
  }
  
  // 移除状态效果
  removeBuff(buffId) {
    this.buffs = this.buffs.filter(b => b.id !== buffId)
  }
  
  removeDebuff(debuffId) {
    this.debuffs = this.debuffs.filter(d => d.id !== debuffId)
  }
  
  // 更新状态效果
  updateStatusEffects() {
    // 更新buff/debuff持续时间
    this.buffs = this.buffs.filter(buff => --buff.duration > 0)
    this.debuffs = this.debuffs.filter(debuff => --debuff.duration > 0)
    
    // 清除眩晕状态（如果持续时间结束）
    if (this.isStunned) {
      const stunDebuff = this.debuffs.find(d => d.type === 'stun')
      if (!stunDebuff) {
        this.isStunned = false
      }
    }
  }
  
  // 获取实际属性值（考虑buff/debuff影响）
  getEffectiveStat(stat) {
    let value = this[stat]
    
    // 应用buff加成
    this.buffs.forEach(buff => {
      if (buff.affects.includes(stat)) {
        value *= (1 + buff.value)
      }
    })
    
    // 应用debuff减成
    this.debuffs.forEach(debuff => {
      if (debuff.affects.includes(stat)) {
        value *= (1 - debuff.value)
      }
    })
    
    // 战斗属性提升（来自特殊属性）
    if (['critRate', 'comboRate', 'counterRate', 'stunRate', 'dodgeRate', 'vampireRate'].includes(stat)) {
      value *= (1 + this.combatBoost)
    }
    
    // 战斗抗性提升（来自特殊属性）
    if (['critResist', 'comboResist', 'counterResist', 'stunResist', 'dodgeResist', 'vampireResist'].includes(stat)) {
      value *= (1 + this.resistanceBoost)
    }
    
    return Math.max(value, 0)
  }
}

/**
 * 战斗管理器类
 */
export class BattleManager {
  constructor(playerData, opponentData, isPvE = false) {
    this.player = new BattleParticipant(playerData)
    this.opponent = new BattleParticipant(opponentData, isPvE)
    this.isPvE = isPvE
    this.turn = 0
    this.logs = []
    this.result = null
    
    // 添加战斗开始日志
    this.addLog(LogType.INFO, `战斗开始！${this.player.name} vs ${this.opponent.name}`)
  }
  
  // 添加日志
  addLog(type, message) {
    this.logs.push({
      type,
      message,
      turn: this.turn
    })
  }
  
  // 执行一个回合
  executeTurn() {
    this.turn++
    
    // 更新状态效果
    this.player.updateStatusEffects()
    this.opponent.updateStatusEffects()
    
    // 决定行动顺序（基于速度）
    let first, second
    if (this.player.speed >= this.opponent.speed) {
      first = this.player
      second = this.opponent
    } else {
      first = this.opponent
      second = this.player
    }
    
    // 执行行动
    this.performAction(first, second)
    
    // 检查战斗是否结束
    if (!second.isAlive) {
      this.result = first === this.player ? BattleResult.VICTORY : BattleResult.DEFEAT
      this.addLog(LogType.INFO, `${first.name} 获胜！`)
      return this.result
    }
    
    // 第二方行动（如果还活着）
    if (second.isAlive && !second.isStunned) {
      this.performAction(second, first)
      
      // 检查战斗是否结束
      if (!first.isAlive) {
        this.result = second === this.player ? BattleResult.VICTORY : BattleResult.DEFEAT
        this.addLog(LogType.INFO, `${second.name} 获胜！`)
        return this.result
      }
    } else if (second.isStunned) {
      this.addLog(LogType.INFO, `${second.name} 被眩晕，无法行动`)
    }
    
    return null
  }
  
  // 执行单个行动
  performAction(attacker, defender) {
    // 检查是否被眩晕
    if (attacker.isStunned) {
      this.addLog(LogType.INFO, `${attacker.name} 被眩晕，无法行动`)
      attacker.isStunned = false // 解除眩晕
      return
    }
    
    // 检查闪避
    const dodgeChance = defender.getEffectiveStat('dodgeRate') - attacker.getEffectiveStat('dodgeResist')
    if (checkEffectTrigger(Math.max(dodgeChance, 0))) {
      this.addLog(LogType.INFO, `${defender.name} 闪避了攻击！`)
      return
    }
    
    // 检查是否暴击
    const critChance = attacker.getEffectiveStat('critRate') - defender.getEffectiveStat('critResist')
    const isCritical = checkEffectTrigger(Math.max(critChance, 0))
    
    // 计算伤害
    const damage = calculateDamage(attacker, defender, isCritical)
    
    // 应用伤害
    defender.currentHealth -= damage
    
    // 添加攻击日志
    const critText = isCritical ? '[暴击!] ' : ''
    this.addLog(LogType.ATTACK, `${critText}${attacker.name} 对 ${defender.name} 造成 ${damage} 点伤害`)
    
    // 检查吸血
    const vampireChance = attacker.getEffectiveStat('vampireRate') - defender.getEffectiveStat('vampireResist')
    if (checkEffectTrigger(Math.max(vampireChance, 0)) && damage > 0) {
      const healAmount = Math.floor(damage * 0.3) // 吸血30%
      attacker.currentHealth = Math.min(attacker.currentHealth + healAmount, attacker.maxHealth)
      this.addLog(LogType.HEAL, `${attacker.name} 吸血 ${healAmount} 点`)
    }
    
    // 检查连击
    const comboChance = attacker.getEffectiveStat('comboRate') - defender.getEffectiveStat('comboResist')
    if (checkEffectTrigger(Math.max(comboChance, 0)) && defender.isAlive) {
      const comboDamage = Math.floor(damage * 0.5) // 连击造成50%伤害
      defender.currentHealth -= comboDamage
      this.addLog(LogType.ATTACK, `[连击!] ${attacker.name} 追加 ${comboDamage} 点伤害`)
    }
    
    // 检查眩晕
    const stunChance = attacker.getEffectiveStat('stunRate') - defender.getEffectiveStat('stunResist')
    if (checkEffectTrigger(Math.max(stunChance, 0)) && defender.isAlive) {
      defender.isStunned = true
      defender.addDebuff({
        id: 'stun',
        type: 'stun',
        duration: 1,
        value: 0,
        affects: []
      })
      this.addLog(LogType.DEBUFF, `${defender.name} 被眩晕了！`)
    }
    
    // 检查反击（只有当defender还活着且没有被眩晕时）
    if (defender.isAlive && !defender.isStunned) {
      const counterChance = defender.getEffectiveStat('counterRate') - attacker.getEffectiveStat('counterResist')
      if (checkEffectTrigger(Math.max(counterChance, 0))) {
        const counterDamage = Math.floor(damage * 0.4) // 反击造成40%伤害
        attacker.currentHealth -= counterDamage
        this.addLog(LogType.ATTACK, `[反击!] ${defender.name} 反击造成 ${counterDamage} 点伤害`)
      }
    }
  }
  
  // 运行完整战斗（最多50回合）
  runBattle() {
    for (let i = 0; i < 50; i++) {
      const result = this.executeTurn()
      if (result) {
        return result
      }
    }
    
    // 50回合后判定为平局
    this.result = BattleResult.DRAW
    this.addLog(LogType.INFO, '战斗超时，判定为平局！')
    return this.result
  }
  
  // 获取战斗日志
  getCombatLog() {
    return this.logs
  }
  
  // 获取战斗结果
  getBattleResult() {
    return {
      result: this.result,
      playerHealth: this.player.currentHealth,
      opponentHealth: this.opponent.currentHealth,
      logs: this.logs
    }
  }
}

/**
 * 主要战斗模拟函数
 * @param {Object} playerData 玩家数据
 * @param {Object} opponentData 对手数据
 * @param {boolean} isPvE 是否是PVE战斗
 * @returns {Object} 战斗结果
 */
export function simulateBattle(playerData, opponentData, isPvE = false) {
  const battle = new BattleManager(playerData, opponentData, isPvE)
  const result = battle.runBattle()
  
  // 生成奖励
  const rewards = []
  if (result === BattleResult.VICTORY) {
    if (isPvE) {
      // PVE奖励
      const baseExp = opponentData.level * 100
      const baseStones = opponentData.level * 10
      
      rewards.push(`修为 ${baseExp}`)
      rewards.push(`灵石 ${baseStones}`)
      
      // 随机掉落
      if (Math.random() < 0.3) {
        rewards.push(`强化石 ${Math.floor(Math.random() * 3) + 1}`)
      }
      if (Math.random() < 0.1) {
        rewards.push(`洗炼石 1`)
      }
    } else {
      // PVP奖励
      const prestige = Math.floor(opponentData.level * 5)
      const stones = Math.floor(opponentData.level * 3)
      
      rewards.push(`声望 ${prestige}`)
      rewards.push(`灵石 ${stones}`)
    }
  }
  
  return {
    result: {
      win: result === BattleResult.VICTORY,
      title: result === BattleResult.VICTORY ? '战斗胜利！' : 
             result === BattleResult.DEFEAT ? '战斗失败' : '平局',
      type: result === BattleResult.VICTORY ? 'success' : 
            result === BattleResult.DEFEAT ? 'error' : 'warning',
      message: result === BattleResult.VICTORY ? '恭喜你获得胜利！' :
               result === BattleResult.DEFEAT ? '再接再厉，继续修炼！' :
               '势均力敌，难分胜负！',
      rewards,
      playerHealth: battle.player.currentHealth,
      opponentHealth: battle.opponent.currentHealth
    },
    logs: battle.getCombatLog()
  }
}