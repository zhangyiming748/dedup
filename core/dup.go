package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"dedup/sqlite"

	"github.com/cespare/xxhash/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/zhangyiming748/finder"
)

func Duplicate(root string, real bool) {
	// 初始化 SQLite
	sqlite.SetSqlite()

	// 清空数据库（每次运行都是全新的）
	sqlite.ClearAll()

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

		// 计算文件哈希值（使用 XXH3）
		hash, err := calculateXXH3(fp)
		if err != nil {
			log.Printf("[错误] 计算哈希失败: %s - %v", fp, err)
			bar.Add(1)
			continue
		}

		// 获取文件大小
		fileInfo, err := os.Stat(fp)
		if err != nil {
			log.Printf("[错误] 获取文件信息失败: %s - %v", fp, err)
			bar.Add(1)
			continue
		}
		fileSize := fileInfo.Size()

		if real {
			// 真实模式：检查是否重复，如果重复则删除
			isDuplicate, originalPath, err := sqlite.CheckAndAdd(hash, fp, fileSize)
			if err != nil {
				log.Printf("[错误] 数据库操作失败: %s - %v", fp, err)
				bar.Add(1)
				continue
			}

			if isDuplicate {
				// 发现重复文件，删除当前文件
				err := os.Remove(fp)
				if err != nil {
					log.Printf("[错误] 删除文件失败: %s - %v", fp, err)
				} else {
					log.Printf("[删除] 重复文件: %s (原件: %s)", fp, originalPath)
				}
			} else {
				log.Printf("[新增] 已记录: %s (hash: %s)", fp, hash)
			}
		} else {
			// 试运行模式：只记录，不检查重复，不删除
			err = sqlite.AddFile(hash, fp, fileSize)
			if err != nil {
				log.Printf("[错误] 写入数据库失败: %s - %v", fp, err)
			} else {
				log.Printf("[新增] 已记录: %s (hash: %s)", fp, hash)
			}
		}

		bar.Add(1)
	}
	bar.Finish()

	// 显示统计信息
	total, err := sqlite.GetTotalFiles()
	if err == nil {
		log.Printf("✓ 数据库中共有 %d 个文件记录", total)
	}

	log.Printf("========== 去重任务完成 ==========")
}

// calculateXXH3 使用 XXH3 算法计算文件哈希 (比 MD5 快 3-5 倍)
func calculateXXH3(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := xxhash.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// XXH3 返回 uint64，转换为字符串
	hashValue := hash.Sum64()
	return strconv.FormatUint(hashValue, 10), nil
}
