package stdlib

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"

	"github.com/gnuos/jpl/engine"
)

// RegisterMath 将数学函数注册到引擎。
//
// 注册的函数包括：
//   - abs: 绝对值
//   - ceil/floor/round: 取整函数
//   - pow/sqrt: 幂运算和平方根
//   - min/max: 最值（支持多个参数）
//   - random/randomInt: 随机数
//   - parseInt/parseFloat: 字符串转数字
//   - isNaN/isFinite: 数值检查
//
// 同时注册到 "math" 模块，可通过 import "math" 使用。
//
// 参数：
//   - e: 引擎实例
func RegisterMath(e *engine.Engine) {
	// 全局注册
	e.RegisterFunc("abs", builtinAbs)
	e.RegisterFunc("ceil", builtinCeil)
	e.RegisterFunc("floor", builtinFloor)
	e.RegisterFunc("round", builtinRound)
	e.RegisterFunc("pow", builtinPow)
	e.RegisterFunc("sqrt", builtinSqrt)
	e.RegisterFunc("min", builtinMin)
	e.RegisterFunc("max", builtinMax)
	e.RegisterFunc("random", builtinRandom)
	e.RegisterFunc("randomInt", builtinRandomInt)
	e.RegisterFunc("parseInt", builtinParseInt)
	e.RegisterFunc("parseFloat", builtinParseFloat)
	e.RegisterFunc("isNaN", builtinIsNaN)
	e.RegisterFunc("isFinite", builtinIsFinite)

	// Phase 7.4: 三角函数
	e.RegisterFunc("sin", builtinSin)
	e.RegisterFunc("cos", builtinCos)
	e.RegisterFunc("tan", builtinTan)
	e.RegisterFunc("asin", builtinAsin)
	e.RegisterFunc("acos", builtinAcos)
	e.RegisterFunc("atan", builtinAtan)
	e.RegisterFunc("atan2", builtinAtan2)

	// Phase 7.4: 双曲函数
	e.RegisterFunc("sinh", builtinSinh)
	e.RegisterFunc("cosh", builtinCosh)
	e.RegisterFunc("tanh", builtinTanh)

	// Phase 7.4: 对数/指数
	e.RegisterFunc("log", builtinMathLog)
	e.RegisterFunc("log10", builtinMathLog10)
	e.RegisterFunc("exp", builtinExp)
	e.RegisterFunc("pi", builtinPi)

	// Phase 7.4: 其他数学函数
	e.RegisterFunc("fmod", builtinFmod)
	e.RegisterFunc("hypot", builtinHypot)
	e.RegisterFunc("deg2rad", builtinDeg2rad)
	e.RegisterFunc("rad2deg", builtinRad2deg)

	// Phase 11.3 数学增强
	e.RegisterFunc("rand_str", builtinRandStr)
	e.RegisterFunc("getrandmax", builtinGetRandMax)
	e.RegisterFunc("dechex", builtinDecHex)
	e.RegisterFunc("decoct", builtinDecOct)
	e.RegisterFunc("decbin", builtinDecBin)
	e.RegisterFunc("hexdec", builtinHexDec)
	e.RegisterFunc("bindec", builtinBinDec)
	e.RegisterFunc("octdec", builtinOctDec)
	e.RegisterFunc("base_convert", builtinBaseConvert)

	// Math P0: 缺失函数
	e.RegisterFunc("cbrt", builtinCbrt)
	e.RegisterFunc("log2", builtinLog2)
	e.RegisterFunc("clamp", builtinClamp)
	e.RegisterFunc("sign", builtinSign)
	e.RegisterFunc("intdiv", builtinIntDiv)

	// Math P1: 常用函数
	e.RegisterFunc("trunc", builtinTrunc)
	e.RegisterFunc("factorial", builtinFactorial)
	e.RegisterFunc("gcd", builtinGcd)
	e.RegisterFunc("lcm", builtinLcm)
	e.RegisterFunc("median", builtinMedian)
	e.RegisterFunc("mean", builtinMean)
	e.RegisterFunc("stddev", builtinStddev)
	e.RegisterFunc("modf", builtinModf)

	// 模块注册 — import "math" 可用
	e.RegisterModule("math", map[string]engine.GoFunction{
		"abs": builtinAbs, "ceil": builtinCeil, "floor": builtinFloor, "round": builtinRound,
		"pow": builtinPow, "sqrt": builtinSqrt, "min": builtinMin, "max": builtinMax,
		"random": builtinRandom, "randomInt": builtinRandomInt,
		"parseInt": builtinParseInt, "parseFloat": builtinParseFloat,
		"isNaN": builtinIsNaN, "isFinite": builtinIsFinite,
		// Phase 7.4
		"sin": builtinSin, "cos": builtinCos, "tan": builtinTan,
		"asin": builtinAsin, "acos": builtinAcos, "atan": builtinAtan, "atan2": builtinAtan2,
		"sinh": builtinSinh, "cosh": builtinCosh, "tanh": builtinTanh,
		"log": builtinMathLog, "log10": builtinMathLog10, "exp": builtinExp, "pi": builtinPi,
		"fmod": builtinFmod, "hypot": builtinHypot,
		"deg2rad": builtinDeg2rad, "rad2deg": builtinRad2deg,
		// Phase 11.3
		"rand_str": builtinRandStr, "getrandmax": builtinGetRandMax,
		"dechex": builtinDecHex, "decoct": builtinDecOct, "decbin": builtinDecBin,
		"hexdec": builtinHexDec, "bindec": builtinBinDec, "octdec": builtinOctDec,
		"base_convert": builtinBaseConvert,
		// P0
		"cbrt": builtinCbrt, "log2": builtinLog2, "clamp": builtinClamp,
		"sign": builtinSign, "intdiv": builtinIntDiv,
		// P1
		"trunc": builtinTrunc, "factorial": builtinFactorial, "gcd": builtinGcd,
		"lcm": builtinLcm, "median": builtinMedian, "mean": builtinMean,
		"stddev": builtinStddev, "modf": builtinModf,
	})
}

