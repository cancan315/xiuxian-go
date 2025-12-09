import { defineStore } from 'pinia'
import { getAuthToken, clearAuthToken } from './db'
import APIService from '../services/api'
import router from '../router'
// 前端统一使用 playerInfoStore 作为数据源
export const usePlayerInfoStore = defineStore('playerInfo', {
  state: () => ({
    // 基础属性
    name: '无名修士',
    nameChangeCount: 0, // 道号修改次数
    level: 1, // 境界等级
    realm: '练气期一层', // 当前境界名称
    cultivation: 0, // 当前修为值
    maxCultivation: 100, // 当前境界最大修为值
    spirit: 0, // 灵力值
    spiritRate: 1, // 灵力获取倍率
    luck: 1, // 幸运值
    cultivationRate: 1, // 修炼速率
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
    
    // 是否新玩家
    isNewPlayer: true,
    
    // 成就与解锁项
    unlockedRealms: ['练气一层'], // 已解锁境界
    unlockedLocations: ['新手村'], // 已解锁地点
    unlockedSkills: [], // 已解锁功法
  }),
  
  actions: {
    // 玩家重命名
    renamePlayer(newName) {
      this.name = newName
      this.nameChangeCount++
    },
    
    // 退出游戏
    async logout() {

      
      // 通知服务器玩家已离线
      const token = getAuthToken()
      if (token) {
        try {
          await APIService.setPlayerOffline(token)
          
          // Also mark player as offline in Redis
          try {
            const userData = await APIService.getUser(token);
            if (userData && userData.id) {
              await APIService.playerOffline(userData.id);
            }
          } catch (error) {
            console.error('设置玩家Redis离线状态失败:', error);
          }
        } catch (error) {
          console.error('设置玩家离线状态失败:', error)
        }
      }
      
      // 清除认证令牌
      clearAuthToken()
      
      // 重置玩家状态
      this.$reset()
      
      // 跳转到登录页面
      router.push('/login')
    }
  }
})