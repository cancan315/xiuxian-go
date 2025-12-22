# Loki 日志集成指南

## 📋 快速开始

### 1. 前置条件
- Docker 和 Docker Compose 已安装
- Go 1.25 或更高版本

### 2. 启动 Loki 和 Grafana

```bash
cd server-go
docker-compose up -d loki grafana
```

**服务访问地址：**
- **Loki API**: http://localhost:3100
- **Grafana**: http://localhost:3001 (用户名: admin，密码: admin)

### 3. 更新依赖并运行服务器

```bash
# 在 server-go 目录下执行
go mod download
go mod tidy
go run cmd/server/main.go
```

## 🔧 配置说明

### .env 配置
```env
# Loki 日志服务配置
LOKI_URL=http://localhost:3100        # Loki 服务地址
LOG_LEVEL=debug                        # 日志级别：debug, info, warn, error
```

### Loki 工作原理

1. **日志收集**：应用日志通过 zap logger 输出到标准输出和 Loki
2. **日志存储**：Loki 将日志存储在本地文件系统（`/loki` 目录）
3. **日志查询**：通过 Grafana 的 Loki 数据源查询和可视化日志

## 📊 在 Grafana 中查询日志

### 访问 Grafana
1. 打开浏览器访问 http://localhost:3001
2. 使用默认凭证登录：
   - 用户名：`admin`
   - 密码：`admin`

### 查询日志
在 Grafana 中创建新 Panel 并使用 LogQL 查询：

```logql
# 查询所有应用日志
{job="xiuxian-server"}

# 查询错误日志
{job="xiuxian-server"} | level="error"

# 查询特定级别的日志
{job="xiuxian-server"} | level=~"error|warn"

# 查询包含特定关键字的日志
{job="xiuxian-server"} | "装备"

# 查询响应时间
{job="xiuxian-server"} | json | latency > 100
```

## 🐳 Docker Compose 服务

### 完整启动所有服务
```bash
docker-compose up -d
```

**服务列表：**
- `postgres`: PostgreSQL 数据库 (5432)
- `redis`: Redis 缓存 (6379)
- `loki`: Loki 日志服务 (3100)
- `grafana`: Grafana 可视化 (3001)

### 查看服务日志
```bash
# 查看 Loki 日志
docker-compose logs -f loki

# 查看 Grafana 日志
docker-compose logs -f grafana

# 查看所有服务日志
docker-compose logs -f
```

### 停止服务
```bash
docker-compose down
```

## 📈 日志存储配置

**存储位置：** `/loki` 目录
- `boltdb-shipper-active/`: 活跃索引
- `boltdb-shipper-cache/`: 缓存
- `chunks/`: 日志块

## 🔍 故障排查

### Loki 无法连接
```bash
# 检查 Loki 服务状态
curl http://localhost:3100/loki/api/v1/status

# 如果失败，重启 Loki
docker-compose restart loki
```

### Grafana 看不到日志
1. 检查 LOKI_URL 是否正确
2. 在 Grafana 中验证 Loki 数据源配置
3. 确保后端应用已启动并生成日志

### 日志未出现在 Loki
1. 检查 .env 中的 LOKI_URL 配置
2. 查看应用日志是否有错误
3. 确保 Loki 容器正常运行

## 💡 最佳实践

1. **日志级别设置**：
   - 开发环境：`debug`
   - 生产环境：`info` 或 `warn`

2. **标签策略**：
   - 在 Loki 配置中添加服务标签便于查询
   - 示例：`{job="xiuxian-server", environment="prod"}`

3. **日志保留**：
   - Loki 默认保留 168 小时（7天）的日志
   - 可在 `loki-config.yml` 中修改 `reject_old_samples_max_age`

4. **性能优化**：
   - 适当调高 `ingestion_rate_mb` 处理高日志量
   - 在 Grafana 中使用 Label Filters 加速查询

## 📚 参考资源

- [Loki 官方文档](https://grafana.com/docs/loki/latest/)
- [Grafana Loki 数据源](https://grafana.com/docs/grafana/latest/datasources/loki/)
- [LogQL 查询语言](https://grafana.com/docs/loki/latest/logql/)
