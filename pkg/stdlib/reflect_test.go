package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// typeof 测试
// ============================================================================

func TestTypeOfInt(t *testing.T) {
	result, err := callBuiltin("typeof", engine.NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "int" {
		t.Errorf("typeof(42) = %q, expected 'int'", result.String())
	}
}

func TestTypeOfString(t *testing.T) {
	result, err := callBuiltin("typeof", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "string" {
		t.Errorf("typeof('hello') = %q, expected 'string'", result.String())
	}
}

func TestTypeOfNull(t *testing.T) {
	result, err := callBuiltin("typeof", engine.NewNull())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "null" {
		t.Errorf("typeof(null) = %q, expected 'null'", result.String())
	}
}

func TestTypeOfArray(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
	result, err := callBuiltin("typeof", arr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "array" {
		t.Errorf("typeof([]) = %q, expected 'array'", result.String())
	}
}

// ============================================================================
// exists 测试
// ============================================================================

func TestExistsGlobal(t *testing.T) {
	script := `$myvar = 42;
varexists("$myvar")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("varexists('$myvar') should be true")
	}
}

func TestExistsUndefined(t *testing.T) {
	script := `varexists("nonexistent_var_12345")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("varexists('nonexistent_var_12345') should be false")
	}
}

// ============================================================================
// getvar/setvar 测试
// ============================================================================

func TestGetVar(t *testing.T) {
	script := `$x = 100;
getvar("$x")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 100 {
		t.Errorf("getvar('$x') = %d, expected 100", result.Int())
	}
}

func TestSetVar(t *testing.T) {
	// setvar modifies VM globals, but the script's local copy is not synced.
	// This is a known limitation. To verify setvar works, use getvar to read it back.
	script := `setvar("$x", 20);
getvar("$x")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("expected 20, got %d", result.Int())
	}
}

// ============================================================================
// listvars 测试
// ============================================================================

func TestListVars(t *testing.T) {
	script := `$a = 1;
$b = 2;
listvars()`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type() != engine.TypeArray {
		t.Fatalf("listvars() should return array, got %s", result.Type())
	}

	// 至少包含 a 和 b
	arr := result.Array()
	found := map[string]bool{}
	for _, v := range arr {
		found[v.String()] = true
	}

	if !found["a"] && !found["$a"] {
		t.Error("listvars() should contain 'a' or '$a'")
	}
	if !found["b"] && !found["$b"] {
		t.Error("listvars() should contain 'b' or '$b'")
	}
}

// ============================================================================
// listfns 测试
// ============================================================================

func TestListFns(t *testing.T) {
	script := `listfns()`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type() != engine.TypeArray {
		t.Fatalf("listfns() should return array, got %s", result.Type())
	}

	// 应该包含内置函数
	arr := result.Array()
	found := map[string]bool{}
	for _, v := range arr {
		found[v.String()] = true
	}

	// 检查是否包含一些内置函数
	if !found["print"] {
		t.Error("listfns() should contain 'print'")
	}
	if !found["len"] {
		t.Error("listfns() should contain 'len'")
	}
}

// ============================================================================
// fn_exists 测试
// ============================================================================

func TestFnExistsBuiltin(t *testing.T) {
	script := `fn_exists("print")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("fn_exists('print') should be true")
	}
}

func TestFnExistsDefined(t *testing.T) {
	script := `fn myfunc() { return 1; }
fn_exists("myfunc")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("fn_exists('myfunc') should be true")
	}
}

func TestFnExistsNonexistent(t *testing.T) {
	script := `fn_exists("nonexistent_func_12345")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("fn_exists('nonexistent_func_12345') should be false")
	}
}

// ============================================================================
// getfninfo 测试
// ============================================================================

func TestGetFnInfoBuiltin(t *testing.T) {
	script := `getfninfo("print")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type() != engine.TypeObject {
		t.Fatalf("getfninfo should return object, got %s", result.Type())
	}

	obj := result.Object()
	if v, ok := obj["name"]; !ok || v.String() != "print" {
		t.Errorf("name should be 'print', got %v", v)
	}
}

func TestGetFnInfoDefined(t *testing.T) {
	script := `fn add(a, b) { return a + b; }
getfninfo("add")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type() != engine.TypeObject {
		t.Fatalf("getfninfo should return object, got %s", result.Type())
	}

	obj := result.Object()
	if v, ok := obj["name"]; !ok || v.String() != "add" {
		t.Errorf("name should be 'add', got %v", v)
	}
	if v, ok := obj["paramCount"]; !ok || v.Int() != 2 {
		t.Errorf("paramCount should be 2, got %v", v)
	}
}

func TestGetFnInfoNonexistent(t *testing.T) {
	script := `getfninfo("no_such_func")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("getfninfo of nonexistent should return null, got %v", result)
	}
}

// ============================================================================
// callfn 测试
// ============================================================================

func TestCallFn(t *testing.T) {
	script := `fn double(n) { return n * 2; }
callfn("double", 21)`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("callfn('double', 21) = %d, expected 42", result.Int())
	}
}

func TestCallFnBuiltin(t *testing.T) {
	script := `callfn("len", "hello")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 5 {
		t.Errorf("callfn('len', 'hello') = %d, expected 5", result.Int())
	}
}

func TestCallFnArrayArg(t *testing.T) {
	script := `fn sum3(a, b, c) { return a + b + c; }
callfn("sum3", [10, 20, 30])`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("callfn('sum3', [10,20,30]) = %d, expected 60", result.Int())
	}
}

func TestCallFnNonexistent(t *testing.T) {
	script := `callfn("nonexistent_func_12345")`

	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("callfn with nonexistent function should return error")
	}
}
