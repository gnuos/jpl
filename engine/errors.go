package engine

import "fmt"

// 错误类型定义
var (
	ErrEngineClosed     = NewEngineError("engine is closed")
	ErrVMClosed         = NewEngineError("vm is closed")
	ErrCompileFailed    = NewCompileError("compilation failed")
	ErrRuntimeFailed    = NewRuntimeError("execution failed")
	ErrInvalidArg       = NewEngineError("invalid argument")
	ErrTypeMismatch     = NewRuntimeError("type mismatch")
	ErrDivideByZero     = NewRuntimeError("division by zero")
	ErrUndefinedVar     = NewRuntimeError("undefined variable")
	ErrUndefinedFunc    = NewRuntimeError("undefined function")
	ErrIndexOutOfBounds = NewRuntimeError("index out of bounds")
	ErrStackOverflow    = NewRuntimeError("stack overflow: maximum call depth exceeded")
	ErrInterrupted      = NewRuntimeError("execution interrupted")
)

// EngineError 引擎级别错误
type EngineError struct {
	Message string
}

func NewEngineError(message string) *EngineError {
	return &EngineError{Message: message}
}

func (e *EngineError) Error() string {
	return fmt.Sprintf("engine error: %s", e.Message)
}

// CompileError 编译错误
type CompileError struct {
	Message string
	Line    int
	Column  int
	File    string
}

func NewCompileError(message string) *CompileError {
	return &CompileError{Message: message}
}

