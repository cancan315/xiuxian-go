const { Sequelize } = require('sequelize');
const User = require('../models/User');
const Item = require('../models/Item');
const Equipment = require('../models/Equipment');
const Pet = require('../models/Pet');
const { v4: uuidv4 } = require('uuid');

// 装备品质概率配置 (总和应为100%)
const EQUIPMENT_QUALITY_PROBABILITIES = {
  mythic: 0.001,     // 极品仙器 0.1%
  legendary: 0.003,  // 仙器 0.3%
  epic: 0.016,       // 极品灵器 1.6%
  rare: 0.03,        // 灵器 3%
  uncommon: 0.15,    // 极品法器 15%
  common: 0.80       // 法器 80%
};

// 灵宠品质概率配置 (总和应为100%)
const PET_RARITY_PROBABILITIES = {
  mythic: 0.001,     // 仙兽 0.1%
  legendary: 0.003,  // 瑞兽 0.3%
  epic: 0.016,       // 上古异兽 1.6%
  rare: 0.03,        // 灵兽 3%
  uncommon: 0.15,    // 妖兽 15%
  common: 0.80       // 凡兽 80%
};

// 装备类型
const equipmentTypes = {
  faqi: {
    name: '法宝',
    slot: 'faqi',
    prefixes: {
      common: ['粗制', '劣质', '破损', '锈蚀'],               // 凡器
      uncommon: ['精制', '优质', '改良', '强化'],             // 法器
      rare: ['上品', '精品', '珍品', '灵韵'],                // 灵器
      epic: ['异宝', '奇珍', '灵光', '宝华'],                // 极品灵器
      legendary: ['法宝', '神兵', '灵宝', '仙锋'],            // 伪仙器
      mythic: ['仙器', '神器', '天器', '圣器']               // 仙器
    }
  },
  guanjin: {
    name: '冠巾',
    slot: 'guanjin',
    prefixes: {
      common: ['布制', '麻织', '粗布', '素巾'],               // 凡器
      uncommon: ['丝织', '锦制', '绣冠', '轻纱'],             // 法器
      rare: ['灵丝', '云锦', '霓裳', '羽冠'],                // 灵器
      epic: ['宝冠', '灵冠', '星辰', '月华'],                // 极品灵器
      legendary: ['仙冠', '神冠', '紫金', '龙纹'],            // 伪仙器
      mythic: ['天冠', '圣冠', '混沌', '太虚']               // 仙器
    }
  },
  daopao: {
    name: '道袍',
    slot: 'daopao',
    prefixes: {
      common: ['粗布', '麻衣', '布衣', '素袍'],               // 凡器
      uncommon: ['丝绸', '锦袍', '绣衣', '轻衫'],             // 法器
      rare: ['灵绸', '云裳', '霞衣', '霓裳'],                // 灵器
      epic: ['宝衣', '灵衣', '星辰', '月华'],                // 极品灵器
      legendary: ['仙衣', '神袍', '紫绶', '龙袍'],            // 伪仙器
      mythic: ['天衣', '圣袍', '混沌', '太虚']               // 仙器
    }
  },
  yunlv: {
    name: '云履',
    slot: 'yunlv',
    prefixes: {
      common: ['布鞋', '草鞋', '麻鞋', '木屐'],               // 凡器
      uncommon: ['皮靴', '丝履', '锦履', '绣鞋'],             // 法器
      rare: ['灵靴', '云履', '霞履', '霓履'],                // 灵器
      epic: ['宝履', '灵履', '星辰', '月华'],                // 极品灵器
      legendary: ['仙履', '神履', '紫霞', '龙履'],            // 伪仙器
      mythic: ['天履', '圣履', '混沌', '太虚']               // 仙器
    }
  },
  fabao: {
    name: '本命法宝',
    slot: 'fabao',
    prefixes: {
      common: ['粗制', '劣质', '仿制', '赝品'],               // 凡器
      uncommon: ['精制', '良品', '上品', '优质'],             // 法器
      rare: ['灵宝', '珍宝', '异宝', '奇宝'],                // 灵器
      epic: ['重宝', '至宝', '灵光', '宝华'],                // 极品灵器
      legendary: ['仙宝', '神宝', '天宝', '圣宝'],            // 伪仙器
      mythic: ['至尊', '无上', '混沌', '太虚']               // 仙器
    }
  }
};

