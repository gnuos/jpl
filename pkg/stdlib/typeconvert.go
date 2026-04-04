package stdlib

import (
	"fmt"
	"strconv"

	"github.com/gnuos/jpl/engine"
)

// RegisterTypeConvert 注册类型转换函数到引擎。
//
// 注册的函数：
//   - intval: 转换为整数（支持多进制）
//   - floatval: 转换为浮点数
//   - strval: 转换为字符串
//   - boolval: 转换为布尔值
//
// 参数：
//   - e: 引擎实例
func RegisterTypeConvert(e *engine.Engine) {
	e.RegisterFunc("intval", builtinIntval)
	e.RegisterFunc("floatval", builtinFloatval)
	e.RegisterFunc("strval", builtinStrval)
	e.RegisterFunc("boolval", builtinBoolval)
}

// TypeConvertNames 返回类型转换函数名称列表。
//
// 返回值：
//   - []string: 函数名列表 ["intval", "floatval", "strval", "boolval"]
func TypeConvertNames() []string {
	return []string{"intval", "floatval", "strval", "boolval"}
}

// builtinIntval 将值转换为整数
//
// 支持多进制字符串解析（2/8/10/16 进制）。转换失败返回 0，不产生错误。
// 浮点数转为整数时会截断小数部分（向零取整）。
//
// 参数：
//   - value: 要转换的值（int/float/string/bool/null/array）
//   - base: 可选，字符串解析进制，默认为 10
//     支持：2(二进制)、8(八进制)、10(十进制)、16(十六进制)
//
// 返回值：
//   - int: 转换后的整数值
//
// 转换规则：
//   - int: 直接返回
//   - float: 截断小数部分
//   - string: 按指定进制解析，失败返回 0
//   - bool: true → 1, false → 0
//   - null/array: 返回 0
//
// 使用示例：
//
//	intval(42)              // 42
//	intval(3.7)             // 3 (截断)
//	intval("123")           // 123
//	intval("FF", 16)        // 255 (十六进制)
//	intval("1010", 2)       // 10 (二进制)
//	intval("abc")           // 0 (解析失败)
//	intval(true)            // 1
func builtinIntval(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("intval() expects 1 or 2 arguments, got %d", len(args))
	}

	// 默认进制为 10
	base := 10
	if len(args) == 2 {
		if args[1].Type() == engine.TypeInt {
			base = int(args[1].Int())
		}
	}

	v := args[0]
	switch v.Type() {
	case engine.TypeInt:
		return v, nil
	case engine.TypeFloat:
		return engine.NewInt(int64(v.Float())), nil
	case engine.TypeString:
		s := v.String()
		if s == "" {
			return engine.NewInt(0), nil
		}
		// 根据进制解析
		var n int64
		var err error
		if base == 10 {
			n, err = strconv.ParseInt(s, 10, 64)
		} else if base == 8 {
			n, err = strconv.ParseInt(s, 8, 64)
		} else if base == 16 {
			n, err = strconv.ParseInt(s, 16, 64)
		} else if base == 2 {
			n, err = strconv.ParseInt(s, 2, 64)
		} else {
			return engine.NewInt(0), nil
		}
		if err != nil {
			return engine.NewInt(0), nil
		}
		return engine.NewInt(n), nil
	case engine.TypeBool:
		if engine.IsTruthy(v) {
			return engine.NewInt(1), nil
		}
		return engine.NewInt(0), nil
	case engine.TypeNull:
		return engine.NewInt(0), nil
	default:
		return engine.NewInt(0), nil
	}
}

// builtinFloatval 将值转换为浮点数。
//
// 转换规则：
//   - float: 直接返回
//   - int: 转换为浮点数
//   - string: 解析为浮点数，失败返回 0.0
//   - bool: true → 1.0, false → 0.0
//   - null: 返回 0.0
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要转换的值
//
// 返回值：
//   - float64: 转换后的浮点数值
//   - error: 参数数量错误
//
// 使用示例：
//
//	floatval(42)             // → 42.0
//	floatval("3.14")         // → 3.14
//	floatval(true)           // → 1.0
func builtinFloatval(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floatval() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	switch v.Type() {
	case engine.TypeFloat:
		return v, nil
	case engine.TypeInt:
		return engine.NewFloat(v.Float()), nil
	case engine.TypeString:
		s := v.String()
		if s == "" {
			return engine.NewFloat(0.0), nil
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return engine.NewFloat(0.0), nil
		}
		return engine.NewFloat(f), nil
	case engine.TypeBool:
		if engine.IsTruthy(v) {
			return engine.NewFloat(1.0), nil
		}
		return engine.NewFloat(0.0), nil
	case engine.TypeNull:
		return engine.NewFloat(0.0), nil
	default:
		return engine.NewFloat(0.0), nil
	}
}

// builtinStrval 将值转换为字符串。
//
// 使用值的 String() 方法进行转换。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要转换的值
//
// 返回值：
//   - string: 转换后的字符串
//   - error: 参数数量错误
//
// 使用示例：
//
//	strval(42)               // → "42"
//	strval(3.14)             // → "3.14"
//	strval(true)             // → "true"
//	strval([1, 2])           // → "[1, 2]"
func builtinStrval(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strval() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	// 使用值的 String() 方法
	return engine.NewString(v.String()), nil
}

// builtinBoolval 将值转换为布尔值。
//
// 使用 IsTruthy 语义进行转换。以下值转换为 false：
//   - null
//   - false
//   - 0 (整数) 或 0.0 (浮点数)
//   - "" (空字符串) 或 "0" (字符串零)
//   - [] (空数组) 或 {} (空对象)
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要转换的值
//
// 返回值：
//   - bool: 转换后的布尔值
//   - error: 参数数量错误
//
// 使用示例：
//
//	boolval(1)               // → true
//	boolval(0)               // → false
//	boolval("hello")         // → true
//	boolval("")              // → false
func builtinBoolval(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("boolval() expects 1 argument, got %d", len(args))
	}

	return engine.NewBool(engine.IsTruthy(args[0])), nil
}

// TypeConvertSigs returns function signatures for REPL :doc command.
func TypeConvertSigs() map[string]string {
	return map[string]string{
		"intval":     "intval(value, [base]) → int  — Convert to integer",
		"floatval":   "floatval(value) → float  — Convert to float",
		"strval":     "strval(value) → string  — Convert to string",
		"boolval":    "boolval(value) → bool  — Convert to boolean",
		"bigint":     "bigint(value) → bigint  — Convert to BigInt",
		"bigdecimal": "bigdecimal(value) → bigdecimal  — Convert to BigDecimal",
	}
}
