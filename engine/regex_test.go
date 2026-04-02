package engine

import (
	"fmt"
	"regexp"
	"testing"
)

// ==================== 正则字面量测试 ====================

func compileAndRunRegex(t *testing.T, script string) string {
	t.Helper()
	prog, err := CompileStringWithGlobals(script, "<test>", nil)
	if err != nil {
		t.Fatalf("编译错误: %v", err)
	}
	eng := NewEngine()
	eng.RegisterFunc("is_regex", func(ctx *Context, args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("is_regex() expects 1 argument")
		}
		return NewBool(IsRegex(args[0])), nil
	})
	eng.RegisterFunc("re_groups_raw", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("re_groups_raw() expects 2 arguments")
		}
		reVal := args[0]
		subject := args[1].String()
		if IsRegex(reVal) {
			type regexIface interface {
				Regexp() *regexp.Regexp
			}
			re := reVal.(regexIface).Regexp()
			matches := re.FindStringSubmatch(subject)
			if matches == nil {
				return NewNull(), nil
			}
			result := make(map[string]Value)
			for i, m := range matches {
				result[fmt.Sprintf("%d", i)] = NewString(m)
			}
			names := re.SubexpNames()
			for i, name := range names {
				if i > 0 && i < len(matches) && name != "" {
					result[name] = NewString(matches[i])
				}
			}
			return NewObject(result), nil
		}
		return NewNull(), nil
	})
	defer eng.Close()
	vm := newVMWithProgram(eng, prog)
	if err := vm.Execute(); err != nil {
		t.Fatalf("执行错误: %v", err)
	}
	return vm.GetResult().Stringify()
}

// TestRegexLiteralBasic 测试基本正则字面量创建和匹配
func TestRegexLiteralBasic(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		expect string
	}{
		{
			name:   "基本匹配 - 数字",
			code:   `return "abc123" =~ #/\d+/#;`,
			expect: "true",
		},
		{
			name:   "基本匹配 - 不匹配",
			code:   `return "abc" =~ #/\d+/#;`,
			expect: "false",
		},
		{
			name:   "精确匹配 - 锚定",
			code:   `return "hello" =~ #/^hello$/#;`,
			expect: "true",
		},
		{
			name:   "精确匹配 - 不完全匹配",
			code:   `return "hello world" =~ #/^hello$/#;`,
			expect: "false",
		},
		{
			name:   "忽略大小写 flag",
			code:   `return "Hello" =~ #/hello/i#;`,
			expect: "true",
		},
		{
			name:   "忽略大小写 - 不匹配",
			code:   `return "Hello" =~ #/hello/#;`,
			expect: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileAndRunRegex(t, tt.code)
			if result != tt.expect {
				t.Errorf("期望 %s，得到 %s", tt.expect, result)
			}
		})
	}
}

// TestRegexLiteralCompileError 测试编译期错误检测
func TestRegexLiteralCompileError(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{name: "空正则", code: `return #//#;`},
		{name: "缺少结尾#", code: `return "abc" =~ #/\d+/;`},
		{name: "无效正则语法", code: `return #/[/invalid/#;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileStringWithGlobals(tt.code, "<test>", nil)
			if err == nil {
				t.Error("预期编译错误，但成功编译")
			}
		})
	}
}

// TestRegexLiteralFlags 测试正则 flags
func TestRegexLiteralFlags(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		expect string
	}{
		{
			name:   "flag i - 忽略大小写",
			code:   `return "ABC" =~ #/abc/i#;`,
			expect: "true",
		},
		{
			name:   "flag m - 多行模式",
			code:   "return \"line1\\nline2\" =~ #/^line2$/m#;",
			expect: "true",
		},
		{
			name:   "flag s - dot匹配换行",
			code:   "return \"a\\nb\" =~ #/a.b/s#;",
			expect: "true",
		},
		{
			name:   "flag 组合 im",
			code:   "return \"LINE1\\nLINE2\" =~ #/^line2$/im#;",
			expect: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileAndRunRegex(t, tt.code)
			if result != tt.expect {
				t.Errorf("期望 %s，得到 %s", tt.expect, result)
			}
		})
	}
}

