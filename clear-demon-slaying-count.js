#!/usr/bin/env node

/**
 * 清除除魔卫道次数缓存脚本
 * 用于清除Redis中所有玩家的除魔卫道每日挑战次数记录
 * 
 * 用法: node clear-demon-slaying-count.js
 * 
 * Redis Key格式: demon-slaying:daily:YYYY-MM-DD:用户ID
 * 例如: demon-slaying:daily:2026-01-02:1
 */

const redis = require('redis');

const client = redis.createClient({
  host: 'localhost',
  port: 6379,
  db: 0
});

async function clearDemonSlayingCount() {
  try {
    // 连接到Redis
    await client.connect();
    console.log('[清除脚本] 已连接到Redis');

    // 查找所有除魔卫道次数key（使用通配符）
    const pattern = 'demon-slaying:daily:*';
    console.log(`[查找] 正在查找匹配模式: ${pattern}`);

    const keys = await client.keys(pattern);
    
    if (keys.length === 0) {
      console.log('[INFO] 没有找到除魔卫道次数记录');
    } else {
      console.log(`[找到] 共找到 ${keys.length} 条除魔卫道次数记录`);
      
      // 批量删除所有匹配的key
      let deletedCount = 0;
      for (const key of keys) {
        const result = await client.del(key);
        if (result === 1) {
          deletedCount++;
          console.log(`[删除] ${key}`);
        }
      }
      
      console.log(`\n[完成] 成功删除 ${deletedCount}/${keys.length} 条记录`);
      console.log('[提示] 所有玩家的除魔卫道次数已重置，可重新挑战');
    }
  } catch (error) {
    console.error('[错误] 清除次数失败:', error);
    process.exit(1);
  } finally {
    await client.disconnect();
    console.log('[断开] 已断开Redis连接');
  }
}

clearDemonSlayingCount();
