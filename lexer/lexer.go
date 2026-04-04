// Package lexer 实现 JPL 脚本语言的词法分析器
package lexer

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"github.com/gnuos/jpl/token"
)

// Lexer 词法分析器
type Lexer struct {
	source []rune // 源代码字符数组
	pos    int    // 当前位置
	line   int    // 当前行号
	column int    // 当前列号
	file   string // 文件名

	// 字符串插值状态（Phase 10.2）
	inString      bool           // 当前是否在字符串内
	stringQuote   rune           // 字符串引号类型（' 或 ")
	isTripleQuote bool           // 是否三引号字符串
	stringPos     token.Position // 字符串开始位置
	interpDepth   int            // 插值嵌套深度（用于处理嵌套 #{ ... #{ ... } ... }）
	inInterp      bool           // 当前是否在插值表达式内（扫描表达式而非字符串）
}

// NewLexer 创建新的词法分析器
func NewLexer(source string, file string) *Lexer {
	return &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		column: 1,
		file:   file,
	}
}

// NextToken 获取下一个 Token
func (l *Lexer) NextToken() token.Token {
	// 字符串插值模式处理
	if l.inString && !l.inInterp {
		// 如果在字符串内（但不在插值表达式内）且遇到 #{，返回 INTERP_START
		if l.current() == '#' && l.peek() == '{' {
			pos := l.currentPos()
			l.advance() // #
			l.advance() // {
			l.interpDepth++
			l.inInterp = true // 进入插值表达式模式
			return l.newToken(token.INTERP_START, "#{", pos)
		}

		// 在字符串内，继续扫描字符串
		pos := l.currentPos()
		if l.isTripleQuote {
			return l.scanTripleStringContinue(pos)
		}
		return l.scanStringContinue(pos)
	}

	// 如果在插值表达式内且遇到 }，返回 INTERP_END
	if l.inInterp && l.current() == '}' {
		pos := l.currentPos()
		l.advance()
		l.interpDepth--
		if l.interpDepth <= 0 {
			l.inInterp = false
			l.interpDepth = 0
			// 如果还在字符串内，下次调用会继续扫描字符串
		}
		return l.newToken(token.INTERP_END, "}", pos)
	}

	// 跳过空白字符（不包括换行符）
	l.skipWhitespace()

	// 记录当前位置
	pos := l.currentPos()

	// 到达文件末尾
	if l.atEnd() {
		return l.newToken(token.EOF, "", pos)
	}

	ch := l.current()

	// 处理换行符
	if ch == '\n' || ch == '\r' {
		return l.scanNewline(pos)
	}

	// 处理注释
	if ch == '/' {
		if l.peek() == '/' {
			return l.scanLineComment(pos)
		}
		if l.peek() == '*' {
			return l.scanBlockComment(pos)
		}
	}

	// 处理数字
	if isDigit(ch) || (ch == '.' && isDigit(l.peek())) {
		return l.scanNumber(pos)
	}

	// 处理标识符（$var 或 var，$ 前缀可选）
	if ch == '$' {
		return l.scanVariable(pos)
	}

	// 处理实例变量（@member）
	if ch == '@' {
		return l.scanInstanceVar(pos)
	}

	// 处理间接变量引用（`name）
	if ch == '`' {
		return l.scanBacktick(pos)
	}

	// 处理普通标识符或关键字
	if isIdentifierStart(ch) {
		return l.scanIdentifier(pos)
	}

	// 处理下划线
	if ch == '_' {
		return l.scanUnderscore(pos)
	}

	// 处理字符串（包括多行字符串）
	if ch == '"' || ch == '\'' {
		// 检查是否为三引号
		if ch == '"' && l.peek() == '"' && l.peekAhead(2) == '"' {
			return l.scanTripleString(pos, '"')
		}
		if ch == '\'' && l.peek() == '\'' && l.peekAhead(2) == '\'' {
			return l.scanTripleString(pos, '\'')
		}
		return l.scanString(pos)
	}

	// 处理正则字面量 #/pattern/flags#
	if ch == '#' && l.peek() == '/' {
		return l.scanRegexLiteral(pos)
	}

	// 处理运算符
	if isOperator(ch) {
		return l.scanOperator(pos)
	}

	// 处理分隔符
	if isDelimiter(ch) {
		return l.scanDelimiter(pos)
	}

	// 非法字符
	l.advance()
	return l.newToken(token.ILLEGAL, string(ch), pos)
}

