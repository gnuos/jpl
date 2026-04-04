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
	e.RegisterFunc("typeof", builtinTypeof)
	e.RegisterFunc("dump", builtinDump)
	e.RegisterFunc("keys", builtinKeys)
	e.RegisterFunc("values", builtinValues)
	e.RegisterFunc("entries", builtinEntries)
	e.RegisterFunc("has_key", builtinHasKey)
	e.RegisterFunc("clone", builtinClone)
}

// UtilNames 返回工具函数名称列表。
func UtilNames() []string {
	return []string{"len", "typeof", "dump", "keys", "values", "entries", "has_key", "clone"}
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

// builtinTypeof 返回值的类型名称。
func builtinTypeof(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("typeof() expects 1 argument, got %d", len(args))
	}
	return engine.NewString(args[0].Type().String()), nil
}

// builtinDump 返回值的详细调试信息。
func builtinDump(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dump() expects 1 argument, got %d", len(args))
	}
	return engine.NewString(dumpValue(args[0], 0)), nil
}

func dumpValue(v engine.Value, indent int) string {
	prefix := ""
	for range indent {
		prefix += "  "
	}

	switch v.Type() {
	case engine.TypeArray:
		arr := v.Array()
		if len(arr) == 0 {
			return "[]"
		}
		result := "[\n"
		for i, item := range arr {
			result += prefix + "  " + dumpValue(item, indent+1)
			if i < len(arr)-1 {
				result += ","
			}
			result += "\n"
		}
		result += prefix + "]"
		return result
	case engine.TypeObject:
		obj := v.Object()
		if len(obj) == 0 {
			return "{}"
		}
		result := "{\n"
		i := 0
		for key, val := range obj {
			result += prefix + "  " + key + ": " + dumpValue(val, indent+1)
			if i < len(obj)-1 {
				result += ","
			}
			result += "\n"
			i++
		}
		result += prefix + "}"
		return result
	case engine.TypeString:
		return fmt.Sprintf("%q", v.String())
	case engine.TypeNull:
		return "null"
	case engine.TypeBool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return v.Stringify()
	}
}

// builtinKeys 返回对象的所有键或数组的所有索引。
func builtinKeys(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("keys() expects 1 argument, got %d", len(args))
	}

	v := args[0]
	switch v.Type() {
	case engine.TypeObject:
		obj := v.Object()
		keys := make([]engine.Value, 0, len(obj))
		for k := range obj {
			keys = append(keys, engine.NewString(k))
		}
		return engine.NewArray(keys), nil
	case engine.TypeArray:
		arr := v.Array()
		keys := make([]engine.Value, len(arr))
		for i := range arr {
			keys[i] = engine.NewInt(int64(i))
		}
		return engine.NewArray(keys), nil
	default:
		return nil, fmt.Errorf("keys() expects object or array, got %s", v.Type())
	}
}

// builtinValues 返回对象的所有值。
func builtinValues(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("values() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("values() expects object, got %s", args[0].Type())
	}

	obj := args[0].Object()
	vals := make([]engine.Value, 0, len(obj))
	for _, v := range obj {
		vals = append(vals, v)
	}
	return engine.NewArray(vals), nil
}

// builtinEntries 返回对象的 [key, value] 对数组。
func builtinEntries(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("entries() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("entries() expects object, got %s", args[0].Type())
	}

	obj := args[0].Object()
	entries := make([]engine.Value, 0, len(obj))
	for k, v := range obj {
		entry := []engine.Value{engine.NewString(k), v}
		entries = append(entries, engine.NewArray(entry))
	}
	return engine.NewArray(entries), nil
}

// builtinHasKey 检查对象是否包含指定键。
func builtinHasKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has_key() expects 2 arguments (obj, key), got %d", len(args))
	}

	if args[0].Type() != engine.TypeObject {
		return nil, fmt.Errorf("has_key() argument 1 must be object, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("has_key() argument 2 must be string, got %s", args[1].Type())
	}

	obj := args[0].Object()
	key := args[1].String()
	_, exists := obj[key]
	return engine.NewBool(exists), nil
}

// builtinClone 深拷贝任意值。
func builtinClone(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("clone() expects 1 argument, got %d", len(args))
	}
	return cloneValue(args[0]), nil
}

func cloneValue(v engine.Value) engine.Value {
	switch v.Type() {
	case engine.TypeArray:
		arr := v.Array()
		result := make([]engine.Value, len(arr))
		for i, item := range arr {
			result[i] = cloneValue(item)
		}
		return engine.NewArray(result)
	case engine.TypeObject:
		obj := v.Object()
		result := make(map[string]engine.Value, len(obj))
		for k, val := range obj {
			result[k] = cloneValue(val)
		}
		return engine.NewObject(result)
	default:
		return v
	}
}

// UtilSigs returns function signatures for REPL :doc command.
func UtilSigs() map[string]string {
	return map[string]string{
		"len":     "len(value) → int  — Return length of string, array, or object",
		"typeof":  "typeof(value) → string  — Return type name of value",
		"dump":    "dump(value) → string  — Return debug representation of value",
		"keys":    "keys(obj_or_arr) → array  — Return keys of object or indices of array",
		"values":  "values(obj) → array  — Return values of object",
		"entries": "entries(obj) → array  — Return [key, value] pairs of object",
		"has_key": "has_key(obj, key) → bool  — Check if object has key",
		"clone":   "clone(value) → value  — Deep clone any value",
	}
}
