// 统一的装备灵宠品质配置
export const itemQualities = {
  // 装备品质
  equipment: {
    common: { 
      name: '凡器', 
      color: '#9e9e9e', 
      statMod: 1.0, 
      maxStatMod: 1.5 
    },
    uncommon: { 
      name: '法器', 
      color: '#4caf50', 
      statMod: 1.2, 
      maxStatMod: 2.0 
    },
    rare: { 
      name: '灵器', 
      color: '#2196f3', 
      statMod: 1.5, 
      maxStatMod: 2.5 
    },
    epic: { 
      name: '极品灵器', 
      color: '#9c27b0', 
      statMod: 2.0, 
      maxStatMod: 3.0 
    },
    legendary: { 
      name: '伪仙器', 
      color: '#ff9800', 
      statMod: 2.5, 
      maxStatMod: 3.5 
    },
    mythic: { 
      name: '仙器', 
      color: '#e91e63', 
      statMod: 3.0, 
      maxStatMod: 4.0 
    }
  },
  
  // 灵宠品质
  pet: {
    common: { 
      name: '凡兽', 
      color: '#9e9e9e', 
      expMod: 1.0, 
      statMod: 1.0 
    },
    uncommon: { 
      name: '妖兽', 
      color: '#4caf50', 
      expMod: 1.2, 
      statMod: 1.2 
    },
    rare: { 
      name: '灵兽', 
      color: '#2196f3', 
      expMod: 1.5, 
      statMod: 1.5 
    },
    epic: { 
      name: '上古异兽', 
      color: '#9c27b0', 
      expMod: 2.0, 
      statMod: 2.0 
    },
    legendary: { 
      name: '瑞兽', 
      color: '#ff9800', 
      expMod: 2.5, 
      statMod: 2.5 
    },
    mythic: { 
      name: '仙兽', 
      color: '#e91e63', 
      expMod: 3.0, 
      statMod: 3.0 
    }
  }
};

// 品质映射表 - 用于处理不同组件间的品质命名差异
export const qualityMappings = {
  // 抽奖系统到背包系统的映射
  gachaToInventory: {
    // 装备品质映射（目前一致，保留以备将来扩展）
    equipment: {
      common: 'common',
      uncommon: 'uncommon',
      rare: 'rare',
      epic: 'epic',
      legendary: 'legendary',
      mythic: 'mythic'
    },
    // 灵宠品质映射（目前一致，保留以备将来扩展）
    pet: {
      common: 'common',
      uncommon: 'uncommon',
      rare: 'rare',
      epic: 'epic',
      legendary: 'legendary',
      mythic: 'mythic'
    }
  },
  
  // 数据库存储到前端显示的映射
  dbToFrontend: {
    // 灵宠品质映射
    pet: {
      // 如果数据库中有不同的命名方式，可以在这里添加映射
      // 示例：'tier1': 'common'
    }
  }
};

// 获取品质配置
export function getQualityConfig(type, quality) {
  if (type === 'pet') {
    return itemQualities.pet[quality] || itemQualities.pet.common;
  } else {
    return itemQualities.equipment[quality] || itemQualities.equipment.common;
  }
}

// 映射品质名称
export function mapQuality(fromSystem, toSystem, itemType, quality) {
  const mapping = qualityMappings[`${fromSystem}To${toSystem}`];
  if (mapping && mapping[itemType] && mapping[itemType][quality]) {
    return mapping[itemType][quality];
  }
  return quality; // 如果没有映射，返回原始品质
}