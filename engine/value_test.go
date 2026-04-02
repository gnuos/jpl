package engine

import (
	"math"
	"math/big"
	"testing"
)

// =============================================================================
// 基础类型创建和访问
// =============================================================================

func TestNewNull(t *testing.T) {
	v := NewNull()
	if v.Type() != TypeNull {
		t.Errorf("expected TypeNull, got %v", v.Type())
	}
	if !v.IsNull() {
		t.Error("expected IsNull() == true")
	}
	if v.Bool() != false {
		t.Error("expected Bool() == false")
	}
	if v.Int() != 0 {
		t.Error("expected Int() == 0")
	}
	if v.Float() != 0.0 {
		t.Error("expected Float() == 0.0")
	}
	if v.String() != "" {
		t.Errorf("expected String() == \"\", got %q", v.String())
	}
	if v.Stringify() != "null" {
		t.Errorf("expected Stringify() == \"null\", got %q", v.Stringify())
	}
}

func TestNewBool(t *testing.T) {
	tests := []struct {
		input    bool
		intVal   int64
		floatVal float64
		strVal   string
	}{
		{true, 1, 1.0, "true"},
		{false, 0, 0.0, "false"},
	}
	for _, tt := range tests {
		v := NewBool(tt.input)
		if v.Type() != TypeBool {
			t.Errorf("expected TypeBool, got %v", v.Type())
		}
		if v.IsNull() {
			t.Error("expected IsNull() == false")
		}
		if v.Bool() != tt.input {
			t.Errorf("expected Bool() == %v", tt.input)
		}
		if v.Int() != tt.intVal {
			t.Errorf("expected Int() == %d, got %d", tt.intVal, v.Int())
		}
		if v.Float() != tt.floatVal {
			t.Errorf("expected Float() == %f, got %f", tt.floatVal, v.Float())
		}
		if v.String() != tt.strVal {
			t.Errorf("expected String() == %q, got %q", tt.strVal, v.String())
		}
	}
}

func TestNewInt(t *testing.T) {
	tests := []int64{0, 1, -1, 42, -999, 1<<63 - 1}
	for _, input := range tests {
		v := NewInt(input)
		if v.Type() != TypeInt {
			t.Errorf("expected TypeInt, got %v", v.Type())
		}
		if v.IsNull() {
			t.Error("expected IsNull() == false")
		}
		if v.Int() != input {
			t.Errorf("expected Int() == %d, got %d", input, v.Int())
		}
		if v.Float() != float64(input) {
			t.Errorf("expected Float() == %f, got %f", float64(input), v.Float())
		}
		if input == 0 && v.Bool() {
			t.Error("expected Bool() == false for 0")
		}
		if input != 0 && !v.Bool() {
			t.Error("expected Bool() == true for non-zero")
		}
	}
}

func TestNewFloat(t *testing.T) {
	tests := []float64{0.0, 1.5, -3.14, 1e10, -1e-10}
	for _, input := range tests {
		v := NewFloat(input)
		if v.Type() != TypeFloat {
			t.Errorf("expected TypeFloat, got %v", v.Type())
		}
		if v.IsNull() {
			t.Error("expected IsNull() == false")
		}
		if v.Float() != input {
			t.Errorf("expected Float() == %f, got %f", input, v.Float())
		}
		if input == 0.0 && v.Bool() {
			t.Error("expected Bool() == false for 0.0")
		}
		if input != 0.0 && !v.Bool() {
			t.Error("expected Bool() == true for non-zero")
		}
	}
}

func TestNewString(t *testing.T) {
	tests := []string{"", "hello", "你好世界", "line\nbreak"}
	for _, input := range tests {
		v := NewString(input)
		if v.Type() != TypeString {
			t.Errorf("expected TypeString, got %v", v.Type())
		}
		if v.IsNull() {
			t.Error("expected IsNull() == false")
		}
		if v.String() != input {
			t.Errorf("expected String() == %q, got %q", input, v.String())
		}
		if input == "" && v.Bool() {
			t.Error("expected Bool() == false for empty string")
		}
		if input != "" && !v.Bool() {
			t.Error("expected Bool() == true for non-empty string")
		}
	}
}

