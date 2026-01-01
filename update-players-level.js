// 注意：这是一个Node.js脚本，用于直接操作数据库
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function updatePlayerLevel() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 2) {
    console.log('使用方法: node update-players-level.js <playerId> <level> [cultivation] [jieYingRate]');
    console.log('示例: node update-players-level.js 1 10');
    console.log('       node update-players-level.js 1 10 500');
    console.log('       node update-players-level.js 1 10 500 0.15');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  const level = parseInt(args[1]);
  const cultivation = args.length > 2 ? parseFloat(args[2]) : null;
  const jieYingRate = args.length > 3 ? parseFloat(args[3]) : null;
  
  // 验证参数
  if (isNaN(playerId) || playerId <= 0) {
    console.error('错误: playerId 必须是正整数');
    process.exit(1);
  }
  
  if (isNaN(level) || level <= 0) {
    console.error('错误: level 必须是正整数');
    process.exit(1);
  }
  
  if (cultivation !== null && isNaN(cultivation)) {
    console.error('错误: cultivation 必须是有效的数字');
    process.exit(1);
  }
  
  if (jieYingRate !== null && (isNaN(jieYingRate) || jieYingRate < 0 || jieYingRate > 1)) {
    console.error('错误: jieYingRate 必须是 0-1 之间的有效数字');
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
    const selectQuery = 'SELECT * FROM users WHERE id = $1';
    const selectResult = await client.query(selectQuery, [playerId]);
    
    if (selectResult.rows.length === 0) {
      console.error(`错误: 未找到ID为 ${playerId} 的玩家`);
      process.exit(1);
    }
    
    const user = selectResult.rows[0];
    
    // 显示更新前的信息
    console.log('更新前的玩家信息:');
    console.log(`  玩家ID: ${user.id}`);
    console.log(`  用户名: ${user.username}`);
    console.log(`  玩家名称: ${user.player_name || '未设置'}`);
    console.log(`  当前等级: ${user.level}`);
    console.log(`  当前境界: ${user.realm}`);
    if (cultivation !== null) {
      console.log(`  当前修为: ${user.cultivation}`);
    }
    if (jieYingRate !== null) {
      const baseAttrs = typeof user.base_attributes === 'string' ? JSON.parse(user.base_attributes) : user.base_attributes;
      console.log(`  当前结婴成功率: ${((baseAttrs?.jieYingRate || 0.05) * 100).toFixed(1)}%`);
    }
    
    // 执行更新
    let updateQuery;
    let params;
    let updates = ['level = $2'];
    let paramCount = 3;
    params = [playerId, level];
    
    if (cultivation !== null) {
      updates.push('cultivation = $' + paramCount++);
      params.push(cultivation);
    }
    
    if (jieYingRate !== null) {
      updates.push('base_attributes = jsonb_set(base_attributes, \'{jieYingRate}\', to_jsonb($' + paramCount++ + '::float))');
      params.push(jieYingRate);
    }
    
    updateQuery = 'UPDATE users SET ' + updates.join(', ') + ' WHERE id = $1';
    await client.query(updateQuery, params);
    
    // 获取更新后的信息
    const updatedResult = await client.query(selectQuery, [playerId]);
    const updatedUser = updatedResult.rows[0];
    
    // 显示更新后的信息
    console.log('\n更新后的玩家信息:');
    console.log(`  玩家ID: ${updatedUser.id}`);
    console.log(`  用户名: ${updatedUser.username}`);
    console.log(`  玩家名称: ${updatedUser.player_name || '未设置'}`);
    console.log(`  新等级: ${updatedUser.level}`);
    console.log(`  境界: ${updatedUser.realm}`);
    if (cultivation !== null) {
      console.log(`  新修为: ${updatedUser.cultivation}`);
    }
    if (jieYingRate !== null) {
      const baseAttrs = typeof updatedUser.base_attributes === 'string' ? JSON.parse(updatedUser.base_attributes) : updatedUser.base_attributes;
      console.log(`  新结婴成功率: ${(baseAttrs?.jieYingRate * 100).toFixed(1)}%`);
    }
    
    console.log('\n✓ 玩家信息更新成功!');

  } catch (error) {
    console.error('更新过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

updatePlayerLevel();
