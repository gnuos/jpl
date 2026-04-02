package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterUtil 将实用工具函数注册到引擎。
//
// 当前注册的函数：
//   - len: 返回字符串、数组或对象的长度
//
// 参数：
//   - e: 引擎实例
//
// 使用示例：
//
//	buildin.RegisterUtil(eng)
//
//	vm, _ := eng.Compile(`
//	    $str = "Hello"
//	    print len($str)        // 输出: 5
//
//	    $arr = [1, 2, 3, 4]
//	    print len($arr)        // 输出: 4
//
//	    $obj = {a: 1, b: 2}
//	    print len($obj)        // 输出: 2
//	`)
func RegisterUtil(e *engine.Engine) {
	e.RegisterFunc("len", builtinLen)
}

// UtilNames 返回工具函数名称列表。
//
// 返回值：
//   - []string: 工具函数名列表 ["len"]
func UtilNames() []string {
	return []string{"len"}
}

// builtinLen 返回值的元素个数或长度。
//
// 根据值的类型返回不同的含义：
//   - 字符串: 字符数（UTF-8 编码的字节数）
//   - 数组: 元素个数
//   - 对象: 键值对数量
//
// 不支持其他类型（如 int、float、null、bool 等），会返回错误。
//
// 注意：在 JPL 中，可以使用 # 运算符作为快捷方式：
//
//	len($arr) 等价于 # $arr
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要计算长度的值
//
// 返回值：
//   - int: 长度值（字符串字符数、数组元素数、对象键数）
//   - error: 参数错误或不支持的类型
//
// 使用示例：
//
//	$str = "Hello, 世界"
//	print len($str)          // 输出: 12（注意：不是 9 个字符，而是 12 个字节）
//
//	$arr = [1, 2, 3, 4, 5]
//	print len($arr)          // 输出: 5
//	print # $arr             // 同上，使用 # 运算符
//
//	$obj = {name: "Alice", age: 30}
//	print len($obj)          // 输出: 2
//
//	// 错误示例
//	print len(42)            // 错误: len() not supported for type int
func builtinLen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	switch v.Type() {
	case engine.TypeString, engine.TypeArray, engine.TypeObject:
		return engine.NewInt(int64(v.Len())), nil
	default:
		return nil, fmt.Errorf("len() not supported for type %s", v.Type())
	}
}
