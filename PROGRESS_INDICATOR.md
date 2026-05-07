# 长时间运行操作的进度提示实现

## 🎯 问题描述

当扫描巨大目录时，`finder.FindAllFiles(root)` 可能需要很长时间（几分钟甚至更久）。在这段时间内，用户看不到任何输出，可能会以为：
- ❌ 程序卡死了
- ❌ 程序崩溃了
- ❌ 出现错误了

## ✅ 解决方案

使用 **goroutine + channel** 在后台定期打印进度提示，让用户知道程序仍在正常运行。

## 📝 实现代码

```go
// 步骤1: 获取所有文件（可能很耗时，使用 goroutine 显示进度提示）
fmt.Println("正在扫描文件，请稍候...")

var fps []string
done := make(chan bool)

// 启动后台 goroutine 定期打印提示
go func() {
    seconds := 0
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(5 * time.Second)
            seconds += 5
            fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)
        }
    }
}()

// 执行文件扫描（阻塞操作）
fps = finder.FindAllFiles(root)

// 通知后台 goroutine 停止
close(done)
fmt.Println() // 换行

log.Printf("✓ 扫描完成，找到 %d 个文件", len(fps))
```

## 🔍 工作原理

### 1. 创建信号通道
```go
done := make(chan bool)
```
用于通知后台 goroutine 何时停止。

### 2. 启动后台 goroutine
```go
go func() {
    seconds := 0
    for {
        select {
        case <-done:
            return  // 收到信号，退出
        default:
            time.Sleep(5 * time.Second)
            seconds += 5
            fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)
        }
    }
}()
```

**关键点：**
- `go func()` - 在后台运行，不阻塞主线程
- `select` - 监听 `done` 通道
- `time.Sleep(5 * time.Second)` - 每 5 秒打印一次
- `\r` - 回车符，覆盖同一行，保持界面整洁

### 3. 执行长时间操作
```go
fps = finder.FindAllFiles(root)
```
这是阻塞调用，可能需要很长时间。在此期间，后台 goroutine 会每 5 秒打印一次提示。

### 4. 停止后台 goroutine
```go
close(done)  // 关闭通道，发送信号
fmt.Println() // 换行
```

## 📊 运行效果

### 实际输出示例

```
正在连接 Redis...
✓ Redis 连接成功

正在扫描文件，请稍候...
  ⏳ 扫描中... 已等待 5 秒
  ⏳ 扫描中... 已等待 10 秒
  ⏳ 扫描中... 已等待 15 秒
  ⏳ 扫描中... 已等待 20 秒
  ⏳ 扫描中... 已等待 25 秒
✓ 扫描完成，找到 150000 个文件

[1/150000] 处理文件: /path/to/file1.txt
[新增] 已记录文件哈希: /path/to/file1.txt
...
```

### 视觉效果

由于使用了 `\r`（回车符），提示会在同一行更新：

```
  ⏳ 扫描中... 已等待 5 秒    ← 5秒后
  ⏳ 扫描中... 已等待 10 秒   ← 10秒后（覆盖上一行）
  ⏳ 扫描中... 已等待 15 秒   ← 15秒后（覆盖上一行）
```

用户看到的是时间在不断更新，而不是多行输出。

## 💡 关键技术点

### 1. Goroutine（协程）
```go
go func() {
    // 后台运行的代码
}()
```
- 轻量级线程
- 与主线程并发执行
- 不会阻塞主流程

### 2. Channel（通道）
```go
done := make(chan bool)
close(done)  // 发送信号
```
- 用于 goroutine 之间的通信
- `close(done)` 会让所有监听该通道的 `select` 收到信号

### 3. Select 语句
```go
select {
case <-done:
    return  // 收到停止信号
default:
    // 继续执行
}
```
- 非阻塞地检查通道
- 如果 `done` 被关闭，立即执行 `case` 分支
- 否则执行 `default` 分支

### 4. 回车符 `\r`
```go
fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)
```
- `\r` 将光标移回行首
- 新输出会覆盖旧输出
- 保持终端界面整洁

## 🎨 自定义配置

### 调整提示间隔

```go
// 每 5 秒提示一次（当前）
time.Sleep(5 * time.Second)

// 改为每 10 秒提示一次
time.Sleep(10 * time.Second)

// 改为每 2 秒提示一次（更频繁）
time.Sleep(2 * time.Second)
```

### 自定义提示消息

```go
// 当前样式
fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)

// 更详细的样式
fmt.Printf("\r  🔍 正在扫描目录: %s (已等待 %d 秒)", root, seconds)

// 简洁样式
fmt.Printf("\r  扫描中... %ds", seconds)
```

### 添加更多统计信息

如果需要显示更多信息（如已扫描的文件数），可以修改为：

```go
go func() {
    seconds := 0
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-done:
            return
        case <-ticker.C:
            seconds += 5
            // 这里可以访问共享变量获取更多统计信息
            fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)
        }
    }
}()
```

## ⚠️ 注意事项

### 1. 资源清理
确保在长时间操作完成后关闭通道：
```go
fps = finder.FindAllFiles(root)
close(done)  // ← 必须调用，否则 goroutine 会泄漏
```

### 2. 线程安全
如果需要在 goroutine 中访问共享变量，需要使用互斥锁：
```go
var mu sync.Mutex
var scannedCount int

go func() {
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(5 * time.Second)
            mu.Lock()
            count := scannedCount
            mu.Unlock()
            fmt.Printf("\r  扫描中... 已扫描 %d 个文件", count)
        }
    }
}()
```

### 3. 错误处理
如果长时间操作可能失败，需要确保 goroutine 能正常退出：
```go
fps, err := finder.FindAllFiles(root)
close(done)  // 无论成功还是失败，都要关闭通道

if err != nil {
    log.Printf("扫描失败: %v", err)
    return
}
```

## 🚀 其他应用场景

这种模式可以用于任何长时间运行的操作：

### 1. 大文件复制
```go
go func() {
    seconds := 0
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(5 * time.Second)
            seconds += 5
            fmt.Printf("\r  📦 复制中... 已等待 %d 秒", seconds)
        }
    }
}()

copyLargeFile(src, dst)
close(done)
```

### 2. 数据库迁移
```go
go func() {
    seconds := 0
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(10 * time.Second)
            seconds += 10
            fmt.Printf("\r  🗄️  迁移中... 已等待 %d 秒", seconds)
        }
    }
}()

migrateDatabase()
close(done)
```

### 3. 网络下载
```go
go func() {
    seconds := 0
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(5 * time.Second)
            seconds += 5
            fmt.Printf("\r  ⬇️  下载中... 已等待 %d 秒", seconds)
        }
    }
}()

downloadFile(url)
close(done)
```

## 📈 性能影响

### 资源消耗
- **CPU**: 几乎无影响（大部分时间在 sleep）
- **内存**: 极小（一个 goroutine 约几 KB）
- **I/O**: 每 5 秒一次控制台输出，可忽略

### 对比
| 方案 | 用户体验 | 资源消耗 | 实现复杂度 |
|------|---------|---------|-----------|
| 无提示 | ❌ 差 | 无 | 简单 |
| 此方案 | ✅ 好 | 极低 | 中等 |
| 真实进度条 | ✅✅ 最好 | 低 | 复杂 |

## 🎉 总结

通过简单的 goroutine + channel 机制，我们实现了：
- ✅ 用户友好的进度提示
- ✅ 避免用户误以为程序卡死
- ✅ 极低的性能开销
- ✅ 简洁优雅的代码
- ✅ 易于扩展和定制

这是一个在生产环境中非常实用的技巧！
