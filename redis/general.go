package redis

import "time"

// Expire 设置键的过期时间
func Expire(key string, expiration time.Duration) error {
	return Rdb.Expire(Ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func TTL(key string) (time.Duration, error) {
	return Rdb.TTL(Ctx, key).Result()
}

// Del 删除键（通用，适用于所有类型）
func Del(keys ...string) error {
	return Rdb.Del(Ctx, keys...).Err()
}

// Exists 检查键是否存在（通用，适用于所有类型）
func Exists(keys ...string) (int64, error) {
	return Rdb.Exists(Ctx, keys...).Result()
}

// Keys 查找所有符合模式的键（谨慎使用，可能影响性能）
func Keys(pattern string) ([]string, error) {
	return Rdb.Keys(Ctx, pattern).Result()
}

// Type 获取键的类型
func Type(key string) (string, error) {
	return Rdb.Type(Ctx, key).Result()
}

// Rename 重命名键
func Rename(oldKey, newKey string) error {
	return Rdb.Rename(Ctx, oldKey, newKey).Err()
}

// FlushDB 清空当前数据库
func FlushDB() error {
	return Rdb.FlushDB(Ctx).Err()
}

// FlushAll 清空所有数据库
func FlushAll() error {
	return Rdb.FlushAll(Ctx).Err()
}

// DBSize 获取当前数据库的键数量
func DBSize() (int64, error) {
	return Rdb.DBSize(Ctx).Result()
}

// Ping 测试连接
func Ping() error {
	return Rdb.Ping(Ctx).Err()
}
