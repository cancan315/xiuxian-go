const Equipment = require('../models/Equipment');
const { Sequelize } = require('sequelize');

// 强化等级配置
const enhanceConfig = {
  maxLevel: 150, // 最大强化等级改为150级
  baseSuccessRates: [
    { level: 0, maxLevel: 10, requiredRealm: 1, successRate: 0.5 },   // 0-10级要求练气期，成功率50%
    { level: 11, maxLevel: 20, requiredRealm: 2, successRate: 0.35 }, // 11-20级要求筑基期，成功率35%
    { level: 21, maxLevel: 30, requiredRealm: 3, successRate: 0.3 },  // 21-30级要求金丹期，成功率30%
    { level: 31, maxLevel: 40, requiredRealm: 4, successRate: 0.25 }, // 31-40级要求元婴期，成功率25%
    { level: 41, maxLevel: 50, requiredRealm: 5, successRate: 0.2 }, // 41-50级要求化神期，成功率20%
    { level: 51, maxLevel: 60, requiredRealm: 6, successRate: 0.15 }, // 51-60级要求返虚期，成功率15%
    { level: 61, maxLevel: 70, requiredRealm: 7, successRate: 0.1 }, // 61-70级要求合体期，成功率10%
    { level: 71, maxLevel: 80, requiredRealm: 8, successRate: 0.05 }, // 71-80级要求大乘期，成功率5%
    { level: 81, maxLevel: 90, requiredRealm: 9, successRate: 0.025 }, // 81-90级要求渡劫期，成功率2.5%
    { level: 91, maxLevel: 100, requiredRealm: 10, successRate: 0.02 }, // 91-100级要求散仙期，成功率2%
    { level: 101, maxLevel: 110, requiredRealm: 11, successRate: 0.015 }, // 101-110级要求仙人期，成功率1.5%
    { level: 111, maxLevel: 120, requiredRealm: 12, successRate: 0.01 }, // 111-120级要求真仙期，成功率1%
    { level: 121, maxLevel: 130, requiredRealm: 13, successRate: 0.005 }, // 121-130级要求金仙期，成功率0.5%
    { level: 131, maxLevel: 140, requiredRealm: 14, successRate: 0.0025 }, // 131-140级要求太乙期，成功率0.25%
    { level: 141, maxLevel: 150, requiredRealm: 15, successRate: 0.001 }, // 141-150级要求大罗期，成功率0.1%
    
  ],
  costPerLevel: 10, // 每级消耗的强化石数量
  statIncrease: 0.1 // 每级属性提升比例（10%）
}

// 洗练配置
const reforgeConfig = {
  costPerAttempt: 10, // 每次洗练消耗的洗练石数量
  minVariation: -0.5, // 最小属性变化（-50%）
  maxVariation: 0.5, // 最大属性变化（+50%）
  newStatChance: 0.3 // 更换属性的概率（30%）
}

// 可洗练的属性池
const reforgeableStats = {
  weapon: ['attack', 'critRate', 'critDamageBoost'],
  head: ['defense', 'health', 'stunResist'],
  body: ['defense', 'health', 'finalDamageReduce'],
  legs: ['defense', 'speed', 'dodgeRate'],
  feet: ['defense', 'speed', 'dodgeRate'],
  shoulder: ['defense', 'health', 'counterRate'],
  hands: ['attack', 'critRate', 'comboRate'],
  wrist: ['defense', 'counterRate', 'vampireRate'],
  necklace: ['health', 'healBoost', 'spiritRate'],
  ring1: ['attack', 'critDamageBoost', 'finalDamageBoost'],
  ring2: ['defense', 'critDamageReduce', 'resistanceBoost'],
  belt: ['health', 'defense', 'combatBoost'],
  artifact: ['attack', 'critRate', 'comboRate'],
  faqi: ['attack', 'critRate', 'critDamageBoost'],
  guanjin: ['defense', 'health', 'stunResist'],
  daopao: ['defense', 'health', 'finalDamageReduce'],
  yunlv: ['defense', 'speed', 'dodgeRate'],
  fabao: ['attack', 'critRate', 'comboRate']
}

/**
 * 根据境界名称获取对应的requiredRealm값
 * 1对应炼气期一层-九层
 * 2对应筑基期一层到筑基期九层
 * ...
 * 14对应太乙期一层到太乙九层
 * 15对应大罗期一层到大罗九层
 * @param {string} realmName - 境界名称
 * @returns {number} requiredRealm값
 */
const getRequiredRealm = (realmName) => {
  if (!realmName) return 1;
  
  // 境界期名称映射到requiredRealm값
  const realmPeriods = [
    { period: '练气期', value: 1 },
    { period: '筑基期', value: 2 },
    { period: '金丹期', value: 3 },
    { period: '元婴期', value: 4 },
    { period: '化神期', value: 5 },
    { period: '返虚期', value: 6 },
    { period: '合体期', value: 7 },
    { period: '大乘期', value: 8 },
    { period: '渡劫期', value: 9 },
    { period: '散仙期', value: 10 },
    { period: '仙人期', value: 11 },
    { period: '真仙期', value: 12 },
    { period: '金仙期', value: 13 },
    { period: '太乙期', value: 14 },
    { period: '大罗期', value: 15 }
  ];
  
  // 查找对应的境界期
  for (const period of realmPeriods) {
    if (realmName.includes(period.period)) {
      return period.value;
    }
  }
  
  // 默认返回1（练气期）
  return 1;
}

