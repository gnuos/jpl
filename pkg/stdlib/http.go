package stdlib

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gnuos/jpl/engine"
)

// =============================================================================
// HTTP Client 模块 - 实现高级 HTTP 客户端功能
// =============================================================================

// HTTPResponseValue 表示 HTTP 响应对象
//
// 响应对象的字段通过 Object() 方法暴露给 JPL 代码：
//   - status: HTTP 状态码（整数），如 200, 404, 500
//   - status_text: 完整状态文本（字符串），如 "200 OK"
//   - headers: 响应头对象（object），键值对形式
//   - body: 响应体内容（字符串）
//   - content_length: 内容长度（整数）
//   - time: 请求耗时（浮点数，单位秒）
//
// 使用示例：
//
//	$resp = http_get("https://api.example.com/data")
//	println $resp.status          // 输出: 200
//	println $resp.body            // 输出: 响应内容
//	println $resp.headers.date   // 输出: 响应头中的 Date
//
//	// 检查响应状态
//	if ($resp.status >= 200 && $resp.status < 300) {
//	    println "请求成功"
//	}
//
// 类型转换行为：
//   - Bool(): 状态码 2xx 返回 true，其他返回 false
//   - Int(): 返回状态码
//   - Float(): 返回状态码的浮点形式
//   - String(): 返回 "HTTPResponse(status status_text)"
type HTTPResponseValue struct {
	Status        int               // HTTP 状态码
	StatusText    string            // 状态文本
	Headers       map[string]string // 响应头
	Body          []byte            // 响应体（字节）
	ContentLength int64             // 内容长度
	Time          float64           // 请求耗时（秒）
}

// Type 返回类型标识
func (r *HTTPResponseValue) Type() engine.ValueType { return engine.TypeObject }

// IsNull 检查是否为 null
func (r *HTTPResponseValue) IsNull() bool { return r == nil }

// Bool 返回布尔值
func (r *HTTPResponseValue) Bool() bool { return r != nil && r.Status >= 200 && r.Status < 300 }

// Int 返回状态码
func (r *HTTPResponseValue) Int() int64 { return int64(r.Status) }

// Float 返回状态码的浮点形式
func (r *HTTPResponseValue) Float() float64 { return float64(r.Status) }

// String 返回字符串表示
func (r *HTTPResponseValue) String() string {
	if r == nil {
		return "null"
	}
	return fmt.Sprintf("HTTPResponse(%d %s)", r.Status, r.StatusText)
}

// Stringify 返回 JSON 序列化字符串
func (r *HTTPResponseValue) Stringify() string { return r.String() }

// Array 返回数组值
func (r *HTTPResponseValue) Array() []engine.Value { return nil }

// Object 返回响应信息的对象形式
func (r *HTTPResponseValue) Object() map[string]engine.Value {
	if r == nil {
		return map[string]engine.Value{}
	}

	headers := make(map[string]engine.Value)
	for k, v := range r.Headers {
		headers[k] = engine.NewString(v)
	}

	return map[string]engine.Value{
		"status":         engine.NewInt(int64(r.Status)),
		"status_text":    engine.NewString(r.StatusText),
		"headers":        engine.NewObject(headers),
		"body":           engine.NewString(string(r.Body)),
		"content_length": engine.NewInt(r.ContentLength),
		"time":           engine.NewFloat(r.Time),
	}
}

// Len 返回响应体的字节数
func (r *HTTPResponseValue) Len() int { return len(r.Body) }

// Equals 等于
func (r *HTTPResponseValue) Equals(v engine.Value) bool { return false }

// Less 小于
func (r *HTTPResponseValue) Less(v engine.Value) bool { return false }

// Greater 大于
func (r *HTTPResponseValue) Greater(v engine.Value) bool { return false }

// LessEqual 小于等于
func (r *HTTPResponseValue) LessEqual(v engine.Value) bool { return false }

// GreaterEqual 大于等于
func (r *HTTPResponseValue) GreaterEqual(v engine.Value) bool { return false }

// ToBigInt 转换为大整数
func (r *HTTPResponseValue) ToBigInt() engine.Value { return engine.NewInt(0) }

