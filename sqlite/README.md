# SQLite 模块使用说明

## 📋 数据库结构

### 表名：file_hashes

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键，自增 |
| Hash | string(64) | 文件哈希值（XXH3/MD5），唯一索引 |
| FilePath | string(2048) | 文件完整路径 |
| FileSize | int64 | 文件大小（字节） |
| CreatedAt | time.Time | 创建时间 |

## 🔧 核心函数

### 1. CheckAndAdd - 检查并添加文件（推荐）

```go
// 检查文件是否重复，如果不重复则添加到数据库
// 参数:
//   - hash: 文件哈希值
//   - filePath: 文件路径
//   - fileSize: 文件大小
// 返回:
//   - isDuplicate: 是否重复
//   - originalPath: 如果重复，返回原始文件路径
//   - error: 错误信息

isDuplicate, originalPath, err := sqlite.CheckAndAdd(hash, filePath, fileSize)
if err != nil {
    log.Printf("错误: %v", err)
    return
}

if isDuplicate {
    fmt.Printf("发现重复文件: %s (原件: %s)\n", filePath, originalPath)
    // 删除重复文件
    os.Remove(filePath)
} else {
    fmt.Println("新文件，已记录")
}
```

### 2. CheckExists - 仅检查是否存在

```go
exists, filePath, err := sqlite.CheckExists(hash)
if err != nil {
    log.Printf("错误: %v", err)
    return
}

if exists {
    fmt.Printf("文件已存在: %s\n", filePath)
}
```

### 3. AddFile - 直接添加文件

```go
err := sqlite.AddFile(hash, filePath, fileSize)
if err != nil {
    log.Printf("添加失败: %v", err)
}
```

### 4. GetTotalFiles - 获取文件总数

```go
count, err := sqlite.GetTotalFiles()
if err == nil {
    fmt.Printf("已记录 %d 个文件\n", count)
}
```

### 5. ClearAll - 清空所有记录（慎用！）

```go
err := sqlite.ClearAll()
if err != nil {
    log.Printf("清空失败: %v", err)
}
```

## 🚀 使用示例

### 在 core/dup.go 中使用

```go
package core

import (
    "dedup/sqlite"
    "os"
    // ... 其他导入
)

func Duplicate(root string, dryrun bool) {
    // 初始化 SQLite
    sqlite.SetSqlite()
    
    // 扫描文件...
    fps := finder.FindAllFiles(root)
    
    for _, fp := range fps {
        // 计算文件哈希
        hash := calculateHash(fp)
        
        // 获取文件大小
        fileInfo, _ := os.Stat(fp)
        fileSize := fileInfo.Size()
        
        // 检查并添加
        isDuplicate, originalPath, err := sqlite.CheckAndAdd(hash, fp, fileSize)
        if err != nil {
            log.Printf("错误: %v", err)
            continue
        }
        
        if isDuplicate {
            if !dryrun {
                os.Remove(fp)
                log.Printf("[删除] 重复文件: %s", fp)
            } else {
                log.Printf("[试运行] 发现重复: %s", fp)
            }
        }
    }
}
```

## 💡 优势

1. **零依赖** - 不需要启动任何服务
2. **轻量级** - 数据库就是一个文件
3. **自动建表** - 程序启动时自动创建表结构
4. **唯一索引** - Hash 字段有唯一索引，防止重复插入
5. **简单易用** - API 简洁，易于集成

## 📊 性能

- **查询速度**: O(log n)，对于几万到几十万文件足够快
- **存储空间**: 每个记录约 2KB，10万文件约 200MB
- **并发安全**: GORM 自带连接池，支持并发访问

## ⚠️ 注意事项

1. **数据库文件位置**: 默认在当前目录创建 `duplicate.db`
2. **定期清理**: 可以使用 `ClearAll()` 清空数据重新开始
3. **备份建议**: 重要数据建议定期备份 `duplicate.db` 文件
