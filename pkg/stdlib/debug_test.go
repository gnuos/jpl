package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// compileAndRunWithEngine 编译脚本并执行，返回结果、引擎和错误
func compileAndRunWithEngine(script string) (engine.Value, *engine.Engine, error) {
	prog, err := engine.CompileString(script)
	if err != nil {
		return nil, nil, err
	}
	e := engine.NewEngine()
	RegisterAll(e)
	vm := engine.NewVMWithProgram(e, prog)
	err = vm.Execute()
	return vm.GetResult(), e, err
}

// ============================================================================
// errors() 测试
// ============================================================================

func TestErrorsEmpty(t *testing.T) {
	result, err := callBuiltin("errors")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr := result.Array()
	if len(arr) != 0 {
		t.Errorf("errors() on clean engine should return empty array, got %d", len(arr))
	}
}

func TestErrorsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("errors", engine.NewInt(1))
	if err == nil {
		t.Error("errors(1) should return error")
	}
}

func TestErrorsAfterTypeError(t *testing.T) {
	script := `"hello" + 5; errors()`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr := result.Array()
	if len(arr) == 0 {
		t.Fatal("expected at least one error")
	}
	if arr[0].String() != "runtime error: cannot add string and int" {
		t.Errorf("unexpected error message: %s", arr[0].String())
	}
}

func TestErrorsMultiple(t *testing.T) {
	script := `"a" + 1; [1] - 2; errors()`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr := result.Array()
	if len(arr) < 2 {
		t.Errorf("expected 2 errors, got %d", len(arr))
	}
}

// ============================================================================
// last_error() 测试
// ============================================================================

func TestLastErrorEmpty(t *testing.T) {
	result, err := callBuiltin("last_error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("last_error() on clean engine should return null, got %v", result)
	}
}

func TestLastErrorWrongArgCount(t *testing.T) {
	_, err := callBuiltin("last_error", engine.NewInt(1))
	if err == nil {
		t.Error("last_error(1) should return error")
	}
}

func TestLastErrorAfterTypeError(t *testing.T) {
	script := `"hello" + 5; last_error()`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "runtime error: cannot add string and int" {
		t.Errorf("unexpected last error: %s", result.String())
	}
}

func TestLastErrorReturnsLast(t *testing.T) {
	script := `"a" + 1; [1] - 2; last_error()`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "runtime error: cannot subtract from array" {
		t.Errorf("expected last error about array, got: %s", result.String())
	}
}

// ============================================================================
// clear_errors() 测试
// ============================================================================

func TestClearErrorsEmpty(t *testing.T) {
	result, err := callBuiltin("clear_errors")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("clear_errors() should return null, got %v", result)
	}
}

func TestClearErrorsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("clear_errors", engine.NewInt(1))
	if err == nil {
		t.Error("clear_errors(1) should return error")
	}
}

func TestClearErrorsEffect(t *testing.T) {
	script := `"hello" + 5; clear_errors(); errors()`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr := result.Array()
	if len(arr) != 0 {
		t.Errorf("errors() after clear should be empty, got %d", len(arr))
	}
}

// ============================================================================
// 集成测试 — 脚本中使用调试函数
// ============================================================================

func TestDebugFunctionsInScript(t *testing.T) {
	script := `"hello" + 5;
$err_count = len(errors());
$err_count`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("expected 1 error, got %d", result.Int())
	}
}

func TestDebugWorkflow(t *testing.T) {
	// 模拟调试工作流：产生错误 → 查看 → 清空 → 确认清空
	script := `"a" + 1;
$b = last_error();
clear_errors();
$empty_len = len(errors());
$empty_len`
	result, e, err := compileAndRunWithEngine(script)
	defer e.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("expected 0 errors after clear, got %d", result.Int())
	}
}
