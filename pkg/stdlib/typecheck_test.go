package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ==================== TypeCheckNames 测试 ====================

// TestTypeCheckNames 测试 TypeCheckNames 返回正确列表
func TestTypeCheckNames(t *testing.T) {
	names := TypeCheckNames()
	expected := []string{
		// 基础类型检查
		"is_null", "is_bool", "is_int", "is_float",
		"is_string", "is_array", "is_object", "is_func",
		// 类型别名
		"is_real", "is_double", "is_integer", "is_long",
		// 扩展类型检查
		"is_numeric", "is_scalar", "empty",
		// 流类型检查
		"is_stream",
		// 正则类型检查
		"is_regex",
		// 大数类型检查
		"is_bigint", "is_bigdecimal",
		// 新增 P1
		"is_error", "is_callable", "is_iterable",
	}

	if len(names) != len(expected) {
		t.Errorf("expected %d functions, got %d", len(expected), len(names))
	}

	for i, name := range expected {
		if i >= len(names) || names[i] != name {
			t.Errorf("position %d: expected %s, got %v", i, name, names)
			break
		}
	}
}

// TestRegisterTypeCheck 测试 RegisterTypeCheck 注册所有函数
func TestRegisterTypeCheck(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	RegisterTypeCheck(e)

	// 验证所有函数都已注册
	names := TypeCheckNames()
	for _, name := range names {
		fn := e.GetRegisteredFunc(name)
		if fn == nil {
			t.Errorf("function %s not registered", name)
		}
	}
}

// ==================== is_null 测试 ====================

// TestIsNullTrue 测试 is_null 返回 true
func TestIsNullTrue(t *testing.T) {
	result, err := callBuiltin("is_null", engine.NewNull())
	if err != nil {
		t.Fatalf("is_null error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_null(null) should be true")
	}
}

// TestIsNullFalse 测试 is_null 返回 false
func TestIsNullFalse(t *testing.T) {
	result, err := callBuiltin("is_null", engine.NewInt(1))
	if err != nil {
		t.Fatalf("is_null error: %v", err)
	}

	if result.Bool() {
		t.Error("is_null(1) should be false")
	}
}

// ==================== is_bool 测试 ====================

// TestIsBoolTrue 测试 is_bool 返回 true
func TestIsBoolTrue(t *testing.T) {
	result, err := callBuiltin("is_bool", engine.NewBool(true))
	if err != nil {
		t.Fatalf("is_bool error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_bool(true) should be true")
	}
}

// TestIsBoolFalse 测试 is_bool 返回 false
func TestIsBoolFalse(t *testing.T) {
	result, err := callBuiltin("is_bool", engine.NewInt(1))
	if err != nil {
		t.Fatalf("is_bool error: %v", err)
	}

	if result.Bool() {
		t.Error("is_bool(1) should be false")
	}
}

// ==================== is_int 测试 ====================

// TestIsIntTrue 测试 is_int 返回 true
func TestIsIntTrue(t *testing.T) {
	result, err := callBuiltin("is_int", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_int error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_int(42) should be true")
	}
}

// TestIsIntFalse 测试 is_int 返回 false
func TestIsIntFalse(t *testing.T) {
	result, err := callBuiltin("is_int", engine.NewFloat(3.14))
	if err != nil {
		t.Fatalf("is_int error: %v", err)
	}

	if result.Bool() {
		t.Error("is_int(3.14) should be false")
	}
}

// ==================== is_float 测试 ====================

// TestIsFloatTrue 测试 is_float 返回 true
func TestIsFloatTrue(t *testing.T) {
	result, err := callBuiltin("is_float", engine.NewFloat(3.14))
	if err != nil {
		t.Fatalf("is_float error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_float(3.14) should be true")
	}
}

// TestIsFloatFalse 测试 is_float 返回 false
func TestIsFloatFalse(t *testing.T) {
	result, err := callBuiltin("is_float", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_float error: %v", err)
	}

	if result.Bool() {
		t.Error("is_float(42) should be false")
	}
}

// ==================== is_string 测试 ====================

// TestIsStringTrue 测试 is_string 返回 true
func TestIsStringTrue(t *testing.T) {
	result, err := callBuiltin("is_string", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("is_string error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_string('hello') should be true")
	}
}

