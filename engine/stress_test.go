package engine

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

// compileAndRunWithBuiltins 编译脚本并执行，注册必要的内置函数
func compileAndRunWithBuiltins(script string) (Value, error) {
	prog, err := CompileString(script)
	if err != nil {
		return nil, err
	}
	eng := NewEngine()
	defer eng.Close()
	registerStressBuiltins(eng)
	vm := newVMWithProgram(eng, prog)
	if err := vm.Execute(); err != nil {
		return nil, err
	}
	return vm.GetResult(), nil
}

// registerStressBuiltins 注册压力测试所需的内置函数
func registerStressBuiltins(e *Engine) {
	// len
	e.RegisterFunc("len", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("len() expects 1 argument")
		}
		return NewInt(int64(args[0].Len())), nil
	})
	// push
	e.RegisterFunc("push", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("push() expects 2 arguments")
		}
		arr := args[0].Array()
		if arr == nil {
			return nil, fmt.Errorf("push() expects array")
		}
		arr = append(arr, args[1])
		return NewArray(arr), nil
	})
	// str
	e.RegisterFunc("str", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("str() expects 1 argument")
		}
		return NewString(args[0].String()), nil
	})
	// strlen
	e.RegisterFunc("strlen", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("strlen() expects 1 argument")
		}
		return NewInt(int64(len(args[0].String()))), nil
	})
	// toUpper
	e.RegisterFunc("toUpper", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("toUpper() expects 1 argument")
		}
		s := args[0].String()
		upper := ""
		for _, ch := range s {
			if ch >= 'a' && ch <= 'z' {
				upper += string(rune(ch - 32))
			} else {
				upper += string(ch)
			}
		}
		return NewString(upper), nil
	})
	// abs
	e.RegisterFunc("abs", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("abs() expects 1 argument")
		}
		v := args[0].Int()
		if v < 0 {
			v = -v
		}
		return NewInt(v), nil
	})
	// ceil
	e.RegisterFunc("ceil", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("ceil() expects 1 argument")
		}
		v := args[0].Float()
		return NewInt(int64(v + 0.999999)), nil
	})
	// floor
	e.RegisterFunc("floor", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("floor() expects 1 argument")
		}
		v := args[0].Float()
		return NewInt(int64(v)), nil
	})
	// filter
	e.RegisterFunc("filter", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("filter() expects 2 arguments")
		}
		arr := args[0].Array()
		fn := args[1]
		if arr == nil {
			return nil, fmt.Errorf("filter() expects array")
		}
		var result []Value
		for _, item := range arr {
			val, err := ctx.VM().CallValue(fn, item)
			if err != nil {
				return nil, err
			}
			if val.Bool() {
				result = append(result, item)
			}
		}
		return NewArray(result), nil
	})
	// map
	e.RegisterFunc("map", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("map() expects 2 arguments")
		}
		arr := args[0].Array()
		fn := args[1]
		if arr == nil {
			return nil, fmt.Errorf("map() expects array")
		}
		var result []Value
		for _, item := range arr {
			val, err := ctx.VM().CallValue(fn, item)
			if err != nil {
				return nil, err
			}
			result = append(result, val)
		}
		return NewArray(result), nil
	})
	// reduce
	e.RegisterFunc("reduce", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("reduce() expects 3 arguments")
		}
		arr := args[0].Array()
		fn := args[1]
		acc := args[2]
		if arr == nil {
			return nil, fmt.Errorf("reduce() expects array")
		}
		for _, item := range arr {
			val, err := ctx.VM().CallValue(fn, acc, item)
			if err != nil {
				return nil, err
			}
			acc = val
		}
		return acc, nil
	})
	// sort
	e.RegisterFunc("sort", func(ctx *Context, args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sort() expects 1 argument")
		}
		arr := args[0].Array()
		if arr == nil {
			return nil, fmt.Errorf("sort() expects array")
		}
		cp := make([]Value, len(arr))
		copy(cp, arr)
		sort.Slice(cp, func(i, j int) bool {
			return cp[i].Less(cp[j])
		})
		return NewArray(cp), nil
	})
}

// ============================================================================
// 大循环压力测试
// ============================================================================

func TestStressLargeLoop(t *testing.T) {
	count := 1_000_000
	script := fmt.Sprintf(`
		$r = 0
		for ($i = 0; $i < %d; $i = $i + 1) {
			$r = $r + 1
		}
		$r
	`, count)

	start := time.Now()
	result, err := compileAndRun(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("large loop failed: %v", err)
	}
	if result.Int() != int64(count) {
		t.Errorf("expected %d, got %d", count, result.Int())
	}
	t.Logf("1M iterations: %v (%.0f ops/sec)", elapsed, float64(count)/elapsed.Seconds())
}

