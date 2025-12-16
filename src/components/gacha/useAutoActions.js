// 处理自动操作
import { itemQualities } from '../../plugins/itemQualities.js'

export const processAutoActions = async (results, playerInfoStore) => {
  let autoReleasedCount = 0
  let autoSoldIncome = 0
  let autoSoldCount = 0

  const equipmentQualities = itemQualities.equipment

  for (const item of results) {
    if (item.type === 'pet') {
      // 检查是否需要自动放生
      const shouldRelease =
        playerInfoStore.autoReleaseRarities?.includes('all') ||
        playerInfoStore.autoReleaseRarities?.includes(item.rarity)

      if (shouldRelease) {
        // 放生灵宠，获得经验和强化石奖励
        const reward = Math.max(1, Math.floor((item.level || 1) * (item.star + 1) * 0.5))
        playerInfoStore.exp += reward
        playerInfoStore.reinforceStones += reward
        autoReleasedCount++
      } else {
        // 添加到灵宠列表
        playerInfoStore.pets.push(item)
      }
    } else if (item.type === 'equipment') {
      // 检查是否需要自动出售
      const shouldSell =
        playerInfoStore.autoSellQualities?.includes('all') ||
        playerInfoStore.autoSellQualities?.includes(item.quality)

      if (shouldSell) {
        // 出售装备，获得强化石奖励
        const reward = Math.max(1, Math.floor((item.level || 1) * (item.qualityInfo ? Object.keys(equipmentQualities).indexOf(item.quality) + 1 : 1)))
        playerInfoStore.reinforceStones += reward
        autoSoldIncome += reward
        autoSoldCount++
      } else {
        // 添加到背包
        playerInfoStore.items.push(item)
      }
    }
  }

  return {
    autoReleasedCount,
    autoSoldIncome,
    autoSoldCount
  }
}