package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"dedup/redis"

	"github.com/schollz/progressbar/v3"
	"github.com/zhangyiming748/finder"

	"github.com/cespare/xxhash/v2"
)

func Duplicate(root string, dryrun bool) {
	// 初始化 Redis 连接
	fmt.Println("正在连接 Redis...")
	err := redis.InitRedis("localhost:6379", "", 0)
	if err != nil {
		fmt.Printf("错误: 无法连接到 Redis: %v\n", err)
		fmt.Println("请确保 Redis 已启动: docker-compose up -d")
		log.Printf("[错误] Redis 连接失败: %v", err)
		return
	}
	defer redis.Close()
	fmt.Println()

	// 步骤1: 获取所有文件（可能很耗时，使用 goroutine 显示进度提示）
	fmt.Println("正在扫描文件，请稍候...")

	var fps []string
	done := make(chan bool)

	// 启动后台 goroutine 定期打印提示
	go func() {
		seconds := 0
		for {
			select {
			case <-done:
				return
			default:
				time.Sleep(5 * time.Second)
				seconds += 5
				fmt.Printf("\r  ⏳ 扫描中... 已等待 %d 秒", seconds)
			}
		}
	}()

	// 执行文件扫描（阻塞操作）
	fps = finder.FindAllFiles(root)

	// 通知后台 goroutine 停止
	close(done)
	fmt.Println() // 换行

	log.Printf("✓ 扫描完成，找到 %d 个文件", len(fps))

	bar := progressbar.New(len(fps))
	for i, fp := range fps {
		log.Printf("[%d/%d] 处理文件: %s", i+1, len(fps), fp)
		if deleteable := redisInsertHash(fp); deleteable {
			// 删除重复的文件
			if !dryrun {
				err := os.Remove(fp)
				if err != nil {
					log.Printf("[错误] 删除文件失败: %s - %v", fp, err)
				} else {
					log.Printf("[删除] 重复文件: %s", fp)
				}
			} else {
				log.Printf("[试运行] 发现重复文件（未删除）: %s", fp)
			}
		} else {
			// 报告在hash中添加了这个文件的哈希值
			log.Printf("[新增] 已记录文件哈希: %s", fp)
		}
		bar.Add(1)
	}
	bar.Finish()
	log.Printf("========== 去重任务完成 ==========")
}

func redisInsertHash(fp string) (deleteable bool) {
	/*
		redis逻辑在这里实现
		key永远是固定的一个 就比如叫dupmission
		field是文件的哈希值 因为field不允许重复 所以一旦插入错误就说明那个文件可以删除 直接返回deleteable = true
		value是文件的路径 因为value可以重复 所以插入时不会报错
	*/

	// 步骤1: 计算文件的 XXH3 哈希值
	hash, err := calculateXXH3(fp)
	if err != nil {
		log.Printf("[错误] 计算文件哈希失败: %s - %v", fp, err)
		return false // 计算失败，不删除
	}

	// 步骤2: 将 uint64 哈希转换为字符串作为 field
	field := strconv.FormatUint(hash, 10)

	// 步骤3: 使用 HSETNX 原子性地检查并插入
	// HSETNX 只在 field 不存在时设置，返回 true 表示是新插入
	// 如果返回 false，说明 field 已存在，是重复文件
	inserted, err := redis.HSetNX("dupmission", field, fp)
	if err != nil {
		log.Printf("[错误] Redis 操作失败: %v", err)
		return false // 操作失败，不删除
	}

	// 如果 inserted 为 false，说明 field 已存在，是重复文件
	if !inserted {
		// 哈希已存在，说明是重复文件，可以删除
		return true
	}

	// 成功添加，不是重复文件
	return false
}

// calculateXXH3 使用 XXH3 算法计算文件哈希 (比 MD5 快 3-5 倍)
func calculateXXH3(filePath string) (uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hash := xxhash.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return 0, err
	}

	return hash.Sum64(), nil
}

