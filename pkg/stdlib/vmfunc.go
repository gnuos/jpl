package stdlib

import (
	"fmt"
	"unicode/utf8"

	"github.com/gnuos/jpl"
	"github.com/gnuos/jpl/engine"
)

// RegisterVMFunc 注册 VM/反射函数到引擎。
//
// 注册的函数：
//   - func_num_args: 返回当前函数参数数量
//   - func_get_arg: 获取指定索引的参数
//   - func_get_args: 获取所有参数数组
//   - function_exists: 检查函数是否存在
//   - is_callable: 检查值是否可调用
//   - get_defined_functions: 获取所有函数名
//   - get_defined_constants: 获取所有常量名
//   - jpl_version: 返回 JPL 版本号
//   - utf8_encode: UTF-8 编码（转十六进制）
//   - utf8_decode: UTF-8 解码（从十六进制）
//
// 参数：
//   - e: 引擎实例
func RegisterVMFunc(e *engine.Engine) {
	// 函数参数获取
	e.RegisterFunc("func_num_args", builtinFuncNumArgs)
	e.RegisterFunc("func_get_arg", builtinFuncGetArg)
	e.RegisterFunc("func_get_args", builtinFuncGetArgs)

	// 函数/可调用检查
	e.RegisterFunc("function_exists", builtinFunctionExists)
	e.RegisterFunc("is_callable", builtinIsCallable)

	// 定义列表
	e.RegisterFunc("get_defined_functions", builtinGetDefinedFunctions)
	e.RegisterFunc("get_defined_constants", builtinGetDefinedConstants)

	// 版本信息
	e.RegisterFunc("jpl_version", builtinJPLVersion)

	// UTF-8 编解码
	e.RegisterFunc("utf8_encode", builtinUTF8Encode)
	e.RegisterFunc("utf8_decode", builtinUTF8Decode)

	// 模块注册
	e.RegisterModule("vm", map[string]engine.GoFunction{
		"func_num_args":         builtinFuncNumArgs,
		"func_get_arg":          builtinFuncGetArg,
		"func_get_args":         builtinFuncGetArgs,
		"function_exists":       builtinFunctionExists,
		"is_callable":           builtinIsCallable,
		"get_defined_functions": builtinGetDefinedFunctions,
		"get_defined_constants": builtinGetDefinedConstants,
		"jpl_version":           builtinJPLVersion,
		"utf8_encode":           builtinUTF8Encode,
		"utf8_decode":           builtinUTF8Decode,
	})
}

// VMFuncNames 返回 VM/反射函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func VMFuncNames() []string {
	return []string{
		"func_num_args", "func_get_arg", "func_get_args",
		"function_exists", "is_callable",
		"get_defined_functions", "get_defined_constants",
		"jpl_version", "utf8_encode", "utf8_decode",
	}
}

// ============================================================================
// 函数参数获取
// ============================================================================

// builtinFuncNumArgs 返回当前用户定义函数的参数数量。
//
// 此函数只能在用户定义的函数内部调用，在主函数或全局作用域调用会返回错误。
// 用于实现可变参数函数或参数验证。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - int: 当前函数的参数数量
//   - error: 在非函数上下文中调用时返回错误
//
// 使用示例：
//
//	fn test(a, b, c) {
//	    return func_num_args()  // → 3
//	}
//	test(1, 2, 3)
func builtinFuncNumArgs(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("func_num_args() expects 0 arguments, got %d", len(args))
	}

	vm := ctx.VM()
	if vm == nil {
		return nil, fmt.Errorf("func_num_args() requires a VM context")
	}

	// 获取当前函数的参数数量
	fn := vm.CurrentFunction()
	if fn == nil {
		return nil, fmt.Errorf("func_num_args() cannot be called outside a function")
	}

	// 检查是否在主函数中调用
	if fn.Name == "<main>" {
		return nil, fmt.Errorf("func_num_args() cannot be called from outside a user-defined function")
	}

	return engine.NewInt(int64(fn.Params)), nil
}

