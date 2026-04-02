package engine

import (
	"fmt"
	"maps"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"sync"

	"github.com/gnuos/jpl/gc"
)

// ValueType 值类型枚举
type ValueType int

const (
	TypeNull ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeArray
	TypeObject
	TypeFunc
	TypeBigInt
	TypeBigDecimal
	TypeStream
	TypeError
	TypeRange
	TypeRegex
)

// 小整数缓存范围
const (
	smallIntMin = -256
	smallIntMax = 1024
)

// 字符串内部化配置
const (
	stringInternMaxLen = 64 // 仅缓存长度 ≤64 的字符串
)

// 全局单例值，消除重复堆分配
var (
	nullSingleton  Value = &nullValue{}
	boolTrueValue  Value = &boolValue{value: true}
	boolFalseValue Value = &boolValue{value: false}
	smallIntCache  []Value
)

// 字符串内部化缓存池
// key: string, value: *stringValue
var stringInternPool sync.Map

func init() {
	smallIntCache = make([]Value, smallIntMax-smallIntMin+1)
	for i := smallIntMin; i <= smallIntMax; i++ {
		smallIntCache[i-smallIntMin] = &intValue{value: int64(i)}
	}
}

func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBool:
		return "bool"
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	case TypeFunc:
		return "func"
	case TypeBigInt:
		return "bigint"
	case TypeBigDecimal:
		return "bigdecimal"
	case TypeStream:
		return "stream"
	case TypeError:
		return "error"
	case TypeRange:
		return "range"
	case TypeRegex:
		return "regex"
	default:
		return "unknown"
	}
}

// Value 是 JPL 脚本引擎中所有值的统一接口。
//
// 所有脚本中的值（null、bool、int、float、string、array、object、function、bigint、bigdecimal、error）
// 都实现了此接口，提供统一的操作方式。
//
// 值类型分为两大类：
//  1. 标量类型（null、bool、int、float、string）：不可变，值传递
//  2. 复合类型（array、object、function、bigint、bigdecimal、error）：引用类型
//
// 接口提供的能力：
//   - 类型查询（Type、IsNull）
//   - 类型转换（Bool、Int、Float、String、Array、Object）
//   - 运算符支持（Add、Sub、Mul、Div、Mod、Negate）
//   - 比较运算（Equals、Less、Greater、LessEqual、GreaterEqual）
//   - 集合操作（Len）
//   - 大数支持（ToBigInt、ToBigDecimal）
//
// 使用示例：
//
//	var v engine.Value = engine.NewInt(42)
//	fmt.Println(v.Type())        // 输出: int
//	fmt.Println(v.Int())         // 输出: 42
//	fmt.Println(v.String())      // 输出: 42
//	fmt.Println(v.Stringify())   // 输出: 42
//
//	// 运算
//	result := v.Add(engine.NewInt(8))
//	fmt.Println(result.Int())    // 输出: 50
//
//	// 比较
//	fmt.Println(v.Equals(engine.NewInt(42))) // 输出: true
//	fmt.Println(v.Less(engine.NewInt(100)))  // 输出: true
type Value interface {
	// Type 返回值的类型标识
	Type() ValueType

	// IsNull 检查值是否为 null
	// 只有 nullValue 返回 true，其他类型返回 false
	IsNull() bool

	// Bool 获取值的布尔表示
	// 遵循脚本语言的类型转换规则：null/false/0/"" 为 false，其他为 true
	Bool() bool

	// Int 获取整数值
	// 如果值不是整数类型，按类型转换规则转换
	// float 截断小数，string 解析整数，其他返回 0
	Int() int64

	// Float 获取浮点值
	// 如果值不是浮点类型，按类型转换规则转换
	// int 转为浮点，string 解析浮点数，其他返回 0
	Float() float64

	// String 获取字符串值
	// 所有类型都可转为字符串，返回值的文本表示
	String() string

	// Array 获取数组值
	// 只有 arrayValue 返回有效数组，其他返回 nil
	Array() []Value

	// Object 获取对象值
	// 只有 objectValue 返回有效映射，其他返回 nil
	Object() map[string]Value

	// Len 返回值的长度
	// string: 字符数（UTF-8）
	// array: 元素数
	// object: 键值对数
	// 其他: 0
	Len() int

	// Equals 比较两个值是否相等
	// 支持类型比较：int 和 float 可比较，其他类型必须严格相同
	Equals(other Value) bool

	// Stringify 返回值的 JSON 序列化字符串
	// 用于输出和序列化，格式与 String() 不同
	Stringify() string

	// ToBigInt 将值转换为大整数
	// int 精确转换，float 截断小数，string 解析，其他返回 0
	ToBigInt() Value

	// ToBigDecimal 将值转换为大数值
	// int/float/string 精确转换，其他返回 0
	ToBigDecimal() Value

	// Add 执行加法运算
	// 支持数字类型相加，字符串拼接，数组合并等
	Add(other Value) Value

	// Sub 执行减法运算
	// 仅支持数字类型
	Sub(other Value) Value

	// Mul 执行乘法运算
	// 仅支持数字类型
	Mul(other Value) Value

	// Div 执行除法运算
	// 支持数字类型，遵循 IEEE 754 标准（除以零返回 Inf/NaN）
	Div(other Value) Value

	// Mod 执行取模运算
	// 支持整数和浮点数，除以零返回 NaN
	Mod(other Value) Value

	// Negate 执行取反运算（一元负号）
	// 仅支持数字类型
	Negate() Value

	// Less 小于比较（<）
	// 支持数字类型比较，string 按字典序比较
	Less(other Value) bool

	// Greater 大于比较（>）
	// 逻辑等同于 !LessEqual(other)
	Greater(other Value) bool

	// LessEqual 小于等于比较（<=）
	// 支持数字类型比较，string 按字典序比较
	LessEqual(other Value) bool

	// GreaterEqual 大于等于比较（>=）
	// 逻辑等同于 !Less(other)
	GreaterEqual(other Value) bool
}

// runtimeError 运行时错误值
type runtimeError struct {
	msg string
}

func (v *runtimeError) Type() ValueType          { return TypeNull }
func (v *runtimeError) IsNull() bool             { return true }
func (v *runtimeError) Bool() bool               { return false }
func (v *runtimeError) Int() int64               { return 0 }
func (v *runtimeError) Float() float64           { return 0.0 }
func (v *runtimeError) String() string           { return "" }
func (v *runtimeError) Array() []Value           { return nil }
func (v *runtimeError) Object() map[string]Value { return nil }
func (v *runtimeError) Equals(other Value) bool  { return other.IsNull() }
func (v *runtimeError) Stringify() string        { return "null" }
func (v *runtimeError) Len() int                 { return 0 }
func (v *runtimeError) ToBigInt() Value          { return NewBigInt(new(big.Int)) }
func (v *runtimeError) ToBigDecimal() Value      { return NewBigDecimal(new(big.Rat)) }
func (v *runtimeError) Add(other Value) Value    { return v }
func (v *runtimeError) Sub(other Value) Value    { return v }
func (v *runtimeError) Mul(other Value) Value    { return v }
func (v *runtimeError) Div(other Value) Value    { return v }
func (v *runtimeError) Mod(other Value) Value    { return v }
func (v *runtimeError) Negate() Value            { return v }
func (v *runtimeError) Less(other Value) bool    { return false }
func (v *runtimeError) Greater(other Value) bool { return false }
func (v *runtimeError) LessEqual(other Value) bool {
	return other.IsNull() || other.Type() == TypeNull
}
func (v *runtimeError) GreaterEqual(other Value) bool {
	return other.IsNull() || other.Type() == TypeNull
}

// newRuntimeError 创建运行时错误值
func newRuntimeError(msg string) Value {
	return &runtimeError{msg: msg}
}

// nullValue 表示 null 值
type nullValue struct{}

func (v *nullValue) Type() ValueType          { return TypeNull }
func (v *nullValue) IsNull() bool             { return true }
func (v *nullValue) Bool() bool               { return false }
func (v *nullValue) Int() int64               { return 0 }
func (v *nullValue) Float() float64           { return 0.0 }
func (v *nullValue) String() string           { return "" }
func (v *nullValue) Array() []Value           { return nil }
func (v *nullValue) Object() map[string]Value { return nil }
func (v *nullValue) Equals(other Value) bool  { return other.IsNull() }
func (v *nullValue) Stringify() string        { return "null" }
func (v *nullValue) Len() int                 { return 0 }
func (v *nullValue) ToBigInt() Value          { return NewBigInt(new(big.Int)) }
func (v *nullValue) ToBigDecimal() Value      { return NewBigDecimal(new(big.Rat)) }

func (v *nullValue) Add(other Value) Value {
	return NewInt(0).Add(other)
}
func (v *nullValue) Sub(other Value) Value {
	return NewInt(0).Sub(other)
}
func (v *nullValue) Mul(other Value) Value {
	return NewInt(0).Mul(other)
}
func (v *nullValue) Div(other Value) Value {
	return NewInt(0).Div(other)
}
func (v *nullValue) Mod(other Value) Value {
	return NewInt(0).Mod(other)
}
func (v *nullValue) Negate() Value { return NewInt(0) }

