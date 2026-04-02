package lexer

import (
	"testing"

	"github.com/gnuos/jpl/token"
)

// TestBasicTokens 测试基本 Token 扫描
func TestBasicTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{"空输入", "", []token.TokenType{token.EOF}},
		{"空白字符", "   \t  ", []token.TokenType{token.EOF}},
		{"单个换行", "\n", []token.TokenType{token.NEWLINE, token.EOF}},
		{"多个换行", "\n\n\n", []token.TokenType{token.NEWLINE, token.NEWLINE, token.NEWLINE, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tokens := l.ScanAll()
			if len(tokens) != len(tt.expected) {
				t.Errorf("期望 %d 个 Token，得到 %d 个", len(tt.expected), len(tokens))
				return
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("Token[%d]: 期望 %v，得到 %v", i, tt.expected[i], tok.Type)
				}
			}
		})
	}
}

// TestKeywords 测试关键字扫描
func TestKeywords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"if", "if", token.IF, "if"},
		{"else", "else", token.ELSE, "else"},
		{"while", "while", token.WHILE, "while"},
		{"for", "for", token.FOR, "for"},
		{"foreach", "foreach", token.FOREACH, "foreach"},
		{"function", "function", token.FUNCTION, "function"},
		{"fn", "fn", token.FUNCTION, "fn"},
		{"return", "return", token.RETURN, "return"},
		{"break", "break", token.BREAK, "break"},
		{"continue", "continue", token.CONTINUE, "continue"},
		{"import", "import", token.IMPORT, "import"},
		{"static", "static", token.STATIC, "static"},
		{"global", "global", token.GLOBAL, "global"},
		{"const", "const", token.CONST, "const"},
		{"try", "try", token.TRY, "try"},
		{"catch", "catch", token.CATCH, "catch"},
		{"throw", "throw", token.THROW, "throw"},
		{"match", "match", token.MATCH, "match"},
		{"case", "case", token.CASE, "case"},
		{"default", "default", token.DEFAULT, "default"},
		{"null", "null", token.KW_NULL, "null"},
		{"NULL", "NULL", token.KW_NULL, "NULL"},
		{"true", "true", token.KW_TRUE, "true"},
		{"TRUE", "TRUE", token.KW_TRUE, "TRUE"},
		{"True", "True", token.KW_TRUE, "True"},
		{"false", "false", token.KW_FALSE, "false"},
		{"FALSE", "FALSE", token.KW_FALSE, "FALSE"},
		{"False", "False", token.KW_FALSE, "False"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestIdentifiers 测试标识符扫描
func TestIdentifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"$变量", "$myVar", token.IDENTIFIER, "$myVar"},
		{"$单字符", "$x", token.IDENTIFIER, "$x"},
		{"$下划线变量", "$my_var", token.IDENTIFIER, "$my_var"},
		{"普通标识符", "myVar", token.IDENTIFIER, "myVar"},
		{"首字母大写", "MyVar", token.IDENTIFIER, "MyVar"},
		{"下划线开头", "_private", token.IDENTIFIER, "_private"},
		{"单个下划线", "_", token.UNDERSCORE, "_"},
		{"特殊变量$_", "$_", token.SPECIAL_VAR, "$_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestNumbers 测试数字扫描
func TestNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"普通整数", "42", token.INTEGER, "42"},
		{"零", "0", token.INTEGER, "0"},
		{"大整数", "9999999999999999999", token.BIGINT, "9999999999999999999"},
		{"浮点数", "3.14", token.FLOAT, "3.14"},
		{"前导零浮点", "0.5", token.FLOAT, "0.5"},
		{"无整数部分浮点", ".5", token.FLOAT, ".5"},
		{"科学计数法小写", "1e10", token.FLOAT, "1e10"},
		{"科学计数法大写", "1E10", token.FLOAT, "1E10"},
		{"科学计数法负指数", "1e-5", token.FLOAT, "1e-5"},
		{"科学计数法正指数", "1e+5", token.FLOAT, "1e+5"},
		{"浮点科学计数法", "1.23e10", token.FLOAT, "1.23e10"},
		{"十六进制", "0xFF", token.INTEGER, "0xFF"},
		{"十六进制小写", "0x1a2b", token.INTEGER, "0x1a2b"},
		{"八进制", "0o77", token.INTEGER, "0o77"},
		{"八进制大写", "0O123", token.INTEGER, "0O123"},
		{"二进制", "0b1010", token.INTEGER, "0b1010"},
		{"二进制大写", "0B1111", token.INTEGER, "0B1111"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestStrings 测试字符串扫描
func TestStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"双引号空字符串", `""`, token.STRING, ""},
		{"单引号空字符串", `''`, token.STRING, ""},
		{"双引号字符串", `"hello"`, token.STRING, "hello"},
		{"单引号字符串", `'world'`, token.STRING, "world"},
		{"转义换行", `"hello\nworld"`, token.STRING, "hello\nworld"},
		{"转义制表符", `"hello\tworld"`, token.STRING, "hello\tworld"},
		{"转义双引号", `"say \"hi\""`, token.STRING, `say "hi"`},
		{"转义单引号", `'it\'s'`, token.STRING, "it's"},
		{"转义反斜杠", `"path\\to\\file"`, token.STRING, `path\to\file`},
		{"Unicode转义", `"中文"`, token.STRING, "中文"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestOperators 测试运算符扫描
func TestOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"加号", "+", token.PLUS, "+"},
		{"减号", "-", token.MINUS, "-"},
		{"星号", "*", token.STAR, "*"},
		{"斜杠", "/", token.SLASH, "/"},
		{"百分号", "%", token.PERCENT, "%"},
		{"与号", "&", token.AMPERSAND, "&"},
		{"或号", "|", token.PIPE, "|"},
		{"异或", "^", token.CARET, "^"},
		{"波浪号", "~", token.TILDE, "~"},
		{"左移", "<<", token.SHIFTLEFT, "<<"},
		{"右移", ">>", token.SHIFTRIGHT, ">>"},
		{"等于", "==", token.EQ, "=="},
		{"不等于", "!=", token.NEQ, "!="},
		{"小于", "<", token.LT, "<"},
		{"大于", ">", token.GT, ">"},
		{"小于等于", "<=", token.LTE, "<="},
		{"大于等于", ">=", token.GTE, ">="},
		{"逻辑与", "&&", token.AND, "&&"},
		{"逻辑或", "||", token.OR, "||"},
		{"逻辑非", "!", token.NOT, "!"},
		{"赋值", "=", token.ASSIGN, "="},
		{"加赋值", "+=", token.PLUS_ASSIGN, "+="},
		{"减赋值", "-=", token.MINUS_ASSIGN, "-="},
		{"乘赋值", "*=", token.STAR_ASSIGN, "*="},
		{"除赋值", "/=", token.SLASH_ASSIGN, "/="},
		{"字符串连接", "..", token.CONCAT, ".."},
		{"问号", "?", token.QUESTION, "?"},
		{"冒号", ":", token.COLON, ":"},
		{"箭头", "->", token.ARROW, "->"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestDelimiters 测试分隔符扫描
func TestDelimiters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected token.TokenType
		literal  string
	}{
		{"分号", ";", token.SEMICOLON, ";"},
		{"逗号", ",", token.COMMA, ","},
		{"左括号", "(", token.LPAREN, "("},
		{"右括号", ")", token.RPAREN, ")"},
		{"左大括号", "{", token.LBRACE, "{"},
		{"右大括号", "}", token.RBRACE, "}"},
		{"左方括号", "[", token.LBRACKET, "["},
		{"右方括号", "]", token.RBRACKET, "]"},
		{"点号", ".", token.DOT, "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tok := l.NextToken()
			if tok.Type != tt.expected {
				t.Errorf("期望 %v，得到 %v", tt.expected, tok.Type)
			}
			if tok.Literal != tt.literal {
				t.Errorf("字面值期望 %q，得到 %q", tt.literal, tok.Literal)
			}
		})
	}
}

