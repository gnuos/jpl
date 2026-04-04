package stdlib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/token"
)

// RegisterObjectParse 注册 parse_object() 函数
func RegisterObjectParse(e *engine.Engine) {
	e.RegisterFunc("parse_object", builtinParseObject)
}

// ObjectParseNames 返回 parse_object 函数名称列表
func ObjectParseNames() []string {
	return []string{"parse_object"}
}

// builtinParseObject 安全地解析对象字面量字符串
// parse_object(str) — 将 JPL 格式的对象字符串转换为对象值
// 只接受对象字面量语法，拒绝函数调用和表达式
//
// 参数：
//   - str: JPL 对象字面量字符串，如 "{a: 1, b: 2}" 或 "{name: 'John'}"
//
// 返回：
//   - 成功：返回解析后的对象值
//   - 失败：返回 null 和错误信息
//
// 示例：
//
//	parse_object("{a: 1, b: 2}")              // → {a: 1, b: 2}
//	parse_object("{name: 'John', age: 30}")   // → {name: "John", age: 30}
//	parse_object("{nested: {x: 1}}")        // → {nested: {x: 1}}
func builtinParseObject(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return engine.NewNull(), fmt.Errorf("parse_object() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewNull(), fmt.Errorf("parse_object() argument must be a string, got %s", args[0].Type())
	}

	code := args[0].String()
	code = strings.TrimSpace(code)

	// 安全检查：确保以 { 开头
	if !strings.HasPrefix(code, "{") {
		return engine.NewNull(), fmt.Errorf("parse_object() argument must be an object literal starting with '{'")
	}

	// 使用安全的对象解析器
	obj, err := safeParseObject(code)
	if err != nil {
		return engine.NewNull(), fmt.Errorf("parse_object() error: %v", err)
	}

	return obj, nil
}

// safeParseObject 安全地解析对象字面量
// 只允许对象字面量语法，拒绝其他表达式
type safeObjectParser struct {
	l      *lexer.Lexer
	cur    token.Token
	errors []string
}

func safeParseObject(code string) (engine.Value, error) {
	p := &safeObjectParser{
		l:      lexer.NewLexer(code, "<parse_object>"),
		errors: []string{},
	}

	// 获取第一个 token
	p.next()

	// 必须以大括号开头
	if p.cur.Type != token.LBRACE {
		return nil, fmt.Errorf("expected '{' at start of object literal, got %s", p.cur.Type)
	}

	// 解析对象
	obj := p.parseObjectLiteral()

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(p.errors, "; "))
	}

	return obj, nil
}

func (p *safeObjectParser) next() {
	p.cur = p.l.NextToken()
	// 跳过换行和注释
	for p.cur.Type == token.NEWLINE || p.cur.Type == token.ILLEGAL && p.cur.Literal == "" {
		p.cur = p.l.NextToken()
	}
}

func (p *safeObjectParser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

// parseObjectLiteral 解析 { key: value, ... }
func (p *safeObjectParser) parseObjectLiteral() engine.Value {
	p.next() // 消费 {

	objData := make(map[string]engine.Value)

	// 空对象 {}
	if p.cur.Type == token.RBRACE {
		p.next()
		return engine.NewObject(objData)
	}

	for {
		// 解析 key
		if p.cur.Type != token.IDENTIFIER && p.cur.Type != token.STRING {
			p.addError(fmt.Sprintf("expected identifier or string as key, got %s", p.cur.Type))
			return engine.NewNull()
		}

		key := p.cur.Literal
		if p.cur.Type == token.STRING {
			// 去掉字符串引号
			key = strings.Trim(key, `"'`)
		}
		p.next()

		// 期望 :
		if p.cur.Type != token.COLON {
			p.addError(fmt.Sprintf("expected ':' after key, got %s", p.cur.Type))
			return engine.NewNull()
		}
		p.next()

		// 解析 value（只允许安全的字面量）
		value := p.parseSafeValue()
		if value == nil {
			return engine.NewNull()
		}

		// 添加到对象
		objData[key] = value

		// 检查下一个 token
		if p.cur.Type == token.RBRACE {
			p.next()
			break
		}

		if p.cur.Type != token.COMMA {
			p.addError(fmt.Sprintf("expected ',' or '}', got %s", p.cur.Type))
			return engine.NewNull()
		}
		p.next() // 消费 ,

		// 检查是否尾随逗号
		if p.cur.Type == token.RBRACE {
			p.next()
			break
		}
	}

	return engine.NewObject(objData)
}

// parseSafeValue 解析安全的值（字面量、嵌套对象、数组）
// 拒绝：函数调用、变量、表达式、运算符
func (p *safeObjectParser) parseSafeValue() engine.Value {
	switch p.cur.Type {
	case token.INTEGER:
		val, _ := strconv.ParseInt(p.cur.Literal, 10, 64)
		p.next()
		return engine.NewInt(val)

	case token.FLOAT:
		val, _ := strconv.ParseFloat(p.cur.Literal, 64)
		p.next()
		return engine.NewFloat(val)

	case token.STRING:
		val := p.cur.Literal
		// 去掉引号
		if (strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`)) ||
			(strings.HasPrefix(val, `'`) && strings.HasSuffix(val, `'`)) {
			val = val[1 : len(val)-1]
		}
		p.next()
		return engine.NewString(val)

	case token.KW_TRUE:
		p.next()
		return engine.NewBool(true)

	case token.KW_FALSE:
		p.next()
		return engine.NewBool(false)

	case token.KW_NULL:
		p.next()
		return engine.NewNull()

	case token.LBRACE:
		// 嵌套对象
		return p.parseObjectLiteral()

	case token.LBRACKET:
		// 数组
		return p.parseArrayLiteral()

	default:
		p.addError(fmt.Sprintf("unsupported value type in object: %s (only literals allowed)", p.cur.Type))
		return nil
	}
}

// parseArrayLiteral 解析 [ value, ... ]
func (p *safeObjectParser) parseArrayLiteral() engine.Value {
	p.next() // 消费 [

	arrData := []engine.Value{}

	// 空数组 []
	if p.cur.Type == token.RBRACKET {
		p.next()
		return engine.NewArray(arrData)
	}

	for {
		value := p.parseSafeValue()
		if value == nil {
			return engine.NewNull()
		}

		// 添加到数组
		arrData = append(arrData, value)

		// 检查下一个 token
		if p.cur.Type == token.RBRACKET {
			p.next()
			break
		}

		if p.cur.Type != token.COMMA {
			p.addError(fmt.Sprintf("expected ',' or ']' in array, got %s", p.cur.Type))
			return engine.NewNull()
		}
		p.next() // 消费 ,

		// 检查是否尾随逗号
		if p.cur.Type == token.RBRACKET {
			p.next()
			break
		}
	}

	return engine.NewArray(arrData)
}

// 辅助函数：确保在 builtin.go 中被调用
func init() {
	// 这会在包初始化时注册函数名
	// 实际的注册在 RegisterObjectParse 中完成
}

// ObjectParseSigs returns function signatures for REPL :doc command.
func ObjectParseSigs() map[string]string {
	return map[string]string{
		"parse_ini":  "parse_ini(str) → object  — Parse INI format string",
		"parse_yaml": "parse_yaml(str) → object  — Parse YAML format string",
		"parse_toml": "parse_toml(str) → object  — Parse TOML format string",
	}
}
