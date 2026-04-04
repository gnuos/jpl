package stdlib

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterURL 注册 URL/URI 处理函数到引擎。
//
// 注册的函数：
//   - urlencode: 标准 URL 编码（空格 → +）
//   - urldecode: 标准 URL 解码（+ → 空格）
//   - rawurlencode: 原始 URL 编码（空格 → %20）
//   - rawurldecode: 原始 URL 解码（%20 → 空格）
//   - parse_url: 解析 URL 为组件对象
//
// 参数：
//   - e: 引擎实例
func RegisterURL(e *engine.Engine) {
	// 全局注册
	e.RegisterFunc("urlencode", builtinUrlencode)
	e.RegisterFunc("urldecode", builtinUrldecode)
	e.RegisterFunc("rawurlencode", builtinRawurlencode)
	e.RegisterFunc("rawurldecode", builtinRawurldecode)
	e.RegisterFunc("parse_url", builtinParseURL)

	// P1
	e.RegisterFunc("build_url", builtinBuildURL)
	e.RegisterFunc("parse_query", builtinParseQuery)

	// 模块注册 — import "url" 可用
	e.RegisterModule("url", map[string]engine.GoFunction{
		"urlencode":    builtinUrlencode,
		"urldecode":    builtinUrldecode,
		"rawurlencode": builtinRawurlencode,
		"rawurldecode": builtinRawurldecode,
		"parse_url":    builtinParseURL,
		// P1
		"build_url":   builtinBuildURL,
		"parse_query": builtinParseQuery,
	})
}

// UrlNames 返回 URL 处理函数名称列表。
//
// 返回值：
//   - []string: 函数名列表 ["urlencode", "urldecode", "rawurlencode", "rawurldecode", "parse_url"]
func UrlNames() []string {
	return []string{"urlencode", "urldecode", "rawurlencode", "rawurldecode", "parse_url", "build_url", "parse_query"}
}

// builtinUrlencode 标准 URL 编码。
//
// 将字符串编码为 URL 安全格式，空格编码为 +。
// 适用于查询字符串参数值的编码。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要编码的字符串
//
// 返回值：
//   - string: URL 编码后的字符串
//   - error: 参数错误
//
// 使用示例：
//
//	urlencode("hello world")   // → "hello+world"
//	urlencode("a=b&c=d")       // → "a%3Db%26c%3Dd"
func builtinUrlencode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("urlencode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("urlencode() argument must be a string, got %s", args[0].Type())
	}
	encoded := url.QueryEscape(args[0].String())
	return engine.NewString(encoded), nil
}

// builtinUrldecode 标准 URL 解码。
//
// 解码 URL 编码的字符串，+ 解码为空格。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要解码的 URL 编码字符串
//
// 返回值：
//   - string: 解码后的字符串
//   - error: 无效的编码序列
//
// 使用示例：
//
//	urldecode("hello+world")   // → "hello world"
//	urldecode("a%3Db")         // → "a=b"
func builtinUrldecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("urldecode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("urldecode() argument must be a string, got %s", args[0].Type())
	}
	decoded, err := url.QueryUnescape(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("urldecode() invalid input: %v", err)
	}
	return engine.NewString(decoded), nil
}

// builtinRawurlencode 原始 URL 编码。
//
// 将字符串编码为 URL 安全格式，空格编码为 %20。
// 适用于 URL 路径部分的编码。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要编码的字符串
//
// 返回值：
//   - string: URL 编码后的字符串
//   - error: 参数错误
//
// 使用示例：
//
//	rawurlencode("hello world")  // → "hello%20world"
//	rawurlencode("/path/to")     // → "%2Fpath%2Fto"
func builtinRawurlencode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rawurlencode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("rawurlencode() argument must be a string, got %s", args[0].Type())
	}
	// url.PathEscape 将空格编码为 %20，与 rawurlencode 行为一致
	encoded := url.PathEscape(args[0].String())
	return engine.NewString(encoded), nil
}

// builtinRawurldecode 原始 URL 解码。
//
// 解码 URL 编码的字符串，%20 解码为空格。
// 注意：+ 保持为字面量，不会解码为空格。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要解码的 URL 编码字符串
//
// 返回值：
//   - string: 解码后的字符串
//   - error: 无效的编码序列
//
// 使用示例：
//
//	rawurldecode("hello%20world")  // → "hello world"
//	rawurldecode("a+b")            // → "a+b"（+ 不解码）
func builtinRawurldecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rawurldecode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("rawurldecode() argument must be a string, got %s", args[0].Type())
	}
	decoded, err := url.PathUnescape(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("rawurldecode() invalid input: %v", err)
	}
	return engine.NewString(decoded), nil
}

