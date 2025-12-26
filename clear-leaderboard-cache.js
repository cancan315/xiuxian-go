#!/usr/bin/env node

/**
 * 清除排行榜缓存脚本（节点版本）
 * 用于清除Redis中的排行榜缓存，使新的缓存数据结构生效
 * 
 * 用法: node clear-leaderboard-cache.js
 */

const redis = require('redis');

const client = redis.createClient({
  host: 'localhost',
  port: 6379,
  db: 0
});

const keys = [
  'leaderboard:realm:top100',
  'leaderboard:spiritStones:top100',
  'leaderboard:equipment:top100',
  'leaderboard:pets:top100'
];

async function clearCache() {
  try {
    // 连接到Redis
    await client.connect();
    console.log('[清除脚本] 已连接到Redis');

    // 删除排行榜相关的所有缓存键
    for (const key of keys) {
      const result = await client.del(key);
      if (result === 1) {
        console.log(`[OK] 已删除缓存: ${key}`);
      } else {
        console.log(`[INFO] 缓存不存在: ${key}`);
      }
    }

    console.log('\n排行榜缓存已清空，下次请求时将重新生成正确格式的缓存数据');
  } catch (error) {
    console.error('[错误] 清除缓存失败:', error);
    process.exit(1);
  } finally {
    await client.disconnect();
  }
}

clearCache();
