// 注意：这是一个Node.js脚本，用于将数据库中的所有玩家ID添加到Redis
// 需要安装相应的库
// npm install pg redis dotenv

const { Client } = require('pg');
const { createClient } = require('redis');
require('dotenv').config();

async function syncPlayersToRedis() {
  // 数据库连接配置
  const dbClient = new Client({
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    database: process.env.DB_NAME || 'xiuxian_db',
    user: process.env.DB_USER || 'xiuxian_user',
    password: process.env.DB_PASSWORD || 'xiuxian_password',
  });

  // Redis连接配置
  const redisUrl = process.env.REDIS_URL || 'redis://localhost:6379';
  const redisClient = createClient({
    url: redisUrl,
  });

  try {
    // 连接到数据库
    await dbClient.connect();
    console.log('✓ 数据库连接成功');

    // 连接到Redis
    redisClient.on('error', (err) => console.log('Redis错误:', err));
    await redisClient.connect();
    console.log('✓ Redis连接成功');

    // 从数据库获取所有玩家ID
    const query = 'SELECT id FROM users ORDER BY id';
    const result = await dbClient.query(query);
    const playerIds = result.rows.map((row) => row.id.toString());

    if (playerIds.length === 0) {
      console.log('⚠ 数据库中没有玩家');
      return;
    }

    console.log(`获取到 ${playerIds.length} 个玩家`);

    // 清空旧的Redis Set
    const key = 'spirit:auto:grow:players';
    const deletedCount = await redisClient.del(key);
    console.log(`✓ 清空旧数据 (删除了 ${deletedCount} 个键)`);

    // 将所有玩家ID添加到Redis Set中
    // 使用 SADD 命令，一次性添加所有ID
    const addedCount = await redisClient.sAdd(key, playerIds);
    console.log(`✓ 成功将 ${addedCount} 个玩家ID添加到 Redis`);

    // 验证结果
    const count = await redisClient.sCard(key);
    console.log(`验证: Redis中现在有 ${count} 个玩家ID`);

    console.log('\n✓ 所有玩家ID已成功同步到Redis的 spirit:auto:grow:players!');

  } catch (error) {
    console.error('执行过程中发生错误:', error);
    process.exit(1);
  } finally {
    await dbClient.end();
    await redisClient.quit();
  }
}

syncPlayersToRedis();