// 按品质分类的灵宠名称
const petNamesByRarity = {
  common: [      // 凡兽 - 12个名称
    '白狐', '灰狼', '黄牛', '黑虎',
    '赤马', '棕熊', '青蛇', '紫貂',
    '银鼠', '金蝉', '彩雀', '田园犬'
  ],
  uncommon: [    // 妖兽 - 12个名称
    '妖猴', '妖驴', '妖豹', '妖蛇',
    '妖虎', '妖熊', '妖鹰', '妖蝎',
    '妖蛛', '妖蝠', '妖蟾', '妖蜈'
  ],
  rare: [        // 灵兽 - 12个名称
    '灵狐', '灵鹿', '灵龟', '灵蛇',
    '灵猿', '灵鹤', '灵鲤', '灵雀',
    '灵虎', '灵豹', '灵猫', '灵犬'
  ],
  epic: [        // 上古异兽 - 12个名称
    '饕餮', '穷奇', '梼杌', '混沌',
    '九婴', '相柳', '凿齿', '修蛇',
    '封豨', '大风', '巴蛇', '朱厌'
  ],
  legendary: [   // 瑞兽 - 12个名称
    '麒麟', '凤凰', '龙龟', '白泽',
    '重明鸟', '当康', '乘黄', '英招',
    '夫诸', '天马', '青牛', '玄龟'
  ],
  mythic: [      // 仙兽 - 12个名称
    '应龙', '夔牛', '毕方', '饕餮',
    '九尾狐', '玉兔', '金蟾', '青鸾',
    '火凤', '水麒麟', '土蝼', '陆吾'
  ]
};

// 按品质分类的灵宠描述
const petDescriptionsByRarity = {
  common: [      // 凡兽描述
    '一只普通的小动物，刚刚开启灵智',
    '刚刚踏上修仙路的凡兽，潜力有限',
    '随处可见的普通凡兽，资质平庸',
    '最低等的凡兽，仅有微弱的灵力',
    '毛茸茸的小家伙，虽然灵力微弱但十分忠诚',
    '山野间常见的小生灵，懵懂单纯'
  ],
  uncommon: [    // 妖兽描述
    '已具初步妖力的妖兽，有一定培养价值',
    '开启了妖智的兽类，具备基本法术能力',
    '经过初步修炼的妖兽，战斗力尚可',
    '初具成长的妖兽，已能施展简单法术',
    '体内蕴含妖力精华，有望成长为强大伙伴',
    '灵智已开，能听懂主人简单的指令'
  ],
  rare: [        // 灵兽描述
    '天生蕴含灵气的珍稀兽类，颇具潜力',
    '灵力充沛的高等灵兽，是不错的伙伴',
    '具有强大灵力的兽类，战斗力出色',
    '千年难遇的灵兽，拥有不俗的天赋',
    '通体散发着柔和灵光，天生与道有缘',
    '灵性非凡，可助主人感悟天地大道'
  ],
  epic: [        // 上古异兽描述
    '来自上古时代的神秘异兽，力量强大',
    '传说中的远古生物，拥有毁天灭地之力',
    '稀世罕见的上古凶兽，威能惊人',
    '承载着远古血脉的强大生物，实力深不可测',
    '血脉中流淌着太古神力，潜力无穷',
    '一吼可震动山河，乃是天地孕育的奇珍'
  ],
  legendary: [   // 瑞兽描述
    '祥瑞降临，世间少有的瑞兽',
    '蕴含天地精华的瑞兽，气运之子',
    '可佑主人大道亨通的瑞兽',
    '拥有莫大威能的瑞兽，极为罕见',
    '身披祥云，踏足之处百邪不侵',
    '瑞气千条，能为主人带来无上机缘和庇佑'
  ],
  mythic: [      // 仙兽描述
    '超越凡俗的仙界仙兽，几近传说',
    '来自仙境的存在，拥有通天彻地之能',
    '近乎神明般的存在，掌握天地法则',
    '传说中只存在于神话中的至高仙兽',
    '仙气缭，举手投足间蕴含大道至理',
    '超脱三界之外，不在五行之中，已窥仙道真谛'
  ]
};