func TestNewArray(t *testing.T) {
	arr := []Value{NewInt(1), NewString("a"), NewBool(true)}
	v := NewArray(arr)
	if v.Type() != TypeArray {
		t.Errorf("expected TypeArray, got %v", v.Type())
	}
	if v.IsNull() {
		t.Error("expected IsNull() == false")
	}
	if len(v.Array()) != 3 {
		t.Errorf("expected len == 3, got %d", len(v.Array()))
	}
	if v.Int() != 3 {
		t.Errorf("expected Int() == 3, got %d", v.Int())
	}

	empty := NewArray([]Value{})
	if empty.Bool() {
		t.Error("expected Bool() == false for empty array")
	}
}

func TestNewObject(t *testing.T) {
	obj := map[string]Value{"a": NewInt(1), "b": NewString("x")}
	v := NewObject(obj)
	if v.Type() != TypeObject {
		t.Errorf("expected TypeObject, got %v", v.Type())
	}
	if len(v.Object()) != 2 {
		t.Errorf("expected len == 2, got %d", len(v.Object()))
	}
	if v.Int() != 2 {
		t.Errorf("expected Int() == 2, got %d", v.Int())
	}

	empty := NewObject(map[string]Value{})
	if empty.Bool() {
		t.Error("expected Bool() == false for empty object")
	}
}

func TestNewBigInt(t *testing.T) {
	bi := big.NewInt(1234567890123456789)
	v := NewBigInt(bi)
	if v.Type() != TypeBigInt {
		t.Errorf("expected TypeBigInt, got %v", v.Type())
	}
	if v.String() != "1234567890123456789" {
		t.Errorf("expected \"1234567890123456789\", got %q", v.String())
	}
}

func TestNewBigDecimal(t *testing.T) {
	bd := new(big.Rat).SetFrac64(1, 3)
	v := NewBigDecimal(bd)
	if v.Type() != TypeBigDecimal {
		t.Errorf("expected TypeBigDecimal, got %v", v.Type())
	}
	if !v.Bool() {
		t.Error("expected Bool() == true for 1/3")
	}
}

// =============================================================================
// Equals 相等性测试
// =============================================================================

func TestEquals(t *testing.T) {
	tests := []struct {
		a, b     Value
		expected bool
		desc     string
	}{
		{NewNull(), NewNull(), true, "null == null"},
		{NewNull(), NewInt(0), false, "null != int(0)"},
		{NewBool(true), NewBool(true), true, "true == true"},
		{NewBool(true), NewBool(false), false, "true != false"},
		{NewInt(42), NewInt(42), true, "int(42) == int(42)"},
		{NewInt(42), NewInt(99), false, "int(42) != int(99)"},
		{NewInt(1), NewFloat(1.0), true, "int(1) == float(1.0)"},
		{NewFloat(3.14), NewFloat(3.14), true, "float(3.14) == float(3.14)"},
		{NewString("hello"), NewString("hello"), true, "string hello == hello"},
		{NewString("hello"), NewString("world"), false, "string hello != world"},
		{NewInt(1), NewString("1"), false, "int(1) != string(1)"},
	}

	for _, tt := range tests {
		result := tt.a.Equals(tt.b)
		if result != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.desc, tt.expected, result)
		}
	}
}

func TestEqualsBigInt(t *testing.T) {
	bi := NewBigInt(big.NewInt(42))
	if !bi.Equals(NewInt(42)) {
		t.Error("BigInt(42) should equal Int(42)")
	}
	if !bi.Equals(NewBigInt(big.NewInt(42))) {
		t.Error("BigInt(42) should equal BigInt(42)")
	}
}

func TestEqualsBigDecimal(t *testing.T) {
	bd := NewBigDecimal(new(big.Rat).SetFrac64(1, 2))
	if !bd.Equals(NewFloat(0.5)) {
		t.Error("BigDecimal(1/2) should equal Float(0.5)")
	}
}

// =============================================================================
// 算术运算 - 整数
// =============================================================================

