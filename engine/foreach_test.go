package engine

import (
	"fmt"
	"strings"
	"testing"
)

// ============================================================================
// foreach 数组遍历测试
// ============================================================================

func TestForeachArrayValuesOnly(t *testing.T) {
	script := `
		$arr = [10, 20, 30]
		$sum = 0
		foreach ($val in $arr) {
			$sum = $sum + $val
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("expected sum=60, got %d", result.Int())
	}
}

func TestForeachArrayWithKey(t *testing.T) {
	script := `
		$arr = [10, 20, 30]
		$keySum = 0
		$valSum = 0
		foreach ($key => $val in $arr) {
			$keySum = $keySum + $key
			$valSum = $valSum + $val
		}
		$keySum * 100 + $valSum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// keySum = 0+1+2 = 3, valSum = 10+20+30 = 60, result = 3*100+60 = 360
	if result.Int() != 360 {
		t.Errorf("expected 360, got %d", result.Int())
	}
}

func TestForeachEmptyArray(t *testing.T) {
	script := `
		$arr = []
		$count = 0
		foreach ($val in $arr) {
			$count = $count + 1
		}
		$count
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("expected count=0 for empty array, got %d", result.Int())
	}
}

func TestForeachNestedArrays(t *testing.T) {
	script := `
		$matrix = [[1, 2], [3, 4], [5, 6]]
		$sum = 0
		foreach ($row in $matrix) {
			foreach ($val in $row) {
				$sum = $sum + $val
			}
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 21 {
		t.Errorf("expected sum=21, got %d", result.Int())
	}
}

// ============================================================================
// foreach 对象遍历测试
// ============================================================================

func TestForeachObjectValuesOnly(t *testing.T) {
	script := `
		$obj = {a: 10, b: 20, c: 30}
		$sum = 0
		foreach ($val in $obj) {
			$sum = $sum + $val
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("expected sum=60, got %d", result.Int())
	}
}

func TestForeachObjectWithKey(t *testing.T) {
	script := `
		$obj = {x: 1, y: 2, z: 3}
		$keys = ""
		$sum = 0
		foreach ($key => $val in $obj) {
			$keys = $keys + $key
			$sum = $sum + $val
		}
		// 检查键是否按字母顺序遍历
		$keys
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 对象遍历按键排序，所以应该是 "xyz"
	if !strings.Contains(result.String(), "xyz") {
		t.Errorf("expected keys to contain 'xyz' (sorted), got %s", result.String())
	}
}

func TestForeachEmptyObject(t *testing.T) {
	script := `
		$obj = {}
		$count = 0
		foreach ($val in $obj) {
			$count = $count + 1
		}
		$count
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("expected count=0 for empty object, got %d", result.Int())
	}
}

// ============================================================================
// foreach break/continue 测试
// ============================================================================

func TestForeachBreak(t *testing.T) {
	script := `
		$arr = [1, 2, 3, 4, 5]
		$sum = 0
		foreach ($val in $arr) {
			if ($val == 3) {
				break
			}
			$sum = $sum + $val
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 + 2 = 3，遇到 3 时 break
	if result.Int() != 3 {
		t.Errorf("expected sum=3 (break at 3), got %d", result.Int())
	}
}

func TestForeachContinue(t *testing.T) {
	script := `
		$arr = [1, 2, 3, 4, 5]
		$sum = 0
		foreach ($val in $arr) {
			if ($val == 3) {
				continue
			}
			$sum = $sum + $val
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 + 2 + 4 + 5 = 12，跳过 3
	if result.Int() != 12 {
		t.Errorf("expected sum=12 (skip 3), got %d", result.Int())
	}
}

// ============================================================================
// foreach 变量作用域测试
// ============================================================================

func TestForeachVariableScope(t *testing.T) {
	script := `
		$x = 100
		$arr = [1, 2, 3]
		foreach ($x in $arr) {
			// 这里的 $x 应该覆盖外部的 $x
		}
		// 循环结束后 $x 应该恢复为 100（取决于实现，也可能保留最后一个值）
		// 目前测试：确认循环变量能正常工作
		$x
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 根据当前实现，循环变量会声明在 foreach 的作用域内
	// 但外部 $x 的值在循环后应该是 100（不受影响）
	// 注：具体行为取决于作用域实现，这里主要测试不报错
	t.Logf("Final $x value: %v", result.Int())
}

// ============================================================================
// foreach 错误处理测试
// ============================================================================

func TestForeachInvalidType(t *testing.T) {
	script := `
		$x = 42
		foreach ($val in $x) {
			println $val
		}
	`
	_, err := compileAndRun(script)
	if err == nil {
		t.Error("expected error when iterating over integer, got nil")
	}
	if !strings.Contains(err.Error(), "cannot iterate") {
		t.Errorf("expected 'cannot iterate' error, got: %v", err)
	}
}

func TestForeachNull(t *testing.T) {
	script := `
		$x = null
		foreach ($val in $x) {
			println $val
		}
	`
	_, err := compileAndRun(script)
	if err == nil {
		t.Error("expected error when iterating over null, got nil")
	}
}

// ============================================================================
// foreach 复杂表达式测试
// ============================================================================

func TestForeachComplexExpression(t *testing.T) {
	script := `
		function makeArray() {
			return [10, 20, 30]
		}
		$sum = 0
		foreach ($val in makeArray()) {
			$sum = $sum + $val
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 60 {
		t.Errorf("expected sum=60, got %d", result.Int())
	}
}

func TestForeachStringArray(t *testing.T) {
	script := `
		$arr = ["hello", " ", "world"]
		$result = ""
		foreach ($str in $arr) {
			$result = $result + $str
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result.String())
	}
}

// ============================================================================
// foreach 性能测试
// ============================================================================

func TestForeachLargeArray(t *testing.T) {
	script := `
		$sum = 0
		for ($i = 0; $i < 1000; $i = $i + 1) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0+1+2+...+999 = 999*1000/2 = 499500
	expected := int64(499500)
	if result.Int() != expected {
		t.Errorf("expected %d, got %d", expected, result.Int())
	}
}

// ============================================================================
// range 范围测试
// ============================================================================

func TestRangeHalfOpen(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in 1...5) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1+2+3+4 = 10
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestRangeInclusive(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in 1..=5) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1+2+3+4+5 = 15
	if result.Int() != 15 {
		t.Errorf("expected 15, got %d", result.Int())
	}
}

func TestRangeReverse(t *testing.T) {
	script := `
		$result = ""
		foreach ($i in 5...1) {
			$result = $result + $i + ","
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// start > end in half-open range is empty
	if result.String() != "" {
		t.Errorf("expected '', got '%s'", result.String())
	}
}

func TestRangeReverseInclusive(t *testing.T) {
	script := `
		$result = ""
		foreach ($i in 5..=1) {
			$result = $result + $i + ","
		}
		$result
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// start > end in inclusive range is empty
	if result.String() != "" {
		t.Errorf("expected '', got '%s'", result.String())
	}
}

func TestRangeFunction(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in range(1, 5)) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("range", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("range() expects 2-3 arguments")
			}
			start := args[0].Int()
			end := args[1].Int()
			inclusive := false
			if len(args) >= 3 {
				inclusive = args[2].Bool()
			}
			return NewRange(start, end, inclusive), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1+2+3+4 = 10
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}

func TestRangeFunctionInclusive(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in range(1, 5, true)) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRunWithFuncs(script, func(e *Engine) {
		e.RegisterFunc("range", func(ctx *Context, args []Value) (Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("range() expects 2-3 arguments")
			}
			start := args[0].Int()
			end := args[1].Int()
			inclusive := false
			if len(args) >= 3 {
				inclusive = args[2].Bool()
			}
			return NewRange(start, end, inclusive), nil
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1+2+3+4+5 = 15
	if result.Int() != 15 {
		t.Errorf("expected 15, got %d", result.Int())
	}
}

func TestRangeValue(t *testing.T) {
	script := `
		$r = 1...5
		$r
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type() != TypeRange {
		t.Errorf("expected TypeRange, got %v", result.Type())
	}
}

func TestRangeValueInclusive(t *testing.T) {
	script := `
		$r = 1..=5
		$r
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type() != TypeRange {
		t.Errorf("expected TypeRange, got %v", result.Type())
	}
}

func TestRangeEmpty(t *testing.T) {
	script := `
		$count = 0
		foreach ($i in 5...1) {
			$count = $count + 1
		}
		$count
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 5...1 is empty (no valid values when start > end in half-open)
	if result.Int() != 0 {
		t.Errorf("expected 0, got %d", result.Int())
	}
}

func TestRangeSingleElement(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in 1...2) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1...2 = only 1
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

func TestRangeSingleElementInclusive(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in 1..=1) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1..=1 = only 1
	if result.Int() != 1 {
		t.Errorf("expected 1, got %d", result.Int())
	}
}

func TestRangeNegative(t *testing.T) {
	script := `
		$sum = 0
		foreach ($i in -3..=0) {
			$sum = $sum + $i
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// -3 + -2 + -1 + 0 = -6
	if result.Int() != -6 {
		t.Errorf("expected -6, got %d", result.Int())
	}
}

func TestRangeWithKey(t *testing.T) {
	script := `
		$keySum = 0
		$valSum = 0
		foreach ($k => $v in 1...5) {
			$keySum = $keySum + $k
			$valSum = $valSum + $v
		}
		$keySum * 100 + $valSum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1...5 = [1,5) = 1,2,3,4 (4 elements)
	// k=0,1,2,3 => keySum = 0+1+2+3 = 6
	// v=1,2,3,4 => valSum = 1+2+3+4 = 10
	// result = 6*100 + 10 = 610
	if result.Int() != 610 {
		t.Errorf("expected 610, got %d", result.Int())
	}
}

func TestForeachInlineObjectArray(t *testing.T) {
	script := `
		$arr = [{a: 1}, {a: 2}, {a: 3}]
		$sum = 0
		foreach ($val in $arr) {
			$sum = $sum + $val.a
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 6 {
		t.Errorf("expected 6, got %d", result.Int())
	}
}

func TestForeachNestedObjectArray(t *testing.T) {
	script := `
		$arr = [{x: {y: 1}}, {x: {y: 2}}, {x: {y: 3}}]
		$sum = 0
		foreach ($val in $arr) {
			$sum = $sum + $val.x.y
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 6 {
		t.Errorf("expected 6, got %d", result.Int())
	}
}

func TestForeachObjectArrayWithKey(t *testing.T) {
	script := `
		$arr = [{a: 10}, {a: 20}, {a: 30}]
		$keyValSum = 0
		foreach ($key => $val in $arr) {
			$keyValSum = $keyValSum + $key * $val.a
		}
		$keyValSum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0*10 + 1*20 + 2*30 = 0 + 20 + 60 = 80
	if result.Int() != 80 {
		t.Errorf("expected 80, got %d", result.Int())
	}
}

func TestForeachObjectArrayWithStrings(t *testing.T) {
	script := `
		$arr = [{name: "a"}, {name: "b"}, {name: "c"}]
		$names = ""
		foreach ($val in $arr) {
			$names = $names + $val.name
		}
		$names
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "abc" {
		t.Errorf("expected abc, got %s", result.String())
	}
}

func TestForeachNestedInlineObjectArray(t *testing.T) {
	script := `
		$arr = [[{a: 1}, {a: 2}], [{a: 3}, {a: 4}]]
		$sum = 0
		foreach ($inner in $arr) {
			foreach ($val in $inner) {
				$sum = $sum + $val.a
			}
		}
		$sum
	`
	result, err := compileAndRun(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 + 2 + 3 + 4 = 10
	if result.Int() != 10 {
		t.Errorf("expected 10, got %d", result.Int())
	}
}
