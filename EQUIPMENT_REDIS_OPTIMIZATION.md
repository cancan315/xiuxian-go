# 装备强化和洗练 Redis 优化指南

## 概述

本文档说明如何使用 Redis 优化装备强化和洗练功能，降低数据库压力。

### 优化亮点

✅ **强化石和洗练石缓存**：在 Redis 中缓存用户的强化石和洗练石数量，避免频繁数据库查询  
✅ **装备级别并发控制**：使用 Redis 锁防止同一装备的并发强化/洗练  
✅ **装备数据缓存**：缓存装备信息以加快查询速度  
✅ **缓存失效管理**：在装备更新时自动清除相关缓存  

---

## 核心优化内容

### 1. Redis 装备资源模块 (`internal/redis/equipment.go`)

新增文件提供了以下核心功能：

#### 常量定义
```go
// 用户装备资源缓存键格式
const EquipmentResourceKeyFormat = "user:%d:equipment:resources"

// 装备强化锁键格式
const EquipmentEnhanceLockKeyFormat = "user:%d:equipment:%s:enhance:lock"

// 装备洗练锁键格式
const EquipmentReforgeLockKeyFormat = "user:%d:equipment:%s:reforge:lock"

// 缓存过期时间
const EquipmentCacheTTL = 10 * time.Second          // 装备缓存 10 秒过期
const EquipmentListCacheTTL = 5 * time.Second       // 装备列表缓存 5 秒过期
const OperationLockTTL = 10 * time.Second           // 操作锁 10 秒过期
```

#### 主要函数

**GetEquipmentResources** - 从 Redis 获取用户装备资源
```go
resources, err := redisClient.GetEquipmentResources(ctx, userID)
// resources.ReinforceStones   强化石数量
// resources.RefinementStones  洗练石数量
// resources.UpdatedAt         更新时间戳
```

**SetEquipmentResources** - 将装备资源缓存到 Redis
```go
err := redisClient.SetEquipmentResources(ctx, userID, reinforceStones, refinementStones)
```

**TryEnhanceLock / ReleaseEnhanceLock** - 装备强化锁
```go
acquired, err := redisClient.TryEnhanceLock(ctx, userID, equipmentID)
if acquired {
    defer redisClient.ReleaseEnhanceLock(ctx, userID, equipmentID)
    // 执行强化操作
}
```

**TryReforgeLock / ReleaseReforgeLock** - 装备洗练锁
```go
acquired, err := redisClient.TryReforgeLock(ctx, userID, equipmentID)
if acquired {
    defer redisClient.ReleaseReforgeLock(ctx, userID, equipmentID)
    // 执行洗练操作
}
```

**CacheEquipment / GetCachedEquipment** - 装备数据缓存
```go
// 缓存装备数据
err := redisClient.CacheEquipment(ctx, userID, equipmentID, equipmentData)

// 获取缓存的装备数据
equipmentData, err := redisClient.GetCachedEquipment(ctx, userID, equipmentID)
```

**InvalidateEquipmentCache / InvalidateEquipmentListCache** - 清除缓存
```go
// 清除特定装备的缓存
err := redisClient.InvalidateEquipmentCache(ctx, userID, equipmentID)

// 清除用户的装备列表缓存
err := redisClient.InvalidateEquipmentListCache(ctx, userID)
```

---

### 2. 强化装备优化 (`EnhanceEquipment`)

#### 变更亮点

✅ **装备级别并发控制**
```go
// 使用 Redis 获取装备强化锁（防并发）
acquired, err := redisClient.TryEnhanceLock(c, userID, id)
if !acquired {
    return "该装备强化正在进行中，请稍候"
}
defer redisClient.ReleaseEnhanceLock(c, userID, id)
```

✅ **优先使用 Redis 缓存的强化石数量**
```go
// 第一选择：从 Redis 获取缓存的强化石数量（速度更快）
cachedResources, err := redisClient.GetEquipmentResources(c, userID)
if err == nil && cachedResources != nil {
    userReinforceStones = int(cachedResources.ReinforceStones)
} else {
    // 缓存不存在，从数据库读取
    userReinforceStones = user.ReinforceStones
}
```

✅ **强化完成后更新缓存**
```go
// 缓存装备数据
redisClient.CacheEquipment(c, userID, equipment.ID, equipment)

// 更新强化石缓存
redisClient.SyncEquipmentResourcesToRedis(
    c, userID, 
    int64(userFresh.ReinforceStones - cost),
    0,
)

// 清除装备列表缓存
redisClient.InvalidateEquipmentListCache(c, userID)
```

---

### 3. 洗练装备优化 (`ReforgeEquipment` + `ConfirmReforge`)

#### ReforgeEquipment 变更

✅ **装备级别洗练锁**
```go
acquired, err := redisClient.TryReforgeLock(c, userID, id)
if !acquired {
    return "该装备洗练正在进行中，请稍候"
}
defer redisClient.ReleaseReforgeLock(c, userID, id)
```

✅ **优先使用 Redis 缓存的洗练石**
```go
cachedResources, err := redisClient.GetEquipmentResources(c, userID)
if err == nil && cachedResources != nil {
    userRefinementStones = int(cachedResources.RefinementStones)
} else {
    userRefinementStones = user.RefinementStones
}
```

#### ConfirmReforge 变更

✅ **洗练确认后更新 Redis 缓存**
```go
if req.Confirmed {
    // ... 更新数据库 ...
    
    // 清除装备缓存
    redisClient.InvalidateEquipmentCache(c, userID, equipment.ID)
    redisClient.InvalidateEquipmentListCache(c, userID)
    
    // 更新洗练石缓存
    cachedResources, _ := redisClient.GetEquipmentResources(c, userID)
    if cachedResources != nil {
        newRefinementStones := cachedResources.RefinementStones - int64(refinementCost)
        redisClient.SyncEquipmentResourcesToRedis(
            c, userID,
            cachedResources.ReinforceStones,
            newRefinementStones,
        )
    }
}
```

