package stdlib

import (
	"fmt"
	"iter"

	"github.com/gnuos/jpl/engine"
)

func toIter(v engine.Value) iter.Seq2[int, engine.Value] {
	return func(yield func(int, engine.Value) bool) {
		switch v.Type() {
		case engine.TypeArray:
			for i, val := range v.Array() {
				if !yield(i, val) {
					return
				}
			}
		case engine.TypeRange:
			rv := v.(interface {
				Start() int64
				End() int64
				IsInclusive() bool
			})
			start, end := rv.Start(), rv.End()
			inclusive := rv.IsInclusive()
			for i := start; i < end || (inclusive && i == end); i++ {
				if !yield(int(i-start), engine.NewInt(i)) {
					return
				}
			}
		}
	}
}

func getLength(v engine.Value) (int, bool) {
	switch v.Type() {
	case engine.TypeArray:
		return len(v.Array()), true
	case engine.TypeRange:
		rv := v.(interface {
			Start() int64
			End() int64
			IsInclusive() bool
		})
		start, end := rv.Start(), rv.End()
		inclusive := rv.IsInclusive()
		var length int64
		if inclusive {
			length = end - start + 1
		} else {
			length = end - start
		}
		if length < 0 {
			return 0, true
		}
		return int(length), true
	}
	return 0, false
}

// RegisterFunctional 注册函数式编程函数到引擎。
//
// 注册的函数（支持数组和 Range）：
//   - map: 对每个元素应用函数，返回新数组
//   - filter: 过滤，保留满足条件的元素
//   - reject: 排除满足条件的元素
//   - reduce: 归约为单个值
//   - find: 查找第一个满足条件的元素
//   - some: 检查是否有元素满足条件
//   - every: 检查是否所有元素都满足条件
//   - contains: 检查是否包含某值
//   - reject: 排除满足条件的元素
//   - partition: 按条件分成两组
//
// 注册的函数（仅支持数组）：
//   - sort: 排序数组
//   - unique: 去除重复元素
//   - flattenDeep: 深度展平数组
//   - difference: 差集运算
//   - union: 并集运算
//   - zip: 合并多个数组
//   - unzip: 拆分元组数组
//
// 注册的函数（支持数组和 Range）：
//   - first: 返回第一个元素
//   - last: 返回最后一个元素
//   - take: 取前 n 个元素
//   - drop: 跳过前 n 个元素
//   - sum: 求和（数值类型）
//   - arrayMin: 最小值（数值类型）
//   - arrayMax: 最大值（数值类型）
//   - size: 元素数量
//
// 参数：
//   - e: 引擎实例
func RegisterFunctional(e *engine.Engine) {
	e.RegisterFunc("map", builtinMap)
	e.RegisterFunc("filter", builtinFilter)
	e.RegisterFunc("reject", builtinReject)
	e.RegisterFunc("reduce", builtinReduce)
	e.RegisterFunc("find", builtinFind)
	e.RegisterFunc("some", builtinSome)
	e.RegisterFunc("every", builtinEvery)
	e.RegisterFunc("sort", builtinSort)
	e.RegisterFunc("contains", builtinContains)
	e.RegisterFunc("unique", builtinUnique)
	e.RegisterFunc("partition", builtinPartition)
	e.RegisterFunc("flattenDeep", builtinFlattenDeep)
	e.RegisterFunc("difference", builtinDifference)
	e.RegisterFunc("union", builtinUnion)
	e.RegisterFunc("zip", builtinZip)
	e.RegisterFunc("unzip", builtinUnzip)
	e.RegisterFunc("first", builtinFirst)
	e.RegisterFunc("last", builtinLast)
	e.RegisterFunc("take", builtinTake)
	e.RegisterFunc("drop", builtinDrop)
	e.RegisterFunc("sum", builtinSum)
	e.RegisterFunc("size", builtinSize)
}

// FunctionalNames 返回函数式编程函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func FunctionalNames() []string {
	return []string{
		"map", "filter", "reject", "reduce", "find", "some", "every",
		"sort", "contains", "unique", "partition", "flattenDeep",
		"difference", "union", "zip", "unzip", "first", "last",
		"take", "drop", "sum", "size",
	}
}

