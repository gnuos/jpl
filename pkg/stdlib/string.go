package stdlib

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gnuos/jpl/engine"
)

// RegisterString 将字符串操作函数注册到引擎。
//
// 注册的函数包括：
//   - strlen: 字符串长度
//   - substr: 截取子串
//   - strpos: 查找子串位置
//   - str_replace: 替换子串
//   - trim: 去除空白
//   - toUpper/toLower: 大小写转换
//   - split: 分割字符串
//   - join: 连接字符串
//   - startsWith/endsWith: 前缀/后缀检查
//   - charAt: 获取字符
//   - repeat: 重复字符串
//   - reverse: 反转字符串
//   - Phase 7.3 新增：
//   - implode/explode: join/split 别名
//   - ltrim/rtrim/chop: 修剪函数
//   - strcmp/strcasecmp/strncmp/strncasecmp: 字符串比较
//   - stripos/strrpos/strripos/strstr/stristr/strchr: 查找函数
//   - sprintf/printf/vsprintf/vprintf: 格式化输出
//   - ord/chr: ASCII 转换
//   - nl2br: 换行转 HTML
//   - bin2hex: 二进制转十六进制
//
// 同时注册到 "strings" 模块，可通过 import "strings" 使用。
//
// 参数：
//   - e: 引擎实例
func RegisterString(e *engine.Engine) {
	// 全局注册
	e.RegisterFunc("strlen", builtinStrlen)
	e.RegisterFunc("substr", builtinSubstr)
	e.RegisterFunc("strpos", builtinStrpos)
	e.RegisterFunc("str_replace", builtinStrReplace)
	e.RegisterFunc("trim", builtinTrim)
	e.RegisterFunc("toUpper", builtinToUpper)
	e.RegisterFunc("toLower", builtinToLower)
	e.RegisterFunc("split", builtinSplit)
	e.RegisterFunc("join", builtinJoin)
	e.RegisterFunc("startsWith", builtinStartsWith)
	e.RegisterFunc("endsWith", builtinEndsWith)
	e.RegisterFunc("charAt", builtinCharAt)
	e.RegisterFunc("repeat", builtinRepeat)
	e.RegisterFunc("reverse", builtinReverse)

	// Phase 7.3 新增
	// 别名函数
	e.RegisterFunc("implode", builtinJoin)  // join 别名
	e.RegisterFunc("explode", builtinSplit) // split 别名
	e.RegisterFunc("chop", builtinRtrim)    // rtrim 别名
	// trim 系列
	e.RegisterFunc("ltrim", builtinLtrim)
	e.RegisterFunc("rtrim", builtinRtrim)
	// 字符串比较
	e.RegisterFunc("strcmp", builtinStrcmp)
	e.RegisterFunc("strcasecmp", builtinStrcasecmp)
	e.RegisterFunc("strncmp", builtinStrncmp)
	e.RegisterFunc("strncasecmp", builtinStrncasecmp)
	// 查找函数
	e.RegisterFunc("stripos", builtinStripos)
	e.RegisterFunc("strrpos", builtinStrrpos)
	e.RegisterFunc("strripos", builtinStrripos)
	e.RegisterFunc("strstr", builtinStrstr)
	e.RegisterFunc("stristr", builtinStristr)
	e.RegisterFunc("strchr", builtinStrchr)
	// 格式化
	e.RegisterFunc("sprintf", builtinSprintf)
	e.RegisterFunc("printf", builtinPrintf)
	e.RegisterFunc("vsprintf", builtinVsprintf)
	e.RegisterFunc("vprintf", builtinVprintf)
	e.RegisterFunc("number_format", builtinNumberFormat)
	// 其他
	e.RegisterFunc("ord", builtinOrd)
	e.RegisterFunc("chr", builtinChr)
	e.RegisterFunc("nl2br", builtinNl2br)
	e.RegisterFunc("bin2hex", builtinBin2hex)
	e.RegisterFunc("hex2bin", builtinHex2bin)

	// Phase 11.1 字符串增强
	e.RegisterFunc("substr_compare", builtinSubstrCompare)
	e.RegisterFunc("substr_count", builtinSubstrCount)
	e.RegisterFunc("str_repeat", builtinStrRepeat)
	e.RegisterFunc("str_pad", builtinStrPad)
	e.RegisterFunc("str_split", builtinStrSplit)
	e.RegisterFunc("strrev", builtinStrrev)
	e.RegisterFunc("htmlspecialchars", builtinHtmlspecialchars)
	e.RegisterFunc("htmlspecialchars_decode", builtinHtmlspecialcharsDecode)
	e.RegisterFunc("strip_tags", builtinStripTags)
	e.RegisterFunc("wordwrap", builtinWordwrap)
	e.RegisterFunc("strtolower", builtinStrtolower)
	e.RegisterFunc("strtoupper", builtinStrtoupper)
	e.RegisterFunc("chunk_split", builtinChunkSplit)

	// 转义编码函数
	e.RegisterFunc("addslashes", builtinAddslashes)
	e.RegisterFunc("stripslashes", builtinStripslashes)
	e.RegisterFunc("addcslashes", builtinAddcslashes)

	// 模块注册 — import "strings" 可用
	e.RegisterModule("strings", map[string]engine.GoFunction{
		"strlen": builtinStrlen, "substr": builtinSubstr, "strpos": builtinStrpos,
		"str_replace": builtinStrReplace, "trim": builtinTrim,
		"toUpper": builtinToUpper, "toLower": builtinToLower,
		"split": builtinSplit, "join": builtinJoin,
		"startsWith": builtinStartsWith, "endsWith": builtinEndsWith,
		"charAt": builtinCharAt, "repeat": builtinRepeat, "reverse": builtinReverse,
		// Phase 7.3
		"implode": builtinJoin, "explode": builtinSplit, "chop": builtinRtrim,
		"ltrim": builtinLtrim, "rtrim": builtinRtrim,
		"strcmp": builtinStrcmp, "strcasecmp": builtinStrcasecmp,
		"strncmp": builtinStrncmp, "strncasecmp": builtinStrncasecmp,
		"stripos": builtinStripos, "strrpos": builtinStrrpos, "strripos": builtinStrripos,
		"strstr": builtinStrstr, "stristr": builtinStristr, "strchr": builtinStrchr,
		"sprintf": builtinSprintf, "printf": builtinPrintf,
		"vsprintf": builtinVsprintf, "vprintf": builtinVprintf,
		"number_format": builtinNumberFormat,
		"ord":           builtinOrd, "chr": builtinChr,
		"nl2br": builtinNl2br, "bin2hex": builtinBin2hex, "hex2bin": builtinHex2bin,
		// 转义编码函数
		"addslashes": builtinAddslashes, "stripslashes": builtinStripslashes,
		"addcslashes": builtinAddcslashes,
		// Email编码函数 (quoted-printable)
		"quoted_printable_encode": builtinQuotedPrintableEncode,
		"quoted_printable_decode": builtinQuotedPrintableDecode,
		// HTML实体编码函数
		"htmlentities":               builtinHtmlEntities,
		"html_entity_decode":         builtinHtmlEntityDecode,
		"get_html_translation_table": builtinGetHtmlTranslationTable,
	})
}

