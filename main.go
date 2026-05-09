package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"dedup/core"
	"dedup/sqlite"
	"dedup/util"

	"github.com/spf13/cobra"
)

var (
	rootDir string
)

// 版本信息（通过 -ldflags 在编译时注入）
var (
	Version   = "dev"     // 版本号，如 v1.0.0
	BuildTime = "unknown" // 编译时间
	GitCommit = "unknown" // Git commit hash
)

var rootCmd = &cobra.Command{
	Use:   "dedup",
	Short: "查找并删除重复文件",
	Long: `dedup 是一个用于查找和删除重复文件的命令行工具.
它会扫描指定目录下的所有文件, 通过计算 XXH3 128-bit 哈希值来识别重复文件,
并自动删除重复文件（保留一份）。`,
	Version: Version,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("[命令执行] 开始解析命令参数")
		log.Printf("  原始参数: %v", os.Args)
		log.Printf("  rootDir 参数值: '%s'", rootDir)

		// 如果没有提供 -d 参数, 显示帮助信息
		if rootDir == "" {
			log.Printf("[警告] 未提供根目录参数 (-d), 显示帮助信息")
			return cmd.Help()
		}

		log.Printf("[验证] 参数验证通过")
		log.Printf("  目标目录: %s", rootDir)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("[执行] 开始执行去重任务...")
		log.Printf("========================================")
		core.Duplicate(rootDir)
		log.Printf("========================================")
		log.Printf("[完成] 去重任务执行完毕")
	},
	SilenceUsage: true,
}

func init() {
	// 设置日志
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("创建本地sqlite数据库目录失败:%s", err.Error())
	}
	location := filepath.Join(home, "dedup.log")
	util.SetLog(location)
	sqlite.SetSqlite()
	log.Printf("[初始化] SQLite 数据库已启动")
	log.Printf("[初始化] 日志系统已启动, 日志文件: dedup.log")

	// 添加命令行标志
	rootCmd.Flags().StringVarP(&rootDir, "dir", "d", "", "要扫描的根目录路径")
	rootCmd.Version = Version // 设置版本信息
	log.Printf("[初始化] 命令行参数注册完成")
	log.Printf("  - dir (短参数: -d): 要扫描的根目录路径, 默认: (空)")

	// 添加详细版本命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示详细版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("dedup version %s\n", Version)
			fmt.Printf("Build time: %s\n", BuildTime)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
	rootCmd.AddCommand(versionCmd)
}

func main() {
	log.Printf("========== 程序启动 ==========")
	log.Printf("[启动] dedup 文件去重工具")
	log.Printf("[启动] 命令行: %v", os.Args)
	log.Printf("[启动] 工作目录: %s", func() string {
		dir, err := os.Getwd()
		if err != nil {
			return "未知"
		}
		return dir
	}())

	log.Printf("[启动] 开始执行 Cobra 命令...")
	if err := rootCmd.Execute(); err != nil {
		log.Printf("[错误] 命令执行失败: %v", err)
		fmt.Println(err)
		log.Printf("[退出] 程序异常退出, 退出码: 1")
		os.Exit(1)
	}

	log.Printf("[退出] 程序正常退出")
	log.Printf("========== 程序结束 ==========")
}