// TestComments 测试注释扫描
func TestComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.TokenType
	}{
		{"单行注释", "// 这是注释", []token.TokenType{token.COMMENT, token.EOF}},
		{"单行注释后代码", "// 注释\n$var", []token.TokenType{token.COMMENT, token.NEWLINE, token.IDENTIFIER, token.EOF}},
		{"多行注释", "/* 多行\n注释 */", []token.TokenType{token.BLOCK_COMMENT, token.EOF}},
		{"多行注释后代码", "/* 注释 */$var", []token.TokenType{token.BLOCK_COMMENT, token.IDENTIFIER, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.jpl")
			tokens := l.ScanAll()
			if len(tokens) != len(tt.expected) {
				t.Errorf("期望 %d 个 Token，得到 %d 个", len(tt.expected), len(tokens))
				return
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("Token[%d]: 期望 %v，得到 %v", i, tt.expected[i], tok.Type)
				}
			}
		})
	}
}

// TestComplexExpression 测试复杂表达式
func TestComplexExpression(t *testing.T) {
	input := `$sum = $a + $b * 2;`
	expected := []token.TokenType{
		token.IDENTIFIER,
		token.ASSIGN,
		token.IDENTIFIER,
		token.PLUS,
		token.IDENTIFIER,
		token.STAR,
		token.INTEGER,
		token.SEMICOLON,
		token.EOF,
	}

	l := NewLexer(input, "test.jpl")
	tokens := l.ScanAll()

	if len(tokens) != len(expected) {
		t.Fatalf("期望 %d 个 Token，得到 %d 个", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token[%d]: 期望 %v，得到 %v (字面值: %q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

// TestFunctionDeclaration 测试函数声明
func TestFunctionDeclaration(t *testing.T) {
	input := `fn add($a, $b) { return $a + $b; }`
	expected := []token.TokenType{
		token.FUNCTION,
		token.IDENTIFIER,
		token.LPAREN,
		token.IDENTIFIER,
		token.COMMA,
		token.IDENTIFIER,
		token.RPAREN,
		token.LBRACE,
		token.RETURN,
		token.IDENTIFIER,
		token.PLUS,
		token.IDENTIFIER,
		token.SEMICOLON,
		token.RBRACE,
		token.EOF,
	}

	l := NewLexer(input, "test.jpl")
	tokens := l.ScanAll()

	if len(tokens) != len(expected) {
		t.Fatalf("期望 %d 个 Token，得到 %d 个", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token[%d]: 期望 %v，得到 %v (字面值: %q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

// TestStringConcatenation 测试字符串连接
func TestStringConcatenation(t *testing.T) {
	input := `"Hello" .. " " .. "World"`
	expected := []token.TokenType{
		token.STRING,
		token.CONCAT,
		token.STRING,
		token.CONCAT,
		token.STRING,
		token.EOF,
	}

	l := NewLexer(input, "test.jpl")
	tokens := l.ScanAll()

	if len(tokens) != len(expected) {
		t.Fatalf("期望 %d 个 Token，得到 %d 个", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token[%d]: 期望 %v，得到 %v", i, expected[i], tok.Type)
		}
	}
}

// TestPosition 测试位置信息
func TestPosition(t *testing.T) {
	input := "line1\nline2\nline3"
	l := NewLexer(input, "test.jpl")

	// 第一行
	tok := l.NextToken()
	if tok.Pos.Line != 1 || tok.Pos.Column != 1 {
		t.Errorf("位置错误: 期望 行1列1，得到 行%d列%d", tok.Pos.Line, tok.Pos.Column)
	}

	// 换行符
	tok = l.NextToken()
	if tok.Pos.Line != 1 || tok.Pos.Column != 6 {
		t.Errorf("位置错误: 期望 行1列6，得到 行%d列%d", tok.Pos.Line, tok.Pos.Column)
	}

	// 第二行
	tok = l.NextToken()
	if tok.Pos.Line != 2 || tok.Pos.Column != 1 {
		t.Errorf("位置错误: 期望 行2列1，得到 行%d列%d", tok.Pos.Line, tok.Pos.Column)
	}
}

// TestIllegalCharacterNotInfiniteLoop 测试非法字符不会导致无限循环
// 修复前：遇到非法字符（如 #）时 lexer 不移动位置，导致无限循环
// 修复后：非法字符应该返回 ILLEGAL token 并正确前进到 EOF
func TestIllegalCharacterNotInfiniteLoop(t *testing.T) {
	input := "#"
	l := NewLexer(input, "test.jpl")

	// 第一个 token 应该是 ILLEGAL
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Errorf("期望第一个 token 是 ILLEGAL，得到 %v", tok.Type)
	}
	if tok.Literal != "#" {
		t.Errorf("期望 ILLEGAL token 的 literal 是 '#'，得到 %q", tok.Literal)
	}

	// 第二个 token 应该是 EOF（不应该无限循环）
	tok = l.NextToken()
	if tok.Type != token.EOF {
		t.Errorf("期望第二个 token 是 EOF，得到 %v（可能存在无限循环）", tok.Type)
	}
}
