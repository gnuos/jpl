package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestStdioConstants 测试 STDIN/STDOUT/STDERR 常量
func TestStdioConstants(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	RegisterPresetConstants(e)

	tests := []struct {
		name     string
		expected string
	}{
		{"STDIN", "stdin"},
		{"STDOUT", "stdout"},
		{"STDERR", "stderr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := e.GetConst(tt.name)
			if !ok {
				t.Fatalf("Constant %s should be registered", tt.name)
			}

			if val.Type() != engine.TypeStream {
				t.Errorf("%s should be a stream, got %s", tt.name, val.Type())
			}

			if val.String() != tt.expected {
				t.Errorf("%s = %q, expected %q", tt.name, val.String(), tt.expected)
			}
		})
	}
}

// TestStdioConstantsUsage 测试 IO 常量的使用场景
func TestStdioConstantsUsage(t *testing.T) {
	// 这些常量目前作为标识符使用
	// 未来可能用于指定输出目标：
	// print("message", STDERR)  // 输出到标准错误
	// input = read(STDIN)      // 从标准输入读取

	t.Log("STDIN, STDOUT, STDERR constants are available for future IO redirection support")
	t.Log("Current usage: identifiers for standard streams")
}
