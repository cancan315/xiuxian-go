const { Sequelize } = require('sequelize');
const User = require('../models/User');
const Item = require('../models/Item');
const Pet = require('../models/Pet');
const Herb = require('../models/Herb');
const Pill = require('../models/Pill');
const { v4: uuidv4 } = require('uuid');

// Get player data (deprecated, use initializePlayer instead)
const getPlayerData = async (req, res) => {
  try {
    const user = await User.findByPk(req.user.id);
    if (!user) {
      return res.status(404).json({ message: '用户不存在' });
    }

    // Get all player items (excluding pets now)
    const items = await Item.findAll({ 
      where: { 
        userId: req.user.id
      } 
    });
    
    // 处理从数据库获取的物品数据，确保quality字段被正确解析
    const processedItems = items.map(item => {
      // 如果quality是字符串，则尝试解析为对象
      if (item.quality && typeof item.quality === 'string') {
        try {
          // 首先检查是否已经是有效的JSON字符串
          if (item.quality.startsWith('{') || item.quality.startsWith('[')) {
            item.quality = JSON.parse(item.quality);
          }
          // 如果不是JSON格式（比如是简单的字符串如"common"），则保持原样
        } catch (e) {
          console.error('解析物品quality失败:', e);
        }
      }
      
      // 如果stats是字符串，则尝试解析为对象
      if (item.stats && typeof item.stats === 'string') {
        try {
          item.stats = JSON.parse(item.stats);
        } catch (e) {
          console.error('解析物品stats失败:', e);
        }
      }
      
      // 如果details是字符串，则尝试解析为对象
      if (item.details && typeof item.details === 'string') {
        try {
          item.details = JSON.parse(item.details);
        } catch (e) {
          console.error('解析物品details失败:', e);
        }
      }
      
      return item;
    });
    
    // Get player pets
    const pets = await Pet.findAll({ 
      where: { 
        userId: req.user.id
      } 
    });
    
    // 处理从数据库获取的宠物数据，确保quality和combatAttributes字段被正确解析
    const processedPets = pets.map(pet => {
      // 如果quality是字符串，则尝试解析为对象
      if (pet.quality && typeof pet.quality === 'string') {
        try {
          // 首先检查是否已经是有效的JSON字符串
          if (pet.quality.startsWith('{') || pet.quality.startsWith('[')) {
            pet.quality = JSON.parse(pet.quality);
          }
          // 如果不是JSON格式（比如是简单的字符串如"common"），则保持原样
        } catch (e) {
          console.error('解析宠物quality失败:', e);
        }
      }
      
      // 如果combatAttributes是字符串，则尝试解析为对象
      if (pet.combatAttributes && typeof pet.combatAttributes === 'string') {
        try {
          pet.combatAttributes = JSON.parse(pet.combatAttributes);
        } catch (e) {
          console.error('解析宠物combatAttributes失败:', e);
        }
      }
      
      // 确保宠物有combatAttributes属性，如果没有则创建一个空对象
      if (!pet.combatAttributes) {
        console.error('严重错误：宠物缺少combatAttributes属性，这可能导致属性计算错误:', {
          petId: pet.id,
          petName: pet.name,
          petRarity: pet.rarity,
          userId: pet.userId
        });
        pet.combatAttributes = {};
      }
      
      // 使用数据库中预计算的属性加成
      pet.bonus = {
        attack: pet.attackBonus || 0,
        defense: pet.defenseBonus || 0,
        health: pet.healthBonus || 0
      };
      
      return pet;
    });
    
    // Get player herbs
    const herbs = await Herb.findAll({ where: { userId: req.user.id } });
    
    // Get player pills
    const pills = await Pill.findAll({ where: { userId: req.user.id } });
    
    // No separate artifacts table anymore
    const artifacts = [];
    
    res.json({
      user,
      items: processedItems,
      pets: processedPets,
      herbs,
      pills: pills,
      artifacts: artifacts
    });
  } catch (error) {
    console.error('获取玩家数据失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// Initialize player data
const initializePlayer = async (req, res) => {
  try {
    const user = await User.findByPk(req.user.id);
    if (!user) {
      return res.status(404).json({ message: '用户不存在' });
    }

    // Get all player items (excluding pets now)
    const items = await Item.findAll({ 
      where: { 
        userId: req.user.id
      } 
    });
    
    // 处理从数据库获取的物品数据，确保quality字段被正确解析
    const processedItems = items.map(item => {
      // 如果quality是字符串，则尝试解析为对象
      if (item.quality && typeof item.quality === 'string') {
        try {
          // 首先检查是否已经是有效的JSON字符串
          if (item.quality.startsWith('{') || item.quality.startsWith('[')) {
            item.quality = JSON.parse(item.quality);
          }
          // 如果不是JSON格式（比如是简单的字符串如"common"），则保持原样
        } catch (e) {
          console.error('解析物品quality失败:', e);
        }
      }
      
      // 如果stats是字符串，则尝试解析为对象
      if (item.stats && typeof item.stats === 'string') {
        try {
          item.stats = JSON.parse(item.stats);
        } catch (e) {
          console.error('解析物品stats失败:', e);
        }
      }
      
      // 如果details是字符串，则尝试解析为对象
      if (item.details && typeof item.details === 'string') {
        try {
          item.details = JSON.parse(item.details);
        } catch (e) {
          console.error('解析物品details失败:', e);
        }
      }
      
      return item;
    });
    
    // Get player pets
    const pets = await Pet.findAll({ 
      where: { 
        userId: req.user.id
      } 
    });
    
    // 处理从数据库获取的宠物数据，确保quality和combatAttributes字段被正确解析
    const processedPets = pets.map(pet => {
      // 如果quality是字符串，则尝试解析为对象
      if (pet.quality && typeof pet.quality === 'string') {
        try {
          // 首先检查是否已经是有效的JSON字符串
          if (pet.quality.startsWith('{') || pet.quality.startsWith('[')) {
            pet.quality = JSON.parse(pet.quality);
          }
          // 如果不是JSON格式（比如是简单的字符串如"common"），则保持原样
        } catch (e) {
          console.error('解析宠物quality失败:', e);
        }
      }
      
      // 如果combatAttributes是字符串，则尝试解析为对象
      if (pet.combatAttributes && typeof pet.combatAttributes === 'string') {
        try {
          pet.combatAttributes = JSON.parse(pet.combatAttributes);
        } catch (e) {
          console.error('解析宠物combatAttributes失败:', e);
        }
      }
      
      // 确保宠物有combatAttributes属性，如果没有则创建一个空对象
      if (!pet.combatAttributes) {
        console.error('严重错误：宠物缺少combatAttributes属性，这可能导致属性计算错误:', {
          petId: pet.id,
          petName: pet.name,
          petRarity: pet.rarity,
          userId: pet.userId
        });
        pet.combatAttributes = {};
      }
      
      // 使用数据库中预计算的属性加成
      pet.bonus = {
        attack: pet.attackBonus || 0,
        defense: pet.defenseBonus || 0,
        health: pet.healthBonus || 0
      };
      
      return pet;
    });
    
    // Get player herbs
    const herbs = await Herb.findAll({ where: { userId: req.user.id } });
    
    // Get player pills
    const pills = await Pill.findAll({ where: { userId: req.user.id } });
    
    // No separate artifacts table anymore
    const artifacts = [];
    
    res.json({
      user,
      items: processedItems,
      pets: processedPets,
      herbs,
      pills: pills,
      artifacts: artifacts
    });
  } catch (error) {
    console.error('初始化玩家数据失败:', error);
    res.status(500).json({ message: 'Server error', error: error.message });
  }
};

// Get player spirit value only
const getPlayerSpirit = async (req, res) => {
  try {
    const user = await User.findByPk(req.user.id, {
      attributes: ['spirit'] // 只选择spirit字段
    });
    
    if (!user) {
      return res.status(404).json({ message: '用户不存在' });
    }
    
    res.json({ spirit: user.spirit });
  } catch (error) {
    console.error('获取玩家灵力值失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// 增量更新灵力值
const updateSpirit = async (req, res) => {
  try {
    const { spirit } = req.body;
    const userId = req.user.id;

    // 只更新灵力字段
    await User.update(
      { spirit: spirit },
      { where: { id: userId } }
    );

    res.status(200).send({ message: '灵力值更新成功' });
  } catch (error) {
    console.error('更新灵力值失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// Get leaderboard data
const getLeaderboard = async (req, res) => {
  try {
    // 获取前100名玩家，按境界(level)和灵石数量(spiritStones)排序
    const leaderboard = await User.findAll({
      attributes: ['id', 'playerName', 'level', 'realm', 'spiritStones'],
      order: [
        ['level', 'DESC'],       // 境界高的排在前面
        ['spiritStones', 'DESC']  // 境界相同时，灵石多的排在前面
      ],
      limit: 100
    });

    res.json(leaderboard);
  } catch (error) {
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

module.exports = {
  initializePlayer, // 初始化玩家数据
  getPlayerData,     // 获取玩家数据（已弃用，推荐使用initializePlayer）
  getPlayerSpirit,   // 仅获取玩家灵力值
  updateSpirit,      // 更新玩家灵力值
  getLeaderboard    // 获取排行榜数据
};