// ToBigDecimal 转换为大浮点数
func (r *HTTPResponseValue) ToBigDecimal() engine.Value { return engine.NewFloat(0) }

// Add 加法
func (r *HTTPResponseValue) Add(v engine.Value) engine.Value { return r }

// Sub 减法
func (r *HTTPResponseValue) Sub(v engine.Value) engine.Value { return r }

// Mul 乘法
func (r *HTTPResponseValue) Mul(v engine.Value) engine.Value { return r }

// Div 除法
func (r *HTTPResponseValue) Div(v engine.Value) engine.Value { return r }

// Mod 取模
func (r *HTTPResponseValue) Mod(v engine.Value) engine.Value { return r }

// Negate 取反
func (r *HTTPResponseValue) Negate() engine.Value { return r }

// RegisterHTTP 注册 HTTP 函数到引擎
func RegisterHTTP(e *engine.Engine) {
	// 简单请求函数
	e.RegisterFunc("http_get", builtinHTTPGet)
	e.RegisterFunc("http_post", builtinHTTPPost)
	e.RegisterFunc("http_put", builtinHTTPPut)
	e.RegisterFunc("http_delete", builtinHTTPDelete)
	e.RegisterFunc("http_head", builtinHTTPHead)
	e.RegisterFunc("http_patch", builtinHTTPPatch)

	// 通用请求函数
	e.RegisterFunc("http_request", builtinHTTPRequest)

	// 模块注册 - import "http" 可用
	e.RegisterModule("http", map[string]engine.GoFunction{
		"get":     builtinHTTPGet,
		"post":    builtinHTTPPost,
		"put":     builtinHTTPPut,
		"delete":  builtinHTTPDelete,
		"head":    builtinHTTPHead,
		"patch":   builtinHTTPPatch,
		"request": builtinHTTPRequest,
	})
}

// HTTPNames 返回 HTTP 函数名称列表
func HTTPNames() []string {
	return []string{
		"http_get", "http_post", "http_put",
		"http_delete", "http_head", "http_patch",
		"http_request",
	}
}

// parseHTTPOptions 解析 HTTP 选项参数
func parseHTTPOptions(args engine.Value) *httpOptions {
	options := &httpOptions{
		Timeout:         30,
		FollowRedirects: true,
		MaxRedirects:    10,
		VerifySSL:       true,
	}

	if args == nil || args.Type() != engine.TypeObject {
		return options
	}

	obj := args.Object()

	// headers
	if headers, ok := obj["headers"]; ok && headers.Type() == engine.TypeObject {
		options.Headers = make(map[string]string)
		for k, v := range headers.Object() {
			options.Headers[k] = v.String()
		}
	}

	// timeout
	if timeout, ok := obj["timeout"]; ok {
		if timeout.Type() == engine.TypeInt {
			options.Timeout = int(timeout.Int())
		} else if timeout.Type() == engine.TypeFloat {
			options.Timeout = int(timeout.Float())
		}
	}

	// follow_redirects
	if follow, ok := obj["follow_redirects"]; ok {
		options.FollowRedirects = follow.Bool()
	}

	// max_redirects
	if maxRedir, ok := obj["max_redirects"]; ok {
		if maxRedir.Type() == engine.TypeInt {
			options.MaxRedirects = int(maxRedir.Int())
		}
	}

	// verify_ssl
	if verify, ok := obj["verify_ssl"]; ok {
		options.VerifySSL = verify.Bool()
	}

	// proxy
	if proxy, ok := obj["proxy"]; ok && proxy.Type() == engine.TypeString {
		options.Proxy = proxy.String()
	}

	// body
	if body, ok := obj["body"]; ok && body.Type() == engine.TypeString {
		options.Body = []byte(body.String())
		options.ContentType = "text/plain"
	}

	// json
	if jsonData, ok := obj["json"]; ok {
		options.Body, _ = json.Marshal(jsonData.Object())
		options.ContentType = "application/json"
	}

	// form
	if formData, ok := obj["form"]; ok && formData.Type() == engine.TypeObject {
		data := url.Values{}
		for k, v := range formData.Object() {
			data.Set(k, v.String())
		}
		options.Body = []byte(data.Encode())
		options.ContentType = "application/x-www-form-urlencoded"
	}

	// auth (基本认证)
	if auth, ok := obj["auth"]; ok && auth.Type() == engine.TypeObject {
		authObj := auth.Object()
		if username, ok := authObj["username"]; ok {
			if password, ok := authObj["password"]; ok {
				authStr := username.String() + ":" + password.String()
				encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
				if options.Headers == nil {
					options.Headers = make(map[string]string)
				}
				options.Headers["Authorization"] = "Basic " + encodedAuth
			}
		}
	}

	return options
}

