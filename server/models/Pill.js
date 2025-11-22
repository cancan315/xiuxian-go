const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// Player pills
const Pill = sequelize.define('Pill', {
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
  pillId: {
    type: DataTypes.STRING,
    allowNull: false
  },
  name: {
    type: DataTypes.STRING,
    allowNull: false
  },
  description: {
    type: DataTypes.TEXT,
    allowNull: true
  },
  effect: {
    type: DataTypes.JSON,
    allowNull: true
  }
});

module.exports = Pill;