package redis

import "fmt"

// FileHashKey 是存储文件哈希映射的 Redis Hash key
const FileHashKey = "dedup:file_hashes"

// CheckAndAddFile 检查文件是否重复并添加到 Redis
// 使用 HSETNX 实现原子性检查和插入
// 返回: isDuplicate (是否重复), originalPath (原始文件路径), error
func CheckAndAddFile(md5Hash string, filePath string) (bool, string, error) {
	// 先检查该哈希是否已存在
	exists, err := HashExists(FileHashKey, md5Hash)
	if err != nil {
		return false, "", fmt.Errorf("检查哈希失败: %v", err)
	}

	if exists {
		// 哈希已存在，获取原始文件路径
		originalPath, err := HashGet(FileHashKey, md5Hash)
		if err != nil {
			return false, "", fmt.Errorf("获取原始路径失败: %v", err)
		}
		return true, originalPath, nil
	}

	// 哈希不存在，添加新记录
	err = HashSet(FileHashKey, md5Hash, filePath)
	if err != nil {
		return false, "", fmt.Errorf("添加哈希记录失败: %v", err)
	}

	return false, "", nil
}

// GetFilePath 根据哈希值获取文件路径
func GetFilePath(md5Hash string) (string, error) {
	return HashGet(FileHashKey, md5Hash)
}

// RemoveFileHash 删除文件哈希记录
func RemoveFileHash(md5Hash string) error {
	return HashDel(FileHashKey, md5Hash)
}

// GetAllFileHashes 获取所有文件哈希映射
func GetAllFileHashes() (map[string]string, error) {
	return HashGetAll(FileHashKey)
}

// GetTotalFiles 获取已记录的文件总数
func GetTotalFiles() (int64, error) {
	return HashLen(FileHashKey)
}

// ClearAllHashes 清空所有哈希记录（慎用！）
func ClearAllHashes() error {
	return Del(FileHashKey)
}
