const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// 宠物模型，用于定义玩家宠物的数据结构
const Pet = sequelize.define('Pet', {
  // 宠物唯一标识符(UUID)
  id: {
    type: DataTypes.UUID,
    defaultValue: DataTypes.UUIDV4,
    primaryKey: true,
    allowNull: false
  },
  // 关联的用户ID
  userId: {
    type: DataTypes.INTEGER,
    allowNull: false,
    references: {
      model: 'Users',
      key: 'id'
    }
  },
  // 宠物类型ID，用于区分不同的宠物种类
  petId: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 宠物名称
  name: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 实体类型，用于区分宠物与其他实体(默认为'pet')
  type: {
    type: DataTypes.STRING,
    defaultValue: 'pet'
  },
  // 稀有度等级，影响宠物的基础属性
  rarity: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 宠物等级(默认为1级)
  level: {
    type: DataTypes.INTEGER,
    defaultValue: 1
  },
  // 星级，进一步细化宠物强度(默认为0星)
  star: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // 当前经验值
  experience: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // 升级所需的最大经验值(默认为100)
  maxExperience: {
    type: DataTypes.INTEGER,
    defaultValue: 100
  },
  // 品质属性，可能包含颜色、前缀等信息
  quality: {
    type: DataTypes.JSON,
    allowNull: true
  },
  // 战斗属性，存储宠物在战斗中的各项能力值
  combatAttributes: {
    type: DataTypes.JSON,
    allowNull: true
  },
  // 是否为活跃状态，决定宠物是否被玩家使用
  isActive: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  // 其他宠物特定字段
  // 力量值，影响宠物的战斗力
  power: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // 升级所需物品数量
  upgradeItems: {
    type: DataTypes.INTEGER,
    defaultValue: 1
  },
  // 宠物描述信息
  description: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // 攻击加成
  attackBonus: {
    type: DataTypes.FLOAT,
    defaultValue: 0
  },
  // 防御加成
  defenseBonus: {
    type: DataTypes.FLOAT,
    defaultValue: 0
  },
  // 生命加成
  healthBonus: {
    type: DataTypes.FLOAT,
    defaultValue: 0
  }
});

module.exports = Pet;