const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// Player artifacts (equipment)
const Artifact = sequelize.define('Artifact', {
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
  artifactId: {
    type: DataTypes.STRING,
    allowNull: false
  },
  name: {
    type: DataTypes.STRING,
    allowNull: false
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
});

module.exports = Artifact;