// MathNames 返回数学函数名称列表。
//
// 返回值：
//   - []string: 数学函数名列表
//
// 包含的函数：
//   - abs, ceil, floor, round（基础运算）
//   - pow, sqrt（幂运算）
//   - min, max（最值）
//   - random, randomInt（随机数）
//   - parseInt, parseFloat（转换）
//   - isNaN, isFinite（检查）
func MathNames() []string {
	return []string{
		// 基础函数
		"abs", "ceil", "floor", "round",
		"pow", "sqrt",
		"min", "max",
		"random", "randomInt",
		"parseInt", "parseFloat",
		"isNaN", "isFinite",
		// Phase 7.4: 三角函数
		"sin", "cos", "tan",
		"asin", "acos", "atan", "atan2",
		// Phase 7.4: 双曲函数
		"sinh", "cosh", "tanh",
		// Phase 7.4: 对数/指数
		"log", "log10", "exp", "pi",
		// Phase 7.4: 其他
		"fmod", "hypot", "deg2rad", "rad2deg",
		// Phase 11.3 数学增强
		"rand_str", "getrandmax",
		"dechex", "decoct", "decbin", "hexdec", "bindec", "octdec", "base_convert",
	}
}

// ============================================================================
// 基础数学函数
// ============================================================================

// builtinAbs 返回数字的绝对值。
//
// 支持 int、float、BigInt、BigDecimal 类型。
// 负数返回其正数形式，正数和零保持不变。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要计算绝对值的数字
//
// 返回值：
//   - 同类型: 绝对值
//   - error: 参数不是数字
//
// 使用示例：
//
//	print abs(-10)          // 输出: 10
//	print abs(10)           // 输出: 10
//	print abs(-3.14)        // 输出: 3.14
//	print abs(0)            // 输出: 0
func builtinAbs(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("abs() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	switch val := v.(type) {
	case *engine.BigIntValue:
		return engine.NewBigInt(new(big.Int).Abs(val.BigInt())), nil
	case *engine.BigDecimalValue:
		return engine.NewBigDecimal(new(big.Rat).Abs(val.BigRat())), nil
	}
	switch v.Type() {
	case engine.TypeInt:
		n := v.Int()
		if n < 0 {
			return engine.NewInt(-n), nil
		}
		return engine.NewInt(n), nil
	case engine.TypeFloat:
		return engine.NewFloat(math.Abs(v.Float())), nil
	default:
		return nil, fmt.Errorf("abs() expects a number, got %s", v.Type())
	}
}

