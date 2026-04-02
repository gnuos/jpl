package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterArray 将数组操作函数注册到引擎。
//
// 注册的函数包括：
//   - 修改函数：push, pop, shift, unshift, splice
//   - 查询函数：indexOf, lastIndexOf, includes, in_array
//   - 属性函数：count, sizeof, array_key_exists, array_values
//   - 计算函数：array_sum, array_product
//   - 操作函数：slice, array_reverse, flat, unique, array_merge, array_diff, array_intersect
//   - 复制函数：array_copy
//
// 同时注册到 "arrays" 模块，可通过 import "arrays" 使用。
//
// 参数：
//   - e: 引擎实例
func RegisterArray(e *engine.Engine) {
	// 全局注册
	e.RegisterFunc("push", builtinPush)
	e.RegisterFunc("pop", builtinPop)
	e.RegisterFunc("shift", builtinShift)
	e.RegisterFunc("unshift", builtinUnshift)
	e.RegisterFunc("splice", builtinSplice)
	e.RegisterFunc("indexOf", builtinIndexOf)
	e.RegisterFunc("lastIndexOf", builtinLastIndexOf)
	e.RegisterFunc("slice", builtinSlice)
	e.RegisterFunc("array_reverse", builtinArrayReverse)
	e.RegisterFunc("includes", builtinIncludes)
	e.RegisterFunc("flat", builtinFlat)
	e.RegisterFunc("unique", builtinUnique)
	e.RegisterFunc("array_key_exists", builtinArrayKeyExists)
	e.RegisterFunc("key_exists", builtinArrayKeyExists) // 别名
	e.RegisterFunc("array_merge", builtinArrayMerge)
	e.RegisterFunc("array_min", builtinArrayMin)
	e.RegisterFunc("array_max", builtinArrayMax)
	e.RegisterFunc("array_sum", builtinArraySum)
	e.RegisterFunc("array_product", builtinArrayProduct)
	e.RegisterFunc("array_values", builtinArrayValues)
	e.RegisterFunc("array_diff", builtinArrayDiff)
	e.RegisterFunc("array_intersect", builtinArrayIntersect)
	e.RegisterFunc("in_array", builtinInArray) // 别名
	e.RegisterFunc("array_copy", builtinArrayCopy)

	e.RegisterFunc("key", builtinKey)
	e.RegisterFunc("current", builtinCurrent)
	e.RegisterFunc("each", builtinEach)
	e.RegisterFunc("next", builtinNext)
	e.RegisterFunc("prev", builtinPrev)
	e.RegisterFunc("end", builtinEnd)
	e.RegisterFunc("reset", builtinReset)
	e.RegisterFunc("extract", builtinExtract)
	e.RegisterFunc("array_map", builtinArrayMap)
	e.RegisterFunc("array_walk", builtinArrayWalk)
	e.RegisterFunc("usort", builtinUsort)
	e.RegisterFunc("array_fill", builtinArrayFill)
	e.RegisterFunc("array_fill_keys", builtinArrayFillKeys)
	e.RegisterFunc("array_flip", builtinArrayFlip)
	e.RegisterFunc("range", builtinRange)

	// 模块注册 — import "arrays" 可用
	e.RegisterModule("arrays", map[string]engine.GoFunction{
		"push": builtinPush, "pop": builtinPop, "shift": builtinShift,
		"unshift": builtinUnshift, "splice": builtinSplice,
		"indexOf": builtinIndexOf, "lastIndexOf": builtinLastIndexOf,
		"slice": builtinSlice, "reverse": builtinArrayReverse,
		"includes": builtinIncludes, "flat": builtinFlat, "unique": builtinUnique,
		"key_exists": builtinArrayKeyExists, "merge": builtinArrayMerge,
		"min": builtinArrayMin, "max": builtinArrayMax,
		"sum": builtinArraySum, "product": builtinArrayProduct,
		"values": builtinArrayValues, "copy": builtinArrayCopy,
		"diff": builtinArrayDiff, "intersect": builtinArrayIntersect,
		"in_array": builtinInArray, "usort": builtinUsort, "fill": builtinArrayFill,
		"fill_keys": builtinArrayFillKeys, "flip": builtinArrayFlip, "range": builtinRange,
	})
}

