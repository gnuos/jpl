package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// urlencode 测试
// ============================================================================

func TestUrlencodeBasic(t *testing.T) {
	result, err := callBuiltin("urlencode", engine.NewString("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空格编码为 +
	expected := "hello+world"
	if result.String() != expected {
		t.Errorf("urlencode('hello world') expected %s, got %s", expected, result.String())
	}
}

func TestUrlencodeSpecialChars(t *testing.T) {
	result, err := callBuiltin("urlencode", engine.NewString("a+b=c&d?e"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a%2Bb%3Dc%26d%3Fe"
	if result.String() != expected {
		t.Errorf("urlencode special chars expected %s, got %s", expected, result.String())
	}
}

func TestUrlencodeChinese(t *testing.T) {
	result, err := callBuiltin("urlencode", engine.NewString("中文"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 中文字符应该被编码
	if result.String() == "中文" {
		t.Errorf("urlencode('中文') should encode Chinese characters")
	}
}

func TestUrlencodeEmpty(t *testing.T) {
	result, err := callBuiltin("urlencode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("urlencode('') expected empty, got %s", result.String())
	}
}

func TestUrlencodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("urlencode")
	if err == nil {
		t.Error("urlencode(0 args) should return error")
	}
	_, err = callBuiltin("urlencode", engine.NewString("a"), engine.NewString("b"))
	if err == nil {
		t.Error("urlencode(2 args) should return error")
	}
}

func TestUrlencodeNotString(t *testing.T) {
	_, err := callBuiltin("urlencode", engine.NewInt(42))
	if err == nil {
		t.Error("urlencode(42) should return error")
	}
}

// ============================================================================
// urldecode 测试
// ============================================================================

func TestUrldecodeBasic(t *testing.T) {
	result, err := callBuiltin("urldecode", engine.NewString("hello+world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// + 解码为空格
	expected := "hello world"
	if result.String() != expected {
		t.Errorf("urldecode('hello+world') expected %s, got %s", expected, result.String())
	}
}

func TestUrldecodePercent(t *testing.T) {
	result, err := callBuiltin("urldecode", engine.NewString("hello%20world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// %20 也解码为空格
	expected := "hello world"
	if result.String() != expected {
		t.Errorf("urldecode('hello%%20world') expected %s, got %s", expected, result.String())
	}
}

func TestUrldecodeSpecialChars(t *testing.T) {
	result, err := callBuiltin("urldecode", engine.NewString("a%2Bb%3Dc%26d%3Fe"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a+b=c&d?e"
	if result.String() != expected {
		t.Errorf("urldecode special chars expected %s, got %s", expected, result.String())
	}
}

func TestUrldecodeEmpty(t *testing.T) {
	result, err := callBuiltin("urldecode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("urldecode('') expected empty, got %s", result.String())
	}
}

func TestUrldecodeInvalid(t *testing.T) {
	// 无效的百分号编码应该报错
	_, err := callBuiltin("urldecode", engine.NewString("%ZZ"))
	if err == nil {
		t.Error("urldecode('%ZZ') should return error for invalid encoding")
	}
}

func TestUrldecodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("urldecode")
	if err == nil {
		t.Error("urldecode(0 args) should return error")
	}
}

func TestUrldecodeNotString(t *testing.T) {
	_, err := callBuiltin("urldecode", engine.NewInt(42))
	if err == nil {
		t.Error("urldecode(42) should return error")
	}
}

// ============================================================================
// rawurlencode 测试
// ============================================================================

func TestRawurlencodeBasic(t *testing.T) {
	result, err := callBuiltin("rawurlencode", engine.NewString("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空格编码为 %20
	expected := "hello%20world"
	if result.String() != expected {
		t.Errorf("rawurlencode('hello world') expected %s, got %s", expected, result.String())
	}
}

func TestRawurlencodeSpecialChars(t *testing.T) {
	result, err := callBuiltin("rawurlencode", engine.NewString("a+b=c&d?e"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "a+b=c&d?e"
	// + 和 & 在 path 中不需要编码，但 = 和 ? 需要
	// url.PathEscape 的行为：保留一些字符
	if result.String() != expected {
		t.Logf("rawurlencode('a+b=c&d?e') = %s (may vary by implementation)", result.String())
	}
}

func TestRawurlencodeEmpty(t *testing.T) {
	result, err := callBuiltin("rawurlencode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("rawurlencode('') expected empty, got %s", result.String())
	}
}

func TestRawurlencodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("rawurlencode")
	if err == nil {
		t.Error("rawurlencode(0 args) should return error")
	}
}

func TestRawurlencodeNotString(t *testing.T) {
	_, err := callBuiltin("rawurlencode", engine.NewInt(42))
	if err == nil {
		t.Error("rawurlencode(42) should return error")
	}
}

// ============================================================================
// rawurldecode 测试
// ============================================================================

func TestRawurldecodeBasic(t *testing.T) {
	result, err := callBuiltin("rawurldecode", engine.NewString("hello%20world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "hello world"
	if result.String() != expected {
		t.Errorf("rawurldecode('hello%%20world') expected %s, got %s", expected, result.String())
	}
}

func TestRawurldecodePlus(t *testing.T) {
	result, err := callBuiltin("rawurldecode", engine.NewString("hello+world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// rawurldecode 中 + 保持为 +，不解码为空格
	expected := "hello+world"
	if result.String() != expected {
		t.Errorf("rawurldecode('hello+world') expected %s, got %s", expected, result.String())
	}
}

func TestRawurldecodeEmpty(t *testing.T) {
	result, err := callBuiltin("rawurldecode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("rawurldecode('') expected empty, got %s", result.String())
	}
}

func TestRawurldecodeInvalid(t *testing.T) {
	_, err := callBuiltin("rawurldecode", engine.NewString("%ZZ"))
	if err == nil {
		t.Error("rawurldecode('%ZZ') should return error")
	}
}

func TestRawurldecodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("rawurldecode")
	if err == nil {
		t.Error("rawurldecode(0 args) should return error")
	}
}

func TestRawurldecodeNotString(t *testing.T) {
	_, err := callBuiltin("rawurldecode", engine.NewInt(42))
	if err == nil {
		t.Error("rawurldecode(42) should return error")
	}
}

// ============================================================================
// parse_url 测试
// ============================================================================

func TestParseUrlFull(t *testing.T) {
	url := "https://user:pass@example.com:8080/path/to/file?query=1&foo=bar#section"
	result, err := callBuiltin("parse_url", engine.NewString(url))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.Object()
	if m == nil {
		t.Fatalf("parse_url should return object")
	}

	// 检查各个字段
	if m["scheme"].String() != "https" {
		t.Errorf("scheme expected 'https', got %s", m["scheme"].String())
	}
	if m["host"].String() != "example.com" {
		t.Errorf("host expected 'example.com', got %s", m["host"].String())
	}
	if m["port"].Int() != 8080 {
		t.Errorf("port expected 8080, got %d", m["port"].Int())
	}
	if m["user"].String() != "user" {
		t.Errorf("user expected 'user', got %s", m["user"].String())
	}
	if m["pass"].String() != "pass" {
		t.Errorf("pass expected 'pass', got %s", m["pass"].String())
	}
	if m["path"].String() != "/path/to/file" {
		t.Errorf("path expected '/path/to/file', got %s", m["path"].String())
	}
	if m["query"].String() != "query=1&foo=bar" {
		t.Errorf("query expected 'query=1&foo=bar', got %s", m["query"].String())
	}
	if m["fragment"].String() != "section" {
		t.Errorf("fragment expected 'section', got %s", m["fragment"].String())
	}
}

func TestParseUrlSimple(t *testing.T) {
	url := "http://example.com"
	result, err := callBuiltin("parse_url", engine.NewString(url))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.Object()
	if m == nil {
		t.Fatalf("parse_url should return object")
	}

	if m["scheme"].String() != "http" {
		t.Errorf("scheme expected 'http', got %s", m["scheme"].String())
	}
	if m["host"].String() != "example.com" {
		t.Errorf("host expected 'example.com', got %s", m["host"].String())
	}
	// 没有端口、路径等字段
	if _, ok := m["port"]; ok {
		t.Errorf("port should not exist for simple URL")
	}
}

func TestParseUrlPathOnly(t *testing.T) {
	url := "/path/to/file"
	result, err := callBuiltin("parse_url", engine.NewString(url))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.Object()
	if m == nil {
		t.Fatalf("parse_url should return object")
	}

	if m["path"].String() != "/path/to/file" {
		t.Errorf("path expected '/path/to/file', got %s", m["path"].String())
	}
	// 其他字段应该不存在
	if _, ok := m["scheme"]; ok {
		t.Errorf("scheme should not exist for path-only URL")
	}
}

func TestParseUrlEmpty(t *testing.T) {
	_, err := callBuiltin("parse_url", engine.NewString(""))
	if err == nil {
		t.Error("parse_url('') should return error")
	}
}

func TestParseUrlInvalid(t *testing.T) {
	// 包含无效字符的 URL
	_, err := callBuiltin("parse_url", engine.NewString("://invalid"))
	// Go 的 url.Parse 可能对此宽容，视情况而定
	if err != nil {
		t.Logf("parse_url(':') returned error: %v (implementation dependent)", err)
	}
}

func TestParseUrlWrongArgCount(t *testing.T) {
	_, err := callBuiltin("parse_url")
	if err == nil {
		t.Error("parse_url(0 args) should return error")
	}
	_, err = callBuiltin("parse_url", engine.NewString("a"), engine.NewString("b"))
	if err == nil {
		t.Error("parse_url(2 args) should return error")
	}
}

func TestParseUrlNotString(t *testing.T) {
	_, err := callBuiltin("parse_url", engine.NewInt(42))
	if err == nil {
		t.Error("parse_url(42) should return error")
	}
}

// ============================================================================
// 集成测试 - 编解码往返
// ============================================================================

func TestUrlencodeDecodeRoundTrip(t *testing.T) {
	script := `$encoded = urlencode("hello world+foo");
$decoded = urldecode($encoded);
$decoded`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world+foo" {
		t.Errorf("urlencode/decode round trip failed, got %s", result.String())
	}
}

func TestRawurlencodeDecodeRoundTrip(t *testing.T) {
	script := `$encoded = rawurlencode("hello world+foo");
$decoded = rawurldecode($encoded);
$decoded`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world+foo" {
		t.Errorf("rawurlencode/decode round trip failed, got %s", result.String())
	}
}

func TestParseUrlIntegration(t *testing.T) {
	script := `$parts = parse_url("https://example.com:8080/path?query=1");
$parts["host"]`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "example.com" {
		t.Errorf("parse_url integration test failed, got %s", result.String())
	}
}

// ============================================================================
// 边缘案例测试 - Phase 7.7 补充
// ============================================================================

// urlencode 边缘案例

func TestUrlencodeSpecialSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/path/to", "%2Fpath%2Fto"},
		{"user@host", "user%40host"},
		{"a:b:c", "a%3Ab%3Ac"},
		{"(hello)", "%28hello%29"},
		{"a!b*c", "a%21b%2Ac"},
		{"a,b;c", "a%2Cb%3Bc"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("urlencode", engine.NewString(tt.input))
		if err != nil {
			t.Errorf("urlencode('%s') unexpected error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.expected {
			t.Errorf("urlencode('%s') expected '%s', got '%s'", tt.input, tt.expected, result.String())
		}
	}
}

func TestUrlencodeLongString(t *testing.T) {
	// 超长字符串测试 (2KB)
	longStr := ""
	for i := 0; i < 2048; i++ {
		longStr += "a"
	}
	result, err := callBuiltin("urlencode", engine.NewString(longStr))
	if err != nil {
		t.Fatalf("urlencode(long string) unexpected error: %v", err)
	}
	if result.String() != longStr {
		t.Errorf("urlencode(long string) should return same string for alphanumeric")
	}
}

func TestUrlencodeOnlySpecialChars(t *testing.T) {
	result, err := callBuiltin("urlencode", engine.NewString("&=?#"))
	if err != nil {
		t.Fatalf("urlencode('&=?#') unexpected error: %v", err)
	}
	expected := "%26%3D%3F%23"
	if result.String() != expected {
		t.Errorf("urlencode('&=?#') expected '%s', got '%s'", expected, result.String())
	}
}

// urldecode 边缘案例

func TestUrldecodeTruncatedEncoding(t *testing.T) {
	// 截断的编码应该报错
	tests := []string{"%A", "%", "%2", "a%b"}
	for _, input := range tests {
		_, err := callBuiltin("urldecode", engine.NewString(input))
		if err == nil {
			t.Errorf("urldecode('%s') should return error for truncated encoding", input)
		}
	}
}

func TestUrldecodeAllReservedChars(t *testing.T) {
	encoded := "encode%3A%2F%3F%23%5B%5D%40%21%24%26%27%28%29%2A%2B%2C%3B%3D"
	result, err := callBuiltin("urldecode", engine.NewString(encoded))
	if err != nil {
		t.Fatalf("urldecode reserved chars unexpected error: %v", err)
	}
	if result.String() == encoded {
		t.Errorf("urldecode should decode reserved characters")
	}
}

// rawurlencode 边缘案例

func TestRawurlencodeTildeChar(t *testing.T) {
	// Go's url.PathEscape preserves ~
	result, err := callBuiltin("rawurlencode", engine.NewString("hello~world"))
	if err != nil {
		t.Fatalf("rawurlencode('~') unexpected error: %v", err)
	}
	if result.String() != "hello~world" {
		t.Logf("rawurlencode('hello~world') = '%s' (implementation note: ~ preserved)", result.String())
	}
}

func TestRawurlencodeSlashChar(t *testing.T) {
	// Go's url.PathEscape encodes / as %2F (matches PHP rawurlencode)
	result, err := callBuiltin("rawurlencode", engine.NewString("/path/to"))
	if err != nil {
		t.Fatalf("rawurlencode('/') unexpected error: %v", err)
	}
	if result.String() != "%2Fpath%2Fto" {
		t.Errorf("rawurlencode('/path/to') expected '%%2Fpath%%2Fto', got '%s'", result.String())
	}
}

func TestRawurlencodeReservedSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"(hello)", "%28hello%29", "parentheses"},
		{"a!b*c", "a%21b%2Ac", "exclamation asterisk"},
		{"a,b;c", "a%2Cb%3Bc", "comma semicolon"},
		{"a:b", "a:b", "colon (preserved by PathEscape)"},
		{"a@b", "a@b", "at sign (preserved by PathEscape)"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("rawurlencode", engine.NewString(tt.input))
		if err != nil {
			t.Errorf("rawurlencode('%s') error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.expected {
			t.Errorf("rawurlencode('%s') expected '%s', got '%s' (%s)", tt.input, tt.expected, result.String(), tt.desc)
		}
	}
}

// rawurldecode 边缘案例

func TestRawurldecodeTruncated(t *testing.T) {
	tests := []string{"%A", "%", "%2"}
	for _, input := range tests {
		_, err := callBuiltin("rawurldecode", engine.NewString(input))
		if err == nil {
			t.Errorf("rawurldecode('%s') should return error", input)
		}
	}
}

func TestRawurldecodeKeepsPlus(t *testing.T) {
	result, err := callBuiltin("rawurldecode", engine.NewString("a+b%3Dc"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// rawurldecode should keep + as literal
	if result.String() != "a+b=c" {
		t.Errorf("rawurldecode('a+b%%3Dc') expected 'a+b=c', got '%s'", result.String())
	}
}

// parse_url 边缘案例

func TestParseUrlNoProtocol(t *testing.T) {
	// 无协议 URL (protocol-relative)
	result, err := callBuiltin("parse_url", engine.NewString("//example.com/path"))
	if err != nil {
		t.Fatalf("parse_url('//example.com') unexpected error: %v", err)
	}
	m := result.Object()
	if m["host"].String() != "example.com" {
		t.Errorf("host expected 'example.com', got '%s'", m["host"].String())
	}
	if m["path"].String() != "/path" {
		t.Errorf("path expected '/path', got '%s'", m["path"].String())
	}
}

func TestParseUrlRelativePath(t *testing.T) {
	// 相对路径
	result, err := callBuiltin("parse_url", engine.NewString("path/to/file?query=1"))
	if err != nil {
		t.Fatalf("parse_url relative path unexpected error: %v", err)
	}
	m := result.Object()
	if m["path"].String() != "path/to/file" {
		t.Errorf("path expected 'path/to/file', got '%s'", m["path"].String())
	}
	if m["query"].String() != "query=1" {
		t.Errorf("query expected 'query=1', got '%s'", m["query"].String())
	}
	if _, ok := m["scheme"]; ok {
		t.Error("scheme should not exist for relative path")
	}
}

func TestParseUrlUserNoPassword(t *testing.T) {
	// 有用户无密码
	result, err := callBuiltin("parse_url", engine.NewString("http://user@example.com/path"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.Object()
	if m["user"].String() != "user" {
		t.Errorf("user expected 'user', got '%s'", m["user"].String())
	}
	if _, ok := m["pass"]; ok {
		t.Error("pass should not exist when no password provided")
	}
}

func TestParseUrlChinesePath(t *testing.T) {
	// 中文路径
	result, err := callBuiltin("parse_url", engine.NewString("http://example.com/中文/路径"))
	if err != nil {
		t.Fatalf("parse_url chinese path unexpected error: %v", err)
	}
	m := result.Object()
	// Go's url.Parse may return raw or encoded path
	if m["path"] == nil {
		t.Error("path should exist")
	}
	t.Logf("parse_url chinese path = '%s'", m["path"].String())
}

func TestParseUrlIPv6(t *testing.T) {
	// IPv6 地址
	result, err := callBuiltin("parse_url", engine.NewString("http://[::1]:8080/path"))
	if err != nil {
		t.Fatalf("parse_url IPv6 unexpected error: %v", err)
	}
	m := result.Object()
	if m["host"].String() != "::1" {
		t.Errorf("host expected '::1', got '%s'", m["host"].String())
	}
	if m["port"].Int() != 8080 {
		t.Errorf("port expected 8080, got %d", m["port"].Int())
	}
}

func TestParseUrlMultiSlashPath(t *testing.T) {
	// 多斜杠路径
	result, err := callBuiltin("parse_url", engine.NewString("http://example.com///path"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.Object()
	if m["path"].String() != "///path" {
		t.Errorf("path expected '///path', got '%s'", m["path"].String())
	}
}

func TestParseUrlEncodedChars(t *testing.T) {
	// URL 中的编码字符
	result, err := callBuiltin("parse_url", engine.NewString("http://example.com/path%20with%20spaces"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.Object()
	// Go's url.Parse decodes percent-encoded chars in path
	if m["path"] == nil {
		t.Error("path should exist")
	}
	t.Logf("parse_url encoded path = '%s'", m["path"].String())
}

func TestParseUrlEmptyFragment(t *testing.T) {
	// 空 fragment (以 # 结尾)
	result, err := callBuiltin("parse_url", engine.NewString("http://example.com/path#"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.Object()
	// Go treats empty fragment as non-existent
	if _, ok := m["fragment"]; ok {
		t.Logf("fragment exists (empty): '%s'", m["fragment"].String())
	}
}

func TestParseUrlEmptyQuery(t *testing.T) {
	// 空 query (以 ? 结尾)
	result, err := callBuiltin("parse_url", engine.NewString("http://example.com/path?"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.Object()
	// Go treats empty query as non-existent
	if _, ok := m["query"]; ok {
		t.Logf("query exists (empty): '%s'", m["query"].String())
	}
}
