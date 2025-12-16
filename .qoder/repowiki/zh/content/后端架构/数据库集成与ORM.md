# 数据库集成与ORM

<cite>
**本文档引用的文件**   
- [db.go](file://server-go/internal/db/db.go)
- [user.go](file://server-go/internal/models/user.go)
- [pill.go](file://server-go/internal/models/pill.go)
- [equipment.go](file://server-go/internal/models/equipment.go)
- [herb.go](file://server-go/internal/models/herb.go)
- [item.go](file://server-go/internal/models/item.go)
- [pet.go](file://server-go/internal/models/pet.go)
- [pill_fragment.go](file://server-go/internal/models/pill_fragment.go)
- [user_alchemy_data.go](file://server-go/internal/models/user_alchemy_data.go)
- [init.sql](file://server-go/init.sql)
- [helpers.go](file://server-go/internal/http/handlers/player/helpers.go)
</cite>

## 目录
1. [数据库连接初始化](#数据库连接初始化)
2. [GORM模型定义与结构体标签](#gorm模型定义与结构体标签)
3. [GORM关联关系](#gorm关联关系)
4. [GORM模型与数据库Schema一致性验证](#gorm模型与数据库schema一致性验证)
5. [GORM查询示例](#gorm查询示例)
6. [性能优化策略](#性能优化策略)
7. [总结](#总结)

## 数据库连接初始化

Go应用通过`db.go`文件中的`Init()`函数初始化与PostgreSQL的连接池。该函数从环境变量中读取数据库连接参数（如主机、端口、数据库名、用户名和密码），并使用GORM库建立连接。如果环境变量未设置，则使用默认值。连接字符串通过`fmt.Sprintf`构建，并使用`gorm.Open`函数打开数据库连接，返回一个`*gorm.DB`实例，该实例被赋值给全局变量`DB`，以便在整个应用中使用。

**Section sources**
- [db.go](file://server-go/internal/db/db.go#L1-L44)

## GORM模型定义与结构体标签

在`models/`目录下，各结构体使用GORM标签来映射数据库字段。例如，在`user.go`中，`User`结构体的`ID`字段使用`gorm:"primaryKey;column:id"`标签指定为主键，并映射到数据库的`id`列。`Username`字段使用`gorm:"size:255;uniqueIndex;not null;column:username"`标签指定大小、唯一索引、非空约束，并映射到`username`列。其他字段如`PlayerName`、`Level`等也使用`gorm:"column:字段名"`标签映射到相应的数据库列。

在`pill.go`中，`Pill`结构体的`ID`字段同样使用`gorm:"primaryKey;column:id"`标签指定为主键，并映射到`id`列。`UserID`字段使用`gorm:"column:user_id"`标签映射到`user_id`列。`PillID`、`Name`、`Description`和`Effect`字段也分别使用`gorm:"column:字段名"`标签映射到相应的数据库列。

**Section sources**
- [user.go](file://server-go/internal/models/user.go#L1-L47)
- [pill.go](file://server-go/internal/models/pill.go#L1-L20)

## GORM关联关系

GORM模型之间通过外键建立关联关系。例如，`Pill`结构体中的`UserID`字段作为外键，关联到`User`结构体的`ID`字段，表示一个用户可以拥有多个丹药。类似地，`Herb`、`Item`、`Pet`等结构体中的`UserID`字段也作为外键，关联到`User`结构体的`ID`字段，表示一个用户可以拥有多个草药、物品和灵宠。

此外，`Equipment`结构体中的`UserID`字段作为外键，关联到`User`结构体的`ID`字段，表示一个用户可以拥有多个装备。`PillFragment`结构体中的`UserID`字段作为外键，关联到`User`结构体的`ID`字段，表示一个用户可以拥有多个丹方残页。

**Section sources**
- [pill.go](file://server-go/internal/models/pill.go#L1-L20)
- [herb.go](file://server-go/internal/models/herb.go#L1-L16)
- [item.go](file://server-go/internal/models/item.go#L1-L25)
- [pet.go](file://server-go/internal/models/pet.go#L1-L34)
- [equipment.go](file://server-go/internal/models/equipment.go#L1-L33)
- [pill_fragment.go](file://server-go/internal/models/pill_fragment.go#L1-L12)

## GORM模型与数据库Schema一致性验证

通过对比`models/`目录下的结构体定义和`init.sql`中的表结构，可以验证GORM模型与数据库Schema的一致性。例如，`user.go`中的`User`结构体与`init.sql`中的`users`表在字段名称、数据类型和约束上完全一致。`pill.go`中的`Pill`结构体与`init.sql`中的`pills`表在字段名称、数据类型和约束上也完全一致。

其他模型如`Herb`、`Item`、`Pet`、`Equipment`和`PillFragment`也与`init.sql`中的相应表结构保持一致。这确保了GORM模型能够正确地映射到数据库表，并且在进行CRUD操作时不会出现字段不匹配的问题。

**Section sources**
- [user.go](file://server-go/internal/models/user.go#L1-L47)
- [pill.go](file://server-go/internal/models/pill.go#L1-L20)
- [herb.go](file://server-go/internal/models/herb.go#L1-L16)
- [item.go](file://server-go/internal/models/item.go#L1-L25)
- [pet.go](file://server-go/internal/models/pet.go#L1-L34)
- [equipment.go](file://server-go/internal/models/equipment.go#L1-L33)
- [pill_fragment.go](file://server-go/internal/models/pill_fragment.go#L1-L12)
- [init.sql](file://server-go/init.sql#L1-L166)

## GORM查询示例

GORM提供了丰富的查询方法，可以方便地进行数据的增删改查操作。例如，创建用户可以通过以下代码实现：

```go
user := models.User{
    Username: "example",
    Password: "password",
    PlayerName: "Example Player",
}
result := db.DB.Create(&user)
if result.Error != nil {
    // 处理错误
}
```

更新修炼数据可以通过以下代码实现：

```go
var user models.User
result := db.DB.First(&user, userID)
if result.Error != nil {
    // 处理错误
}
user.Cultivation += 10
result = db.DB.Save(&user)
if result.Error != nil {
    // 处理错误
}
```

这些示例展示了如何使用GORM进行基本的数据库操作，确保数据的正确性和一致性。

**Section sources**
- [helpers.go](file://server-go/internal/http/handlers/player/helpers.go#L224-L471)

## 性能优化策略

为了提高数据库操作的性能，可以采用以下策略：

1. **预加载（Preload）**：在查询主模型时，可以使用`Preload`方法一次性加载关联的子模型，避免N+1查询问题。例如，查询用户及其所有丹药时，可以使用`db.DB.Preload("Pills").Find(&users)`。

2. **索引使用**：在经常用于查询条件的字段上创建索引，可以显著提高查询速度。例如，在`users`表的`username`字段和`last_spirit_gain_time`字段上创建索引。

3. **连接池配置**：合理配置数据库连接池的大小，可以避免因连接过多或过少导致的性能问题。GORM默认使用连接池，可以通过`gorm.Config`进行配置。

这些策略有助于提高应用的整体性能，确保在高并发场景下仍能保持良好的响应速度。

**Section sources**
- [init.sql](file://server-go/init.sql#L154-L165)
- [db.go](file://server-go/internal/db/db.go#L1-L44)

## 总结

本文档详细介绍了Go应用如何通过`db.go`中的`Init()`函数初始化并管理与PostgreSQL的连接池，以及GORM在`models/`目录下各结构体中的使用方式。通过验证GORM模型与`init.sql`中表结构的一致性，确保了数据模型的正确性。同时，提供了GORM查询示例和性能优化策略，帮助开发者更好地理解和使用GORM进行数据库操作。