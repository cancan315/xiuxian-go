import { defineStore } from 'pinia'
import { getRealmName, getRealmLength } from '../plugins/realm'

export const useCultivationStore = defineStore('cultivation', {
  actions: {
    // 修炼增加修为
    cultivate(amount, playerInfoStore, persistenceStore) {
      // 确保amount是数字类型
      const numAmount = Number(String(amount).replace(/[^0-9.-]/g, '')) || 0
      playerInfoStore.cultivation = Number(String(playerInfoStore.cultivation).replace(/[^0-9.-]/g, '')) || 0
      playerInfoStore.cultivation += numAmount
      // 增加修炼时间统计（需要在stats模块中处理）
      // this.totalCultivationTime += 1 // 增加修炼时间统计
      if (playerInfoStore.cultivation >= playerInfoStore.maxCultivation) {
        this.tryBreakthrough(playerInfoStore, persistenceStore)
      }
    },
    
    // 尝试突破
    tryBreakthrough(playerInfoStore, persistenceStore) {
      // 境界等级对应的境界名称和修为上限
      const realmsLength = getRealmLength()
      // 检查是否可以突破到下一个境界
      if (playerInfoStore.level < realmsLength) {
        const nextRealm = getRealmName(playerInfoStore.level)
        // 更新境界信息
        playerInfoStore.level += 1
        playerInfoStore.realm = nextRealm.name // 使用完整的境界名称（如：练气一层）
        playerInfoStore.maxCultivation = nextRealm.maxCultivation
        playerInfoStore.cultivation = 0 // 重置修为值
        // 增加突破次数（需要在stats模块中处理）
        // this.breakthroughCount += 1 // 增加突破次数
        // 解锁新境界
        if (!playerInfoStore.unlockedRealms.includes(nextRealm.name)) {
          playerInfoStore.unlockedRealms.push(nextRealm.name)
        }
        // 突破奖励
        playerInfoStore.spirit += 100 * playerInfoStore.level // 获得灵力奖励
        playerInfoStore.spiritRate *= 1.2 // 提升灵力获取倍率
        return true
      }
      return false
    },
    
    // 获取灵力
    gainSpirit(amount, playerInfoStore) {
      playerInfoStore.spirit += amount * playerInfoStore.spiritRate;
      // 由于灵力获取已经移到后端处理，这里只需更新本地状态
      // 实际的数据库更新由后端定时任务完成
    },
  }
})