const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const User = require('../models/User');

// Generate JWT token
const generateToken = (userId) => {
  return jwt.sign({ id: userId }, process.env.JWT_SECRET, {
    expiresIn: '30d'
  });
};

// Register user
const register = async (req, res) => {
  try {
    const { username, password } = req.body;

    // Check if user already exists
    const existingUser = await User.findOne({ where: { username } });
    if (existingUser) {
      return res.status(400).json({ message: '用户名已存在' });
    }

    // Hash password
    const salt = await bcrypt.genSalt(10);
    const hashedPassword = await bcrypt.hash(password, salt);

    // Create user
    const user = await User.create({
      username,
      password: hashedPassword
    });

    // Generate token
    const token = generateToken(user.id);

    res.status(201).json({
      id: user.id,
      username: user.username,
      token
    });
  } catch (error) {
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// Login user
const login = async (req, res) => {
  try {
    const { username, password } = req.body;

    // Check if user exists
    const user = await User.findOne({ where: { username } });
    if (!user) {
      return res.status(400).json({ message: '用户名或密码错误' });
    }

    // Check password
    const isMatch = await bcrypt.compare(password, user.password);
    if (!isMatch) {
      return res.status(400).json({ message: '用户名或密码错误' });
    }

    // Generate token
    const token = generateToken(user.id);

    res.json({
      id: user.id,
      username: user.username,
      token
    });
  } catch (error) {
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

// Get user data
const getUser = async (req, res) => {
  try {
    const user = await User.findByPk(req.user.id, {
      attributes: { exclude: ['password'] } // Exclude password from response
    });
    
    if (!user) {
      return res.status(404).json({ message: '用户不存在' });
    }
    
    res.json(user);
  } catch (error) {
    res.status(500).json({ message: '服务器错误', error: error.message });
  }
};

module.exports = {
  register,
  login,
  getUser
};