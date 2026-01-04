// 注意：这是一个Node.js脚本，用于直接操作数据库
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function updatePlayerHerbs() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 2) {
    console.log('使用方法: node update-player-herbs.js <playerId> <herbCountToAdd>');
    console.log('示例: node update-player-herbs.js 1 100');
    console.log('说明: 增加指定玩家所有灵草的数量');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  const herbCountToAdd = parseInt(args[1]);
  
  // 验证参数
  if (isNaN(playerId) || playerId <= 0) {
    console.error('错误: playerId 必须是正整数');
    process.exit(1);
  }
  
  if (isNaN(herbCountToAdd) || herbCountToAdd <= 0) {
    console.error('错误: herbCountToAdd 必须是正整数');
    process.exit(1);
  }
  
  // 数据库连接配置
  const client = new Client({
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    database: process.env.DB_NAME || 'xiuxian_db',
    user: process.env.DB_USER || 'xiuxian_user',
    password: process.env.DB_PASSWORD || 'xiuxian_password',
  });
  
  try {
    // 连接到数据库
    await client.connect();
    console.log('数据库连接成功');
    
    // 查找玩家
    const userQuery = 'SELECT * FROM users WHERE id = $1';
    const userResult = await client.query(userQuery, [playerId]);
    
    if (userResult.rows.length === 0) {
      console.error(`错误: 未找到ID为 ${playerId} 的玩家`);
      process.exit(1);
    }
    
    const user = userResult.rows[0];
    console.log(`\n玩家信息:`);
    console.log(`  玩家ID: ${user.id}`);
    console.log(`  用户名: ${user.username}`);
    console.log(`  玩家名称: ${user.player_name || '未设置'}`);
    
    // 查询玩家当前灵草
    const selectQuery = 'SELECT * FROM herbs WHERE user_id = $1 ORDER BY herb_id';
    const selectResult = await client.query(selectQuery, [playerId]);
    
    if (selectResult.rows.length === 0) {
      console.log(`\n该玩家没有任何灵草记录`);
      process.exit(0);
    }
    
    // 显示更新前的灵草信息
    console.log(`\n更新前的灵草列表 (共 ${selectResult.rows.length} 种):`);
    selectResult.rows.forEach(herb => {
      console.log(`  ${herb.name} (${herb.herb_id}): ${herb.count} 个 [${herb.quality || '普通'}]`);
    });
    
    // 执行更新 - 增加所有灵草数量
    const updateQuery = 'UPDATE herbs SET count = count + $2 WHERE user_id = $1';
    const updateResult = await client.query(updateQuery, [playerId, herbCountToAdd]);
    
    console.log(`\n已更新 ${updateResult.rowCount} 条灵草记录`);
    
    // 获取更新后的灵草信息
    const updatedResult = await client.query(selectQuery, [playerId]);
    
    // 显示更新后的灵草信息
    console.log(`\n更新后的灵草列表:`);
    updatedResult.rows.forEach(herb => {
      console.log(`  ${herb.name} (${herb.herb_id}): ${herb.count} 个 [${herb.quality || '普通'}]`);
    });
    
    console.log(`\n每种灵草增加: +${herbCountToAdd}`);
    console.log('\n✓ 玩家灵草更新成功!');

  } catch (error) {
    console.error('更新过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

updatePlayerHerbs();
