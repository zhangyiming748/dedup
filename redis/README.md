# Redis 模块使用说明

## 概述

本模块提供了完整的 Redis 操作封装，包含五种基本数据类型的所有常用操作。

## 快速开始

### 1. 初始化连接

```go
import "dedup/redis"

// 初始化 Redis 连接
err := redis.InitRedis("localhost:6379", "", 0)
if err != nil {
    log.Fatal(err)
}
defer redis.Close()
```

### 2. 字符串操作 (String)

```go
// 设置值（带过期时间）
redis.StringSet("name", "张三", 5*time.Minute)

// 获取值
value, err := redis.StringGet("name")

// 原子性设置（只在键不存在时设置）
success, err := redis.StringSetNX("lock", "locked", 10*time.Second)

// 批量操作
redis.StringMSet(map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
})
```

### 3. 哈希操作 (Hash) - 文件去重核心

```go
// 设置哈希字段
redis.HashSet("user:1001", "name", "张三")
redis.HashSet("user:1001", "age", 25)

// 获取哈希字段
name, _ := redis.HashGet("user:1001", "name")

// 批量设置
redis.HashMSet("user:1001", map[string]interface{}{
    "email": "zhangsan@example.com",
    "phone": "13800138000",
})

// 获取所有字段
allData, _ := redis.HashGetAll("user:1001")
```

#### 文件去重示例

```go
// 检查文件是否重复
isDuplicate, originalPath, err := redis.CheckAndAddFile(md5Hash, filePath)
if err != nil {
    log.Printf("检查失败: %v", err)
    return
}

if isDuplicate {
    fmt.Printf("发现重复文件: %s (原件: %s)\n", filePath, originalPath)
    // 删除重复文件
    os.Remove(filePath)
} else {
    fmt.Printf("新文件已记录: %s\n", filePath)
}
```

### 4. 列表操作 (List)

```go
// 从左侧推入
redis.ListLPush("tasks", "task1", "task2")

// 从右侧弹出
task, err := redis.ListRPop("tasks")

// 获取列表范围
tasks, err := redis.ListRange("tasks", 0, -1) // 获取所有元素
```

### 5. 集合操作 (Set)

```go
// 添加成员
redis.SetAdd("tags", "go", "redis", "docker")

// 检查成员是否存在
exists, _ := redis.SetIsMember("tags", "go")

// 获取所有成员
members, _ := redis.SetMembers("tags")

// 集合运算
intersection, _ := redis.SetInter("set1", "set2") // 交集
union, _ := redis.SetUnion("set1", "set2")         // 并集
diff, _ := redis.SetDiff("set1", "set2")           // 差集
```

### 6. 有序集合操作 (ZSet)

```go
// 添加成员（带分数）
redis.ZSetAdd("leaderboard", 100.5, "player1")
redis.ZSetAdd("leaderboard", 200.3, "player2")

// 获取排名（从小到大）
rank, _ := redis.ZSetRank("leaderboard", "player1")

// 获取前10名（分数从高到低）
top10, _ := redis.ZSetRevRange("leaderboard", 0, 9)

// 获取带分数的排名
top10WithScores, _ := redis.ZSetRevRangeWithScores("leaderboard", 0, 9)
```

### 7. 通用操作

```go
// 设置过期时间
redis.Expire("key", 10*time.Minute)

// 获取剩余生存时间
ttl, _ := redis.TTL("key")

// 删除键
redis.Del("key1", "key2")

// 检查键是否存在
count, _ := redis.Exists("key1", "key2")

// 获取键类型
keyType, _ := redis.Type("key")

// 清空数据库（慎用！）
redis.FlushDB()
```

## API 参考

### 连接管理 (connect.go)
- `InitRedis(addr, password, db)` - 初始化连接
- `Close()` - 关闭连接

### 字符串操作 (string.go)
- `StringSet`, `StringGet`, `StringDel`
- `StringSetNX`, `StringExists`
- `StringIncr`, `StringDecr`
- `StringMSet`, `StringMGet`

### 哈希操作 (hash.go)
- `HashSet`, `HashGet`, `HashDel`
- `HashExists`, `HashGetAll`
- `HashKeys`, `HashVals`, `HashLen`
- `HashMSet`, `HashMGet`
- `HashIncrBy`, `HashIncrByFloat`

### 列表操作 (sort.go)
- `ListLPush`, `ListRPush`
- `ListLPop`, `ListRPop`
- `ListRange`, `ListLen`, `ListIndex`
- `ListRem`, `ListLTrim`

### 集合操作 (set.go)
- `SetAdd`, `SetMembers`, `SetIsMember`
- `SetCard`, `SetRem`, `SetPop`
- `SetDiff`, `SetInter`, `SetUnion`

### 有序集合操作 (zset.go)
- `ZSetAdd`, `ZSetRem`, `ZSetScore`
- `ZSetRank`, `ZSetRevRank`
- `ZSetRange`, `ZSetRevRange`
- `ZSetRangeByScore`, `ZSetCount`

### 通用操作 (general.go)
- `Expire`, `TTL`, `Del`, `Exists`
- `Keys`, `Type`, `Rename`
- `FlushDB`, `FlushAll`, `DBSize`

### 文件去重专用 (file_dedup.go)
- `CheckAndAddFile(md5Hash, filePath)` - 检查并添加文件
- `GetFilePath(md5Hash)` - 获取文件路径
- `GetTotalFiles()` - 获取文件总数
- `ClearAllHashes()` - 清空所有记录

## 注意事项

1. **连接池**: 默认配置了 100 个连接池大小，可根据实际需求调整
2. **超时设置**: 连接超时 5s，读写超时 3s
3. **持久化**: Docker Compose 已配置 AOF + RDB 双重持久化
4. **性能**: 批量操作时使用 Pipeline 可获得更好性能
5. **内存**: 定期清理不需要的键，避免内存泄漏

## 错误处理

所有操作都返回 error，请务必进行错误检查：

```go
value, err := redis.StringGet("key")
if err != nil {
    if err == redis.Nil {
        // 键不存在
        fmt.Println("Key not found")
    } else {
        // 其他错误
        log.Printf("Error: %v", err)
    }
}
```
