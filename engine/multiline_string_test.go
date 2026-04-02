package engine

import (
	"strings"
	"testing"
)

// ============================================================================
// 多行字符串测试
// ============================================================================

func TestTripleSingleQuoteString(t *testing.T) {
	// 基本多行字符串
	script := `
		$text = '''Hello
World'''
		$text
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Hello") || !strings.Contains(result.String(), "World") {
		t.Errorf("expected 'Hello\\nWorld', got '%s'", result.String())
	}
}

func TestTripleDoubleQuoteString(t *testing.T) {
	// 双引号多行字符串
	script := `
		$text = """Line1
Line2
Line3"""
		$text
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Line1") || !strings.Contains(result.String(), "Line2") {
		t.Errorf("expected multi-line content, got '%s'", result.String())
	}
}

func TestTripleQuoteJSON(t *testing.T) {
	// JSON 格式的多行字符串
	script := `
		$json = '''
{
    "name": "JPL",
    "version": "1.0"
}
'''
		$json
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "name") || !strings.Contains(result.String(), "JPL") {
		t.Errorf("expected JSON content, got '%s'", result.String())
	}
}

func TestTripleQuoteEmpty(t *testing.T) {
	// 空的多行字符串
	script := `
		$text = ''''''
		$text == ""
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Errorf("expected empty string, got '%s'", result.String())
	}
}

func TestTripleQuoteWithEscapes(t *testing.T) {
	// 带转义字符的多行字符串
	script := `
		$text = '''Line1\nLine2\tTab'''
		$text
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 检查转义是否被正确处理
	if !strings.Contains(result.String(), "Line1") {
		t.Errorf("expected escaped content, got '%s'", result.String())
	}
}

func TestTripleQuoteVsSingleQuote(t *testing.T) {
	// 对比单引号字符串和三引号字符串
	script := `
		$single = 'hello'
		$triple = '''hello'''
		$single == $triple
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type().String() != "bool" || !result.Bool() {
		t.Errorf("expected single and triple quote strings to be equal")
	}
}

func TestTripleQuoteAssignment(t *testing.T) {
	// 多行字符串赋值和使用
	script := `
		$template = """Dear User,

Welcome to our service!

Best regards,
Team"""
		$template != ""
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Errorf("expected non-empty template")
	}
}