// currentPos 获取当前位置信息
func (l *Lexer) currentPos() token.Position {
	return token.Position{
		Filename: l.file,
		Line:     l.line,
		Column:   l.column,
		Offset:   l.pos,
	}
}

// current 获取当前字符
func (l *Lexer) current() rune {
	if l.atEnd() {
		return 0
	}
	return l.source[l.pos]
}

// peek 查看下一个字符（不移动位置）
func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.source) {
		return 0
	}
	return l.source[l.pos+1]
}

// peekAhead 查看指定偏移位置的字符（不移动位置）
func (l *Lexer) peekAhead(offset int) rune {
	if l.pos+offset >= len(l.source) {
		return 0
	}
	return l.source[l.pos+offset]
}

// advance 移动到下一个字符
func (l *Lexer) advance() rune {
	if l.atEnd() {
		return 0
	}
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return ch
}

// atEnd 检查是否到达文件末尾
func (l *Lexer) atEnd() bool {
	return l.pos >= len(l.source)
}

// skipWhitespace 跳过空白字符（不包括换行符）
func (l *Lexer) skipWhitespace() {
	for !l.atEnd() {
		ch := l.current()
		if ch == ' ' || ch == '\t' || ch == '\v' || ch == '\f' {
			l.advance()
		} else {
			break
		}
	}
}

// newToken 创建新的 Token
func (l *Lexer) newToken(typ token.TokenType, literal string, pos token.Position) token.Token {
	return token.Token{
		Type:    typ,
		Literal: literal,
		Pos:     pos,
	}
}

// scanNewline 扫描换行符
func (l *Lexer) scanNewline(pos token.Position) token.Token {
	ch := l.advance()
	if ch == '\r' && l.current() == '\n' {
		l.advance()
	}
	return l.newToken(token.NEWLINE, "\n", pos)
}

// scanLineComment 扫描单行注释
func (l *Lexer) scanLineComment(pos token.Position) token.Token {
	// 跳过 //
	l.advance()
	l.advance()

	// 收集注释内容直到换行符
	var sb strings.Builder
	sb.WriteString("//")
	for !l.atEnd() && l.current() != '\n' && l.current() != '\r' {
		sb.WriteRune(l.advance())
	}

	// 不消费换行符，返回 COMMENT token
	return l.newToken(token.COMMENT, sb.String(), pos)
}

// scanBlockComment 扫描多行注释
func (l *Lexer) scanBlockComment(pos token.Position) token.Token {
	// 跳过 /*
	l.advance()
	l.advance()

	var sb strings.Builder
	sb.WriteString("/*")
	for !l.atEnd() {
		ch := l.advance()
		if ch == '*' && l.current() == '/' {
			l.advance()
			sb.WriteString("*/")
			break
		}
		sb.WriteRune(ch)
	}

	// 返回 BLOCK_COMMENT token
	return l.newToken(token.BLOCK_COMMENT, sb.String(), pos)
}

