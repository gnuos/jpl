package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gnuos/jpl/pkg/format"
)

var (
	fmtWrite bool // 是否原地写入
	fmtCheck bool // 是否仅检查格式
)

// fmtCmd 实现 jpl fmt 子命令
var fmtCmd = &cobra.Command{
	Use:   "fmt <file>",
	Short: "格式化 JPL 脚本",
	Long: `格式化 JPL 脚本文件，规范代码风格。

默认输出格式化后的代码到标准输出。
使用 --write 原地修改文件。
使用 --check 检查文件是否已格式化（不输出，退出码表示状态）。

示例：
  jpl fmt script.jpl            # 输出到 stdout
  jpl fmt --write script.jpl    # 原地格式化
  jpl fmt --check script.jpl    # 检查格式`,
	Args: cobra.MinimumNArgs(1),
	Run:  fmtScript,
}

func init() {
	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "原地格式化文件")
	fmtCmd.Flags().BoolVarP(&fmtCheck, "check", "c", false, "检查文件是否已格式化")
	rootCmd.AddCommand(fmtCmd)
}

func fmtScript(cmd *cobra.Command, args []string) {
	hasError := false
	hasDiff := false

	for _, filename := range args {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: 无法读取文件: %v\n", filename, err)
			hasError = true
			continue
		}

		formatted, err := format.Format(string(content), filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filename, err)
			hasError = true
			continue
		}

		if fmtCheck {
			// 检查模式：比较格式化前后是否一致
			if string(content) != formatted {
				fmt.Printf("%s\n", filename)
				hasDiff = true
			}
		} else if fmtWrite {
			// 原地写入
			if string(content) != formatted {
				if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "%s: 无法写入文件: %v\n", filename, err)
					hasError = true
					continue
				}
				if verbose {
					fmt.Fprintf(os.Stderr, "已格式化: %s\n", filename)
				}
			}
		} else {
			// 输出到 stdout
			fmt.Print(formatted)
		}
	}

	if hasError {
		os.Exit(exitCompileError)
	}
	if fmtCheck && hasDiff {
		os.Exit(1)
	}
}
