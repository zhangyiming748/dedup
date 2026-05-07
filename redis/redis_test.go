package redis

import (
	"testing"
	"time"
)

// TestInitRedis 测试 Redis 连接
func TestInitRedis(t *testing.T) {
	err := InitRedis("localhost:6379", "", 0)
	if err != nil {
		t.Fatalf("无法连接到 Redis: %v", err)
	}
	defer Close()

	// 测试 Ping
	err = Ping()
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}

	t.Log("✓ Redis 连接成功")
}

// TestStringOperations 测试字符串操作
func TestStringOperations(t *testing.T) {
	err := InitRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skip("Redis 未运行，跳过测试")
	}
	defer Close()
	defer FlushDB() // 清理测试数据

	// 测试 Set/Get
	key := "test:string"
	value := "hello redis"
	err = StringSet(key, value, 5*time.Minute)
	if err != nil {
		t.Fatalf("StringSet 失败: %v", err)
	}

	got, err := StringGet(key)
	if err != nil {
		t.Fatalf("StringGet 失败: %v", err)
	}

	if got != value {
		t.Errorf("期望 %s, 得到 %s", value, got)
	}

	t.Log("✓ 字符串操作测试通过")
}

// TestHashOperations 测试哈希操作
func TestHashOperations(t *testing.T) {
	err := InitRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skip("Redis 未运行，跳过测试")
	}
	defer Close()
	defer FlushDB()

	// 测试 Hash Set/Get
	key := "test:hash"
	field := "name"
	value := "test"

	err = HashSet(key, field, value)
	if err != nil {
		t.Fatalf("HashSet 失败: %v", err)
	}

	got, err := HashGet(key, field)
	if err != nil {
		t.Fatalf("HashGet 失败: %v", err)
	}

	if got != value {
		t.Errorf("期望 %s, 得到 %s", value, got)
	}

	// 测试 HashExists
	exists, err := HashExists(key, field)
	if err != nil {
		t.Fatalf("HashExists 失败: %v", err)
	}

	if !exists {
		t.Error("字段应该存在")
	}

	t.Log("✓ 哈希操作测试通过")
}

// TestFileDedup 测试文件去重功能
func TestFileDedup(t *testing.T) {
	err := InitRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skip("Redis 未运行，跳过测试")
	}
	defer Close()
	defer ClearAllHashes()

	// 模拟第一个文件
	hash1 := "abc123def456"
	path1 := "/path/to/file1.txt"

	isDup, originalPath, err := CheckAndAddFile(hash1, path1)
	if err != nil {
		t.Fatalf("CheckAndAddFile 失败: %v", err)
	}

	if isDup {
		t.Error("第一个文件不应该被标记为重复")
	}

	// 模拟第二个相同哈希的文件（重复）
	path2 := "/path/to/file2.txt"
	isDup, originalPath, err = CheckAndAddFile(hash1, path2)
	if err != nil {
		t.Fatalf("CheckAndAddFile 失败: %v", err)
	}

	if !isDup {
		t.Error("第二个文件应该被标记为重复")
	}

	if originalPath != path1 {
		t.Errorf("期望原始路径 %s, 得到 %s", path1, originalPath)
	}

	// 测试获取文件总数
	total, err := GetTotalFiles()
	if err != nil {
		t.Fatalf("GetTotalFiles 失败: %v", err)
	}

	if total != 1 {
		t.Errorf("期望 1 个文件，得到 %d", total)
	}

	t.Log("✓ 文件去重功能测试通过")
}
