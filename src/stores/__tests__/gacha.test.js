import { setActivePinia, createPinia } from 'pinia'
import { describe, it, expect, beforeEach } from 'vitest'
import { useGachaStore } from '../gacha'

describe('Gacha Store', () => {
  beforeEach(() => {
    // 创建新的 pinia 实例
    setActivePinia(createPinia())
  })

  it('initializes with correct default state', () => {
    const store = useGachaStore()
    
    // 检查默认状态
    expect(store.wishlistEnabled).toBe(false)
    expect(store.selectedWishEquipQuality).toBe('')
    expect(store.selectedWishPetRarity).toBe('')
    expect(store.autoSellQualities).toEqual([])
    expect(store.autoReleaseRarities).toEqual([])
    expect(store.gachaResults).toEqual([])
    expect(store.showResultModal).toBe(false)
    expect(store.currentPage).toBe(1)
    expect(store.pageSize).toBe(12)
    expect(store.isDrawing).toBe(false)
    expect(store.isShaking).toBe(false)
    expect(store.isOpening).toBe(false)
  })

  it('toggles wishlist correctly', () => {
    const store = useGachaStore()
    
    store.toggleWishlist(true)
    expect(store.wishlistEnabled).toBe(true)
    
    store.toggleWishlist(false)
    expect(store.wishlistEnabled).toBe(false)
  })

  it('sets wish qualities correctly', () => {
    const store = useGachaStore()
    
    store.setSelectedWishEquipQuality('epic')
    expect(store.selectedWishEquipQuality).toBe('epic')
    
    store.setSelectedWishPetRarity('legendary')
    expect(store.selectedWishPetRarity).toBe('legendary')
  })

  it('handles gacha results correctly', () => {
    const store = useGachaStore()
    const mockResults = [{ id: 1, name: 'Test Item' }]
    
    store.setGachaResults(mockResults)
    expect(store.gachaResults).toEqual(mockResults)
  })

  it('computes currentPageResults correctly', () => {
    const store = useGachaStore()
    
    // 创建20个测试结果
    const mockResults = Array.from({ length: 20 }, (_, i) => ({ 
      id: i + 1, 
      name: `Item ${i + 1}` 
    }))
    
    store.setGachaResults(mockResults)
    
    // 检查第一页结果
    expect(store.currentPageResults.length).toBe(12)
    expect(store.currentPageResults[0].id).toBe(1)
    expect(store.currentPageResults[11].id).toBe(12)
    
    // 切换到第二页
    store.setCurrentPage(2)
    expect(store.currentPageResults.length).toBe(8)
    expect(store.currentPageResults[0].id).toBe(13)
  })

  it('computes totalPages correctly', () => {
    const store = useGachaStore()
    
    // 创建20个测试结果
    const mockResults = Array.from({ length: 20 }, (_, i) => ({ 
      id: i + 1, 
      name: `Item ${i + 1}` 
    }))
    
    store.setGachaResults(mockResults)
    expect(store.totalPages).toBe(2) // 20个项目，每页12个 = 2页
    
    // 更改页面大小
    store.pageSize = 10
    expect(store.totalPages).toBe(2) // 20个项目，每页10个 = 2页
    
    store.pageSize = 5
    expect(store.totalPages).toBe(4) // 20个项目，每页5个 = 4页
  })
})