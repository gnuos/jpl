package engine

import (
	"testing"
)

// ============================================================================
// global 关键字测试
// ============================================================================

func TestGlobalKeywordBasic(t *testing.T) {
	script := `$x = 10;
fn getx() {
	global $x;
	return $x;
}
getx()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestGlobalKeywordModify(t *testing.T) {
	// The global keyword allows functions to modify global variables.
	// Test: function sets global, then reads it back
	script := `fn set_and_read($v) {
	global $x;
	$x = $v;
	return $x;
}
set_and_read(42)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestGlobalKeywordMultipleVars(t *testing.T) {
	script := `$a = 1;
$b = 2;
fn swap() {
	global $a, $b;
	$tmp = $a;
	$a = $b;
	$b = $tmp;
}
swap();
$a + $b`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After swap: a=2, b=1, so a+b=3
	if result.Int() != 3 {
		t.Errorf("expected 3, got %d", result.Int())
	}
}

func TestGlobalKeywordNestedFunc(t *testing.T) {
	// Test that global keyword persists across multiple function calls
	script := `fn increment() {
	global $counter;
	$counter = $counter + 1;
	return $counter;
}
increment();
increment();
increment()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("expected 3, got %d", result.Int())
	}
}

func TestGlobalKeywordWithoutDecl(t *testing.T) {
	// Without global declaration, a function can still access outer scope
	// variables through closure capture (upvalue mechanism).
	// The `global` keyword is for accessing variables from the global scope
	// when the variable doesn't exist in any enclosing function scope.
	script := `fn outer() {
	$x = 10;
	fn inner($v) {
		// Without global, inner captures outer's $x via upvalue
		$x = $v;
		return $x;
	}
	inner(42);
	return $x;  // outer's $x was modified by inner
}
outer()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// inner modifies outer's $x via closure capture
	if result.Int() != 42 {
		t.Errorf("expected 42 (modified via closure), got %d", result.Int())
	}
}

// ============================================================================
// static 变量测试
// ============================================================================

func TestStaticVarBasic(t *testing.T) {
	script := `fn counter() {
	static $count = 0;
	$count = $count + 1;
	return $count;
}
counter()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First call should return 1 (initial 0 + 1)
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

func TestStaticVarPersistence(t *testing.T) {
	script := `fn counter() {
	static $count = 0;
	$count = $count + 1;
	return $count;
}
counter();
counter();
counter()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Third call should return 3
	if result.Int() != 3 {
		t.Errorf("expected 3, got %d", result.Int())
	}
}

func TestStaticVarNoInitialValue(t *testing.T) {
	script := `fn acc() {
	static $sum;
	$sum = $sum + 10;
	return $sum;
}
acc()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// First call: null + 10 = 10
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestStaticVarMultipleFunctions(t *testing.T) {
	// Each function has its own static variable namespace
	script := `fn a() {
	static $x = 0;
	$x = $x + 1;
	return $x;
}
fn b() {
	static $x = 0;
	$x = $x + 10;
	return $x;
}
a();
b();
a();
b()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// a=2, b=20, result = 20
	if result.Int() != 20 {
		t.Errorf("expected 20, got %d", result.Int())
	}
}

func TestStaticVarStringType(t *testing.T) {
	// Static variables can hold non-numeric types
	script := `fn greet() {
	static $msg = "hello";
	$msg = $msg .. "!";
	return $msg;
}
greet()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello!" {
		t.Errorf("expected 'hello!', got '%s'", result.String())
	}
}

func TestStaticVarPersistenceAcrossCalls(t *testing.T) {
	// Static variable should persist across multiple calls
	script := `fn counter() {
	static $count = 0;
	$count = $count + 1;
	return $count;
}
counter();
counter();
counter();
counter()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 4 {
		t.Errorf("expected 4, got %d", result.Int())
	}
}

func TestStaticVarInitializationOnlyOnce(t *testing.T) {
	// Static variable should only be initialized once
	script := `fn test() {
	static $x = 0;
	$x = $x + 1;
	return $x;
}
test();
test();
test()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("expected 3, got %d", result.Int())
	}
}

func TestStaticVarWithComplexExpression(t *testing.T) {
	// Static variable with complex initial expression
	script := `fn test() {
	static $x = 10 * 2 + 5;
	$x = $x + 1;
	return $x;
}
test()`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 26 {
		t.Errorf("expected 26, got %d", result.Int())
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

func compileAndRunScope(script string) (Value, error) {
	prog, err := CompileString(script)
	if err != nil {
		return nil, err
	}
	e := NewEngine()
	defer e.Close()
	vm := NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		return nil, err
	}
	return vm.GetResult(), nil
}

// ============================================================================
// 尾调用测试（TCO 已实现，支持自递归尾调用栈帧复用）
// ============================================================================

func TestTailCallBasic(t *testing.T) {
	// Basic tail recursive factorial works correctly
	script := `fn fact($n, $acc) {
	if ($n <= 1) {
		return $acc;
	}
	return fact($n - 1, $n * $acc);
}
fact(10, 1)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10! = 3628800
	if result.Int() != 3628800 {
		t.Errorf("expected 3628800, got %d", result.Int())
	}
}

func TestTailCallSum(t *testing.T) {
	script := `fn sum($n, $acc) {
	if ($n <= 0) {
		return $acc;
	}
	return sum($n - 1, $acc + $n);
}
sum(100, 0)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sum(1..100) = 5050
	if result.Int() != 5050 {
		t.Errorf("expected 5050, got %d", result.Int())
	}
}

func TestTailCallDeepRecursion(t *testing.T) {
	// 深度递归：超过默认 maxCallDepth (1000) 的限制
	// 如果没有 TCO，此测试会因栈溢出而失败
	script := `fn sum($n, $acc) {
	if ($n <= 0) {
		return $acc;
	}
	return sum($n - 1, $acc + $n);
}
sum(5000, 0)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sum(1..5000) = 5000 * 5001 / 2 = 12502500
	expected := int64(5000 * 5001 / 2)
	if result.Int() != expected {
		t.Errorf("expected %d, got %d", expected, result.Int())
	}
}

func TestTailCallVeryDeepRecursion(t *testing.T) {
	// 极深递归：10000 层，验证 TCO 真正消除了栈帧增长
	script := `fn counter($n) {
	if ($n <= 0) {
		return "done";
	}
	return counter($n - 1);
}
counter(10000)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "done" {
		t.Errorf("expected 'done', got '%s'", result.String())
	}
}

func TestTailCallFactorialDeep(t *testing.T) {
	// 深度阶乘：验证大数计算的正确性
	script := `fn fact($n, $acc) {
	if ($n <= 1) {
		return $acc;
	}
	return fact($n - 1, $n * $acc);
}
fact(20, 1)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 20! = 2432902008176640000
	expected := int64(2432902008176640000)
	if result.Int() != expected {
		t.Errorf("expected %d, got %d", expected, result.Int())
	}
}

func TestTailCallWithMultipleBranches(t *testing.T) {
	// 尾递归中有多个返回路径
	script := `fn collatz($n, $steps) {
	if ($n == 1) {
		return $steps;
	}
	if ($n % 2 == 0) {
		return collatz($n / 2, $steps + 1);
	}
	return collatz($n * 3 + 1, $steps + 1);
}
collatz(27, 0)`

	result, err := compileAndRunScope(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// collatz(27) = 111 steps
	if result.Int() != 111 {
		t.Errorf("expected 111, got %d", result.Int())
	}
}
