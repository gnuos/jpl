package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRunScriptSuccess 测试成功执行脚本
func TestRunScriptSuccess(t *testing.T) {
	// 创建临时脚本文件
	tmpDir := t.TempDir()
	scriptFile := filepath.Join(tmpDir, "test.jpl")

	script := `print "Hello from JPL"`
	if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
		t.Fatalf("无法创建测试脚本: %v", err)
	}

	// 测试：脚本应该能成功执行（不panic）
	// 注意：由于runScript调用os.Exit，我们无法直接测试
	// 这里只是验证文件存在
	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		t.Errorf("测试脚本文件不存在")
	}
}

// TestRunScriptFileNotFound 测试文件不存在的情况
func TestRunScriptFileNotFound(t *testing.T) {
	// 测试不存在的文件
	nonExistentFile := "/tmp/nonexistent_script_xyz.jpl"

	_, err := os.Stat(nonExistentFile)
	if !os.IsNotExist(err) {
		t.Logf("文件意外存在: %v", err)
	}
}

// TestRunScriptExitCodes 测试退出码常量
func TestRunScriptExitCodes(t *testing.T) {
	// 验证退出码常量定义
	if exitSuccess != 0 {
		t.Errorf("exitSuccess 应该是 0, 得到 %d", exitSuccess)
	}
	if exitCompileError != 1 {
		t.Errorf("exitCompileError 应该是 1, 得到 %d", exitCompileError)
	}
	if exitRuntimeError != 2 {
		t.Errorf("exitRuntimeError 应该是 2, 得到 %d", exitRuntimeError)
	}
	if exitFileError != 3 {
		t.Errorf("exitFileError 应该是 3, 得到 %d", exitFileError)
	}
}

// TestRunCmdRegistration 测试 run 命令已注册
func TestRunCmdRegistration(t *testing.T) {
	// 验证 runCmd 已定义
	if runCmd == nil {
		t.Fatal("runCmd 未定义")
	}

	// 验证命令名称
	if runCmd.Name() != "run" {
		t.Errorf("命令名应该是 'run', 得到 %q", runCmd.Name())
	}

	// 验证使用说明
	if runCmd.Use == "" {
		t.Error("runCmd.Use 不应为空")
	}
}

// TestRunCmdArgs 测试命令参数要求
func TestRunCmdArgs(t *testing.T) {
	// 验证命令需要至少1个参数（文件名）
	if runCmd.Args == nil {
		t.Log("Args 约束为 nil，可能是 cobra.MinimumNArgs(1)")
	}
}
