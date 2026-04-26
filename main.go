package dedup

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"

	"github.com/zhangyiming748/finder"
)

func duplicate(folder string, dryrun bool) {
	/*
		1. 遍历目录，获取所有文件的md5值
		2. 创建一个map，key为文件的md5值，value为文件的路径
		3. 遍历md5 map，如果md5已经出现过，则删除当前文件 如果没有 则创建这个key
	*/
	files := finder.FindAllFiles(folder)
	dupMap := make(map[string]string)
	for _, fp := range files {
		log.Println(fp)
		// 获得当前文件的md5
		md5Hash, err := calculateMD5(fp)
		if err != nil {
			log.Printf("Error calculating MD5 for %s: %v\n", fp, err)
			continue
		}
		log.Printf("MD5: %s\n", md5Hash)

		if v, ok := dupMap[md5Hash]; ok {
			log.Printf("Duplicate found: %s and %s\n", v, fp)
			if !dryrun {
				log.Println("Deleting file:", fp)
				os.Remove(fp)
			} else {
				dupMap[md5Hash] = fp
			}
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