// StringNames 返回字符串函数名称列表。
//
// 返回值：
//   - []string: 字符串函数名列表
//
// 包含的函数：
//   - strlen, substr, strpos, str_replace（核心操作）
//   - trim, toUpper, toLower, ltrim, rtrim, chop（格式化）
//   - split, join, implode, explode（转换）
//   - startsWith, endsWith（检查）
//   - charAt, repeat, reverse（其他）
//   - strcmp, strcasecmp, strncmp, strncasecmp（比较）
//   - stripos, strrpos, strripos, strstr, stristr, strchr（查找）
//   - sprintf, printf, vsprintf, vprintf（格式化输出）
//   - ord, chr, nl2br, bin2hex（其他）
func StringNames() []string {
	return []string{
		"strlen", "substr", "strpos", "str_replace",
		"trim", "toUpper", "toLower", "ltrim", "rtrim", "chop",
		"split", "join", "implode", "explode",
		"startsWith", "endsWith",
		"charAt", "repeat", "reverse",
		"strcmp", "strcasecmp", "strncmp", "strncasecmp",
		"stripos", "strrpos", "strripos", "strstr", "stristr", "strchr",
		"sprintf", "printf", "vsprintf", "vprintf", "number_format",
		"ord", "chr", "nl2br", "bin2hex", "hex2bin",
		// Phase 11.1 字符串增强
		"substr_compare", "substr_count", "str_repeat", "str_pad",
		"str_split", "strrev", "htmlspecialchars", "htmlspecialchars_decode",
		"strip_tags", "wordwrap", "strtolower", "strtoupper",
		"chunk_split",
		// 转义编码函数
		"addslashes", "stripslashes", "addcslashes",
		// Email编码函数 (quoted-printable)
		"quoted_printable_encode", "quoted_printable_decode",
		// HTML实体编码函数
		"htmlentities", "html_entity_decode", "get_html_translation_table",
	}
}

// ============================================================================
// 核心函数
// ============================================================================

// builtinStrlen 返回字符串的字符数（不是字节数）。
//
// 使用 UTF-8 解码，正确计算 Unicode 字符数量。
// 对于 ASCII 字符串，字符数等于字节数；对于包含多字节字符（如中文）的字符串，
// 字符数小于字节数。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目标字符串
//
// 返回值：
//   - int: 字符数量
//   - error: 参数错误
//
// 使用示例：
//
//	print strlen("Hello")          // 输出: 5（5个ASCII字符）
//	print strlen("你好世界")       // 输出: 4（4个Unicode字符，但12个字节）
//	print strlen("Héllo")          // 输出: 5（é 是2字节字符）
//
// 注意：如果需要字节数，使用 # 运算符或 len() 函数：
//
//	print # "你好"                 // 输出字节数
func builtinStrlen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strlen() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strlen() argument must be a string, got %s", args[0].Type())
	}
	return engine.NewInt(int64(utf8.RuneCountInString(args[0].String()))), nil
}

// substr(s, start, length) 截取子串
// start: 起始位置（支持负数，-1 表示最后一个字符）
// length: 截取长度（可选，默认到末尾）
func builtinSubstr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("substr() expects 2-3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("substr() argument 1 must be a string, got %s", args[0].Type())
	}

	s := args[0].String()
	runes := []rune(s)
	runeLen := len(runes)

	start := int(args[1].Int())

	// 负数索引：从末尾计算
	if start < 0 {
		start = runeLen + start
	}
	if start < 0 {
		start = 0
	}
	if start >= runeLen {
		return engine.NewString(""), nil
	}

	// 计算长度
	length := runeLen - start
	if len(args) == 3 {
		length = max(int(args[2].Int()), 0)
	}

	end := min(start+length, runeLen)

	return engine.NewString(string(runes[start:end])), nil
}

// strpos(s, needle) 查找 needle 在 s 中的位置，未找到返回 -1
func builtinStrpos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("strpos() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strpos() argument 1 must be a string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strpos() argument 2 must be a string, got %s", args[1].Type())
	}

	s := args[0].String()
	needle := args[1].String()

	idx := strings.Index(s, needle)
	if idx < 0 {
		return engine.NewInt(-1), nil
	}

	// 转换为字符位置（而非字节位置）
	charPos := utf8.RuneCountInString(s[:idx])
	return engine.NewInt(int64(charPos)), nil
}

// str_replace(s, search, replace) 替换所有匹配的子串
func builtinStrReplace(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("str_replace() expects 3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("str_replace() argument 1 must be a string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("str_replace() argument 2 must be a string, got %s", args[1].Type())
	}

	s := args[0].String()
	search := args[1].String()
	replace := args[2].String()

	if search == "" {
		return engine.NewString(s), nil
	}

	result := strings.ReplaceAll(s, search, replace)
	return engine.NewString(result), nil
}

// ============================================================================
// 格式化函数
// ============================================================================

// trim(s) 去除首尾空白字符
func builtinTrim(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trim() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("trim() argument must be a string, got %s", args[0].Type())
	}
	return engine.NewString(strings.TrimSpace(args[0].String())), nil
}

// toUpper(s) 转换为大写
func builtinToUpper(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toUpper() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("toUpper() argument must be a string, got %s", args[0].Type())
	}
	return engine.NewString(strings.ToUpper(args[0].String())), nil
}

// toLower(s) 转换为小写
func builtinToLower(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toLower() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("toLower() argument must be a string, got %s", args[0].Type())
	}
	return engine.NewString(strings.ToLower(args[0].String())), nil
}

