const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// User model
const User = sequelize.define('User', {
  id: {
    type: DataTypes.INTEGER,
    primaryKey: true,
    autoIncrement: true
  },
  username: {
    type: DataTypes.STRING,
    allowNull: false,
    unique: true
  },
  password: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // Player basic info
  playerName: {
    type: DataTypes.STRING,
    defaultValue: '无名修士'
  },
  level: {
    type: DataTypes.INTEGER,
    defaultValue: 1
  },
  realm: {
    type: DataTypes.STRING,
    defaultValue: '练气期一层'
  },
  cultivation: {
    type: DataTypes.FLOAT,
    defaultValue: 0
  },
  maxCultivation: {
    type: DataTypes.FLOAT,
    defaultValue: 100
  },
  spirit: {
    type: DataTypes.FLOAT,
    defaultValue: 0
  },
  spiritStones: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  reinforceStones: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  refinementStones: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // Player attributes
  baseAttributes: {
    type: DataTypes.JSON,
    defaultValue: {
      attack: 10,
      health: 100,
      defense: 5,
      speed: 10
    }
  },
  combatAttributes: {
    type: DataTypes.JSON,
    defaultValue: {
      critRate: 0,
      comboRate: 0,
      counterRate: 0,
      stunRate: 0,
      dodgeRate: 0,
      vampireRate: 0
    }
  },
  combatResistance: {
    type: DataTypes.JSON,
    defaultValue: {
      critResist: 0,
      comboResist: 0,
      counterResist: 0,
      stunResist: 0,
      dodgeResist: 0,
      vampireResist: 0
    }
  },
  specialAttributes: {
    type: DataTypes.JSON,
    defaultValue: {
      healBoost: 0,
      critDamageBoost: 0,
      critDamageReduce: 0,
      finalDamageBoost: 0,
      finalDamageReduce: 0,
      combatBoost: 0,
      resistanceBoost: 0
    }
  },
  // Game statistics
  totalCultivationTime: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  breakthroughCount: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  explorationCount: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  itemsFound: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  eventTriggered: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // Settings and preferences
  isDarkMode: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  autoSellQualities: {
    type: DataTypes.JSON,
    defaultValue: []
  },
  autoReleaseRarities: {
    type: DataTypes.JSON,
    defaultValue: []
  },
  wishlistEnabled: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  selectedWishEquipQuality: {
    type: DataTypes.STRING,
    allowNull: true
  },
  selectedWishPetRarity: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // Unlocks
  unlockedRealms: {
    type: DataTypes.JSON,
    defaultValue: ['练气一层']
  },
  unlockedLocations: {
    type: DataTypes.JSON,
    defaultValue: ['新手村']
  },
  unlockedSkills: {
    type: DataTypes.JSON,
    defaultValue: []
  },
  // New player flag
  isNewPlayer: {
    type: DataTypes.BOOLEAN,
    defaultValue: true
  }
});

module.exports = User;