// builtinMap 对数组每个元素应用函数，返回新数组。
//
// 遍历数组的每个元素，调用回调函数并将返回值组成新数组。
// 原数组不会被修改。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → newValue
//
// 返回值：
//   - array: 映射后的新数组
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3]
//	$doubled = map($nums, ($x) -> $x * 2)    // → [2, 4, 6]
//	$strs = map([1, 2, 3], ($x) -> str($x))  // → ["1", "2", "3"]
func builtinMap(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("map() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("map() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("map() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	var result []engine.Value
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("map() callback error: %w", err)
		}
		result = append(result, val)
	}

	return engine.NewArray(result), nil
}

// builtinFilter 过滤数组，保留回调函数返回 true 的元素。
//
// 遍历数组的每个元素，调用回调函数，保留返回值为 true 的元素。
// 原数组不会被修改。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - array: 过滤后的新数组
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	$evens = filter($nums, ($x) -> $x % 2 == 0)  // → [2, 4]
//	$big = filter($nums, ($x) -> $x > 3)          // → [4, 5]
func builtinFilter(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("filter() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("filter() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("filter() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	var result []engine.Value
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("filter() callback error: %w", err)
		}
		if val.Bool() {
			result = append(result, elem)
		}
	}

	return engine.NewArray(result), nil
}

// builtinReduce 将数组归约为单个值。
//
// 从左到右遍历数组，将累积值和当前元素传入回调函数，
// 用回调函数的返回值更新累积值。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(accumulator, element) → newAccumulator
//   - args[2]: 可选的初始值（不提供则使用数组第一个元素）
//
// 返回值：
//   - Value: 归约后的最终累积值
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	$sum = reduce($nums, ($acc, $x) -> $acc + $x, 0)    // → 15
//	$prod = reduce($nums, ($acc, $x) -> $acc * $x, 1)   // → 120
//	$max = reduce($nums, ($acc, $x) -> $x > $acc ? $x : $acc)  // → 5
func builtinReduce(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("reduce() expects 2 or 3 arguments (array, fn, [initial]), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("reduce() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("reduce() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	hasInitial := len(args) == 3
	started := false
	var acc engine.Value

	for _, elem := range toIter(args[0]) {
		if !started {
			if hasInitial {
				acc = args[2]
				started = true
			} else {
				acc = elem
				started = true
				continue
			}
		}
		val, err := vm.CallValue(fn, acc, elem)
		if err != nil {
			return nil, fmt.Errorf("reduce() callback error: %w", err)
		}
		acc = val
	}

	if !started {
		if hasInitial {
			return args[2], nil
		}
		return engine.NewNull(), nil
	}

	return acc, nil
}

// builtinFind 返回数组中第一个满足条件的元素。
//
// 从左到右遍历数组，返回第一个回调函数返回 true 的元素。
// 如果没有元素满足条件，返回 null。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - Value: 找到的元素，未找到返回 null
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	find($nums, ($x) -> $x > 3)      // → 4
//	find($nums, ($x) -> $x > 10)     // → null
func builtinFind(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("find() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("find() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("find() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	idx := 0
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("find() callback error at index %d: %w", idx, err)
		}
		if val.Bool() {
			return elem, nil
		}
		idx++
	}

	return engine.NewNull(), nil
}

// builtinSome 检查数组是否有元素满足条件。
//
// 从左到右遍历数组，如果任一元素的回调函数返回 true，则立即返回 true。
// 如果所有元素都不满足条件，返回 false。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - bool: 有元素满足条件返回 true
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	some($nums, ($x) -> $x > 3)      // → true
//	some($nums, ($x) -> $x > 10)     // → false
func builtinSome(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("some() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("some() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("some() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	idx := 0
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("some() callback error at index %d: %w", idx, err)
		}
		if val.Bool() {
			return engine.NewBool(true), nil
		}
		idx++
	}

	return engine.NewBool(false), nil
}

// builtinEvery 检查数组是否所有元素都满足条件。
//
// 从左到右遍历数组，如果任一元素的回调函数返回 false，则立即返回 false。
// 如果所有元素都满足条件，返回 true。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - bool: 所有元素满足条件返回 true
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	every($nums, ($x) -> $x > 0)     // → true
//	every($nums, ($x) -> $x > 3)     // → false
func builtinEvery(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("every() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("every() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("every() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	idx := 0
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("every() callback error at index %d: %w", idx, err)
		}
		if !val.Bool() {
			return engine.NewBool(false), nil
		}
		idx++
	}

	return engine.NewBool(true), nil
}

// builtinSort 对数组进行排序，可选自定义比较函数。
//
// 不修改原数组，返回排好序的新数组。
// 默认按升序排序，可传入比较函数自定义排序规则。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 可选的比较函数 fn(a, b) → bool（a 应排在 b 前面时返回 true）
//
// 返回值：
//   - array: 排序后的新数组
//   - error: 参数错误
//
// 使用示例：
//
//	$nums = [3, 1, 4, 1, 5]
//	sort($nums)                        // → [1, 1, 3, 4, 5]
//	sort($nums, ($a, $b) -> $b - $a > 0) // → [5, 4, 3, 1, 1]（降序）
func builtinSort(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("sort() expects 1 or 2 arguments (array, [fn]), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("sort() first argument must be array, got %s", args[0].Type())
	}
	if len(args) == 2 && args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("sort() second argument must be function, got %s", args[1].Type())
	}

	arr := args[0].Array()
	result := make([]engine.Value, len(arr))
	copy(result, arr)

	vm := ctx.VM()

	if len(args) == 2 {
		// 自定义比较函数
		fn := args[1]
		less := func(i, j int) bool {
			val, err := vm.CallValue(fn, result[i], result[j])
			if err != nil {
				return false
			}
			return val.Bool()
		}
		sortSlice(result, less)
	} else {
		// 默认排序（使用 Less 方法）
		less := func(i, j int) bool {
			return result[i].Less(result[j])
		}
		sortSlice(result, less)
	}

	return engine.NewArray(result), nil
}

// sortSlice 简单的插入排序（数组规模通常不大）
func sortSlice(arr []engine.Value, less func(i, j int) bool) {
	n := len(arr)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && less(j, j-1); j-- {
			arr[j], arr[j-1] = arr[j-1], arr[j]
		}
	}
}

