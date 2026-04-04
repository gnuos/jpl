package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterIO 将 I/O 相关的内置函数注册到引擎。
//
// 注册的函数：
//   - print: 输出到 stdout（无换行）
//   - println: 输出到 stdout（带换行）
//   - echo: 拼接参数为字符串（不输出）
//   - format: 格式化字符串
//   - assert: 断言检查
//
// 这些函数同时注册到全局命名空间和 "io" 模块，可以通过以下方式使用：
//
//	print "hello"           // 全局调用
//	import "io"; io.print "hello"  // 模块调用
//
// 参数：
//   - e: 引擎实例
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	buildin.RegisterIO(eng)  // 注册 I/O 函数
//
//	vm, _ := eng.Compile(`
//	    print "Hello"
//	    println "World"
//	    $str = echo "test"
//	    $fmt = format("Name: %s", "Alice")
//	    assert($x > 0, "x must be positive")
//	`)
func RegisterIO(e *engine.Engine) {
	e.RegisterFunc("print", builtinPrint)
	e.RegisterFunc("println", builtinPrintln)
	e.RegisterFunc("puts", builtinPuts)
	e.RegisterFunc("pp", builtinPP)
	e.RegisterFunc("echo", builtinEcho)
	e.RegisterFunc("format", builtinFormat)
	e.RegisterFunc("assert", builtinAssert)

	e.RegisterFunc("fopen", builtinFopen)
	e.RegisterFunc("fread", builtinFread)
	e.RegisterFunc("fgets", builtinFgets)
	e.RegisterFunc("fwrite", builtinFwrite)
	e.RegisterFunc("fclose", builtinFclose)
	e.RegisterFunc("feof", builtinFeof)
	e.RegisterFunc("fflush", builtinFflush)
	e.RegisterFunc("stream_get_meta_data", builtinStreamGetMetaData)
	e.RegisterFunc("is_readable", builtinStreamIsReadable)
	e.RegisterFunc("is_writable", builtinStreamIsWritable)

	e.RegisterModule("io", map[string]engine.GoFunction{
		"print": builtinPrint, "println": builtinPrintln, "puts": builtinPuts, "pp": builtinPP,
		"echo": builtinEcho, "format": builtinFormat, "assert": builtinAssert,
		"fopen": builtinFopen, "fread": builtinFread, "fgets": builtinFgets,
		"fwrite": builtinFwrite, "fclose": builtinFclose, "feof": builtinFeof,
		"fflush": builtinFflush, "stream_get_meta_data": builtinStreamGetMetaData,
		"is_readable": builtinStreamIsReadable, "is_writable": builtinStreamIsWritable,
	})
}

// IONames 返回 I/O 相关内置函数的名称列表。
//
// 用于代码补全和函数枚举。
//
// 返回值：
//   - []string: I/O 函数名列表 ["print", "println", "puts", "pp", "echo", "format", "assert"]
func IONames() []string {
	return []string{"print", "println", "puts", "pp", "echo", "format", "assert"}
}

// builtinPrint 将参数输出到标准输出（stdout），不换行。
//
// 参数之间用空格分隔。此函数是最常用的输出函数，
// 支持所有数据类型，会自动调用 Stringify() 方法进行格式化。
//
// 特例语法支持：print 语句级调用时可省略括号
//
//	print "hello"        // 合法
//	print("hello")       // 也合法
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要输出的值列表（变长参数）
//
// 返回值：
//   - null: 总是返回 null
//   - nil: 无错误
//
// 使用示例：
//
//	print "Hello"                    // 输出: Hello
//	print "The answer is" 42         // 输出: The answer is 42
//	print [1, 2, 3]                  // 输出: [1, 2, 3]
//	print {name: "Alice"}            // 输出: {"name": "Alice"}
func builtinPrint(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 检查首参数是否为流（支持 print(STDERR, "msg") 语法）
	var output io.Writer = os.Stdout
	startIdx := 0
	if len(args) > 0 && engine.IsStream(args[0]) {
		sv := engine.ToStreamValue(args[0])
		if sv != nil && sv.IsWritable() {
			output = sv
			startIdx = 1
		}
	}

	parts := make([]string, len(args)-startIdx)
	for i, arg := range args[startIdx:] {
		if arg.Type() == engine.TypeString {
			parts[i] = arg.String()
		} else {
			parts[i] = arg.Stringify()
		}
	}
	fmt.Fprint(output, strings.Join(parts, " "))
	return engine.NewNull(), nil
}

