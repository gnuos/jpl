// Package token 定义 JPL 脚本语言的词法单元
package token

// TokenType Token 类型枚举
type TokenType int

const (
	// 特殊 Token
	ILLEGAL TokenType = iota // 非法字符
	EOF                      // 文件结束
	NEWLINE                  // 换行符

	// 注释（格式化专用）
	COMMENT       // 单行注释 // ...
	BLOCK_COMMENT // 多行注释 /* ... */

	// 标识符和字面量
	IDENTIFIER   // 标识符（$var 或 var）
	UNDERSCORE   // 下划线占位符（_）
	SPECIAL_VAR  // 特殊变量（$_）
	INSTANCE_VAR // 实例变量（@member，用于闭包内访问对象成员）
	INTEGER      // 整数字面量
	FLOAT        // 浮点数字面量
	BIGINT       // 大整数字面量
	BIGDECIMAL   // 高精度小数字面量
	STRING       // 字符串字面量

	// 多行字符串（Phase 10.1）
	TRIPLE_SINGLE // ''' 单引号多行字符串开始/结束
	TRIPLE_DOUBLE // """ 双引号多行字符串开始/结束
	STRING_FRAG   // 多行字符串片段（插值前的文本）

	// 字符串插值（Phase 10.2）
	INTERP_START // #{ 插值表达式开始（Ruby 风格）
	INTERP_END   // } 插值表达式结束

	// 运算符 - 算术
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %

	// 运算符 - 位运算
	AMPERSAND  // &
	PIPE       // |
	CARET      // ^
	TILDE      // ~
	SHIFTLEFT  // <<
	SHIFTRIGHT // >>

	// 运算符 - 比较
	EQ  // ==
	NEQ // !=
	LT  // <
	GT  // >
	LTE // <=
	GTE // >=

	// 运算符 - 逻辑
	AND // &&
	OR  // ||
	NOT // !

	// 运算符 - 赋值
	ASSIGN       // =
	PLUS_ASSIGN  // +=
	MINUS_ASSIGN // -=
	STAR_ASSIGN  // *=
	SLASH_ASSIGN // /=

	// 运算符 - 位运算组合赋值
	AND_ASSIGN // &=
	OR_ASSIGN  // |=
	XOR_ASSIGN // ^=
	SHL_ASSIGN // <<=
	SHR_ASSIGN // >>=

	// 运算符 - 字符串组合赋值
	CONCAT_ASSIGN // .=

	// 运算符 - 字符串连接
	CONCAT // ..

	// 运算符 - 范围
	ELLIPSIS        // ... 半开区间 [start, end)
	RANGE_INCLUSIVE // ..= 闭区间 [start, end]

	// 运算符 - 三元
	QUESTION // :
	COLON    // :

	// 运算符 - Lambda
	ARROW // ->

	// 运算符 - foreach 键值对
	ROCKET // =>

	// 运算符 - 管道
	PIPE_FWD // |> 管道前向运算（左到右）
	PIPE_BWD // <| 管道反向运算（右到左）

	// 运算符 - 正则匹配
	MATCH_EQ // =~ 正则匹配运算符

	// 字面量 - 正则
	REGEX // #/pattern/flags# 正则字面量

	// 分隔符
	SEMICOLON // ;
	COMMA     // ,
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	DOT       // .

	// 关键字
	IF           // if
	ELSE         // else
	WHILE        // while
	FOR          // for
	FOREACH      // foreach
	FUNCTION     // function, fn
	RETURN       // return
	BREAK        // break
	CONTINUE     // continue
	IMPORT       // import
	STATIC       // static
	GLOBAL       // global
	TRY          // try
	CATCH        // catch
	THROW        // throw
	MATCH        // match
	CASE         // case
	DEFAULT      // default
	CONST        // const
	IN           // in（用于 foreach）
	INCLUDE      // include
	INCLUDE_ONCE // include_once
	FROM         // from
	AS           // as
	WHEN         // when（用于条件捕获）

	// 保留字面量值名称
	KW_NULL  // null, NULL
	KW_TRUE  // true, TRUE, True
	KW_FALSE // false, FALSE, False

	// 类型关键字（用于类型转换语法）
	KW_INT    // int
	KW_FLOAT  // float
	KW_STRING // string
	KW_BOOL   // bool
)

// keywords 关键字映射表
var keywords = map[string]TokenType{
	// 关键字
	"if":           IF,
	"else":         ELSE,
	"while":        WHILE,
	"for":          FOR,
	"foreach":      FOREACH,
	"function":     FUNCTION,
	"fn":           FUNCTION,
	"return":       RETURN,
	"break":        BREAK,
	"continue":     CONTINUE,
	"import":       IMPORT,
	"static":       STATIC,
	"global":       GLOBAL,
	"try":          TRY,
	"catch":        CATCH,
	"throw":        THROW,
	"match":        MATCH,
	"case":         CASE,
	"default":      DEFAULT,
	"const":        CONST,
	"in":           IN,
	"include":      INCLUDE,
	"include_once": INCLUDE_ONCE,
	"from":         FROM,
	"as":           AS,
	"when":         WHEN,

	// 保留字面量值名称 - null
	"null": KW_NULL,
	"NULL": KW_NULL,

	// 保留字面量值名称 - true
	"true": KW_TRUE,
	"TRUE": KW_TRUE,
	"True": KW_TRUE,

	// 保留字面量值名称 - false
	"false": KW_FALSE,
	"FALSE": KW_FALSE,
	"False": KW_FALSE,

	// 类型关键字 - 用于类型转换语法
	"int":    KW_INT,
	"float":  KW_FLOAT,
	"string": KW_STRING,
	"bool":   KW_BOOL,
}

