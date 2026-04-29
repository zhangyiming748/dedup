package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	// Rdb 全局 Redis 客户端实例
	Rdb *redis.Client
	// Ctx 全局上下文
	Ctx = context.Background()
)

// InitRedis 初始化 Redis 连接
func InitRedis(addr string, password string, db int) error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:         addr,            // Redis 地址，格式: "localhost:6379"
		Password:     password,        // 密码，默认为空
		DB:           db,              // 数据库编号，默认 0
		PoolSize:     100,             // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
	})

	// 测试连接
	if err := Rdb.Ping(Ctx).Err(); err != nil {
		return fmt.Errorf("无法连接到 Redis: %v", err)
	}

	fmt.Println("✓ Redis 连接成功")
	return nil
}

// Close 关闭 Redis 连接
func Close() error {
	if Rdb != nil {
		return Rdb.Close()
	}
	return nil
}
