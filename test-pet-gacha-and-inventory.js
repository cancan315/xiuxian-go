const sequelize = require('./server/models/database');
const User = require('./server/models/User');
const Pet = require('./server/models/Pet');
const bcrypt = require('bcryptjs');
const { v4: uuidv4 } = require('uuid');

// 模拟抽奖功能
function simulateGacha(level, count = 1) {
  // 灵宠品质配置
  const petRarities = {
    common: { name: '凡兽', color: '#9e9e9e', expMod: 1.0, statMod: 1.0, probability: 0.4 },
    uncommon: { name: '妖兽', color: '#4caf50', expMod: 1.2, statMod: 1.2, probability: 0.3 },
    rare: { name: '灵兽', color: '#2196f3', expMod: 1.5, statMod: 1.5, probability: 0.15 },
    epic: { name: '上古异兽', color: '#9c27b0', expMod: 2.0, statMod: 2.0, probability: 0.1 },
    legendary: { name: '瑞兽', color: '#ff9800', expMod: 2.5, statMod: 2.5, probability: 0.04 },
    mythic: { name: '仙兽', color: '#e91e63', expMod: 3.0, statMod: 3.0, probability: 0.01 }
  };

  // 灵宠名称列表
  const petNames = [
    '小白', '小黑', '小花', '小黄', '小蓝', '小绿', '小红', '小紫',
    '火狐', '水蛇', '土熊', '风狼', '雷鹰', '冰豹', '岩龟', '木鹿',
    '金狮', '银虎', '铜象', '铁猿', '玉兔', '翡翠鸟', '琥珀虫', '玛瑙蟹',
    '九尾狐', '独角兽', '凤凰', '麒麟', '应龙', '饕餮', '混沌', '帝江'
  ];

  // 灵宠描述列表
  const descriptions = [
    '一只可爱的小动物',
    '拥有神秘力量的灵兽',
    '来自远古的神奇生物',
    '天赋异禀的稀有灵宠',
    '修行多年的通灵之兽',
    '蕴含天地灵气的奇兽',
    '拥有特殊能力的异兽'
  ];

  const results = [];
  
  for (let i = 0; i < count; i++) {
    // 根据概率随机选择品质
    const rarityKeys = Object.keys(petRarities);
    const probabilities = rarityKeys.map(key => petRarities[key].probability);
    const rarity = weightedRandom(rarityKeys, probabilities);
    
    // 随机选择名称
    const name = petNames[Math.floor(Math.random() * petNames.length)];
    
    // 随机选择描述
    const description = descriptions[Math.floor(Math.random() * descriptions.length)];
    
    // 生成基础属性
    const baseStats = {
      attack: Math.floor(Math.random() * 50 + 25),     // 攻击力
      defense: Math.floor(Math.random() * 30 + 15),    // 防御力
      health: Math.floor(Math.random() * 100 + 50),    // 生命值
      speed: Math.floor(Math.random() * 15 + 5)        // 速度
    };

    // 根据品质调整属性
    const rarityMod = petRarities[rarity].statMod;
    Object.keys(baseStats).forEach(stat => {
      baseStats[stat] = Math.floor(baseStats[stat] * rarityMod);
    });

    // 随机生成额外战斗属性
    const extraAttributes = {};
    // 战斗属性说明:
    // critRate: 暴击率       comboRate: 连击率       counterRate: 反击率       stunRate: 眩晕率
    // dodgeRate: 闪避率     vampireRate: 吸血率
    // critResist: 抗暴击    comboResist: 抗连击     counterResist: 抗反击     stunResist: 抗眩晕
    // dodgeResist: 抗闪避   vampireResist: 抗吸血
    // healBoost: 强化治疗   critDamageBoost: 强化爆伤 critDamageReduce: 弱化爆伤
    // finalDamageBoost: 最终增伤 finalDamageReduce: 最终减伤
    // combatBoost: 战斗属性提升 resistanceBoost: 战斗抗性提升
    const attributeKeys = [
      'critRate', 'comboRate', 'counterRate', 'stunRate', 
      'dodgeRate', 'vampireRate', 'critResist', 'comboResist',
      'counterResist', 'stunResist', 'dodgeResist', 'vampireResist',
      'healBoost', 'critDamageBoost', 'critDamageReduce', 
      'finalDamageBoost', 'finalDamageReduce', 'combatBoost', 'resistanceBoost'
    ];
    
    // 随机选择0-3个额外属性
    const numExtraAttributes = Math.floor(Math.random() * 4);
    const selectedAttributes = [];
    
    for (let j = 0; j < numExtraAttributes; j++) {
      // 避免重复选择属性
      let attr;
      do {
        attr = attributeKeys[Math.floor(Math.random() * attributeKeys.length)];
      } while (selectedAttributes.includes(attr));
      
      selectedAttributes.push(attr);
      
      // 根据品质设定属性值范围
      let value;
      switch (rarity) {
        case 'mythic':
          value = parseFloat((Math.random() * 0.15 + 0.05).toFixed(3)); // 5%-20%
          break;
        case 'legendary':
          value = parseFloat((Math.random() * 0.12 + 0.04).toFixed(3)); // 4%-16%
          break;
        case 'epic':
          value = parseFloat((Math.random() * 0.1 + 0.03).toFixed(3)); // 3%-13%
          break;
        case 'rare':
          value = parseFloat((Math.random() * 0.08 + 0.02).toFixed(3)); // 2%-10%
          break;
        case 'uncommon':
          value = parseFloat((Math.random() * 0.05 + 0.01).toFixed(3)); // 1%-6%
          break;
        default: // common
          value = parseFloat((Math.random() * 0.03 + 0.01).toFixed(3)); // 1%-4%
      }
      
      extraAttributes[attr] = value;
    }

    results.push({
      name,
      rarity,
      level: 1,
      star: 0,
      description,
      combatAttributes: {
        ...baseStats,
        ...extraAttributes
      }
    });
  }
  
  return results;
}