// 属性类型定义
const attributeTypes = {
  // 基础属性
  baseAttributes: ['attack', 'health', 'defense', 'speed'],
  
  // 战斗属性
  combatAttributes: [
    'critRate', 'comboRate', 'counterRate', 
    'stunRate', 'dodgeRate', 'vampireRate'
  ],
  
  // 战斗抗性
  combatResistance: [
    'critResist', 'comboResist', 'counterResist',
    'stunResist', 'dodgeResist', 'vampireResist'
  ],
  
  // 特殊属性
  specialAttributes: [
    'healBoost', 'critDamageBoost', 'critDamageReduce',
    'finalDamageBoost', 'finalDamageReduce', 
    'combatBoost', 'resistanceBoost'
  ]
};

// 不同品质的属性生成规则
const qualityAttributeRules = {
  // 装备品质规则
  equipment: {
    common: { base: 1 },        // 凡器: 1条基础属性
    uncommon: { base: 2 },      // 法器: 2条基础属性
    rare: { base: 3 },          // 灵器: 3条基础属性
    epic: { base: 4, combat: 1 }, // 极品灵器: 4条基础属性 + 1条战斗属性
    legendary: { base: 4, combat: 6, resistance: 6 }, // 伪仙器: 4条基础属性 + 6条战斗属性 + 6条战斗抗性
    mythic: { base: 4, combat: 6, resistance: 6, special: 7 } // 仙器: 4条基础属性 + 6条战斗属性 + 6条战斗抗性 + 7条特殊属性
  },
  
  // 灵宠品质规则
  pet: {
    common: { base: 1 },        // 凡兽: 1条基础属性
    uncommon: { base: 2 },      // 妖兽: 2条基础属性
    rare: { base: 3 },          // 灵兽: 3条基础属性
    epic: { base: 4, combat: 1 }, // 上古异兽: 4条基础属性 + 1条战斗属性
    legendary: { base: 4, combat: 6, resistance: 6 }, // 瑞兽: 4条基础属性 + 6条战斗属性 + 6条战斗抗性
    mythic: { base: 4, combat: 6, resistance: 6, special: 7 } // 仙兽: 4条基础属性 + 6条战斗属性 + 6条战斗抗性 + 7条特殊属性
  }
};

// 加权随机选择
const weightedRandom = (weights) => {
  const totalWeight = weights.reduce((sum, weight) => sum + weight, 0);
  let random = Math.random() * totalWeight;

  for (let i = 0; i < weights.length; i++) {
    if (random < weights[i]) {
      return i;
    }
    random -= weights[i];
  }

  return weights.length - 1;
};

// 随机选择指定数量的属性
const selectRandomAttributes = (attributeList, count) => {
  const shuffled = [...attributeList].sort(() => 0.5 - Math.random());
  return shuffled.slice(0, count);
};

// 生成指定属性的随机值
const generateAttributeValue = (attrType, quality, isPet = false) => {
  // 基础属性范围
  const baseRanges = {
    attack: { min: 30, max: 100 },
    health: { min: 60, max: 200 },
    defense: { min: 15, max: 50 },
    speed: { min: 5, max: 20 }
  };
  
  // 战斗属性范围 (百分比)
  const combatRateRanges = {
    min: 0.01, // 1%
    max: 0.10  // 10%
  };
  
  // 战斗抗性范围 (百分比)
  const resistanceRateRanges = {
    min: 0.01, // 1%
    max: 0.08  // 8%
  };
  
  // 特殊属性范围 (百分比)
  const specialRateRanges = {
    min: 0.005, // 0.5%
    max: 0.05   // 5%
  };
  
  // 根据属性类型生成值
  if (attributeTypes.baseAttributes.includes(attrType)) {
    const range = baseRanges[attrType] || { min: 10, max: 50 };
    // 根据品质调整数值
    const qualityMods = {
      common: 1.0,
      uncommon: 1.2,
      rare: 1.5,
      epic: 2.0,
      legendary: 2.5,
      mythic: 3.0
    };
    
    const mod = qualityMods[quality] || 1.0;
    return Math.floor((Math.random() * (range.max - range.min) + range.min) * mod);
  } 
  else if (attributeTypes.combatAttributes.includes(attrType)) {
    return Number((Math.random() * (combatRateRanges.max - combatRateRanges.min) + combatRateRanges.min).toFixed(3));
  }
  else if (attributeTypes.combatResistance.includes(attrType)) {
    return Number((Math.random() * (resistanceRateRanges.max - resistanceRateRanges.min) + resistanceRateRanges.min).toFixed(3));
  }
  else if (attributeTypes.specialAttributes.includes(attrType)) {
    return Number((Math.random() * (specialRateRanges.max - specialRateRanges.min) + specialRateRanges.min).toFixed(3));
  }
  
  // 默认值
  return Math.floor(Math.random() * 50) + 10;
};

