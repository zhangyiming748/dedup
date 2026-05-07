package redis

// SetAdd 向集合添加成员
func SetAdd(key string, members ...interface{}) error {
	return Rdb.SAdd(Ctx, key, members...).Err()
}

// SetMembers 获取集合所有成员
func SetMembers(key string) ([]string, error) {
	return Rdb.SMembers(Ctx, key).Result()
}

// SetIsMember 检查成员是否在集合中
func SetIsMember(key string, member interface{}) (bool, error) {
	return Rdb.SIsMember(Ctx, key, member).Result()
}

// SetCard 获取集合成员数量
func SetCard(key string) (int64, error) {
	return Rdb.SCard(Ctx, key).Result()
}

// SetRem 从集合移除成员
func SetRem(key string, members ...interface{}) error {
	return Rdb.SRem(Ctx, key, members...).Err()
}

// SetPop 随机移除并返回一个成员
func SetPop(key string) (string, error) {
	return Rdb.SPop(Ctx, key).Result()
}

// SetRandMember 随机获取一个成员（不移除）
func SetRandMember(key string) (string, error) {
	return Rdb.SRandMember(Ctx, key).Result()
}

// SetMove 将成员从一个集合移动到另一个集合
func SetMove(source, destination string, member interface{}) (bool, error) {
	return Rdb.SMove(Ctx, source, destination, member).Result()
}

// SetDiff 计算集合差集
func SetDiff(keys ...string) ([]string, error) {
	return Rdb.SDiff(Ctx, keys...).Result()
}

// SetInter 计算集合交集
func SetInter(keys ...string) ([]string, error) {
	return Rdb.SInter(Ctx, keys...).Result()
}

// SetUnion 计算集合并集
func SetUnion(keys ...string) ([]string, error) {
	return Rdb.SUnion(Ctx, keys...).Result()
}

// SetDiffStore 计算集合差集并存储到新集合
func SetDiffStore(destination string, keys ...string) (int64, error) {
	return Rdb.SDiffStore(Ctx, destination, keys...).Result()
}

// SetInterStore 计算集合交集并存储到新集合
func SetInterStore(destination string, keys ...string) (int64, error) {
	return Rdb.SInterStore(Ctx, destination, keys...).Result()
}

// SetUnionStore 计算集合并集并存储到新集合
func SetUnionStore(destination string, keys ...string) (int64, error) {
	return Rdb.SUnionStore(Ctx, destination, keys...).Result()
}
