package core

import (
	"os"
	"path/filepath"
	"testing"

	"dedup/redis"
)

func TestDuplicate(t *testing.T) {
	t.Log("This is a test")
	Duplicate("W:\\音乐\\audio", true)
}

// TestScanAndGroupByHash 测试 ScanAndGroupByHash 函数
func TestScanAndGroupByHash(t *testing.T) {
	// 步骤1: 初始化 Redis 连接
	t.Log("正在连接 Redis...")
	err := redis.InitRedis("localhost:6379", "", 0)
	if err != nil {
		t.Fatalf("无法连接到 Redis: %v\n请确保 Redis 已启动", err)
	}
	defer redis.Close()
	t.Log("✓ Redis 连接成功")

	// 步骤2: 创建临时测试目录和文件
	tempDir, err := os.MkdirTemp("", "dedup-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	t.Logf("测试目录: %s", tempDir)

	// 创建测试文件
	testFiles := map[string]string{
		"file1.txt":        "Hello World",
		"file2.txt":        "Hello World", // 与 file1.txt 内容相同，哈希值相同
		"file3.txt":        "Different Content",
		"subdir/file4.txt": "Hello World", // 与 file1.txt 内容相同
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(tempDir, relPath)
		// 创建子目录
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("创建子目录失败: %v", err)
		}
		// 写入文件内容
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %s - %v", fullPath, err)
		}
		t.Logf("创建测试文件: %s (内容: %s)", relPath, content)
	}

	// 步骤3: 执行 ScanAndGroupByHash
	t.Log("\n开始执行 ScanAndGroupByHash...")
	err = ScanAndGroupByHash(tempDir, true) // 使用 XXH3 算法
	if err != nil {
		t.Fatalf("ScanAndGroupByHash 执行失败: %v", err)
	}

	// 步骤4: 验证 Redis 中的数据
	t.Log("\n验证 Redis 中的数据...")

	// 获取所有键（哈希值作为 key）
	allKeys, err := redis.Keys("*")
	if err != nil {
		t.Fatalf("获取所有键失败: %v", err)
	}

	t.Logf("找到 %d 个不同的哈希值（key）", len(allKeys))
	if len(allKeys) == 0 {
		t.Fatal("Redis 中没有存储任何数据")
	}

	// 检查每个哈希值对应的文件
	for _, hashKey := range allKeys {
		t.Logf("\n--- 检查哈希值: %s ---", hashKey)
		files, err := redis.HashGetAll(hashKey)
		if err != nil {
			t.Errorf("获取哈希 %s 的文件列表失败: %v", hashKey, err)
			continue
		}

		t.Logf("该哈希值对应 %d 个文件:", len(files))
		for fileName, filePath := range files {
			t.Logf("  - %s -> %s", fileName, filePath)
		}

		// 验证：如果有多个文件，它们应该是重复文件
		if len(files) > 1 {
			t.Logf("✓ 发现重复文件组，共 %d 个文件", len(files))
		}
	}

	// 步骤5: 验证预期结果
	// 应该有 2 个不同的哈希值（"Hello World" 和 "Different Content"）
	if len(allKeys) != 2 {
		t.Errorf("预期 2 个不同的哈希值，实际得到 %d 个", len(allKeys))
	}

	// 查找包含 3 个文件的哈希值（应该是 "Hello World" 的内容）
	foundTriple := false
	for _, hashKey := range allKeys {
		files, _ := redis.HashGetAll(hashKey)
		if len(files) == 3 {
			foundTriple = true
			t.Logf("✓ 找到包含 3 个文件的哈希组（预期结果）")
			break
		}
	}

	if !foundTriple {
		t.Error("未找到包含 3 个文件的哈希组（file1.txt, file2.txt, subdir/file4.txt 应该具有相同哈希）")
	}

	t.Log("\n========== 测试完成 ==========")
}
