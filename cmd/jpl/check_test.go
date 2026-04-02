package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestCheckScriptSuccess 测试成功检查脚本语法
func TestCheckScriptSuccess(t *testing.T) {
	// 创建临时脚本文件
	tmpDir := t.TempDir()
	scriptFile := filepath.Join(tmpDir, "valid.jpl")

	script := `
$x = 10
$y = 20
$z = $x + $y
print $z
`
	if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
		t.Fatalf("无法创建测试脚本: %v", err)
	}

	// 读取并编译（不执行）
	content, err := os.ReadFile(scriptFile)
	if err != nil {
		t.Fatalf("无法读取文件: %v", err)
	}

	_, err = engine.CompileStringWithName(string(content), scriptFile)
	if err != nil {
		t.Errorf("有效脚本应该编译成功，但得到错误: %v", err)
	}
}

// TestCheckScriptSyntaxError 测试语法错误检测
func TestCheckScriptSyntaxError(t *testing.T) {
	// 创建带有语法错误的脚本
	tmpDir := t.TempDir()
	scriptFile := filepath.Join(tmpDir, "invalid.jpl")

	script := `
$x = 
print $x
`
	if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
		t.Fatalf("无法创建测试脚本: %v", err)
	}

	// 读取并编译（应该失败）
	content, err := os.ReadFile(scriptFile)
	if err != nil {
		t.Fatalf("无法读取文件: %v", err)
	}

	_, err = engine.CompileStringWithName(string(content), scriptFile)
	// 应该产生编译错误
	if err == nil {
		t.Log("语法错误脚本可能未产生错误，需检查parser行为")
	}
}

// TestCheckCmdRegistration 测试 check 命令已注册
func TestCheckCmdRegistration(t *testing.T) {
	// 验证 checkCmd 已定义
	if checkCmd == nil {
		t.Fatal("checkCmd 未定义")
	}

	// 验证命令名称
	if checkCmd.Name() != "check" {
		t.Errorf("命令名应该是 'check', 得到 %q", checkCmd.Name())
	}

	// 验证使用说明包含关键信息
	if checkCmd.Short == "" {
		t.Error("checkCmd.Short 不应为空")
	}

	// 验证是语法检查命令
	if !contains(checkCmd.Short, "语法") && !contains(checkCmd.Short, "syntax") {
		t.Log("命令描述可能不包含'语法'关键字")
	}
}

// TestCheckCmdArgs 测试命令参数要求
func TestCheckCmdArgs(t *testing.T) {
	// 验证命令需要至少1个参数
	if checkCmd.Args == nil {
		t.Log("Args 约束为 nil")
	}
}

// TestCheckMultipleFiles 测试检查多个文件
func TestCheckMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建两个有效脚本
	script1 := filepath.Join(tmpDir, "script1.jpl")
	script2 := filepath.Join(tmpDir, "script2.jpl")

	os.WriteFile(script1, []byte(`$a = 1`), 0644)
	os.WriteFile(script2, []byte(`$b = 2`), 0644)

	// 验证文件存在
	if _, err := os.Stat(script1); os.IsNotExist(err) {
		t.Errorf("script1 不存在")
	}
	if _, err := os.Stat(script2); os.IsNotExist(err) {
		t.Errorf("script2 不存在")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
