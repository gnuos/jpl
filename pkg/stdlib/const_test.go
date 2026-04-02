package stdlib

import (
	"math"
	"math/big"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// define 函数测试
// ============================================================================

func TestDefineBasic(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	fn := e.GetRegisteredFunc("define")
	ctx := engine.NewContext(e, nil)

	result, err := fn(ctx, []engine.Value{engine.NewString("MY_PI"), engine.NewFloat(3.14159)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("define should return null, got %v", result)
	}

	// 验证常量已存储
	v, ok := e.GetConst("MY_PI")
	if !ok {
		t.Fatal("constant MY_PI should exist")
	}
	if v.Float() != 3.14159 {
		t.Errorf("MY_PI should be 3.14159, got %v", v)
	}
}

func TestDefineDuplicate(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	fn := e.GetRegisteredFunc("define")
	ctx := engine.NewContext(e, nil)

	_, err := fn(ctx, []engine.Value{engine.NewString("X"), engine.NewInt(1)})
	if err != nil {
		t.Fatalf("first define should succeed: %v", err)
	}

	_, err = fn(ctx, []engine.Value{engine.NewString("X"), engine.NewInt(2)})
	if err == nil {
		t.Error("duplicate define should return error")
	}
}

func TestDefineWrongArgCount(t *testing.T) {
	_, err := callBuiltin("define", engine.NewString("X"))
	if err == nil {
		t.Error("define(1 arg) should return error")
	}

	_, err = callBuiltin("define", engine.NewString("X"), engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("define(3 args) should return error")
	}
}

func TestDefineNameNotString(t *testing.T) {
	_, err := callBuiltin("define", engine.NewInt(42), engine.NewInt(1))
	if err == nil {
		t.Error("define(42, 1) should return error")
	}
}

func TestDefineEmptyName(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	fn := e.GetRegisteredFunc("define")
	ctx := engine.NewContext(e, nil)

	_, err := fn(ctx, []engine.Value{engine.NewString(""), engine.NewInt(1)})
	if err == nil {
		t.Error("define('', value) should return error")
	}
}

func TestDefineVariousTypes(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	fn := e.GetRegisteredFunc("define")
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		name  string
		value engine.Value
	}{
		{"int_const", engine.NewInt(42)},
		{"float_const", engine.NewFloat(3.14)},
		{"str_const", engine.NewString("hello")},
		{"bool_const", engine.NewBool(true)},
		{"null_const", engine.NewNull()},
		{"array_const", engine.NewArray([]engine.Value{engine.NewInt(1)})},
		{"obj_const", engine.NewObject(map[string]engine.Value{"a": engine.NewInt(1)})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn(ctx, []engine.Value{engine.NewString(tt.name), tt.value})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsNull() {
				t.Errorf("define should return null, got %v", result)
			}
		})
	}
}

// ============================================================================
// defined 函数测试
// ============================================================================

func TestDefinedTrue(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	ctx := engine.NewContext(e, nil)

	defineFn := e.GetRegisteredFunc("define")
	definedFn := e.GetRegisteredFunc("defined")

	_, err := defineFn(ctx, []engine.Value{engine.NewString("MY_CONST"), engine.NewInt(99)})
	if err != nil {
		t.Fatalf("define failed: %v", err)
	}

	result, err := definedFn(ctx, []engine.Value{engine.NewString("MY_CONST")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("defined('MY_CONST') should be true")
	}
}

func TestDefinedFalse(t *testing.T) {
	result, err := callBuiltin("defined", engine.NewString("NONEXISTENT"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("defined('NONEXISTENT') should be false")
	}
}

func TestDefinedWrongArgCount(t *testing.T) {
	_, err := callBuiltin("defined")
	if err == nil {
		t.Error("defined(0 args) should return error")
	}

	_, err = callBuiltin("defined", engine.NewString("X"), engine.NewString("Y"))
	if err == nil {
		t.Error("defined(2 args) should return error")
	}
}

func TestDefinedArgNotString(t *testing.T) {
	_, err := callBuiltin("defined", engine.NewInt(42))
	if err == nil {
		t.Error("defined(42) should return error")
	}
}

// ============================================================================
// 集成测试 — 通过脚本执行
// ============================================================================

func TestDefineIntegration(t *testing.T) {
	script := `define("MAX", 100);
defined("MAX")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("defined('MAX') should be true after define")
	}
}

func TestDefinedBeforeDefine(t *testing.T) {
	script := `defined("NOPE")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("defined('NOPE') should be false")
	}
}

func TestDefineDuplicateIntegration(t *testing.T) {
	script := `define("DUP", 1);
define("DUP", 2)`

	_, err := compileAndRunBuiltins(script)
	if err == nil {
		t.Error("duplicate define in script should return error")
	}
}

func TestDefineNoParen(t *testing.T) {
	script := `define "MY_PI", 3.14159;
defined("MY_PI")`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("defined('PI') should be true after define without parens")
	}
}

