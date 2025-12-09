const express = require('express');
const router = express.Router();
const { getRedisClient } = require('../utils/redis');

// 玩家登录，标记在线
router.post('/login', async (req, res) => {
  try {
    const { playerId, ip } = req.body;
    
    if (!playerId) {
      return res.status(400).json({ error: '缺少玩家ID' });
    }
    
    const redis = getRedisClient();
    const loginTime = Date.now();
    
    // 存储玩家在线状态
    const playerKey = `player:online:${playerId}`;
    await redis.HSET(playerKey, {
      playerId,
      loginTime,
      lastHeartbeat: loginTime,
      ip: ip || '',
      status: 'online'
    });
    
    // 设置过期时间（例如24小时）
    await redis.EXPIRE(playerKey, 24 * 60 * 60);
    
    // 将玩家添加到在线玩家集合中
    await redis.SADD('server:online:players', String(playerId));
    
    console.log(`玩家 ${playerId} 已上线`);
    res.json({ message: '成功标记为在线', playerId, loginTime });
  } catch (error) {
    console.error('玩家登录标记在线失败:', error);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

// 玩家心跳更新
router.post('/heartbeat', async (req, res) => {
  try {
    const { playerId } = req.body;
    
    if (!playerId) {
      return res.status(400).json({ error: '缺少玩家ID' });
    }
    
    const redis = getRedisClient();
    const playerKey = `player:online:${playerId}`;
    
    // 检查玩家是否存在
    const exists = await redis.EXISTS(playerKey);
    if (!exists) {
      return res.status(404).json({ error: '玩家不在线' });
    }
    
    // 更新最后心跳时间
    const lastHeartbeat = Date.now();
    await redis.HSET(playerKey, 'lastHeartbeat', lastHeartbeat);
    
    // 更新过期时间
    await redis.EXPIRE(playerKey, 24 * 60 * 60);
    
    res.json({ message: '心跳更新成功', playerId, lastHeartbeat });
  } catch (error) {
    console.error('玩家心跳更新失败:', error);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

// 玩家退出，标记离线
router.post('/logout', async (req, res) => {
  try {
    const { playerId } = req.body;
    
    if (!playerId) {
      return res.status(400).json({ error: '缺少玩家ID' });
    }
    
    const redis = getRedisClient();
    const playerKey = `player:online:${playerId}`;
    
    // 检查玩家是否存在
    const exists = await redis.EXISTS(playerKey);
    if (!exists) {
      return res.status(404).json({ error: '玩家不在线' });
    }
    
    // 更新玩家状态为离线
    const logoutTime = Date.now();
    await redis.HSET(playerKey, {
      status: 'offline',
      logoutTime
    });
    
    // 从在线玩家集合中移除
    await redis.SREM('server:online:players', String(playerId));
    
    console.log(`玩家 ${playerId} 已离线`);
    res.json({ message: '成功标记为离线', playerId, logoutTime });
  } catch (error) {
    console.error('玩家退出标记离线失败:', error);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

// 获取在线玩家列表
router.get('/players', async (req, res) => {
  try {
    const redis = getRedisClient();
    
    // 获取所有在线玩家ID
    const onlinePlayerIds = await redis.SMEMBERS('server:online:players');
    
    // 获取每个在线玩家的详细信息
    const players = [];
    for (const playerId of onlinePlayerIds) {
      const playerKey = `player:online:${playerId}`;
      const playerData = await redis.HGETALL(playerKey);
      if (playerData && playerData.status === 'online') {
        players.push({
          playerId: playerData.playerId,
          loginTime: parseInt(playerData.loginTime),
          lastHeartbeat: parseInt(playerData.lastHeartbeat),
          ip: playerData.ip
        });
      }
    }
    
    res.json({ players });
  } catch (error) {
    console.error('获取在线玩家列表失败:', error);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

// 获取指定玩家的在线状态
router.get('/player/:playerId', async (req, res) => {
  try {
    const { playerId } = req.params;
    
    const redis = getRedisClient();
    const playerKey = `player:online:${playerId}`;
    
    // 获取玩家在线状态
    const playerData = await redis.HGETALL(playerKey);
    
    if (!playerData) {
      return res.status(404).json({ error: '玩家不在线' });
    }
    
    res.json({
      playerId: playerData.playerId,
      status: playerData.status,
      loginTime: parseInt(playerData.loginTime),
      lastHeartbeat: parseInt(playerData.lastHeartbeat),
      ip: playerData.ip
    });
  } catch (error) {
    console.error('获取玩家在线状态失败:', error);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

module.exports = router;