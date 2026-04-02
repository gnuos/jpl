package engine

import (
	"fmt"
	"sync"
)

// GoFunction 是 Go 语言注册函数类型，用于将 Go 函数暴露给 JPL 脚本调用。
//
// 参数说明：
//   - ctx: 执行上下文，包含当前 VM 引用和调用信息
//   - args: 调用时传入的参数列表
//
// 返回值：
//   - Value: 函数执行结果
//   - error: 如果执行出错，返回错误信息
//
// 示例：
//
//	engine.RegisterFunc("add", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    if len(args) != 2 {
//	        return nil, fmt.Errorf("add() requires 2 arguments")
//	    }
//	    a := args[0].Int()
//	    b := args[1].Int()
//	    return engine.NewInt(a + b), nil
//	})
type GoFunction func(ctx *Context, args []Value) (Value, error)

// ModuleCache 表示已缓存的模块，包含模块导出的所有符号。
//
// 模块系统将模块的导出内容缓存在 ModuleCache 中，避免重复加载。
// Exports 映射中包含所有导出的变量和函数，键为符号名称（不包含 $ 前缀）。
//
// 使用场景：
//   - 文件模块加载后将导出符号缓存
//   - Go 注册模块的导出符号存储
//   - URL 导入模块的磁盘缓存
//
// 相关方法：
//   - Engine.GetModule: 获取已缓存的模块
//   - Engine.CacheModule: 将模块添加到缓存
//   - Engine.RegisterModule: 注册 Go 实现的模块
type ModuleCache struct {
	Exports map[string]Value // 导出的符号（变量和函数）
}

// ModuleLoader 是模块加载器接口，定义了加载模块的契约。
//
// 模块加载器负责从各种来源（文件系统、URL、内置模块等）加载 JPL 模块。
// 实现此接口可以自定义模块加载行为，例如添加特定的模块搜索路径或协议支持。
//
// 内置实现：
//   - FileModuleLoader: 从文件系统加载 .jpl 文件
//
// 使用示例：
//
//	loader := &engine.FileModuleLoader{
//	    SearchPaths: []string{"./jpl_modules", "~/.jpl/modules"},
//	}
//	engine.SetModuleLoader(loader)
//	vm, _ := engine.Compile(`import "my_module"`)
//	vm.Execute()
type ModuleLoader interface {
	// LoadModule 加载指定来源的模块，返回导出的符号映射。
	//
	// 参数：
	//   - source: 模块来源标识符（文件路径、URL、模块名等）
	//   - engine: 当前引擎实例，用于编译和执行模块代码
	//
	// 返回值：
	//   - *ModuleCache: 包含导出符号的缓存对象
	//   - error: 加载失败时返回错误
	LoadModule(source string, engine *Engine) (*ModuleCache, error)
}

// Engine 是 JPL 脚本引擎的核心结构体，管理脚本执行的全局环境。
//
// Engine 提供了线程安全的脚本执行环境，包括：
//   - 变量管理：设置/获取全局变量
//   - 函数注册：将 Go 函数暴露给脚本调用
//   - 常量定义：预定义的不可变值
//   - 错误日志：记录执行过程中的错误
//   - 模块系统：支持 import/include 语句的模块化编程
//
// 线程安全：
//
//	Engine 的所有导出方法都是线程安全的，内部使用 sync.RWMutex 保护共享状态。
//	可以同时从多个 goroutine 调用 Engine 的方法。
//
// 生命周期：
//  1. NewEngine() 创建实例
//  2. 注册函数/变量/常量
//  3. Compile() 编译脚本
//  4. vm.Execute() 执行脚本
//  5. Close() 释放资源（可选，但建议显式调用）
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册 Go 函数
//	eng.RegisterFunc("hello", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    return engine.NewString("Hello, World!"), nil
//	})
//
//	// 编译并执行脚本
//	vm, err := eng.Compile(`hello()`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := vm.Execute(); err != nil {
//	    log.Fatal(err)
//	}
type Engine struct {
	mu        sync.RWMutex          // 保护并发访问
	closed    bool                  // 引擎是否已关闭
	variables map[string]Value      // 全局变量映射
	functions map[string]GoFunction // 注册的 Go 函数映射
	constants map[string]Value      // 预定义常量映射
	errLog    []error               // 错误日志
	stdLib    map[string]any        // 标准库内部使用

	// 模块系统
	modules      map[string]*ModuleCache // 模块缓存（name -> cached module）
	goModules    map[string]*ModuleCache // Go 注册的模块（用于 Go 实现的模块）
	moduleLoader ModuleLoader            // 模块加载器（加载 .jpl 文件等）
}