// ceil(n) 向上取整 — 支持 int/float/BigDecimal
func builtinCeil(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ceil() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	if _, ok := v.(*engine.BigDecimalValue); ok {
		f := v.Float()
		return engine.NewBigInt(big.NewInt(int64(math.Ceil(f)))), nil
	}
	switch v.Type() {
	case engine.TypeInt:
		return v, nil
	case engine.TypeFloat:
		return engine.NewInt(int64(math.Ceil(v.Float()))), nil
	default:
		return nil, fmt.Errorf("ceil() expects a number, got %s", v.Type())
	}
}

// floor(n) 向下取整 — 支持 int/float/BigDecimal
func builtinFloor(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	if _, ok := v.(*engine.BigDecimalValue); ok {
		f := v.Float()
		return engine.NewBigInt(big.NewInt(int64(math.Floor(f)))), nil
	}
	switch v.Type() {
	case engine.TypeInt:
		return v, nil
	case engine.TypeFloat:
		return engine.NewInt(int64(math.Floor(v.Float()))), nil
	default:
		return nil, fmt.Errorf("floor() expects a number, got %s", v.Type())
	}
}

// round(n, precision) 四舍五入 — 支持 int/float/BigDecimal
func builtinRound(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("round() expects 1-2 arguments, got %d", len(args))
	}
	v := args[0]

	precision := 0
	if len(args) == 2 {
		precision = int(args[1].Int())
	}

	// BigDecimal
	if bv, ok := v.(*engine.BigDecimalValue); ok {
		if precision == 0 {
			rat := bv.BigRat()
			// 四舍五入到整数
			f, _ := rat.Float64()
			return engine.NewBigInt(big.NewInt(int64(math.Round(f)))), nil
		}
		// 带精度四舍五入
		f, _ := bv.BigRat().Float64()
		mult := math.Pow(10, float64(precision))
		return engine.NewFloat(math.Round(f*mult) / mult), nil
	}

	if v.Type() == engine.TypeInt {
		return v, nil
	}
	if v.Type() != engine.TypeFloat {
		return nil, fmt.Errorf("round() expects a number, got %s", v.Type())
	}

	f := v.Float()
	if precision == 0 {
		return engine.NewInt(int64(math.Round(f))), nil
	}
	mult := math.Pow(10, float64(precision))
	return engine.NewFloat(math.Round(f*mult) / mult), nil
}

