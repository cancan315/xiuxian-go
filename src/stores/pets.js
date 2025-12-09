import { defineStore } from 'pinia'
import { itemQualities } from '../plugins/itemQualities'

export const usePetsStore = defineStore('pets', {
  state: () => ({
    petConfig: {
      rarityMap: itemQualities.pet
    },
    pets: [], // 灵宠库存
  }),
  
  actions: {
    // 添加灵宠到库存
    addPet(pet) {
      // 此方法已废弃，灵宠状态完全由后端管理
    },
    
    // 召回灵宠
    recallPet(playerInfoStore, persistenceStore) {
      // 此方法已废弃，灵宠状态完全由后端管理
      return { success: true, message: '召回成功' }
    },
    
    // 出战灵宠
    deployPet(pet, playerInfoStore, persistenceStore) {
      // 此方法已废弃，灵宠状态完全由后端管理
      return { success: true, message: '出战成功' }
    },
    
    // 重置灵宠属性加成
    resetPetBonuses(playerInfoStore) {
      // 此方法已废弃，灵宠属性加成完全由后端管理
    },
    
    // 应用灵宠属性加成
    applyPetBonuses(playerInfoStore) {
      // 此方法已废弃，灵宠属性加成完全由后端管理
    },
    
    // 升级灵宠
    upgradePet(pet, essenceCount, playerInfoStore, persistenceStore) {
      // 此方法已废弃，灵宠状态完全由后端管理
      return { success: true, message: '升级成功' }
    },
    
    // 升星灵宠
    evolvePet(pet, foodPet, playerInfoStore, persistenceStore) {
      // 此方法已废弃，灵宠状态完全由后端管理
      return { success: true, message: '升星成功！' };
    },
    
    // 批量放生灵宠
    batchReleasePets(rarity, playerInfoStore, persistenceStore, inventoryStore) {
      // 此方法已废弃，灵宠状态完全由后端管理
      return { success: true, message: '放生成功' };
    }
  }
})