// NewEngine 创建一个新的 JPL 脚本引擎实例。
//
// 新建的引擎实例处于就绪状态，可以立即用于：
//   - 注册 Go 函数（RegisterFunc）
//   - 设置变量（Set）
//   - 编译脚本（Compile）
//
// 资源管理：
//
//	建议使用 defer engine.Close() 确保资源正确释放，虽然 Go 的垃圾回收
//	最终也会回收，但显式关闭可以立即释放内部映射占用的内存。
//
// 返回值：
//   - *Engine: 新创建的引擎实例，永远不会返回 nil
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//	vm, err := eng.Compile(`print("Hello")`)
func NewEngine() *Engine {
	return &Engine{
		variables: make(map[string]Value),
		functions: make(map[string]GoFunction),
		constants: make(map[string]Value),
		stdLib:    make(map[string]any),
		modules:   make(map[string]*ModuleCache),
		goModules: make(map[string]*ModuleCache),
	}
}

// Close 关闭引擎实例，释放所有内部资源。
//
// 关闭后，引擎不能再使用，所有方法都会返回 ErrEngineClosed 错误。
// 此方法会：
//   - 清空所有变量映射
//   - 清空所有函数映射
//   - 清空所有常量映射
//   - 释放标准库资源
//
// 线程安全：可以安全地从多个 goroutine 调用，但只有第一次调用有效。
//
// 返回值：
//   - nil: 成功关闭
//   - ErrEngineClosed: 引擎已经关闭过
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	// ... 使用引擎 ...
//	if err := eng.Close(); err != nil {
//	    log.Printf("关闭引擎失败: %v", err)
//	}
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrEngineClosed
	}

	e.closed = true
	e.variables = nil
	e.functions = nil
	e.constants = nil
	e.stdLib = nil

	return nil
}

// IsClosed 检查引擎是否已关闭。
//
// 返回 true 表示引擎已经关闭，此时再调用其他方法都会返回错误。
// 此方法常用于测试或调试时检查引擎状态。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 返回值：
//   - true: 引擎已关闭
//   - false: 引擎处于活动状态
func (e *Engine) IsClosed() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.closed
}

// Set 设置引擎全局变量的值。
//
// 此方法用于在脚本执行前预设置变量值，或在脚本执行后与脚本交互。
// 设置的变量可以在脚本中通过 $name 语法访问。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 变量名，不能为空字符串
//   - value: 变量值，可以是任何 Value 类型
//
// 返回值：
//   - nil: 设置成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrInvalidArg: 变量名为空
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 设置脚本可以访问的变量
//	eng.Set("username", engine.NewString("Alice"))
//	eng.Set("count", engine.NewInt(42))
//
//	vm, _ := eng.Compile(`print($username)`)
//	vm.Execute()
func (e *Engine) Set(name string, value Value) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrEngineClosed
	}

	if name == "" {
		return ErrInvalidArg
	}

	e.variables[name] = value
	return nil
}

// Get 获取引擎全局变量的值。
//
// 此方法用于获取之前在脚本执行前设置的变量，或脚本执行后导出的变量。
// 如果变量不存在，返回 ErrUndefinedVar 错误。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 变量名，区分大小写
//
// 返回值：
//   - Value: 变量的当前值
//   - nil: 获取成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrUndefinedVar: 变量不存在
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	vm, _ := eng.Compile(`$result = 42 * 2`)
//	vm.Execute()
//
//	value, err := eng.Get("result")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Result: %v\n", value.Int()) // 输出: 84
func (e *Engine) Get(name string) (Value, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	// 首先检查变量
	value, exists := e.variables[name]
	if exists {
		return value, nil
	}

	// 检查是否是已注册的函数
	if fn, exists := e.functions[name]; exists {
		return &funcValue{name: name, fn: fn}, nil
	}

	return nil, ErrUndefinedVar
}

