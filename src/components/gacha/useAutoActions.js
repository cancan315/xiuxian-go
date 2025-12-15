// 处理自动操作
import { itemQualities } from '../../plugins/itemQualities.js'

export const processAutoActions = async (results, inventoryStore, petsStore, playerInfoStore) => {
  let autoReleasedCount = 0
  let autoSoldIncome = 0
  let autoSoldCount = 0

  const equipmentQualities = itemQualities.equipment

  for (const item of results) {
    if (item.type === 'pet') {
      // 检查是否需要自动放生
      const shouldRelease =
        inventoryStore.autoReleaseRarities?.includes('all') ||
        inventoryStore.autoReleaseRarities?.includes(item.rarity)

      if (shouldRelease) {
        // 放生灵宠，获得经验和强化石奖励
        const reward = Math.max(1, Math.floor((item.level || 1) * (item.star + 1) * 0.5))
        playerInfoStore.exp += reward
        inventoryStore.reinforceStones += reward
        autoReleasedCount++
      } else {
        // 添加到灵宠列表
        petsStore.addPet(item)
      }
    } else if (item.type === 'equipment') {
      // 检查是否需要自动出售
      const shouldSell =
        inventoryStore.autoSellQualities?.includes('all') ||
        inventoryStore.autoSellQualities?.includes(item.quality)

      if (shouldSell) {
        // 出售装备，获得强化石奖励
        const reward = Math.max(1, Math.floor((item.level || 1) * (item.qualityInfo ? Object.keys(equipmentQualities).indexOf(item.quality) + 1 : 1)))
        inventoryStore.reinforceStones += reward
        autoSoldIncome += reward
        autoSoldCount++
      } else {
        // 添加到背包
        inventoryStore.addEquipment(item)
      }
    }
  }

  return {
    autoReleasedCount,
    autoSoldIncome,
    autoSoldCount
  }
}