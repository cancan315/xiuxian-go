const { Sequelize } = require('sequelize');
require('dotenv').config();

// Check if we're running inside Docker
const isDocker = process.env.DB_HOST === 'postgres';

// Database configuration
const sequelize = new Sequelize(
  process.env.DB_NAME || 'xiuxian_db',
  process.env.DB_USER || 'xiuxian_user',
  process.env.DB_PASSWORD || 'xiuxian_password',
  {
    host: isDocker ? 'postgres' : 'localhost', // Use postgres service name in Docker, localhost otherwise
    port: process.env.DB_PORT || 5432,
    dialect: 'postgres',
    logging: false,
    pool: {
      max: 5,
      min: 0,
      acquire: 30000,
      idle: 10000
    }
  }
);

module.exports = sequelize;