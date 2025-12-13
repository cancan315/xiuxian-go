import { defineStore } from 'pinia'

export const useGachaStore = defineStore('gacha', {
  state: () => ({
    // 心愿单设置
    wishlistEnabled: false,
    selectedWishEquipQuality: '',
    selectedWishPetRarity: '',
    
    // 自动处理设置
    autoSellQualities: [],
    autoReleaseRarities: [],
    
    // 抽卡结果相关
    gachaResults: [],
    showResultModal: false,
    
    // 当前页码和分页设置
    currentPage: 1,
    pageSize: 12,
    
    // 筛选设置
    selectedQuality: 'all',
    selectedRarity: 'all',
    
    // 动画状态
    isDrawing: false,
    isShaking: false,
    isOpening: false,
    
    // 弹窗显示状态
    showProbabilityInfo: false,
    showWishlistSettings: false,
    showAutoSettings: false,
    showPetDetails: false,
    showEquipmentDetails: false,
    
    // 详情查看
    selectedPet: null,
    selectedEquipment: null
  }),
  
  actions: {
    // 切换心愿单启用状态
    toggleWishlist(enabled) {
      this.wishlistEnabled = enabled
    },
    
    // 设置心愿装备品质
    setSelectedWishEquipQuality(quality) {
      this.selectedWishEquipQuality = quality
    },
    
    // 设置心愿灵宠品质
    setSelectedWishPetRarity(rarity) {
      this.selectedWishPetRarity = rarity
    },
    
    // 更新自动出售品质设置
    setAutoSellQualities(qualities) {
      this.autoSellQualities = qualities
    },
    
    // 更新自动放生品质设置
    setAutoReleaseRarities(rarities) {
      this.autoReleaseRarities = rarities
    },
    
    // 设置抽卡结果
    setGachaResults(results) {
      this.gachaResults = results
    },
    
    // 切换结果弹窗显示状态
    toggleResultModal(show) {
      this.showResultModal = show
    },
    
    // 设置当前页码
    setCurrentPage(page) {
      this.currentPage = page
    },
    
    // 设置筛选品质
    setSelectedQuality(quality) {
      this.selectedQuality = quality
    },
    
    // 设置筛选稀有度
    setSelectedRarity(rarity) {
      this.selectedRarity = rarity
    },
    
    // 设置抽卡动画状态
    setDrawingState(isDrawing) {
      this.isDrawing = isDrawing
    },
    
    // 设置摇晃动画状态
    setShakingState(isShaking) {
      this.isShaking = isShaking
    },
    
    // 设置开启动画状态
    setOpeningState(isOpening) {
      this.isOpening = isOpening
    },
    
    // 重置分页
    resetPagination() {
      this.currentPage = 1
    }
  },
  
  getters: {
    // 获取当前页结果
    currentPageResults: (state) => {
      if (!state.gachaResults) return []
      
      // 添加筛选逻辑
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
    
    // 获取总页数（需要考虑筛选后的结果）
    totalPages: (state) => {
      if (!state.gachaResults) return 0
      
      // 添加筛选逻辑
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