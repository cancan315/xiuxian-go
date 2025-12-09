const sequelize = require('./server/models/database');
const User = require('./server/models/User');

/**
 * 根据玩家ID增加强化石数量
 * @param {number} playerId - 玩家ID
 * @param {number} amount - 要增加的强化石数量
 */
async function addReinforceStones(playerId, amount) {
  try {
    // 查找玩家
    const user = await User.findByPk(playerId);
    
    if (!user) {
      console.log(`未找到ID为 ${playerId} 的玩家`);
      return;
    }
    
    // 增加强化石数量
    const currentStones = user.reinforceStones || 0;
    const newStones = currentStones + amount;
    
    // 更新玩家数据
    await User.update(
      { reinforceStones: newStones },
      { where: { id: playerId } }
    );
    
    console.log(`成功为玩家 ${user.username} (ID: ${playerId}) 增加 ${amount} 个强化石`);
    console.log(`强化石数量从 ${currentStones} 变更为 ${newStones}`);
    
  } catch (error) {
    console.error('增加强化石时发生错误:', error);
  } finally {
    // 关闭数据库连接
    await sequelize.close();
    console.log('数据库连接已关闭');
  }
}

// 检查命令行参数
if (process.argv.length !== 4) {
  console.log('使用方法: node add-reinforce-stones.js <playerId> <amount>');
  console.log('示例: node add-reinforce-stones.js 1 100');
  process.exit(1);
}

// 获取命令行参数
const playerId = parseInt(process.argv[2]);
const amount = parseInt(process.argv[3]);

// 验证参数
if (isNaN(playerId) || isNaN(amount)) {
  console.log('错误: playerId 和 amount 必须是数字');
  process.exit(1);
}

if (amount <= 0) {
  console.log('错误: amount 必须大于0');
  process.exit(1);
}

// 执行增加强化石操作
addReinforceStones(playerId, amount);