// builtinContains 检查数组是否包含指定值。
//
// 使用 Equals 方法进行比较，支持所有类型。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 要查找的值
//
// 返回值：
//   - bool: 包含返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	$arr = [1, 2, 3, "hello"]
//	contains($arr, 2)          // → true
//	contains($arr, "hello")    // → true
//	contains($arr, 5)          // → false
func builtinContains(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("contains() expects 2 arguments (array, value), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("contains() first argument must be array or range, got %s", args[0].Type())
	}

	target := args[1]

	for _, elem := range toIter(args[0]) {
		if elem.Equals(target) {
			return engine.NewBool(true), nil
		}
	}

	return engine.NewBool(false), nil
}

// builtinReject 排除满足条件的元素，返回新数组。
//
// 遍历数组的每个元素，调用回调函数，返回回调函数返回 false 的元素。
// 原数组不会被修改。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - array: 排除后的新数组
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	$odd = reject($nums, ($x) -> $x % 2 == 0)  // → [1, 3, 5]
func builtinReject(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("reject() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("reject() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("reject() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	var result []engine.Value
	idx := 0
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("reject() callback error at index %d: %w", idx, err)
		}
		if !val.Bool() {
			result = append(result, elem)
		}
		idx++
	}

	return engine.NewArray(result), nil
}

