package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestArrayNames(t *testing.T) {
	names := ArrayNames()
	if len(names) != 41 {
		t.Errorf("expected 41 array function names, got %d", len(names))
	}
}

// ============================================================================
// push
// ============================================================================

func TestPush(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2)})
	result, err := callBuiltin("push", arr, engine.NewInt(3))
	if err != nil {
		t.Fatalf("push error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("push() returned %d, want 3", result.Int())
	}
}

func TestPushMultiple(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1)})
	result, err := callBuiltin("push", arr, engine.NewInt(2), engine.NewInt(3))
	if err != nil {
		t.Fatalf("push error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("push() returned %d, want 3", result.Int())
	}
}

func TestPushNotArray(t *testing.T) {
	_, err := callBuiltin("push", engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("push(int) should return error")
	}
}

// ============================================================================
// pop
// ============================================================================

func TestPop(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("pop", arr)
	if err != nil {
		t.Fatalf("pop error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("pop() returned %d, want 3", result.Int())
	}
}

func TestPopEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})
	result, err := callBuiltin("pop", arr)
	if err != nil {
		t.Fatalf("pop error: %v", err)
	}
	if !result.IsNull() {
		t.Error("pop([]) should return null")
	}
}

// ============================================================================
// shift
// ============================================================================

func TestShift(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("shift", arr)
	if err != nil {
		t.Fatalf("shift error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("shift() returned %d, want 1", result.Int())
	}
}

func TestShiftEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})
	result, err := callBuiltin("shift", arr)
	if err != nil {
		t.Fatalf("shift error: %v", err)
	}
	if !result.IsNull() {
		t.Error("shift([]) should return null")
	}
}

// ============================================================================
// unshift
// ============================================================================

func TestUnshift(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("unshift", arr, engine.NewInt(1))
	if err != nil {
		t.Fatalf("unshift error: %v", err)
	}
	if result.Int() != 3 {
		t.Errorf("unshift() returned %d, want 3", result.Int())
	}
}

// ============================================================================
// splice
// ============================================================================

func TestSpliceDelete(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3), engine.NewInt(4)})
	result, err := callBuiltin("splice", arr, engine.NewInt(1), engine.NewInt(2))
	if err != nil {
		t.Fatalf("splice error: %v", err)
	}
	// 返回被删除的元素
	deleted := result.Array()
	if len(deleted) != 2 {
		t.Fatalf("splice returned %d elements, want 2", len(deleted))
	}
	if deleted[0].Int() != 2 || deleted[1].Int() != 3 {
		t.Errorf("splice deleted = [%d, %d], want [2, 3]", deleted[0].Int(), deleted[1].Int())
	}
}

func TestSpliceInsert(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(4)})
	result, err := callBuiltin("splice", arr, engine.NewInt(1), engine.NewInt(0), engine.NewInt(2), engine.NewInt(3))
	if err != nil {
		t.Fatalf("splice error: %v", err)
	}
	// 返回被删除的元素（0 个）
	deleted := result.Array()
	if len(deleted) != 0 {
		t.Errorf("splice returned %d elements, want 0", len(deleted))
	}
}

func TestSpliceNegativeStart(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("splice", arr, engine.NewInt(-1), engine.NewInt(1))
	if err != nil {
		t.Fatalf("splice error: %v", err)
	}
	deleted := result.Array()
	if len(deleted) != 1 || deleted[0].Int() != 3 {
		t.Errorf("splice(-1, 1) deleted = %v, want [3]", deleted)
	}
}

// ============================================================================
// indexOf / lastIndexOf
// ============================================================================

func TestIndexOf(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("indexOf", arr, engine.NewInt(2))
	if err != nil {
		t.Fatalf("indexOf error: %v", err)
	}
	if result.Int() != 1 {
		t.Errorf("indexOf(2) = %d, want 1", result.Int())
	}
}

func TestIndexOfNotFound(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2)})
	result, err := callBuiltin("indexOf", arr, engine.NewInt(5))
	if err != nil {
		t.Fatalf("indexOf error: %v", err)
	}
	if result.Int() != -1 {
		t.Errorf("indexOf(5) = %d, want -1", result.Int())
	}
}