// ============================================================================
// 拼接/分割函数
// ============================================================================

// builtinSplit 使用分隔符将字符串分割为数组。
//
// 按照分隔符分割字符串，返回字符串数组。如果字符串不包含分隔符，
// 返回包含原字符串的单元素数组。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要分割的字符串
//   - args[1]: 分隔符字符串
//
// 返回值：
//   - array: 分割后的字符串数组
//   - error: 参数错误
//
// 使用示例：
//
//	$csv = "apple,banana,cherry"
//	$fruits = split($csv, ",")
//	print $fruits               // 输出: ["apple", "banana", "cherry"]
//
//	$path = "/home/user/docs"
//	$parts = split($path, "/")
//	print $parts                // 输出: ["", "home", "user", "docs"]
//
//	$text = "Hello World"
//	$words = split($text, " ")  // ["Hello", "World"]
//
// 注意：连续的分隔符会产生空字符串元素
//
//	split("a,,b", ",")  // ["a", "", "b"]
func builtinSplit(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("split() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("split() argument 1 must be a string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("split() argument 2 must be a string, got %s", args[1].Type())
	}

	s := args[0].String()
	delim := args[1].String()

	parts := strings.Split(s, delim)
	result := make([]engine.Value, len(parts))
	for i, p := range parts {
		result[i] = engine.NewString(p)
	}
	return engine.NewArray(result), nil
}

// builtinJoin 将数组元素连接为字符串。
//
// 将数组中的每个元素转换为字符串，然后用分隔符连接。
// 元素通过 String() 方法转换为字符串。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要连接的数组
//   - args[1]: 分隔符字符串
//
// 返回值：
//   - string: 连接后的字符串
//   - error: 参数错误
//
// 使用示例：
//
//	$words = ["Hello", "World"]
//	$sentence = join($words, " ")
//	print $sentence             // 输出: "Hello World"
//
//	$nums = [1, 2, 3, 4, 5]
//	$str = join($nums, "-")     // "1-2-3-4-5"
//
//	$paths = ["home", "user", "docs"]
//	$path = join($paths, "/")   // "home/user/docs"
//
// 注意：这是 split 的逆操作
//
//	$arr = split($str, $delim)
//	$str2 = join($arr, $delim)  // $str2 == $str
func builtinJoin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("join() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("join() argument 1 must be an array, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("join() argument 2 must be a string, got %s", args[1].Type())
	}

	arr := args[0].Array()
	delim := args[1].String()

	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = v.String()
	}
	return engine.NewString(strings.Join(parts, delim)), nil
}

// ============================================================================
// 前缀/后缀检查
// ============================================================================

// startsWith(s, prefix) 检查字符串是否以 prefix 开头
func builtinStartsWith(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("startsWith() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("startsWith() argument 1 must be a string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("startsWith() argument 2 must be a string, got %s", args[1].Type())
	}
	return engine.NewBool(strings.HasPrefix(args[0].String(), args[1].String())), nil
}

// endsWith(s, suffix) 检查字符串是否以 suffix 结尾
func builtinEndsWith(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("endsWith() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("endsWith() argument 1 must be a string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("endsWith() argument 2 must be a string, got %s", args[1].Type())
	}
	return engine.NewBool(strings.HasSuffix(args[0].String(), args[1].String())), nil
}

// ============================================================================
// 字符操作
// ============================================================================

// charAt(s, index) 获取指定位置的字符
func builtinCharAt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("charAt() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("charAt() argument 1 must be a string, got %s", args[0].Type())
	}

	s := args[0].String()
	runes := []rune(s)
	idx := int(args[1].Int())

	if idx < 0 || idx >= len(runes) {
		return engine.NewString(""), nil
	}
	return engine.NewString(string(runes[idx])), nil
}

// repeat(s, count) 重复字符串 N 次
func builtinRepeat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("repeat() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("repeat() argument 1 must be a string, got %s", args[0].Type())
	}

	count := int(args[1].Int())
	if count < 0 {
		return nil, fmt.Errorf("repeat() count must be non-negative, got %d", count)
	}
	if count == 0 {
		return engine.NewString(""), nil
	}
	// 限制最大重复次数防止内存爆炸
	const maxRepeat = 100000
	if count > maxRepeat {
		return nil, fmt.Errorf("repeat() count exceeds maximum (%d)", maxRepeat)
	}
	return engine.NewString(strings.Repeat(args[0].String(), count)), nil
}

// reverse(s) 反转字符串
func builtinReverse(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("reverse() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("reverse() argument must be a string, got %s", args[0].Type())
	}

	runes := []rune(args[0].String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return engine.NewString(string(runes)), nil
}

// ============================================================================
// Phase 7.3 新增函数
// ============================================================================

// ============================================================================
// trim 系列
// ============================================================================

// builtinLtrim 去除字符串左侧空白字符
func builtinLtrim(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("ltrim() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ltrim() argument 1 must be a string, got %s", args[0].Type())
	}

	str := args[0].String()
	chars := " \t\n\r\x00\x0B" // 默认空白字符

	if len(args) == 2 && args[1].Type() == engine.TypeString {
		chars = args[1].String()
	}

	return engine.NewString(strings.TrimLeft(str, chars)), nil
}

// builtinRtrim 去除字符串右侧空白字符
func builtinRtrim(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("rtrim() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("rtrim() argument 1 must be a string, got %s", args[0].Type())
	}

	str := args[0].String()
	chars := " \t\n\r\x00\x0B" // 默认空白字符

	if len(args) == 2 && args[1].Type() == engine.TypeString {
		chars = args[1].String()
	}

	return engine.NewString(strings.TrimRight(str, chars)), nil
}

// ============================================================================
// 字符串比较
// ============================================================================

// builtinStrcmp 二进制安全字符串比较
func builtinStrcmp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("strcmp() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strcmp() arguments must be strings")
	}

	result := strings.Compare(args[0].String(), args[1].String())
	return engine.NewInt(int64(result)), nil
}

// builtinStrcasecmp 不区分大小写字符串比较
func builtinStrcasecmp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("strcasecmp() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strcasecmp() arguments must be strings")
	}

	s1 := strings.ToLower(args[0].String())
	s2 := strings.ToLower(args[1].String())
	result := strings.Compare(s1, s2)
	return engine.NewInt(int64(result)), nil
}

