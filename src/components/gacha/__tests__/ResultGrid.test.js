import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import ResultGrid from '../ResultGrid.vue'

describe('ResultGrid', () => {
  const mockPetRarities = {
    common: { name: '凡兽', color: '#9e9e9e' },
    rare: { name: '灵兽', color: '#2196f3' }
  }

  const mockEquipmentTypesMap = {
    faqi: { name: '法宝' },
    guanjin: { name: '冠巾' }
  }

  const mockResults = [
    {
      id: '1',
      name: '测试装备',
      type: 'equipment',
      quality: 'uncommon',
      qualityInfo: { name: '法器', color: '#4caf50' },
      equipType: 'weapon'
    },
    {
      id: '2',
      name: '测试灵宠',
      type: 'pet',
      rarity: 'rare',
      description: '一只测试灵宠'
    }
  ]

  it('renders result items correctly', () => {
    const wrapper = mount(ResultGrid, {
      props: {
        currentPageResults: mockResults,
        wishlistEnabled: false,
        selectedWishEquipQuality: '',
        selectedWishPetRarity: '',
        petRarities: mockPetRarities,
        equipmentTypesMap: mockEquipmentTypesMap
      }
    })
    
    // 应该渲染2个结果项
    const items = wrapper.findAll('.result-item')
    expect(items.length).toBe(2)
    
    // 检查第一个装备项
    expect(items[0].text()).toContain('测试装备')
    expect(items[0].text()).toContain('法器')
    
    // 检查第二个灵宠项
    expect(items[1].text()).toContain('测试灵宠')
    expect(items[1].text()).toContain('灵兽')
    expect(items[1].text()).toContain('一只测试灵宠')
  })

  it('applies wish-bonus class when wishlist is enabled and matches', () => {
    const wrapper = mount(ResultGrid, {
      props: {
        currentPageResults: [mockResults[0]], // 只测试装备
        wishlistEnabled: true,
        selectedWishEquipQuality: 'uncommon', // 与装备质量匹配
        selectedWishPetRarity: '',
        petRarities: mockPetRarities,
        equipmentTypesMap: mockEquipmentTypesMap
      }
    })
    
    // 第一个装备应该有 wish-bonus 类
    const items = wrapper.findAll('.result-item')
    expect(items[0].classes()).toContain('wish-bonus')
  })
})