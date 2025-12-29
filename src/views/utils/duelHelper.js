/**
 * 根据境界等级获取标签类型
 * @param {number} level - 境界等级
 * @returns {string} 标签类型
 */
export const getRealmTagType = (level) => {
  if (level >= 8) return 'error'    // 高境界显示为红色
  if (level >= 5) return 'warning'  // 中等境界显示为黄色
  if (level >= 3) return 'info'     // 低境界显示为蓝色
  return 'success'                  // 最低境界显示为绿色
}

/**
 * 根据难度获取标签类型
 * @param {string} difficulty - 难度
 * @returns {string} 标签类型
 */
export const getDifficultyTagType = (difficulty) => {
  const map = {
    normal: 'info',
    hard: 'warning',
    nightmare: 'error'
  }
  return map[difficulty] || 'default'
}

/**
 * 根据难度获取中文名称
 * @param {string} difficulty - 难度
 * @returns {string} 中文名称
 */
export const getDifficultyName = (difficulty) => {
  const map = {
    normal: '普通',
    hard: '困难',
    nightmare: '噩梦'
  }
  return map[difficulty] || difficulty
}

/**
 * 根据日志类型获取标签类型
 * @param {string} type - 日志类型
 * @returns {string} 标签类型
 */
export const getLogTagType = (type) => {
  const map = {
    attack: 'error',  // 攻击日志显示为红色
    heal: 'success',  // 治疗日志显示为绿色
    buff: 'warning',  // 增益日志显示为黄色
    info: 'info',     // 信息日志显示为蓝色
    special: 'primary' // 特殊日志显示为紫色
  }
  return map[type] || 'default'
}

/**
 * 难度选项列表
 */
export const difficulties = [
  { label: '普通', value: 'normal' },
  { label: '困难', value: 'hard' },
  { label: '噩梦', value: 'nightmare' }
]
