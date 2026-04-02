package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// func_num_args 测试
// ============================================================================

func TestFuncNumArgsBasic(t *testing.T) {
	script := `
fn test(a, b, c) {
	return func_num_args();
}
test(1, 2, 3)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("func_num_args() expected 3, got %d", result.Int())
	}
}

func TestFuncNumArgsZero(t *testing.T) {
	script := `
fn test() {
	return func_num_args();
}
test()`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("func_num_args() expected 0, got %d", result.Int())
	}
}

func TestFuncNumArgsInMain(t *testing.T) {
	script := `func_num_args()`
	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("func_num_args() in main should return error")
	}
}

func TestFuncNumArgsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("func_num_args", engine.NewInt(1))
	if err == nil {
		t.Error("func_num_args(1) should return error")
	}
}

// ============================================================================
// func_get_arg 测试
// ============================================================================

func TestFuncGetArgBasic(t *testing.T) {
	script := `
fn test(a, b, c) {
	return func_get_arg(1);
}
test(10, 20, 30)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("func_get_arg(1) expected 20, got %d", result.Int())
	}
}

func TestFuncGetArgFirst(t *testing.T) {
	script := `
fn test(a, b) {
	return func_get_arg(0);
}
test("hello", "world")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello" {
		t.Errorf("func_get_arg(0) expected 'hello', got '%s'", result.String())
	}
}

func TestFuncGetArgOutOfRange(t *testing.T) {
	script := `
fn test(a) {
	return func_get_arg(5);
}
test(1)`
	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("func_get_arg(5) out of range should return error")
	}
}

func TestFuncGetArgNegativeIndex(t *testing.T) {
	script := `
fn test(a) {
	return func_get_arg(-1);
}
test(1)`
	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("func_get_arg(-1) should return error")
	}
}

func TestFuncGetArgWrongArgType(t *testing.T) {
	script := `
fn test(a) {
	return func_get_arg("0");
}
test(1)`
	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("func_get_arg('0') should return error (non-integer)")
	}
}

// ============================================================================
// func_get_args 测试
// ============================================================================

func TestFuncGetArgsBasic(t *testing.T) {
	script := `
fn test(a, b, c) {
	return func_get_args()[1];
}
test(10, 20, 30)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("func_get_args()[1] expected 20, got %d", result.Int())
	}
}

func TestFuncGetArgsAll(t *testing.T) {
	script := `
fn test(a, b, c) {
	return len(func_get_args());
}
test(1, 2, 3)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("len(func_get_args()) expected 3, got %d", result.Int())
	}
}

func TestFuncGetArgsEmpty(t *testing.T) {
	script := `
fn test() {
	return len(func_get_args());
}
test()`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("len(func_get_args()) expected 0, got %d", result.Int())
	}
}

func TestFuncGetArgsInMain(t *testing.T) {
	script := `func_get_args()`
	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("func_get_args() in main should return error")
	}
}

// ============================================================================
// function_exists 测试
// ============================================================================

func TestFunctionExistsTrue(t *testing.T) {
	script := `
fn myFunc() {
	return 42;
}
function_exists("myFunc")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("function_exists('myFunc') should return true")
	}
}

func TestFunctionExistsFalse(t *testing.T) {
	script := `function_exists("nonexistent")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("function_exists('nonexistent') should return false")
	}
}

func TestFunctionExistsBuiltin(t *testing.T) {
	script := `function_exists("print")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("function_exists('print') should return true")
	}
}

func TestFunctionExistsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("function_exists")
	if err == nil {
		t.Error("function_exists() should return error")
	}
}

func TestFunctionExistsNotString(t *testing.T) {
	_, err := callBuiltin("function_exists", engine.NewInt(42))
	if err == nil {
		t.Error("function_exists(42) should return error")
	}
}

// ============================================================================
// is_callable 测试
// ============================================================================

func TestIsCallableFunctionName(t *testing.T) {
	script := `
fn myFunc() {
	return 42;
}
is_callable("myFunc")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("is_callable('myFunc') should return true")
	}
}

func TestIsCallableBuiltin(t *testing.T) {
	script := `is_callable("print")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("is_callable('print') should return true")
	}
}

func TestIsCallableNonexistent(t *testing.T) {
	script := `is_callable("nonexistent")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("is_callable('nonexistent') should return false")
	}
}

func TestIsCallableNotString(t *testing.T) {
	script := `is_callable(42)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("is_callable(42) should return false")
	}
}

func TestIsCallableWrongArgCount(t *testing.T) {
	_, err := callBuiltin("is_callable")
	if err == nil {
		t.Error("is_callable() should return error")
	}
}

// ============================================================================
// get_defined_functions 测试
// ============================================================================

func TestGetDefinedFunctions(t *testing.T) {
	script := `
fn myFunc1() {}
fn myFunc2() {}
len(get_defined_functions())`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 至少包含 myFunc1, myFunc2 和内置函数
	if result.Int() < 2 {
		t.Errorf("get_defined_functions() should return at least 2 functions, got %d", result.Int())
	}
}

func TestGetDefinedFunctionsContainsBuiltin(t *testing.T) {
	script := `
fn myFunc() {}
found = false
foreach (f in get_defined_functions()) {
	if (f == "print") {
		found = true
	}
}
found`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("get_defined_functions() should contain 'print'")
	}
}

func TestGetDefinedFunctionsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("get_defined_functions", engine.NewInt(1))
	if err == nil {
		t.Error("get_defined_functions(1) should return error")
	}
}

// ============================================================================
// get_defined_constants 测试
// ============================================================================

func TestGetDefinedConstants(t *testing.T) {
	script := `len(get_defined_constants())`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该包含一些预定义常量
	if result.Int() < 1 {
		t.Errorf("get_defined_constants() should return at least 1 constant, got %d", result.Int())
	}
}

func TestGetDefinedConstantsContainsPI(t *testing.T) {
	script := `
found = false
foreach (c in get_defined_constants()) {
	if (c == "PI") {
		found = true
	}
}
found`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("get_defined_constants() should contain 'PI'")
	}
}

func TestGetDefinedConstantsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("get_defined_constants", engine.NewInt(1))
	if err == nil {
		t.Error("get_defined_constants(1) should return error")
	}
}

// ============================================================================
// jpl_version 测试
// ============================================================================

func TestJPLVersion(t *testing.T) {
	result, err := callBuiltin("jpl_version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() == "" {
		t.Error("jpl_version() should return non-empty string")
	}
}

func TestJPLVersionWrongArgCount(t *testing.T) {
	_, err := callBuiltin("jpl_version", engine.NewInt(1))
	if err == nil {
		t.Error("jpl_version(1) should return error")
	}
}

// ============================================================================
// utf8_encode 测试
// ============================================================================

func TestUTF8EncodeBasic(t *testing.T) {
	result, err := callBuiltin("utf8_encode", engine.NewString("Hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "Hello" 的 UTF-8 十六进制是 48656c6c6f
	if result.String() != "48656c6c6f" {
		t.Errorf("utf8_encode('Hello') expected '48656c6c6f', got '%s'", result.String())
	}
}

func TestUTF8EncodeChinese(t *testing.T) {
	result, err := callBuiltin("utf8_encode", engine.NewString("中文"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "中文" 的 UTF-8 十六进制是 e4b8ade69687
	if result.String() != "e4b8ade69687" {
		t.Errorf("utf8_encode('中文') expected 'e4b8ade69687', got '%s'", result.String())
	}
}

func TestUTF8EncodeEmpty(t *testing.T) {
	result, err := callBuiltin("utf8_encode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("utf8_encode('') expected '', got '%s'", result.String())
	}
}

func TestUTF8EncodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("utf8_encode")
	if err == nil {
		t.Error("utf8_encode() should return error")
	}
}

func TestUTF8EncodeNotString(t *testing.T) {
	_, err := callBuiltin("utf8_encode", engine.NewInt(42))
	if err == nil {
		t.Error("utf8_encode(42) should return error")
	}
}

// ============================================================================
// utf8_decode 测试
// ============================================================================

func TestUTF8DecodeBasic(t *testing.T) {
	result, err := callBuiltin("utf8_decode", engine.NewString("48656c6c6f"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "Hello" {
		t.Errorf("utf8_decode('48656c6c6f') expected 'Hello', got '%s'", result.String())
	}
}

func TestUTF8DecodeChinese(t *testing.T) {
	result, err := callBuiltin("utf8_decode", engine.NewString("e4b8ade69687"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "中文" {
		t.Errorf("utf8_decode('e4b8ade69687') expected '中文', got '%s'", result.String())
	}
}

func TestUTF8DecodeEmpty(t *testing.T) {
	result, err := callBuiltin("utf8_decode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("utf8_decode('') expected '', got '%s'", result.String())
	}
}

func TestUTF8DecodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("utf8_decode")
	if err == nil {
		t.Error("utf8_decode() should return error")
	}
}

func TestUTF8DecodeNotString(t *testing.T) {
	_, err := callBuiltin("utf8_decode", engine.NewInt(42))
	if err == nil {
		t.Error("utf8_decode(42) should return error")
	}
}

func TestUTF8DecodeInvalidHex(t *testing.T) {
	_, err := callBuiltin("utf8_decode", engine.NewString("invalid"))
	if err == nil {
		t.Error("utf8_decode('invalid') should return error for invalid hex")
	}
}

// ============================================================================
// 集成测试
// ============================================================================

func TestFuncArgsIntegration(t *testing.T) {
	script := `
fn sum(a, b, c) {
	return func_get_args()[0] + func_get_args()[1] + func_get_args()[2];
}
sum(10, 20, 30)`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("sum(10, 20, 30) expected 60, got %d", result.Int())
	}
}

func TestFunctionExistsIntegration(t *testing.T) {
	script := `
fn greet(name) {
	return "Hello, " + name;
}
function_exists("greet") && is_callable("greet")`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("function_exists('greet') && is_callable('greet') should return true")
	}
}

func TestUTF8RoundTrip(t *testing.T) {
	original := "Hello, 世界!"
	script := `$encoded = utf8_encode("` + original + `");
$decoded = utf8_decode($encoded);
$decoded`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != original {
		t.Errorf("UTF8 round trip failed, expected '%s', got '%s'", original, result.String())
	}
}
