import { defineStore } from 'pinia'
import APIService from '../services/api.js'

export const useEquipmentStore = defineStore('equipment', {
  state: () => ({
    // 装备栏位（仅用于UI展示，实际状态由后端管理）
    equippedArtifacts: {
      faqi: null, // 法宝
      guanjin: null, // 冠巾
      daopao: null, // 道袍
      yunlv: null, // 云履
      fabao: null // 本命法宝
    }
  }),

  actions: {
    // 添加装备到库存
    addEquipment(equipment) {
      // 此方法已废弃，装备状态完全由后端管理
    },

    // 应用装备属性加成（已废弃，装备属性加成由后端计算）
    applyArtifactBonuses(artifactStats, operation = 'add') {
      // 此方法已废弃，装备属性加成完全由后端管理
    },

    // 穿上装备
    async equipArtifact(artifact, slot, inventoryStore, persistenceStore, playerInfoStore, token) {
      try {
        // 添加日志记录
        console.log(`[Equipment Store] 玩家尝试装备物品: ${artifact.name} (${artifact.id}), 槽位: ${slot}`)
        console.log(`[Equipment Store] 接收到的认证令牌:`, { 
          tokenAvailable: !!token, 
          tokenLength: token ? token.length : 0 
        })
        
        // 确保传递的槽位参数有效
        if (!slot) {
          console.error('[Equipment Store] 装备失败：槽位参数无效', { artifact, slot });
          return { success: false, message: '槽位参数无效', equipment: artifact };
        }
        
        // 调用后端API穿戴装备
        const response = await APIService.equipEquipment(token, artifact.id, slot);
        
        if (response.success) {
          // 检查境界要求
          if (response.equipment.requiredRealm && playerInfoStore.level < response.equipment.requiredRealm) {
            console.log(`[Equipment Store] 装备失败，境界不足: 玩家境界=${playerInfoStore.level}, 装备要求=${response.equipment.requiredRealm}`)
            return { success: false, message: '境界不足，无法装备此装备' }
          }
          
          console.log(`[Equipment Store] 成功装备物品: ${response.equipment.name} (${response.equipment.id})`)
          return { success: true, message: '装备成功', equipment: response.equipment, user: response.user }
        } else {
          console.log(`[Equipment Store] 装备失败: ${response.message || '未知错误'}`);
          // 添加更详细的日志记录
          console.log(`[Equipment Store] 尝试装备的物品详情:`, artifact);
          
          // 如果是同类型装备已装备的错误，尝试获取已装备的同类型装备信息
          if (response.message === '同类型装备已装备，无法重复装备') {
            console.log(`[Equipment Store] 当前已装备的同类型装备:`, this.equippedArtifacts[slot]);
          }
          
          return { success: false, message: response.message || '装备失败', equipment: artifact };
        }
      } catch (error) {
        console.error('[Equipment Store] 装备穿戴失败:', error);
        return { success: false, message: '装备穿戴失败: ' + error.message, equipment: artifact }
      }
    },

    // 卸下装备
    async unequipArtifact(slot, inventoryStore, persistenceStore, token) {
      const artifact = this.equippedArtifacts[slot]
      if (artifact) {
        try {
          // 添加日志记录
          console.log(`[Equipment Store] 玩家尝试卸下装备: ${artifact.name} (${artifact.id}) 从槽位: ${slot}`)
          
          // 调用后端API卸下装备
          const response = await APIService.unequipEquipment(token, artifact.id);
          
          if (response.success) {
            console.log(`[Equipment Store] 成功卸下装备: ${response.equipment.name} (${response.equipment.id})`)
            // 返回装备和用户属性（如果有的话）
            return { 
              success: true, 
              equipment: response.equipment,
              user: response.user
            }
          } else {
            console.log(`[Equipment Store] 卸下装备失败: ${response.message || '未知错误'}`)
            return { success: false, message: response.message || '卸下装备失败', equipment: artifact }
          }
        } catch (error) {
          console.error('[Equipment Store] 装备卸下失败:', error);
          return { success: false, message: '装备卸下失败: ' + error.message, equipment: artifact }
        }
      }
      console.log(`[Equipment Store] 尝试卸下装备失败: 指定槽位(${slot})没有装备`)
      return { success: false, message: '指定槽位没有装备', equipment: null }
    },

    // 卖出装备
    async sellEquipment(equipment, inventoryStore, persistenceStore, token) {
      try {
        // 添加日志记录
        console.log(`[Equipment Store] 玩家尝试卖出装备: ${equipment.name} (${equipment.id})`)
        
        const response = await APIService.sellEquipment(token, equipment.id);
        
        if (response.success) {
          const index = inventoryStore.items.findIndex(i => i.id === equipment.id)
          if (index > -1) {
            inventoryStore.items.splice(index, 1)
          }
          
          // 更新玩家的强化石数量（如果返回了stonesReceived）
          if (response.stonesReceived !== undefined) {
            // 注意：这里需要更新玩家的强化石数量，可能需要通过playerInfoStore或者其他方式
            // playerInfoStore.reinforceStones += response.stonesReceived;
          }
          
          console.log(`[Equipment Store] 成功卖出装备: ${equipment.name} (${equipment.id}), 获得强化石: ${response.stonesReceived || 0}`)
          return { success: true, message: `成功卖出装备，获得${response.stonesReceived || 0}个强化石` }
        } else {
          console.log(`[Equipment Store] 卖出装备失败: ${response.message || '未知错误'}`)
          return { success: false, message: response.message || '出售失败' }
        }
      } catch (error) {
        console.error('[Equipment Store] 装备出售失败:', error);
        return { success: false, message: '装备出售失败: ' + error.message }
      }
    },

    // 批量卖出装备
    async batchSellEquipments(quality = null, equipmentType = null, inventoryStore, persistenceStore, token) {
      try {
        // 添加日志记录
        console.log(`[Equipment Store] 玩家尝试批量卖出装备: 品质=${quality || '全部'}, 类型=${equipmentType || '全部'}`)
        
        const params = {};
        if (quality) params.quality = quality;
        if (equipmentType) params.type = equipmentType;
        
        const response = await APIService.batchSellEquipment(token, params);
        
        if (response.success) {
          // 从背包中移除已出售的装备
          const originalCount = inventoryStore.items.length;
          inventoryStore.items = inventoryStore.items.filter(item => {
            if (!item || !item.type || item.type === 'pill' || item.type === 'pet') return true
            if (equipmentType && item.type !== equipmentType) return true
            if (quality && item.quality !== quality) return true
            return false
          })
          
          const removedCount = originalCount - inventoryStore.items.length;
          
          // 更新玩家的强化石数量（如果返回了stonesReceived）
          if (response.stonesReceived !== undefined) {
            // 注意：这里需要更新玩家的强化石数量，可能需要通过playerInfoStore或者其他方式
            // playerInfoStore.reinforceStones += response.stonesReceived;
          }
          
          console.log(`[Equipment Store] 成功批量卖出装备: 数量=${response.equipmentSold || 0}, 移除数量=${removedCount}, 获得强化石=${response.stonesReceived || 0}`)
          return {
            success: true,
            message: `成功卖出${response.equipmentSold || 0}件装备，获得${response.stonesReceived || 0}个强化石`
          }
        } else {
          console.log(`[Equipment Store] 批量卖出装备失败: ${response.message || '未知错误'}`)
          return { success: false, message: response.message || '批量出售失败' }
        }
      } catch (error) {
        console.error('[Equipment Store] 批量出售装备失败:', error);
        return { success: false, message: '批量出售装备失败: ' + error.message }
      }
    }
  }
})