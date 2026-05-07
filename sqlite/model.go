package sqlite

import (
	"time"
)

// FileHash 文件哈希记录表
type FileHash struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`     // 主键ID
	Hash      string    `gorm:"uniqueIndex;size:64;not null"` // 文件哈希值（XXH3），唯一索引
	FilePath  string    `gorm:"size:2048;not null"`           // 文件完整路径
	FileSize  int64     `gorm:"not null"`                     // 文件大小（字节）
	CreatedAt time.Time `gorm:"autoCreateTime"`               // 创建时间
}

// TableName 指定表名
func (FileHash) TableName() string {
	return "file_hashes"
}

// CheckExists 检查哈希是否已存在
// 返回: exists (是否存在), filePath (如果存在，返回文件路径), error
func CheckExists(hash string) (bool, string, error) {
	var fileHash FileHash
	result := GetSqlite().Where("hash = ?", hash).First(&fileHash)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			// 不存在
			return false, "", nil
		}
		return false, "", result.Error
	}

	// 存在
	return true, fileHash.FilePath, nil
}

// AddFile 添加文件哈希记录
// 如果哈希已存在（唯一索引冲突），返回错误
func AddFile(hash string, filePath string, fileSize int64) error {
	fileHash := FileHash{
		Hash:     hash,
		FilePath: filePath,
		FileSize: fileSize,
	}
	return GetSqlite().Create(&fileHash).Error
}

// CheckAndAdd 检查并添加文件（原子操作）
// 返回: isDuplicate (是否重复), originalPath (原始文件路径), error
func CheckAndAdd(hash string, filePath string, fileSize int64) (bool, string, error) {
	// 先检查是否存在
	exists, originalPath, err := CheckExists(hash)
	if err != nil {
		return false, "", err
	}

	if exists {
		// 已存在，是重复文件
		return true, originalPath, nil
	}

	// 不存在，添加新记录
	err = AddFile(hash, filePath, fileSize)
	if err != nil {
		return false, "", err
	}

	return false, "", nil
}

// GetTotalFiles 获取已记录的文件总数
func GetTotalFiles() (int64, error) {
	var count int64
	return count, GetSqlite().Model(&FileHash{}).Count(&count).Error
}

// ClearAll 清空所有记录（慎用！）
func ClearAll() error {
	return GetSqlite().Exec("DELETE FROM file_hashes").Error
}