func TestLastIndexOf(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(1)})
	result, err := callBuiltin("lastIndexOf", arr, engine.NewInt(1))
	if err != nil {
		t.Fatalf("lastIndexOf error: %v", err)
	}
	if result.Int() != 2 {
		t.Errorf("lastIndexOf(1) = %d, want 2", result.Int())
	}
}

// ============================================================================
// includes
// ============================================================================

func TestIncludes(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("includes", arr, engine.NewInt(2))
	if err != nil {
		t.Fatalf("includes error: %v", err)
	}
	if !result.Bool() {
		t.Error("includes(2) should be true")
	}

	result, err = callBuiltin("includes", arr, engine.NewInt(5))
	if err != nil {
		t.Fatalf("includes error: %v", err)
	}
	if result.Bool() {
		t.Error("includes(5) should be false")
	}
}

// ============================================================================
// slice
// ============================================================================

func TestSlice(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3), engine.NewInt(4)})
	result, err := callBuiltin("slice", arr, engine.NewInt(1), engine.NewInt(3))
	if err != nil {
		t.Fatalf("slice error: %v", err)
	}
	sliced := result.Array()
	if len(sliced) != 2 {
		t.Fatalf("slice() returned %d elements, want 2", len(sliced))
	}
	if sliced[0].Int() != 2 || sliced[1].Int() != 3 {
		t.Errorf("slice() = [%d, %d], want [2, 3]", sliced[0].Int(), sliced[1].Int())
	}
}

func TestSliceNoEnd(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("slice", arr, engine.NewInt(1))
	if err != nil {
		t.Fatalf("slice error: %v", err)
	}
	sliced := result.Array()
	if len(sliced) != 2 {
		t.Fatalf("slice(1) returned %d elements, want 2", len(sliced))
	}
}

func TestSliceNegative(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("slice", arr, engine.NewInt(-2))
	if err != nil {
		t.Fatalf("slice error: %v", err)
	}
	sliced := result.Array()
	if len(sliced) != 2 {
		t.Fatalf("slice(-2) returned %d elements, want 2", len(sliced))
	}
	if sliced[0].Int() != 2 || sliced[1].Int() != 3 {
		t.Errorf("slice(-2) = [%d, %d], want [2, 3]", sliced[0].Int(), sliced[1].Int())
	}
}

// ============================================================================
// array_reverse
// ============================================================================

func TestArrayReverse(t *testing.T) {
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	result, err := callBuiltin("array_reverse", arr)
	if err != nil {
		t.Fatalf("array_reverse error: %v", err)
	}
	reversed := result.Array()
	if len(reversed) != 3 {
		t.Fatalf("array_reverse() returned %d elements, want 3", len(reversed))
	}
	expected := []int64{3, 2, 1}
	for i, v := range reversed {
		if v.Int() != expected[i] {
			t.Errorf("array_reverse()[%d] = %d, want %d", i, v.Int(), expected[i])
		}
	}
}

func TestArrayReverseEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})
	result, err := callBuiltin("array_reverse", arr)
	if err != nil {
		t.Fatalf("array_reverse error: %v", err)
	}
	if len(result.Array()) != 0 {
		t.Error("array_reverse([]) should return empty array")
	}
}

// ============================================================================
// flat
// ============================================================================

func TestFlat(t *testing.T) {
	inner1 := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2)})
	inner2 := engine.NewArray([]engine.Value{engine.NewInt(3), engine.NewInt(4)})
	arr := engine.NewArray([]engine.Value{inner1, inner2})

	result, err := callBuiltin("flat", arr)
	if err != nil {
		t.Fatalf("flat error: %v", err)
	}
	flattened := result.Array()
	if len(flattened) != 4 {
		t.Fatalf("flat() returned %d elements, want 4", len(flattened))
	}
	for i, v := range flattened {
		if v.Int() != int64(i+1) {
			t.Errorf("flat()[%d] = %d, want %d", i, v.Int(), i+1)
		}
	}
}