// pow(base, exp) 幂运算 — BigInt 底数 + 整数指数返回精确 BigInt
func builtinPow(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pow() expects 2 arguments, got %d", len(args))
	}
	base := args[0]
	exp := args[1]

	// BigInt 底数 + 整数指数 → 精确 BigInt
	if bv, ok := base.(*engine.BigIntValue); ok && exp.Type() == engine.TypeInt {
		e := exp.Int()
		if e < 0 {
			return engine.NewFloat(math.Pow(base.Float(), exp.Float())), nil
		}
		result := new(big.Int).Exp(bv.BigInt(), big.NewInt(e), nil)
		return engine.NewBigInt(result), nil
	}

	// BigDecimal 底数 + 整数指数
	if bv, ok := base.(*engine.BigDecimalValue); ok && exp.Type() == engine.TypeInt {
		e := exp.Int()
		if e == 0 {
			return engine.NewBigDecimal(new(big.Rat).SetInt64(1)), nil
		}
		result := new(big.Rat).Set(bv.BigRat())
		if e < 0 {
			result.Inv(result)
			e = -e
		}
		// 连乘
		for i := int64(1); i < e; i++ {
			result.Mul(result, bv.BigRat())
		}
		return engine.NewBigDecimal(result), nil
	}

	// 通用 float 计算
	if (base.Type() != engine.TypeInt && base.Type() != engine.TypeFloat) ||
		(exp.Type() != engine.TypeInt && exp.Type() != engine.TypeFloat) {
		return nil, fmt.Errorf("pow() expects numbers")
	}
	return engine.NewFloat(math.Pow(base.Float(), exp.Float())), nil
}

// sqrt(n) 平方根
func builtinSqrt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sqrt() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	if v.Type() != engine.TypeInt && v.Type() != engine.TypeFloat {
		return nil, fmt.Errorf("sqrt() expects a number, got %s", v.Type())
	}
	return engine.NewFloat(math.Sqrt(v.Float())), nil
}

// ============================================================================
// 比较函数
// ============================================================================

// min(a, b, ...) 返回最小值
func builtinMin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("min() expects at least 1 argument")
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i].Less(result) {
			result = args[i]
		}
	}
	return result, nil
}

// max(a, b, ...) 返回最大值
func builtinMax(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("max() expects at least 1 argument")
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i].Greater(result) {
			result = args[i]
		}
	}
	return result, nil
}

// ============================================================================
// 随机数函数
// ============================================================================

// random() 返回 [0, 1) 之间的随机浮点数
func builtinRandom(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("random() expects 0 arguments, got %d", len(args))
	}
	return engine.NewFloat(rand.Float64()), nil
}

// randomInt(min, max) 返回 [min, max] 之间的随机整数
func builtinRandomInt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("randomInt() expects 2 arguments, got %d", len(args))
	}
	minVal := args[0].Int()
	maxVal := args[1].Int()
	if minVal > maxVal {
		return nil, fmt.Errorf("randomInt() min must be <= max")
	}
	if minVal == maxVal {
		return engine.NewInt(minVal), nil
	}
	return engine.NewInt(minVal + rand.Int63n(maxVal-minVal+1)), nil
}

// ============================================================================
// 类型转换函数
// ============================================================================

// builtinParseInt 将字符串或数字转换为整数。
//
// 支持以下类型：
//   - 字符串：解析整数字符串（支持十进制、十六进制 0x、八进制 0o、二进制 0b）
//   - int：直接返回
//   - float：截断小数部分
//
// 如果解析失败，返回 null（不是错误）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要转换的值
//
// 返回值：
//   - int: 转换后的整数
//   - null: 转换失败
//
// 使用示例：
//
//	print parseInt("42")          // 42
//	print parseInt("0xFF")        // 255（十六进制）
//	print parseInt("0b1010")    // 10（二进制）
//	print parseInt(3.14)         // 3（截断）
//	print parseInt("abc")        // null（失败）
func builtinParseInt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parseInt() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	switch v.Type() {
	case engine.TypeInt:
		return v, nil
	case engine.TypeFloat:
		return engine.NewInt(v.Int()), nil
	case engine.TypeString:
		s := v.String()
		n, err := strconv.ParseInt(s, 0, 64)
		if err != nil {
			return engine.NewNull(), nil
		}
		return engine.NewInt(n), nil
	default:
		return engine.NewNull(), nil
	}
}

