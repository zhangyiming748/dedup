package util

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/zhangyiming748/lumberjack"
)

func SetLog(l string) {
	log.SetFlags(log.Ltime | log.Lshortfile)

	// 设置全局时区为Asia/Shanghai
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("无法加载时区 Asia/Shanghai: %v", err)
	} else {
		time.Local = location
	}
	// 创建一个用于写入文件的Logger实例
	fileLogger := &lumberjack.Logger{
		Filename:   l,
		MaxSize:    1, // MB
		MaxBackups: 1,
		MaxAge:     28, // days
	}
	err = fileLogger.Rotate()
	if err != nil {
		log.Println("转换新日志文件失败", err)
	}
	consoleLogger := log.New(os.Stdout, "CONSOLE: ", log.LstdFlags)
	log.SetOutput(io.MultiWriter(fileLogger, consoleLogger.Writer()))
}