// 强化装备
function enhanceEquipmentLogic(equipment, playerReinforceStones, playerRealm) {
  if (!equipment || !equipment.stats) {
    return { success: false, message: '无效的装备' }
  }
  const currentLevel = equipment.enhanceLevel || 0
  if (currentLevel >= enhanceConfig.maxLevel) {
    return { success: false, message: '装备已达到最大强化等级' }
  }
  
  // 检查玩家境界是否满足强化要求
  let requiredRealm = 1; // 默认需要练气期
  let successRate = 0.5; // 默认成功率50%
  
  // 根据强化等级确定所需的境界和成功率
  for (const rateConfig of enhanceConfig.baseSuccessRates) {
    if (currentLevel >= rateConfig.level && currentLevel <= rateConfig.maxLevel) {
      requiredRealm = rateConfig.requiredRealm;
      successRate = rateConfig.successRate;
      break;
    }
  }
  
  // 如果没有匹配到预设配置，则使用原有的递减逻辑
  if (requiredRealm === 1 && successRate === 0.5 && currentLevel >= 31) {
    requiredRealm = Math.max(1, Math.floor(currentLevel / 10)); // 每10级提高境界要求
    successRate = Math.max(0.1, 1 - currentLevel * 0.05); // 成功率递减
  }
  
  // 检查玩家境界
  if (playerRealm < requiredRealm) {
    return { success: false, message: '玩家境界不足，无法强化' }
  }
  
  const cost = enhanceConfig.costPerLevel * (currentLevel + 1)
  if (playerReinforceStones < cost) {
    return { success: false, message: '强化石不足' }
  }
  
  const isSuccess = Math.random() < successRate
  if (!isSuccess) {
    return {
      success: false,
      message: '强化失败',
      cost,
      oldStats: { ...equipment.stats },
      newStats: { ...equipment.stats }
    }
  }
  
  // 保存旧属性用于对比
  const oldStats = { ...equipment.stats }
  // 提升装备属性
  Object.keys(equipment.stats).forEach(stat => {
    if (typeof equipment.stats[stat] === 'number') {
      equipment.stats[stat] *= 1 + enhanceConfig.statIncrease
      // 对百分比属性进行特殊处理
      if (
        ['critRate', 'critDamageBoost', 'dodgeRate', 'vampireRate', 'finalDamageBoost', 'finalDamageReduce'].includes(
          stat
        )
      ) {
        equipment.stats[stat] = Math.round(equipment.stats[stat] * 100) / 100
      } else {
        equipment.stats[stat] = Math.round(equipment.stats[stat])
      }
    }
  })
  // 更新强化等级
  equipment.enhanceLevel = (equipment.enhanceLevel || 0) + 1
  
  // 根据强化等级更新装备的境界要求
  let newRequiredRealm = 1;
  if (equipment.enhanceLevel >= 0 && equipment.enhanceLevel <= 10) {
    newRequiredRealm = 1;
  } else if (equipment.enhanceLevel >= 11 && equipment.enhanceLevel <= 20) {
    newRequiredRealm = 2;
  } else if (equipment.enhanceLevel >= 21 && equipment.enhanceLevel <= 30) {
    newRequiredRealm = 3;
  } else if (equipment.enhanceLevel >= 31 && equipment.enhanceLevel <= 40) {
    newRequiredRealm = 4;
  } else if (equipment.enhanceLevel >= 41 && equipment.enhanceLevel <= 50) {
    newRequiredRealm = 5;
  } else if (equipment.enhanceLevel >= 51 && equipment.enhanceLevel <= 60) {
    newRequiredRealm = 6;
  } else if (equipment.enhanceLevel >= 61 && equipment.enhanceLevel <= 70) {
    newRequiredRealm = 7;
  } else if (equipment.enhanceLevel >= 71 && equipment.enhanceLevel <= 80) {
    newRequiredRealm = 8;
  } else if (equipment.enhanceLevel >= 81 && equipment.enhanceLevel <= 90) {
    newRequiredRealm = 9;
  } else if (equipment.enhanceLevel >= 91 && equipment.enhanceLevel <= 100) {
    newRequiredRealm = 10;
  } else if (equipment.enhanceLevel >= 101 && equipment.enhanceLevel <= 110) {
    newRequiredRealm = 11;
  } else if (equipment.enhanceLevel >= 111 && equipment.enhanceLevel <= 120) {
    newRequiredRealm = 12;
  } else if (equipment.enhanceLevel >= 121 && equipment.enhanceLevel <= 130) {
    newRequiredRealm = 13;
  } else if (equipment.enhanceLevel >= 131 && equipment.enhanceLevel <= 140) {
    newRequiredRealm = 14;
  } else if (equipment.enhanceLevel >= 141 && equipment.enhanceLevel <= 150) {
    newRequiredRealm = 15;
  }
  
  // 更新装备的境界要求
  equipment.requiredRealm = newRequiredRealm;
  
  return {
    success: true,
    message: '强化成功',
    cost,
    oldStats,
    newStats: equipment.stats,
    newLevel: equipment.enhanceLevel,
    newRequiredRealm: equipment.requiredRealm
  }
}

// 洗练装备
function reforgeEquipmentLogic(equipment, playerRefinementStones, confirmNewStats = true) {
  if (!equipment || !equipment.stats || !equipment.type) {
    return { success: false, message: '无效的装备' }
  }
  if (playerRefinementStones < reforgeConfig.costPerAttempt) {
    return { success: false, message: '洗练石不足' }
  }
  const oldStats = { ...equipment.stats }
  const availableStats = reforgeableStats[equipment.type]
  const tempStats = { ...equipment.stats }
  const originStats = Object.keys(tempStats)
  // 生成要处理的属性索引（1-3个随机）
  const modifyIndexes = [
    ...new Set(
      Array.from({ length: Math.floor(Math.random() * 3) + 1 }, () => Math.floor(Math.random() * originStats.length))
    )
  ].slice(0, 3) // 确保最多处理3个属性
  modifyIndexes.forEach(index => {
    const originStat = originStats[index]
    let currentStat = originStat
    const baseValue = tempStats[originStat]
    // Step 1: 尝试替换属性
    if (Math.random() < reforgeConfig.newStatChance) {
      // 过滤可用属性（不包含现有属性）
      const availableNew = availableStats.filter(s => !originStats.includes(s) && s !== originStat)
      if (availableNew.length > 0) {
        const newStat = availableNew[Math.floor(Math.random() * availableNew.length)]
        // 替换属性名但保留当前数值（会在步骤2中调整）
        delete tempStats[originStat]
        currentStat = newStat
      }
    }
    // Step 2：强制数值调整（基于原始值±50%）
    const delta = Math.random() * 1.0 - 0.5 // [-0.5, 0.5]
    const newValue = baseValue * (1 + delta)
    // 根据属性类型处理数值精度
    if (
      ['critRate', 'critDamageBoost', 'dodgeRate', 'vampireRate', 'finalDamageBoost', 'finalDamageReduce'].includes(
        currentStat
      )
    ) {
      tempStats[currentStat] = Math.min(Math.max(Number(newValue.toFixed(2)), baseValue * 0.5), baseValue * 1.5)
    } else {
      tempStats[currentStat] = Math.min(
        Math.max(Math.round(newValue), Math.round(baseValue * 0.5)),
        Math.round(baseValue * 1.5)
      )
    }
  })
  // 强制属性数量校验
  if (Object.keys(tempStats).length !== originStats.length) {
    console.error('属性数量异常', { old: originStats, new: tempStats })
    return {
      success: false,
      message: '洗练过程出现异常',
      cost: 0,
      oldStats,
      newStats: oldStats
    }
  }
  if (confirmNewStats) {
    equipment.stats = { ...tempStats }
  }
  return {
    success: true,
    message: confirmNewStats ? '洗练成功' : '保留原有属性',
    cost: reforgeConfig.costPerAttempt,
    oldStats,
    newStats: tempStats,
    confirmed: confirmNewStats
  }
}

