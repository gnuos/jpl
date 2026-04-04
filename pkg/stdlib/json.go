package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterJSON 注册 JSON 处理函数
func RegisterJSON(e *engine.Engine) {
	e.RegisterFunc("json_encode", builtinJSONEncode)
	e.RegisterFunc("json_decode", builtinJSONDecode)
	e.RegisterFunc("json_pretty", builtinJSONPretty)
	e.RegisterFunc("json_validate", builtinJSONValidate)
}

// JSONNames 返回 JSON 相关函数名
func JSONNames() []string {
	return []string{"json_encode", "json_decode", "json_pretty", "json_validate"}
}

// builtinJSONEncode 将值序列化为 JSON 字符串
//
// 用法: json_encode($value) -> string
// 用法: json_encode($value, $pretty) -> string
//
// 参数:
//   - $value: 要序列化的值（任意类型）
//   - $pretty: 是否美化输出（可选，默认 false）
//
// 示例:
//
//	json_encode([1, 2, 3])              // "[1,2,3]"
//	json_encode({"name": "Alice"})      // "{"name":"Alice"}"
//	json_encode({"a": 1}, true)         // 格式化输出
func builtinJSONEncode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("json_encode() requires at least 1 argument")
	}

	value := args[0]
	pretty := false
	if len(args) > 1 {
		pretty = args[1].Bool()
	}

	// 转换为 Go 值然后序列化
	goValue := jplValueToGo(value)

	var bytes []byte
	var err error
	if pretty {
		bytes, err = json.MarshalIndent(goValue, "", "  ")
	} else {
		bytes, err = json.Marshal(goValue)
	}

	if err != nil {
		return nil, fmt.Errorf("json_encode() failed: %v", err)
	}

	return engine.NewString(string(bytes)), nil
}

// builtinJSONDecode 将 JSON 字符串解析为值
//
// 用法: json_decode($json_string) -> value
//
// 参数:
//   - $json_string: JSON 格式的字符串
//
// 返回:
//   - 解析后的值（可能是对象、数组、字符串、数字、布尔或 null）
//
// 数字解析策略（内存优化）：
//   - 纯整数：int64 范围内 → Int，超出 → BigInt
//   - 科学计数法：计算后是整数 → 按纯整数处理；有小数 → Float
//   - 普通小数：能精确表示 → Float，否则 → BigDecimal
//
// 示例:
//
//	json_decode("[1, 2, 3]")              // [1, 2, 3]
//	json_decode("{\"name\":\"Alice\"}")   // {"name": "Alice"}
//	json_decode("null")                 // null
//	json_decode("1e20")                   // BigInt(100000000000000000000)
func builtinJSONDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("json_decode() requires exactly 1 argument")
	}

	// 使用 String() 获取原始字符串值，而不是 Stringify()（后者会添加引号和转义）
	jsonStr := args[0].String()

	// 去除首尾空白
	jsonStr = strings.TrimSpace(jsonStr)

	// 使用 Decoder 并启用 UseNumber 以获取原始数字字符串
	dec := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	dec.UseNumber()

	var raw any
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("json_decode() failed: %v", err)
	}

	// 转换回 JPL 值
	result := goValueToJPL(raw)
	return result, nil
}

// builtinJSONPretty 返回美化的 JSON 字符串
//
// 用法: json_pretty($value) -> string
//
// 这是 json_encode($value, true) 的简写形式
func builtinJSONPretty(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("json_pretty() requires at least 1 argument")
	}

	value := args[0]
	goValue := jplValueToGo(value)

	bytes, err := json.MarshalIndent(goValue, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json_pretty() failed: %v", err)
	}

	return engine.NewString(string(bytes)), nil
}

// builtinJSONValidate 验证 JSON 字符串是否合法，不返回解析结果。
func builtinJSONValidate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("json_validate() requires exactly 1 argument")
	}

	jsonStr := args[0].String()
	jsonStr = strings.TrimSpace(jsonStr)

	dec := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	dec.UseNumber()

	var raw any
	err := dec.Decode(&raw)
	if err != nil {
		return engine.NewBool(false), nil
	}
	return engine.NewBool(true), nil
}

// jplValueToGo 将 JPL 值转换为 Go 值
func jplValueToGo(v engine.Value) any {
	switch v.Type() {
	case engine.TypeNull:
		return nil
	case engine.TypeBool:
		return v.Bool()
	case engine.TypeInt:
		return v.Int()
	case engine.TypeFloat:
		return v.Float()
	case engine.TypeString:
		return v.String()
	case engine.TypeArray:
		arr := v.Array()
		result := make([]any, len(arr))
		for i, item := range arr {
			result[i] = jplValueToGo(item)
		}
		return result
	case engine.TypeObject:
		obj := v.Object()
		result := make(map[string]any)
		for key, val := range obj {
			result[key] = jplValueToGo(val)
		}
		return result
	default:
		return v.Stringify()
	}
}

