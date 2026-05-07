# Redis 配置文件说明

## 📄 文件位置

- **配置文件**: `redis.conf` - 完整的 Redis 配置
- **Docker 配置**: `docker-compose.yml` - 挂载配置文件到容器

## 🔧 配置方式

### 之前的方式（不优雅）
```yaml
command: >
  redis-server
  --appendonly yes
  --appendfsync everysec
  --save 900 1
  ...
```

### 现在的方式（优雅）
```yaml
volumes:
  - ./redis.conf:/usr/local/etc/redis/redis.conf:ro
command: redis-server /usr/local/etc/redis/redis.conf
```

## 📋 主要配置项说明

### 网络配置
```conf
bind 0.0.0.0          # 接受所有网络接口连接
port 6379             # 端口号
timeout 0             # 连接超时（0=禁用）
tcp-keepalive 300     # TCP keepalive
```

### RDB 持久化（快照）
```conf
save 900 1            # 900秒内至少1个key改变则保存
save 300 10           # 300秒内至少10个key改变则保存
save 60 10000         # 60秒内至少10000个key改变则保存
dbfilename dump.rdb   # RDB文件名
dir /data             # 数据目录
```

### AOF 持久化（追加日志）
```conf
appendonly yes                    # 开启AOF
appendfilename "appendonly.aof"   # AOF文件名
appendfsync everysec              # 每秒同步（推荐）
auto-aof-rewrite-percentage 100   # AOF文件大小增长100%时重写
auto-aof-rewrite-min-size 64mb    # AOF文件最小64MB才重写
```

### 内存管理
```conf
maxclients 10000      # 最大客户端连接数
# maxmemory <bytes>   # 最大内存限制（按需设置）
# maxmemory-policy    # 内存淘汰策略
```

### 性能优化
```conf
hz 10                 # 后台任务频率
dynamic-hz yes        # 动态调整hz
activerehashing yes   # 主动重新哈希
lazyfree-lazy-eviction no  # 惰性删除配置
```

### 安全配置（可选）
```conf
# requirepass your_password_here  # 设置密码

# 禁用危险命令
# rename-command FLUSHDB ""
# rename-command FLUSHALL ""
# rename-command DEBUG ""
```

## 🎯 针对 dedup 项目的优化

### 1. 持久化策略
- **RDB + AOF 双重持久化**：确保数据不丢失
- **AOF everysec**：平衡性能和安全性
- **自动重写**：防止 AOF 文件无限增长

### 2. 性能配置
- **连接池**：支持 10000 个客户端连接
- **惰性删除**：减少延迟峰值
- **动态 hz**：根据负载自动调整

### 3. 数据结构优化
```conf
hash-max-listpack-entries 128    # Hash 优化
set-max-intset-entries 512       # Set 优化
zset-max-listpack-entries 128    # ZSet 优化
```

## 🚀 使用方法

### 启动 Redis
```bash
docker-compose up -d
```

### 查看配置是否生效
```bash
# 进入 Redis CLI
docker exec -it dedup-redis redis-cli

# 查看 AOF 状态
CONFIG GET appendonly

# 查看保存策略
CONFIG GET save

# 查看当前配置
CONFIG GET *
```

### 修改配置

#### 方法 1：修改配置文件后重启
```bash
# 编辑 redis.conf
vim redis.conf

# 重启 Redis
docker-compose restart
```

#### 方法 2：运行时动态修改（临时）
```bash
docker exec -it dedup-redis redis-cli

# 修改配置
CONFIG SET save "900 1 300 10 60 10000"

# 保存配置到文件
CONFIG REWRITE
```

### 监控和调试

```bash
# 查看实时日志
docker-compose logs -f redis

# 查看慢查询
docker exec -it dedup-redis redis-cli SLOWLOG GET 10

# 查看内存使用
docker exec -it dedup-redis redis-cli INFO memory

# 查看持久化状态
docker exec -it dedup-redis redis-cli INFO persistence
```

## 📊 配置调优建议

### 小数据量（< 1GB）
```conf
save 900 1
save 300 10
save 60 10000
appendfsync everysec
```

### 中等数据量（1GB - 10GB）
```conf
save 300 10
save 60 1000
appendfsync everysec
maxmemory 8gb
maxmemory-policy allkeys-lru
```

### 大数据量（> 10GB）
```conf
save 300 100
appendfsync everysec
maxmemory 16gb
maxmemory-policy volatile-lru
hz 100
```

## ⚠️ 注意事项

1. **配置文件权限**：`:ro` 表示只读挂载，容器内无法修改
2. **数据目录**：`./redis-data` 会自动创建，用于存储持久化文件
3. **生产环境**：建议设置 `requirepass` 密码
4. **内存限制**：根据服务器内存设置 `maxmemory`
5. **备份策略**：定期备份 `redis-data` 目录

## 🔍 验证配置

启动后运行以下命令验证配置是否正确加载：

```bash
docker exec -it dedup-redis redis-cli

# 1. 检查 AOF 是否启用
127.0.0.1:6379> CONFIG GET appendonly
1) "appendonly"
2) "yes"

# 2. 检查保存策略
127.0.0.1:6379> CONFIG GET save
1) "save"
2) "900 1 300 10 60 10000"

# 3. 检查数据目录
127.0.0.1:6379> CONFIG GET dir
1) "dir"
2) "/data"

# 4. 测试写入
127.0.0.1:6379> SET test "hello"
OK

# 5. 检查持久化文件
exit
ls -lh redis-data/
# 应该看到 dump.rdb 和 appendonly.aof
```

## 📝 常用配置模板

### 开发环境（快速、不安全）
```conf
appendonly no
save ""
loglevel debug
```

### 生产环境（安全、稳定）
```conf
requirepass strong_password
appendonly yes
appendfsync everysec
maxmemory 4gb
maxmemory-policy allkeys-lru
rename-command FLUSHDB ""
rename-command FLUSHALL ""
```

### 缓存场景（高性能）
```conf
maxmemory 8gb
maxmemory-policy allkeys-lru
appendonly no
save ""
```

## 🎉 总结

现在您的 Redis 配置更加：
- ✅ **优雅**：使用独立的配置文件
- ✅ **灵活**：易于修改和维护
- ✅ **清晰**：所有配置一目了然
- ✅ **专业**：符合生产环境最佳实践

配置文件已针对 dedup 项目优化，可以直接使用！
