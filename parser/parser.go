package parser

import (
	"fmt"
	"strconv"

	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/token"
)

// ============================================================================
// 特例函数（支持无括号调用）
// ============================================================================

// specialFuncs 特例函数表 — 语句级可省略括号调用
// 仅包含作为语句使用的 I/O 类函数，返回值有意义的函数（len/is_*/typeof）不在此表
var specialFuncs = map[string]bool{
	"echo":    true,
	"print":   true,
	"println": true,
	"puts":    true,
	"pp":      true,
	"log":     true,
	"format":  true,
	"assert":  true,
	"define":  true,
	"exit":    true,
	"die":     true,
}

// isValueToken 判断 Token 是否为值类型（标识符/字面量/左方括号/左花括号/前缀运算符）
// 注意: LPAREN 不包含在内，因为 foo(x) 是合法的普通函数调用
func isValueToken(t token.TokenType) bool {
	switch t {
	case token.IDENTIFIER, token.SPECIAL_VAR, token.UNDERSCORE,
		token.STRING, token.STRING_FRAG, token.INTEGER, token.FLOAT,
		token.BIGINT, token.BIGDECIMAL,
		token.KW_TRUE, token.KW_FALSE, token.KW_NULL,
		token.LBRACKET, token.LBRACE,
		token.TILDE, token.REGEX:
		return true
	}
	return false
}

// ============================================================================
// 优先级定义
// ============================================================================

// Precedence 运算符优先级
type Precedence int

const (
	_ Precedence = iota
	LOWEST
	TERNARY_PREC     // ?: （最低优先级）
	ASSIGN_PREC      // =
	ARROW_PREC       // ->
	PIPE_FWD_PREC    // |> 管道前向运算
	PIPE_BWD_PREC    // <| 管道反向运算
	OR_PREC          // ||
	AND_PREC         // &&
	EQUALITY_PREC    // ==
	COMPARISON_PREC  // < > <= >=
	CONCAT_PREC      // ..
	RANGE_PREC       // ... ..= 范围运算符（优先级低于字符串连接）
	BITWISE_OR_PREC  // |
	BITWISE_XOR_PREC // ^
	BITWISE_AND_PREC // &
	SHIFT_PREC       // << >>
	SUM_PREC         // + -
	PRODUCT_PREC     // * / %
	PREFIX_PREC      // -x !x ~x
	CALL_PREC        // function(x)
	INDEX_PREC       // array[index]
	MEMBER_PREC      // object.member
)

// precedences Token 类型到优先级的映射
var precedences = map[token.TokenType]Precedence{
	token.OR:              OR_PREC,
	token.AND:             AND_PREC,
	token.EQ:              EQUALITY_PREC,
	token.NEQ:             EQUALITY_PREC,
	token.LT:              COMPARISON_PREC,
	token.GT:              COMPARISON_PREC,
	token.LTE:             COMPARISON_PREC,
	token.GTE:             COMPARISON_PREC,
	token.CONCAT:          CONCAT_PREC,
	token.ELLIPSIS:        RANGE_PREC,
	token.RANGE_INCLUSIVE: RANGE_PREC,
	token.PIPE:            BITWISE_OR_PREC,
	token.CARET:           BITWISE_XOR_PREC,
	token.AMPERSAND:       BITWISE_AND_PREC,
	token.SHIFTLEFT:       SHIFT_PREC,
	token.SHIFTRIGHT:      SHIFT_PREC,
	token.PLUS:            SUM_PREC,
	token.MINUS:           SUM_PREC,
	token.STAR:            PRODUCT_PREC,
	token.SLASH:           PRODUCT_PREC,
	token.PERCENT:         PRODUCT_PREC,
	token.LPAREN:          CALL_PREC,
	token.STRING:          CALL_PREC, // 支持无括号函数调用：println "hello"
	token.STRING_FRAG:     CALL_PREC, // 支持无括号函数调用：println "hello #{$name}"
	token.LBRACKET:        INDEX_PREC,
	token.DOT:             MEMBER_PREC,
	token.ASSIGN:          ASSIGN_PREC,
	token.PLUS_ASSIGN:     ASSIGN_PREC,
	token.MINUS_ASSIGN:    ASSIGN_PREC,
	token.STAR_ASSIGN:     ASSIGN_PREC,
	token.SLASH_ASSIGN:    ASSIGN_PREC,
	token.QUESTION:        TERNARY_PREC,
	token.ARROW:           ARROW_PREC,
	token.PIPE_FWD:        PIPE_FWD_PREC,
	token.PIPE_BWD:        PIPE_BWD_PREC,
	token.MATCH_EQ:        EQUALITY_PREC, // =~ 与 == != 同优先级
}

// ============================================================================
// Parser 结构体
// ============================================================================