// builtinPartition 将数组按条件分成两组，返回包含两个数组的数组。
//
// 第一个数组包含回调函数返回 true 的元素，第二个数组包含回调函数返回 false 的元素。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 回调函数 fn(element) → bool
//
// 返回值：
//   - array: [[matching...], [non_matching...]]
//   - error: 参数错误或回调函数错误
//
// 使用示例：
//
//	$nums = [1, 2, 3, 4, 5]
//	$parts = partition($nums, ($x) -> $x % 2 == 0)
//	// → [[2, 4], [1, 3, 5]]
func builtinPartition(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("partition() expects 2 arguments (array, fn), got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("partition() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeFunc {
		return nil, fmt.Errorf("partition() second argument must be function, got %s", args[1].Type())
	}

	fn := args[1]
	vm := ctx.VM()

	var matching, nonMatching []engine.Value
	idx := 0
	for _, elem := range toIter(args[0]) {
		val, err := vm.CallValue(fn, elem)
		if err != nil {
			return nil, fmt.Errorf("partition() callback error at index %d: %w", idx, err)
		}
		if val.Bool() {
			matching = append(matching, elem)
		} else {
			nonMatching = append(nonMatching, elem)
		}
		idx++
	}

	return engine.NewArray([]engine.Value{
		engine.NewArray(matching),
		engine.NewArray(nonMatching),
	}), nil
}

// builtinFlattenDeep 深度展平嵌套数组，返回新数组。
//
// 将多层嵌套的数组递归展平为单一层级。默认无限深度，可指定最大深度。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组
//   - args[1]: 可选的最大深度（负数或省略表示无限深度）
//
// 返回值：
//   - array: 展平后的新数组
//   - error: 参数错误
//
// 使用示例：
//
//	$nested = [1, [2, [3, [4, [5]]]]]
//	flattenDeep($nested)                    // → [1, 2, 3, 4, 5]
//	flattenDeep($nested, 2)                // → [1, 2, [3, [4, [5]]]]
func builtinFlattenDeep(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("flattenDeep() expects 1-2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("flattenDeep() first argument must be array, got %s", args[0].Type())
	}

	depth := -1 // -1 表示无限深度
	if len(args) == 2 {
		depth = int(args[1].Int())
	}

	arr := args[0].Array()
	result := flattenDeepRecursive(arr, depth)

	return engine.NewArray(result), nil
}

// flattenDeepRecursive 递归展平数组
func flattenDeepRecursive(arr []engine.Value, depth int) []engine.Value {
	if depth == 0 {
		result := make([]engine.Value, len(arr))
		copy(result, arr)
		return result
	}

	var result []engine.Value
	for _, v := range arr {
		if v.Type() == engine.TypeArray {
			nested := flattenDeepRecursive(v.Array(), depth-1)
			result = append(result, nested...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

// builtinDifference 返回存在于第一个数组但不在其他数组中的元素。
//
// 计算数组差集，返回只在第一个数组中出现的元素（其他数组不包含）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 第一个数组
//   - args[1...]: 其他数组（至少一个）
//
// 返回值：
//   - array: 差集结果
//   - error: 参数错误
//
// 使用示例：
//
//	$a = [1, 2, 3, 4, 5]
//	$b = [2, 4, 6]
//	$c = [3, 5]
//	difference($a, $b, $c)  // → [1]
func builtinDifference(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("difference() expects at least 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("difference() first argument must be array, got %s", args[0].Type())
	}

	// 构建所有其他数组的元素的 Set
	excludeSet := make(map[string]bool)
	for i := 1; i < len(args); i++ {
		if args[i].Type() != engine.TypeArray {
			return nil, fmt.Errorf("difference() argument %d must be array, got %s", i+1, args[i].Type())
		}
		for _, v := range args[i].Array() {
			excludeSet[v.Stringify()] = true
		}
	}

	// 收集差集元素
	var result []engine.Value
	for _, v := range args[0].Array() {
		key := v.Stringify()
		if !excludeSet[key] {
			result = append(result, v)
		}
	}

	return engine.NewArray(result), nil
}

// builtinUnion 合并多个数组并去重，返回新数组。
//
// 将所有数组的元素合并为一个数组，自动去除重复元素。保留元素的原始顺序。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0...]: 多个数组（至少一个）
//
// 返回值：
//   - array: 合并去重后的新数组
//   - error: 参数错误
//
// 使用示例：
//
//	$a = [1, 2, 3]
//	$b = [3, 4, 5]
//	$c = [5, 6, 7]
//	union($a, $b, $c)  // → [1, 2, 3, 4, 5, 6, 7]
func builtinUnion(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("union() expects at least 1 argument, got %d", len(args))
	}

	seen := make(map[string]bool)
	var result []engine.Value

	for i, arg := range args {
		if arg.Type() != engine.TypeArray {
			return nil, fmt.Errorf("union() argument %d must be array, got %s", i+1, arg.Type())
		}
		for _, v := range arg.Array() {
			key := v.Stringify()
			if !seen[key] {
				seen[key] = true
				result = append(result, v)
			}
		}
	}

	return engine.NewArray(result), nil
}

// builtinZip 将多个数组相同索引的元素组成元组数组。
//
// 将 n 个数组的相同索引位置元素组成一个数组，返回 n 个元组组成的数组。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0...]: 多个数组（至少一个）
//
// 返回值：
//   - array: 元组数组 [[a1, b1, c1], [a2, b2, c2], ...]
//   - error: 参数错误
//
// 使用示例：
//
//	$names = ["Alice", "Bob", "Charlie"]
//	$ages = [25, 30, 35]
//	zip($names, $ages)
//	// → [["Alice", 25], ["Bob", 30], ["Charlie", 35]]
func builtinZip(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("zip() expects at least 1 argument, got %d", len(args))
	}

	// 验证所有参数都是数组
	for i, arg := range args {
		if arg.Type() != engine.TypeArray {
			return nil, fmt.Errorf("zip() argument %d must be array, got %s", i+1, arg.Type())
		}
	}

	// 获取最小长度
	minLen := -1
	for _, arg := range args {
		l := len(arg.Array())
		if minLen == -1 || l < minLen {
			minLen = l
		}
	}

	// 构建元组数组
	var result []engine.Value
	for i := 0; i < minLen; i++ {
		var tuple []engine.Value
		for _, arg := range args {
			tuple = append(tuple, arg.Array()[i])
		}
		result = append(result, engine.NewArray(tuple))
	}

	return engine.NewArray(result), nil
}

// builtinUnzip 将元组数组拆分回多个数组。
//
// 将由元组组成的数组拆分回多个独立数组。与 zip 互为逆操作。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 元组数组
//
// 返回值：
//   - array: 拆分后的多个数组 [[a1, a2, a3], [b1, b2, b3], ...]
//   - error: 参数错误
//
// 使用示例：
//
//	$tuples = [["Alice", 25], ["Bob", 30], ["Charlie", 35]]
//	unzip($tuples)
//	// → [["Alice", "Bob", "Charlie"], [25, 30, 35]]
func builtinUnzip(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unzip() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray {
		return nil, fmt.Errorf("unzip() argument must be array, got %s", args[0].Type())
	}

	tuples := args[0].Array()
	if len(tuples) == 0 {
		return engine.NewArray([]engine.Value{}), nil
	}

	// 检查元组结构
	firstTuple := tuples[0].Array()
	if len(firstTuple) == 0 {
		return engine.NewArray([]engine.Value{}), nil
	}

	numArrays := len(firstTuple)
	result := make([][]engine.Value, numArrays)

	// 初始化结果数组
	for i := range numArrays {
		result[i] = make([]engine.Value, 0, len(tuples))
	}

	// 拆分元组
	for _, tupleVal := range tuples {
		tuple := tupleVal.Array()
		for i := 0; i < numArrays && i < len(tuple); i++ {
			result[i] = append(result[i], tuple[i])
		}
	}

	// 转换为 Value 数组
	var finalResult []engine.Value
	for _, arr := range result {
		finalResult = append(finalResult, engine.NewArray(arr))
	}

	return engine.NewArray(finalResult), nil
}

// builtinFirst 返回数组或范围的第一个元素。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//
// 返回值：
//   - Value: 第一个元素，空数组返回 null
//   - error: 参数错误
//
// 使用示例：
//
//	first([1, 2, 3])      // → 1
//	first(1...5)          // → 1
//	first([])            // → null
func builtinFirst(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("first() argument must be array or range, got %s", args[0].Type())
	}

	firstElem, ok := engine.Value(nil), false
	for i, elem := range toIter(args[0]) {
		if i == 0 {
			firstElem = elem
			ok = true
			break
		}
	}
	if !ok {
		return engine.NewNull(), nil
	}
	return firstElem, nil
}

// builtinLast 返回数组或范围的最后一个元素。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//
// 返回值：
//   - Value: 最后一个元素，空数组返回 null
//   - error: 参数错误
//
// 使用示例：
//
//	last([1, 2, 3])      // → 3
//	last(1...5)          // → 5 (非包含) 或 4 (包含?)
//	last([])             // → null
func builtinLast(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("last() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("last() argument must be array or range, got %s", args[0].Type())
	}

	var result engine.Value
	for _, elem := range toIter(args[0]) {
		result = elem
	}

	if result == nil {
		return engine.NewNull(), nil
	}
	return result, nil
}

// builtinTake 返回数组或范围的前 n 个元素。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//   - args[1]: 要获取的元素数量
//
// 返回值：
//   - array: 前 n 个元素组成的数组
//   - error: 参数错误
//
// 使用示例：
//
//	take([1, 2, 3, 4, 5], 3)  // → [1, 2, 3]
//	take(1...10, 3)          // → [1, 2, 3]
//	take(1...10, 20)        // → [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
func builtinTake(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("take() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("take() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeInt {
		return nil, fmt.Errorf("take() second argument must be int, got %s", args[1].Type())
	}

	n := int(args[1].Int())
	if n <= 0 {
		return engine.NewArray([]engine.Value{}), nil
	}

	var result []engine.Value
	for i, elem := range toIter(args[0]) {
		if i >= n {
			break
		}
		result = append(result, elem)
	}

	return engine.NewArray(result), nil
}

// builtinDrop 跳过数组或范围的前 n 个元素，返回剩余元素。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//   - args[1]: 要跳过的元素数量
//
// 返回值：
//   - array: 跳过前 n 个元素后的数组
//   - error: 参数错误
//
// 使用示例：
//
//	drop([1, 2, 3, 4, 5], 2)  // → [3, 4, 5]
//	drop(1...10, 5)          // → [6, 7, 8, 9, 10]
func builtinDrop(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("drop() expects 2 arguments, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("drop() first argument must be array or range, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeInt {
		return nil, fmt.Errorf("drop() second argument must be int, got %s", args[1].Type())
	}

	n := max(int(args[1].Int()), 0)

	var result []engine.Value
	idx := 0
	for _, elem := range toIter(args[0]) {
		if idx >= n {
			result = append(result, elem)
		}
		idx++
	}

	return engine.NewArray(result), nil
}

// builtinSum 计算数组或范围所有元素的和（仅限数值类型）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组或范围
//
// 返回值：
//   - int: 所有元素的总和
//   - error: 参数错误或元素类型不支持
//
// 使用示例：
//
//	sum([1, 2, 3, 4, 5])  // → 15
//	sum(1...10)          // → 55
func builtinSum(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sum() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeArray && args[0].Type() != engine.TypeRange {
		return nil, fmt.Errorf("sum() argument must be array or range, got %s", args[0].Type())
	}

	var sum int64
	for _, elem := range toIter(args[0]) {
		if elem.Type() != engine.TypeInt && elem.Type() != engine.TypeFloat {
			return nil, fmt.Errorf("sum() expects numeric elements, got %s", elem.Type())
		}
		sum += elem.Int()
	}

	return engine.NewInt(sum), nil
}

// builtinSize 返回数组、范围或对象的元素数量。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 输入数组、范围或对象
//
// 返回值：
//   - int: 元素数量
//   - error: 参数错误
//
// 使用示例：
//
//	size([1, 2, 3])    // → 3
//	size(1...10)      // → 9
//	size({a: 1, b: 2}) // → 2
//	size("hello")     // → 1（非空字符串返回 1）
//	size(null)        // → 0
func builtinSize(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("size() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	switch v.Type() {
	case engine.TypeArray, engine.TypeRange:
		if hasLen, ok := getLength(v); ok {
			return engine.NewInt(int64(hasLen)), nil
		}
		count := 0
		for range toIter(v) {
			count++
		}
		return engine.NewInt(int64(count)), nil
	case engine.TypeObject:
		return engine.NewInt(int64(v.Len())), nil
	default:
		if v.Type() == engine.TypeNull {
			return engine.NewInt(0), nil
		}
		return engine.NewInt(1), nil
	}
}

// FunctionalSigs returns function signatures for REPL :doc command.
func FunctionalSigs() map[string]string {
	return map[string]string{
		"map":         "map(array_or_range, fn(element) → newValue) → array  — Apply function to each element",
		"filter":      "filter(array_or_range, fn(element) → bool) → array  — Filter elements by predicate",
		"reduce":      "reduce(array_or_range, fn(acc, element) → newAcc, [initial]) → value  — Reduce to single value",
		"find":        "find(array_or_range, fn(element) → bool) → value  — Find first matching element",
		"some":        "some(array_or_range, fn(element) → bool) → bool  — Check if any element matches",
		"every":       "every(array_or_range, fn(element) → bool) → bool  — Check if all elements match",
		"sort":        "sort(array, [fn(a, b) → bool]) → array  — Sort array with optional comparator",
		"contains":    "contains(array_or_range, value) → bool  — Check if value exists",
		"reject":      "reject(array_or_range, fn(element) → bool) → array  — Remove matching elements",
		"partition":   "partition(array_or_range, fn(element) → bool) → array  — Split into two groups",
		"unique":      "unique(array) → array  — Remove duplicates",
		"flattenDeep": "flattenDeep(array, [depth]) → array  — Flatten nested arrays",
		"difference":  "difference(array, ...other_arrays) → array  — Elements only in first array",
		"union":       "union(...arrays) → array  — Merge and deduplicate arrays",
		"zip":         "zip(...arrays) → array  — Combine arrays by index",
		"unzip":       "unzip(tuple_array) → array  — Split tuple array into arrays",
		"first":       "first(array_or_range) → value  — Return first element",
		"last":        "last(array_or_range) → value  — Return last element",
		"take":        "take(array_or_range, n) → array  — Take first n elements",
		"drop":        "drop(array_or_range, n) → array  — Skip first n elements",
		"sum":         "sum(array_or_range) → int  — Sum numeric elements",
		"size":        "size(value) → int  — Return element count",
	}
}