// builtinStrncmp 比较字符串前 n 个字符
func builtinStrncmp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("strncmp() expects 3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strncmp() arguments 1 and 2 must be strings")
	}
	if args[2].Type() != engine.TypeInt {
		return nil, fmt.Errorf("strncmp() argument 3 must be an integer")
	}

	s1 := args[0].String()
	s2 := args[1].String()
	n := max(int(args[2].Int()), 0)

	n = min(n, len(s1))
	n = min(n, len(s2))

	result := strings.Compare(s1[:n], s2[:n])
	return engine.NewInt(int64(result)), nil
}

// builtinStrncasecmp 不区分大小写比较前 n 个字符
func builtinStrncasecmp(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("strncasecmp() expects 3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strncasecmp() arguments 1 and 2 must be strings")
	}
	if args[2].Type() != engine.TypeInt {
		return nil, fmt.Errorf("strncasecmp() argument 3 must be an integer")
	}

	s1 := strings.ToLower(args[0].String())
	s2 := strings.ToLower(args[1].String())
	n := max(int(args[2].Int()), 0)

	n = min(n, len(s1))
	n = min(n, len(s2))

	result := strings.Compare(s1[:n], s2[:n])
	return engine.NewInt(int64(result)), nil
}

// ============================================================================
// 查找函数
// ============================================================================

// builtinStripos 不区分大小写查找子串位置
func builtinStripos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("stripos() expects 2-3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("stripos() arguments 1 and 2 must be strings")
	}

	haystack := strings.ToLower(args[0].String())
	needle := strings.ToLower(args[1].String())
	offset := 0

	if len(args) == 3 {
		if args[2].Type() != engine.TypeInt {
			return nil, fmt.Errorf("stripos() argument 3 must be an integer")
		}
		offset = int(args[2].Int())
		if offset < 0 {
			offset = 0
		}
	}

	if offset > len(haystack) {
		return engine.NewBool(false), nil
	}

	pos := strings.Index(haystack[offset:], needle)
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewInt(int64(pos + offset)), nil
}

// builtinStrrpos 查找最后一次出现的位置
func builtinStrrpos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("strrpos() expects 2-3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strrpos() arguments 1 and 2 must be strings")
	}

	haystack := args[0].String()
	needle := args[1].String()
	offset := 0

	if len(args) == 3 {
		if args[2].Type() != engine.TypeInt {
			return nil, fmt.Errorf("strrpos() argument 3 must be an integer")
		}
		offset = int(args[2].Int())
	}

	if offset < 0 {
		offset = 0
	}
	if offset > len(haystack) {
		return engine.NewBool(false), nil
	}

	pos := strings.LastIndex(haystack[offset:], needle)
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewInt(int64(pos + offset)), nil
}

// builtinStrripos 不区分大小写查找最后一次出现的位置
func builtinStrripos(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("strripos() expects 2-3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strripos() arguments 1 and 2 must be strings")
	}

	haystack := strings.ToLower(args[0].String())
	needle := strings.ToLower(args[1].String())
	offset := 0

	if len(args) == 3 {
		if args[2].Type() != engine.TypeInt {
			return nil, fmt.Errorf("strripos() argument 3 must be an integer")
		}
		offset = int(args[2].Int())
	}

	if offset < 0 {
		offset = 0
	}
	if offset > len(haystack) {
		return engine.NewBool(false), nil
	}

	pos := strings.LastIndex(haystack[offset:], needle)
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewInt(int64(pos + offset)), nil
}

// builtinStrstr 查找子串并返回首次出现到结尾的字符串
func builtinStrstr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("strstr() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strstr() arguments must be strings")
	}

	haystack := args[0].String()
	needle := args[1].String()

	pos := strings.Index(haystack, needle)
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewString(haystack[pos:]), nil
}

// builtinStristr 不区分大小写查找子串
func builtinStristr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("stristr() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("stristr() arguments must be strings")
	}

	haystack := strings.ToLower(args[0].String())
	needle := strings.ToLower(args[1].String())
	original := args[0].String()

	pos := strings.Index(haystack, needle)
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewString(original[pos:]), nil
}

// builtinStrchr 查找字符首次出现（strstr 的别名行为）
func builtinStrchr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("strchr() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("strchr() arguments must be strings")
	}

	haystack := args[0].String()
	needle := args[1].String()

	if len(needle) == 0 {
		return engine.NewBool(false), nil
	}

	// 只取第一个字符查找
	pos := strings.IndexByte(haystack, needle[0])
	if pos == -1 {
		return engine.NewBool(false), nil
	}

	return engine.NewString(haystack[pos:]), nil
}

// ============================================================================
// 格式化函数
// ============================================================================

// builtinSprintf 格式化字符串（简化版）
//
// 支持基本格式化占位符：
//   - %s: 字符串（调用值的 String() 方法）
//   - %d: 十进制整数
//   - %f: 浮点数（默认 6 位小数）
//   - %%: 字面量百分号
//
// 注意：此为简化实现，不支持：
//   - 精度控制（如 %.2f）
//   - 宽度控制（如 %10s）
//   - 标志位（如 %+d, %05d）
//   - 其他占位符（%b, %x, %o 等）
//
// 参数：
//   - format: 格式字符串
//   - ...args: 要格式化的值
//
// 返回值：
//   - string: 格式化后的字符串
//
// 使用示例：
//
//	sprintf("Hello %s", "World")          // "Hello World"
//	sprintf("Number: %d", 42)             // "Number: 42"
//	sprintf("Float: %f", 3.14)            // "Float: 3.140000"
//	sprintf("%%")                          // "%"
//	sprintf("Name: %s, Age: %d", "John", 30)
//	// "Name: John, Age: 30"
//
// 相关函数：printf(输出到 stdout)、vsprintf(数组参数)
func builtinSprintf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("sprintf() expects at least 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("sprintf() argument 1 must be a string")
	}

	format := args[0].String()

	// 简单格式化：只处理 %s, %d, %f, %%
	result := format
	argIdx := 1

	for i := 0; i < len(format); i++ {
		if format[i] == '%' && i+1 < len(format) {
			switch format[i+1] {
			case 's':
				if argIdx < len(args) {
					result = strings.Replace(result, "%s", args[argIdx].String(), 1)
					argIdx++
				}
			case 'd':
				if argIdx < len(args) {
					result = strings.Replace(result, "%d", fmt.Sprintf("%d", args[argIdx].Int()), 1)
					argIdx++
				}
			case 'f':
				if argIdx < len(args) {
					result = strings.Replace(result, "%f", fmt.Sprintf("%f", args[argIdx].Float()), 1)
					argIdx++
				}
			case '%':
				result = strings.Replace(result, "%%", "%", 1)
			}
		}
	}

	return engine.NewString(result), nil
}

