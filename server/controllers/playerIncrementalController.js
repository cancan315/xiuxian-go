const { Sequelize } = require('sequelize');
const User = require('../models/User');
const Item = require('../models/Item');
const Equipment = require('../models/Equipment');
const Pet = require('../models/Pet');
const Herb = require('../models/Herb');
const Pill = require('../models/Pill');
const { v4: uuidv4 } = require('uuid');

// 增量更新玩家数据
const updatePlayerData = async (req, res) => {
  try {
    const { user, items, pets, herbs, pills } = req.body;
    const userId = req.user.id;

    console.log('接收到玩家数据增量更新请求:', {
      userId,
      userName: user?.playerName,
      itemCount: items?.length,
      petCount: pets?.length,
      herbCount: herbs?.length,
      pillCount: pills?.length
    });

    // 只更新提供的用户数据字段
    if (user) {
      await User.update(user, {
        where: { id: userId }
      });
    }

    // 增量更新物品 - 只更新提供的物品
    if (items && items.length > 0) {
      // 分批处理以避免数据库超时
      const batchSize = 50;
      for (let i = 0; i < items.length; i += batchSize) {
        const batch = items.slice(i, i + batchSize);
        const promises = batch.map(async (item) => {
          // 处理 stats 字段，确保它可以被正确序列化
          let processedStats = item.stats;
          if (typeof item.stats === 'object' && item.stats !== null) {
            try {
              processedStats = JSON.stringify(item.stats);
            } catch (e) {
              console.error('stats序列化失败:', e);
              processedStats = '{}';
            }
          }
          
          // 处理 details 字段，确保它可以被正确序列化
          let processedDetails = item.details;
          if (typeof item.details === 'object' && item.details !== null) {
            try {
              processedDetails = JSON.stringify(item.details);
            } catch (e) {
              console.error('details序列化失败:', e);
              processedDetails = '{}';
            }
          }
          
          // 处理 quality 字段，确保它可以被正确序列化
          let processedQuality = item.quality;
          if (typeof item.quality === 'object' && item.quality !== null) {
            try {
              processedQuality = JSON.stringify(item.quality);
            } catch (e) {
              console.error('quality序列化失败:', e);
              processedQuality = '';
            }
          }
          
          // 确保必要字段存在
          const itemId = item.itemId || item.id || `item_${Date.now()}_${Math.random()}`;
          
          // 确保装备有slot字段
          if ((item.type !== 'pet' && item.type !== 'pill' && item.type !== 'herb') && !item.slot) {
            // 从前端定义的装备类型中获取slot信息
            // 这里我们假设前端会在type字段中传递装备类型（例如"faqi"），然后我们可以映射到slot
            item.slot = item.type; // 默认使用type作为slot
          }
          
          // 查找现有物品（只查找装备类型）
          let existingItem = null;
          if (item.type !== 'pet' && item.type !== 'pill' && item.type !== 'herb') {
            existingItem = await Equipment.findOne({
              where: { 
                [Sequelize.Op.or]: [
                  { equipmentId: itemId },
                  { id: item.id }
                ].filter(condition => {
                  const key = Object.keys(condition)[0];
                  return condition[key] !== undefined;
                }), // 过滤掉undefined的条件
                userId: userId
              }
            });
          }
          
          if (existingItem) {
            // 更新现有物品 - 不更新主键id字段
            const updateData = {
              ...item,
              userId,
              equipmentId: itemId,
              stats: processedStats,
              details: processedDetails,
              quality: processedQuality,
              equipped: item.equipped || false,
              slot: item.slot || item.type // 确保slot字段存在
            };
            
            // 删除 updateData.id 以避免主键冲突
            delete updateData.id;
            
            return Equipment.update(updateData, {
              where: { id: existingItem.id }
            });
          } else if (item.type !== 'pet' && item.type !== 'pill' && item.type !== 'herb') {
            // 创建新物品（只创建装备类型）
            // 为了避免主键冲突，我们不传递 id 字段，让数据库自动生成 UUID
            const newItemData = {
              ...item,
              userId,
              equipmentId: itemId,
              stats: processedStats,
              details: processedDetails,
              quality: processedQuality,
              equipped: item.equipped || false,
              slot: item.slot || item.type // 确保slot字段存在
            };
            
            // 删除 newItemData.id 以避免主键冲突
            delete newItemData.id;
            
            return Equipment.create(newItemData);
          } else if (item.type === 'pet' || item.type === 'pill' || item.type === 'herb') {
            // 对于非装备类型，仍使用原来的 Item 模型
            existingItem = await Item.findOne({
              where: { 
                [Sequelize.Op.or]: [
                  { itemId: itemId },
                  { id: item.id }
                ].filter(condition => {
                  const key = Object.keys(condition)[0];
                  return condition[key] !== undefined;
                }), // 过滤掉undefined的条件
                userId: userId
              }
            });
            
            if (existingItem) {
              // 更新现有物品 - 不更新主键id字段
              const updateData = {
                ...item,
                userId,
                itemId: itemId,
                stats: processedStats,
                details: processedDetails,
                quality: processedQuality,
                equipped: item.equipped || false,
                slot: item.slot || null
              };
              
              // 删除 updateData.id 以避免主键冲突
              delete updateData.id;
              
              return Item.update(updateData, {
                where: { id: existingItem.id }
              });
            } else {
              // 创建新物品
              // 为了避免主键冲突，我们不传递 id 字段，让数据库自动生成 UUID
              const newItemData = {
                ...item,
                userId,
                itemId: itemId,
                stats: processedStats,
                details: processedDetails,
                quality: processedQuality,
                equipped: item.equipped || false,
                slot: item.slot || null
              };
              
              // 删除 newItemData.id 以避免主键冲突
              delete newItemData.id;
              
              return Item.create(newItemData);
            }
          }
        });
        
        // 并行处理批次内的所有项目
        await Promise.all(promises);
      }
      
      console.log(`成功增量更新 ${items.length} 个物品到数据库`);
    }

    // 增量更新宠物 - 只更新提供的宠物
    if (pets && pets.length > 0) {
      // 分批处理以避免数据库超时
      const batchSize = 50;
      for (let i = 0; i < pets.length; i += batchSize) {
        const batch = pets.slice(i, i + batchSize);
        const promises = batch.map(async (pet) => {
          // 处理 quality 字段，确保它可以被正确序列化
          let processedQuality = pet.quality;
          if (typeof pet.quality === 'object' && pet.quality !== null) {
            try {
              processedQuality = JSON.stringify(pet.quality);
            } catch (e) {
              console.error('宠物quality序列化失败:', e);
              processedQuality = '{}';
            }
          }
          
          // 处理 combatAttributes 字段，确保它可以被正确序列化
          let processedCombatAttributes = pet.combatAttributes;
          if (typeof pet.combatAttributes === 'object' && pet.combatAttributes !== null) {
            try {
              processedCombatAttributes = JSON.stringify(pet.combatAttributes);
            } catch (e) {
              console.error('宠物combatAttributes序列化失败:', e);
              processedCombatAttributes = '{}';
            }
          } else if (!pet.combatAttributes) {
            // 如果combatAttributes是null或undefined，设置为空对象字符串
            processedCombatAttributes = '{}';
          }
          
          // 从quality中提取rarity信息（如果存在）
          let petRarity = pet.rarity;
          if (!petRarity && pet.quality) {
            if (typeof pet.quality === 'object' && pet.quality.rarity) {
              petRarity = pet.quality.rarity;
            } else if (typeof pet.quality === 'string') {
              try {
                const parsedQuality = JSON.parse(pet.quality);
                if (parsedQuality.rarity) {
                  petRarity = parsedQuality.rarity;
                }
              } catch (e) {
                // 如果解析失败，保持petRarity为undefined
              }
            }
          }
          
          // 确保必要字段存在
          const petId = pet.petId || pet.id || `pet_${Date.now()}_${Math.random()}`;
          
          // 查找现有宠物
          const existingPet = await Pet.findOne({
            where: { 
              [Sequelize.Op.or]: [
                { petId: petId },
                { id: pet.id }
              ].filter(condition => {
                const key = Object.keys(condition)[0];
                return condition[key] !== undefined;
              }), // 过滤掉undefined的条件
              userId: userId
            }
          });
          
          if (existingPet) {
            // 更新现有宠物 - 不更新主键id字段
            const updateData = {
              ...pet,
              userId,
              petId: petId,
              rarity: petRarity || 'mortal',
              quality: processedQuality,
              combatAttributes: processedCombatAttributes,
              isActive: pet.isActive || false
            };
            
            // 删除 updateData.id 以避免主键冲突
            delete updateData.id;
            
            return Pet.update(updateData, {
              where: { id: existingPet.id }
            });
          } else {
            // 创建新宠物
            // 为了避免主键冲突，我们不传递 id 字段，让数据库自动生成 UUID
            const newPetData = {
              ...pet,
              userId,
              petId: petId,
              rarity: petRarity || 'mortal',
              quality: processedQuality,
              combatAttributes: processedCombatAttributes,
              isActive: pet.isActive || false
            };
            
            // 删除 newPetData.id 以避免主键冲突
            delete newPetData.id;
            
            return Pet.create(newPetData);
          }
        });
        
        // 并行处理批次内的所有宠物
        await Promise.all(promises);
      }
      
      console.log(`成功增量更新 ${pets.length} 个灵宠到数据库`);
    }

    // 增量更新灵草 - 只更新提供的灵草
    if (herbs && herbs.length > 0) {
      // 分批处理以避免数据库超时
      const batchSize = 50;
      for (let i = 0; i < herbs.length; i += batchSize) {
        const batch = herbs.slice(i, i + batchSize);
        const promises = batch.map(async (herb) => {
          const herbId = herb.herbId || herb.id || `herb_${Date.now()}_${Math.random()}`;
          
          // 查找现有灵草
          const existingHerb = await Herb.findOne({
            where: { 
              [Sequelize.Op.or]: [
                { herbId: herbId },
                { id: herb.id }
              ].filter(condition => {
                const key = Object.keys(condition)[0];
                return condition[key] !== undefined;
              }), // 过滤掉undefined的条件
              userId: userId
            }
          });
          
          if (existingHerb) {
            // 更新现有灵草 - 不更新主键id字段
            const updateData = {
              ...herb,
              userId,
              herbId: herbId
            };
            
            // 删除 updateData.id 以避免主键冲突
            delete updateData.id;
            
            return Herb.update(updateData, {
              where: { id: existingHerb.id }
            });
          } else {
            // 创建新灵草
            // 为了避免主键冲突，我们不传递 id 字段
            const newHerbData = {
              ...herb,
              userId,
              herbId: herbId
            };
            
            // 删除 newHerbData.id 以避免主键冲突
            delete newHerbData.id;
            
            return Herb.create(newHerbData);
          }
        });
        
        // 并行处理批次内的所有灵草
        await Promise.all(promises);
      }
      
      console.log(`成功增量更新 ${herbs.length} 个灵草到数据库`);
    }

    // 增量更新丹药 - 只更新提供的丹药
    if (pills && pills.length > 0) {
      // 分批处理以避免数据库超时
      const batchSize = 50;
      for (let i = 0; i < pills.length; i += batchSize) {
        const batch = pills.slice(i, i + batchSize);
        const promises = batch.map(async (pill) => {
          const pillId = pill.pillId || pill.id || `pill_${Date.now()}_${Math.random()}`;
          
          // 查找现有丹药
          const existingPill = await Pill.findOne({
            where: { 
              [Sequelize.Op.or]: [
                { pillId: pillId },
                { id: pill.id }
              ].filter(condition => {
                const key = Object.keys(condition)[0];
                return condition[key] !== undefined;
              }), // 过滤掉undefined的条件
              userId: userId
            }
          });
          
          if (existingPill) {
            // 更新现有丹药 - 不更新主键id字段
            const updateData = {
              ...pill,
              userId,
              pillId: pillId
            };
            
            // 删除 updateData.id 以避免主键冲突
            delete updateData.id;
            
            return Pill.update(updateData, {
              where: { id: existingPill.id }
            });
          } else {
            // 创建新丹药
            // 为了避免主键冲突，我们不传递 id 字段
            const newPillData = {
              ...pill,
              userId,
              pillId: pillId
            };
            
            // 删除 newPillData.id 以避免主键冲突
            delete newPillData.id;
            
            return Pill.create(newPillData);
          }
        });
        
        // 并行处理批次内的所有丹药
        await Promise.all(promises);
      }
      
      console.log(`成功增量更新 ${pills.length} 个丹药到数据库`);
    }

    console.log('玩家数据增量更新完成:', userId);
    res.status(200).send({ message: '数据增量更新成功' });
  } catch (error) {
    console.error('服务器增量更新数据失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// 删除物品
const deleteItems = async (req, res) => {
  try {
    const { itemIds } = req.body;
    const userId = req.user.id;

    if (!itemIds || !Array.isArray(itemIds)) {
      return res.status(400).json({ message: '无效的物品ID列表' });
    }

    // 删除指定的物品（从Equipment表中删除装备）
    await Equipment.destroy({
      where: {
        userId: userId,
        [Sequelize.Op.or]: [
          { equipmentId: { [Sequelize.Op.in]: itemIds } },
          { id: { [Sequelize.Op.in]: itemIds } }
        ]
      }
    });

    console.log(`成功删除 ${itemIds.length} 个物品`);
    res.status(200).send({ message: '物品删除成功' });
  } catch (error) {
    console.error('删除物品失败:', error);
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

module.exports = {
  updatePlayerData,
  deleteItems
};