package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ==================== FunctionalNames 测试 ====================

// TestFunctionalNames 测试 FunctionalNames 返回正确列表
func TestFunctionalNames(t *testing.T) {
	names := FunctionalNames()
	expected := []string{
		"map", "filter", "reject", "reduce", "find", "some", "every",
		"sort", "contains", "unique", "partition", "flattenDeep",
		"difference", "union", "zip", "unzip", "first", "last",
		"take", "drop", "sum", "size",
	}

	if len(names) != len(expected) {
		t.Errorf("expected %d functions, got %d", len(expected), len(names))
	}

	for i, name := range expected {
		if i >= len(names) || names[i] != name {
			t.Errorf("position %d: expected %s, got %v", i, name, names)
			break
		}
	}
}

// TestRegisterFunctional 测试 RegisterFunctional 注册所有函数
func TestRegisterFunctional(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	RegisterFunctional(e)

	// 验证所有函数都已注册
	names := FunctionalNames()
	for _, name := range names {
		fn := e.GetRegisteredFunc(name)
		if fn == nil {
			t.Errorf("function %s not registered", name)
		}
	}
}

// ==================== contains 测试（不需要回调）====================

// TestFunctionalContainsTrue 测试 contains 返回 true
func TestFunctionalContainsTrue(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(1),
		engine.NewInt(2),
		engine.NewInt(3),
	})

	result, err := callBuiltin("contains", arr, engine.NewInt(2))
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}

	if !result.Bool() {
		t.Error("expected true (contains 2)")
	}
}

// TestFunctionalContainsFalse 测试 contains 返回 false
func TestFunctionalContainsFalse(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(1),
		engine.NewInt(2),
		engine.NewInt(3),
	})

	result, err := callBuiltin("contains", arr, engine.NewInt(4))
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}

	if result.Bool() {
		t.Error("expected false (doesn't contain 4)")
	}
}

// TestFunctionalContainsString 测试 contains 字符串
func TestFunctionalContainsString(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewString("apple"),
		engine.NewString("banana"),
		engine.NewString("cherry"),
	})

	result, err := callBuiltin("contains", arr, engine.NewString("banana"))
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}

	if !result.Bool() {
		t.Error("expected true (contains 'banana')")
	}
}

// TestFunctionalContainsEmpty 测试 contains 空数组
func TestFunctionalContainsEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})

	result, err := callBuiltin("contains", arr, engine.NewInt(1))
	if err != nil {
		t.Fatalf("contains failed: %v", err)
	}

	if result.Bool() {
		t.Error("expected false for empty array")
	}
}

// ==================== sort 测试（简单情况，无回调）====================

// TestFunctionalSortDefault 测试 sort 默认升序
func TestFunctionalSortDefault(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(5),
		engine.NewInt(2),
		engine.NewInt(8),
		engine.NewInt(1),
		engine.NewInt(9),
	})

	result, err := callBuiltin("sort", arr)
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 2, 5, 8, 9}

	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}

	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// TestFunctionalSortEmpty 测试 sort 空数组
func TestFunctionalSortEmpty(t *testing.T) {
	arr := engine.NewArray([]engine.Value{})

	result, err := callBuiltin("sort", arr)
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 0 {
		t.Errorf("expected empty array, got %d elements", len(resultArr))
	}
}

// TestFunctionalSortSingle 测试 sort 单元素
func TestFunctionalSortSingle(t *testing.T) {
	arr := engine.NewArray([]engine.Value{
		engine.NewInt(42),
	})

	result, err := callBuiltin("sort", arr)
	if err != nil {
		t.Fatalf("sort failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 1 || resultArr[0].Int() != 42 {
		t.Errorf("expected [42], got %v", resultArr)
	}
}

// ==================== 错误处理测试 ====================

// TestFunctionalContainsWrongArgCount 测试 contains 参数数量错误
func TestFunctionalContainsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("contains", engine.NewArray([]engine.Value{}))
	if err == nil {
		t.Error("expected error for wrong argument count")
	}
}

// TestFunctionalContainsNonArray 测试 contains 非数组参数
func TestFunctionalContainsNonArray(t *testing.T) {
	_, err := callBuiltin("contains", engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("expected error for non-array argument")
	}
}

// TestFunctionalSortWrongArgCount 测试 sort 参数数量错误
func TestFunctionalSortWrongArgCount(t *testing.T) {
	_, err := callBuiltin("sort", engine.NewArray([]engine.Value{}), engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("expected error for too many arguments")
	}
}

// TestFunctionalSortNonArray 测试 sort 非数组参数
func TestFunctionalSortNonArray(t *testing.T) {
	_, err := callBuiltin("sort", engine.NewInt(1))
	if err == nil {
		t.Error("expected error for non-array argument")
	}
}

// ==================== reject 测试 ====================

// TestFunctionalRejectBasic 测试 reject 基本功能
func TestFunctionalRejectBasic(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = reject($arr, $x -> $x % 2 == 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 3, 5}
	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}
	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// TestFunctionalRejectEmpty 测试 reject 空数组
func TestFunctionalRejectEmpty(t *testing.T) {
	script := `$result = reject([], $x -> true);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("reject failed: %v", err)
	}
	if len(result.Array()) != 0 {
		t.Error("expected empty array")
	}
}

// ==================== partition 测试 ====================

// TestFunctionalPartitionBasic 测试 partition 基本功能
func TestFunctionalPartitionBasic(t *testing.T) {
	script := `$arr = [1, 2, 3, 4, 5];
$result = partition($arr, $x -> $x % 2 == 0);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("partition failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 2 {
		t.Fatalf("expected 2 arrays, got %d", len(resultArr))
	}

	matching := resultArr[0].Array()
	nonMatching := resultArr[1].Array()

	if len(matching) != 2 || matching[0].Int() != 2 || matching[1].Int() != 4 {
		t.Errorf("expected matching [2, 4], got %v", matching)
	}
	if len(nonMatching) != 3 || nonMatching[0].Int() != 1 || nonMatching[1].Int() != 3 || nonMatching[2].Int() != 5 {
		t.Errorf("expected nonMatching [1, 3, 5], got %v", nonMatching)
	}
}