func TestIntArithmetic(t *testing.T) {
	a := NewInt(10)
	b := NewInt(3)

	// 加法
	result := a.Add(b)
	if result.Int() != 13 {
		t.Errorf("10 + 3: expected 13, got %d", result.Int())
	}

	// 减法
	result = a.Sub(b)
	if result.Int() != 7 {
		t.Errorf("10 - 3: expected 7, got %d", result.Int())
	}

	// 乘法
	result = a.Mul(b)
	if result.Int() != 30 {
		t.Errorf("10 * 3: expected 30, got %d", result.Int())
	}

	// 除法（int / int → float）
	result = a.Div(b)
	if result.Type() != TypeFloat {
		t.Errorf("10 / 3: expected TypeFloat, got %v", result.Type())
	}
	expected := 10.0 / 3.0
	if result.Float() != expected {
		t.Errorf("10 / 3: expected %f, got %f", expected, result.Float())
	}

	// 取模
	result = a.Mod(b)
	if result.Int() != 1 {
		t.Errorf("10 %% 3: expected 1, got %d", result.Int())
	}

	// 取反
	result = a.Negate()
	if result.Int() != -10 {
		t.Errorf("-10: expected -10, got %d", result.Int())
	}
}

func TestIntArithmeticWithFloat(t *testing.T) {
	a := NewInt(10)
	b := NewFloat(3.5)

	result := a.Add(b)
	if result.Type() != TypeFloat {
		t.Errorf("int + float: expected TypeFloat, got %v", result.Type())
	}
	if result.Float() != 13.5 {
		t.Errorf("10 + 3.5: expected 13.5, got %f", result.Float())
	}

	result = a.Mul(b)
	if result.Float() != 35.0 {
		t.Errorf("10 * 3.5: expected 35.0, got %f", result.Float())
	}
}

func TestIntArithmeticWithNull(t *testing.T) {
	a := NewInt(10)
	n := NewNull()

	result := a.Add(n)
	if result.Int() != 10 {
		t.Errorf("10 + null: expected 10, got %d", result.Int())
	}

	result = a.Sub(n)
	if result.Int() != 10 {
		t.Errorf("10 - null: expected 10, got %d", result.Int())
	}

	result = a.Mul(n)
	if result.Int() != 0 {
		t.Errorf("10 * null: expected 0, got %d", result.Int())
	}
}

func TestIntArithmeticWithBool(t *testing.T) {
	a := NewInt(10)
	tr := NewBool(true)
	fl := NewBool(false)

	if a.Add(tr).Int() != 11 {
		t.Errorf("10 + true: expected 11, got %d", a.Add(tr).Int())
	}
	if a.Add(fl).Int() != 10 {
		t.Errorf("10 + false: expected 10, got %d", a.Add(fl).Int())
	}
	if a.Mul(fl).Int() != 0 {
		t.Errorf("10 * false: expected 0, got %d", a.Mul(fl).Int())
	}
}

// =============================================================================
// 算术运算 - 浮点数
// =============================================================================

func TestFloatArithmetic(t *testing.T) {
	a := NewFloat(10.5)
	b := NewFloat(2.5)

	if a.Add(b).Float() != 13.0 {
		t.Errorf("10.5 + 2.5: expected 13.0, got %f", a.Add(b).Float())
	}
	if a.Sub(b).Float() != 8.0 {
		t.Errorf("10.5 - 2.5: expected 8.0, got %f", a.Sub(b).Float())
	}
	if a.Mul(b).Float() != 26.25 {
		t.Errorf("10.5 * 2.5: expected 26.25, got %f", a.Mul(b).Float())
	}
	if a.Div(b).Float() != 4.2 {
		t.Errorf("10.5 / 2.5: expected 4.2, got %f", a.Div(b).Float())
	}
	if a.Negate().Float() != -10.5 {
		t.Errorf("-10.5: expected -10.5, got %f", a.Negate().Float())
	}
}

func TestFloatArithmeticWithInt(t *testing.T) {
	a := NewFloat(10.5)
	b := NewInt(2)

	if a.Add(b).Float() != 12.5 {
		t.Errorf("10.5 + 2: expected 12.5, got %f", a.Add(b).Float())
	}
}

// =============================================================================
// 算术运算 - 字符串
// =============================================================================

func TestStringArithmetic(t *testing.T) {
	a := NewString("hello")
	b := NewString("world")

	// 字符串 + 字符串 → 拼接
	result := a.Add(b)
	if result.String() != "helloworld" {
		t.Errorf("hello + world: expected \"helloworld\", got %q", result.String())
	}

	// 字符串 + 非字符串 → 错误
	err := a.Add(NewInt(1))
	if !err.IsNull() {
		t.Error("string + int should return error value (null)")
	}

	// 字符串减法 → 错误
	err = a.Sub(b)
	if !err.IsNull() {
		t.Error("string - string should return error value")
	}
}

