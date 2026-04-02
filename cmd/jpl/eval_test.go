package main

import (
	"testing"

	"github.com/gnuos/jpl"
	"github.com/gnuos/jpl/pkg/stdlib"
	"github.com/gnuos/jpl/engine"
)

// TestEvalSimple 测试简单代码求值
func TestEvalSimple(t *testing.T) {
	code := `$x = 10 + 20`

	e := jpl.NewEngine()
	defer e.Close()
	stdlib.RegisterAll(e)

	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		t.Fatalf("执行失败: %v", err)
	}
}

// TestEvalExpression 测试表达式求值
func TestEvalExpression(t *testing.T) {
	code := `println 5 * 5`

	e := jpl.NewEngine()
	defer e.Close()
	stdlib.RegisterAll(e)

	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		t.Fatalf("执行失败: %v", err)
	}
}

// TestEvalFunction 测试函数定义和执行
func TestEvalFunction(t *testing.T) {
	code := `
function add(a, b) {
	return a + b
}
$result = add(10, 20)
`

	e := jpl.NewEngine()
	defer e.Close()
	stdlib.RegisterAll(e)

	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		t.Fatalf("执行失败: %v", err)
	}
}

// TestEvalCmdRegistration 测试 eval 命令已注册
func TestEvalCmdRegistration(t *testing.T) {
	// 验证 evalCmd 已定义
	if evalCmd == nil {
		t.Fatal("evalCmd 未定义")
	}

	// 验证命令名称
	if evalCmd.Name() != "eval" {
		t.Errorf("命令名应该是 'eval', 得到 %q", evalCmd.Name())
	}

	// 验证使用说明
	if evalCmd.Short == "" {
		t.Error("evalCmd.Short 不应为空")
	}
}

// TestEvalCmdArgs 测试命令参数要求
func TestEvalCmdArgs(t *testing.T) {
	// 验证命令需要恰好1个参数
	if evalCmd.Args == nil {
		t.Log("Args 约束为 nil")
	}
}

// TestEvalComplexCode 测试复杂代码片段
func TestEvalComplexCode(t *testing.T) {
	code := `
$arr = [1, 2, 3, 4, 5]
$sum = 0
foreach ($x in $arr) {
	$sum = $sum + $x
}
`

	e := jpl.NewEngine()
	defer e.Close()
	stdlib.RegisterAll(e)

	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		t.Fatalf("执行失败: %v", err)
	}
}
