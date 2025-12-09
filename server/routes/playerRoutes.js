const express = require('express');
const router = express.Router();
const { initializePlayer, getPlayerData, getPlayerSpirit, updateSpirit, getLeaderboard } = require('../controllers/playerController');
const { updatePlayerData, deleteItems } = require('../controllers/playerIncrementalController');
const { deletePets, deployPet, recallPet, upgradePet, evolvePet, batchReleasePets } = require('../controllers/petController');
const { getPlayerInventory, getItemDetails } = require('../controllers/inventoryController');
const { 
  getPlayerEquipment,
  getEquipmentDetails,
  enhanceEquipment,
  reforgeEquipment,
  confirmReforge,
  equipEquipment,
  unequipEquipment,
  sellEquipment,
  batchSellEquipment
} = require('../controllers/equipmentController');
const { protect } = require('../middleware/authMiddleware');

// Get leaderboard - 公开访问，不需要认证
router.get('/leaderboard', getLeaderboard);

// Apply authentication middleware to all routes
router.use(protect);

// Initialize player data
router.get('/init', initializePlayer);

// Get player data (deprecated, use /init instead)
router.get('/data', getPlayerData);

// Get player spirit value only
router.get('/spirit', getPlayerSpirit);

// Incremental update player data
router.patch('/data', updatePlayerData);

// Delete items
router.delete('/items', deleteItems);

// Delete pets
router.delete('/pets', deletePets);

// 增量更新灵力值
router.put('/spirit', updateSpirit);

// 获取玩家装备数据
router.get('/inventory/:userId?', getPlayerInventory);

// 获取特定装备详情
router.get('/inventory/item/:itemId', getItemDetails);

// 装备系统API
// 获取玩家装备列表
router.get('/equipment/:userId?', getPlayerEquipment);

// 获取特定装备详情
router.get('/equipment/details/:id', getEquipmentDetails);

// 装备强化
router.post('/equipment/:id/enhance', enhanceEquipment);

// 装备洗练
router.post('/equipment/:id/reforge', reforgeEquipment);

// 确认洗练结果
router.post('/equipment/:id/reforge-confirm', confirmReforge);

// 装备穿戴
router.post('/equipment/:id/equip', equipEquipment);

// 装备卸下
router.post('/equipment/:id/unequip', unequipEquipment);

// 出售装备
router.delete('/equipment/:id', sellEquipment);

// 批量出售装备
router.post('/equipment/batch-sell', batchSellEquipment);

// 灵宠系统API
// 出战灵宠
router.post('/pets/:id/deploy', deployPet);

// 召回灵宠
router.post('/pets/:id/recall', recallPet);

// 升级灵宠
router.post('/pets/:id/upgrade', upgradePet);

// 升星灵宠
router.post('/pets/:id/evolve', evolvePet);

// 批量放生灵宠
router.post('/pets/batch-release', batchReleasePets);

module.exports = router;