func TestFlatDepth2(t *testing.T) {
	inner := engine.NewArray([]engine.Value{
		engine.NewArray([]engine.Value{engine.NewInt(1)}),
	})
	outer := engine.NewArray([]engine.Value{inner})

	result, err := callBuiltin("flat", outer, engine.NewInt(2))
	if err != nil {
		t.Fatalf("flat error: %v", err)
	}
	flattened := result.Array()
	if len(flattened) != 1 || flattened[0].Int() != 1 {
		t.Errorf("flat(depth=2) = %v, want [1]", flattened)
	}
}

// ============================================================================
// unique
// ============================================================================

func TestUnique(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(1), engine.NewInt(2), engine.NewInt(1), engine.NewInt(3), engine.NewInt(2),
	})
	result, err := callBuiltin("unique", arr)
	if err != nil {
		t.Fatalf("unique error: %v", err)
	}
	unique := result.Array()
	if len(unique) != 3 {
		t.Fatalf("unique() returned %d elements, want 3", len(unique))
	}
	expected := []int64{1, 2, 3}
	for i, v := range unique {
		if v.Int() != expected[i] {
			t.Errorf("unique()[%d] = %d, want %d", i, v.Int(), expected[i])
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewString("a"), engine.NewString("b"), engine.NewString("a"),
	})
	result, err := callBuiltin("unique", arr)
	if err != nil {
		t.Fatalf("unique error: %v", err)
	}
	unique := result.Array()
	if len(unique) != 2 {
		t.Errorf("unique() returned %d elements, want 2", len(unique))
	}
}

func TestUniqueEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})
	result, err := callBuiltin("unique", arr)
	if err != nil {
		t.Fatalf("unique error: %v", err)
	}
	if len(result.Array()) != 0 {
		t.Error("unique([]) should return empty array")
	}
}

// ============================================================================
// 参数校验测试
// ============================================================================

func TestArrayFuncArgErrors(t *testing.T) {
	funcs := []string{
		"pop", "shift", "array_reverse", "unique",
	}
	for _, fn := range funcs {
		_, err := callBuiltin(fn, engine.NewInt(42))
		if err == nil {
			t.Errorf("%s(int) should return error", fn)
		}
	}

	funcs2 := []string{"indexOf", "lastIndexOf", "includes"}
	for _, fn := range funcs2 {
		_, err := callBuiltin(fn, engine.NewInt(42), engine.NewInt(1))
		if err == nil {
			t.Errorf("%s(int, val) should return error", fn)
		}
	}
}

// TestCount 测试 count/sizeof 函数
func TestCount(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		name     string
		value    engine.Value
		expected int64
	}{
		{"array with 3 elements", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)}), 3},
		{"empty array", engine.NewArray([]engine.Value{}), 0},
		{"object with 2 keys", engine.NewObject(map[string]engine.Value{"a": engine.NewInt(1), "b": engine.NewInt(2)}), 2},
		{"empty object", engine.NewObject(map[string]engine.Value{}), 0},
		{"null", engine.NewNull(), 0},
		{"int", engine.NewInt(42), 1},
		{"string", engine.NewString("hello"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinSize(ctx, []engine.Value{tt.value})
			if err != nil {
				t.Fatalf("size() error: %v", err)
			}
			if result.Int() != tt.expected {
				t.Errorf("size() = %d, expected %d", result.Int(), tt.expected)
			}
		})
	}
}

// TestArrayKeyExists 测试 array_key_exists 函数
func TestArrayKeyExists(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr := engine.NewArray([]engine.Value{engine.NewString("a"), engine.NewString("b"), engine.NewString("c")})

	tests := []struct {
		name     string
		key      engine.Value
		expected bool
	}{
		{"key 0", engine.NewInt(0), true},
		{"key 1", engine.NewInt(1), true},
		{"key 2", engine.NewInt(2), true},
		{"key 3 (out of range)", engine.NewInt(3), false},
		{"negative key -1", engine.NewInt(-1), true}, // 最后一个元素
		{"negative key -3", engine.NewInt(-3), true}, // 第一个元素
		{"negative key -4", engine.NewInt(-4), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinArrayKeyExists(ctx, []engine.Value{arr, tt.key})
			if err != nil {
				t.Fatalf("array_key_exists() error: %v", err)
			}
			if result.Bool() != tt.expected {
				t.Errorf("array_key_exists(arr, %d) = %v, expected %v", tt.key.Int(), result.Bool(), tt.expected)
			}
		})
	}
}