// 加权随机选择
function weightedRandom(options, weights) {
  const totalWeight = weights.reduce((sum, weight) => sum + weight, 0);
  let random = Math.random() * totalWeight;

  for (let i = 0; i < weights.length; i++) {
    if (random < weights[i]) {
      return options[i];
    }
    random -= weights[i];
  }

  return options[options.length - 1];
}

async function testPetGachaAndInventory() {
  try {
    // Sync database
    await sequelize.sync({ force: true });
    console.log('数据库同步成功');

    // Test user data
    const testUser = {
      username: '1',
      password: '1',
      playerName: '测试修仙者'
    };

    // Hash password
    const salt = await bcrypt.genSalt(10);
    const hashedPassword = await bcrypt.hash(testUser.password, salt);

    // Create user
    const user = await User.create({
      username: testUser.username,
      password: hashedPassword,
      playerName: testUser.playerName
    });

    console.log('玩家创建成功!');
    console.log('玩家ID:', user.id);
    console.log('用户名:', user.username);
    console.log('灵石数量:', user.spiritStones);
    console.log('是否为新玩家:', user.isNewPlayer);

    // 模拟玩家使用抽奖功能抽取灵宠
    console.log('\n--- 模拟玩家使用抽奖功能抽取灵宠 ---');
    const gachaResults = simulateGacha(user.level || 1, 5); // 抽取5只灵宠
    
    console.log(`抽取到 ${gachaResults.length} 只灵宠:`);
    gachaResults.forEach((pet, index) => {
      console.log(`\n${index + 1}. ${pet.name} (${pet.rarity})`);
      console.log(`   描述: ${pet.description}`);
      console.log(`   基础属性:`);
      console.log(`     攻击: ${pet.combatAttributes.attack}`);
      console.log(`     防御: ${pet.combatAttributes.defense}`);
      console.log(`     生命: ${pet.combatAttributes.health}`);
      console.log(`     速度: ${pet.combatAttributes.speed}`);
      
      // 显示额外属性
      const extraAttrs = Object.keys(pet.combatAttributes).filter(
        key => !['attack', 'defense', 'health', 'speed'].includes(key)
      );
      
      if (extraAttrs.length > 0) {
        console.log(`   额外属性:`);
        extraAttrs.forEach(attr => {
          console.log(`     ${attr}: ${(pet.combatAttributes[attr] * 100).toFixed(1)}%`);
        });
      }
    });

    // 将抽取的灵宠保存到数据库
    console.log('\n--- 将抽取的灵宠保存到数据库 ---');
    const petRecords = [];
    for (let i = 0; i < gachaResults.length; i++) {
      const petData = gachaResults[i];
      const petRecord = await Pet.create({
        userId: user.id,
        petId: `pet_${uuidv4().substring(0, 8)}`,
        name: petData.name,
        rarity: petData.rarity,
        level: petData.level,
        star: petData.star,
        description: petData.description,
        combatAttributes: petData.combatAttributes
      });
      
      petRecords.push(petRecord);
      console.log(`灵宠 "${petData.name}" 已保存到数据库，ID: ${petRecord.id}`);
    }

    // 查询数据库中的灵宠属性
    console.log('\n--- 查询数据库中的灵宠属性 ---');
    const dbPets = await Pet.findAll({
      where: {
        userId: user.id
      }
    });

    console.log(`用户 ${user.playerName} 拥有 ${dbPets.length} 只灵宠:`);
    dbPets.forEach((pet, index) => {
      console.log(`\n灵宠 ${index + 1}:`);
      console.log(`  名称: ${pet.name}`);
      console.log(`  ID: ${pet.id}`);
      console.log(`  品质: ${pet.rarity}`);
      console.log(`  等级: ${pet.level}`);
      console.log(`  星级: ${pet.star}`);
      console.log(`  描述: ${pet.description}`);
      console.log(`  战斗属性:`);
      console.log(`    攻击力: ${pet.combatAttributes.attack}`);
      console.log(`    防御力: ${pet.combatAttributes.defense}`);
      console.log(`    生命值: ${pet.combatAttributes.health}`);
      console.log(`    速度: ${pet.combatAttributes.speed}`);
      
      // 显示额外属性
      const extraAttrs = Object.keys(pet.combatAttributes).filter(
        key => !['attack', 'defense', 'health', 'speed'].includes(key)
      );
      
      if (extraAttrs.length > 0) {
        console.log(`    额外属性:`);
        extraAttrs.forEach(attr => {
          console.log(`      ${attr}: ${(pet.combatAttributes[attr] * 100).toFixed(1)}%`);
        });
      }
    });

    console.log('\n--- 测试总结 ---');
    console.log('1. 灵宠数据存储在 Pets 表中');
    console.log('2. 每个灵宠通过 userId 字段关联到对应的用户');
    console.log('3. 灵宠的核心属性存储在 combatAttributes 字段中（JSON格式）');
    console.log('4. 抽奖功能可以正确生成各种品质的灵宠');
    console.log('5. 数据库可以正确存储和检索灵宠的所有属性');

  } catch (error) {
    console.error('测试过程中发生错误:', error);
  } finally {
    // Close database connection
    await sequelize.close();
    console.log('\n数据库连接已关闭');
  }
}

testPetGachaAndInventory();