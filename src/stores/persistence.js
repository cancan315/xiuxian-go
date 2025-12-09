import { defineStore } from 'pinia'
import { getAuthToken } from './db'
import APIService from '../services/api'
import { mapQuality } from '../plugins/itemQualities'

export const usePersistenceStore = defineStore('persistence', {
  actions: {
    // 同步灵力值到后端
    async syncSpiritToBackend(spiritValue) {
      const token = getAuthToken();
      if (token) {
        try {
          await APIService.updateSpirit(token, spiritValue);
        } catch (error) {
          console.error('同步灵力值失败:', error);
        }
      }
    },
    
    // 仅同步灵力值（用于自动获取灵力功能）
    async syncSpiritOnly() {
      // 这个方法原本用于同步灵力值，现在我们可以让它调用 syncSpiritToBackend
      // 或者根据需要执行其他与灵力相关的同步操作
      // 目前留空，可以根据实际需求添加实现
    },
  }
})