// =============================================================================
// 算术运算 - 数组
// =============================================================================

func TestArrayArithmetic(t *testing.T) {
	a := NewArray([]Value{NewInt(1), NewInt(2)})
	b := NewArray([]Value{NewInt(3), NewInt(4)})

	// 数组 + 数组 → 拼接
	result := a.Add(b)
	if result.Type() != TypeArray {
		t.Errorf("array + array: expected TypeArray, got %v", result.Type())
	}
	if len(result.Array()) != 4 {
		t.Errorf("expected 4 elements, got %d", len(result.Array()))
	}
	if result.Array()[0].Int() != 1 {
		t.Errorf("expected first element 1, got %d", result.Array()[0].Int())
	}
	if result.Array()[3].Int() != 4 {
		t.Errorf("expected last element 4, got %d", result.Array()[3].Int())
	}

	// 数组 + 非数组 → 错误
	err := a.Add(NewInt(1))
	if !err.IsNull() {
		t.Error("array + int should return error value")
	}
}

// =============================================================================
// 算术运算 - 对象
// =============================================================================

func TestObjectArithmetic(t *testing.T) {
	a := NewObject(map[string]Value{"x": NewInt(1)})
	b := NewObject(map[string]Value{"y": NewInt(2)})

	// 对象 + 对象 → 合并
	result := a.Add(b)
	if result.Type() != TypeObject {
		t.Errorf("object + object: expected TypeObject, got %v", result.Type())
	}
	if len(result.Object()) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result.Object()))
	}
	if result.Object()["x"].Int() != 1 {
		t.Errorf("expected x == 1")
	}
	if result.Object()["y"].Int() != 2 {
		t.Errorf("expected y == 2")
	}
}

// =============================================================================
// 算术运算 - BigInt
// =============================================================================

func TestBigIntArithmetic(t *testing.T) {
	a := NewBigInt(big.NewInt(1000000000000))
	b := NewBigInt(big.NewInt(2000000000000))

	result := a.Add(b)
	expected := "3000000000000"
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}

	result = a.Mul(b)
	expected = "2000000000000000000000000"
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}
}

func TestBigIntWithInt(t *testing.T) {
	bi := NewBigInt(big.NewInt(100))
	in := NewInt(50)

	result := bi.Add(in)
	if result.String() != "150" {
		t.Errorf("expected 150, got %s", result.String())
	}
}

func TestBigIntWithFloat(t *testing.T) {
	bi := NewBigInt(big.NewInt(100))
	fl := NewFloat(0.5)

	result := bi.Add(fl)
	if result.Type() != TypeFloat {
		t.Errorf("expected TypeFloat, got %v", result.Type())
	}
	if result.Float() != 100.5 {
		t.Errorf("expected 100.5, got %f", result.Float())
	}
}

func TestBigIntWithBigDecimal(t *testing.T) {
	bi := NewBigInt(big.NewInt(100))
	bd := NewBigDecimal(new(big.Rat).SetFrac64(1, 4))

	result := bi.Add(bd)
	if result.Type() != TypeBigDecimal {
		t.Errorf("expected TypeBigDecimal, got %v", result.Type())
	}
	if !result.Equals(NewBigDecimal(new(big.Rat).SetFrac64(401, 4))) {
		t.Errorf("expected 401/4, got %s", result.String())
	}
}

// =============================================================================
// 算术运算 - BigDecimal
// =============================================================================

