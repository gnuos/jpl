package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterReflect 注册反射 API 函数到引擎。
//
// 注册的函数：
//   - typeof: 返回值的类型字符串
//   - varexists: 检查变量或函数是否存在
//   - getvar: 按名称获取变量值
//   - setvar: 按名称设置变量
//   - listvars: 返回所有全局变量名数组
//   - listfns: 返回所有函数名数组
//   - fn_exists: 检查函数是否存在
//   - getfninfo: 获取函数详细信息
//   - callfn: 按名称动态调用函数
//
// 参数：
//   - e: 引擎实例
func RegisterReflect(e *engine.Engine) {
	// 变量查询
	e.RegisterFunc("typeof", builtinTypeOf)
	e.RegisterFunc("varexists", builtinVarExists)
	e.RegisterFunc("getvar", builtinGetVar)
	e.RegisterFunc("setvar", builtinSetVar)
	e.RegisterFunc("listvars", builtinListVars)

	// 函数查询
	e.RegisterFunc("listfns", builtinListFns)
	e.RegisterFunc("fn_exists", builtinFnExists)
	e.RegisterFunc("getfninfo", builtinGetFnInfo)
	e.RegisterFunc("callfn", builtinCallFn)
}

// ReflectNames 返回反射函数名称列表。
//
// 返回值：
//   - []string: 函数名列表 ["typeof", "varexists", "getvar", "setvar", "listvars", "listfns", "fn_exists", "getfninfo", "callfn"]
func ReflectNames() []string {
	return []string{
		"typeof", "varexists", "getvar", "setvar", "listvars",
		"listfns", "fn_exists", "getfninfo", "callfn",
	}
}

// ============================================================================
// 变量查询
// ============================================================================

// builtinTypeOf 返回值的类型字符串。
//
// 返回值的类型名称，可用于运行时类型检查和调试。
// 返回的类型字符串包括：null, bool, int, float, string, array, object, func, error 等。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查类型的值
//
// 返回值：
//   - string: 类型名称字符串
//   - error: 参数数量错误
//
// 使用示例：
//
//	typeof(null)           // → "null"
//	typeof(true)           // → "bool"
//	typeof(42)             // → "int"
//	typeof(3.14)           // → "float"
//	typeof("hello")        // → "string"
//	typeof([1, 2, 3])      // → "array"
//	typeof({a: 1})         // → "object"
//	typeof(fn() {})        // → "func"
func builtinTypeOf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("typeof() expects 1 argument, got %d", len(args))
	}
	return engine.NewString(args[0].Type().String()), nil
}

// builtinVarExists 检查变量、函数或常量是否存在。
//
// 检查顺序：全局变量 → 引擎注册变量 → 常量 → 函数
// 只要任一项存在就返回 true。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要检查的名称（字符串，不需要 $ 前缀）
//
// 返回值：
//   - bool: 存在返回 true，否则返回 false
//   - error: 参数错误
//
// 使用示例：
//
//	$x = 42
//	varexists("$x")        // → true
//	varexists("print")     // → true（内置函数）
//	varexists("PI")        // → true（预设常量）
//	varexists("nonexist")  // → false
func builtinVarExists(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("exists() argument must be a string, got %s", args[0].Type())
	}
	name := args[0].String()

	vm := ctx.VM()
	if vm != nil {
		// 检查全局变量
		if _, ok := vm.GetGlobal(name); ok {
			return engine.NewBool(true), nil
		}
		// 检查引擎注册变量
		if _, err := ctx.Engine().Get(name); err == nil {
			return engine.NewBool(true), nil
		}
		// 检查常量
		if _, ok := ctx.Engine().GetConst(name); ok {
			return engine.NewBool(true), nil
		}
		// 检查函数
		if vm.FunctionExists(name) {
			return engine.NewBool(true), nil
		}
	}
	return engine.NewBool(false), nil
}