func (v *nullValue) Less(other Value) bool {
	switch other.Type() {
	case TypeNull:
		return false
	case TypeBool:
		return !other.Bool()
	case TypeInt:
		return 0 < other.Int()
	case TypeFloat:
		return 0.0 < other.Float()
	case TypeBigDecimal:
		return other.ToBigDecimal().(*BigDecimalValue).value.Sign() > 0
	default:
		return false
	}
}
func (v *nullValue) Greater(other Value) bool {
	switch other.Type() {
	case TypeNull:
		return false
	case TypeBool:
		return other.Bool()
	case TypeInt:
		return 0 > other.Int()
	case TypeFloat:
		return 0.0 > other.Float()
	case TypeBigDecimal:
		return other.ToBigDecimal().(*BigDecimalValue).value.Sign() < 0
	default:
		return false
	}
}
func (v *nullValue) LessEqual(other Value) bool    { return !v.Greater(other) }
func (v *nullValue) GreaterEqual(other Value) bool { return !v.Less(other) }

// boolValue 表示布尔值
type boolValue struct {
	value bool
}

func (v *boolValue) Type() ValueType { return TypeBool }
func (v *boolValue) IsNull() bool    { return false }
func (v *boolValue) Bool() bool      { return v.value }
func (v *boolValue) Int() int64 {
	if v.value {
		return 1
	}
	return 0
}
func (v *boolValue) Float() float64 {
	if v.value {
		return 1.0
	}
	return 0.0
}
func (v *boolValue) String() string {
	if v.value {
		return "true"
	}
	return "false"
}
func (v *boolValue) Array() []Value           { return nil }
func (v *boolValue) Object() map[string]Value { return nil }
func (v *boolValue) Equals(other Value) bool {
	if other.Type() != TypeBool {
		return false
	}
	return v.value == other.Bool()
}
func (v *boolValue) Stringify() string {
	if v.value {
		return "true"
	}
	return "false"
}
func (v *boolValue) Len() int { return 0 }
func (v *boolValue) ToBigInt() Value {
	if v.value {
		return NewBigInt(big.NewInt(1))
	}
	return NewBigInt(big.NewInt(0))
}
func (v *boolValue) ToBigDecimal() Value {
	if v.value {
		return NewBigDecimal(new(big.Rat).SetInt64(1))
	}
	return NewBigDecimal(new(big.Rat))
}

func (v *boolValue) Add(other Value) Value {
	return NewInt(v.Int()).Add(other)
}
func (v *boolValue) Sub(other Value) Value {
	return NewInt(v.Int()).Sub(other)
}
func (v *boolValue) Mul(other Value) Value {
	return NewInt(v.Int()).Mul(other)
}
func (v *boolValue) Div(other Value) Value {
	return NewInt(v.Int()).Div(other)
}
func (v *boolValue) Mod(other Value) Value {
	return NewInt(v.Int()).Mod(other)
}
func (v *boolValue) Negate() Value {
	return NewInt(-v.Int())
}

func (v *boolValue) Less(other Value) bool {
	return NewInt(v.Int()).Less(other)
}
func (v *boolValue) Greater(other Value) bool {
	return NewInt(v.Int()).Greater(other)
}
func (v *boolValue) LessEqual(other Value) bool {
	return NewInt(v.Int()).LessEqual(other)
}
func (v *boolValue) GreaterEqual(other Value) bool {
	return NewInt(v.Int()).GreaterEqual(other)
}

// intValue 表示 64 位整数值。
//
// 这是 JPL 中最常用的数字类型，存储 int64 范围的整数。
// 小整数 [-256, 1024] 使用全局缓存，无需分配内存。
//
// 支持的运算：
//   - 算术：+、-、*、/、%、-x（一元负号）
//   - 比较：<、>、<=、>=、==、!=
//   - 类型转换：to float、to bool、to string、to bigint、to bigdecimal
//
// 除法语义：
//   - 整数除法返回浮点数（如 5/2 = 2.5）
//   - 除以零返回 Inf（遵循 IEEE 754）
type intValue struct {
	value int64
}

func (v *intValue) Type() ValueType          { return TypeInt }
func (v *intValue) IsNull() bool             { return false }
func (v *intValue) Bool() bool               { return v.value != 0 }
func (v *intValue) Int() int64               { return v.value }
func (v *intValue) Float() float64           { return float64(v.value) }
func (v *intValue) String() string           { return fmt.Sprintf("%d", v.value) }
func (v *intValue) Array() []Value           { return nil }
func (v *intValue) Object() map[string]Value { return nil }
func (v *intValue) Equals(other Value) bool {
	switch other.Type() {
	case TypeInt:
		return v.value == other.Int()
	case TypeFloat:
		return float64(v.value) == other.Float()
	case TypeBool:
		return v.Bool() == other.Bool()
	case TypeBigDecimal:
		return new(big.Rat).SetInt64(v.value).Cmp(other.ToBigDecimal().(*BigDecimalValue).value) == 0
	default:
		return false
	}
}
func (v *intValue) Stringify() string {
	return fmt.Sprintf("%d", v.value)
}
func (v *intValue) Len() int { return 0 }
func (v *intValue) ToBigInt() Value {
	return NewBigInt(big.NewInt(v.value))
}
func (v *intValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat).SetInt64(v.value))
}