// TestIsStringFalse 测试 is_string 返回 false
func TestIsStringFalse(t *testing.T) {
	result, err := callBuiltin("is_string", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_string error: %v", err)
	}

	if result.Bool() {
		t.Error("is_string(42) should be false")
	}
}

// ==================== is_array 测试 ====================

// TestIsArrayTrue 测试 is_array 返回 true
func TestIsArrayTrue(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2)})
	result, err := callBuiltin("is_array", arr)
	if err != nil {
		t.Fatalf("is_array error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_array([1,2]) should be true")
	}
}

// TestIsArrayFalse 测试 is_array 返回 false
func TestIsArrayFalse(t *testing.T) {
	result, err := callBuiltin("is_array", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_array error: %v", err)
	}

	if result.Bool() {
		t.Error("is_array(42) should be false")
	}
}

// ==================== is_object 测试 ====================

// TestIsObjectTrue 测试 is_object 返回 true
func TestIsObjectTrue(t *testing.T) {
	obj := engine.NewObject(map[string]engine.Value{"a": engine.NewInt(1)})
	result, err := callBuiltin("is_object", obj)
	if err != nil {
		t.Fatalf("is_object error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_object({a:1}) should be true")
	}
}

// TestIsObjectFalse 测试 is_object 返回 false
func TestIsObjectFalse(t *testing.T) {
	result, err := callBuiltin("is_object", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_object error: %v", err)
	}

	if result.Bool() {
		t.Error("is_object(42) should be false")
	}
}

// ==================== is_func 测试 ====================

// TestIsFuncTrue 测试 is_func 返回 true
func TestIsFuncTrue(t *testing.T) {
	fn := engine.NewFunc("test", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
		return engine.NewNull(), nil
	})
	result, err := callBuiltin("is_func", fn)
	if err != nil {
		t.Fatalf("is_func error: %v", err)
	}

	if !result.Bool() {
		t.Error("is_func(function) should be true")
	}
}

// TestIsFuncFalse 测试 is_func 返回 false
func TestIsFuncFalse(t *testing.T) {
	result, err := callBuiltin("is_func", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_func error: %v", err)
	}

	if result.Bool() {
		t.Error("is_func(42) should be false")
	}
}

// ==================== 错误处理测试 ====================

// TestTypeCheckWrongArgCount 测试类型检查函数参数数量错误
func TestTypeCheckWrongArgCount(t *testing.T) {
	funcs := []string{
		"is_null", "is_bool", "is_int", "is_float",
		"is_string", "is_array", "is_object", "is_func",
		"is_real", "is_double", "is_integer", "is_long",
		"is_numeric", "is_scalar", "empty",
	}

	for _, name := range funcs {
		_, err := callBuiltin(name) // 无参数
		if err == nil {
			t.Errorf("%s() with no args should return error", name)
		}
	}
}

// TestTypeCheckMultipleArgs 测试类型检查函数多余参数
func TestTypeCheckMultipleArgs(t *testing.T) {
	_, err := callBuiltin("is_int", engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("is_int with 2 args should return error")
	}
}

// ==================== 类型别名测试 ====================

// TestIsRealAlias 测试 is_real 是 is_float 的别名
func TestIsRealAlias(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterTypeCheck(e)

	// 测试别名函数存在
	fn := e.GetRegisteredFunc("is_real")
	if fn == nil {
		t.Fatal("is_real should be registered as alias")
	}

	// 测试行为与 is_float 相同
	result, err := callBuiltin("is_real", engine.NewFloat(3.14))
	if err != nil {
		t.Fatalf("is_real error: %v", err)
	}
	if !result.Bool() {
		t.Error("is_real(3.14) should be true")
	}
}

// TestIsIntegerAlias 测试 is_integer 是 is_int 的别名
func TestIsIntegerAlias(t *testing.T) {
	result, err := callBuiltin("is_integer", engine.NewInt(42))
	if err != nil {
		t.Fatalf("is_integer error: %v", err)
	}
	if !result.Bool() {
		t.Error("is_integer(42) should be true")
	}

	result, err = callBuiltin("is_integer", engine.NewFloat(3.14))
	if err != nil {
		t.Fatalf("is_integer error: %v", err)
	}
	if result.Bool() {
		t.Error("is_integer(3.14) should be false")
	}
}

