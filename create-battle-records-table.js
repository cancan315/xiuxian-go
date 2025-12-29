// 注意：这是一个Node.js脚本，用于创建战斗记录表
// 支持 PostgreSQL 数据库
// npm install pg dotenv

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

// 数据库类型（固定为postgres）
const DB_TYPE = 'postgres'; // postgres



/**
 * 创建 PostgreSQL 数据库表
 */
async function createPostgreSQLTable() {
  const client = new Client({
    host: process.env.DB_HOST || 'localhost',
    port: process.env.DB_PORT || 5432,
    database: process.env.DB_NAME || 'xiuxian_db',
    user: process.env.DB_USER || 'xiuxian_user',
    password: process.env.DB_PASSWORD || 'xiuxian_password',
  });

  try {
    console.log('正在连接 PostgreSQL 数据库...');
    await client.connect();
    
    // 创建表SQL
    const createTableSQL = `
    CREATE TABLE IF NOT EXISTS battle_records (
      id BIGSERIAL PRIMARY KEY,
      player_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      opponent_id BIGINT NOT NULL,
      opponent_name VARCHAR(255) NOT NULL,
      result VARCHAR(10) NOT NULL CHECK (result IN ('胜利', '失败')),
      battle_type VARCHAR(10) NOT NULL CHECK (battle_type IN ('pvp', 'pve')),
      rewards TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `;
    
    // 创建索引
    const createIndexSQL = `
    CREATE INDEX IF NOT EXISTS idx_battle_records_player_id ON battle_records(player_id);
    CREATE INDEX IF NOT EXISTS idx_battle_records_created_at ON battle_records(created_at);
    CREATE INDEX IF NOT EXISTS idx_battle_records_battle_type ON battle_records(battle_type);
    `;

    console.log('执行建表SQL...');
    await client.query(createTableSQL);
    await client.query(createIndexSQL);
    
    console.log('✓ 战斗记录表创建成功！');
    
    // 验证表是否存在
    const result = await client.query(
      `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'battle_records')`
    );
    
    if (result.rows[0].exists) {
      console.log('\n✓ 表验证成功：battle_records 表已创建');
      
      // 获取表的列信息
      const columnsResult = await client.query(
        `SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = 'battle_records' ORDER BY ordinal_position`
      );
      
      console.log('\n表结构信息：');
      console.log('┌─────────────────┬─────────────┬──────┐');
      console.log('│ 列名            │ 类型        │ 空值 │');
      console.log('├─────────────────┼─────────────┼──────┤');
      
      for (const col of columnsResult.rows) {
        const colName = col.column_name.padEnd(15);
        const colType = col.data_type.padEnd(11);
        const nullable = col.is_nullable === 'YES' ? '是' : '否';
        console.log(`│ ${colName} │ ${colType} │ ${nullable.padEnd(4)} │`);
      }
      
      console.log('└─────────────────┴─────────────┴──────┘');
    }
    
  } catch (error) {
    console.error('❌ 创建表时发生错误:', error.message);
    process.exit(1);
  } finally {
    await client.end();
  }
}

/**
 * 主函数
 */
async function main() {
  console.log('='.repeat(60));
  console.log('战斗记录表创建脚本');
  console.log('='.repeat(60));
  console.log(`\n数据库类型: ${DB_TYPE.toUpperCase()}`);
  console.log(`数据库主机: ${process.env.DB_HOST || 'localhost'}`);
  console.log(`数据库名称: ${process.env.DB_NAME || 'xiuxian_db'}`);
  console.log(`数据库用户: ${process.env.DB_USER}`);
  console.log('');

  try {
    if (DB_TYPE.toLowerCase() === 'postgres') {
      await createPostgreSQLTable();
    } else {
      console.error(`❌ 不支持的数据库类型: ${DB_TYPE}`);
      console.log('支持的类型: postgres');
      process.exit(1);
    }
    
    console.log('\n' + '='.repeat(60));
    console.log('✓ 表创建完成！');
    console.log('='.repeat(60));
    
  } catch (error) {
    console.error('❌ 脚本执行失败:', error.message);
    process.exit(1);
  }
}

// 执行主函数
main();