// 生成指定品质的属性集合
const generateQualityBasedAttributes = (type, quality) => {
  const rules = qualityAttributeRules[type][quality];
  const attributes = {};
  
  // 生成基础属性
  if (rules.base) {
    const selectedBaseAttrs = selectRandomAttributes(attributeTypes.baseAttributes, rules.base);
    selectedBaseAttrs.forEach(attr => {
      attributes[attr] = generateAttributeValue(attr, quality, type === 'pet');
    });
  }
  
  // 生成战斗属性
  if (rules.combat) {
    const selectedCombatAttrs = selectRandomAttributes(attributeTypes.combatAttributes, rules.combat);
    selectedCombatAttrs.forEach(attr => {
      attributes[attr] = generateAttributeValue(attr, quality, type === 'pet');
    });
  }
  
  // 生成战斗抗性属性
  if (rules.resistance) {
    const selectedResistanceAttrs = selectRandomAttributes(attributeTypes.combatResistance, rules.resistance);
    selectedResistanceAttrs.forEach(attr => {
      attributes[attr] = generateAttributeValue(attr, quality, type === 'pet');
    });
  }
  
  // 生成特殊属性
  if (rules.special) {
    const selectedSpecialAttrs = selectRandomAttributes(attributeTypes.specialAttributes, rules.special);
    selectedSpecialAttrs.forEach(attr => {
      attributes[attr] = generateAttributeValue(attr, quality, type === 'pet');
    });
  }
  
  return attributes;
};