// ==================== is_numeric 测试 ====================

// TestIsNumericTrue 测试 is_numeric 对数字返回 true
func TestIsNumericTrue(t *testing.T) {
	tests := []engine.Value{
		engine.NewInt(42),
		engine.NewFloat(3.14),
	}

	for _, v := range tests {
		result, err := callBuiltin("is_numeric", v)
		if err != nil {
			t.Fatalf("is_numeric error: %v", err)
		}
		if !result.Bool() {
			t.Errorf("is_numeric(%v) should be true", v)
		}
	}
}

// TestIsNumericFalse 测试 is_numeric 对非数字返回 false
func TestIsNumericFalse(t *testing.T) {
	tests := []engine.Value{
		engine.NewString("hello"),
		engine.NewBool(true),
		engine.NewNull(),
		engine.NewArray([]engine.Value{}),
	}

	for _, v := range tests {
		result, err := callBuiltin("is_numeric", v)
		if err != nil {
			t.Fatalf("is_numeric error: %v", err)
		}
		if result.Bool() {
			t.Errorf("is_numeric(%v) should be false", v)
		}
	}
}

// ==================== is_scalar 测试 ====================

// TestIsScalarTrue 测试 is_scalar 对标量返回 true
func TestIsScalarTrue(t *testing.T) {
	tests := []engine.Value{
		engine.NewNull(),
		engine.NewBool(true),
		engine.NewInt(42),
		engine.NewFloat(3.14),
		engine.NewString("hello"),
	}

	for _, v := range tests {
		result, err := callBuiltin("is_scalar", v)
		if err != nil {
			t.Fatalf("is_scalar error: %v", err)
		}
		if !result.Bool() {
			t.Errorf("is_scalar(%v) should be true", v)
		}
	}
}

// TestIsScalarFalse 测试 is_scalar 对非标量返回 false
func TestIsScalarFalse(t *testing.T) {
	tests := []engine.Value{
		engine.NewArray([]engine.Value{}),
		engine.NewObject(map[string]engine.Value{}),
	}

	for _, v := range tests {
		result, err := callBuiltin("is_scalar", v)
		if err != nil {
			t.Fatalf("is_scalar error: %v", err)
		}
		if result.Bool() {
			t.Errorf("is_scalar(%v) should be false", v)
		}
	}
}

// ==================== empty 测试 ====================

// TestEmptyTrue 测试 empty 对各种"空"值返回 true
func TestEmptyTrue(t *testing.T) {
	tests := []struct {
		name  string
		value engine.Value
	}{
		{"null", engine.NewNull()},
		{"false", engine.NewBool(false)},
		{"zero int", engine.NewInt(0)},
		{"zero float", engine.NewFloat(0.0)},
		{"empty string", engine.NewString("")},
		{"string zero", engine.NewString("0")},
		{"empty array", engine.NewArray([]engine.Value{})},
		{"empty object", engine.NewObject(map[string]engine.Value{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("empty", tt.value)
			if err != nil {
				t.Fatalf("empty error: %v", err)
			}
			if !result.Bool() {
				t.Errorf("empty(%s) should be true", tt.name)
			}
		})
	}
}

// TestEmptyFalse 测试 empty 对非"空"值返回 false
func TestEmptyFalse(t *testing.T) {
	tests := []struct {
		name  string
		value engine.Value
	}{
		{"true", engine.NewBool(true)},
		{"non-zero int", engine.NewInt(42)},
		{"non-zero float", engine.NewFloat(0.1)},
		{"non-empty string", engine.NewString("hello")},
		{"non-empty array", engine.NewArray([]engine.Value{engine.NewInt(1)})},
		{"non-empty object", engine.NewObject(map[string]engine.Value{"a": engine.NewInt(1)})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("empty", tt.value)
			if err != nil {
				t.Fatalf("empty error: %v", err)
			}
			if result.Bool() {
				t.Errorf("empty(%s) should be false", tt.name)
			}
		})
	}
}