// builtinParseFloat 将字符串或数字转换为浮点数。
//
// 支持以下类型：
//   - 字符串：解析为浮点数
//   - float：直接返回
//   - int：转为浮点数
//
// 如果解析失败，返回 null（不是错误）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要转换的值
//
// 返回值：
//   - float: 转换后的浮点数
//   - null: 转换失败
//
// 使用示例：
//
//	print parseFloat("3.14")      // 3.14
//	print parseFloat("-0.5")      // -0.5
//	print parseFloat(42)          // 42.0
//	print parseFloat("1e3")       // 1000.0（科学计数法）
//	print parseFloat("abc")        // null（失败）
func builtinParseFloat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parseFloat() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	switch v.Type() {
	case engine.TypeFloat:
		return v, nil
	case engine.TypeInt:
		return engine.NewFloat(v.Float()), nil
	case engine.TypeString:
		s := v.String()
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return engine.NewNull(), nil
		}
		return engine.NewFloat(f), nil
	default:
		return engine.NewNull(), nil
	}
}

// ============================================================================
// 检查函数
// ============================================================================

// isNaN(v) 检查值是否为 NaN
func builtinIsNaN(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isNaN() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	if v.Type() == engine.TypeFloat {
		return engine.NewBool(math.IsNaN(v.Float())), nil
	}
	return engine.NewBool(false), nil
}

// isFinite(v) 检查值是否为有限数
func builtinIsFinite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isFinite() expects 1 argument, got %d", len(args))
	}
	v := args[0]
	switch v.Type() {
	case engine.TypeInt:
		return engine.NewBool(true), nil
	case engine.TypeFloat:
		return engine.NewBool(!math.IsNaN(v.Float()) && !math.IsInf(v.Float(), 0)), nil
	default:
		return engine.NewBool(false), nil
	}
}

// ============================================================================
// Phase 7.4: 扩展数学函数
// ============================================================================

// builtinSin 返回角度的正弦值（弧度）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 弧度值
//
// 返回值：
//   - float64: 正弦值（-1 到 1）
//   - error: 参数错误
//
// 使用示例：
//
//	sin(0)                   // → 0
//	sin(PI / 2)              // → 1
func builtinSin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sin() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Sin(args[0].Float())), nil
}

// builtinCos 返回角度的余弦值（弧度）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 弧度值
//
// 返回值：
//   - float64: 余弦值（-1 到 1）
//   - error: 参数错误
//
// 使用示例：
//
//	cos(0)                   // → 1
//	cos(PI)                  // → -1
func builtinCos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cos() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Cos(args[0].Float())), nil
}

// builtinTan 返回角度的正切值
func builtinTan(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tan() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Tan(args[0].Float())), nil
}

// builtinAsin 返回反正弦值（弧度）
func builtinAsin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("asin() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Asin(args[0].Float())), nil
}

// builtinAcos 返回反余弦值（弧度）
func builtinAcos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("acos() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Acos(args[0].Float())), nil
}

// builtinAtan 返回反正切值（弧度）
func builtinAtan(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("atan() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Atan(args[0].Float())), nil
}

// builtinAtan2 返回 y/x 的反正切值（考虑象限）
func builtinAtan2(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("atan2() expects 2 arguments, got %d", len(args))
	}
	y := args[0].Float()
	x := args[1].Float()
	return engine.NewFloat(math.Atan2(y, x)), nil
}

// builtinSinh 返回双曲正弦值
func builtinSinh(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sinh() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Sinh(args[0].Float())), nil
}

// builtinCosh 返回双曲余弦值
func builtinCosh(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cosh() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Cosh(args[0].Float())), nil
}

// builtinTanh 返回双曲正切值
func builtinTanh(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tanh() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Tanh(args[0].Float())), nil
}

// builtinLog 返回自然对数（以 e 为底）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 正数
//
// 返回值：
//   - float64: 自然对数值
//   - error: 参数错误
//
// 使用示例：
//
//	log(1)                   // → 0
//	log(E)                   // → 1
func builtinMathLog(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("log() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Log(args[0].Float())), nil
}

// builtinLog10 返回常用对数（以 10 为底）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 正数
//
// 返回值：
//   - float64: 常用对数值
//   - error: 参数错误
//
// 使用示例：
//
//	log10(1)                 // → 0
//	log10(10)                // → 1
//	log10(100)               // → 2
func builtinMathLog10(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("log10() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Log10(args[0].Float())), nil
}

// builtinExp 返回 e 的指数幂
func builtinExp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exp() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Exp(args[0].Float())), nil
}

