const { Sequelize } = require('sequelize');
const User = require('../models/User');
const Pet = require('../models/Pet');

// 计算灵宠属性加成
const calculatePetBonus = (pet) => {
  if (!pet) return null;

  // 品质基础加成映射
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

  const baseBonus = qualityBonusMap[pet.rarity] || 0;
  const starBonus = (pet.star || 0) * (starBonusPerQuality[pet.rarity] || 0);
  const levelBonus = ((pet.level || 1) - 1) * (baseBonus * 0.1);
  const phase = Math.floor((pet.star || 0) / 5);
  const phaseBonus = phase * (baseBonus * 0.5);
  const finalBonus = baseBonus + starBonus + levelBonus + phaseBonus;
  const combatBonus = finalBonus * 0.5;

  return {
    attack: finalBonus,
    defense: finalBonus,
    health: finalBonus,
    critRate: 0,
    comboRate: 0,
    counterRate: 0,
    stunRate: 0,
    dodgeRate: 0,
    vampireRate: 0,
    critResist: 0,
    comboResist: 0,
    counterResist: 0,
    stunResist: 0,
    dodgeResist: 0,
    vampireResist: 0,
    healBoost: 0,
    critDamageBoost: 0,
    critDamageReduce: 0,
    finalDamageBoost: 0,
    finalDamageReduce: 0,
    combatBoost: 0,
    resistanceBoost: 0
  };
};

