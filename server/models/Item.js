const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// Player inventory items
const Item = sequelize.define('Item', {
  id: {
    type: DataTypes.INTEGER,
    primaryKey: true,
    autoIncrement: true
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
  }
});

module.exports = Item;