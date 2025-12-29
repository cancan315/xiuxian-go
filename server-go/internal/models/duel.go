package models

import "time"

// BattleRecord 战斗记录模型
type BattleRecord struct {
	ID           int64     `db:"id" json:"id"`
	PlayerID     int64     `db:"player_id" json:"playerId"`
	OpponentID   int64     `db:"opponent_id" json:"opponentId"`
	OpponentName string    `db:"opponent_name" json:"opponentName"`
	Result       string    `db:"result" json:"result"`          // '胜利' 或 '失败'
	BattleType   string    `db:"battle_type" json:"battleType"` // 'pvp' 或 'pve'
	Rewards      string    `db:"rewards" json:"rewards"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

// DuelStats 斗法统计
type DuelStats struct {
	TotalBattles     int `json:"totalBattles"`
	Wins             int `json:"wins"`
	Losses           int `json:"losses"`
	WinRate          int `json:"winRate"`
	CurrentWinStreak int `json:"currentWinStreak"`
	MaxWinStreak     int `json:"maxWinStreak"`
}

// Opponent 对手信息
type Opponent struct {
	ID                int64       `json:"id"`
	Name              string      `json:"name"`
	Level             int         `json:"level"`
	Cultivation       int64       `json:"cultivation"`
	MaxCultivation    int64       `json:"maxCultivation"`
	SpiritStones      int64       `json:"spiritStones"`
	BaseAttributes    interface{} `json:"baseAttributes"`
	CombatAttributes  interface{} `json:"combatAttributes"`
	CombatResistance  interface{} `json:"combatResistance"`
	SpecialAttributes interface{} `json:"specialAttributes"`
}