// goValueToJPL 将 Go 值转换为 JPL 值
func goValueToJPL(v any) engine.Value {
	if v == nil {
		return engine.NewNull()
	}

	switch val := v.(type) {
	case bool:
		return engine.NewBool(val)
	case float64:
		// JSON 数字默认是 float64，检查是否为整数
		if val == float64(int64(val)) {
			return engine.NewInt(int64(val))
		}
		return engine.NewFloat(val)
	case json.Number:
		// 使用原始数字字符串进行智能解析
		return parseJSONNumber(string(val))
	case string:
		return engine.NewString(val)
	case []any:
		arr := make([]engine.Value, len(val))
		for i, item := range val {
			arr[i] = goValueToJPL(item)
		}
		return engine.NewArray(arr)
	case map[string]any:
		obj := make(map[string]engine.Value)
		for key, item := range val {
			obj[key] = goValueToJPL(item)
		}
		return engine.NewObject(obj)
	default:
		// 其他类型转为字符串
		return engine.NewString(fmt.Sprintf("%v", val))
	}
}

// scientificNotationRegex 匹配科学计数法格式
var scientificNotationRegex = regexp.MustCompile(`^[+-]?(\d+\.?\d*|\.\d+)[eE][+-]?\d+$`)

// parseJSONNumber 智能解析 JSON 数字字符串
//
// 解析策略（按内存占用优先级）：
// 1. 纯整数：
//   - 在 int64 范围内 → Int
//   - 超出范围 → BigInt
//
// 2. 科学计数法（如 1e10, 1.5e-5）：
//   - 计算后如果是整数且在 int64 范围内 → Int
//   - 计算后如果是整数但超出 int64 范围 → BigInt
//   - 有小数部分 → Float（优先）
//
// 3. 普通小数（如 1.5, 3.14159）：
//   - 能精确表示 → Float
//   - 不能精确表示 → BigDecimal
func parseJSONNumber(s string) engine.Value {
	s = strings.TrimSpace(s)

	// 检查是否为科学计数法
	if scientificNotationRegex.MatchString(s) {
		return parseScientificNotation(s)
	}

	// 检查是否有小数点
	if strings.Contains(s, ".") {
		return parseDecimal(s)
	}

	// 纯整数
	return parseInteger(s)
}

// parseInteger 解析纯整数
func parseInteger(s string) engine.Value {
	// 尝试 int64
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return engine.NewInt(i)
	}

	// 超出 int64 范围，使用 BigInt
	if bi, ok := new(big.Int).SetString(s, 10); ok {
		return engine.NewBigInt(bi)
	}

	// 解析失败，返回字符串
	return engine.NewString(s)
}

// parseDecimal 解析普通小数（无科学计数法）
func parseDecimal(s string) engine.Value {
	// 优先尝试 float64
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		// 检查 float64 是否能精确表示该值
		// 方法：转回字符串，看是否与原字符串在有效数字上一致
		if isFloatPrecise(s, f) {
			return engine.NewFloat(f)
		}
	}

	// 使用 BigDecimal 获得精确表示
	if br, ok := new(big.Rat).SetString(s); ok {
		return engine.NewBigDecimal(br)
	}

	// 解析失败，返回字符串
	return engine.NewString(s)
}

// parseScientificNotation 解析科学计数法
func parseScientificNotation(s string) engine.Value {
	// 使用 big.Rat 直接解析，获得精确值
	br, ok := new(big.Rat).SetString(s)
	if !ok {
		// 解析失败，尝试 float64
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return engine.NewFloat(f)
		}
		return engine.NewString(s)
	}

	// 检查是否为整数
	if br.IsInt() {
		// 尝试转为 int64
		num := br.Num()
		if num.IsInt64() {
			return engine.NewInt(num.Int64())
		}
		// 超出 int64 范围，使用 BigInt
		return engine.NewBigInt(num)
	}

	// 有小数部分
	// 尝试转为 float64（如果能精确表示）
	f, _ := br.Float64()
	if !math.IsInf(f, 0) {
		// 检查是否能精确表示：用高精度格式化后再解析
		// %.17g 是 float64 的精确表示所需精度
		formatted := fmt.Sprintf("%.17g", f)
		if backToRat, ok := new(big.Rat).SetString(formatted); ok && backToRat.Cmp(br) == 0 {
			return engine.NewFloat(f)
		}
	}

	// 使用 BigDecimal 获得精确表示
	return engine.NewBigDecimal(br)
}

