#!/usr/bin/env node

/**
 * 清除所有PvE挑战次数缓存脚本
 * 用于清除Redis中所有玩家的PvE挑战次数记录（降服妖兽 + 除魔卫道）
 * 
 * 用法: 
 *   node clear-pve-count.js              # 清除所有PvE次数
 *   node clear-pve-count.js --demon      # 仅清除除魔卫道次数
 *   node clear-pve-count.js --monster    # 仅清除降服妖兽次数
 * 
 * Redis Key格式:
 *   - 降服妖兽: pve:daily:YYYY-MM-DD:用户ID
 *   - 除魔卫道: demon-slaying:daily:YYYY-MM-DD:用户ID
 */

const redis = require('redis');

const client = redis.createClient({
  host: 'localhost',
  port: 6379,
  db: 0
});

// 解析命令行参数
const args = process.argv.slice(2);
const onlyDemon = args.includes('--demon');
const onlyMonster = args.includes('--monster');

async function clearPvECount() {
  try {
    // 连接到Redis
    await client.connect();
    console.log('[清除脚本] 已连接到Redis\n');

    let patterns = [];
    
    if (onlyDemon) {
      patterns = ['demon-slaying:daily:*'];
      console.log('[模式] 仅清除除魔卫道次数');
    } else if (onlyMonster) {
      patterns = ['pve:daily:*'];
      console.log('[模式] 仅清除降服妖兽次数');
    } else {
      patterns = ['pve:daily:*', 'demon-slaying:daily:*'];
      console.log('[模式] 清除所有PvE挑战次数（降服妖兽 + 除魔卫道）');
    }

    let totalDeleted = 0;
    let totalFound = 0;

    for (const pattern of patterns) {
      console.log(`\n[查找] 正在查找匹配模式: ${pattern}`);
      const keys = await client.keys(pattern);
      
      if (keys.length === 0) {
        console.log(`[INFO] 没有找到匹配的记录`);
      } else {
        totalFound += keys.length;
        console.log(`[找到] 共找到 ${keys.length} 条记录`);
        
        // 批量删除所有匹配的key
        for (const key of keys) {
          const result = await client.del(key);
          if (result === 1) {
            totalDeleted++;
            console.log(`  ✓ ${key}`);
          }
        }
      }
    }

    console.log('\n' + '='.repeat(60));
    console.log(`[完成] 成功删除 ${totalDeleted}/${totalFound} 条记录`);
    console.log('[提示] 所有玩家的PvE挑战次数已重置，可重新挑战');
    console.log('='.repeat(60));
  } catch (error) {
    console.error('\n[错误] 清除次数失败:', error);
    process.exit(1);
  } finally {
    await client.disconnect();
    console.log('\n[断开] 已断开Redis连接');
  }
}

clearPvECount();
