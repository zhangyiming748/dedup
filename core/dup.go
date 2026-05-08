package core

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"dedup/sqlite"

	"github.com/schollz/progressbar/v3"
	"github.com/zeebo/xxh3"
	"github.com/zhangyiming748/finder"
)

func Duplicate(root string) {
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

	// 步骤2: 按文件大小分组
	fmt.Println("正在按文件大小分组...")
	sizeGroups, uniqueCount := groupFilesBySize(fps)
	log.Printf("✓ 大小分组完成: %d 个唯一大小（跳过）, %d 个候选重复组", uniqueCount, len(sizeGroups))

	// 统计需要计算哈希的文件数
	totalHashFiles := 0
	for _, files := range sizeGroups {
		totalHashFiles += len(files)
	}
	log.Printf("✓ 需要计算哈希的文件: %d / %d (跳过 %d 个唯一大小文件)",
		totalHashFiles, len(fps), len(fps)-totalHashFiles)

	if totalHashFiles == 0 {
		log.Printf("========== 没有发现可能的重复文件 ==========")
		return
	}

	bar := progressbar.New(totalHashFiles)

	// 步骤3: 对相同大小的文件计算哈希并写入 SQLite（发现重复立即删除）
	fmt.Println("\n正在计算哈希并检测重复...")
	processFilesWithUniqueIndex(sizeGroups, bar)

	bar.Finish()

	// 显示统计信息
	total, err := sqlite.GetTotalFiles()
	if err == nil {
		log.Printf("✓ 数据库中共有 %d 个唯一文件记录", total)
	}

	log.Printf("========== 去重任务完成 ==========")

	// 清理：删除数据库文件，避免二次运行误删
	cleanupDatabase()
}

