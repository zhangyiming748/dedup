package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"

	"dedup/redis"
)

// 这个文件展示了如何使用 Redis 模块进行文件去重
// 运行前请确保 Redis 已通过 Docker Compose 启动: docker-compose up -d

func main() {
	// 1. 初始化 Redis 连接
	err := redis.InitRedis("localhost:6379", "", 0)
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}
	defer redis.Close()

	fmt.Println("=== Redis 文件去重示例 ===\n")

	// 2. 清空之前的测试数据（可选）
	err = redis.ClearAllHashes()
	if err != nil {
		log.Printf("清空数据失败: %v", err)
	}

	// 3. 模拟处理多个文件
	files := []string{
		"test1.txt",
		"test2.txt",
		"test3.txt",
	}

	for _, filePath := range files {
		processFile(filePath)
	}

	// 4. 显示统计信息
	total, err := redis.GetTotalFiles()
	if err != nil {
		log.Printf("获取统计信息失败: %v", err)
	} else {
		fmt.Printf("\n总共记录了 %d 个唯一文件\n", total)
	}
}

// processFile 处理单个文件
func processFile(filePath string) {
	fmt.Printf("\n处理文件: %s\n", filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 创建测试文件
		createTestFile(filePath)
	}

	// 计算文件 MD5
	md5Hash, err := calculateMD5(filePath)
	if err != nil {
		log.Printf("计算哈希失败: %v", err)
		return
	}

	fmt.Printf("  MD5: %s\n", md5Hash)

	// 使用 Redis 检查是否重复
	isDuplicate, originalPath, err := redis.CheckAndAddFile(md5Hash, filePath)
	if err != nil {
		log.Printf("检查重复失败: %v", err)
		return
	}

	if isDuplicate {
		fmt.Printf("  ⚠️  发现重复文件！\n")
		fmt.Printf("  当前文件: %s\n", filePath)
		fmt.Printf("  原始文件: %s\n", originalPath)
		fmt.Printf("  建议操作: 删除当前文件\n")
		// 实际应用中可以这里删除文件: os.Remove(filePath)
	} else {
		fmt.Printf("  ✓ 新文件，已记录到 Redis\n")
	}
}

// calculateMD5 计算文件的 MD5 哈希值
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createTestFile 创建测试文件
func createTestFile(filePath string) {
	content := fmt.Sprintf("这是测试文件: %s\n时间戳: %d", filePath, os.Getpid())
	os.WriteFile(filePath, []byte(content), 0644)
	fmt.Printf("  创建测试文件: %s\n", filePath)
}
