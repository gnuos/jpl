package main

import (
	"fmt"
	"os"

	"github.com/gnuos/jpl"
	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/pkg/stdlib"
	"github.com/spf13/cobra"
)

// evalCmd 实现 jpl eval 子命令
var evalCmd = &cobra.Command{
	Use:   "eval <code>",
	Short: "执行代码片段",
	Long: `直接从命令行执行 JPL 代码片段。

示例：
  jpl eval 'print "Hello, World!"'
  jpl eval 'println 1 + 2'
  jpl eval 'for (i = 0; i < 3; i = i + 1) println i'`,
	Args: cobra.ExactArgs(1),
	Run:  evalCode,
}

func init() {
	rootCmd.AddCommand(evalCmd)
}

func evalCode(cmd *cobra.Command, args []string) {
	code := args[0]

	// 创建引擎
	e := jpl.NewEngine()
	defer e.Close()

	// 注册所有内置函数
	stdlib.RegisterAll(e)

	// 编译并执行
	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		if ce, ok := err.(*engine.CompileError); ok {
			fmt.Fprintf(os.Stderr, "编译错误: %s\n", ce.Message)
		} else {
			fmt.Fprintf(os.Stderr, "编译错误: %v\n", err)
		}
		os.Exit(exitCompileError)
	}

	vm := engine.NewVMWithProgram(e, prog)
	vm.SetDebugMode(debug)
	if err := vm.Execute(); err != nil {
		if re, ok := err.(*engine.RuntimeError); ok {
			fmt.Fprintf(os.Stderr, "运行时错误: %s\n", re.Message)
		} else {
			fmt.Fprintf(os.Stderr, "运行时错误: %v\n", err)
		}
		os.Exit(exitRuntimeError)
	}

	// 检查是否是 exit/die 导致的正常终止
	exitCode := vm.GetExitCode()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