// builtinGetVar 按名称获取变量值。
//
// 检查顺序：全局变量 → 引擎注册变量 → 常量
// 如果都不存在，返回 null。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 变量名（字符串，不需要 $ 前缀）
//
// 返回值：
//   - Value: 变量值，不存在返回 null
//   - error: 参数错误
//
// 使用示例：
//
//	$x = 100
//	getvar("$x")           // → 100
//	getvar("nonexist")     // → null
func builtinGetVar(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("getvar() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("getvar() argument must be a string, got %s", args[0].Type())
	}
	name := args[0].String()

	vm := ctx.VM()
	if vm != nil {
		// 检查全局变量
		if val, ok := vm.GetGlobal(name); ok {
			return val, nil
		}
		// 检查引擎注册变量
		if val, err := ctx.Engine().Get(name); err == nil {
			return val, nil
		}
		// 检查常量
		if val, ok := ctx.Engine().GetConst(name); ok {
			return val, nil
		}
	}
	return engine.NewNull(), nil
}

// builtinSetVar 按名称设置变量值。
//
// 如果变量已存在则更新，不存在则创建。
// 设置的变量为全局变量，可在脚本任何位置访问。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 变量名（字符串，不需要 $ 前缀）
//   - args[1]: 要设置的值
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	setvar("$x", 42)       // 创建 $x = 42
//	setvar("$name", "Alice")
//	getvar("$x")           // → 42
func builtinSetVar(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("setvar() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("setvar() argument 1 must be a string, got %s", args[0].Type())
	}
	name := args[0].String()
	value := args[1]

	vm := ctx.VM()
	if vm != nil {
		vm.SetGlobal(name, value)
		return engine.NewBool(true), nil
	}
	return engine.NewBool(false), nil
}

// builtinListVars 返回所有全局变量名数组。
//
// 只返回全局变量的名称，不包含局部变量、常量或函数。
// 返回的名称带有 $ 前缀。
//
// 参数：
//   - ctx: 执行上下文
//
// 返回值：
//   - array: 变量名字符串数组
//   - error: 无
//
// 使用示例：
//
//	$a = 1
//	$b = 2
//	listvars()             // → ["$a", "$b"]
func builtinListVars(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("listvars() expects 0 arguments, got %d", len(args))
	}

	vm := ctx.VM()
	if vm == nil {
		return engine.NewArray(nil), nil
	}

	names := vm.GetGlobalNames()
	values := make([]engine.Value, len(names))
	for i, name := range names {
		values[i] = engine.NewString(name)
	}
	return engine.NewArray(values), nil
}

// ============================================================================
// 函数查询
// ============================================================================

