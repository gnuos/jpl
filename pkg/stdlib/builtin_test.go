package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/engine"
)

func callBuiltin(name string, args ...engine.Value) (engine.Value, error) {
	e := engine.NewEngine()
	RegisterAll(e)
	fn := e.GetRegisteredFunc(name)
	if fn == nil {
		return nil, nil
	}
	ctx := engine.NewContext(e, nil)
	return fn(ctx, args)
}

// compileAndRunBuiltins 编译脚本并执行（注册内置函数）
func compileAndRunBuiltins(script string) (engine.Value, error) {
	prog, err := engine.CompileString(script)
	if err != nil {
		return nil, err
	}
	e := engine.NewEngine()
	RegisterAll(e)
	defer e.Close()
	vm := engine.NewVMWithProgram(e, prog)
	err = vm.Execute()
	if err != nil {
		return nil, err
	}
	return vm.GetResult(), nil
}

// ============================================================================
// I/O 函数测试
// ============================================================================

func TestBuiltinPrint(t *testing.T) {
	result, err := callBuiltin("print", engine.NewString("hello"), engine.NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("print should return null, got %v", result)
	}
}

func TestBuiltinPrintln(t *testing.T) {
	result, err := callBuiltin("println", engine.NewString("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("println should return null, got %v", result)
	}
}

// TestPrintToStream 测试 print(stream, "msg") 流参数
func TestPrintToStream(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "print_test.txt")
	stream := engine.NewFileStream(tmpFile, engine.StreamWrite)

	result, err := callBuiltin("print", stream, engine.NewString("hello"))
	if err != nil {
		t.Fatalf("print to stream error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("print should return null, got %v", result)
	}

	// 关闭流并读取文件
	callBuiltin("fclose", stream)
	data, _ := os.ReadFile(tmpFile)
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
}

// TestPrintlnToStream 测试 println(stream, "msg") 流参数
func TestPrintlnToStream(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "println_test.txt")
	stream := engine.NewFileStream(tmpFile, engine.StreamWrite)

	_, err := callBuiltin("println", stream, engine.NewString("error"))
	if err != nil {
		t.Fatalf("println to stream error: %v", err)
	}

	callBuiltin("fclose", stream)
	data, _ := os.ReadFile(tmpFile)
	if string(data) != "error\n" {
		t.Errorf("expected 'error\\n', got %q", string(data))
	}
}

// TestPrintMultipleArgsToStream 测试多参数输出到流
func TestPrintMultipleArgsToStream(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "print_multi_test.txt")
	stream := engine.NewFileStream(tmpFile, engine.StreamWrite)

	_, err := callBuiltin("print", stream, engine.NewString("count:"), engine.NewInt(42))
	if err != nil {
		t.Fatalf("print to stream error: %v", err)
	}

	callBuiltin("fclose", stream)
	data, _ := os.ReadFile(tmpFile)
	if string(data) != "count: 42" {
		t.Errorf("expected 'count: 42', got %q", string(data))
	}
}

// TestPrintNoStreamFirstArg 测试首参数非流时的正常行为
func TestPrintNoStreamFirstArg(t *testing.T) {
	// 首参数是字符串，应作为普通输出
	result, err := callBuiltin("print", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("print should return null, got %v", result)
	}
}