// Parser 语法分析器
type Parser struct {
	lexer  *lexer.Lexer // 词法分析器
	cur    token.Token  // 当前 Token
	peek   token.Token  // 下一个 Token
	errors []string     // 错误列表

	// 解析函数映射
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// prefixParseFn 前缀解析函数
type prefixParseFn func() Expression

// infixParseFn 中缀解析函数
type infixParseFn func(left Expression) Expression

// NewParser 创建新的 Parser
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer: l,
	}

	// 注册前缀解析函数
	p.prefixParseFns = map[token.TokenType]prefixParseFn{
		token.IDENTIFIER:    p.parseIdentifier,
		token.SPECIAL_VAR:   p.parseIdentifier,
		token.UNDERSCORE:    p.parseIdentifier,
		token.INSTANCE_VAR:  p.parseInstanceVar,
		token.INTEGER:       p.parseNumberLiteral,
		token.FLOAT:         p.parseNumberLiteral,
		token.BIGINT:        p.parseNumberLiteral,
		token.BIGDECIMAL:    p.parseNumberLiteral,
		token.STRING:        p.parseStringLiteral,
		token.STRING_FRAG:   p.parseStringLiteral,
		token.TRIPLE_SINGLE: p.parseStringLiteral,
		token.TRIPLE_DOUBLE: p.parseStringLiteral,
		token.KW_TRUE:       p.parseBoolLiteral,
		token.KW_FALSE:      p.parseBoolLiteral,
		token.KW_NULL:       p.parseNullLiteral,
		// 类型关键字（用于类型转换语法）
		token.KW_INT:    p.parseTypeCast,
		token.KW_FLOAT:  p.parseTypeCast,
		token.KW_STRING: p.parseTypeCast,
		token.KW_BOOL:   p.parseTypeCast,
		token.LPAREN:    p.parseGroupedExpression,
		token.LBRACKET:  p.parseArrayLiteral,
		token.LBRACE:    p.parseObjectLiteral,
		token.MINUS:     p.parsePrefixExpression,
		token.NOT:       p.parsePrefixExpression,
		token.TILDE:     p.parsePrefixExpression,
		token.FUNCTION:  p.parseLambdaExpression,
		token.IF:        p.parseIfExpression,
		token.MATCH:     p.parseMatchExpression,
		token.REGEX:     p.parseRegexLiteral,
		// 字符串插值标记（不应在字符串外出现）
		token.INTERP_START: p.parseInterpStartError,
		token.INTERP_END:   p.parseInterpEndError,
	}

	// 注册中缀解析函数
	p.infixParseFns = map[token.TokenType]infixParseFn{
		token.PLUS:            p.parseInfixExpression,
		token.MINUS:           p.parseInfixExpression,
		token.STAR:            p.parseInfixExpression,
		token.SLASH:           p.parseInfixExpression,
		token.PERCENT:         p.parseInfixExpression,
		token.EQ:              p.parseInfixExpression,
		token.NEQ:             p.parseInfixExpression,
		token.LT:              p.parseInfixExpression,
		token.GT:              p.parseInfixExpression,
		token.LTE:             p.parseInfixExpression,
		token.GTE:             p.parseInfixExpression,
		token.AND:             p.parseInfixExpression,
		token.OR:              p.parseInfixExpression,
		token.AMPERSAND:       p.parseInfixExpression,
		token.PIPE:            p.parseInfixExpression,
		token.CARET:           p.parseInfixExpression,
		token.SHIFTLEFT:       p.parseInfixExpression,
		token.SHIFTRIGHT:      p.parseInfixExpression,
		token.CONCAT:          p.parseInfixExpression,
		token.ELLIPSIS:        p.parseInfixExpression,
		token.RANGE_INCLUSIVE: p.parseInfixExpression,
		token.LPAREN:          p.parseCallExpression,
		token.STRING:          p.parseImplicitCallExpression, // 无括号函数调用
		token.STRING_FRAG:     p.parseImplicitCallExpression, // 无括号函数调用（插值字符串）
		token.LBRACKET:        p.parseIndexExpression,
		token.DOT:             p.parseMemberExpression,
		token.ASSIGN:          p.parseAssignExpression,
		token.PLUS_ASSIGN:     p.parseAssignExpression,
		token.MINUS_ASSIGN:    p.parseAssignExpression,
		token.STAR_ASSIGN:     p.parseAssignExpression,
		token.SLASH_ASSIGN:    p.parseAssignExpression,
		token.QUESTION:        p.parseTernaryExpression,
		token.ARROW:           p.parseArrowExpression,
		token.PIPE_FWD:        p.parsePipeExpression,
		token.PIPE_BWD:        p.parsePipeExpression,
		token.MATCH_EQ:        p.parseInfixExpression, // =~ 正则匹配运算符
	}

	// 读取两个 Token
	p.nextToken()
	p.nextToken()

	return p
}

// ============================================================================
// 基础方法
// ============================================================================

// nextToken 移动到下一个 Token（跳过注释）
func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
	// 跳过注释 token（编译时不需要，格式化时另行处理）
	for p.peek.Type == token.COMMENT || p.peek.Type == token.BLOCK_COMMENT {
		p.peek = p.lexer.NextToken()
	}
}

// skipSeparators 跳过换行符、注释和分号（空语句）
func (p *Parser) skipSeparators() {
	for p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.COMMENT) || p.curTokenIs(token.BLOCK_COMMENT) || p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
}

// peekIsNewline 检查下一个 Token 是否为换行符或注释
func (p *Parser) peekIsNewline() bool {
	return p.peek.Type == token.NEWLINE || p.peek.Type == token.COMMENT || p.peek.Type == token.BLOCK_COMMENT
}

// curTokenIs 检查当前 Token 类型
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.cur.Type == t
}

// peekTokenIs 检查下一个 Token 类型
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peek.Type == t
}

// expectPeek 期望下一个 Token 为指定类型，如果是则移动
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekPrecedence 获取下一个 Token 的优先级
func (p *Parser) peekPrecedence() Precedence {
	if p, ok := precedences[p.peek.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence 获取当前 Token 的优先级
func (p *Parser) curPrecedence() Precedence {
	if p, ok := precedences[p.cur.Type]; ok {
		return p
	}
	return LOWEST
}

// Errors 获取错误列表
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError 添加 peek 错误
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("期望下一个 Token 为 %s，实际为 %s (字面值: %q)",
		t, p.peek.Type, p.peek.Literal)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 添加无前缀解析函数错误
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	if t == token.ILLEGAL {
		// ILLEGAL token 包含词法分析器的错误信息
		p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.cur.Pos.Line, p.cur.Literal))
	} else {
		msg := fmt.Sprintf("line %d: unexpected token %s (%q)", p.cur.Pos.Line, t, p.cur.Literal)
		p.errors = append(p.errors, msg)
	}
}

// ============================================================================
// 解析入口
// ============================================================================