// builtinParseURL 解析 URL 字符串，返回各组成部分的对象。
//
// 解析 URL 并返回包含各组件的对象。不存在的组件不会出现在对象中。
//
// 返回对象可能包含的字段：
//   - scheme: 协议（如 "http", "https"）
//   - host: 主机名
//   - port: 端口号（整数）
//   - user: 用户名
//   - pass: 密码
//   - path: 路径
//   - query: 查询字符串
//   - fragment: 片段标识符
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: URL 字符串（非空）
//
// 返回值：
//   - object: URL 组件对象
//   - error: 无效的 URL 或空字符串
//
// 使用示例：
//
//	parse_url("https://user:pass@example.com:8080/path?query=1#section")
//	// → {scheme: "https", host: "example.com", port: 8080, user: "user", pass: "pass", path: "/path", query: "query=1", fragment: "section"}
//
//	$parts = parse_url("http://example.com")
//	$parts["host"]  // → "example.com"
func builtinParseURL(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parse_url() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("parse_url() argument must be a string, got %s", args[0].Type())
	}

	urlStr := args[0].String()
	if urlStr == "" {
		return nil, fmt.Errorf("parse_url() expects a non-empty URL string")
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("parse_url() invalid URL: %v", err)
	}

	// 构建返回数组
	result := make(map[string]engine.Value)

	if parsed.Scheme != "" {
		result["scheme"] = engine.NewString(parsed.Scheme)
	}
	if parsed.Host != "" {
		// 分离 host 和 port
		host := parsed.Hostname()
		port := parsed.Port()
		if host != "" {
			result["host"] = engine.NewString(host)
		}
		if port != "" {
			// 转换 port 为整数
			if portNum, err := strconv.ParseInt(port, 10, 64); err == nil {
				result["port"] = engine.NewInt(portNum)
			}
		}
	}
	if parsed.User != nil {
		result["user"] = engine.NewString(parsed.User.Username())
		if password, ok := parsed.User.Password(); ok {
			result["pass"] = engine.NewString(password)
		}
	}
	if parsed.Path != "" {
		result["path"] = engine.NewString(parsed.Path)
	}
	if parsed.Query() != nil && parsed.RawQuery != "" {
		result["query"] = engine.NewString(parsed.RawQuery)
	}
	if parsed.Fragment != "" {
		result["fragment"] = engine.NewString(parsed.Fragment)
	}

	return engine.NewObject(result), nil
}

// UrlSigs returns function signatures for REPL :doc command.
func UrlSigs() map[string]string {
	return map[string]string{
		"urlencode":    "urlencode(str) → string  — URL encode (space → +)",
		"urldecode":    "urldecode(str) → string  — URL decode (+ → space)",
		"rawurlencode": "rawurlencode(str) → string  — Raw URL encode (space → %20)",
		"rawurldecode": "rawurldecode(str) → string  — Raw URL decode",
		"parse_url":    "parse_url(url) → object  — Parse URL into components",
		"build_url":    "build_url(parts) → string  — Build URL from parts object",
		"parse_query":  "parse_query(str) → object  — Parse query string to key-value object",
	}
}

// builtinBuildURL 从 parts 对象构建 URL。
func builtinBuildURL(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("build_url() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("build_url() argument must be object, got %s", args[0].Type())
	}
	obj := args[0].Object()
	scheme := ""
	if v, ok := obj["scheme"]; ok {
		scheme = v.String()
	}
	host := ""
	if v, ok := obj["host"]; ok {
		host = v.String()
	}
	port := ""
	if v, ok := obj["port"]; ok {
		port = v.String()
	}
	path := ""
	if v, ok := obj["path"]; ok {
		path = v.String()
	}
	query := ""
	if v, ok := obj["query"]; ok {
		query = v.String()
	}

	var result string
	if scheme != "" {
		result = scheme + "://"
	}
	result += host
	if port != "" {
		result += ":" + port
	}
	result += path
	if query != "" {
		result += "?" + query
	}
	return engine.NewString(result), nil
}

// builtinParseQuery 解析查询字符串为键值对对象。
func builtinParseQuery(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parse_query() expects 1 argument, got %d", len(args))
	}
	str := args[0].String()
	result := make(map[string]engine.Value)
	pairs := strings.Split(str, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		key := kv[0]
		val := ""
		if len(kv) == 2 {
			val = kv[1]
		}
		decodedKey, _ := url.QueryUnescape(key)
		decodedVal, _ := url.QueryUnescape(val)
		result[decodedKey] = engine.NewString(decodedVal)
	}
	return engine.NewObject(result), nil
}