// 出战灵宠
const deployPet = async (req, res) => {
  try {
    const { id: petId } = req.params;
    const userId = req.user.id;

    // 查找要出战的灵宠
    let pet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: petId },
          { id: petId }
        ]
      }
    });

    if (!pet) {
      return res.status(404).json({ message: '灵宠未找到' });
    }

    // 打印出战前灵宠属性
    console.log(`[Pet Deploy] 出战前灵宠属性:`, {
      petId: pet.id,
      petName: pet.name,
      petRarity: pet.rarity,
      petLevel: pet.level,
      petStar: pet.star,
      petDescription: pet.description,
      petCombatAttributes: pet.combatAttributes,
      petAttackBonus: pet.attackBonus,
      petDefenseBonus: pet.defenseBonus,
      petHealthBonus: pet.healthBonus,
      petIsActive: pet.isActive
    });

    // 检查是否已经有出战的灵宠
    const activePet = await Pet.findOne({
      where: {
        userId: userId,
        isActive: true
      }
    });
    
    // 如果已经有出战的灵宠，先将其召回
    if (activePet && activePet.id !== pet.id) {
      console.log(`召回当前出战的灵宠: ${activePet.name}(${activePet.id})`);
      await Pet.update(
        { isActive: false },
        {
          where: {
            userId: userId,
            isActive: true
          }
        }
      );
    } else if (activePet && activePet.id === pet.id) {
      // 如果要出战的灵宠已经出战，直接返回成功
      console.log(`灵宠 ${pet.name}(${petId}) 已经处于出战状态`);
      return res.status(200).send({ success: true, message: '灵宠已经出战' });
    }

    // 标记目标灵宠为出战状态
    const [updatedRows] = await Pet.update(
      { isActive: true },
      {
        where: {
          userId: userId,
          [Sequelize.Op.or]: [
            { petId: petId },
            { id: petId }
          ]
        }
      }
    );
      
    // 检查更新是否成功
    if (updatedRows === 0) {
      console.log(`未能成功出战灵宠: ${pet.name}(${petId})`);
      return res.status(500).json({ message: '未能成功出战灵宠' });
    }
      
    console.log(`成功出战灵宠: ${pet.name}(${petId})`);

    

    // 重新查询更新后的灵宠对象，确保获取最新状态
    const updatedPet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: petId },
          { id: petId }
        ]
      }
    });
    
    if (!updatedPet) {
      console.log(`未能找到更新后的灵宠: ${pet.name}(${petId})`);
      return res.status(500).json({ message: '未能找到更新后的灵宠' });
    }
    
    // 用更新后的灵宠对象替换原有对象
    pet = updatedPet;

    // 计算灵宠属性加成
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

    const baseBonus = qualityBonusMap[pet.rarity] || 0;
    const starBonus = (pet.star || 0) * (starBonusPerQuality[pet.rarity] || 0);
    const levelBonus = ((pet.level || 1) - 1) * (baseBonus * 0.1);
    const phase = Math.floor((pet.star || 0) / 5);
    const phaseBonus = phase * (baseBonus * 0.5);
    const finalBonus = baseBonus + starBonus + levelBonus + phaseBonus;
    
    // 使用预定义的攻击、防御、生命加成值
    const petBonus = {
      attack: pet.attackBonus || 0,
      defense: pet.defenseBonus || 0,
      health: pet.healthBonus || 0
    };
    
    // 打印出战后灵宠属性
    console.log(`[Pet Deploy] 出战后灵宠属性:`, {
      petId: pet.id,
      petName: pet.name,
      petRarity: pet.rarity,
      petLevel: pet.level,
      petStar: pet.star,
      petDescription: pet.description,
      petCombatAttributes: pet.combatAttributes,
      petAttackBonus: pet.attackBonus,
      petDefenseBonus: pet.defenseBonus,
      petHealthBonus: pet.healthBonus,
      petIsActive: pet.isActive
    });

    // 获取用户当前属性
    const user = await User.findByPk(userId);
    if (user) {
      // 打印出战前玩家属性
      console.log(`[Pet Deploy] 出战前玩家属性:`, {
        baseAttributes: user.baseAttributes,
        combatAttributes: user.combatAttributes,
        combatResistance: user.combatResistance,
        specialAttributes: user.specialAttributes
      });
      
      // 应用属性加成到用户数据
      const updatedBaseAttributes = { ...user.baseAttributes };
      const updatedCombatAttributes = { ...user.combatAttributes };
      const updatedCombatResistance = { ...user.combatResistance };
      const updatedSpecialAttributes = { ...user.specialAttributes };

      // 先将灵宠的combatAttributes中的字段属性累加到玩家属性中的相同字段
      if (pet.combatAttributes) {
        // 解析combatAttributes（如果是字符串）
        let combatAttributes = pet.combatAttributes;
        if (typeof combatAttributes === 'string') {
          try {
            combatAttributes = JSON.parse(combatAttributes);
          } catch (e) {
            console.error('解析combatAttributes失败:', e);
          }
        }
        
        // 累加基础属性
        if (combatAttributes.attack) {
          updatedBaseAttributes.attack = (updatedBaseAttributes.attack || 0) + combatAttributes.attack;
        }
        if (combatAttributes.defense) {
          updatedBaseAttributes.defense = (updatedBaseAttributes.defense || 0) + combatAttributes.defense;
        }
        if (combatAttributes.health) {
          updatedBaseAttributes.health = (updatedBaseAttributes.health || 0) + combatAttributes.health;
        }
        if (combatAttributes.speed) {
          updatedBaseAttributes.speed = (updatedBaseAttributes.speed || 0) + combatAttributes.speed;
        }
        
        // 累加战斗属性
        if (combatAttributes.critRate !== undefined) {
          updatedCombatAttributes.critRate = Math.min(1, (updatedCombatAttributes.critRate || 0) + combatAttributes.critRate);
        }
        if (combatAttributes.comboRate !== undefined) {
          updatedCombatAttributes.comboRate = Math.min(1, (updatedCombatAttributes.comboRate || 0) + combatAttributes.comboRate);
        }
        if (combatAttributes.counterRate !== undefined) {
          updatedCombatAttributes.counterRate = Math.min(1, (updatedCombatAttributes.counterRate || 0) + combatAttributes.counterRate);
        }
        if (combatAttributes.stunRate !== undefined) {
          updatedCombatAttributes.stunRate = Math.min(1, (updatedCombatAttributes.stunRate || 0) + combatAttributes.stunRate);
        }
        if (combatAttributes.dodgeRate !== undefined) {
          updatedCombatAttributes.dodgeRate = Math.min(1, (updatedCombatAttributes.dodgeRate || 0) + combatAttributes.dodgeRate);
        }
        if (combatAttributes.vampireRate !== undefined) {
          updatedCombatAttributes.vampireRate = Math.min(1, (updatedCombatAttributes.vampireRate || 0) + combatAttributes.vampireRate);
        }
        
        // 累加战斗抗性
        if (combatAttributes.critResist !== undefined) {
          updatedCombatResistance.critResist = Math.min(1, (updatedCombatResistance.critResist || 0) + combatAttributes.critResist);
        }
        if (combatAttributes.comboResist !== undefined) {
          updatedCombatResistance.comboResist = Math.min(1, (updatedCombatResistance.comboResist || 0) + combatAttributes.comboResist);
        }
        if (combatAttributes.counterResist !== undefined) {
          updatedCombatResistance.counterResist = Math.min(1, (updatedCombatResistance.counterResist || 0) + combatAttributes.counterResist);
        }
        if (combatAttributes.stunResist !== undefined) {
          updatedCombatResistance.stunResist = Math.min(1, (updatedCombatResistance.stunResist || 0) + combatAttributes.stunResist);
        }
        if (combatAttributes.dodgeResist !== undefined) {
          updatedCombatResistance.dodgeResist = Math.min(1, (updatedCombatResistance.dodgeResist || 0) + combatAttributes.dodgeResist);
        }
        if (combatAttributes.vampireResist !== undefined) {
          updatedCombatResistance.vampireResist = Math.min(1, (updatedCombatResistance.vampireResist || 0) + combatAttributes.vampireResist);
        }
        
        // 累加特殊属性
        if (combatAttributes.healBoost !== undefined) {
          updatedSpecialAttributes.healBoost = (updatedSpecialAttributes.healBoost || 0) + combatAttributes.healBoost;
        }
        if (combatAttributes.critDamageBoost !== undefined) {
          updatedSpecialAttributes.critDamageBoost = (updatedSpecialAttributes.critDamageBoost || 0) + combatAttributes.critDamageBoost;
        }
        if (combatAttributes.critDamageReduce !== undefined) {
          updatedSpecialAttributes.critDamageReduce = (updatedSpecialAttributes.critDamageReduce || 0) + combatAttributes.critDamageReduce;
        }
        if (combatAttributes.finalDamageBoost !== undefined) {
          updatedSpecialAttributes.finalDamageBoost = (updatedSpecialAttributes.finalDamageBoost || 0) + combatAttributes.finalDamageBoost;
        }
        if (combatAttributes.finalDamageReduce !== undefined) {
          updatedSpecialAttributes.finalDamageReduce = (updatedSpecialAttributes.finalDamageReduce || 0) + combatAttributes.finalDamageReduce;
        }
        if (combatAttributes.combatBoost !== undefined) {
          updatedSpecialAttributes.combatBoost = (updatedSpecialAttributes.combatBoost || 0) + combatAttributes.combatBoost;
        }
        if (combatAttributes.resistanceBoost !== undefined) {
          updatedSpecialAttributes.resistanceBoost = (updatedSpecialAttributes.resistanceBoost || 0) + combatAttributes.resistanceBoost;
        }
      }

      // 再应用基础属性百分比加成
      updatedBaseAttributes.attack = (updatedBaseAttributes.attack || 0) * (1 + (pet.attackBonus || 0));
      updatedBaseAttributes.defense = (updatedBaseAttributes.defense || 0) * (1 + (pet.defenseBonus || 0));
      updatedBaseAttributes.health = (updatedBaseAttributes.health || 0) * (1 + (pet.healthBonus || 0));

      // 打印出战后玩家属性
      console.log(`[Pet Deploy] 出战后玩家属性:`, {
        updatedBaseAttributes,
        updatedCombatAttributes,
        updatedCombatResistance,
        updatedSpecialAttributes
      });

      // 更新用户数据
      await User.update(
        {
          baseAttributes: updatedBaseAttributes,
          combatAttributes: updatedCombatAttributes,
          combatResistance: updatedCombatResistance,
          specialAttributes: updatedSpecialAttributes
        },
        {
          where: { id: userId }
        }
      );
    }

    res.status(200).send({ 
      success: true, 
      message: '出战成功', 
      pet: { 
        id: pet.id, 
        name: pet.name, 
        isActive: pet.isActive 
      } 
    });
  } catch (error) {
    console.error('出战灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

    
// 召回灵宠
const recallPet = async (req, res) => {
  try {
    const { id: petId } = req.params;
    const userId = req.user.id;
    
    // 查找要召回的灵宠
    let pet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: petId },
          { id: petId }
        ]
      }
    });

    if (!pet) {
      return res.status(404).json({ message: '灵宠未找到' });
    }

    // 标记灵宠为召回状态
    await Pet.update(
      { isActive: false },
      {
        where: {
          userId: userId,
          [Sequelize.Op.or]: [
            { petId: petId },
            { id: petId }
          ]
        }
      }
    );

    // 计算灵宠属性加成（用于移除）
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

    const baseBonus = qualityBonusMap[pet.rarity] || 0;
    const starBonus = (pet.star || 0) * (starBonusPerQuality[pet.rarity] || 0);
    const levelBonus = ((pet.level || 1) - 1) * (baseBonus * 0.1);
    const phase = Math.floor((pet.star || 0) / 5);
    const phaseBonus = phase * (baseBonus * 0.5);
    const finalBonus = baseBonus + starBonus + levelBonus + phaseBonus;
    const combatBonus = finalBonus * 0.5;

    const petBonus = {
      attack: finalBonus,
      defense: finalBonus,
      health: finalBonus,
      critRate: combatBonus,
      comboRate: combatBonus,
      counterRate: combatBonus,
      stunRate: combatBonus,
      dodgeRate: combatBonus,
      vampireRate: combatBonus,
      critResist: combatBonus,
      comboResist: combatBonus,
      counterResist: combatBonus,
      stunResist: combatBonus,
      dodgeResist: combatBonus,
      vampireResist: combatBonus,
      healBoost: combatBonus,
      critDamageBoost: combatBonus,
      critDamageReduce: combatBonus,
      finalDamageBoost: combatBonus,
      finalDamageReduce: combatBonus,
      combatBoost: combatBonus,
      resistanceBoost: combatBonus
    };
    
    // 打印召回灵宠的属性，用于排查属性移除问题
    console.log(`[Pet Recall] 打印召回后灵宠属性:`, {
      petId: pet.id,
      petName: pet.name,
      petRarity: pet.rarity,
      petLevel: pet.level,
      petStar: pet.star,
      petDescription: pet.description,
      petCombatAttributes: pet.combatAttributes,
      petAttackBonus: pet.attackBonus,
      petDefenseBonus: pet.defenseBonus,
      petHealthBonus: pet.healthBonus,
      petIsActive: pet.isActive
    });

    // 获取用户当前属性
    const user = await User.findByPk(userId);
    if (user) {
      // 打印玩家原始属性
      console.log(`[Pet Recall] 玩家原始属性:`, {
        baseAttributes: user.baseAttributes,
        combatAttributes: user.combatAttributes,
        combatResistance: user.combatResistance,
        specialAttributes: user.specialAttributes
      });
      
      // 移除属性加成
      const updatedBaseAttributes = { ...user.baseAttributes };
      const updatedCombatAttributes = { ...user.combatAttributes };
      const updatedCombatResistance = { ...user.combatResistance };
      const updatedSpecialAttributes = { ...user.specialAttributes };

      // 先移除基础属性百分比加成
      updatedBaseAttributes.attack = (updatedBaseAttributes.attack || 0) / (1 + (pet.attackBonus || 0));
      updatedBaseAttributes.defense = (updatedBaseAttributes.defense || 0) / (1 + (pet.defenseBonus || 0));
      updatedBaseAttributes.health = (updatedBaseAttributes.health || 0) / (1 + (pet.healthBonus || 0));

      // 再移除combatAttributes中的属性加成
      if (pet.combatAttributes) {
        // 解析combatAttributes（如果是字符串）
        let combatAttributes = pet.combatAttributes;
        if (typeof combatAttributes === 'string') {
          try {
            combatAttributes = JSON.parse(combatAttributes);
          } catch (e) {
            console.error('解析combatAttributes失败:', e);
          }
        }
        
        // 移除基础属性
        if (combatAttributes.attack) {
          updatedBaseAttributes.attack = (updatedBaseAttributes.attack || 0) - combatAttributes.attack;
        }
        if (combatAttributes.defense) {
          updatedBaseAttributes.defense = (updatedBaseAttributes.defense || 0) - combatAttributes.defense;
        }
        if (combatAttributes.health) {
          updatedBaseAttributes.health = (updatedBaseAttributes.health || 0) - combatAttributes.health;
        }
        if (combatAttributes.speed) {
          updatedBaseAttributes.speed = (updatedBaseAttributes.speed || 0) - combatAttributes.speed;
        }
        
        // 移除战斗属性
        if (combatAttributes.critRate !== undefined) {
          updatedCombatAttributes.critRate = Math.max(0, (updatedCombatAttributes.critRate || 0) - combatAttributes.critRate);
        }
        if (combatAttributes.comboRate !== undefined) {
          updatedCombatAttributes.comboRate = Math.max(0, (updatedCombatAttributes.comboRate || 0) - combatAttributes.comboRate);
        }
        if (combatAttributes.counterRate !== undefined) {
          updatedCombatAttributes.counterRate = Math.max(0, (updatedCombatAttributes.counterRate || 0) - combatAttributes.counterRate);
        }
        if (combatAttributes.stunRate !== undefined) {
          updatedCombatAttributes.stunRate = Math.max(0, (updatedCombatAttributes.stunRate || 0) - combatAttributes.stunRate);
        }
        if (combatAttributes.dodgeRate !== undefined) {
          updatedCombatAttributes.dodgeRate = Math.max(0, (updatedCombatAttributes.dodgeRate || 0) - combatAttributes.dodgeRate);
        }
        if (combatAttributes.vampireRate !== undefined) {
          updatedCombatAttributes.vampireRate = Math.max(0, (updatedCombatAttributes.vampireRate || 0) - combatAttributes.vampireRate);
        }
        
        // 移除战斗抗性
        if (combatAttributes.critResist !== undefined) {
          updatedCombatResistance.critResist = Math.max(0, (updatedCombatResistance.critResist || 0) - combatAttributes.critResist);
        }
        if (combatAttributes.comboResist !== undefined) {
          updatedCombatResistance.comboResist = Math.max(0, (updatedCombatResistance.comboResist || 0) - combatAttributes.comboResist);
        }
        if (combatAttributes.counterResist !== undefined) {
          updatedCombatResistance.counterResist = Math.max(0, (updatedCombatResistance.counterResist || 0) - combatAttributes.counterResist);
        }
        if (combatAttributes.stunResist !== undefined) {
          updatedCombatResistance.stunResist = Math.max(0, (updatedCombatResistance.stunResist || 0) - combatAttributes.stunResist);
        }
        if (combatAttributes.dodgeResist !== undefined) {
          updatedCombatResistance.dodgeResist = Math.max(0, (updatedCombatResistance.dodgeResist || 0) - combatAttributes.dodgeResist);
        }
        if (combatAttributes.vampireResist !== undefined) {
          updatedCombatResistance.vampireResist = Math.max(0, (updatedCombatResistance.vampireResist || 0) - combatAttributes.vampireResist);
        }
        
        // 移除特殊属性
        if (combatAttributes.healBoost !== undefined) {
          updatedSpecialAttributes.healBoost = Math.max(0, (updatedSpecialAttributes.healBoost || 0) - combatAttributes.healBoost);
        }
        if (combatAttributes.critDamageBoost !== undefined) {
          updatedSpecialAttributes.critDamageBoost = Math.max(0, (updatedSpecialAttributes.critDamageBoost || 0) - combatAttributes.critDamageBoost);
        }
        if (combatAttributes.critDamageReduce !== undefined) {
          updatedSpecialAttributes.critDamageReduce = Math.max(0, (updatedSpecialAttributes.critDamageReduce || 0) - combatAttributes.critDamageReduce);
        }
        if (combatAttributes.finalDamageBoost !== undefined) {
          updatedSpecialAttributes.finalDamageBoost = Math.max(0, (updatedSpecialAttributes.finalDamageBoost || 0) - combatAttributes.finalDamageBoost);
        }
        if (combatAttributes.finalDamageReduce !== undefined) {
          updatedSpecialAttributes.finalDamageReduce = Math.max(0, (updatedSpecialAttributes.finalDamageReduce || 0) - combatAttributes.finalDamageReduce);
        }
        if (combatAttributes.combatBoost !== undefined) {
          updatedSpecialAttributes.combatBoost = Math.max(0, (updatedSpecialAttributes.combatBoost || 0) - combatAttributes.combatBoost);
        }
        if (combatAttributes.resistanceBoost !== undefined) {
          updatedSpecialAttributes.resistanceBoost = Math.max(0, (updatedSpecialAttributes.resistanceBoost || 0) - combatAttributes.resistanceBoost);
        }
      }
      
      // 打印更新后的玩家属性
      console.log(`[Pet Recall] 更新后玩家属性:`, {
        updatedBaseAttributes,
        updatedCombatAttributes,
        updatedCombatResistance,
        updatedSpecialAttributes
      });

      // 更新用户数据
      await User.update(
        {
          baseAttributes: updatedBaseAttributes,
          combatAttributes: updatedCombatAttributes,
          combatResistance: updatedCombatResistance,
          specialAttributes: updatedSpecialAttributes
        },
        {
          where: { id: userId }
        }
      );
    }

    res.status(200).send({ 
      success: true, 
      message: '召回成功', 
      pet: { 
        id: pet.id, 
        name: pet.name, 
        isActive: pet.isActive 
      } 
    });
  } catch (error) {
    console.error('召回灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};



// 升级灵宠
const upgradePet = async (req, res) => {
  try {
    const { id: petId } = req.params;
    const { essenceCount } = req.body;
    const userId = req.user.id;

    // 查找要升级的灵宠
    let pet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: petId },
          { id: petId }
        ]
      }
    });

    if (!pet) {
      return res.status(404).json({ message: '灵宠未找到' });
    }

    // 检查玩家是否有足够的灵宠精华
    const user = await User.findByPk(userId);
    if (!user) {
      return res.status(404).json({ message: '用户未找到' });
    }

    if (user.petEssence < essenceCount) {
      return res.status(400).json({ message: '灵宠精华不足' });
    }

    // 消耗灵宠精华
    await User.update(
      { petEssence: user.petEssence - essenceCount },
      { where: { id: userId } }
    );

    // 升级灵宠
    const newLevel = pet.level + 1;
    
    // 计算新的属性加成
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

    const baseBonus = qualityBonusMap[pet.rarity] || 0;
    const starBonus = (pet.star || 0) * (starBonusPerQuality[pet.rarity] || 0);
    const levelBonus = (newLevel - 1) * (baseBonus * 0.1);
    const phase = Math.floor((pet.star || 0) / 5);
    const phaseBonus = phase * (baseBonus * 0.5);
    const newBonus = baseBonus + starBonus + levelBonus + phaseBonus;
    
    await Pet.update(
      { 
        level: newLevel,
        attackBonus: newBonus,
        defenseBonus: newBonus,
        healthBonus: newBonus
      },
      {
        where: {
          userId: userId,
          [Sequelize.Op.or]: [
            { petId: petId },
            { id: petId }
          ]
        }
      }
    );

    console.log(`灵宠 ${pet.name}(${petId}) 升级至 ${newLevel} 级`);
    res.status(200).send({ success: true, message: '升级成功', newLevel });
  } catch (error) {
    console.error('升级灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// 升星灵宠
const evolvePet = async (req, res) => {
  try {
    const { id: petId } = req.params;
    const { foodPetId } = req.body;
    const userId = req.user.id;

    // 查找目标灵宠
    let targetPet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: petId },
          { id: petId }
        ]
      }
    });

    if (!targetPet) {
      return res.status(404).json({ message: '目标灵宠未找到' });
    }

    // 查找作为材料的灵宠
    const foodPet = await Pet.findOne({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: foodPetId },
          { id: foodPetId }
        ]
      }
    });

    if (!foodPet) {
      return res.status(404).json({ message: '材料灵宠未找到' });
    }

    // 检查是否是相同品质的灵宠
    if (targetPet.rarity !== foodPet.rarity) {
      return res.status(400).json({ message: '只能使用相同品质的灵宠进行升星' });
    }

    // 计算成功率
    let successRate = 1.0; // 相同名字的成功率
    if (targetPet.name !== foodPet.name) {
      successRate = 0.3; // 不同名字的成功率
    }

    // 判断升星是否成功
    const isSuccess = Math.random() < successRate;

    if (isSuccess) {
      // 升星成功，提升目标灵宠星级
      const newStar = targetPet.star + 1;
      
      // 计算新的属性加成
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
      
      const baseBonus = qualityBonusMap[targetPet.rarity] || 0;
      const starBonus = newStar * (starBonusPerQuality[targetPet.rarity] || 0);
      const levelBonus = ((targetPet.level || 1) - 1) * (baseBonus * 0.1);
      const phase = Math.floor(newStar / 5);
      const phaseBonus = phase * (baseBonus * 0.5);
      const newBonus = baseBonus + starBonus + levelBonus + phaseBonus;
      
      await Pet.update(
        { 
          star: newStar,
          attackBonus: newBonus,
          defenseBonus: newBonus,
          healthBonus: newBonus
        },
        {
          where: {
            userId: userId,
            [Sequelize.Op.or]: [
              { petId: petId },
              { id: petId }
            ]
          }
        }
      );

      console.log(`灵宠 ${targetPet.name}(${petId}) 升星成功，星级提升至 ${newStar}`);
      res.status(200).send({ success: true, message: '升星成功', newStar });
    } else {
      console.log(`灵宠 ${targetPet.name}(${petId}) 升星失败`);
      res.status(200).send({ success: false, message: '升星失败，材料已消耗' });
    }

    // 删除作为材料的灵宠
    await Pet.destroy({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: foodPetId },
          { id: foodPetId }
        ]
      }
    });
  } catch (error) {
    console.error('升星灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// 批量放生灵宠
const batchReleasePets = async (req, res) => {
  try {
    const { rarity } = req.body;
    const userId = req.user.id;

    // 构建查询条件
    const whereCondition = {
      userId: userId,
      isActive: false // 排除出战的灵宠
    };

    // 如果指定了品质，则添加品质筛选条件
    if (rarity) {
      whereCondition.rarity = rarity;
    }

    // 查找符合条件的灵宠
    const petsToRelease = await Pet.findAll({ where: whereCondition });

    if (petsToRelease.length === 0) {
      return res.status(400).json({ message: '没有符合条件的灵宠可放生' });
    }

    // 计算总共可以获得的经验值
    let totalExp = 0;
    petsToRelease.forEach(pet => {
      // 根据等级和星级计算经验值
      const exp = Math.max(1, Math.floor((pet.level || 1) * (pet.star + 1) * 0.5));
      totalExp += exp;
    });

    // 删除这些宠物
    await Pet.destroy({ where: whereCondition });

    // 增加经验值
    const user = await User.findByPk(userId);
    if (user) {
      await User.update(
        { exp: user.exp + totalExp },
        { where: { id: userId } }
      );
    }

    console.log(`成功放生 ${petsToRelease.length} 只灵宠，获得 ${totalExp} 点经验`);
    res.status(200).send({ 
      success: true, 
      message: `成功放生${petsToRelease.length}只灵宠，获得${totalExp}点经验`,
      releasedCount: petsToRelease.length,
      expGained: totalExp
    });
  } catch (error) {
    console.error('批量放生灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// 删除宠物
const deletePets = async (req, res) => {
  try {
    const { petIds } = req.body;
    const userId = req.user.id;

    if (!petIds || !Array.isArray(petIds)) {
      return res.status(400).json({ message: '无效的宠物ID列表' });
    }

    // 删除指定的宠物
    await Pet.destroy({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { petId: { [Sequelize.Op.in]: petIds } },
          { id: { [Sequelize.Op.in]: petIds } }
        ]
      }
    });

    console.log(`成功删除 ${petIds.length} 个灵宠`);
    res.status(200).send({ message: '灵宠删除成功' });
  } catch (error) {
    console.error('删除灵宠失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

module.exports = {
  deployPet,
  recallPet,
  upgradePet,
  evolvePet,
  batchReleasePets,
  deletePets
};