// TestFunctionalPartitionEmpty 测试 partition 空数组
func TestFunctionalPartitionEmpty(t *testing.T) {
	script := `$result = partition([], $x -> true);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("partition failed: %v", err)
	}
	resultArr := result.Array()
	if len(resultArr[0].Array()) != 0 || len(resultArr[1].Array()) != 0 {
		t.Error("expected both arrays to be empty")
	}
}

// ==================== flattenDeep 测试 ====================

// TestFunctionalFlattenDeepBasic 测试 flattenDeep 基本功能
func TestFunctionalFlattenDeepBasic(t *testing.T) {
	script := `$arr = [1, [2, [3, [4, [5]]]]];
$result = flattenDeep($arr);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("flattenDeep failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 2, 3, 4, 5}
	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}
	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// TestFunctionalFlattenDeepWithDepth 测试 flattenDeep 指定深度
func TestFunctionalFlattenDeepWithDepth(t *testing.T) {
	script := `$arr = [1, [2, [3]]];
$result = flattenDeep($arr, 1);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("flattenDeep failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 3 {
		t.Errorf("expected 3 elements [1, 2, [3]], got %d", len(resultArr))
	}
}

// ==================== difference 测试 ====================

// TestFunctionalDifferenceBasic 测试 difference 基本功能
func TestFunctionalDifferenceBasic(t *testing.T) {
	script := `$a = [1, 2, 3, 4, 5];
$b = [2, 4, 6];
$result = difference($a, $b);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("difference failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 3, 5}
	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}
	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// TestFunctionalDifferenceMultiple 测试 difference 多个数组
func TestFunctionalDifferenceMultiple(t *testing.T) {
	script := `$a = [1, 2, 3, 4, 5];
$b = [2, 4];
$c = [3, 5];
$result = difference($a, $b, $c);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("difference failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 1 || resultArr[0].Int() != 1 {
		t.Errorf("expected [1], got %v", resultArr)
	}
}

// ==================== union 测试 ====================

// TestFunctionalUnionBasic 测试 union 基本功能
func TestFunctionalUnionBasic(t *testing.T) {
	script := `$a = [1, 2, 3];
$b = [3, 4, 5];
$result = union($a, $b);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("union failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 2, 3, 4, 5}
	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}
	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// TestFunctionalUnionMultiple 测试 union 多个数组
func TestFunctionalUnionMultiple(t *testing.T) {
	script := `$a = [1, 2];
$b = [2, 3];
$c = [3, 4];
$result = union($a, $b, $c);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("union failed: %v", err)
	}

	resultArr := result.Array()
	expected := []int64{1, 2, 3, 4}
	if len(resultArr) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(resultArr))
	}
	for i, v := range resultArr {
		if v.Int() != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v.Int())
		}
	}
}

// ==================== zip 测试 ====================

// TestFunctionalZipBasic 测试 zip 基本功能
func TestFunctionalZipBasic(t *testing.T) {
	script := `$names = ["Alice", "Bob", "Charlie"];
$ages = [25, 30, 35];
$result = zip($names, $ages);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("zip failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 3 {
		t.Fatalf("expected 3 tuples, got %d", len(resultArr))
	}

	for i, tuple := range resultArr {
		arr := tuple.Array()
		if len(arr) != 2 {
			t.Errorf("tuple %d: expected 2 elements, got %d", i, len(arr))
		}
	}
}

// TestFunctionalZipDifferentLengths 测试 zip 不同长度数组
func TestFunctionalZipDifferentLengths(t *testing.T) {
	script := `$a = [1, 2, 3];
$b = [10];
$result = zip($a, $b);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("zip failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 1 {
		t.Fatalf("expected 1 tuple, got %d", len(resultArr))
	}
}

// ==================== unzip 测试 ====================

// TestFunctionalUnzipBasic 测试 unzip 基本功能
func TestFunctionalUnzipBasic(t *testing.T) {
	script := `$tuples = [["Alice", 25], ["Bob", 30], ["Charlie", 35]];
$result = unzip($tuples);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 2 {
		t.Fatalf("expected 2 arrays, got %d", len(resultArr))
	}

	names := resultArr[0].Array()
	ages := resultArr[1].Array()

	if len(names) != 3 || names[0].String() != "Alice" {
		t.Errorf("expected names [Alice, Bob, Charlie], got %v", names)
	}
	if len(ages) != 3 || ages[0].Int() != 25 {
		t.Errorf("expected ages [25, 30, 35], got %v", ages)
	}
}

// TestFunctionalUnzipEmpty 测试 unzip 空数组
func TestFunctionalUnzipEmpty(t *testing.T) {
	script := `$result = unzip([]);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unzip failed: %v", err)
	}
	if len(result.Array()) != 0 {
		t.Error("expected empty result")
	}
}

// TestFunctionalUnzipInverseOfZip 测试 unzip 是 zip 的逆操作
func TestFunctionalUnzipInverseOfZip(t *testing.T) {
	script := `$a = [1, 2, 3];
$b = [10, 20, 30];
$zipped = zip($a, $b);
$result = unzip($zipped);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unzip failed: %v", err)
	}

	resultArr := result.Array()
	if len(resultArr) != 2 {
		t.Fatalf("expected 2 arrays, got %d", len(resultArr))
	}
}
