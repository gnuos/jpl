package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterTypeCheck 注册类型检查函数到引擎。
//
// 注册的函数：
//   - 基础类型: is_null, is_bool, is_int, is_float, is_string, is_array, is_object, is_func
//   - 类型别名: is_real, is_double (is_float 别名), is_integer, is_long (is_int 别名)
//   - 扩展检查: is_numeric, is_scalar, empty
//
// 参数：
//   - e: 引擎实例
func RegisterTypeCheck(e *engine.Engine) {
	// 基础类型检查
	e.RegisterFunc("is_null", builtinIsNull)
	e.RegisterFunc("is_bool", builtinIsBool)
	e.RegisterFunc("is_int", builtinIsInt)
	e.RegisterFunc("is_float", builtinIsFloat)
	e.RegisterFunc("is_string", builtinIsString)
	e.RegisterFunc("is_array", builtinIsArray)
	e.RegisterFunc("is_object", builtinIsObject)
	e.RegisterFunc("is_func", builtinIsFunc)

	// 类型别名（兼容 Jx9 风格）
	e.RegisterFunc("is_real", builtinIsFloat)   // is_float 的别名
	e.RegisterFunc("is_double", builtinIsFloat) // is_float 的别名
	e.RegisterFunc("is_integer", builtinIsInt)  // is_int 的别名
	e.RegisterFunc("is_long", builtinIsInt)     // is_int 的别名

	// 扩展类型检查
	e.RegisterFunc("is_numeric", builtinIsNumeric)
	e.RegisterFunc("is_scalar", builtinIsScalar)
	e.RegisterFunc("empty", builtinEmpty)

	// 流类型检查
	e.RegisterFunc("is_stream", builtinIsStream)

	// 正则类型检查
	e.RegisterFunc("is_regex", builtinIsRegex)

	// 大数类型检查
	e.RegisterFunc("is_bigint", builtinIsBigInt)
	e.RegisterFunc("is_bigdecimal", builtinIsBigDecimal)
}

// TypeCheckNames 返回类型检查函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func TypeCheckNames() []string {
	return []string{
		// 基础类型检查
		"is_null", "is_bool", "is_int", "is_float",
		"is_string", "is_array", "is_object", "is_func",
		// 类型别名
		"is_real", "is_double", "is_integer", "is_long",
		// 扩展类型检查
		"is_numeric", "is_scalar", "empty",
		// 流类型检查
		"is_stream",
		// 正则类型检查
		"is_regex",
		// 大数类型检查
		"is_bigint", "is_bigdecimal",
	}
}

// builtinIsNull 检查值是否为 null。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是 null 返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_null(null)            // → true
//	is_null(0)               // → false
//	is_null("")              // → false
func builtinIsNull(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_null() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].IsNull()), nil
}

// builtinIsBool 检查值是否为布尔类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是布尔类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_bool(true)            // → true
//	is_bool(1)               // → false
func builtinIsBool(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_bool() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeBool), nil
}

// builtinIsInt 检查值是否为整数类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是整数类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_int(42)               // → true
//	is_int(3.14)             // → false
//	is_int("42")             // → false
func builtinIsInt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_int() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeInt), nil
}

// builtinIsFloat 检查值是否为浮点类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是浮点类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_float(3.14)           // → true
//	is_float(42)             // → false
func builtinIsFloat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_float() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeFloat), nil
}

// builtinIsString 检查值是否为字符串类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是字符串类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_string("hello")       // → true
//	is_string(42)            // → false
func builtinIsString(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_string() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeString), nil
}

// builtinIsArray 检查值是否为数组类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是数组类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_array([1, 2, 3])      // → true
//	is_array("hello")        // → false
func builtinIsArray(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_array() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeArray), nil
}

// builtinIsObject 检查值是否为对象类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是对象类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_object({a: 1})        // → true
//	is_object([1, 2])        // → false
func builtinIsObject(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_object() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeObject), nil
}

// builtinIsFunc 检查值是否为函数类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是函数类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_func(fn() {})         // → true
//	is_func("hello")         // → false
func builtinIsFunc(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_func() expects 1 argument, got %d", len(args))
	}
	return engine.NewBool(args[0].Type() == engine.TypeFunc), nil
}

// builtinIsNumeric 检查值是否为数字类型（整数或浮点数）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是数字类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_numeric(42)           // → true
//	is_numeric(3.14)         // → true
//	is_numeric("42")         // → false（字符串不是数字）
func builtinIsNumeric(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_numeric() expects 1 argument, got %d", len(args))
	}
	t := args[0].Type()
	return engine.NewBool(t == engine.TypeInt || t == engine.TypeFloat || t == engine.TypeBigInt || t == engine.TypeBigDecimal), nil
}

// builtinIsScalar 检查值是否为标量类型。
//
// 标量类型：null, bool, int, float, string, bigint, bigdecimal
// 非标量类型：array, object, func
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 是标量类型返回 true
//   - error: 参数数量错误
//
// 使用示例：
//
//	is_scalar(42)            // → true
//	is_scalar("hello")       // → true
//	is_scalar([1, 2])        // → false
func builtinIsScalar(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_scalar() expects 1 argument, got %d", len(args))
	}
	t := args[0].Type()
	// 标量类型：null, bool, int, float, string, bigint, bigdecimal
	return engine.NewBool(
		t == engine.TypeNull ||
			t == engine.TypeBool ||
			t == engine.TypeInt ||
			t == engine.TypeFloat ||
			t == engine.TypeString ||
			t == engine.TypeBigInt ||
			t == engine.TypeBigDecimal,
	), nil
}

