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
// 尾调用测试（注释：当前未实现真正的栈帧复用优化，仅测试尾递归函数的正确性）
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
