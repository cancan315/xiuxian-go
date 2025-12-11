package models

type QualityConfig struct {
	Name       string
	Color      string
	StatMod    float64
	MaxStatMod *float64 // 只对装备有效
	ExpMod     *float64 // 只对灵宠有效
}

var (
	// 装备品质配置
	EquipmentQualityConfigs = map[string]QualityConfig{
		"common": {
			Name:       "凡器",
			Color:      "#9e9e9e",
			StatMod:    1.0,
			MaxStatMod: float64Ptr(1.5),
		},
		"uncommon": {
			Name:       "法器",
			Color:      "#4caf50",
			StatMod:    1.2,
			MaxStatMod: float64Ptr(2.0),
		},
		"rare": {
			Name:       "灵器",
			Color:      "#2196f3",
			StatMod:    1.5,
			MaxStatMod: float64Ptr(2.5),
		},
		"epic": {
			Name:       "极品灵器",
			Color:      "#9c27b0",
			StatMod:    2.0,
			MaxStatMod: float64Ptr(3.0),
		},
		"legendary": {
			Name:       "伪仙器",
			Color:      "#ff9800",
			StatMod:    2.5,
			MaxStatMod: float64Ptr(3.5),
		},
		"mythic": {
			Name:       "仙器",
			Color:      "#e91e63",
			StatMod:    3.0,
			MaxStatMod: float64Ptr(4.0),
		},
	}

	// 灵宠品质配置
	PetQualityConfigs = map[string]QualityConfig{
		"common": {
			Name:    "凡兽",
			Color:   "#9e9e9e",
			StatMod: 1.0,
			ExpMod:  float64Ptr(1.0),
		},
		"uncommon": {
			Name:    "妖兽",
			Color:   "#4caf50",
			StatMod: 1.2,
			ExpMod:  float64Ptr(1.2),
		},
		"rare": {
			Name:    "灵兽",
			Color:   "#2196f3",
			StatMod: 1.5,
			ExpMod:  float64Ptr(1.5),
		},
		"epic": {
			Name:    "上古异兽",
			Color:   "#9c27b0",
			StatMod: 2.0,
			ExpMod:  float64Ptr(2.0),
		},
		"legendary": {
			Name:    "瑞兽",
			Color:   "#ff9800",
			StatMod: 2.5,
			ExpMod:  float64Ptr(2.5),
		},
		"mythic": {
			Name:    "仙兽",
			Color:   "#e91e63",
			StatMod: 3.0,
			ExpMod:  float64Ptr(3.0),
		},
	}
)

// 辅助函数用于创建float64指针
func float64Ptr(v float64) *float64 {
	return &v
}