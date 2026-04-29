# dedup

一个基于 Go 语言开发的命令行文件去重工具，使用 **Redis** 和 **XXH3** 哈希算法高效识别并删除重复文件。

## 功能特性

- 🔍 **智能去重**: 使用 XXH3 哈希算法（比 MD5 快 3-5 倍）精确识别内容完全相同的文件
- 🚀 **Redis 支持**: 利用 Redis Hash 实现 O(1) 时间复杂度的重复检测，支持分布式场景
- 🛡️ **安全模式**: 支持试运行模式 (dry-run)，先预览再执行删除操作
- 📊 **详细日志**: 完整的操作日志记录，支持文件和控制台双输出
- 🎯 **递归扫描**: 自动遍历指定目录下的所有子文件夹
- ⚡ **高效处理**: 保留首个发现的原件，删除后续重复文件
- 📝 **统计报告**: 每个文件夹处理完成后输出详细统计数据
- 💬 **进度提示**: 长时间扫描时显示实时进度，避免用户误以为程序卡死

## 安装

### 方式一: 下载预编译版本 (推荐)

从 [GitHub Releases](https://github.com/zhangyiming748/dedup/releases/latest) 下载最新版本的二进制文件:

| 平台 | 架构 | 下载链接 |
|------|------|----------|
| Windows | amd64 | [dedup_windows_amd64.exe](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_windows_amd64.exe) |
| Windows | arm64 | [dedup_windows_arm64.exe](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_windows_arm64.exe) |
| Linux | amd64 | [dedup_linux_amd64](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_linux_amd64) |
| Linux | arm64 | [dedup_linux_arm64](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_linux_arm64) |
| macOS | amd64 | [dedup_darwin_amd64](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_darwin_amd64) |
| macOS | arm64 | [dedup_darwin_arm64](https://github.com/zhangyiming748/dedup/releases/latest/download/dedup_darwin_arm64) |

**提示**: 以上链接会自动指向最新版本, 无需手动查找.

#### Linux/macOS 用户

下载后添加执行权限:

```bash
chmod +x dedup_linux_amd64
./dedup_linux_amd64 --help
```

#### Windows 用户

直接双击运行或在命令行中使用:

```cmd
dedup_windows_amd64.exe --help
```

### 方式二: 从源码编译

```bash
# 克隆仓库
git clone https://github.com/zhangyiming748/dedup.git
cd dedup

# 编译
go build -o dedup.exe .

# Windows 用户可以直接使用编译好的 dedup.exe
# Linux/Mac 用户可以使用 ./dedup
```

### 依赖要求

- Go 1.26.2 或更高版本
- **Redis 7.0+** (用于存储文件哈希映射)
  - 可通过 Docker Compose 快速启动：`docker-compose up -d`
  - 配置文件：`redis.conf`（已优化，支持 AOF + RDB 持久化）

## 使用方法

### 前置准备：启动 Redis

在使用去重功能前，需要先启动 Redis 服务：

```bash
# 启动 Redis（后台运行）
docker-compose up -d

# 验证 Redis 是否运行
docker ps | grep redis

# 查看日志
docker-compose logs -f redis
```

**提示**：Redis 配置文件 `redis.conf` 已针对 dedup 项目优化，支持 AOF + RDB 双重持久化。

### 基本语法

```bash
dedup [flags]
```

### 命令行参数

| 参数 | 短参数 | 说明 | 默认值 |
|------|--------|------|--------|
| `--dir` | `-d` | 要扫描的根目录路径 | 必填 |
| `--test` | `-t` | 试运行模式, 只打印不删除 | `false` |
| `--help` | `-h` | 显示帮助信息 | - |

### 使用示例

#### 1. 查看帮助信息

```bash
./dedup --help
```

#### 2. 试运行模式 (推荐首次使用)

在正式删除前, 先使用试运行模式预览将要删除的文件:

```bash
./dedup -t -d /path/to/scan
```

#### 3. 正式执行去重

确认无误后, 移除 `-t` 参数执行实际删除:

```bash
./dedup -d /path/to/scan
```

#### 4. 扫描当前目录

```bash
./dedup -d .
```

## 工作原理

### 去重流程

1. **Redis 连接**: 程序启动时自动连接到 Redis 服务
2. **目录扫描**: 递归查找指定根目录下的所有文件（显示进度提示）
3. **XXH3 计算**: 为每个文件计算 XXH3 哈希值（比 MD5 快 3-5 倍）
4. **重复检测**: 使用 Redis Hash 存储哈希映射
   - Key: `"dupmission"` (固定)
   - Field: 文件的 XXH3 哈希值（字符串形式）
   - Value: 文件路径
   - 使用 `HSETNX` 原子性检查并插入：
     - 返回 `true` → 新文件，记录到 Redis
     - 返回 `false` → 重复文件，可以删除
5. **文件处理**: 根据运行模式处理重复文件
   - 试运行模式: 仅打印信息，不删除
   - 正式模式: 删除重复文件，保留原件

### 技术优势

#### 1. Redis Hash 结构
- **O(1) 时间复杂度**: 哈希查找和插入都是常数时间
- **原子性操作**: `HSETNX` 保证检查和插入的原子性，无竞态条件
- **持久化支持**: AOF + RDB 双重持久化，数据不丢失
- **分布式友好**: 多个实例可共享同一 Redis

#### 2. XXH3 哈希算法
- **速度快**: 比 MD5 快 3-5 倍
- **碰撞率低**: 64 位哈希值，碰撞概率极低
- **非 CGO**: 纯 Go 实现，跨平台兼容性好

#### 3. 进度提示
- **长时间操作反馈**: 扫描大目录时每 5 秒显示一次进度
- **用户友好**: 避免用户误以为程序卡死
- **资源消耗低**: 后台 goroutine，几乎无性能影响

### 日志系统

程序使用 `lumberjack` 实现日志轮转, 同时输出到:
- **控制台**: 实时查看处理进度
- **日志文件**: `dedup.log`, 完整记录所有操作

#### 日志内容包括:

- ✅ 程序启动和退出信息
- ✅ 命令行参数解析结果
- ✅ 每个文件夹的处理进度和统计
- ✅ 每个文件的 MD5 计算过程
- ✅ 重复文件检测结果 (原件 vs 副本)
- ✅ 文件删除操作及结果
- ✅ 错误信息和异常处理

#### 日志示例

```
========== 程序启动 ==========
[启动] dedup 文件去重工具
[启动] 命令行: [dedup.exe -t -d .]
[启动] 工作目录: C:\Users\example

正在连接 Redis...
✓ Redis 连接成功

正在扫描文件，请稍候...
  ⏳ 扫描中... 已等待 5 秒
  ⏳ 扫描中... 已等待 10 秒
  ⏳ 扫描中... 已等待 15 秒
✓ 扫描完成，找到 150000 个文件

[1/150000] 处理文件: file1.txt
[新增] 已记录文件哈希: file1.txt

[2/150000] 处理文件: file2.txt
[删除] 重复文件: file2.txt

========== 去重任务完成 ==========
[退出] 程序正常退出
========== 程序结束 ==========
```

## 注意事项

⚠️ **重要提示**:

1. **必须先启动 Redis**: 运行前请执行 `docker-compose up -d` 启动 Redis 服务
2. **首次使用请务必使用 `-t` 试运行模式**, 确认要删除的文件无误后再正式执行
3. 程序会**永久删除**重复文件，且**无法恢复**，请确保有备份
4. 保留的是**第一个扫描到的文件**，删除的是后续发现的重复文件
5. 对于无法读取的文件 (权限不足、被占用等)，程序会跳过并记录错误
6. 日志文件会自动轮转，最大 1MB，保留最近 28 天的备份
7. Redis 数据持久化在 `./redis-data` 目录，建议定期备份

## 技术栈

- **语言**: Go 1.26.2
- **CLI 框架**: [Cobra](https://github.com/spf13/cobra) - 强大的命令行工具库
- **Redis 客户端**: [go-redis/v9](https://github.com/redis/go-redis) - Redis Go 客户端
- **文件查找**: [finder](https://github.com/zhangyiming748/finder) - 文件和文件夹查找工具
- **哈希算法**: [xxhash/v2](https://github.com/cespare/xxhash) - 高性能 XXH3 哈希
- **日志轮转**: [lumberjack](https://github.com/zhangyiming748/lumberjack) - 日志文件管理
- **进度条**: [progressbar/v3](https://github.com/schollz/progressbar) - 终端进度条

## 项目结构

```
dedup/
├── main.go              # CLI 入口, Cobra 命令定义
├── core/
│   └── dup.go           # 核心去重逻辑 (Redis + XXH3)
├── redis/               # Redis 客户端模块
│   ├── connect.go       # Redis 连接管理
│   ├── hash.go          # Hash 操作
│   ├── string.go        # String 操作
│   ├── set.go           # Set 操作
│   ├── sort.go          # List 操作
│   ├── zset.go          # ZSet 操作
│   ├── general.go       # 通用操作
│   ├── file_dedup.go    # 文件去重专用函数
│   ├── redis_test.go    # 单元测试
│   └── README.md        # Redis 模块文档
├── util/
│   └── log.go           # 日志配置和管理
├── examples/            # 示例代码
│   └── redis_dedup_example.go
├── docker-compose.yml   # Redis Docker 配置
├── redis.conf           # Redis 配置文件
├── go.mod               # Go 模块依赖
├── go.sum               # 依赖校验文件
└── README.md            # 项目文档
```

## 开发

### 启动 Redis

```bash
# 启动 Redis 服务
docker-compose up -d

# 验证配置
docker exec -it dedup-redis redis-cli CONFIG GET appendonly
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行 Redis 模块测试
go test ./redis -v
```

### 代码格式化

```bash
go fmt ./...
```

### 相关文档

- **Redis 模块文档**: [redis/README.md](redis/README.md)
- **快速开始指南**: [redis/QUICKSTART.md](redis/QUICKSTART.md)
- **配置说明**: [REDIS_CONFIG_GUIDE.md](REDIS_CONFIG_GUIDE.md)
- **使用示例**: [examples/redis_dedup_example.go](examples/redis_dedup_example.go)

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 贡献

欢迎提交 Issue 和 Pull Request!

## 作者

zhangyiming748

---

**提示**: 如有任何问题或建议, 欢迎通过 GitHub Issues 反馈.
