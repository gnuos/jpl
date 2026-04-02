package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnuos/jpl"
	"github.com/gnuos/jpl/pkg/stdlib"
	"github.com/gnuos/jpl/engine"
	"github.com/spf13/cobra"
)

// runCmd 实现 jpl run 子命令
var runCmd = &cobra.Command{
	Use:   "run <file> [args...]",
	Short: "执行 JPL 脚本文件",
	Long: `执行指定的 JPL 脚本文件。

示例：
  jpl run script.jpl
  jpl run script.jpl arg1 arg2`,
	Args: cobra.MinimumNArgs(1),
	Run:  runScript,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// exitCode 退出码常量
const (
	exitSuccess      = 0
	exitCompileError = 1
	exitRuntimeError = 2
	exitFileError    = 3
)

func runScript(cmd *cobra.Command, args []string) {
	filename := args[0]
	scriptArgs := args[1:]

	// 读取脚本文件
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法读取文件 %q: %v\n", filename, err)
		os.Exit(exitFileError)
	}

	// 创建引擎
	e := jpl.NewEngine()
	defer e.Close()

	// 注册所有内置函数
	stdlib.RegisterAll(e)

	// 注册 __FILE__ 常量（PHP 风格）
	if err := e.RegisterConst("__FILE__", jpl.NewString(filename)); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法设置 __FILE__: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	// 注册 __DIR__ 常量（PHP 风格）
	dir := filepath.Dir(filename)
	if err := e.RegisterConst("__DIR__", jpl.NewString(dir)); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法设置 __DIR__: %v\n", err)
		os.Exit(exitRuntimeError)
	}

	// 准备命令行参数数组（ARGV/ARGC 魔术常量使用）
	// ARGV[0] = 脚本文件名
	// ARGV[1], ARGV[2], ... = 命令行参数
	argvArray := make([]engine.Value, 0, 1+len(scriptArgs))
	argvArray = append(argvArray, engine.NewString(filename))
	for _, arg := range scriptArgs {
		argvArray = append(argvArray, engine.NewString(arg))
	}

	// 编译并执行
	// 传入内置函数名作为预定义全局变量，确保编译器能识别这些函数
	builtinFuncs := stdlib.FunctionNames()
	prog, err := engine.CompileStringWithGlobals(string(content), filename, builtinFuncs)
	if err != nil {
		// 编译错误
		if ce, ok := err.(*engine.CompileError); ok {
			fmt.Fprintf(os.Stderr, "编译错误: %s:%d:%d: %s\n",
				ce.File, ce.Line, ce.Column, ce.Message)
		} else {
			fmt.Fprintf(os.Stderr, "编译错误: %v\n", err)
		}
		os.Exit(exitCompileError)
	}

	// 创建 VM 并设置命令行参数
	vm := engine.NewVMWithProgram(e, prog)
	vm.SetDebugMode(debug)
	vm.SetArgs(argvArray)
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] 执行 %s\n", filename)
	}

	if err := vm.Execute(); err != nil {
		// 运行时错误
		if re, ok := err.(*engine.RuntimeError); ok {
			fmt.Fprintf(os.Stderr, "运行时错误: %s:%d:%d: %s\n",
				re.File, re.Line, re.Column, re.Message)
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

	// 成功退出
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] 执行完成\n")
	}
}
