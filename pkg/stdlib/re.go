package stdlib

import (
	"fmt"
	"regexp"

	"github.com/gnuos/jpl/engine"
)

// =============================================================================
// 正则表达式模块 - 基于 Go RE2 语法，参考 Python re 模块命名
// =============================================================================

// RegisterRe 注册正则表达式函数到引擎
func RegisterRe(e *engine.Engine) {
	e.RegisterFunc("re_match", builtinReMatch)
	e.RegisterFunc("re_search", builtinReSearch)
	e.RegisterFunc("re_findall", builtinReFindall)
	e.RegisterFunc("re_sub", builtinReSub)
	e.RegisterFunc("re_split", builtinReSplit)
	e.RegisterFunc("re_groups", builtinReGroups)

	// 内部函数：用于 match/case 正则模式捕获组提取
	// re_groups_raw(regexValue, subject) → object
	e.RegisterFunc("re_groups_raw", builtinReGroupsRaw)

	// P1
	e.RegisterFunc("re_quote", builtinReQuote)
	e.RegisterFunc("re_fullmatch", builtinReFullmatch)

	// 模块注册 - import "re" 可用
	e.RegisterModule("re", map[string]engine.GoFunction{
		"match":   builtinReMatch,
		"search":  builtinReSearch,
		"findall": builtinReFindall,
		"sub":     builtinReSub,
		"split":   builtinReSplit,
		"groups":  builtinReGroups,
		// P1
		"quote":     builtinReQuote,
		"fullmatch": builtinReFullmatch,
	})
}

// ReNames 返回正则表达式函数名称列表
func ReNames() []string {
	return []string{
		"re_match", "re_search", "re_findall",
		"re_sub", "re_split", "re_groups",
		"re_quote", "re_fullmatch",
	}
}

// builtinReMatch 检查正则是否匹配
// re_match(pattern, string) → bool
//
// 示例：
//
//	re_match("\\d+", "abc123")  // → true
func builtinReMatch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_match() expects 2 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_match() expects string pattern, got %s", args[0].Type())
	}

	subject := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_match() expects string subject, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_match() invalid pattern: %v", err)
	}

	return engine.NewBool(re.MatchString(subject)), nil
}

// builtinReSearch 查找第一个匹配
// re_search(pattern, string) → string | ""
//
// 示例：
//
//	re_search("\\d+", "Room 101, Floor 5")  // → "101"
func builtinReSearch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_search() expects 2 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_search() expects string pattern, got %s", args[0].Type())
	}

	subject := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_search() expects string subject, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_search() invalid pattern: %v", err)
	}

	match := re.FindString(subject)
	return engine.NewString(match), nil
}

// builtinReFindall 查找所有匹配
// re_findall(pattern, string) → [match1, match2, ...]
//
// 示例：
//
//	re_findall("\\d+", "Room 101, Floor 5")  // → ["101", "5"]
func builtinReFindall(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_findall() expects 2 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_findall() expects string pattern, got %s", args[0].Type())
	}

	subject := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_findall() expects string subject, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_findall() invalid pattern: %v", err)
	}

	matches := re.FindAllString(subject, -1)
	if matches == nil {
		matches = []string{}
	}

	result := make([]engine.Value, len(matches))
	for i, m := range matches {
		result[i] = engine.NewString(m)
	}

	return engine.NewArray(result), nil
}

// builtinReSub 替换所有匹配
// re_sub(pattern, replacement, string) → string
//
// 示例：
//
//	re_sub("\\d+", "[NUM]", "Room 101")  // → "Room [NUM]"
func builtinReSub(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("re_sub() expects 3 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_sub() expects string pattern, got %s", args[0].Type())
	}

	replacement := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_sub() expects string replacement, got %s", args[1].Type())
	}

	subject := args[2].String()
	if args[2].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_sub() expects string subject, got %s", args[2].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_sub() invalid pattern: %v", err)
	}

	result := re.ReplaceAllString(subject, replacement)
	return engine.NewString(result), nil
}

// builtinReSplit 按正则分割字符串
// re_split(pattern, string) → [part1, part2, ...]
//
// 示例：
//
//	re_split("\\s*,\\s*", "a, b ,c")  // → ["a", "b", "c"]
func builtinReSplit(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_split() expects 2 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_split() expects string pattern, got %s", args[0].Type())
	}

	subject := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_split() expects string subject, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_split() invalid pattern: %v", err)
	}

	parts := re.Split(subject, -1)
	result := make([]engine.Value, len(parts))
	for i, p := range parts {
		result[i] = engine.NewString(p)
	}

	return engine.NewArray(result), nil
}

