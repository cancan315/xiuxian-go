const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// Player inventory items
const Item = sequelize.define('Item', {
  id: {
    type: DataTypes.UUID,
    defaultValue: DataTypes.UUIDV4,
    primaryKey: true,
    allowNull: false
  },
  userId: {
    type: DataTypes.INTEGER,
    allowNull: false,
    references: {
      model: 'Users',
      key: 'id'
    }
  },
  itemId: {
    type: DataTypes.STRING,
    allowNull: false
  },
  name: {
    type: DataTypes.STRING,
    allowNull: false
  },
  type: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // Store item details as JSON
  details: {
    type: DataTypes.JSON,
    allowNull: true
  },
  slot: {
    type: DataTypes.STRING,
    allowNull: true
  },
  stats: {
    type: DataTypes.JSON,
    allowNull: true
  },
  quality: {
    type: DataTypes.STRING,
    allowNull: true
  },
  equipped: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  }
  // Removed isActive field since it doesn't exist in the database table
});

// Export the model with its associated methods
module.exports = Item;