import { defineStore } from 'pinia'
import { getAuthToken, clearAuthToken } from './db'
import APIService from '../services/api'
import router from '../router'
import { getRealmName, getRealmLength } from '../plugins/realm'
import { itemQualities } from '../plugins/itemQualities'
import { pillRecipes, tryCreatePill, calculatePillEffect } from '../plugins/pills'
// 前端统一使用 playerInfoStore 作为数据源
export const usePlayerInfoStore = defineStore('playerInfo', {
  state: () => ({
    // 基础属性
    id: 0, // 玩家ID（来自后端，用于WebSocket连接）
    name: '无名修士',
    nameChangeCount: 0, // 道号修改次数
    level: 1, // 境界等级
    realm: '练气期一层', // 当前境界名称
    cultivation: 0, // 当前修为值
    maxCultivation: 100, // 当前境界最大修为值
    spirit: 0, // 灵力值
    spiritRate: 1, // 灵力获取倍率
    spiritGainRate: 1, // 灵力获取速率
    luck: 1, // 幸运值
    cultivationRate: 1, // 修炼速率
    cultivationCost: 1, // 修炼消耗灵力
    cultivationGain: 1, // 修炼获得修为
    spiritStones: 0, // 灵石数量
    reinforceStones: 0, // 强化石数量
    refinementStones: 0, // 洗炼石数量
    petEssence: 0, // 灵兽精华
    
    herbRate: 1, // 灵草获取倍率
    alchemyRate: 1, // 炼丹成功率加成
    
    // 战斗属性
    baseAttributes: {
      attack: 10, // 攻击
      health: 100, // 生命
      defense: 5, // 防御
      speed: 10 // 速度
    },
    // 战斗属性
    combatAttributes: {
      critRate: 0, // 暴击率
      comboRate: 0, // 连击率
      counterRate: 0, // 反击率
      stunRate: 0, // 眩晕率
      dodgeRate: 0, // 闪避率
      vampireRate: 0 // 吸血率
    },
    // 战斗抗性
    combatResistance: {
      critResist: 0, // 抗暴击
      comboResist: 0, // 抗连击
      counterResist: 0, // 抗反击
      stunResist: 0, // 抗眩晕
      dodgeResist: 0, // 抗闪避
      vampireResist: 0 // 抗吸血
    },
    // 特殊属性
    specialAttributes: {
      healBoost: 0, // 强化治疗
      critDamageBoost: 0, // 强化爆伤
      critDamageReduce: 0, // 弱化爆伤
      finalDamageBoost: 0, // 最终增伤
      finalDamageReduce: 0, // 最终减伤
      combatBoost: 0, // 战斗属性提升
      resistanceBoost: 0 // 战斗抗性提升
    },
    
    // 灵宠系统
    activePet: null, // 当前出战的灵宠
    petEssence: 0, // 灵宠精华
    
    // 库存与物品系统
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
    
    // 成就与解锁项
    unlockedRealms: ['练气一层'], // 已解锁境界
    unlockedLocations: ['新手村'], // 已解锁地点
    unlockedSkills: [], // 已解锁功法
    
    // ===== 来自 combatStore =====
    // 战斗属性已在上方定义
    
    // ===== 来自 cultivationStore =====
    // 修炼相关属性已在上方定义
    
    // ===== 来自 equipmentStore =====
    equippedArtifacts: {
      faqi: null, // 法宝
      guanjin: null, // 冠巾
      daopao: null, // 道袍
      yunlv: null, // 云履
      fabao: null // 本命法宝
    },
    
    // ===== 来自 gachaStore =====
    gachaResults: [],
    showResultModal: false,
    currentPage: 1,
    pageSize: 12,
    selectedQuality: 'all',
    selectedRarity: 'all',
    isDrawing: false,
    isShaking: false,
    isOpening: false,
    showProbabilityInfo: false,
    showWishlistSettings: false,
    showAutoSettings: false,
    showPetDetails: false,
    showEquipmentDetails: false,
    selectedPet: null,
    selectedEquipment: null,
    
    // ===== 来自 petsStore =====
    petConfig: {
      rarityMap: itemQualities.pet
    },
    pets: [], // 灵宠库存
    
    // ===== 来自 pillsStore =====
    pills: [], // 丹药库存
    pillFragments: {}, // 丹方残页
    pillRecipes: [], // 已获得的完整丹方
    activeEffects: [], // 当前生效的丹药效果列表
    pillsCrafted: 0, // 炼制丹药次数
    pillsConsumed: 0, // 服用丹药次数
    
    // ===== 来自 settingsStore =====
    isDarkMode: false, // 主题设置
    
    // ===== 来自 statsStore =====
    totalCultivationTime: 0, // 总修炼时间
    breakthroughCount: 0, // 突破次数
    explorationCount: 0, // 探索次数
    itemsFound: 0, // 获得物品数量
    eventTriggered: 0, // 触发事件次数
    unlockedPillRecipes: 0, // 解锁丹方数量
    dungeonDifficulty: 'easy', // 难度选择（默认凡界）
    dungeonHighestFloor: 0, // 最高通关层数
    dungeonHighestFloor_2: 0,
    dungeonHighestFloor_5: 0,
    dungeonHighestFloor_10: 0,
    dungeonHighestFloor_100: 0,
    dungeonLastFailedFloor: 0, // 最后失败层数
    dungeonTotalRuns: 0, // 总探索次数
    dungeonBossKills: 0, // Boss击杀数
    dungeonEliteKills: 0, // 精英击杀数
    dungeonTotalKills: 0, // 总击杀数
    dungeonDeathCount: 0, // 死亡次数
    dungeonTotalRewards: 0, // 获得奖励次数
  }),
  
  actions: {
    // ===== 原有 playerInfo 方法 =====
    // 玩家重命名
    renamePlayer(newName) {
      this.name = newName
      this.nameChangeCount++
    },
    
    // 退出游戰
    async logout() {
      const token = getAuthToken()
      if (token) {
        try {
          const userData = await APIService.getUser(token);
          if (userData && userData.id) {
            await APIService.playerOffline(userData.id);
          }
        } catch (error) {
          console.error('设置玩家离线状态失败:', error)
        }
      }
      clearAuthToken()
      this.$reset()
      router.push('/login')
    },
    
    // ===== 来自 cultivationStore 方法 =====
    // 修炼剂添加修为
    cultivate(amount) {
      const numAmount = Number(String(amount).replace(/[^0-9.-]/g, '')) || 0
      this.cultivation = Number(String(this.cultivation).replace(/[^0-9.-]/g, '')) || 0
      this.cultivation += numAmount
      if (this.cultivation >= this.maxCultivation) {
        this.tryBreakthrough()
      }
    },
    
    // 尝试突破
    tryBreakthrough() {
      const realmsLength = getRealmLength()
      if (this.level < realmsLength) {
        const nextRealm = getRealmName(this.level)
        this.level += 1
        this.realm = nextRealm.name
        this.maxCultivation = nextRealm.maxCultivation
        this.cultivation = 0
        if (!this.unlockedRealms.includes(nextRealm.name)) {
          this.unlockedRealms.push(nextRealm.name)
        }
        this.spirit += 100 * this.level
        this.spiritRate *= 1.2
        this.breakthroughCount++
        return true
      }
      return false
    },
    
    // 获取灵力
    gainSpirit(amount) {
      this.spirit += amount * this.spiritRate;
    },
    
    // ===== 来自 equipmentStore 方法 =====
    // 穿上装备
    async equipArtifact(artifact, slot, token) {
      try {
        console.log(`[装备] 玩家尝试装备物品: ${artifact.name} (${artifact.id}), 槽位: ${slot}`)
        
        if (!slot) {
          console.error('[装备] 装备失败：槽位参数无效', { artifact, slot });
          return { success: false, message: '槽位参数无效', equipment: artifact };
        }
        
        const response = await APIService.equipEquipment(token, artifact.id, slot);
        
        if (response.success) {
          if (response.equipment.requiredRealm && this.level < response.equipment.requiredRealm) {
            console.log(`[装备] 装备失败，境界不足`)
            return { success: false, message: '境界不足，无法装备此装备' }
          }
          
          console.log(`[装备] 成功装备物品: ${response.equipment.name}`)
          this.equippedArtifacts[slot] = response.equipment;
          return { success: true, message: '装备成功', equipment: response.equipment, user: response.user }
        } else {
          console.log(`[装备] 装备失败: ${response.message || '\u672a\u77e5\u9519\u8bef'}`);
          return { success: false, message: response.message || '装备失败', equipment: artifact };
        }
      } catch (error) {
        console.error('[装备] 装备穿戴失败:', error);
        return { success: false, message: '装备穿戴失败: ' + error.message, equipment: artifact }
      } 
    },

    // 卸下装备
    async unequipArtifact(slot, token) {
      const artifact = this.equippedArtifacts[slot]
      if (artifact) {
        try {
          console.log(`[装备] 玩家尝试卸下装备: ${artifact.name}`)
          
          const response = await APIService.unequipEquipment(token, artifact.id);
          
          if (response.success) {
            console.log(`[装备] 成功卸下装备: ${response.equipment.name}`)
            this.equippedArtifacts[slot] = null;
            return { 
              success: true, 
              equipment: response.equipment,
              user: response.user
            }
          } else {
            console.log(`[装备] 卸下装备失败: ${response.message || '\u672a\u77e5\u9519\u8bef'}`);
            return { success: false, message: response.message || '卸下装备失败', equipment: artifact }
          }
        } catch (error) {
          console.error('[装备] 装备卸下失败:', error);
          return { success: false, message: '装备卸下失败: ' + error.message, equipment: artifact }
        }
      }
      console.log(`[装备] 尝试卸下装备失败: 指定槽位没有装备`)
      return { success: false, message: '指定槽位没有装备', equipment: null }
    },

    // 卖出装备
    async sellEquipment(equipment, token) {
      try {
        console.log(`[装备] 玩家尝试卖出装备: ${equipment.name}`)
        
        const response = await APIService.sellEquipment(token, equipment.id);
        
        if (response.success) {
          const index = this.items.findIndex(i => i.id === equipment.id)
          if (index > -1) {
            this.items.splice(index, 1)
          }
          console.log(`[装备] 成功卖出装备: ${equipment.name}`)
          return { success: true, message: `成功卖出装备，获得${response.spiritStones || 0}个灵石` }
        } else {
          console.log(`[装备] 卖出装备失败: ${response.message || '\u672a\u77e5\u9519\u8bef'}`);
          return { success: false, message: response.message || '出售失败' }
        }
      } catch (error) {
        console.error('[装备] 装备出售失败:', error);
        return { success: false, message: '装备出售失败: ' + error.message }
      }
    },

    // 批量卖出装备
    async batchSellEquipments(quality = null, equipmentType = null, token) {
      try {
        console.log(`[装备] 玩家尝试批量卖出装备`)
        
        const params = {};
        if (quality) params.quality = quality;
        if (equipmentType) params.type = equipmentType;
        
        const response = await APIService.batchSellEquipment(token, params);
        
        if (response.success) {
          const originalCount = this.items.length;
          this.items = this.items.filter(item => {
            if (!item || !item.type || item.type === 'pill' || item.type === 'pet') return true
            if (equipmentType && item.type !== equipmentType) return true
            if (quality && item.quality !== quality) return true
            return false
          })
          
          console.log(`[装备] 批量卖出装备成功`)
          return {
            success: true,
            message: `成功卖出${response.equipmentSold || 0}件装备，获得${response.stonesReceived || 0}个强化石，${response.stonesReceived || 0}个洗练石`
          }
        } else {
          console.log(`[装备] 批量卖出装备失败`)
          return { success: false, message: response.message || '批量出售失败' }
        }
      } catch (error) {
        console.error('[装备] 批量出售装备失败:', error);
        return { success: false, message: '批量出售装备失败: ' + error.message }
      }
    },
    
    // 从后端获取玩家装备数据
    async fetchPlayerEquipment(token, params = {}) {
      try {
        console.log('[装备] 从后端获取玩家装备数据');
        const response = await APIService.getEquipmentList(token, params);
        console.log('[装备] 成功获取玩家装备数据:', response);
        return response;
      } catch (error) {
        console.error('[装备] 获取玩家装备数据失败:', error);
        throw error;
      }
    },
    
    // ===== 来自 gachaStore 方法 =====
    setCurrentPage(page) {
      this.currentPage = page
    },
    
    setSelectedQuality(quality) {
      this.selectedQuality = quality
    },
    
    setSelectedRarity(rarity) {
      this.selectedRarity = rarity
    },
    
    setDrawingState(isDrawing) {
      this.isDrawing = isDrawing
    },
    
    setShakingState(isShaking) {
      this.isShaking = isShaking
    },
    
    setOpeningState(isOpening) {
      this.isOpening = isOpening
    },
    
    resetPagination() {
      this.currentPage = 1
    },
    
    setGachaResults(results) {
      this.gachaResults = results
    },
    
    toggleResultModal(show) {
      this.showResultModal = show
    },
    
    // ===== 来自 pillsStore 方法 =====
    gainPillFragment(recipeId) {
      if (!this.pillFragments[recipeId]) {
        this.pillFragments[recipeId] = 0
      }
      this.pillFragments[recipeId]++
      const recipe = pillRecipes.find(r => r.id === recipeId)
      if (recipe && this.pillFragments[recipeId] >= recipe.fragmentsNeeded) {
        this.pillFragments[recipeId] -= recipe.fragmentsNeeded
        if (!this.pillRecipes.includes(recipeId)) {
          this.pillRecipes.push(recipeId)
          this.unlockedPillRecipes++
        }
      }
    },
    
    craftPill(recipeId) {
      const recipe = pillRecipes.find(r => r.id === recipeId)
      if (!recipe || !this.pillRecipes.includes(recipeId)) {
        return { success: false, message: '未掉握丹方' }
      }
      const fragments = this.pillFragments[recipeId] || 0
      const result = tryCreatePill(recipe, this.herbs, this, fragments, this.luck * this.alchemyRate)
      if (result.success) {
        recipe.materials.forEach(material => {
          for (let i = 0; i < material.count; i++) {
            const index = this.herbs.findIndex(h => h.id === material.herb)
            if (index > -1) {
              this.herbs.splice(index, 1)
            }
          }
        })
        const effect = calculatePillEffect(recipe, this.level)
        const pill = {
          id: `${recipe.id}_${Date.now()}`,
          name: recipe.name,
          description: recipe.description,
          type: 'pill',
          effect
        }
        this.pillsCrafted++
      }
      return result
    },
    
    usePill(pill) {
      const now = Date.now()
      this.activeEffects.push({
        ...pill.effect,
        startTime: now,
        endTime: now + pill.effect.duration * 1000
      })
      const index = this.items.findIndex(i => i.id === pill.id)
      if (index > -1) {
        this.items.splice(index, 1)
        this.pillsConsumed++
      }
      this.activeEffects = this.activeEffects.filter(effect => effect.endTime > now)
      return { success: true, message: '使用丹药成功' }
    },
    
    getActiveEffects() {
      const now = Date.now()
      return this.activeEffects.filter(effect => effect.endTime > now)
    },
    
    // ===== 来自 settingsStore 方法 =====
    updateHtmlDarkMode(isDarkMode) {
      const htmlEl = document.documentElement
      if (isDarkMode) {
        htmlEl.classList.add('dark')
      } else {
        htmlEl.classList.remove('dark')
      }
    },
    
    toggleDarkMode() {
      this.isDarkMode = !this.isDarkMode
      this.updateHtmlDarkMode(this.isDarkMode)
    }
  },
  
  getters: {
    // ===== 来自 gachaStore getters =====
    currentPageResults: (state) => {
      if (!state.gachaResults) return []
      
      let filteredResults = state.gachaResults
      if (state.selectedQuality && state.selectedQuality !== 'all') {
        filteredResults = filteredResults.filter(item => item.quality === state.selectedQuality)
      }
      if (state.selectedRarity && state.selectedRarity !== 'all') {
        filteredResults = filteredResults.filter(item => item.rarity === state.selectedRarity)
      }
      
      const start = (state.currentPage - 1) * state.pageSize
      const end = start + state.pageSize
      return filteredResults.slice(start, end)
    },
    
    totalPages: (state) => {
      if (!state.gachaResults) return 0
      
      let filteredResults = state.gachaResults
      if (state.selectedQuality && state.selectedQuality !== 'all') {
        filteredResults = filteredResults.filter(item => item.quality === state.selectedQuality)
      }
      if (state.selectedRarity && state.selectedRarity !== 'all') {
        filteredResults = filteredResults.filter(item => item.rarity === state.selectedRarity)
      }
      
      return Math.ceil(filteredResults.length / state.pageSize)
    }
  }
})