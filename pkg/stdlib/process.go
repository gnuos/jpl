package stdlib

import (
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// RegisterProcess 注册进程控制相关的内置函数。
//
// 注册的函数：
//   - exit: 终止脚本执行并返回退出码
//   - die: 输出消息并终止脚本执行（模仿 PHP 的 die）
//
// 这些函数用于控制脚本的生命周期，可以立即终止执行。
// 与 throw 不同，exit/die 无视 try/catch，总是立即终止。
//
// 参数：
//   - e: 引擎实例
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	buildin.RegisterProcess(eng)  // 注册进程控制函数
//
//	vm, _ := eng.Compile(`
//	    if ($error) {
//	        exit(1)
//	    }
//	    die("Fatal error occurred", 2)
//	`)
func RegisterProcess(e *engine.Engine) {
	e.RegisterFunc("exit", builtinExit)
	e.RegisterFunc("die", builtinDie)
}

// ProcessNames 返回进程控制函数名称列表
//
// 用于代码补全和函数枚举。
//
// 返回值：
//   - []string: 进程控制函数名列表 ["exit", "die"]
func ProcessNames() []string {
	return []string{"exit", "die"}
}

// builtinExit 终止脚本执行并返回指定的退出码。
//
// **内部实现说明**：
// 1. 该函数通过返回 ExitError 来触发脚本终止
// 2. ExitError 不是真正的错误，而是向 VM 发送的终止信号
// 3. VM 在收到 ExitError 时会：
//   - 停止当前指令执行
//   - 将退出码保存到 vm.exitCode
//   - 将错误标记为 nil（表示正常终止而非异常）
//   - 返回 nil 给调用者
//
// 4. 此机制确保 exit 无视 try/catch，因为 VM 不将 ExitError 当作异常处理
//
// 参数：
//   - ctx: 执行上下文（包含引擎和 VM 实例）
//   - args[0]: 可选的退出码（整数，默认为 0）
//
// 返回值：
//   - Value: 总是返回 nil（因为函数永不正常返回）
//   - error: 总是返回 *ExitError（包含退出码信息）
//
// 使用示例：
//
//	exit()       // 退出码 0
//	exit(0)      // 退出码 0（成功）
//	exit(1)      // 退出码 1（一般错误）
//
//	// 无视 try/catch
//	try {
//	    exit(1)  // 直接终止，不会进入 catch
//	} catch ($e) {
//	    println "这不会执行"
//	}
func builtinExit(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 解析退出码参数，默认为 0
	code := 0
	if len(args) >= 1 {
		code = int(args[0].Int())
	}

	// 创建 ExitError 并返回，通知 VM 需要终止执行
	// 注意：这不是错误，而是正常终止的信号
	return nil, engine.NewExitError(code, "")
}

// builtinDie 输出消息并终止脚本执行。
//
// **内部实现说明**：
// 1. 该函数首先输出消息到 stdout（如果提供了消息参数）
// 2. 然后返回 ExitError 触发脚本终止
// 3. 与 exit 的区别：die 在终止前会输出消息，而 exit 静默终止
// 4. 消息参数通过 fmt.Println 输出，支持任意类型的字符串表示
// 5. 输出和终止是原子操作，确保消息一定能被打印
//
// **参数处理逻辑**：
//   - 0 参数：静默退出，退出码 0
//   - 1 参数：输出消息，退出码 0
//   - 2 参数：输出消息，指定退出码
//
// 参数：
//   - ctx: 执行上下文（包含引擎和 VM 实例）
//   - args[0]: 可选的消息字符串（任何类型都会被转换为字符串）
//   - args[1]: 可选的退出码（整数，默认为 0）
//
// 返回值：
//   - Value: 总是返回 nil（因为函数永不正常返回）
//   - error: 总是返回 *ExitError（包含退出码和消息）
//
// 使用示例：
//
//	die()                      // 直接退出，退出码 0
//	die("配置加载失败")         // 输出消息，退出码 0
//	die("数据库连接失败", 2)    // 输出消息，退出码 2
//
//	// 实际应用
//	if ($config == null) {
//	    die("配置文件未找到", 1)
//	}
func builtinDie(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 初始化消息和退出码
	message := ""
	code := 0

	// 解析参数
	// args[0] = 消息（可选）
	if len(args) >= 1 {
		message = args[0].String()
	}
	// args[1] = 退出码（可选）
	if len(args) >= 2 {
		code = int(args[1].Int())
	}

	// 如果有消息，先输出到 stdout
	// 使用 fmt.Println 确保消息末尾有换行符
	if message != "" {
		fmt.Println(message)
	}

	// 创建 ExitError 并返回
	// 消息被包含在 ExitError 中，用于调试和日志
	return nil, engine.NewExitError(code, message)
}
