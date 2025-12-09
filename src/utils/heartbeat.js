import APIService from '../services/api';
import { getAuthToken } from '../stores/db';

let heartbeatInterval = null;

// 启动心跳机制
export const startHeartbeat = async (playerId) => {
  // 先清除现有的心跳定时器（如果有的话）
  stopHeartbeat();
  
  // 启动新的心跳定时器，每5秒发送一次心跳
  heartbeatInterval = setInterval(async () => {
    try {
      await APIService.updateHeartbeat(playerId);
    } catch (error) {
      console.error('心跳更新失败:', error);
    }
  }, 5000); // 每5秒发送一次心跳
};

// 停止心跳机制
export const stopHeartbeat = () => {
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval);
    heartbeatInterval = null;
  }
};

// 页面卸载时主动离线
export const handleBeforeUnload = async (playerId) => {
  try {
    await APIService.playerOffline(playerId);
  } catch (error) {
    console.error('主动离线失败:', error);
  }
};