func (v *intValue) Add(other Value) Value {
	switch other.Type() {
	case TypeInt:
		return NewInt(v.value + other.Int())
	case TypeFloat:
		return NewFloat(v.Float() + other.Float())
	case TypeBigDecimal:
		r := new(big.Rat).SetInt64(v.value)
		r.Add(r, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(r)
	case TypeNull:
		return v
	case TypeBool:
		return NewInt(v.value + other.Int())
	default:
		return newRuntimeError("cannot add int and " + other.Type().String())
	}
}

func (v *intValue) Sub(other Value) Value {
	switch other.Type() {
	case TypeInt:
		return NewInt(v.value - other.Int())
	case TypeFloat:
		return NewFloat(v.Float() - other.Float())
	case TypeBigDecimal:
		r := new(big.Rat).SetInt64(v.value)
		r.Sub(r, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(r)
	case TypeNull:
		return v
	case TypeBool:
		return NewInt(v.value - other.Int())
	default:
		return newRuntimeError("cannot subtract " + other.Type().String() + " from int")
	}
}

func (v *intValue) Mul(other Value) Value {
	switch other.Type() {
	case TypeInt:
		return NewInt(v.value * other.Int())
	case TypeFloat:
		return NewFloat(v.Float() * other.Float())
	case TypeBigDecimal:
		r := new(big.Rat).SetInt64(v.value)
		r.Mul(r, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(r)
	case TypeNull:
		return NewInt(0)
	case TypeBool:
		return NewInt(v.value * other.Int())
	default:
		return newRuntimeError("cannot multiply int and " + other.Type().String())
	}
}

func (v *intValue) Div(other Value) Value {
	switch other.Type() {
	case TypeInt:
		return NewFloat(v.Float() / other.Float())
	case TypeFloat:
		return NewFloat(v.Float() / other.Float())
	case TypeBigDecimal:
		rat := other.ToBigDecimal().(*BigDecimalValue).value
		if rat.Sign() == 0 {
			return NewFloat(math.Inf(sign(v.value)))
		}
		r := new(big.Rat).SetInt64(v.value)
		r.Quo(r, rat)
		return NewBigDecimal(r)
	case TypeNull:
		return newRuntimeError("division by null")
	case TypeBool:
		return NewFloat(v.Float() / other.Float())
	default:
		return newRuntimeError("cannot divide int by " + other.Type().String())
	}
}

func (v *intValue) Mod(other Value) Value {
	switch other.Type() {
	case TypeInt:
		if other.Int() == 0 {
			return NewFloat(math.NaN())
		}
		return NewInt(v.value % other.Int())
	case TypeFloat:
		if other.Float() == 0.0 {
			return NewFloat(math.NaN())
		}
		return NewFloat(float64(int64(v.Float()) % int64(other.Float())))
	case TypeNull:
		return newRuntimeError("modulo by null")
	case TypeBool:
		return NewInt(v.value % other.Int())
	default:
		return newRuntimeError("cannot modulo int and " + other.Type().String())
	}
}

func (v *intValue) Negate() Value {
	return NewInt(-v.value)
}

func (v *intValue) Less(other Value) bool {
	switch other.Type() {
	case TypeInt:
		return v.value < other.Int()
	case TypeFloat:
		return v.Float() < other.Float()
	case TypeBigDecimal:
		r := new(big.Rat).SetInt64(v.value)
		return r.Cmp(other.ToBigDecimal().(*BigDecimalValue).value) < 0
	case TypeNull:
		return v.value < 0
	case TypeBool:
		return v.value < other.Int()
	default:
		return false
	}
}

func (v *intValue) Greater(other Value) bool {
	switch other.Type() {
	case TypeInt:
		return v.value > other.Int()
	case TypeFloat:
		return v.Float() > other.Float()
	case TypeBigDecimal:
		r := new(big.Rat).SetInt64(v.value)
		return r.Cmp(other.ToBigDecimal().(*BigDecimalValue).value) > 0
	case TypeNull:
		return v.value > 0
	case TypeBool:
		return v.value > other.Int()
	default:
		return false
	}
}

func (v *intValue) LessEqual(other Value) bool    { return !v.Greater(other) }
func (v *intValue) GreaterEqual(other Value) bool { return !v.Less(other) }

// sign 返回 int64 的符号：-1、0、1
func sign(n int64) int {
	if n > 0 {
		return 1
	}
	if n < 0 {
		return -1
	}
	return 0
}

// floatToRat 将 float64 通过十进制字符串转换为 big.Rat，避免 SetFloat64 的二进制精度问题
func floatToRat(f float64) *big.Rat {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return new(big.Rat).SetFloat64(f)
	}
	return r
}

// signFloat 返回 float64 的符号：-1、0、1（NaN 返回 0）
func signFloat(f float64) int {
	if math.IsNaN(f) {
		return 0
	}
	if f > 0 {
		return 1
	}
	if f < 0 {
		return -1
	}
	return 0
}

// floatValue 表示 64 位浮点数值（IEEE 754 双精度）。
//
// 用于存储有小数部分的数字，支持科学计数法表示。
//
// 特殊值支持：
//   - Inf（正无穷）：1.0/0.0
//   - -Inf（负无穷）：-1.0/0.0
//   - NaN（非数字）：0.0/0.0
//
// 支持的运算：
//   - 算术：+、-、*、/、%-x（一元负号）
//   - 比较：<、>、<=、>=、==、!=（注意：NaN 与任何值都不等，包括自身）
//   - 类型转换：to int（截断）、to bool、to string、to bigdecimal
//
// 与 intValue 运算时会自动提升为浮点数。
type floatValue struct {
	value float64
}

func (v *floatValue) Type() ValueType          { return TypeFloat }
func (v *floatValue) IsNull() bool             { return false }
func (v *floatValue) Bool() bool               { return v.value != 0.0 }
func (v *floatValue) Int() int64               { return int64(v.value) }
func (v *floatValue) Float() float64           { return v.value }
func (v *floatValue) String() string           { return fmt.Sprintf("%g", v.value) }
func (v *floatValue) Array() []Value           { return nil }
func (v *floatValue) Object() map[string]Value { return nil }
func (v *floatValue) Equals(other Value) bool {
	switch other.Type() {
	case TypeFloat:
		return v.value == other.Float()
	case TypeInt:
		return v.value == float64(other.Int())
	case TypeBool:
		return v.Bool() == other.Bool()
	case TypeBigDecimal:
		f, _ := new(big.Float).SetRat(other.ToBigDecimal().(*BigDecimalValue).value).Float64()
		return v.value == f
	default:
		return false
	}
}
func (v *floatValue) Stringify() string {
	return fmt.Sprintf("%g", v.value)
}
func (v *floatValue) Len() int { return 0 }
func (v *floatValue) ToBigInt() Value {
	return NewBigInt(big.NewInt(int64(v.value)))
}
func (v *floatValue) ToBigDecimal() Value {
	return NewBigDecimal(floatToRat(v.value))
}

func (v *floatValue) Add(other Value) Value {
	switch other.Type() {
	case TypeFloat:
		return NewFloat(v.value + other.Float())
	case TypeInt:
		return NewFloat(v.value + other.Float())
	case TypeBigDecimal:
		f := floatToRat(v.value)
		f.Add(f, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(f)
	case TypeNull:
		return v
	case TypeBool:
		return NewFloat(v.value + other.Float())
	default:
		return newRuntimeError("cannot add float and " + other.Type().String())
	}
}

func (v *floatValue) Sub(other Value) Value {
	switch other.Type() {
	case TypeFloat:
		return NewFloat(v.value - other.Float())
	case TypeInt:
		return NewFloat(v.value - other.Float())
	case TypeBigDecimal:
		f := floatToRat(v.value)
		f.Sub(f, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(f)
	case TypeNull:
		return v
	case TypeBool:
		return NewFloat(v.value - other.Float())
	default:
		return newRuntimeError("cannot subtract " + other.Type().String() + " from float")
	}
}

func (v *floatValue) Mul(other Value) Value {
	switch other.Type() {
	case TypeFloat:
		return NewFloat(v.value * other.Float())
	case TypeInt:
		return NewFloat(v.value * other.Float())
	case TypeBigDecimal:
		f := floatToRat(v.value)
		f.Mul(f, other.ToBigDecimal().(*BigDecimalValue).value)
		return NewBigDecimal(f)
	case TypeNull:
		return NewFloat(0.0)
	case TypeBool:
		return NewFloat(v.value * other.Float())
	default:
		return newRuntimeError("cannot multiply float and " + other.Type().String())
	}
}

func (v *floatValue) Div(other Value) Value {
	switch other.Type() {
	case TypeFloat:
		return NewFloat(v.value / other.Float())
	case TypeInt:
		return NewFloat(v.value / other.Float())
	case TypeBigDecimal:
		rat := other.ToBigDecimal().(*BigDecimalValue).value
		if rat.Sign() == 0 {
			return NewFloat(math.Inf(signFloat(v.value)))
		}
		f := floatToRat(v.value)
		f.Quo(f, rat)
		return NewBigDecimal(f)
	case TypeNull:
		return newRuntimeError("division by null")
	case TypeBool:
		return NewFloat(v.value / other.Float())
	default:
		return newRuntimeError("cannot divide float by " + other.Type().String())
	}
}

func (v *floatValue) Mod(other Value) Value {
	switch other.Type() {
	case TypeFloat:
		return NewFloat(math.Mod(v.value, other.Float()))
	case TypeInt:
		return NewFloat(math.Mod(v.value, other.Float()))
	case TypeNull:
		return newRuntimeError("modulo by null")
	case TypeBool:
		return NewFloat(math.Mod(v.value, other.Float()))
	default:
		return newRuntimeError("cannot modulo float and " + other.Type().String())
	}
}

func (v *floatValue) Negate() Value {
	return NewFloat(-v.value)
}

func (v *floatValue) Less(other Value) bool {
	switch other.Type() {
	case TypeFloat:
		return v.value < other.Float()
	case TypeInt:
		return v.value < other.Float()
	case TypeBigDecimal:
		f := floatToRat(v.value)
		return f.Cmp(other.ToBigDecimal().(*BigDecimalValue).value) < 0
	case TypeNull:
		return v.value < 0.0
	case TypeBool:
		return v.value < other.Float()
	default:
		return false
	}
}

func (v *floatValue) Greater(other Value) bool {
	switch other.Type() {
	case TypeFloat:
		return v.value > other.Float()
	case TypeInt:
		return v.value > other.Float()
	case TypeBigDecimal:
		f := floatToRat(v.value)
		return f.Cmp(other.ToBigDecimal().(*BigDecimalValue).value) > 0
	case TypeNull:
		return v.value > 0.0
	case TypeBool:
		return v.value > other.Float()
	default:
		return false
	}
}

func (v *floatValue) LessEqual(other Value) bool    { return !v.Greater(other) }
func (v *floatValue) GreaterEqual(other Value) bool { return !v.Less(other) }

// stringValue 表示字符串值。
//
// 存储 UTF-8 编码的字符串，支持 Unicode 字符。
// 短字符串（≤64 字节）使用内部化优化，重复字符串共享同一实例。
//
// 性能优化：
//   - 短字符串缓存：相同内容共享实例，字符串比较从 O(n) 变为 O(1)
//   - 避免重复分配：频繁使用的字符串（如变量名）自动共享
//
// 支持的操作：
//   - 拼接：+ 运算符
//   - 比较：==、!=、<、>、<=、>=（字典序）
//   - 长度：Len() 返回字节数（UTF-8 编码）
//   - 类型转换：to bool（空字符串为 false）、to array（字符数组）
//
// 注意：字符串是不可变的，修改操作会创建新字符串。
type stringValue struct {
	value string
}

func (v *stringValue) Type() ValueType          { return TypeString }
func (v *stringValue) IsNull() bool             { return false }
func (v *stringValue) Bool() bool               { return v.value != "" }
func (v *stringValue) Int() int64               { return 0 }
func (v *stringValue) Float() float64           { return 0.0 }
func (v *stringValue) String() string           { return v.value }
func (v *stringValue) Array() []Value           { return nil }
func (v *stringValue) Object() map[string]Value { return nil }
func (v *stringValue) Equals(other Value) bool {
	if other.Type() != TypeString {
		return false
	}
	return v.value == other.String()
}
func (v *stringValue) Stringify() string {
	return fmt.Sprintf("%q", v.value)
}
func (v *stringValue) Len() int { return len(v.value) }
func (v *stringValue) ToBigInt() Value {
	return NewBigInt(new(big.Int))
}
func (v *stringValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat))
}

func (v *stringValue) Add(other Value) Value {
	if other.Type() == TypeString {
		return NewString(v.value + other.String())
	}
	return newRuntimeError("cannot add string and " + other.Type().String())
}

func (v *stringValue) Sub(other Value) Value {
	return newRuntimeError("cannot subtract from string")
}

func (v *stringValue) Mul(other Value) Value {
	return newRuntimeError("cannot multiply string")
}

func (v *stringValue) Div(other Value) Value {
	return newRuntimeError("cannot divide string")
}

func (v *stringValue) Mod(other Value) Value {
	return newRuntimeError("cannot modulo string")
}

func (v *stringValue) Negate() Value {
	return newRuntimeError("cannot negate string")
}

func (v *stringValue) Less(other Value) bool {
	if other.Type() == TypeString {
		return v.value < other.String()
	}
	return false
}

func (v *stringValue) Greater(other Value) bool {
	if other.Type() == TypeString {
		return v.value > other.String()
	}
	return false
}

func (v *stringValue) LessEqual(other Value) bool {
	if other.Type() == TypeString {
		return v.value <= other.String()
	}
	return false
}

func (v *stringValue) GreaterEqual(other Value) bool {
	if other.Type() == TypeString {
		return v.value >= other.String()
	}
	return false
}

// arrayValue 表示数组值（有序集合）。
//
// 数组存储任意类型的 Value 列表，支持索引访问（从 0 开始）。
// 数组是引用类型，支持 GC 管理（可选）。
//
// 创建方式：
//   - 字面量：[1, 2, 3]
//   - 构造函数：NewArray() / NewArrayGC()
//   - 内置函数：array()、range()、split() 等
//
// 支持的运算：
//   - 索引访问：arr[0]、arr[$i]
//   - 长度：Len()、# 运算符
//   - 拼接：+ 运算符（合并数组）
//   - 比较：==、!=（深度比较）
//   - 遍历：for...of 循环
//
// 数组方法（内置）：
//   - push、pop、shift、unshift、splice
//   - indexOf、lastIndexOf、slice、includes
//   - map、filter、reduce、find、sort
//
// GC 支持：
//
//	使用 NewArrayGC() 创建的数组会被 GC 跟踪，循环引用会自动回收。
type arrayValue struct {
	value []Value // 底层切片
	// GC 字段
	gcID     uint64 // GC 对象 ID
	refCount int    // 引用计数
	gcPtr    *gc.GC // GC 管理器指针
	alive    bool   // 对象是否存活
}

func (v *arrayValue) Type() ValueType          { return TypeArray }
func (v *arrayValue) IsNull() bool             { return false }
func (v *arrayValue) Bool() bool               { return len(v.value) > 0 }
func (v *arrayValue) Int() int64               { return int64(len(v.value)) }
func (v *arrayValue) Float() float64           { return float64(len(v.value)) }
func (v *arrayValue) String() string           { return v.Stringify() }
func (v *arrayValue) Array() []Value           { return v.value }
func (v *arrayValue) Object() map[string]Value { return nil }
func (v *arrayValue) Equals(other Value) bool {
	if other.Type() != TypeArray {
		return false
	}
	otherArr := other.Array()
	if len(v.value) != len(otherArr) {
		return false
	}
	for i, val := range v.value {
		if !val.Equals(otherArr[i]) {
			return false
		}
	}
	return true
}
func (v *arrayValue) Stringify() string {
	result := "["
	for i, val := range v.value {
		if i > 0 {
			result += ", "
		}
		result += val.Stringify()
	}
	result += "]"
	return result
}
func (v *arrayValue) Len() int { return len(v.value) }
func (v *arrayValue) ToBigInt() Value {
	return NewBigInt(big.NewInt(int64(len(v.value))))
}
func (v *arrayValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat).SetInt64(int64(len(v.value))))
}

