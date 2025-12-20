package battle

// BattleLog 战斗日志管理
type BattleLog struct {
	entries []string
}

// NewBattleLog 创建新的战斗日志
func NewBattleLog() *BattleLog {
	return &BattleLog{
		entries: make([]string, 0),
	}
}

// Add 添加日志项
func (bl *BattleLog) Add(entry string) {
	bl.entries = append(bl.entries, entry)
}

// GetAll 获取所有日志
func (bl *BattleLog) GetAll() []string {
	return bl.entries
}

// Clear 清空日志
func (bl *BattleLog) Clear() {
	bl.entries = make([]string, 0)
}
