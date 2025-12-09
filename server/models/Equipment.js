const { DataTypes } = require('sequelize');
const sequelize = require('./database');

// 装备模型，用于定义玩家装备的数据结构
const Equipment = sequelize.define('Equipment', {
  // 装备唯一标识符(UUID)
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
  // 装备类型ID，
  equipmentId: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 装备名称
  name: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 实体类型，用于区分装备与其他实体
  type: {
    type: DataTypes.STRING,
    allowNull: false
  },
  // 装备槽位类型(faqi, guanjin, daopao, yunlv, fabao)
  slot: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // 抽奖系统中的装备类型字段，用于与抽奖系统保持一致性(faqi, guanjin, daopao, yunlv, fabao)
  equipType: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // 装备详细信息
  details: {
    type: DataTypes.JSON,
    allowNull: true
  },
  // 装备属性，存储装备的各项能力值
  stats: {
    type: DataTypes.JSON,
    allowNull: true
  },
  // 装备品质(mythic:仙器, legendary:伪仙器, epic:极品灵器, rare:灵器, uncommon:法器, common:凡器)
  quality: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // 强化等级
  enhanceLevel: {
    type: DataTypes.INTEGER,
    defaultValue: 0
  },
  // 是否已装备
  equipped: {
    type: DataTypes.BOOLEAN,
    defaultValue: false
  },
  // 装备描述信息
  description: {
    type: DataTypes.STRING,
    allowNull: true
  },
  // 装备等级要求
  requiredRealm: {
    type: DataTypes.INTEGER,
    defaultValue: 1
  },
  // 装备等级
  level: {
    type: DataTypes.INTEGER,
    defaultValue: 1
  }
});

module.exports = Equipment;