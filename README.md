# dedup

一个基于 Go 语言开发的命令行文件去重工具, 通过计算 MD5 哈希值识别并删除重复文件.

## 功能特性

- 🔍 **智能去重**: 使用 MD5 哈希算法精确识别内容完全相同的文件
- 🛡️ **安全模式**: 支持试运行模式 (dry-run), 先预览再执行删除操作
- 📊 **详细日志**: 完整的操作日志记录, 支持文件和控制台双输出
- 🎯 **递归扫描**: 自动遍历指定目录下的所有子文件夹
- ⚡ **高效处理**: 保留首个发现的原件, 删除后续重复文件
- 📝 **统计报告**: 每个文件夹处理完成后输出详细统计数据

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

## 使用方法

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

1. **目录扫描**: 递归查找指定根目录下的所有文件夹
2. **文件遍历**: 对每个文件夹中的所有文件进行处理
3. **MD5 计算**: 为每个文件计算 MD5 哈希值作为唯一标识
4. **重复检测**: 使用哈希表记录已出现的 MD5 值
   - 如果 MD5 已存在 -> 标记为重复文件
   - 如果 MD5 不存在 -> 记录为原件
5. **文件处理**: 根据运行模式处理重复文件
   - 试运行模式: 仅打印信息, 不删除
   - 正式模式: 删除重复文件, 保留原件

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

[命令执行] 开始解析命令参数
  rootDir 参数值: '.'
  dryRun 参数值: true
[验证] 参数验证通过
  目标目录: .
  运行模式: 试运行 (不删除)

========== 开始去重任务 ==========
根目录: .
试运行模式: true
找到 62 个文件夹

[1/62] 处理文件夹: C:\Users\example\docs
  >> 开始扫描文件夹: C:\Users\example\docs
  >> 找到 91 个文件
  >> 初始化 MD5 映射表
    [1/91] 处理文件: file1.txt
      [计算] 打开文件: file1.txt
      [信息] 文件大小: 1024 bytes
      [计算] 开始计算 MD5...
      [完成] MD5 计算完成: abc123...
    [记录] 记录为新文件 (原件)
    
    [2/91] 处理文件: file2.txt
      [MD5] abc123...
    [重复] 发现重复文件!
      原件: file1.txt
      副本: file2.txt
      [试运行] 跳过删除 (dryrun 模式)
  
  << 文件夹扫描完成
     处理文件数: 91
     发现重复数: 15
     删除文件数: 0
     错误次数: 0
     唯一 MD5 数: 76

========== 去重任务完成 ==========
[退出] 程序正常退出
========== 程序结束 ==========
```

## 注意事项

⚠️ **重要提示**:

1. **首次使用请务必使用 `-t` 试运行模式**, 确认要删除的文件无误后再正式执行
2. 程序会**永久删除**重复文件, 且**无法恢复**, 请确保有备份
3. 保留的是**第一个扫描到的文件**, 删除的是后续发现的重复文件
4. 对于无法读取的文件 (权限不足、被占用等), 程序会跳过并记录错误
5. 日志文件会自动轮转, 最大 1MB, 保留最近 28 天的备份

## 技术栈

- **语言**: Go 1.26.2
- **CLI 框架**: [Cobra](https://github.com/spf13/cobra) - 强大的命令行工具库
- **文件查找**: [finder](https://github.com/zhangyiming748/finder) - 文件和文件夹查找工具
- **日志轮转**: [lumberjack](https://github.com/zhangyiming748/lumberjack) - 日志文件管理

## 项目结构

```
dedup/
├── main.go              # CLI 入口, Cobra 命令定义
├── core/
│   ├── dup.go           # 核心去重逻辑
│   └── dup_test.go      # 单元测试
├── util/
│   └── log.go           # 日志配置和管理
├── go.mod               # Go 模块依赖
├── go.sum               # 依赖校验文件
└── README.md            # 项目文档
```

## 开发

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
```

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 贡献

欢迎提交 Issue 和 Pull Request!

## 作者

zhangyiming748

---

**提示**: 如有任何问题或建议, 欢迎通过 GitHub Issues 反馈.