// builtinFuncGetArg 返回当前函数指定索引的参数值。
//
// 此函数只能在用户定义的函数内部调用。索引从 0 开始。
// 如果索引超出范围，返回错误。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 参数索引（整数，0-based）
//
// 返回值：
//   - Value: 指定索引的参数值
//   - error: 索引越界或在非函数上下文中调用
//
// 使用示例：
//
//	fn test(a, b, c) {
//	    return func_get_arg(1)  // → b 的值
//	}
//	test(10, 20, 30)  // → 20
func builtinFuncGetArg(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("func_get_arg() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeInt {
		return nil, fmt.Errorf("func_get_arg() argument must be an integer, got %s", args[0].Type())
	}

	vm := ctx.VM()
	if vm == nil {
		return nil, fmt.Errorf("func_get_arg() requires a VM context")
	}

	fn := vm.CurrentFunction()
	if fn == nil || fn.Name == "<main>" {
		return nil, fmt.Errorf("func_get_arg() cannot be called from outside a user-defined function")
	}

	n := args[0].Int()
	if n < 0 || int(n) >= fn.Params {
		return nil, fmt.Errorf("func_get_arg() argument %d out of range (0-%d)", n, fn.Params-1)
	}

	// 获取参数值
	registers := vm.CurrentRegisters()
	if registers == nil || int(n) >= len(registers) {
		return engine.NewNull(), nil
	}

	return registers[n], nil
}

// builtinFuncGetArgs 返回当前函数的所有参数数组。
//
// 此函数只能在用户定义的函数内部调用。
// 返回的数组按参数顺序排列，可直接通过索引访问。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - array: 参数值数组
//   - error: 在非函数上下文中调用时返回错误
//
// 使用示例：
//
//	fn sum(a, b, c) {
//	    return func_get_args()[0] + func_get_args()[1] + func_get_args()[2]
//	}
//	sum(10, 20, 30)  // → 60
func builtinFuncGetArgs(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("func_get_args() expects 0 arguments, got %d", len(args))
	}

	vm := ctx.VM()
	if vm == nil {
		return nil, fmt.Errorf("func_get_args() requires a VM context")
	}

	fn := vm.CurrentFunction()
	if fn == nil || fn.Name == "<main>" {
		return nil, fmt.Errorf("func_get_args() cannot be called from outside a user-defined function")
	}

	// 获取所有参数
	registers := vm.CurrentRegisters()
	if registers == nil {
		return engine.NewArray(nil), nil
	}

	paramCount := min(fn.Params, len(registers))

	params := make([]engine.Value, paramCount)
	for i := range paramCount {
		params[i] = registers[i]
	}

	return engine.NewArray(params), nil
}

// ============================================================================
// 函数/可调用检查
// ============================================================================

// builtinFunctionExists 检查指定名称的函数是否存在。
//
// 检查编译后的用户函数和引擎注册的内置函数。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 函数名（字符串）
//
// 返回值：
//   - bool: 函数存在返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	fn greet(name) { return "Hello, " + name }
//	function_exists("greet")     // → true
//	function_exists("print")     // → true
//	function_exists("nonexist")  // → false
func builtinFunctionExists(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("function_exists() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("function_exists() argument must be a string, got %s", args[0].Type())
	}

	name := args[0].String()
	vm := ctx.VM()
	if vm != nil && vm.FunctionExists(name) {
		return engine.NewBool(true), nil
	}
	return engine.NewBool(false), nil
}

// builtinIsCallable 检查值是否可调用。
//
// 可调用的类型：
//   - 函数值（闭包、lambda）
//   - 有效的函数名字符串
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的值
//
// 返回值：
//   - bool: 可调用返回 true
//   - error: 无
//
// 使用示例：
//
//	fn add(a, b) { return a + b }
//	is_callable(add)           // → true
//	is_callable("add")         // → true
//	is_callable("print")       // → true
//	is_callable(42)            // → false
func builtinIsCallable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_callable() expects 1 argument, got %d", len(args))
	}

	vm := ctx.VM()
	if vm == nil {
		return engine.NewBool(false), nil
	}

	switch args[0].Type() {
	case engine.TypeFunc:
		return engine.NewBool(true), nil
	case engine.TypeString:
		// 检查字符串是否为有效的函数名
		name := args[0].String()
		return engine.NewBool(vm.FunctionExists(name)), nil
	default:
		return engine.NewBool(false), nil
	}
}

// ============================================================================
// 定义列表
// ============================================================================

// builtinGetDefinedFunctions 返回所有已定义函数名的数组。
//
// 包括用户定义的函数和引擎注册的内置函数。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - array: 函数名字符串数组
//   - error: 无
//
// 使用示例：
//
//	fn myFunc() {}
//	get_defined_functions()  // → ["myFunc", "print", "len", ...]
func builtinGetDefinedFunctions(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("get_defined_functions() expects 0 arguments, got %d", len(args))
	}

	vm := ctx.VM()
	if vm == nil {
		return engine.NewArray(nil), nil
	}

	names := vm.GetAllFunctionNames()
	values := make([]engine.Value, len(names))
	for i, name := range names {
		values[i] = engine.NewString(name)
	}
	return engine.NewArray(values), nil
}