// 生成随机装备
const generateRandomEquipment = (level) => {
  // 随机选择装备品质（按固定概率）
  const qualities = Object.keys(EQUIPMENT_QUALITY_PROBABILITIES);
  // 按照品质顺序重新排列，保证高品质在前
  qualities.sort((a, b) => {
    const order = ['mythic', 'legendary', 'epic', 'rare', 'uncommon', 'common'];
    return order.indexOf(a) - order.indexOf(b);
  });
  
  const qualityWeights = qualities.map(q => EQUIPMENT_QUALITY_PROBABILITIES[q] * 100);
  const qualityIndex = weightedRandom(qualityWeights);
  const quality = qualities[qualityIndex];

  // 随机选择装备类型
  const equipTypes = Object.keys(equipmentTypes);
  const equipType = equipTypes[Math.floor(Math.random() * equipTypes.length)];

  // 根据装备类型和品质生成装备名称
  const prefixes = equipmentTypes[equipType].prefixes[quality];
  const prefix = prefixes[Math.floor(Math.random() * prefixes.length)];
  // 移除品质字段，只使用前缀和装备类型名称
  const typeName = equipmentTypes[equipType].name;
  const name = `${prefix}${typeName}`;

  // 根据品质生成属性
  const stats = generateQualityBasedAttributes('equipment', quality);

  // 计算装备的境界要求（装备等级的一半，最小为1）
  // 确保传入的level是一个有效数字，默认为1
  const numericLevel = isNaN(level) || level === undefined || level === null ? 1 : level;
  
  // 根据玩家境界确定装备的requiredRealm值
  // 1对应炼气期一层-九层，2对应筑基期一层到筑基期九层...以此类推
  let requiredRealm = 1;
  if (numericLevel >= 1 && numericLevel <= 9) {
    requiredRealm = 1; // 练气期
  } else if (numericLevel >= 10 && numericLevel <= 18) {
    requiredRealm = 2; // 筑基期
  } else if (numericLevel >= 19 && numericLevel <= 27) {
    requiredRealm = 3; // 金丹期
  } else if (numericLevel >= 28 && numericLevel <= 36) {
    requiredRealm = 4; // 元婴期
  } else if (numericLevel >= 37 && numericLevel <= 45) {
    requiredRealm = 5; // 化神期
  } else if (numericLevel >= 46 && numericLevel <= 54) {
    requiredRealm = 6; // 返虚期
  } else if (numericLevel >= 55 && numericLevel <= 63) {
    requiredRealm = 7; // 合体期
  } else if (numericLevel >= 64 && numericLevel <= 72) {
    requiredRealm = 8; // 大乘期
  } else if (numericLevel >= 73 && numericLevel <= 81) {
    requiredRealm = 9; // 渡劫期
  } else if (numericLevel >= 82 && numericLevel <= 90) {
    requiredRealm = 10; // 散仙期
  } else if (numericLevel >= 91 && numericLevel <= 99) {
    requiredRealm = 11; // 仙人期
  } else if (numericLevel >= 100 && numericLevel <= 108) {
    requiredRealm = 12; // 真仙期
  } else if (numericLevel >= 109 && numericLevel <= 117) {
    requiredRealm = 13; // 金仙期
  } else if (numericLevel >= 118 && numericLevel <= 126) {
    requiredRealm = 14; // 太乙期
  } else if (numericLevel >= 127 && numericLevel <= 135) {
    requiredRealm = 15; // 大罗期
  } else {
    requiredRealm = 15; // 大罗期（最高境界）
  }

  return {
    id: uuidv4(),
    name,
    type: 'equipment',
    quality: quality,
    equipType,
    level: numericLevel, // 装备要求字段，默认为炼气期
    requiredRealm: requiredRealm, // 添加境界要求字段
    enhanceLevel: 0, // 添加强化等级字段，默认为0
    stats: stats,
    extraAttributes: {},
    createdAt: new Date().toISOString()
  };
};

// 生成随机灵宠
const generateRandomPet = (level) => {
  // 随机选择灵宠品质（按固定概率）
  const rarities = Object.keys(PET_RARITY_PROBABILITIES);
  // 按照品质顺序重新排列，保证高品质在前
  rarities.sort((a, b) => {
    const order = ['mythic', 'legendary', 'epic', 'rare', 'uncommon', 'common'];
    return order.indexOf(a) - order.indexOf(b);
  });
  
  const rarityWeights = rarities.map(r => PET_RARITY_PROBABILITIES[r] * 1000);
  const rarityIndex = weightedRandom(rarityWeights);
  const rarity = rarities[rarityIndex];

  // 根据品质选择对应名称
  const petNames = petNamesByRarity[rarity] || petNamesByRarity.common;
  const name = petNames[Math.floor(Math.random() * petNames.length)];

  // 根据品质生成属性
  const combatAttributes = generateQualityBasedAttributes('pet', rarity);

  // 根据品质选择对应描述
  const descriptions = petDescriptionsByRarity[rarity] || petDescriptionsByRarity.common;
  const description = descriptions[Math.floor(Math.random() * descriptions.length)];

  // 计算属性加成
  const qualityBonusMap = {
    mythic: 0.15, // 仙兽基础加成15%
    legendary: 0.12, // 瑞兽基础加成12%
    epic: 0.09, // 上古异兽基础加成9%
    rare: 0.06, // 灵兽基础加成6%
    uncommon: 0.03, // 妖兽基础加成3%
    common: 0.03 // 凡兽品质基础加成3%
  }
  const starBonusPerQuality = {
    mythic: 0.02, // 仙兽每星+2%
    legendary: 0.01, // 瑞兽每星+1%
    epic: 0.01, // 上古异兽每星+1%
    rare: 0.01, // 灵兽每星+1%
    uncommon: 0.01, // 妖兽每星+1%
    common: 0.01 // 凡兽每星+1% 
  };

  const baseBonus = qualityBonusMap[rarity] || 0;
  const starBonus = 0 * (starBonusPerQuality[rarity] || 0); // 初始星级为0
  const levelBonus = (1 - 1) * (baseBonus * 0.1); // 初始等级为1
  const phase = Math.floor(0 / 5); // 初始星级为0
  const phaseBonus = phase * (baseBonus * 0.5);
  const finalBonus = baseBonus + starBonus + levelBonus + phaseBonus;

  return {
    id: uuidv4(),
    name,
    type: 'pet',
    rarity: rarity,
    level: 1,
    star: 0,
    exp: 0,
    description,
    combatAttributes: combatAttributes,
    attackBonus: finalBonus,
    defenseBonus: finalBonus,
    healthBonus: finalBonus,
    createdAt: new Date().toISOString()
  };
};

