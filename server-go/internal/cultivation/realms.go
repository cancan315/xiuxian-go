package cultivation

// RealmConfiguration 境界配置
var realms = []RealmInfo{
	// 练气期 1-9层
	{Level: 1, Name: "练气期一层", MaxCultivation: 100},
	{Level: 2, Name: "练气期二层", MaxCultivation: 200},
	{Level: 3, Name: "练气期三层", MaxCultivation: 300},
	{Level: 4, Name: "练气期四层", MaxCultivation: 400},
	{Level: 5, Name: "练气期五层", MaxCultivation: 500},
	{Level: 6, Name: "练气期六层", MaxCultivation: 600},
	{Level: 7, Name: "练气期七层", MaxCultivation: 700},
	{Level: 8, Name: "练气期八层", MaxCultivation: 800},
	{Level: 9, Name: "练气期九层", MaxCultivation: 900},

	// 筑基期 10-18层
	{Level: 10, Name: "筑基期一层", MaxCultivation: 1000},
	{Level: 11, Name: "筑基期二层", MaxCultivation: 1200},
	{Level: 12, Name: "筑基期三层", MaxCultivation: 1400},
	{Level: 13, Name: "筑基期四层", MaxCultivation: 1600},
	{Level: 14, Name: "筑基期五层", MaxCultivation: 1800},
	{Level: 15, Name: "筑基期六层", MaxCultivation: 2000},
	{Level: 16, Name: "筑基期七层", MaxCultivation: 2200},
	{Level: 17, Name: "筑基期八层", MaxCultivation: 2400},
	{Level: 18, Name: "筑基期九层", MaxCultivation: 2600},

	// 金丹期 19-27层
	{Level: 19, Name: "金丹期一层", MaxCultivation: 3000},
	{Level: 20, Name: "金丹期二层", MaxCultivation: 3500},
	{Level: 21, Name: "金丹期三层", MaxCultivation: 4000},
	{Level: 22, Name: "金丹期四层", MaxCultivation: 4500},
	{Level: 23, Name: "金丹期五层", MaxCultivation: 5000},
	{Level: 24, Name: "金丹期六层", MaxCultivation: 5500},
	{Level: 25, Name: "金丹期七层", MaxCultivation: 6000},
	{Level: 26, Name: "金丹期八层", MaxCultivation: 6500},
	{Level: 27, Name: "金丹期九层", MaxCultivation: 7000},

	// 元婴期 28-36层
	{Level: 28, Name: "元婴期一层", MaxCultivation: 8000},
	{Level: 29, Name: "元婴期二层", MaxCultivation: 9000},
	{Level: 30, Name: "元婴期三层", MaxCultivation: 10000},
	{Level: 31, Name: "元婴期四层", MaxCultivation: 11000},
	{Level: 32, Name: "元婴期五层", MaxCultivation: 12000},
	{Level: 33, Name: "元婴期六层", MaxCultivation: 13000},
	{Level: 34, Name: "元婴期七层", MaxCultivation: 14000},
	{Level: 35, Name: "元婴期八层", MaxCultivation: 15000},
	{Level: 36, Name: "元婴期九层", MaxCultivation: 16000},

	// 化神期 37-45层
	{Level: 37, Name: "化神期一层", MaxCultivation: 18000},
	{Level: 38, Name: "化神期二层", MaxCultivation: 20000},
	{Level: 39, Name: "化神期三层", MaxCultivation: 22000},
	{Level: 40, Name: "化神期四层", MaxCultivation: 24000},
	{Level: 41, Name: "化神期五层", MaxCultivation: 26000},
	{Level: 42, Name: "化神期六层", MaxCultivation: 28000},
	{Level: 43, Name: "化神期七层", MaxCultivation: 30000},
	{Level: 44, Name: "化神期八层", MaxCultivation: 32000},
	{Level: 45, Name: "化神期九层", MaxCultivation: 35000},

	// 返虚期 46-54层
	{Level: 46, Name: "返虚期一层", MaxCultivation: 40000},
	{Level: 47, Name: "返虚期二层", MaxCultivation: 45000},
	{Level: 48, Name: "返虚期三层", MaxCultivation: 50000},
	{Level: 49, Name: "返虚期四层", MaxCultivation: 55000},
	{Level: 50, Name: "返虚期五层", MaxCultivation: 60000},
	{Level: 51, Name: "返虚期六层", MaxCultivation: 65000},
	{Level: 52, Name: "返虚期七层", MaxCultivation: 70000},
	{Level: 53, Name: "返虚期八层", MaxCultivation: 75000},
	{Level: 54, Name: "返虚期九层", MaxCultivation: 80000},

	// 合体期 55-63层
	{Level: 55, Name: "合体期一层", MaxCultivation: 90000},
	{Level: 56, Name: "合体期二层", MaxCultivation: 100000},
	{Level: 57, Name: "合体期三层", MaxCultivation: 110000},
	{Level: 58, Name: "合体期四层", MaxCultivation: 120000},
	{Level: 59, Name: "合体期五层", MaxCultivation: 130000},
	{Level: 60, Name: "合体期六层", MaxCultivation: 140000},
	{Level: 61, Name: "合体期七层", MaxCultivation: 150000},
	{Level: 62, Name: "合体期八层", MaxCultivation: 160000},
	{Level: 63, Name: "合体期九层", MaxCultivation: 170000},

	// 大乘期 64-72层
	{Level: 64, Name: "大乘期一层", MaxCultivation: 200000},
	{Level: 65, Name: "大乘期二层", MaxCultivation: 230000},
	{Level: 66, Name: "大乘期三层", MaxCultivation: 260000},
	{Level: 67, Name: "大乘期四层", MaxCultivation: 290000},
	{Level: 68, Name: "大乘期五层", MaxCultivation: 320000},
	{Level: 69, Name: "大乘期六层", MaxCultivation: 350000},
	{Level: 70, Name: "大乘期七层", MaxCultivation: 380000},
	{Level: 71, Name: "大乘期八层", MaxCultivation: 410000},
	{Level: 72, Name: "大乘期九层", MaxCultivation: 450000},

	// 渡劫期 73-81层
	{Level: 73, Name: "渡劫期一层", MaxCultivation: 500000},
	{Level: 74, Name: "渡劫期二层", MaxCultivation: 550000},
	{Level: 75, Name: "渡劫期三层", MaxCultivation: 600000},
	{Level: 76, Name: "渡劫期四层", MaxCultivation: 650000},
	{Level: 77, Name: "渡劫期五层", MaxCultivation: 700000},
	{Level: 78, Name: "渡劫期六层", MaxCultivation: 750000},
	{Level: 79, Name: "渡劫期七层", MaxCultivation: 800000},
	{Level: 80, Name: "渡劫期八层", MaxCultivation: 850000},
	{Level: 81, Name: "渡劫期九层", MaxCultivation: 900000},

	// 散仙期 82-90层
	{Level: 82, Name: "散仙期一层", MaxCultivation: 500000},
	{Level: 83, Name: "散仙期二层", MaxCultivation: 550000},
	{Level: 84, Name: "散仙期三层", MaxCultivation: 600000},
	{Level: 85, Name: "散仙期四层", MaxCultivation: 650000},
	{Level: 86, Name: "散仙期五层", MaxCultivation: 700000},
	{Level: 87, Name: "散仙期六层", MaxCultivation: 750000},
	{Level: 88, Name: "散仙期七层", MaxCultivation: 800000},
	{Level: 89, Name: "散仙期八层", MaxCultivation: 850000},
	{Level: 90, Name: "散仙期九层", MaxCultivation: 900000},

	// 仙人期 91-99层
	{Level: 91, Name: "仙人期一层", MaxCultivation: 1000000},
	{Level: 92, Name: "仙人期二层", MaxCultivation: 1200000},
	{Level: 93, Name: "仙人期三层", MaxCultivation: 1400000},
	{Level: 94, Name: "仙人期四层", MaxCultivation: 1600000},
	{Level: 95, Name: "仙人期五层", MaxCultivation: 1800000},
	{Level: 96, Name: "仙人期六层", MaxCultivation: 2000000},
	{Level: 97, Name: "仙人期七层", MaxCultivation: 2200000},
	{Level: 98, Name: "仙人期八层", MaxCultivation: 2400000},
	{Level: 99, Name: "仙人期九层", MaxCultivation: 2600000},

	// 真仙期 100-108层
	{Level: 100, Name: "真仙期一层", MaxCultivation: 3000000},
	{Level: 101, Name: "真仙期二层", MaxCultivation: 3500000},
	{Level: 102, Name: "真仙期三层", MaxCultivation: 4000000},
	{Level: 103, Name: "真仙期四层", MaxCultivation: 4500000},
	{Level: 104, Name: "真仙期五层", MaxCultivation: 5000000},
	{Level: 105, Name: "真仙期六层", MaxCultivation: 5500000},
	{Level: 106, Name: "真仙期七层", MaxCultivation: 6000000},
	{Level: 107, Name: "真仙期八层", MaxCultivation: 6500000},
	{Level: 108, Name: "真仙期九层", MaxCultivation: 7000000},

	// 金仙期 109-117层
	{Level: 109, Name: "金仙期一层", MaxCultivation: 8000000},
	{Level: 110, Name: "金仙期二层", MaxCultivation: 9000000},
	{Level: 111, Name: "金仙期三层", MaxCultivation: 10000000},
	{Level: 112, Name: "金仙期四层", MaxCultivation: 11000000},
	{Level: 113, Name: "金仙期五层", MaxCultivation: 12000000},
	{Level: 114, Name: "金仙期六层", MaxCultivation: 13000000},
	{Level: 115, Name: "金仙期七层", MaxCultivation: 14000000},
	{Level: 116, Name: "金仙期八层", MaxCultivation: 15000000},
	{Level: 117, Name: "金仙期九层", MaxCultivation: 16000000},

	// 太乙期 118-126层
	{Level: 118, Name: "太乙期一层", MaxCultivation: 20000000},
	{Level: 119, Name: "太乙期二层", MaxCultivation: 24000000},
	{Level: 120, Name: "太乙期三层", MaxCultivation: 28000000},
	{Level: 121, Name: "太乙期四层", MaxCultivation: 32000000},
	{Level: 122, Name: "太乙期五层", MaxCultivation: 36000000},
	{Level: 123, Name: "太乙期六层", MaxCultivation: 40000000},
	{Level: 124, Name: "太乙期七层", MaxCultivation: 44000000},
	{Level: 125, Name: "太乙期八层", MaxCultivation: 48000000},
	{Level: 126, Name: "太乙期九层", MaxCultivation: 52000000},

	// 大罗期 127-135层
	{Level: 127, Name: "大罗期一层", MaxCultivation: 60000000},
	{Level: 128, Name: "大罗期二层", MaxCultivation: 70000000},
	{Level: 129, Name: "大罗期三层", MaxCultivation: 80000000},
	{Level: 130, Name: "大罗期四层", MaxCultivation: 90000000},
	{Level: 131, Name: "大罗期五层", MaxCultivation: 100000000},
	{Level: 132, Name: "大罗期六层", MaxCultivation: 110000000},
	{Level: 133, Name: "大罗期七层", MaxCultivation: 120000000},
	{Level: 134, Name: "大罗期八层", MaxCultivation: 130000000},
	{Level: 135, Name: "大罗期九层", MaxCultivation: 140000000},
}

// GetRealmByLevel 根据等级获取境界信息
func GetRealmByLevel(level int) *RealmInfo {
	if level < 1 || level > len(realms) {
		return nil
	}
	return &realms[level-1]
}

// GetNextRealm 获取下一个境界
func GetNextRealm(currentLevel int) *RealmInfo {
	nextLevel := currentLevel + 1
	if nextLevel > len(realms) {
		return nil
	}
	return &realms[nextLevel-1]
}

// GetMaxLevel 获取最大等级
func GetMaxLevel() int {
	return len(realms)
}