// TestArrayMerge 测试 array_merge 函数
func TestArrayMerge(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		name     string
		arrays   []engine.Value
		expected int // 期望的元素数量
	}{
		{
			"merge two arrays",
			[]engine.Value{
				engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2)}),
				engine.NewArray([]engine.Value{engine.NewInt(3), engine.NewInt(4)}),
			},
			4,
		},
		{
			"merge array with scalar",
			[]engine.Value{
				engine.NewArray([]engine.Value{engine.NewInt(1)}),
				engine.NewInt(2),
			},
			2,
		},
		{
			"merge three arrays",
			[]engine.Value{
				engine.NewArray([]engine.Value{engine.NewInt(1)}),
				engine.NewArray([]engine.Value{engine.NewInt(2)}),
				engine.NewArray([]engine.Value{engine.NewInt(3)}),
			},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinArrayMerge(ctx, tt.arrays)
			if err != nil {
				t.Fatalf("array_merge() error: %v", err)
			}
			if result.Type() != engine.TypeArray {
				t.Fatal("array_merge() should return array")
			}
			if result.Len() != tt.expected {
				t.Errorf("array_merge() length = %d, expected %d", result.Len(), tt.expected)
			}
		})
	}
}

// TestArraySum 测试 array_sum 函数
func TestArraySum(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		name     string
		arr      engine.Value
		expected float64
		isInt    bool
	}{
		{"sum of integers", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)}), 6.0, true},
		{"sum of floats", engine.NewArray([]engine.Value{engine.NewFloat(1.5), engine.NewFloat(2.5)}), 4.0, false},
		{"mixed sum", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewFloat(2.5)}), 3.5, false},
		{"sum with bool", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewBool(true)}), 2.0, true},
		{"empty array", engine.NewArray([]engine.Value{}), 0.0, true},
		{"single element", engine.NewArray([]engine.Value{engine.NewInt(42)}), 42.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinArraySum(ctx, []engine.Value{tt.arr})
			if err != nil {
				t.Fatalf("array_sum() error: %v", err)
			}
			var got float64
			if result.Type() == engine.TypeInt {
				got = float64(result.Int())
			} else {
				got = result.Float()
			}
			if got != tt.expected {
				t.Errorf("array_sum() = %f, expected %f", got, tt.expected)
			}
		})
	}
}

// TestArrayProduct 测试 array_product 函数
func TestArrayProduct(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		name     string
		arr      engine.Value
		expected float64
		isInt    bool
	}{
		{"product of 1,2,3", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)}), 6.0, true},
		{"product with float", engine.NewArray([]engine.Value{engine.NewInt(2), engine.NewFloat(3.5)}), 7.0, false},
		{"empty array", engine.NewArray([]engine.Value{}), 1.0, true}, // 单位元
		{"single element", engine.NewArray([]engine.Value{engine.NewInt(5)}), 5.0, true},
		{"product with zero", engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(0), engine.NewInt(3)}), 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinArrayProduct(ctx, []engine.Value{tt.arr})
			if err != nil {
				t.Fatalf("array_product() error: %v", err)
			}
			var got float64
			if result.Type() == engine.TypeInt {
				got = float64(result.Int())
			} else {
				got = result.Float()
			}
			if got != tt.expected {
				t.Errorf("array_product() = %f, expected %f", got, tt.expected)
			}
		})
	}
}

// TestArrayValues 测试 array_values 函数
func TestArrayValues(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewString("two"), engine.NewBool(true)})

	ctx := engine.NewContext(e, nil)
	result, err := builtinArrayValues(ctx, []engine.Value{arr})
	if err != nil {
		t.Fatalf("array_values() error: %v", err)
	}

	if result.Len() != 3 {
		t.Errorf("array_values() length = %d, expected 3", result.Len())
	}

	// 检查元素是否相同
	for i := range 3 {
		if !result.Array()[i].Equals(arr.Array()[i]) {
			t.Errorf("array_values()[%d] != arr[%d]", i, i)
		}
	}
}

