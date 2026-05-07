package core

import (
	"fmt"
	"log"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/zhangyiming748/finder"
)

func Duplicate(root string, dryrun bool) {
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

	bar := progressbar.New(len(fps))
	for i, fp := range fps {
		log.Printf("[%d/%d] 处理文件: %s", i+1, len(fps), fp)
		// TODO: 在这里添加 SQLite 去重逻辑
		// if deleteable := checkDuplicate(fp); deleteable {
		//     if !dryrun {
		//         os.Remove(fp)
		//     }
		// }
		bar.Add(1)
	}
	bar.Finish()
	log.Printf("========== 去重任务完成 ==========")
}