func TestDefineReturnsNull(t *testing.T) {
	script := `define "X", 42`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("define should return null, got %v", result)
	}
}

// ============================================================================
// 预设常量测试
// ============================================================================

func TestPresetConstantINF(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("INF")
	if !ok {
		t.Fatal("preset constant INF should exist")
	}
	if v.Type() != engine.TypeFloat {
		t.Errorf("INF should be float, got %s", v.Type())
	}
	if !math.IsInf(v.Float(), 1) {
		t.Errorf("INF should be +Inf, got %v", v.Float())
	}
}

func TestPresetConstantNaN(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("NaN")
	if !ok {
		t.Fatal("preset constant NaN should exist")
	}
	if !math.IsNaN(v.Float()) {
		t.Errorf("NaN should be NaN, got %v", v.Float())
	}
}

func TestPresetConstantPI(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("PI")
	if !ok {
		t.Fatal("preset constant PI should exist")
	}
	if v.Float() != math.Pi {
		t.Errorf("PI should be %v, got %v", math.Pi, v.Float())
	}
}

func TestPresetConstantE(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("E")
	if !ok {
		t.Fatal("preset constant E should exist")
	}
	if v.Float() != math.E {
		t.Errorf("E should be %v, got %v", math.E, v.Float())
	}
}

func TestPresetConstantTAU(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("TAU")
	if !ok {
		t.Fatal("preset constant TAU should exist")
	}
	if v.Float() != 2*math.Pi {
		t.Errorf("TAU should be %v, got %v", 2*math.Pi, v.Float())
	}
}

func TestPresetConstantSQRT2(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("SQRT2")
	if !ok {
		t.Fatal("preset constant SQRT2 should exist")
	}
	if v.Float() != math.Sqrt2 {
		t.Errorf("SQRT2 should be %v, got %v", math.Sqrt2, v.Float())
	}
}

func TestPresetConstantLN2(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("LN2")
	if !ok {
		t.Fatal("preset constant LN2 should exist")
	}
	if v.Float() != math.Ln2 {
		t.Errorf("LN2 should be %v, got %v", math.Ln2, v.Float())
	}
}

func TestPresetConstantLN10(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	v, ok := e.GetConst("LN10")
	if !ok {
		t.Fatal("preset constant LN10 should exist")
	}
	if v.Float() != math.Ln10 {
		t.Errorf("LN10 should be %v, got %v", math.Ln10, v.Float())
	}
}

func TestPresetConstantDefined(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	ctx := engine.NewContext(e, nil)
	fn := e.GetRegisteredFunc("defined")

	for _, name := range []string{"INF", "NaN", "PI", "TAU", "E", "SQRT2", "LN2", "LN10"} {
		result, err := fn(ctx, []engine.Value{engine.NewString(name)})
		if err != nil {
			t.Fatalf("defined(%q) error: %v", name, err)
		}
		if !result.Bool() {
			t.Errorf("defined(%q) should be true", name)
		}
	}
}