// RegisterFunc 将 Go 函数注册到引擎，使其可以在 JPL 脚本中被调用。
//
// 注册的函数可以在脚本中像普通函数一样使用，例如：
//
//	my_func($arg1, $arg2)
//
// 如果同名函数已存在，将被新函数覆盖。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 函数名，不能为空字符串，将作为脚本中调用的函数名
//   - fn: Go 函数实现，必须符合 GoFunction 签名
//
// 返回值：
//   - nil: 注册成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrInvalidArg: 函数名为空或函数为 nil
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册一个简单的加法函数
//	err := eng.RegisterFunc("add", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    if len(args) != 2 {
//	        return nil, fmt.Errorf("add() requires 2 arguments, got %d", len(args))
//	    }
//	    result := args[0].Int() + args[1].Int()
//	    return engine.NewInt(result), nil
//	})
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 在脚本中使用
//	vm, _ := eng.Compile(`print(add(10, 20))`)
//	vm.Execute() // 输出: 30
func (e *Engine) RegisterFunc(name string, fn GoFunction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrEngineClosed
	}

	if name == "" || fn == nil {
		return ErrInvalidArg
	}

	e.functions[name] = fn
	return nil
}

// GetRegisteredFunc 获取已注册的 Go 函数。
//
// 此方法用于查询引擎中是否存在指定名称的 Go 函数。
// 如果函数不存在，返回 nil。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 函数名，区分大小写
//
// 返回值：
//   - GoFunction: 注册的函数实现，如果不存在返回 nil
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	eng.RegisterFunc("test", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    return engine.NewNull(), nil
//	})
//
//	fn := eng.GetRegisteredFunc("test")
//	if fn != nil {
//	    fmt.Println("Function 'test' is registered")
//	}
func (e *Engine) GetRegisteredFunc(name string) GoFunction {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.functions[name]
}

// GetConst 查询预定义常量是否存在并获取其值。
//
// 常量是在引擎初始化时或通过 RegisterConst 设置的不可变值。
// 脚本可以通过常量名访问这些值，但不能修改它们。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 常量名，区分大小写
//
// 返回值：
//   - Value: 常量的值，如果不存在返回 nil
//   - bool: 如果常量存在返回 true，否则返回 false
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册常量
//	eng.RegisterConst("PI", engine.NewFloat(3.14159))
//
//	// 查询常量
//	if value, ok := eng.GetConst("PI"); ok {
//	    fmt.Printf("PI = %v\n", value.Float())
//	}
func (e *Engine) GetConst(name string) (Value, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	v, ok := e.constants[name]
	return v, ok
}

// RegisterConst 注册预定义常量到引擎。
//
// 常量是在引擎级别定义的全局不可变值，脚本可以读取但不能修改。
// 这与通过 Set() 设置的变量不同，变量可以被脚本修改。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 常量名，不能为空字符串
//   - value: 常量值
//
// 返回值：
//   - nil: 注册成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrInvalidArg: 常量名为空
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册数学常量
//	eng.RegisterConst("PI", engine.NewFloat(3.14159265359))
//	eng.RegisterConst("E", engine.NewFloat(2.71828182846))
//
//	// 脚本中使用（注意：常量名不需要 $ 前缀）
//	vm, _ := eng.Compile(`print(PI)`)
//	vm.Execute()
func (e *Engine) RegisterConst(name string, value Value) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrEngineClosed
	}

	if name == "" {
		return ErrInvalidArg
	}

	e.constants[name] = value
	return nil
}

// GetConstantNames 返回所有已注册的常量名称列表。
//
// 此方法用于反射功能，获取引擎中所有常量的名称。
// 返回的名称列表不包含重复项，顺序不确定。
//
// 返回值：
//   - []string: 常量名称列表，如果引擎已关闭则返回 nil
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	eng.RegisterConst("PI", engine.NewFloat(3.14159))
//	eng.RegisterConst("E", engine.NewFloat(2.71828))
//
//	names := eng.GetConstantNames()
//	fmt.Printf("常量: %v\n", names) // 输出: [PI E]
func (e *Engine) GetConstantNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil
	}

	names := make([]string, 0, len(e.constants))
	for name := range e.constants {
		names = append(names, name)
	}
	return names
}

