// 修复数据库中已有渡劫丹的效果值
// 将 effect.value 从错误的 0.1 改为正确的 0.05
// 使用方法: node fix-du-jie-pill-effect.js

const { Client } = require('pg');
require('dotenv').config();

async function fixDuJiePillEffect() {
  const client = new Client({
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    database: process.env.DB_NAME || 'xiuxian_db',
    user: process.env.DB_USER || 'xiuxian_user',
    password: process.env.DB_PASSWORD || 'xiuxian_password',
  });
  
  try {
    await client.connect();
    console.log('数据库连接成功');
    
    // 查询所有渡劫丹
    const selectQuery = `SELECT id, user_id, pill_id, name, effect FROM pills WHERE pill_id = 'du_jie_pill'`;
    const selectResult = await client.query(selectQuery);
    
    if (selectResult.rows.length === 0) {
      console.log('没有找到渡劫丹记录');
      process.exit(0);
    }
    
    console.log(`\n找到 ${selectResult.rows.length} 个渡劫丹记录:`);
    selectResult.rows.forEach(pill => {
      const effect = typeof pill.effect === 'string' ? JSON.parse(pill.effect) : pill.effect;
      console.log(`  ID: ${pill.id}, 玩家ID: ${pill.user_id}, 当前效果值: ${effect.value}`);
    });
    
    // 更新所有渡劫丹的效果值为 0.05
    const correctEffect = JSON.stringify({
      type: 'duJieRate',
      value: 0.05,
      duration: 0,
      successRate: 0.4  // grade3 的成功率
    });
    
    const updateQuery = `UPDATE pills SET effect = $1 WHERE pill_id = 'du_jie_pill'`;
    const updateResult = await client.query(updateQuery, [correctEffect]);
    
    console.log(`\n已更新 ${updateResult.rowCount} 个渡劫丹的效果值`);
    
    // 验证更新结果
    const verifyResult = await client.query(selectQuery);
    console.log(`\n更新后的渡劫丹效果:`);
    verifyResult.rows.forEach(pill => {
      const effect = typeof pill.effect === 'string' ? JSON.parse(pill.effect) : pill.effect;
      console.log(`  ID: ${pill.id}, 玩家ID: ${pill.user_id}, 新效果值: ${effect.value}`);
    });
    
    console.log('\n✓ 渡劫丹效果修复完成!');

  } catch (error) {
    console.error('修复过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

fixDuJiePillEffect();
