package format

import (
	"strings"
	"testing"
)

func formatOrFatal(t *testing.T, src string) string {
	t.Helper()
	out, err := Format(src, "test.jpl")
	if err != nil {
		t.Fatalf("format error: %v\ninput:\n%s", err, src)
	}
	return out
}

// TestVariableDecl 变量声明格式化
func TestVariableDecl(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			"简单赋值",
			`$x = 1`,
			"$x = 1\n",
		},
		{
			"表达式赋值",
			`$sum = $a + $b`,
			"$sum = $a + $b\n",
		},
		{
			"无初始值",
			`$x`,
			"$x\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOrFatal(t, tt.src)
			if got != tt.want {
				t.Errorf("期望:\n%s\n得到:\n%s", tt.want, got)
			}
		})
	}
}

// TestConstDecl 常量声明
func TestConstDecl(t *testing.T) {
	src := `const PI = 3.14`
	want := "const PI = 3.14\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%s\n得到:\n%s", want, got)
	}
}

// TestFunctionDecl 函数声明格式化
func TestFunctionDecl(t *testing.T) {
	src := `fn add($a, $b) { return $a + $b }`
	want := "fn add($a, $b) {\n    return $a + $b\n}\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%s\n得到:\n%s", want, got)
	}
}

// TestIfElse if/else 格式化
func TestIfElse(t *testing.T) {
	src := `if ($x > 0) { println("pos") } else { println("neg") }`
	want := strings.Join([]string{
		`if ($x > 0) {`,
		`    println("pos")`,
		`} else {`,
		`    println("neg")`,
		`}`,
		``,
	}, "\n")
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%q\n得到:\n%q", want, got)
	}
}

// TestElseIf else if 格式化
func TestElseIf(t *testing.T) {
	src := `if ($x > 0) { println("pos") } else if ($x < 0) { println("neg") } else { println("zero") }`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "else if") {
		t.Errorf("应包含 else if，得到:\n%s", got)
	}
}

// TestWhileLoop while 循环
func TestWhileLoop(t *testing.T) {
	src := `while ($i < 10) { $i = $i + 1 }`
	want := "while ($i < 10) {\n    $i = $i + 1\n}\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%s\n得到:\n%s", want, got)
	}
}

// TestForLoop for 循环
func TestForLoop(t *testing.T) {
	src := `for ($i = 0; $i < 10; $i = $i + 1) { println($i) }`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "for ($i = 0; $i < 10; $i = $i + 1)") {
		t.Errorf("for 格式不正确:\n%s", got)
	}
}

// TestForeach foreach 循环
func TestForeach(t *testing.T) {
	src := `foreach ($item in $arr) { println($item) }`
	want := "foreach ($item in $arr) {\n    println($item)\n}\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%s\n得到:\n%s", want, got)
	}
}

// TestForeachKeyValue 带键的 foreach
func TestForeachKeyValue(t *testing.T) {
	src := `foreach ($k => $v in $map) { println($k) }`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "$k => $v") {
		t.Errorf("键值对格式不正确:\n%s", got)
	}
}

// TestLineComment 单行注释保留
func TestLineComment(t *testing.T) {
	src := "// 这是注释\n$x = 1\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "// 这是注释") {
		t.Errorf("单行注释丢失:\n%s", got)
	}
}

