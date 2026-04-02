package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterBitwise 注册位运算函数
func RegisterBitwise(e *engine.Engine) {
	e.RegisterFunc("band", builtinBitAnd)
	e.RegisterFunc("bor", builtinBitOr)
	e.RegisterFunc("bxor", builtinBitXor)
	e.RegisterFunc("bnot", builtinBitNot)
	e.RegisterFunc("shl", builtinShl)
	e.RegisterFunc("shr", builtinShr)

	// 模块注册 — import "bitwise" 可用
	e.RegisterModule("bitwise", map[string]engine.GoFunction{
		"band": builtinBitAnd,
		"bor":  builtinBitOr,
		"bxor": builtinBitXor,
		"bnot": builtinBitNot,
		"shl":  builtinShl,
		"shr":  builtinShr,
	})
}

// BitwiseNames 返回位运算函数名
func BitwiseNames() []string {
	return []string{"band", "bor", "bxor", "bnot", "shl", "shr"}
}

// builtinBitAnd 按位与
func builtinBitAnd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("band() requires exactly 2 arguments")
	}
	return engine.NewInt(args[0].Int() & args[1].Int()), nil
}

// builtinBitOr 按位或
func builtinBitOr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("bor() requires exactly 2 arguments")
	}
	return engine.NewInt(args[0].Int() | args[1].Int()), nil
}

// builtinBitXor 按位异或
func builtinBitXor(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("bxor() requires exactly 2 arguments")
	}
	return engine.NewInt(args[0].Int() ^ args[1].Int()), nil
}

// builtinBitNot 按位取反
func builtinBitNot(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bnot() requires exactly 1 argument")
	}
	return engine.NewInt(^args[0].Int()), nil
}

// builtinShl 左移
func builtinShl(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("shl() requires exactly 2 arguments")
	}
	return engine.NewInt(args[0].Int() << uint(args[1].Int())), nil
}

// builtinShr 右移
func builtinShr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("shr() requires exactly 2 arguments")
	}
	return engine.NewInt(args[0].Int() >> uint(args[1].Int())), nil
}