// calculateMD5 计算文件的 MD5 哈希值
func calculateMD5(filePath string) (string, error) {
	log.Printf("      [计算] 打开文件: %s", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("      [错误] 无法打开文件: %v", err)
		return "", err
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err == nil {
		log.Printf("      [信息] 文件大小: %d bytes", fileInfo.Size())
	}

	hash := md5.New()
	log.Printf("      [计算] 开始计算 MD5...")
	_, err = io.Copy(hash, file)
	if err != nil {
		log.Printf("      [错误] MD5 计算失败: %v", err)
		return "", err
	}

	md5Result := hex.EncodeToString(hash.Sum(nil))
	log.Printf("      [完成] MD5 计算完成: %s", md5Result)
	return md5Result, nil
}

// ScanAndGroupByHash 扫描目录并将文件按哈希值分组存储到 Redis
// 数据结构: Redis Hash - key=哈希值, field=文件名, value=文件完整路径
// 示例: HSET "abc123" "file1.txt" "/path/to/file1.txt"
//
//	HSET "abc123" "file2.txt" "/path/to/file2.txt"
//
// 参数:
//   - root: 要扫描的根目录路径
//   - useXXH3: true=使用XXH3算法, false=使用MD5算法
//
// 注意: 调用此函数前需要先调用 redis.InitRedis() 初始化连接
func ScanAndGroupByHash(root string, useXXH3 bool) error {
	fmt.Println("========== 开始扫描并分组存储到 Redis ==========")
	fmt.Printf("根目录: %s\n", root)
	fmt.Printf("哈希算法: %s\n", map[bool]string{true: "XXH3", false: "MD5"}[useXXH3])

	// 步骤1: 获取所有文件
	fmt.Println("正在扫描文件，请稍候...")
	allPaths := finder.FindAllFiles(root)

	// 过滤掉目录，只保留文件
	var fps []string
	for _, path := range allPaths {
		info, err := os.Stat(path)
		if err != nil {
			log.Printf("[警告] 无法获取文件信息: %s - %v", path, err)
			continue
		}
		if !info.IsDir() {
			fps = append(fps, path)
		}
	}

	fmt.Printf("✓ 扫描完成，找到 %d 个文件\n\n", len(fps))

	if len(fps) == 0 {
		fmt.Println("没有找到任何文件")
		return nil
	}

	bar := progressbar.New(len(fps))

	// 步骤2: 遍历所有文件，计算哈希并存储到 Redis
	processedCount := 0
	errorCount := 0

	for _, fp := range fps {
		var hashStr string

		// 计算文件哈希值
		if useXXH3 {
			hash, calcErr := calculateXXH3(fp)
			if calcErr != nil {
				log.Printf("[错误] 计算文件哈希失败: %s - %v", fp, calcErr)
				errorCount++
				bar.Add(1)
				continue
			}
			hashStr = strconv.FormatUint(hash, 10)
		} else {
			hash, calcErr := calculateMD5(fp)
			if calcErr != nil {
				log.Printf("[错误] 计算文件哈希失败: %s - %v", fp, calcErr)
				errorCount++
				bar.Add(1)
				continue
			}
			hashStr = hash
		}

		// 提取文件名作为 field
		fileName := fp
		if idx := len(fp); idx > 0 {
			for i := idx - 1; i >= 0; i-- {
				if fp[i] == '/' || fp[i] == '\\' {
					fileName = fp[i+1:]
					break
				}
			}
		}

		// 存储到 Redis Hash
		// Key: 哈希值 (例如 "1234567890")
		// Field: 文件名 (例如 "photo.jpg")
		// Value: 完整路径 (例如 "/home/user/photos/photo.jpg")
		if err := redis.HashSet(hashStr, fileName, fp); err != nil {
			log.Printf("[错误] Redis 存储失败: %s - %v", fp, err)
			errorCount++
			bar.Add(1)
			continue
		}

		processedCount++
		bar.Add(1)
	}

	bar.Finish()
	fmt.Println()

	// 输出统计信息
	fmt.Printf("========== 扫描完成 ==========\n")
	fmt.Printf("总文件数: %d\n", len(fps))
	fmt.Printf("成功处理: %d\n", processedCount)
	fmt.Printf("失败数量: %d\n", errorCount)
	fmt.Printf("\n数据已存储到 Redis，可以使用以下命令查询:\n")
	fmt.Printf("  redis-cli HKEYS dupmission          # 查看所有哈希值\n")
	fmt.Printf("  redis-cli HGETALL <hash_value>      # 查看某个哈希值对应的所有文件\n")
	fmt.Printf("  redis-cli HLEN <hash_value>         # 查看某个哈希值的文件数量\n")

	log.Printf("[完成] 扫描并存储完成: 总计=%d, 成功=%d, 失败=%d", len(fps), processedCount, errorCount)

	return nil
}