// scanNumber 扫描数字
func (l *Lexer) scanNumber(pos token.Position) token.Token {
	var sb strings.Builder

	// 检查十六进制、八进制、二进制
	if l.current() == '0' {
		sb.WriteRune(l.advance())
		if !l.atEnd() {
			next := l.current()
			if next == 'x' || next == 'X' {
				// 十六进制
				sb.WriteRune(l.advance())
				for !l.atEnd() && isHexDigit(l.current()) {
					sb.WriteRune(l.advance())
				}
				return l.newToken(token.INTEGER, sb.String(), pos)
			}
			if next == 'o' || next == 'O' {
				// 八进制
				sb.WriteRune(l.advance())
				for !l.atEnd() && isOctalDigit(l.current()) {
					sb.WriteRune(l.advance())
				}
				return l.newToken(token.INTEGER, sb.String(), pos)
			}
			if next == 'b' || next == 'B' {
				// 二进制
				sb.WriteRune(l.advance())
				for !l.atEnd() && isBinaryDigit(l.current()) {
					sb.WriteRune(l.advance())
				}
				return l.newToken(token.INTEGER, sb.String(), pos)
			}
		}
	}

	// 扫描整数部分
	for !l.atEnd() && isDigit(l.current()) {
		sb.WriteRune(l.advance())
	}

	isFloat := false

	// 检查小数点
	if !l.atEnd() && l.current() == '.' && l.peek() != '.' {
		isFloat = true
		sb.WriteRune(l.advance())
		for !l.atEnd() && isDigit(l.current()) {
			sb.WriteRune(l.advance())
		}
	}

	// 检查科学计数法
	if !l.atEnd() && (l.current() == 'e' || l.current() == 'E') {
		isFloat = true
		sb.WriteRune(l.advance())
		if !l.atEnd() && (l.current() == '+' || l.current() == '-') {
			sb.WriteRune(l.advance())
		}
		for !l.atEnd() && isDigit(l.current()) {
			sb.WriteRune(l.advance())
		}
	}

	literal := sb.String()

	// 检查显式类型后缀：n = BigInt, d = BigDecimal
	if !l.atEnd() && (l.current() == 'n' || l.current() == 'd') {
		suffix := l.current()
		// 确保 n/d 后面不是标识符字符（避免匹配变量名如 "name"）
		if l.atEnd() || !isIdentifierPart(l.peek()) {
			l.advance() // 消费 n 或 d
			if suffix == 'n' {
				return l.newToken(token.BIGINT, literal, pos)
			}
			// suffix == 'd'
			return l.newToken(token.BIGDECIMAL, literal, pos)
		}
	}

	// 检查是否需要转为大数类型
	if isFloat {
		// 尝试解析为 float64
		if _, _, err := big.ParseFloat(literal, 10, 256, big.ToNearestEven); err == nil {
			// 检查精度是否超过 float64
			if len(literal) > 15 {
				return l.newToken(token.BIGDECIMAL, literal, pos)
			}
		}
		return l.newToken(token.FLOAT, literal, pos)
	}

	// 整数处理
	if isInt64Overflow(literal) {
		return l.newToken(token.BIGINT, literal, pos)
	}
	return l.newToken(token.INTEGER, literal, pos)
}

// scanVariable 扫描 $ 前缀变量（$ 前缀可选）
func (l *Lexer) scanVariable(pos token.Position) token.Token {
	// 跳过 $
	l.advance()

	// 检查特殊变量 $_
	if !l.atEnd() && l.current() == '_' {
		next := l.peek()
		if !isIdentifierPart(next) {
			l.advance()
			return l.newToken(token.SPECIAL_VAR, "$_", pos)
		}
	}

	// $ 后必须紧跟标识符字符
	if l.atEnd() || !isIdentifierStart(l.current()) && l.current() != '_' {
		return l.newToken(token.ILLEGAL, "$", pos)
	}

	// 扫描变量名
	var sb strings.Builder
	sb.WriteRune('$')

	for !l.atEnd() && isIdentifierPart(l.current()) {
		sb.WriteRune(l.advance())
	}

	literal := sb.String()

	return l.newToken(token.IDENTIFIER, literal, pos)
}

// scanInstanceVar 扫描实例变量（@member）
func (l *Lexer) scanInstanceVar(pos token.Position) token.Token {
	l.advance() // 跳过 @

	// @ 后必须紧跟标识符字符
	if l.atEnd() || !isIdentifierStart(l.current()) && l.current() != '_' {
		return l.newToken(token.ILLEGAL, "@", pos)
	}

	// 扫描成员名
	var sb strings.Builder
	sb.WriteRune('@')

	for !l.atEnd() && isIdentifierPart(l.current()) {
		sb.WriteRune(l.advance())
	}

	literal := sb.String()

	return l.newToken(token.INSTANCE_VAR, literal, pos)
}

// scanBacktick 扫描间接变量引用（`name）
func (l *Lexer) scanBacktick(pos token.Position) token.Token {
	l.advance() // 跳过 `

	// ` 后必须紧跟标识符字符
	if l.atEnd() || !isIdentifierStart(l.current()) && l.current() != '_' {
		return l.newToken(token.ILLEGAL, "`", pos)
	}

	// 扫描变量名
	var sb strings.Builder
	sb.WriteRune('`')

	for !l.atEnd() && isIdentifierPart(l.current()) {
		sb.WriteRune(l.advance())
	}

	literal := sb.String()

	return l.newToken(token.BACKTICK, literal, pos)
}