// 执行抽奖
const drawGacha = async (req, res) => {
  try {
    const { poolType, count, useWishlist } = req.body;
    const userId = req.user.id;
    
    console.log(`[Server Gacha] 用户开始抽奖，用户ID: ${userId}，类型: ${poolType}，数量: ${count}，使用心愿单: ${useWishlist}`);

    // 定义抽卡消耗
    const cost = useWishlist ? 
      { 1: 200, 10: 2000, 50: 10000, 100: 20000 } :
      { 1: 100, 10: 1000, 50: 5000, 100: 10000 };
    
    const requiredSpiritStones = cost[count] || (useWishlist ? 200 : 100) * count;
    
    // 获取用户信息（用于确定物品等级和检查灵石）
    const user = await User.findByPk(userId);
    if (!user) {
      console.error(`[Server Gacha] 用户不存在，用户ID: ${userId}`);
      return res.status(404).json({ success: false, message: '用户不存在' });
    }
    
    console.log(`[Server Gacha] 用户灵石余额: ${user.spiritStones}，所需灵石: ${requiredSpiritStones}`);
    
    // 检查灵石是否足够
    if (user.spiritStones < requiredSpiritStones) {
      console.warn(`[Server Gacha] 用户灵石不足，用户ID: ${userId}`);
      return res.status(400).json({ success: false, message: '灵石不足' });
    }
    
    // 生成物品
    const items = [];
    for (let i = 0; i < count; i++) {
      if (poolType === 'equipment') {
        const equipment = generateRandomEquipment(user.realm);
        console.log(`[Server Gacha] 生成装备: ${equipment.name}，品质: ${equipment.quality}`);
        // 添加详细日志，打印装备的全部字段信息
        console.log(`[Server Gacha] 生成的装备详细信息:`, JSON.stringify(equipment, null, 2));
        items.push(equipment);
      } else if (poolType === 'pet') {
        const pet = generateRandomPet(user.realm);
        console.log(`[Server Gacha] 生成灵宠: ${pet.name}，品质: ${pet.rarity}`);
        items.push(pet);
      }
    }
    
    // 保存物品到数据库
    console.log(`[Server Gacha] 开始保存物品到数据库，物品数量: ${items.length}`);
    for (const item of items) {
      if (item.type === 'equipment') {
        console.log(`[Server Gacha] 准备保存装备到数据库: ${item.name}, ID: ${item.id}, 品质: ${item.quality}, 类型: ${item.equipType}, 强化等级: ${item.enhanceLevel}`);
        // 添加详细日志，打印即将保存到数据库的装备信息
        console.log(`[Server Gacha] 保存到数据库的装备详细信息:`, JSON.stringify(item, null, 2));
        await Equipment.create({
          userId,
          equipmentId: item.id,
          name: item.name,
          type: item.type,
          quality: item.quality,
          equipType: item.equipType,  // 只保存 equipType 字段，不映射到 slot 字段
          level: item.level,
          enhanceLevel: item.enhanceLevel, // 添加 enhanceLevel 字段映射
          requiredRealm: item.requiredRealm, // 添加境界要求字段
          stats: JSON.stringify(item.stats),
          extraAttributes: JSON.stringify(item.extraAttributes),
          equipped: false
        });
        console.log(`[Server Gacha] 装备保存成功: ${item.name}, ID: ${item.id}`);
      } else if (item.type === 'pet') {
        console.log(`[Server Gacha] 准备保存灵宠到数据库: ${item.name}, ID: ${item.id}, 品质: ${item.rarity}`);
        await Pet.create({
          userId,
          petId: item.id,
          name: item.name,
          type: item.type,
          rarity: item.rarity,
          level: item.level,
          star: item.star,
          exp: item.exp,
          description: item.description,
          combatAttributes: JSON.stringify(item.combatAttributes),
          attackBonus: item.attackBonus,
          defenseBonus: item.defenseBonus,
          healthBonus: item.healthBonus,
          isActive: false
        });
        console.log(`[Server Gacha] 灵宠保存成功: ${item.name}, ID: ${item.id}`);
      }
    }
    console.log(`[Server Gacha] 物品保存完成，总共保存数量: ${items.length}`);
    
    // 扣除灵石
    console.log(`[Server Gacha] 扣除灵石，原数量: ${user.spiritStones}，扣除数量: ${requiredSpiritStones}`);
    await User.update(
      { spiritStones: user.spiritStones - requiredSpiritStones },
      { where: { id: userId } }
    );
    
    res.status(200).json({ 
      success: true, 
      items, 
      message: '抽奖成功',
      spiritStones: user.spiritStones - requiredSpiritStones
    });
    console.log(`[Server Gacha] 抽奖完成，用户ID: ${userId}，获得物品数量: ${items.length}`);
  } catch (error) {
    console.error('[Server Gacha] 抽奖失败:', error);
    res.status(500).json({ success: false, message: '服务器错误' });
  }
};

