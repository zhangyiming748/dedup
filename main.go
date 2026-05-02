package main

import (
	"fmt"
	"log"
	"os"

	"dedup/core"
	"dedup/util"

	"github.com/spf13/cobra"
)

var (
	rootDir string
	dryRun  bool
)

// rootCmd 是根命令,作为所有子命令的父命令
var rootCmd = &cobra.Command{
	Use:   "dedup",
	Short: "dedup 文件去重工具",
	Long:  `dedup 是一个用于查找和删除重复文件的命令行工具.`,
}

// dupCmd 是去重子命令
var dupCmd = &cobra.Command{
	Use:   "dup [flags]",
	Short: "查找并删除重复文件",
	Long: `dup 命令会扫描指定目录下的所有文件, 通过计算 MD5 哈希值来识别重复文件,
并可以选择性地删除重复文件.

示例:
  dedup dup -d /path/to/scan          # 正式模式, 会删除重复文件
  dedup dup -d /path/to/scan -t       # 试运行模式, 只打印不删除`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("[命令执行] 开始解析命令参数")
		log.Printf("  原始参数: %v", os.Args)
		log.Printf("  rootDir 参数值: '%s'", rootDir)
		log.Printf("  dryRun 参数值: %v", dryRun)

		// 如果没有提供 -d 参数, 显示帮助信息
		if rootDir == "" {
			log.Printf("[警告] 未提供根目录参数 (-d), 显示帮助信息")
			return cmd.Help()
		}

		log.Printf("[验证] 参数验证通过")
		log.Printf("  目标目录: %s", rootDir)
		log.Printf("  运行模式: %s", map[bool]string{true: "试运行 (不删除)", false: "正式运行 (会删除)"}[dryRun])
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("[执行] 开始执行去重任务...")
		log.Printf("========================================")
		core.Duplicate(rootDir, dryRun)
		log.Printf("========================================")
		log.Printf("[完成] 去重任务执行完毕")
	},
	SilenceUsage: true,
}

func init() {
	// 设置日志
	util.SetLog("dedup.log")
	log.Printf("[初始化] 日志系统已启动, 日志文件: dedup.log")

	// 为 dup 子命令添加命令行标志
	dupCmd.Flags().StringVarP(&rootDir, "dir", "d", "", "要扫描的根目录路径")
	dupCmd.Flags().BoolVarP(&dryRun, "test", "t", false, "试运行模式, 只打印不删除")
	log.Printf("[初始化] 命令行参数注册完成")
	log.Printf("  - dir (短参数: -d): 要扫描的根目录路径, 默认: (空)")
	log.Printf("  - test (短参数: -t): 试运行模式, 默认: false")

	// 将 dup 子命令添加到根命令
	rootCmd.AddCommand(dupCmd)
	log.Printf("[初始化] 子命令注册完成: dup")
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