// builtinPrintln 将参数输出到标准输出（stdout），末尾自动添加换行符。
//
// 参数之间用空格分隔，输出结束后换行。
// 这是最常用的输出函数，等价于 print 后加换行。
//
// 特例语法支持：println 语句级调用时可省略括号
//
//	println "hello"        // 合法
//	println("hello")       // 也合法
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要输出的值列表（变长参数）
//
// 返回值：
//   - null: 总是返回 null
//   - nil: 无错误
//
// 使用示例：
//
//	println "Hello"                    // 输出: Hello\n
//	println "Line 1"
//	println "Line 2"                    // 每行单独输出
//	println "Sum:" (10 + 20)            // 输出: Sum: 30\n
func builtinPrintln(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 检查首参数是否为流（支持 println(STDERR, "msg") 语法）
	var output io.Writer = os.Stdout
	startIdx := 0
	if len(args) > 0 && engine.IsStream(args[0]) {
		sv := engine.ToStreamValue(args[0])
		if sv != nil && sv.IsWritable() {
			output = sv
			startIdx = 1
		}
	}

	parts := make([]string, len(args)-startIdx)
	for i, arg := range args[startIdx:] {
		if arg.Type() == engine.TypeString {
			parts[i] = arg.String()
		} else {
			parts[i] = arg.Stringify()
		}
	}
	fmt.Fprintln(output, strings.Join(parts, " "))
	return engine.NewNull(), nil
}

// builtinEcho 将参数拼接为字符串返回，不输出到 stdout。
//
// 与 print/println 不同，echo 不会直接输出，而是返回拼接后的字符串。
// 参数之间用空格分隔。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要拼接的值列表（变长参数）
//
// 返回值：
//   - string: 拼接后的字符串
//   - nil: 无错误
//
// 使用示例：
//
//	$msg = echo "Hello" "World"     // $msg = "Hello World"
//	$str = echo 10 20 30            // $str = "10 20 30"
//	println echo "The" "answer" "is" 42   // 输出: The answer is 42
func builtinEcho(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = valueToString(arg)
	}
	return engine.NewString(strings.Join(parts, " ")), nil
}

// builtinFormat 格式化字符串，使用 %s 作为占位符替换为参数值。
//
// 第一个参数是模板字符串，后续参数按顺序替换模板中的 %s。
// 支持多个 %s 占位符，参数数量不足时保留未替换的 %s。
//
// 格式化规则：
//   - %s 会被替换为对应参数的字符串表示（调用 String() 方法）
//   - 多余的参数会被忽略
//   - 参数不足时，剩余的 %s 保持不变
//
// 参数：
//   - ctx: 执行上下文
//   - args: 第一个参数是模板字符串，后续是替换值
//
// 返回值：
//   - string: 格式化后的字符串
//   - error: 参数不足时返回错误
//
// 使用示例：
//
//	$str = format("Hello, %s!", "World")           // "Hello, World!"
//	$msg = format("Name: %s, Age: %s", "Alice", 30) // "Name: Alice, Age: 30"
//	$tpl = format("File: %s, Line: %s", "test.jpl") // "File: test.jpl, Line: %s" (保留未替换)
//
// 注意：使用字符串插值 {$var} 通常是更好的选择：
//
//	$str = "Hello, {$name}!"
func builtinFormat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("format() expects at least 1 argument, got %d", len(args))
	}

	tmpl := valueToString(args[0])
	result := tmpl
	argIdx := 1

	for argIdx < len(args) {
		idx := strings.Index(result, "%s")
		if idx == -1 {
			break
		}
		result = result[:idx] + valueToString(args[argIdx]) + result[idx+2:]
		argIdx++
	}

	return engine.NewString(result), nil
}

// builtinAssert 断言检查，条件为 false 时抛出运行时错误。
//
// 用于调试和验证，当断言条件不满足时立即终止执行并报告错误。
// 支持自定义错误消息，如果不提供则使用默认消息 "assertion failed"。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 断言条件（布尔表达式）
//   - args[1]: 可选的错误消息（字符串）
//
// 返回值：
//   - null: 断言通过时返回 null
//   - error: 断言失败时返回错误
//
// 使用示例：
//
//	assert($x > 0)                      // 如果 $x <= 0，抛出 "assertion failed"
//	assert($x > 0, "x must be positive") // 自定义错误消息
//	assert(len($arr) > 0, "Array cannot be empty")
//
// 在正式代码中，建议使用条件语句代替：
//
//	if ($x <= 0) { throw error("x must be positive", 4001) }
func builtinAssert(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("assert() expects 1 or 2 arguments, got %d", len(args))
	}

	if !args[0].Bool() {
		msg := "assertion failed"
		if len(args) == 2 {
			msg = args[1].Stringify()
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return engine.NewNull(), nil
}

func builtinFopen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("fopen() expects 1 or 2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("fopen() argument 1 must be a string, got %s", args[0].Type())
	}

	path := args[0].String()

	mode := engine.StreamRead
	if len(args) == 2 {
		if args[1].Type() != engine.TypeString {
			return nil, fmt.Errorf("fopen() argument 2 must be a string, got %s", args[1].Type())
		}
		modeStr := args[1].String()
		switch modeStr {
		case "r":
			mode = engine.StreamRead
		case "w":
			mode = engine.StreamWrite
		case "rw", "r+":
			mode = engine.StreamReadWrite
		default:
			return nil, fmt.Errorf("fopen() invalid mode: %s (supported: r, w, rw)", modeStr)
		}
	}

	return engine.NewFileStream(path, mode), nil
}

