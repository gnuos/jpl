package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 全局配置
var (
	verbose bool
	debug   bool
)

// rootCmd 是 JPL 命令行工具的根命令
var rootCmd = &cobra.Command{
	Use:   "jpl",
	Short: "JPL - 一门轻量级脚本语言",
	Long: `JPL (Jx9-like Programming Language) 是一门轻量级、高性能的脚本语言。
支持动态类型、闭包、异常处理、模块系统等现代语言特性。`,
	SilenceUsage: true,
}

func init() {
	// 全局 flag
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "调试模式")
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