func TestBigDecimalArithmetic(t *testing.T) {
	a := NewBigDecimal(new(big.Rat).SetFrac64(1, 3))
	b := NewBigDecimal(new(big.Rat).SetFrac64(1, 6))

	// 1/3 + 1/6 = 1/2
	result := a.Add(b)
	if !result.Equals(NewBigDecimal(new(big.Rat).SetFrac64(1, 2))) {
		t.Errorf("1/3 + 1/6: expected 1/2, got %s", result.String())
	}

	// 1/3 - 1/6 = 1/6
	result = a.Sub(b)
	if !result.Equals(NewBigDecimal(new(big.Rat).SetFrac64(1, 6))) {
		t.Errorf("1/3 - 1/6: expected 1/6, got %s", result.String())
	}

	// 1/3 * 1/6 = 1/18
	result = a.Mul(b)
	if !result.Equals(NewBigDecimal(new(big.Rat).SetFrac64(1, 18))) {
		t.Errorf("1/3 * 1/6: expected 1/18, got %s", result.String())
	}

	// (1/3) / (1/6) = 2
	result = a.Div(b)
	if !result.Equals(NewInt(2)) {
		t.Errorf("(1/3) / (1/6): expected 2, got %s", result.String())
	}
}

func TestBigDecimalPrecision(t *testing.T) {
	// 测试精度：0.1 + 0.2 != 0.3（float64），但 BigDecimal 精确
	a := NewBigDecimal(new(big.Rat).SetFrac64(1, 10))
	b := NewBigDecimal(new(big.Rat).SetFrac64(2, 10))
	c := NewBigDecimal(new(big.Rat).SetFrac64(3, 10))

	result := a.Add(b)
	if !result.Equals(c) {
		t.Errorf("0.1 + 0.2 should equal 0.3, got %s", result.String())
	}
}

// =============================================================================
// 除法和取模 - 边界情况
// =============================================================================

func TestDivisionByZero(t *testing.T) {
	// IEEE 754: 除零返回 Inf/NaN，非 error
	tests := []struct {
		desc    string
		fn      func() Value
		wantInf int // 0=不检查, 1=+Inf, -1=-Inf
		wantNaN bool
	}{
		{"int / int(0)", func() Value { return NewInt(10).Div(NewInt(0)) }, 1, false},
		{"int / float(0)", func() Value { return NewInt(10).Div(NewFloat(0.0)) }, 1, false},
		{"int / null", func() Value { return NewInt(10).Div(NewNull()) }, 0, false}, // 仍为 error
		{"float / int(0)", func() Value { return NewFloat(10.0).Div(NewInt(0)) }, 1, false},
		{"float / float(0)", func() Value { return NewFloat(10.0).Div(NewFloat(0.0)) }, 1, false},
		{"-int / int(0)", func() Value { return NewInt(-10).Div(NewInt(0)) }, -1, false},
		{"0 / int(0)", func() Value { return NewInt(0).Div(NewInt(0)) }, 0, true},
		{"bigint / int(0)", func() Value { return NewBigInt(big.NewInt(10)).Div(NewInt(0)) }, 1, false},
		{"bigdecimal / bigdecimal(0)", func() Value {
			return NewBigDecimal(new(big.Rat).SetInt64(10)).Div(NewBigDecimal(new(big.Rat)))
		}, 1, false},
	}

	for _, tt := range tests {
		result := tt.fn()
		if tt.wantInf != 0 {
			if !math.IsInf(result.Float(), tt.wantInf) {
				t.Errorf("%s: expected Inf(%d), got %v", tt.desc, tt.wantInf, result)
			}
		} else if tt.wantNaN {
			if !math.IsNaN(result.Float()) {
				t.Errorf("%s: expected NaN, got %v", tt.desc, result)
			}
		} else {
			// int/null 除法仍为 error
			if !result.IsNull() {
				t.Errorf("%s: expected error (null), got %v", tt.desc, result)
			}
		}
	}
}

func TestModuloByZero(t *testing.T) {
	tests := []struct {
		desc    string
		fn      func() Value
		wantNaN bool
	}{
		{"int % int(0)", func() Value { return NewInt(10).Mod(NewInt(0)) }, true},
		{"int % null", func() Value { return NewInt(10).Mod(NewNull()) }, false}, // 仍为 error
	}

	for _, tt := range tests {
		result := tt.fn()
		if tt.wantNaN {
			if !math.IsNaN(result.Float()) {
				t.Errorf("%s: expected NaN, got %v", tt.desc, result)
			}
		} else {
			if !result.IsNull() {
				t.Errorf("%s: expected error (null), got %v", tt.desc, result)
			}
		}
	}
}

// =============================================================================
// 比较运算
// =============================================================================

