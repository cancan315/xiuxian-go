const Item = require('../models/Item');
const Equipment = require('../models/Equipment');

// 获取玩家装备数据
const getPlayerInventory = async (req, res) => {
  try {
    const { userId } = req.params;
    const { type, quality, equipped } = req.query;
    
    // 优先使用URL参数中的userId，其次使用认证用户ID，最后才是请求体中的userId
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
    // 注意：数据库中使用equipType字段存储装备类型
    if (type) {
      whereConditions.equipType = type;
    } else if (req.query.equipType) {
      whereConditions.equipType = req.query.equipType;
    }
    
    // 如果提供了品质过滤条件
    if (quality) {
      whereConditions.quality = quality;
    }
    
    // 如果提供了装备状态过滤条件
    if (equipped !== undefined) {
      whereConditions.equipped = equipped === 'true';
    }
    
    // 查询符合条件的物品
    const items = await Equipment.findAll({
      where: whereConditions,
      order: [['createdAt', 'DESC']] // 按创建时间倒序排列
    });
    
    // 处理从数据库获取的物品数据
    const processedItems = items.map(item => {
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
    
    console.log(`[Inventory Controller] 获取玩家装备数据: 用户ID=${actualUserId}, 类型=${type}, 品质=${quality}, 装备状态=${equipped}, 结果数量=${processedItems.length}`);
    
    res.json({
      success: true,
      items: processedItems
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
const getItemDetails = async (req, res) => {
  try {
    const { itemId } = req.params;
    
    // 查询特定物品
    const item = await Equipment.findOne({
      where: {
        id: itemId,
        userId: req.user.id // 确保只能访问自己的物品
      }
    });
    
    if (!item) {
      return res.status(404).json({
        success: false,
        message: '物品未找到'
      });
    }
    
    // 处理物品数据
    const processedItem = { ...item.dataValues };
    
    // 如果stats是字符串，则尝试解析为对象
    if (processedItem.stats && typeof processedItem.stats === 'string') {
      try {
        processedItem.stats = JSON.parse(processedItem.stats);
      } catch (e) {
        console.error('解析物品stats失败:', e);
      }
    }
    
    // 如果details是字符串，则尝试解析为对象
    if (processedItem.details && typeof processedItem.details === 'string') {
      try {
        processedItem.details = JSON.parse(processedItem.details);
      } catch (e) {
        console.error('解析物品details失败:', e);
      }
    }
    
    res.json({
      success: true,
      item: processedItem
    });
  } catch (error) {
    console.error('获取物品详情失败:', error);
    res.status(500).json({ 
      success: false, 
      message: '服务器错误', 
      error: error.message 
    });
  }
};

module.exports = {
  getPlayerInventory,
  getItemDetails
};