func builtinFread(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("fread() expects 2 arguments, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("fread() argument 1 must be a stream, got %s", args[0].Type())
	}

	if args[1].Type() != engine.TypeInt && args[1].Type() != engine.TypeFloat {
		return nil, fmt.Errorf("fread() argument 2 must be an integer, got %s", args[1].Type())
	}

	length := int(args[1].Int())
	if length <= 0 {
		length = 8192
	}
	if length > 8192 {
		length = 8192
	}

	buf := make([]byte, length)
	n, err := stream.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("fread() error: %v", err)
	}

	return engine.NewString(string(buf[:n])), nil
}

func builtinFgets(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("fgets() expects 1 argument, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("fgets() argument 1 must be a stream, got %s", args[0].Type())
	}

	reader := bufio.NewReader(stream)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("fgets() error: %v", err)
	}

	return engine.NewString(line), nil
}

func builtinFwrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("fwrite() expects 2 arguments, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("fwrite() argument 1 must be a stream, got %s", args[0].Type())
	}

	data := args[1].Stringify()
	n, err := stream.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("fwrite() error: %v", err)
	}

	return engine.NewInt(int64(n)), nil
}

func builtinFclose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("fclose() expects 1 argument, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("fclose() argument 1 must be a stream, got %s", args[0].Type())
	}

	err := stream.Close()
	if err != nil {
		return nil, fmt.Errorf("fclose() error: %v", err)
	}

	return engine.NewNull(), nil
}

func builtinFeof(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("feof() expects 1 argument, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("feof() argument 1 must be a stream, got %s", args[0].Type())
	}

	if stream.IsClosed() {
		return engine.NewBool(true), nil
	}

	return engine.NewBool(false), nil
}

func builtinFflush(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("fflush() expects 1 argument, got %d", len(args))
	}

	stream := engine.ToStreamValue(args[0])
	if stream == nil {
		return nil, fmt.Errorf("fflush() argument 1 must be a stream, got %s", args[0].Type())
	}

	if stream.IsClosed() {
		return nil, fmt.Errorf("fflush() stream is closed")
	}

	return engine.NewBool(true), nil
}

func builtinStreamGetMetaData(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("stream_get_meta_data() expects 1 argument, got %d", len(args))
	}

	meta := engine.StreamMeta(args[0])
	if meta == nil {
		return nil, fmt.Errorf("stream_get_meta_data() argument 1 must be a stream, got %s", args[0].Type())
	}

	return engine.NewObject(meta), nil
}

func builtinStreamIsReadable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_readable() expects 1 argument, got %d", len(args))
	}

	if engine.IsStream(args[0]) {
		stream := engine.ToStreamValue(args[0])
		return engine.NewBool(stream.IsReadable()), nil
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("is_readable() argument must be string or stream, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}
	mode := info.Mode()
	return engine.NewBool(mode&0444 != 0), nil
}

func builtinStreamIsWritable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("is_writable() expects 1 argument, got %d", len(args))
	}

	if engine.IsStream(args[0]) {
		stream := engine.ToStreamValue(args[0])
		return engine.NewBool(stream.IsWritable()), nil
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("is_writable() argument must be string or stream, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}
	mode := info.Mode()
	return engine.NewBool(mode&0222 != 0), nil
}

// builtinPuts 将参数输出到标准输出（stdout），所有值都不带引号。
//
// 与 print/println 不同，puts 对所有类型都使用 String() 方法，
// 不添加调试用的引号。适合用户友好的纯文本输出。
//
// 参数之间用空格分隔，输出结束后自动换行。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要输出的值列表（变长参数）
//
// 返回值：
//   - null: 总是返回 null
//   - nil: 无错误
//
// 使用示例：
//
//	puts "Hello"                    // 输出: Hello
//	puts [1, 2, 3]                  // 输出: [1, 2, 3]（不带引号）
//	puts {name: "Alice"}            // 输出: {name: Alice}（不带引号）
//	puts "Sum:" (10 + 20)          // 输出: Sum: 30
func builtinPuts(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.String()
	}
	fmt.Println(strings.Join(parts, " "))
	return engine.NewNull(), nil
}

