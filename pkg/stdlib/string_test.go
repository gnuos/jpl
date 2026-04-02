package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestStringNames(t *testing.T) {
	names := StringNames()
	if len(names) != 60 {
		t.Errorf("expected 60 string function names, got %d", len(names))
	}
}

// ============================================================================
// strlen
// ============================================================================

func TestStrlen(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"hello", 5},
		{"", 0},
		{"你好世界", 4},
		{"a", 1},
	}
	for _, tt := range tests {
		result, err := callBuiltin("strlen", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("strlen(%q) error: %v", tt.input, err)
		}
		if result.Int() != tt.want {
			t.Errorf("strlen(%q) = %d, want %d", tt.input, result.Int(), tt.want)
		}
	}
}

func TestStrlenNotString(t *testing.T) {
	_, err := callBuiltin("strlen", engine.NewInt(42))
	if err == nil {
		t.Error("strlen(int) should return error")
	}
}

func TestStrlenArgCount(t *testing.T) {
	_, err := callBuiltin("strlen", engine.NewString("a"), engine.NewString("b"))
	if err == nil {
		t.Error("strlen(a, b) should return error")
	}
}

// ============================================================================
// substr
// ============================================================================

func TestSubstr(t *testing.T) {
	tests := []struct {
		s, want string
		start   int64
		length  int64
		hasLen  bool
	}{
		{"hello", "llo", 2, 0, false},
		{"hello", "ll", 2, 2, true},
		{"hello", "hello", 0, 5, true},
		{"hello", "", 5, 0, false},
		{"hello", "o", -1, 0, false},
		{"hello", "el", -4, 2, true},
		{"hello", "", 10, 0, false},
		{"hello", "hello", -10, 0, false},
		{"", "", 0, 0, false},
	}
	for _, tt := range tests {
		var result engine.Value
		var err error
		if tt.hasLen {
			result, err = callBuiltin("substr", engine.NewString(tt.s), engine.NewInt(tt.start), engine.NewInt(tt.length))
		} else {
			result, err = callBuiltin("substr", engine.NewString(tt.s), engine.NewInt(tt.start))
		}
		if err != nil {
			t.Fatalf("substr(%q, %d) error: %v", tt.s, tt.start, err)
		}
		if result.String() != tt.want {
			t.Errorf("substr(%q, %d) = %q, want %q", tt.s, tt.start, result.String(), tt.want)
		}
	}
}

func TestSubstrUnicode(t *testing.T) {
	result, err := callBuiltin("substr", engine.NewString("你好世界"), engine.NewInt(1), engine.NewInt(2))
	if err != nil {
		t.Fatalf("substr error: %v", err)
	}
	if result.String() != "好世" {
		t.Errorf("substr('你好世界', 1, 2) = %q, want '好世'", result.String())
	}
}

// ============================================================================
// strpos
// ============================================================================

func TestStrpos(t *testing.T) {
	tests := []struct {
		s, needle string
		want      int64
	}{
		{"hello world", "world", 6},
		{"hello world", "hello", 0},
		{"hello world", "xyz", -1},
		{"hello", "", 0},
		{"", "a", -1},
		{"aaa", "a", 0},
	}
	for _, tt := range tests {
		result, err := callBuiltin("strpos", engine.NewString(tt.s), engine.NewString(tt.needle))
		if err != nil {
			t.Fatalf("strpos(%q, %q) error: %v", tt.s, tt.needle, err)
		}
		if result.Int() != tt.want {
			t.Errorf("strpos(%q, %q) = %d, want %d", tt.s, tt.needle, result.Int(), tt.want)
		}
	}
}

// ============================================================================
// str_replace
// ============================================================================

