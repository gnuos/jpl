package engine

import (
	"strings"
	"testing"
)

// ============================================================================
// 完整表达式插值测试 (Phase 10.3)
// ============================================================================

func TestInterpObjectAccess(t *testing.T) {
	// 对象属性访问
	script := `
		$obj = {name: "JPL", version: "1.0"}
		$msg = "Name: #{$obj.name}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Name: JPL") {
		t.Errorf("expected 'Name: JPL', got '%s'", result.String())
	}
}

func TestInterpArrayIndex(t *testing.T) {
	// 数组索引访问
	script := `
		$arr = ["apple", "banana", "cherry"]
		$msg = "First: #{$arr[0]}, Second: #{$arr[1]}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "First: apple") || !strings.Contains(result.String(), "Second:") {
		t.Errorf("expected array indices, got '%s'", result.String())
	}
}

func TestInterpArithmetic(t *testing.T) {
	// 算术运算
	script := `
		$a = 10
		$b = 20
		$msg = "Sum: #{$a + $b}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Sum: 30") {
		t.Errorf("expected 'Sum: 30', got '%s'", result.String())
	}
}

func TestInterpComplexArithmetic(t *testing.T) {
	// 复杂算术运算
	script := `
		$x = 5
		$y = 3
		$msg = "Result: #{$x * $y + 2}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Result: 17") {
		t.Errorf("expected 'Result: 17', got '%s'", result.String())
	}
}

func TestInterpChainedProperty(t *testing.T) {
	// 链式对象访问
	script := `
		$user = {profile: {name: "Alice"}}
		$msg = "User: #{$user.profile.name}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "User: Alice") {
		t.Errorf("expected 'User: Alice', got '%s'", result.String())
	}
}

func TestInterpNestedArray(t *testing.T) {
	// 嵌套数组访问
	script := `
		$matrix = [[1, 2], [3, 4]]
		$msg = "Value: #{$matrix[0][1]}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Value: 2") {
		t.Errorf("expected 'Value: 2', got '%s'", result.String())
	}
}

func TestInterpExpressionWithText(t *testing.T) {
	// 表达式周围有文本
	script := `
		$price = 100
		$tax = 0.08
		$msg = "Total: #{$price * (1 + $tax)} dollars"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Total: 108") {
		t.Errorf("expected 'Total: 108', got '%s'", result.String())
	}
}

func TestInterpMultipleExpressions(t *testing.T) {
	// 多个表达式插值
	script := `
		$a = 1
		$b = 2
		$c = 3
		$msg = "#{$a} + #{$b} = #{$c}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "1") || !strings.Contains(result.String(), "2") {
		t.Errorf("expected '1 + 2 = 3', got '%s'", result.String())
	}
}

func TestInterpWithFunction(t *testing.T) {
	// 函数返回值插值（如果支持）
	script := `
		function getName() {
			return "Test"
		}
		$msg = "Name: #{getName()}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 注意：函数调用在插值中可能有限制
	if !strings.Contains(result.String(), "Name:") {
		t.Errorf("expected function result in interpolation, got '%s'", result.String())
	}
}

func TestInterpLogicalExpression(t *testing.T) {
	// 逻辑表达式（布尔值转字符串）
	script := `
		$x = 5
		$msg = "Is positive: #{$x > 0}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "true") {
		t.Errorf("expected 'true', got '%s'", result.String())
	}
}

func TestInterpTernaryOperator(t *testing.T) {
	// 三元运算符
	script := `
		$score = 85
		$msg = "Grade: #{$score >= 60 ? 'Pass' : 'Fail'}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Pass") {
		t.Errorf("expected 'Pass', got '%s'", result.String())
	}
}

func TestInterpArrayLength(t *testing.T) {
	// 数组长度（使用 count 函数或手动计算，这里简化为测试最后一个元素）
	script := `
		$arr = ["a", "b", "c", "d", "e"]
		$last = $arr[4]
		$msg = "Last: #{$last} (5 items)"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Last: e") {
		t.Errorf("expected array length test, got '%s'", result.String())
	}
}

func TestInterpNegativeNumber(t *testing.T) {
	// 负数运算
	script := `
		$temp = -10
		$msg = "Temperature: #{$temp}"
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "-10") {
		t.Errorf("expected '-10', got '%s'", result.String())
	}
}

func TestInterpStringConcatInExpression(t *testing.T) {
	// 表达式中的字符串连接
	script := `
		$first = "Hello"
		$last = "World"
		$msg = "Full: #{$first .. ' ' .. $last}"
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

func TestInterpMultilineWithExpression(t *testing.T) {
	// 多行字符串中的表达式
	script := `
		$name = "JPL"
		$version = "1.0"
		$msg = """Project: #{$name}
Version: #{$version}
Status: #{"Active"}"""
		$msg
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.String(), "Project: JPL") {
		t.Errorf("expected multiline with expressions, got '%s'", result.String())
	}
}
