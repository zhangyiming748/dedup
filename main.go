package main

import (
	"fmt"
	"os"

	"dedup/core"
	"dedup/util"

	"github.com/spf13/cobra"
)

var (
	rootDir string
	dryRun  bool
)

var rootCmd = &cobra.Command{
	Use:   "dedup",
	Short: "查找并删除重复文件",
	Long: `dedup 是一个用于查找和删除重复文件的命令行工具。
它会扫描指定目录下的所有文件，通过计算 MD5 哈希值来识别重复文件，
并可以选择性地删除重复文件。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供 -d 参数，显示帮助信息
		if rootDir == "" {
			cmd.Help()
			return
		}
		core.Duplicate(rootDir, dryRun)
	},
}

func init() {
	// 设置日志
	util.SetLog("dedup.log")
	// 添加命令行标志
	rootCmd.Flags().StringVarP(&rootDir, "dir", "d", "", "要扫描的根目录路径")
	rootCmd.Flags().BoolVarP(&dryRun, "test", "t", false, "试运行模式，只打印不删除")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}