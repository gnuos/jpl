package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterEval 注册 eval() 函数
func RegisterEval(e *engine.Engine) {
	e.RegisterFunc("eval", builtinEval)
}

// EvalNames 返回 eval 函数名称列表
func EvalNames() []string {
	return []string{"eval"}
}

// builtinEval 执行字符串中的 JPL 代码
//
// ⚠️ 安全警告：eval() 是一个危险的函数，使用不当会导致严重安全漏洞
//
// 主要风险：
//   - 代码注入攻击：不要将用户输入直接传给 eval()
//   - 恶意代码执行：任何在 JPL 中可执行的代码都可能通过 eval() 运行
//   - 性能问题：每次调用都会重新编译代码，比直接执行慢得多
//
// 安全的替代方案：
//   - 尽量避免使用 eval()，重新设计代码结构
//   - 如果需要解析配置，使用 JSON/YAML 等数据格式而非代码
//   - 如果必须使用，确保输入来源可信并进行严格验证
//
// eval(code) — code 必须为字符串
// 返回脚本执行结果
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 要执行的 JPL 代码字符串
//
// 返回值：
//   - Value: 脚本执行结果（可以是任意类型）
//   - error: 编译错误或运行时错误
//
// 使用示例：
//
//	// 简单的表达式计算
//	$result = eval("1 + 2 * 3")  // → 7
//
//	// 动态执行函数
//	$code = '
//	    $x = 10
//	    $y = 20
//	    return $x + $y
//	'
//	$result = eval($code)  // → 30
//
//	// 条件执行
//	$condition = true
//	if ($condition) {
//	    eval("println('条件为真')")
//	}
//
// 错误处理示例：
//
//	// 编译错误会被捕获
//	$result = eval("for $i in [1,2] {")  // 编译错误
//
//	// 运行时错误也会被捕获
//	$result = eval("x + y")  // 运行时错误：变量未定义
func builtinEval(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("eval() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("eval() argument must be a string, got %s", args[0].Type())
	}

	code := args[0].String()

	// 获取引擎实例
	eng := ctx.Engine()
	if eng == nil {
		return nil, fmt.Errorf("eval() requires an engine context")
	}

	// 编译并执行
	prog, err := engine.CompileStringWithName(code, "<eval>")
	if err != nil {
		return nil, fmt.Errorf("eval() compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(eng, prog)
	if err := vm.Execute(); err != nil {
		return nil, fmt.Errorf("eval() runtime error: %v", err)
	}

	return vm.GetResult(), nil
}
