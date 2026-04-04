package stdlib

import (
	"math"
	"math/big"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// abs
// ============================================================================

func TestAbs(t *testing.T) {
	tests := []struct {
		input engine.Value
		want  int64
	}{
		{engine.NewInt(5), 5},
		{engine.NewInt(-5), 5},
		{engine.NewInt(0), 0},
	}
	for _, tt := range tests {
		result, err := callBuiltin("abs", tt.input)
		if err != nil {
			t.Fatalf("abs error: %v", err)
		}
		if result.Int() != tt.want {
			t.Errorf("abs(%v) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

func TestAbsFloat(t *testing.T) {
	result, err := callBuiltin("abs", engine.NewFloat(-3.14))
	if err != nil {
		t.Fatalf("abs error: %v", err)
	}
	if result.Float() != 3.14 {
		t.Errorf("abs(-3.14) = %f, want 3.14", result.Float())
	}
}

// ============================================================================
// ceil / floor
// ============================================================================

func TestCeil(t *testing.T) {
	tests := []struct {
		input float64
		want  int64
	}{
		{3.14, 4},
		{3.0, 3},
		{-2.7, -2},
		{0.1, 1},
	}
	for _, tt := range tests {
		result, err := callBuiltin("ceil", engine.NewFloat(tt.input))
		if err != nil {
			t.Fatalf("ceil error: %v", err)
		}
		if result.Int() != tt.want {
			t.Errorf("ceil(%f) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

func TestCeilInt(t *testing.T) {
	result, err := callBuiltin("ceil", engine.NewInt(5))
	if err != nil {
		t.Fatalf("ceil error: %v", err)
	}
	if result.Int() != 5 {
		t.Errorf("ceil(5) = %d, want 5", result.Int())
	}
}

func TestFloor(t *testing.T) {
	tests := []struct {
		input float64
		want  int64
	}{
		{3.14, 3},
		{3.99, 3},
		{-2.7, -3},
		{0.9, 0},
	}
	for _, tt := range tests {
		result, err := callBuiltin("floor", engine.NewFloat(tt.input))
		if err != nil {
			t.Fatalf("floor error: %v", err)
		}
		if result.Int() != tt.want {
			t.Errorf("floor(%f) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

// ============================================================================
// round
// ============================================================================

func TestRound(t *testing.T) {
	tests := []struct {
		input float64
		want  int64
	}{
		{3.5, 4},
		{3.4, 3},
		{-2.5, -3},
		{-2.6, -3},
	}
	for _, tt := range tests {
		result, err := callBuiltin("round", engine.NewFloat(tt.input))
		if err != nil {
			t.Fatalf("round error: %v", err)
		}
		if result.Int() != tt.want {
			t.Errorf("round(%f) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

func TestRoundPrecision(t *testing.T) {
	result, err := callBuiltin("round", engine.NewFloat(3.14159), engine.NewInt(2))
	if err != nil {
		t.Fatalf("round error: %v", err)
	}
	if result.Float() != 3.14 {
		t.Errorf("round(3.14159, 2) = %f, want 3.14", result.Float())
	}
}

func TestRoundInt(t *testing.T) {
	result, err := callBuiltin("round", engine.NewInt(5))
	if err != nil {
		t.Fatalf("round error: %v", err)
	}
	if result.Int() != 5 {
		t.Errorf("round(5) = %d, want 5", result.Int())
	}
}

// ============================================================================
// pow / sqrt
// ============================================================================

func TestPow(t *testing.T) {
	result, err := callBuiltin("pow", engine.NewInt(2), engine.NewInt(10))
	if err != nil {
		t.Fatalf("pow error: %v", err)
	}
	if result.Float() != 1024.0 {
		t.Errorf("pow(2, 10) = %f, want 1024", result.Float())
	}
}

func TestPowFloat(t *testing.T) {
	result, err := callBuiltin("pow", engine.NewFloat(2.0), engine.NewFloat(0.5))
	if err != nil {
		t.Fatalf("pow error: %v", err)
	}
	if math.Abs(result.Float()-1.4142135623730951) > 0.0001 {
		t.Errorf("pow(2.0, 0.5) = %f, want ~1.414", result.Float())
	}
}

func TestSqrt(t *testing.T) {
	result, err := callBuiltin("sqrt", engine.NewFloat(9.0))
	if err != nil {
		t.Fatalf("sqrt error: %v", err)
	}
	if result.Float() != 3.0 {
		t.Errorf("sqrt(9) = %f, want 3", result.Float())
	}
}

func TestSqrtInt(t *testing.T) {
	result, err := callBuiltin("sqrt", engine.NewInt(16))
	if err != nil {
		t.Fatalf("sqrt error: %v", err)
	}
	if result.Float() != 4.0 {
		t.Errorf("sqrt(16) = %f, want 4", result.Float())
	}
}

// ============================================================================
// min / max
// ============================================================================

func TestMin(t *testing.T) {
	result, err := callBuiltin("min", engine.NewInt(3), engine.NewInt(1), engine.NewInt(2))
	if err != nil {
		t.Fatalf("min error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("min(3,1,2) = %d, want 1", result.Int())
	}
}

func TestMinSingle(t *testing.T) {
	result, err := callBuiltin("min", engine.NewInt(42))
	if err != nil {
		t.Fatalf("min error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("min(42) = %d, want 42", result.Int())
	}
}

func TestMax(t *testing.T) {
	result, err := callBuiltin("max", engine.NewInt(3), engine.NewInt(1), engine.NewInt(2))
	if err != nil {
		t.Fatalf("max error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("max(3,1,2) = %d, want 3", result.Int())
	}
}

func TestMaxFloat(t *testing.T) {
	result, err := callBuiltin("max", engine.NewFloat(1.5), engine.NewFloat(2.5), engine.NewFloat(1.9))
	if err != nil {
		t.Fatalf("max error: %v", err)
	}
	if result.Float() != 2.5 {
		t.Errorf("max(1.5,2.5,1.9) = %f, want 2.5", result.Float())
	}
}

// ============================================================================
// random / randomInt
// ============================================================================

func TestRandom(t *testing.T) {
	result, err := callBuiltin("random")
	if err != nil {
		t.Fatalf("random error: %v", err)
	}
	f := result.Float()
	if f < 0 || f >= 1 {
		t.Errorf("random() = %f, want [0, 1)", f)
	}
}

func TestRandomInt(t *testing.T) {
	for range 100 {
		result, err := callBuiltin("randomInt", engine.NewInt(1), engine.NewInt(10))
		if err != nil {
			t.Fatalf("randomInt error: %v", err)
		}
		n := result.Int()
		if n < 1 || n > 10 {
			t.Errorf("randomInt(1,10) = %d, want [1, 10]", n)
		}
	}
}

func TestRandomIntSame(t *testing.T) {
	result, err := callBuiltin("randomInt", engine.NewInt(5), engine.NewInt(5))
	if err != nil {
		t.Fatalf("randomInt error: %v", err)
	}
	if result.Int() != 5 {
		t.Errorf("randomInt(5,5) = %d, want 5", result.Int())
	}
}

func TestRandomIntMinMaxError(t *testing.T) {
	_, err := callBuiltin("randomInt", engine.NewInt(10), engine.NewInt(1))
	if err == nil {
		t.Error("randomInt(10,1) should return error")
	}
}

// ============================================================================
// parseInt / parseFloat
// ============================================================================

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"42", 42},
		{"-10", -10},
		{"0", 0},
		{"0xff", 255},
	}
	for _, tt := range tests {
		result, err := callBuiltin("parseInt", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("parseInt error: %v", err)
		}
		if result.Int() != tt.want {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

func TestParseIntInvalid(t *testing.T) {
	result, err := callBuiltin("parseInt", engine.NewString("abc"))
	if err != nil {
		t.Fatalf("parseInt error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("parseInt('abc') should return null, got %v", result)
	}
}

func TestParseIntFromFloat(t *testing.T) {
	result, err := callBuiltin("parseInt", engine.NewFloat(3.7))
	if err != nil {
		t.Fatalf("parseInt error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("parseInt(3.7) = %d, want 3", result.Int())
	}
}

func TestParseFloat(t *testing.T) {
	result, err := callBuiltin("parseFloat", engine.NewString("3.14"))
	if err != nil {
		t.Fatalf("parseFloat error: %v", err)
	}
	if result.Float() != 3.14 {
		t.Errorf("parseFloat('3.14') = %f, want 3.14", result.Float())
	}
}

func TestParseFloatInvalid(t *testing.T) {
	result, err := callBuiltin("parseFloat", engine.NewString("abc"))
	if err != nil {
		t.Fatalf("parseFloat error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("parseFloat('abc') should return null")
	}
}

func TestParseFloatFromInt(t *testing.T) {
	result, err := callBuiltin("parseFloat", engine.NewInt(42))
	if err != nil {
		t.Fatalf("parseFloat error: %v", err)
	}
	if result.Float() != 42.0 {
		t.Errorf("parseFloat(42) = %f, want 42.0", result.Float())
	}
}

// ============================================================================
// isNaN / isFinite
// ============================================================================

func TestIsNaN(t *testing.T) {
	nan := engine.NewFloat(math.NaN())
	result, err := callBuiltin("isNaN", nan)
	if err != nil {
		t.Fatalf("isNaN error: %v", err)
	}
	if !result.Bool() {
		t.Error("isNaN(NaN) should be true")
	}

	result, err = callBuiltin("isNaN", engine.NewFloat(1.0))
	if err != nil {
		t.Fatalf("isNaN error: %v", err)
	}
	if result.Bool() {
		t.Error("isNaN(1.0) should be false")
	}

	result, err = callBuiltin("isNaN", engine.NewInt(42))
	if err != nil {
		t.Fatalf("isNaN error: %v", err)
	}
	if result.Bool() {
		t.Error("isNaN(42) should be false")
	}
}

func TestIsFinite(t *testing.T) {
	result, err := callBuiltin("isFinite", engine.NewFloat(1.0))
	if err != nil {
		t.Fatalf("isFinite error: %v", err)
	}
	if !result.Bool() {
		t.Error("isFinite(1.0) should be true")
	}

	result, err = callBuiltin("isFinite", engine.NewFloat(math.Inf(1)))
	if err != nil {
		t.Fatalf("isFinite error: %v", err)
	}
	if result.Bool() {
		t.Error("isFinite(+Inf) should be false")
	}

	result, err = callBuiltin("isFinite", engine.NewFloat(math.NaN()))
	if err != nil {
		t.Fatalf("isFinite error: %v", err)
	}
	if result.Bool() {
		t.Error("isFinite(NaN) should be false")
	}

	result, err = callBuiltin("isFinite", engine.NewInt(42))
	if err != nil {
		t.Fatalf("isFinite error: %v", err)
	}
	if !result.Bool() {
		t.Error("isFinite(42) should be true")
	}

	result, err = callBuiltin("isFinite", engine.NewString("abc"))
	if err != nil {
		t.Fatalf("isFinite error: %v", err)
	}
	if result.Bool() {
		t.Error("isFinite('abc') should be false")
	}
}

// ============================================================================
// BigInt/BigDecimal 支持测试
// ============================================================================

func TestAbsBigInt(t *testing.T) {
	neg := engine.NewBigInt(big.NewInt(-42))
	result, err := callBuiltin("abs", neg)
	if err != nil {
		t.Fatalf("abs error: %v", err)
	}
	bi, ok := result.(*engine.BigIntValue)
	if !ok {
		t.Fatal("abs(BigInt) should return BigInt")
	}
	if bi.BigInt().Cmp(big.NewInt(42)) != 0 {
		t.Errorf("abs(bigint(-42)) = %s, want 42", bi.BigInt().String())
	}
}

func TestAbsBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("-3.14")
	neg := engine.NewBigDecimal(rat)
	result, err := callBuiltin("abs", neg)
	if err != nil {
		t.Fatalf("abs error: %v", err)
	}
	bd, ok := result.(*engine.BigDecimalValue)
	if !ok {
		t.Fatal("abs(BigDecimal) should return BigDecimal")
	}
	expected, _ := new(big.Rat).SetString("3.14")
	if bd.BigRat().Cmp(expected) != 0 {
		t.Errorf("abs(BigDecimal(-3.14)) = %s, want 3.14", bd.BigRat().FloatString(2))
	}
}

func TestPowBigInt(t *testing.T) {
	base := engine.NewBigInt(big.NewInt(2))
	result, err := callBuiltin("pow", base, engine.NewInt(100))
	if err != nil {
		t.Fatalf("pow error: %v", err)
	}
	bi, ok := result.(*engine.BigIntValue)
	if !ok {
		t.Fatal("pow(BigInt, int) should return BigInt")
	}
	// 2^100 = 1267650600228229401496703205376
	expected := new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil)
	if bi.BigInt().Cmp(expected) != 0 {
		t.Errorf("pow(BigInt(2), 100) = %s, want %s", bi.BigInt().String(), expected.String())
	}
}

func TestPowBigIntNegativeExp(t *testing.T) {
	base := engine.NewBigInt(big.NewInt(2))
	result, err := callBuiltin("pow", base, engine.NewInt(-1))
	if err != nil {
		t.Fatalf("pow error: %v", err)
	}
	// 负指数返回 float
	if result.Type() != engine.TypeFloat {
		t.Errorf("pow(BigInt, -1) should return float, got %s", result.Type())
	}
	if result.Float() != 0.5 {
		t.Errorf("pow(BigInt(2), -1) = %f, want 0.5", result.Float())
	}
}

func TestPowBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("1.5")
	base := engine.NewBigDecimal(rat)
	result, err := callBuiltin("pow", base, engine.NewInt(3))
	if err != nil {
		t.Fatalf("pow error: %v", err)
	}
	bd, ok := result.(*engine.BigDecimalValue)
	if !ok {
		t.Fatal("pow(BigDecimal, int) should return BigDecimal")
	}
	expected, _ := new(big.Rat).SetString("3.375") // 1.5^3 = 3.375
	if bd.BigRat().Cmp(expected) != 0 {
		t.Errorf("pow(BigDecimal(1.5), 3) = %s, want 3.375", bd.BigRat().FloatString(3))
	}
}

func TestCeilBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("3.14")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("ceil", bd)
	if err != nil {
		t.Fatalf("ceil error: %v", err)
	}
	bi, ok := result.(*engine.BigIntValue)
	if !ok {
		t.Fatal("ceil(BigDecimal) should return BigInt")
	}
	if bi.BigInt().Cmp(big.NewInt(4)) != 0 {
		t.Errorf("ceil(BigDecimal(3.14)) = %s, want 4", bi.BigInt().String())
	}
}

func TestCeilBigDecimalNegative(t *testing.T) {
	rat, _ := new(big.Rat).SetString("-3.14")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("ceil", bd)
	if err != nil {
		t.Fatalf("ceil error: %v", err)
	}
	bi := result.(*engine.BigIntValue)
	// ceil(-3.14) = -3（向正无穷取整）
	if bi.BigInt().Cmp(big.NewInt(-3)) != 0 {
		t.Errorf("ceil(BigDecimal(-3.14)) = %s, want -3", bi.BigInt().String())
	}
}

func TestFloorBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("3.99")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("floor", bd)
	if err != nil {
		t.Fatalf("floor error: %v", err)
	}
	bi, ok := result.(*engine.BigIntValue)
	if !ok {
		t.Fatal("floor(BigDecimal) should return BigInt")
	}
	if bi.BigInt().Cmp(big.NewInt(3)) != 0 {
		t.Errorf("floor(BigDecimal(3.99)) = %s, want 3", bi.BigInt().String())
	}
}

func TestFloorBigDecimalNegative(t *testing.T) {
	rat, _ := new(big.Rat).SetString("-3.14")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("floor", bd)
	if err != nil {
		t.Fatalf("floor error: %v", err)
	}
	bi := result.(*engine.BigIntValue)
	// floor(-3.14) = -4（向负无穷取整）
	if bi.BigInt().Cmp(big.NewInt(-4)) != 0 {
		t.Errorf("floor(BigDecimal(-3.14)) = %s, want -4", bi.BigInt().String())
	}
}

func TestRoundBigDecimal(t *testing.T) {
	rat, _ := new(big.Rat).SetString("3.14159")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("round", bd, engine.NewInt(2))
	if err != nil {
		t.Fatalf("round error: %v", err)
	}
	if result.Float() != 3.14 {
		t.Errorf("round(BigDecimal(3.14159), 2) = %f, want 3.14", result.Float())
	}
}

func TestRoundBigDecimalZero(t *testing.T) {
	rat, _ := new(big.Rat).SetString("3.7")
	bd := engine.NewBigDecimal(rat)
	result, err := callBuiltin("round", bd)
	if err != nil {
		t.Fatalf("round error: %v", err)
	}
	bi, ok := result.(*engine.BigIntValue)
	if !ok {
		t.Fatal("round(BigDecimal(3.7)) should return BigInt")
	}
	if bi.BigInt().Cmp(big.NewInt(4)) != 0 {
		t.Errorf("round(BigDecimal(3.7)) = %s, want 4", bi.BigInt().String())
	}
}
