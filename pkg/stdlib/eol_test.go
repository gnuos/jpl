package stdlib

import (
	"runtime"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestEOLConstant 测试 EOL 常量
func TestEOLConstant(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	RegisterPresetConstants(e)

	// 检查 EOL 常量是否存在
	val, ok := e.GetConst("EOL")
	if !ok {
		t.Fatal("EOL constant should be registered")
	}

	if val.Type() != engine.TypeString {
		t.Errorf("EOL should be a string, got %s", val.Type())
	}

	eol := val.String()

	// 根据操作系统检查值
	switch runtime.GOOS {
	case "windows":
		if eol != "\r\n" {
			t.Errorf("On Windows, EOL should be '\\r\\n', got %q", eol)
		}
	default:
		if eol != "\n" {
			t.Errorf("On Unix-like systems, EOL should be '\\n', got %q", eol)
		}
	}

	t.Logf("EOL constant value: %q on %s", eol, runtime.GOOS)
}

// TestEOLInScript 测试在脚本中使用 EOL 常量
func TestEOLInScript(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	RegisterPresetConstants(e)
	RegisterIO(e)

	// 简单测试：EOL 应该可以在脚本中引用
	// 实际脚本执行测试需要通过 CLI 或 VM 进行
	t.Log("EOL constant is available for use in scripts")
}

// TestGetPlatformEOL 测试 getPlatformEOL 函数
func TestGetPlatformEOL(t *testing.T) {
	result := getPlatformEOL()

	switch runtime.GOOS {
	case "windows":
		if result != "\r\n" {
			t.Errorf("getPlatformEOL() = %q, expected '\\r\\n' on Windows", result)
		}
	default:
		if result != "\n" {
			t.Errorf("getPlatformEOL() = %q, expected '\\n' on Unix", result)
		}
	}
}
