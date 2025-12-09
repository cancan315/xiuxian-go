import { defineStore } from 'pinia'

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    // 主题设置
    isDarkMode: false,
  }),
  
  actions: {
    // 更新HTML暗黑模式类
    updateHtmlDarkMode(isDarkMode) {
      const htmlEl = document.documentElement
      if (isDarkMode) {
        htmlEl.classList.add('dark')
      } else {
        htmlEl.classList.remove('dark')
      }
    },
    
    // 切换暗黑模式
    toggle(persistenceStore) {
      this.isDarkMode = !this.isDarkMode
      // 更新html标签的class
      this.updateHtmlDarkMode(this.isDarkMode)

    }
  }
})