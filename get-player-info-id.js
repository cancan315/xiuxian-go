// 注意：这是一个Node.js脚本，用于查询数据库中指定玩家的所有信息
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function viewPlayerInfo() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 1) {
    console.log('使用方法: node view-player-info.js <playerId>');
    console.log('示例: node view-player-info.js 1');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  
  // 验证参数
  if (isNaN(playerId) || playerId <= 0) {
    console.error('错误: playerId 必须是正整数');
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
    
    // 查询指定玩家的所有信息
    const selectQuery = 'SELECT * FROM users WHERE id = $1';
    const selectResult = await client.query(selectQuery, [playerId]);
    
    if (selectResult.rows.length === 0) {
      console.error(`错误: 未找到ID为 ${playerId} 的玩家`);
      process.exit(1);
    }
    
    const user = selectResult.rows[0];
    
    console.log(`\nID为 ${playerId} 的玩家详细信息:`);
    console.log('==================================================');
    console.log(`ID: ${user.id}`);
    console.log(`用户名: ${user.username}`);
    console.log(`玩家名称: ${user.player_name || '未设置'}`);
    console.log(`等级: ${user.level || 0}`);
    console.log(`境界: ${user.realm || '凡人'}`);
    console.log(`修为: ${user.cultivation || 0}`);
    console.log(`最大修为: ${user.max_cultivation || 0}`);
    console.log(`灵力: ${user.spirit || 0}`);
    console.log(`灵石: ${user.spirit_stones || 0}`);
    console.log(`强化石: ${user.reinforce_stones || 0}`);
    console.log(`洗练石: ${user.refinement_stones || 0}`);
    console.log(`灵宠精华: ${user.pet_essence || 0}`);
    console.log(`道号修改次数: ${user.name_change_count || 0}`);
    console.log(`最后灵力增长时间: ${user.last_spirit_gain_time || '未记录'}`);
    console.log(`创建时间: ${user.created_at}`);
    console.log(`更新时间: ${user.updated_at}`);
    
    // 如果有基础属性、战斗属性等JSON字段，也显示出来
    if (user.base_attributes) {
      console.log(`基础属性: ${JSON.stringify(user.base_attributes, null, 2)}`);
    }
    if (user.combat_attributes) {
      console.log(`战斗属性: ${JSON.stringify(user.combat_attributes, null, 2)}`);
    }
    if (user.combat_resistance) {
      console.log(`战斗抗性: ${JSON.stringify(user.combat_resistance, null, 2)}`);
    }
    if (user.special_attributes) {
      console.log(`特殊属性: ${JSON.stringify(user.special_attributes, null, 2)}`);
    }
    
    console.log('==================================================');
    console.log('✓ 玩家信息查询完成!');

  } catch (error) {
    console.error('查询过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

viewPlayerInfo();