// httpOptions 表示 HTTP 请求选项
//
// 字段说明：
//   - Headers: 自定义请求头，键值对形式
//   - Timeout: 请求超时时间（秒），默认 30 秒
//   - FollowRedirects: 是否自动跟随重定向，默认 true
//   - MaxRedirects: 最大重定向次数，默认 10
//   - VerifySSL: 是否验证 SSL 证书，默认 true（生产环境应保持 true）
//   - Proxy: 代理服务器地址，如 "http://proxy:8080"
//   - Body: 请求体字节数组
//   - ContentType: 请求 Content-Type 头
type httpOptions struct {
	Headers         map[string]string
	Timeout         int
	FollowRedirects bool
	MaxRedirects    int
	VerifySSL       bool
	Proxy           string
	Body            []byte
	ContentType     string
}

// doHTTPRequest 执行 HTTP 请求
func doHTTPRequest(method, urlStr string, options *httpOptions) (*HTTPResponseValue, error) {
	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !options.VerifySSL,
	}

	// 创建传输层
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// 创建客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(options.Timeout) * time.Second,
	}

	// 设置重定向策略
	if !options.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 创建请求
	var bodyReader io.Reader
	if options.Body != nil {
		bodyReader = bytes.NewReader(options.Body)
	}

	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}

	// 设置 Content-Type
	if options.ContentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", options.ContentType)
	}

	// 设置 User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "JPL/1.0")
	}

	// 自动设置 Accept-Encoding，支持 gzip、deflate 和 brotli 压缩
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}

	// 执行请求
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 自动解压响应体（如果服务器返回压缩内容）
	contentEncoding := strings.ToLower(resp.Header.Get("Content-Encoding"))
	switch contentEncoding {
	case "gzip":
		gr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to decompress gzip response: %v", err)
		}
		defer gr.Close()
		body, _ = io.ReadAll(gr)
	case "deflate":
		fr := flate.NewReader(bytes.NewReader(body))
		defer fr.Close()
		body, _ = io.ReadAll(fr)
	case "br":
		br := brotli.NewReader(bytes.NewReader(body))
		body, _ = io.ReadAll(br)
	}

	// 构建响应头
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &HTTPResponseValue{
		Status:        resp.StatusCode,
		StatusText:    resp.Status,
		Headers:       headers,
		Body:          body,
		ContentLength: resp.ContentLength,
		Time:          time.Since(start).Seconds(),
	}, nil
}

// builtinHTTPGet 执行 GET 请求
// http_get(url, options?) → HTTPResponseValue
//
// 参数：
//   - args[0]: URL 地址（字符串）
//   - args[1]: 选项对象（可选）
//
// 示例：
//
//	$resp = http_get("https://api.example.com/users")
//	$users = $resp.json()
func builtinHTTPGet(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_get() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_get() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("GET", urlStr, options)
}

// builtinHTTPPost 执行 POST 请求
// http_post(url, options?) → HTTPResponseValue
//
// 参数：
//   - args[0]: URL 地址（字符串）
//   - args[1]: 选项对象（可选）
//   - body: 原始请求体
//   - json: JSON 对象（自动设置 Content-Type）
//   - form: Form 对象（自动设置 Content-Type）
//
// 示例：
//
//	$resp = http_post("https://api.example.com/users", {
//	    json: {name: "Alice"}
//	})
func builtinHTTPPost(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_post() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_post() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("POST", urlStr, options)
}

// builtinHTTPPut 执行 PUT 请求
func builtinHTTPPut(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_put() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_put() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("PUT", urlStr, options)
}

// builtinHTTPDelete 执行 DELETE 请求
func builtinHTTPDelete(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_delete() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_delete() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("DELETE", urlStr, options)
}

