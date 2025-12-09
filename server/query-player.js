#!/usr/bin/env node

/**
 * 查询玩家数据的工具脚本
 * 使用方法: 
 *   node server/query-player.js <playerId>
 * 示例:
 *   node server/query-player.js 4
 */

const User = require('./models/User');
const Item = require('./models/Item');
const Pet = require('./models/Pet');
const Herb = require('./models/Herb');
const Pill = require('./models/Pill');
const sequelize = require('./models/database');

async function queryPlayer(playerId) {
  try {
    // Connect to the database
    await sequelize.authenticate();
    console.log('数据库连接成功');

    if (!playerId) {
      console.log('请提供玩家ID');
      console.log('使用方法: node server/query-player.js <playerId>');
      return;
    }

    // Fetch user data
    const user = await User.findByPk(playerId);
    if (!user) {
      console.log(`未找到ID为 ${playerId} 的玩家`);
      return;
    }

    console.log('=== 玩家基本信息 ===');
    console.log(`ID: ${user.id}`);
    console.log(`用户名: ${user.username}`);
    console.log(`玩家名称: ${user.playerName}`);
    console.log(`境界等级: ${user.level}`);
    console.log(`当前境界: ${user.realm}`);
    console.log(`当前修为: ${user.cultivation}`);
    console.log(`最大修为: ${user.maxCultivation}`);
    console.log(`灵力值: ${user.spirit}`);
    console.log(`灵石数量: ${user.spiritStones}`);
    console.log(`强化石数量: ${user.reinforceStones}`);
    console.log(`洗练石数量: ${user.refinementStones}`);
    console.log(`是否为新玩家: ${user.isNewPlayer}`);

    // Fetch items
    const items = await Item.findAll({ 
      where: { userId: playerId }
    });
    console.log('\n=== 玩家物品 ===');
    console.log(`物品总数: ${items.length}`);
    if (items.length > 0) {
      items.forEach((item, index) => {
        console.log(`${index + 1}. ${item.name} (类型: ${item.type}, 品质: ${item.quality})`);
      });
    }

    // Fetch pets
    const pets = await Pet.findAll({ where: { userId: playerId } });
    console.log('\n=== 玩家灵宠 ===');
    console.log(`灵宠总数: ${pets.length}`);
    if (pets.length > 0) {
      pets.forEach((pet, index) => {
        console.log(`${index + 1}. ${pet.name} (品质: ${pet.rarity}, 等级: ${pet.level}, 星级: ${pet.star})`);
      });
    }

    // Fetch herbs
    const herbs = await Herb.findAll({ where: { userId: playerId } });
    console.log('\n=== 玩家灵草 ===');
    console.log(`灵草总数: ${herbs.length}`);
    if (herbs.length > 0) {
      herbs.forEach((herb, index) => {
        console.log(`${index + 1}. ${herb.name} (数量: ${herb.count})`);
      });
    }

    // Fetch pills
    const pills = await Pill.findAll({ where: { userId: playerId } });
    console.log('\n=== 玩家丹药 ===');
    console.log(`丹药总数: ${pills.length}`);
    if (pills.length > 0) {
      pills.forEach((pill, index) => {
        console.log(`${index + 1}. ${pill.name}`);
      });
    }

    // Additional stats
    console.log('\n=== 游戏统计数据 ===');
    console.log(`总修炼时间: ${user.totalCultivationTime}`);
    console.log(`突破次数: ${user.breakthroughCount}`);
    console.log(`探索次数: ${user.explorationCount}`);
    console.log(`获得物品数: ${user.itemsFound}`);
    console.log(`触发事件数: ${user.eventTriggered}`);

    console.log('\n=== 解锁内容 ===');
    console.log(`已解锁境界: ${Array.isArray(user.unlockedRealms) ? user.unlockedRealms.join(', ') : user.unlockedRealms}`);
    console.log(`已解锁地点: ${Array.isArray(user.unlockedLocations) ? user.unlockedLocations.join(', ') : user.unlockedLocations}`);

  } catch (error) {
    console.error('查询玩家数据时出错:', error.message);
  } finally {
    // Close the database connection
    await sequelize.close();
    console.log('\n数据库连接已关闭');
  }
}

// Get player ID from command line arguments
const playerId = process.argv[2] ? parseInt(process.argv[2], 10) : null;

queryPlayer(playerId);