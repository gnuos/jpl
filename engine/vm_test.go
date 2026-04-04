package engine

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// 辅助函数
// ============================================================================

// compileAndRun 编译脚本并执行，返回结果和错误
func compileAndRun(script string) (Value, error) {
	prog, err := CompileString(script)
	if err != nil {
		return nil, err
	}
	engine := NewEngine()
	defer engine.Close()
	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		return nil, err
	}
	return vm.GetResult(), nil
}

// compileAndRunWithFuncs 编译脚本并执行，支持注册自定义函数
func compileAndRunWithFuncs(script string, registerFuncs func(*Engine)) (Value, error) {
	prog, err := CompileString(script)
	if err != nil {
		return nil, err
	}
	eng := NewEngine()
	defer eng.Close()
	if registerFuncs != nil {
		registerFuncs(eng)
	}
	vm := newVMWithProgram(eng, prog)
	err = vm.Execute()
	if err != nil {
		return nil, err
	}
	return vm.GetResult(), nil
}

// ============================================================================
// 基础算术测试
// ============================================================================

func TestVMAdd(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = 1 + 2; $a", 3},
		{"$a = 10 + 20; $a", 30},
		{"$a = 0 + 0; $a", 0},
		{"$a = -5 + 3; $a", -2},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

func TestVMSub(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = 5 - 3; $a", 2},
		{"$a = 10 - 20; $a", -10},
		{"$a = 0 - 5; $a", -5},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

func TestVMMul(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = 3 * 4; $a", 12},
		{"$a = -2 * 5; $a", -10},
		{"$a = 0 * 100; $a", 0},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

func TestVMDiv(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = 10 / 2; $a", 5},
		{"$a = 15 / 3; $a", 5},
		{"$a = 7 / 2; $a", 3}, // 整数除法
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

func TestVMMod(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = 10 % 3; $a", 1},
		{"$a = 15 % 5; $a", 0},
		{"$a = 7 % 4; $a", 3},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

func TestVMNeg(t *testing.T) {
	tests := []struct {
		script   string
		expected int64
	}{
		{"$a = -5; $a", -5},
		{"$a = -(-3); $a", 3},
		{"$a = -0; $a", 0},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int())
			}
		})
	}
}

// ============================================================================
// 比较运算测试
// ============================================================================

func TestVMComparison(t *testing.T) {
	tests := []struct {
		script   string
		expected bool
	}{
		{"$a = 5 == 5; $a", true},
		{"$a = 5 == 3; $a", false},
		{"$a = 5 != 3; $a", true},
		{"$a = 5 != 5; $a", false},
		{"$a = 3 < 5; $a", true},
		{"$a = 5 < 3; $a", false},
		{"$a = 5 > 3; $a", true},
		{"$a = 3 > 5; $a", false},
		{"$a = 5 <= 5; $a", true},
		{"$a = 5 <= 3; $a", false},
		{"$a = 5 >= 5; $a", true},
		{"$a = 3 >= 5; $a", false},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.Bool())
			}
		})
	}
}

// ============================================================================
// 逻辑运算测试
// ============================================================================

func TestVMLogic(t *testing.T) {
	tests := []struct {
		script   string
		expected bool
	}{
		{"$a = true && true; $a", true},
		{"$a = true && false; $a", false},
		{"$a = false && true; $a", false},
		{"$a = false && false; $a", false},
		{"$a = true || true; $a", true},
		{"$a = true || false; $a", true},
		{"$a = false || true; $a", true},
		{"$a = false || false; $a", false},
		{"$a = !true; $a", false},
		{"$a = !false; $a", true},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.Bool())
			}
		})
	}
}

// ============================================================================
// 字符串测试
// ============================================================================