// TestArrayDiff 测试 array_diff 函数
func TestArrayDiff(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		name         string
		arrays       []engine.Value
		expectedLen  int
		expectedVals []int64
	}{
		{
			"diff with one array",
			[]engine.Value{
				engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)}),
				engine.NewArray([]engine.Value{engine.NewInt(2)}),
			},
			2,
			[]int64{1, 3},
		},
		{
			"diff with multiple arrays",
			[]engine.Value{
				engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3), engine.NewInt(4)}),
				engine.NewArray([]engine.Value{engine.NewInt(2)}),
				engine.NewArray([]engine.Value{engine.NewInt(4)}),
			},
			2,
			[]int64{1, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := engine.NewContext(e, nil)
			result, err := builtinArrayDiff(ctx, tt.arrays)
			if err != nil {
				t.Fatalf("array_diff() error: %v", err)
			}
			if result.Len() != tt.expectedLen {
				t.Errorf("array_diff() length = %d, expected %d", result.Len(), tt.expectedLen)
			}
			// 验证结果元素
			for i, exp := range tt.expectedVals {
				if !result.Array()[i].Equals(engine.NewInt(exp)) {
					t.Errorf("array_diff()[%d] = %v, expected %d", i, result.Array()[i], exp)
				}
			}
		})
	}
}

// TestArrayIntersect 测试 array_intersect 函数
func TestArrayIntersect(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr1 := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})
	arr2 := engine.NewArray([]engine.Value{engine.NewInt(2), engine.NewInt(3), engine.NewInt(4)})

	ctx := engine.NewContext(e, nil)
	result, err := builtinArrayIntersect(ctx, []engine.Value{arr1, arr2})
	if err != nil {
		t.Fatalf("array_intersect() error: %v", err)
	}

	// 应该返回 [2, 3]
	if result.Len() != 2 {
		t.Errorf("array_intersect() length = %d, expected 2", result.Len())
	}

	// 检查包含 2 和 3
	has2, has3 := false, false
	for _, v := range result.Array() {
		if v.Equals(engine.NewInt(2)) {
			has2 = true
		}
		if v.Equals(engine.NewInt(3)) {
			has3 = true
		}
	}
	if !has2 || !has3 {
		t.Errorf("array_intersect() should contain 2 and 3, got %v", result.Array())
	}
}

// TestInArray 测试 in_array 函数（includes 的别名）
func TestInArray(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3)})

	ctx := engine.NewContext(e, nil)

	// 测试存在的元素
	result, err := builtinInArray(ctx, []engine.Value{arr, engine.NewInt(2)})
	if err != nil {
		t.Fatalf("in_array() error: %v", err)
	}
	if !result.Bool() {
		t.Error("in_array(arr, 2) should be true")
	}

	// 测试不存在的元素
	result, err = builtinInArray(ctx, []engine.Value{arr, engine.NewInt(4)})
	if err != nil {
		t.Fatalf("in_array() error: %v", err)
	}
	if result.Bool() {
		t.Error("in_array(arr, 4) should be false")
	}
}

// TestArrayCopy 测试 array_copy 函数
func TestArrayCopy(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	// 创建嵌套数组
	inner := engine.NewArray([]engine.Value{engine.NewInt(4), engine.NewInt(5)})
	arr := engine.NewArray([]engine.Value{engine.NewInt(1), engine.NewInt(2), engine.NewInt(3), inner})

	ctx := engine.NewContext(e, nil)
	result, err := builtinArrayCopy(ctx, []engine.Value{arr})
	if err != nil {
		t.Fatalf("array_copy() error: %v", err)
	}

	// 检查长度相同
	if result.Len() != arr.Len() {
		t.Errorf("array_copy() length = %d, expected %d", result.Len(), arr.Len())
	}

	// 检查基本元素相等
	for i := range 3 {
		if !result.Array()[i].Equals(arr.Array()[i]) {
			t.Errorf("array_copy()[%d] != arr[%d]", i, i)
		}
	}

	// 检查嵌套数组是否独立（浅拷贝检查）
	copiedInner := result.Array()[3]
	if copiedInner.Type() != engine.TypeArray {
		t.Fatal("copied inner should be array")
	}

	// 修改原嵌套数组
	inner.Array()[0] = engine.NewInt(100)

	// 复制应该不受影响
	if copiedInner.Array()[0].Int() == 100 {
		t.Error("array_copy() did not create independent copy of nested array")
	}
}

