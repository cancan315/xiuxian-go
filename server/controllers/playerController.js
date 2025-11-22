const User = require('../models/User');
const Item = require('../models/Item');
const Herb = require('../models/Herb');
const Pill = require('../models/Pill');
const Artifact = require('../models/Artifact');

// Get player data
const getPlayerData = async (req, res) => {
  try {
    const user = await User.findByPk(req.user.id);
    if (!user) {
      return res.status(404).json({ message: '用户不存在' });
    }

    // Get player items
    const items = await Item.findAll({ where: { userId: req.user.id } });
    
    // Get player herbs
    const herbs = await Herb.findAll({ where: { userId: req.user.id } });
    
    // Get player pills
    const pills = await Pill.findAll({ where: { userId: req.user.id } });
    
    // Get player artifacts
    const artifacts = await Artifact.findAll({ where: { userId: req.user.id } });

    res.json({
      user,
      items,
      herbs,
      pills,
      artifacts
    });
  } catch (error) {
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// Save player data
const savePlayerData = async (req, res) => {
  try {
    const { user, items, herbs, pills, artifacts } = req.body;
    const userId = req.user.id;

    // Update user data
    await User.update(user, {
      where: { id: userId }
    });

    // Update items - First delete all existing items, then insert new ones
    await Item.destroy({ where: { userId } });
    if (items && items.length > 0) {
      const itemsWithUserId = items.map(item => ({
        ...item,
        userId
      }));
      await Item.bulkCreate(itemsWithUserId);
    }

    // Update herbs
    await Herb.destroy({ where: { userId } });
    if (herbs && herbs.length > 0) {
      const herbsWithUserId = herbs.map(herb => ({
        ...herb,
        userId
      }));
      await Herb.bulkCreate(herbsWithUserId);
    }

    // Update pills
    await Pill.destroy({ where: { userId } });
    if (pills && pills.length > 0) {
      const pillsWithUserId = pills.map(pill => ({
        ...pill,
        userId
      }));
      await Pill.bulkCreate(pillsWithUserId);
    }

    // Update artifacts
    await Artifact.destroy({ where: { userId } });
    if (artifacts && artifacts.length > 0) {
      const artifactsWithUserId = artifacts.map(artifact => ({
        ...artifact,
        userId
      }));
      await Artifact.bulkCreate(artifactsWithUserId);
    }

    res.json({ message: '数据保存成功' });
  } catch (error) {
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
  getPlayerData,
  savePlayerData,
  getLeaderboard
};