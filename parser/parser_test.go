package parser

import (
	"testing"

	"github.com/gnuos/jpl/lexer"
)

// TestParseNumberLiteral 测试数字字面量解析
func TestParseNumberLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"整数", "42", "42"},
		{"浮点数", "3.14", "3.14"},
		{"科学计数法", "1e10", "1e10"},
		{"十六进制", "0xFF", "0xFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			literal, ok := stmt.Expression.(*NumberLiteral)
			if !ok {
				t.Errorf("期望 NumberLiteral，得到 %T", stmt.Expression)
				return
			}

			if literal.Value != tt.expected {
				t.Errorf("期望 %q，得到 %q", tt.expected, literal.Value)
			}
		})
	}
}

// TestParseStringLiteral 测试字符串字面量解析
func TestParseStringLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"双引号", `"hello"`, "hello"},
		{"单引号", `'world'`, "world"},
		{"空字符串", `""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			literal, ok := stmt.Expression.(*StringLiteral)
			if !ok {
				t.Errorf("期望 StringLiteral，得到 %T", stmt.Expression)
				return
			}

			if literal.Value != tt.expected {
				t.Errorf("期望 %q，得到 %q", tt.expected, literal.Value)
			}
		})
	}
}

// TestParseBoolLiteral 测试布尔字面量解析
func TestParseBoolLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"TRUE", "TRUE", true},
		{"FALSE", "FALSE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			literal, ok := stmt.Expression.(*BoolLiteral)
			if !ok {
				t.Errorf("期望 BoolLiteral，得到 %T", stmt.Expression)
				return
			}

			if literal.Value != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, literal.Value)
			}
		})
	}
}

// TestParseIdentifier 测试标识符解析
func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"$变量", "$myVar", "$myVar"},
		{"普通变量", "myVar", "myVar"},
		{"特殊变量", "$_", "$_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			ident, ok := stmt.Expression.(*Identifier)
			if !ok {
				t.Errorf("期望 Identifier，得到 %T", stmt.Expression)
				return
			}

			if ident.Value != tt.expected {
				t.Errorf("期望 %q，得到 %q", tt.expected, ident.Value)
			}
		})
	}
}

// TestParseBinaryExpression 测试二元表达式解析
func TestParseBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		operator string
	}{
		{"加法", "$a + $b", "+"},
		{"减法", "$a - $b", "-"},
		{"乘法", "$a * $b", "*"},
		{"除法", "$a / $b", "/"},
		{"等于", "$a == $b", "=="},
		{"不等于", "$a != $b", "!="},
		{"小于", "$a < $b", "<"},
		{"大于", "$a > $b", ">"},
		{"逻辑与", "$a && $b", "&&"},
		{"逻辑或", "$a || $b", "||"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			binary, ok := stmt.Expression.(*BinaryExpr)
			if !ok {
				t.Errorf("期望 BinaryExpr，得到 %T", stmt.Expression)
				return
			}

			if binary.Operator != tt.operator {
				t.Errorf("期望运算符 %q，得到 %q", tt.operator, binary.Operator)
			}
		})
	}
}

// TestParseFunctionCall 测试函数调用解析
func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"无参数调用", "print()"},
		{"单参数调用", "print($msg)"},
		{"多参数调用", "add($a, $b)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			_, ok = stmt.Expression.(*CallExpr)
			if !ok {
				t.Errorf("期望 CallExpr，得到 %T", stmt.Expression)
			}
		})
	}
}

// TestParseFuncDecl 测试函数声明解析
func TestParseFuncDecl(t *testing.T) {
	input := `fn add($a, $b) { return $a + $b; }`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	if len(program.Statements) != 1 {
		t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
		return
	}

	decl, ok := program.Statements[0].(*FuncDecl)
	if !ok {
		t.Errorf("期望 FuncDecl，得到 %T", program.Statements[0])
		return
	}

	if decl.Name.Value != "add" {
		t.Errorf("期望函数名 'add'，得到 %q", decl.Name.Value)
	}

	if len(decl.Parameters) != 2 {
		t.Errorf("期望 2 个参数，得到 %d 个", len(decl.Parameters))
	}
}

// TestParseIfStatement 测试 if 语句解析
func TestParseIfStatement(t *testing.T) {
	input := `if ($x > 0) { echo "positive"; }`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	_, ok := program.Statements[0].(*IfStmt)
	if !ok {
		t.Errorf("期望 IfStmt，得到 %T", program.Statements[0])
	}
}

// TestParseWhileStatement 测试 while 语句解析
func TestParseWhileStatement(t *testing.T) {
	input := `while ($i < 10) { $i = $i + 1; }`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	_, ok := program.Statements[0].(*WhileStmt)
	if !ok {
		t.Errorf("期望 WhileStmt，得到 %T", program.Statements[0])
	}
}

// TestParseVarDecl 测试变量声明解析
func TestParseVarDecl(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"带$前缀", "$x = 42"},
		{"不带$前缀", "x = 42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			_, ok := program.Statements[0].(*VarDecl)
			if !ok {
				t.Errorf("期望 VarDecl，得到 %T", program.Statements[0])
			}
		})
	}
}

// TestParseConstDecl 测试常量声明解析
func TestParseConstDecl(t *testing.T) {
	input := `const MAX = 100;`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	_, ok := program.Statements[0].(*ConstDecl)
	if !ok {
		t.Errorf("期望 ConstDecl，得到 %T", program.Statements[0])
	}
}

// TestParseArrayLiteral 测试数组字面量解析
func TestParseArrayLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"空数组", "[]", 0},
		{"单元素数组", "[1]", 1},
		{"多元素数组", "[1, 2, 3]", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			array, ok := stmt.Expression.(*ArrayLiteral)
			if !ok {
				t.Errorf("期望 ArrayLiteral，得到 %T", stmt.Expression)
				return
			}

			if len(array.Elements) != tt.expected {
				t.Errorf("期望 %d 个元素，得到 %d 个", tt.expected, len(array.Elements))
			}
		})
	}
}

// TestParseStringConcat 测试字符串连接解析
func TestParseStringConcat(t *testing.T) {
	input := `"Hello" .. " " .. "World"`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	stmt, ok := program.Statements[0].(*ExprStmt)
	if !ok {
		t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
		return
	}

	// 第一个 ..
	_, ok = stmt.Expression.(*ConcatExpr)
	if !ok {
		t.Errorf("期望 ConcatExpr，得到 %T", stmt.Expression)
	}
}

// TestParseArrowFunction 测试箭头函数解析
func TestParseArrowFunction(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"表达式体", "$x -> $x * 2"},
		{"块体", "$x -> { return $x * 2; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			_, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
			}
		})
	}
}

// TestParseTernaryExpression 测试三元表达式解析
func TestParseTernaryExpression(t *testing.T) {
	input := `$x > 0 ? "positive" : "negative"`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	stmt, ok := program.Statements[0].(*ExprStmt)
	if !ok {
		t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
		return
	}

	_, ok = stmt.Expression.(*TernaryExpr)
	if !ok {
		t.Errorf("期望 TernaryExpr，得到 %T", stmt.Expression)
	}
}

// TestParseMatchStatement 测试 match 语句解析
func TestParseMatchStatement(t *testing.T) {
	input := `match (42) {
    case 0: {
        println "值为零"
    }
    case 1: {
        println "值为一"
    }
    case 2 || 3 || 5 || 7 || 11 || 13 || 17 || 19: {
        println "值是小于20的质数"
    }
    case 20...30: {
        println "值在20到30之间（不包括边界）"
    }
    case 30..=50: {
        println "值在30到50之间（包括边界）"
    }
    case $s if $s > 50: {
        println "值大于50"
    }
    case _: {
        println "其他值: " .. $value
    }
}`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	matchStmt, ok := program.Statements[0].(*MatchStmt)
	if !ok {
		t.Errorf("期望 MatchStmt，得到 %T", program.Statements[0])
		return
	}

	if len(matchStmt.Cases) != 7 {
		t.Errorf("期望 7 个 case，得到 %d 个", len(matchStmt.Cases))
	}
}

// TestParseImportStatement 测试 import 语句解析
func TestParseImportStatement(t *testing.T) {
	input := `import "utils.jpl";`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	importStmt, ok := program.Statements[0].(*ImportStmt)
	if !ok {
		t.Errorf("期望 ImportStmt，得到 %T", program.Statements[0])
		return
	}

	if importStmt.Source != "utils.jpl" {
		t.Errorf("期望导入源 'utils.jpl'，得到 %q", importStmt.Source)
	}
}

// TestParseImportAsStatement 测试 import ... as ... 语句解析
func TestParseImportAsStatement(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		source string
		alias  string
	}{
		{"基本别名", `import "math" as m;`, "math", "m"},
		{"URL 别名", `import "https://example.com/lib.jpl" as lib;`, "https://example.com/lib.jpl", "lib"},
		{"无分号", `import "utils" as u`, "utils", "u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ImportStmt)
			if !ok {
				t.Errorf("期望 ImportStmt，得到 %T", program.Statements[0])
				return
			}

			if stmt.Source != tt.source {
				t.Errorf("期望源 %q，得到 %q", tt.source, stmt.Source)
			}
			if stmt.Alias == nil {
				t.Fatal("Alias 不应为 nil")
			}
			if stmt.Alias.Value != tt.alias {
				t.Errorf("期望别名 %q，得到 %q", tt.alias, stmt.Alias.Value)
			}
		})
	}
}

// TestParseIncludeStatement 测试 include 语句解析
func TestParseIncludeStatement(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		source string
		once   bool
	}{
		{"include", `include "utils.jpl";`, "utils.jpl", false},
		{"include_once", `include_once "config.jpl";`, "config.jpl", true},
		{"include 无分号", `include "helpers.jpl"`, "helpers.jpl", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*IncludeStmt)
			if !ok {
				t.Errorf("期望 IncludeStmt，得到 %T", program.Statements[0])
				return
			}

			if stmt.Source != tt.source {
				t.Errorf("期望源 %q，得到 %q", tt.source, stmt.Source)
			}
			if stmt.Once != tt.once {
				t.Errorf("期望 Once=%v，得到 %v", tt.once, stmt.Once)
			}
		})
	}
}

// TestParseFromImportStatement 测试 from ... import ... 语句解析
func TestParseFromImportStatement(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		source string
		names  []string
	}{
		{"单个导入", `from "math" import sqrt;`, "math", []string{"sqrt"}},
		{"多个导入", `from "math" import sqrt, abs, pow;`, "math", []string{"sqrt", "abs", "pow"}},
		{"无分号", `from "utils" import helper`, "utils", []string{"helper"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ImportStmt)
			if !ok {
				t.Errorf("期望 ImportStmt，得到 %T", program.Statements[0])
				return
			}

			if stmt.Source != tt.source {
				t.Errorf("期望源 %q，得到 %q", tt.source, stmt.Source)
			}

			if len(stmt.Names) != len(tt.names) {
				t.Errorf("期望 %d 个名称，得到 %d", len(tt.names), len(stmt.Names))
				return
			}

			for i, name := range tt.names {
				if stmt.Names[i].Value != name {
					t.Errorf("名称[%d] 期望 %q，得到 %q", i, name, stmt.Names[i].Value)
				}
			}
		})
	}
}

// TestParseSpecialCallSingleArg 测试特例函数单参数无括号调用
func TestParseSpecialCallSingleArg(t *testing.T) {
	tests := []struct {
		name  string
		input string
		fn    string
	}{
		{"print 字符串", `print "hello"`, "print"},
		{"println 字符串", `println "world"`, "println"},
		{"echo 字符串", `echo "test"`, "echo"},
		{"log 字符串", `log "debug"`, "log"},
		{"assert 条件", `assert ok`, "assert"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			call, ok := stmt.Expression.(*CallExpr)
			if !ok {
				t.Errorf("期望 CallExpr，得到 %T", stmt.Expression)
				return
			}

			ident, ok := call.Function.(*Identifier)
			if !ok {
				t.Errorf("期望函数名为 Identifier，得到 %T", call.Function)
				return
			}

			if ident.Value != tt.fn {
				t.Errorf("期望函数名 %q，得到 %q", tt.fn, ident.Value)
			}

			if len(call.Arguments) != 1 {
				t.Errorf("期望 1 个参数，得到 %d 个", len(call.Arguments))
			}
		})
	}
}

// TestParseSpecialCallMultiArg 测试特例函数多参数逗号分隔调用
func TestParseSpecialCallMultiArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fn       string
		argCount int
	}{
		{"format 多参数", `format "%d %s", x, name`, "format", 3},
		{"echo 多参数", `echo "value:", count`, "echo", 2},
		{"print 多参数", `print a, b, c`, "print", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			call, ok := stmt.Expression.(*CallExpr)
			if !ok {
				t.Errorf("期望 CallExpr，得到 %T", stmt.Expression)
				return
			}

			ident, ok := call.Function.(*Identifier)
			if !ok || ident.Value != tt.fn {
				t.Errorf("期望函数名 %q", tt.fn)
				return
			}

			if len(call.Arguments) != tt.argCount {
				t.Errorf("期望 %d 个参数，得到 %d 个", tt.argCount, len(call.Arguments))
			}
		})
	}
}

// TestParseSpecialCallWithSemicolon 测试特例函数带分号结尾
func TestParseSpecialCallWithSemicolon(t *testing.T) {
	input := `print "hello";`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	if len(program.Statements) != 1 {
		t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
	}
}

// TestParseSpecialCallNested 测试特例函数嵌套调用（必须用括号）
func TestParseSpecialCallNested(t *testing.T) {
	input := `print(len(arr))`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	stmt, ok := program.Statements[0].(*ExprStmt)
	if !ok {
		t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
		return
	}

	call, ok := stmt.Expression.(*CallExpr)
	if !ok {
		t.Errorf("期望 CallExpr，得到 %T", stmt.Expression)
		return
	}

	// print 的参数应该是一个 CallExpr (len(arr))
	if len(call.Arguments) != 1 {
		t.Errorf("期望 print 有 1 个参数，得到 %d 个", len(call.Arguments))
		return
	}

	innerCall, ok := call.Arguments[0].(*CallExpr)
	if !ok {
		t.Errorf("期望嵌套 CallExpr，得到 %T", call.Arguments[0])
		return
	}

	innerIdent, ok := innerCall.Function.(*Identifier)
	if !ok || innerIdent.Value != "len" {
		t.Errorf("期望内部调用为 len")
	}
}

// TestParseNonSpecialFuncError 测试非特例函数后跟值类型报语法错误
func TestParseNonSpecialFuncError(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"普通函数字符串", `myFunc "hello"`},
		{"普通函数数字", `foo 42`},
		{"普通函数变量", `bar x`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			p.Parse()

			if len(p.Errors()) == 0 {
				t.Errorf("期望语法错误，但解析成功")
				return
			}

			// 检查错误信息包含关键词
			errMsg := p.Errors()[0]
			if errMsg == "" {
				t.Errorf("错误信息为空")
			}
		})
	}
}

// TestParseSpecialCallWithExpression 测试特例函数参数为表达式
func TestParseSpecialCallWithExpression(t *testing.T) {
	input := `print a + b, c * d`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	stmt, ok := program.Statements[0].(*ExprStmt)
	if !ok {
		t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
		return
	}

	call, ok := stmt.Expression.(*CallExpr)
	if !ok {
		t.Errorf("期望 CallExpr，得到 %T", stmt.Expression)
		return
	}

	if len(call.Arguments) != 2 {
		t.Errorf("期望 2 个参数，得到 %d 个", len(call.Arguments))
		return
	}

	// 第一个参数应该是 BinaryExpr (a + b)
	_, ok = call.Arguments[0].(*BinaryExpr)
	if !ok {
		t.Errorf("期望第一个参数为 BinaryExpr，得到 %T", call.Arguments[0])
	}

	// 第二个参数应该是 BinaryExpr (c * d)
	_, ok = call.Arguments[1].(*BinaryExpr)
	if !ok {
		t.Errorf("期望第二个参数为 BinaryExpr，得到 %T", call.Arguments[1])
	}
}

// TestParseMultilineArrayLiteral 测试多行数组字面量
func TestParseMultilineArrayLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"单行数组", `[1, 2, 3]`, 3},
		{"多行数组带逗号", "[\n  1,\n  2,\n  3\n]", 3},
		{"多行数组无逗号", "[\n  1\n  2\n  3\n]", 3},
		{"多行数组混合", "[\n  1,\n  2\n  3\n]", 3},
		{"空多行数组", "[\n]", 0},
		{"连续换行数组", "[\n\n\n  1\n\n  2\n\n]", 2},
		{"单元素多行数组", "[\n  42\n]", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
				return
			}

			stmt, ok := program.Statements[0].(*ExprStmt)
			if !ok {
				t.Errorf("期望 ExprStmt，得到 %T", program.Statements[0])
				return
			}

			arr, ok := stmt.Expression.(*ArrayLiteral)
			if !ok {
				t.Errorf("期望 ArrayLiteral，得到 %T", stmt.Expression)
				return
			}

			if len(arr.Elements) != tt.expected {
				t.Errorf("期望 %d 个元素，得到 %d 个", tt.expected, len(arr.Elements))
			}
		})
	}
}

// TestParseMultilineObjectLiteral 测试多行对象字面量
// 对象字面量必须在赋值或表达式中使用，否则 { 会被解析为代码块
func TestParseMultilineObjectLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"单行对象", `$obj = {name: "John", age: 30}`, 2},
		{"多行对象带逗号", "$obj = {\n  name: \"John\",\n  age: 30\n}", 2},
		{"多行对象无逗号", "$obj = {\n  name: \"John\"\n  age: 30\n}", 2},
		{"多行对象混合", "$obj = {\n  a: 1,\n  b: 2\n  c: 3\n}", 3},
		{"空多行对象", "$obj = {\n}", 0},
		{"连续换行对象", "$obj = {\n\n\n  a: 1\n\n  b: 2\n\n}", 2},
		{"单键多行对象", "$obj = {\n  key: \"value\"\n}", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.NewLexer(tt.input, "test.jpl")
			p := NewParser(l)
			program := p.Parse()

			if len(p.Errors()) > 0 {
				t.Errorf("解析错误: %v", p.Errors())
				return
			}

			if len(program.Statements) != 1 {
				t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
				return
			}

			stmt, ok := program.Statements[0].(*VarDecl)
			if !ok {
				t.Errorf("期望 VarDecl，得到 %T", program.Statements[0])
				return
			}

			obj, ok := stmt.Value.(*ObjectLiteral)
			if !ok {
				t.Errorf("期望 ObjectLiteral，得到 %T", stmt.Value)
				return
			}

			if len(obj.Pairs) != tt.expected {
				t.Errorf("期望 %d 个键值对，得到 %d 个", tt.expected, len(obj.Pairs))
			}
		})
	}
}

// TestParseNestedMultilineLiteral 测试嵌套多行字面量
func TestParseNestedMultilineLiteral(t *testing.T) {
	// 使用变量赋值，避免 { 被解析为代码块
	input := `$obj = {
	name: "test",
	items: [
		1,
		2,
		3
	],
	config: {
		enabled: true
		timeout: 30
	}
}`

	l := lexer.NewLexer(input, "test.jpl")
	p := NewParser(l)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		t.Errorf("解析错误: %v", p.Errors())
		return
	}

	if len(program.Statements) != 1 {
		t.Errorf("期望 1 个语句，得到 %d 个", len(program.Statements))
		return
	}

	stmt, ok := program.Statements[0].(*VarDecl)
	if !ok {
		t.Errorf("期望 VarDecl，得到 %T", program.Statements[0])
		return
	}

	obj, ok := stmt.Value.(*ObjectLiteral)
	if !ok {
		t.Errorf("期望 ObjectLiteral，得到 %T", stmt.Value)
		return
	}

	if len(obj.Pairs) != 3 {
		t.Errorf("期望 3 个键值对，得到 %d 个", len(obj.Pairs))
		return
	}
}
