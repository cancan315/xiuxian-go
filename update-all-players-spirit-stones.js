// 注意：这是一个Node.js脚本，用于直接操作数据库
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function updateAllPlayersSpiritStones() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 1) {
    console.log('使用方法: node update-all-players-spirit-stones.js <spiritStonesAmount>');
    console.log('示例: node update-all-players-spirit-stones.js 1000');
    process.exit(1);
  }
  
  const spiritStonesAmount = parseInt(args[0]);
  
  // 验证参数
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
    
    // 获取所有玩家数量
    const countQuery = 'SELECT COUNT(*) FROM users';
    const countResult = await client.query(countQuery);
    const totalPlayers = parseInt(countResult.rows[0].count);
    
    console.log(`正在为 ${totalPlayers} 名玩家增加灵石...`);
    
    // 更新所有玩家的灵石
    const updateQuery = 'UPDATE users SET spirit_stones = spirit_stones + $1';
    await client.query(updateQuery, [spiritStonesAmount]);
    
    console.log(`✓ 成功为 ${totalPlayers} 名玩家增加了 ${spiritStonesAmount} 灵石!`);
    
  } catch (error) {
    console.error('更新过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

updateAllPlayersSpiritStones();