// builtinPi 返回圆周率 PI（函数形式）
func builtinPi(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	return engine.NewFloat(math.Pi), nil
}

// builtinFmod 返回浮点数除法的余数
func builtinFmod(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("fmod() expects 2 arguments, got %d", len(args))
	}
	x := args[0].Float()
	y := args[1].Float()
	return engine.NewFloat(math.Mod(x, y)), nil
}

// builtinHypot 返回直角三角形斜边长度
func builtinHypot(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("hypot() expects 2 arguments, got %d", len(args))
	}
	x := args[0].Float()
	y := args[1].Float()
	return engine.NewFloat(math.Hypot(x, y)), nil
}

// builtinDeg2rad 将角度转换为弧度
func builtinDeg2rad(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("deg2rad() expects 1 argument, got %d", len(args))
	}
	deg := args[0].Float()
	return engine.NewFloat(deg * math.Pi / 180), nil
}

// builtinRad2deg 将弧度转换为角度
func builtinRad2deg(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rad2deg() expects 1 argument, got %d", len(args))
	}
	rad := args[0].Float()
	return engine.NewFloat(rad * 180 / math.Pi), nil
}

func builtinRandStr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	length := 16
	if len(args) >= 1 {
		length = int(args[0].Int())
	}
	if length <= 0 {
		length = 16
	}

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return engine.NewString(string(result)), nil
}

func builtinGetRandMax(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	return engine.NewInt(2147483647), nil
}

func builtinDecHex(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dechex() expects 1 argument, got %d", len(args))
	}

	num := args[0].Int()
	return engine.NewString(fmt.Sprintf("%x", num)), nil
}

func builtinDecOct(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("decoct() expects 1 argument, got %d", len(args))
	}

	num := args[0].Int()
	return engine.NewString(fmt.Sprintf("%o", num)), nil
}

func builtinDecBin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("decbin() expects 1 argument, got %d", len(args))
	}

	num := args[0].Int()
	return engine.NewString(fmt.Sprintf("%b", num)), nil
}

func builtinHexDec(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("hexdec() expects 1 argument, got %d", len(args))
	}

	hexStr := args[0].String()
	num, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		return engine.NewInt(0), nil
	}

	return engine.NewInt(num), nil
}

func builtinBinDec(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bindec() expects 1 argument, got %d", len(args))
	}

	binStr := args[0].String()
	num, err := strconv.ParseInt(binStr, 2, 64)
	if err != nil {
		return engine.NewInt(0), nil
	}

	return engine.NewInt(num), nil
}

func builtinOctDec(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("octdec() expects 1 argument, got %d", len(args))
	}

	octStr := args[0].String()
	num, err := strconv.ParseInt(octStr, 8, 64)
	if err != nil {
		return engine.NewInt(0), nil
	}

	return engine.NewInt(num), nil
}

func builtinBaseConvert(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("base_convert() expects 3 arguments, got %d", len(args))
	}

	numStr := args[0].String()
	fromBase := int(args[1].Int())
	toBase := int(args[2].Int())

	num, err := strconv.ParseInt(numStr, fromBase, 64)
	if err != nil {
		return engine.NewString("0"), nil
	}

	switch toBase {
	case 16:
		return engine.NewString(fmt.Sprintf("%x", num)), nil
	case 8:
		return engine.NewString(fmt.Sprintf("%o", num)), nil
	case 2:
		return engine.NewString(fmt.Sprintf("%b", num)), nil
	default:
		return engine.NewString(fmt.Sprintf("%d", num)), nil
	}
}

