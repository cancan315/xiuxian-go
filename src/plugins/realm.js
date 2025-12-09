// 境界名称配置
// prettier-ignore
const realms = [
  // 练气期
  { name: '练气期一层', maxCultivation: 100 }, { name: '练气期二层', maxCultivation: 200 },
  { name: '练气期三层', maxCultivation: 300 }, { name: '练气期四层', maxCultivation: 400 },
  { name: '练气期五层', maxCultivation: 500 }, { name: '练气期六层', maxCultivation: 600 },
  { name: '练气期七层', maxCultivation: 700 }, { name: '练气期八层', maxCultivation: 800 },
  { name: '练气期九层', maxCultivation: 900 },
  // 筑基期
  { name: '筑基期一层', maxCultivation: 1000 }, { name: '筑基期二层', maxCultivation: 1200 },
  { name: '筑基期三层', maxCultivation: 1400 }, { name: '筑基期四层', maxCultivation: 1600 },
  { name: '筑基期五层', maxCultivation: 1800 }, { name: '筑基期六层', maxCultivation: 2000 },
  { name: '筑基期七层', maxCultivation: 2200 }, { name: '筑基期八层', maxCultivation: 2400 },
  { name: '筑基期九层', maxCultivation: 2600 },
  // 金丹期
  { name: '金丹期一层', maxCultivation: 3000 }, { name: '金丹期二层', maxCultivation: 3500 },
  { name: '金丹期三层', maxCultivation: 4000 }, { name: '金丹期四层', maxCultivation: 4500 },
  { name: '金丹期五层', maxCultivation: 5000 }, { name: '金丹期六层', maxCultivation: 5500 },
  { name: '金丹期七层', maxCultivation: 6000 }, { name: '金丹期八层', maxCultivation: 6500 },
  { name: '金丹期九层', maxCultivation: 7000 },
  // 元婴期
  { name: '元婴期一层', maxCultivation: 8000 }, { name: '元婴期二层', maxCultivation: 9000 },
  { name: '元婴期三层', maxCultivation: 10000 }, { name: '元婴期四层', maxCultivation: 11000 },
  { name: '元婴期五层', maxCultivation: 12000 }, { name: '元婴期六层', maxCultivation: 13000 },
  { name: '元婴期七层', maxCultivation: 14000 }, { name: '元婴期八层', maxCultivation: 15000 },
  { name: '元婴期九层', maxCultivation: 16000 },
  // 化神期
  { name: '化神期一层', maxCultivation: 18000 }, { name: '化神期二层', maxCultivation: 20000 },
  { name: '化神期三层', maxCultivation: 22000 }, { name: '化神期四层', maxCultivation: 24000 },
  { name: '化神期五层', maxCultivation: 26000 }, { name: '化神期六层', maxCultivation: 28000 },
  { name: '化神期七层', maxCultivation: 30000 }, { name: '化神期八层', maxCultivation: 32000 },
  { name: '化神期九层', maxCultivation: 35000 },
  // 返虚期
  { name: '返虚期一层', maxCultivation: 40000 }, { name: '返虚期二层', maxCultivation: 45000 },
  { name: '返虚期三层', maxCultivation: 50000 }, { name: '返虚期四层', maxCultivation: 55000 },
  { name: '返虚期五层', maxCultivation: 60000 }, { name: '返虚期六层', maxCultivation: 65000 },
  { name: '返虚期七层', maxCultivation: 70000 }, { name: '返虚期八层', maxCultivation: 75000 },
  { name: '返虚期九层', maxCultivation: 80000 },
  // 合体期
  { name: '合体期一层', maxCultivation: 90000 }, { name: '合体期二层', maxCultivation: 100000 },
  { name: '合体期三层', maxCultivation: 110000 }, { name: '合体期四层', maxCultivation: 120000 },
  { name: '合体期五层', maxCultivation: 130000 }, { name: '合体期六层', maxCultivation: 140000 },
  { name: '合体期七层', maxCultivation: 150000 }, { name: '合体期八层', maxCultivation: 160000 },
  { name: '合体期九层', maxCultivation: 170000 },
  // 大乘期
  { name: '大乘期一层', maxCultivation: 200000 }, { name: '大乘期二层', maxCultivation: 230000 },
  { name: '大乘期三层', maxCultivation: 260000 }, { name: '大乘期四层', maxCultivation: 290000 },
  { name: '大乘期五层', maxCultivation: 320000 }, { name: '大乘期六层', maxCultivation: 350000 },
  { name: '大乘期七层', maxCultivation: 380000 }, { name: '大乘期八层', maxCultivation: 410000 },
  { name: '大乘期九层', maxCultivation: 450000 },
  // 渡劫期
  { name: '渡劫期一层', maxCultivation: 500000 }, { name: '渡劫期二层', maxCultivation: 550000 },
  { name: '渡劫期三层', maxCultivation: 600000 }, { name: '渡劫期四层', maxCultivation: 650000 },
  { name: '渡劫期五层', maxCultivation: 700000 }, { name: '渡劫期六层', maxCultivation: 750000 },
  { name: '渡劫期七层', maxCultivation: 800000 }, { name: '渡劫期八层', maxCultivation: 850000 },
  { name: '渡劫期九层', maxCultivation: 900000 },
  // 散仙镜
  { name: '散仙镜一层', maxCultivation: 500000 }, { name: '散仙镜二层', maxCultivation: 550000 },
  { name: '散仙镜三层', maxCultivation: 600000 }, { name: '散仙镜四层', maxCultivation: 650000 },
  { name: '散仙镜五层', maxCultivation: 700000 }, { name: '散仙镜六层', maxCultivation: 750000 },
  { name: '散仙镜七层', maxCultivation: 800000 }, { name: '散仙镜八层', maxCultivation: 850000 },
  { name: '散仙镜九层', maxCultivation: 900000 },
  // 仙人境
  { name: '仙人期一层', maxCultivation: 1000000 }, { name: '仙人期二层', maxCultivation: 1200000 },
  { name: '仙人期三层', maxCultivation: 1400000 }, { name: '仙人期四层', maxCultivation: 1600000 },
  { name: '仙人期五层', maxCultivation: 1800000 }, { name: '仙人期六层', maxCultivation: 2000000 },
  { name: '仙人期七层', maxCultivation: 2200000 }, { name: '仙人期八层', maxCultivation: 2400000 },
  { name: '仙人期九层', maxCultivation: 2600000 },
  // 真仙境
  { name: '真仙期一层', maxCultivation: 3000000 }, { name: '真仙期二层', maxCultivation: 3500000 },
  { name: '真仙期三层', maxCultivation: 4000000 }, { name: '真仙期四层', maxCultivation: 4500000 },
  { name: '真仙期五层', maxCultivation: 5000000 }, { name: '真仙期六层', maxCultivation: 5500000 },
  { name: '真仙期七层', maxCultivation: 6000000 }, { name: '真仙期八层', maxCultivation: 6500000 },
  { name: '真仙期九层', maxCultivation: 7000000 },
  // 金仙境
  { name: '金仙期一层', maxCultivation: 8000000 }, { name: '金仙期二层', maxCultivation: 9000000 },
  { name: '金仙期三层', maxCultivation: 10000000 }, { name: '金仙期四层', maxCultivation: 11000000 },
  { name: '金仙期五层', maxCultivation: 12000000 }, { name: '金仙期六层', maxCultivation: 13000000 },
  { name: '金仙期七层', maxCultivation: 14000000 }, { name: '金仙期八层', maxCultivation: 15000000 },
  { name: '金仙期九层', maxCultivation: 16000000 },
  // 太乙境
  { name: '太乙期一层', maxCultivation: 20000000 }, { name: '太乙期二层', maxCultivation: 24000000 },
  { name: '太乙期三层', maxCultivation: 28000000 }, { name: '太乙期四层', maxCultivation: 32000000 },
  { name: '太乙期五层', maxCultivation: 36000000 }, { name: '太乙期六层', maxCultivation: 40000000 },
  { name: '太乙期七层', maxCultivation: 44000000 }, { name: '太乙期八层', maxCultivation: 48000000 },
  { name: '太乙期九层', maxCultivation: 52000000 },
  // 大罗境
  { name: '大罗期一层', maxCultivation: 60000000 }, { name: '大罗期二层', maxCultivation: 70000000 },
  { name: '大罗期三层', maxCultivation: 80000000 }, { name: '大罗期四层', maxCultivation: 90000000 },
  { name: '大罗期五层', maxCultivation: 100000000 }, { name: '大罗期六层', maxCultivation: 110000000 },
  { name: '大罗期七层', maxCultivation: 120000000 }, { name: '大罗期八层', maxCultivation: 130000000 },
  { name: '大罗期九层', maxCultivation: 140000000 }
]

