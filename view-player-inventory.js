const sequelize = require('./server/models/database');
const User = require('./server/models/User');
const Item = require('./server/models/Equipment');

async function viewPlayerAttributes(playerId) {
  try {
    // 查找玩家
    const user = await User.findByPk(playerId);
    
    if (!user) {
      console.log(`未找到ID为 ${playerId} 的玩家`);
      return;
    }

    console.log(`玩家信息:`);
    console.log(`  ID: ${user.id}`);
    console.log(`  用户名: ${user.username}`);
    console.log(`  角色名: ${user.playerName}`);
    console.log(`  等级: ${user.level}`);
    console.log(`  境界: ${user.realm}`);
    console.log(`  修为: ${user.cultivation}/${user.maxCultivation}`);
    console.log(`  灵力: ${user.spirit}`);
    console.log(`  灵石: ${user.spiritStones}`);
    console.log(`  强化石: ${user.reinforceStones}`);
    console.log(`  洗练石: ${user.refinementStones}`);
    console.log('');
    
    console.log('基础属性:');
    const baseAttributes = typeof user.baseAttributes === 'string' ? JSON.parse(user.baseAttributes) : user.baseAttributes;
    Object.entries(baseAttributes).forEach(([key, value]) => {
      console.log(`  ${key}: ${value}`);
    });
    
    console.log('\n战斗属性:');
    const combatAttributes = typeof user.combatAttributes === 'string' ? JSON.parse(user.combatAttributes) : user.combatAttributes;
    Object.entries(combatAttributes).forEach(([key, value]) => {
      console.log(`  ${key}: ${value}`);
    });
    
    console.log('\n战斗抗性:');
    const combatResistance = typeof user.combatResistance === 'string' ? JSON.parse(user.combatResistance) : user.combatResistance;
    Object.entries(combatResistance).forEach(([key, value]) => {
      console.log(`  ${key}: ${value}`);
    });
    
    console.log('\n特殊属性:');
    const specialAttributes = typeof user.specialAttributes === 'string' ? JSON.parse(user.specialAttributes) : user.specialAttributes;
    Object.entries(specialAttributes).forEach(([key, value]) => {
      console.log(`  ${key}: ${value}`);
    });
    
    console.log('\n统计数据:');
    console.log(`  总修炼时间: ${user.totalCultivationTime}`);
    console.log(`  突破次数: ${user.breakthroughCount}`);
    console.log(`  探索次数: ${user.explorationCount}`);
    console.log(`  发现物品: ${user.itemsFound}`);
    console.log(`  触发事件: ${user.eventTriggered}`);
    
    console.log('\n已解锁内容:');
    console.log(`  已解锁境界: ${Array.isArray(user.unlockedRealms) ? user.unlockedRealms.join(', ') : user.unlockedRealms}`);
    console.log(`  已解锁地点: ${Array.isArray(user.unlockedLocations) ? user.unlockedLocations.join(', ') : user.unlockedLocations}`);
    console.log(`  已解锁技能: ${Array.isArray(user.unlockedSkills) ? user.unlockedSkills.join(', ') : user.unlockedSkills}`);
    
    console.log('\n设置:');
    console.log(`  暗黑模式: ${user.isDarkMode ? '是' : '否'}`);
    
  } catch (error) {
    console.error('查询过程中发生错误:', error);
  } finally {
    // 关闭数据库连接
    await sequelize.close();
    console.log('数据库连接已关闭');
  }
}

// 获取命令行参数中的玩家ID
const args = process.argv.slice(2);
const playerId = args.length > 0 ? parseInt(args[0]) : 1;

// 调用函数执行查询
viewPlayerAttributes(playerId);