// builtinHTTPHead 执行 HEAD 请求
func builtinHTTPHead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_head() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_head() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("HEAD", urlStr, options)
}

// builtinHTTPPatch 执行 PATCH 请求
func builtinHTTPPatch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("http_patch() expects at least 1 argument, got %d", len(args))
	}

	urlStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_patch() expects string URL, got %s", args[0].Type())
	}

	var options *httpOptions
	if len(args) >= 2 {
		options = parseHTTPOptions(args[1])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest("PATCH", urlStr, options)
}

// builtinHTTPRequest 执行通用 HTTP 请求
//
// http_request(method, url, options?) → HTTPResponseValue
//
// 参数说明：
//   - method: HTTP 方法（字符串），支持 GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS 等
//   - url: 目标 URL 地址（字符串）
//   - options: 可选的配置对象，包含以下字段：
//   - headers: 请求头对象，如 {Content-Type: "application/json"}
//   - timeout: 超时秒数（整数），默认 30
//   - follow_redirects: 是否跟随重定向（布尔），默认 true
//   - max_redirects: 最大重定向次数（整数），默认 10
//   - verify_ssl: 是否验证 SSL 证书（布尔），默认 true
//   - proxy: 代理地址，如 "http://proxy:8080"
//   - body: 原始请求体（字符串）
//   - json: JSON 对象（自动设置 Content-Type 为 application/json）
//   - form: Form 对象（自动设置 Content-Type 为 application/x-www-form-urlencoded）
//   - auth: 认证对象，包含 username 和 password 字段
//
// 返回值：
//   - HTTPResponseValue: 响应对象，包含 status, headers, body 等字段
//
// 使用示例：
//
//	// 简单 GET 请求
//	$resp = http_request("GET", "https://api.example.com/users")
//
//	// 带自定义头的请求
//	$resp = http_request("GET", "https://api.example.com/users", {
//	    headers: {Authorization: "Bearer token123"}
//	})
//
//	// POST JSON 数据
//	$resp = http_request("POST", "https://api.example.com/users", {
//	    json: {name: "Alice", age: 30}
//	})
//
//	// POST Form 数据
//	$resp = http_request("POST", "https://example.com/form", {
//	    form: {username: "alice", password: "secret"}
//	})
//
//	// 基本认证
//	$resp = http_request("GET", "https://api.example.com/private", {
//	    auth: {username: "admin", password: "password"}
//	})
//
//	// 禁用 SSL 验证（仅用于测试）
//	$resp = http_request("GET", "https://self-signed.example.com", {
//	    verify_ssl: false
//	})
//
//	// 使用代理
//	$resp = http_request("GET", "https://example.com", {
//	    proxy: "http://proxy:8080"
//	})
func builtinHTTPRequest(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("http_request() expects at least 2 arguments, got %d", len(args))
	}

	method := strings.ToUpper(args[0].String())
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_request() expects string method, got %s", args[0].Type())
	}

	urlStr := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("http_request() expects string URL, got %s", args[1].Type())
	}

	var options *httpOptions
	if len(args) >= 3 {
		options = parseHTTPOptions(args[2])
	} else {
		options = &httpOptions{
			Timeout:         30,
			FollowRedirects: true,
			MaxRedirects:    10,
			VerifySSL:       true,
		}
	}

	return doHTTPRequest(method, urlStr, options)
}

// HTTPSigs returns function signatures for REPL :doc command.
func HTTPSigs() map[string]string {
	return map[string]string{
		"http_get":     "http_get(url, [options]) → HTTPResponse  — HTTP GET request",
		"http_post":    "http_post(url, [options]) → HTTPResponse  — HTTP POST request",
		"http_put":     "http_put(url, [options]) → HTTPResponse  — HTTP PUT request",
		"http_delete":  "http_delete(url, [options]) → HTTPResponse  — HTTP DELETE request",
		"http_head":    "http_head(url, [options]) → HTTPResponse  — HTTP HEAD request",
		"http_patch":   "http_patch(url, [options]) → HTTPResponse  — HTTP PATCH request",
		"http_request": "http_request(method, url, [options]) → HTTPResponse  — Generic HTTP request",
	}
}