// TestRegexLiteralVariable 测试正则赋值给变量
func TestRegexLiteralVariable(t *testing.T) {
	result := compileAndRunRegex(t, `
		$re = #/\d+/#
		return "abc123" =~ $re;
	`)
	if result != "true" {
		t.Errorf("期望 true，得到 %s", result)
	}
}

// TestRegexIsRegex 测试 is_regex 函数
func TestRegexIsRegex(t *testing.T) {
	result := compileAndRunRegex(t, `
		$re = #/\d+/#
		$x = "hello"
		return is_regex($re) && !is_regex($x);
	`)
	if result != "true" {
		t.Errorf("期望 true，得到 %s", result)
	}
}

// ==================== match/case 正则模式测试 ====================

// TestMatchRegexBasic 测试 match/case 基本正则匹配
func TestMatchRegexBasic(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		expect string
	}{
		{
			name: "正则匹配第一个分支",
			code: `
				$input = "hello123"
				return match ($input) {
					case #/^hello\d+$/#: "greeting with number"
					case #/^\d+$/#: "just number"
					case _: "unknown"
				}
			`,
			expect: `"greeting with number"`,
		},
		{
			name: "正则匹配第二个分支",
			code: `
				$input = "456"
				return match ($input) {
					case #/^hello\d+$/#: "greeting with number"
					case #/^\d+$/#: "just number"
					case _: "unknown"
				}
			`,
			expect: `"just number"`,
		},
		{
			name: "正则匹配兜底分支",
			code: `
				$input = "xyz"
				return match ($input) {
					case #/^hello\d+$/#: "greeting with number"
					case #/^\d+$/#: "just number"
					case _: "unknown"
				}
			`,
			expect: `"unknown"`,
		},
		{
			name: "正则忽略大小写",
			code: `
				$input = "QUIT"
				return match ($input) {
					case #/^quit$/i#: "exit"
					case _: "continue"
				}
			`,
			expect: `"exit"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileAndRunRegex(t, tt.code)
			if result != tt.expect {
				t.Errorf("期望 %s，得到 %s", tt.expect, result)
			}
		})
	}
}

// TestMatchRegexAsBinding 测试 match/case 正则 as 绑定
func TestMatchRegexAsBinding(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		expect string
	}{
		{
			name: "捕获组绑定 - 单个组",
			code: `
				$input = "set name=Alice"
				return match ($input) {
					case #/^set (\w+)=(.+)$/# as $m: $m[1]
					case _: "none"
				}
			`,
			expect: `"name"`,
		},
		{
			name: "捕获组绑定 - 多个组",
			code: `
				$input = "2024-03"
				return match ($input) {
					case #/^(\d{4})-(\d{2})$/# as $m: $m[1]
					case _: "none"
				}
			`,
			expect: `"2024"`,
		},
		{
			name: "捕获组绑定 - 完整匹配",
			code: `
				$input = "123-456"
				return match ($input) {
					case #/^(\d{3})-(\d{3})$/# as $m: $m[0]
					case _: "none"
				}
			`,
			expect: `"123-456"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compileAndRunRegex(t, tt.code)
			if result != tt.expect {
				t.Errorf("期望 %s，得到 %s", tt.expect, result)
			}
		})
	}
}

// TestMatchRegexMixed 测试字面量和正则混合匹配
func TestMatchRegexMixed(t *testing.T) {
	result := compileAndRunRegex(t, `
		$input = "quit"
		return match ($input) {
			case "quit", "exit": "bye"
			case #/^\d+$/#: "number"
			case #/^hello$/i#: "greeting"
			case _: "other"
		}
	`)
	if result != `"bye"` {
		t.Errorf("期望 \"bye\"，得到 %s", result)
	}
}