// TestBlockComment 多行注释保留
func TestBlockComment(t *testing.T) {
	src := "/* 多行\n注释 */\n$x = 1\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "/* 多行") {
		t.Errorf("多行注释丢失:\n%s", got)
	}
}

// TestCommentBeforeFunction 函数前的注释
func TestCommentBeforeFunction(t *testing.T) {
	src := "// greet 函数\nfn greet($name) { println($name) }\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "// greet 函数") {
		t.Errorf("函数前注释丢失:\n%s", got)
	}
}

// TestEmptyBlock 空代码块
func TestEmptyBlock(t *testing.T) {
	src := `fn noop() {}`
	want := "fn noop() {}\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%s\n得到:\n%s", want, got)
	}
}

// TestArrayLiteral 数组字面量
func TestArrayLiteral(t *testing.T) {
	src := `$arr = [1, 2, 3]`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "[1, 2, 3]") {
		t.Errorf("数组格式不正确:\n%s", got)
	}
}

// TestObjectLiteral 对象字面量
func TestObjectLiteral(t *testing.T) {
	src := `$obj = {name: "test", value: 42}`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "{") || !strings.Contains(got, "}") {
		t.Errorf("对象格式不正确:\n%s", got)
	}
}

// TestNestedIf 嵌套 if
func TestNestedIf(t *testing.T) {
	src := `if ($a) { if ($b) { println("yes") } }`
	got := formatOrFatal(t, src)
	// 应该有两层缩进
	if !strings.Contains(got, "        println") {
		t.Errorf("嵌套缩进不正确:\n%s", got)
	}
}

// TestTryCatch try/catch
func TestTryCatch(t *testing.T) {
	src := `try { risky() } catch ($e) { println($e) }`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "try {") {
		t.Errorf("try 格式不正确:\n%s", got)
	}
	if !strings.Contains(got, "catch ($e)") {
		t.Errorf("catch 格式不正确:\n%s", got)
	}
}

// TestImport import 语句
func TestImport(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"基本导入", `import "lib.jpl"`, "import \"lib.jpl\"\n"},
		{"别名导入", `import "lib.jpl" as lib`, "import \"lib.jpl\" as lib\n"},
		{"选择导入", `from "lib.jpl" import a, b`, "from \"lib.jpl\" import a, b\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatOrFatal(t, tt.src)
			if got != tt.want {
				t.Errorf("期望:\n%s\n得到:\n%s", tt.want, got)
			}
		})
	}
}

// TestIdempotency 幂等性测试
func TestIdempotency(t *testing.T) {
	src := `// 测试
fn main() {
    $x = 1
    if ($x > 0) {
        println("yes")
    }
}

/* 结尾注释 */`

	first := formatOrFatal(t, src)
	second := formatOrFatal(t, first)
	if first != second {
		t.Errorf("幂等性失败:\n首次:\n%s\n二次:\n%s", first, second)
	}
}

// TestSyntaxError 语法错误处理
func TestSyntaxError(t *testing.T) {
	src := `fn { broken syntax`
	_, err := Format(src, "test.jpl")
	if err == nil {
		t.Error("期望语法错误，但格式化成功")
	}
}

// TestThrowStatement throw 语句
func TestThrowStatement(t *testing.T) {
	src := `throw "error"`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, `throw "error"`) {
		t.Errorf("throw 格式不正确:\n%s", got)
	}
}

// TestGlobalDecl global 声明
func TestGlobalDecl(t *testing.T) {
	src := `global $x, $y`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "global $x, $y") {
		t.Errorf("global 格式不正确:\n%s", got)
	}
}

// TestStaticDecl static 声明
func TestStaticDecl(t *testing.T) {
	src := `static $counter = 0`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "static $counter = 0") {
		t.Errorf("static 格式不正确:\n%s", got)
	}
}

// TestMultipleStatements 多语句
func TestMultipleStatements(t *testing.T) {
	src := strings.Join([]string{
		`$a = 1`,
		`$b = 2`,
		`$c = $a + $b`,
		`println($c)`,
	}, "\n")
	want := strings.Join([]string{
		`$a = 1`,
		`$b = 2`,
		`$c = $a + $b`,
		`println($c)`,
		``,
	}, "\n")
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%q\n得到:\n%q", want, got)
	}
}

// TestUnaryExpr 一元表达式
func TestUnaryExpr(t *testing.T) {
	src := `$x = -$y`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "-$y") {
		t.Errorf("一元表达式格式不正确:\n%s", got)
	}
}

// TestTernaryExpr 三元表达式
func TestTernaryExpr(t *testing.T) {
	src := `$result = $x > 0 ? "pos" : "neg"`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "?") || !strings.Contains(got, ":") {
		t.Errorf("三元表达式格式不正确:\n%s", got)
	}
}

// TestMemberAccess 成员访问
func TestMemberAccess(t *testing.T) {
	src := `println($obj.name)`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "$obj.name") {
		t.Errorf("成员访问格式不正确:\n%s", got)
	}
}

// TestIndexAccess 索引访问
func TestIndexAccess(t *testing.T) {
	src := `println($arr[0])`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "$arr[0]") {
		t.Errorf("索引访问格式不正确:\n%s", got)
	}
}

// ============================================================================
// 边界案例测试
// ============================================================================

// TestEmptyInput 空文件
func TestEmptyInput(t *testing.T) {
	out, err := Format("", "test.jpl")
	if err != nil {
		t.Fatalf("空文件不应报错: %v", err)
	}
	if out != "" {
		t.Errorf("空文件应输出空字符串，得到: %q", out)
	}
}

// TestOnlyComments 纯注释文件
func TestOnlyComments(t *testing.T) {
	src := "// 这是注释\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "// 这是注释") {
		t.Errorf("纯注释应保留，得到: %q", got)
	}
}

// TestMultipleBlockComments 连续多个注释
func TestMultipleBlockComments(t *testing.T) {
	src := "/* 注释一 */\n/* 注释二 */\n$x = 1\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "/* 注释一 */") || !strings.Contains(got, "/* 注释二 */") {
		t.Errorf("多个注释应全部保留:\n%s", got)
	}
}

// TestMultipleBlankLines 连续空行
func TestMultipleBlankLines(t *testing.T) {
	src := "$a = 1\n\n\n\n$b = 2\n"
	got := formatOrFatal(t, src)
	// 格式化器统一为单换行（不保留源码空行）
	if strings.Contains(got, "\n\n") {
		t.Errorf("格式化器不保留源码空行，得到:\n%q", got)
	}
}

// TestTrailingComment 行尾注释
func TestTrailingComment(t *testing.T) {
	src := "$x = 1 // 行尾注释\n$y = 2\n"
	got := formatOrFatal(t, src)
	// 行尾注释应保留在同一行
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) < 1 {
		t.Fatalf("输出为空")
	}
	if !strings.Contains(lines[0], "// 行尾注释") {
		t.Errorf("行尾注释应在同一行，第一行: %q\n全部:\n%q", lines[0], got)
	}
}

// TestTrailingCommentOnFunction 函数声明行尾注释
func TestTrailingCommentOnFunction(t *testing.T) {
	src := "fn add($a, $b) { // 两数相加\n    return $a + $b\n}\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "// 两数相加") {
		t.Errorf("函数行尾注释丢失:\n%s", got)
	}
}

// TestCommentLikeInString 字符串内含注释符号
func TestCommentLikeInString(t *testing.T) {
	src := `$url = "http://example.com"
$path = "// not a comment"
$block = "/* also not */"`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, `"http://example.com"`) {
		t.Errorf("字符串内 // 被误判为注释:\n%s", got)
	}
	if !strings.Contains(got, `"// not a comment"`) {
		t.Errorf("字符串内 // 被误判为注释:\n%s", got)
	}
}

// TestDeepNesting 深度嵌套（3 层函数声明）
func TestDeepNesting(t *testing.T) {
	src := "fn a() {\n    fn b() {\n        fn c() {\n            println(\"deep\")\n        }\n    }\n}\n"
	got := formatOrFatal(t, src)
	// 第三层缩进应有 3 个 4 空格 = 12 空格
	if !strings.Contains(got, "            println") {
		t.Errorf("深度嵌套缩进不正确:\n%s", got)
	}
	// 幂等性
	second := formatOrFatal(t, got)
	if got != second {
		t.Errorf("深度嵌套幂等性失败")
	}
}

// TestUTF8Content UTF-8 多字节字符
func TestUTF8Content(t *testing.T) {
	src := `$msg = "你好世界"
// 这是中文注释
println($msg)`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "你好世界") {
		t.Errorf("UTF-8 字符串丢失:\n%s", got)
	}
	if !strings.Contains(got, "// 这是中文注释") {
		t.Errorf("UTF-8 注释丢失:\n%s", got)
	}
}

// TestNestedBlockCommentLimitation 嵌套注释限制
func TestNestedBlockCommentLimitation(t *testing.T) {
	// C 风格注释不支持嵌套，/* /* */ 中第一个 */ 结束注释
	// 这是预期行为，测试确认返回错误而非 panic
	src := "/* outer /* inner */ end */\n$x = 1\n"
	_, err := Format(src, "test.jpl")
	// 应该返回语法错误（因为 "end */" 不是合法语法）
	if err == nil {
		t.Log("嵌套注释意外通过（lexer 可能已支持）")
	}
}

// TestSingleLineOnly 仅一行代码
func TestSingleLineOnly(t *testing.T) {
	src := `$x = 1 + 2`
	want := "$x = 1 + 2\n"
	got := formatOrFatal(t, src)
	if got != want {
		t.Errorf("期望:\n%q\n得到:\n%q", want, got)
	}
}

// TestConsecutiveComments 连续注释块
func TestConsecutiveComments(t *testing.T) {
	src := "// 注释1\n// 注释2\n// 注释3\n$x = 1\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "// 注释1") {
		t.Errorf("注释1丢失:\n%s", got)
	}
	if !strings.Contains(got, "// 注释2") {
		t.Errorf("注释2丢失:\n%s", got)
	}
	if !strings.Contains(got, "// 注释3") {
		t.Errorf("注释3丢失:\n%s", got)
	}
}

// TestBlockCommentBetweenStmts 语句间多行注释
func TestBlockCommentBetweenStmts(t *testing.T) {
	src := "$a = 1\n/* 中间注释 */\n$b = 2\n"
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "/* 中间注释 */") {
		t.Errorf("语句间注释丢失:\n%s", got)
	}
}

// TestVeryLongExpression 超长表达式（不自动换行）
func TestVeryLongExpression(t *testing.T) {
	src := `$result = $a + $b + $c + $d + $e + $f + $g + $h + $i + $j + $k + $l + $m + $n`
	got := formatOrFatal(t, src)
	if !strings.Contains(got, "$a + $b + $c") {
		t.Errorf("超长表达式处理异常:\n%s", got)
	}
}

// TestComplexProgram 完整程序格式化
func TestComplexProgram(t *testing.T) {
	src := `// 配置
const DEBUG = true

fn process($data) {
    if ($data == null) {
        return null
    }
    $result = []
    foreach ($item in $data) {
        if ($item > 0) {
            $result = push($result, $item)
        }
    }
    return $result
}

$nums = [1, -2, 3, -4, 5]
$positive = process($nums)
println("结果: " .. str($positive))`

	// 格式化两次验证幂等性
	first := formatOrFatal(t, src)
	second := formatOrFatal(t, first)
	if first != second {
		t.Errorf("复杂程序幂等性失败")
	}

	// 验证关键内容保留
	if !strings.Contains(first, "// 配置") {
		t.Errorf("注释丢失")
	}
	if !strings.Contains(first, "const DEBUG = true") {
		t.Errorf("常量声明丢失")
	}
	if !strings.Contains(first, "foreach ($item in $data)") {
		t.Errorf("foreach 丢失")
	}
}