func TestVMString(t *testing.T) {
	tests := []struct {
		script   string
		expected string
	}{
		{`$a = "hello"; $a`, "hello"},
		{`$a = "hello" .. " " .. "world"; $a`, "hello world"},
		{`$a = ""; $a`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

// ============================================================================
// 数组测试
// ============================================================================

func TestVMArray(t *testing.T) {
	script := `$a = [1, 2, 3]; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("expected array length 3, got %d", len(arr))
	}
	for i, expected := range []int64{1, 2, 3} {
		if arr[i].Int() != expected {
			t.Errorf("expected arr[%d] = %d, got %d", i, expected, arr[i].Int())
		}
	}
}

func TestVMArrayIndex(t *testing.T) {
	script := `$a = [10, 20, 30]; $b = $a[1]; $b`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("expected 20, got %d", result.Int())
	}
}

// TestVMArrayNegativeIndex 测试负数索引功能
// -1 表示最后一个元素，-2 表示倒数第二个，以此类推
func TestVMArrayNegativeIndex(t *testing.T) {
	script := `
		$a = [10, 20, 30, 40, 50];
		$first = $a[0];
		$last = $a[-1];
		$secondLast = $a[-2];
		$first + $last + $secondLast
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10 + 50 + 40 = 100
	if result.Int() != 100 {
		t.Errorf("expected 100 (10+50+40), got %d", result.Int())
	}
}

// TestVMArrayNegativeIndexOutOfBounds 测试负数索引越界情况
func TestVMArrayNegativeIndexOutOfBounds(t *testing.T) {
	script := `
		$a = [1, 2, 3];
		$a[-10]  // 越界，应该返回 null
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("expected null for out of bounds negative index, got %v", result)
	}
}

// ============================================================================
// 对象测试
// ============================================================================

func TestVMObject(t *testing.T) {
	script := `$a = {"name": "test", "value": 42}; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	obj := result.Object()
	if obj["name"].String() != "test" {
		t.Errorf("expected name = 'test', got %q", obj["name"].String())
	}
	if obj["value"].Int() != 42 {
		t.Errorf("expected value = 42, got %d", obj["value"].Int())
	}
}

func TestVMObjectMember(t *testing.T) {
	script := `$a = {"name": "test", "value": 42}; $b = $a.name; $b`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "test" {
		t.Errorf("expected 'test', got %q", result.String())
	}
}

// ============================================================================
// 控制流测试
// ============================================================================

func TestVMIf(t *testing.T) {
	// 简化测试：只测试条件跳转是否正确工作
	script := `$result = 0; $a = 10; if ($a > 5) { $result = 1; } $result`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

func TestVMWhile(t *testing.T) {
	done := make(chan bool)
	go func() {
		script := `$sum = 0; $i = 0; while ($i < 5) { $sum = $sum + $i; $i = $i + 1; } $sum`
		result, err := compileAndRun(script)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if result.Int() != 10 { // 0+1+2+3+4 = 10
			t.Errorf("expected 10, got %d", result.Int())
		}
		done <- true
	}()
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("test timed out after 5 seconds")
	}
}

func TestVMFor(t *testing.T) {
	done := make(chan bool)
	go func() {
		script := `$sum = 0; for ($i = 0; $i < 5; $i = $i + 1) { $sum = $sum + $i; } $sum`
		result, err := compileAndRun(script)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if result.Int() != 10 { // 0+1+2+3+4 = 10
			t.Errorf("expected 10, got %d", result.Int())
		}
		done <- true
	}()
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("test timed out after 5 seconds")
	}
}

// ============================================================================
// 函数调用测试
// ============================================================================

func TestVMFunctionCall(t *testing.T) {
	script := `
	fn add($a, $b) {
		return $a + $b;
	}
	$result = add(3, 4);
	$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 7 {
		t.Errorf("expected 7, got %d", result.Int())
	}
}

func TestVMFunctionMultipleCalls(t *testing.T) {
	script := `
	fn double($x) {
		return $x * 2;
	}
	$a = double(5);
	$b = double(10);
	$c = $a + $b;
	$c
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 30 { // 10 + 20
		t.Errorf("expected 30, got %d", result.Int())
	}
}

func TestVMFunctionNoReturn(t *testing.T) {
	script := `
	fn greet() {
		$msg = "hello";
	}
	greet();
	$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// msg 在 greet 函数作用域内，外部不可访问
	// 应该返回 null
	if !result.IsNull() {
		t.Errorf("expected null, got %v", result)
	}
}

func TestVMFunctionWithLocalVars(t *testing.T) {
	script := `
	fn compute($x, $y) {
		$sum = $x + $y;
		$product = $x * $y;
		return $sum + $product;
	}
	compute(2, 3)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (2+3) + (2*3) = 5 + 6 = 11
	if result.Int() != 11 {
		t.Errorf("expected 11, got %d", result.Int())
	}
}

// ============================================================================
// 递归测试
// ============================================================================

func TestVMRecursion(t *testing.T) {
	script := `
	fn factorial($n) {
		if ($n <= 1) {
			return 1;
		}
		return $n * factorial($n - 1);
	}
	factorial(5)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5! = 120
	if result.Int() != 120 {
		t.Errorf("expected 120, got %d", result.Int())
	}
}

func TestVMRecursionFibonacci(t *testing.T) {
	script := `
	fn fib($n) {
		if ($n <= 1) {
			return $n;
		}
		return fib($n - 1) + fib($n - 2);
	}
	fib(10)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// fib(10) = 55
	if result.Int() != 55 {
		t.Errorf("expected 55, got %d", result.Int())
	}
}

// ============================================================================
// 栈溢出保护测试
// ============================================================================

func TestVMStackOverflow(t *testing.T) {
	// Non-tail-recursive function that will overflow
	script := `
	fn infinite() {
		infinite();
		return 1;
	}
	infinite()
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	engine := NewEngine()
	defer engine.Close()

	vm := newVMWithProgram(engine, prog)
	vm.SetMaxCallDepth(100) // 设置较小的调用深度限制

	err = vm.Execute()
	if err == nil {
		t.Fatal("expected stack overflow error, got nil")
	}

	// 检查是否为栈溢出错误
	var runtimeErr *RuntimeError
	if !errors.As(err, &runtimeErr) {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}

	if runtimeErr.Message != "stack overflow: maximum call depth exceeded" {
		t.Errorf("expected stack overflow message, got %q", runtimeErr.Message)
	}
}

func TestVMStackOverflowDefault(t *testing.T) {
	// Non-tail-recursive function that will overflow
	script := `
	fn infinite() {
		infinite();
		return 1;
	}
	infinite()
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	engine := NewEngine()
	defer engine.Close()

	vm := newVMWithProgram(engine, prog)

	// 默认调用深度为 1000
	err = vm.Execute()
	if err == nil {
		t.Fatal("expected stack overflow error, got nil")
	}
}

func TestVMCallDepthGetterSetter(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	vm := newVM(engine)

	// 测试默认值
	if vm.GetMaxCallDepth() != defaultMaxCallDepth {
		t.Errorf("expected default max call depth %d, got %d", defaultMaxCallDepth, vm.GetMaxCallDepth())
	}

	// 测试设置
	vm.SetMaxCallDepth(500)
	if vm.GetMaxCallDepth() != 500 {
		t.Errorf("expected max call depth 500, got %d", vm.GetMaxCallDepth())
	}

	// 测试无效值
	vm.SetMaxCallDepth(-1)
	if vm.GetMaxCallDepth() != 500 {
		t.Errorf("expected max call depth to remain 500 after negative input, got %d", vm.GetMaxCallDepth())
	}

	// 测试当前调用深度
	if vm.GetCallDepth() != 0 {
		t.Errorf("expected call depth 0, got %d", vm.GetCallDepth())
	}
}

// ============================================================================
// 全局变量测试
// ============================================================================

func TestVMGlobalVariable(t *testing.T) {
	script := `$globalVar = 42; $globalVar`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMGlobalVariableInFunction(t *testing.T) {
	script := `
	$globalVar = 100;
	fn addGlobal($x) {
		return $globalVar + $x;
	}
	addGlobal(23)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 123 {
		t.Errorf("expected 123, got %d", result.Int())
	}
}

// ============================================================================
// 复杂表达式测试
// ============================================================================

func TestVMComplexExpression(t *testing.T) {
	script := `$a = (2 + 3) * 4 - 1; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (2+3)*4-1 = 5*4-1 = 20-1 = 19
	if result.Int() != 19 {
		t.Errorf("expected 19, got %d", result.Int())
	}
}

func TestVMTernary(t *testing.T) {
	script := `$a = 10 > 5 ? 1 : 0; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

// ============================================================================
// VM 生命周期测试
// ============================================================================

func TestVMClose(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	vm := newVM(engine)

	// 测试关闭
	err := vm.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 测试重复关闭
	err = vm.Close()
	if err == nil {
		t.Fatal("expected error on double close")
	}

	// 测试关闭后执行
	err = vm.Execute()
	if err == nil {
		t.Fatal("expected error on execute after close")
	}
}

func TestVMReset(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	vm := newVM(engine)
	vm.result = NewInt(42)
	vm.err = NewRuntimeError("test error")

	vm.Reset()

	if !vm.GetResult().IsNull() {
		t.Errorf("expected null result after reset, got %v", vm.GetResult())
	}
}

// ============================================================================
// 新 VM 实例测试
// ============================================================================

func TestNewVMWithProgram(t *testing.T) {
	script := `$a = 1 + 2; $a`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	engine := NewEngine()
	defer engine.Close()

	vm := newVMWithProgram(engine, prog)

	if vm.Engine() != engine {
		t.Error("vm engine mismatch")
	}

	if vm.program != prog {
		t.Error("vm program mismatch")
	}

	if len(vm.funcMap) != 0 {
		t.Errorf("expected empty func map for main-only program, got %d entries", len(vm.funcMap))
	}
}

func TestNewVMWithFunctions(t *testing.T) {
	script := `
	fn add($a, $b) { return $a + $b; }
	fn sub($a, $b) { return $a - $b; }
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	engine := NewEngine()
	defer engine.Close()

	vm := newVMWithProgram(engine, prog)

	// 应该有 2 个函数
	if len(vm.funcMap) != 2 {
		t.Errorf("expected 2 functions in func map, got %d", len(vm.funcMap))
	}

	// 检查函数名
	if _, ok := vm.funcMap["add"]; !ok {
		t.Error("expected 'add' function in func map")
	}
	if _, ok := vm.funcMap["sub"]; !ok {
		t.Error("expected 'sub' function in func map")
	}
}

// ============================================================================
// 错误处理测试
// ============================================================================

func TestVMNoProgram(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	vm := newVM(engine)
	err := vm.Execute()
	if err == nil {
		t.Fatal("expected error when no program")
	}
}

func TestVMDisassemble(t *testing.T) {
	script := `$a = 1 + 2; $a`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	engine := NewEngine()
	defer engine.Close()

	vm := newVMWithProgram(engine, prog)

	disassembly := vm.Disassemble()
	if disassembly == "" {
		t.Error("expected non-empty disassembly")
	}
}

// ============================================================================
// 类型测试
// ============================================================================

func TestVMTypeOf(t *testing.T) {
	script := `$a = typeof(42); $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "int" {
		t.Errorf("expected 'int', got %q", result.String())
	}
}

// ============================================================================
// 空值测试
// ============================================================================

func TestVMNull(t *testing.T) {
	script := `$a = null; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("expected null, got %v", result)
	}
}

// ============================================================================
// 布尔测试
// ============================================================================

func TestVMBool(t *testing.T) {
	tests := []struct {
		script   string
		expected bool
	}{
		{"$a = true; $a", true},
		{"$a = false; $a", false},
		{"$a = TRUE; $a", true},
		{"$a = FALSE; $a", false},
	}

	for _, tt := range tests {
		t.Run(tt.script, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result.Bool())
			}
		})
	}
}