// builtinPrintf 格式化并输出字符串
func builtinPrintf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	result, err := builtinSprintf(ctx, args)
	if err != nil {
		return nil, err
	}

	// 输出到 stdout
	fmt.Print(result.String())
	return result, nil
}

// builtinVsprintf 用数组参数格式化字符串
func builtinVsprintf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("vsprintf() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("vsprintf() argument 1 must be a string")
	}
	if args[1].Type() != engine.TypeArray {
		return nil, fmt.Errorf("vsprintf() argument 2 must be an array")
	}

	params := args[1].Array()

	// 构建新参数列表
	newArgs := []engine.Value{args[0]}
	newArgs = append(newArgs, params...)

	return builtinSprintf(ctx, newArgs)
}

// builtinVprintf 用数组参数格式化并输出
func builtinVprintf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	result, err := builtinVsprintf(ctx, args)
	if err != nil {
		return nil, err
	}

	fmt.Print(result.String())
	return result, nil
}

// ============================================================================
// 其他函数
// ============================================================================

// builtinOrd 获取字符串首字符的 ASCII/Unicode 码点值
//
// 返回字符串第一个字节的 ASCII 值（0-255）。
// 对于多字节 UTF-8 字符，返回首字节值而非完整码点。
// 空字符串返回 0。
//
// 参数：
//   - str: 输入字符串
//
// 返回值：
//   - int: 首字符的字节值（0-255），空字符串返回 0
//
// 使用示例：
//
//	ord("A")    // 65
//	ord("a")    // 97
//	ord("0")    // 48
//	ord("")     // 0
//	ord("ABC")  // 65 (只取第一个字符)
//
// 注意：Unicode 字符可能由多字节组成，ord 只返回首字节。
//
//	如需完整 Unicode 码点，建议配合 utf8 包使用。
//
// 相关函数：chr() —— 逆向转换
func builtinOrd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ord() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ord() argument must be a string")
	}

	str := args[0].String()
	if len(str) == 0 {
		return engine.NewInt(0), nil
	}

	return engine.NewInt(int64(str[0])), nil
}

// builtinChr 将 ASCII 值转换为单字符字符串
//
// 将 0-255 的整数转换为对应的 ASCII 字符。
// 超出范围（< 0 或 > 255）返回空字符串。
//
// 参数：
//   - ascii: ASCII 码值（0-255）
//
// 返回值：
//   - string: 对应的 ASCII 字符，越界返回空字符串
//
// 使用示例：
//
//	chr(65)     // "A"
//	chr(97)     // "a"
//	chr(48)     // "0"
//	chr(255)    // 字符 0xFF
//	chr(256)    // "" (越界)
//	chr(-1)     // "" (越界)
//
// 注意：chr 仅支持单字节字符（0-255）。
//
//	如需输出 Unicode 字符，建议使用转义序列或直接输入。
//
// 相关函数：ord() —— 逆向转换
func builtinChr(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("chr() expects 1 argument, got %d", len(args))
	}

	ascii := int(args[0].Int())
	if ascii < 0 || ascii > 255 {
		return engine.NewString(""), nil
	}

	return engine.NewString(string(byte(ascii))), nil
}

// builtinNl2br 将换行符转换为 <br> 标签（HTML/XHTML 输出）
//
// 将字符串中的换行符替换为 <br> 或 <br /> 标签，方便在 HTML 中显示。
// 支持三种换行符风格：\n (Unix)、\r (Mac)、\r\n (Windows)。
//
// 参数：
//   - str: 输入字符串
//   - isXHTML: 可选，是否使用 XHTML 风格（<br />），默认 true
//     false 时使用 HTML 风格（<br>）
//
// 返回值：
//   - string: 替换后的字符串
//
// 换行符处理：
//   - \r\n (Windows): 转为 "<br />\r\n" 或 "<br>\r\n"
//   - \n (Unix): 转为 "<br />\n" 或 "<br>\n"
//   - \r (Mac): 转为 "<br />\r" 或 "<br>\r"
//
// 使用示例：
//
//	nl2br("line1\nline2")           // "line1<br />\nline2"
//	nl2br("line1\nline2", false)   // "line1<br>\nline2"
//	nl2br("line1\r\nline2")        // "line1<br />\r\nline2"
//
// 注意：此函数保留原始换行符，只是在前面插入 <br> 标签。
//
//	如需移除换行符，需配合其他字符串函数。
func builtinNl2br(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("nl2br() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("nl2br() argument must be a string")
	}

	str := args[0].String()
	isXHTML := true

	if len(args) == 2 {
		isXHTML = engine.IsTruthy(args[1])
	}

	br := "<br />"
	if !isXHTML {
		br = "<br>"
	}

	// 替换各种换行符
	// 策略：先将 \r\n 替换为占位符，处理完其他后再恢复
	const placeholder = "\x00NEWLINE\x00"
	result := strings.ReplaceAll(str, "\r\n", placeholder)
	result = strings.ReplaceAll(result, "\n", br+"\n")
	result = strings.ReplaceAll(result, "\r", br+"\r")
	result = strings.ReplaceAll(result, placeholder, br+"\r\n")

	return engine.NewString(result), nil
}

// builtinBin2hex 将二进制数据转换为十六进制
func builtinBin2hex(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bin2hex() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("bin2hex() argument must be a string")
	}

	str := args[0].String()
	result := fmt.Sprintf("%x", str)

	return engine.NewString(result), nil
}

