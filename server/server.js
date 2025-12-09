const express = require('express');
const cors = require('cors');
require('dotenv').config();
const sequelize = require('./models/database');
const User = require('./models/User');
const { connectRedis } = require('./utils/redis');

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ limit: '10mb', extended: true }));

// Routes
app.use('/api/auth', require('./routes/authRoutes'));
app.use('/api/player', require('./routes/playerRoutes'));
app.use('/api/gacha', require('./routes/gachaRoutes'));
app.use('/api/online', require('./routes/onlineRoutes'));

// 提供一个接口用于玩家主动标记为离线（退出游戏时调用）
app.post('/api/player/offline', async (req, res) => {
  try {
    const { playerId } = req.body;
    
    if (!playerId) {
      return res.status(400).json({ error: '缺少玩家ID' });
    }
    
    const { getRedisClient } = require('./utils/redis');
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
    return res.status(200).json({ message: '玩家已标记为离线' });
  } catch (error) {
    console.error('玩家退出标记离线失败:', error);
    return res.status(500).json({ error: '内部服务器错误' });
  }
});

// 导入定时任务
const { checkOnlineStatus } = require('./tasks/checkOnlineStatus');
const { increaseSpiritForOnlinePlayers } = require('./tasks/increaseSpirit');

// Export app for testing purposes
module.exports = app;

// Database connection and server start
sequelize.sync({ force: false, alter: true }).then(async () => {
  // 连接Redis
  await connectRedis();
  
  const server = app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
  });
  
  // 启动定时任务，每5秒检查一次在线状态
  setInterval(checkOnlineStatus, 5000);
  
  // 启动定时任务，每秒给在线玩家增加灵力值
  setInterval(increaseSpiritForOnlinePlayers, 1000);
}).catch(err => {
  console.error('Unable to connect to the database:', err);
});