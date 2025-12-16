// 注意：这是一个Node.js脚本，用于直接操作数据库
// 需要安装相应的PostgreSQL客户端库
// npm install pg

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

async function updatePlayerResources() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 6) {
    console.log('使用方法: node update-player-resources.js <playerId> <spiritStones> <reinforceStones> <refinementStones> <petEssence> <spirit>');
    console.log('示例: node update-player-resources.js 1 1000 500 300 200 100');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  const spiritStones = parseInt(args[1]);
  const reinforceStones = parseInt(args[2]);
  const refinementStones = parseInt(args[3]);
  const petEssence = parseInt(args[4]);
  const spirit = parseFloat(args[5]);
  
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
    console.log(`  灵石: ${user.spirit_stones}`);
    console.log(`  强化石: ${user.reinforce_stones}`);
    console.log(`  洗练石: ${user.refinement_stones}`);
    console.log(`  灵宠精华: ${user.pet_essence}`);
    console.log(`  灵力: ${user.spirit}`);
    
    // 构建更新查询
    const updates = [];
    const values = [playerId];
    let valueIndex = 2;
    
    if (!isNaN(spiritStones)) {
      updates.push(`spirit_stones = spirit_stones + $${valueIndex}`);
      values.push(spiritStones);
      valueIndex++;
    }
    
    if (!isNaN(reinforceStones)) {
      updates.push(`reinforce_stones = reinforce_stones + $${valueIndex}`);
      values.push(reinforceStones);
      valueIndex++;
    }
    
    if (!isNaN(refinementStones)) {
      updates.push(`refinement_stones = refinement_stones + $${valueIndex}`);
      values.push(refinementStones);
      valueIndex++;
    }
    
    if (!isNaN(petEssence)) {
      updates.push(`pet_essence = pet_essence + $${valueIndex}`);
      values.push(petEssence);
      valueIndex++;
    }
    
    if (!isNaN(spirit)) {
      updates.push(`spirit = spirit + $${valueIndex}`);
      values.push(spirit);
      valueIndex++;
    }
    
    if (updates.length === 0) {
      console.log('没有指定任何有效的资源更新');
      process.exit(1);
    }
    
    // 执行更新
    const updateQuery = `UPDATE users SET ${updates.join(', ')} WHERE id = $1`;
    await client.query(updateQuery, values);
    
    // 获取更新后的信息
    const updatedResult = await client.query(selectQuery, [playerId]);
    const updatedUser = updatedResult.rows[0];
    
    // 显示更新后的信息
    console.log('\n更新后的玩家信息:');
    console.log(`  玩家ID: ${updatedUser.id}`);
    console.log(`  用户名: ${updatedUser.username}`);
    console.log(`  灵石: ${updatedUser.spirit_stones}`);
    console.log(`  强化石: ${updatedUser.reinforce_stones}`);
    console.log(`  洗练石: ${updatedUser.refinement_stones}`);
    console.log(`  灵宠精华: ${updatedUser.pet_essence}`);
    console.log(`  灵力: ${parseFloat(updatedUser.spirit).toFixed(2)}`);
    
    console.log('\n✓ 玩家资源更新成功!');
    
  } catch (error) {
    console.error('更新过程中发生错误:', error);
    process.exit(1);
  } finally {
    await client.end();
  }
}

updatePlayerResources();