func (v *arrayValue) Add(other Value) Value {
	if other.Type() == TypeArray {
		result := make([]Value, len(v.value)+len(other.Array()))
		copy(result, v.value)
		copy(result[len(v.value):], other.Array())
		return NewArray(result)
	}
	return newRuntimeError("cannot add array and " + other.Type().String())
}

func (v *arrayValue) Sub(other Value) Value {
	return newRuntimeError("cannot subtract from array")
}

func (v *arrayValue) Mul(other Value) Value {
	return newRuntimeError("cannot multiply array")
}

func (v *arrayValue) Div(other Value) Value {
	return newRuntimeError("cannot divide array")
}

func (v *arrayValue) Mod(other Value) Value {
	return newRuntimeError("cannot modulo array")
}

func (v *arrayValue) Negate() Value {
	return newRuntimeError("cannot negate array")
}

func (v *arrayValue) Less(other Value) bool    { return false }
func (v *arrayValue) Greater(other Value) bool { return false }
func (v *arrayValue) LessEqual(other Value) bool {
	return v.Equals(other)
}
func (v *arrayValue) GreaterEqual(other Value) bool {
	return v.Equals(other)
}

// arrayValue 实现 gc.ManagedObject 接口
func (v *arrayValue) ObjID() uint64    { return v.gcID }
func (v *arrayValue) GetRefCount() int { return v.refCount }
func (v *arrayValue) IsAlive() bool    { return v.alive }
func (v *arrayValue) OnFree()          { v.value = nil }

func (v *arrayValue) IncRef() {
	v.refCount++
}

func (v *arrayValue) DecRef() {
	v.refCount--
	if v.refCount <= 0 {
		v.alive = false
	}
}

func (v *arrayValue) MarkChildren(marker func(child any)) {
	for _, elem := range v.value {
		marker(elem)
	}
}

// SetupGC 设置 GC 并注册到 GC
func (v *arrayValue) SetupGC(g *gc.GC) {
	if g == nil {
		return
	}
	v.gcPtr = g
	v.gcID = g.NextID()
	v.alive = true
	v.refCount = 1
	g.Register(v)
}

// rangeValue 表示范围值（用于迭代）。
//
// 范围可以是半开区间 [start, end) 或闭区间 [start, end]。
// 用于 for...in 循环和 range() 函数。
//
// 创建方式：
//   - 范围语法：1...10, 1..=10
//   - range() 函数
//
// 遍历：
//   - for i in 1...10 { ... }  // 1,2,3,4,5,6,7,8,9
//   - for i in 1..=10 { ... } // 1,2,3,4,5,6,7,8,9,10
type rangeValue struct {
	start     int64 // 起始值
	end       int64 // 结束值
	inclusive bool  // true 表示闭区间
}

// Type 实现 Value 接口
func (v *rangeValue) Type() ValueType          { return TypeRange }
func (v *rangeValue) IsNull() bool             { return false }
func (v *rangeValue) Bool() bool               { return v.start <= v.end || v.inclusive && v.start == v.end }
func (v *rangeValue) Int() int64               { return v.end - v.start + 1 }
func (v *rangeValue) Float() float64           { return float64(v.Int()) }
func (v *rangeValue) String() string           { return v.Stringify() }
func (v *rangeValue) Array() []Value           { return nil }
func (v *rangeValue) Object() map[string]Value { return nil }

func (v *rangeValue) Equals(other Value) bool {
	if other.Type() != TypeRange {
		return false
	}
	o := other.(*rangeValue)
	return v.start == o.start && v.end == o.end && v.inclusive == o.inclusive
}

func (v *rangeValue) Stringify() string {
	if v.inclusive {
		return fmt.Sprintf("%d..=%d", v.start, v.end)
	}
	return fmt.Sprintf("%d...%d", v.start, v.end)
}

func (v *rangeValue) Len() int { return int(v.Int()) }

func (v *rangeValue) Start() int64      { return v.start }
func (v *rangeValue) End() int64        { return v.end }
func (v *rangeValue) IsInclusive() bool { return v.inclusive }

func (v *rangeValue) Add(other Value) Value         { return newRuntimeError("cannot add range") }
func (v *rangeValue) Sub(other Value) Value         { return newRuntimeError("cannot subtract from range") }
func (v *rangeValue) Mul(other Value) Value         { return newRuntimeError("cannot multiply range") }
func (v *rangeValue) Div(other Value) Value         { return newRuntimeError("cannot divide range") }
func (v *rangeValue) Mod(other Value) Value         { return newRuntimeError("cannot modulo range") }
func (v *rangeValue) ToBigInt() Value               { return NewBigInt(big.NewInt(v.Int())) }
func (v *rangeValue) ToBigDecimal() Value           { return NewBigDecimal(new(big.Rat).SetInt64(v.Int())) }
func (v *rangeValue) Negate() Value                 { return newRuntimeError("cannot negate range") }
func (v *rangeValue) Less(other Value) bool         { return false }
func (v *rangeValue) Greater(other Value) bool      { return false }
func (v *rangeValue) LessEqual(other Value) bool    { return true }
func (v *rangeValue) GreaterEqual(other Value) bool { return true }

// regexValue 表示正则表达式值。
//
// 正则字面量 #/pattern/flags# 创建此类型，底层使用 Go regexp.Regexp。
//
// 支持的操作：
//   - =~ 匹配运算符：$str =~ #/pattern/#
//   - match/case 模式匹配
//   - .test() / .match() 等方法调用
type regexValue struct {
	pattern string         // 原始模式字符串
	flags   string         // flags 字符串（imsU）
	re      *regexp.Regexp // 编译后的 Go 正则
}

func (v *regexValue) Type() ValueType          { return TypeRegex }
func (v *regexValue) IsNull() bool             { return false }
func (v *regexValue) Bool() bool               { return true }
func (v *regexValue) Int() int64               { return 0 }
func (v *regexValue) Float() float64           { return 0 }
func (v *regexValue) String() string           { return fmt.Sprintf("#/%s/%s#", v.pattern, v.flags) }
func (v *regexValue) Array() []Value           { return nil }
func (v *regexValue) Object() map[string]Value { return nil }
func (v *regexValue) Len() int                 { return 0 }

func (v *regexValue) Equals(other Value) bool {
	if other.Type() != TypeRegex {
		return false
	}
	o := other.(*regexValue)
	return v.pattern == o.pattern && v.flags == o.flags
}

func (v *regexValue) Stringify() string {
	return fmt.Sprintf("#/%s/%s#", v.pattern, v.flags)
}

func (v *regexValue) ToBigInt() Value               { return NewBigInt(big.NewInt(0)) }
func (v *regexValue) ToBigDecimal() Value           { return NewBigDecimal(new(big.Rat).SetInt64(0)) }
func (v *regexValue) Add(other Value) Value         { return newRuntimeError("cannot add regex") }
func (v *regexValue) Sub(other Value) Value         { return newRuntimeError("cannot subtract from regex") }
func (v *regexValue) Mul(other Value) Value         { return newRuntimeError("cannot multiply regex") }
func (v *regexValue) Div(other Value) Value         { return newRuntimeError("cannot divide regex") }
func (v *regexValue) Mod(other Value) Value         { return newRuntimeError("cannot modulo regex") }
func (v *regexValue) Negate() Value                 { return newRuntimeError("cannot negate regex") }
func (v *regexValue) Less(other Value) bool         { return false }
func (v *regexValue) Greater(other Value) bool      { return false }
func (v *regexValue) LessEqual(other Value) bool    { return true }
func (v *regexValue) GreaterEqual(other Value) bool { return true }

// Pattern 返回原始模式字符串
func (v *regexValue) Pattern() string { return v.pattern }