// builtinGetDefinedConstants 返回所有已定义常量名的数组。
//
// 包括预设常量（PI、E、INF 等）和用户定义的常量。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - array: 常量名字符串数组
//   - error: 无
//
// 使用示例：
//
//	get_defined_constants()  // → ["PI", "E", "INF", "NaN", ...]
func builtinGetDefinedConstants(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("get_defined_constants() expects 0 arguments, got %d", len(args))
	}

	eng := ctx.Engine()
	if eng == nil {
		return engine.NewArray(nil), nil
	}

	// 获取所有常量名
	constNames := eng.GetConstantNames()
	values := make([]engine.Value, len(constNames))
	for i, name := range constNames {
		values[i] = engine.NewString(name)
	}
	return engine.NewArray(values), nil
}

// ============================================================================
// 版本信息
// ============================================================================

// builtinJPLVersion 返回 JPL 版本号字符串。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - string: 版本号（如 "1.0.0"）
//   - error: 无
//
// 使用示例：
//
//	jpl_version()  // → "1.0.0"
func builtinJPLVersion(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("jpl_version() expects 0 arguments, got %d", len(args))
	}
	return engine.NewString(jpl.Version), nil
}

// ============================================================================
// UTF-8 编解码
// ============================================================================

// builtinUTF8Encode 将字符串编码为 UTF-8 十六进制表示。
//
// 将字符串转换为其 UTF-8 字节序列的十六进制字符串。
// 可用于数据传输或存储场景。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要编码的字符串
//
// 返回值：
//   - string: UTF-8 字节的十六进制字符串
//   - error: 参数错误
//
// 使用示例：
//
//	utf8_encode("Hello")  // → "48656c6c6f"
//	utf8_encode("中文")   // → "e4b8ade69687"
func builtinUTF8Encode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("utf8_encode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("utf8_encode() argument must be a string, got %s", args[0].Type())
	}

	str := args[0].String()
	// 返回 UTF-8 字节的十六进制表示
	encoded := fmt.Sprintf("%x", []byte(str))
	return engine.NewString(encoded), nil
}

// builtinUTF8Decode 将 UTF-8 十六进制字符串解码为普通字符串。
//
// 将十六进制字符串解析为字节序列，验证 UTF-8 有效性后转换为字符串。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 十六进制字符串（长度必须为偶数）
//
// 返回值：
//   - string: 解码后的字符串
//   - error: 无效的十六进制或 UTF-8 序列
//
// 使用示例：
//
//	utf8_decode("48656c6c6f")    // → "Hello"
//	utf8_decode("e4b8ade69687")  // → "中文"
func builtinUTF8Decode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("utf8_decode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("utf8_decode() argument must be a string, got %s", args[0].Type())
	}

	hexStr := args[0].String()
	if hexStr == "" {
		return engine.NewString(""), nil
	}

	// 解析十六进制字符串
	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("utf8_decode() invalid hex string length: must be even")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
		if err != nil {
			return nil, fmt.Errorf("utf8_decode() invalid hex string: %s", hexStr)
		}
		bytes[i/2] = b
	}

	// 验证是否为有效的 UTF-8
	if !utf8.Valid(bytes) {
		return nil, fmt.Errorf("utf8_decode() invalid UTF-8 sequence")
	}

	return engine.NewString(string(bytes)), nil
}

// VMFuncSigs returns function signatures for REPL :doc command.
func VMFuncSigs() map[string]string {
	return map[string]string{
		"func_num_args":         "func_num_args() → int  — Get current function arg count",
		"func_get_arg":          "func_get_arg(index) → value  — Get arg by index",
		"func_get_args":         "func_get_args() → array  — Get all args as array",
		"function_exists":       "function_exists(name) → bool  — Check if function exists",
		"is_callable":           "is_callable(value) → bool  — Check if value is callable",
		"get_defined_functions": "get_defined_functions() → array  — Get all function names",
		"get_defined_constants": "get_defined_constants() → array  — Get all constant names",
		"jpl_version":           "jpl_version() → string  — Get JPL version",
		"utf8_encode":           "utf8_encode(str) → string  — Encode to UTF-8 hex",
		"utf8_decode":           "utf8_decode(hex) → string  — Decode UTF-8 hex",
	}
}
