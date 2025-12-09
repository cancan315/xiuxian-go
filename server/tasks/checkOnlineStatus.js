const { getRedisClient } = require('../utils/redis');

// 定时检查玩家在线状态的任务
const checkOnlineStatus = async () => {
  try {
    const redis = getRedisClient();
    
    // 获取所有在线玩家ID
    const onlinePlayerIds = await redis.SMEMBERS('server:online:players');
    
    // 检查每个玩家的最后心跳时间
    const currentTime = Date.now();
    const timeoutThreshold = 10 * 1000; // 10秒超时
    
    for (const playerId of onlinePlayerIds) {
      const playerKey = `player:online:${playerId}`;
      const playerData = await redis.HGETALL(playerKey);
      
      if (playerData && playerData.lastHeartbeat) {
        const lastHeartbeat = parseInt(playerData.lastHeartbeat);
        
        // 如果超过10秒没有心跳，则标记为离线
        if (currentTime - lastHeartbeat > timeoutThreshold) {
          // 更新玩家状态为离线
          await redis.HSET(playerKey, 'status', 'offline');
          
          // 从在线玩家集合中移除
          await redis.SREM('server:online:players', playerId);
          
          console.log(`玩家 ${playerId} 因超时被标记为离线`);
        }
      }
    }
  } catch (error) {
    console.error('检查在线状态任务失败:', error);
  }
};

module.exports = { checkOnlineStatus };