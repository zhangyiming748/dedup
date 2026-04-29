package redis

import (
	"time"
)

// StringSet 设置字符串键值对
func StringSet(key string, value interface{}, expiration time.Duration) error {
	return Rdb.Set(Ctx, key, value, expiration).Err()
}

// StringGet 获取字符串值
func StringGet(key string) (string, error) {
	return Rdb.Get(Ctx, key).Result()
}

// StringDel 删除字符串键
func StringDel(keys ...string) error {
	return Rdb.Del(Ctx, keys...).Err()
}

// StringExists 检查键是否存在
func StringExists(keys ...string) (int64, error) {
	return Rdb.Exists(Ctx, keys...).Result()
}

// StringSetNX 只在键不存在时设置值（原子操作）
// 返回 true 表示设置成功，false 表示键已存在
func StringSetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return Rdb.SetNX(Ctx, key, value, expiration).Result()
}

// StringIncr 自增
func StringIncr(key string) (int64, error) {
	return Rdb.Incr(Ctx, key).Result()
}

// StringDecr 自减
func StringDecr(key string) (int64, error) {
	return Rdb.Decr(Ctx, key).Result()
}

// StringAppend 追加字符串
func StringAppend(key string, value string) (int64, error) {
	return Rdb.Append(Ctx, key, value).Result()
}

// StringGetRange 获取子字符串
func StringGetRange(key string, start, end int64) (string, error) {
	return Rdb.GetRange(Ctx, key, start, end).Result()
}

// StringSetRange 覆盖部分字符串
func StringSetRange(key string, offset int64, value string) (int64, error) {
	return Rdb.SetRange(Ctx, key, offset, value).Result()
}

// StringStrLen 获取字符串长度
func StringStrLen(key string) (int64, error) {
	return Rdb.StrLen(Ctx, key).Result()
}

// StringMSet 批量设置键值对
func StringMSet(pairs map[string]interface{}) error {
	return Rdb.MSet(Ctx, pairs).Err()
}

// StringMGet 批量获取值
func StringMGet(keys ...string) ([]interface{}, error) {
	return Rdb.MGet(Ctx, keys...).Result()
}