func TestPresetConstantCannotRedefine(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)
	ctx := engine.NewContext(e, nil)
	fn := e.GetRegisteredFunc("define")

	_, err := fn(ctx, []engine.Value{engine.NewString("PI"), engine.NewFloat(9.9)})
	if err == nil {
		t.Error("redefining PI should return error")
	}
}

// ============================================================================
// 预设常量脚本集成测试
// ============================================================================

func TestPresetConstantInScript(t *testing.T) {
	script := `PI`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float() != math.Pi {
		t.Errorf("script PI = %v, want %v", result.Float(), math.Pi)
	}
}

func TestPresetConstantArithmetic(t *testing.T) {
	script := `PI * 2`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Float() != math.Pi*2 {
		t.Errorf("PI * 2 = %v, want %v", result.Float(), math.Pi*2)
	}
}

// ============================================================================
// IEEE 754 除零测试
// ============================================================================

func TestDivByZeroIntOverInt(t *testing.T) {
	v := engine.NewInt(5).Div(engine.NewInt(0))
	if v.Type() != engine.TypeFloat {
		t.Errorf("5 / 0 should return float, got %s", v.Type())
	}
	if !math.IsInf(v.Float(), 1) {
		t.Errorf("5 / 0 should be +Inf, got %v", v.Float())
	}
}

func TestDivByZeroNegativeOverInt(t *testing.T) {
	v := engine.NewInt(-5).Div(engine.NewInt(0))
	if !math.IsInf(v.Float(), -1) {
		t.Errorf("-5 / 0 should be -Inf, got %v", v.Float())
	}
}

func TestDivByZeroZeroOverZero(t *testing.T) {
	v := engine.NewInt(0).Div(engine.NewInt(0))
	if !math.IsNaN(v.Float()) {
		t.Errorf("0 / 0 should be NaN, got %v", v.Float())
	}
}

func TestDivByZeroFloatOverFloat(t *testing.T) {
	v := engine.NewFloat(3.14).Div(engine.NewFloat(0.0))
	if !math.IsInf(v.Float(), 1) {
		t.Errorf("3.14 / 0.0 should be +Inf, got %v", v.Float())
	}
}

func TestDivByZeroNegativeFloat(t *testing.T) {
	v := engine.NewFloat(-2.5).Div(engine.NewFloat(0.0))
	if !math.IsInf(v.Float(), -1) {
		t.Errorf("-2.5 / 0.0 should be -Inf, got %v", v.Float())
	}
}

func TestModByZeroReturnsNaN(t *testing.T) {
	v := engine.NewInt(5).Mod(engine.NewInt(0))
	if !math.IsNaN(v.Float()) {
		t.Errorf("5 %% 0 should be NaN, got %v", v.Float())
	}
}

func TestModByZeroFloatReturnsNaN(t *testing.T) {
	v := engine.NewFloat(3.14).Mod(engine.NewFloat(0.0))
	if !math.IsNaN(v.Float()) {
		t.Errorf("3.14 %% 0.0 should be NaN, got %v", v.Float())
	}
}

func TestDivByZeroBigInt(t *testing.T) {
	v := engine.NewBigInt(big.NewInt(100)).Div(engine.NewInt(0))
	if !math.IsInf(v.Float(), 1) {
		t.Errorf("BigInt(100) / 0 should be +Inf, got %v", v.Float())
	}
}

func TestDivByZeroBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("3.14")
	v := engine.NewBigDecimal(rat).Div(engine.NewBigDecimal(new(big.Rat)))
	if !math.IsInf(v.Float(), 1) {
		t.Errorf("BigDecimal(3.14) / BigDecimal(0) should be +Inf, got %v", v.Float())
	}
}

func TestModByZeroBigInt(t *testing.T) {
	v := engine.NewBigInt(big.NewInt(100)).Mod(engine.NewInt(0))
	if !math.IsNaN(v.Float()) {
		t.Errorf("BigInt(100) %% 0 should be NaN, got %v", v.Float())
	}
}