// builtinPP Pretty Print 格式化输出对象和数组。
//
// 类似于 Ruby 的 pp 函数，对复杂数据结构进行缩进格式化，
// 使其更易读。输出结束后自动换行。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 要格式化的值列表（变长参数）
//
// 返回值：
//   - null: 总是返回 null
//   - nil: 无错误
//
// 使用示例：
//
//	pp {name: "Alice", items: [1, 2, 3]}
//	// 输出:
//	// {
//	//   name: "Alice",
//	//   items: [1, 2, 3]
//	// }
//
//	pp [1, [2, 3], 4]
//	// 输出:
//	// [1, [2, 3], 4]
func builtinPP(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	for _, arg := range args {
		output := prettyPrintValue(arg, 0)
		fmt.Println(output)
	}
	return engine.NewNull(), nil
}

// prettyPrintValue 递归格式化值，带缩进
func prettyPrintValue(v engine.Value, indent int) string {
	prefix := strings.Repeat("  ", indent)

	switch v.Type() {
	case engine.TypeArray:
		arr := v.Array()
		if len(arr) == 0 {
			return "[]"
		}

		// 检查是否所有元素都是简单类型（非数组/对象）
		allSimple := true
		for _, item := range arr {
			if item.Type() == engine.TypeArray || item.Type() == engine.TypeObject {
				allSimple = false
				break
			}
		}

		if allSimple && len(arr) <= 5 {
			// 简单数组，单行输出
			parts := make([]string, len(arr))
			for i, item := range arr {
				parts[i] = prettyPrintValue(item, 0)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}

		// 复杂数组，多行输出
		result := "[\n"
		for i, item := range arr {
			result += prefix + "  " + prettyPrintValue(item, indent+1)
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

		// 获取所有 key 并排序（保持一致性）
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}

		// 简单对象检查（所有值都是简单类型）
		allSimple := true
		for _, key := range keys {
			val := obj[key]
			if val.Type() == engine.TypeArray || val.Type() == engine.TypeObject {
				allSimple = false
				break
			}
		}

		if allSimple && len(keys) <= 3 {
			// 简单对象，单行输出
			parts := make([]string, 0, len(keys))
			for _, key := range keys {
				val := obj[key]
				parts = append(parts, fmt.Sprintf("%s: %s", key, prettyPrintValue(val, 0)))
			}
			return "{" + strings.Join(parts, ", ") + "}"
		}

		// 复杂对象，多行输出
		result := "{\n"
		for i, key := range keys {
			val := obj[key]
			valStr := prettyPrintValue(val, indent+1)
			// 如果值是多行的，需要特殊处理
			if strings.Contains(valStr, "\n") {
				result += fmt.Sprintf("%s  %s:\n%s", prefix, key, valStr)
			} else {
				result += fmt.Sprintf("%s  %s: %s", prefix, key, valStr)
			}
			if i < len(keys)-1 {
				result += ","
			}
			result += "\n"
		}
		result += prefix + "}"
		return result

	case engine.TypeString:
		return fmt.Sprintf("%q", v.String())

	default:
		return v.Stringify()
	}
}

// IOSigs returns function signatures for REPL :doc command.
func IOSigs() map[string]string {
	return map[string]string{
		"print":                "print(args...) → null  — Output to stdout without newline",
		"println":              "println(args...) → null  — Output to stdout with newline",
		"puts":                 "puts(args...) → null  — Output to stdout without quotes, with newline",
		"pp":                   "pp(args...) → null  — Pretty print formatted output",
		"echo":                 "echo(args...) → string  — Concatenate arguments as string",
		"format":               "format(template, args...) → string  — Format string with %s placeholders",
		"assert":               "assert(condition, [message]) → null  — Assertion check",
		"fopen":                "fopen(path, [mode]) → stream  — Open file stream",
		"fread":                "fread(stream, length) → string  — Read bytes from stream",
		"fgets":                "fgets(stream) → string  — Read line from stream",
		"fwrite":               "fwrite(stream, data) → int  — Write data to stream",
		"fclose":               "fclose(stream) → null  — Close stream",
		"feof":                 "feof(stream) → bool  — Check if stream is at EOF",
		"fflush":               "fflush(stream) → bool  — Flush stream buffer",
		"stream_get_meta_data": "stream_get_meta_data(stream) → object  — Get stream metadata",
		"is_readable":          "is_readable(path_or_stream) → bool  — Check if readable",
		"is_writable":          "is_writable(path_or_stream) → bool  — Check if writable",
	}
}