// builtinHex2bin 将十六进制字符串转换回二进制数据
// hex2bin($hex) → string
//
// 参数：
//   - args[0]: 十六进制字符串（必须包含偶数个字符，0-9, a-f, A-F）
//
// 返回值：
//   - 二进制字符串（原始字节）
//   - null: 输入包含无效字符或长度为奇数
//
// 示例：
//
//	hex2bin("48656c6c6f") → "Hello"
//	hex2bin("c0a80101") → 4字节的二进制数据
func builtinHex2bin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("hex2bin() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("hex2bin() argument must be a string")
	}

	hexStr := strings.ToLower(args[0].String())

	// 去除可能的 0x 或 0X 前缀
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")

	// 检查长度是否为偶数
	if len(hexStr)%2 != 0 {
		return engine.NewNull(), nil
	}

	// 解析十六进制
	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		// 获取两个十六进制字符
		high := hexCharToByte(hexStr[i])
		low := hexCharToByte(hexStr[i+1])

		if high < 0 || low < 0 {
			return engine.NewNull(), nil
		}

		b := byte((high << 4) | low)
		result.WriteByte(b)
	}

	return engine.NewString(result.String()), nil
}

// hexCharToByte 将十六进制字符转换为字节值
func hexCharToByte(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c - 'a' + 10)
	case c >= 'A' && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}

// 十六进制字符查找表（大写）
var hexUpper = []byte("0123456789ABCDEF")

func builtinSubstrCompare(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("substr_compare() expects 2 to 4 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("substr_compare() arguments must be strings")
	}

	mainStr := args[0].String()
	compareStr := args[1].String()

	offset := 0
	if len(args) >= 3 {
		offset = int(args[2].Int())
	}

	length := len(mainStr)
	if len(args) >= 4 {
		length = int(args[3].Int())
	}

	if offset < 0 {
		offset = max(len(mainStr)+offset, 0)
	}

	if offset >= len(mainStr) {
		return engine.NewInt(0), nil
	}

	if length < 0 {
		length = len(mainStr) - offset + length
	}

	end := min(offset+length, len(mainStr))

	mainSub := mainStr[offset:end]
	if len(compareStr) > len(mainSub) {
		compareStr = compareStr[:len(mainSub)]
	}

	return engine.NewInt(int64(strings.Compare(mainSub, compareStr))), nil
}

func builtinSubstrCount(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("substr_count() expects 2 to 4 arguments, got %d", len(args))
	}

	haystack := args[0].String()
	needle := args[1].String()

	if needle == "" {
		return engine.NewInt(0), nil
	}

	offset := 0
	if len(args) >= 3 {
		offset = int(args[2].Int())
	}

	length := len(haystack)
	if len(args) >= 4 {
		length = int(args[3].Int())
	}

	if offset < 0 {
		offset = len(haystack) + offset
	}

	if offset >= len(haystack) {
		return engine.NewInt(0), nil
	}

	if length > 0 {
		haystack = haystack[offset : offset+length]
	} else {
		haystack = haystack[offset:]
	}

	count := 0
	idx := 0
	for {
		i := strings.Index(haystack[idx:], needle)
		if i == -1 {
			break
		}
		count++
		idx += i + len(needle)
	}

	return engine.NewInt(int64(count)), nil
}

func builtinStrRepeat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("str_repeat() expects 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("str_repeat() argument 1 must be a string")
	}

	str := args[0].String()
	multiplier := max(int(args[1].Int()), 0)

	if len(str)*multiplier > 1000000 {
		return nil, fmt.Errorf("str_repeat(): Result too large")
	}

	result := strings.Repeat(str, multiplier)
	return engine.NewString(result), nil
}

func builtinStrPad(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("str_pad() expects 3 to 4 arguments, got %d", len(args))
	}

	input := args[0].String()
	padLength := int(args[1].Int())
	padString := " "
	if len(args) >= 4 {
		padString = args[3].String()
	}

	if padLength <= len(input) {
		return engine.NewString(input), nil
	}

	padType := 0
	if len(args) >= 4 {
		padStr := args[3].String()
		switch padStr {
		case "STR_PAD_RIGHT":
			padType = 0
		case "STR_PAD_LEFT":
			padType = 1
		case "STR_PAD_BOTH":
			padType = 2
		}
	}

	result := input
	remain := padLength - len(input)

	switch padType {
	case 1:
		result = strings.Repeat(padString, (remain+len(padString)-1)/len(padString)) + result
	case 2:
		{
			left := remain / 2
			right := remain - left
			result = strings.Repeat(padString, (left+len(padString)-1)/len(padString)) + result +
				strings.Repeat(padString, (right+len(padString)-1)/len(padString))
		}
	default:
		result = result + strings.Repeat(padString, (remain+len(padString)-1)/len(padString))
	}

	return engine.NewString(result), nil
}

func builtinStrSplit(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("str_split() expects 1 or 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("str_split() argument 1 must be a string")
	}

	str := args[0].String()
	splitLen := 1
	if len(args) >= 2 {
		splitLen = int(args[1].Int())
	}

	if splitLen <= 0 {
		splitLen = 1
	}

	var result []engine.Value
	for i := 0; i < len(str); i += splitLen {
		end := min(i+splitLen, len(str))
		result = append(result, engine.NewString(str[i:end]))
	}

	return engine.NewArray(result), nil
}

func builtinStrrev(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strrev() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strrev() argument must be a string")
	}

	str := args[0].String()
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return engine.NewString(string(runes)), nil
}

func builtinHtmlspecialchars(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("htmlspecialchars() expects 1 to 3 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("htmlspecialchars() argument 1 must be a string")
	}

	str := args[0].String()

	flags := 0
	if len(args) >= 2 {
		flags = int(args[1].Int())
	}

	result := html.EscapeString(str)

	if flags&1 == 0 {
		result = strings.ReplaceAll(result, "&quot;", "\"")
	}

	if flags&2 == 0 {
		result = strings.ReplaceAll(result, "&#039;", "'")
	}

	if flags&4 == 0 {
		result = strings.ReplaceAll(result, "&lt;", "<")
		result = strings.ReplaceAll(result, "&gt;", ">")
	}

	if flags&8 == 0 {
		result = strings.ReplaceAll(result, "&amp;", "&")
	}

	return engine.NewString(result), nil
}

func builtinHtmlspecialcharsDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("htmlspecialchars_decode() expects 1 to 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("htmlspecialchars_decode() argument 1 must be a string")
	}

	str := args[0].String()
	result := html.UnescapeString(str)

	return engine.NewString(result), nil
}

