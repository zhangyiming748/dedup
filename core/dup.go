package core

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"

	"github.com/zhangyiming748/finder"
)

func Duplicate(root string, dryrun bool) {
	folders := finder.FindAllFolders(root)
	for _, folder := range folders {
		duplicate(folder, dryrun)
	}
}

// duplicate 查找并删除指定目录下的重复文件
// 参数:
//   - folder: 要扫描的目录路径
//   - dryrun: 是否为试运行模式。true=只打印不删除，false=实际删除重复文件
//
// 工作流程:
//  1. 遍历目录下所有文件，计算每个文件的 MD5 哈希值
//  2. 使用 map 记录已出现的 MD5 值及其对应的文件路径（key=MD5, value=文件路径）
//  3. 对于每个文件：
//     - 如果其 MD5 已在 map 中存在 → 说明是重复文件，根据 dryrun 决定是否删除
//     - 如果其 MD5 不在 map 中 → 将该 MD5 和文件路径存入 map，作为该内容的"原件"
//
// 注意:
//   - 保留第一个扫描到的文件，删除后续发现的重复文件
//   - dryrun=true 时只会打印信息，不会实际删除文件，用于预览效果
func duplicate(folder string, dryrun bool) {
	// 步骤1: 获取目录下所有文件的完整路径列表
	files := finder.FindAllFiles(folder)

	// 步骤2: 创建 map 用于存储已见过的文件 MD5 值和对应的文件路径
	// key: MD5 哈希值（字符串形式），value: 文件路径（第一个出现该 MD5 的文件）
	dupMap := make(map[string]string)

	// 步骤3: 遍历所有文件，逐个检查是否为重复文件
	for _, fp := range files {
		log.Println(fp)

		// 步骤3.1: 计算当前文件的 MD5 哈希值（文件的"指纹"）
		// MD5 相同意味着文件内容完全相同
		md5Hash, err := calculateMD5(fp)
		if err != nil {
			// 如果计算失败（如文件被占用、权限不足等），记录错误并跳过该文件
			log.Printf("Error calculating MD5 for %s: %v\n", fp, err)
			continue
		}
		log.Printf("MD5: %s\n", md5Hash)

		// 步骤3.2: 检查该 MD5 是否已经在 map 中出现过
		if v, ok := dupMap[md5Hash]; ok {
			// 情况A: MD5 已存在 → 当前文件是重复文件
			// v 是第一个出现的文件（原件），fp 是当前扫描到的重复文件
			log.Printf("Duplicate found: %s and %s\n", v, fp)
			if !dryrun {
				// 非试运行模式：实际删除重复文件
				log.Println("Deleting file:", fp)
				os.Remove(fp)
			}
			// 如果是 dryrun 模式，只打印信息，不执行删除操作
		} else {
			// 情况B: MD5 不存在 → 这是第一次遇到该内容类型的文件
			// 将其记录到 map 中，作为该 MD5 的"原件"，后续相同 MD5 的文件会被视为重复
			dupMap[md5Hash] = fp
		}
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
