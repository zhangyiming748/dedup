package redis

// ListLPush 从左侧推入列表
func ListLPush(key string, values ...interface{}) error {
	return Rdb.LPush(Ctx, key, values...).Err()
}

// ListRPush 从右侧推入列表
func ListRPush(key string, values ...interface{}) error {
	return Rdb.RPush(Ctx, key, values...).Err()
}

// ListLPop 从左侧弹出元素
func ListLPop(key string) (string, error) {
	return Rdb.LPop(Ctx, key).Result()
}

// ListRPop 从右侧弹出元素
func ListRPop(key string) (string, error) {
	return Rdb.RPop(Ctx, key).Result()
}

// ListRange 获取列表指定范围的元素
func ListRange(key string, start, stop int64) ([]string, error) {
	return Rdb.LRange(Ctx, key, start, stop).Result()
}

// ListLen 获取列表长度
func ListLen(key string) (int64, error) {
	return Rdb.LLen(Ctx, key).Result()
}

// ListIndex 获取列表指定位置的元素
func ListIndex(key string, index int64) (string, error) {
	return Rdb.LIndex(Ctx, key, index).Result()
}

// ListSet 设置列表指定位置的值
func ListSet(key string, index int64, value interface{}) error {
	return Rdb.LSet(Ctx, key, index, value).Err()
}

// ListRem 移除列表中的元素
// count > 0: 从头部开始删除 count 个值为 value 的元素
// count < 0: 从尾部开始删除 |count| 个值为 value 的元素
// count = 0: 删除所有值为 value 的元素
func ListRem(key string, count int64, value interface{}) (int64, error) {
	return Rdb.LRem(Ctx, key, count, value).Result()
}

// ListLTrim 修剪列表，只保留指定范围内的元素
func ListLTrim(key string, start, stop int64) error {
	return Rdb.LTrim(Ctx, key, start, stop).Err()
}

// ListLInsert 在列表中某个元素之前或之后插入值
func ListLInsert(key string, op string, pivot, value interface{}) (int64, error) {
	var position string
	if op == "BEFORE" {
		position = "BEFORE"
	} else {
		position = "AFTER"
	}
	return Rdb.LInsert(Ctx, key, position, pivot, value).Result()
}
