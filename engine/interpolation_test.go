package engine

import (
	"strings"
	"testing"
)

// ============================================================================
// 字符串插值测试 (Phase 10.2)
// ============================================================================

func TestStringInterpolationBasic(t *testing.T) {
	// 基本变量插值
	script := `
		$name = "World"
		$greeting = "Hello #{$name}!"
		$greeting
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Hello") || !strings.Contains(result.String(), "World") {
		t.Errorf("expected 'Hello World!', got '%s'", result.String())
	}
}

func TestStringInterpolationMultipleVars(t *testing.T) {
	// 多个变量插值
	script := `
		$first = "John"
		$last = "Doe"
		$name = "#{$first} #{$last}"
		$name
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "John") || !strings.Contains(result.String(), "Doe") {
		t.Errorf("expected 'John Doe', got '%s'", result.String())
	}
}

func TestStringInterpolationWithText(t *testing.T) {
	// 插值周围有文本
	script := `
		$x = 42
		$msg = "The answer is #{$x}."
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "The answer is 42") {
		t.Errorf("expected 'The answer is 42.', got '%s'", result.String())
	}
}

func TestStringInterpolationEmpty(t *testing.T) {
	// 空变量插值
	script := `
		$empty = ""
		$result = "[#{$empty}]"
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "[]" {
		t.Errorf("expected '[]', got '%s'", result.String())
	}
}

func TestStringInterpolationMultiline(t *testing.T) {
	// 多行字符串插值
	script := `
		$name = "Alice"
		$msg = """Hello #{$name},
Welcome!"""
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Hello Alice") || !strings.Contains(result.String(), "Welcome") {
		t.Errorf("expected multiline with interpolation, got '%s'", result.String())
	}
}

func TestStringInterpolationAtStart(t *testing.T) {
	// 插值在开头
	script := `
		$x = "Hello"
		$msg = "#{$x} World"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Hello World") {
		t.Errorf("expected 'Hello World', got '%s'", result.String())
	}
}

func TestStringInterpolationAtEnd(t *testing.T) {
	// 插值在结尾
	script := `
		$x = "World"
		$msg = "Hello #{$x}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Hello World") {
		t.Errorf("expected 'Hello World', got '%s'", result.String())
	}
}

func TestStringInterpolationOnly(t *testing.T) {
	// 只有插值，没有其他文本
	script := `
		$x = "test"
		$msg = "#{$x}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "test" {
		t.Errorf("expected 'test', got '%s'", result.String())
	}
}

func TestStringInterpolationComparison(t *testing.T) {
	// 插值结果与手动连接对比
	script := `
		$name = "JPL"
		$interp = "Hello #{$name}!"
		$concat = "Hello " .. $name .. "!"
		$interp == $concat
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Errorf("interpolated string should equal concatenated string")
	}
}

func TestStringInterpolationSpecialVar(t *testing.T) {
	// 特殊变量插值
	script := `
		$_ = "special"
		$msg = "Value: #{$_}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "special") {
		t.Errorf("expected 'Value: special', got '%s'", result.String())
	}
}

func TestStringInterpolationEscaped(t *testing.T) {
	// 转义 \# 防止插值（在 JPL 中，\# 被解析为字面量 #，从而打断 #{ 的识别）
	script := `
		$msg = "Use \#{$var} for interpolation"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该输出字面量 "Use #{$var} for interpolation"，而不是插值
	expected := "Use #{$var} for interpolation"
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result.String())
	}
}

func TestStringInterpolationEscapedWithVar(t *testing.T) {
	// 混合：转义部分 + 实际插值
	script := `
		$name = "World"
		$msg = "Say \#{$name}, Hello #{$name}!"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该包含字面量 #{$name} 和插值后的 World
	expected := "Say #{$name}, Hello World!"
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result.String())
	}
}

func TestStringInterpolationEscapedAtEnd(t *testing.T) {
	// 在字符串末尾转义
	script := `
		$msg = "End with \#{$"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "End with #{$"
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result.String())
	}
}

func TestStringInterpolationMultipleEscaped(t *testing.T) {
	// 多个转义
	script := `
		$msg = "\#{$a} and \#{$b} are escaped"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "#{$a} and #{$b} are escaped"
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result.String())
	}
}
