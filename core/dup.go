package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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

	// 步骤1: 获取所有文件夹
	fps := finder.FindAllFiles(root)
	log.Printf("找到 %d 个文件夹", len(fps))

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