func (e *CompileError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("compile error at %s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("compile error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("compile error: %s", e.Message)
}

// RuntimeError 运行时错误
type RuntimeError struct {
	Message string
	Line    int
	Column  int
	File    string
}

func NewRuntimeError(message string) *RuntimeError {
	return &RuntimeError{Message: message}
}

// NewRuntimeErrorWithLocation 创建带位置的运行时错误
func NewRuntimeErrorWithLocation(message string, line, column int, file string) *RuntimeError {
	return &RuntimeError{
		Message: message,
		Line:    line,
		Column:  column,
		File:    file,
	}
}

// WithSourceContext 为运行时错误附加源码上下文
// 返回新的 RuntimeError，包含格式化的源码上下文信息
func (e *RuntimeError) WithSourceContext(sourceLines []string) *RuntimeError {
	if e.Line <= 0 || sourceLines == nil {
		return e
	}
	return e
}

func (e *RuntimeError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("runtime error at %s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("runtime error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("runtime error: %s", e.Message)
}

// FormatWithContext 返回带源码上下文的格式化错误信息
func (e *RuntimeError) FormatWithContext(sourceLines []string) string {
	if e.Line <= 0 || sourceLines == nil || len(sourceLines) == 0 {
		return e.Error()
	}

	lineIdx := e.Line - 1 // 转为 0-based 索引
	if lineIdx >= len(sourceLines) {
		return e.Error()
	}

	// 计算显示范围（前后各 2 行）
	start := lineIdx - 2
	if start < 0 {
		start = 0
	}
	end := lineIdx + 2
	if end >= len(sourceLines) {
		end = len(sourceLines) - 1
	}

	var buf string
	buf += e.Error() + "\n"

	// 计算行号宽度用于对齐
	lineNumWidth := 0
	for n := end + 1; n > 0; n /= 10 {
		lineNumWidth++
	}

	for i := start; i <= end; i++ {
		lineNum := i + 1
		prefix := "   "
		if i == lineIdx {
			prefix = " → "
		}
		buf += fmt.Sprintf(fmt.Sprintf("%%s%%%dd | %%s\n", lineNumWidth), prefix, lineNum, sourceLines[i])

		// 在错误行添加下划线标记
		if i == lineIdx && e.Column > 0 {
			marker := fmt.Sprintf("   %s", spaces(lineNumWidth+1))
			marker += spaces(e.Column - 1)
			marker += "^"
			buf += marker + "\n"
		}
	}

	return buf
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	s := make([]byte, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}

// ExitError 脚本退出错误（用于 exit/die 函数）
//
// **内部实现说明**：
// 1. 这不是真正的错误，而是请求脚本正常终止的信号
// 2. 实现了 Go 的 error 接口，因此可以通过函数返回 error 类型传递
// 3. VM 在收到 ExitError 时会特殊处理：
//   - 不将其视为异常（不触发 catch）
//   - 保存退出码到 vm.exitCode
//   - 将 vm.err 设为 nil
//   - 返回 nil 表示正常终止
//     4. 这种设计允许 exit/die 函数通过 error 返回值与 VM 通信
//     同时保持 Go 语言的错误处理习惯的兼容性
//
// **使用场景**：
//   - exit() 函数创建 ExitError 并返回
//   - die() 函数创建 ExitError（包含消息）并返回
//   - VM 在多个位置检查 ExitError：
//   - opCall 中检查 Go 函数返回值
//   - opTailCall 中检查尾调用返回值
//   - run() 主循环中检查 vm.err
//   - CLI 工具通过 vm.GetExitCode() 获取退出码
//
// **字段说明**：
//   - Code: 退出码，传递给操作系统的值（0-255）
//   - Message: 可选的退出消息，用于调试和日志
//
// **与 RuntimeError 的区别**：
//   - RuntimeError: 真正的错误，会触发 catch，执行失败
//   - ExitError: 正常终止信号，不触发 catch，执行成功但提前结束
type ExitError struct {
	Code    int    // 退出码（0-255），传递给操作系统
	Message string // 可选的退出消息，用于调试和日志
}

// NewExitError 创建新的退出错误
//
// **参数说明**：
//   - code: 退出码（0-255），0 表示成功，非 0 表示各种错误
//   - message: 可选的退出消息，空字符串表示无消息
//
// **返回值**：
//   - *ExitError: 包含退出码和消息的退出错误对象
//
// **实现细节**：
//   - 直接构造 ExitError 结构体
//   - 不验证退出码范围（由调用者确保在 0-255 之间）
//   - 消息可以为空，此时 Error() 方法返回简化格式
func NewExitError(code int, message string) *ExitError {
	return &ExitError{Code: code, Message: message}
}

// Error 实现 error 接口，返回格式化错误字符串
//
// **实现细节**：
//   - 如果有消息，格式为 "exit(code): message"
//   - 如果无消息，格式为 "exit(code)"
//   - 此格式用于调试和日志记录
//
// **注意**：这个错误字符串不会被 VM 当作错误显示给用户
// 因为 VM 会将 ExitError 识别为正常终止信号
func (e *ExitError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("exit(%d): %s", e.Code, e.Message)
	}
	return fmt.Sprintf("exit(%d)", e.Code)
}

// IsExitError 检查错误是否为退出错误
//
// **用途**：
//   - 在 VM 中用于区分 ExitError 和真正的错误
//   - CLI 工具可以用它来判断脚本是正常退出还是异常终止
//
// **参数**：
//   - err: 任何 error 类型（可能为 nil）
//
// **返回值**：
//   - true: 是 ExitError（脚本正常终止）
//   - false: 不是 ExitError（可能是 RuntimeError 或其他错误）
//
// **实现细节**：
//   - 使用类型断言检查 err 是否为 *ExitError
//   - 如果 err 为 nil，返回 false
func IsExitError(err error) bool {
	_, ok := err.(*ExitError)
	return ok
}

// GetExitCode 从错误中获取退出码
//
// **用途**：
//   - CLI 工具获取 exit/die 函数设置的退出码
//   - 将退出码传递给操作系统（os.Exit）
//
// **参数**：
//   - err: 任何 error 类型（可能为 nil）
//
// **返回值**：
//   - >= 0: 是 ExitError，返回实际的退出码
//   - -1: 不是 ExitError（可能是其他错误或 nil）
//
// **实现细节**：
//   - 使用类型断言检查 err 是否为 *ExitError
//   - 如果是，返回 exitErr.Code
//   - 如果不是，返回 -1 作为哨兵值
//
// **使用示例**：
//
//	err := vm.Execute()
//	exitCode := engine.GetExitCode(err)
//	if exitCode >= 0 {
//	    os.Exit(exitCode)  // 使用 exit/die 设置的退出码
//	} else if err != nil {
//	    os.Exit(1)  // 其他错误，使用默认退出码
//	}
func GetExitCode(err error) int {
	if exitErr, ok := err.(*ExitError); ok {
		return exitErr.Code
	}
	return -1
}
