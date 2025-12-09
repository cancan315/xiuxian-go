import { defineStore } from 'pinia'
import { pillRecipes, tryCreatePill, calculatePillEffect } from '../plugins/pills'

export const usePillsStore = defineStore('pills', {
  state: () => ({
    // 丹药系统
    pills: [], // 丹药库存
    pillFragments: {}, // 丹方残页（key为丹方ID，value为数量）
    pillRecipes: [], // 已获得的完整丹方
    activeEffects: [], // 当前生效的丹药效果列表
    pillsCrafted: 0, // 炼制丹药次数
    pillsConsumed: 0, // 服用丹药次数
  }),
  
  actions: {
    // 获得丹方残页
    gainPillFragment(recipeId, persistenceStore) {
      if (!this.pillFragments[recipeId]) {
        this.pillFragments[recipeId] = 0
      }
      this.pillFragments[recipeId]++
      // 检查是否可以合成完整丹方
      const recipe = pillRecipes.find(r => r.id === recipeId)
      if (recipe && this.pillFragments[recipeId] >= recipe.fragmentsNeeded) {
        this.pillFragments[recipeId] -= recipe.fragmentsNeeded
        if (!this.pillRecipes.includes(recipeId)) {
          this.pillRecipes.push(recipeId)
          // 需要在statsStore中增加unlockedPillRecipes统计
          // this.unlockedPillRecipes++
        }
      }
    },
    
    // 炼制丹药
    craftPill(recipeId, herbsStore, playerInfoStore, persistenceStore) {
      const recipe = pillRecipes.find(r => r.id === recipeId)
      if (!recipe || !this.pillRecipes.includes(recipeId)) {
        return { success: false, message: '未掌握丹方' }
      }
      const fragments = this.pillFragments[recipeId] || 0
      const result = tryCreatePill(recipe, herbsStore.herbs, playerInfoStore, fragments, playerInfoStore.luck * playerInfoStore.alchemyRate)
      if (result.success) {
        // 消耗材料
        recipe.materials.forEach(material => {
          for (let i = 0; i < material.count; i++) {
            const index = herbsStore.herbs.findIndex(h => h.id === material.herb)
            if (index > -1) {
              herbsStore.herbs.splice(index, 1)
            }
          }
        })
        // 创建丹药
        const effect = calculatePillEffect(recipe, playerInfoStore.level)
        const pill = {
          id: `${recipe.id}_${Date.now()}`,
          name: recipe.name,
          description: recipe.description,
          type: 'pill',
          effect
        }
        // 需要在inventoryStore.items中添加
        // this.items.push(pill)
        this.pillsCrafted++
      }
      return result
    },
    
    // 使用丹药
    usePill(pill, inventoryStore, persistenceStore) {
      const now = Date.now()
      // 添加效果
      this.activeEffects.push({
        ...pill.effect,
        startTime: now,
        endTime: now + pill.effect.duration * 1000
      })
      // 移除已使用的丹药
      const index = inventoryStore.items.findIndex(i => i.id === pill.id)
      if (index > -1) {
        inventoryStore.items.splice(index, 1)
        this.pillsConsumed++
      }
      // 清理过期效果
      this.activeEffects = this.activeEffects.filter(effect => effect.endTime > now)
      return { success: true, message: '使用丹药成功' }
    },
    
    // 获取当前有效的丹药效果
    getActiveEffects() {
      const now = Date.now()
      return this.activeEffects.filter(effect => effect.endTime > now)
    }
  }
})