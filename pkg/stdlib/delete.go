package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterDelete 注册 delete 和 unset 函数。
//
// 当前注册的函数：
//   - delete: 删除对象成员或数组元素
//   - unset: 将变量设为 null
//
// 参数：
//   - e: 引擎实例
func RegisterDelete(e *engine.Engine) {
	e.RegisterFunc("delete", builtinDelete)
	e.RegisterFunc("unset", builtinUnset)
}

// DeleteNames 返回删除函数名称列表。
func DeleteNames() []string {
	return []string{"delete", "unset"}
}

// builtinDelete 删除对象成员或数组元素。
//
// 用法：
//   - delete($obj, "key"): 删除对象的成员
//   - delete($arr, index): 删除数组指定索引的元素
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 对象或数组
//   - args[1]: 成员名（字符串）或索引（整数）
//
// 返回值：
//   - bool: 是否成功删除
//   - error: 参数错误
func builtinDelete(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("delete() expects 2 arguments, got %d", len(args))
	}

	target := args[0]
	key := args[1]

	switch target.Type() {
	case engine.TypeObject:
		objMap := target.Object()
		switch key.Type() {
		case engine.TypeString:
			keyStr := key.String()
			if _, exists := objMap[keyStr]; exists {
				delete(objMap, keyStr)
				return engine.NewBool(true), nil
			}
			return engine.NewBool(false), nil
		default:
			return nil, fmt.Errorf("delete(): object key must be a string, got %s", key.Type())
		}

	case engine.TypeArray:
		arr := target.Array()
		switch key.Type() {
		case engine.TypeInt:
			idx := int(key.Int())
			if idx < 0 || idx >= len(arr) {
				return engine.NewBool(false), nil
			}
			// 删除元素并返回新数组
			newArr := make([]engine.Value, 0, len(arr)-1)
			for i, v := range arr {
				if i != idx {
					newArr = append(newArr, v)
				}
			}
			return engine.NewArray(newArr), nil
		default:
			return nil, fmt.Errorf("delete(): array index must be an integer, got %s", key.Type())
		}

	default:
		return nil, fmt.Errorf("delete(): first argument must be an object or array, got %s", target.Type())
	}
}

// builtinUnset 将变量设为 null（模拟 PHP 的 unset）。
//
// 用法：
//   - unset($var): 将变量设为 null
//
// 注意：此函数会将变量的值设为 null，而不是真正从作用域中删除变量。
// 这是因为 JPL 的变量是值类型，无法真正"删除"变量。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要unset的变量
//
// 返回值：
//   - null: 总是返回 null
//   - error: 参数错误
func builtinUnset(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unset() expects 1 argument, got %d", len(args))
	}

	// 将值设为 null
	args[0] = engine.NewNull()

	return engine.NewNull(), nil
}