// scanUnderscore 扫描下划线
func (l *Lexer) scanUnderscore(pos token.Position) token.Token {
	l.advance()

	// 检查是否为标识符的一部分
	if !l.atEnd() && isIdentifierPart(l.current()) {
		var sb strings.Builder
		sb.WriteRune('_')
		for !l.atEnd() && isIdentifierPart(l.current()) {
			sb.WriteRune(l.advance())
		}
		literal := sb.String()

		// 检查是否为关键字
		tokType := token.Keyword(literal)
		if tokType != token.IDENTIFIER {
			return l.newToken(tokType, literal, pos)
		}
		return l.newToken(token.IDENTIFIER, literal, pos)
	}

	// 单个下划线
	return l.newToken(token.UNDERSCORE, "_", pos)
}

// scanIdentifier 扫描标识符或关键字
func (l *Lexer) scanIdentifier(pos token.Position) token.Token {
	var sb strings.Builder

	for !l.atEnd() && isIdentifierPart(l.current()) {
		sb.WriteRune(l.advance())
	}

	literal := sb.String()

	// 检查是否为关键字
	tokType := token.Keyword(literal)
	return l.newToken(tokType, literal, pos)
}

// scanString 扫描字符串（支持插值）
func (l *Lexer) scanString(pos token.Position) token.Token {
	quote := l.advance() // 记录引号类型并跳过

	var sb strings.Builder
	for !l.atEnd() {
		ch := l.current()

		// 结束引号
		if ch == quote {
			l.advance()
			// 普通字符串（无插值），返回完整字符串
			return l.newToken(token.STRING, sb.String(), pos)
		}

		// 字符串插值：双引号字符串遇到 #{ 开始插值（Phase 10.2）
		if quote == '"' && ch == '#' && l.peek() == '{' {
			// 如果有积累的字符串内容，先返回字符串片段，并设置状态让下次处理 INTERP_START
			if sb.Len() > 0 {
				l.inString = true
				l.stringQuote = quote
				l.isTripleQuote = false
				l.stringPos = pos
				return l.newToken(token.STRING_FRAG, sb.String(), pos)
			}
			// 没有内容，设置状态，NextToken 下次会处理 INTERP_START
			l.inString = true
			l.stringQuote = quote
			l.isTripleQuote = false
			// 返回空字符串标记，实际 INTERP_START 由 NextToken 处理
			return l.newToken(token.STRING, "", pos)
		}

		// 转义字符
		if ch == '\\' {
			l.advance()
			if l.atEnd() {
				return l.newToken(token.ILLEGAL, sb.String(), pos)
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			case '0':
				sb.WriteRune('\000')
			case 'x':
				// 十六进制转义 \xHH
				if !l.atEnd() && isHexDigit(l.current()) {
					high := hexValue(l.advance())
					var low byte
					if !l.atEnd() && isHexDigit(l.current()) {
						low = hexValue(l.advance())
					}
					sb.WriteByte(high<<4 | low)
				}
			case 'u':
				// Unicode 转义 \uHHHH
				sb.WriteRune(l.scanUnicode(4))
			case 'U':
				// Unicode 转义 \UHHHHHHHH
				sb.WriteRune(l.scanUnicode(8))
			default:
				sb.WriteRune(escaped)
			}
			continue
		}

		// 换行符（单引号字符串不允许）
		if ch == '\n' || ch == '\r' {
			if quote == '\'' {
				return l.newToken(token.ILLEGAL, sb.String(), pos)
			}
		}

		sb.WriteRune(l.advance())
	}

	// 未闭合的字符串
	return l.newToken(token.ILLEGAL, sb.String(), pos)
}

