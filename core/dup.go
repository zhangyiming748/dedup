package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
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

	// 第一阶段：按文件大小分组（快速筛选）
	fmt.Println("正在按文件大小分组...")
	sizeGroups, uniqueCount := groupFilesBySize(fps)
	log.Printf("✓ 大小分组完成: %d 个唯一大小, %d 个候选重复组", uniqueCount, len(sizeGroups))

	// 统计需要计算哈希的文件数
	totalHashFiles := 0
	for _, files := range sizeGroups {
		totalHashFiles += len(files)
	}
	log.Printf("✓ 需要计算哈希的文件: %d / %d (跳过 %d 个唯一大小文件)",
		totalHashFiles, len(fps), len(fps)-totalHashFiles)

	bar := progressbar.New(totalHashFiles)

	if real {
		// 真实模式：线性处理，避免幻读和竞态条件
		processFilesSequential(sizeGroups, bar)
	} else {
		// 试运行模式：并行处理，提升性能
		processFilesParallel(sizeGroups, bar)
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

// groupFilesBySize 按文件大小分组，只返回有重复大小的文件组
// 返回: sizeGroups (大小相同的文件组), uniqueCount (唯一大小的数量)
func groupFilesBySize(fps []string) (map[int64][]string, int) {
	sizeMap := make(map[int64][]string)

	for _, fp := range fps {
		fileInfo, err := os.Stat(fp)
		if err != nil {
			log.Printf("[警告] 获取文件信息失败: %s - %v", fp, err)
			continue
		}
		size := fileInfo.Size()
		sizeMap[size] = append(sizeMap[size], fp)
	}

	// 过滤出只有唯一大小的文件（不可能重复）
	uniqueCount := 0
	result := make(map[int64][]string)
	for size, files := range sizeMap {
		if len(files) > 1 {
			// 有多个相同大小的文件，需要进一步检查
			result[size] = files
		} else {
			// 唯一大小，不可能重复，跳过
			uniqueCount++
		}
	}

	return result, uniqueCount
}

// processFilesSequential 线性处理文件（真实模式）
func processFilesSequential(sizeGroups map[int64][]string, bar *progressbar.ProgressBar) {
	// 将 map 转换为切片，便于遍历
	var allFiles []string
	for _, files := range sizeGroups {
		allFiles = append(allFiles, files...)
	}

	processed := 0
	for _, fp := range allFiles {
		processed++
		log.Printf("[%d] 处理文件: %s", processed, fp)

		// 计算文件哈希值
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

		// 检查是否重复，如果重复则删除
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

		bar.Add(1)
	}
}

// fileResult 文件处理结果
type fileResult struct {
	filePath string
	hash     string
	fileSize int64
	err      error
}

// processFilesParallel 并行处理文件（试运行模式）
func processFilesParallel(sizeGroups map[int64][]string, bar *progressbar.ProgressBar) {
	// 将 map 转换为切片，便于分发任务
	var allFiles []string
	for _, files := range sizeGroups {
		allFiles = append(allFiles, files...)
	}

	// 根据 CPU 核心数确定 worker 数量，最多 8 个
	numWorkers := runtime.NumCPU()
	if numWorkers > 8 {
		numWorkers = 8
	}

	log.Printf("[并行模式] 启动 %d 个 worker 进行并行哈希计算", numWorkers)

	// 创建任务 channel 和结果 channel
	tasks := make(chan string, numWorkers*2)
	results := make(chan fileResult, numWorkers*2)

	// 启动 worker pool
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for fp := range tasks {
				// 计算哈希
				hash, err := calculateXXH3(fp)
				if err != nil {
					results <- fileResult{filePath: fp, err: err}
					continue
				}

				// 获取文件大小
				fileInfo, err := os.Stat(fp)
				if err != nil {
					results <- fileResult{filePath: fp, err: err}
					continue
				}

				results <- fileResult{
					filePath: fp,
					hash:     hash,
					fileSize: fileInfo.Size(),
				}
			}
		}(w)
	}

	// 启动 goroutine 发送任务
	go func() {
		for _, fp := range allFiles {
			tasks <- fp
		}
		close(tasks)
	}()

	// 启动 goroutine 等待所有 worker 完成并关闭结果 channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// 主线程收集结果并写入数据库
	processed := 0
	for result := range results {
		processed++
		log.Printf("[%d] 处理文件: %s", processed, result.filePath)

		if result.err != nil {
			log.Printf("[错误] 处理失败: %s - %v", result.filePath, result.err)
			bar.Add(1)
			continue
		}

		// 检查是否重复（不删除，只提示）
		isDuplicate, originalPath, err := sqlite.CheckAndAdd(result.hash, result.filePath, result.fileSize)
		if err != nil {
			log.Printf("[错误] 数据库操作失败: %s - %v", result.filePath, err)
			bar.Add(1)
			continue
		}

		if isDuplicate {
			log.Printf("[重复] 发现重复文件: %s (原件: %s)", result.filePath, originalPath)
		} else {
			log.Printf("[新增] 已记录: %s (hash: %s)", result.filePath, result.hash)
		}

		bar.Add(1)
	}
}
