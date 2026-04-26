package core

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/zhangyiming748/finder"
)

func Duplicate(root string, dryrun bool) {
	if !dryrun {
		fmt.Println("⚠️  警告: 当前运行在正式模式下, 找到的重复文件会被直接删除!")
		fmt.Println("确定要继续吗? (y/n)")
		var confirm string
		fmt.Scanln(&confirm)
		confirm = strings.ToLower(confirm)
		if confirm != "y" {
			fmt.Println("已取消操作")
			log.Printf("[取消] 用户取消了去重任务")
			return
		}
		fmt.Println("开始执行去重任务...")
		fmt.Println()
	}
	fmt.Println("")
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

// duplicate 查找并删除指定目录下的重复文件
// 参数:
//   - folder: 要扫描的目录路径
//   - dryrun: 是否为试运行模式. true=只打印不删除, false=实际删除重复文件
//
// 工作流程:
//  1. 遍历目录下所有文件, 计算每个文件的 MD5 哈希值
//  2. 使用 map 记录已出现的 MD5 值及其对应的文件路径 (key=MD5, value=文件路径)
//  3. 对于每个文件:
//     - 如果其 MD5 已在 map 中存在 -> 说明是重复文件, 根据 dryrun 决定是否删除
//     - 如果其 MD5 不在 map 中 -> 将该 MD5 和文件路径存入 map, 作为该内容的"原件"
//
// 注意:
//   - 保留第一个扫描到的文件, 删除后续发现的重复文件
//   - dryrun=true 时只会打印信息, 不会实际删除文件, 用于预览效果
func duplicate(folder string, dryrun bool) {
	log.Printf("  >> 开始扫描文件夹: %s", folder)

	// 步骤1: 获取目录下所有文件的完整路径列表
	files := finder.FindAllFiles(folder)
	log.Printf("  >> 找到 %d 个文件", len(files))

	// 步骤2: 创建 map 用于存储已见过的文件 MD5 值和对应的文件路径
	// key: MD5 哈希值（字符串形式），value: 文件路径（第一个出现该 MD5 的文件）
	dupMap := make(map[string]string)
	log.Printf("  >> 初始化 MD5 映射表")

	// 统计变量
	processedCount := 0
	duplicateCount := 0
	deletedCount := 0
	errorCount := 0

	// 步骤3: 遍历所有文件，逐个检查是否为重复文件
	for fileIdx, fp := range files {
		processedCount++
		log.Printf("    [%d/%d] 处理文件: %s", fileIdx+1, len(files), fp)

		// 步骤3.1: 计算当前文件的 MD5 哈希值 (文件的"指纹")
		// MD5 相同意味着文件内容完全相同
		md5Hash, err := calculateMD5(fp)
		if err != nil {
			// 如果计算失败 (如文件被占用、权限不足等), 记录错误并跳过该文件
			errorCount++
			log.Printf("    [错误] 计算 MD5 失败: %v", err)
			continue
		}
		log.Printf("    [MD5] %s", md5Hash)

		// 步骤3.2: 检查该 MD5 是否已经在 map 中出现过
		if v, ok := dupMap[md5Hash]; ok {
			// 情况A: MD5 已存在 -> 当前文件是重复文件
			// v 是第一个出现的文件 (原件), fp 是当前扫描到的重复文件
			duplicateCount++
			log.Printf("    [重复] 发现重复文件!")
			log.Printf("      原件: %s", v)
			log.Printf("      副本: %s", fp)

			if !dryrun {
				// 非试运行模式: 实际删除重复文件
				log.Printf("      [操作] 删除副本文件...")
				err := os.Remove(fp)
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
			dupMap[md5Hash] = fp
			log.Printf("    [记录] 记录为新文件 (原件)")
		}
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