// scanTripleString 扫描三引号多行字符串（”' 或 """）
func (l *Lexer) scanTripleString(pos token.Position, quote rune) token.Token {
	// 跳过三个引号
	l.advance()
	l.advance()
	l.advance()

	// 确定 token 类型
	var tokType token.TokenType
	if quote == '\'' {
		tokType = token.TRIPLE_SINGLE
	} else {
		tokType = token.TRIPLE_DOUBLE
	}

	var sb strings.Builder
	for !l.atEnd() {
		// 检查结束三引号
		if l.current() == quote && l.peek() == quote && l.peekAhead(2) == quote {
			// 跳过结束引号
			l.advance()
			l.advance()
			l.advance()
			return l.newToken(tokType, sb.String(), pos)
		}

		// 字符串插值：双引号三引号遇到 #{ 开始插值（Phase 10.2）
		if quote == '"' && l.current() == '#' && l.peek() == '{' {
			// 如果有积累的字符串内容，先返回字符串片段，并设置状态让下次处理 INTERP_START
			if sb.Len() > 0 {
				l.inString = true
				l.stringQuote = quote
				l.isTripleQuote = true
				l.stringPos = pos
				return l.newToken(token.STRING_FRAG, sb.String(), pos)
			}
			// 没有内容，设置状态，NextToken 下次会处理 INTERP_START
			l.inString = true
			l.stringQuote = quote
			l.isTripleQuote = true
			// 返回空字符串标记，实际 INTERP_START 由 NextToken 处理
			return l.newToken(token.TRIPLE_DOUBLE, "", pos)
		}

		// 转义字符处理（与单引号字符串相同，支持基本的转义）
		if l.current() == '\\' {
			l.advance()
			if l.atEnd() {
				return l.newToken(token.ILLEGAL, sb.String(), pos)
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune(escaped)
			}
			continue
		}

		// 普通字符（包括换行符，多行字符串允许换行）
		sb.WriteRune(l.advance())
	}

	// 未闭合的多行字符串
	return l.newToken(token.ILLEGAL, sb.String(), pos)
}

// scanStringContinue 在插值结束后继续扫描双引号字符串
func (l *Lexer) scanStringContinue(pos token.Position) token.Token {
	quote := l.stringQuote
	var sb strings.Builder

	for !l.atEnd() {
		ch := l.current()

		// 结束引号
		if ch == quote {
			l.advance()
			l.inString = false
			// 字符串结束，返回 STRING 标记（即使有内容也返回 STRING，表示字符串结束）
			return l.newToken(token.STRING, sb.String(), pos)
		}

		// 字符串插值：遇到 #{ 开始插值
		if quote == '"' && ch == '#' && l.peek() == '{' {
			if sb.Len() > 0 {
				return l.newToken(token.STRING_FRAG, sb.String(), pos)
			}
			l.advance() // #
			l.advance() // {
			l.interpDepth++
			return l.newToken(token.INTERP_START, "#{", pos)
		}

		// 转义字符
		if ch == '\\' {
			l.advance()
			if l.atEnd() {
				l.inString = false
				return l.newToken(token.ILLEGAL, sb.String(), pos)
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			case '0':
				sb.WriteRune('\000')
			default:
				sb.WriteRune(escaped)
			}
			continue
		}

		sb.WriteRune(l.advance())
	}

	l.inString = false
	return l.newToken(token.ILLEGAL, sb.String(), pos)
}

// scanTripleStringContinue 在插值结束后继续扫描三引号字符串
func (l *Lexer) scanTripleStringContinue(pos token.Position) token.Token {
	quote := l.stringQuote
	var sb strings.Builder

	for !l.atEnd() {
		// 检查结束三引号
		if l.current() == quote && l.peek() == quote && l.peekAhead(2) == quote {
			l.advance()
			l.advance()
			l.advance()
			l.inString = false
			l.isTripleQuote = false
			// 在插值字符串中，返回片段（即使为空也返回 TRIPLE_DOUBLE 作为结束标记）
			if sb.Len() > 0 {
				return l.newToken(token.STRING_FRAG, sb.String(), pos)
			}
			return l.newToken(token.TRIPLE_DOUBLE, "", pos)
		}

		// 字符串插值：双引号三引号遇到 #{ 开始插值
		if quote == '"' && l.current() == '#' && l.peek() == '{' {
			if sb.Len() > 0 {
				return l.newToken(token.STRING_FRAG, sb.String(), pos)
			}
			l.advance() // #
			l.advance() // {
			l.interpDepth++
			return l.newToken(token.INTERP_START, "#{", pos)
		}

		// 转义字符
		if l.current() == '\\' {
			l.advance()
			if l.atEnd() {
				l.inString = false
				l.isTripleQuote = false
				return l.newToken(token.ILLEGAL, sb.String(), pos)
			}
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune(escaped)
			}
			continue
		}

		sb.WriteRune(l.advance())
	}

	l.inString = false
	l.isTripleQuote = false
	return l.newToken(token.ILLEGAL, sb.String(), pos)
}