func TestStrReplace(t *testing.T) {
	tests := []struct {
		s, search, replace, want string
	}{
		{"hello world", "world", "go", "hello go"},
		{"aaa", "a", "b", "bbb"},
		{"hello", "xyz", "abc", "hello"},
		{"hello", "", "x", "hello"},
		{"abcabc", "bc", "X", "aXaX"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("str_replace", engine.NewString(tt.s), engine.NewString(tt.search), engine.NewString(tt.replace))
		if err != nil {
			t.Fatalf("str_replace error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("str_replace(%q, %q, %q) = %q, want %q", tt.s, tt.search, tt.replace, result.String(), tt.want)
		}
	}
}

// ============================================================================
// trim
// ============================================================================

func TestTrim(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"  \t\n", ""},
		{"", ""},
		{"\t\n hello world \r\n", "hello world"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("trim", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("trim error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("trim(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

// ============================================================================
// toUpper / toLower
// ============================================================================

func TestToUpper(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "HELLO"},
		{"Hello World", "HELLO WORLD"},
		{"ABC", "ABC"},
		{"", ""},
		{"123", "123"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("toUpper", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("toUpper error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("toUpper(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"abc", "abc"},
		{"", ""},
		{"123", "123"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("toLower", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("toLower error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("toLower(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

// ============================================================================
// split / join
// ============================================================================

func TestSplit(t *testing.T) {
	result, err := callBuiltin("split", engine.NewString("a,b,c"), engine.NewString(","))
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("split returned %d elements, want 3", len(arr))
	}
	expected := []string{"a", "b", "c"}
	for i, v := range arr {
		if v.String() != expected[i] {
			t.Errorf("split()[%d] = %q, want %q", i, v.String(), expected[i])
		}
	}
}

func TestSplitEmpty(t *testing.T) {
	result, err := callBuiltin("split", engine.NewString(""), engine.NewString(","))
	if err != nil {
		t.Fatalf("split error: %v", err)
	}
	arr := result.Array()
	if len(arr) != 1 {
		t.Errorf("split('') returned %d elements, want 1", len(arr))
	}
}

func TestJoin(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewString("a"),
		engine.NewString("b"),
		engine.NewString("c"),
	})
	result, err := callBuiltin("join", arr, engine.NewString(", "))
	if err != nil {
		t.Fatalf("join error: %v", err)
	}
	if result.String() != "a, b, c" {
		t.Errorf("join() = %q, want 'a, b, c'", result.String())
	}
}

func TestJoinEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})
	result, err := callBuiltin("join", arr, engine.NewString(","))
	if err != nil {
		t.Fatalf("join error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("join([]) = %q, want ''", result.String())
	}
}

// ============================================================================
// startsWith / endsWith
// ============================================================================

func TestStartsWith(t *testing.T) {
	tests := []struct {
		s, prefix string
		want      bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
		{"hello", "", true},
		{"", "a", false},
		{"", "", true},
	}
	for _, tt := range tests {
		result, err := callBuiltin("startsWith", engine.NewString(tt.s), engine.NewString(tt.prefix))
		if err != nil {
			t.Fatalf("startsWith error: %v", err)
		}
		if result.Bool() != tt.want {
			t.Errorf("startsWith(%q, %q) = %v, want %v", tt.s, tt.prefix, result.Bool(), tt.want)
		}
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		s, suffix string
		want      bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", false},
		{"hello", "", true},
		{"", "a", false},
		{"", "", true},
	}
	for _, tt := range tests {
		result, err := callBuiltin("endsWith", engine.NewString(tt.s), engine.NewString(tt.suffix))
		if err != nil {
			t.Fatalf("endsWith error: %v", err)
		}
		if result.Bool() != tt.want {
			t.Errorf("endsWith(%q, %q) = %v, want %v", tt.s, tt.suffix, result.Bool(), tt.want)
		}
	}
}

// ============================================================================
// charAt
// ============================================================================

func TestCharAt(t *testing.T) {
	tests := []struct {
		s    string
		idx  int64
		want string
	}{
		{"hello", 0, "h"},
		{"hello", 4, "o"},
		{"hello", 5, ""},
		{"hello", -1, ""},
		{"", 0, ""},
		{"你好", 1, "好"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("charAt", engine.NewString(tt.s), engine.NewInt(tt.idx))
		if err != nil {
			t.Fatalf("charAt error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("charAt(%q, %d) = %q, want %q", tt.s, tt.idx, result.String(), tt.want)
		}
	}
}

// ============================================================================
// repeat
// ============================================================================

func TestRepeat(t *testing.T) {
	tests := []struct {
		s     string
		count int64
		want  string
	}{
		{"ab", 3, "ababab"},
		{"x", 0, ""},
		{"hello", 1, "hello"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		result, err := callBuiltin("repeat", engine.NewString(tt.s), engine.NewInt(tt.count))
		if err != nil {
			t.Fatalf("repeat error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("repeat(%q, %d) = %q, want %q", tt.s, tt.count, result.String(), tt.want)
		}
	}
}

func TestRepeatNegative(t *testing.T) {
	_, err := callBuiltin("repeat", engine.NewString("a"), engine.NewInt(-1))
	if err == nil {
		t.Error("repeat() with negative count should return error")
	}
}

func TestRepeatTooLarge(t *testing.T) {
	_, err := callBuiltin("repeat", engine.NewString("a"), engine.NewInt(100001))
	if err == nil {
		t.Error("repeat() with excessive count should return error")
	}
}

// ============================================================================
// reverse
// ============================================================================

func TestReverse(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"你好", "好你"},
	}
	for _, tt := range tests {
		result, err := callBuiltin("reverse", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("reverse error: %v", err)
		}
		if result.String() != tt.want {
			t.Errorf("reverse(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

// ============================================================================
// 转义编码函数测试
// ============================================================================

func TestAddslashes(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{`It's a test`, `It\'s a test`},
		{`Say "Hello"`, `Say \"Hello\"`},
		{`C:\Users\name`, `C:\\Users\\name`},
		{`normal text`, `normal text`},
		{``, ``},
		{`\`, `\\`},
	}

	for _, tt := range tests {
		result, err := builtinAddslashes(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("addslashes(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("addslashes(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestStripslashes(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{`It\'s a test`, `It's a test`},
		{`Say \"Hello\"`, `Say "Hello"`},
		{`C:\\Users\\name`, `C:\Users\name`},
		{`normal text`, `normal text`},
		{``, ``},
	}

	for _, tt := range tests {
		result, err := builtinStripslashes(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("stripslashes(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("stripslashes(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestAddslashesStripslashesRoundTrip(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"Hello World",
		"It's a test",
		`Say "Hello"`,
		"Test\nNewline",
		"Special chars: !@#$%",
	}

	for _, original := range tests {
		escaped, err := builtinAddslashes(ctx, []engine.Value{engine.NewString(original)})
		if err != nil {
			t.Errorf("addslashes(%q) error: %v", original, err)
			continue
		}

		unescaped, err := builtinStripslashes(ctx, []engine.Value{escaped})
		if err != nil {
			t.Errorf("stripslashes() error: %v", err)
			continue
		}

		if unescaped.String() != original {
			t.Errorf("Round trip failed: original=%q, unescaped=%q", original, unescaped.String())
		}
	}
}

func TestAddcslashes(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		str      string
		charlist string
		want     string
	}{
		{"Hello\nWorld", "\n", "Hello\\nWorld"},
		{"Hello\tWorld", "\t", "Hello\\tWorld"},
		{"Test\r\n", "\r", "Test\\r\n"},
		{"no escape needed", "z", "no escape needed"},
		{"all chars", "a..z", "\\a\\l\\l \\c\\h\\a\\r\\s"},
	}

	for _, tt := range tests {
		result, err := builtinAddcslashes(ctx, []engine.Value{
			engine.NewString(tt.str),
			engine.NewString(tt.charlist),
		})
		if err != nil {
			t.Errorf("addcslashes(%q, %q) error: %v", tt.str, tt.charlist, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("addcslashes(%q, %q) = %q, want %q", tt.str, tt.charlist, result.String(), tt.want)
		}
	}
}

// ============================================================================
// hex2bin 测试
// ============================================================================

func TestHex2bin(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		hex      string
		expected string
	}{
		{"48656c6c6f", "Hello"}, // "Hello" 的 hex
		{"48656c6c6f20576f726c64", "Hello World"},
		{"", ""},                  // 空字符串
		{"00", "\x00"},            // NUL 字符
		{"0x48656c6c6f", "Hello"}, // 带 0x 前缀
		{"0X48656c6c6f", "Hello"}, // 带 0X 前缀
	}

	for _, tt := range tests {
		result, err := builtinHex2bin(ctx, []engine.Value{engine.NewString(tt.hex)})
		if err != nil {
			t.Errorf("hex2bin(%q) error: %v", tt.hex, err)
			continue
		}
		if result.String() != tt.expected {
			t.Errorf("hex2bin(%q) = %q, expected %q", tt.hex, result.String(), tt.expected)
		}
	}
}

func TestHex2binInvalid(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"486",   // 奇数长度
		"zz",    // 无效字符
		"hello", // 非十六进制字符
		"xyz",   // 无效字符
	}

	for _, tt := range tests {
		result, err := builtinHex2bin(ctx, []engine.Value{engine.NewString(tt)})
		if err != nil {
			t.Errorf("hex2bin(%q) error: %v", tt, err)
			continue
		}
		if result.Type() != engine.TypeNull {
			t.Errorf("hex2bin(%q) should return null for invalid input, got %v", tt, result.Type())
		}
	}
}

func TestBin2hexHex2binRoundTrip(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"Hello World",
		"Test123!@#",
		"Binary\x00Data",
		"Unicode: 你好",
	}

	for _, original := range tests {
		// bin2hex
		hex, err := builtinBin2hex(ctx, []engine.Value{engine.NewString(original)})
		if err != nil {
			t.Errorf("bin2hex(%q) error: %v", original, err)
			continue
		}

		// hex2bin
		result, err := builtinHex2bin(ctx, []engine.Value{hex})
		if err != nil {
			t.Errorf("hex2bin() error: %v", err)
			continue
		}

		if result.String() != original {
			t.Errorf("Round trip failed: original=%q, result=%q", original, result.String())
		}
	}
}

// ============================================================================
// Email encoding (quoted-printable) tests
// ============================================================================

func TestQuotedPrintableEncode(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "Hello=20World"},
		{"Hello World!", "Hello=20World!"},
		{"Hello=World", "Hello=3DWorld"},
		{"Hello\tWorld", "Hello=09World"},
		{"Hello\nWorld", "Hello=0AWorld"},
		{"\x00", "=00"},
		{"=", "=3D"},
		{"?", "=3F"},
		{" ", "=20"},
		{"\t", "=09"},
		{"\n", "=0A"},
		{"\r", "=0D"},
		{string([]byte{127}), "=7F"},
		{string([]byte{255}), "=FF"},
	}

	for _, tt := range tests {
		result, err := builtinQuotedPrintableEncode(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("quoted_printable_encode(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("quoted_printable_encode(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestQuotedPrintableDecode(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "Hello World"},
		{"Hello=20World", "Hello World"},
		{"Hello=3DWorld", "Hello=World"},
		{"Hello\tWorld", "Hello\tWorld"},
		{"Hello\nWorld", "Hello\nWorld"},
		{"=00", "\x00"},
		{"=3D", "="},
		{"=3F", "?"},
		{"=20", " "},
		{"=09", "\t"},
		{"=0A", "\n"},
		{"=0D", "\r"},
		{"=7F", string([]byte{127})},
		{"=FF", string([]byte{255})},
		{"=0a", "\n"}, // lowercase hex
		{"=0A=0B", "\n\x0B"},
	}

	for _, tt := range tests {
		result, err := builtinQuotedPrintableDecode(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("quoted_printable_decode(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("quoted_printable_decode(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestQuotedPrintableInvalid(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"=",   // incomplete
		"=1",  // incomplete (only one hex)
		"=1Z", // invalid hex (second char not hex)
		"=ZZ", // invalid hex (both chars not hex)
	}

	for _, tt := range tests {
		result, err := builtinQuotedPrintableDecode(ctx, []engine.Value{engine.NewString(tt)})
		if err != nil {
			t.Errorf("quoted_printable_decode(%q) error: %v", tt, err)
			continue
		}
		// Should return the input unchanged for invalid sequences
		if result.String() != tt {
			t.Errorf("quoted_printable_decode(%q) = %q, want %q (unchanged)", tt, result.String(), tt)
		}
	}
}

func TestQuotedPrintableRoundTrip(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"Hello World",
		"Hello=World",
		"Test\twith\ttabs",
		"Test\nwith\nnewlines",
		"Special chars: !@#$%^&*()",
		"",
		"\x00\x01\x02\xFF",
	}

	for _, original := range tests {
		encoded, err := builtinQuotedPrintableEncode(ctx, []engine.Value{engine.NewString(original)})
		if err != nil {
			t.Errorf("quoted_printable_encode(%q) error: %v", original, err)
			continue
		}

		decoded, err := builtinQuotedPrintableDecode(ctx, []engine.Value{encoded})
		if err != nil {
			t.Errorf("quoted_printable_decode() error: %v", err)
			continue
		}

		if decoded.String() != original {
			t.Errorf("Round trip failed: original=%q, encoded=%q, decoded=%q", original, encoded.String(), decoded.String())
		}
	}
}

// ============================================================================
// HTML entity encoding tests
// ============================================================================

func TestHtmlEntities(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{"<script>", "&lt;script&gt;"},
		{">", "&gt;"},
		{"&", "&amp;"},
		{`"`, "&quot;"},
		{"'", "&#039;"},
		{"Hello World", "Hello World"},
		{"a<b>c", "a&lt;b&gt;c"},
		{"a&b", "a&amp;b"},
		{`a"b`, `a&quot;b`},
		{"a'b", "a&#039;b"},
		{"<a&\"'>", "&lt;a&amp;&quot;&#039;&gt;"},
	}

	for _, tt := range tests {
		result, err := builtinHtmlEntities(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("htmlentities(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("htmlentities(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestHtmlEntityDecode(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		input string
		want  string
	}{
		{"&lt;script&gt;", "<script>"},
		{"&gt;", ">"},
		{"&amp;", "&"},
		{"&quot;", `"`},
		{"&#039;", "'"},
		{"Hello World", "Hello World"},
		{"a&lt;b&gt;c", "a<b>c"},
		{"a&amp;b", "a&b"},
		{`a&quot;b`, `a"b`},
		{"a&#039;b", "a'b"},
		{"&lt;a&amp;&quot;&#039;&gt;", "<a&\"'>"},
	}

	for _, tt := range tests {
		result, err := builtinHtmlEntityDecode(ctx, []engine.Value{engine.NewString(tt.input)})
		if err != nil {
			t.Errorf("html_entity_decode(%q) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("html_entity_decode(%q) = %q, want %q", tt.input, result.String(), tt.want)
		}
	}
}

func TestGetHtmlTranslationTable(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)
	ctx := engine.NewContext(e, nil)

	result, err := builtinGetHtmlTranslationTable(ctx, []engine.Value{})
	if err != nil {
		t.Fatalf("get_html_translation_table() error: %v", err)
	}

	obj := result.Object()
	expected := map[string]string{
		"<":  "&lt;",
		">":  "&gt;",
		"&":  "&amp;",
		"\"": "&quot;",
		"'":  "&#039;",
	}

	for k, v := range expected {
		if obj[k] == nil {
			t.Errorf("get_html_translation_table() missing key %q", k)
			continue
		}
		if obj[k].String() != v {
			t.Errorf("get_html_translation_table()[%q] = %q, want %q", k, obj[k].String(), v)
		}
	}
	// Check no extra keys
	if len(obj) != len(expected) {
		t.Errorf("get_html_translation_table() expected %d keys, got %d", len(expected), len(obj))
	}
}

// TestLtrim 测试 ltrim 函数
func TestLtrim(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		name     string
		str      string
		chars    string
		expected string
		useChars bool
	}{
		{"default whitespace", "  hello  ", "", "hello  ", false},
		{"tab and space", "\t\thello", "", "hello", false},
		{"custom chars", "xxxyhello", "xy", "hello", true},
		{"empty string", "", "", "", false},
		{"no leading chars", "hello  ", "", "hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			args := []engine.Value{engine.NewString(tt.str)}
			if tt.useChars {
				args = append(args, engine.NewString(tt.chars))
			}
			result, err := builtinLtrim(ctx, args)
			if err != nil {
				t.Fatalf("ltrim() error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("ltrim() = %q, expected %q", result.String(), tt.expected)
			}
		})
	}
}

// TestRtrim 测试 rtrim 函数
func TestRtrim(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		name     string
		str      string
		chars    string
		expected string
		useChars bool
	}{
		{"default whitespace", "  hello  ", "", "  hello", false},
		{"trailing newline", "hello\n\n", "", "hello", false},
		{"custom chars", "helloxxx", "x", "hello", true},
		{"empty string", "", "", "", false},
		{"no trailing chars", "  hello", "", "  hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			args := []engine.Value{engine.NewString(tt.str)}
			if tt.useChars {
				args = append(args, engine.NewString(tt.chars))
			}
			result, err := builtinRtrim(ctx, args)
			if err != nil {
				t.Fatalf("rtrim() error: %v", err)
			}
			if result.String() != tt.expected {
				t.Errorf("rtrim() = %q, expected %q", result.String(), tt.expected)
			}
		})
	}
}

// TestStrcmp 测试 strcmp 函数
func TestStrcmp(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		s1       string
		s2       string
		expected int64
	}{
		{"abc", "abc", 0},
		{"abc", "def", -1},
		{"def", "abc", 1},
		{"a", "b", -1},
		{"hello", "world", -1},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinStrcmp(ctx, []engine.Value{engine.NewString(tt.s1), engine.NewString(tt.s2)})
		if err != nil {
			t.Fatalf("strcmp() error: %v", err)
		}
		sign := result.Int()
		if (tt.expected < 0 && sign >= 0) || (tt.expected > 0 && sign <= 0) || (tt.expected == 0 && sign != 0) {
			t.Errorf("strcmp(%q, %q) = %d, expected sign %d", tt.s1, tt.s2, sign, tt.expected)
		}
	}
}

// TestStrcasecmp 测试 strcasecmp 函数（不区分大小写）
func TestStrcasecmp(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		s1       string
		s2       string
		expected int64
	}{
		{"ABC", "abc", 0},
		{"Hello", "hello", 0},
		{"abc", "DEF", -1},
		{"DEF", "abc", 1},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinStrcasecmp(ctx, []engine.Value{engine.NewString(tt.s1), engine.NewString(tt.s2)})
		if err != nil {
			t.Fatalf("strcasecmp() error: %v", err)
		}
		sign := result.Int()
		if (tt.expected < 0 && sign >= 0) || (tt.expected > 0 && sign <= 0) || (tt.expected == 0 && sign != 0) {
			t.Errorf("strcasecmp(%q, %q) = %d, expected sign %d", tt.s1, tt.s2, sign, tt.expected)
		}
	}
}

// TestStrncmp 测试 strncmp 函数
func TestStrncmp(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		s1       string
		s2       string
		n        int64
		expected int64
	}{
		{"abc", "abc", 3, 0},
		{"abc", "abd", 2, 0}, // 前2个相同
		{"abc", "abd", 3, -1},
		{"hello", "world", 5, -1},
		{"abc", "abcdef", 3, 0},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinStrncmp(ctx, []engine.Value{
			engine.NewString(tt.s1),
			engine.NewString(tt.s2),
			engine.NewInt(tt.n),
		})
		if err != nil {
			t.Fatalf("strncmp() error: %v", err)
		}
		sign := result.Int()
		if (tt.expected < 0 && sign >= 0) || (tt.expected > 0 && sign <= 0) || (tt.expected == 0 && sign != 0) {
			t.Errorf("strncmp(%q, %q, %d) = %d, expected sign %d", tt.s1, tt.s2, tt.n, sign, tt.expected)
		}
	}
}

// TestStripos 测试 stripos 函数（不区分大小写查找）
func TestStripos(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		haystack  string
		needle    string
		offset    int64
		useOffset bool
		expected  any // int64 for position, bool for false
	}{
		{"Hello World", "world", 0, false, int64(6)},
		{"Hello World", "WORLD", 0, false, int64(6)},
		{"abc abc", "ABC", 0, false, int64(0)},
		{"abc abc", "ABC", 1, true, int64(4)},
		{"hello", "xyz", 0, false, false},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		args := []engine.Value{engine.NewString(tt.haystack), engine.NewString(tt.needle)}
		if tt.useOffset {
			args = append(args, engine.NewInt(tt.offset))
		}
		result, err := builtinStripos(ctx, args)
		if err != nil {
			t.Fatalf("stripos() error: %v", err)
		}
		if expected, ok := tt.expected.(int64); ok {
			if result.Int() != expected {
				t.Errorf("stripos(%q, %q) = %d, expected %d", tt.haystack, tt.needle, result.Int(), expected)
			}
		} else {
			if result.Type() != engine.TypeBool || result.Bool() != false {
				t.Errorf("stripos(%q, %q) should return false", tt.haystack, tt.needle)
			}
		}
	}
}

// TestStrrpos 测试 strrpos 函数（反向查找）
func TestStrrpos(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		haystack string
		needle   string
		expected any // int64 or false
	}{
		{"abc abc abc", "abc", int64(8)},
		{"hello world", "o", int64(7)},
		{"hello", "xyz", false},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinStrrpos(ctx, []engine.Value{
			engine.NewString(tt.haystack),
			engine.NewString(tt.needle),
		})
		if err != nil {
			t.Fatalf("strrpos() error: %v", err)
		}
		if expected, ok := tt.expected.(int64); ok {
			if result.Int() != expected {
				t.Errorf("strrpos(%q, %q) = %d, expected %d", tt.haystack, tt.needle, result.Int(), expected)
			}
		} else {
			if result.Type() != engine.TypeBool || result.Bool() != false {
				t.Errorf("strrpos(%q, %q) should return false", tt.haystack, tt.needle)
			}
		}
	}
}

// TestStrstr 测试 strstr 函数
func TestStrstr(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		haystack string
		needle   string
		expected any // string or false
	}{
		{"user@example.com", "@", "@example.com"},
		{"hello world", "world", "world"},
		{"hello", "xyz", false},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinStrstr(ctx, []engine.Value{
			engine.NewString(tt.haystack),
			engine.NewString(tt.needle),
		})
		if err != nil {
			t.Fatalf("strstr() error: %v", err)
		}
		if expected, ok := tt.expected.(string); ok {
			if result.String() != expected {
				t.Errorf("strstr(%q, %q) = %q, expected %q", tt.haystack, tt.needle, result.String(), expected)
			}
		} else {
			if result.Type() != engine.TypeBool || result.Bool() != false {
				t.Errorf("strstr(%q, %q) should return false", tt.haystack, tt.needle)
			}
		}
	}
}

// TestSprintf 测试 sprintf 函数
func TestSprintf(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		format   string
		args     []engine.Value
		expected string
	}{
		{"Hello %s", []engine.Value{engine.NewString("World")}, "Hello World"},
		{"Number: %d", []engine.Value{engine.NewInt(42)}, "Number: 42"},
		{"Float: %f", []engine.Value{engine.NewFloat(3.14)}, "Float: 3.140000"},
		{"%%", []engine.Value{}, "%"},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		args := append([]engine.Value{engine.NewString(tt.format)}, tt.args...)
		result, err := builtinSprintf(ctx, args)
		if err != nil {
			t.Fatalf("sprintf() error: %v", err)
		}
		if result.String() != tt.expected {
			t.Errorf("sprintf(%q) = %q, expected %q", tt.format, result.String(), tt.expected)
		}
	}
}

// TestOrd 测试 ord 函数
func TestOrd(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		str      string
		expected int64
	}{
		{"A", 65},
		{"a", 97},
		{"0", 48},
		{"", 0},
		{"ABC", 65}, // 只取第一个字符
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinOrd(ctx, []engine.Value{engine.NewString(tt.str)})
		if err != nil {
			t.Fatalf("ord() error: %v", err)
		}
		if result.Int() != tt.expected {
			t.Errorf("ord(%q) = %d, expected %d", tt.str, result.Int(), tt.expected)
		}
	}
}

// TestChr 测试 chr 函数
func TestChr(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		ascii    int64
		expected string
	}{
		{65, "A"},
		{97, "a"},
		{48, "0"},
		{255, string(byte(255))},
		{256, ""}, // 超出范围
		{-1, ""},  // 负数
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinChr(ctx, []engine.Value{engine.NewInt(tt.ascii)})
		if err != nil {
			t.Fatalf("chr() error: %v", err)
		}
		if result.String() != tt.expected {
			t.Errorf("chr(%d) = %q, expected %q", tt.ascii, result.String(), tt.expected)
		}
	}
}

// TestNl2br 测试 nl2br 函数
func TestNl2br(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		str      string
		isXHTML  bool
		useXHTML bool
		expected string
	}{
		{"hello\nworld", true, false, "hello<br />\nworld"},
		{"hello\r\nworld", true, false, "hello<br />\r\nworld"},
		{"hello\nworld", false, true, "hello<br>\nworld"},
		{"no newline", true, false, "no newline"},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		args := []engine.Value{engine.NewString(tt.str)}
		if tt.useXHTML {
			args = append(args, engine.NewBool(tt.isXHTML))
		}
		result, err := builtinNl2br(ctx, args)
		if err != nil {
			t.Fatalf("nl2br() error: %v", err)
		}
		if result.String() != tt.expected {
			t.Errorf("nl2br(%q) = %q, expected %q", tt.str, result.String(), tt.expected)
		}
	}
}

// TestBin2hex 测试 bin2hex 函数
func TestBin2hex(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	tests := []struct {
		str      string
		expected string
	}{
		{"A", "41"},
		{"hello", "68656c6c6f"},
		{"", ""},
		{"\x00\x01\x02", "000102"},
	}

	for _, tt := range tests {
		ctx := engine.NewContext(e, nil)
		result, err := builtinBin2hex(ctx, []engine.Value{engine.NewString(tt.str)})
		if err != nil {
			t.Fatalf("bin2hex() error: %v", err)
		}
		if result.String() != tt.expected {
			t.Errorf("bin2hex(%q) = %q, expected %q", tt.str, result.String(), tt.expected)
		}
	}
}

// TestStringAliases 测试字符串函数别名
func TestStringAliases(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterString(e)

	aliases := []string{"implode", "explode", "chop"}
	for _, alias := range aliases {
		fn := e.GetRegisteredFunc(alias)
		if fn == nil {
			t.Errorf("%s should be registered as alias", alias)
		}
	}
}
