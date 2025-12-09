import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import GachaCard from '../GachaCard.vue'

describe('GachaCard', () => {
  it('renders correctly with default props', () => {
    const wrapper = mount(GachaCard, {
      props: {
        gachaType: 'equipment'
      }
    })
    
    expect(wrapper.find('.gacha-item').exists()).toBe(true)
    // 由于我们没有模拟 NIcon 组件，所以我们检查文本内容
    expect(wrapper.text()).toContain('🥋')
  })

  it('renders pet emoji when gachaType is pet', () => {
    const wrapper = mount(GachaCard, {
      props: {
        gachaType: 'pet'
      }
    })
    
    expect(wrapper.text()).toContain('🥚')
  })

  it('applies shake class when isShaking is true', () => {
    const wrapper = mount(GachaCard, {
      props: {
        gachaType: 'equipment',
        isShaking: true
      }
    })
    
    expect(wrapper.find('.shake').exists()).toBe(true)
  })

  it('applies open class when isOpening is true', () => {
    const wrapper = mount(GachaCard, {
      props: {
        gachaType: 'equipment',
        isOpening: true
      }
    })
    
    expect(wrapper.find('.open').exists()).toBe(true)
  })
})