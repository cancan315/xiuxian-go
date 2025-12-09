import { ref } from 'vue'
import APIService from '../../services/api'

// 抽奖功能
export const useGacha = () => {
  // 执行抽奖
  const performGacha = async (token, poolType, count, useWishlist) => {
    try {
      console.log(`[Gacha] 开始执行抽奖，类型: ${poolType}, 数量: ${count}, 使用心愿单: ${useWishlist}`);
      const response = await APIService.drawGacha(token, { poolType, count, useWishlist });
      console.log(`[Gacha] 抽奖成功，获得物品数量: ${response.items?.length || 0}`);
      return response;
    } catch (error) {
      console.error('[Gacha] 抽奖失败:', error);
      throw error;
    }
  };

  // 执行自动处理
  const processAutoActions = async (token, items, autoSellQualities, autoReleaseRarities) => {
    try {
      console.log(`[Gacha] 开始执行自动处理，物品数量: ${items?.length || 0}`);
      const response = await APIService.processAutoActions(token, { 
        items, 
        autoSellQualities, 
        autoReleaseRarities 
      });
      console.log(`[Gacha] 自动处理完成，卖出物品: ${response.soldItems?.length || 0}, 放生灵宠: ${response.releasedPets?.length || 0}`);
      return response;
    } catch (error) {
      console.error('[Gacha] 自动处理失败:', error);
      throw error;
    }
  };

  return {
    performGacha,
    processAutoActions
  };
};