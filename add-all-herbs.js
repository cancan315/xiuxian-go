// 为指定玩家添加指定数量的所有灵草
// 需要安装相应的PostgreSQL客户端库
// npm install pg dotenv

const { Client } = require('pg');
require('dotenv').config(); // 加载.env文件中的环境变量

// 所有灵草配置（与后端 exploration/config.go 中的 HerbConfigs 一致）
const HERB_CONFIGS = [
  {
    ID:          "spirit_grass",
    Name:        "灵精草",
    Description: "最常见的灵草，蕴含少量灵气",
    BaseValue:   10,
    Category:    "spirit",
  },
  {
    ID:          "cloud_flower",
    Name:        "云雾花",
    Description: "生长在云雾缭绕处的灵花，有助于修炼",
    BaseValue:   15,
    Category:    "cultivation",
  },
  {
    ID:          "thunder_root",
    Name:        "雷击根",
    Description: "经过雷霆淬炼的灵根，蕴含强大能量",
    BaseValue:   25,
    Category:    "attribute",
  },
  {
    ID:          "dragon_breath_herb",
    Name:        "龙息草",
    Description: "吸收龙气孕育的灵草，极为珍贵",
    BaseValue:   40,
    Category:    "special",
  },
  {
    ID:          "immortal_jade_grass",
    Name:        "仙玉草",
    Description: "传说中生长在仙境的灵草，可遇不可求",
    BaseValue:   60,
    Category:    "special",
  },
  {
    ID:          "dark_yin_grass",
    Name:        "玄阴草",
    Description: "生长在阴暗处的奇特灵草，具有独特的灵气属性",
    BaseValue:   30,
    Category:    "spirit",
  },
  {
    ID:          "nine_leaf_lingzhi",
    Name:        "九叶灵芝",
    Description: "传说中的灵芝，拥有九片叶子，蕴含强大的生命力",
    BaseValue:   45,
    Category:    "cultivation",
  },
  {
    ID:          "purple_ginseng",
    Name:        "紫金参",
    Description: "千年紫参，散发着淡淡的黄金，大补元气",
    BaseValue:   50,
    Category:    "attribute",
  },
  {
    ID:          "frost_lotus",
    Name:        "寒霜莲",
    Description: "生长在极寒之地的莲花，可以提升修炼者的灵力纯度",
    BaseValue:   55,
    Category:    "spirit",
  },
  {
    ID:          "fire_heart_flower",
    Name:        "火心花",
    Description: "生长在火山口的奇花，花心似火焰跳动",
    BaseValue:   35,
    Category:    "attribute",
  },
  {
    ID:          "moonlight_orchid",
    Name:        "月华兰",
    Description: "只在月圆之夜绽放的神秘兰花，能吸收月华精华",
    BaseValue:   70,
    Category:    "spirit",
  },
  {
    ID:          "sun_essence_flower",
    Name:        "日精花",
    Description: "吸收太阳精华的奇花，蕴含纯阳之力",
    BaseValue:   75,
    Category:    "cultivation",
  },
  {
    ID:          "five_elements_grass",
    Name:        "五行草",
    Description: "一株草同时具备金木水火土五种属性的奇珍",
    BaseValue:   80,
    Category:    "attribute",
  },
  {
    ID:          "phoenix_feather_herb",
    Name:        "凤羽草",
    Description: "传说生长在不死火凤栖息地的神草，具有涅槃之力",
    BaseValue:   85,
    Category:    "special",
  },
  {
    ID:          "celestial_dew_grass",
    Name:        "天露草",
    Description: "凝聚天地精华的仙草，千年一遇",
    BaseValue:   90,
    Category:    "special",
  },
];

async function addAllHerbsToPlayer() {
  // 获取命令行参数
  const args = process.argv.slice(2);
  
  if (args.length < 2) {
    console.log('使用方法: node add-all-herbs.js <playerId> <herbAmount>');
    console.log('示例: node add-all-herbs.js 3 10');
    console.log('\n说明:');
    console.log('  playerId: 玩家ID');
    console.log('  herbAmount: 每种灵草添加的数量');
    process.exit(1);
  }
  
  const playerId = parseInt(args[0]);
  const herbAmount = parseInt(args[1]);
  
  // 验证参数
  if (isNaN(playerId) || playerId <= 0) {
    console.error('❌ 错误: playerId 必须是正整数');
    process.exit(1);
  }
  
  if (isNaN(herbAmount) || herbAmount <= 0) {
    console.error('❌ 错误: herbAmount 必须是正整数');
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
    console.log('✓ 数据库连接成功\n');
    
    // 验证玩家是否存在
    const selectQuery = 'SELECT * FROM users WHERE id = $1';
    const selectResult = await client.query(selectQuery, [playerId]);
    
    if (selectResult.rows.length === 0) {
      console.error(`❌ 错误: 未找到ID为 ${playerId} 的玩家`);
      process.exit(1);
    }
    
    const user = selectResult.rows[0];
    console.log('📋 玩家信息:');
    console.log(`  玩家ID: ${user.id}`);
    console.log(`  用户名: ${user.username}`);
    console.log(`  玩家名称: ${user.player_name || '未设置'}\n`);
    
    console.log(`🌿 开始为玩家添加灵草 (每种 ${herbAmount} 个)...\n`);
    
    let addedCount = 0;
    
    // 为每种灵草添加指定数量
    for (const herb of HERB_CONFIGS) {
      // 首先检查玩家是否已有该灵草
      const checkQuery = 'SELECT * FROM herbs WHERE user_id = $1 AND herb_id = $2';
      const checkResult = await client.query(checkQuery, [playerId, herb.ID]);
      
      if (checkResult.rows.length > 0) {
        // 灵草已存在，更新数量
        const herbRecord = checkResult.rows[0];
        const updateQuery = 'UPDATE herbs SET count = count + $3 WHERE user_id = $1 AND herb_id = $2';
        await client.query(updateQuery, [playerId, herb.ID, herbAmount]);
        console.log(`✓ 更新 ${herb.Name} (原数量: ${herbRecord.count} → 新数量: ${herbRecord.count + herbAmount})`);
      } else {
        // 灵草不存在，创建新记录
        const insertQuery = 
          'INSERT INTO herbs (user_id, herb_id, name, count, quality) VALUES ($1, $2, $3, $4, $5)';
        await client.query(insertQuery, [playerId, herb.ID, herb.Name, herbAmount, 'common']);
        console.log(`✓ 新增 ${herb.Name} (数量: ${herbAmount})`);
      }
      
      addedCount++;
    }
    
    console.log(`\n✨ 完成！共处理 ${addedCount} 种灵草`);
    console.log('\n📊 最终灵草库存:');
    
    // 显示更新后的灵草列表
    const finalQuery = 'SELECT * FROM herbs WHERE user_id = $1 ORDER BY herb_id';
    const finalResult = await client.query(finalQuery, [playerId]);
    
    for (const herbRecord of finalResult.rows) {
      console.log(`  - ${herbRecord.name}: ${herbRecord.count} 个`);
    }
    
    console.log('\n✅ 灵草添加成功!');

  } catch (error) {
    console.error('❌ 操作过程中发生错误:', error.message);
    process.exit(1);
  } finally {
    await client.end();
  }
}

addAllHerbsToPlayer();
