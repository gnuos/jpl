package token

import "testing"

// TestKeyword 测试关键字映射
func TestKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TokenType
	}{
		// 控制流关键字
		{"if关键字", "if", IF},
		{"else关键字", "else", ELSE},
		{"while关键字", "while", WHILE},
		{"for关键字", "for", FOR},
		{"foreach关键字", "foreach", FOREACH},
		{"function关键字", "function", FUNCTION},
		{"fn别名", "fn", FUNCTION},
		{"return关键字", "return", RETURN},
		{"break关键字", "break", BREAK},
		{"continue关键字", "continue", CONTINUE},

		// 模块关键字
		{"import关键字", "import", IMPORT},

		// 作用域关键字
		{"static关键字", "static", STATIC},
		{"global关键字", "global", GLOBAL},
		{"const关键字", "const", CONST},

		// 异常关键字
		{"try关键字", "try", TRY},
		{"catch关键字", "catch", CATCH},
		{"throw关键字", "throw", THROW},

		// 匹配关键字
		{"match关键字", "match", MATCH},
		{"case关键字", "case", CASE},
		{"default关键字", "default", DEFAULT},

		// null 字面量
		{"null小写", "null", KW_NULL},
		{"NULL大写", "NULL", KW_NULL},

		// true 字面量
		{"true小写", "true", KW_TRUE},
		{"TRUE大写", "TRUE", KW_TRUE},
		{"True首字母大写", "True", KW_TRUE},

		// false 字面量
		{"false小写", "false", KW_FALSE},
		{"FALSE大写", "FALSE", KW_FALSE},
		{"False首字母大写", "False", KW_FALSE},

		// 非关键字
		{"普通标识符", "myVar", IDENTIFIER},
		{"随机字符串", "hello", IDENTIFIER},
		{"空字符串", "", IDENTIFIER},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Keyword(tt.input)
			if result != tt.expected {
				t.Errorf("Keyword(%q) = %v, 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTokenTypeString 测试 TokenType 的 String 方法
func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		name     string
		typ      TokenType
		expected string
	}{
		{"ILLEGAL", ILLEGAL, "ILLEGAL"},
		{"EOF", EOF, "EOF"},
		{"NEWLINE", NEWLINE, "NEWLINE"},
		{"IDENTIFIER", IDENTIFIER, "IDENTIFIER"},
		{"INTEGER", INTEGER, "INTEGER"},
		{"FLOAT", FLOAT, "FLOAT"},
		{"STRING", STRING, "STRING"},
		{"PLUS", PLUS, "PLUS"},
		{"MINUS", MINUS, "MINUS"},
		{"STAR", STAR, "STAR"},
		{"SLASH", SLASH, "SLASH"},
		{"PERCENT", PERCENT, "PERCENT"},
		{"AMPERSAND", AMPERSAND, "AMPERSAND"},
		{"PIPE", PIPE, "PIPE"},
		{"CARET", CARET, "CARET"},
		{"TILDE", TILDE, "TILDE"},
		{"SHIFTLEFT", SHIFTLEFT, "SHIFTLEFT"},
		{"SHIFTRIGHT", SHIFTRIGHT, "SHIFTRIGHT"},
		{"EQ", EQ, "EQ"},
		{"NEQ", NEQ, "NEQ"},
		{"LT", LT, "LT"},
		{"GT", GT, "GT"},
		{"LTE", LTE, "LTE"},
		{"GTE", GTE, "GTE"},
		{"AND", AND, "AND"},
		{"OR", OR, "OR"},
		{"NOT", NOT, "NOT"},
		{"ASSIGN", ASSIGN, "ASSIGN"},
		{"PLUS_ASSIGN", PLUS_ASSIGN, "PLUS_ASSIGN"},
		{"MINUS_ASSIGN", MINUS_ASSIGN, "MINUS_ASSIGN"},
		{"STAR_ASSIGN", STAR_ASSIGN, "STAR_ASSIGN"},
		{"SLASH_ASSIGN", SLASH_ASSIGN, "SLASH_ASSIGN"},
		{"CONCAT", CONCAT, "CONCAT"},
		{"QUESTION", QUESTION, "QUESTION"},
		{"COLON", COLON, "COLON"},
		{"ARROW", ARROW, "ARROW"},
		{"SEMICOLON", SEMICOLON, "SEMICOLON"},
		{"COMMA", COMMA, "COMMA"},
		{"LPAREN", LPAREN, "LPAREN"},
		{"RPAREN", RPAREN, "RPAREN"},
		{"LBRACE", LBRACE, "LBRACE"},
		{"RBRACE", RBRACE, "RBRACE"},
		{"LBRACKET", LBRACKET, "LBRACKET"},
		{"RBRACKET", RBRACKET, "RBRACKET"},
		{"DOT", DOT, "DOT"},
		{"IF", IF, "IF"},
		{"ELSE", ELSE, "ELSE"},
		{"WHILE", WHILE, "WHILE"},
		{"FOR", FOR, "FOR"},
		{"FOREACH", FOREACH, "FOREACH"},
		{"FUNCTION", FUNCTION, "FUNCTION"},
		{"RETURN", RETURN, "RETURN"},
		{"BREAK", BREAK, "BREAK"},
		{"CONTINUE", CONTINUE, "CONTINUE"},
		{"IMPORT", IMPORT, "IMPORT"},
		{"STATIC", STATIC, "STATIC"},
		{"GLOBAL", GLOBAL, "GLOBAL"},
		{"TRY", TRY, "TRY"},
		{"CATCH", CATCH, "CATCH"},
		{"THROW", THROW, "THROW"},
		{"MATCH", MATCH, "MATCH"},
		{"CASE", CASE, "CASE"},
		{"DEFAULT", DEFAULT, "DEFAULT"},
		{"CONST", CONST, "CONST"},
		{"KW_NULL", KW_NULL, "KW_NULL"},
		{"KW_TRUE", KW_TRUE, "KW_TRUE"},
		{"KW_FALSE", KW_FALSE, "KW_FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.typ.String()
			if result != tt.expected {
				t.Errorf("TokenType(%d).String() = %q, 期望 %q", tt.typ, result, tt.expected)
			}
		})
	}
}

