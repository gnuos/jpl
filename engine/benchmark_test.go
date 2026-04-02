package engine

import (
	"fmt"
	"testing"
)

// ============================================================================
// 编译基准测试
// ============================================================================

func BenchmarkCompileSimple(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		_, _ = CompileString("$a = 1 + 2")
	}
}

func BenchmarkCompileFunction(b *testing.B) {
	script := `
		fn add($x, $y) { return $x + $y }
		$a = add(1, 2)
	`
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		_, _ = CompileString(script)
	}
}

func BenchmarkCompileComplex(b *testing.B) {
	script := `
		fn fib($n) {
			if ($n <= 1) { return $n }
			return fib($n - 1) + fib($n - 2)
		}
		$result = fib(10)
		for ($i = 0; $i < 100; $i = $i + 1) {
			$result = $result + $i
		}
	`
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		_, _ = CompileString(script)
	}
}

// ============================================================================
// 算术运算基准测试
// ============================================================================

func BenchmarkArithmeticAdd(b *testing.B) {
	script := "$r = 0; for ($i = 0; $i < N; $i = $i + 1) { $r = $r + 1 }"
	script = fmt.Sprintf("$r = 0; for ($i = 0; $i < %d; $i = $i + 1) { $r = $r + 1 }", b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkArithmeticMixed(b *testing.B) {
	script := fmt.Sprintf(`
		$r = 1
		for ($i = 1; $i <= %d; $i = $i + 1) {
			$r = $r + $i * 2 - 1
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

// ============================================================================
// 控制流基准测试
// ============================================================================

func BenchmarkLoopWhile(b *testing.B) {
	script := fmt.Sprintf(`
		$i = 0
		while ($i < %d) {
			$i = $i + 1
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkLoopFor(b *testing.B) {
	script := fmt.Sprintf(`
		for ($i = 0; $i < %d; $i = $i + 1) {
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkConditionalBranching(b *testing.B) {
	script := fmt.Sprintf(`
		$r = 0
		for ($i = 0; $i < %d; $i = $i + 1) {
			if ($i %% 2 == 0) {
				$r = $r + 1
			} else {
				$r = $r - 1
			}
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

// ============================================================================
// 函数调用基准测试
// ============================================================================

func BenchmarkFunctionCall(b *testing.B) {
	script := fmt.Sprintf(`
		fn inc($x) { return $x + 1 }
		$r = 0
		for ($i = 0; $i < %d; $i = $i + 1) {
			$r = inc($r)
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkRecursionFib(b *testing.B) {
	// fib(20) = 6765, 递归深度 ~20
	script := `
		fn fib($n) {
			if ($n <= 1) { return $n }
			return fib($n - 1) + fib($n - 2)
		}
	`
	// 预编译
	prog, err := CompileString(script + "fib(20)")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		eng := NewEngine()
		vm := newVMWithProgram(eng, prog)
		if err := vm.Execute(); err != nil {
			b.Fatal(err)
		}
		eng.Close()
	}
}

// ============================================================================
// 闭包基准测试
// ============================================================================

func BenchmarkClosure(b *testing.B) {
	script := fmt.Sprintf(`
		fn makeCounter() {
			$count = 0
			return fn() {
				$count = $count + 1
				return $count
			}
		}
		$counter = makeCounter()
		for ($i = 0; $i < %d; $i = $i + 1) {
			$counter()
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

// ============================================================================
// 字符串操作基准测试
// ============================================================================

func BenchmarkStringConcat(b *testing.B) {
	script := fmt.Sprintf(`
		$s = ""
		for ($i = 0; $i < %d; $i = $i + 1) {
			$s = $s + "a"
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkStringFunctions(b *testing.B) {
	script := fmt.Sprintf(`
		for ($i = 0; $i < %d; $i = $i + 1) {
			$s = "hello world"
			$s = strlen($s)
			$s = toUpper("hello")
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRunWithBuiltins(script)
	if err != nil {
		b.Fatal(err)
	}
}

// ============================================================================
// 数组操作基准测试
// ============================================================================

func BenchmarkArrayPush(b *testing.B) {
	script := fmt.Sprintf(`
		$arr = []
		for ($i = 0; $i < %d; $i = $i + 1) {
			$arr = push($arr, $i)
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRunWithBuiltins(script)
	if err != nil {
		b.Fatal(err)
	}
}

func BenchmarkArrayAccess(b *testing.B) {
	script := fmt.Sprintf(`
		$arr = [0,1,2,3,4,5,6,7,8,9]
		$r = 0
		for ($i = 0; $i < %d; $i = $i + 1) {
			$r = $r + $arr[$i %% 10]
		}
	`, b.N)
	b.ResetTimer()
	_, err := compileAndRun(script)
	if err != nil {
		b.Fatal(err)
	}
}