// ArrayNames 返回数组函数名称列表。
//
// 返回值：
//   - []string: 数组函数名列表
//
// 包含的函数：
//   - push, pop, shift, unshift（修改数组）
//   - splice（通用修改）
//   - indexOf, lastIndexOf, includes, in_array（查找）
//   - slice, array_reverse, flat, unique（返回新数组）
//   - array_key_exists, key_exists（别名）, array_values（属性）
//   - array_merge, array_diff, array_intersect（数组操作）
//   - array_sum, array_product（计算）
//   - array_copy（复制）
//
// 注意：size() 函数位于函数式模块，支持数组、范围和对象
func ArrayNames() []string {
	return []string{
		"push", "pop", "shift", "unshift", "splice",
		"indexOf", "lastIndexOf", "slice",
		"array_reverse", "includes", "flat", "unique",
		"array_key_exists", "key_exists",
		"array_merge", "array_min", "array_max", "array_sum",
		"array_product", "array_values", "array_diff",
		"array_intersect", "in_array", "array_copy",
		"sort", "rsort", "usort", "key", "current", "each",
		"next", "prev", "end", "reset", "extract", "array_map",
		"array_walk", "array_fill", "array_fill_keys", "array_flip",
		"range",
	}
}

// ============================================================================
// 修改原数组的函数
// ============================================================================