// TestTokenStructure 测试 Token 结构体
func TestTokenStructure(t *testing.T) {
	pos := Position{
		Filename: "test.jpl",
		Line:     1,
		Column:   5,
		Offset:   4,
	}

	tok := Token{
		Type:    INTEGER,
		Literal: "42",
		Pos:     pos,
	}

	if tok.Type != INTEGER {
		t.Errorf("Token.Type = %v, 期望 INTEGER", tok.Type)
	}
	if tok.Literal != "42" {
		t.Errorf("Token.Literal = %q, 期望 \"42\"", tok.Literal)
	}
	if tok.Pos.Filename != "test.jpl" {
		t.Errorf("Token.Pos.Filename = %q, 期望 \"test.jpl\"", tok.Pos.Filename)
	}
	if tok.Pos.Line != 1 {
		t.Errorf("Token.Pos.Line = %d, 期望 1", tok.Pos.Line)
	}
	if tok.Pos.Column != 5 {
		t.Errorf("Token.Pos.Column = %d, 期望 5", tok.Pos.Column)
	}
	if tok.Pos.Offset != 4 {
		t.Errorf("Token.Pos.Offset = %d, 期望 4", tok.Pos.Offset)
	}
}

// TestFunctionAlias 测试 function 和 fn 别名
func TestFunctionAlias(t *testing.T) {
	functionTok := Keyword("function")
	fnTok := Keyword("fn")

	if functionTok != FUNCTION {
		t.Errorf("Keyword(\"function\") = %v, 期望 FUNCTION", functionTok)
	}
	if fnTok != FUNCTION {
		t.Errorf("Keyword(\"fn\") = %v, 期望 FUNCTION", fnTok)
	}
	if functionTok != fnTok {
		t.Errorf("function 和 fn 应该映射到相同的 TokenType")
	}
}

// TestCaseInsensitiveKeywords 测试大小写变体
func TestCaseInsensitiveKeywords(t *testing.T) {
	// null 变体
	if Keyword("null") != KW_NULL {
		t.Errorf("Keyword(\"null\") 应该返回 KW_NULL")
	}
	if Keyword("NULL") != KW_NULL {
		t.Errorf("Keyword(\"NULL\") 应该返回 KW_NULL")
	}

	// true 变体
	if Keyword("true") != KW_TRUE {
		t.Errorf("Keyword(\"true\") 应该返回 KW_TRUE")
	}
	if Keyword("TRUE") != KW_TRUE {
		t.Errorf("Keyword(\"TRUE\") 应该返回 KW_TRUE")
	}
	if Keyword("True") != KW_TRUE {
		t.Errorf("Keyword(\"True\") 应该返回 KW_TRUE")
	}

	// false 变体
	if Keyword("false") != KW_FALSE {
		t.Errorf("Keyword(\"false\") 应该返回 KW_FALSE")
	}
	if Keyword("FALSE") != KW_FALSE {
		t.Errorf("Keyword(\"FALSE\") 应该返回 KW_FALSE")
	}
	if Keyword("False") != KW_FALSE {
		t.Errorf("Keyword(\"False\") 应该返回 KW_FALSE")
	}
}

// TestNonKeywordReturnsIdentifier 测试非关键字返回 IDENTIFIER
func TestNonKeywordReturnsIdentifier(t *testing.T) {
	nonKeywords := []string{
		"myVar",
		"hello",
		"world",
		"test123",
		"_private",
		"camelCase",
		"PascalCase",
		"snake_case",
	}

	for _, name := range nonKeywords {
		if Keyword(name) != IDENTIFIER {
			t.Errorf("Keyword(%q) 应该返回 IDENTIFIER", name)
		}
	}
}

// TestKeywords 测试关键字列表导出
func TestKeywords(t *testing.T) {
	names := Keywords()

	if len(names) == 0 {
		t.Fatal("Keywords() 返回空列表")
	}

	// 检查关键关键字是否存在
	required := map[string]bool{
		"if": true, "else": true, "while": true, "for": true,
		"foreach": true, "function": true, "fn": true,
		"return": true, "break": true, "continue": true,
		"import": true, "const": true, "true": true, "false": true, "null": true,
	}

	nameSet := make(map[string]bool, len(names))
	for _, name := range names {
		nameSet[name] = true
	}

	for kw := range required {
		if !nameSet[kw] {
			t.Errorf("Keywords() 缺少关键字: %q", kw)
		}
	}
}