// builtinEmpty 检查值是否为空（PHP 风格）
//
// empty() 函数判断一个值是否被视为"空"。以下情况返回 true：
//   - null
//   - false
//   - 0 (整数) 或 0.0 (浮点数)
//   - "" (空字符串) 或 "0" (字符串零)
//   - [] (空数组) 或 {} (空对象)
//   - 函数永不为空
//
// 参数：
//   - value: 要检查的值
//
// 返回值：
//   - bool: true 如果值为空，false 否则
//
// 注意：与 JavaScript 不同，字符串 "0" 被视为空
//
// 使用示例：
//
//	$val = 0
//	if (empty($val)) {
//	    print "值为空"
//	}
//
//	$arr = []
//	print empty($arr)     // true
//	print empty("hello")  // false
//	print empty("0")      // true (注意！)
func builtinEmpty(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	isEmpty := false

	switch v.Type() {
	case engine.TypeNull:
		isEmpty = true
	case engine.TypeBool:
		isEmpty = !v.Bool()
	case engine.TypeInt:
		isEmpty = v.Int() == 0
	case engine.TypeFloat:
		isEmpty = v.Float() == 0.0
	case engine.TypeString:
		s := v.String()
		isEmpty = s == "" || s == "0"
	case engine.TypeArray:
		isEmpty = v.Len() == 0
	case engine.TypeObject:
		isEmpty = v.Len() == 0
	case engine.TypeFunc:
		// 函数永不为空
		isEmpty = false
	}

	return engine.NewBool(isEmpty), nil
}

func builtinIsStream(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_stream() expects 1 argument, got %d", len(args))
	}

	return engine.NewBool(engine.IsStream(args[0])), nil
}

// builtinIsRegex 检查值是否为正则表达式类型。
//
// 语法：is_regex($value)
//
// 参数：
//   - $value: 要检查的值
//
// 返回值：
//   - true: 值是正则表达式
//   - false: 值不是正则表达式
//
// 示例：
//
//	$re = #/\d+/#
//	is_regex($re)  // true
//	is_regex("x")  // false
func builtinIsRegex(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_regex() expects 1 argument, got %d", len(args))
	}

	return engine.NewBool(engine.IsRegex(args[0])), nil
}

// builtinIsBigInt 检查值是否为大整数类型（BigInt）。
//
// 语法：is_bigint($value)
//
// 参数：
//   - $value: 要检查的值
//
// 返回值：
//   - true: 值是大整数（超出 int64 范围的整数）
//   - false: 值不是大整数
//
// 示例：
//
//	$x = 999999999999999999999
//	is_bigint($x)   // true
//	is_bigint(42)    // false
//	is_bigint(3.14)  // false
func builtinIsBigInt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_bigint() expects 1 argument, got %d", len(args))
	}
	_, ok := args[0].(*engine.BigIntValue)
	return engine.NewBool(ok), nil
}

// builtinIsBigDecimal 检查值是否为大浮点数类型（BigDecimal）。
//
// 语法：is_bigdecimal($value)
//
// 参数：
//   - $value: 要检查的值
//
// 返回值：
//   - true: 值是大浮点数（高精度小数）
//   - false: 值不是大浮点数
//
// 示例：
//
//	$x = 1.234567890123456789012345
//	is_bigdecimal($x)   // true
//	is_bigdecimal(3.14)  // false
//	is_bigdecimal(42)    // false
func builtinIsBigDecimal(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_bigdecimal() expects 1 argument, got %d", len(args))
	}
	_, ok := args[0].(*engine.BigDecimalValue)
	return engine.NewBool(ok), nil
}

// TypeCheckSigs returns function signatures for REPL :doc command.
func TypeCheckSigs() map[string]string {
	return map[string]string{
		"is_null":       "is_null(value) → bool  — Check if value is null",
		"is_bool":       "is_bool(value) → bool  — Check if value is boolean",
		"is_int":        "is_int(value) → bool  — Check if value is integer",
		"is_float":      "is_float(value) → bool  — Check if value is float",
		"is_string":     "is_string(value) → bool  — Check if value is string",
		"is_array":      "is_array(value) → bool  — Check if value is array",
		"is_object":     "is_object(value) → bool  — Check if value is object",
		"is_func":       "is_func(value) → bool  — Check if value is function",
		"is_numeric":    "is_numeric(value) → bool  — Check if value is numeric (int/float/bigint/bigdecimal)",
		"is_scalar":     "is_scalar(value) → bool  — Check if value is scalar type",
		"is_bigint":     "is_bigint(value) → bool  — Check if value is BigInt",
		"is_bigdecimal": "is_bigdecimal(value) → bool  — Check if value is BigDecimal",
		"is_regex":      "is_regex(value) → bool  — Check if value is regex",
		"is_stream":     "is_stream(value) → bool  — Check if value is stream",
		"is_error":      "is_error(value) → bool  — Check if value is error",
		"empty":         "empty(value) → bool  — Check if value is empty (PHP style)",
	}
}