// scanUnicode 扫描 Unicode 转义序列
func (l *Lexer) scanUnicode(length int) rune {
	var value rune
	for i := 0; i < length && !l.atEnd() && isHexDigit(l.current()); i++ {
		value = value<<4 | rune(hexValue(l.advance()))
	}
	return value
}

// scanOperator 扫描运算符
func (l *Lexer) scanOperator(pos token.Position) token.Token {
	ch := l.advance()

	switch ch {
	case '+':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.PLUS_ASSIGN, "+=", pos)
		}
		return l.newToken(token.PLUS, "+", pos)
	case '-':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.MINUS_ASSIGN, "-=", pos)
		}
		if l.current() == '>' {
			l.advance()
			return l.newToken(token.ARROW, "->", pos)
		}
		return l.newToken(token.MINUS, "-", pos)
	case '*':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.STAR_ASSIGN, "*=", pos)
		}
		return l.newToken(token.STAR, "*", pos)
	case '/':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.SLASH_ASSIGN, "/=", pos)
		}
		return l.newToken(token.SLASH, "/", pos)
	case '%':
		return l.newToken(token.PERCENT, "%", pos)
	case '&':
		if l.current() == '&' {
			l.advance()
			return l.newToken(token.AND, "&&", pos)
		}
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.AND_ASSIGN, "&=", pos)
		}
		return l.newToken(token.AMPERSAND, "&", pos)
	case '|':
		if l.current() == '>' {
			l.advance()
			return l.newToken(token.PIPE_FWD, "|>", pos)
		}
		if l.current() == '|' {
			l.advance()
			return l.newToken(token.OR, "||", pos)
		}
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.OR_ASSIGN, "|=", pos)
		}
		return l.newToken(token.PIPE, "|", pos)
	case '^':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.XOR_ASSIGN, "^=", pos)
		}
		return l.newToken(token.CARET, "^", pos)
	case '~':
		return l.newToken(token.TILDE, "~", pos)
	case '<':
		if l.current() == '|' {
			l.advance()
			return l.newToken(token.PIPE_BWD, "<|", pos)
		}
		if l.current() == '<' {
			l.advance()
			if l.current() == '=' {
				l.advance()
				return l.newToken(token.SHL_ASSIGN, "<<=", pos)
			}
			return l.newToken(token.SHIFTLEFT, "<<", pos)
		}
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.LTE, "<=", pos)
		}
		return l.newToken(token.LT, "<", pos)
	case '>':
		if l.current() == '>' {
			l.advance()
			if l.current() == '=' {
				l.advance()
				return l.newToken(token.SHR_ASSIGN, ">>=", pos)
			}
			return l.newToken(token.SHIFTRIGHT, ">>", pos)
		}
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.GTE, ">=", pos)
		}
		return l.newToken(token.GT, ">", pos)
	case '=':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.EQ, "==", pos)
		}
		if l.current() == '>' {
			l.advance()
			return l.newToken(token.ROCKET, "=>", pos)
		}
		if l.current() == '~' {
			l.advance()
			return l.newToken(token.MATCH_EQ, "=~", pos)
		}
		return l.newToken(token.ASSIGN, "=", pos)
	case '!':
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.NEQ, "!=", pos)
		}
		return l.newToken(token.NOT, "!", pos)
	case '.':
		if l.current() == '.' {
			l.advance()
			if l.current() == '.' {
				l.advance()
				return l.newToken(token.ELLIPSIS, "...", pos)
			}
			if l.current() == '=' {
				l.advance()
				return l.newToken(token.RANGE_INCLUSIVE, "..=", pos)
			}
			return l.newToken(token.CONCAT, "..", pos)
		}
		if l.current() == '=' {
			l.advance()
			return l.newToken(token.RANGE_INCLUSIVE, "..=", pos)
		}
		return l.newToken(token.DOT, ".", pos)
	case '?':
		if l.current() == '?' {
			l.advance()
			return l.newToken(token.NULL_COALESCING, "??", pos)
		}
		return l.newToken(token.QUESTION, "?", pos)
	case ':':
		return l.newToken(token.COLON, ":", pos)
	}

	return l.newToken(token.ILLEGAL, string(ch), pos)
}