func TestIntComparison(t *testing.T) {
	a := NewInt(5)
	b := NewInt(10)

	if !a.Less(b) {
		t.Error("5 < 10 should be true")
	}
	if a.Greater(b) {
		t.Error("5 > 10 should be false")
	}
	if !a.LessEqual(b) {
		t.Error("5 <= 10 should be true")
	}
	if a.GreaterEqual(b) {
		t.Error("5 >= 10 should be false")
	}
	if !a.LessEqual(NewInt(5)) {
		t.Error("5 <= 5 should be true")
	}
	if !a.GreaterEqual(NewInt(5)) {
		t.Error("5 >= 5 should be true")
	}
}

func TestFloatComparison(t *testing.T) {
	a := NewFloat(3.14)
	b := NewFloat(2.71)

	if a.Less(b) {
		t.Error("3.14 < 2.71 should be false")
	}
	if !a.Greater(b) {
		t.Error("3.14 > 2.71 should be true")
	}
}

func TestMixedComparison(t *testing.T) {
	// int vs float
	a := NewInt(5)
	b := NewFloat(5.5)

	if !a.Less(b) {
		t.Error("int(5) < float(5.5) should be true")
	}
	if b.Greater(a) == false {
		t.Error("float(5.5) > int(5) should be true")
	}
}

func TestNullComparison(t *testing.T) {
	n := NewNull()

	if n.Less(NewInt(0)) {
		t.Error("null < 0 should be false")
	}
	if n.Greater(NewInt(0)) {
		t.Error("null > 0 should be false")
	}
	if !n.Less(NewInt(1)) {
		t.Error("null < 1 should be true")
	}
	if n.LessEqual(NewInt(-1)) {
		t.Error("null <= -1 should be false")
	}
}

func TestStringComparison(t *testing.T) {
	a := NewString("apple")
	b := NewString("banana")

	if !a.Less(b) {
		t.Errorf("%q < %q should be true", a.String(), b.String())
	}
	if a.Greater(b) {
		t.Errorf("%q > %q should be false", a.String(), b.String())
	}
}

func TestBigIntComparison(t *testing.T) {
	a := NewBigInt(big.NewInt(100))
	b := NewBigInt(big.NewInt(200))

	if !a.Less(b) {
		t.Error("BigInt(100) < BigInt(200) should be true")
	}
	if a.Greater(b) {
		t.Error("BigInt(100) > BigInt(200) should be false")
	}
}

func TestBigDecimalComparison(t *testing.T) {
	a := NewBigDecimal(new(big.Rat).SetFrac64(1, 3))
	b := NewBigDecimal(new(big.Rat).SetFrac64(1, 2))

	if !a.Less(b) {
		t.Error("1/3 < 1/2 should be true")
	}
	if a.Greater(b) {
		t.Error("1/3 > 1/2 should be false")
	}
}

// =============================================================================
// 类型转换
// =============================================================================

func TestToBigInt(t *testing.T) {
	tests := []struct {
		input    Value
		expected string
		desc     string
	}{
		{NewInt(42), "42", "int to bigint"},
		{NewBool(true), "1", "bool(true) to bigint"},
		{NewBool(false), "0", "bool(false) to bigint"},
		{NewNull(), "0", "null to bigint"},
		{NewFloat(3.7), "3", "float to bigint (truncated)"},
	}

	for _, tt := range tests {
		result := tt.input.ToBigInt()
		if result.Type() != TypeBigInt {
			t.Errorf("%s: expected TypeBigInt, got %v", tt.desc, result.Type())
		}
		if result.String() != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.desc, tt.expected, result.String())
		}
	}
}

func TestToBigDecimal(t *testing.T) {
	tests := []struct {
		input    Value
		expected string
		desc     string
	}{
		{NewInt(42), "42", "int to bigdecimal"},
		{NewFloat(0.5), "1/2", "float to bigdecimal"},
		{NewBool(true), "1", "bool(true) to bigdecimal"},
		{NewNull(), "0", "null to bigdecimal"},
	}

	for _, tt := range tests {
		result := tt.input.ToBigDecimal()
		if result.Type() != TypeBigDecimal {
			t.Errorf("%s: expected TypeBigDecimal, got %v", tt.desc, result.Type())
		}
		if result.String() != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.desc, tt.expected, result.String())
		}
	}
}

