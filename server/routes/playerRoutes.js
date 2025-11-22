const express = require('express');
const router = express.Router();
const { getPlayerData, savePlayerData, getLeaderboard } = require('../controllers/playerController');
const { protect } = require('../middleware/authMiddleware');

// Player routes
router.get('/data', protect, getPlayerData);
router.post('/data', protect, savePlayerData);
router.get('/leaderboard', getLeaderboard);

module.exports = router;