# Redis 配置验证脚本 (PowerShell)
# 用于验证 Docker Compose 启动的 Redis 是否正确加载了配置文件

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  Redis 配置验证脚本" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 检查容器是否运行
Write-Host "1. 检查 Redis 容器状态..." -ForegroundColor Yellow
$container = docker ps --filter "name=dedup-redis" --format "{{.Names}}"
if ($container -eq "dedup-redis") {
    Write-Host "   ✓ Redis 容器正在运行" -ForegroundColor Green
} else {
    Write-Host "   ✗ Redis 容器未运行" -ForegroundColor Red
    Write-Host "   请先执行: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# 验证配置
Write-Host "2. 验证配置是否加载..." -ForegroundColor Yellow
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET appendonly
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET appendfsync
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET save
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET bind
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET port
Write-Host ""

docker exec dedup-redis redis-cli CONFIG GET maxclients
Write-Host ""

# 功能测试
Write-Host "3. 功能测试..." -ForegroundColor Yellow
docker exec dedup-redis redis-cli SET test_key "test_value" | Out-Null
$result = docker exec dedup-redis redis-cli GET test_key
if ($result -eq "test_value") {
    Write-Host "   ✓ 写入/读取测试通过" -ForegroundColor Green
} else {
    Write-Host "   ✗ 写入/读取测试失败" -ForegroundColor Red
}
docker exec dedup-redis redis-cli DEL test_key | Out-Null

Write-Host ""

# 检查持久化文件
Write-Host "4. 检查持久化文件..." -ForegroundColor Yellow
if (Test-Path "redis-data") {
    Write-Host "   数据目录内容:" -ForegroundColor Gray
    Get-ChildItem "redis-data" -ErrorAction SilentlyContinue | Format-Table Name, Length, LastWriteTime
} else {
    Write-Host "   ✗ 数据目录不存在" -ForegroundColor Red
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  验证完成！" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "提示：" -ForegroundColor Yellow
Write-Host "- 如果看到配置值与您设置的相符，说明配置加载成功"
Write-Host "- 持久化文件会在有数据写入后自动创建"
Write-Host "- 查看详细日志: docker-compose logs -f redis"
