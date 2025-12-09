const { getRedisClient } = require('../utils/redis');
const User = require('../models/User');

// 给所有在线玩家增加灵力值的任务
const increaseSpiritForOnlinePlayers = async () => {
  try {
    const redis = getRedisClient();
    
    // 获取所有在线玩家ID
    const onlinePlayerIds = await redis.SMEMBERS('server:online:players');
    
    if (onlinePlayerIds.length === 0) {
      return; // 没有在线玩家，直接返回
    }
    
    // 给每个在线玩家增加1点灵力值
    for (const playerId of onlinePlayerIds) {
      try {
        // 增加玩家的灵力值
        await User.increment('spirit', { 
          by: 1, 
          where: { id: playerId } 
        });
      } catch (error) {
        console.error(`给玩家 ${playerId} 增加灵力值失败:`, error);
      }
    }
    
    console.log(`已给 ${onlinePlayerIds.length} 名在线玩家各增加1点灵力值`);
  } catch (error) {
    console.error('增加在线玩家灵力值任务失败:', error);
  }
};

module.exports = { increaseSpiritForOnlinePlayers };