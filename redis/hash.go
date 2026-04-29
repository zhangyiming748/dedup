package redis

// HashSet 设置哈希字段的值
func HashSet(key string, field string, value interface{}) error {
	return Rdb.HSet(Ctx, key, field, value).Err()
}

// HashGet 获取哈希字段的值
func HashGet(key string, field string) (string, error) {
	return Rdb.HGet(Ctx, key, field).Result()
}

// HashDel 删除哈希字段
func HashDel(key string, fields ...string) error {
	return Rdb.HDel(Ctx, key, fields...).Err()
}

// HashExists 检查哈希字段是否存在
// 返回 true 表示存在，false 表示不存在
func HashExists(key string, field string) (bool, error) {
	return Rdb.HExists(Ctx, key, field).Result()
}

// HashGetAll 获取哈希所有字段和值
func HashGetAll(key string) (map[string]string, error) {
	return Rdb.HGetAll(Ctx, key).Result()
}

// HashKeys 获取哈希所有字段名
func HashKeys(key string) ([]string, error) {
	return Rdb.HKeys(Ctx, key).Result()
}

// HashVals 获取哈希所有值
func HashVals(key string) ([]string, error) {
	return Rdb.HVals(Ctx, key).Result()
}

// HashLen 获取哈希字段数量
func HashLen(key string) (int64, error) {
	return Rdb.HLen(Ctx, key).Result()
}

// HashMSet 批量设置哈希字段
func HashMSet(key string, values map[string]interface{}) error {
	return Rdb.HMSet(Ctx, key, values).Err()
}

// HashMGet 批量获取哈希字段的值
func HashMGet(key string, fields ...string) ([]interface{}, error) {
	return Rdb.HMGet(Ctx, key, fields...).Result()
}

// HashIncrBy 哈希字段自增
func HashIncrBy(key string, field string, increment int64) (int64, error) {
	return Rdb.HIncrBy(Ctx, key, field, increment).Result()
}

// HashIncrByFloat 哈希字段自增（浮点数）
func HashIncrByFloat(key string, field string, increment float64) (float64, error) {
	return Rdb.HIncrByFloat(Ctx, key, field, increment).Result()
}

// HashScan 遍历哈希字段（使用游标）
func HashScan(key string, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return Rdb.HScan(Ctx, key, cursor, match, count).Result()
}
