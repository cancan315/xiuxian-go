// 注意：这是一个Node.js脚本，用于直接操作数据库
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function updatePlayerSpiritStones() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 2) {
    console.log('使用方法: node add-player-spirit-stones.js <playerId> <spiritStonesAmount>');
    console.log('示例: node add-player-spirit-stones.js 1 1000');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  const spiritStonesAmount = parseInt(args[1]);
  
  // 验证参数
  if (isNaN(playerId) || playerId <= 0) {
    console.error('错误: playerId 必须是正整数');
    process.exit(1);
  }
  
  if (isNaN(spiritStonesAmount) || spiritStonesAmount <= 0) {
    console.error('错误: spiritStonesAmount 必须是正整数');
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
    console.log(`  当前灵石: ${user.spirit_stones}`);
    
    // 执行更新
    const updateQuery = 'UPDATE users SET spirit_stones = spirit_stones + $2 WHERE id = $1';
    await client.query(updateQuery, [playerId, spiritStonesAmount]);
    
    // 获取更新后的信息
    const updatedResult = await client.query(selectQuery, [playerId]);
    const updatedUser = updatedResult.rows[0];
    
    // 显示更新后的信息
    console.log('\n更新后的玩家信息:');
    console.log(`  玩家ID: ${updatedUser.id}`);
    console.log(`  用户名: ${updatedUser.username}`);
    console.log(`  玩家名称: ${updatedUser.player_name || '未设置'}`);
    console.log(`  新灵石数量: ${updatedUser.spirit_stones}`);
    console.log(`  增加灵石: ${spiritStonesAmount}`);
    
    console.log('\n✓ 玩家灵石更新成功!');

  } catch (error) {
    console.error('更新过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

updatePlayerSpiritStones();
