const express = require('express');
const router = express.Router();
const { register, login, getUser } = require('../controllers/authController');
const { protect } = require('../middleware/authMiddleware');

// Auth routes
router.post('/register', register);
router.post('/login', login);
router.get('/user', protect, getUser);

module.exports = router;