// TestArrayAliases 测试函数别名
func TestArrayAliases(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	tests := []struct {
		alias    string
		original string
	}{
		{"key_exists", "array_key_exists"},
		{"in_array", "includes"},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			aliasFn := e.GetRegisteredFunc(tt.alias)
			if aliasFn == nil {
				t.Errorf("%s should be registered", tt.alias)
			}
		})
	}
}

// ============================================================================
// usort 测试
// ============================================================================

func TestUsortBasic(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	// Test: usort([3, 1, 2]) should return [1, 2, 3] (default comparison)
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(3),
		engine.NewInt(1),
		engine.NewInt(4),
		engine.NewInt(1),
		engine.NewInt(5),
	})

	ctx := engine.NewContext(e, nil)
	result, err := builtinUsort(ctx, []engine.Value{arr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 5 {
		t.Fatalf("expected 5 elements, got %d", len(resultArr))
	}
	if resultArr[0].Int() != 1 || resultArr[1].Int() != 1 || resultArr[2].Int() != 3 || resultArr[3].Int() != 4 || resultArr[4].Int() != 5 {
		t.Errorf("expected [1, 1, 3, 4, 5], got %v", resultArr)
	}
}

func TestUsortDescending(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$sorted = usort($arr, fn($a, $b) { return $a > $b; });
$sorted`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	if arr[0].Int() != 5 || arr[1].Int() != 4 || arr[2].Int() != 3 || arr[3].Int() != 2 || arr[4].Int() != 1 {
		t.Errorf("expected [5, 4, 3, 2, 1], got %v", arr)
	}
}

func TestUsortStrings(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr := engine.NewArray([]engine.Value{
		engine.NewString("banana"),
		engine.NewString("apple"),
		engine.NewString("cherry"),
	})

	ctx := engine.NewContext(e, nil)
	result, err := builtinUsort(ctx, []engine.Value{arr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultArr := result.Array()
	if resultArr[0].String() != "apple" || resultArr[1].String() != "banana" || resultArr[2].String() != "cherry" {
		t.Errorf("expected [apple, banana, cherry], got [%s, %s, %s]",
			resultArr[0].String(), resultArr[1].String(), resultArr[2].String())
	}
}

func TestUsortEmpty(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)

	arr := engine.NewArray([]engine.Value{})

	ctx := engine.NewContext(e, nil)
	result, err := builtinUsort(ctx, []engine.Value{arr})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 0 {
		t.Errorf("expected empty array, got %d elements", len(resultArr))
	}
}

func TestUsortWrongArgCount(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)
	ctx := engine.NewContext(e, nil)

	_, err := builtinUsort(ctx, []engine.Value{})
	if err == nil {
		t.Error("usort(0 args) should return error")
	}

	_, err = builtinUsort(ctx, []engine.Value{
		engine.NewArray([]engine.Value{}),
		engine.NewNull(),
		engine.NewNull(),
	})
	if err == nil {
		t.Error("usort(3 args) should return error")
	}
}

func TestUsortNotArray(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)
	ctx := engine.NewContext(e, nil)

	_, err := builtinUsort(ctx, []engine.Value{engine.NewInt(42), engine.NewNull()})
	if err == nil {
		t.Error("usort(int, fn) should return error")
	}
}

func TestUsortNotFunction(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterArray(e)
	ctx := engine.NewContext(e, nil)

	_, err := builtinUsort(ctx, []engine.Value{
		engine.NewArray([]engine.Value{engine.NewInt(1)}),
		engine.NewString("not a function"),
	})
	if err == nil {
		t.Error("usort(array, string) should return error")
	}
}