// builtinReGroups 返回捕获组
// re_groups(pattern, string) → {0: full_match, 1: group1, ...} | null
//
// 支持命名捕获组 (?P<name>...)
//
// 示例：
//
//	re_groups("(\\d{4})-(\\d{2})", "2024-03")
//	// → {0: "2024-03", 1: "2024", 2: "03"}
//
//	re_groups("(?P<year>\\d{4})-(?P<month>\\d{2})", "2024-03")
//	// → {0: "2024-03", year: "2024", month: "03", 1: "2024", 2: "03"}
func builtinReGroups(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_groups() expects 2 arguments, got %d", len(args))
	}

	pattern := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_groups() expects string pattern, got %s", args[0].Type())
	}

	subject := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("re_groups() expects string subject, got %s", args[1].Type())
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("re_groups() invalid pattern: %v", err)
	}

	matches := re.FindStringSubmatch(subject)
	if matches == nil {
		return engine.NewNull(), nil
	}

	result := make(map[string]engine.Value)

	for i, m := range matches {
		result[fmt.Sprintf("%d", i)] = engine.NewString(m)
	}

	names := re.SubexpNames()
	for i, name := range names {
		if i > 0 && i < len(matches) && name != "" {
			result[name] = engine.NewString(matches[i])
		}
	}

	return engine.NewObject(result), nil
}

// builtinReGroupsRaw 内部函数：从 regexValue 提取捕获组
//
// 用于 match/case 正则模式的 as $var 绑定。
// 参数：args[0] = regexValue, args[1] = subject string
// 返回：捕获组对象 {0: "full", 1: "group1", ...} 或 null
func builtinReGroupsRaw(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_groups_raw() expects 2 arguments, got %d", len(args))
	}

	reVal := args[0]
	subject := args[1].String()

	// 如果是 regexValue，直接使用预编译的正则
	if engine.IsRegex(reVal) {
		reObj := reVal.(interface {
			Regexp() *regexp.Regexp
			Pattern() string
			Flags() string
		})
		re := reObj.Regexp()
		matches := re.FindStringSubmatch(subject)
		if matches == nil {
			return engine.NewNull(), nil
		}

		result := make(map[string]engine.Value)
		for i, m := range matches {
			result[fmt.Sprintf("%d", i)] = engine.NewString(m)
		}
		names := re.SubexpNames()
		for i, name := range names {
			if i > 0 && i < len(matches) && name != "" {
				result[name] = engine.NewString(matches[i])
			}
		}
		return engine.NewObject(result), nil
	}

	// fallback：字符串模式
	pattern := reVal.String()
	re, err := regexp.Compile(pattern)
	if err != nil {
		return engine.NewNull(), nil
	}
	matches := re.FindStringSubmatch(subject)
	if matches == nil {
		return engine.NewNull(), nil
	}

	result := make(map[string]engine.Value)
	for i, m := range matches {
		result[fmt.Sprintf("%d", i)] = engine.NewString(m)
	}
	return engine.NewObject(result), nil
}

// ReSigs returns function signatures for REPL :doc command.

// builtinReQuote 转义正则特殊字符。
func builtinReQuote(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("re_quote() expects 1 argument, got %d", len(args))
	}
	return engine.NewString(regexp.QuoteMeta(args[0].String())), nil
}

// builtinReFullmatch 检查是否完全匹配。
func builtinReFullmatch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("re_fullmatch() expects 2 arguments, got %d", len(args))
	}
	pattern := args[0].String()
	str := args[1].String()
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return nil, fmt.Errorf("re_fullmatch() invalid pattern: %v", err)
	}
	return engine.NewBool(re.MatchString(str)), nil
}
func ReSigs() map[string]string {
	return map[string]string{
		"re_match":      "re_match(pattern, str) → bool  — Check if pattern matches",
		"re_replace":    "re_replace(pattern, replacement, str) → string  — Replace all matches",
		"re_split":      "re_split(pattern, str) → array  — Split by pattern",
		"re_find":       "re_find(pattern, str) → string  — Find first match",
		"re_find_all":   "re_find_all(pattern, str) → array  — Find all matches",
		"re_groups":     "re_groups(pattern, str) → object  — Get capture groups",
		"re_groups_raw": "re_groups_raw(regex, str) → object  — Get capture groups from regex value",
		"re_quote":      "re_quote(str) → string  — Escape regex special characters",
		"re_fullmatch":  "re_fullmatch(pattern, str) → bool  — Check full string match",
	}
}