// Parse 解析整个程序
func (p *Parser) Parse() *Program {
	program := &Program{}

	for !p.curTokenIs(token.EOF) {
		// 跳过换行符、注释和空语句（分号）
		p.skipSeparators()
		if p.curTokenIs(token.EOF) {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// ============================================================================
// 语句解析
// ============================================================================

// parseStatement 解析语句
func (p *Parser) parseStatement() Statement {
	switch p.cur.Type {
	case token.CONST:
		return p.parseConstDecl()
	case token.FUNCTION:
		if p.peekTokenIs(token.IDENTIFIER) {
			return p.parseFuncDecl()
		}
		return p.parseExpressionStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.FOREACH:
		return p.parseForeachStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.FROM:
		return p.parseFromImportStatement()
	case token.INCLUDE, token.INCLUDE_ONCE:
		return p.parseIncludeStatement()
	case token.TRY:
		return p.parseTryCatchStatement()
	case token.THROW:
		return p.parseThrowStatement()
	case token.GLOBAL:
		return p.parseGlobalDecl()
	case token.STATIC:
		return p.parseStaticDecl()
	case token.MATCH:
		return p.parseMatchStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.ILLEGAL:
		// ILLEGAL token 包含词法错误信息，记录并跳过
		p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.cur.Pos.Line, p.cur.Literal))
		p.nextToken() // 跳过非法 token
		return nil
	case token.IDENTIFIER, token.SPECIAL_VAR:
		if p.peekTokenIs(token.ASSIGN) ||
			p.peekTokenIs(token.PLUS_ASSIGN) ||
			p.peekTokenIs(token.MINUS_ASSIGN) ||
			p.peekTokenIs(token.STAR_ASSIGN) ||
			p.peekTokenIs(token.SLASH_ASSIGN) ||
			p.peekTokenIs(token.AND_ASSIGN) ||
			p.peekTokenIs(token.OR_ASSIGN) ||
			p.peekTokenIs(token.XOR_ASSIGN) ||
			p.peekTokenIs(token.SHL_ASSIGN) ||
			p.peekTokenIs(token.SHR_ASSIGN) ||
			p.peekTokenIs(token.CONCAT_ASSIGN) {
			return p.parseAssignStatement()
		}
		// 特例函数无括号调用
		if specialFuncs[p.cur.Literal] && isValueToken(p.peek.Type) {
			return p.parseSpecialCallStatement()
		}
		// 非特例函数后跟值类型 → 语法错误
		// 但 $var[...] 是合法的索引访问，排除 LBRACKET
		if !specialFuncs[p.cur.Literal] && isValueToken(p.peek.Type) && p.peek.Type != token.LBRACKET {
			p.errors = append(p.errors, fmt.Sprintf(
				"syntax error: unexpected %s after identifier, use parentheses for function calls (line %d)",
				p.peek.Type, p.peek.Pos.Line))
			return p.parseExpressionStatement()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseSpecialCallStatement 解析特例函数无括号调用
// 格式: print "hello" 或 format "%d", x, y
func (p *Parser) parseSpecialCallStatement() Statement {
	funcName := &Identifier{Token: p.cur, Value: p.cur.Literal}

	// 前进到第一个参数
	p.nextToken()

	// 解析逗号分隔的参数列表
	var args []Expression
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 消费逗号
		p.nextToken() // 前进到下一个表达式
		args = append(args, p.parseExpression(LOWEST))
	}

	// 消费可选的分号
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	callExpr := &CallExpr{
		Token:     funcName.Token,
		Function:  funcName,
		Arguments: args,
	}

	return &ExprStmt{Token: funcName.Token, Expression: callExpr}
}

// parseConstDecl 解析常量声明
func (p *Parser) parseConstDecl() Statement {
	stmt := &ConstDecl{Token: p.cur}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.cur, Value: p.cur.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseGlobalDecl 解析 global 声明
// 格式: global $varname; 或 global $a, $b, $c;
func (p *Parser) parseGlobalDecl() Statement {
	stmt := &GlobalDecl{Token: p.cur}

	// 解析变量名列表（逗号分隔）
	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}
	stmt.Names = append(stmt.Names, &Identifier{Token: p.cur, Value: p.cur.Literal})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 消费逗号
		if !p.expectPeek(token.IDENTIFIER) {
			return nil
		}
		stmt.Names = append(stmt.Names, &Identifier{Token: p.cur, Value: p.cur.Literal})
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseStaticDecl 解析 static 声明
// 格式: static $varname = initialValue;
func (p *Parser) parseStaticDecl() Statement {
	stmt := &StaticDecl{Token: p.cur}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.cur, Value: p.cur.Literal}

	// 可选的初始值
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // 消费 =
		p.nextToken() // 前进到表达式
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseFuncDecl 解析函数声明
func (p *Parser) parseFuncDecl() Statement {
	stmt := &FuncDecl{Token: p.cur}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	stmt.Name = &Identifier{Token: p.cur, Value: p.cur.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionParameters 解析函数参数列表
func (p *Parser) parseFunctionParameters() []*Identifier {
	var params []*Identifier

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	// 处理第一个参数
	params = append(params, p.parseIdentifier().(*Identifier))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		params = append(params, p.parseIdentifier().(*Identifier))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// parseReturnStatement 解析 return 语句
func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStmt{Token: p.cur}

	// 检查是否有返回值
	if !p.peekTokenIs(token.SEMICOLON) && !p.peekIsNewline() && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseBreakStatement 解析 break 语句
func (p *Parser) parseBreakStatement() Statement {
	stmt := &BreakStmt{Token: p.cur}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseContinueStatement 解析 continue 语句
func (p *Parser) parseContinueStatement() Statement {
	stmt := &ContinueStmt{Token: p.cur}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseIfStatement 解析 if 语句
func (p *Parser) parseIfStatement() Statement {
	stmt := &IfStmt{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Cond = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// 处理 else 分支
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if p.peekTokenIs(token.IF) {
			p.nextToken()
			stmt.Else = p.parseIfStatement()
		} else if p.peekTokenIs(token.LBRACE) {
			p.nextToken()
			stmt.Else = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseWhileStatement 解析 while 语句
func (p *Parser) parseWhileStatement() Statement {
	stmt := &WhileStmt{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Cond = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForStatement 解析 for 语句
func (p *Parser) parseForStatement() Statement {
	stmt := &ForStmt{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// 解析初始化语句（直接解析，不用 parseStatement 避免分号冲突）
	p.nextToken()
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Init = p.parseForInitOrPost()
		// parseForInitOrPost 不消耗尾部分号
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	// 解析条件表达式
	p.nextToken()
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Cond = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	// 解析后置表达式
	p.nextToken()
	if !p.curTokenIs(token.RPAREN) {
		stmt.Post = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseForInitOrPost 解析 for 循环的 init/post 子句
// 与 parseStatement 不同，此函数不消耗尾部分号
func (p *Parser) parseForInitOrPost() Statement {
	switch p.cur.Type {
	case token.CONST:
		return p.parseConstDecl()
	case token.IDENTIFIER, token.SPECIAL_VAR:
		if p.peekTokenIs(token.ASSIGN) ||
			p.peekTokenIs(token.PLUS_ASSIGN) ||
			p.peekTokenIs(token.MINUS_ASSIGN) ||
			p.peekTokenIs(token.STAR_ASSIGN) ||
			p.peekTokenIs(token.SLASH_ASSIGN) ||
			p.peekTokenIs(token.AND_ASSIGN) ||
			p.peekTokenIs(token.OR_ASSIGN) ||
			p.peekTokenIs(token.XOR_ASSIGN) ||
			p.peekTokenIs(token.SHL_ASSIGN) ||
			p.peekTokenIs(token.SHR_ASSIGN) ||
			p.peekTokenIs(token.CONCAT_ASSIGN) {
			return p.parseAssignNoSemi()
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseAssignNoSemi 解析赋值语句（支持组合赋值），不消耗尾部分号
// 用于 for 循环的 init/post 子句
func (p *Parser) parseAssignNoSemi() Statement {
	name := &Identifier{Token: p.cur, Value: p.cur.Literal}

	// 消费赋值运算符
	opToken := p.peek
	p.nextToken()

	stmt := &VarDecl{
		Token: name.Token,
		Name:  name,
	}

	// 移动到值的位置
	p.nextToken()

	// 检查是否为组合赋值运算符
	if opToken.Type == token.PLUS_ASSIGN || opToken.Type == token.MINUS_ASSIGN ||
		opToken.Type == token.STAR_ASSIGN || opToken.Type == token.SLASH_ASSIGN ||
		opToken.Type == token.AND_ASSIGN || opToken.Type == token.OR_ASSIGN ||
		opToken.Type == token.XOR_ASSIGN || opToken.Type == token.SHL_ASSIGN ||
		opToken.Type == token.SHR_ASSIGN || opToken.Type == token.CONCAT_ASSIGN {
		// 组合赋值：转换为二元运算表达式
		var op string
		switch opToken.Type {
		case token.PLUS_ASSIGN:
			op = "+"
		case token.MINUS_ASSIGN:
			op = "-"
		case token.STAR_ASSIGN:
			op = "*"
		case token.SLASH_ASSIGN:
			op = "/"
		case token.AND_ASSIGN:
			op = "&"
		case token.OR_ASSIGN:
			op = "|"
		case token.XOR_ASSIGN:
			op = "^"
		case token.SHL_ASSIGN:
			op = "<<"
		case token.SHR_ASSIGN:
			op = ">>"
		case token.CONCAT_ASSIGN:
			op = ".."
		}

		// 解析右侧表达式
		right := p.parseExpression(LOWEST)

		// 创建二元运算：name op right
		stmt.Value = &BinaryExpr{
			Token:    opToken,
			Left:     name,
			Operator: op,
			Right:    right,
		}
	} else {
		// 普通赋值
		stmt.Value = p.parseExpression(LOWEST)
	}

	// 不消耗尾部分号
	return stmt
}

// parseForeachStatement 解析 foreach 循环语句
//
// 语法支持两种形式：
//  1. 只遍历值：foreach ($value in $array) { ... }
//  2. 遍历键值对：foreach ($key => $value in $array) { ... }
//
// 解析流程：
//   - 期望 '(' 开始
//   - 解析第一个变量（可能是值或键）
//   - 检查是否存在 "=>"（ROCKET 令牌），如果存在则第一个变量是键，继续解析值变量
//   - 期望 'in' 关键字
//   - 解析被遍历的数组或对象表达式
//   - 期望 ')' 结束循环头
//   - 期望 '{' 开始循环体
//   - 解析循环体语句块
//
// 注意：支持特殊变量（$var）和普通标识符作为循环变量
func (p *Parser) parseForeachStatement() Statement {
	stmt := &ForeachStmt{Token: p.cur}

	// 期望 '(' 开始 foreach 条件
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	// 检测语法类型：
	// - foreach ($value in $arr)     // 完整语法（只有值）
	// - foreach ($key => $value in $arr)  // 完整语法（键值对）
	// 先检查是否是键值语法（第一个 token 后是 ROCKET）
	if p.peekTokenIs(token.ROCKET) {
		// 键值语法：foreach ($key => $value in $arr)
		// 解析键变量
		if p.curTokenIs(token.IDENTIFIER) || p.curTokenIs(token.SPECIAL_VAR) {
			stmt.Key = &Identifier{Token: p.cur, Value: p.cur.Literal}
		}
		// 消费 "=>"
		p.nextToken()
		// 移动到值变量
		p.nextToken()
		// 解析值变量
		if p.curTokenIs(token.IDENTIFIER) || p.curTokenIs(token.SPECIAL_VAR) {
			stmt.Value = &Identifier{Token: p.cur, Value: p.cur.Literal}
		}
		// 期望 'in' 关键字
		if !p.expectPeek(token.IN) {
			return nil
		}
	} else if p.peekTokenIs(token.IN) {
		// 完整语法：foreach ($value in $arr)
		// 解析值变量
		if p.curTokenIs(token.IDENTIFIER) || p.curTokenIs(token.SPECIAL_VAR) {
			stmt.Value = &Identifier{Token: p.cur, Value: p.cur.Literal}
		}
		// 消费 'in' 关键字
		p.nextToken()
	} else {
		// 语法错误：必须有 'in' 关键字
		p.errors = append(p.errors, "foreach语法错误：期望 'in' 关键字")
		return nil
	}

	p.nextToken()

	// 解析被遍历的数组或对象表达式
	// 注意：不要跳过，因为 cur 已经在正确的位置（in 之后的 token）
	stmt.Array = p.parseExpression(LOWEST)

	// 期望 ')' 结束条件部分
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 期望 '{' 开始循环体
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析循环体语句块
	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseImportStatement 解析 import 语句
// 格式: import "source"; 或 import "source" as name;
func (p *Parser) parseImportStatement() Statement {
	stmt := &ImportStmt{Token: p.cur}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Source = p.cur.Literal

	// 检查 as 别名
	if p.peekTokenIs(token.AS) {
		p.nextToken() // consume 'as'
		if !p.expectPeek(token.IDENTIFIER) {
			return nil
		}
		stmt.Alias = &Identifier{Token: p.cur, Value: p.cur.Literal}
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseFromImportStatement 解析 from ... import ... 语句
// 格式: from "source" import name1, name2;
func (p *Parser) parseFromImportStatement() Statement {
	stmt := &ImportStmt{Token: p.cur}

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Source = p.cur.Literal

	if !p.expectPeek(token.IMPORT) {
		return nil
	}

	// 解析导入名称列表
	p.nextToken()

	// 第一个名称
	if p.curTokenIs(token.IDENTIFIER) {
		stmt.Names = append(stmt.Names, &Identifier{Token: p.cur, Value: p.cur.Literal})
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"from import 后期望标识符，实际为 %s (line %d)",
			p.cur.Type, p.cur.Pos.Line))
		return nil
	}

	// 后续逗号分隔的名称
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // 消费逗号
		p.nextToken() // 前进到下一个标识符
		if p.curTokenIs(token.IDENTIFIER) {
			stmt.Names = append(stmt.Names, &Identifier{Token: p.cur, Value: p.cur.Literal})
		} else {
			p.errors = append(p.errors, fmt.Sprintf(
				"from import 逗号后期望标识符，实际为 %s (line %d)",
				p.cur.Type, p.cur.Pos.Line))
			return nil
		}
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseIncludeStatement 解析 include/include_once 语句
// 格式: include "source"; 或 include_once "source";
func (p *Parser) parseIncludeStatement() Statement {
	stmt := &IncludeStmt{Token: p.cur}
	stmt.Once = p.cur.Type == token.INCLUDE_ONCE

	if !p.expectPeek(token.STRING) {
		return nil
	}

	stmt.Source = p.cur.Literal

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseTryCatchStatement 解析 try/catch 语句
func (p *Parser) parseTryCatchStatement() Statement {
	stmt := &TryCatchStmt{Token: p.cur}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.TryBody = p.parseBlockStatement()

	// 解析第一个 catch 分支
	if !p.expectPeek(token.CATCH) {
		return nil
	}

	stmt.CatchClauses = append(stmt.CatchClauses, p.parseCatchClause())

	// 循环解析后续的 catch 分支
	for p.peekTokenIs(token.CATCH) {
		p.nextToken() // consume CATCH
		stmt.CatchClauses = append(stmt.CatchClauses, p.parseCatchClause())
	}

	return stmt
}

// parseCatchClause 解析单个 catch 分支
func (p *Parser) parseCatchClause() *CatchClause {
	clause := &CatchClause{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	clause.CatchVar = &Identifier{Token: p.cur, Value: p.cur.Literal}

	// 可选的 when 条件
	if p.peekTokenIs(token.WHEN) {
		p.nextToken() // consume WHEN
		p.nextToken() // move to condition expression
		clause.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	clause.Body = p.parseBlockStatement()

	return clause
}

// parseThrowStatement 解析 throw 语句
func (p *Parser) parseThrowStatement() Statement {
	stmt := &ThrowStmt{Token: p.cur}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseMatchStatement 解析 match 语句
func (p *Parser) parseMatchStatement() Statement {
	stmt := &MatchStmt{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// 解析 case 列表
	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.COMMENT) || p.curTokenIs(token.BLOCK_COMMENT) {
			continue
		}

		if p.curTokenIs(token.CASE) {
			caseNode := p.parseMatchCase()
			if caseNode != nil {
				stmt.Cases = append(stmt.Cases, caseNode)
			}
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return stmt
}

// parseMatchCase 解析 match 的 case 分支
func (p *Parser) parseMatchCase() *MatchCase {
	caseNode := &MatchCase{Token: p.cur}

	p.nextToken()
	caseNode.Pattern = p.parsePattern()

	// 检查守卫条件
	if p.peekTokenIs(token.IF) {
		p.nextToken()
		p.nextToken()
		caseNode.Guard = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.COLON) {
		return nil
	}

	// 解析分支体
	p.nextToken()
	if p.curTokenIs(token.LBRACE) {
		caseNode.Body = p.parseBlockStatement()
	} else {
		// 表达式体
		caseNode.Body = &ExprStmt{
			Token:      p.cur,
			Expression: p.parseExpression(LOWEST),
		}
	}

	return caseNode
}

// parseBlockStatement 解析代码块
func (p *Parser) parseBlockStatement() *BlockStmt {
	block := &BlockStmt{Token: p.cur}
	block.Statements = []Statement{}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.COMMENT) || p.curTokenIs(token.BLOCK_COMMENT) {
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return block
}

// parseAssignStatement 解析赋值语句（支持组合赋值运算符）
// 处理：$x = expr, $x += expr, $x -= expr, $x *= expr, $x /= expr
func (p *Parser) parseAssignStatement() Statement {
	name := &Identifier{Token: p.cur, Value: p.cur.Literal}

	// 消费赋值运算符（=, +=, -=, *=, /=）
	opToken := p.peek
	p.nextToken()

	stmt := &VarDecl{
		Token: name.Token,
		Name:  name,
	}

	// 移动到值的位置
	p.nextToken()

	// 检查是否为组合赋值运算符
	if opToken.Type == token.PLUS_ASSIGN || opToken.Type == token.MINUS_ASSIGN ||
		opToken.Type == token.STAR_ASSIGN || opToken.Type == token.SLASH_ASSIGN ||
		opToken.Type == token.AND_ASSIGN || opToken.Type == token.OR_ASSIGN ||
		opToken.Type == token.XOR_ASSIGN || opToken.Type == token.SHL_ASSIGN ||
		opToken.Type == token.SHR_ASSIGN || opToken.Type == token.CONCAT_ASSIGN {
		// 组合赋值：转换为二元运算表达式
		// a += b 转换为 a = a + b
		var op string
		switch opToken.Type {
		case token.PLUS_ASSIGN:
			op = "+"
		case token.MINUS_ASSIGN:
			op = "-"
		case token.STAR_ASSIGN:
			op = "*"
		case token.SLASH_ASSIGN:
			op = "/"
		case token.AND_ASSIGN:
			op = "&"
		case token.OR_ASSIGN:
			op = "|"
		case token.XOR_ASSIGN:
			op = "^"
		case token.SHL_ASSIGN:
			op = "<<"
		case token.SHR_ASSIGN:
			op = ">>"
		case token.CONCAT_ASSIGN:
			op = ".."
		}

		// 解析右侧表达式
		right := p.parseExpression(LOWEST)

		// 创建二元运算：name op right
		stmt.Value = &BinaryExpr{
			Token:    opToken,
			Left:     name,
			Operator: op,
			Right:    right,
		}
	} else {
		// 普通赋值
		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement 解析表达式语句
func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExprStmt{Token: p.cur}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// ============================================================================
// 表达式解析
// ============================================================================

// parseExpression 解析表达式（Pratt 算法核心）
func (p *Parser) parseExpression(prec Precedence) Expression {
	prefix := p.prefixParseFns[p.cur.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.cur.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && !p.peekIsNewline() && prec < p.peekPrecedence() {
		infix := p.infixParseFns[p.peek.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier 解析标识符
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Token: p.cur, Value: p.cur.Literal}
}

// parseInstanceVar 解析实例变量（@member）
func (p *Parser) parseInstanceVar() Expression {
	return &Identifier{Token: p.cur, Value: p.cur.Literal}
}

// parseNumberLiteral 解析数字字面量
func (p *Parser) parseNumberLiteral() Expression {
	return &NumberLiteral{Token: p.cur, Value: p.cur.Literal}
}

// parseStringLiteral 解析字符串字面量（支持插值）
func (p *Parser) parseStringLiteral() Expression {
	// 获取第一个字符串片段
	frag := &StringLiteral{Token: p.cur, Value: p.cur.Literal}

	// 检查是否有插值 #{...}
	if p.peekTokenIs(token.INTERP_START) {
		return p.parseInterpolatedString(frag)
	}

	// 普通字符串，直接返回
	return frag
}

// parseRegexLiteral 解析正则字面量 #/pattern/flags#
//
// Token literal 格式：#/pattern/flags# （pattern 含原始转义）
// 需要解析出 pattern 和 flags 两部分
func (p *Parser) parseRegexLiteral() Expression {
	lit := p.cur.Literal

	// 解析 #/pattern/flags# 格式
	// 去掉首尾的 #/
	if len(lit) < 4 || lit[0] != '#' || lit[1] != '/' || lit[len(lit)-1] != '#' {
		p.errors = append(p.errors, fmt.Sprintf("invalid regex literal: %s", lit))
		return nil
	}

	// 去掉 #/ 前缀和 # 后缀
	inner := lit[2 : len(lit)-1]

	// 找到最后一个 / 来分隔 pattern 和 flags
	// 需要跳过转义的 \/
	lastSlash := -1
	for i := len(inner) - 1; i >= 0; i-- {
		if inner[i] == '/' {
			// 检查前面是否有奇数个 \
			slashCount := 0
			for j := i - 1; j >= 0 && inner[j] == '\\'; j-- {
				slashCount++
			}
			if slashCount%2 == 0 {
				lastSlash = i
				break
			}
		}
	}

	if lastSlash == -1 {
		p.errors = append(p.errors, fmt.Sprintf("invalid regex literal: %s", lit))
		return nil
	}

	pattern := inner[:lastSlash]
	flags := inner[lastSlash+1:]

	return &RegexLiteral{
		Token:   p.cur,
		Pattern: pattern,
		Flags:   flags,
	}
}

// parseInterpolatedString 解析插值字符串，构建 ConcatExpr 链
// 处理："Hello #{$name}!" → ConcatExpr(ConcatExpr("Hello ", $name), "!")
func (p *Parser) parseInterpolatedString(firstFrag *StringLiteral) Expression {
	// 从第一个字符串片段开始构建连接表达式
	var result Expression = firstFrag

	for p.peekTokenIs(token.INTERP_START) {
		// 消费 #{
		p.nextToken() // INTERP_START

		// 解析插值表达式（Phase 10.3：支持完整表达式）
		p.nextToken()

		// 解析任意表达式（对象访问、数组索引、算术运算等）
		// 注意：INTERP_END (}) 会终止表达式解析
		var expr = p.parseExpression(LOWEST)

		// 期望 INTERP_END (})
		// parseExpression 可能在遇到 INTERP_END 前停止，所以检查 cur 和 peek
		if p.curTokenIs(token.INTERP_END) {
			// 表达式刚好在 } 前结束，cur 是 INTERP_END
		} else if p.peekTokenIs(token.INTERP_END) {
			// 需要消费 INTERP_END
			p.nextToken()
		} else {
			// 错误：缺少 }
			p.errors = append(p.errors, "expected '}' to close interpolation")
			return result
		}
		p.nextToken() // 消费 INTERP_END

		// 构建 ConcatExpr：result .. expr
		result = &ConcatExpr{
			Token: p.cur,
			Left:  result,
			Right: expr,
		}

		// 检查后面的字符串片段（此时 p.cur 已经是 INTERP_END 后面的 token）
		// 注意：第二次 p.nextToken() 已经把 STRING_FRAG 或 STRING 移到了 p.cur
		if p.curTokenIs(token.STRING_FRAG) {
			frag := &StringLiteral{Token: p.cur, Value: p.cur.Literal}
			// 继续连接：result .. frag
			result = &ConcatExpr{
				Token: p.cur,
				Left:  result,
				Right: frag,
			}
		} else if p.curTokenIs(token.STRING) {
			// 结束标记（空字符串或非空结尾）
			// 如果是空结尾，不添加
			if p.cur.Literal != "" {
				frag := &StringLiteral{Token: p.cur, Value: p.cur.Literal}
				result = &ConcatExpr{
					Token: p.cur,
					Left:  result,
					Right: frag,
				}
			}
		}
	}

	return result
}

// parseBoolLiteral 解析布尔字面量
func (p *Parser) parseBoolLiteral() Expression {
	value := p.cur.Type == token.KW_TRUE
	return &BoolLiteral{Token: p.cur, Value: value}
}

// parseNullLiteral 解析 null 字面量
func (p *Parser) parseNullLiteral() Expression {
	return &NullLiteral{Token: p.cur}
}

// parseTypeCast 解析 Go 风格的类型转换表达式
// 支持语法: int(x), float(x), string(x), bool(x)
//
// 类型关键字（KW_INT, KW_FLOAT, KW_STRING, KW_BOOL）在 prefixParseFns 中注册，
// 当 parser 遇到这些 token 时会调用此函数。
//
// 解析过程：
//  1. 获取当前 token 作为目标类型（int/float/string/bool）
//  2. 期望并消费左括号 '('
//  3. 使用 parseExpression(LOWEST) 解析被转换的表达式
//  4. 期望并消费右括号 ')'
//  5. 返回 TypeCast AST 节点
//
// 错误处理：
//   - 如果类型关键字后没有 '('，添加错误信息并返回 nil
//   - 如果表达式后没有 ')'，添加错误信息并返回 nil
//
// AST 节点：
//
//	TypeCast.Token = 类型关键字 token
//	TypeCast.Type  = 目标类型字符串（"int"/"float"/"string"/"bool"）
//	TypeCast.Expr  = 被转换的表达式节点
//
// 示例：
//
//	int("42") 解析为 TypeCast{Type: "int", Expr: StringLiteral{Value: "42"}}
//	float($x + 1) 解析为 TypeCast{Type: "float", Expr: BinaryExpr{...}}
func (p *Parser) parseTypeCast() Expression {
	cast := &TypeCast{
		Token: p.cur,
		Type:  p.cur.Literal,
	}

	// 期望下一个 token 是左括号 (
	if !p.expectPeek(token.LPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("expected '(' after type '%s'", cast.Type))
		return nil
	}

	// 消费左括号
	p.nextToken()

	// 解析被转换的表达式
	cast.Expr = p.parseExpression(LOWEST)

	// 期望右括号
	if !p.expectPeek(token.RPAREN) {
		p.errors = append(p.errors, "expected ')' after type cast expression")
		return nil
	}

	return cast
}

// parseInterpStartError 处理字符串外的 INTERP_START 错误
func (p *Parser) parseInterpStartError() Expression {
	p.errors = append(p.errors, "unexpected '#{' outside of string interpolation")
	return nil
}

// parseInterpEndError 处理字符串外的 INTERP_END 错误
func (p *Parser) parseInterpEndError() Expression {
	p.errors = append(p.errors, "unexpected '}' outside of string interpolation")
	return nil
}

// parseArrayLiteral 解析数组字面量
func (p *Parser) parseArrayLiteral() Expression {
	array := &ArrayLiteral{Token: p.cur}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseObjectLiteral 解析对象字面量
func (p *Parser) parseObjectLiteral() Expression {
	obj := &ObjectLiteral{Token: p.cur}
	obj.Pairs = make(map[Expression]Expression)

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.COMMENT) || p.curTokenIs(token.BLOCK_COMMENT) {
			continue
		}

		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		obj.Pairs[key] = value

		// 跳过可能的换行符
		for p.peekIsNewline() {
			p.nextToken()
		}

		if !p.peekTokenIs(token.RBRACE) {
			// 需要逗号或换行作为分隔符
			if p.peekTokenIs(token.COMMA) {
				p.nextToken() // 消费逗号
				// 消费逗号后跳过可能的换行符
				for p.peekIsNewline() {
					p.nextToken()
				}
			} else if !p.peekTokenIs(token.IDENTIFIER) && !p.peekTokenIs(token.STRING) && !p.peekTokenIs(token.SPECIAL_VAR) {
				// 后面不是另一个键（标识符、字符串或特殊变量），说明缺少逗号
				p.peekError(token.COMMA)
				return nil
			}
			// 如果是另一个键（标识符、字符串或特殊变量），继续解析（换行作为分隔符）
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return obj
}

// parseExpressionList 解析表达式列表
func (p *Parser) parseExpressionList(end token.TokenType) []Expression {
	var list []Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	// 跳过开头可能的换行符（支持空多行数组/对象）
	for p.peekIsNewline() {
		p.nextToken()
	}

	// 跳过换行后再次检查是否是结束符
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()

	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) || p.peekIsNewline() {
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // 消费逗号
			// 逗号后跳过可能的换行
			for p.peekIsNewline() {
				p.nextToken()
			}
		} else {
			p.nextToken() // 消费换行
			// 跳过可能的连续换行
			for p.peekIsNewline() {
				p.nextToken()
			}
		}

		// 如果换行后是闭合括号，结束解析
		if p.peekTokenIs(end) {
			break
		}

		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseGroupedExpression 解析分组表达式或箭头函数参数
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()

	// 检查是否为空参数列表的箭头函数
	if p.curTokenIs(token.RPAREN) && p.peekTokenIs(token.ARROW) {
		return p.parseArrowFunctionParams([]*Identifier{})
	}

	// 解析第一个表达式或标识符
	exp := p.parseExpression(LOWEST)

	// 检查是否为逗号（可能是箭头函数参数列表）
	if p.peekTokenIs(token.COMMA) {
		// 尝试解析为参数列表
		params := []*Identifier{}

		// 第一个参数
		if ident, ok := exp.(*Identifier); ok {
			params = append(params, ident)
		} else {
			// 不是标识符，回退为普通表达式
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
			return exp
		}

		// 解析剩余参数
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // 消费逗号
			p.nextToken() // 移动到下一个 token

			if ident, ok := p.parseIdentifier().(*Identifier); ok {
				params = append(params, ident)
			} else {
				// 不是标识符，回退为普通表达式
				if !p.expectPeek(token.RPAREN) {
					return nil
				}
				return exp
			}
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}

		// 检查是否为箭头函数
		if p.peekTokenIs(token.ARROW) {
			return p.parseArrowFunctionParams(params)
		}

		// 不是箭头函数，这是语法错误
		return nil
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// 检查是否为箭头函数（单参数）
	if p.peekTokenIs(token.ARROW) {
		// 将表达式转换为参数列表
		if ident, ok := exp.(*Identifier); ok {
			return p.parseArrowFunctionParams([]*Identifier{ident})
		}
	}

	return exp
}

// parseLambdaExpression 解析 Lambda 表达式
// 语法: fn($x) { return $x * 2 }
func (p *Parser) parseLambdaExpression() Expression {
	lambda := &LambdaExpr{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lambda.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lambda.Body = p.parseBlockStatement()

	return lambda
}

// parseArrowFunctionParams 解析箭头函数参数并返回箭头函数表达式
func (p *Parser) parseArrowFunctionParams(params []*Identifier) Expression {
	arrow := &ArrowExpr{Token: p.cur}
	arrow.Parameters = params

	p.nextToken() // 消费 ->

	// 检查是否为块体
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		arrow.BlockBody = p.parseBlockStatement()
	} else {
		p.nextToken()
		arrow.Body = p.parseExpression(LOWEST)
	}

	return arrow
}

// parseIfExpression 解析 if 表达式
func (p *Parser) parseIfExpression() Expression {
	stmt := &IfStmt{Token: p.cur}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Cond = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if p.peekTokenIs(token.LBRACE) {
			p.nextToken()
			stmt.Else = p.parseBlockStatement()
		}
	}

	return stmt
}

// parseMatchExpression 解析 match 表达式
func (p *Parser) parseMatchExpression() Expression {
	stmt := p.parseMatchStatement().(*MatchStmt)
	stmt.IsExpr = true
	return stmt
}

// ============================================================================
// 中缀解析函数
// ============================================================================

// parseInfixExpression 解析中缀表达式
func (p *Parser) parseInfixExpression(left Expression) Expression {
	// 特殊处理字符串连接
	if p.cur.Type == token.CONCAT {
		exp := &ConcatExpr{
			Token: p.cur,
			Left:  left,
		}
		precedence := p.curPrecedence()
		p.nextToken()
		exp.Right = p.parseExpression(precedence)
		return exp
	}

	// 特殊处理范围表达式
	if p.cur.Type == token.ELLIPSIS || p.cur.Type == token.RANGE_INCLUSIVE {
		exp := &RangeExpr{
			Token:     p.cur,
			Start:     left,
			Inclusive: p.cur.Type == token.RANGE_INCLUSIVE,
		}
		precedence := p.curPrecedence()
		p.nextToken()
		exp.End = p.parseExpression(precedence)
		return exp
	}

	exp := &BinaryExpr{
		Token:    p.cur,
		Left:     left,
		Operator: p.cur.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)

	return exp
}

// parseCallExpression 解析函数调用表达式
func (p *Parser) parseCallExpression(left Expression) Expression {
	exp := &CallExpr{Token: p.cur, Function: left}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseImplicitCallExpression 解析无括号函数调用表达式
// 处理：println "hello" → 等同于 println("hello")
// 在 Pratt 解析器中，当函数名后面紧跟字符串字面量时触发
func (p *Parser) parseImplicitCallExpression(left Expression) Expression {
	exp := &CallExpr{Token: p.cur, Function: left}
	// 当前 token 已经是 STRING 或 STRING_FRAG，解析它作为参数
	arg := p.parseExpression(LOWEST)
	exp.Arguments = []Expression{arg}
	return exp
}

// parseIndexExpression 解析索引表达式
func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpr{Token: p.cur, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseMemberExpression 解析成员访问表达式
func (p *Parser) parseMemberExpression(left Expression) Expression {
	exp := &MemberExpr{Token: p.cur, Object: left}

	if !p.expectPeek(token.IDENTIFIER) {
		return nil
	}

	exp.Member = &Identifier{Token: p.cur, Value: p.cur.Literal}

	return exp
}

// parseAssignExpression 解析赋值表达式
func (p *Parser) parseAssignExpression(left Expression) Expression {
	exp := &AssignExpr{
		Token: p.cur,
		Left:  left,
	}

	operator := p.cur.Literal
	precedence := curPrecedence(p.cur.Type)
	p.nextToken()

	if operator == "=" {
		exp.Value = p.parseExpression(precedence)
	} else {
		// 复合赋值运算符转换为普通运算
		// 例如: a += b 转换为 a = a + b
		op := operator[:len(operator)-1]
		exp.Value = &BinaryExpr{
			Token:    exp.Token,
			Left:     left,
			Operator: op,
			Right:    p.parseExpression(precedence),
		}
	}

	return exp
}

// curPrecedence 获取 Token 类型的优先级
func curPrecedence(t token.TokenType) Precedence {
	if p, ok := precedences[t]; ok {
		return p
	}
	return LOWEST
}

// parseTernaryExpression 解析三元表达式
func (p *Parser) parseTernaryExpression(left Expression) Expression {
	exp := &TernaryExpr{
		Token:     p.cur,
		Condition: left,
	}

	p.nextToken()
	exp.TrueExpr = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	p.nextToken()
	exp.FalseExpr = p.parseExpression(LOWEST)

	return exp
}

// parseArrowExpression 解析箭头函数表达式
func (p *Parser) parseArrowExpression(left Expression) Expression {
	arrow := &ArrowExpr{Token: p.cur}

	// 将左侧表达式转换为参数
	if ident, ok := left.(*Identifier); ok {
		arrow.Parameters = []*Identifier{ident}
	} else {
		p.errors = append(p.errors, "箭头函数左侧必须是标识符")
		return nil
	}

	// 检查是否为块体
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		arrow.BlockBody = p.parseBlockStatement()
	} else {
		p.nextToken()
		arrow.Body = p.parseExpression(LOWEST)
	}

	return arrow
}

// parsePipeExpression 解析管道表达式
// |> 正向管道（左结合）: a |> f(b,c) = f(a, b, c)
// <| 反向管道（右结合）: f(b,c) <| a = f(b, c, a)
func (p *Parser) parsePipeExpression(left Expression) Expression {
	pipe := &PipeExpr{
		Token: p.cur,
		Left:  left,
	}

	if p.cur.Type == token.PIPE_FWD {
		pipe.Forward = true
		// 左结合: 右侧用相同优先级解析
		precedence := p.curPrecedence()
		p.nextToken()
		pipe.Right = p.parseExpression(precedence)
	} else {
		// PIPE_BWD
		pipe.Forward = false
		// 右结合: 右侧用较低优先级解析（PIPE_FWD_PREC < PIPE_BWD_PREC）
		p.nextToken()
		pipe.Right = p.parseExpression(PIPE_FWD_PREC)
	}

	return pipe
}

// ============================================================================
// 前缀表达式解析
// ============================================================================

// parsePrefixExpression 解析前缀表达式
func (p *Parser) parsePrefixExpression() Expression {
	exp := &UnaryExpr{
		Token:    p.cur,
		Operator: p.cur.Literal,
	}

	p.nextToken()
	exp.Operand = p.parseExpression(PREFIX_PREC)

	return exp
}

// ============================================================================
// 模式解析
// ============================================================================

// parsePattern 解析模式
func (p *Parser) parsePattern() Pattern {
	switch p.cur.Type {
	case token.UNDERSCORE:
		return &WildcardPattern{Token: p.cur}
	case token.IDENTIFIER, token.SPECIAL_VAR:
		ident := &Identifier{Token: p.cur, Value: p.cur.Literal}

		// 检查是否为 OR 模式 (逗号分隔)
		if p.peekTokenIs(token.COMMA) {
			firstPattern := &IdentifierPattern{Token: ident.Token, Name: ident}
			return p.parseOrPattern(firstPattern)
		}

		return &IdentifierPattern{Token: p.cur, Name: ident}
	case token.LBRACKET:
		return p.parseArrayPattern()
	case token.LBRACE:
		return p.parseObjectPattern()
	case token.INTEGER, token.FLOAT, token.STRING,
		token.KW_TRUE, token.KW_FALSE, token.KW_NULL:
		lit := p.parseExpression(LOWEST)
		if re, ok := lit.(*RangeExpr); ok {
			return &RangePattern{
				Token:     re.Token,
				Start:     re.Start,
				End:       re.End,
				Inclusive: re.Inclusive,
			}
		}
		if p.peekTokenIs(token.COMMA) {
			firstPattern := &LiteralPattern{Token: p.cur, Value: lit}
			return p.parseOrPattern(firstPattern)
		}
		return &LiteralPattern{Token: p.cur, Value: lit}
	case token.REGEX:
		// 解析正则模式 #/pattern/flags# [as $var]
		regexLit := p.parseRegexLiteral().(*RegexLiteral)
		rp := &RegexPattern{
			Token:   regexLit.Token,
			Pattern: regexLit.Pattern,
			Flags:   regexLit.Flags,
		}

		// 检查是否有 as $var 绑定
		if p.peekTokenIs(token.AS) {
			p.nextToken() // consume peek (AS)
			p.nextToken() // move to identifier
			if p.cur.Type != token.IDENTIFIER && p.cur.Type != token.SPECIAL_VAR {
				p.errors = append(p.errors, fmt.Sprintf("expected identifier after 'as', got %s", p.cur.Literal))
				return rp
			}
			rp.Binding = &Identifier{Token: p.cur, Value: p.cur.Literal}
		}

		return rp
	default:
		lit := p.parseExpression(LOWEST)
		return &LiteralPattern{Token: p.cur, Value: lit}
	}
}

// parseOrPattern 解析 OR 模式 (逗号分隔)
func (p *Parser) parseOrPattern(first Pattern) Pattern {
	or := &OrPattern{Token: p.peek}
	or.Patterns = []Pattern{first}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next token

		var pat Pattern
		switch p.cur.Type {
		case token.UNDERSCORE:
			pat = &WildcardPattern{Token: p.cur}
		case token.IDENTIFIER, token.SPECIAL_VAR:
			ident := &Identifier{Token: p.cur, Value: p.cur.Literal}
			pat = &IdentifierPattern{Token: p.cur, Name: ident}
		case token.INTEGER, token.FLOAT, token.STRING,
			token.KW_TRUE, token.KW_FALSE, token.KW_NULL:
			lit := p.parseExpression(LOWEST)
			pat = &LiteralPattern{Token: p.cur, Value: lit}
		default:
			lit := p.parseExpression(LOWEST)
			pat = &LiteralPattern{Token: p.cur, Value: lit}
		}
		or.Patterns = append(or.Patterns, pat)
	}

	return or
}

// parseArrayPattern 解析数组模式
func (p *Parser) parseArrayPattern() Pattern {
	pattern := &ArrayPattern{Token: p.cur}

	for !p.peekTokenIs(token.RBRACKET) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTokenIs(token.COMMA) {
			continue
		}

		// 检查剩余元素
		if p.curTokenIs(token.DOT) && p.peekTokenIs(token.DOT) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(token.IDENTIFIER) || p.curTokenIs(token.SPECIAL_VAR) {
				pattern.Rest = &Identifier{Token: p.cur, Value: p.cur.Literal}
			}
			break
		}

		pattern.Elements = append(pattern.Elements, p.parsePattern())
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return pattern
}

// parseObjectPattern 解析对象模式
func (p *Parser) parseObjectPattern() Pattern {
	pattern := &ObjectPattern{Token: p.cur}
	pattern.Pairs = make(map[string]Pattern)

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()

		if p.curTokenIs(token.COMMA) {
			continue
		}

		// 检查剩余属性
		if p.curTokenIs(token.DOT) && p.peekTokenIs(token.DOT) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(token.IDENTIFIER) || p.curTokenIs(token.SPECIAL_VAR) {
				pattern.Rest = &Identifier{Token: p.cur, Value: p.cur.Literal}
			}
			break
		}

		// 解析键
		key := p.cur.Literal

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		pattern.Pairs[key] = p.parsePattern()
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return pattern
}

// ============================================================================
// 辅助函数
// ============================================================================

// ParseInt 解析整数字符串
func ParseInt(s string) (int64, error) {
	// 处理十六进制、八进制、二进制
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			return strconv.ParseInt(s[2:], 16, 64)
		case 'o', 'O':
			return strconv.ParseInt(s[2:], 8, 64)
		case 'b', 'B':
			return strconv.ParseInt(s[2:], 2, 64)
		}
	}
	return strconv.ParseInt(s, 10, 64)
}

// ParseFloat 解析浮点数字符串
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