// Flags 返回 flags 字符串
func (v *regexValue) Flags() string { return v.flags }

// Regexp 返回编译后的 Go regexp.Regexp
func (v *regexValue) Regexp() *regexp.Regexp { return v.re }

// Match 检查字符串是否匹配正则（子串匹配）
func (v *regexValue) Match(s string) bool {
	return v.re.MatchString(s)
}

// FindStringSubmatch 返回捕获组匹配结果
func (v *regexValue) FindStringSubmatch(s string) []string {
	return v.re.FindStringSubmatch(s)
}

// NewRegex 创建正则表达式值
func NewRegex(pattern, flags string) (Value, error) {
	// 构建 Go regexp 的 inline flags
	goFlags := ""
	if containsFlag(flags, 'i') {
		goFlags += "(?i)"
	}
	if containsFlag(flags, 'm') {
		goFlags += "(?m)"
	}
	if containsFlag(flags, 's') {
		goFlags += "(?s)"
	}
	if containsFlag(flags, 'U') {
		goFlags += "(?U)"
	}

	fullPattern := goFlags + pattern
	re, err := regexp.Compile(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %v", err)
	}

	return &regexValue{
		pattern: pattern,
		flags:   flags,
		re:      re,
	}, nil
}

// NewRegexUnsafe 创建正则表达式值（已验证，不检查错误）
func NewRegexUnsafe(pattern, flags string) *regexValue {
	goFlags := ""
	if containsFlag(flags, 'i') {
		goFlags += "(?i)"
	}
	if containsFlag(flags, 'm') {
		goFlags += "(?m)"
	}
	if containsFlag(flags, 's') {
		goFlags += "(?s)"
	}
	if containsFlag(flags, 'U') {
		goFlags += "(?U)"
	}

	fullPattern := goFlags + pattern
	re := regexp.MustCompile(fullPattern)

	return &regexValue{
		pattern: pattern,
		flags:   flags,
		re:      re,
	}
}

func containsFlag(flags string, f byte) bool {
	for i := 0; i < len(flags); i++ {
		if flags[i] == f {
			return true
		}
	}
	return false
}

// IsRegex 检查值是否为正则表达式类型
func IsRegex(v Value) bool {
	_, ok := v.(*regexValue)
	return ok
}

// objectValue 表示对象值（键值对映射）。
//
// 对象存储 string 到 Value 的无序映射，类似 JavaScript 的对象或 PHP 的关联数组。
// 对象是引用类型，支持 GC 管理（可选）。
//
// 创建方式：
//   - 字面量：{name: "Alice", age: 30}
//   - 构造函数：NewObject() / NewObjectGC()
//   - 内置函数：object()、json_decode() 等
//
// 支持的运算：
//   - 属性访问：obj.name、obj["name"]
//   - 动态访问：obj[$key]
//   - 长度：Len()（键值对数量）
//   - 比较：==、!=（深度比较）
//   - 遍历：for...in 循环（遍历键）
//
// 对象方法（内置）：
//   - keys()、values()、has()、delete()
//   - merge()（合并对象）
//
// GC 支持：
//
//	使用 NewObjectGC() 创建的对象会被 GC 跟踪，循环引用会自动回收。
//
// 注意：键必须是字符串类型，非字符串会被自动转换。
type objectValue struct {
	value map[string]Value // 底层映射
	// GC 字段
	gcID     uint64 // GC 对象 ID
	refCount int    // 引用计数
	gcPtr    *gc.GC // GC 管理器指针
	alive    bool   // 对象是否存活
}

func (v *objectValue) Type() ValueType          { return TypeObject }
func (v *objectValue) IsNull() bool             { return false }
func (v *objectValue) Bool() bool               { return len(v.value) > 0 }
func (v *objectValue) Int() int64               { return int64(len(v.value)) }
func (v *objectValue) Float() float64           { return float64(len(v.value)) }
func (v *objectValue) String() string           { return v.Stringify() }
func (v *objectValue) Array() []Value           { return nil }
func (v *objectValue) Object() map[string]Value { return v.value }
func (v *objectValue) Equals(other Value) bool {
	if other.Type() != TypeObject {
		return false
	}
	otherObj := other.Object()
	if len(v.value) != len(otherObj) {
		return false
	}
	for key, val := range v.value {
		otherVal, exists := otherObj[key]
		if !exists || !val.Equals(otherVal) {
			return false
		}
	}
	return true
}
func (v *objectValue) Stringify() string {
	result := "{"
	first := true
	for key, val := range v.value {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%q: %s", key, val.Stringify())
		first = false
	}
	result += "}"
	return result
}
func (v *objectValue) Len() int { return len(v.value) }
func (v *objectValue) ToBigInt() Value {
	return NewBigInt(big.NewInt(int64(len(v.value))))
}
func (v *objectValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat).SetInt64(int64(len(v.value))))
}

func (v *objectValue) Add(other Value) Value {
	if other.Type() == TypeObject {
		result := make(map[string]Value, len(v.value)+len(other.Object()))
		maps.Copy(result, v.value)
		maps.Copy(result, other.Object())
		return NewObject(result)
	}
	return newRuntimeError("cannot add object and " + other.Type().String())
}

func (v *objectValue) Sub(other Value) Value {
	return newRuntimeError("cannot subtract from object")
}

func (v *objectValue) Mul(other Value) Value {
	return newRuntimeError("cannot multiply object")
}

func (v *objectValue) Div(other Value) Value {
	return newRuntimeError("cannot divide object")
}

func (v *objectValue) Mod(other Value) Value {
	return newRuntimeError("cannot modulo object")
}

func (v *objectValue) Negate() Value {
	return newRuntimeError("cannot negate object")
}

func (v *objectValue) Less(other Value) bool    { return false }
func (v *objectValue) Greater(other Value) bool { return false }
func (v *objectValue) LessEqual(other Value) bool {
	return v.Equals(other)
}
func (v *objectValue) GreaterEqual(other Value) bool {
	return v.Equals(other)
}

// objectValue 实现 gc.ManagedObject 接口
func (v *objectValue) ObjID() uint64    { return v.gcID }
func (v *objectValue) GetRefCount() int { return v.refCount }
func (v *objectValue) IsAlive() bool    { return v.alive }
func (v *objectValue) OnFree()          { v.value = nil }

func (v *objectValue) IncRef() {
	v.refCount++
}

func (v *objectValue) DecRef() {
	v.refCount--
	if v.refCount <= 0 {
		v.alive = false
	}
}

func (v *objectValue) MarkChildren(marker func(child any)) {
	for _, elem := range v.value {
		marker(elem)
	}
}

// SetupGC 设置 GC 并注册到 GC
func (v *objectValue) SetupGC(g *gc.GC) {
	if g == nil {
		return
	}
	v.gcPtr = g
	v.gcID = g.NextID()
	v.alive = true
	v.refCount = 1
	g.Register(v)
}

// funcValue 表示函数值
type funcValue struct {
	name string
	fn   GoFunction
}

func (v *funcValue) Type() ValueType          { return TypeFunc }
func (v *funcValue) IsNull() bool             { return false }
func (v *funcValue) Bool() bool               { return true }
func (v *funcValue) Int() int64               { return 0 }
func (v *funcValue) Float() float64           { return 0.0 }
func (v *funcValue) String() string           { return fmt.Sprintf("function(%s)", v.name) }
func (v *funcValue) Array() []Value           { return nil }
func (v *funcValue) Object() map[string]Value { return nil }
func (v *funcValue) Equals(other Value) bool {
	if other.Type() != TypeFunc {
		return false
	}
	return v.name == other.String()
}
func (v *funcValue) Stringify() string {
	return fmt.Sprintf("function(%s)", v.name)
}
func (v *funcValue) Len() int { return 0 }
func (v *funcValue) ToBigInt() Value {
	return NewBigInt(new(big.Int))
}
func (v *funcValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat))
}

func (v *funcValue) Add(other Value) Value {
	return newRuntimeError("cannot add function")
}
func (v *funcValue) Sub(other Value) Value {
	return newRuntimeError("cannot subtract from function")
}
func (v *funcValue) Mul(other Value) Value {
	return newRuntimeError("cannot multiply function")
}
func (v *funcValue) Div(other Value) Value {
	return newRuntimeError("cannot divide function")
}
func (v *funcValue) Mod(other Value) Value {
	return newRuntimeError("cannot modulo function")
}
func (v *funcValue) Negate() Value {
	return newRuntimeError("cannot negate function")
}

func (v *funcValue) Less(other Value) bool    { return false }
func (v *funcValue) Greater(other Value) bool { return false }
func (v *funcValue) LessEqual(other Value) bool {
	return v.Equals(other)
}
func (v *funcValue) GreaterEqual(other Value) bool {
	return v.Equals(other)
}

// BigIntValue 表示大整数值
type BigIntValue struct {
	value *big.Int
}

// BigInt 返回底层 *big.Int 值
func (v *BigIntValue) BigInt() *big.Int { return v.value }

func (v *BigIntValue) Type() ValueType          { return TypeBigInt }
func (v *BigIntValue) IsNull() bool             { return false }
func (v *BigIntValue) Bool() bool               { return v.value.Sign() != 0 }
func (v *BigIntValue) Int() int64               { return v.value.Int64() }
func (v *BigIntValue) Float() float64           { f, _ := new(big.Float).SetInt(v.value).Float64(); return f }
func (v *BigIntValue) String() string           { return v.value.String() }
func (v *BigIntValue) Array() []Value           { return nil }
func (v *BigIntValue) Object() map[string]Value { return nil }
func (v *BigIntValue) Equals(other Value) bool {
	switch o := other.(type) {
	case *BigIntValue:
		return v.value.Cmp(o.value) == 0
	case *intValue:
		return v.value.Int64() == o.value
	case *floatValue:
		f, _ := new(big.Float).SetInt(v.value).Float64()
		return f == o.value
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		return r.Cmp(o.value) == 0
	default:
		return false
	}
}
func (v *BigIntValue) Stringify() string {
	return v.value.String()
}
func (v *BigIntValue) Len() int { return 0 }
func (v *BigIntValue) ToBigInt() Value {
	return v
}
func (v *BigIntValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat).SetInt(v.value))
}

