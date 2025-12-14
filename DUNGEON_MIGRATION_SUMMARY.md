# 秘境功能后端迁移总结

## 完成内容

### 后端实现 (Go)
已将秘境功能的所有业务逻辑从前端迁移到Go后端。

#### 创建的文件：
1. **`/server-go/internal/dungeon/models.go`**
   - 定义秘境系统的数据结构
   - BuffOption: 增益选项
   - FightResult: 战斗结果
   - BuffConfig: 增益配置
   - DifficultyModifier: 难度修饰符
   - DungeonRequest: API请求类型

2. **`/server-go/internal/dungeon/service.go`**
   - 核心业务逻辑实现
   - `GetRandomBuffs(floor)`: 根据层数生成随机增益选项
   - `SelectBuff(buffID)`: 查找并返回选中增益的详细信息
   - `StartFight(floor, difficulty)`: 执行战斗逻辑
   - `EndDungeon(floor, victory)`: 结束秘境并计算奖励

3. **`/server-go/internal/http/handlers/dungeon/dungeon.go`**
   - HTTP API处理器
   - 5个端点实现

4. **修改 `/server-go/internal/http/router/router.go`**
   - 注册秘境路由组
   - 添加5个API端点

### 前端改造 (Vue3)
改造`src/views/Dungeon.vue`使其使用后端API替代本地计算。

#### 改动的方法：
1. **`startDungeon()`** - 改为async，调用后端API `/api/dungeon/start`
2. **`generateOptions()`** - 改为async，调用后端API `/api/dungeon/buffs/:floor`
3. **`handleRefreshOptions()`** - 改为async，支持等待generateOptions完成
4. **`selectOption(option)`** - 改为async，调用后端API `/api/dungeon/select-buff`
5. **`startFight()`** - 新增，调用后端API `/api/dungeon/fight`
6. **`endDungeon(victory)`** - 新增，调用后端API `/api/dungeon/end`

#### 移除的导入：
- 移除了对`getRandomOptions`的导入（现由后端提供）
- 移除了对本地增益生成函数的依赖

## API端点详情

### 1. 开始秘境 `POST /api/dungeon/start`
**请求：**
```json
{
  "difficulty": "easy|normal|hard|expert"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "floor": 1,
    "difficulty": "normal",
    "refreshCount": 3
  },
  "message": "秘境已开启"
}
```

### 2. 获取增益选项 `GET /api/dungeon/buffs/:floor`
**响应：**
```json
{
  "success": true,
  "data": {
    "floor": 1,
    "options": [
      {
        "id": "heal",
        "name": "气血增加",
        "description": "增加10%血量",
        "type": "common",
        "effect": {"health": 0.1}
      },
      ...
    ]
  }
}
```

### 3. 选择增益 `POST /api/dungeon/select-buff`
**请求：**
```json
{
  "buffID": "heal"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "id": "heal",
    "name": "气血增加",
    "description": "增加10%血量",
    "effects": {"health": 0.1}
  },
  "message": "增益已选择"
}
```

### 4. 开始战斗 `POST /api/dungeon/fight`
**请求：**
```json
{
  "floor": 1,
  "difficulty": "normal"
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "success": true,
    "victory": true,
    "floor": 1,
    "message": "战斗胜利！",
    "rewards": [
      {
        "type": "spirit_stone",
        "amount": 100
      }
    ]
  }
}
```

### 5. 结束秘境 `POST /api/dungeon/end`
**请求：**
```json
{
  "floor": 5,
  "victory": true
}
```

**响应：**
```json
{
  "success": true,
  "data": {
    "floor": 5,
    "totalReward": 250,
    "victory": true,
    "spiritStones": 1500
  },
  "message": "秘境已结束"
}
```

## 增益配置

### 普通增益 (Common) - 8个
- 气血增加、小幅强化、铁壁、疾风、会心、轻身、吸血、战意

### 稀有增益 (Rare) - 6个
- 防御大师、会心精通、无影、连击精通、血魔、震慑

### 史诗增益 (Epic) - 6个
- 极限突破、天道庇护、战斗大师、不朽之躯、天人合一、战圣至尊

## 难度修饰符

| 难度 | 健康修饰符 | 伤害修饰符 | 奖励修饰符 |
|------|-----------|----------|----------|
| easy | 0.8 | 0.8 | 0.8 |
| normal | 1.0 | 1.0 | 1.0 |
| hard | 1.2 | 1.2 | 1.5 |
| expert | 1.5 | 1.5 | 2.0 |

## 增益概率

**基础概率：**
- 普通: 70%
- 稀有: 25%
- 史诗: 5%

**每5层调整：**
- 普通: 50%
- 稀有: 35%
- 史诗: 15%

**每10层调整：**
- 普通: 50%
- 稀有: 30%
- 史诗: 20%

## 后端编译状态
✓ 编译成功，无错误

## 前端改动验证
✓ Dungeon.vue语法正确，所有新方法已正确定义

## 后续工作
1. 集成测试：运行前后端验证API调用是否正常
2. 完整战斗系统：将目前的简化战斗逻辑替换为完整的combat系统
3. 增益效果应用：在后端实现真实的增益属性修改逻辑
4. 会话管理：使用数据库跟踪玩家秘境进度