// builtinCbrt 计算立方根。
func builtinCbrt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cbrt() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Cbrt(args[0].Float())), nil
}

// builtinLog2 计算以 2 为底的对数。
func builtinLog2(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("log2() expects 1 argument, got %d", len(args))
	}
	return engine.NewFloat(math.Log2(args[0].Float())), nil
}

// builtinClamp 将值限制在 [min, max] 范围内。
func builtinClamp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("clamp() expects 3 arguments (value, min, max), got %d", len(args))
	}

	val := args[0].Float()
	minVal := args[1].Float()
	maxVal := args[2].Float()

	if val < minVal {
		val = minVal
	} else if val > maxVal {
		val = maxVal
	}

	if val == float64(int64(val)) {
		return engine.NewInt(int64(val)), nil
	}
	return engine.NewFloat(val), nil
}

// builtinSign 返回数字的符号：-1（负数）、0（零）、1（正数）。
func builtinSign(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sign() expects 1 argument, got %d", len(args))
	}

	val := args[0].Float()
	if val < 0 {
		return engine.NewInt(-1), nil
	} else if val > 0 {
		return engine.NewInt(1), nil
	}
	return engine.NewInt(0), nil
}

// builtinIntDiv 整数除法。
func builtinIntDiv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("intdiv() expects 2 arguments (dividend, divisor), got %d", len(args))
	}

	a := args[0].Int()
	b := args[1].Int()

	if b == 0 {
		return nil, fmt.Errorf("intdiv() division by zero")
	}

	return engine.NewInt(a / b), nil
}

// builtinTrunc 向零取整。
func builtinTrunc(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trunc() expects 1 argument, got %d", len(args))
	}
	return engine.NewInt(int64(math.Trunc(args[0].Float()))), nil
}

// builtinFactorial 计算阶乘。
func builtinFactorial(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("factorial() expects 1 argument, got %d", len(args))
	}
	n := int(args[0].Int())
	if n < 0 {
		return nil, fmt.Errorf("factorial() argument must be non-negative")
	}
	result := int64(1)
	for i := 2; i <= n; i++ {
		result *= int64(i)
	}
	return engine.NewInt(result), nil
}

// builtinGcd 计算最大公约数。
func builtinGcd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("gcd() expects 2 arguments, got %d", len(args))
	}
	a, b := args[0].Int(), args[1].Int()
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return engine.NewInt(a), nil
}

// builtinLcm 计算最小公倍数。
func builtinLcm(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("lcm() expects 2 arguments, got %d", len(args))
	}
	a, b := args[0].Int(), args[1].Int()
	if a == 0 || b == 0 {
		return engine.NewInt(0), nil
	}
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	return engine.NewInt(a / (a % b) * b), nil
}

// builtinMedian 计算中位数。
func builtinMedian(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("median() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("median() argument must be array, got %s", args[0].Type())
	}
	arr := args[0].Array()
	if len(arr) == 0 {
		return nil, fmt.Errorf("median() array must not be empty")
	}
	vals := make([]float64, len(arr))
	for i, v := range arr {
		vals[i] = v.Float()
	}
	// 插入排序
	for i := 1; i < len(vals); i++ {
		for j := i; j > 0 && vals[j] < vals[j-1]; j-- {
			vals[j], vals[j-1] = vals[j-1], vals[j]
		}
	}
	n := len(vals)
	if n%2 == 1 {
		return engine.NewFloat(vals[n/2]), nil
	}
	return engine.NewFloat((vals[n/2-1] + vals[n/2]) / 2), nil
}

// builtinMean 计算算术平均值。
func builtinMean(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mean() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("mean() argument must be array, got %s", args[0].Type())
	}
	arr := args[0].Array()
	if len(arr) == 0 {
		return nil, fmt.Errorf("mean() array must not be empty")
	}
	sum := 0.0
	for _, v := range arr {
		sum += v.Float()
	}
	return engine.NewFloat(sum / float64(len(arr))), nil
}