// 执行自动处理操作
const processAutoActions = async (req, res) => {
  try {
    const { items, autoSellQualities, autoReleaseRarities } = req.body;
    const userId = req.user.id;
    
    console.log(`[Server Gacha] 开始执行自动处理，用户ID: ${userId}，物品数量: ${items?.length || 0}`);

    const soldItems = [];
    const releasedPets = [];
    let stonesGained = 0;
    
    // 处理自动出售装备
    for (const item of items) {
      if (item.type === 'equipment' && autoSellQualities.includes(item.quality)) {
        // 删除装备
        await Equipment.destroy({
          where: {
            userId,
            equipmentId: item.id
          }
        });
        
        // 计算获得的强化石数量（根据品质）
        const stoneValues = {
          common: 1,
          uncommon: 2,
          rare: 5,
          epic: 10,
          legendary: 20,
          mythic: 50
        };
        
        stonesGained += stoneValues[item.quality] || 1;
        soldItems.push(item);
        console.log(`[Server Gacha] 自动出售装备: ${item.name}，品质: ${item.quality}`);
      } else if (item.type === 'pet' && autoReleaseRarities.includes(item.rarity)) {
        // 删除灵宠
        await Pet.destroy({
          where: {
            userId,
            petId: item.id
          }
        });
        
        // 计算获得的强化石数量（根据品质）
        const stoneValues = {
          common: 1,
          uncommon: 2,
          rare: 5,
          epic: 10,
          legendary: 20,
          mythic: 50
        };
        
        stonesGained += stoneValues[item.rarity] || 1;
        releasedPets.push(item);
        console.log(`[Server Gacha] 自动放生灵宠: ${item.name}，品质: ${item.rarity}`);
      }
    }
    
    // 更新用户的强化石数量
    if (stonesGained > 0) {
      console.log(`[Server Gacha] 更新用户强化石数量，增加: ${stonesGained}`);
      await User.increment('reinforceStones', {
        by: stonesGained,
        where: { id: userId }
      });
    }
    
    res.status(200).json({ 
      success: true, 
      soldItems, 
      releasedPets, 
      stonesGained,
      message: '自动处理完成' 
    });
    
    console.log(`[Server Gacha] 自动处理完成，卖出装备: ${soldItems.length}，放生灵宠: ${releasedPets.length}，获得强化石: ${stonesGained}`);
  } catch (error) {
    console.error('[Server Gacha] 自动处理失败:', error);
    res.status(500).json({ success: false, message: '服务器错误' });
  }
};

module.exports = {
  drawGacha,
  processAutoActions
};