// =============================================================================
// 运算入口函数（value_ops.go）
// =============================================================================

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    Value
		expected bool
	}{
		{NewInt(1), true},
		{NewFloat(1.0), true},
		{NewBigInt(big.NewInt(1)), true},
		{NewBigDecimal(new(big.Rat).SetInt64(1)), true},
		{NewNull(), false},
		{NewBool(true), false},
		{NewString("1"), false},
		{NewArray([]Value{}), false},
	}

	for _, tt := range tests {
		if IsNumeric(tt.input) != tt.expected {
			t.Errorf("IsNumeric(%v): expected %v", tt.input, tt.expected)
		}
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		input    Value
		expected bool
	}{
		{NewNull(), false},
		{NewBool(true), true},
		{NewBool(false), false},
		{NewInt(0), false},
		{NewInt(1), true},
		{NewInt(-1), true},
		{NewFloat(0.0), false},
		{NewFloat(0.1), true},
		{NewString(""), false},
		{NewString("a"), true},
		{NewArray([]Value{}), false},
		{NewArray([]Value{NewInt(1)}), true},
		{NewObject(map[string]Value{}), false},
		{NewObject(map[string]Value{"a": NewInt(1)}), true},
		{NewBigDecimal(new(big.Rat)), false},
		{NewBigDecimal(new(big.Rat).SetInt64(1)), true},
	}

	for _, tt := range tests {
		if IsTruthy(tt.input) != tt.expected {
			t.Errorf("IsTruthy(%v [%s]): expected %v", tt.input.Stringify(), tt.input.Type(), tt.expected)
		}
	}
}

func TestValueOpsDispatch(t *testing.T) {
	a := NewInt(5)
	b := NewInt(3)

	if ValueAdd(a, b).Int() != 8 {
		t.Error("ValueAdd(5, 3) expected 8")
	}
	if ValueSub(a, b).Int() != 2 {
		t.Error("ValueSub(5, 3) expected 2")
	}
	if ValueMul(a, b).Int() != 15 {
		t.Error("ValueMul(5, 3) expected 15")
	}
	if ValueNegate(a).Int() != -5 {
		t.Error("ValueNegate(5) expected -5")
	}
	if !ValueLess(b, a) {
		t.Error("ValueLess(3, 5) expected true")
	}
	if !ValueGreater(a, b) {
		t.Error("ValueGreater(5, 3) expected true")
	}
}

func TestConcatValues(t *testing.T) {
	a := NewString("hello ")
	b := NewString("world")
	result := ConcatValues(a, b)
	if result.String() != "hello world" {
		t.Errorf("expected \"hello world\", got %q", result.String())
	}
}

// =============================================================================
// ValueType.String()
// =============================================================================

func TestValueTypeString(t *testing.T) {
	tests := []struct {
		vt       ValueType
		expected string
	}{
		{TypeNull, "null"},
		{TypeBool, "bool"},
		{TypeInt, "int"},
		{TypeFloat, "float"},
		{TypeString, "string"},
		{TypeArray, "array"},
		{TypeObject, "object"},
		{TypeFunc, "func"},
		{TypeBigDecimal, "bigdecimal"},
		{ValueType(99), "unknown"},
	}

	for _, tt := range tests {
		if tt.vt.String() != tt.expected {
			t.Errorf("ValueType(%d).String(): expected %q, got %q", tt.vt, tt.expected, tt.vt.String())
		}
	}
}

// =============================================================================
// Stringify
// =============================================================================

func TestStringify(t *testing.T) {
	tests := []struct {
		input    Value
		expected string
	}{
		{NewNull(), "null"},
		{NewBool(true), "true"},
		{NewInt(42), "42"},
		{NewFloat(3.14), "3.14"},
		{NewString("hello"), `"hello"`},
		{NewArray([]Value{NewInt(1), NewInt(2)}), "[1, 2]"},
		{NewBigInt(big.NewInt(999)), "999"},
	}

	for _, tt := range tests {
		if tt.input.Stringify() != tt.expected {
			t.Errorf("Stringify(%v): expected %q, got %q", tt.input.Type(), tt.expected, tt.input.Stringify())
		}
	}
}

// =============================================================================
// 非法操作
// =============================================================================

