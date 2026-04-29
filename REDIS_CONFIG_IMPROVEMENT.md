# Redis 配置优化说明

## 📋 改进内容

### 之前的问题
使用 Docker Compose 的 `command` 参数直接传递配置项：
```yaml
command: >
  redis-server
  --appendonly yes
  --appendfsync everysec
  --save 900 1
  ...
```

**缺点：**
- ❌ 配置项分散，不易维护
- ❌ 无法使用注释说明配置含义
- ❌ 配置项过多时命令行过长
- ❌ 不符合生产环境最佳实践
- ❌ 难以进行复杂的配置（如安全加固）

### 现在的方案
使用独立的配置文件并挂载到容器：
```yaml
volumes:
  - ./redis.conf:/usr/local/etc/redis/redis.conf:ro
command: redis-server /usr/local/etc/redis/redis.conf
```

**优点：**
- ✅ 配置集中管理，一目了然
- ✅ 支持详细注释和说明
- ✅ 符合 Redis 官方推荐方式
- ✅ 易于版本控制和分享
- ✅ 支持所有配置项（包括复杂配置）
- ✅ 可以在运行时通过 CONFIG SET 动态调整

## 📁 相关文件

### 1. redis.conf
**位置**: 项目根目录  
**作用**: Redis 完整配置文件  
**特点**:
- 212 行完整配置
- 包含详细中文注释
- 针对 dedup 项目优化
- 涵盖网络、持久化、内存、性能等所有方面

### 2. docker-compose.yml
**修改内容**:
```yaml
volumes:
  # 挂载 Redis 配置文件（只读）
  - ./redis.conf:/usr/local/etc/redis/redis.conf:ro
  # 挂载数据目录（持久化）
  - ./redis-data:/data
command: redis-server /usr/local/etc/redis/redis.conf
```

### 3. REDIS_CONFIG_GUIDE.md
**位置**: 项目根目录  
**作用**: 详细的配置说明文档  
**内容**:
- 所有配置项的详细说明
- 针对不同场景的配置模板
- 调优建议和最佳实践
- 验证和调试方法

### 4. verify_redis_config.ps1 / .sh
**位置**: 项目根目录  
**作用**: 配置验证脚本  
**功能**:
- 自动检查容器状态
- 验证关键配置是否加载
- 测试基本功能
- 检查持久化文件

## 🎯 配置亮点

### 1. 双重持久化
```conf
# RDB 快照
save 900 1
save 300 10
save 60 10000

# AOF 日志
appendonly yes
appendfsync everysec
```
**优势**: 既保证性能，又确保数据安全

### 2. 性能优化
```conf
hz 10
dynamic-hz yes
activerehashing yes
maxclients 10000
```
**优势**: 根据负载自动调整，支持高并发

### 3. 数据结构优化
```conf
hash-max-listpack-entries 128
set-max-intset-entries 512
zset-max-listpack-entries 128
```
**优势**: 针对 dedup 使用的 Hash/Set/ZSet 优化内存使用

### 4. 安全性（可选）
```conf
# requirepass your_password_here
# rename-command FLUSHDB ""
# rename-command FLUSHALL ""
```
**优势**: 生产环境可快速启用安全加固

## 🚀 使用方法

### 启动 Redis
```bash
docker-compose up -d
```

### 验证配置
```bash
# Windows PowerShell
.\verify_redis_config.ps1

# Linux/Mac
chmod +x verify_redis_config.sh
./verify_redis_config.sh
```

### 修改配置

#### 方法 1：编辑配置文件后重启
```bash
# 编辑配置
notepad redis.conf  # Windows
vim redis.conf      # Linux/Mac

# 重启生效
docker-compose restart
```

#### 方法 2：运行时动态修改
```bash
docker exec -it dedup-redis redis-cli

# 修改配置
CONFIG SET save "900 1 300 10"

# 保存到文件
CONFIG REWRITE
```

### 查看当前配置
```bash
docker exec -it dedup-redis redis-cli

# 查看所有配置
CONFIG GET *

# 查看特定配置
CONFIG GET appendonly
CONFIG GET save
```

## 📊 配置对比

| 配置项 | 之前 | 现在 |
|--------|------|------|
| **配置方式** | 命令行参数 | 独立配置文件 |
| **可维护性** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **可读性** | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **灵活性** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **专业性** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **注释支持** | ❌ | ✅ |
| **动态修改** | ❌ | ✅ |
| **版本控制** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

## 🔍 验证示例

运行验证脚本后的输出示例：
```
=========================================
  Redis 配置验证脚本
=========================================

1. 检查 Redis 容器状态...
   ✓ Redis 容器正在运行

2. 验证配置是否加载...

1) "appendonly"
2) "yes"

1) "appendfsync"
2) "everysec"

1) "save"
2) "900 1 300 10 60 10000"

1) "bind"
2) "0.0.0.0"

1) "port"
2) "6379"

3. 功能测试...
   ✓ 写入/读取测试通过

=========================================
  验证完成！
=========================================
```

## 💡 最佳实践

### 1. 开发环境
保持当前配置即可，已经过优化。

### 2. 生产环境建议
```conf
# 启用密码认证
requirepass StrongP@ssw0rd!2024

# 限制内存
maxmemory 4gb
maxmemory-policy allkeys-lru

# 禁用危险命令
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command DEBUG ""

# 启用 TLS（需要额外配置证书）
# tls-port 6380
# tls-cert-file /path/to/cert.pem
# tls-key-file /path/to/key.pem
```

### 3. 监控建议
```bash
# 定期检查内存使用
docker exec dedup-redis redis-cli INFO memory

# 检查持久化状态
docker exec dedup-redis redis-cli INFO persistence

# 查看慢查询
docker exec dedup-redis redis-cli SLOWLOG GET 10

# 监控连接数
docker exec dedup-redis redis-cli INFO clients
```

## ⚠️ 注意事项

1. **配置文件权限**: `:ro` 表示只读挂载，容器内无法修改
2. **修改后重启**: 修改 `redis.conf` 后需要重启容器才能生效
3. **数据备份**: 定期备份 `redis-data` 目录
4. **日志轮转**: Docker 会自动处理日志，无需额外配置
5. **端口冲突**: 确保 6379 端口未被占用

## 📝 常见问题

### Q1: 如何确认配置已加载？
```bash
docker exec -it dedup-redis redis-cli CONFIG GET <配置项>
```

### Q2: 配置修改后不生效？
```bash
# 检查配置文件语法
docker exec dedup-redis redis-server --test-memory 1

# 重启容器
docker-compose restart

# 查看日志
docker-compose logs redis
```

### Q3: 如何在运行时临时修改配置？
```bash
docker exec -it dedup-redis redis-cli CONFIG SET <key> <value>
```

### Q4: 持久化文件在哪里？
```bash
ls -lh redis-data/
# dump.rdb - RDB 快照
# appendonly.aof - AOF 日志
```

## 🎉 总结

通过这次优化：
- ✅ 配置更加优雅和专业
- ✅ 易于维护和扩展
- ✅ 符合生产环境标准
- ✅ 提供了完整的文档和工具
- ✅ 为后续优化打下基础

您的 Redis 配置现在已经达到生产级别的标准！
