package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterError 注册错误相关内置函数
func RegisterError(e *engine.Engine) {
	e.RegisterFunc("error", builtinError)
}

// ErrorNames 返回错误函数名称列表
func ErrorNames() []string {
	return []string{"error"}
}

// error(message) / error(message, code) / error(message, code, type)
// 创建结构化错误对象，用于 throw 抛出
func builtinError(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("error() expects 1-3 arguments, got %d", len(args))
	}

	message := args[0].String()
	var code int64
	var errType string

	if len(args) >= 2 {
		code = args[1].Int()
	}
	if len(args) >= 3 {
		errType = args[2].String()
	}

	return engine.NewError(message, code, errType), nil
}

// ErrorSigs returns function signatures for REPL :doc command.
func ErrorSigs() map[string]string {
	return map[string]string{
		"error": "error(message, [code], [type]) → error  — Create error object",
		"throw": "throw(error)  — Throw error (language construct)",
	}
}
