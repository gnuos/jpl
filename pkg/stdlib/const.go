package stdlib

import (
	"fmt"
	"math"
	"runtime"

	"github.com/gnuos/jpl/engine"
)

// RegisterPresetConstants 注册预设数学常量和特殊值
func RegisterPresetConstants(e *engine.Engine) {
	// 特殊值
	e.RegisterConst("INF", engine.NewFloat(math.Inf(1)))
	e.RegisterConst("NaN", engine.NewFloat(math.NaN()))

	// 数学常量
	e.RegisterConst("PI", engine.NewFloat(math.Pi))
	e.RegisterConst("TAU", engine.NewFloat(2*math.Pi))
	e.RegisterConst("E", engine.NewFloat(math.E))
	e.RegisterConst("SQRT2", engine.NewFloat(math.Sqrt2))
	e.RegisterConst("LN2", engine.NewFloat(math.Ln2))
	e.RegisterConst("LN10", engine.NewFloat(math.Ln10))

	// 平台常量
	e.RegisterConst("EOL", engine.NewString(getPlatformEOL()))

	// 标准 IO 流
	// 流资源类型，可用于 fopen/fread/fwrite/fclose 等 IO 函数
	e.RegisterConst("STDIN", engine.NewStdinStream())
	e.RegisterConst("STDOUT", engine.NewStdoutStream())
	e.RegisterConst("STDERR", engine.NewStderrStream())
}

// getPlatformEOL 返回当前平台的换行符
// Windows: "\r\n", Unix/Linux/macOS: "\n"
func getPlatformEOL() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

// RegisterConstFunc 注册动态常量函数
func RegisterConstFunc(e *engine.Engine) {
	e.RegisterFunc("define", builtinDefine)
	e.RegisterFunc("defined", builtinDefined)
}

// ConstFuncNames 返回动态常量函数名称列表
func ConstFuncNames() []string {
	return []string{"define", "defined"}
}

// builtinDefine 定义一个常量
// define(name, value) — name 必须为字符串
// 重复定义同名常量将报错
func builtinDefine(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("define() expects 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("define() argument 1 must be a string, got %s", args[0].Type())
	}

	name := args[0].String()
	value := args[1]

	// 检查是否已定义
	if _, ok := ctx.Engine().GetConst(name); ok {
		return nil, fmt.Errorf("constant %q already defined", name)
	}

	if err := ctx.Engine().RegisterConst(name, value); err != nil {
		return nil, err
	}

	return engine.NewNull(), nil
}

// builtinDefined 检查常量是否已定义
// defined(name) — name 必须为字符串，返回 bool
func builtinDefined(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("defined() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("defined() argument must be a string, got %s", args[0].Type())
	}

	name := args[0].String()
	_, ok := ctx.Engine().GetConst(name)
	return engine.NewBool(ok), nil
}

// ConstFuncSigs returns function signatures for REPL :doc command.
func ConstFuncSigs() map[string]string {
	return map[string]string{
		"define":  "define(name, value) → null  — Define a constant",
		"defined": "defined(name) → bool  — Check if constant is defined",
	}
}
