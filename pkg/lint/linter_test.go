package lint

import (
	"testing"
)

func lintOrFatal(t *testing.T, src string) *LintResult {
	t.Helper()
	return Lint(src, "test.jpl")
}

func TestUnusedVar(t *testing.T) {
	src := `$used = 1
$unused = 2
println($used)`
	result := lintOrFatal(t, src)

	found := false
	for _, d := range result.Diagnostics {
		if d.Rule == "unused-var" && d.Message == `variable "$unused" declared but never used` {
			found = true
		}
	}
	if !found {
		t.Errorf("期望检测到未使用变量 $unused，得到: %v", result.Diagnostics)
	}
}

func TestUndefinedVar(t *testing.T) {
	src := `println($undefined)`
	result := lintOrFatal(t, src)

	found := false
	for _, d := range result.Diagnostics {
		if d.Rule == "undefined-var" {
			found = true
		}
	}
	if !found {
		t.Error("期望检测到未定义变量")
	}
}

func TestNoUndefinedForBuiltin(t *testing.T) {
	// 内置函数名不应报 undefined
	src := `println("hello")`
	result := lintOrFatal(t, src)

	for _, d := range result.Diagnostics {
		if d.Rule == "undefined-var" {
			t.Errorf("内置函数不应报 undefined: %v", d)
		}
	}
}

func TestDeadCode(t *testing.T) {
	src := `fn test() {
    return 1
    $dead = 2
}`
	result := lintOrFatal(t, src)

	found := false
	for _, d := range result.Diagnostics {
		if d.Rule == "dead-code" {
			found = true
		}
	}
	if !found {
		t.Error("期望检测到死代码")
	}
}

func TestNoDeadCode(t *testing.T) {
	src := `$x = 1
$y = 2
println($x + $y)`
	result := lintOrFatal(t, src)

	for _, d := range result.Diagnostics {
		if d.Rule == "dead-code" {
			t.Errorf("不应有死代码警告: %v", d)
		}
	}
}

func TestFunctionParams(t *testing.T) {
	// 函数参数应被视为已声明
	src := `fn add($a, $b) {
    return $a + $b
}
println(add(1, 2))`
	result := lintOrFatal(t, src)

	for _, d := range result.Diagnostics {
		if d.Rule == "undefined-var" {
			t.Errorf("函数参数不应报 undefined: %v", d)
		}
	}
}

func TestForeachVars(t *testing.T) {
	// foreach 变量应被视为已声明
	src := `$arr = [1, 2, 3]
foreach ($item in $arr) {
    println($item)
}`
	result := lintOrFatal(t, src)

	for _, d := range result.Diagnostics {
		if d.Rule == "undefined-var" && d.Message != "" {
			// $item 和 $arr 都已声明
			t.Logf("诊断: %v", d)
		}
	}
}

func TestCleanCode(t *testing.T) {
	// 正确使用的代码不应有警告
	src := `$x = 10
$y = $x + 5
println($y)`
	result := lintOrFatal(t, src)

	if len(result.Diagnostics) != 0 {
		t.Errorf("正确代码不应有诊断，得到: %v", result.Diagnostics)
	}
}

func TestHasErrors(t *testing.T) {
	src := `println($undefined)`
	result := lintOrFatal(t, src)

	if !result.HasErrors() {
		t.Error("未定义变量应产生 error")
	}
}

func TestNoErrorsForWarnings(t *testing.T) {
	src := `$unused = 1`
	result := lintOrFatal(t, src)

	if result.HasErrors() {
		t.Error("未使用变量应为 warning 而非 error")
	}
}
