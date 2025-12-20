package battle

// 此文件提供了从旧类型到新battle包类型的适配

// ToCombatStats 从玩家数据转换为战斗属性
func ToCombatStats(
	health, attack, defense, speed float64,
	critRate, comboRate, counterRate, stunRate, dodgeRate, vampireRate float64,
	critResist, comboResist, counterResist, stunResist, dodgeResist, vampireResist float64,
	healBoost, critDamageBoost, critDamageReduce, finalDamageBoost, finalDamageReduce, combatBoost, resistanceBoost float64,
) *CombatStats {
	return &CombatStats{
		Health:            health,
		MaxHealth:         health,
		Damage:            attack,
		Defense:           defense,
		Speed:             speed,
		CritRate:          critRate,
		ComboRate:         comboRate,
		CounterRate:       counterRate,
		StunRate:          stunRate,
		DodgeRate:         dodgeRate,
		VampireRate:       vampireRate,
		CritResist:        critResist,
		ComboResist:       comboResist,
		CounterResist:     counterResist,
		StunResist:        stunResist,
		DodgeResist:       dodgeResist,
		VampireResist:     vampireResist,
		HealBoost:         healBoost,
		CritDamageBoost:   critDamageBoost,
		CritDamageReduce:  critDamageReduce,
		FinalDamageBoost:  finalDamageBoost,
		FinalDamageReduce: finalDamageReduce,
		CombatBoost:       combatBoost,
		ResistanceBoost:   resistanceBoost,
	}
}