// builtinListFns 返回所有已定义函数名数组。
//
// 包括用户定义的函数和引擎注册的内置函数。
// 返回的数组是去重的，顺序不确定。
//
// 参数：
//   - ctx: 执行上下文
//
// 返回值：
//   - array: 函数名字符串数组
//   - error: 无
//
// 使用示例：
//
//	fn add(a, b) { return a + b }
//	listfns()              // → ["add", "print", "len", ...]
func builtinListFns(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("listfns() expects 0 arguments, got %d", len(args))
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

// builtinFnExists 检查函数是否存在。
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
//	fn_exists("greet")     // → true
//	fn_exists("print")     // → true（内置函数）
//	fn_exists("nonexist")  // → false
func builtinFnExists(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("fn_exists() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("fn_exists() argument must be a string, got %s", args[0].Type())
	}
	name := args[0].String()

	vm := ctx.VM()
	if vm != nil && vm.FunctionExists(name) {
		return engine.NewBool(true), nil
	}
	return engine.NewBool(false), nil
}

// builtinGetFnInfo 获取函数的详细信息。
//
// 返回函数的名称、参数列表、参数数量和重载数量。
// 如果函数不存在，返回 null。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 函数名（字符串）
//
// 返回值：
//   - object: 函数信息对象 {name, params, paramCount, overloads}
//   - null: 函数不存在
//   - error: 参数错误
//
// 返回对象字段：
//   - name: 函数名
//   - params: 参数名数组
//   - paramCount: 参数数量
//   - overloads: 重载版本数量
//
// 使用示例：
//
//	fn add(a, b) { return a + b }
//	getfninfo("add")
//	// → {name: "add", params: ["$a", "$b"], paramCount: 2, overloads: 1}
func builtinGetFnInfo(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("getfninfo() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("getfninfo() argument must be a string, got %s", args[0].Type())
	}
	name := args[0].String()

	vm := ctx.VM()
	if vm == nil {
		return engine.NewNull(), nil
	}

	// 检查编译后的函数
	infos, ok := vm.GetFunctionInfo(name)
	if ok && len(infos) > 0 {
		// 返回第一个重载版本的信息
		info := infos[0]
		params := make([]engine.Value, len(info.ParamNames))
		for i, p := range info.ParamNames {
			params[i] = engine.NewString(p)
		}
		obj := map[string]engine.Value{
			"name":       engine.NewString(info.Name),
			"params":     engine.NewArray(params),
			"paramCount": engine.NewInt(int64(info.ParamCount)),
			"overloads":  engine.NewInt(int64(len(infos))),
		}
		return engine.NewObject(obj), nil
	}

	// 检查引擎注册的 Go 函数
	if vm.FunctionExists(name) {
		obj := map[string]engine.Value{
			"name":       engine.NewString(name),
			"params":     engine.NewArray(nil),
			"paramCount": engine.NewInt(0),
			"overloads":  engine.NewInt(1),
		}
		return engine.NewObject(obj), nil
	}

	return engine.NewNull(), nil
}

// builtinCallFn 按名称动态调用函数。
//
// 支持两种参数传递方式：
//   - 展开参数：callfn("fn", arg1, arg2, ...)
//   - 数组参数：callfn("fn", [arg1, arg2, ...])
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 函数名（字符串）
//   - args[1:]: 函数参数（展开或数组）
//
// 返回值：
//   - Value: 函数返回值
//   - error: 函数不存在或调用错误
//
// 使用示例：
//
//	fn add(a, b) { return a + b }
//	callfn("add", 10, 20)           // → 30
//	callfn("add", [10, 20])         // → 30
//	callfn("print", "Hello")        // 输出: Hello
func builtinCallFn(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("callfn() expects at least 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("callfn() argument 1 must be a string, got %s", args[0].Type())
	}
	name := args[0].String()

	vm := ctx.VM()
	if vm == nil {
		return nil, fmt.Errorf("callfn() requires a VM context")
	}

	// 收集函数参数
	var fnArgs []engine.Value
	if len(args) == 2 && args[1].Type() == engine.TypeArray {
		// 第二个参数是数组，展开为参数列表
		fnArgs = args[1].Array()
	} else if len(args) > 1 {
		// 直接使用后续参数
		fnArgs = args[1:]
	}

	return vm.CallByName(name, fnArgs...)
}

// ReflectSigs returns function signatures for REPL :doc command.
func ReflectSigs() map[string]string {
	return map[string]string{
		"typeof":                "typeof(value) → string  — Get type name",
		"getvar":                "getvar(name) → value  — Get variable by name",
		"setvar":                "setvar(name, value) → bool  — Set variable by name",
		"defined":               "defined(name) → bool  — Check if constant is defined",
		"define":                "define(name, value) → null  — Define a constant",
		"func_num_args":         "func_num_args() → int  — Get argument count in current function",
		"func_get_arg":          "func_get_arg(index) → value  — Get argument by index",
		"func_get_args":         "func_get_args() → array  — Get all arguments",
		"function_exists":       "function_exists(name) → bool  — Check if function exists",
		"is_callable":           "is_callable(value) → bool  — Check if value is callable",
		"get_defined_functions": "get_defined_functions() → array  — Get all function names",
		"get_defined_constants": "get_defined_constants() → array  — Get all constant names",
		"jpl_version":           "jpl_version() → string  — Get JPL version",
		"utf8_encode":           "utf8_encode(str) → string  — Encode string to UTF-8 hex",
		"utf8_decode":           "utf8_decode(hex) → string  — Decode UTF-8 hex to string",
	}
}