// 获取境界名称
export const getRealmName = level => {
  return realms[level - 1]
}

export const getRealmLength = () => {
  return realms.length
}

/**
 * 根据境界层级获取对应的境界期名称
 * @param {number} realmLevel - 境界层级 (1-15)
 * @returns {string} 境界期名称
 */
export const getRealmPeriodName = (realmLevel) => {
  const realmPeriods = [
    { period: '练气期', value: 1 },
    { period: '筑基期', value: 2 },
    { period: '金丹期', value: 3 },
    { period: '元婴期', value: 4 },
    { period: '化神期', value: 5 },
    { period: '返虚期', value: 6 },
    { period: '合体期', value: 7 },
    { period: '大乘期', value: 8 },
    { period: '渡劫期', value: 9 },
    { period: '散仙期', value: 10 },
    { period: '仙人期', value: 11 },
    { period: '真仙期', value: 12 },
    { period: '金仙期', value: 13 },
    { period: '太乙期', value: 14 },
    { period: '大罗期', value: 15 }
  ];

  const period = realmPeriods.find(p => p.value === realmLevel);
  return period ? period.period : '未知境界';
}

/**
 * 根据境界名称获取对应的requiredRealm值
 * 1对应炼气期一层-九层
 * 2对应筑基期一层到筑基期九层
 * ...
 * 14对应太乙期一层到太乙九层
 * 15对应大罗期一层到大罗九层
 * @param {string} realmName - 境界名称
 * @returns {number} requiredRealm值
 */
export const getRequiredRealm = (realmName) => {
  if (!realmName) return 1;
  
  // 境界期名称映射到requiredRealm值
  const realmPeriods = [
    { period: '练气期', value: 1 },
    { period: '筑基期', value: 2 },
    { period: '金丹期', value: 3 },
    { period: '元婴期', value: 4 },
    { period: '化神期', value: 5 },
    { period: '返虚期', value: 6 },
    { period: '合体期', value: 7 },
    { period: '大乘期', value: 8 },
    { period: '渡劫期', value: 9 },
    { period: '散仙期', value: 10 },
    { period: '仙人期', value: 11 },
    { period: '真仙期', value: 12 },
    { period: '金仙期', value: 13 },
    { period: '太乙期', value: 14 },
    { period: '大罗期', value: 15 }
  ];
  
  // 查找对应的境界期
  for (const period of realmPeriods) {
    if (realmName.includes(period.period)) {
      return period.value;
    }
  }
  
  // 默认返回1（练气期）
  return 1;
}