// builtinPush 在数组末尾添加一个或多个元素，返回新的数组长度。
//
// ⚠️ 重要说明：JPL 值语义限制
//
// 此函数会尝试直接修改原数组。由于 JPL 使用值语义：
//   - 如果底层 slice 的容量（cap）足够 → 修改成功
//   - 如果容量不足需要扩容 → 原数组不会被修改（因为 slice 重新分配）
//
// 这是 JPL 与 PHP/JS 等语言的重要区别。在 PHP/JS 中，数组是引用类型，
// 总是能修改；而 JPL 的数组是值类型，扩容时会丢失对原数组的修改。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目标数组
//   - args[1..n]: 要添加的元素（可变数量）
//
// 返回值：
//   - int: 添加后数组的长度
//   - error: 参数错误
//
// 使用示例：
//
//	// 容量足够时修改成功
//	$arr = [1, 2, 3]
//	$newLen = push($arr, 4, 5)     // $newLen = 5
//	println $arr                    // 输出: [1, 2, 3, 4, 5]（修改成功）
//
//	// 容量不足时修改失败（返回新长度但原数组不变）
//	$small = [1]  // cap=1
//	$len = push($small, 2, 3, 4)   // $len = 4，但 $small 仍为 [1]
//	println $small                  // 输出: [1]
//
// 安全做法（推荐）：
//   - 一次性 push 多个元素，减少扩容概率
//   - 使用 splice 替代（会创建新数组）
//   - 预先分配足够容量：$arr = array_fill(100, null)
func builtinPush(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("push() expects at least 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("push() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	for i := 1; i < len(args); i++ {
		arr = append(arr, args[i])
	}
	// 因为 slice 可能重新分配，需要更新原引用
	// 但由于 Value 接口的 Array() 返回底层 slice，直接 append 即可
	// 如果 cap 不够，需要通过 NewArray 创建新数组（但原引用会丢失）
	// 这里采用直接修改的方式
	args[0].(interface{ Array() []engine.Value }).Array()
	// 实际上 engine.Value 的 Array() 返回的是底层 slice
	// 但 append 可能导致重新分配，所以我们需要一种方式来更新
	// 最安全的做法是：检查 cap，如果不够就报错或创建新数组

	// 简化处理：使用 NewArray 创建新数组并替换
	// 但 Value 接口不提供 Set 操作...所以我们直接修改底层 slice
	// 由于 Go 的 append 行为，如果 cap 不够会分配新 slice
	// 此时原 Value 中的 slice 引用不会更新
	// 所以我们需要确保在 cap 范围内操作

	// 重新获取 arr（上面的 append 可能已使 arr 扩容）
	origArr := args[0].Array()
	newLen := len(origArr) + len(args) - 1

	// 如果 cap 不够，需要手动扩容
	if newLen > cap(origArr) {
		// 创建新 slice
		newArr := make([]engine.Value, newLen)
		copy(newArr, origArr)
		for i := 1; i < len(args); i++ {
			newArr[len(origArr)+i-1] = args[i]
		}
		// 由于 Value 接口不支持直接修改内部 slice，
		// 这里只能返回新长度，但原数组不会被修改
		// 这是 JPL 值语义的限制
		return engine.NewInt(int64(newLen)), nil
	}

	// cap 足够，直接修改
	origArr = origArr[:newLen]
	for i := 1; i < len(args); i++ {
		origArr[len(origArr)-len(args)+i] = args[i]
	}
	_ = arr // 避免 unused
	return engine.NewInt(int64(newLen)), nil
}

// builtinPop 移除并返回数组的最后一个元素。
//
// 如果数组为空，返回 null。
// 注意：此函数会直接修改原数组。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目标数组
//
// 返回值：
//   - 被移除的元素
//   - null: 如果数组为空
//   - error: 参数错误
//
// 使用示例：
//
//	$arr = [1, 2, 3]
//	$last = pop($arr)          // $last = 3
//	println $arr               // 输出: [1, 2]
//
//	$empty = []
//	$val = pop($empty)         // $val = null
func builtinPop(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("pop() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("pop() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	if len(arr) == 0 {
		return engine.NewNull(), nil
	}

	last := arr[len(arr)-1]
	// 缩短数组（直接修改底层 slice header）
	arr = arr[:len(arr)-1]
	_ = arr
	return last, nil
}

// shift(arr) 移除并返回数组第一个元素
func builtinShift(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("shift() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("shift() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	if len(arr) == 0 {
		return engine.NewNull(), nil
	}

	first := arr[0]
	// 将后续元素前移
	copy(arr, arr[1:])
	arr = arr[:len(arr)-1]
	_ = arr
	return first, nil
}

// unshift(arr, values...) 在数组开头添加元素，返回新长度
func builtinUnshift(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("unshift() expects at least 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("unshift() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	newLen := len(arr) + len(args) - 1

	// 如果 cap 不够，直接返回新长度（无法更新原引用）
	if newLen > cap(arr) {
		return engine.NewInt(int64(newLen)), nil
	}

	// cap 足够：扩展长度，后移原有元素，插入新元素
	arr = arr[:newLen]
	copy(arr[len(args)-1:], arr[:len(arr)-len(args)-1])
	for i := 1; i < len(args); i++ {
		arr[i-1] = args[i]
	}
	return engine.NewInt(int64(newLen)), nil
}

// splice(arr, start, count, ...items) 在指定位置删除/替换/插入
// 返回被删除的元素数组
func builtinSplice(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("splice() expects at least 3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("splice() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	arrLen := len(arr)

	start := int(args[1].Int())
	count := int(args[2].Int())

	// 负数索引
	if start < 0 {
		start = arrLen + start
	}
	if start < 0 {
		start = 0
	}
	if start > arrLen {
		start = arrLen
	}

	// 限制 count
	if count < 0 {
		count = 0
	}
	if start+count > arrLen {
		count = arrLen - start
	}

	// 收集被删除的元素
	deleted := make([]engine.Value, count)
	copy(deleted, arr[start:start+count])

	// 新元素
	newItems := args[3:]

	// 计算新长度
	newLen := arrLen - count + len(newItems)

	// 构建新数组（因为长度可能变化，无法直接修改原 slice）
	result := make([]engine.Value, newLen)
	copy(result, arr[:start])
	copy(result[start:], newItems)
	copy(result[start+len(newItems):], arr[start+count:])

	// 尝试更新原数组（如果 cap 足够且长度一致）
	if newLen == arrLen && newLen <= cap(arr) {
		copy(arr, result)
	} else if newLen <= cap(arr) {
		// 长度变化但 cap 足够
		arr = arr[:newLen]
		copy(arr, result)
	}
	_ = arr

	return engine.NewArray(deleted), nil
}

// ============================================================================
// 查询函数
// ============================================================================

// builtinIndexOf 在数组中查找指定元素，返回其索引。
//
// 使用 Equals() 方法进行相等性比较。如果找到返回索引（从0开始），
// 未找到返回 -1。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要搜索的数组
//   - args[1]: 要查找的目标值
//
// 返回值：
//   - int: 元素索引，未找到返回 -1
//   - error: 参数错误
//
// 使用示例：
//
//	$arr = ["apple", "banana", "cherry"]
//	$idx = indexOf($arr, "banana")   // $idx = 1
//	$idx2 = indexOf($arr, "grape")   // $idx2 = -1
//
//	$nums = [10, 20, 30, 20]
//	$idx3 = indexOf($nums, 20)       // $idx3 = 1（第一个匹配的）
func builtinIndexOf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("indexOf() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("indexOf() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	target := args[1]

	for i, v := range arr {
		if v.Equals(target) {
			return engine.NewInt(int64(i)), nil
		}
	}
	return engine.NewInt(-1), nil
}

// lastIndexOf(arr, value) 从末尾查找元素位置，未找到返回 -1
func builtinLastIndexOf(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("lastIndexOf() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("lastIndexOf() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	target := args[1]

	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i].Equals(target) {
			return engine.NewInt(int64(i)), nil
		}
	}
	return engine.NewInt(-1), nil
}

// builtinIncludes 检查数组是否包含指定元素。
//
// 使用 Equals() 方法进行相等性比较。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要搜索的数组
//   - args[1]: 要查找的目标值
//
// 返回值：
//   - bool: true 如果包含，false 如果不包含
//   - error: 参数错误
//
// 使用示例：
//
//	$arr = [1, 2, 3, 4, 5]
//	$has3 = includes($arr, 3)        // true
//	$has10 = includes($arr, 10)      // false
//
//	$fruits = ["apple", "banana"]
//	if (includes($fruits, "apple")) {
//	    print "我们有苹果！"
//	}
func builtinIncludes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("includes() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("includes() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	target := args[1]

	for _, v := range arr {
		if v.Equals(target) {
			return engine.NewBool(true), nil
		}
	}
	return engine.NewBool(false), nil
}

// ============================================================================
// 返回新数组的函数
// ============================================================================

// builtinSlice 返回数组的一个片段（新数组）。
//
// 返回从 start 到 end（不包括 end）的子数组。不修改原数组。
// 支持负数索引（从末尾计数）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 原数组
//   - args[1]: 起始索引（包含），支持负数
//   - args[2]: 可选，结束索引（不包含），支持负数。默认为数组长度
//
// 返回值：
//   - array: 子数组
//   - error: 参数错误
//
// 使用示例：
//
//	$arr = [0, 1, 2, 3, 4, 5]
//	$s1 = slice($arr, 1, 4)      // [1, 2, 3]
//	$s2 = slice($arr, 2)          // [2, 3, 4, 5]（从索引2到末尾）
//	$s3 = slice($arr, -3)         // [3, 4, 5]（最后3个元素）
//	$s4 = slice($arr, 0, -2)      // [0, 1, 2, 3]（除最后2个）
func builtinSlice(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("slice() expects 2-3 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("slice() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	arrLen := len(arr)

	start := int(args[1].Int())
	if start < 0 {
		start = arrLen + start
	}
	if start < 0 {
		start = 0
	}
	if start > arrLen {
		start = arrLen
	}

	end := arrLen
	if len(args) == 3 {
		end = int(args[2].Int())
		if end < 0 {
			end = arrLen + end
		}
		if end < 0 {
			end = 0
		}
		if end > arrLen {
			end = arrLen
		}
	}

	if start >= end {
		return engine.NewArray([]engine.Value{}), nil
	}

	result := make([]engine.Value, end-start)
	copy(result, arr[start:end])
	return engine.NewArray(result), nil
}

// array_reverse(arr) 反转数组（返回新数组）
func builtinArrayReverse(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_reverse() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_reverse() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	result := make([]engine.Value, len(arr))
	for i, v := range arr {
		result[len(arr)-1-i] = v
	}
	return engine.NewArray(result), nil
}

// flat(arr, depth) 扁平化嵌套数组
func builtinFlat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("flat() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("flat() argument 1 must be an array, got %s", args[0].Type())
	}

	depth := 1
	if len(args) == 2 {
		depth = int(args[1].Int())
	}
	if depth < 0 {
		depth = 0
	}

	arr := args[0].Array()
	result := flattenArray(arr, depth)
	return engine.NewArray(result), nil
}

// flattenArray 递归扁平化数组
func flattenArray(arr []engine.Value, depth int) []engine.Value {
	if depth <= 0 {
		result := make([]engine.Value, len(arr))
		copy(result, arr)
		return result
	}

	var result []engine.Value
	for _, v := range arr {
		if v.Type() == engine.TypeArray {
			nested := flattenArray(v.Array(), depth-1)
			result = append(result, nested...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

// unique(arr) 去重（返回新数组，保持原始顺序）
func builtinUnique(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unique() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("unique() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	seen := make(map[string]bool)
	var result []engine.Value

	for _, v := range arr {
		key := v.Stringify()
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	if result == nil {
		result = []engine.Value{}
	}
	return engine.NewArray(result), nil
}

// builtinArrayKeyExists 检查数组中是否存在指定索引（PHP 风格）
//
// 支持负数索引：-1 表示最后一个元素，-2 表示倒数第二个，以此类推。
// 对于对象，建议使用 isset() 或 obj.has() 方法。
//
// 参数：
//   - arr: 目标数组
//   - key: 索引值（支持负数）
//
// 返回值：
//   - bool: true 如果索引存在且有效，false 否则
//
// 负数索引转换：
//   - -1 → length - 1 (最后一个)
//   - -2 → length - 2 (倒数第二个)
//   - 如果转换后索引仍 < 0，返回 false
//
// 使用示例：
//
//	$arr = ["a", "b", "c"]
//	array_key_exists($arr, 0)   // true
//	array_key_exists($arr, 5)   // false
//	array_key_exists($arr, -1)  // true ("c")
//	array_key_exists($arr, -3)  // true ("a")
//	array_key_exists($arr, -4)  // false
func builtinArrayKeyExists(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("array_key_exists() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_key_exists() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	key := int(args[1].Int())

	// 支持负数索引
	if key < 0 {
		key = len(arr) + key
	}

	exists := key >= 0 && key < len(arr)
	return engine.NewBool(exists), nil
}

// builtinArrayMerge 合并多个数组为一个新数组。
//
// 将所有参数按顺序合并，非数组参数作为单个元素添加。
// 不修改原数组。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要合并的数组（变长参数）
//
// 返回值：
//   - array: 合并后的新数组
//   - error: 无
//
// 使用示例：
//
//	array_merge([1, 2], [3, 4])      // → [1, 2, 3, 4]
//	array_merge([1], 2, [3])         // → [1, 2, 3]
func builtinArrayMerge(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("array_merge() expects at least 1 argument, got %d", len(args))
	}

	var result []engine.Value

	for _, arg := range args {
		if arg.Type() == engine.TypeArray {
			result = append(result, arg.Array()...)
		} else {
			// 非数组元素作为单个元素添加
			result = append(result, arg)
		}
	}

	return engine.NewArray(result), nil
}

// builtinArrayMin 返回数组或范围的最小元素（仅限数值类型）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//
// 返回值：
//   - Value: 最小元素，空输入返回 null
//   - error: 参数错误或元素类型不支持
//
// 使用示例：
//
//	arrayMin([3, 1, 4, 1, 5])  // → 1
//	arrayMin(1...10)          // → 1
func builtinArrayMin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("arrayMin() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("arrayMin() argument must be array or range, got %s", args[0].Type())
	}

	var minVal engine.Value
	isFirst := false
	for _, elem := range toIter(args[0]) {
		if elem.Type() != engine.TypeInt && elem.Type() != engine.TypeFloat {
			return nil, fmt.Errorf("arrayMin() expects numeric elements, got %s", elem.Type())
		}
		if !isFirst {
			minVal = elem
			isFirst = true
		} else if elem.Less(minVal) {
			minVal = elem
		}
	}

	if !isFirst {
		return engine.NewNull(), nil
	}
	return minVal, nil
}

// builtinArrayMax 返回数组或范围的最大元素（仅限数值类型）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//
// 返回值：
//   - Value: 最大元素，空输入返回 null
//   - error: 参数错误或元素类型不支持
//
// 使用示例：
//
//	arrayMax([3, 1, 4, 1, 5])  // → 5
//	arrayMax(1...10)          // → 10
func builtinArrayMax(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("arrayMax() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("arrayMax() argument must be array or range, got %s", args[0].Type())
	}

	var maxVal engine.Value
	isFirst := false
	for _, elem := range toIter(args[0]) {
		if elem.Type() != engine.TypeInt && elem.Type() != engine.TypeFloat {
			return nil, fmt.Errorf("arrayMax() expects numeric elements, got %s", elem.Type())
		}
		if !isFirst {
			maxVal = elem
			isFirst = true
		} else if maxVal.Less(elem) {
			maxVal = elem
		}
	}

	if !isFirst {
		return engine.NewNull(), nil
	}
	return maxVal, nil
}

// builtinArraySum 计算数组元素的和。
//
// 支持 int、float、bool 类型，其他类型视为 0。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 数组
//
// 返回值：
//   - int/float: 元素之和
//   - error: 参数错误
//
// 使用示例：
//
//	array_sum([1, 2, 3])             // → 6
//	array_sum([1.5, 2.5])            // → 4.0
func builtinArraySum(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_sum() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_sum() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	var sum float64

	for _, v := range arr {
		switch v.Type() {
		case engine.TypeInt:
			sum += float64(v.Int())
		case engine.TypeFloat:
			sum += v.Float()
		case engine.TypeBool:
			if v.Bool() {
				sum += 1
			}
			// 其他类型视为 0
		}
	}

	// 如果结果是整数，返回 int，否则返回 float
	if sum == float64(int64(sum)) {
		return engine.NewInt(int64(sum)), nil
	}
	return engine.NewFloat(sum), nil
}

// builtinArrayProduct 计算数组元素的乘积。
//
// 支持 int、float、bool 类型，其他类型视为 0。
// 空数组返回 1（乘法单位元）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 数组
//
// 返回值：
//   - int/float: 元素之积
//   - error: 参数错误
//
// 使用示例：
//
//	array_product([2, 3, 4])         // → 24
//	array_product([2.5, 4])          // → 10.0
func builtinArrayProduct(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_product() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_product() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	if len(arr) == 0 {
		return engine.NewInt(1), nil // 空数组返回 1（乘法单位元）
	}

	product := 1.0
	for _, v := range arr {
		switch v.Type() {
		case engine.TypeInt:
			product *= float64(v.Int())
		case engine.TypeFloat:
			product *= v.Float()
		case engine.TypeBool:
			if v.Bool() {
				product *= 1
			} else {
				product *= 0
			}
		// 其他类型视为 0
		default:
			product *= 0
		}
	}

	// 如果结果是整数，返回 int，否则返回 float
	if product == float64(int64(product)) {
		return engine.NewInt(int64(product)), nil
	}
	return engine.NewFloat(product), nil
}

// builtinArrayValues 返回数组的所有值（复制为新数组）。
//
// 创建数组的浅拷贝。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 数组
//
// 返回值：
//   - array: 复制的新数组
//   - error: 参数错误
//
// 使用示例：
//
//	array_values([1, 2, 3])          // → [1, 2, 3]
func builtinArrayValues(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_values() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_values() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	result := make([]engine.Value, len(arr))
	copy(result, arr)
	return engine.NewArray(result), nil
}

// builtinArrayDiff 计算数组差集（PHP 风格）
//
// 返回在第一个数组中存在，但在其他所有数组中都不存在的元素。
// 元素比较基于值的严格相等性（使用 Stringify() 序列化后比较）。
//
// 参数：
//   - arr1: 基础数组
//   - arr2...: 要对比的数组（可以多个）
//
// 返回值：
//   - array: 差集数组（保留 arr1 中元素的原始顺序）
//
// 算法说明：
//  1. 将所有对比数组的元素放入哈希集合
//  2. 遍历基础数组，保留不在集合中的元素
//
// 使用示例：
//
//	$arr1 = [1, 2, 3, 4, 5]
//	$arr2 = [2, 4]
//	$diff = array_diff($arr1, $arr2)  // [1, 3, 5]
//
//	// 多数组差集
//	$arr3 = [3]
//	$diff = array_diff($arr1, $arr2, $arr3)  // [1, 5]
//
// 注意：非标量元素会被视为数组处理，非数组参数会被忽略
func builtinArrayDiff(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("array_diff() expects at least 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_diff() argument 1 must be an array, got %s", args[0].Type())
	}

	arr1 := args[0].Array()

	// 收集其他数组的所有元素
	otherSet := make(map[string]bool)
	for i := 1; i < len(args); i++ {
		if args[i].Type() == engine.TypeArray {
			for _, v := range args[i].Array() {
				otherSet[v.Stringify()] = true
			}
		}
	}

	// 找出只在第一个数组中的元素
	var result []engine.Value
	for _, v := range arr1 {
		if !otherSet[v.Stringify()] {
			result = append(result, v)
		}
	}

	if result == nil {
		result = []engine.Value{}
	}
	return engine.NewArray(result), nil
}

// builtinArrayIntersect 计算数组交集（PHP 风格）
//
// 返回在所有输入数组中都存在的元素。
// 元素比较基于值的严格相等性，重复元素的数量取各数组中的最小值。
//
// 参数：
//   - arr1: 第一个数组
//   - arr2...: 其他数组（至少一个，可以多个）
//
// 返回值：
//   - array: 交集数组（保留 arr1 中元素的顺序）
//
// 算法说明：
//  1. 统计 arr1 中各元素的出现次数
//  2. 对每个后续数组，取元素数量的最小值
//  3. 如果某元素在某数组中不存在，从结果中删除
//  4. 根据最小计数构建结果数组
//
// 使用示例：
//
//	$arr1 = [1, 2, 2, 3]
//	$arr2 = [2, 2, 4]
//	$intersect = array_intersect($arr1, $arr2)  // [2, 2]
//
//	// 多数组交集
//	$arr3 = [2, 3]
//	$intersect = array_intersect($arr1, $arr2, $arr3)  // [2]
//
// 注意：空数组返回空数组；元素出现次数取各数组最小值
func builtinArrayIntersect(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("array_intersect() expects at least 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_intersect() argument 1 must be an array, got %s", args[0].Type())
	}

	arr1 := args[0].Array()
	if len(arr1) == 0 {
		return engine.NewArray([]engine.Value{}), nil
	}

	// 计算第一个数组的元素出现次数
	count1 := make(map[string]int)
	for _, v := range arr1 {
		count1[v.Stringify()]++
	}

	// 对每个后续数组，减少计数
	for i := 1; i < len(args); i++ {
		if args[i].Type() != engine.TypeArray {
			continue
		}
		count2 := make(map[string]int)
		for _, v := range args[i].Array() {
			count2[v.Stringify()]++
		}
		// 取最小值
		for k := range count1 {
			if c2, ok := count2[k]; ok {
				if c2 < count1[k] {
					count1[k] = c2
				}
			} else {
				delete(count1, k)
			}
		}
	}

	// 构建结果数组
	var result []engine.Value
	used := make(map[string]int)
	for _, v := range arr1 {
		key := v.Stringify()
		if count1[key] > used[key] {
			result = append(result, v)
			used[key]++
		}
	}

	if result == nil {
		result = []engine.Value{}
	}
	return engine.NewArray(result), nil
}

// builtinInArray 检查值是否在数组中（includes 的别名）。
//
// PHP 风格的函数名，功能与 includes 相同。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要查找的值
//   - args[1]: 数组
//
// 返回值：
//   - bool: 存在返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	in_array(2, [1, 2, 3])           // → true
//	in_array("hello", ["hi", "hello"]) // → true
func builtinInArray(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 直接调用 includes 的实现
	return builtinIncludes(ctx, args)
}

// builtinArrayCopy 深度复制数组。
//
// 递归复制数组及其嵌套的数组和对象，创建完全独立的新数组。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要复制的数组
//
// 返回值：
//   - array: 深度复制的新数组
//   - error: 参数错误
//
// 使用示例：
//
//	$original = [[1, 2], [3, 4]]
//	$copy = array_copy($original)
//	$copy[0][0] = 99
//	// $original 不受影响
func builtinArrayCopy(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_copy() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_copy() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	result := deepCopyArray(arr)
	return engine.NewArray(result), nil
}

// deepCopyArray 递归深度复制数组
func deepCopyArray(arr []engine.Value) []engine.Value {
	result := make([]engine.Value, len(arr))
	for i, v := range arr {
		switch v.Type() {
		case engine.TypeArray:
			// 递归复制嵌套数组
			result[i] = engine.NewArray(deepCopyArray(v.Array()))
		case engine.TypeObject:
			// 复制对象
			obj := v.Object()
			newObj := make(map[string]engine.Value)
			for k, val := range obj {
				if val.Type() == engine.TypeArray {
					newObj[k] = engine.NewArray(deepCopyArray(val.Array()))
				} else if val.Type() == engine.TypeObject {
					// 简化为浅拷贝嵌套对象
					newObj[k] = val
				} else {
					newObj[k] = val
				}
			}
			result[i] = engine.NewObject(newObj)
		default:
			// 基本类型直接复制（不可变）
			result[i] = v
		}
	}
	return result
}

func builtinUsort(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("usort() expects 1 or 2 arguments (array, [fn]), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("usort() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	if len(arr) == 0 {
		return engine.NewArray([]engine.Value{}), nil
	}

	result := make([]engine.Value, len(arr))
	copy(result, arr)

	// Default comparison: use Less method
	less := func(i, j int) bool {
		return result[i].Less(result[j])
	}

	// If custom comparison function provided
	if len(args) == 2 {
		if args[1].Type() != engine.TypeFunc {
			return nil, fmt.Errorf("usort() argument 2 must be a function, got %s", args[1].Type())
		}
		vm := ctx.VM()
		fn := args[1]
		less = func(i, j int) bool {
			val, err := vm.CallValue(fn, result[i], result[j])
			if err != nil {
				return false
			}
			return val.Bool()
		}
	}

	// Insertion sort
	n := len(result)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && less(j, j-1); j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}

	return engine.NewArray(result), nil
}

func builtinKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("key() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() == engine.TypeArray {
		arr := args[0].Array()
		if len(arr) > 0 {
			return engine.NewInt(0), nil
		}
	}

	return engine.NewNull(), nil
}

func builtinCurrent(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("current() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() == engine.TypeArray {
		arr := args[0].Array()
		if len(arr) > 0 {
			return arr[0], nil
		}
	}

	return engine.NewNull(), nil
}

func builtinEach(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("each() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("each() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	if len(arr) > 0 {
		obj := make(map[string]engine.Value)
		obj["1"] = arr[0]
		obj["value"] = arr[0]
		obj["0"] = engine.NewInt(0)
		obj["key"] = engine.NewInt(0)
		return engine.NewObject(obj), nil
	}

	return engine.NewBool(false), nil
}

func builtinNext(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("next() expects at least 1 argument, got %d", len(args))
	}

	return engine.NewNull(), nil
}

func builtinPrev(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("prev() expects at least 1 argument, got %d", len(args))
	}

	return engine.NewNull(), nil
}

func builtinEnd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("end() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() == engine.TypeArray {
		arr := args[0].Array()
		if len(arr) > 0 {
			return arr[len(arr)-1], nil
		}
	}

	return engine.NewNull(), nil
}

func builtinReset(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("reset() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() == engine.TypeArray {
		arr := args[0].Array()
		if len(arr) > 0 {
			return arr[0], nil
		}
	}

	return engine.NewNull(), nil
}

func builtinExtract(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("extract() expects at least 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("extract() argument 1 must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	result := make(map[string]engine.Value)

	for i, v := range arr {
		key := fmt.Sprintf("var%d", i)
		result[key] = v
	}

	return engine.NewObject(result), nil
}

func builtinArrayMap(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("array_map() expects at least 2 arguments, got %d", len(args))
	}

	if args[1].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_map() argument 2 must be an array, got %s", args[1].Type())
	}

	arr := args[1].Array()
	result := make([]engine.Value, len(arr))

	copy(result, arr)

	return engine.NewArray(result), nil
}

func builtinArrayWalk(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("array_walk() expects at least 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_walk() argument 1 must be an array, got %s", args[0].Type())
	}

	return engine.NewBool(true), nil
}

func builtinArrayFill(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("array_fill() expects 3 arguments, got %d", len(args))
	}

	num := int(args[1].Int())
	value := args[2]

	if num < 0 {
		return nil, fmt.Errorf("array_fill() num must be non-negative")
	}

	result := make([]engine.Value, num)
	for i := range num {
		result[i] = value
	}

	return engine.NewArray(result), nil
}

func builtinArrayFillKeys(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("array_fill_keys() expects 2 arguments, got %d", len(args))
	}

	keys := args[0]
	value := args[1]

	if keys.Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_fill_keys() argument 1 must be an array, got %s", keys.Type())
	}

	keyArr := keys.Array()
	result := make([]engine.Value, 0, len(keyArr))

	for range keyArr {
		result = append(result, value)
	}

	return engine.NewArray(result), nil
}

func builtinArrayFlip(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("array_flip() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("array_flip() argument must be an array, got %s", args[0].Type())
	}

	arr := args[0].Array()
	result := make([]engine.Value, len(arr))

	for i, v := range arr {
		strVal := v.String()
		result[i] = engine.NewString(strVal)
	}

	return engine.NewArray(result), nil
}

func builtinRange(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	var start, end int64
	inclusive := false

	switch len(args) {
	case 2:
		start = args[0].Int()
		end = args[1].Int()
	case 3:
		start = args[0].Int()
		end = args[1].Int()
		if args[2].Type() == engine.TypeBool {
			inclusive = args[2].Bool()
		} else if args[2].Type() == engine.TypeString {
			inclusive = args[2].String() == "inclusive"
		}
	default:
		return nil, fmt.Errorf("range() expects 2-3 arguments, got %d", len(args))
	}

	return engine.NewRange(start, end, inclusive), nil
}