func (v *BigIntValue) Add(other Value) Value {
	switch o := other.(type) {
	case *BigIntValue:
		return NewBigInt(new(big.Int).Add(v.value, o.value))
	case *intValue:
		return NewBigInt(new(big.Int).Add(v.value, big.NewInt(o.value)))
	case *floatValue:
		return NewFloat(v.Float() + o.value)
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		r.Add(r, o.value)
		return NewBigDecimal(r)
	default:
		return NewInt(v.Int()).Add(other)
	}
}

func (v *BigIntValue) Sub(other Value) Value {
	switch o := other.(type) {
	case *BigIntValue:
		return NewBigInt(new(big.Int).Sub(v.value, o.value))
	case *intValue:
		return NewBigInt(new(big.Int).Sub(v.value, big.NewInt(o.value)))
	case *floatValue:
		return NewFloat(v.Float() - o.value)
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		r.Sub(r, o.value)
		return NewBigDecimal(r)
	default:
		return NewInt(v.Int()).Sub(other)
	}
}

func (v *BigIntValue) Mul(other Value) Value {
	switch o := other.(type) {
	case *BigIntValue:
		return NewBigInt(new(big.Int).Mul(v.value, o.value))
	case *intValue:
		return NewBigInt(new(big.Int).Mul(v.value, big.NewInt(o.value)))
	case *floatValue:
		return NewFloat(v.Float() * o.value)
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		r.Mul(r, o.value)
		return NewBigDecimal(r)
	default:
		return NewInt(v.Int()).Mul(other)
	}
}

func (v *BigIntValue) Div(other Value) Value {
	switch o := other.(type) {
	case *BigIntValue:
		if o.value.Sign() == 0 {
			return NewFloat(math.Inf(v.value.Sign()))
		}
		return NewFloat(v.Float() / o.Float())
	case *intValue:
		if o.value == 0 {
			return NewFloat(math.Inf(sign(v.Int())))
		}
		return NewFloat(v.Float() / o.Float())
	case *floatValue:
		return NewFloat(v.Float() / o.value)
	case *BigDecimalValue:
		if o.value.Sign() == 0 {
			return NewFloat(math.Inf(v.value.Sign()))
		}
		r := new(big.Rat).SetInt(v.value)
		r.Quo(r, o.value)
		return NewBigDecimal(r)
	default:
		return NewInt(v.Int()).Div(other)
	}
}

func (v *BigIntValue) Mod(other Value) Value {
	switch o := other.(type) {
	case *BigIntValue:
		if o.value.Sign() == 0 {
			return NewFloat(math.NaN())
		}
		return NewBigInt(new(big.Int).Mod(v.value, o.value))
	case *intValue:
		if o.value == 0 {
			return NewFloat(math.NaN())
		}
		return NewBigInt(new(big.Int).Mod(v.value, big.NewInt(o.value)))
	default:
		return NewInt(v.Int()).Mod(other)
	}
}

func (v *BigIntValue) Negate() Value {
	return NewBigInt(new(big.Int).Neg(v.value))
}

func (v *BigIntValue) Less(other Value) bool {
	switch o := other.(type) {
	case *BigIntValue:
		return v.value.Cmp(o.value) < 0
	case *intValue:
		return v.value.Cmp(big.NewInt(o.value)) < 0
	case *floatValue:
		return v.Float() < o.value
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		return r.Cmp(o.value) < 0
	default:
		return NewInt(v.Int()).Less(other)
	}
}

func (v *BigIntValue) Greater(other Value) bool {
	switch o := other.(type) {
	case *BigIntValue:
		return v.value.Cmp(o.value) > 0
	case *intValue:
		return v.value.Cmp(big.NewInt(o.value)) > 0
	case *floatValue:
		return v.Float() > o.value
	case *BigDecimalValue:
		r := new(big.Rat).SetInt(v.value)
		return r.Cmp(o.value) > 0
	default:
		return NewInt(v.Int()).Greater(other)
	}
}

func (v *BigIntValue) LessEqual(other Value) bool    { return !v.Greater(other) }
func (v *BigIntValue) GreaterEqual(other Value) bool { return !v.Less(other) }

// BigDecimalValue 表示大数值（有理数精度）
type BigDecimalValue struct {
	value *big.Rat
}

// BigRat 返回底层 *big.Rat 值
func (v *BigDecimalValue) BigRat() *big.Rat { return v.value }

func (v *BigDecimalValue) Type() ValueType          { return TypeBigDecimal }
func (v *BigDecimalValue) IsNull() bool             { return false }
func (v *BigDecimalValue) Bool() bool               { return v.value.Sign() != 0 }
func (v *BigDecimalValue) Int() int64               { f, _ := v.value.Float64(); return int64(f) }
func (v *BigDecimalValue) Float() float64           { f, _ := v.value.Float64(); return f }
func (v *BigDecimalValue) String() string           { return v.value.RatString() }
func (v *BigDecimalValue) Array() []Value           { return nil }
func (v *BigDecimalValue) Object() map[string]Value { return nil }
func (v *BigDecimalValue) Equals(other Value) bool {
	o := other.ToBigDecimal()
	return v.value.Cmp(o.(*BigDecimalValue).value) == 0
}
func (v *BigDecimalValue) Stringify() string {
	if v.value.IsInt() {
		return v.value.Num().String()
	}
	f, _ := v.value.Float64()
	return fmt.Sprintf("%g", f)
}
func (v *BigDecimalValue) Len() int { return 0 }
func (v *BigDecimalValue) ToBigInt() Value {
	n := new(big.Int).Div(v.value.Num(), v.value.Denom())
	return NewBigInt(n)
}
func (v *BigDecimalValue) ToBigDecimal() Value {
	return v
}

func (v *BigDecimalValue) Add(other Value) Value {
	r := new(big.Rat).Set(v.value)
	r.Add(r, other.ToBigDecimal().(*BigDecimalValue).value)
	return NewBigDecimal(r)
}

func (v *BigDecimalValue) Sub(other Value) Value {
	r := new(big.Rat).Set(v.value)
	r.Sub(r, other.ToBigDecimal().(*BigDecimalValue).value)
	return NewBigDecimal(r)
}

func (v *BigDecimalValue) Mul(other Value) Value {
	r := new(big.Rat).Set(v.value)
	r.Mul(r, other.ToBigDecimal().(*BigDecimalValue).value)
	return NewBigDecimal(r)
}

func (v *BigDecimalValue) Div(other Value) Value {
	rat := other.ToBigDecimal().(*BigDecimalValue).value
	if rat.Sign() == 0 {
		return NewFloat(math.Inf(v.value.Sign()))
	}
	r := new(big.Rat).Set(v.value)
	r.Quo(r, rat)
	return NewBigDecimal(r)
}

func (v *BigDecimalValue) Mod(other Value) Value {
	return NewInt(v.Int()).Mod(other)
}

func (v *BigDecimalValue) Negate() Value {
	return NewBigDecimal(new(big.Rat).Neg(v.value))
}

func (v *BigDecimalValue) Less(other Value) bool {
	o := other.ToBigDecimal()
	return v.value.Cmp(o.(*BigDecimalValue).value) < 0
}

func (v *BigDecimalValue) Greater(other Value) bool {
	o := other.ToBigDecimal()
	return v.value.Cmp(o.(*BigDecimalValue).value) > 0
}

func (v *BigDecimalValue) LessEqual(other Value) bool    { return !v.Greater(other) }
func (v *BigDecimalValue) GreaterEqual(other Value) bool { return !v.Less(other) }

// ============================================================================
// Error 类型
// ============================================================================

type errorValue struct {
	message string
	code    int64
	errType string
}

func (v *errorValue) Type() ValueType { return TypeError }
func (v *errorValue) IsNull() bool    { return false }
func (v *errorValue) Bool() bool      { return true }
func (v *errorValue) Int() int64      { return v.code }
func (v *errorValue) Float() float64  { return float64(v.code) }
func (v *errorValue) String() string {
	if v.code != 0 {
		return fmt.Sprintf("%s: %s (code: %d)", v.errType, v.message, v.code)
	}
	return fmt.Sprintf("%s: %s", v.errType, v.message)
}
func (v *errorValue) Array() []Value           { return nil }
func (v *errorValue) Object() map[string]Value { return v.fields() }
func (v *errorValue) Len() int                 { return 0 }
func (v *errorValue) Equals(other Value) bool {
	if other.Type() != TypeError {
		return false
	}
	oe := other.(*errorValue)
	return v.message == oe.message && v.code == oe.code && v.errType == oe.errType
}
func (v *errorValue) Stringify() string {
	return fmt.Sprintf("{\"type\": %q, \"code\": %d, \"message\": %q}", v.errType, v.code, v.message)
}
func (v *errorValue) ToBigInt() Value {
	return NewBigInt(big.NewInt(v.code))
}
func (v *errorValue) ToBigDecimal() Value {
	return NewBigDecimal(new(big.Rat).SetInt64(v.code))
}