func TestInvalidOperations(t *testing.T) {
	fn := NewFunc("test", nil)
	obj := NewObject(map[string]Value{})
	arr := NewArray([]Value{})

	// func 算术
	if !fn.Add(NewInt(1)).IsNull() {
		t.Error("func + int should return error")
	}
	if !fn.Sub(NewInt(1)).IsNull() {
		t.Error("func - int should return error")
	}
	if !fn.Negate().IsNull() {
		t.Error("-func should return error")
	}

	// object 非法操作
	if !obj.Sub(NewInt(1)).IsNull() {
		t.Error("object - int should return error")
	}
	if !obj.Mul(NewInt(1)).IsNull() {
		t.Error("object * int should return error")
	}

	// array 非法操作
	if !arr.Sub(NewInt(1)).IsNull() {
		t.Error("array - int should return error")
	}

	// string 非法操作
	s := NewString("hello")
	if !s.Mul(NewInt(3)).IsNull() {
		t.Error("string * int should return error")
	}
	if !s.Div(NewInt(1)).IsNull() {
		t.Error("string / int should return error")
	}
	if !s.Mod(NewInt(1)).IsNull() {
		t.Error("string % int should return error")
	}
	if !s.Negate().IsNull() {
		t.Error("-string should return error")
	}
}

// =============================================================================
// floatToRat 精度测试
// =============================================================================

func TestFloatToRatPrecision(t *testing.T) {
	expected := new(big.Rat).SetFrac(big.NewInt(1), big.NewInt(10)) // 1/10
	got := floatToRat(0.1)
	if got.Cmp(expected) != 0 {
		t.Errorf("floatToRat(0.1) = %s, want 1/10", got.RatString())
	}

	expected2, _ := new(big.Rat).SetString("3.14")
	got2 := floatToRat(3.14)
	if got2.Cmp(expected2) != 0 {
		t.Errorf("floatToRat(3.14) = %s, want 157/50", got2.RatString())
	}
}

func TestFloatToBigDecimalArithmetic(t *testing.T) {
	f := NewFloat(0.1)
	bd := NewBigDecimal(new(big.Rat).SetInt64(2))

	// 0.1 + 2 = 2.1
	add := f.Add(bd)
	expectedAdd, _ := new(big.Rat).SetString("2.1")
	gotRat := add.(*BigDecimalValue).BigRat()
	if gotRat.Cmp(expectedAdd) != 0 {
		t.Errorf("0.1 + BigDecimal(2) = %s, want 21/10", gotRat.RatString())
	}

	// 0.1 - 2 = -1.9
	sub := f.Sub(bd)
	expectedSub, _ := new(big.Rat).SetString("-1.9")
	gotSub := sub.(*BigDecimalValue).BigRat()
	if gotSub.Cmp(expectedSub) != 0 {
		t.Errorf("0.1 - BigDecimal(2) = %s, want -19/10", gotSub.RatString())
	}

	// 0.1 * 2 = 0.2
	mul := f.Mul(bd)
	expectedMul, _ := new(big.Rat).SetString("0.2")
	gotMul := mul.(*BigDecimalValue).BigRat()
	if gotMul.Cmp(expectedMul) != 0 {
		t.Errorf("0.1 * BigDecimal(2) = %s, want 1/5", gotMul.RatString())
	}

	// 0.1 / 2 = 0.05
	div := f.Div(bd)
	expectedDiv, _ := new(big.Rat).SetString("0.05")
	gotDiv := div.(*BigDecimalValue).BigRat()
	if gotDiv.Cmp(expectedDiv) != 0 {
		t.Errorf("0.1 / BigDecimal(2) = %s, want 1/20", gotDiv.RatString())
	}
}

func TestFloatToBigDecimalComparison(t *testing.T) {
	f := NewFloat(0.3)
	bd := NewBigDecimal(new(big.Rat).SetFrac(big.NewInt(3), big.NewInt(10))) // 3/10

	if !f.Equals(bd) {
		t.Error("float(0.3) should equal BigDecimal(3/10)")
	}

	if f.Less(bd) {
		t.Error("float(0.3) should not be less than BigDecimal(3/10)")
	}

	if f.Greater(bd) {
		t.Error("float(0.3) should not be greater than BigDecimal(3/10)")
	}
}