func TestStressNestedLoops(t *testing.T) {
	n := 1000
	script := fmt.Sprintf(`
		$r = 0
		for ($i = 0; $i < %d; $i = $i + 1) {
			for ($j = 0; $j < %d; $j = $j + 1) {
				$r = $r + 1
			}
		}
		$r
	`, n, n)

	start := time.Now()
	result, err := compileAndRun(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("nested loops failed: %v", err)
	}
	expected := int64(n) * int64(n)
	if result.Int() != expected {
		t.Errorf("expected %d, got %d", expected, result.Int())
	}
	t.Logf("1M nested iterations: %v (%.0f ops/sec)", elapsed, float64(expected)/elapsed.Seconds())
}

// ============================================================================
// 深递归压力测试
// ============================================================================

func TestStressDeepRecursion(t *testing.T) {
	depth := 500
	script := fmt.Sprintf(`
		fn countdown($n) {
			if ($n <= 0) { return 0 }
			return countdown($n - 1) + 1
		}
		countdown(%d)
	`, depth)

	start := time.Now()
	result, err := compileAndRun(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("deep recursion failed: %v", err)
	}
	if result.Int() != int64(depth) {
		t.Errorf("expected %d, got %d", depth, result.Int())
	}
	t.Logf("recursion depth %d: %v", depth, elapsed)
}

func TestStressStackOverflow(t *testing.T) {
	script := `
		fn recurse($n) {
			return recurse($n + 1)
		}
		recurse(0)
	`

	_, err := compileAndRun(script)
	if err == nil {
		t.Fatal("expected stack overflow error")
	}
	re, ok := err.(*RuntimeError)
	if !ok {
		t.Fatalf("expected RuntimeError, got %T: %v", err, err)
	}
	if re.Message != "stack overflow: maximum call depth exceeded" {
		t.Errorf("unexpected error message: %s", re.Message)
	}
}

// ============================================================================
// 大内存压力测试
// ============================================================================

func TestStressLargeArray(t *testing.T) {
	count := 100_000
	script := fmt.Sprintf(`
		$arr = []
		for ($i = 0; $i < %d; $i = $i + 1) {
			$arr = push($arr, $i)
		}
		len($arr)
	`, count)

	start := time.Now()
	result, err := compileAndRunWithBuiltins(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("large array failed: %v", err)
	}
	if result.Int() != int64(count) {
		t.Errorf("expected len %d, got %d", count, result.Int())
	}
	t.Logf("100K array push: %v", elapsed)
}

func TestStressLargeObject(t *testing.T) {
	count := 10_000
	script := fmt.Sprintf(`
		$obj = {}
		for ($i = 0; $i < %d; $i = $i + 1) {
			$obj["key" + str($i)] = $i
		}
		len($obj)
	`, count)

	start := time.Now()
	result, err := compileAndRunWithBuiltins(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("large object failed: %v", err)
	}
	if result.Int() != int64(count) {
		t.Errorf("expected len %d, got %d", count, result.Int())
	}
	t.Logf("10K object insert: %v", elapsed)
}

func TestStressArraySort(t *testing.T) {
	count := 10_000
	script := fmt.Sprintf(`
		$arr = []
		for ($i = %d; $i >= 0; $i = $i - 1) {
			$arr = push($arr, $i)
		}
		$arr = sort($arr)
		$arr[0]
	`, count-1)

	start := time.Now()
	result, err := compileAndRunWithBuiltins(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("array sort failed: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("expected 0, got %d", result.Int())
	}
	t.Logf("10K array sort: %v", elapsed)
}

// ============================================================================
// 复合压力测试
// ============================================================================

func TestStressFibHeavy(t *testing.T) {
	script := `
		fn fib($n) {
			if ($n <= 1) { return $n }
			return fib($n - 1) + fib($n - 2)
		}
		fib(30)
	`

	start := time.Now()
	result, err := compileAndRun(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("fib(30) failed: %v", err)
	}
	if result.Int() != 832040 {
		t.Errorf("expected 832040, got %d", result.Int())
	}
	t.Logf("fib(30): %v", elapsed)
}

func TestStressMixedWorkload(t *testing.T) {
	script := `
		$arr = []
		for ($i = 0; $i < 10000; $i = $i + 1) {
			$arr = push($arr, $i)
		}
		$evens = filter($arr, fn($x) { return $x % 2 == 0 })
		$doubled = map($evens, fn($x) { return $x * 2 })
		$sum = reduce($doubled, fn($acc, $x) { return $acc + $x }, 0)
		$sum
	`

	start := time.Now()
	result, err := compileAndRunWithBuiltins(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("mixed workload failed: %v", err)
	}
	if result.Int() != 49990000 {
		t.Errorf("expected 49990000, got %d", result.Int())
	}
	t.Logf("mixed workload: %v", elapsed)
}

func TestStressStringConcat(t *testing.T) {
	count := 10_000
	script := fmt.Sprintf(`
		$s = ""
		for ($i = 0; $i < %d; $i = $i + 1) {
			$s = $s + "a"
		}
		strlen($s)
	`, count)

	start := time.Now()
	result, err := compileAndRunWithBuiltins(script)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("string concat failed: %v", err)
	}
	if result.Int() != int64(count) {
		t.Errorf("expected %d, got %d", count, result.Int())
	}
	t.Logf("10K string concat: %v", elapsed)
}
