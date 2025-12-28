// 注意：这是一个Node.js脚本，用于查询数据库中的所有玩家ID和名称
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function viewAllPlayers() {
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

    // 查询所有玩家的ID和名称
    const selectQuery = 'SELECT id, username FROM users ORDER BY id';
    const selectResult = await client.query(selectQuery);

    if (selectResult.rows.length === 0) {
      console.log('数据库中没有玩家');
      return;
    }

    console.log(`共找到 ${selectResult.rows.length} 名玩家:`);
    console.log('----------------------------------------');
    console.log('ID\t\t用户名');
    console.log('----------------------------------------');

    // 显示所有玩家信息
    selectResult.rows.forEach((user) => {
      console.log(`${user.id}\t\t${user.username}`);
    });

    console.log('----------------------------------------');
    console.log(`✓ 共显示 ${selectResult.rows.length} 名玩家信息!`);

  } catch (error) {
    console.error('查询过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

viewAllPlayers();