// scanRegexLiteral 扫描正则字面量 #/pattern/flags#
//
// 格式：#/pattern/flags#
// - 模式内容中 \/ 表示字面量 /
// - 模式内容中 \# 表示字面量 #
// - 模式内容中 \\ 表示字面量 \
// - flags 可选：i (忽略大小写), m (多行), s (dot匹配换行), U (非贪婪)
// - 编译期验证正则合法性
func (l *Lexer) scanRegexLiteral(pos token.Position) token.Token {
	l.advance() // #
	l.advance() // /

	patternStart := l.pos

	// 扫描模式内容直到遇到未转义的 /
	for !l.atEnd() {
		ch := l.current()
		if ch == '\\' && !l.atEnd() {
			l.advance() // 转义符
			l.advance() // 转义的字符
			continue
		}
		if ch == '/' {
			break
		}
		if ch == '\n' || ch == '\r' {
			// 不支持跨行正则
			return l.newToken(token.ILLEGAL, "regex literal cannot contain newline", pos)
		}
		l.advance()
	}

	patternEnd := l.pos
	pattern := string(l.source[patternStart:patternEnd])

	// 检查空模式
	if pattern == "" {
		return l.newToken(token.ILLEGAL, "empty regex pattern", pos)
	}

	// 消费 /
	if l.atEnd() || l.current() != '/' {
		return l.newToken(token.ILLEGAL, "unterminated regex literal", pos)
	}
	l.advance()

	// 扫描 flags
	flagsStart := l.pos
	for !l.atEnd() {
		ch := l.current()
		if ch == 'i' || ch == 'm' || ch == 's' || ch == 'U' {
			l.advance()
		} else {
			break
		}
	}
	flags := string(l.source[flagsStart:l.pos])

	// 验证 flags（已在上面过滤，这里不需要额外检查）

	// 检查结尾 #
	if l.atEnd() || l.current() != '#' {
		return l.newToken(token.ILLEGAL, "unterminated regex literal, expected #", pos)
	}
	l.advance() // #

	// 处理转义后的模式用于验证
	unescapedPattern := unescapeRegexPattern(pattern)

	// 预编译验证（使用 unescaped pattern）
	if _, err := compileRegexForValidation(unescapedPattern, flags); err != nil {
		return l.newToken(token.ILLEGAL, fmt.Sprintf("invalid regex: %v", err), pos)
	}

	// 构造完整的 literal（包含原始 pattern 和 flags）
	literal := "#/" + pattern + "/" + flags + "#"
	return l.newToken(token.REGEX, literal, pos)
}

// unescapeRegexPattern 处理正则模式中的转义
// \/ -> /, \# -> #, \\ -> \ (其他转义保留给 Go regexp)
func unescapeRegexPattern(pattern string) string {
	var result strings.Builder
	result.Grow(len(pattern))
	i := 0
	for i < len(pattern) {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			switch next {
			case '/':
				result.WriteByte('/')
				i += 2
				continue
			case '#':
				result.WriteByte('#')
				i += 2
				continue
			case '\\':
				result.WriteByte('\\')
				i += 2
				continue
			}
		}
		result.WriteByte(pattern[i])
		i++
	}
	return result.String()
}

// compileRegexForValidation 编译正则表达式用于编译期验证
func compileRegexForValidation(pattern, flags string) (*regexp.Regexp, error) {
	goFlags := ""
	for _, f := range flags {
		switch f {
		case 'i':
			goFlags += "(?i)"
		case 'm':
			goFlags += "(?m)"
		case 's':
			goFlags += "(?s)"
		case 'U':
			goFlags += "(?U)"
		}
	}
	return regexp.Compile(goFlags + pattern)
}

