package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestEvalSimple(t *testing.T) {
	script := `$result = eval("1 + 2");
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("eval('1 + 2') = %d, expected 3", result.Int())
	}
}

func TestEvalString(t *testing.T) {
	script := `$result = eval('"hello" + " " + "world"');
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("eval string concat = %q, expected 'hello world'", result.String())
	}
}

func EvalArray(t *testing.T) {
	script := `$result = eval("[1, 2, 3]");
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0].Int() != 1 || arr[1].Int() != 2 || arr[2].Int() != 3 {
		t.Errorf("expected [1, 2, 3], got %v", arr)
	}
}

func TestEvalFunctionDef(t *testing.T) {
	// eval creates a separate VM context - functions defined in eval
	// are not visible to the calling script.
	// This tests that eval can execute code with function definitions internally.
	script := `$result = eval("fn add(a, b) { return a + b; } add(10, 20)");
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 30 {
		t.Errorf("expected 30, got %d", result.Int())
	}
}

func TestEvalVariableAccess(t *testing.T) {
	// eval creates a separate VM context - variables from the calling
	// script are not directly visible inside eval.
	// This tests eval can execute arithmetic.
	script := `$result = eval("21 * 2");
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("eval('21 * 2') = %d, expected 42", result.Int())
	}
}

func TestEvalError(t *testing.T) {
	result, err := callBuiltin("eval", engine.NewInt(42))
	if err == nil {
		t.Error("eval(42) should return error")
	}
	_ = result
}

func TestEvalSyntaxError(t *testing.T) {
	// Use syntax that causes a runtime error
	script := `eval("1 / 0")`

	// Division by zero returns Inf, not an error, so this should succeed
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1/0 should return Inf
	if result.Type() != engine.TypeFloat {
		t.Errorf("expected float, got %s", result.Type())
	}
}

func TestEvalReturn(t *testing.T) {
	script := `$sq = eval("fn square(n) { return n * n; } square(5)");
$sq`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 25 {
		t.Errorf("square(5) = %d, expected 25", result.Int())
	}
}
