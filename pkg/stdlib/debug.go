package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterDebug 注册调试相关内置函数
func RegisterDebug(e *engine.Engine) {
	e.RegisterFunc("errors", builtinErrors)
	e.RegisterFunc("last_error", builtinLastError)
	e.RegisterFunc("clear_errors", builtinClearErrors)
}

// DebugNames 返回调试函数名称列表
func DebugNames() []string {
	return []string{"errors", "last_error", "clear_errors"}
}

// errors() 返回所有错误消息列表
func builtinErrors(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("errors() expects 0 arguments, got %d", len(args))
	}
	logs := ctx.Engine().GetErrorLog()
	result := make([]engine.Value, len(logs))
	for i, err := range logs {
		result[i] = engine.NewString(err.Error())
	}
	return engine.NewArray(result), nil
}

// last_error() 返回最后一条错误消息，无错误返回 null
func builtinLastError(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("last_error() expects 0 arguments, got %d", len(args))
	}
	err := ctx.Engine().GetLastError()
	if err == nil {
		return engine.NewNull(), nil
	}
	return engine.NewString(err.Error()), nil
}

// clear_errors() 清空错误日志
func builtinClearErrors(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("clear_errors() expects 0 arguments, got %d", len(args))
	}
	ctx.Engine().ClearErrorLog()
	return engine.NewNull(), nil
}

// DebugSigs returns function signatures for REPL :doc command.
func DebugSigs() map[string]string {
	return map[string]string{
		"errors":       "errors() → array  — Get all error messages",
		"last_error":   "last_error() → string  — Get last error message",
		"clear_errors": "clear_errors() → null  — Clear error log",
	}
}