// calculateXXH3 使用 XXH3 128-bit 算法计算文件哈希
// 返回 32 字符的 hex 字符串，碰撞概率极低
func calculateXXH3(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 使用 XXH3 128-bit，降低碰撞概率
	hash := xxh3.New128()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// 128-bit 输出为 16 bytes，转为 hex 字符串（32 chars）
	return hex.EncodeToString(hash.Sum(nil)), nil
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

// fileResult 文件处理结果
type fileResult struct {
	filePath string
	hash     string
	fileSize int64
	err      error
}

// processFilesWithUniqueIndex 处理文件并利用唯一索引检测重复
// 发现重复文件立即删除
func processFilesWithUniqueIndex(sizeGroups map[int64][]string, bar *progressbar.ProgressBar) {
	// 以追加模式打开删除记录文件
	deletedFile, err := os.OpenFile("deleted.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[错误] 创建/打开删除记录文件失败: %v", err)
		return
	}
	defer deletedFile.Close()

	// 如果文件是新建的，写入文件头
	fileInfo, err := deletedFile.Stat()
	if err == nil && fileInfo.Size() == 0 {
		fmt.Fprintf(deletedFile, "# 已删除的重复文件列表\n")
		fmt.Fprintf(deletedFile, "# 警告：这些文件已被永久删除！\n")
		fmt.Fprintf(deletedFile, "# ========================================\n\n")
	}

	// 写入本次运行的时间戳分隔线
	fmt.Fprintf(deletedFile, "\n## 运行时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	// 用于同步写入文件
	var fileMutex sync.Mutex
	deletedCount := 0
	// 将 map 转换为切片，便于处理
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

	// 主线程收集结果并写入数据库（利用唯一索引检测重复）
	processed := 0
	for result := range results {
		processed++
		bar.Add(1)

		if result.err != nil {
			log.Printf("[错误] 处理失败: %s - %v", result.filePath, result.err)
			continue
		}

		// 尝试写入数据库，如果哈希已存在（唯一索引冲突），则立即删除
		err := sqlite.AddFile(result.hash, result.filePath, result.fileSize)
		if err != nil {
			// 判断是否为唯一索引冲突（重复文件）
			errMsg := err.Error()
			// SQLite 唯一约束错误的典型特征
			isUniqueConstraint := strings.Contains(errMsg, "UNIQUE constraint") ||
				strings.Contains(errMsg, "UNIQUE constraint failed") ||
				(strings.Contains(errMsg, "constraint failed") && strings.Contains(errMsg, "file_hashes.hash"))

			if isUniqueConstraint {
				// 唯一索引冲突，说明是重复文件，立即删除
				log.Printf("[检测到重复] %s", result.filePath)
				delErr := os.Remove(result.filePath)
				if delErr != nil {
					log.Printf("  ❌ [删除失败] %s - 错误: %v", result.filePath, delErr)
				} else {
					log.Printf("  ✅ [删除成功] %s", result.filePath)

					// 记录到 deleted.txt
					fileMutex.Lock()
					deletedCount++
					fmt.Fprintf(deletedFile, "%d. %s\n", deletedCount, result.filePath)
					fileMutex.Unlock()
				}
			} else {
				// 其他错误（数据库错误、IO错误等），记录但不删除
				log.Printf("  ⚠️  [写入失败] %s - 错误: %v (未删除)", result.filePath, err)
			}
		} else {
			log.Printf("[新增记录] %s", result.filePath)
		}
	}

	// 输出统计信息
	fmt.Println()
	if deletedCount > 0 {
		log.Printf("========== 删除统计 ==========")
		log.Printf("✅ 成功删除: %d 个文件", deletedCount)
		log.Printf("📄 删除列表已保存到: deleted.txt")
	} else {
		log.Printf("========== 检查结果 ==========")
		log.Printf("✅ 未发现重复文件")
	}
}

// deleteFiles 批量删除文件
func deleteFiles(files []string) {
	fmt.Printf("\n开始删除 %d 个重复文件...\n", len(files))
	deletedCount := 0
	failedCount := 0

	for i, fp := range files {
		err := os.Remove(fp)
		if err != nil {
			log.Printf("[错误] 删除文件失败 [%d/%d]: %s - %v", i+1, len(files), fp, err)
			failedCount++
		} else {
			deletedCount++
			if (i+1)%100 == 0 || i == len(files)-1 {
				fmt.Printf("  进度: %d/%d\n", i+1, len(files))
			}
		}
	}

	log.Printf("✓ 删除完成: 成功 %d 个, 失败 %d 个", deletedCount, failedCount)
}

// saveFilesToList 将待删除文件列表保存到文本文件
func saveFilesToList(files []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件头信息
	_, err = fmt.Fprintf(file, "# 待删除的重复文件列表\n")
	if err != nil {
		return fmt.Errorf("写入文件头失败: %w", err)
	}
	_, err = fmt.Fprintf(file, "# 总数: %d 个文件\n", len(files))
	if err != nil {
		return fmt.Errorf("写入统计信息失败: %w", err)
	}
	_, err = fmt.Fprintf(file, "# 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("写入时间戳失败: %w", err)
	}
	_, err = fmt.Fprintf(file, "# \n# 警告：这些文件将在确认后永久删除！\n")
	if err != nil {
		return fmt.Errorf("写入警告信息失败: %w", err)
	}
	_, err = fmt.Fprintf(file, "# ================================\n\n")
	if err != nil {
		return fmt.Errorf("写入分隔线失败: %w", err)
	}

	// 逐行写入文件路径
	for i, fp := range files {
		_, err = fmt.Fprintf(file, "%d. %s\n", i+1, fp)
		if err != nil {
			return fmt.Errorf("写入文件路径失败: %w", err)
		}
	}

	log.Printf("✓ 文件列表已保存到: %s", filename)
	return nil
}

// cleanupDatabase 清理数据库文件，避免二次运行误删
func cleanupDatabase() {
	dbPath := "duplicate.db"

	// 关闭数据库连接
	db := sqlite.GetSqlite()
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
			log.Printf("[清理] 数据库连接已关闭")
		}
	}

	// 删除数据库文件
	err := os.Remove(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[清理] 数据库文件不存在: %s", dbPath)
		} else {
			log.Printf("[警告] 删除数据库文件失败: %s - %v", dbPath, err)
		}
	} else {
		log.Printf("[清理] 数据库文件已删除: %s", dbPath)
	}

	// 同时删除 WAL 和 SHM 文件（SQLite 可能生成）
	for _, ext := range []string{"-wal", "-shm"} {
		walPath := dbPath + ext
		if err := os.Remove(walPath); err == nil {
			log.Printf("[清理] 已删除: %s", walPath)
		}
	}
}
