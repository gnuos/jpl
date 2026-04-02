package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gnuos/jpl/pkg/lint"
)

// lintCmd 实现 jpl lint 子命令
var lintCmd = &cobra.Command{
	Use:   "lint <file>",
	Short: "静态分析 JPL 脚本",
	Long: `检查 JPL 脚本的常见问题。

检测规则：
  unused-var     声明但未使用的变量（warning）
  undefined-var  使用未声明的变量（error）
  dead-code      return/break/continue/throw 后的不可达代码（warning）

示例：
  jpl lint script.jpl
  jpl lint *.jpl`,
	Args: cobra.MinimumNArgs(1),
	Run:  lintScript,
}

func init() {
	rootCmd.AddCommand(lintCmd)
}

func lintScript(cmd *cobra.Command, args []string) {
	hasError := false

	for _, filename := range args {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: 无法读取文件: %v\n", filename, err)
			hasError = true
			continue
		}

		result := lint.Lint(string(content), filename)

		if len(result.Diagnostics) == 0 {
			if verbose {
				fmt.Fprintf(os.Stderr, "%s: OK\n", filename)
			}
			continue
		}

		for _, d := range result.Diagnostics {
			fmt.Printf("%s\n", d.String())
		}

		if result.HasErrors() {
			hasError = true
		}
	}

	if hasError {
		os.Exit(exitCompileError)
	}
}
