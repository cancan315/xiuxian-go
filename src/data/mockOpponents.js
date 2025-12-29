/**
 * 模拟对手数据
 * 在开发阶段使用，上线后应替换为真实API数据
 */

export const mockOpponents = [
  {
    id: 101,
    name: '张三丰',
    level: 3,
    realm: 3,
    cultivation: 4500,
    maxCultivation: 10000,
    spiritStones: 1500,
    baseAttributes: {
      health: 300,
      attack: 45,
      defense: 25,
      speed: 20
    },
    combatAttributes: {
      critRate: 0.15,
      comboRate: 0.05,
      counterRate: 0.1,
      stunRate: 0.08,
      dodgeRate: 0.12,
      vampireRate: 0
    }
  },
  {
    id: 102,
    name: '李逍遥',
    level: 4,
    realm: 4,
    cultivation: 7800,
    maxCultivation: 15000,
    spiritStones: 2300,
    baseAttributes: {
      health: 350,
      attack: 55,
      defense: 30,
      speed: 25
    },
    combatAttributes: {
      critRate: 0.18,
      comboRate: 0.08,
      counterRate: 0.12,
      stunRate: 0.1,
      dodgeRate: 0.15,
      vampireRate: 0.05
    }
  },
  {
    id: 103,
    name: '赵灵儿',
    level: 2,
    realm: 2,
    cultivation: 2000,
    maxCultivation: 5000,
    spiritStones: 800,
    baseAttributes: {
      health: 250,
      attack: 35,
      defense: 20,
      speed: 18
    },
    combatAttributes: {
      critRate: 0.12,
      comboRate: 0.03,
      counterRate: 0.08,
      stunRate: 0.05,
      dodgeRate: 0.1,
      vampireRate: 0
    }
  },
  {
    id: 104,
    name: '王重阳',
    level: 5,
    realm: 5,
    cultivation: 12000,
    maxCultivation: 20000,
    spiritStones: 3500,
    baseAttributes: {
      health: 400,
      attack: 65,
      defense: 35,
      speed: 22
    },
    combatAttributes: {
      critRate: 0.2,
      comboRate: 0.1,
      counterRate: 0.15,
      stunRate: 0.12,
      dodgeRate: 0.1,
      vampireRate: 0.08
    }
  },
  {
    id: 105,
    name: '周芷若',
    level: 3,
    realm: 3,
    cultivation: 5200,
    maxCultivation: 10000,
    spiritStones: 1800,
    baseAttributes: {
      health: 280,
      attack: 48,
      defense: 22,
      speed: 26
    },
    combatAttributes: {
      critRate: 0.16,
      comboRate: 0.06,
      counterRate: 0.09,
      stunRate: 0.1,
      dodgeRate: 0.18,
      vampireRate: 0.03
    }
  }
]

// 根据玩家等级获取合适的对手
export function getSuitableOpponents(playerLevel) {
  const minLevel = Math.max(1, playerLevel - 2)
  const maxLevel = playerLevel + 2
  
  return mockOpponents.filter(opponent => 
    opponent.level >= minLevel && opponent.level <= maxLevel
  )
}

// 获取随机对手
export function getRandomOpponent(playerLevel) {
  const suitable = getSuitableOpponents(playerLevel)
  if (suitable.length === 0) {
    return mockOpponents[Math.floor(Math.random() * mockOpponents.length)]
  }
  return suitable[Math.floor(Math.random() * suitable.length)]
}