// builtinStddev 计算标准差。
func builtinStddev(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("stddev() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("stddev() argument must be array, got %s", args[0].Type())
	}
	arr := args[0].Array()
	if len(arr) == 0 {
		return nil, fmt.Errorf("stddev() array must not be empty")
	}
	n := float64(len(arr))
	sum := 0.0
	for _, v := range arr {
		sum += v.Float()
	}
	mean := sum / n
	variance := 0.0
	for _, v := range arr {
		diff := v.Float() - mean
		variance += diff * diff
	}
	variance /= n
	return engine.NewFloat(math.Sqrt(variance)), nil
}

// builtinModf 返回整数和小数部分。
func builtinModf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("modf() expects 1 argument, got %d", len(args))
	}
	intPart, fracPart := math.Modf(args[0].Float())
	return engine.NewArray([]engine.Value{
		engine.NewFloat(fracPart),
		engine.NewFloat(intPart),
	}), nil
}

// MathSigs returns function signatures for REPL :doc command.
func MathSigs() map[string]string {
	return map[string]string{
		"abs":          "abs(n) → number  — Absolute value",
		"ceil":         "ceil(n) → int  — Round up",
		"floor":        "floor(n) → int  — Round down",
		"round":        "round(n, [precision]) → number  — Round to precision",
		"sqrt":         "sqrt(n) → float  — Square root",
		"pow":          "pow(base, exp) → number  — Power",
		"min":          "min(a, b, ...) → number  — Minimum value",
		"max":          "max(a, b, ...) → number  — Maximum value",
		"rand":         "rand() → float  — Random float [0, 1)",
		"rand_int":     "rand_int(min, max) → int  — Random integer in range",
		"rand_str":     "rand_str([length]) → string  — Random alphanumeric string",
		"getrandmax":   "getrandmax() → int  — Maximum random value",
		"dechex":       "dechex(n) → string  — Decimal to hex",
		"decoct":       "decoct(n) → string  — Decimal to octal",
		"decbin":       "decbin(n) → string  — Decimal to binary",
		"hexdec":       "hexdec(str) → int  — Hex to decimal",
		"bindec":       "bindec(str) → int  — Binary to decimal",
		"octdec":       "octdec(str) → int  — Octal to decimal",
		"base_convert": "base_convert(num, from_base, to_base) → string  — Convert between bases",
		"sin":          "sin(radians) → float  — Sine",
		"cos":          "cos(radians) → float  — Cosine",
		"tan":          "tan(radians) → float  — Tangent",
		"asin":         "asin(n) → float  — Arc sine",
		"acos":         "acos(n) → float  — Arc cosine",
		"atan":         "atan(n) → float  — Arc tangent",
		"atan2":        "atan2(y, x) → float  — Arc tangent of y/x",
		"sinh":         "sinh(n) → float  — Hyperbolic sine",
		"cosh":         "cosh(n) → float  — Hyperbolic cosine",
		"tanh":         "tanh(n) → float  — Hyperbolic tangent",
		"log":          "log(n) → float  — Natural logarithm",
		"log10":        "log10(n) → float  — Base-10 logarithm",
		"exp":          "exp(n) → float  — e raised to power n",
		"pi":           "pi() → float  — Return PI constant",
		"fmod":         "fmod(x, y) → float  — Floating point modulo",
		"hypot":        "hypot(x, y) → float  — Hypotenuse",
		"deg2rad":      "deg2rad(degrees) → float  — Degrees to radians",
		"rad2deg":      "rad2deg(radians) → float  — Radians to degrees",
		"cbrt":         "cbrt(n) → float  — Cube root",
		"log2":         "log2(n) → float  — Base-2 logarithm",
		"clamp":        "clamp(value, min, max) → number  — Clamp value between min and max",
		"sign":         "sign(n) → int  — Sign of number (-1, 0, 1)",
		"intdiv":       "intdiv(a, b) → int  — Integer division",
	}
}
