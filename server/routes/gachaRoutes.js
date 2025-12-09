const express = require('express');
const router = express.Router();
const { drawGacha, processAutoActions } = require('../controllers/gachaController');
const { protect } = require('../middleware/authMiddleware');

// Apply authentication middleware to all routes
router.use(protect);

// 执行抽奖
router.post('/draw', drawGacha);

// 执行自动处理操作
router.post('/auto-actions', processAutoActions);

module.exports = router;