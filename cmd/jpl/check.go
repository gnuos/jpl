package main

import (
	"fmt"
	"os"

	"github.com/gnuos/jpl/engine"
	"github.com/spf13/cobra"
)

// checkCmd 实现 jpl check 子命令
var checkCmd = &cobra.Command{
	Use:   "check <file>",
	Short: "检查脚本语法",
	Long: `检查 JPL 脚本文件的语法是否正确，仅编译不执行。

示例：
  jpl check script.jpl
  jpl check *.jpl`,
	Args: cobra.MinimumNArgs(1),
	Run:  checkScript,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func checkScript(cmd *cobra.Command, args []string) {
	hasError := false

	for _, filename := range args {
		// 读取文件
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: 无法读取文件: %v\n", filename, err)
			hasError = true
			continue
		}

		// 仅编译，不执行
		_, err = engine.CompileStringWithName(string(content), filename)
		if err != nil {
			// 编译错误
			if ce, ok := err.(*engine.CompileError); ok {
				fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n",
					ce.File, ce.Line, ce.Column, ce.Message)
			} else {
				fmt.Fprintf(os.Stderr, "%s: %v\n", filename, err)
			}
			hasError = true
			continue
		}

		// 成功时如果有 verbose 输出
		if verbose {
			fmt.Fprintf(os.Stderr, "%s: OK\n", filename)
		}
	}

	if hasError {
		os.Exit(exitCompileError)
	}
}