func (v *errorValue) fields() map[string]Value {
	return map[string]Value{
		"message": NewString(v.message),
		"code":    NewInt(v.code),
		"type":    NewString(v.errType),
	}
}

func (v *errorValue) Add(other Value) Value         { return newRuntimeError("cannot add error") }
func (v *errorValue) Sub(other Value) Value         { return newRuntimeError("cannot subtract error") }
func (v *errorValue) Mul(other Value) Value         { return newRuntimeError("cannot multiply error") }
func (v *errorValue) Div(other Value) Value         { return newRuntimeError("cannot divide error") }
func (v *errorValue) Mod(other Value) Value         { return newRuntimeError("cannot modulo error") }
func (v *errorValue) Negate() Value                 { return newRuntimeError("cannot negate error") }
func (v *errorValue) Less(other Value) bool         { return false }
func (v *errorValue) Greater(other Value) bool      { return false }
func (v *errorValue) LessEqual(other Value) bool    { return false }
func (v *errorValue) GreaterEqual(other Value) bool { return false }

// IsError 检查 Value 是否为 error 类型。
//
// 此方法快速检查 Value 的类型是否为 TypeError，用于错误处理和条件捕获。
//
// 参数：
//   - v: 要检查的 Value
//
// 返回值：
//   - true: 值是 error 类型
//   - false: 值不是 error 类型
//
// 使用示例：
//
//	err := engine.NewError("something went wrong", 500, "Error")
//	if engine.IsError(err) {
//	    fmt.Println("这是一个错误值")
//	}
//
//	// 在脚本中
//	// catch ($e) {
//	//     if (is_error($e)) { ... }
//	// }
func IsError(v Value) bool {
	return v.Type() == TypeError
}

// GetErrorField 获取 error 值的指定字段。
//
// 此方法从 errorValue 中提取特定字段的值，支持以下字段：
//   - "message": 错误消息字符串
//   - "code": 错误码（int64）
//   - "type": 错误类型字符串
//
// 用于错误处理和自定义错误展示。
//
// 参数：
//   - v: error 类型的 Value
//   - field: 字段名（"message"、"code" 或 "type"）
//
// 返回值：
//   - Value: 字段的值
//   - true: 字段存在且获取成功
//   - nil, false: 值不是 error 类型或字段不存在
//
// 使用示例：
//
//	err := engine.NewError("not found", 404, "NotFoundError")
//
//	if msg, ok := engine.GetErrorField(err, "message"); ok {
//	    fmt.Printf("错误: %s\n", msg.String()) // 输出: 错误: not found
//	}
//
//	if code, ok := engine.GetErrorField(err, "code"); ok {
//	    fmt.Printf("错误码: %d\n", code.Int()) // 输出: 错误码: 404
//	}
//
//	// 在脚本中
//	// catch ($e when $e.code == 404) { ... }
func GetErrorField(v Value, field string) (Value, bool) {
	if v.Type() != TypeError {
		return nil, false
	}
	ev := v.(*errorValue)
	switch field {
	case "message":
		return NewString(ev.message), true
	case "code":
		return NewInt(ev.code), true
	case "type":
		return NewString(ev.errType), true
	}
	return nil, false
}

// 工厂函数

// NewNull 创建一个 null 值。
//
// 此方法返回全局 null 单例，不分配新内存。
// null 表示空值或不存在的值，在脚本中对应 null 关键字。
//
// 性能优化：使用全局单例避免重复分配。
//
// 返回值：
//   - Value: null 值（单例）
//
// 使用示例：
//
//	nullVal := engine.NewNull()
//	fmt.Println(nullVal.Type()) // 输出: null
//	fmt.Println(nullVal.IsNull()) // 输出: true
func NewNull() Value {
	return nullSingleton
}

// NewBool 创建布尔值。
//
// 此方法返回全局布尔单例（true 或 false），不分配新内存。
// 在脚本中对应 true/false 关键字。
//
// 性能优化：使用两个全局单例（boolTrueValue/boolFalseValue）
// 避免重复分配，所有布尔值共享同一实例。
//
// 参数：
//   - v: 布尔值 true 或 false
//
// 返回值：
//   - Value: 布尔值（单例）
//
// 使用示例：
//
//	trueVal := engine.NewBool(true)
//	falseVal := engine.NewBool(false)
//	fmt.Println(trueVal.Bool())  // 输出: true
//	fmt.Println(falseVal.Bool()) // 输出: false
func NewBool(v bool) Value {
	if v {
		return boolTrueValue
	}
	return boolFalseValue
}

// NewInt 创建整数值。
//
// 此方法创建 int64 类型的整数值。
// 对于小整数 [-256, 1024] 范围，使用预分配缓存，无需内存分配。
// 超出范围的整数会创建新的 intValue 实例。
//
// 性能优化：小整数缓存消除了频繁创建整数对象的 GC 压力。
//
// 参数：
//   - v: int64 整数值
//
// 返回值：
//   - Value: 整数值
//
// 使用示例：
//
//	// 小整数（缓存）
//	one := engine.NewInt(1)
//	hundred := engine.NewInt(100)
//
//	// 大整数（新建实例）
//	billion := engine.NewInt(1000000000)
//
//	fmt.Println(one.Int())       // 输出: 1
//	fmt.Println(hundred.Int())   // 输出: 100
//	fmt.Println(billion.Int())   // 输出: 1000000000
func NewInt(v int64) Value {
	if v >= smallIntMin && v <= smallIntMax {
		return smallIntCache[v-smallIntMin]
	}
	return &intValue{value: v}
}

// NewFloat 创建浮点数值。
//
// 此方法创建 float64 类型的浮点数值。
// 浮点值没有缓存机制，每次调用都会创建新的 floatValue 实例。
//
// IEEE 754 语义：
//   - 支持特殊值：Inf、-Inf、NaN
//   - 除零操作返回 Inf 或 NaN（遵循 IEEE 754 标准）
//
// 参数：
//   - v: float64 浮点值
//
// 返回值：
//   - Value: 浮点值
//
// 使用示例：
//
//	pi := engine.NewFloat(3.14159)
//	inf := engine.NewFloat(math.Inf(1))
//	nan := engine.NewFloat(math.NaN())
//
//	fmt.Println(pi.Float())  // 输出: 3.14159
//	fmt.Println(inf.Float()) // 输出: +Inf
//	fmt.Println(math.IsNaN(nan.Float())) // 输出: true
func NewFloat(v float64) Value {
	return &floatValue{value: v}
}

// NewString 创建字符串值。
//
// 此方法创建 string 类型的字符串值。
// 对于短字符串（≤64 字节），使用内部化（interning）优化，
// 相同内容的字符串共享同一实例，字符串比较从 O(n) 变为 O(1)（指针比较）。
//
// 性能优化：
//   - 短字符串缓存：消除重复字符串的内存开销
//   - 快速比较：相同字符串指针必然相等
//
// 参数：
//   - v: 字符串值
//
// 返回值：
//   - Value: 字符串值（可能共享实例）
//
// 使用示例：
//
//	// 短字符串（内部化缓存）
//	s1 := engine.NewString("hello")
//	s2 := engine.NewString("hello")
//	fmt.Println(s1 == s2) // 输出: true（同一实例）
//
//	// 长字符串（新建实例）
//	long := engine.NewString(strings.Repeat("a", 100))
//
//	fmt.Println(s1.String()) // 输出: hello
func NewString(v string) Value {
	// 仅缓存较短的字符串，避免内存爆炸
	if len(v) > stringInternMaxLen {
		return &stringValue{value: v}
	}

	// 尝试从缓存池获取
	if cached, ok := stringInternPool.Load(v); ok {
		return cached.(*stringValue)
	}

	// 创建新实例并存入缓存池
	sv := &stringValue{value: v}
	stringInternPool.Store(v, sv)
	return sv
}

// NewArray 创建数组值。
//
// 此方法创建 array 类型的数组值，存储任意类型的 Value 列表。
// 数组是有序集合，支持索引访问（从 0 开始）。
//
// 注意：此方法创建的数组不支持自动 GC 管理。
// 如果需要 GC 支持，使用 NewArrayGC()。
//
// 参数：
//   - v: Value 切片（可为 nil 或空切片）
//
// 返回值：
//   - Value: 数组值
//
// 使用示例：
//
//	// 创建数组
//	arr := engine.NewArray([]engine.Value{
//	    engine.NewInt(1),
//	    engine.NewInt(2),
//	    engine.NewInt(3),
//	})
//
//	// 获取数组元素
//	first := arr.Array()[0]
//	fmt.Println(first.Int()) // 输出: 1
//
//	// 获取数组长度
//	fmt.Println(arr.Len())   // 输出: 3
func NewArray(v []Value) Value {
	return &arrayValue{value: v}
}

// NewRange 创建范围值。
//
// 此方法创建 range 类型的范围值，用于迭代。
// 范围可以是半开区间 [start, end) 或闭区间 [start, end]。
//
// 参数：
//   - start: 起始值
//   - end: 结束值
//   - inclusive: true 表示闭区间 (..=)，false 表示半开区间 (...)
//
// 返回值：
//   - Value: 范围值
func NewRange(start, end int64, inclusive bool) Value {
	return &rangeValue{start: start, end: end, inclusive: inclusive}
}