// scanDelimiter 扫描分隔符
func (l *Lexer) scanDelimiter(pos token.Position) token.Token {
	ch := l.advance()

	switch ch {
	case ';':
		return l.newToken(token.SEMICOLON, ";", pos)
	case ',':
		return l.newToken(token.COMMA, ",", pos)
	case '(':
		return l.newToken(token.LPAREN, "(", pos)
	case ')':
		return l.newToken(token.RPAREN, ")", pos)
	case '{':
		return l.newToken(token.LBRACE, "{", pos)
	case '}':
		return l.newToken(token.RBRACE, "}", pos)
	case '[':
		return l.newToken(token.LBRACKET, "[", pos)
	case ']':
		return l.newToken(token.RBRACKET, "]", pos)
	}

	return l.newToken(token.ILLEGAL, string(ch), pos)
}

// 辅助函数

// isDigit 检查是否为数字
func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

// isHexDigit 检查是否为十六进制数字
func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// isOctalDigit 检查是否为八进制数字
func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

// isBinaryDigit 检查是否为二进制数字
func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
}

// hexValue 获取十六进制数字的值
func hexValue(ch rune) byte {
	if ch >= '0' && ch <= '9' {
		return byte(ch - '0')
	}
	if ch >= 'a' && ch <= 'f' {
		return byte(ch - 'a' + 10)
	}
	if ch >= 'A' && ch <= 'F' {
		return byte(ch - 'A' + 10)
	}
	return 0
}

// isIdentifierStart 检查是否为标识符起始字符
func isIdentifierStart(ch rune) bool {
	return unicode.IsLetter(ch)
}

// isIdentifierPart 检查是否为标识符组成部分
func isIdentifierPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// isOperator 检查是否为运算符字符
func isOperator(ch rune) bool {
	switch ch {
	case '+', '-', '*', '/', '%',
		'&', '|', '^', '~',
		'<', '>', '=',
		'!', '.', '?', ':':
		return true
	}
	return false
}

// isDelimiter 检查是否为分隔符
func isDelimiter(ch rune) bool {
	switch ch {
	case ';', ',', '(', ')', '{', '}', '[', ']':
		return true
	}
	return false
}

// isInt64Overflow 检查整数是否溢出 int64
func isInt64Overflow(s string) bool {
	// 处理十六进制、八进制、二进制
	if len(s) > 2 && s[0] == '0' {
		switch s[1] {
		case 'x', 'X':
			// 十六进制
			_, ok := new(big.Int).SetString(s[2:], 16)
			if !ok {
				return true
			}
			// 检查是否超过 int64 范围
			maxInt64 := big.NewInt(1<<63 - 1)
			minInt64 := big.NewInt(-1 << 63)
			val, _ := new(big.Int).SetString(s[2:], 16)
			return val.Cmp(maxInt64) > 0 || val.Cmp(minInt64) < 0
		case 'o', 'O':
			// 八进制
			_, ok := new(big.Int).SetString(s[2:], 8)
			if !ok {
				return true
			}
			maxInt64 := big.NewInt(1<<63 - 1)
			minInt64 := big.NewInt(-1 << 63)
			val, _ := new(big.Int).SetString(s[2:], 8)
			return val.Cmp(maxInt64) > 0 || val.Cmp(minInt64) < 0
		case 'b', 'B':
			// 二进制
			_, ok := new(big.Int).SetString(s[2:], 2)
			if !ok {
				return true
			}
			maxInt64 := big.NewInt(1<<63 - 1)
			minInt64 := big.NewInt(-1 << 63)
			val, _ := new(big.Int).SetString(s[2:], 2)
			return val.Cmp(maxInt64) > 0 || val.Cmp(minInt64) < 0
		}
	}

	// 普通十进制
	val, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return true
	}
	maxInt64 := big.NewInt(1<<63 - 1)
	minInt64 := big.NewInt(-1 << 63)
	return val.Cmp(maxInt64) > 0 || val.Cmp(minInt64) < 0
}

// ScanAll 扫描所有 Token 并返回
func (l *Lexer) ScanAll() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// Errors 返回扫描过程中的错误信息
func (l *Lexer) Errors() []string {
	var errors []string
	tokens := l.ScanAll()
	for _, tok := range tokens {
		if tok.Type == token.ILLEGAL {
			errors = append(errors, fmt.Sprintf("非法字符 '%s' 位于 %s:%d:%d",
				tok.Literal, tok.Pos.Filename, tok.Pos.Line, tok.Pos.Column))
		}
	}
	return errors
}
