我现在有一个新的思路
golang负责记录文件名和md5
每个文件名对应一个md5
计算完每一个文件按照key value的形式上传到redis作为hash结构
使用HSET命令: HSET file_hashes <md5_hash> <file_path>
由于HSET的特性是当field已存在时返回0，表示该哈希值已经出现过
所以哈希作为hash的field，文件路径作为hash的value
找到的文件直接算好哈希值往里怼，HSET返回0就说明文件重复，直接就删除
这样复杂度就是on
方便快捷 简单粗暴

## ✅ 已完成

1. ✅ Redis 模块完整实现（70+ 函数）
2. ✅ Docker Compose 配置（AOF + RDB 持久化）
3. ✅ 独立配置文件 redis.conf
4. ✅ 核心去重逻辑集成到 core/dup.go
5. ✅ 完整的文档和使用说明

## 🚀 下一步

- [ ] 测试 Redis 去重功能
- [ ] 性能基准测试
- [ ] 添加命令行选项选择存储后端（内存 vs Redis）
- [ ] 实现批量操作的 Pipeline 优化