func builtinStripTags(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("strip_tags() expects 1 or 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strip_tags() argument 1 must be a string")
	}

	str := args[0].String()
	allowableTags := ""
	if len(args) >= 2 {
		allowableTags = args[1].String()
	}

	re := regexp.MustCompile(`(?s)<[^>]*>`)
	result := re.ReplaceAllString(str, "")

	if allowableTags != "" {
		for tag := range strings.SplitSeq(allowableTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			result = strings.ReplaceAll(result, "&lt;"+tag+"&gt;", "<"+tag+">")
		}
	}

	wsRe := regexp.MustCompile(`\s+`)
	result = wsRe.ReplaceAllString(result, " ")

	return engine.NewString(strings.TrimSpace(result)), nil
}

func builtinWordwrap(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 4 {
		return nil, fmt.Errorf("wordwrap() expects 2 to 4 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("wordwrap() argument 1 must be a string")
	}

	str := args[0].String()
	width := int(args[1].Int())
	breakChar := "\n"
	cut := false

	if len(args) >= 3 {
		breakChar = args[2].String()
	}
	if len(args) >= 4 {
		cut = args[3].Bool()
	}

	words := strings.Fields(str)
	if len(words) == 0 {
		return engine.NewString(""), nil
	}

	result := words[0]
	lineLen := len(words[0])

	for i := 1; i < len(words); i++ {
		wordLen := len(words[i])
		if lineLen+1+wordLen > width && lineLen > 0 {
			result += breakChar + words[i]
			lineLen = wordLen
		} else {
			result += " " + words[i]
			lineLen += 1 + wordLen
		}
	}

	if cut && lineLen > width {
		var wrapped strings.Builder
		for i := 0; i < len(result); i += width {
			end := min(i+width, len(result))
			wrapped.WriteString(result[i:end])
			if end < len(result) {
				wrapped.WriteString(breakChar)
			}
		}
		return engine.NewString(wrapped.String()), nil
	}

	return engine.NewString(result), nil
}

func builtinStrtolower(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strtolower() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strtolower() argument must be a string")
	}

	return engine.NewString(strings.ToLower(args[0].String())), nil
}

func builtinStrtoupper(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("strtoupper() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strtoupper() argument must be a string")
	}

	return engine.NewString(strings.ToUpper(args[0].String())), nil
}

func builtinChunkSplit(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("chunk_split() expects 2 or 3 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("chunk_split() argument 1 must be a string")
	}

	str := args[0].String()
	chunkLen := int(args[1].Int())
	end := "\r\n"
	if len(args) >= 3 {
		end = args[2].String()
	}

	if chunkLen <= 0 {
		return engine.NewString(str), nil
	}

	var result strings.Builder
	for i := 0; i < len(str); i += chunkLen {
		endIdx := min(i+chunkLen, len(str))
		result.WriteString(str[i:endIdx])
		if endIdx < len(str) {
			result.WriteString(end)
		}
	}

	return engine.NewString(result.String()), nil
}

// builtinAddslashes 使用反斜杠引用字符串
// addslashes($str) → string
//
// 在字符 ' 、" 、\ 以及 NUL（\0） 前添加反斜杠
// 用于转义 SQL 查询中的字符串、JSON 字符串等场景
//
// 参数：
//   - args[0]: 要转义的字符串
//
// 返回值：
//   - 转义后的字符串
//
// 示例：
//
//	addslashes("It's a test") → "It\\'s a test"
//	addslashes('Say "Hello"') → 'Say \"Hello\"'
func builtinAddslashes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("addslashes() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("addslashes() argument must be a string")
	}

	str := args[0].String()
	var result strings.Builder

	for _, r := range str {
		switch r {
		case '\'', '"', '\\', '\x00':
			result.WriteByte('\\')
		}
		result.WriteRune(r)
	}

	return engine.NewString(result.String()), nil
}

// builtinStripslashes 反引用一个使用 addslashes() 转义的字符串
// stripslashes($str) → string
//
// 移除由 addslashes() 添加的反斜杠转义，还原原始字符串
//
// 参数：
//   - args[0]: 要还原的字符串
//
// 返回值：
//   - 还原后的字符串
//
// 示例：
//
//	stripslashes("It\\'s a test") → "It's a test"
func builtinStripslashes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("stripslashes() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("stripslashes() argument must be a string")
	}

	str := args[0].String()
	var result strings.Builder

	for i := 0; i < len(str); i++ {
		if str[i] == '\\' && i < len(str)-1 {
			// 跳过反斜杠，保留后面的字符
			i++
			result.WriteByte(str[i])
		} else {
			result.WriteByte(str[i])
		}
	}

	return engine.NewString(result.String()), nil
}

// builtinAddcslashes 以C语言风格使用反斜杠转义字符串
// addcslashes($str, $charlist) → string
//
// 只对 charlist 中列出的字符进行转义，使用C风格转义序列
// 支持的C风格转义：\n(换行), \t(制表符), \r(回车), \0(NUL), \\等
// charlist 中的字符可以是单个字符或范围（如 "a..z"）
//
// 参数：
//   - args[0]: 要转义的字符串
//   - args[1]: 需要转义的字符列表
//
// 返回值：
//   - 转义后的字符串
//
// 示例：
//
//	addcslashes("Hello\nWorld", "\n") → "Hello\\nWorld"
//	addcslashes("test", "a..z") → "\\t\\e\\s\\t"
func builtinAddcslashes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("addcslashes() expects 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("addcslashes() argument 1 must be a string")
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("addcslashes() argument 2 must be a string")
	}

	str := args[0].String()
	charlist := args[1].String()

	// 解析 charlist，支持范围（如 "a..z"）
	charsToEscape := make(map[rune]bool)
	for i := 0; i < len(charlist); i++ {
		if i+2 < len(charlist) && charlist[i+1] == '.' && charlist[i+2] == '.' {
			// 范围：start..end
			start := rune(charlist[i])
			end := rune(charlist[i+3])
			if start > end {
				start, end = end, start
			}
			for c := start; c <= end; c++ {
				charsToEscape[c] = true
			}
			i += 3
		} else {
			charsToEscape[rune(charlist[i])] = true
		}
	}

	// C风格转义映射
	escapeMap := map[rune]string{
		'\n':   "\\n",
		'\t':   "\\t",
		'\r':   "\\r",
		'\x00': "\\0",
		'\\':   "\\\\",
		'"':    "\\\"",
		'\'':   "\\'",
	}

	var result strings.Builder
	for _, r := range str {
		if charsToEscape[r] {
			if esc, ok := escapeMap[r]; ok {
				result.WriteString(esc)
			} else {
				result.WriteByte('\\')
				result.WriteRune(r)
			}
		} else {
			result.WriteRune(r)
		}
	}

	return engine.NewString(result.String()), nil
}

