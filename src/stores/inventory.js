import { defineStore } from 'pinia'
import { v4 as uuidv4 } from 'uuid'

export const useInventoryStore = defineStore('inventory', {
  state: () => ({
    // 资源
    spiritStones: 0, // 灵石数量
    reinforceStones: 0, // 强化石数量
    refinementStones: 0, // 洗练石数量
    herbs: [], // 灵草库存
    items: [], // 物品库存 (不包括灵宠)
    artifacts: [], // 法宝装备
    
    // 自动出售相关设置
    autoSellQualities: [], // 选中的装备品质
    autoReleaseRarities: [], // 选中的灵宠品质
    // 心愿单相关设置
    wishlistEnabled: false, // 心愿单开关
    selectedWishEquipQuality: null,
    selectedWishPetRarity: null,
  }),
  
  actions: {
    // 获得物品
    gainItem(item, persistenceStore) {
      this.items.push({
        ...item,
        // 确保物品有itemId以便能正确保存到数据库
        itemId: item.itemId || item.id || uuidv4(),
        id: uuidv4()
      });
      // 需要在statsStore中增加itemsFound统计
      // this.itemsFound++ // 增加获得物品统计
    },
    
    // 使用物品（丹药或灵宠）
    useItem(item, pillsStore, petsStore, playerInfoStore, persistenceStore) {
      if (item.type === 'pill') {
        return pillsStore.usePill(item, this, persistenceStore);
      } else if (item.type === 'pet') {
        return petsStore.usePet(item, playerInfoStore, persistenceStore);
      }
      return { success: false, message: '无法使用该物品' }
    },
    
    // 添加装备到背包
    addEquipment(equipment, persistenceStore) {
      if (!this.items) {
        this.items = []
      }
      this.items.push(equipment)
    }
  }
})