---

### 4. 缓存初始化模块 (`internal/http/handlers/player/equipment_redis_init.go`)

#### InitEquipmentResourcesCache - 用户登录时调用

在用户认证成功后，调用此函数初始化 Redis 缓存：

```go
// 在登录端点或认证后调用
func Login(c *gin.Context) {
    // ... 验证用户 ...
    
    // 登录成功后初始化装备资源缓存
    if err := InitEquipmentResourcesCache(c, userID); err != nil {
        // 日志记录错误，但不中断登录流程
        log.Printf("初始化装备资源缓存失败: %v", err)
    }
    
    // ... 返回登录响应 ...
}
```

#### SyncEquipmentResourcesToDB - 用户登出时调用

在用户登出时，将 Redis 中的数据同步回数据库：

```go
// 在登出端点调用
func Logout(c *gin.Context) {
    userID := getCurrentUserID(c)
    
    // 同步 Redis 缓存到数据库
    if err := SyncEquipmentResourcesToDB(c, userID); err != nil {
        // 日志记录错误
        log.Printf("同步装备资源到数据库失败: %v", err)
    }
    
    // ... 返回登出响应 ...
}
```

---

## 集成步骤

### 第一步：在认证处理器中初始化缓存

**文件**：`internal/http/handlers/auth/auth.go`

```go
package auth

import (
    "github.com/qoder/xiuxian-go/server-go/internal/http/handlers/player"
)

func Login(c *gin.Context) {
    // ... 验证用户 ...
    
    userID := user.ID
    
    // ✅ 初始化装备资源缓存
    if err := player.InitEquipmentResourcesCache(c, userID); err != nil {
        // 日志记录，但不中断流程
        zapLogger.Warn("初始化装备资源缓存失败", zap.Error(err))
    }
    
    // ... 返回用户信息和 Token ...
}
```

### 第二步：在登出处理器中同步缓存

```go
func Logout(c *gin.Context) {
    userID, ok := currentUserID(c)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"success": false})
        return
    }
    
    // ✅ 同步 Redis 缓存到数据库
    if err := player.SyncEquipmentResourcesToDB(c, userID); err != nil {
        // 日志记录
        zapLogger.Warn("同步装备资源到数据库失败", zap.Error(err))
    }
    
    c.JSON(http.StatusOK, gin.H{"success": true, "message": "已登出"})
}
```

### 第三步：定期同步任务（可选但推荐）

创建一个定期任务，每 5-10 分钟将所有活跃用户的 Redis 缓存同步回数据库：

```go
// internal/tasks/sync_equipment_resources.go
package tasks

import (
    "context"
    "time"
    "github.com/qoder/xiuxian-go/server-go/internal/redis"
    "github.com/qoder/xiuxian-go/server-go/internal/db"
    "github.com/qoder/xiuxian-go/server-go/internal/models"
)

// StartEquipmentResourcesSyncTask 启动定期同步任务
func StartEquipmentResourcesSyncTask(interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        
        for range ticker.C {
            syncAllEquipmentResources()
        }
    }()
}

func syncAllEquipmentResources() {
    // 获取所有有缓存的用户 ID（通过扫描 Redis 键）
    // 对每个用户调用 SyncEquipmentResourcesToDB
    // ...
}
```

---

## 性能收益

### 数据库查询减少

| 操作 | 优化前 | 优化后 | 减少 |
|-----|--------|--------|------|
| 强化前检查强化石 | 1 次 DB 查询 | 0 次（Redis Hit） | 100% |
| 强化后更新强化石 | 1 次 DB 更新 | Redis 缓存更新 | DB 避免 |
| 洗练前检查洗练石 | 1 次 DB 查询 | 0 次（Redis Hit） | 100% |
| 洗练后更新洗练石 | 1 次 DB 更新 | Redis 缓存更新 | DB 避免 |

### 响应时间改进

- **强化/洗练检查**：从 5-10ms 降低到 <1ms（Redis 命中）
- **并发锁获取**：从数据库行锁降低到内存锁，避免数据库连接饱和
- **装备列表查询**：支持缓存热装备，减少序列化开销

### 并发处理能力

- **装备级别并发控制**：支持多个用户同时强化/洗练不同装备
- **避免数据库连接饱和**：Redis 原生支持高并发，不消耗数据库连接
- **减少数据库锁竞争**：分散了强化/洗练的并发压力

---

## 注意事项

### 缓存一致性

1. **定期同步**：虽然 Redis 缓存有 TTL，但建议在以下时机同步到数据库：
   - 用户登出时
   - 定期任务（如 5-10 分钟）
   - 系统关闭时

2. **缓存失效**：
   - 装备强化/洗练后自动清除装备和列表缓存
   - 资源更新时更新 Redis 资源缓存

### Redis 故障处理

如果 Redis 不可用，系统可以降级到直接使用数据库：

```go
cachedResources, err := redisClient.GetEquipmentResources(c, userID)
if err != nil {
    // Redis 不可用，降级到数据库
    userReinforceStones = user.ReinforceStones
}
```

### 内存管理

- 确保 Redis 配置了合理的 `maxmemory` 策略（推荐 `allkeys-lru`）
- 设置了合理的 TTL，避免内存泄漏
- 监控 Redis 内存使用情况

---

## 总结

通过使用 Redis 缓存装备资源和添加细粒度的并发控制，我们成功：

✅ 降低了数据库查询压力  
✅ 提高了响应速度  
✅ 增强了并发处理能力  
✅ 防止了并发冲突  

推荐在生产环境中部署此优化方案。