// 获取玩家装备列表
const getPlayerEquipment = async (req, res) => {
  try {
    const { userId } = req.params;
    const { type, quality, equipped } = req.query;
    
    // 优先使用URL参数中的userId，其次使用认证用户ID
    const actualUserId = userId || (req.user && req.user.id);
    
    // 如果没有提供有效的userId，返回错误
    if (!actualUserId) {
      return res.status(400).json({
        success: false,
        message: '缺少用户ID参数'
      });
    }
    
    // 构建查询条件
    const whereConditions = {
      userId: actualUserId
    };
    
    // 如果提供了类型过滤条件
    if (type) {
      whereConditions.equipType = type;
    }
    
    // 如果提供了equipType过滤条件
    if (equipType) {
      whereConditions.equipType = equipType;
    }
    
    // 如果提供了品质过滤条件
    if (quality) {
      whereConditions.quality = quality;
    }
    
    // 如果提供了装备状态过滤条件
    if (equipped !== undefined) {
      whereConditions.equipped = equipped === 'true';
    }
    
    // 查询符合条件的装备
    const equipment = await Equipment.findAll({
      where: whereConditions,
      order: [['createdAt', 'DESC']] // 按创建时间倒序排列
    });
    
    // 处理从数据库获取的装备数据
    const processedEquipment = equipment.map(item => {
      // 如果stats是字符串，则尝试解析为对象
      if (item.stats && typeof item.stats === 'string') {
        try {
          item.stats = JSON.parse(item.stats);
        } catch (e) {
          console.error('解析装备stats失败:', e);
        }
      }
      
      // 如果details是字符串，则尝试解析为对象
      if (item.details && typeof item.details === 'string') {
        try {
          item.details = JSON.parse(item.details);
        } catch (e) {
          console.error('解析装备details失败:', e);
        }
      }
      
      return item;
    });
    
    res.json({
      success: true,
      equipment: processedEquipment
    });
  } catch (error) {
    console.error('获取玩家装备数据失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 获取特定装备详情
const getEquipmentDetails = async (req, res) => {
  try {
    const { id } = req.params;
    
    // 查询特定装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id // 确保只能访问自己的装备
      }
    });
    
    if (!equipment) {
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    // 处理装备数据
    const processedEquipment = { ...equipment.dataValues };
    
    // 如果stats是字符串，则尝试解析为对象
    if (processedEquipment.stats && typeof processedEquipment.stats === 'string') {
      try {
        processedEquipment.stats = JSON.parse(processedEquipment.stats);
      } catch (e) {
        console.error('解析装备stats失败:', e);
      }
    }
    
    // 如果details是字符串，则尝试解析为对象
    if (processedEquipment.details && typeof processedEquipment.details === 'string') {
      try {
        processedEquipment.details = JSON.parse(processedEquipment.details);
      } catch (e) {
        console.error('解析装备details失败:', e);
      }
    }
    
    res.json({
      success: true,
      equipment: processedEquipment
    });
  } catch (error) {
    console.error('获取装备详情失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 装备强化
const enhanceEquipment = async (req, res) => {
  try {
    const { id } = req.params;
    const { reinforceStones } = req.body; // 客户端传递的强化石数量
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    if (!equipment) {
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    // 获取玩家信息
    const User = require('../models/User');
    const user = await User.findByPk(req.user.id);
    if (!user) {
      return res.status(404).json({
        success: false,
        message: '玩家未找到'
      });
    }
    
    // 解析stats字段
    if (equipment.stats && typeof equipment.stats === 'string') {
      try {
        equipment.stats = JSON.parse(equipment.stats);
      } catch (e) {
        console.error('解析装备stats失败:', e);
      }
    }
    
    // 获取玩家的境界值（requiredRealm）
    const playerRequiredRealm = getRequiredRealm(user.realm);
    
    // 执行强化逻辑，传入玩家境界
    const result = enhanceEquipmentLogic(equipment, reinforceStones, playerRequiredRealm);
    
    if (result.success) {
      // 保存更新后的装备数据
      equipment.stats = JSON.stringify(equipment.stats);
      equipment.enhanceLevel = result.newLevel;
      equipment.requiredRealm = result.newRequiredRealm; // 更新装备的境界要求
      await equipment.save();
      
      // 扣除强化石
      await User.update(
        { reinforceStones: user.reinforceStones - result.cost },
        { where: { id: req.user.id } }
      );
      
      // 检查装备要求境界是否大于玩家当前境界，如果是则自动卸下装备
      if (equipment.equipped && result.newRequiredRealm > playerRequiredRealm) {
        // 更新装备状态为未装备
        equipment.equipped = false;
        equipment.slot = null; // 保持slot为null，表示未装备状态
        await equipment.save();
        
        console.log(`[Equipment Controller] 装备强化后境界要求超过玩家境界，自动卸下装备: ${equipment.name}`);
      }
      
      // 返回结果
      res.json({
        success: true,
        message: result.message,
        cost: result.cost,
        oldStats: result.oldStats,
        newStats: result.newStats,
        newLevel: result.newLevel,
        newRequiredRealm: result.newRequiredRealm
      });
    } else {
      res.json({
        success: false,
        message: result.message,
        cost: result.cost,
        oldStats: result.oldStats,
        newStats: result.newStats
      });
    }
  } catch (error) {
    console.error('装备强化失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 装备洗练
const reforgeEquipment = async (req, res) => {
  try {
    const { id } = req.params;
    const { refinementStones } = req.body;
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    if (!equipment) {
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    // 解析stats字段
    if (equipment.stats && typeof equipment.stats === 'string') {
      try {
        equipment.stats = JSON.parse(equipment.stats);
      } catch (e) {
        console.error('解析装备stats失败:', e);
      }
    }
    
    // 执行洗练逻辑
    const result = reforgeEquipmentLogic(equipment, refinementStones, false);
    
    if (result.success) {
      // 注意：这里我们不直接保存洗练结果，而是返回给客户端确认
      res.json({
        success: true,
        message: result.message,
        cost: result.cost,
        oldStats: result.oldStats,
        newStats: result.newStats
      });
    } else {
      res.json({
        success: false,
        message: result.message
      });
    }
  } catch (error) {
    console.error('装备洗练失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 确认洗练结果
const confirmReforge = async (req, res) => {
  try {
    const { id } = req.params;
    const { confirmed, newStats } = req.body;
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    if (!equipment) {
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    if (confirmed) {
      // 应用新的属性
      equipment.stats = JSON.stringify(newStats);
      await equipment.save();
      
      res.json({
        success: true,
        message: '洗练属性已应用',
        stats: newStats
      });
    } else {
      res.json({
        success: true,
        message: '已保留原有属性'
      });
    }
  } catch (error) {
    console.error('确认洗练结果失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 装备穿戴
const equipEquipment = async (req, res) => {
  try {
    const { id } = req.params;
    const { slot } = req.body;
    
    console.log(`[Equipment Controller] 装备穿戴请求: 装备ID=${id}, 用户ID=${req.user.id}, 槽位=${slot}`);
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    if (!equipment) {
      console.log(`[Equipment Controller] 装备未找到: 装备ID=${id}, 用户ID=${req.user.id}`);
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    console.log(`[Equipment Controller] 找到装备: ${equipment.name}, 当前装备状态: equipped=${equipment.equipped}, slot=${equipment.slot}`);
    // 添加装备穿戴前的详细信息日志
    console.log(`[Equipment Controller] 装备穿戴前的装备详细信息:`, JSON.stringify(equipment, null, 2));
    
    // 解析stats字段
    if (equipment.stats && typeof equipment.stats === 'string') {
      try {
        equipment.stats = JSON.parse(equipment.stats);
        console.log(`[Equipment Controller] 成功解析装备stats:`, equipment.stats);
      } catch (e) {
        console.error('[Equipment Controller] 解析装备stats失败:', e);
      }
    }
    
    // 检查同类型的装备是否已经装备
    const existingEquipped = await Equipment.findOne({
      where: {
        userId: req.user.id,
        equipType: equipment.equipType, // 使用 equipType 而不是 type 来检查同类型装备
        equipped: true
      }
    });
    
    if (existingEquipped && existingEquipped.id !== equipment.id) {
      console.log(`[Equipment Controller] 同类型装备已装备: ${existingEquipped.name}, 无法重复装备同类型装备: ${equipment.name}`);
      console.log(`[Equipment Controller] 已装备的同类型装备详情:`, JSON.stringify(existingEquipped, null, 2));
      return res.status(400).json({
        success: false,
        message: '同类型装备已装备，无法重复装备'
      });
    }
    
    // 更新装备状态为已装备
    equipment.equipped = true;
    // 使用 equipType 作为槽位类型，但不覆盖已有的 equipType 值
    equipment.slot = equipment.slot || equipment.equipType;
    await equipment.save();
    
    // 应用装备属性加成
    const User = require('../models/User');
    const user = await User.findByPk(req.user.id);
    
    console.log(`[Equipment Controller] 装备穿戴前玩家属性:`, JSON.stringify({
      baseAttributes: user.baseAttributes,
      combatAttributes: user.combatAttributes,
      combatResistance: user.combatResistance,
      specialAttributes: user.specialAttributes
    }, null, 2));
    
    let updatedBaseAttributes = user.baseAttributes;
    let updatedCombatAttributes = user.combatAttributes;
    let updatedCombatResistance = user.combatResistance;
    let updatedSpecialAttributes = user.specialAttributes;
    
    if (equipment.stats) {
      // 克隆当前属性以避免直接修改
      let baseAttributes = JSON.parse(JSON.stringify(user.baseAttributes || {}));
      let combatAttributes = JSON.parse(JSON.stringify(user.combatAttributes || {}));
      let combatResistance = JSON.parse(JSON.stringify(user.combatResistance || {}));
      let specialAttributes = JSON.parse(JSON.stringify(user.specialAttributes || {}));
      
      console.log(`[Equipment Controller] 装备属性:`, JSON.stringify(equipment.stats, null, 2));
      
      // 检查是否有出战的灵宠
      const Pet = require('../models/Pet');
      const activePet = await Pet.findOne({
        where: {
          userId: req.user.id,
          isActive: true
        }
      });
      
      // 如果有出战的灵宠，先移除灵宠加成
      if (activePet) {
        console.log(`[Equipment Controller] 检测到出战灵宠: ${activePet.name}, 先移除灵宠加成`);
        
        // 移除灵宠的基础属性百分比加成
        const currentAttack = baseAttributes.attack || 0;
        const currentDefense = baseAttributes.defense || 0;
        const currentHealth = baseAttributes.health || 0;
        
        baseAttributes.attack = currentAttack / (1 + (activePet.attackBonus || 0));
        baseAttributes.defense = currentDefense / (1 + (activePet.defenseBonus || 0));
        baseAttributes.health = currentHealth / (1 + (activePet.healthBonus || 0));
        
        // 移除灵宠的combatAttributes加成
        if (activePet.combatAttributes) {
          let petCombatAttributes = activePet.combatAttributes;
          if (typeof petCombatAttributes === 'string') {
            try {
              petCombatAttributes = JSON.parse(petCombatAttributes);
            } catch (e) {
              console.error('解析灵宠combatAttributes失败:', e);
            }
          }
          
          // 移除基础属性
          if (petCombatAttributes.attack) {
            baseAttributes.attack = (baseAttributes.attack || 0) - petCombatAttributes.attack;
          }
          if (petCombatAttributes.defense) {
            baseAttributes.defense = (baseAttributes.defense || 0) - petCombatAttributes.defense;
          }
          if (petCombatAttributes.health) {
            baseAttributes.health = (baseAttributes.health || 0) - petCombatAttributes.health;
          }
          if (petCombatAttributes.speed) {
            baseAttributes.speed = (baseAttributes.speed || 0) - petCombatAttributes.speed;
          }
          
          // 移除战斗属性
          if (petCombatAttributes.critRate !== undefined) {
            combatAttributes.critRate = Math.max(0, (combatAttributes.critRate || 0) - petCombatAttributes.critRate);
          }
          if (petCombatAttributes.comboRate !== undefined) {
            combatAttributes.comboRate = Math.max(0, (combatAttributes.comboRate || 0) - petCombatAttributes.comboRate);
          }
          if (petCombatAttributes.counterRate !== undefined) {
            combatAttributes.counterRate = Math.max(0, (combatAttributes.counterRate || 0) - petCombatAttributes.counterRate);
          }
          if (petCombatAttributes.stunRate !== undefined) {
            combatAttributes.stunRate = Math.max(0, (combatAttributes.stunRate || 0) - petCombatAttributes.stunRate);
          }
          if (petCombatAttributes.dodgeRate !== undefined) {
            combatAttributes.dodgeRate = Math.max(0, (combatAttributes.dodgeRate || 0) - petCombatAttributes.dodgeRate);
          }
          if (petCombatAttributes.vampireRate !== undefined) {
            combatAttributes.vampireRate = Math.max(0, (combatAttributes.vampireRate || 0) - petCombatAttributes.vampireRate);
          }
          
          // 移除战斗抗性
          if (petCombatAttributes.critResist !== undefined) {
            combatResistance.critResist = Math.max(0, (combatResistance.critResist || 0) - petCombatAttributes.critResist);
          }
          if (petCombatAttributes.comboResist !== undefined) {
            combatResistance.comboResist = Math.max(0, (combatResistance.comboResist || 0) - petCombatAttributes.comboResist);
          }
          if (petCombatAttributes.counterResist !== undefined) {
            combatResistance.counterResist = Math.max(0, (combatResistance.counterResist || 0) - petCombatAttributes.counterResist);
          }
          if (petCombatAttributes.stunResist !== undefined) {
            combatResistance.stunResist = Math.max(0, (combatResistance.stunResist || 0) - petCombatAttributes.stunResist);
          }
          if (petCombatAttributes.dodgeResist !== undefined) {
            combatResistance.dodgeResist = Math.max(0, (combatResistance.dodgeResist || 0) - petCombatAttributes.dodgeResist);
          }
          if (petCombatAttributes.vampireResist !== undefined) {
            combatResistance.vampireResist = Math.max(0, (combatResistance.vampireResist || 0) - petCombatAttributes.vampireResist);
          }
          
          // 移除特殊属性
          if (petCombatAttributes.healBoost !== undefined) {
            specialAttributes.healBoost = Math.max(0, (specialAttributes.healBoost || 0) - petCombatAttributes.healBoost);
          }
          if (petCombatAttributes.critDamageBoost !== undefined) {
            specialAttributes.critDamageBoost = Math.max(0, (specialAttributes.critDamageBoost || 0) - petCombatAttributes.critDamageBoost);
          }
          if (petCombatAttributes.critDamageReduce !== undefined) {
            specialAttributes.critDamageReduce = Math.max(0, (specialAttributes.critDamageReduce || 0) - petCombatAttributes.critDamageReduce);
          }
          if (petCombatAttributes.finalDamageBoost !== undefined) {
            specialAttributes.finalDamageBoost = Math.max(0, (specialAttributes.finalDamageBoost || 0) - petCombatAttributes.finalDamageBoost);
          }
          if (petCombatAttributes.finalDamageReduce !== undefined) {
            specialAttributes.finalDamageReduce = Math.max(0, (specialAttributes.finalDamageReduce || 0) - petCombatAttributes.finalDamageReduce);
          }
          if (petCombatAttributes.combatBoost !== undefined) {
            specialAttributes.combatBoost = Math.max(0, (specialAttributes.combatBoost || 0) - petCombatAttributes.combatBoost);
          }
          if (petCombatAttributes.resistanceBoost !== undefined) {
            specialAttributes.resistanceBoost = Math.max(0, (specialAttributes.resistanceBoost || 0) - petCombatAttributes.resistanceBoost);
          }
        }
      }
      
      // 遍历装备属性并添加到玩家属性中
      Object.entries(equipment.stats).forEach(([stat, value]) => {
        if (typeof value === 'number') {
          // 根据属性类型将其添加到相应的属性组中
          if (['attack', 'health', 'defense', 'speed'].includes(stat)) {
            baseAttributes[stat] = (baseAttributes[stat] || 0) + value;
          } else if (['critRate', 'comboRate', 'counterRate', 'stunRate', 'dodgeRate', 'vampireRate'].includes(stat)) {
            combatAttributes[stat] = (combatAttributes[stat] || 0) + value;
          } else if (['critResist', 'comboResist', 'counterResist', 'stunResist', 'dodgeResist', 'vampireResist'].includes(stat)) {
            combatResistance[stat] = (combatResistance[stat] || 0) + value;
          } else {
            // 其他特殊属性
            specialAttributes[stat] = (specialAttributes[stat] || 0) + value;
          }
          console.log(`[Equipment Controller] 属性变化 - ${stat}: ${value}`);
        }
      });
      
      // 如果有出战的灵宠，重新应用灵宠加成
      if (activePet) {
        console.log(`[Equipment Controller] 重新应用出战灵宠加成: ${activePet.name}`);
        
        // 重新应用灵宠的基础属性百分比加成
        baseAttributes.attack = (baseAttributes.attack || 0) * (1 + (activePet.attackBonus || 0));
        baseAttributes.defense = (baseAttributes.defense || 0) * (1 + (activePet.defenseBonus || 0));
        baseAttributes.health = (baseAttributes.health || 0) * (1 + (activePet.healthBonus || 0));
        
        // 应用灵宠的combatAttributes加成
        if (activePet.combatAttributes) {
          let petCombatAttributes = activePet.combatAttributes;
          if (typeof petCombatAttributes === 'string') {
            try {
              petCombatAttributes = JSON.parse(petCombatAttributes);
            } catch (e) {
              console.error('解析灵宠combatAttributes失败:', e);
            }
          }
          
          // 累加基础属性
          if (petCombatAttributes.attack) {
            baseAttributes.attack = (baseAttributes.attack || 0) + petCombatAttributes.attack;
          }
          if (petCombatAttributes.defense) {
            baseAttributes.defense = (baseAttributes.defense || 0) + petCombatAttributes.defense;
          }
          if (petCombatAttributes.health) {
            baseAttributes.health = (baseAttributes.health || 0) + petCombatAttributes.health;
          }
          if (petCombatAttributes.speed) {
            baseAttributes.speed = (baseAttributes.speed || 0) + petCombatAttributes.speed;
          }
          
          // 累加战斗属性
          if (petCombatAttributes.critRate !== undefined) {
            combatAttributes.critRate = Math.min(1, (combatAttributes.critRate || 0) + petCombatAttributes.critRate);
          }
          if (petCombatAttributes.comboRate !== undefined) {
            combatAttributes.comboRate = Math.min(1, (combatAttributes.comboRate || 0) + petCombatAttributes.comboRate);
          }
          if (petCombatAttributes.counterRate !== undefined) {
            combatAttributes.counterRate = Math.min(1, (combatAttributes.counterRate || 0) + petCombatAttributes.counterRate);
          }
          if (petCombatAttributes.stunRate !== undefined) {
            combatAttributes.stunRate = Math.min(1, (combatAttributes.stunRate || 0) + petCombatAttributes.stunRate);
          }
          if (petCombatAttributes.dodgeRate !== undefined) {
            combatAttributes.dodgeRate = Math.min(1, (combatAttributes.dodgeRate || 0) + petCombatAttributes.dodgeRate);
          }
          if (petCombatAttributes.vampireRate !== undefined) {
            combatAttributes.vampireRate = Math.min(1, (combatAttributes.vampireRate || 0) + petCombatAttributes.vampireRate);
          }
          
          // 累加战斗抗性
          if (petCombatAttributes.critResist !== undefined) {
            combatResistance.critResist = Math.min(1, (combatResistance.critResist || 0) + petCombatAttributes.critResist);
          }
          if (petCombatAttributes.comboResist !== undefined) {
            combatResistance.comboResist = Math.min(1, (combatResistance.comboResist || 0) + petCombatAttributes.comboResist);
          }
          if (petCombatAttributes.counterResist !== undefined) {
            combatResistance.counterResist = Math.min(1, (combatResistance.counterResist || 0) + petCombatAttributes.counterResist);
          }
          if (petCombatAttributes.stunResist !== undefined) {
            combatResistance.stunResist = Math.min(1, (combatResistance.stunResist || 0) + petCombatAttributes.stunResist);
          }
          if (petCombatAttributes.dodgeResist !== undefined) {
            combatResistance.dodgeResist = Math.min(1, (combatResistance.dodgeResist || 0) + petCombatAttributes.dodgeResist);
          }
          if (petCombatAttributes.vampireResist !== undefined) {
            combatResistance.vampireResist = Math.min(1, (combatResistance.vampireResist || 0) + petCombatAttributes.vampireResist);
          }
          
          // 累加特殊属性
          if (petCombatAttributes.healBoost !== undefined) {
            specialAttributes.healBoost = (specialAttributes.healBoost || 0) + petCombatAttributes.healBoost;
          }
          if (petCombatAttributes.critDamageBoost !== undefined) {
            specialAttributes.critDamageBoost = (specialAttributes.critDamageBoost || 0) + petCombatAttributes.critDamageBoost;
          }
          if (petCombatAttributes.critDamageReduce !== undefined) {
            specialAttributes.critDamageReduce = (specialAttributes.critDamageReduce || 0) + petCombatAttributes.critDamageReduce;
          }
          if (petCombatAttributes.finalDamageBoost !== undefined) {
            specialAttributes.finalDamageBoost = (specialAttributes.finalDamageBoost || 0) + petCombatAttributes.finalDamageBoost;
          }
          if (petCombatAttributes.finalDamageReduce !== undefined) {
            specialAttributes.finalDamageReduce = (specialAttributes.finalDamageReduce || 0) + petCombatAttributes.finalDamageReduce;
          }
          if (petCombatAttributes.combatBoost !== undefined) {
            specialAttributes.combatBoost = (specialAttributes.combatBoost || 0) + petCombatAttributes.combatBoost;
          }
          if (petCombatAttributes.resistanceBoost !== undefined) {
            specialAttributes.resistanceBoost = (specialAttributes.resistanceBoost || 0) + petCombatAttributes.resistanceBoost;
          }
        }
      }
      
      console.log(`[Equipment Controller] 装备穿戴后玩家属性:`, JSON.stringify({
        baseAttributes,
        combatAttributes,
        combatResistance,
        specialAttributes
      }, null, 2));
      
      // 保存更新后的属性
      await User.update(
        {
          baseAttributes,
          combatAttributes,
          combatResistance,
          specialAttributes
        },
        { where: { id: req.user.id } }
      );
      
      // 保存更新后的属性供后续使用
      updatedBaseAttributes = baseAttributes;
      updatedCombatAttributes = combatAttributes;
      updatedCombatResistance = combatResistance;
      updatedSpecialAttributes = specialAttributes;
    }
    
    console.log(`[Equipment Controller] 装备穿戴成功: ${equipment.name}, 新装备状态: equipped=${equipment.equipped}, slot=${equipment.slot}, equipType=${equipment.equipType}`);
    
    // 重新加载装备信息以确保状态是最新的
    const updatedEquipment = await Equipment.findByPk(equipment.id);
    
    // 确保返回的装备信息包含所有必要字段
    console.log(`[Equipment Controller] 装备穿戴完成后最终装备信息:`, JSON.stringify(updatedEquipment, null, 2));
    
    res.json({
      success: true,
      message: '装备穿戴成功',
      equipment: updatedEquipment,
      user: {
        baseAttributes: updatedBaseAttributes,
        combatAttributes: updatedCombatAttributes,
        combatResistance: updatedCombatResistance,
        specialAttributes: updatedSpecialAttributes
      }
    });
  } catch (error) {
    console.error('装备穿戴失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 装备卸下
const unequipEquipment = async (req, res) => {
  try {
    const { id } = req.params;
    
    console.log(`[Equipment Controller] 装备卸下请求: 装备ID=${id}, 用户ID=${req.user.id}`);
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    console.log(`[Equipment Controller] 查找到的装备信息:`, JSON.stringify(equipment, null, 2));
    
    if (!equipment) {
      console.log(`[Equipment Controller] 装备未找到: 装备ID=${id}, 用户ID=${req.user.id}`);
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    console.log(`[Equipment Controller] 找到装备: ${equipment.name}, 当前装备状态: equipped=${equipment.equipped}, slot=${equipment.slot}`);
    // 添加装备卸下前的详细信息日志
    console.log(`[Equipment Controller] 装备卸下前的装备详细信息:`, JSON.stringify(equipment, null, 2));
    
    // 解析stats字段
    if (equipment.stats && typeof equipment.stats === 'string') {
      try {
        equipment.stats = JSON.parse(equipment.stats);
        console.log(`[Equipment Controller] 成功解析装备stats:`, equipment.stats);
      } catch (e) {
        console.error('[Equipment Controller] 解析装备stats失败:', e);
      }
    }
    
    console.log(`[Equipment Controller] 处理后的装备stats:`, JSON.stringify(equipment.stats, null, 2));
    
    // 检查装备是否处于装备状态
    if (!equipment.equipped) {
      console.log(`[Equipment Controller] 装备未处于装备状态，无法卸下: ${equipment.name}`);
      return res.status(400).json({
        success: false,
        message: '装备未处于装备状态，无法卸下'
      });
    }
    
    // 检查装备是否有有效的槽位
    if (!equipment.slot && !equipment.equipType) {
      console.log(`[Equipment Controller] 装备槽位信息缺失，无法卸下: ${equipment.name}`);
      return res.status(400).json({
        success: false,
        message: '装备槽位信息缺失，无法卸下'
      });
    }
    
    // 更新装备状态为未装备
    equipment.equipped = false;
    equipment.slot = null;
    // 保留 equipType 字段，不将其设置为 null
    await equipment.save();
    
    console.log(`[Equipment Controller] 装备状态更新完成，equipped=${equipment.equipped}, slot=${equipment.slot}`);
    
    // 移除装备属性加成
    const User = require('../models/User');
    const user = await User.findByPk(req.user.id);
    
    console.log(`[Equipment Controller] 装备卸下前玩家属性:`, JSON.stringify({
      baseAttributes: user.baseAttributes,
      combatAttributes: user.combatAttributes,
      combatResistance: user.combatResistance,
      specialAttributes: user.specialAttributes
    }, null, 2));
    
    let updatedBaseAttributes = user.baseAttributes;
    let updatedCombatAttributes = user.combatAttributes;
    let updatedCombatResistance = user.combatResistance;
    let updatedSpecialAttributes = user.specialAttributes;
    
    console.log(`[Equipment Controller] 检查装备stats:`, equipment.stats);
    if (equipment.stats) {
      console.log(`[Equipment Controller] 装备stats存在且为:`, typeof equipment.stats);
      // 克隆当前属性以避免直接修改
      let baseAttributes = JSON.parse(JSON.stringify(user.baseAttributes || {}));
      let combatAttributes = JSON.parse(JSON.stringify(user.combatAttributes || {}));
      let combatResistance = JSON.parse(JSON.stringify(user.combatResistance || {}));
      let specialAttributes = JSON.parse(JSON.stringify(user.specialAttributes || {}));
      
      console.log(`[Equipment Controller] 装备属性:`, JSON.stringify(equipment.stats, null, 2));
      
      // 检查是否有出战的灵宠
      const Pet = require('../models/Pet');
      const activePet = await Pet.findOne({
        where: {
          userId: req.user.id,
          isActive: true
        }
      });
      
      // 如果有出战的灵宠，先移除灵宠加成
      if (activePet) {
        console.log(`[Equipment Controller] 检测到出战灵宠: ${activePet.name}, 先移除灵宠加成`);
        
        // 移除灵宠的基础属性百分比加成
        const currentAttack = baseAttributes.attack || 0;
        const currentDefense = baseAttributes.defense || 0;
        const currentHealth = baseAttributes.health || 0;
        
        baseAttributes.attack = currentAttack / (1 + (activePet.attackBonus || 0));
        baseAttributes.defense = currentDefense / (1 + (activePet.defenseBonus || 0));
        baseAttributes.health = currentHealth / (1 + (activePet.healthBonus || 0));
        
        // 移除灵宠的combatAttributes加成
        if (activePet.combatAttributes) {
          let petCombatAttributes = activePet.combatAttributes;
          if (typeof petCombatAttributes === 'string') {
            try {
              petCombatAttributes = JSON.parse(petCombatAttributes);
            } catch (e) {
              console.error('解析灵宠combatAttributes失败:', e);
            }
          }
          
          // 移除基础属性
          if (petCombatAttributes.attack) {
            baseAttributes.attack = (baseAttributes.attack || 0) - petCombatAttributes.attack;
          }
          if (petCombatAttributes.defense) {
            baseAttributes.defense = (baseAttributes.defense || 0) - petCombatAttributes.defense;
          }
          if (petCombatAttributes.health) {
            baseAttributes.health = (baseAttributes.health || 0) - petCombatAttributes.health;
          }
          if (petCombatAttributes.speed) {
            baseAttributes.speed = (baseAttributes.speed || 0) - petCombatAttributes.speed;
          }
          
          // 移除战斗属性
          if (petCombatAttributes.critRate !== undefined) {
            combatAttributes.critRate = Math.max(0, (combatAttributes.critRate || 0) - petCombatAttributes.critRate);
          }
          if (petCombatAttributes.comboRate !== undefined) {
            combatAttributes.comboRate = Math.max(0, (combatAttributes.comboRate || 0) - petCombatAttributes.comboRate);
          }
          if (petCombatAttributes.counterRate !== undefined) {
            combatAttributes.counterRate = Math.max(0, (combatAttributes.counterRate || 0) - petCombatAttributes.counterRate);
          }
          if (petCombatAttributes.stunRate !== undefined) {
            combatAttributes.stunRate = Math.max(0, (combatAttributes.stunRate || 0) - petCombatAttributes.stunRate);
          }
          if (petCombatAttributes.dodgeRate !== undefined) {
            combatAttributes.dodgeRate = Math.max(0, (combatAttributes.dodgeRate || 0) - petCombatAttributes.dodgeRate);
          }
          if (petCombatAttributes.vampireRate !== undefined) {
            combatAttributes.vampireRate = Math.max(0, (combatAttributes.vampireRate || 0) - petCombatAttributes.vampireRate);
          }
          
          // 移除战斗抗性
          if (petCombatAttributes.critResist !== undefined) {
            combatResistance.critResist = Math.max(0, (combatResistance.critResist || 0) - petCombatAttributes.critResist);
          }
          if (petCombatAttributes.comboResist !== undefined) {
            combatResistance.comboResist = Math.max(0, (combatResistance.comboResist || 0) - petCombatAttributes.comboResist);
          }
          if (petCombatAttributes.counterResist !== undefined) {
            combatResistance.counterResist = Math.max(0, (combatResistance.counterResist || 0) - petCombatAttributes.counterResist);
          }
          if (petCombatAttributes.stunResist !== undefined) {
            combatResistance.stunResist = Math.max(0, (combatResistance.stunResist || 0) - petCombatAttributes.stunResist);
          }
          if (petCombatAttributes.dodgeResist !== undefined) {
            combatResistance.dodgeResist = Math.max(0, (combatResistance.dodgeResist || 0) - petCombatAttributes.dodgeResist);
          }
          if (petCombatAttributes.vampireResist !== undefined) {
            combatResistance.vampireResist = Math.max(0, (combatResistance.vampireResist || 0) - petCombatAttributes.vampireResist);
          }
          
          // 移除特殊属性
          if (petCombatAttributes.healBoost !== undefined) {
            specialAttributes.healBoost = Math.max(0, (specialAttributes.healBoost || 0) - petCombatAttributes.healBoost);
          }
          if (petCombatAttributes.critDamageBoost !== undefined) {
            specialAttributes.critDamageBoost = Math.max(0, (specialAttributes.critDamageBoost || 0) - petCombatAttributes.critDamageBoost);
          }
          if (petCombatAttributes.critDamageReduce !== undefined) {
            specialAttributes.critDamageReduce = Math.max(0, (specialAttributes.critDamageReduce || 0) - petCombatAttributes.critDamageReduce);
          }
          if (petCombatAttributes.finalDamageBoost !== undefined) {
            specialAttributes.finalDamageBoost = Math.max(0, (specialAttributes.finalDamageBoost || 0) - petCombatAttributes.finalDamageBoost);
          }
          if (petCombatAttributes.finalDamageReduce !== undefined) {
            specialAttributes.finalDamageReduce = Math.max(0, (specialAttributes.finalDamageReduce || 0) - petCombatAttributes.finalDamageReduce);
          }
          if (petCombatAttributes.combatBoost !== undefined) {
            specialAttributes.combatBoost = Math.max(0, (specialAttributes.combatBoost || 0) - petCombatAttributes.combatBoost);
          }
          if (petCombatAttributes.resistanceBoost !== undefined) {
            specialAttributes.resistanceBoost = Math.max(0, (specialAttributes.resistanceBoost || 0) - petCombatAttributes.resistanceBoost);
          }
        }
      }
      
      // 遍历装备属性并从玩家属性中减去
      Object.entries(equipment.stats).forEach(([stat, value]) => {
        if (typeof value === 'number') {
          // 根据属性类型将其从相应的属性组中减去
          if (['attack', 'health', 'defense', 'speed'].includes(stat)) {
            baseAttributes[stat] = Math.max(0, (baseAttributes[stat] || 0) - value);
          } else if (['critRate', 'comboRate', 'counterRate', 'stunRate', 'dodgeRate', 'vampireRate'].includes(stat)) {
            combatAttributes[stat] = Math.max(0, (combatAttributes[stat] || 0) - value);
          } else if (['critResist', 'comboResist', 'counterResist', 'stunResist', 'dodgeResist', 'vampireResist'].includes(stat)) {
            combatResistance[stat] = Math.max(0, (combatResistance[stat] || 0) - value);
          } else {
            // 其他特殊属性
            specialAttributes[stat] = Math.max(0, (specialAttributes[stat] || 0) - value);
          }
          console.log(`[Equipment Controller] 属性变化 - ${stat}: -${value}`);
        }
      });
      
      // 如果有出战的灵宠，重新应用灵宠加成
      if (activePet) {
        console.log(`[Equipment Controller] 重新应用出战灵宠加成: ${activePet.name}`);
        
        // 重新应用灵宠的基础属性百分比加成
        baseAttributes.attack = (baseAttributes.attack || 0) * (1 + (activePet.attackBonus || 0));
        baseAttributes.defense = (baseAttributes.defense || 0) * (1 + (activePet.defenseBonus || 0));
        baseAttributes.health = (baseAttributes.health || 0) * (1 + (activePet.healthBonus || 0));
        
        // 应用灵宠的combatAttributes加成
        if (activePet.combatAttributes) {
          let petCombatAttributes = activePet.combatAttributes;
          if (typeof petCombatAttributes === 'string') {
            try {
              petCombatAttributes = JSON.parse(petCombatAttributes);
            } catch (e) {
              console.error('解析灵宠combatAttributes失败:', e);
            }
          }
          
          // 累加基础属性
          if (petCombatAttributes.attack) {
            baseAttributes.attack = (baseAttributes.attack || 0) + petCombatAttributes.attack;
          }
          if (petCombatAttributes.defense) {
            baseAttributes.defense = (baseAttributes.defense || 0) + petCombatAttributes.defense;
          }
          if (petCombatAttributes.health) {
            baseAttributes.health = (baseAttributes.health || 0) + petCombatAttributes.health;
          }
          if (petCombatAttributes.speed) {
            baseAttributes.speed = (baseAttributes.speed || 0) + petCombatAttributes.speed;
          }
          
          // 累加战斗属性
          if (petCombatAttributes.critRate !== undefined) {
            combatAttributes.critRate = Math.min(1, (combatAttributes.critRate || 0) + petCombatAttributes.critRate);
          }
          if (petCombatAttributes.comboRate !== undefined) {
            combatAttributes.comboRate = Math.min(1, (combatAttributes.comboRate || 0) + petCombatAttributes.comboRate);
          }
          if (petCombatAttributes.counterRate !== undefined) {
            combatAttributes.counterRate = Math.min(1, (combatAttributes.counterRate || 0) + petCombatAttributes.counterRate);
          }
          if (petCombatAttributes.stunRate !== undefined) {
            combatAttributes.stunRate = Math.min(1, (combatAttributes.stunRate || 0) + petCombatAttributes.stunRate);
          }
          if (petCombatAttributes.dodgeRate !== undefined) {
            combatAttributes.dodgeRate = Math.min(1, (combatAttributes.dodgeRate || 0) + petCombatAttributes.dodgeRate);
          }
          if (petCombatAttributes.vampireRate !== undefined) {
            combatAttributes.vampireRate = Math.min(1, (combatAttributes.vampireRate || 0) + petCombatAttributes.vampireRate);
          }
          
          // 累加战斗抗性
          if (petCombatAttributes.critResist !== undefined) {
            combatResistance.critResist = Math.min(1, (combatResistance.critResist || 0) + petCombatAttributes.critResist);
          }
          if (petCombatAttributes.comboResist !== undefined) {
            combatResistance.comboResist = Math.min(1, (combatResistance.comboResist || 0) + petCombatAttributes.comboResist);
          }
          if (petCombatAttributes.counterResist !== undefined) {
            combatResistance.counterResist = Math.min(1, (combatResistance.counterResist || 0) + petCombatAttributes.counterResist);
          }
          if (petCombatAttributes.stunResist !== undefined) {
            combatResistance.stunResist = Math.min(1, (combatResistance.stunResist || 0) + petCombatAttributes.stunResist);
          }
          if (petCombatAttributes.dodgeResist !== undefined) {
            combatResistance.dodgeResist = Math.min(1, (combatResistance.dodgeResist || 0) + petCombatAttributes.dodgeResist);
          }
          if (petCombatAttributes.vampireResist !== undefined) {
            combatResistance.vampireResist = Math.min(1, (combatResistance.vampireResist || 0) + petCombatAttributes.vampireResist);
          }
          
          // 累加特殊属性
          if (petCombatAttributes.healBoost !== undefined) {
            specialAttributes.healBoost = (specialAttributes.healBoost || 0) + petCombatAttributes.healBoost;
          }
          if (petCombatAttributes.critDamageBoost !== undefined) {
            specialAttributes.critDamageBoost = (specialAttributes.critDamageBoost || 0) + petCombatAttributes.critDamageBoost;
          }
          if (petCombatAttributes.critDamageReduce !== undefined) {
            specialAttributes.critDamageReduce = (specialAttributes.critDamageReduce || 0) + petCombatAttributes.critDamageReduce;
          }
          if (petCombatAttributes.finalDamageBoost !== undefined) {
            specialAttributes.finalDamageBoost = (specialAttributes.finalDamageBoost || 0) + petCombatAttributes.finalDamageBoost;
          }
          if (petCombatAttributes.finalDamageReduce !== undefined) {
            specialAttributes.finalDamageReduce = (specialAttributes.finalDamageReduce || 0) + petCombatAttributes.finalDamageReduce;
          }
          if (petCombatAttributes.combatBoost !== undefined) {
            specialAttributes.combatBoost = (specialAttributes.combatBoost || 0) + petCombatAttributes.combatBoost;
          }
          if (petCombatAttributes.resistanceBoost !== undefined) {
            specialAttributes.resistanceBoost = (specialAttributes.resistanceBoost || 0) + petCombatAttributes.resistanceBoost;
          }
        }
      }
      
      console.log(`[Equipment Controller] 装备卸下后玩家属性:`, JSON.stringify({
        baseAttributes,
        combatAttributes,
        combatResistance,
        specialAttributes
      }, null, 2));
      
      // 保存更新后的属性
      await User.update(
        {
          baseAttributes,
          combatAttributes,
          combatResistance,
          specialAttributes
        },
        { where: { id: req.user.id } }
      );
      
      // 保存更新后的属性供后续使用
      updatedBaseAttributes = baseAttributes;
      updatedCombatAttributes = combatAttributes;
      updatedCombatResistance = combatResistance;
      updatedSpecialAttributes = specialAttributes;
    } else {
      console.log(`[Equipment Controller] 装备stats不存在或为空`);
    }
    
    console.log(`[Equipment Controller] 装备卸下成功: ${equipment.name}, 新装备状态: equipped=${equipment.equipped}, slot=${equipment.slot}, equipType=${equipment.equipType}`);
    
    // 重新加载装备信息以确保状态是最新的
    const updatedEquipment = await Equipment.findByPk(equipment.id);
    
    console.log(`[Equipment Controller] 返回给前端的用户属性:`, JSON.stringify({
      baseAttributes: updatedBaseAttributes,
      combatAttributes: updatedCombatAttributes,
      combatResistance: updatedCombatResistance,
      specialAttributes: updatedSpecialAttributes
    }, null, 2));
    
    res.json({
      success: true,
      message: '装备卸下成功',
      equipment: updatedEquipment,
      user: {
        baseAttributes: updatedBaseAttributes,
        combatAttributes: updatedCombatAttributes,
        combatResistance: updatedCombatResistance,
        specialAttributes: updatedSpecialAttributes
      }
    });
  } catch (error) {
    console.error('装备卸下失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 出售装备
const sellEquipment = async (req, res) => {
  try {
    const { id } = req.params;
    
    // 查找装备
    const equipment = await Equipment.findOne({
      where: {
        id: id,
        userId: req.user.id
      }
    });
    
    if (!equipment) {
      return res.status(404).json({
        success: false,
        message: '装备未找到'
      });
    }
    
    // 删除装备
    await equipment.destroy();
    
    // 根据装备品质计算返还的资源数量
    const qualityStoneMap = {
      mythic: 6,
      legendary: 5,
      epic: 4,
      rare: 3,
      uncommon: 2,
      common: 1
    };
    
    const stonesReceived = qualityStoneMap[equipment.quality] || 1;
    
    res.json({
      success: true,
      message: '装备出售成功',
      stonesReceived: stonesReceived
    });
  } catch (error) {
    console.error('装备出售失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

// 批量出售装备
const batchSellEquipment = async (req, res) => {
  try {
    const { quality, type } = req.body;
    
    // 构建查询条件
    const whereConditions = {
      userId: req.user.id
    };
    
    // 如果提供了品质过滤条件
    if (quality) {
      whereConditions.quality = quality;
    }
    
    // 如果提供了类型过滤条件
    if (type) {
      whereConditions.equipType = type;
    }
    
    // 查找符合条件的装备
    const equipmentList = await Equipment.findAll({
      where: whereConditions
    });
    
    if (equipmentList.length === 0) {
      return res.status(404).json({
        success: false,
        message: '没有找到符合条件的装备'
      });
    }
    
    // 计算返还的资源总数
    const qualityStoneMap = {
      mythic: 6,
      legendary: 5,
      epic: 4,
      rare: 3,
      uncommon: 2,
      common: 1
    };
    
    let totalStonesReceived = 0;
    equipmentList.forEach(eq => {
      totalStonesReceived += qualityStoneMap[eq.quality] || 1;
    });
    
    // 删除所有符合条件的装备
    const equipmentIds = equipmentList.map(eq => eq.id);
    await Equipment.destroy({
      where: {
        id: {
          [Sequelize.Op.in]: equipmentIds
        }
      }
    });
    
    res.json({
      success: true,
      message: `成功出售${equipmentList.length}件装备`,
      equipmentSold: equipmentList.length,
      stonesReceived: totalStonesReceived
    });
  } catch (error) {
    console.error('批量出售装备失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

module.exports = {
  getPlayerEquipment,
  getEquipmentDetails,
  enhanceEquipment,
  reforgeEquipment,
  confirmReforge,
  equipEquipment,
  unequipEquipment,
  sellEquipment,
  batchSellEquipment
};