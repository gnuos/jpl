package stdlib

import (
	"fmt"
	"math/bits"

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

	// P1
	e.RegisterFunc("bit_count", builtinBitCount)
	e.RegisterFunc("bit_length", builtinBitLength)

	// 模块注册 — import "bitwise" 可用
	e.RegisterModule("bitwise", map[string]engine.GoFunction{
		"band":       builtinBitAnd,
		"bor":        builtinBitOr,
		"bxor":       builtinBitXor,
		"bnot":       builtinBitNot,
		"shl":        builtinShl,
		"shr":        builtinShr,
		"bit_count":  builtinBitCount,
		"bit_length": builtinBitLength,
	})
}

// BitwiseNames 返回位运算函数名
func BitwiseNames() []string {
	return []string{"band", "bor", "bxor", "bnot", "shl", "shr", "bit_count", "bit_length"}
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

// BitwiseSigs returns function signatures for REPL :doc command.

// builtinBitCount 计算设置位的数量。
func builtinBitCount(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bit_count() expects 1 argument, got %d", len(args))
	}
	n := args[0].Int()
	return engine.NewInt(int64(bits.OnesCount64(uint64(n)))), nil
}

// builtinBitLength 计算表示数字所需的位数。
func builtinBitLength(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bit_length() expects 1 argument, got %d", len(args))
	}
	n := args[0].Int()
	if n < 0 {
		n = -n
	}
	if n == 0 {
		return engine.NewInt(0), nil
	}
	return engine.NewInt(int64(64 - bits.LeadingZeros64(uint64(n)))), nil
}
func BitwiseSigs() map[string]string {
	return map[string]string{
		"bit_and":    "bit_and(a, b) → int  — Bitwise AND",
		"bit_or":     "bit_or(a, b) → int  — Bitwise OR",
		"bit_xor":    "bit_xor(a, b) → int  — Bitwise XOR",
		"bit_not":    "bit_not(a) → int  — Bitwise NOT",
		"bit_shl":    "bit_shl(a, b) → int  — Bitwise left shift",
		"bit_shr":    "bit_shr(a, b) → int  — Bitwise right shift",
		"bit_count":  "bit_count(n) → int  — Count number of set bits",
		"bit_length": "bit_length(n) → int  — Number of bits to represent integer",
	}
}