// ============================================================================
// 浮点数测试
// ============================================================================

func TestVMFloat(t *testing.T) {
	script := `$a = 3.14; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float() != 3.14 {
		t.Errorf("expected 3.14, got %f", result.Float())
	}
}

func TestVMFloatArithmetic(t *testing.T) {
	script := `$a = 1.5 + 2.5; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float() != 4.0 {
		t.Errorf("expected 4.0, got %f", result.Float())
	}
}

// ============================================================================
// 混合类型测试
// ============================================================================

func TestVMStringConcat(t *testing.T) {
	script := `$a = "Hello" .. " " .. "World"; $a`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", result.String())
	}
}

// ============================================================================
// 嵌套函数调用测试
// ============================================================================

func TestVMNestedFunctionCalls(t *testing.T) {
	script := `
	fn double($x) { return $x * 2; }
	fn triple($x) { return $x * 3; }
	fn compute($a, $b) {
		return double($a) + triple($b);
	}
	compute(5, 10)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// double(5) + triple(10) = 10 + 30 = 40
	if result.Int() != 40 {
		t.Errorf("expected 40, got %d", result.Int())
	}
}

// ============================================================================
// 多次递归调用测试
// ============================================================================

func TestVMDeepRecursion(t *testing.T) {
	script := `
	fn countdown($n) {
		if ($n <= 0) {
			return 0;
		}
		return countdown($n - 1);
	}
	countdown(50)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("expected 0, got %d", result.Int())
	}
}

// ============================================================================
// 数组操作测试
// ============================================================================

func TestVMArraySetIndex(t *testing.T) {
	script := `$a = [1, 2, 3]; $a[1] = 20; $a[1]`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 20 {
		t.Errorf("expected 20, got %d", result.Int())
	}
}