// NewObject 创建对象值。
//
// 此方法创建 object 类型的对象值，存储键值对映射。
// 对象是无序集合，键为字符串，值为任意 Value 类型。
//
// 注意：此方法创建的对象不支持自动 GC 管理。
// 如果需要 GC 支持，使用 NewObjectGC()。
//
// 参数：
//   - v: 键值对映射（可为 nil 或空 map）
//
// 返回值：
//   - Value: 对象值
//
// 使用示例：
//
//	// 创建对象
//	obj := engine.NewObject(map[string]engine.Value{
//	    "name": engine.NewString("Alice"),
//	    "age":  engine.NewInt(30),
//	})
//
//	// 获取对象属性
//	name := obj.Object()["name"]
//	fmt.Println(name.String()) // 输出: Alice
//
//	// 获取对象键数
//	fmt.Println(obj.Len())     // 输出: 2
func NewObject(v map[string]Value) Value {
	return &objectValue{value: v}
}

// NewFunc 创建函数值。
//
// 此方法创建 func 类型的函数值，将 Go 函数包装为可在脚本中调用的值。
// 创建的函数值可通过 Engine.RegisterFunc 注册到引擎，
// 或作为参数传递给脚本中的高阶函数。
//
// 参数：
//   - name: 函数名（用于调试和错误报告）
//   - fn: Go 函数实现，必须符合 GoFunction 签名
//
// 返回值：
//   - Value: 函数值
//
// 使用示例：
//
//	// 创建函数值
//	addFunc := engine.NewFunc("add", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    a := args[0].Int()
//	    b := args[1].Int()
//	    return engine.NewInt(a + b), nil
//	})
//
//	// 注册到引擎
//	eng.RegisterFunc("add", addFunc.(*engine.FuncValue).Fn)
//
//	// 或在 Go 代码中直接调用
//	result, _ := vm.CallValue(addFunc, engine.NewInt(10), engine.NewInt(20))
func NewFunc(name string, fn GoFunction) Value {
	return &funcValue{name: name, fn: fn}
}

// NewBigInt 创建大整数值。
//
// 此方法创建 bigint 类型的大整数值，使用 math/big.Int 存储任意精度的整数。
// 适用于超出 int64 范围的大整数运算（如密码学、科学计算）。
//
// 注意：大整数运算比普通整数慢，仅在需要时 사용。
//
// 参数：
//   - v: *big.Int 指针（会被复制，不会修改原值）
//
// 返回值：
//   - Value: 大整数值
//
// 使用示例：
//
//	// 创建大整数
//	bigNum := big.NewInt(0)
//	bigNum.SetString("123456789012345678901234567890", 10)
//
//	val := engine.NewBigInt(bigNum)
//	fmt.Println(val.Type()) // 输出: bigint
//
//	// 大整数运算
//	result := val.Add(engine.NewBigInt(big.NewInt(1)))
//	fmt.Println(result.String()) // 输出大整数字符串
func NewBigInt(v *big.Int) Value {
	return &BigIntValue{value: v}
}

// NewBigDecimal 创建大数值（任意精度小数）。
//
// 此方法创建 bigdecimal 类型的大数值，使用 math/big.Rat 存储任意精度的有理数。
// 适用于需要精确小数运算的场景（如金融计算、科学计算），避免浮点数精度问题。
//
// 注意：
//   - 大数值运算比普通浮点数慢
//   - 分母为 0 时会产生无效值
//
// 参数：
//   - v: *big.Rat 指针（会被复制，不会修改原值）
//
// 返回值：
//   - Value: 大数值
//
// 使用示例：
//
//	// 创建精确小数 0.1（避免浮点误差）
//	rat := new(big.Rat)
//	rat.SetString("0.1") // 精确表示 1/10
//
//	val := engine.NewBigDecimal(rat)
//	fmt.Println(val.Type()) // 输出: bigdecimal
//
//	// 精确运算：0.1 + 0.2
//	point2 := new(big.Rat)
//	point2.SetString("0.2")
//	result := val.Add(engine.NewBigDecimal(point2))
//	fmt.Println(result.String()) // 输出: 0.3（精确）
func NewBigDecimal(v *big.Rat) Value {
	return &BigDecimalValue{value: v}
}

// NewError 创建错误值。
//
// 此方法创建 error 类型的错误值，用于异常处理和错误报告。
// 错误值包含消息、错误码和类型信息，支持条件捕获（catch when 语法）。
//
// 错误码机制：
//   - 使用 error() 函数在脚本中创建错误值
//   - 使用 catch ($e when $e.code == 500) 按错误码捕获
//   - 支持自定义错误类型
//
// 参数：
//   - message: 错误消息（描述性文本）
//   - code: 错误码（数值，用于条件捕获）
//   - errType: 错误类型（如 "Error", "TypeError"），为空时默认为 "Error"
//
// 返回值：
//   - Value: 错误值
//
// 使用示例：
//
//	// 在 Go 中创建错误值
//	err := engine.NewError("division by zero", 1001, "RuntimeError")
//	fmt.Println(err.String()) // 输出错误信息
//
//	// 在脚本中使用
//	// throw error("invalid input", 400)
//	// catch ($e when $e.code == 400) { ... }
func NewError(message string, code int64, errType string) Value {
	if errType == "" {
		errType = "Error"
	}
	return &errorValue{message: message, code: code, errType: errType}
}

// ============================================================================
// GC 辅助函数
// ============================================================================

// IsManagedObject 检查 Value 是否为 GC 可管理的堆对象。
//
// 此方法检查 Value 是否实现了 gc.ManagedObject 接口。
// 只有堆对象（array、object）支持 GC 管理，标量类型（int、float、string 等）
// 不支持 GC。
//
// 参数：
//   - v: 要检查的 Value
//
// 返回值：
//   - true: Value 是 GC 可管理的堆对象
//   - false: Value 不是堆对象或不支持 GC
//
// 使用示例：
//
//	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
//	fmt.Println(engine.IsManagedObject(arr)) // 输出: true
//
//	num := engine.NewInt(42)
//	fmt.Println(engine.IsManagedObject(num)) // 输出: false
func IsManagedObject(v Value) bool {
	_, ok := v.(gc.ManagedObject)
	return ok
}

// AsManagedObject 将 Value 转换为 GC 管理对象接口。
//
// 此方法将 Value 断言为 gc.ManagedObject 接口类型。
// 如果 Value 不是堆对象，返回 nil。
//
// 参数：
//   - v: 要转换的 Value
//
// 返回值：
//   - gc.ManagedObject: GC 管理对象接口，如果不是堆对象返回 nil
//
// 使用示例：
//
//	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
//	if mo := engine.AsManagedObject(arr); mo != nil {
//	    fmt.Printf("引用计数: %d\n", mo.GetRefCount())
//	    mo.IncRef() // 手动增加引用
//	}
func AsManagedObject(v Value) gc.ManagedObject {
	mo, _ := v.(gc.ManagedObject)
	return mo
}

// SetupGCValue 为堆类型 Value 设置 GC 管理器。
//
// 此方法为 arrayValue 和 objectValue 设置垃圾回收器。
// 如果 Value 不是堆类型或 nil，则静默忽略。
//
// 注意：通常不需要直接调用此方法，GC 会在创建对象时自动设置。
//
// 参数：
//   - v: 要设置 GC 的 Value
//   - g: GC 管理器实例
//
// 使用示例：
//
//	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
//	gc := gc.NewGC()
//	engine.SetupGCValue(arr, gc)
//	// 现在数组受 GC 管理
func SetupGCValue(v Value, g *gc.GC) {
	if g == nil || v == nil {
		return
	}
	switch obj := v.(type) {
	case *arrayValue:
		obj.SetupGC(g)
	case *objectValue:
		obj.SetupGC(g)
	}
}

// NewArrayGC 创建 GC 托管的数组值。
//
// 此方法创建 array 类型的数组值，并立即注册到 GC 管理器。
// 与 NewArray() 不同，此方法创建的数组会被 GC 跟踪和管理，
// 支持自动内存回收。
//
// 参数：
//   - v: Value 切片（可为 nil 或空切片）
//   - g: GC 管理器实例，如果为 nil 则等同于 NewArray()
//
// 返回值：
//   - Value: GC 托管的数组值
//
// 使用示例：
//
//	gc := gc.NewGC()
//
//	// 创建 GC 托管数组
//	arr := engine.NewArrayGC([]engine.Value{
//	    engine.NewInt(1),
//	    engine.NewInt(2),
//	}, gc)
//
//	// 数组受 GC 管理，循环引用会自动回收
//	gc.Sweep() // 触发垃圾回收
func NewArrayGC(v []Value, g *gc.GC) Value {
	av := &arrayValue{value: v, alive: true}
	if g != nil {
		av.gcPtr = g
		av.gcID = g.NextID()
		av.refCount = 1
		g.Register(av)
	}
	return av
}

// NewObjectGC 创建 GC 托管的对象值。
//
// 此方法创建 object 类型的对象值，并立即注册到 GC 管理器。
// 与 NewObject() 不同，此方法创建的对象会被 GC 跟踪和管理，
// 支持自动内存回收。
//
// 参数：
//   - v: 键值对映射（可为 nil 或空 map）
//   - g: GC 管理器实例，如果为 nil 则等同于 NewObject()
//
// 返回值：
//   - Value: GC 托管的对象值
//
// 使用示例：
//
//	gc := gc.NewGC()
//
//	// 创建 GC 托管对象
//	obj := engine.NewObjectGC(map[string]engine.Value{
//	    "name": engine.NewString("Alice"),
//	    "refs": engine.NewArray([]engine.Value{...}), // 循环引用
//	}, gc)
//
//	// 对象受 GC 管理，循环引用会自动回收
//	gc.Sweep() // 触发垃圾回收
func NewObjectGC(v map[string]Value, g *gc.GC) Value {
	ov := &objectValue{value: v, alive: true}
	if g != nil {
		ov.gcPtr = g
		ov.gcID = g.NextID()
		ov.refCount = 1
		g.Register(ov)
	}
	return ov
}
