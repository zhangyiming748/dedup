#!/bin/bash

# Redis 配置验证脚本
# 用于验证 Docker Compose 启动的 Redis 是否正确加载了配置文件

echo "========================================="
echo "  Redis 配置验证脚本"
echo "========================================="
echo ""

# 检查容器是否运行
echo "1. 检查 Redis 容器状态..."
if docker ps | grep -q dedup-redis; then
    echo "   ✓ Redis 容器正在运行"
else
    echo "   ✗ Redis 容器未运行"
    echo "   请先执行: docker-compose up -d"
    exit 1
fi

echo ""

# 进入 Redis CLI 执行验证
echo "2. 验证配置是否加载..."
docker exec -it dedup-redis redis-cli << 'EOF'

# 检查 AOF 配置
ECHO "=== AOF 配置 ==="
CONFIG GET appendonly
CONFIG GET appendfsync
CONFIG GET appendfilename

# 检查 RDB 配置
ECHO ""
ECHO "=== RDB 配置 ==="
CONFIG GET save
CONFIG GET dbfilename
CONFIG GET dir

# 检查网络配置
ECHO ""
ECHO "=== 网络配置 ==="
CONFIG GET bind
CONFIG GET port
CONFIG GET tcp-keepalive

# 检查内存配置
ECHO ""
ECHO "=== 内存配置 ==="
CONFIG GET maxclients
CONFIG GET hz
CONFIG GET dynamic-hz

# 检查数据结构优化
ECHO ""
ECHO "=== 数据结构优化 ==="
CONFIG GET hash-max-listpack-entries
CONFIG GET set-max-intset-entries
CONFIG GET zset-max-listpack-entries

# 测试写入和读取
ECHO ""
ECHO "=== 功能测试 ==="
SET test_key "test_value"
GET test_key
DEL test_key

ECHO ""
ECHO "=== 持久化文件检查 ==="
INFO persistence | grep rdb_bgsave_in_progress
INFO persistence | grep aof_enabled

EOF

echo ""
echo "3. 检查持久化文件..."
if [ -d "redis-data" ]; then
    echo "   数据目录内容:"
    ls -lh redis-data/ 2>/dev/null || echo "   (暂无持久化文件，这是正常的)"
else
    echo "   ✗ 数据目录不存在"
fi

echo ""
echo "========================================="
echo "  验证完成！"
echo "========================================="
echo ""
echo "提示："
echo "- 如果看到配置值与您设置的相符，说明配置加载成功"
echo "- 持久化文件会在有数据写入后自动创建"
echo "- 查看详细日志: docker-compose logs -f redis"