// Compile 编译 JPL 脚本字符串，返回可执行的虚拟机实例。
//
// 此方法将 JPL 源代码编译为字节码，并创建一个配置好的 VM 实例。
// 编译过程包括词法分析、语法分析、AST 构建和字节码生成。
//
// 注意：编译后需要调用 vm.Execute() 执行脚本。
// 编译时不会执行脚本，只是准备执行环境。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - script: JPL 源代码字符串
//
// 返回值：
//   - *VM: 编译后的虚拟机实例，可以执行脚本
//   - nil: 编译成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrInvalidArg: 脚本为空
//   - ErrCompileFailed: 编译失败（语法错误等）
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 编译脚本
//	vm, err := eng.Compile(`
//	    function greet(name) {
//	        return "Hello, " + name + "!"
//	    }
//	    $result = greet("World")
//	`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 执行脚本
//	if err := vm.Execute(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// 获取结果
//	result, _ := eng.Get("result")
//	fmt.Printf("Result: %s\n", result.String())
func (e *Engine) Compile(script string) (*VM, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.closed {
		return nil, ErrEngineClosed
	}

	if script == "" {
		return nil, ErrInvalidArg
	}

	// TODO: 实现编译器，将脚本编译为字节码
	// 这里返回一个空的 VM 作为占位
	return newVM(e), nil
}

// CompileFile 编译 JPL 脚本文件，返回可执行的虚拟机实例。
//
// 此方法读取指定文件的内容并调用 Compile 进行编译。
// 文件路径可以是相对路径或绝对路径。
//
// 注意：此方法会读取整个文件到内存，对于大文件请考虑分块处理。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - filename: 脚本文件路径
//
// 返回值：
//   - *VM: 编译后的虚拟机实例
//   - nil: 编译成功
//   - ErrCompileFailed: 编译失败
//   - 其他错误: 文件读取失败等
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 编译脚本文件
//	vm, err := eng.CompileFile("script.jpl")
//	if err != nil {
//	    log.Fatalf("编译失败: %v", err)
//	}
//
//	// 执行脚本
//	if err := vm.Execute(); err != nil {
//	    log.Fatalf("执行失败: %v", err)
//	}
func (e *Engine) CompileFile(filename string) (*VM, error) {
	// TODO: 读取文件并编译
	return nil, ErrCompileFailed
}

// GetLastError 获取引擎错误日志中的最后一个错误。
//
// 此方法返回引擎执行过程中记录的最后一个错误。
// 如果错误日志为空，返回 nil。
//
// 注意：错误日志在 ClearErrorLog 被调用或引擎关闭后被清空。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 返回值：
//   - error: 最后一个错误，如果没有错误返回 nil
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// ... 执行一些操作 ...
//
//	if err := eng.GetLastError(); err != nil {
//	    fmt.Printf("Last error: %v\n", err)
//	}
func (e *Engine) GetLastError() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.errLog) == 0 {
		return nil
	}
	return e.errLog[len(e.errLog)-1]
}

// GetErrorLog 获取引擎中的所有错误日志。
//
// 此方法返回引擎执行过程中记录的所有错误的副本。
// 返回的切片是独立的副本，修改它不会影响引擎内部状态。
//
// 注意：错误日志可以通过 ClearErrorLog 手动清空。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 返回值：
//   - []error: 错误日志的副本，如果没有错误返回空切片（不是 nil）
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// ... 执行一些操作 ...
//
//	errors := eng.GetErrorLog()
//	for i, err := range errors {
//	    fmt.Printf("Error %d: %v\n", i+1, err)
//	}
//
//	fmt.Printf("Total errors: %d\n", len(errors))
func (e *Engine) GetErrorLog() []error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]error, len(e.errLog))
	copy(result, e.errLog)
	return result
}

// ClearErrorLog 清空引擎的错误日志。
//
// 此方法清空引擎记录的所有错误，释放相关内存。
// 在长时间运行的应用中，建议定期清理错误日志以避免内存泄漏。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// ... 执行一些操作，可能产生错误 ...
//
//	// 清空错误日志，准备下一批操作
//	eng.ClearErrorLog()
//
//	// 继续执行...
func (e *Engine) ClearErrorLog() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errLog = nil
}

// logError 将错误记录到引擎的错误日志中。
//
// 此方法由引擎内部使用，用于记录执行过程中的错误。
// 用户代码通常不需要直接调用此方法，错误会自动记录。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - err: 要记录的错误
func (e *Engine) logError(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errLog = append(e.errLog, err)
}

// ============================================================================
// 模块系统
// ============================================================================

