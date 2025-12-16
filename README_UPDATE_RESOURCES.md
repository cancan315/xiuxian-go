# 玩家资源更新脚本使用说明

这个脚本用于给指定ID的玩家增加各种资源，包括灵石、强化石、洗练石、灵宠精华和灵力。

## 安装依赖

在项目根目录下运行：

```bash
npm install
```

## 配置数据库连接

1. 复制 `.env.example` 文件为 `.env`：
   ```bash
   cp .env.example .env
   ```

2. 根据你的实际数据库配置修改 `.env` 文件中的参数：
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=xiuxian_db
   DB_USER=your_database_user
   DB_PASSWORD=your_database_password
   ```

## 使用方法

### 命令行方式

```bash
node update-player-resources.js <playerId> <spiritStones> <reinforceStones> <refinementStones> <petEssence> <spirit>
```

参数说明：
- `playerId`: 玩家ID（正整数）
- `spiritStones`: 要增加的灵石数量
- `reinforceStones`: 要增加的强化石数量
- `refinementStones`: 要增加的洗练石数量
- `petEssence`: 要增加的灵宠精华数量
- `spirit`: 要增加的灵力值

### 示例

给ID为1的玩家增加1000灵石、500强化石、300洗练石、200灵宠精华和100灵力：

```bash
node update-player-resources.js 1 1000 500 300 200 100
```

### 使用npm脚本

你也可以使用npm脚本来运行：

```bash
npm run update-resources 1 1000 500 300 200 100
```

## 注意事项

1. 脚本会对指定玩家的资源进行累加操作，而不是直接设置值
2. 如果某个资源的数量为负数，将会减少该资源
3. 确保数据库服务正在运行并且可以连接
4. 确保提供的玩家ID存在于数据库中