// ============================================================================
// Email encoding functions (quoted-printable)
// ============================================================================

// builtinQuotedPrintableEncode  quoted-printable 编码
// quoted_printable_encode($str) → string
//
// 将字符串编码为 quoted-printable 格式，适用于 email 传输
// 将不可打印的字符和特殊字符转换为 =XX 形式（XX 为十六进制）
//
// 参数：
//   - args[0]: 要编码的字符串
//
// 返回值：
//   - 编码后的字符串
func builtinQuotedPrintableEncode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("quoted_printable_encode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("quoted_printable_encode() argument must be a string")
	}

	input := args[0].String()
	var result strings.Builder

	for i := 0; i < len(input); i++ {
		char := input[i]
		// 检查是否需要编码：不可打印字符、=、?、_、空格、制表符等
		if char == '=' || char == '?' || char == '_' ||
			(char >= 0 && char <= 32) || char >= 127 {
			// 需要编码为 =XX 形式
			result.WriteByte('=')
			result.WriteByte(hexUpper[char>>4])
			result.WriteByte(hexUpper[char&0x0F])
		} else {
			// 直接写入字符
			result.WriteByte(char)
		}
	}

	return engine.NewString(result.String()), nil
}

// builtinQuotedPrintableDecode  quoted-printable 解码
// quoted_printable_decode($str) → string
//
// 将 quoted-printable 编码的字符串解码回原始字符串
// 将 =XX 形式转换回对应的字节
// 无效的编码形式（如 =1, =ZZ）保持原样
//
// 参数：
//   - args[0]: 要解码的字符串
//
// 返回值：
//   - 解码后的字符串
func builtinQuotedPrintableDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("quoted_printable_decode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("quoted_printable_decode() argument must be a string")
	}

	input := args[0].String()
	var result strings.Builder
	i := 0

	for i < len(input) {
		if input[i] == '=' && i+2 < len(input) {
			// 检查是否是有效的 =XX 编码形式
			high := hexCharToByte(input[i+1])
			low := hexCharToByte(input[i+2])

			if high >= 0 && low >= 0 {
				// 有效的 =XX 编码，直接解码
				result.WriteByte(byte((high << 4) | low))
				i += 3
				continue
			}
		}
		// 普通字符或无效的 = 形式直接写入
		result.WriteByte(input[i])
		i++
	}

	return engine.NewString(result.String()), nil
}

// ============================================================================
// HTML entity encoding functions
// ============================================================================

// builtinHtmlEntities  HTML 实体编码
// htmlentities($str) → string
//
// 将特殊字符转换为 HTML 实体
// 例如：< → &lt;, > → &gt;, & → &amp;, " → &quot;, ' → &#039;
//
// 参数：
//   - args[0]: 要编码的字符串
//
// 返回值：
//   - 编码后的字符串
func builtinHtmlEntities(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("htmlentities() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("htmlentities() argument 1 must be a string")
	}

	input := args[0].String()
	// 手动替换以确保正确的顺序（避免双重转义）
	result := strings.ReplaceAll(input, "&", "&amp;") // 必须先处理 &
	result = strings.ReplaceAll(result, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")
	result = strings.ReplaceAll(result, "\"", "&quot;")
	result = strings.ReplaceAll(result, "'", "&#039;")

	return engine.NewString(result), nil
}

// builtinHtmlEntityDecode  HTML 实体解码
// html_entity_decode($str) → string
//
// 将 HTML 实体转换回普通字符
// 例如：&lt; → <, &gt; → >, &amp; → &, &quot; → ", &#039; → '
//
// 参数：
//   - args[0]: 要解码的字符串
//
// 返回值：
//   - 解码后的字符串
func builtinHtmlEntityDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("html_entity_decode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("html_entity_decode() argument must be a string")
	}

	input := args[0].String()
	// 顺序很重要：先处理长的实体，避免错误替换
	result := strings.ReplaceAll(input, "&quot;", "\"")
	result = strings.ReplaceAll(result, "&#039;", "'")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&amp;", "&")

	return engine.NewString(result), nil
}

// builtinGetHtmlTranslationTable 获取 HTML 实体转换表
// get_html_translation_table() → array
//
// 返回 HTML 特殊字符及其对应的实体映射表
// 用于了解 htmlentities 和 html_entity_decode 的转换规则
//
// 参数：无
//
// 返回值：
//   - 包含字符到实体映射的关联数组
func builtinGetHtmlTranslationTable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("get_html_translation_table() expects 0 arguments, got %d", len(args))
	}

	// 返回标准的 HTML 实体映射表
	table := map[string]engine.Value{
		"<":  engine.NewString("&lt;"),
		">":  engine.NewString("&gt;"),
		"&":  engine.NewString("&amp;"),
		"\"": engine.NewString("&quot;"),
		"'":  engine.NewString("&#039;"),
	}

	return engine.NewObject(table), nil
}

func builtinNumberFormat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	numArgs := len(args)
	if numArgs < 1 || numArgs > 4 {
		return nil, fmt.Errorf("number_format() expects 1-4 arguments, got %d", numArgs)
	}

	num := args[0].Float()

	decimals := 0
	decPoint := "."
	thousandsSep := ","

	if numArgs >= 2 {
		decimals = int(args[1].Int())
	}
	if numArgs >= 3 {
		decPoint = args[2].String()
	}
	if numArgs >= 4 {
		thousandsSep = args[3].String()
	}

	format := fmt.Sprintf("%."+fmt.Sprintf("%d", decimals)+"f", num)
	parts := strings.Split(format, ".")

	intPart := parts[0]
	fracPart := ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	result := ""
	n := len(intPart)
	for i, ch := range intPart {
		if i > 0 && (n-i)%3 == 0 {
			result += thousandsSep
		}
		result += string(ch)
	}

	if decimals > 0 && len(fracPart) > 0 {
		result += decPoint + fracPart
	} else if decimals > 0 {
		result += decPoint + strings.Repeat("0", decimals)
	}

	return engine.NewString(result), nil
}