// Keyword 查询标识符是否为关键字
// 如果是关键字返回对应的 TokenType，否则返回 IDENTIFIER
func Keyword(name string) TokenType {
	if tok, ok := keywords[name]; ok {
		return tok
	}
	return IDENTIFIER
}

// Keywords 返回所有关键字名称列表
func Keywords() []string {
	names := make([]string, 0, len(keywords))
	for name := range keywords {
		names = append(names, name)
	}
	return names
}

// Position 位置信息结构
type Position struct {
	Filename string // 文件名
	Line     int    // 行号
	Column   int    // 列号
	Offset   int    // 字节偏移
}

// Token Token 结构体
type Token struct {
	Type    TokenType // Token 类型
	Literal string    // 字面值
	Pos     Position  // 位置信息
}

// String 返回 Token 类型的字符串表示
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case COMMENT:
		return "COMMENT"
	case BLOCK_COMMENT:
		return "BLOCK_COMMENT"
	case IDENTIFIER:
		return "IDENTIFIER"
	case UNDERSCORE:
		return "UNDERSCORE"
	case SPECIAL_VAR:
		return "SPECIAL_VAR"
	case INSTANCE_VAR:
		return "INSTANCE_VAR"
	case INTEGER:
		return "INTEGER"
	case FLOAT:
		return "FLOAT"
	case BIGINT:
		return "BIGINT"
	case BIGDECIMAL:
		return "BIGDECIMAL"
	case STRING:
		return "STRING"
	case TRIPLE_SINGLE:
		return "TRIPLE_SINGLE"
	case TRIPLE_DOUBLE:
		return "TRIPLE_DOUBLE"
	case STRING_FRAG:
		return "STRING_FRAG"
	case INTERP_START:
		return "INTERP_START"
	case INTERP_END:
		return "INTERP_END"
	case PLUS:
		return "PLUS"
	case MINUS:
		return "MINUS"
	case STAR:
		return "STAR"
	case SLASH:
		return "SLASH"
	case PERCENT:
		return "PERCENT"
	case AMPERSAND:
		return "AMPERSAND"
	case PIPE:
		return "PIPE"
	case CARET:
		return "CARET"
	case TILDE:
		return "TILDE"
	case SHIFTLEFT:
		return "SHIFTLEFT"
	case SHIFTRIGHT:
		return "SHIFTRIGHT"
	case EQ:
		return "EQ"
	case NEQ:
		return "NEQ"
	case LT:
		return "LT"
	case GT:
		return "GT"
	case LTE:
		return "LTE"
	case GTE:
		return "GTE"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case ASSIGN:
		return "ASSIGN"
	case PLUS_ASSIGN:
		return "PLUS_ASSIGN"
	case MINUS_ASSIGN:
		return "MINUS_ASSIGN"
	case STAR_ASSIGN:
		return "STAR_ASSIGN"
	case SLASH_ASSIGN:
		return "SLASH_ASSIGN"
	case AND_ASSIGN:
		return "AND_ASSIGN"
	case OR_ASSIGN:
		return "OR_ASSIGN"
	case XOR_ASSIGN:
		return "XOR_ASSIGN"
	case SHL_ASSIGN:
		return "SHL_ASSIGN"
	case SHR_ASSIGN:
		return "SHR_ASSIGN"
	case CONCAT_ASSIGN:
		return "CONCAT_ASSIGN"
	case CONCAT:
		return "CONCAT"
	case QUESTION:
		return "QUESTION"
	case COLON:
		return "COLON"
	case ARROW:
		return "ARROW"
	case ROCKET:
		return "ROCKET"
	case PIPE_FWD:
		return "PIPE_FWD"
	case PIPE_BWD:
		return "PIPE_BWD"
	case MATCH_EQ:
		return "MATCH_EQ"
	case REGEX:
		return "REGEX"
	case SEMICOLON:
		return "SEMICOLON"
	case COMMA:
		return "COMMA"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	case LBRACE:
		return "LBRACE"
	case RBRACE:
		return "RBRACE"
	case LBRACKET:
		return "LBRACKET"
	case RBRACKET:
		return "RBRACKET"
	case DOT:
		return "DOT"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case WHILE:
		return "WHILE"
	case FOR:
		return "FOR"
	case FOREACH:
		return "FOREACH"
	case FUNCTION:
		return "FUNCTION"
	case RETURN:
		return "RETURN"
	case BREAK:
		return "BREAK"
	case CONTINUE:
		return "CONTINUE"
	case IMPORT:
		return "IMPORT"
	case STATIC:
		return "STATIC"
	case GLOBAL:
		return "GLOBAL"
	case TRY:
		return "TRY"
	case CATCH:
		return "CATCH"
	case THROW:
		return "THROW"
	case MATCH:
		return "MATCH"
	case CASE:
		return "CASE"
	case DEFAULT:
		return "DEFAULT"
	case CONST:
		return "CONST"
	case IN:
		return "IN"
	case INCLUDE:
		return "INCLUDE"
	case INCLUDE_ONCE:
		return "INCLUDE_ONCE"
	case FROM:
		return "FROM"
	case AS:
		return "AS"
	case KW_NULL:
		return "KW_NULL"
	case KW_TRUE:
		return "KW_TRUE"
	case KW_FALSE:
		return "KW_FALSE"
	case KW_INT:
		return "KW_INT"
	case KW_FLOAT:
		return "KW_FLOAT"
	case KW_STRING:
		return "KW_STRING"
	case KW_BOOL:
		return "KW_BOOL"
	default:
		return "UNKNOWN"
	}
}
