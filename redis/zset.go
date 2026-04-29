package redis

import "github.com/redis/go-redis/v9"

// ZSetAdd 向有序集合添加成员
func ZSetAdd(key string, score float64, member interface{}) error {
	return Rdb.ZAdd(Ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// ZSetAddMany 批量向有序集合添加成员
func ZSetAddMany(key string, members ...redis.Z) error {
	return Rdb.ZAdd(Ctx, key, members...).Err()
}

// ZSetRem 从有序集合移除成员
func ZSetRem(key string, members ...interface{}) error {
	return Rdb.ZRem(Ctx, key, members...).Err()
}

// ZSetScore 获取成员的分数
func ZSetScore(key string, member string) (float64, error) {
	return Rdb.ZScore(Ctx, key, member).Result()
}

// ZSetRank 获取成员的排名（从小到大，0-based）
func ZSetRank(key string, member string) (int64, error) {
	return Rdb.ZRank(Ctx, key, member).Result()
}

// ZSetRevRank 获取成员的排名（从大到小，0-based）
func ZSetRevRank(key string, member string) (int64, error) {
	return Rdb.ZRevRank(Ctx, key, member).Result()
}

// ZSetRange 获取指定排名范围的成员（从小到大）
func ZSetRange(key string, start, stop int64) ([]string, error) {
	return Rdb.ZRange(Ctx, key, start, stop).Result()
}

// ZSetRevRange 获取指定排名范围的成员（从大到小）
func ZSetRevRange(key string, start, stop int64) ([]string, error) {
	return Rdb.ZRevRange(Ctx, key, start, stop).Result()
}

// ZSetRangeWithScores 获取指定排名范围的成员和分数（从小到大）
func ZSetRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return Rdb.ZRangeWithScores(Ctx, key, start, stop).Result()
}

// ZSetRevRangeWithScores 获取指定排名范围的成员和分数（从大到小）
func ZSetRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return Rdb.ZRevRangeWithScores(Ctx, key, start, stop).Result()
}

// ZSetRangeByScore 根据分数范围获取成员（从小到大）
func ZSetRangeByScore(key string, min, max string) ([]string, error) {
	return Rdb.ZRangeByScore(Ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

// ZSetRevRangeByScore 根据分数范围获取成员（从大到小）
func ZSetRevRangeByScore(key string, min, max string) ([]string, error) {
	return Rdb.ZRevRangeByScore(Ctx, key, &redis.ZRangeBy{Min: min, Max: max}).Result()
}

// ZSetCount 计算分数范围内的成员数量
func ZSetCount(key string, min, max string) (int64, error) {
	return Rdb.ZCount(Ctx, key, min, max).Result()
}

// ZSetCard 获取有序集合的成员数量
func ZSetCard(key string) (int64, error) {
	return Rdb.ZCard(Ctx, key).Result()
}

// ZSetRemRangeByRank 根据排名范围移除成员
func ZSetRemRangeByRank(key string, start, stop int64) (int64, error) {
	return Rdb.ZRemRangeByRank(Ctx, key, start, stop).Result()
}

// ZSetRemRangeByScore 根据分数范围移除成员
func ZSetRemRangeByScore(key string, min, max string) (int64, error) {
	return Rdb.ZRemRangeByScore(Ctx, key, min, max).Result()
}

// ZSetIncrBy 增加成员的分数
func ZSetIncrBy(key string, increment float64, member string) (float64, error) {
	return Rdb.ZIncrBy(Ctx, key, increment, member).Result()
}

// ZSetInterStore 计算多个有序集合的交集并存储
func ZSetInterStore(destination string, keys ...string) (int64, error) {
	store := &redis.ZStore{Keys: keys}
	return Rdb.ZInterStore(Ctx, destination, store).Result()
}

// ZSetUnionStore 计算多个有序集合的并集并存储
func ZSetUnionStore(destination string, keys ...string) (int64, error) {
	store := &redis.ZStore{Keys: keys}
	return Rdb.ZUnionStore(Ctx, destination, store).Result()
}
