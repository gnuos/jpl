package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestIntval 测试 intval 函数
func TestIntval(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeConvert(e)

	tests := []struct {
		name     string
		args     []engine.Value
		expected int64
	}{
		{"int to int", []engine.Value{engine.NewInt(42)}, 42},
		{"float to int (truncates)", []engine.Value{engine.NewFloat(3.7)}, 3},
		{"float negative to int", []engine.Value{engine.NewFloat(-3.7)}, -3},
		{"string decimal", []engine.Value{engine.NewString("123")}, 123},
		{"string hex", []engine.Value{engine.NewString("FF"), engine.NewInt(16)}, 255},
		{"string octal", []engine.Value{engine.NewString("77"), engine.NewInt(8)}, 63},
		{"string binary", []engine.Value{engine.NewString("1010"), engine.NewInt(2)}, 10},
		{"string invalid", []engine.Value{engine.NewString("abc")}, 0},
		{"empty string", []engine.Value{engine.NewString("")}, 0},
		{"true to int", []engine.Value{engine.NewBool(true)}, 1},
		{"false to int", []engine.Value{engine.NewBool(false)}, 0},
		{"null to int", []engine.Value{engine.NewNull()}, 0},
		{"array to int", []engine.Value{engine.NewArray([]engine.Value{engine.NewInt(1)})}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinIntval(ctx, tt.args)
			if err != nil {
				t.Fatalf("intval() error: %v", err)
			}
			if result.Type() != engine.TypeInt {
				t.Errorf("intval() returned %s, expected int", result.Type())
			}
			if result.Int() != tt.expected {
				t.Errorf("intval() = %d, expected %d", result.Int(), tt.expected)
			}
		})
	}
}

// TestIntvalErrors 测试 intval 错误处理
func TestIntvalErrors(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeConvert(e)

	// 测试参数数量错误
	t.Run("too few args", func(t *testing.T) {
		ctx := engine.NewContext(e, nil)
		_, err := builtinIntval(ctx, []engine.Value{})
		if err == nil {
			t.Error("expected error for 0 arguments")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		ctx := engine.NewContext(e, nil)
		_, err := builtinIntval(ctx, []engine.Value{
			engine.NewString("123"),
			engine.NewInt(10),
			engine.NewInt(16),
		})
		if err == nil {
			t.Error("expected error for 3 arguments")
		}
	})
}

// TestFloatval 测试 floatval 函数
func TestFloatval(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeConvert(e)

	tests := []struct {
		name     string
		args     []engine.Value
		expected float64
	}{
		{"float to float", []engine.Value{engine.NewFloat(3.14)}, 3.14},
		{"int to float", []engine.Value{engine.NewInt(42)}, 42.0},
		{"string float", []engine.Value{engine.NewString("3.14159")}, 3.14159},
		{"string scientific", []engine.Value{engine.NewString("1.5e3")}, 1500.0},
		{"string invalid", []engine.Value{engine.NewString("not a number")}, 0.0},
		{"empty string", []engine.Value{engine.NewString("")}, 0.0},
		{"true to float", []engine.Value{engine.NewBool(true)}, 1.0},
		{"false to float", []engine.Value{engine.NewBool(false)}, 0.0},
		{"null to float", []engine.Value{engine.NewNull()}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinFloatval(ctx, tt.args)
			if err != nil {
				t.Fatalf("floatval() error: %v", err)
			}
			if result.Type() != engine.TypeFloat {
				t.Errorf("floatval() returned %s, expected float", result.Type())
			}
			// 使用小误差范围比较浮点数
			got := result.Float()
			if got < tt.expected-0.0001 || got > tt.expected+0.0001 {
				t.Errorf("floatval() = %f, expected %f", got, tt.expected)
			}
		})
	}
}

// TestStrval 测试 strval 函数
func TestStrval(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeConvert(e)

	tests := []struct {
		name     string
		args     []engine.Value
		expected string
	}{
		{"int to string", []engine.Value{engine.NewInt(42)}, "42"},
		{"float to string", []engine.Value{engine.NewFloat(3.14)}, "3.14"},
		{"string to string", []engine.Value{engine.NewString("hello")}, "hello"},
		{"true to string", []engine.Value{engine.NewBool(true)}, "true"},
		{"false to string", []engine.Value{engine.NewBool(false)}, "false"},
		{"null to string", []engine.Value{engine.NewNull()}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinStrval(ctx, tt.args)
			if err != nil {
				t.Fatalf("strval() error: %v", err)
			}
			if result.Type() != engine.TypeString {
				t.Errorf("strval() returned %s, expected string", result.Type())
			}
			if result.String() != tt.expected {
				t.Errorf("strval() = %q, expected %q", result.String(), tt.expected)
			}
		})
	}
}

// TestBoolval 测试 boolval 函数
func TestBoolval(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeConvert(e)

	tests := []struct {
		name     string
		args     []engine.Value
		expected bool
	}{
		{"true to bool", []engine.Value{engine.NewBool(true)}, true},
		{"false to bool", []engine.Value{engine.NewBool(false)}, false},
		{"non-zero int to bool", []engine.Value{engine.NewInt(42)}, true},
		{"zero int to bool", []engine.Value{engine.NewInt(0)}, false},
		{"non-zero float to bool", []engine.Value{engine.NewFloat(0.1)}, true},
		{"zero float to bool", []engine.Value{engine.NewFloat(0.0)}, false},
		{"non-empty string to bool", []engine.Value{engine.NewString("hello")}, true},
		{"empty string to bool", []engine.Value{engine.NewString("")}, false},
		{"non-empty array to bool", []engine.Value{engine.NewArray([]engine.Value{engine.NewInt(1)})}, true},
		{"empty array to bool", []engine.Value{engine.NewArray([]engine.Value{})}, false},
		{"null to bool", []engine.Value{engine.NewNull()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinBoolval(ctx, tt.args)
			if err != nil {
				t.Fatalf("boolval() error: %v", err)
			}
			if result.Type() != engine.TypeBool {
				t.Errorf("boolval() returned %s, expected bool", result.Type())
			}
			if result.Bool() != tt.expected {
				t.Errorf("boolval() = %v, expected %v", result.Bool(), tt.expected)
			}
		})
	}
}

// TestTypeConvertNames 测试函数名称列表
func TestTypeConvertNames(t *testing.T) {
	names := TypeConvertNames()
	expected := []string{"intval", "floatval", "strval", "boolval"}

	if len(names) != len(expected) {
		t.Errorf("TypeConvertNames() returned %d names, expected %d", len(names), len(expected))
	}

	for i, name := range expected {
		if i >= len(names) || names[i] != name {
			t.Errorf("TypeConvertNames()[%d] = %s, expected %s", i, names[i], name)
		}
	}
}