// TestVMArraySetNegativeIndex 测试负数索引赋值
func TestVMArraySetNegativeIndex(t *testing.T) {
	script := `
		$a = [1, 2, 3, 4, 5];
		$a[-1] = 50;  // 修改最后一个元素
		$a[-2] = 40;  // 修改倒数第二个元素
		$a[-1] + $a[-2] + $a[0]
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 50 + 40 + 1 = 91
	if result.Int() != 91 {
		t.Errorf("expected 91 (50+40+1), got %d", result.Int())
	}
}

// ============================================================================
// 对象操作测试
// ============================================================================

func TestVMObjectSetMember(t *testing.T) {
	script := `$a = {"name": "old"}; $a.name = "new"; $a.name`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "new" {
		t.Errorf("expected 'new', got %q", result.String())
	}
}

// ============================================================================
// while 循环 break/continue 测试
// ============================================================================

func TestVMBreak(t *testing.T) {
	script := `$sum = 0; $i = 0; while (true) { if ($i >= 5) { break; } $sum = $sum + $i; $i = $i + 1; } $sum`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestVMContinue(t *testing.T) {
	script := `$sum = 0; $i = 0; while ($i < 5) { $i = $i + 1; if ($i == 3) { continue; } $sum = $sum + $i; } $sum`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// i=1: sum=1, i=2: sum=3, i=3: skip, i=4: sum=7, i=5: sum=12
	if result.Int() != 12 {
		t.Errorf("expected 12, got %d", result.Int())
	}
}

// ============================================================================
// _ 前缀块级私有作用域测试
// ============================================================================

func TestVMUnderscoreLocalScope(t *testing.T) {
	// _ 前缀变量在块内声明，块外不可访问
	script := `
	$result = 0;
	{
		_local = 42;
		$result = _local;
	}
	$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMUnderscoreScopeIsolation(t *testing.T) {
	// _ 前缀变量在外部作用域不可访问
	script := `
	{
		_secret = 100;
	}
	$x = 1;
	$x
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

func TestVMUnderscoreInFunction(t *testing.T) {
	// 函数内 _ 前缀变量仅在函数作用域内有效
	script := `
	fn compute() {
		_tmp = 99;
		return _tmp;
	}
	compute()
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 99 {
		t.Errorf("expected 99, got %d", result.Int())
	}
}

func TestVMUnderscoreInFunctionNotLeaked(t *testing.T) {
	// 函数内 _ 前缀变量不应泄漏到调用者作用域
	script := `
	fn init() {
		_val = 50;
	}
	init();
	$val = 10;
	$val
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestVMDollarUnderscoreNormalScope(t *testing.T) {
	// $name 遵循普通作用域链，if 块内可访问外层变量
	script := `
	$base = 50;
	$result = 0;
	if ($base > 0) {
		$result = $base + 27;
	}
	$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 77 {
		t.Errorf("expected 77, got %d", result.Int())
	}
}

func TestVMNestedUnderscoreIndependent(t *testing.T) {
	// 嵌套块中同名 _ 前缀变量互不干扰
	script := `
	$sum = 0;
	{
		_x = 10;
		{
			_x = 20;
			$sum = $sum + _x;
		}
		$sum = $sum + _x;
	}
	$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 内层 _x=20, 外层 _x=10, sum = 20 + 10 = 30
	if result.Int() != 30 {
		t.Errorf("expected 30, got %d", result.Int())
	}
}

// ============================================================================
// 动态变量访问（OP_GETVAR / OP_SETVAR）
// =============================================================================

func TestVMGetVarGlobal(t *testing.T) {
	// 测试 OP_GETVAR 读取全局变量
	prog := &Program{
		Main: &CompiledFunction{
			Name:      "<main>",
			Params:    0,
			Registers: 3,
			VarNames:  []string{"$x", "$name", "$result"},
			Constants: []Value{NewInt(42), NewString("$x")},
			Bytecode: []Instruction{
				NewABx(OP_LOADK, 0, 0),     // R0 = 42
				NewABx(OP_SETGLOBAL, 0, 1), // Globals["$x"] = R0
				NewABx(OP_LOADK, 1, 1),     // R1 = "$x"
				NewABC(OP_GETVAR, 2, 1, 0), // R2 = lookup(R1)
				NewABC(OP_RETURN, 2, 0, 0), // return R2
			},
		},
		Constants: []Value{NewInt(42), NewString("$x")},
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err := vm.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := vm.GetResult()
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMSetVarGlobal(t *testing.T) {
	// 测试 OP_SETVAR 写入全局变量
	prog := &Program{
		Main: &CompiledFunction{
			Name:      "<main>",
			Params:    0,
			Registers: 3,
			VarNames:  []string{"$name", "$val", "$result"},
			Constants: []Value{NewString("$x"), NewInt(100)},
			Bytecode: []Instruction{
				NewABx(OP_LOADK, 0, 0),     // R0 = "$x"
				NewABx(OP_LOADK, 1, 1),     // R1 = 100
				NewABC(OP_SETVAR, 0, 0, 1), // assign(R0, R1)
				NewABx(OP_GETGLOBAL, 2, 0), // R2 = Globals["$x"]
				NewABC(OP_RETURN, 2, 0, 0), // return R2
			},
		},
		Constants: []Value{NewString("$x"), NewInt(100)},
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err := vm.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := vm.GetResult()
	if result.Int() != 100 {
		t.Errorf("expected 100, got %d", result.Int())
	}
}

func TestVMGetVarLocal(t *testing.T) {
	// 测试 OP_GETVAR 读取局部变量
	prog := &Program{
		Main: &CompiledFunction{
			Name:      "<main>",
			Params:    0,
			Registers: 3,
			VarNames:  []string{"$x", "$name", "$result"},
			Constants: []Value{NewInt(42), NewString("$x")},
			Bytecode: []Instruction{
				NewABx(OP_LOADK, 0, 0),     // R0 = 42
				NewABx(OP_LOADK, 1, 1),     // R1 = "$x"
				NewABC(OP_GETVAR, 2, 1, 0), // R2 = lookup(R1) -> should find R0
				NewABC(OP_RETURN, 2, 0, 0), // return R2
			},
		},
		Constants: []Value{NewInt(42), NewString("$x")},
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err := vm.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := vm.GetResult()
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMSetVarLocal(t *testing.T) {
	// 测试 OP_SETVAR 写入局部变量
	prog := &Program{
		Main: &CompiledFunction{
			Name:      "<main>",
			Params:    0,
			Registers: 3,
			VarNames:  []string{"$x", "$name", "$result"},
			Constants: []Value{NewInt(42), NewString("$x"), NewInt(99)},
			Bytecode: []Instruction{
				NewABx(OP_LOADK, 0, 0),     // R0 = 42
				NewABx(OP_LOADK, 1, 1),     // R1 = "$x"
				NewABx(OP_LOADK, 2, 2),     // R2 = 99
				NewABC(OP_SETVAR, 0, 1, 2), // assign(R1, R2) -> R0 = 99
				NewABC(OP_RETURN, 0, 0, 0), // return R0
			},
		},
		Constants: []Value{NewInt(42), NewString("$x"), NewInt(99)},
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err := vm.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := vm.GetResult()
	if result.Int() != 99 {
		t.Errorf("expected 99, got %d", result.Int())
	}
}

// ============================================================================
// 反射 API（ListFunctions / GetFunctionInfo）
// =============================================================================

func TestVMListFunctions(t *testing.T) {
	script := `
		fn add($a, $b) {
			return $a + $b;
		}
		fn greet($name) {
			return "hello " + $name;
		}
		$x = 1;
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	funcs := vm.ListFunctions()
	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(funcs))
	}

	// 检查函数名（顺序不确定）
	hasAdd := false
	hasGreet := false
	for _, name := range funcs {
		switch name {
		case "add":
			hasAdd = true
		case "greet":
			hasGreet = true
		}
	}
	if !hasAdd {
		t.Error("missing function 'add'")
	}
	if !hasGreet {
		t.Error("missing function 'greet'")
	}
}