// isFloatPrecise 检查 float64 是否能精确表示十进制数
// 通过比较原始字符串和 float64 转回的字符串
func isFloatPrecise(original string, f float64) bool {
	// 处理特殊情况
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return false
	}

	// 移除原始字符串的前导零和末尾零
	original = strings.ToLower(strings.TrimSpace(original))

	// 将 float64 转回字符串
	floatStr := strconv.FormatFloat(f, 'g', -1, 64)
	floatStr = strings.ToLower(floatStr)

	// 如果直接相等，说明精确
	if original == floatStr {
		return true
	}

	// 检查是否科学计数法匹配
	// 例如原始是 "1e10"，float64 转成 "1e+10"
	if strings.Contains(original, "e") && strings.Contains(floatStr, "e") {
		// 提取有效数字和指数进行比较
		origParts := strings.Split(original, "e")
		floatParts := strings.Split(floatStr, "e")
		if len(origParts) == 2 && len(floatParts) == 2 {
			origMantissa := strings.TrimRight(origParts[0], "0")
			origMantissa = strings.TrimLeft(origMantissa, "0")
			if origMantissa == "" || origMantissa == "." {
				origMantissa = "0"
			}

			floatMantissa := strings.TrimRight(floatParts[0], "0")
			floatMantissa = strings.TrimLeft(floatMantissa, "0")
			if floatMantissa == "" || floatMantissa == "." {
				floatMantissa = "0"
			}

			origExp, _ := strconv.Atoi(origParts[1])
			floatExp, _ := strconv.Atoi(floatParts[1])

			// 将指数调整到相同量级比较
			origExp += len(origParts[0]) - strings.Index(origParts[0], ".") - 1
			floatExp += len(floatParts[0]) - strings.Index(floatParts[0], ".") - 1

			if origMantissa == floatMantissa && origExp == floatExp {
				return true
			}
		}
	}

	// 对于普通小数，检查精度损失是否在可接受范围
	// 这里使用保守策略：如果 float64 的精度（约15-16位十进制数字）
	// 能完整表示原始数字的所有有效数字，则认为精确
	return hasPrecisionLoss(original, f)
}

// hasPrecisionLoss 检查从字符串解析为 float64 是否丢失精度
func hasPrecisionLoss(original string, f float64) bool {
	// 移除符号
	original = strings.TrimPrefix(original, "+")
	original = strings.TrimPrefix(original, "-")
	original = strings.ToLower(original)

	// 计算原始字符串的有效数字位数
	significantDigits := countSignificantDigits(original)

	// float64 有约 15-17 位十进制精度
	// 如果有效数字超过 15 位，可能丢失精度
	if significantDigits > 15 {
		return false
	}

	// 将 float64 格式化回字符串，使用足够精度
	formatted := strconv.FormatFloat(f, 'f', significantDigits, 64)
	formatted = strings.ToLower(formatted)

	// 比较（去除末尾的零）
	normalizedOrig := normalizeNumber(original)
	normalizedFmt := normalizeNumber(formatted)

	return normalizedOrig == normalizedFmt
}

// countSignificantDigits 计算有效数字位数
func countSignificantDigits(s string) int {
	s = strings.ToLower(s)

	// 处理科学计数法
	if idx := strings.Index(s, "e"); idx != -1 {
		s = s[:idx]
	}

	count := 0
	foundNonZero := false
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			if ch != '0' {
				foundNonZero = true
			}
			if foundNonZero {
				count++
			}
		} else if ch == '.' {
			continue
		}
	}

	return count
}

// normalizeNumber 规范化数字字符串以便比较
func normalizeNumber(s string) string {
	s = strings.ToLower(s)

	// 处理科学计数法
	if strings.Contains(s, "e") {
		return s
	}

	// 分割整数和小数部分
	parts := strings.Split(s, ".")
	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	// 移除整数部分的前导零
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	// 移除小数部分的末尾零
	fracPart = strings.TrimRight(fracPart, "0")

	if fracPart == "" {
		return intPart
	}

	return intPart + "." + fracPart
}

// JSONSigs returns function signatures for REPL :doc command.
func JSONSigs() map[string]string {
	return map[string]string{
		"json_encode":   "json_encode(value, [pretty]) → string  — Serialize to JSON",
		"json_decode":   "json_decode(str) → value  — Parse JSON string",
		"json_pretty":   "json_pretty(value) → string  — Pretty-print JSON",
		"json_validate": "json_validate(str) → bool  — Validate JSON without parsing",
	}
}
