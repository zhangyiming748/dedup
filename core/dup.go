package core

import (
	"errors"
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
	"gorm.io/gorm"
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

	// 步骤3: 对相同大小的文件计算哈希并写入 SQLite
	fmt.Println("\n正在计算哈希并检测重复...")
	filesToDelete := processFilesWithUniqueIndex(sizeGroups, bar)

	bar.Finish()

	// 显示统计信息
	total, err := sqlite.GetTotalFiles()
	if err == nil {
		log.Printf("✓ 数据库中共有 %d 个唯一文件记录", total)
	}

	log.Printf("✓ 发现 %d 个重复文件待删除", len(filesToDelete))

	if len(filesToDelete) == 0 {
		log.Printf("========== 没有发现重复文件 ==========")
		return
	}

	// 步骤4: 保存待删除文件列表到文本文件
	listFile := "files_to_delete.txt"
	err = saveFilesToList(filesToDelete, listFile)
	if err != nil {
		log.Printf("[错误] 保存文件列表失败: %v", err)
		fmt.Printf("\n⚠️  警告：保存文件列表失败\n")
		return
	}

	fmt.Printf("\n⚠️  警告：即将永久删除以下重复文件！\n")
	fmt.Printf("共 %d 个文件将被删除\n", len(filesToDelete))
	fmt.Printf("📄 待删除文件列表已保存到: %s\n", listFile)
	fmt.Println("请查看该文件，确认无误后继续...")
	fmt.Print("是否继续删除？(yes/no): ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" && confirm != "y" {
		fmt.Println("已取消操作")
		log.Printf("[取消] 用户取消了删除操作")
		return
	}

	// 步骤5: 执行删除
	deleteFiles(filesToDelete)

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

// fileResult 文件处理结果
type fileResult struct {
	filePath string
	hash     string
	fileSize int64
	err      error
}

// processFilesWithUniqueIndex 处理文件并利用唯一索引检测重复
// 返回待删除的文件列表
func processFilesWithUniqueIndex(sizeGroups map[int64][]string, bar *progressbar.ProgressBar) []string {
	var filesToDelete []string

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

		// 尝试写入数据库，如果哈希已存在（唯一索引冲突），则标记为重复
		err := sqlite.AddFile(result.hash, result.filePath, result.fileSize)
		if err != nil {
			// 精确判断错误类型：只有唯一索引冲突才是重复文件
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				// 唯一索引冲突，说明是重复文件，加入待删除列表
				filesToDelete = append(filesToDelete, result.filePath)
				log.Printf("[重复] 发现重复文件: %s (hash: %s)", result.filePath, result.hash)
			} else {
				// 其他错误（数据库错误、IO错误等），记录但不删除
				log.Printf("[错误] 写入数据库失败: %s - %v", result.filePath, err)
			}
		} else {
			log.Printf("[新增] 已记录: %s (hash: %s)", result.filePath, result.hash)
		}
	}

	return filesToDelete
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