func TestVMGetFunctionInfo(t *testing.T) {
	script := `
		fn add($a, $b) {
			return $a + $b;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	// 测试存在的函数
	infos, ok := vm.GetFunctionInfo("add")
	if !ok {
		t.Fatal("function 'add' not found")
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 overload, got %d", len(infos))
	}
	info := infos[0]
	if info.Name != "add" {
		t.Errorf("expected name 'add', got '%s'", info.Name)
	}
	if info.ParamCount != 2 {
		t.Errorf("expected 2 params, got %d", info.ParamCount)
	}
	if len(info.ParamNames) != 2 {
		t.Fatalf("expected 2 param names, got %d", len(info.ParamNames))
	}
	if info.ParamNames[0] != "$a" {
		t.Errorf("expected first param '$a', got '%s'", info.ParamNames[0])
	}
	if info.ParamNames[1] != "$b" {
		t.Errorf("expected second param '$b', got '%s'", info.ParamNames[1])
	}

	// 测试不存在的函数
	_, ok = vm.GetFunctionInfo("nonexistent")
	if ok {
		t.Error("expected false for nonexistent function")
	}
}

// ============================================================================
// CallByName 动态调用
// =============================================================================

func TestVMCallByNameSimple(t *testing.T) {
	script := `
		fn add($a, $b) {
			return $a + $b;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	result, err := vm.CallByName("add", NewInt(10), NewInt(20))
	if err != nil {
		t.Fatalf("CallByName error: %v", err)
	}
	if result.Int() != 30 {
		t.Errorf("expected 30, got %d", result.Int())
	}
}

func TestVMCallByNameStringConcat(t *testing.T) {
	script := `
		fn greet($first, $last) {
			return "Hello, " + $first + " " + $last;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	result, err := vm.CallByName("greet", NewString("John"), NewString("Doe"))
	if err != nil {
		t.Fatalf("CallByName error: %v", err)
	}
	expected := "Hello, John Doe"
	if result.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.String())
	}
}

func TestVMCallByNameNotFound(t *testing.T) {
	script := `
		fn test() {
			return 1;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	_, err = vm.CallByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent function")
	}
}

func TestVMCallByNameRecursive(t *testing.T) {
	script := `
		fn factorial($n) {
			if ($n <= 1) {
				return 1;
			}
			return $n * factorial($n - 1);
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	result, err := vm.CallByName("factorial", NewInt(5))
	if err != nil {
		t.Fatalf("CallByName error: %v", err)
	}
	if result.Int() != 120 {
		t.Errorf("expected 120, got %d", result.Int())
	}
}

func TestVMCallByNamePreservesState(t *testing.T) {
	script := `
		fn double($x) {
			return $x * 2;
		}
		$global = 100;
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	// 调用前设置全局变量
	vm.SetGlobal("$global", NewInt(100))

	// 调用函数
	result, err := vm.CallByName("double", NewInt(5))
	if err != nil {
		t.Fatalf("CallByName error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}

	// 验证全局变量未被破坏
	val, ok := vm.GetGlobal("$global")
	if !ok || val.Int() != 100 {
		t.Error("global variable was corrupted")
	}
}

// ============================================================================
// Lambda 和箭头函数
// =============================================================================

func TestVMLambdaSimple(t *testing.T) {
	script := `
		$add = fn($a, $b) {
			return $a + $b;
		};
		$result = $add(10, 20);
	`

	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err = vm.Execute()
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	// 获取 $result 变量
	result, ok := vm.GetGlobal("$result")
	if !ok {
		t.Fatal("$result not found in globals")
	}

	if result.Int() != 30 {
		t.Errorf("expected 30, got %d", result.Int())
	}
}

func TestVMArrowFunction(t *testing.T) {
	script := `
		$double = $x -> $x * 2;
		$result = $double(5);
	`

	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err = vm.Execute()
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	result, ok := vm.GetGlobal("$result")
	if !ok {
		t.Fatal("$result not found in globals")
	}

	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestVMLambdaAsArgument(t *testing.T) {
	script := `
		fn apply($fn, $x) {
			return $fn($x);
		}
		$result = apply($n -> $n * 3, 7);
	`

	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err = vm.Execute()
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}

	result, ok := vm.GetGlobal("$result")
	if !ok {
		t.Fatal("$result not found in globals")
	}

	if result.Int() != 21 {
		t.Errorf("expected 21, got %d", result.Int())
	}
}

// ============================================================================
// 异常处理测试（try/catch/throw）
// =============================================================================

func TestVMTryCatchBasic(t *testing.T) {
	script := `
		$result = 0;
		try {
			throw "error";
			$result = 1;
		} catch ($e) {
			$result = 42;
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMTryCatchNoThrow(t *testing.T) {
	script := `
		$result = 0;
		try {
			$result = 10;
		} catch ($e) {
			$result = 99;
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestVMTryCatchErrorMessage(t *testing.T) {
	script := `
		$errMsg = "";
		try {
			throw "something went wrong";
		} catch ($e) {
			$errMsg = $e;
		}
		$errMsg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", result.String())
	}
}

func TestVMTryCatchNested(t *testing.T) {
	script := `
		$result = 0;
		try {
			try {
				throw "inner";
			} catch ($inner) {
				$result = $result + 1;
				throw "outer";
			}
		} catch ($outer) {
			$result = $result + 10;
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// inner catch: result = 1, then throw outer
	// outer catch: result = 1 + 10 = 11
	if result.Int() != 11 {
		t.Errorf("expected 11, got %d", result.Int())
	}
}

func TestVMTryCatchInFunction(t *testing.T) {
	script := `
		fn safe($x) {
			try {
				if ($x < 0) {
					throw "negative";
				}
				return $x * 2;
			} catch ($e) {
				return -1;
			}
		}

		$a = safe(5);
		$b = safe(-1);
		$a + $b
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// safe(5) = 10, safe(-1) = -1, total = 9
	if result.Int() != 9 {
		t.Errorf("expected 9, got %d", result.Int())
	}
}

func TestVMTryCatchNoCatchHandler(t *testing.T) {
	script := `
		throw "unhandled error";
	`
	_, err := compileAndRun(script)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVMThrowInt(t *testing.T) {
	script := `
		$result = 0;
		try {
			throw 42;
		} catch ($e) {
			$result = $e;
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestVMThrowObject(t *testing.T) {
	script := `
		$result = null;
		try {
			throw {"code": 500, "msg": "server error"};
		} catch ($e) {
			$result = $e;
		}
		$result.code
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 500 {
		t.Errorf("expected 500, got %d", result.Int())
	}
}

func TestVMTryCatchSkipRest(t *testing.T) {
	script := `
		$before = 0;
		$after = 0;
		try {
			$before = 1;
			throw "error";
			$before = 2;
		} catch ($e) {
			$after = 1;
		}
		$before + $after
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// before = 1 (set before throw), after = 1 (catch executed)
	// throw skips the rest of try block, so before stays 1
	if result.Int() != 2 {
		t.Errorf("expected 2, got %d", result.Int())
	}
}

// ============================================================================
// Error 类型 + 条件捕获测试
// ============================================================================

// error() 基本创建
func TestVMErrorBasic(t *testing.T) {
	script := `
		$e = error("something failed");
		$e.message
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "something failed" {
		t.Errorf("expected 'something failed', got %q", result.String())
	}
}

// error() 带 code 和 type
func TestVMErrorWithCodeAndType(t *testing.T) {
	script := `
		$e = error("division by zero", 1001, "MathError");
		$e.code .. $e.type
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "1001MathError" {
		t.Errorf("expected '1001MathError', got %q", result.String())
	}
}

// error() 默认值
func TestVMErrorDefaults(t *testing.T) {
	script := `
		$e = error("oops");
		$e.code .. " " .. $e.type
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "0 Error" {
		t.Errorf("expected '0 Error', got %q", result.String())
	}
}

// throw error() + 字段访问
func TestVMThrowError(t *testing.T) {
	script := `
		$msg = "";
		$code = 0;
		try {
			throw error("not found", 404, "HttpError");
		} catch ($e) {
			$msg = $e.message;
			$code = $e.code;
		}
		$msg .. ":" .. $code
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "not found:404" {
		t.Errorf("expected 'not found:404', got %q", result.String())
	}
}

// 条件捕获 — 匹配
func TestVMCatchConditionMatch(t *testing.T) {
	script := `
		$result = 0;
		try {
			throw error("server error", 500, "HttpError");
		} catch ($e when $e.code == 500) {
			$result = $e.code;
		}
		$result
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 500 {
		t.Errorf("expected 500, got %d", result.Int())
	}
}

// 条件捕获 — 不匹配 re-throw 到外层
func TestVMCatchConditionNoMatch(t *testing.T) {
	script := `
		$result = 0;
		try {
			try {
				throw error("not found", 404, "HttpError");
			} catch ($e when $e.code == 500) {
				$result = 999;
			}
		} catch ($e) {
			$result = $e.code;
		}
		$result
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 404 {
		t.Errorf("expected 404, got %d", result.Int())
	}
}

// 条件捕获 — 多个条件分支
func TestVMCatchConditionMultiple(t *testing.T) {
	script := `
		$result = "";
		try {
			throw error("timeout", 408, "HttpError");
		} catch ($e when $e.code == 404) {
			$result = "not found";
		}
		$result
	`
	_, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	// 条件不匹配，re-throw 无外层 handler，变成 RuntimeError
	if err == nil {
		t.Errorf("expected runtime error for uncaught conditional catch")
	}
}

// 条件捕获 — type 字段匹配
func TestVMCatchConditionByType(t *testing.T) {
	script := `
		$result = "";
		try {
			throw error("db error", 1, "DBError");
		} catch ($e when $e.type == "DBError") {
			$result = "caught DB error";
		}
		$result
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "caught DB error" {
		t.Errorf("expected 'caught DB error', got %q", result.String())
	}
}

// typeof 返回 "error"
func TestVMErrorTypeOf(t *testing.T) {
	script := `
		$e = error("test");
		typeof($e)
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("error", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, errors.New("error() expects at least 1 argument")
			}
			msg := args[0].String()
			var code int64
			var errType string
			if len(args) >= 2 {
				code = args[1].Int()
			}
			if len(args) >= 3 {
				errType = args[2].String()
			}
			return NewError(msg, code, errType), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "error" {
		t.Errorf("expected 'error', got %q", result.String())
	}
}

// ============================================================================
// 作用域链测试
// =============================================================================

// 1. 同一函数内块级作用域访问外层变量
func TestScopeChainBlockAccessOuter(t *testing.T) {
	script := `
		$outer = 10;
		$result = 0;
		{
			$result = $outer + 5;
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 15 {
		t.Errorf("expected 15, got %d", result.Int())
	}
}

// 2. 嵌套块级作用域链
func TestScopeChainNestedBlocks(t *testing.T) {
	script := `
		$a = 1;
		{
			$b = $a + 1;
			{
				$c = $b + 1;
				{
					$d = $c + 1;
				}
			}
		}
		$d
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=1, b=2, c=3, d=4
	if result.Int() != 4 {
		t.Errorf("expected 4, got %d", result.Int())
	}
}

// 3. 函数访问全局变量
func TestScopeChainFuncAccessGlobal(t *testing.T) {
	script := `
		$global = 100;
		fn getGlobal() {
			return $global;
		}
		getGlobal()
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 100 {
		t.Errorf("expected 100, got %d", result.Int())
	}
}

// 4. 函数修改全局变量
// 注意：当前实现中，函数通过 SETGLOBAL 修改全局 map，
// 但主作用域的本地变量不会自动同步。如需读取函数修改后的值，
// 应使用独立变量接收，或在函数内直接返回修改后的值。
func TestScopeChainFuncModifyGlobal(t *testing.T) {
	script := `
		$counter = 0;
		fn increment() {
			$counter = $counter + 1;
			return $counter;
		}
		$a = increment();
		$b = increment();
		$c = increment();
		$a + $b + $c
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=1, b=2, c=3, sum=6
	if result.Int() != 6 {
		t.Errorf("expected 6, got %d", result.Int())
	}
}

// 5. 嵌套函数访问外层函数参数
func TestScopeChainNestedFuncAccessOuterParam(t *testing.T) {
	script := `
		fn outer($x) {
			fn inner() {
				return $x + 10;
			}
			return inner();
		}
		outer(5)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 15 {
		t.Errorf("expected 15, got %d", result.Int())
	}
}

// 6. 嵌套函数访问外层函数局部变量
func TestScopeChainNestedFuncAccessOuterLocal(t *testing.T) {
	script := `
		fn outer() {
			$local = 42;
			fn inner() {
				return $local;
			}
			return inner();
		}
		outer()
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

// 7. 闭包捕获变量
func TestScopeChainClosureCapture(t *testing.T) {
	script := `
		fn makeAdder($n) {
			return fn($x) {
				return $x + $n;
			};
		}
		$add5 = makeAdder(5);
		$add10 = makeAdder(10);
		$add5(3) + $add10(3)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// add5(3) = 8, add10(3) = 13, sum = 21
	if result.Int() != 21 {
		t.Errorf("expected 21, got %d", result.Int())
	}
}

// 8. 闭包修改捕获的变量（计数器模式）
func TestScopeChainClosureModify(t *testing.T) {
	script := `
		fn makeCounter() {
			$count = 0;
			return fn() {
				$count = $count + 1;
				return $count;
			};
		}
		$counter = makeCounter();
		$a = $counter();
		$b = $counter();
		$c = $counter();
		$a + $b + $c
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=1, b=2, c=3, sum=6
	if result.Int() != 6 {
		t.Errorf("expected 6, got %d", result.Int())
	}
}

// 9. 三层嵌套函数作用域
func TestScopeChainThreeLevelNested(t *testing.T) {
	script := `
		fn level1($a) {
			fn level2($b) {
				fn level3($c) {
					return $a + $b + $c;
				}
				return level3(30);
			}
			return level2(20);
		}
		level1(10)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("expected 60, got %d", result.Int())
	}
}

// 10. 函数内 if 块访问函数参数
func TestScopeChainIfBlockAccessParam(t *testing.T) {
	script := `
		fn check($value) {
			$result = 0;
			if ($value > 0) {
				$result = $value * 2;
			} else {
				$result = $value * -1;
			}
			return $result;
		}
		$a = check(5);
		$b = check(-3);
		$a + $b
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=10, b=3, sum=13
	if result.Int() != 13 {
		t.Errorf("expected 13, got %d", result.Int())
	}
}

// 11. 循环内访问外层变量
func TestScopeChainLoopAccessOuter(t *testing.T) {
	script := `
		$multiplier = 10;
		$sum = 0;
		$i = 0;
		while ($i < 5) {
			$sum = $sum + $i * $multiplier;
			$i = $i + 1;
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0*10 + 1*10 + 2*10 + 3*10 + 4*10 = 100
	if result.Int() != 100 {
		t.Errorf("expected 100, got %d", result.Int())
	}
}

// 12. Lambda 表达式访问外层变量
func TestScopeChainLambdaAccessOuter(t *testing.T) {
	script := `
		$base = 100;
		$addBase = $x -> $x + $base;
		$addBase(25)
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 125 {
		t.Errorf("expected 125, got %d", result.Int())
	}
}

// ============================================================================
// 函数重载测试（Phase 3.3）
// ============================================================================

// TestOverloadBasic 测试基本的参数数量重载
func TestOverloadBasic(t *testing.T) {
	script := `
		fn greet() {
			return "hello";
		}
		fn greet($name) {
			return "hello " + $name;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	e := NewEngine()
	defer e.Close()
	vm := NewVMWithProgram(e, prog)

	// 调用无参数版本
	result, err := vm.CallByName("greet")
	if err != nil {
		t.Fatalf("call greet() error: %v", err)
	}
	if result.String() != "hello" {
		t.Errorf("greet() expected 'hello', got '%s'", result.String())
	}

	// 调用单参数版本
	result, err = vm.CallByName("greet", NewString("world"))
	if err != nil {
		t.Fatalf("call greet(name) error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("greet('world') expected 'hello world', got '%s'", result.String())
	}
}

// TestOverloadThreeParams 测试三个重载版本
func TestOverloadThreeParams(t *testing.T) {
	script := `
		fn add($a) {
			return $a;
		}
		fn add($a, $b) {
			return $a + $b;
		}
		fn add($a, $b, $c) {
			return $a + $b + $c;
		}
	`
	result, err := compileAndRun(script + `add(10)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("add(10) expected 10, got %d", result.Int())
	}

	result, err = compileAndRun(script + `add(10, 20)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 30 {
		t.Errorf("add(10, 20) expected 30, got %d", result.Int())
	}

	result, err = compileAndRun(script + `add(10, 20, 30)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("add(10, 20, 30) expected 60, got %d", result.Int())
	}
}

// TestOverloadFallback 测试参数数量不匹配时的容错
func TestOverloadFallback(t *testing.T) {
	script := `
		fn identity($x) {
			return $x;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	e := NewEngine()
	defer e.Close()
	vm := NewVMWithProgram(e, prog)

	// 多传参数 — 多余参数被忽略
	result, err := vm.CallByName("identity", NewString("world"), NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "world" {
		t.Errorf("expected 'world', got '%s'", result.String())
	}

	// 少传参数 — 缺少的参数为 null
	result, err = vm.CallByName("identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("expected null, got %v", result)
	}
}

// TestOverloadGetFunctionInfo 测试重载函数的反射 API
func TestOverloadGetFunctionInfo(t *testing.T) {
	script := `
		fn calc($x) {
			return $x;
		}
		fn calc($x, $y) {
			return $x + $y;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	infos, ok := vm.GetFunctionInfo("calc")
	if !ok {
		t.Fatal("function 'calc' not found")
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 overloads, got %d", len(infos))
	}

	// 验证两个重载版本
	found1 := false
	found2 := false
	for _, info := range infos {
		if info.ParamCount == 1 {
			found1 = true
		}
		if info.ParamCount == 2 {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Error("expected overloads with 1 and 2 params")
	}
}

// TestOverloadRecursive 测试重载函数递归调用
func TestOverloadRecursive(t *testing.T) {
	script := `
		fn sum($n) {
			return $n;
		}
		fn sum($n, $rest) {
			return $n + $rest;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	e := NewEngine()
	defer e.Close()
	vm := NewVMWithProgram(e, prog)

	result, err := vm.CallByName("sum", NewInt(10))
	if err != nil {
		t.Fatalf("sum(10) error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("sum(10) expected 10, got %d", result.Int())
	}

	result, err = vm.CallByName("sum", NewInt(10), NewInt(5))
	if err != nil {
		t.Fatalf("sum(10, 5) error: %v", err)
	}
	if result.Int() != 15 {
		t.Errorf("sum(10, 5) expected 15, got %d", result.Int())
	}
}

// TestOverloadInlineCall 测试通过 CallByName 调用重载函数
func TestOverloadInlineCall(t *testing.T) {
	script := `
		fn format() {
			return "empty";
		}
		fn format($val) {
			return "value: " + $val;
		}
		fn format($key, $val) {
			return $key + "=" + $val;
		}
	`
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	e := NewEngine()
	defer e.Close()
	vm := NewVMWithProgram(e, prog)

	// 0 参数版本
	r0, err := vm.CallByName("format")
	if err != nil {
		t.Fatalf("format() error: %v", err)
	}
	if r0.String() != "empty" {
		t.Errorf("format() expected 'empty', got '%s'", r0.String())
	}

	// 1 参数版本
	r1, err := vm.CallByName("format", NewString("hello"))
	if err != nil {
		t.Fatalf("format(val) error: %v", err)
	}
	if r1.String() != "value: hello" {
		t.Errorf("format('hello') expected 'value: hello', got '%s'", r1.String())
	}

	// 2 参数版本
	r2, err := vm.CallByName("format", NewString("name"), NewString("jpl"))
	if err != nil {
		t.Fatalf("format(key, val) error: %v", err)
	}
	if r2.String() != "name=jpl" {
		t.Errorf("format('name', 'jpl') expected 'name=jpl', got '%s'", r2.String())
	}
}

// ============================================================================
// logError 集成测试
// ============================================================================

func TestLogErrorOnRuntimeError(t *testing.T) {
	// "hello" + 5 类型不匹配 → runtimeError，应被 logError 记录
	engine := NewEngine()
	defer engine.Close()

	prog, err := CompileString(`"hello" + 5`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	_ = vm.Execute() // 执行可能因 runtimeError 返回 null

	// 检查错误日志
	logs := engine.GetErrorLog()
	if len(logs) == 0 {
		t.Fatal("expected at least one error in log, got none")
	}
	if logs[0].Error() != "runtime error: cannot add string and int" {
		t.Errorf("unexpected error: %v", logs[0])
	}
}

func TestLogErrorOnArithmeticTypeMismatch(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	prog, err := CompileString(`[1,2] * 3`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	_ = vm.Execute()

	logs := engine.GetErrorLog()
	if len(logs) == 0 {
		t.Fatal("expected error logged for array * int")
	}
}

func TestLogErrorMultiple(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	// 两个连续的类型错误
	prog, err := CompileString(`"a" + 1; [1] - 2`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	_ = vm.Execute()

	logs := engine.GetErrorLog()
	if len(logs) < 2 {
		t.Errorf("expected 2 errors, got %d", len(logs))
	}
}

func TestLogErrorClear(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	prog, err := CompileString(`"a" + 1`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	_ = vm.Execute()

	if len(engine.GetErrorLog()) == 0 {
		t.Fatal("expected errors before clear")
	}

	engine.ClearErrorLog()
	if len(engine.GetErrorLog()) != 0 {
		t.Error("expected empty log after clear")
	}
}

func TestLogErrorGetLastError(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	prog, err := CompileString(`"a" + 1; [1] - 2`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	_ = vm.Execute()

	last := engine.GetLastError()
	if last == nil {
		t.Fatal("expected non-nil last error")
	}
	if last.Error() != "runtime error: cannot subtract from array" {
		t.Errorf("unexpected last error: %v", last)
	}
}

func TestNoLogErrorOnValidOperation(t *testing.T) {
	engine := NewEngine()
	defer engine.Close()

	prog, err := CompileString(`1 + 2`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logs := engine.GetErrorLog()
	if len(logs) != 0 {
		t.Errorf("expected no errors for valid operation, got %d", len(logs))
	}
}

// ============================================================================
// 错误消息源码上下文测试
// ============================================================================

func TestRuntimeErrorSourceContext(t *testing.T) {
	script := `fn test() {
    $x = 10
    throw "boom"
}
test()`

	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err = vm.Execute()
	if err == nil {
		t.Fatal("expected runtime error")
	}

	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T", err)
	}

	// 验证行号被正确设置
	if re.Line == 0 {
		t.Error("expected line number to be set")
	}

	// 验证源码上下文格式化
	formatted := re.FormatWithContext(prog.SourceLines)
	if !strings.Contains(formatted, "boom") {
		t.Errorf("expected formatted output to contain error message, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "throw") {
		t.Errorf("expected formatted output to contain source line, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "→") {
		t.Errorf("expected formatted output to contain arrow marker, got:\n%s", formatted)
	}
}

func TestRuntimeErrorSourceContextMultiLine(t *testing.T) {
	script := `fn greet() {
    $msg = "hello"
    throw "something went wrong"
}

greet()`

	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := newVM(nil)
	vm.program = prog
	vm.buildFuncMap()

	err = vm.Execute()
	if err == nil {
		t.Fatal("expected runtime error")
	}

	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T", err)
	}

	// 行号应该是 3（throw 所在行）
	if re.Line != 3 {
		t.Errorf("expected line 3, got %d", re.Line)
	}

	formatted := re.FormatWithContext(prog.SourceLines)

	// 应该包含错误行前后的上下文
	if !strings.Contains(formatted, "fn greet()") {
		t.Errorf("expected context to show function definition, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "something went wrong") {
		t.Errorf("expected context to show error message, got:\n%s", formatted)
	}
}

func TestRuntimeErrorFormatFallback(t *testing.T) {
	// 测试没有源码时的回退格式
	re := NewRuntimeError("test error")
	re.Line = 5

	// 没有 sourceLines 时应回退到简单格式
	formatted := re.FormatWithContext(nil)
	if !strings.Contains(formatted, "test error") {
		t.Errorf("expected simple format, got: %s", formatted)
	}

	// 行号为 0 时也应回退
	re2 := NewRuntimeError("no line")
	formatted2 := re2.FormatWithContext([]string{"line1", "line2"})
	if !strings.Contains(formatted2, "no line") {
		t.Errorf("expected simple format for zero line, got: %s", formatted2)
	}
}

// ============================================================================
// 间接变量引用测试（Phase 20）
// ============================================================================

func TestVMIndirectRef(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "basic indirect ref",
			script:   `a = "hello"; x = "a"; ` + "`x",
			expected: "hello",
		},
		{
			name:     "indirect ref with $-prefixed variable",
			script:   `$a = "world"; x = "$a"; ` + "`x",
			expected: "world",
		},
		{
			name:     "chained indirect ref",
			script:   `a = "hello"; b = "a"; x = "b"; ` + "`x",
			expected: "a",
		},
		{
			name:     "indirect ref with integer value",
			script:   `a = 42; x = "a"; ` + "`x",
			expected: "42",
		},
		{
			name:     "indirect ref in expression",
			script:   `a = 10; x = "a"; y = ` + "`x; y + 5",
			expected: "15",
		},
		{
			name:     "indirect ref undefined variable",
			script:   `x = "nonexistent"; ` + "`x",
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := compileAndRun(tt.script)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := result.String()
			if result.Type() == TypeString {
				got = result.String()
			} else {
				got = result.Stringify()
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