func TestBuiltinEcho(t *testing.T) {
	result, err := callBuiltin("echo", engine.NewString("hello"), engine.NewString("world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("echo expected 'hello world', got '%s'", result.String())
	}
}

func TestBuiltinEchoSingle(t *testing.T) {
	result, err := callBuiltin("echo", engine.NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "42" {
		t.Errorf("echo expected '42', got '%s'", result.String())
	}
}

// ============================================================================
// 工具函数测试
// ============================================================================

func TestBuiltinLenString(t *testing.T) {
	result, err := callBuiltin("len", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 5 {
		t.Errorf("len('hello') expected 5, got %d", result.Int())
	}
}

func TestBuiltinLenArray(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(1),
		engine.NewInt(2),
		engine.NewInt(3),
	})
	result, err := callBuiltin("len", arr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("len([1,2,3]) expected 3, got %d", result.Int())
	}
}

func TestBuiltinLenObject(t *testing.T) {
	obj := engine.NewObject(map[string]engine.Value{
		"a": engine.NewInt(1),
		"b": engine.NewInt(2),
	})
	result, err := callBuiltin("len", obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 2 {
		t.Errorf("len({a:1, b:2}) expected 2, got %d", result.Int())
	}
}

func TestBuiltinLenEmptyString(t *testing.T) {
	result, err := callBuiltin("len", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("len('') expected 0, got %d", result.Int())
	}
}

func TestBuiltinLenInvalidType(t *testing.T) {
	_, err := callBuiltin("len", engine.NewInt(42))
	if err == nil {
		t.Error("len(42) should return error")
	}
}

func TestBuiltinLenWrongArgCount(t *testing.T) {
	_, err := callBuiltin("len", engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("len(1, 2) should return error")
	}
}

func TestBuiltinAssertTrue(t *testing.T) {
	result, err := callBuiltin("assert", engine.NewBool(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("assert(true) should return null, got %v", result)
	}
}

func TestBuiltinAssertFalse(t *testing.T) {
	_, err := callBuiltin("assert", engine.NewBool(false))
	if err == nil {
		t.Error("assert(false) should return error")
	}
}

func TestBuiltinAssertFalseWithMessage(t *testing.T) {
	_, err := callBuiltin("assert", engine.NewBool(false), engine.NewString("custom error"))
	if err == nil {
		t.Error("assert(false, 'custom error') should return error")
	}
}

func TestBuiltinFormat(t *testing.T) {
	result, err := callBuiltin("format",
		engine.NewString("Hello, %s!"),
		engine.NewString("World"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "Hello, World!" {
		t.Errorf("format expected 'Hello, World!', got '%s'", result.String())
	}
}

func TestBuiltinFormatMultiple(t *testing.T) {
	result, err := callBuiltin("format",
		engine.NewString("%s is %s"),
		engine.NewString("name"),
		engine.NewString("value"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "name is value" {
		t.Errorf("expected 'name is value', got '%s'", result.String())
	}
}

func TestBuiltinFormatNoArgs(t *testing.T) {
	result, err := callBuiltin("format", engine.NewString("no placeholders"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "no placeholders" {
		t.Errorf("expected 'no placeholders', got '%s'", result.String())
	}
}

// ============================================================================
// 类型检查函数测试
// ============================================================================

func TestBuiltinIsNull(t *testing.T) {
	tests := []struct {
		name     string
		arg      engine.Value
		expected bool
	}{
		{"null", engine.NewNull(), true},
		{"int", engine.NewInt(0), false},
		{"string", engine.NewString(""), false},
		{"bool", engine.NewBool(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("is_null", tt.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("is_null(%s) expected %v, got %v", tt.name, tt.expected, result.Bool())
			}
		})
	}
}

func TestBuiltinIsBool(t *testing.T) {
	tests := []struct {
		name     string
		arg      engine.Value
		expected bool
	}{
		{"true", engine.NewBool(true), true},
		{"false", engine.NewBool(false), true},
		{"int", engine.NewInt(0), false},
		{"null", engine.NewNull(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("is_bool", tt.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("is_bool(%s) expected %v, got %v", tt.name, tt.expected, result.Bool())
			}
		})
	}
}

func TestBuiltinIsInt(t *testing.T) {
	result, _ := callBuiltin("is_int", engine.NewInt(42))
	if !result.Bool() {
		t.Error("is_int(42) should be true")
	}

	result, _ = callBuiltin("is_int", engine.NewFloat(3.14))
	if result.Bool() {
		t.Error("is_int(3.14) should be false")
	}
}

func TestBuiltinIsFloat(t *testing.T) {
	result, _ := callBuiltin("is_float", engine.NewFloat(3.14))
	if !result.Bool() {
		t.Error("is_float(3.14) should be true")
	}

	result, _ = callBuiltin("is_float", engine.NewInt(42))
	if result.Bool() {
		t.Error("is_float(42) should be false")
	}
}

func TestBuiltinIsString(t *testing.T) {
	result, _ := callBuiltin("is_string", engine.NewString("hello"))
	if !result.Bool() {
		t.Error("is_string('hello') should be true")
	}

	result, _ = callBuiltin("is_string", engine.NewInt(42))
	if result.Bool() {
		t.Error("is_string(42) should be false")
	}
}

func TestBuiltinIsArray(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
	result, _ := callBuiltin("is_array", arr)
	if !result.Bool() {
		t.Error("is_array([1]) should be true")
	}

	result, _ = callBuiltin("is_array", engine.NewInt(42))
	if result.Bool() {
		t.Error("is_array(42) should be false")
	}
}

func TestBuiltinIsObject(t *testing.T) {
	obj := engine.NewObject(map[string]engine.Value{"a": engine.NewInt(1)})
	result, _ := callBuiltin("is_object", obj)
	if !result.Bool() {
		t.Error("is_object({a:1}) should be true")
	}

	result, _ = callBuiltin("is_object", engine.NewInt(42))
	if result.Bool() {
		t.Error("is_object(42) should be false")
	}
}

func TestBuiltinIsFunc(t *testing.T) {
	fn := engine.NewFunc("test", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
		return engine.NewNull(), nil
	})
	result, _ := callBuiltin("is_func", fn)
	if !result.Bool() {
		t.Error("is_func(fn) should be true")
	}

	result, _ = callBuiltin("is_func", engine.NewInt(42))
	if result.Bool() {
		t.Error("is_func(42) should be false")
	}
}

// ============================================================================
// 函数式编程集成测试
// ============================================================================

func TestMapIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = map($arr, $x -> $x * 2);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{2, 4, 6, 8, 10}
	if len(arr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(arr))
	}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

func TestMapEmptyArray(t *testing.T) {
	script := `$result = map([], $x -> $x * 2);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Array()) != 0 {
		t.Errorf("expected empty array, got %d elements", len(result.Array()))
	}
}

func TestFilterIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5, 6];
$result = filter($arr, $x -> $x > 3);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{4, 5, 6}
	if len(arr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(arr))
	}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

func TestFilterWithEven(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5, 6];
$result = filter($arr, $x -> $x % 2 == 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{2, 4, 6}
	if len(arr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(arr))
	}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

func TestReduceIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = reduce($arr, fn($acc, $x) { return $acc + $x; });
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Int() != 15 {
		t.Errorf("expected 15, got %d", result.Int())
	}
}

func TestReduceWithInitial(t *testing.T) {
	script := `$arr = [1, 2, 3];
$result = reduce($arr, fn($acc, $x) { return $acc + $x; }, 100);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Int() != 106 {
		t.Errorf("expected 106, got %d", result.Int())
	}
}

func TestReduceEmptyArray(t *testing.T) {
	script := `$result = reduce([], fn($acc, $x) { return $acc + $x; }, 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Int() != 0 {
		t.Errorf("expected 0, got %d", result.Int())
	}
}

func TestReduceMultiply(t *testing.T) {
	script := `$arr = [1, 2, 3, 4];
$result = reduce($arr, fn($acc, $x) { return $acc * $x; });
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Int() != 24 {
		t.Errorf("expected 24, got %d", result.Int())
	}
}

func TestMapWithLambda(t *testing.T) {
	script := `$double = fn($n) { return $n * 2; };
$arr = [10, 20, 30];
$result = map($arr, $double);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{20, 40, 60}
	if len(arr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(arr))
	}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// ============================================================================
// find/some/every/sort/contains 测试
// ============================================================================

func TestFindIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = find($arr, $x -> $x > 3);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 4 {
		t.Errorf("expected 4, got %d", result.Int())
	}
}

func TestFindNotFound(t *testing.T) {
	script := `$arr = [1, 2, 3];
$result = find($arr, $x -> $x > 10);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("expected null, got %v", result)
	}
}

func TestSomeIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = some($arr, $x -> $x > 4);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}
}

func TestSomeFalse(t *testing.T) {
	script := `$arr = [1, 2, 3];
$result = some($arr, $x -> $x > 10);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false")
	}
}

func TestEveryIntegration(t *testing.T) {
	script := `$arr = [2, 4, 6];
$result = every($arr, $x -> $x % 2 == 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}
}

func TestEveryFalse(t *testing.T) {
	script := `$arr = [2, 3, 6];
$result = every($arr, $x -> $x % 2 == 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false")
	}
}

func TestSortDefault(t *testing.T) {
	script := `$arr = [3, 1, 4, 1, 5, 9, 2, 6];
$result = sort($arr);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{1, 1, 2, 3, 4, 5, 6, 9}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

func TestSortDescending(t *testing.T) {
	script := `$arr = [3, 1, 4, 1, 5];
$result = sort($arr, fn($a, $b) { return $a > $b; });
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []int64{5, 4, 3, 1, 1}
	for i, v := range arr {
		if v.Int() != expected[i] {
			t.Errorf("index %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

func TestSortDoesNotMutate(t *testing.T) {
	script := `$arr = [3, 1, 2];
$sorted = sort($arr);
len($arr)`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("original array should still have 3 elements, got %d", result.Int())
	}
}

func TestContainsIntegration(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = contains($arr, 3);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}
}

func TestContainsFalse(t *testing.T) {
	script := `$arr = [1, 2, 3];
$result = contains($arr, 99);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false")
	}
}

// TestFunctionNames 测试内置函数名列表导出
func TestFunctionNames(t *testing.T) {
	names := FunctionNames()

	if len(names) == 0 {
		t.Fatal("FunctionNames() 返回空列表")
	}

	// 检查关键函数是否存在
	required := map[string]bool{
		"print": true, "println": true, "echo": true,
		"len": true, "assert": true, "format": true,
		"map": true, "filter": true, "reduce": true,
		"find": true, "some": true, "every": true,
		"sort": true, "contains": true,
		"is_null": true, "is_bool": true, "is_int": true, "is_string": true,
		"is_array": true, "is_object": true, "is_func": true,
	}

	nameSet := make(map[string]bool, len(names))
	for _, name := range names {
		nameSet[name] = true
	}

	for fn := range required {
		if !nameSet[fn] {
			t.Errorf("FunctionNames() 缺少函数: %q", fn)
		}
	}
}