// SetModuleLoader 设置引擎的模块加载器。
//
// 模块加载器负责从各种来源（文件系统、URL 等）加载 JPL 模块。
// 设置后，引擎可以使用 import 和 include 语句加载外部模块。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - loader: 模块加载器实例，如果为 nil 则禁用模块加载
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 配置模块加载器
//	loader := &engine.FileModuleLoader{
//	    SearchPaths: []string{"./jpl_modules", "~/.jpl/modules"},
//	}
//	eng.SetModuleLoader(loader)
//
//	// 现在可以使用 import 语句
//	vm, _ := eng.Compile(`import "my_module"`)
//	vm.Execute()
func (e *Engine) SetModuleLoader(loader ModuleLoader) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.moduleLoader = loader
}

// RegisterModule 注册一个 Go 实现的模块。
//
// 此方法用于将 Go 实现的函数打包为一个模块，使其可以通过 import 语句导入。
// 注册的模块函数会被包装为 funcValue 类型，存储在 goModules 中。
//
// 与 RegisterFunc 不同，此方法注册的函数属于特定模块，不会污染全局命名空间。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - name: 模块名称，将作为 import 语句中的标识符
//   - exports: 模块导出的函数映射，键为函数名，值为 GoFunction
//
// 返回值：
//   - nil: 注册成功
//   - ErrEngineClosed: 引擎已关闭
//   - ErrInvalidArg: 模块名为空
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册 Go 模块
//	err := eng.RegisterModule("math_ext", map[string]engine.GoFunction{
//	    "cube": func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	        n := args[0].Int()
//	        return engine.NewInt(n * n * n), nil
//	    },
//	})
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 在脚本中导入使用
//	vm, _ := eng.Compile(`
//	    import "math_ext"
//	    print(math_ext.cube(3))  // 输出: 27
//	`)
//	vm.Execute()
func (e *Engine) RegisterModule(name string, exports map[string]GoFunction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return ErrEngineClosed
	}

	if name == "" {
		return ErrInvalidArg
	}

	cache := &ModuleCache{
		Exports: make(map[string]Value),
	}
	for fnName, fn := range exports {
		cache.Exports[fnName] = NewFunc(fnName, fn)
	}
	e.goModules[name] = cache
	return nil
}

// GetModule 获取已缓存或已注册的模块。
//
// 此方法按以下顺序查找模块：
//  1. 已缓存的模块（通过 CacheModule 添加）
//  2. Go 注册的模块（通过 RegisterModule 添加）
//
// 如果找到模块，返回 ModuleCache 和 true；否则返回 nil 和 false。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - source: 模块来源标识符（可以是模块名或路径）
//
// 返回值：
//   - *ModuleCache: 模块缓存，如果不存在返回 nil
//   - bool: 如果模块存在返回 true，否则返回 false
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册模块
//	eng.RegisterModule("test_module", map[string]engine.GoFunction{
//	    "hello": func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	        return engine.NewString("Hello!"), nil
//	    },
//	})
//
//	// 查询模块
//	if cache, ok := eng.GetModule("test_module"); ok {
//	    fmt.Printf("Module has %d exports\n", len(cache.Exports))
//	}
func (e *Engine) GetModule(source string) (*ModuleCache, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 1. 查找缓存
	if cache, ok := e.modules[source]; ok {
		return cache, true
	}
	// 2. 查找 Go 注册模块
	if cache, ok := e.goModules[source]; ok {
		return cache, true
	}
	return nil, false
}

// CacheModule 将模块缓存到引擎中。
//
// 此方法用于手动将模块添加到缓存，避免重复加载。
// 通常由模块加载器在成功加载模块后调用。
//
// 如果同名模块已存在，将被新模块覆盖。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - source: 模块来源标识符
//   - cache: 包含导出符号的模块缓存
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 手动创建模块缓存
//	exports := map[string]engine.Value{
//	    "version": engine.NewString("1.0.0"),
//	    "author":  engine.NewString("John Doe"),
//	}
//	cache := &engine.ModuleCache{Exports: exports}
//
//	// 缓存模块
//	eng.CacheModule("my_meta", cache)
//
//	// 之后可以通过 GetModule 获取
//	cached, _ := eng.GetModule("my_meta")
func (e *Engine) CacheModule(source string, cache *ModuleCache) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.modules[source] = cache
}

