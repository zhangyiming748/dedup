package core

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/zhangyiming748/finder"
)

func Duplicate(root string, dryrun bool) {
	if !dryrun {
		fmt.Println("⚠️  警告: 当前运行在正式模式下, 找到的重复文件会被直接删除!")
		log.Print("确定要继续吗? (y/n): ")

		// 使用 bufio.Reader 替代 log.Scanln，跨平台兼容性更好
		reader := bufio.NewReader(os.Stdin)
		confirm, err := reader.ReadString('\n')
		if err != nil {
			log.Println("读取输入失败, 已取消操作")
			log.Printf("[取消] 读取用户输入失败: %v", err)
			return
		}

		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" {
			log.Println("已取消操作")
			log.Printf("[取消] 用户取消了去重任务")
			return
		}
		log.Println("开始执行去重任务...")
		log.Println()
	}
	log.Println("")
	log.Printf("========== 开始去重任务 ==========")
	log.Printf("根目录: %s", root)
	log.Printf("试运行模式: %v", dryrun)

	folders := finder.FindAllFolders(root)
	log.Printf("找到 %d 个文件夹", len(folders))

	for i, folder := range folders {
		log.Printf("[%d/%d] 处理文件夹: %s", i+1, len(folders), folder)
		duplicate(folder, dryrun)
	}

	log.Printf("========== 去重任务完成 ==========")
}

// duplicate 查找并删除指定目录下的重复文件 (并发版本)
// 参数:
//   - folder: 要扫描的目录路径
//   - dryrun: 是否为试运行模式. true=只打印不删除, false=实际删除重复文件
//
// 工作流程:
//  1. 遍历目录下所有文件, 计算每个文件的 MD5 哈希值 (并发计算)
//  2. 使用 map 记录已出现的 MD5 值及其对应的文件路径 (key=MD5, value=文件路径)
//  3. 对于每个文件:
//     - 如果其 MD5 已在 map 中存在 -> 说明是重复文件, 根据 dryrun 决定是否删除
//     - 如果其 MD5 不在 map 中 -> 将该 MD5 和文件路径存入 map, 作为该内容的"原件"
//
// 注意:
//   - 保留第一个扫描到的文件, 删除后续发现的重复文件
//   - dryrun=true 时只会打印信息, 不会实际删除文件, 用于预览效果
//   - 使用多协程并发计算 MD5, 提升处理速度
func duplicate(folder string, dryrun bool) {
	log.Printf("  >> 开始扫描文件夹: %s", folder)

	// 步骤1: 获取目录下所有文件的完整路径列表
	files := finder.FindAllFiles(folder)
	log.Printf("  >> 找到 %d 个文件", len(files))

	// 步骤2: 创建 map 用于存储已见过的文件 MD5 值和对应的文件路径
	// key: MD5 哈希值 (字符串形式), value: 文件路径 (第一个出现该 MD5 的文件)
	dupMap := make(map[string]string)
	var mapMutex sync.Mutex // 用于保护 dupMap 的并发访问
	log.Printf("  >> 初始化 MD5 映射表")

	// 统计变量
	processedCount := 0
	duplicateCount := 0
	deletedCount := 0
	errorCount := 0

	// 步骤3: 并发计算所有文件的 MD5
	log.Printf("  >> 启动并发 MD5 计算 (使用 %d 个协程)", runtime.NumCPU())

	type md5Result struct {
		filePath string
		md5Hash  string
		err      error
	}

	results := make(chan md5Result, len(files))
	var wg sync.WaitGroup

	// 限制并发数量，避免过多协程导致系统负载过高
	// 并发数 = CPU核心数 * 2，适合 IO 密集型任务
	maxConcurrency := runtime.NumCPU() * 2
	semaphore := make(chan struct{}, maxConcurrency)

	log.Printf("  >> 启动并发 MD5 计算 (CPU核心数: %d, 最大并发数: %d)", runtime.NumCPU(), maxConcurrency)

	for _, fp := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			md5Hash, err := calculateMD5(filePath)

			// 发送结果到 channel
			results <- md5Result{
				filePath: filePath,
				md5Hash:  md5Hash,
				err:      err,
			}
		}(fp)
	}

	// 关闭 results channel，当所有协程完成后
	go func() {
		wg.Wait()
		close(results)
	}()

	// 步骤4: 处理 MD5 计算结果，检测并删除重复文件
	fileIndex := 0
	for result := range results {
		fileIndex++
		processedCount++

		if result.err != nil {
			// 如果计算失败 (如文件被占用、权限不足等), 记录错误并跳过该文件
			errorCount++
			log.Printf("    [%d/%d] [错误] 计算 MD5 失败: %s - %v", fileIndex, len(files), result.filePath, result.err)
			continue
		}

		log.Printf("    [%d/%d] [MD5] %s - %s", fileIndex, len(files), result.md5Hash, result.filePath)

		// 检查该 MD5 是否已经在 map 中出现过
		mapMutex.Lock()
		if v, ok := dupMap[result.md5Hash]; ok {
			// 情况A: MD5 已存在 -> 当前文件是重复文件
			// v 是第一个出现的文件 (原件), result.filePath 是当前扫描到的重复文件
			duplicateCount++
			log.Printf("    [重复] 发现重复文件!")
			log.Printf("      原件: %s", v)
			log.Printf("      副本: %s", result.filePath)

			if !dryrun {
				// 非试运行模式: 实际删除重复文件
				log.Printf("      [操作] 删除副本文件...")
				err := os.Remove(result.filePath)
				if err != nil {
					errorCount++
					log.Printf("      [错误] 删除文件失败: %v", err)
				} else {
					deletedCount++
					log.Printf("      [成功] 文件已删除")
				}
			} else {
				log.Printf("      [试运行] 跳过删除 (dryrun 模式)")
			}
		} else {
			// 情况B: MD5 不存在 -> 这是第一次遇到该内容类型的文件
			// 将其记录到 map 中, 作为该 MD5 的"原件", 后续相同 MD5 的文件会被视为重复
			dupMap[result.md5Hash] = result.filePath
			log.Printf("    [记录] 记录为新文件 (原件)")
		}
		mapMutex.Unlock()
	}

	// 输出统计信息
	log.Printf("  << 文件夹扫描完成")
	log.Printf("     处理文件数: %d", processedCount)
	log.Printf("     发现重复数: %d", duplicateCount)
	log.Printf("     删除文件数: %d", deletedCount)
	log.Printf("     错误次数: %d", errorCount)
	log.Printf("     唯一 MD5 数: %d", len(dupMap))
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