// LoadModule 使用配置的模块加载器加载指定模块。
//
// 此方法调用引擎配置的 ModuleLoader 来加载模块。
// 如果未设置模块加载器（SetModuleLoader），将返回错误。
//
// 注意：此方法不会缓存加载的模块，如果需要缓存，应在加载成功后
// 调用 CacheModule 方法。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - source: 模块来源标识符（文件路径、URL、模块名等）
//
// 返回值：
//   - *ModuleCache: 加载的模块缓存
//   - error: 加载失败时返回错误信息
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 配置模块加载器
//	loader := &engine.FileModuleLoader{
//	    SearchPaths: []string{"./jpl_modules"},
//	}
//	eng.SetModuleLoader(loader)
//
//	// 加载模块
//	module, err := eng.LoadModule("./my_module.jpl")
//	if err != nil {
//	    log.Fatalf("加载模块失败: %v", err)
//	}
//
//	// 使用模块导出的内容
//	for name, value := range module.Exports {
//	    fmt.Printf("Export: %s = %v\n", name, value)
//	}
func (e *Engine) LoadModule(source string) (*ModuleCache, error) {
	e.mu.RLock()
	loader := e.moduleLoader
	e.mu.RUnlock()

	if loader == nil {
		return nil, fmt.Errorf("module not found: %q (no module loader configured)", source)
	}

	return loader.LoadModule(source, e)
}

// RunScriptString 编译并执行 JPL 脚本字符串，返回所有导出的符号。
//
// 此方法是一个高级 API，将编译、执行、导出合并为一个步骤。
// 它会：
//  1. 编译脚本字符串
//  2. 创建 VM 并执行
//  3. 收集所有全局变量和函数作为导出
//
// 导出的全局变量名会自动去掉 $ 前缀（如果存在）。
// 此方法主要用于模块加载器加载 .jpl 文件模块。
//
// 线程安全：可以安全地从多个 goroutine 调用。
//
// 参数：
//   - script: JPL 源代码字符串
//   - filename: 脚本文件名（用于错误报告）
//
// 返回值：
//   - map[string]Value: 导出的符号映射，键为符号名，值为对应 Value
//   - error: 编译或执行失败时返回错误
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	exports, err := eng.RunScriptString(`
//	    $version = "1.0.0"
//	    function greet(name) {
//	        return "Hello, " + name
//	    }
//	`, "module.jpl")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 使用导出的内容
//	fmt.Printf("Version: %s\n", exports["version"].String())
//	greetFunc := exports["greet"]
//	result, _ := eng.VM().CallValue(greetFunc, engine.NewString("World"))
func (e *Engine) RunScriptString(script string, filename string) (map[string]Value, error) {
	prog, err := CompileStringWithName(script, filename)
	if err != nil {
		return nil, err
	}

	vm := NewVMWithProgram(e, prog)
	if err := vm.Execute(); err != nil {
		return nil, err
	}

	// 收集全局变量和函数作为导出
	exports := make(map[string]Value)
	for name, idx := range vm.globalNames {
		// 去掉 $ 前缀用于模块导出
		exportName := name
		if len(exportName) > 0 && exportName[0] == '$' {
			exportName = exportName[1:]
		}
		exports[exportName] = vm.globals[idx]
	}
	// 导出编译后的函数
	for name, fns := range vm.funcMap {
		if len(fns) > 0 {
			exports[name] = wrapCompiledFunc(vm, name)
		}
	}

	return exports, nil
}

// wrapCompiledFunc 将 VM 中的编译函数包装为可在 Go 中调用的 Value。
//
// 此方法创建一个新的 funcValue，将编译后的 JPL 函数包装为 GoFunction。
// 包装后的函数可以通过 VM.CallValue 调用，或在 Go 代码中直接调用。
//
// 注意：这是一个内部辅助函数，用于模块系统导出 JPL 函数。
// 用户代码通常不需要直接调用此方法。
//
// 参数：
//   - vm: 包含函数的虚拟机实例
//   - name: 函数名
//
// 返回值：
//   - Value: 包装后的函数值，类型为 funcValue
//
// 使用示例（内部使用）：
//
//	// 导出编译后的函数供 Go 代码调用
//	fn := wrapCompiledFunc(vm, "myFunc")
//	result, err := vm.CallValue(fn, args...)
func wrapCompiledFunc(vm *VM, name string) Value {
	fn := func(ctx *Context, args []Value) (Value, error) {
		return vm.CallByName(name, args...)
	}
	return NewFunc(name, fn)
}
