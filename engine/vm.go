package engine

import (
	"fmt"
	"maps"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gnuos/jpl/gc"
)

// ============================================================================
// 调用栈帧
// ============================================================================

// callFrame 函数调用栈帧
type callFrame struct {
	ip           int               // 返回指令指针
	function     *CompiledFunction // 调用者函数
	registers    []Value           // 调用者寄存器窗口
	resultReg    int               // 返回值存储位置
	registerBase int               // 调用者寄存器基址（兼容）
	upvals       []*upvalue        // 调用者的 upvalue 列表
	tryHandlers  []tryHandler      // 调用者的 try 处理器栈
}

// ============================================================================
// Upvalue 和 Closure
// ============================================================================

// upvalue 捕获的外部变量
type upvalue struct {
	location *Value // 指向寄存器或栈上的值
	closed   Value  // 关闭后的值（函数返回时）
	isClosed bool   // 是否已关闭
}

// closure 闭包对象
type closure struct {
	function *CompiledFunction // 编译后的函数
	upvals   []*upvalue        // 捕获的 upvalue 列表
	// GC 字段
	gcID     uint64
	refCount int
	gcPtr    *gc.GC
	alive    bool
}

// 实现 Value 接口
func (c *closure) Type() ValueType          { return TypeFunc }
func (c *closure) IsNull() bool             { return false }
func (c *closure) Bool() bool               { return true }
func (c *closure) Int() int64               { return 0 }
func (c *closure) Float() float64           { return 0 }
func (c *closure) String() string           { return "<closure:" + c.function.Name + ">" }
func (c *closure) Array() []Value           { return nil }
func (c *closure) Object() map[string]Value { return nil }
func (c *closure) Stringify() string        { return c.String() }
func (c *closure) Len() int                 { return 0 }
func (c *closure) IsTrue() bool             { return true }
func (c *closure) ToBigInt() Value          { return c }
func (c *closure) ToBigDecimal() Value      { return c }

func (c *closure) Equals(other Value) bool {
	if other.Type() != TypeFunc {
		return false
	}
	if o, ok := other.(*closure); ok {
		return c.function == o.function
	}
	return false
}

func (c *closure) Add(other Value) Value { return NewString(c.Stringify() + other.Stringify()) }
func (c *closure) Sub(other Value) Value { return NewNull() }
func (c *closure) Mul(other Value) Value { return NewNull() }
func (c *closure) Div(other Value) Value { return NewNull() }
func (c *closure) Mod(other Value) Value { return NewNull() }
func (c *closure) Negate() Value         { return NewNull() }

func (c *closure) Less(other Value) bool         { return false }
func (c *closure) Greater(other Value) bool      { return false }
func (c *closure) LessEqual(other Value) bool    { return false }
func (c *closure) GreaterEqual(other Value) bool { return false }

// closure 实现 gc.ManagedObject 接口
func (c *closure) ObjID() uint64    { return c.gcID }
func (c *closure) GetRefCount() int { return c.refCount }
func (c *closure) IsAlive() bool    { return c.alive }
func (c *closure) OnFree()          { c.upvals = nil }

func (c *closure) IncRef() {
	c.refCount++
}

func (c *closure) DecRef() {
	c.refCount--
	if c.refCount <= 0 {
		c.alive = false
	}
}

func (c *closure) MarkChildren(marker func(child any)) {
	// upvalue 中可能持有 Value 引用
	for _, uv := range c.upvals {
		if uv != nil {
			if uv.isClosed {
				marker(uv.closed)
			} else if uv.location != nil {
				marker(*uv.location)
			}
		}
	}
}

// SetupGC 设置 GC 并注册到 GC
func (c *closure) SetupGC(g *gc.GC) {
	if g == nil {
		return
	}
	c.gcPtr = g
	c.gcID = g.NextID()
	c.alive = true
	c.refCount = 1
	g.Register(c)
}

// ============================================================================
// 虚拟机
// ============================================================================

const (
	defaultMaxCallDepth          = 1000 // 默认最大调用深度
	defaultInterruptCheckInteval = 1000 // 默认中断检查频率（每 N 条指令）
)

// tryHandler try 块处理器
type tryHandler struct {
	catchPC     int // catch 块的起始 PC
	catchVarReg int // catch 变量的寄存器索引
}

// VM 是 JPL 虚拟机实例，负责执行编译后的字节码程序。
//
// VM 实现了基于寄存器的字节码执行引擎，主要特性包括：
//   - 字节码执行：执行编译后的 JPL 程序
//   - 函数调用：支持递归调用和闭包
//   - 全局变量管理：通过索引数组实现 O(1) 访问
//   - 异常处理：支持 try/catch/throw 语句
//   - 调试追踪：可选的指令级调试
//   - 垃圾回收：可选的引用计数 GC
//   - 中断控制：支持超时和外部中断
//
// 生命周期：
//  1. Engine.Compile() 或 NewVMWithProgram() 创建 VM
//  2. vm.Execute() 执行程序
//  3. vm.Close() 释放资源（可选）
//
// 线程安全：
//
//	VM 的导出方法不是线程安全的，同一时间只能有一个 goroutine 执行。
//	如果需要并发执行，应使用多个 VM 实例。
//
// 资源管理：
//
//	建议显式调用 Close() 释放资源，特别是启用了 GC 的 VM。
type VM struct {
	engine  *Engine  // 关联的引擎实例
	program *Program // 编译后的程序
	closed  bool     // VM 是否已关闭
	result  Value    // 执行结果
	err     error    // 执行错误
	// exitCode 脚本退出码（exit/die 函数设置）
	// 当 exit() 或 die() 函数被调用时，VM 将退出码保存到这里
	// 允许 CLI 工具获取 exit/die 设置的退出码并传递给操作系统
	// 初始值为 0，表示正常退出
	// 范围：0-255（遵循 Unix 退出码约定）
	exitCode int

	// 执行状态
	registers         []Value           // 当前寄存器窗口（局部变量/临时值）
	registerBase      int               // 寄存器基址（当前函数的寄存器偏移）
	ip                int               // 指令指针（当前执行位置）
	function          *CompiledFunction // 当前执行的函数
	currentLine       int               // 当前执行的源码行号
	globals           []Value           // 全局变量数组（编译期优化的数组索引访问）
	globalNames       map[string]int    // 全局变量名 -> 数组索引映射
	globalIndexToName []string          // 数组索引 -> 全局变量名（用于反射）
	upvals            []*upvalue        // 当前函数的 upvalue 列表

	// 函数表（名称 -> 编译后函数列表，支持按参数数量重载）
	funcMap map[string][]*CompiledFunction

	// 调用栈
	callStack    []callFrame // 调用栈帧
	maxCallDepth int         // 最大调用深度（防止栈溢出）
	callDepth    int         // 当前调用深度

	// 异常处理栈
	tryHandlers []tryHandler // try 块处理器栈

	// 运行时魔术常量（命令行参数）
	args []Value // 命令行参数数组（ARGV/ARGC）

	// 追踪调试（可选）
	traceConfig *TraceConfig // 调试追踪配置
	debugMode   bool         // REPL 调试模式（打印执行步骤）

	// GC（可选，nil 表示禁用垃圾回收）
	gcField *gc.GC // 垃圾回收器

	// 中断机制（用于 REPL 超时控制）
	interrupt              chan struct{} // 关闭即中断执行的信号通道
	interruptOnce          sync.Once     // 防止并发 close channel panic
	interruptCheckInterval int           // 每 N 条指令检查一次中断
	running                atomic.Bool   // 当前是否正在执行（原子操作）
}

// newVM 创建新的虚拟机实例（内部使用）
func newVM(engine *Engine) *VM {
	return &VM{
		engine:                 engine,
		globals:                make([]Value, 0, 64), // 初始容量 64
		globalNames:            make(map[string]int),
		funcMap:                make(map[string][]*CompiledFunction),
		maxCallDepth:           defaultMaxCallDepth,
		result:                 NewNull(),
		interrupt:              make(chan struct{}),
		interruptCheckInterval: defaultInterruptCheckInteval,
	}
}

// NewVMWithProgram 使用程序创建虚拟机
func NewVMWithProgram(engine *Engine, prog *Program) *VM {
	vm := newVM(engine)
	vm.program = prog
	// 根据编译期确定的全局变量名初始化 globals
	if prog != nil && len(prog.GlobalNames) > 0 {
		vm.globals = make([]Value, len(prog.GlobalNames))
		vm.globalIndexToName = make([]string, len(prog.GlobalNames))
		for i, name := range prog.GlobalNames {
			vm.globalNames[name] = i
			vm.globalIndexToName[i] = name
			vm.globals[i] = NewNull()
		}
	}

	// 将引擎中注册的 Go 函数设置为全局变量
	// 这样编译器在编译时就能识别这些函数名
	if engine != nil && len(engine.functions) > 0 {
		for name, fn := range engine.functions {
			funcVal := &funcValue{name: name, fn: fn}
			vm.SetGlobal(name, funcVal)
		}
	}

	vm.buildFuncMap()
	return vm
}

// NewTestVM 创建一个无程序的 VM（仅用于测试）
func NewTestVM(engine *Engine) *VM {
	return newVM(engine)
}

// newVMWithProgram 使用程序创建虚拟机（内部使用）
func newVMWithProgram(engine *Engine, prog *Program) *VM {
	return NewVMWithProgram(engine, prog)
}

// buildFuncMap 从程序构建函数表（支持重载）
func (vm *VM) buildFuncMap() {
	if vm.program == nil {
		return
	}
	for _, fn := range vm.program.Functions {
		if fn.Name != "<main>" {
			vm.funcMap[fn.Name] = append(vm.funcMap[fn.Name], fn)
		}
	}
}

// ============================================================================
// 执行入口
// ============================================================================

// Execute 执行编译后的字节码程序。
//
// 此方法启动虚拟机执行引擎，按顺序执行字节码指令直到程序结束或发生错误。
// 执行期间会：
//   - 按指令指针顺序执行字节码
//   - 管理函数调用和返回
//   - 处理异常（try/catch/throw）
//   - 检查中断信号（用于超时控制）
//   - 可选：输出调试追踪信息
//
// 线程安全：此方法不是线程安全的，同一时间只能执行一次。
//
// 返回值：
//   - nil: 执行成功完成
//   - ErrVMClosed: VM 已关闭
//   - 其他错误: 执行过程中的各种错误（如运行时错误）
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	vm, err := eng.Compile(`$result = 10 + 20`)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := vm.Execute(); err != nil {
//	    log.Fatalf("执行失败: %v", err)
//	}
//
//	// 获取执行结果
//	result := vm.GetResult()
//	fmt.Printf("Result: %v\n", result.Int()) // 输出: 30
func (vm *VM) Execute() error {
	if vm.closed {
		return ErrVMClosed
	}

	if vm.program == nil || vm.program.Main == nil {
		return fmt.Errorf("no program to execute")
	}

	vm.err = nil
	vm.result = NewNull()
	vm.exitCode = 0 // 重置退出码

	// 初始化函数表
	vm.buildFuncMap()

	// 初始化全局变量数组大小（如果有编译期全局变量）
	// 需要扩容的场景：REPL 中多次编译，新程序可能有更多全局变量
	requiredSize := len(vm.program.GlobalNames)
	if requiredSize > 0 {
		if len(vm.globals) == 0 {
			// 首次初始化
			vm.globals = make([]Value, requiredSize)
			vm.globalIndexToName = make([]string, requiredSize)
			copy(vm.globalIndexToName, vm.program.GlobalNames)
		} else if requiredSize > len(vm.globals) {
			// 需要扩容（REPL 场景）
			oldGlobals := vm.globals
			oldNames := vm.globalIndexToName
			vm.globals = make([]Value, requiredSize)
			vm.globalIndexToName = make([]string, requiredSize)
			copy(vm.globals, oldGlobals)
			copy(vm.globalIndexToName, oldNames)
			// 补充新的全局变量名，并将新位置初始化为 NewNull()
			for i := len(oldNames); i < requiredSize; i++ {
				vm.globalIndexToName[i] = vm.program.GlobalNames[i]
				vm.globals[i] = NewNull() // 初始化新位置，避免 nil 指针
			}
		}
	}

	// 设置 globalNames 映射（用于运行时动态查找）
	for i, name := range vm.program.GlobalNames {
		vm.globalNames[name] = i
	}

	// 准备主函数（程序入口）
	mainFunc := vm.program.Main
	if mainFunc == nil {
		return fmt.Errorf("程序没有主函数")
	}

	// 创建主函数闭包
	closure := &closure{
		// function: mainFunc,
		upvals: make([]*upvalue, len(mainFunc.Upvals)),
	}

	// 分配主函数的寄存器
	vm.registers = make([]Value, mainFunc.Registers)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}
	vm.registerBase = 0
	vm.function = mainFunc
	vm.upvals = closure.upvals

	// 设置指令指针为入口点（从头开始执行）
	vm.ip = 0

	// 标记正在运行
	vm.running.Store(true)
	defer vm.running.Store(false)

	// 执行主循环
	return vm.run()
}

// ExecuteJSON 执行脚本并将结果序列化为 JSON 字符串返回。
//
// 此方法先调用 Execute() 执行脚本，然后将执行结果转换为 JSON 格式字符串。
// 适用于需要将脚本执行结果以 JSON 格式输出的场景。
//
// 注意：结果的 Stringify() 方法会被调用以生成 JSON 字符串。
// 对于复杂类型（数组、对象），会递归序列化其内容。
//
// 线程安全：此方法不是线程安全的，同一时间只能执行一次。
//
// 返回值：
//   - string: 执行结果的 JSON 序列化字符串
//   - nil: 执行成功
//   - ErrVMClosed: VM 已关闭
//   - 其他错误: Execute() 或序列化过程中的错误
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	vm, _ := eng.Compile(`
//	    $user = {
//	        name: "Alice",
//	        age: 30
//	    }
//	    return $user
//	`)
//
//	jsonStr, err := vm.ExecuteJSON()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(jsonStr) // 输出: {"name":"Alice","age":30}
func (vm *VM) ExecuteJSON() (string, error) {
	if vm.closed {
		return "", ErrVMClosed
	}

	err := vm.Execute()
	if err != nil {
		return "", err
	}

	return vm.result.Stringify(), nil
}

// ============================================================================
// 主执行循环
// ============================================================================

// run 主执行循环
func (vm *VM) run() error {
	instrCount := 0
	for vm.ip < len(vm.function.Bytecode) {
		// 中断检查（每 N 条指令）
		if instrCount%vm.interruptCheckInterval == 0 {
			select {
			case <-vm.interrupt:
				return ErrInterrupted
			default:
			}
		}
		instrCount++

		ins := vm.function.Bytecode[vm.ip]
		op := ins.OP()

		// 更新当前源码行号
		if vm.function.SourceLines != nil && vm.ip < len(vm.function.SourceLines) {
			vm.currentLine = vm.function.SourceLines[vm.ip]
		}

		vm.ip++

		// 追踪点（非侵入式，未配置时无开销）
		vm.trace(ins, op)

		switch op {
		// 加载/存储
		case OP_LOAD:
			vm.opLoad(ins)
		case OP_LOADK:
			vm.opLoadK(ins)
		case OP_LOADNULL:
			vm.opLoadNull(ins)
		case OP_LOADBOOL:
			vm.opLoadBool(ins)
		case OP_GETGLOBAL:
			vm.opGetGlobal(ins)
		case OP_SETGLOBAL:
			vm.opSetGlobal(ins)
		case OP_GETVAR:
			vm.opGetVar(ins)
		case OP_SETVAR:
			vm.opSetVar(ins)

		// 算术运算
		case OP_ADD:
			vm.opAdd(ins)
		case OP_SUB:
			vm.opSub(ins)
		case OP_MUL:
			vm.opMul(ins)
		case OP_DIV:
			vm.opDiv(ins)
		case OP_MOD:
			vm.opMod(ins)
		case OP_NEG:
			vm.opNeg(ins)

		// 位运算（新增）
		case OP_BITAND:
			vm.opBitAnd(ins)
		case OP_BITOR:
			vm.opBitOr(ins)
		case OP_BITXOR:
			vm.opBitXor(ins)
		case OP_BITNOT:
			vm.opBitNot(ins)
		case OP_SHL:
			vm.opShl(ins)
		case OP_SHR:
			vm.opShr(ins)

		// 比较运算
		case OP_EQ:
			vm.opEq(ins)
		case OP_NEQ:
			vm.opNeq(ins)
		case OP_LT:
			vm.opLt(ins)
		case OP_GT:
			vm.opGt(ins)
		case OP_LTE:
			vm.opLte(ins)
		case OP_GTE:
			vm.opGte(ins)

		// 字符串连接
		case OP_CONCAT:
			vm.opConcat(ins)

		// 范围
		case OP_RANGE:
			vm.opRange(ins, false)
		case OP_RANGE_INCLUSIVE:
			vm.opRange(ins, true)

		// 逻辑运算
		case OP_AND:
			vm.opAnd(ins)
		case OP_OR:
			vm.opOr(ins)
		case OP_NOT:
			vm.opNot(ins)

		// 数组和对象
		case OP_NEWARRAY:
			vm.opNewArray(ins)
		case OP_NEWOBJECT:
			vm.opNewObject(ins)
		case OP_GETINDEX:
			vm.opGetIndex(ins)
		case OP_SETINDEX:
			vm.opSetIndex(ins)
		case OP_GETMEMBER:
			vm.opGetMember(ins)
		case OP_SETMEMBER:
			vm.opSetMember(ins)

		// 控制流
		case OP_JMP:
			vm.opJmp(ins)
		case OP_JMPIF:
			vm.opJmpIf(ins)
		case OP_JMPIFNOT:
			vm.opJmpIfNot(ins)

		// 函数调用
		case OP_CALL:
			if err := vm.opCall(ins); err != nil {
				if exitErr, ok := err.(*ExitError); ok {
					vm.exitCode = exitErr.Code
					return nil
				}
				return vm.enrichError(err)
			}
		case OP_TAIL_CALL:
			if err := vm.opTailCall(ins); err != nil {
				if exitErr, ok := err.(*ExitError); ok {
					vm.exitCode = exitErr.Code
					return nil
				}
				return vm.enrichError(err)
			}
		case OP_RETURN:
			vm.opReturn(ins)

		// 闭包
		case OP_CLOSURE:
			vm.opClosure(ins)
		case OP_GETUPVAL:
			vm.opGetUpval(ins)
		case OP_SETUPVAL:
			vm.opSetUpval(ins)
		case OP_CLOSE_UPVALS:
			vm.opCloseUpvals(ins)

		// 异常处理
		case OP_TRY_BEGIN:
			vm.opTryBegin(ins)
		case OP_TRY_END:
			vm.opTryEnd()
		case OP_THROW:
			vm.opThrow(ins)

		// 其他
		case OP_TYPEOF:
			vm.opTypeOf(ins)
		case OP_CAST:
			vm.opCast(ins)
		case OP_FORMAT:
			vm.opFormat(ins)
		case OP_NOP:
			// 空操作
		case OP_POP:
			// 弹出栈顶（丢弃值）

		// 模块系统
		case OP_IMPORT:
			if err := vm.opImport(ins); err != nil {
				return err
			}
		case OP_INCLUDE:
			if err := vm.opInclude(ins); err != nil {
				return err
			}

		// 迭代器（foreach 支持）
		case OP_ITERINIT:
			if err := vm.opIterInit(ins); err != nil {
				return err
			}
		case OP_ITERNEXT:
			done, err := vm.opIterNext(ins)
			if err != nil {
				return err
			}
			if !done {
				// 迭代未结束，跳过下一条 JMP（继续执行循环体）
				vm.ip++
			}
			// 如果 done=true（迭代结束），执行下一条 JMP 跳出循环
		case OP_ITEREND:
			// 清理迭代器资源（目前数组和对象不需要额外清理）

		case OP_GETARGV:
			if err := vm.opGetArgv(ins); err != nil {
				return err
			}

		case OP_GETARGC:
			if err := vm.opGetArgc(ins); err != nil {
				return err
			}

		case OP_REGEX_MATCH:
			vm.opRegexMatch(ins)

		case OP_GET_INDIRECT:
			vm.opGetIndirect(ins)

		default:
			return NewRuntimeError(fmt.Sprintf("unknown opcode: %s", op))
		}

		// 检查错误
		if vm.err != nil {
			// ExitError 处理（run() 主循环中的检查点）：
			// 这是 VM 执行过程中的全局错误检查点
			// 如果之前的操作设置了 vm.err 为 ExitError（如 opThrow 或其他路径）
			// 我们在这里识别并处理它：
			// 1. 保存退出码到 vm.exitCode
			// 2. 返回 nil（表示正常终止）
			// 3. 不将 ExitError 当作异常传播
			//
			// 注意：exit/die 实际上不会走这条路径，因为它们会在 opCall 阶段被捕获
			// 这里作为防御性编程，防止其他意外情况
			if exitErr, ok := vm.err.(*ExitError); ok {
				vm.exitCode = exitErr.Code
				return nil
			}
			return vm.enrichError(vm.err)
		}
	}

	return nil
}

// enrichError 为运行时错误附加当前行号信息
func (vm *VM) enrichError(err error) error {
	if re, ok := err.(*RuntimeError); ok {
		if re.Line == 0 && vm.currentLine > 0 {
			re.Line = vm.currentLine
		}
	}
	return err
}

// ============================================================================
// 加载/存储操作
// ============================================================================

func (vm *VM) opLoad(ins Instruction) {
	a := ins.A()
	b := ins.B()
	vm.gcSetRegister(a, vm.registers[b])
}

func (vm *VM) opLoadK(ins Instruction) {
	a := ins.A()
	bx := ins.Bx()
	vm.gcSetRegister(a, vm.function.Constants[bx])
}

func (vm *VM) opLoadNull(ins Instruction) {
	a := ins.A()
	vm.gcSetRegister(a, NewNull())
}

func (vm *VM) opLoadBool(ins Instruction) {
	a := ins.A()
	b := ins.B()
	vm.gcSetRegister(a, NewBool(b != 0))
}

func (vm *VM) opGetGlobal(ins Instruction) {
	a := ins.A()
	bx := ins.Bx()

	// 首先尝试从全局数组获取
	val, ok := vm.GetGlobalByIndex(bx)
	if ok && !val.IsNull() {
		vm.gcSetRegister(a, val)
		return
	}

	// 如果值为 null 或不存在，尝试回退查找
	// 获取变量名用于回退查找
	var name string
	if bx >= 0 && bx < len(vm.globalIndexToName) {
		name = vm.globalIndexToName[bx]
	}

	if name != "" {
		// 检查函数表（支持递归调用）
		if fns, found := vm.funcMap[name]; found && len(fns) > 0 {
			val = NewString(fns[0].Name)
			vm.gcSetRegister(a, val)
			return
		}

		// 检查引擎注册的变量
		if vm.engine != nil {
			if ev, err := vm.engine.Get(name); err == nil {
				vm.gcSetRegister(a, ev)
				return
			}
		}

		// 检查预设常量和 define() 定义的常量
		if vm.engine != nil {
			if cv, found := vm.engine.GetConst(name); found {
				vm.gcSetRegister(a, cv)
				return
			}
		}

		// 检查引擎注册的 Go 函数（返回函数名字符串，供 opCall 查找）
		if vm.engine != nil {
			vm.engine.mu.RLock()
			_, exists := vm.engine.functions[name]
			vm.engine.mu.RUnlock()
			if exists {
				val = NewString(name)
				vm.gcSetRegister(a, val)
				return
			}
		}
	}

	// 所有回退查找都失败，返回 null（保持原有行为）
	vm.gcSetRegister(a, NewNull())
}

func (vm *VM) opSetGlobal(ins Instruction) {
	a := ins.A()
	bx := ins.Bx()
	// bx 现在是全局变量的直接索引
	vm.SetGlobalByIndex(bx, vm.registers[a])
}

func (vm *VM) opGetVar(ins Instruction) {
	a := ins.A()
	b := ins.B()
	name := vm.registers[b].String()

	// 先查找局部变量（通过 VarNames 映射）
	if vm.function.VarNames != nil {
		for i, varName := range vm.function.VarNames {
			if varName == name && i < len(vm.registers) {
				vm.gcSetRegister(a, vm.registers[i])
				return
			}
		}
	}

	// 再查找全局变量
	if val, ok := vm.GetGlobal(name); ok {
		vm.gcSetRegister(a, val)
		return
	}

	// 检查函数表
	if fns, found := vm.funcMap[name]; found && len(fns) > 0 {
		vm.gcSetRegister(a, NewString(fns[0].Name))
		return
	}

	// 检查引擎注册的变量
	if vm.engine != nil {
		if ev, err := vm.engine.Get(name); err == nil {
			vm.gcSetRegister(a, ev)
			return
		}
	}

	// 检查预设常量和 define() 定义的常量
	if vm.engine != nil {
		if cv, found := vm.engine.GetConst(name); found {
			vm.gcSetRegister(a, cv)
			return
		}
	}

	// 变量未定义，返回 null（保持原有行为）
	vm.gcSetRegister(a, NewNull())
}

func (vm *VM) opSetVar(ins Instruction) {
	b := ins.B()
	c := ins.C()
	name := vm.registers[b].String()
	value := vm.registers[c]

	// 先尝试设置局部变量（通过 VarNames 映射）
	if vm.function.VarNames != nil {
		for i, varName := range vm.function.VarNames {
			if varName == name && i < len(vm.registers) {
				vm.gcSetRegister(i, value)
				return
			}
		}
	}

	// 否则设置全局变量
	vm.gcSetGlobal(name, value)
}

func (vm *VM) opGetIndirect(ins Instruction) {
	a := ins.A()
	b := ins.B()
	name := vm.registers[b].String()

	// 先查找局部变量（通过 VarNames 映射）
	if vm.function.VarNames != nil {
		for i, varName := range vm.function.VarNames {
			if varName == name && i < len(vm.registers) {
				vm.gcSetRegister(a, vm.registers[i])
				return
			}
		}
	}

	// 再查找全局变量
	if val, ok := vm.GetGlobal(name); ok {
		vm.gcSetRegister(a, val)
		return
	}

	// 检查函数表
	if fns, found := vm.funcMap[name]; found && len(fns) > 0 {
		vm.gcSetRegister(a, NewString(fns[0].Name))
		return
	}

	// 检查引擎注册的变量
	if vm.engine != nil {
		if ev, err := vm.engine.Get(name); err == nil {
			vm.gcSetRegister(a, ev)
			return
		}
	}

	// 检查预设常量和 define() 定义的常量
	if vm.engine != nil {
		if cv, found := vm.engine.GetConst(name); found {
			vm.gcSetRegister(a, cv)
			return
		}
	}

	// 变量未定义，返回 null
	vm.gcSetRegister(a, NewNull())
}

// ============================================================================
// 算术运算
// ============================================================================

func (vm *VM) opAdd(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	// int-int 快速路径：跳过接口派发
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value+ri.value))
			return
		}
	}
	vm.gcSetRegister(a, ValueAdd(lhs, rhs))
}

func (vm *VM) opSub(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value-ri.value))
			return
		}
	}
	vm.gcSetRegister(a, ValueSub(lhs, rhs))
}

func (vm *VM) opMul(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value*ri.value))
			return
		}
	}
	vm.gcSetRegister(a, ValueMul(lhs, rhs))
}

func (vm *VM) opDiv(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	// int-int 快速路径：保留原始 float 除法语义
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewFloat(float64(li.value)/float64(ri.value)))
			return
		}
	}
	vm.gcSetRegister(a, ValueDiv(lhs, rhs))
}

func (vm *VM) opMod(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	// int-int 快速路径：保留原始取模语义（零返回 NaN）
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			if ri.value == 0 {
				vm.gcSetRegister(a, NewFloat(math.NaN()))
			} else {
				vm.gcSetRegister(a, NewInt(li.value%ri.value))
			}
			return
		}
	}
	vm.gcSetRegister(a, ValueMod(lhs, rhs))
}

func (vm *VM) opNeg(ins Instruction) {
	a := ins.A()
	b := ins.B()
	val := vm.registers[b]
	if vi, ok := val.(*intValue); ok {
		vm.gcSetRegister(a, NewInt(-vi.value))
		return
	}
	vm.gcSetRegister(a, ValueNegate(val))
}

// ============================================================================
// 位运算（新增）
// ============================================================================

func (vm *VM) opBitAnd(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	// int-int 快速路径
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value&ri.value))
			return
		}
	}
	// 转换为整数后位运算
	vm.gcSetRegister(a, NewInt(lhs.Int()&rhs.Int()))
}

func (vm *VM) opBitOr(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value|ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewInt(lhs.Int()|rhs.Int()))
}

func (vm *VM) opBitXor(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value^ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewInt(lhs.Int()^rhs.Int()))
}

func (vm *VM) opBitNot(ins Instruction) {
	a := ins.A()
	b := ins.B()
	val := vm.registers[b]
	if vi, ok := val.(*intValue); ok {
		vm.gcSetRegister(a, NewInt(^vi.value))
		return
	}
	vm.gcSetRegister(a, NewInt(^val.Int()))
}

func (vm *VM) opShl(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value<<uint(ri.value)))
			return
		}
	}
	vm.gcSetRegister(a, NewInt(lhs.Int()<<uint(rhs.Int())))
}

func (vm *VM) opShr(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewInt(li.value>>uint(ri.value)))
			return
		}
	}
	vm.gcSetRegister(a, NewInt(lhs.Int()>>uint(rhs.Int())))
}

// ============================================================================
// 比较运算
// ============================================================================

func (vm *VM) opEq(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	// int-int 快速路径
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value == ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(lhs.Equals(rhs)))
}

func (vm *VM) opNeq(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value != ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(!lhs.Equals(rhs)))
}

// opRegexMatch 执行正则匹配运算符 =~
//
// R[A] = R[B] =~ R[C]
// 左操作数: 字符串
// 右操作数: 正则表达式值
// 返回: bool（子串匹配）
func (vm *VM) opRegexMatch(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]

	re, ok := rhs.(*regexValue)
	if !ok {
		// 右操作数不是正则，尝试作为字符串进行精确匹配
		vm.gcSetRegister(a, NewBool(lhs.String() == rhs.String()))
		return
	}

	// 正则子串匹配
	subject := lhs.String()
	vm.gcSetRegister(a, NewBool(re.Match(subject)))
}

func (vm *VM) opLt(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value < ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(ValueLess(lhs, rhs)))
}

func (vm *VM) opGt(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value > ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(ValueGreater(lhs, rhs)))
}

func (vm *VM) opLte(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value <= ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(ValueLessEqual(lhs, rhs)))
}

func (vm *VM) opGte(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	lhs := vm.registers[b]
	rhs := vm.registers[c]
	if li, lok := lhs.(*intValue); lok {
		if ri, rok := rhs.(*intValue); rok {
			vm.gcSetRegister(a, NewBool(li.value >= ri.value))
			return
		}
	}
	vm.gcSetRegister(a, NewBool(ValueGreaterEqual(lhs, rhs)))
}

// ============================================================================
// 字符串连接
// ============================================================================

func (vm *VM) opConcat(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	vm.gcSetRegister(a, ConcatValues(vm.registers[b], vm.registers[c]))
}

// ============================================================================
// 范围
// ============================================================================

func (vm *VM) opRange(ins Instruction, inclusive bool) {
	a := ins.A()
	b := ins.B()
	c := ins.C()

	startVal := vm.registers[b]
	endVal := vm.registers[c]

	start := startVal.Int()
	end := endVal.Int()

	vm.gcSetRegister(a, NewRange(start, end, inclusive))
}

// ============================================================================
// 逻辑运算
// ============================================================================

func (vm *VM) opAnd(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	if IsTruthy(vm.registers[b]) {
		vm.gcSetRegister(a, vm.registers[c])
	} else {
		vm.gcSetRegister(a, vm.registers[b])
	}
}

func (vm *VM) opOr(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	if IsTruthy(vm.registers[b]) {
		vm.gcSetRegister(a, vm.registers[b])
	} else {
		vm.gcSetRegister(a, vm.registers[c])
	}
}

func (vm *VM) opNot(ins Instruction) {
	a := ins.A()
	b := ins.B()
	vm.gcSetRegister(a, NewBool(!IsTruthy(vm.registers[b])))
}

// ============================================================================
// 数组和对象
// ============================================================================

func (vm *VM) opNewArray(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	arr := make([]Value, c)
	for i := range c {
		arr[i] = vm.registers[b+i]
	}
	vm.gcSetRegister(a, NewArrayGC(arr, vm.gcField))
}

func (vm *VM) opNewObject(ins Instruction) {
	a := ins.A()
	vm.gcSetRegister(a, NewObjectGC(make(map[string]Value), vm.gcField))
}

func (vm *VM) opGetIndex(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	obj := vm.registers[b]
	idx := vm.registers[c]

	switch obj.Type() {
	case TypeArray:
		i := int(idx.Int())
		// 支持负数索引：-1 表示最后一个元素
		if i < 0 {
			i = obj.Len() + i
		}
		if i < 0 || i >= obj.Len() {
			vm.gcSetRegister(a, NewNull())
		} else {
			vm.gcSetRegister(a, obj.Array()[i])
		}
	case TypeObject:
		objMap := obj.Object()
		key := idx.String()
		if val, ok := objMap[key]; ok {
			vm.gcSetRegister(a, val)
		} else {
			vm.gcSetRegister(a, NewNull())
		}
	default:
		vm.gcSetRegister(a, NewNull())
	}
}

func (vm *VM) opSetIndex(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	obj := vm.registers[b]
	idx := vm.registers[c]
	val := vm.registers[a]

	switch obj.Type() {
	case TypeArray:
		i := int(idx.Int())
		// 支持负数索引：-1 表示最后一个元素
		if i < 0 {
			i = obj.Len() + i
		}
		if i >= 0 && i < obj.Len() {
			obj.Array()[i] = val
		}
	case TypeObject:
		objMap := obj.Object()
		key := idx.String()
		objMap[key] = val
	}
}

func (vm *VM) opGetMember(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	obj := vm.registers[b]
	propName := vm.function.Constants[c].String()

	if obj.Type() == TypeObject || obj.Type() == TypeError {
		objMap := obj.Object()
		if val, ok := objMap[propName]; ok {
			vm.gcSetRegister(a, val)
			return
		}
	}
	vm.gcSetRegister(a, NewNull())
}

func (vm *VM) opSetMember(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()
	obj := vm.registers[b]
	propName := vm.function.Constants[c].String()
	val := vm.registers[a]

	if obj.Type() == TypeObject {
		objMap := obj.Object()
		objMap[propName] = val
	}
}

// ============================================================================
// 控制流
// ============================================================================

func (vm *VM) opJmp(ins Instruction) {
	sbx := ins.AsBx()
	vm.ip += sbx
}

func (vm *VM) opJmpIf(ins Instruction) {
	a := ins.A()
	sbx := ins.AsBx()
	if IsTruthy(vm.registers[a]) {
		vm.ip += sbx
	}
}

func (vm *VM) opJmpIfNot(ins Instruction) {
	a := ins.A()
	sbx := ins.AsBx()
	if !IsTruthy(vm.registers[a]) {
		vm.ip += sbx
	}
}

// ============================================================================
// 函数调用
// ============================================================================

// opCall 执行函数调用
func (vm *VM) opCall(ins Instruction) error {
	a := ins.A()
	c := ins.C()

	// 获取函数值
	funcVal := vm.registers[a]

	// 收集参数
	args := make([]Value, c-1)
	for i := 0; i < c-1; i++ {
		args[i] = vm.registers[a+1+i]
	}

	// 检查是否为闭包
	if cl, ok := funcVal.(*closure); ok {
		return vm.callClosure(cl, args, a)
	}

	// 检查是否为 Go 函数（通过 funcValue 类型）
	if funcVal.Type() == TypeFunc {
		// Go 函数调用（直接调用 funcValue 包装的 GoFunction）
		goFn, ok := funcVal.(*funcValue)
		if !ok {
			return NewRuntimeError("invalid function value")
		}
		ctx := &Context{engine: vm.engine, vm: vm}
		result, err := goFn.fn(ctx, args)
		if err != nil {
			// ExitError 处理（Go 函数路径）：
			// exit/die 函数返回 ExitError 时，直接向上传播
			// 不做任何包装，让调用者（opCall 的处理逻辑）识别并处理
			// 这样可以保持原始退出码和消息
			if _, ok := err.(*ExitError); ok {
				return err
			}
			return NewRuntimeError(err.Error())
		}
		vm.gcSetRegister(a, result)
		return nil
	}

	// 查找编译后的 JPL 函数
	var targetFunc *CompiledFunction
	if funcVal.Type() == TypeString {
		funcName := funcVal.String()
		targetFunc = vm.findFunction(funcName, len(args))
	}

	// 如果找不到编译函数，检查是否为引擎注册的 Go 函数
	if targetFunc == nil && vm.engine != nil {
		if funcVal.Type() == TypeString {
			funcName := funcVal.String()
			vm.engine.mu.RLock()
			goFn, exists := vm.engine.functions[funcName]
			vm.engine.mu.RUnlock()
			if exists {
				ctx := &Context{engine: vm.engine, vm: vm}
				result, err := goFn(ctx, args)
				if err != nil {
					// ExitError 处理（引擎注册函数路径）：
					// 与上面 Go 函数路径类似，exit/die 返回 ExitError 时直接传播
					// 通过函数名字符串查找的注册函数也会经过此处
					if _, ok := err.(*ExitError); ok {
						return err
					}
					return NewRuntimeError(err.Error())
				}
				vm.gcSetRegister(a, result)
				return nil
			}
		}
	}

	if targetFunc == nil {
		return NewRuntimeError(fmt.Sprintf("undefined function: %s", funcVal.Stringify()))
	}

	// 栈溢出保护
	if vm.callDepth >= vm.maxCallDepth {
		return ErrStackOverflow
	}

	// 保存当前帧
	frame := callFrame{
		ip:           vm.ip,
		function:     vm.function,
		registers:    vm.registers,
		resultReg:    a,
		registerBase: vm.registerBase,
		upvals:       vm.upvals,
		tryHandlers:  vm.tryHandlers,
	}
	vm.callStack = append(vm.callStack, frame)
	vm.callDepth++

	// 设置新帧
	vm.function = targetFunc
	vm.ip = 0
	vm.registerBase = 0
	vm.tryHandlers = nil // 清空 try 处理器栈，新函数有自己的 try 块

	// 追踪函数调用
	vm.traceCall(targetFunc.Name, len(args))

	// 创建新寄存器窗口
	regCount := max(targetFunc.Registers, targetFunc.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}

	// 复制参数到新寄存器窗口
	for i := 0; i < len(args) && i < targetFunc.Params; i++ {
		vm.registers[i] = args[i]
	}

	return nil
}

// opTailCall 尾调用优化 — 复用当前栈帧或执行调用并返回
// 对自递归调用进行栈帧复用优化；对其他调用执行普通调用并立即返回结果
func (vm *VM) opTailCall(ins Instruction) error {
	a := ins.A()
	c := ins.C()

	funcVal := vm.registers[a]

	// 收集参数
	args := make([]Value, c-1)
	for i := 0; i < c-1; i++ {
		args[i] = vm.registers[a+1+i]
	}

	// 检查是否为自递归调用
	isSelfRecursive := false

	// 情况 1：通过函数名字符串调用
	if funcVal.Type() == TypeString {
		funcName := funcVal.String()
		if funcName == vm.function.Name {
			isSelfRecursive = true
		}
	}

	// 情况 2：通过闭包调用（最常见的自递归形式）
	if cl, ok := funcVal.(*closure); ok {
		if cl.function == vm.function {
			isSelfRecursive = true
		}
	}

	if isSelfRecursive {
		return vm.tailCallSelfFunction(args)
	}

	// 非自递归：执行调用并立即返回结果（模拟 OP_CALL + OP_RETURN 的语义）
	return vm.tailCallOtherFunction(funcVal, args, a)
}

// tailCallSelfFunction 执行自递归尾调用优化
// 复用当前寄存器窗口和调用帧，仅更新参数并跳转到函数开头
func (vm *VM) tailCallSelfFunction(args []Value) error {
	fn := vm.function

	// 原地更新参数寄存器
	paramCount := max(len(args), fn.Params)
	for i := 0; i < len(args) && i < len(vm.registers); i++ {
		vm.registers[i] = args[i]
	}

	// 清零剩余寄存器（避免残留临时值影响下一轮执行）
	for i := paramCount; i < len(vm.registers); i++ {
		vm.registers[i] = NewNull()
	}

	// 跳转到函数开头重新执行
	vm.ip = 0

	return nil
}

// tailCallOtherFunction 非自递归尾调用：执行调用并立即返回结果
// 用于 return otherFn(args) 场景，语义等价于 OP_CALL + OP_RETURN
func (vm *VM) tailCallOtherFunction(funcVal Value, args []Value, resultReg int) error {
	// 闭包调用
	if cl, ok := funcVal.(*closure); ok {
		return vm.tailCallClosure(cl, args, resultReg)
	}

	// Go 函数调用
	if funcVal.Type() == TypeFunc {
		goFn, ok := funcVal.(*funcValue)
		if !ok {
			return NewRuntimeError("invalid function value")
		}
		ctx := &Context{engine: vm.engine, vm: vm}
		result, err := goFn.fn(ctx, args)
		if err != nil {
			if _, ok := err.(*ExitError); ok {
				return err
			}
			return NewRuntimeError(err.Error())
		}
		vm.gcSetRegister(resultReg, result)
		// 立即返回
		vm.returnFromCurrentFunction(resultReg)
		return nil
	}

	// JPL 函数调用（通过名字）
	if funcVal.Type() == TypeString {
		funcName := funcVal.String()
		targetFunc := vm.findFunction(funcName, len(args))

		if targetFunc == nil && vm.engine != nil {
			vm.engine.mu.RLock()
			goFn, exists := vm.engine.functions[funcName]
			vm.engine.mu.RUnlock()
			if exists {
				ctx := &Context{engine: vm.engine, vm: vm}
				result, err := goFn(ctx, args)
				if err != nil {
					if _, ok := err.(*ExitError); ok {
						return err
					}
					return NewRuntimeError(err.Error())
				}
				vm.gcSetRegister(resultReg, result)
				vm.returnFromCurrentFunction(resultReg)
				return nil
			}
		}

		if targetFunc == nil {
			return NewRuntimeError(fmt.Sprintf("undefined function: %s", funcVal.Stringify()))
		}

		return vm.tailCallJPLFunction(targetFunc, args, resultReg)
	}

	return NewRuntimeError(fmt.Sprintf("cannot call value of type %s", funcVal.Type()))
}

// tailCallClosure 尾调用闭包：创建新帧执行调用，返回时立即向上传递结果
func (vm *VM) tailCallClosure(cl *closure, args []Value, resultReg int) error {
	if vm.callDepth >= vm.maxCallDepth {
		return ErrStackOverflow
	}

	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedUpvals := vm.upvals
	savedTryHandlers := vm.tryHandlers

	vm.function = cl.function
	vm.ip = 0
	vm.registerBase = 0
	vm.upvals = cl.upvals
	vm.tryHandlers = nil

	frame := callFrame{
		ip:           savedIP,
		function:     savedFunc,
		registers:    savedRegs,
		resultReg:    resultReg,
		registerBase: savedRegBase,
		upvals:       savedUpvals,
		tryHandlers:  savedTryHandlers,
	}
	vm.callStack = append(vm.callStack, frame)
	vm.callDepth++

	vm.traceCall(cl.function.Name, len(args))

	regCount := max(cl.function.Registers, cl.function.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}
	for i := 0; i < len(args) && i < cl.function.Params; i++ {
		vm.registers[i] = args[i]
	}
	vm.upvals = cl.upvals

	return nil
}

// tailCallJPLFunction 尾调用 JPL 函数：创建新帧执行调用，返回时立即向上传递结果
func (vm *VM) tailCallJPLFunction(targetFunc *CompiledFunction, args []Value, resultReg int) error {
	if vm.callDepth >= vm.maxCallDepth {
		return ErrStackOverflow
	}

	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedUpvals := vm.upvals
	savedTryHandlers := vm.tryHandlers

	vm.function = targetFunc
	vm.ip = 0
	vm.registerBase = 0
	vm.tryHandlers = nil

	frame := callFrame{
		ip:           savedIP,
		function:     savedFunc,
		registers:    savedRegs,
		resultReg:    resultReg,
		registerBase: savedRegBase,
		upvals:       savedUpvals,
		tryHandlers:  savedTryHandlers,
	}
	vm.callStack = append(vm.callStack, frame)
	vm.callDepth++

	vm.traceCall(targetFunc.Name, len(args))

	regCount := max(targetFunc.Registers, targetFunc.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}
	for i := 0; i < len(args) && i < targetFunc.Params; i++ {
		vm.registers[i] = args[i]
	}

	return nil
}

// returnFromCurrentFunction 将寄存器中的值作为当前函数的返回值，并向调用者返回
func (vm *VM) returnFromCurrentFunction(resultReg int) {
	returnVal := vm.registers[resultReg]
	vm.traceReturn(returnVal)

	if len(vm.callStack) > 0 {
		frame := vm.callStack[len(vm.callStack)-1]
		vm.callStack = vm.callStack[:len(vm.callStack)-1]
		vm.callDepth--

		vm.ip = frame.ip
		vm.function = frame.function
		vm.registers = frame.registers
		vm.registerBase = frame.registerBase
		vm.upvals = frame.upvals
		vm.tryHandlers = frame.tryHandlers

		vm.gcSetRegister(frame.resultReg, returnVal)
	} else {
		vm.result = returnVal
		vm.ip = len(vm.function.Bytecode)
	}
}

// opReturn 从函数返回
func (vm *VM) opReturn(ins Instruction) {
	a := ins.A()
	returnVal := vm.registers[a]

	// 追踪函数返回
	vm.traceReturn(returnVal)

	// 恢复调用者帧
	if len(vm.callStack) > 0 {
		frame := vm.callStack[len(vm.callStack)-1]
		vm.callStack = vm.callStack[:len(vm.callStack)-1]
		vm.callDepth--

		// 恢复调用者状态
		vm.ip = frame.ip
		vm.function = frame.function
		vm.registers = frame.registers
		vm.registerBase = frame.registerBase
		vm.upvals = frame.upvals
		vm.tryHandlers = frame.tryHandlers

		// 存储返回值
		vm.gcSetRegister(frame.resultReg, returnVal)

		// 如果恢复后的 IP 已在函数末尾（尾调用优化场景），继续向上传递返回值
		for vm.ip >= len(vm.function.Bytecode) && len(vm.callStack) > 0 {
			parentFrame := vm.callStack[len(vm.callStack)-1]
			vm.callStack = vm.callStack[:len(vm.callStack)-1]
			vm.callDepth--

			vm.ip = parentFrame.ip
			vm.function = parentFrame.function
			vm.registers = parentFrame.registers
			vm.registerBase = parentFrame.registerBase
			vm.upvals = parentFrame.upvals
			vm.tryHandlers = parentFrame.tryHandlers

			vm.gcSetRegister(parentFrame.resultReg, returnVal)
		}
	} else {
		// 主函数返回
		vm.result = returnVal
		vm.ip = len(vm.function.Bytecode) // 结束执行
	}
}

// callClosure 调用闭包
func (vm *VM) callClosure(cl *closure, args []Value, resultReg int) error {
	// 栈溢出保护
	if vm.callDepth >= vm.maxCallDepth {
		return ErrStackOverflow
	}

	// 保存当前帧的状态
	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedUpvals := vm.upvals
	savedTryHandlers := vm.tryHandlers

	// 设置新帧
	vm.function = cl.function
	vm.ip = 0
	vm.registerBase = 0
	vm.upvals = cl.upvals
	vm.tryHandlers = nil // 清空 try 处理器栈，新函数有自己的 try 块

	// 将旧帧保存到栈
	frame := callFrame{
		ip:           savedIP,
		function:     savedFunc,
		registers:    savedRegs,
		resultReg:    resultReg,
		registerBase: savedRegBase,
		upvals:       savedUpvals,
		tryHandlers:  savedTryHandlers,
	}
	vm.callStack = append(vm.callStack, frame)
	vm.callDepth++

	// 追踪函数调用
	vm.traceCall(cl.function.Name, len(args))

	// 创建新寄存器窗口
	regCount := max(cl.function.Registers, cl.function.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}

	// 复制参数到新寄存器窗口
	for i := 0; i < len(args) && i < cl.function.Params; i++ {
		vm.registers[i] = args[i]
	}

	// 设置 upvalue 列表
	vm.upvals = cl.upvals

	return nil
}

// findFunction 查找编译后的函数（支持重载）
// argCount < 0 表示不按参数数量匹配，返回第一个
func (vm *VM) findFunction(name string, argCount int) *CompiledFunction {
	// 先从函数表查找
	if fns, ok := vm.funcMap[name]; ok && len(fns) > 0 {
		if argCount < 0 {
			return fns[0]
		}
		// 精确匹配参数数量
		for _, fn := range fns {
			if fn.Params == argCount {
				return fn
			}
		}
		// 无精确匹配，返回参数最多的（多余参数被截断，缺少参数补 null）
		best := fns[0]
		for _, fn := range fns[1:] {
			if fn.Params > best.Params {
				best = fn
			}
		}
		return best
	}
	// 从程序函数列表查找
	if vm.program != nil {
		for _, fn := range vm.program.Functions {
			if fn.Name == name {
				return fn
			}
		}
	}
	return nil
}

// ============================================================================
// 闭包操作
// ============================================================================

func (vm *VM) opClosure(ins Instruction) {
	a := ins.A()
	bx := ins.Bx()

	// 从常量池获取函数索引
	fnConst := vm.function.Constants[bx]
	fnIdx := int(fnConst.Int())

	// 从程序函数列表获取函数
	var fn *CompiledFunction
	if vm.program != nil && fnIdx >= 0 && fnIdx < len(vm.program.Functions) {
		fn = vm.program.Functions[fnIdx]
	}
	if fn == nil {
		vm.err = NewRuntimeError(fmt.Sprintf("undefined function at index %d", fnIdx))
		return
	}

	// 获取 upvalue 数量
	numUpvals := fn.NumUpvals

	// 创建闭包
	cl := &closure{
		function: fn,
		upvals:   make([]*upvalue, numUpvals),
	}

	// 捕获 upvalue
	for i := range numUpvals {
		nextIns := vm.function.Bytecode[vm.ip]
		vm.ip++
		isLocal := nextIns.A() == 1
		idx := nextIns.B()

		if isLocal {
			// 捕获局部变量（指向寄存器）
			cl.upvals[i] = &upvalue{
				location: &vm.registers[idx],
			}
		} else {
			// 捕获外层 upvalue
			if idx < len(vm.upvals) {
				cl.upvals[i] = vm.upvals[idx]
			} else {
				cl.upvals[i] = &upvalue{isClosed: true, closed: NewNull()}
			}
		}
	}

	// 设置 GC
	if vm.gcField != nil {
		cl.SetupGC(vm.gcField)
	}

	vm.gcSetRegister(a, cl)
}

func (vm *VM) opGetUpval(ins Instruction) {
	a := ins.A()
	b := ins.B()
	uv := vm.upvals[b]
	if uv.isClosed {
		vm.gcSetRegister(a, uv.closed)
	} else {
		vm.gcSetRegister(a, *uv.location)
	}
}

func (vm *VM) opSetUpval(ins Instruction) {
	a := ins.A()
	b := ins.B()
	uv := vm.upvals[a]
	if uv.isClosed {
		uv.closed = vm.registers[b]
	} else {
		*uv.location = vm.registers[b]
	}
}

func (vm *VM) opCloseUpvals(_ Instruction) {
	// 关闭所有 upvalue，将值复制到 closed 字段
	for _, uv := range vm.upvals {
		if !uv.isClosed {
			uv.closed = *uv.location
			uv.isClosed = true
		}
	}
}

// ============================================================================
// 异常处理
// ============================================================================

func (vm *VM) opTryBegin(ins Instruction) {
	a := ins.A() // catch 变量的寄存器索引
	sbx := ins.AsBx()
	catchPC := vm.ip + sbx

	vm.tryHandlers = append(vm.tryHandlers, tryHandler{
		catchPC:     catchPC,
		catchVarReg: a,
	})
}

func (vm *VM) opTryEnd() {
	if len(vm.tryHandlers) > 0 {
		vm.tryHandlers = vm.tryHandlers[:len(vm.tryHandlers)-1]
	}
}

func (vm *VM) opThrow(ins Instruction) {
	a := ins.A()
	thrownValue := vm.registers[a]

	// 查找 tryHandler，支持跨函数的异常传播
	for len(vm.tryHandlers) == 0 && len(vm.callStack) > 0 {
		// 当前函数没有 tryHandler，返回到调用者函数
		frame := vm.callStack[len(vm.callStack)-1]
		vm.callStack = vm.callStack[:len(vm.callStack)-1]
		vm.callDepth--

		// 恢复调用者状态
		vm.ip = frame.ip
		vm.function = frame.function
		vm.registers = frame.registers
		vm.registerBase = frame.registerBase
		vm.upvals = frame.upvals
		vm.tryHandlers = frame.tryHandlers
	}

	if len(vm.tryHandlers) == 0 {
		// 没有 try 块，返回运行时错误
		vm.err = NewRuntimeError(thrownValue.String())
		return
	}

	// 弹出最近的 tryHandler
	handler := vm.tryHandlers[len(vm.tryHandlers)-1]
	vm.tryHandlers = vm.tryHandlers[:len(vm.tryHandlers)-1]

	// 将异常值存入 catch 变量寄存器
	if handler.catchVarReg < len(vm.registers) {
		vm.gcSetRegister(handler.catchVarReg, thrownValue)
	}

	// 跳转到 catch 块
	vm.ip = handler.catchPC
}

// ============================================================================
// 其他操作
// ============================================================================

func (vm *VM) opTypeOf(ins Instruction) {
	a := ins.A()
	b := ins.B()
	vm.gcSetRegister(a, NewString(vm.registers[b].Type().String()))
}

// opCast 执行类型转换指令 OP_CAST
// 根据指令中的转换类型，将源寄存器的值转换为目标类型，并存储到目标寄存器
//
// 指令格式: OP_CAST A B C
//   - A: 目标寄存器索引（存储转换结果）
//   - B: 源寄存器索引（被转换的值）
//   - C: 转换类型
//   - 0 = int: 转换为整数（字符串解析为十进制、浮点截断小数、布尔转 0/1、null 转 0）
//   - 1 = float: 转换为浮点数（字符串解析为浮点、整数转浮点、布尔转 0.0/1.0、null 转 0.0）
//   - 2 = string: 转换为字符串（调用 String() 方法）
//   - 3 = bool: 转换为布尔值（使用 IsTruthy 语义：null/0/空字符串/空数组 为 false）
//
// 参数：
//   - ins: 指令，包含寄存器索引和转换类型
//
// 示例转换：
//
//	"42" (string) → int → 42 (int)
//	3.7 (float) → int → 3 (int，截断小数)
//	42 (int) → string → "42" (string)
//	"hello" (string) → bool → true (bool，非空字符串)
func (vm *VM) opCast(ins Instruction) {
	a := ins.A()
	b := ins.B()
	castType := ins.C()

	val := vm.registers[b]

	switch castType {
	case 0: // int
		vm.gcSetRegister(a, vm.castToInt(val))
	case 1: // float
		vm.gcSetRegister(a, vm.castToFloat(val))
	case 2: // string
		vm.gcSetRegister(a, NewString(val.String()))
	case 3: // bool
		vm.gcSetRegister(a, NewBool(IsTruthy(val)))
	default:
		vm.gcSetRegister(a, NewInt(0))
	}
}

// opFormat 执行 OP_FORMAT 指令 — 字符串格式化（用于插值格式化）
// R[A] = sprintf(R[B], R[C])
// B 寄存器包含格式字符串（如 "%.2f"），C 寄存器包含要格式化的值
func (vm *VM) opFormat(ins Instruction) {
	a := ins.A()
	b := ins.B()
	c := ins.C()

	fmtStr := vm.registers[b].String()
	val := vm.registers[c]

	// 将值转为 Go 原生类型
	var goVal any
	switch val.Type() {
	case TypeInt:
		goVal = val.Int()
	case TypeFloat:
		goVal = val.Float()
	case TypeString:
		goVal = val.String()
	case TypeBool:
		goVal = val.Bool()
	case TypeBigInt:
		goVal = val.String() // BigInt 用字符串表示
	case TypeBigDecimal:
		goVal = val.String() // BigDecimal 用字符串表示
	case TypeNull:
		goVal = nil
	default:
		goVal = val.Stringify()
	}

	result := fmt.Sprintf(fmtStr, goVal)
	vm.gcSetRegister(a, NewString(result))
}

// castToInt 将任意类型的值转换为整数
// 实现了 JPL 的 int() 类型转换语义
//
// 转换规则：
//   - int: 直接返回原值
//   - float: 截断小数部分（向零取整），如 3.7 → 3, -3.7 → -3
//   - string: 按十进制解析，解析失败返回 0，空字符串返回 0
//   - "42" → 42, "abc" → 0, "" → 0
//   - bool: true → 1, false → 0
//   - null: 返回 0
//   - 其他类型: 返回 0
//
// 注意：字符串解析只支持十进制，不支持十六进制、八进制前缀
//
//	如需多进制支持，使用 intval() 函数
//
// 参数：
//   - v: 要转换的值
//
// 返回值：
//   - 转换后的整数值（int64 包装为 Value）
func (vm *VM) castToInt(v Value) Value {
	switch v.Type() {
	case TypeInt:
		return v
	case TypeFloat:
		return NewInt(int64(v.Float()))
	case TypeString:
		s := v.String()
		if s == "" {
			return NewInt(0)
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return NewInt(n)
		}
		return NewInt(0)
	case TypeBool:
		if v.Bool() {
			return NewInt(1)
		}
		return NewInt(0)
	case TypeNull:
		return NewInt(0)
	default:
		return NewInt(0)
	}
}

// castToFloat 将任意类型的值转换为浮点数
// 实现了 JPL 的 float() 类型转换语义
//
// 转换规则：
//   - float: 直接返回原值
//   - int: 转换为浮点数，如 42 → 42.0
//   - string: 解析为浮点数，解析失败返回 0.0，空字符串返回 0.0
//   - "3.14" → 3.14, "1.5e3" → 1500.0, "abc" → 0.0, "" → 0.0
//   - bool: true → 1.0, false → 0.0
//   - null: 返回 0.0
//   - 其他类型: 返回 0.0
//
// 注意：字符串支持科学计数法（如 "1.5e3"）和特殊值（如 "NaN", "Inf"）
//
// 参数：
//   - v: 要转换的值
//
// 返回值：
//   - 转换后的浮点数值（float64 包装为 Value）
func (vm *VM) castToFloat(v Value) Value {
	switch v.Type() {
	case TypeFloat:
		return v
	case TypeInt:
		return NewFloat(float64(v.Int()))
	case TypeString:
		s := v.String()
		if s == "" {
			return NewFloat(0.0)
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return NewFloat(f)
		}
		return NewFloat(0.0)
	case TypeBool:
		if v.Bool() {
			return NewFloat(1.0)
		}
		return NewFloat(0.0)
	case TypeNull:
		return NewFloat(0.0)
	default:
		return NewFloat(0.0)
	}
}

// ============================================================================
// 运行时魔术常量（命令行参数）
// ============================================================================

// SetArgs 设置命令行参数（用于 ARGV/ARGC 魔术常量）
func (vm *VM) SetArgs(args []Value) {
	vm.args = args
}

// opGetArgv 获取命令行参数数组（运行时魔术常量 ARGV）
// OP_GETARGV A B C - A=目标寄存器
func (vm *VM) opGetArgv(ins Instruction) error {
	a := ins.A()
	vm.gcSetRegister(a, NewArray(vm.args))
	return nil
}

// opGetArgc 获取命令行参数数量（运行时魔术常量 ARGC）
// OP_GETARGC A B C - A=目标寄存器
func (vm *VM) opGetArgc(ins Instruction) error {
	a := ins.A()
	vm.gcSetRegister(a, NewInt(int64(len(vm.args))))
	return nil
}

// opIterInit 初始化迭代器
// OP_ITERINIT A B C - A=迭代器寄存器, B=被遍历对象寄存器
// 在迭代器寄存器中创建一个数组 [对象, 当前索引]，用于跟踪迭代状态
func (vm *VM) opIterInit(ins Instruction) error {
	iterReg := ins.A()
	objReg := ins.B()
	obj := vm.registers[objReg]

	// 检查是否为可遍历类型
	switch obj.Type() {
	case TypeArray, TypeRange, TypeObject:
		// 支持数组、对象和范围
		// 创建迭代器状态：使用数组存储 [对象, 当前索引=-1]
		// 索引从 -1 开始，因为 ITERNEXT 会先递增再取值
		state := NewArray([]Value{obj, NewInt(-1)})
		vm.gcSetRegister(iterReg, state)
		return nil
	default:
		return NewRuntimeError(fmt.Sprintf("cannot iterate over %s", obj.Type()))
	}
}

// opIterNext 迭代器获取下一个元素
// OP_ITERNEXT A B C - A=迭代器寄存器, B=键寄存器(255表示不需要), C=值寄存器
// 返回 (done bool, err error) - done=true 表示迭代结束
func (vm *VM) opIterNext(ins Instruction) (bool, error) {
	iterReg := ins.A()
	keyReg := ins.B()
	valReg := ins.C()

	// 获取迭代器状态
	state := vm.registers[iterReg]
	if state.Type() != TypeArray {
		return false, NewRuntimeError("invalid iterator state")
	}

	stateArr := state.Array()
	if len(stateArr) < 2 {
		return false, NewRuntimeError("corrupted iterator state")
	}

	obj := stateArr[0]
	idxVal := stateArr[1]
	idx := int(idxVal.Int()) + 1 // 递增索引

	switch obj.Type() {
	case TypeArray:
		arr := obj.Array()
		if idx >= len(arr) {
			// 迭代结束
			return true, nil
		}
		// 设置键（索引）
		if keyReg != 255 {
			vm.gcSetRegister(keyReg, NewInt(int64(idx)))
		}
		// 设置值
		vm.gcSetRegister(valReg, arr[idx])
		// 更新索引
		stateArr[1] = NewInt(int64(idx))
		return false, nil

	case TypeObject:
		objData := obj.Object()
		keys := make([]string, 0, len(objData))
		for k := range objData {
			keys = append(keys, k)
		}
		// 按键排序以保证遍历顺序的一致性
		sort.Strings(keys)

		if idx >= len(keys) {
			// 迭代结束
			return true, nil
		}
		key := keys[idx]
		// 设置键
		if keyReg != 255 {
			vm.gcSetRegister(keyReg, NewString(key))
		}
		// 设置值
		vm.gcSetRegister(valReg, objData[key])
		// 更新索引
		stateArr[1] = NewInt(int64(idx))
		return false, nil

	case TypeRange:
		rangeVal := obj.(*rangeValue)
		start := rangeVal.start
		end := rangeVal.end
		inclusive := rangeVal.inclusive

		current := start + int64(idx)
		maxEnd := end
		if inclusive {
			maxEnd = end
		} else {
			maxEnd = end - 1
		}

		if current > maxEnd {
			return true, nil
		}

		if keyReg != 255 {
			vm.gcSetRegister(keyReg, NewInt(int64(idx)))
		}
		vm.gcSetRegister(valReg, NewInt(current))
		stateArr[1] = NewInt(int64(idx))
		return false, nil

	default:
		return false, NewRuntimeError(fmt.Sprintf("cannot iterate over %s", obj.Type().String()))
	}
}

// ============================================================================
// 公共方法
// ============================================================================

// GetResult 获取脚本执行的结果值。
//
// 此方法返回 Execute() 执行完成后的结果值。
// 如果脚本没有显式返回值，结果为 null。
//
// 返回值：
//   - Value: 执行结果，可能是 null、bool、int、float、string、array、object 等类型
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$x = 10 + 20`)
//	vm.Execute()
//	result := vm.GetResult()
//	fmt.Printf("Result type: %v, value: %v\n", result.Type(), result.Int())
func (vm *VM) GetResult() Value {
	return vm.result
}

// GetExitCode 获取脚本执行的退出码。
//
// **内部实现说明**：
// 当 exit() 或 die() 函数被调用时，它们创建 ExitError 并返回
// VM 在多个检查点识别 ExitError：
//   - opCall 指令执行时
//   - opTailCall 指令执行时
//   - run() 主循环的错误检查点
//
// 识别到 ExitError 后，VM 会：
//  1. 将 ExitError.Code 保存到 vm.exitCode
//  2. 返回 nil（表示正常终止，不是错误）
//
// 此方法允许 CLI 工具在执行完成后获取退出码
//
// **使用场景**：
//   - CLI 工具需要在脚本执行后获取 exit/die 设置的退出码
//   - 将退出码传递给操作系统（os.Exit）
//   - 测试时验证 exit/die 的行为
//
// **注意事项**：
//   - 如果脚本没有调用 exit/die，返回 0（默认值）
//   - 如果脚本因 RuntimeError 终止，也返回 0（因为 exitCode 未被设置）
//   - 调用此方法前应先调用 Execute() 并检查错误
//
// 返回值：
//   - int: 退出码（0-255）
//
// 使用示例：
//
//	vm, _ := eng.Compile(`exit(5)`)
//	vm.Execute()  // 返回 nil（不是错误）
//	code := vm.GetExitCode()  // code = 5
//	// 在 CLI 中，可以传递退出码给操作系统
//	// os.Exit(vm.GetExitCode())
func (vm *VM) GetExitCode() int {
	return vm.exitCode
}

// SetResult 设置脚本的执行结果值。
//
// 此方法用于手动设置执行结果，通常在以下场景使用：
//   - 在 Go 代码中调用脚本函数后设置返回值
//   - 测试时模拟执行结果
//   - 自定义执行流程
//
// 参数：
//   - value: 要设置的结果值
func (vm *VM) SetResult(value Value) {
	vm.result = value
}

// Engine 返回 VM 关联的引擎实例。
//
// 每个 VM 实例都与创建它的 Engine 实例关联。
// 通过此方法可以访问引擎的变量、函数和模块系统。
//
// 返回值：
//   - *Engine: 关联的引擎实例
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$x = 42`)
//	vm.Execute()
//
//	// 通过 VM 获取引擎
//	engine := vm.Engine()
//	value, _ := engine.Get("x")
//	fmt.Printf("x = %v\n", value.Int()) // 输出: 42
func (vm *VM) Engine() *Engine {
	return vm.engine
}

// CurrentFunction 返回当前正在执行的编译函数。
//
// 此方法主要用于内置函数获取调用者的函数信息，
// 支持 func_num_args()、func_get_arg() 等反射函数的实现。
//
// 返回值：
//   - *CompiledFunction: 当前函数，如果不在函数上下文中则返回 nil
//
// 使用示例：
//
//	// 在内置函数中获取调用者信息
//	func myBuiltin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    fn := ctx.VM().CurrentFunction()
//	    if fn != nil {
//	        fmt.Printf("调用者: %s, 参数数: %d\n", fn.Name, fn.Params)
//	    }
//	    return engine.NewNull(), nil
//	}
func (vm *VM) CurrentFunction() *CompiledFunction {
	return vm.function
}

// CurrentRegisters 返回当前函数的寄存器窗口。
//
// 此方法主要用于内置函数获取调用者的参数值，
// 支持 func_get_arg()、func_get_args() 等反射函数的实现。
//
// 返回值：
//   - []Value: 当前寄存器窗口，如果不在函数上下文中则返回 nil
//
// 使用示例：
//
//	// 在内置函数中获取调用者的参数
//	func myBuiltin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    regs := ctx.VM().CurrentRegisters()
//	    if regs != nil && len(regs) > 0 {
//	        return regs[0], nil // 返回第一个参数
//	    }
//	    return engine.NewNull(), nil
//	}
func (vm *VM) CurrentRegisters() []Value {
	return vm.registers
}

// SetMaxCallDepth 设置函数调用的最大深度限制。
//
// 此限制防止因无限递归导致的栈溢出崩溃。
// 默认值为 1000 层调用。
//
// 参数：
//   - depth: 最大调用深度，必须大于 0
//
// 使用示例：
//
//	vm, _ := eng.Compile(`function fib(n) { return n <= 1 ? n : fib(n-1) + fib(n-2) }`)
//	vm.SetMaxCallDepth(5000) // 允许更深的递归
//	vm.Execute()
func (vm *VM) SetMaxCallDepth(depth int) {
	if depth > 0 {
		vm.maxCallDepth = depth
	}
}

// GetMaxCallDepth 获取当前的调用深度限制。
//
// 返回值：
//   - int: 当前设置的最大调用深度
func (vm *VM) GetMaxCallDepth() int {
	return vm.maxCallDepth
}

// GetCallDepth 获取当前调用栈的深度。
//
// 此方法返回从主函数开始的当前调用深度，用于监控递归深度
// 或调试复杂的调用链。
//
// 返回值：
//   - int: 当前调用深度（0 表示主函数层）
//
// 使用示例：
//
//	vm.Execute()
//	fmt.Printf("调用深度: %d\n", vm.GetCallDepth())
func (vm *VM) GetCallDepth() int {
	return vm.callDepth
}

// AllocateGlobalIndex 分配全局变量的数组索引。
//
// 此方法用于编译期和运行时动态分配全局变量索引。
// 如果变量已存在，返回现有索引；否则分配新索引。
//
// 性能优化：通过数组索引替代 map 查找，实现 O(1) 的全局变量访问。
// 编译器在编译期调用此方法为每个全局变量分配固定索引。
//
// 参数：
//   - name: 全局变量名（通常以 $ 开头）
//
// 返回值：
//   - int: 分配的数组索引（>= 0）
//
// 使用示例（编译器内部）：
//
//	// 编译期：为全局变量分配索引
//	idx := vm.AllocateGlobalIndex("$counter")
//	// 生成字节码使用索引而非变量名
//	// GETGLOBAL R0, idx  // 而不是 GETGLOBAL R0, "$counter"
func (vm *VM) AllocateGlobalIndex(name string) int {
	if idx, ok := vm.globalNames[name]; ok {
		return idx
	}
	idx := len(vm.globals)
	vm.globalNames[name] = idx
	vm.globals = append(vm.globals, NewNull())
	return idx
}

// SetGlobal 设置全局变量的值。
//
// 此方法用于在 Go 代码中直接操作脚本的全局变量。
// 如果变量不存在，会自动分配索引并创建。
// 如果启用了 GC，会自动管理引用计数。
//
// 参数：
//   - name: 变量名（会自动添加 $ 前缀如果未指定）
//   - value: 变量值
//
// 使用示例：
//
//	vm, _ := eng.Compile(`print($message)`)
//	vm.SetGlobal("message", engine.NewString("Hello from Go!"))
//	vm.Execute()
func (vm *VM) SetGlobal(name string, value Value) {
	idx := vm.AllocateGlobalIndex(name)
	old := vm.globals[idx]
	vm.globals[idx] = value
	// 管理引用计数
	vm.gcIncRef(value)
	vm.gcDecRef(old)
}

// GetGlobal 获取全局变量的值。
//
// 此方法通过变量名获取全局变量的当前值。
// 如果变量不存在，返回 false。
//
// 参数：
//   - name: 变量名（区分大小写）
//
// 返回值：
//   - Value: 变量的值，如果不存在返回 null
//   - bool: 如果变量存在返回 true，否则返回 false
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$counter = 100`)
//	vm.Execute()
//
//	if value, ok := vm.GetGlobal("counter"); ok {
//	    fmt.Printf("Counter: %v\n", value.Int()) // 输出: 100
//	}
func (vm *VM) GetGlobal(name string) (Value, bool) {
	idx, ok := vm.globalNames[name]
	if !ok {
		return NewNull(), false
	}
	return vm.globals[idx], true
}

// GetGlobalByIndex 通过索引获取全局变量（优化路径）
// GetGlobalByIndex 通过数组索引获取全局变量值（优化路径）。
//
// 此方法使用数组索引直接访问全局变量，避免 map 查找开销。
// 适用于编译器已知的索引（如字节码中的 GETGLOBAL/SETGLOBAL 指令）。
//
// 参数：
//   - idx: 全局变量的数组索引（由 AllocateGlobalIndex 分配）
//
// 返回值：
//   - Value: 变量的值
//   - true: 索引有效
//   - null, false: 索引越界
//
// 使用示例：
//
//	// 编译器生成的字节码使用索引访问
//	// 假设编译器已分配 $counter 的索引为 5
//	value, ok := vm.GetGlobalByIndex(5)
//	if ok {
//	    fmt.Printf("counter = %v\n", value.Int())
//	}
func (vm *VM) GetGlobalByIndex(idx int) (Value, bool) {
	if idx < 0 || idx >= len(vm.globals) {
		return NewNull(), false
	}
	return vm.globals[idx], true
}

// SetGlobalByIndex 通过数组索引设置全局变量值（优化路径）。
//
// 此方法使用数组索引直接设置全局变量，避免 map 查找开销。
// 适用于编译器已知的索引（如字节码中的 SETGLOBAL 指令）。
// 如果启用了 GC，会自动管理引用计数。
//
// 参数：
//   - idx: 全局变量的数组索引（由 AllocateGlobalIndex 分配）
//   - value: 要设置的新值
//
// 注意：如果索引越界，此方法无操作。
//
// 使用示例：
//
//	// 编译器生成的字节码使用索引设置
//	// 假设编译器已分配 $counter 的索引为 5
//	vm.SetGlobalByIndex(5, engine.NewInt(100))
func (vm *VM) SetGlobalByIndex(idx int, value Value) {
	if idx < 0 || idx >= len(vm.globals) {
		return
	}
	old := vm.globals[idx]
	vm.globals[idx] = value
	vm.gcIncRef(value)
	vm.gcDecRef(old)
}

// ListFunctions 返回所有已注册函数名列表
// ListFunctions 返回所有已编译函数的名称列表。
//
// 此方法获取当前 VM 中所有已编译的函数名（不包括引擎注册的 Go 函数）。
// 适用于：
//   - 代码补全
//   - 函数列表展示
//   - 反射和元编程
//
// 返回值：
//   - []string: 编译函数名列表
//   - nil: 如果函数表未初始化
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    function add(a, b) { return a + b }
//	    function sub(a, b) { return a - b }
//	`)
//	vm.Execute()
//
//	names := vm.ListFunctions()
//	fmt.Printf("定义了 %d 个函数: %v\n", len(names), names)
//	// 输出: 定义了 2 个函数: [add, sub]
func (vm *VM) ListFunctions() []string {
	if vm.funcMap == nil {
		return nil
	}
	names := make([]string, 0, len(vm.funcMap))
	for name := range vm.funcMap {
		names = append(names, name)
	}
	return names
}

// FunctionInfo 包含函数的元数据信息，用于反射和代码分析。
//
// 此结构体提供了函数的完整描述，包括：
//   - 函数名
//   - 参数名列表（调试和文档生成用）
//   - 参数数量（函数重载匹配用）
//
// 应用场景：
//   - IDE 代码补全
//   - 自动生成文档
//   - 参数验证
//   - 函数签名匹配
type FunctionInfo struct {
	Name       string   // 函数名
	ParamNames []string // 参数名列表（按顺序）
	ParamCount int      // 参数数量
}

// GetFunctionInfo 获取指定函数的详细信息。
//
// 此方法返回函数的所有重载版本的元数据信息（FunctionInfo）。
// 支持函数重载，因此同一个函数名可能返回多个 FunctionInfo。
//
// 参数：
//   - name: 要查询的函数名
//
// 返回值：
//   - []FunctionInfo: 所有重载版本的信息列表
//   - true: 函数存在
//   - nil, false: 函数不存在
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    // 重载函数 greet
//	    function greet(name) { return "Hello, " + name }
//	    function greet(title, name) { return title + " " + name }
//	`)
//	vm.Execute()
//
//	infos, ok := vm.GetFunctionInfo("greet")
//	if ok {
//	    fmt.Printf("函数 'greet' 有 %d 个重载:\n", len(infos))
//	    for i, info := range infos {
//	        fmt.Printf("  %d. %s(%d params): %v\n",
//	            i+1, info.Name, info.ParamCount, info.ParamNames)
//	    }
//	}
func (vm *VM) GetFunctionInfo(name string) ([]FunctionInfo, bool) {
	fns, ok := vm.funcMap[name]
	if !ok || len(fns) == 0 {
		return nil, false
	}
	infos := make([]FunctionInfo, len(fns))
	for i, fn := range fns {
		infos[i] = FunctionInfo{
			Name:       fn.Name,
			ParamNames: fn.ParamNames,
			ParamCount: fn.Params,
		}
	}
	return infos, true
}

// CallByName 按函数名动态调用函数
// CallByName 通过函数名调用脚本中定义的函数。
//
// 此方法允许 Go 代码调用 JPL 脚本中定义的函数，支持：
//   - 脚本编译的函数
//   - 引擎注册的 Go 函数
//   - 函数重载（根据参数数量匹配）
//
// 调用时会：
//   - 查找匹配的函数（考虑参数数量）
//   - 保存当前执行状态
//   - 设置新的调用帧
//   - 执行函数
//   - 恢复之前的执行状态
//
// 栈溢出保护：如果调用深度超过 SetMaxCallDepth 设置的限制，返回 ErrStackOverflow。
//
// 线程安全：不是线程安全的，应在 Execute() 完成后或同步调用。
//
// 参数：
//   - name: 要调用的函数名
//   - args: 传递给函数的参数列表（变长参数）
//
// 返回值：
//   - Value: 函数的返回值
//   - nil: 调用成功
//   - ErrVMClosed: VM 已关闭
//   - ErrStackOverflow: 栈溢出
//   - 其他错误: 函数未定义、运行时错误等
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    function greet(name) {
//	        return "Hello, " + name
//	    }
//	    function add(a, b) {
//	        return a + b
//	    }
//	`)
//	vm.Execute()
//
//	// 调用 greet 函数
//	result1, err := vm.CallByName("greet", engine.NewString("World"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result1.String()) // 输出: Hello, World
//
//	// 调用 add 函数
//	result2, err := vm.CallByName("add", engine.NewInt(10), engine.NewInt(20))
//	fmt.Println(result2.Int()) // 输出: 30
func (vm *VM) CallByName(name string, args ...Value) (Value, error) {
	if vm.closed {
		return nil, ErrVMClosed
	}

	// 查找函数
	fn := vm.findFunction(name, len(args))
	if fn == nil {
		// 检查引擎注册的 Go 函数
		if vm.engine != nil {
			vm.engine.mu.RLock()
			goFn, exists := vm.engine.functions[name]
			vm.engine.mu.RUnlock()
			if exists {
				ctx := &Context{engine: vm.engine, vm: vm}
				return goFn(ctx, args)
			}
		}
		return nil, NewRuntimeError(fmt.Sprintf("undefined function: %s", name))
	}

	// 栈溢出保护
	if vm.callDepth >= vm.maxCallDepth {
		return nil, ErrStackOverflow
	}

	// 保存当前执行状态
	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedCallDepth := vm.callDepth
	savedCallStack := vm.callStack

	// 设置新帧
	vm.function = fn
	vm.ip = 0
	vm.registerBase = 0
	vm.callDepth = 0
	vm.callStack = nil

	// 创建寄存器窗口
	regCount := max(fn.Registers, fn.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}

	var err error

	// 创建闭包（支持递归调用）
	if fn.NumUpvals > 0 {
		cl := &closure{
			function: fn,
			upvals:   make([]*upvalue, fn.NumUpvals),
		}
		// 对于递归函数，upvalue 应该是闭包自身
		// 创建一个指向闭包的 upvalue
		for i := range cl.upvals {
			cl.upvals[i] = &upvalue{
				isClosed: true,
				closed:   cl,
			}
		}
		// 设置 GC
		if vm.gcField != nil {
			cl.SetupGC(vm.gcField)
		}
		// 关键：先设置 vm.upvals，再调用 callClosure
		// 这样 callClosure 保存的帧会有正确的 upvals
		vm.upvals = cl.upvals
		// 使用 callClosure 调用
		err = vm.callClosure(cl, args, 0)
		if err != nil {
			return nil, err
		}
		// 执行被调用的函数
		err = vm.run()
	} else {
		// 复制参数
		for i := 0; i < len(args) && i < fn.Params; i++ {
			vm.registers[i] = args[i]
		}

		// 追踪函数调用
		vm.traceCall(fn.Name, len(args))

		// 执行函数
		err = vm.run()
	}

	// 获取返回值
	var result Value
	if fn.NumUpvals > 0 {
		// 对于闭包调用，返回值在寄存器 0（callClosure 的 resultReg=0）
		if len(vm.registers) > 0 {
			result = vm.registers[0]
		} else {
			result = vm.result
		}
	} else {
		result = vm.result
	}

	// 恢复执行状态
	vm.ip = savedIP
	vm.function = savedFunc
	vm.registers = savedRegs
	vm.registerBase = savedRegBase
	vm.callDepth = savedCallDepth
	vm.callStack = savedCallStack

	if err != nil {
		return nil, err
	}
	return result, nil
}

// CallValue 调用任意函数值（闭包、Go 函数、字符串名函数）。
//
// 此方法是通用的函数调用接口，支持多种函数类型：
//   - funcValue: 引擎注册的 Go 函数（直接调用）
//   - closure: 脚本定义的闭包/lambda（通过 VM 执行）
//   - stringValue: 函数名字符串（查找后调用）
//
// 主要用于以下场景：
//   - Go 注册函数需要回调脚本传递的闭包/lambda
//   - 函数式编程中的高阶函数（map、filter、reduce 等）
//   - 动态调用函数（函数名从变量获取）
//
// 线程安全：不是线程安全的，应在 Execute() 完成后或同步调用。
//
// 参数：
//   - funcVal: 要调用的函数值，可以是 funcValue、closure 或 string
//   - args: 传递给函数的参数列表
//
// 返回值：
//   - Value: 函数的返回值
//   - nil: 调用成功
//   - ErrVMClosed: VM 已关闭
//   - runtimeError: 函数未定义、类型不匹配、运行时错误等
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    function apply(fn, arr) {
//	        return map(arr, fn)
//	    }
//	    function double(x) { return x * 2 }
//	`)
//	vm.Execute()
//
//	// 获取 apply 和 double 函数
//	applyFn, _ := vm.GetGlobal("apply")
//	arr := engine.NewArray([]engine.Value{
//	    engine.NewInt(1),
//	    engine.NewInt(2),
//	    engine.NewInt(3),
//	})
//
//	// 调用 apply 函数，传入 double 作为回调
//	doubleFn := engine.NewString("double")
//	result, err := vm.CallValue(applyFn, doubleFn, arr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Array()) // 输出: [2, 4, 6]
func (vm *VM) CallValue(funcVal Value, args ...Value) (Value, error) {
	if vm.closed {
		return nil, ErrVMClosed
	}

	// Go 函数值 — 直接调用
	if fn, ok := funcVal.(*funcValue); ok {
		ctx := &Context{engine: vm.engine, vm: vm}
		return fn.fn(ctx, args)
	}

	// 闭包 — 需要 VM 执行
	if cl, ok := funcVal.(*closure); ok {
		return vm.callValueClosure(cl, args)
	}

	// 字符串名 — 查找编译函数或注册函数
	if funcVal.Type() == TypeString {
		name := funcVal.String()

		// 检查引擎注册的 Go 函数
		if vm.engine != nil {
			vm.engine.mu.RLock()
			goFn, exists := vm.engine.functions[name]
			vm.engine.mu.RUnlock()
			if exists {
				ctx := &Context{engine: vm.engine, vm: vm}
				return goFn(ctx, args)
			}
		}

		// 查找编译后的 JPL 函数
		fn := vm.findFunction(name, len(args))
		if fn != nil {
			return vm.callValueCompiled(fn, args)
		}

		return nil, NewRuntimeError(fmt.Sprintf("undefined function: %s", name))
	}

	return nil, NewRuntimeError(fmt.Sprintf("cannot call value of type %s", funcVal.Type()))
}

// callValueClosure 调用闭包（保存/恢复 VM 状态）
func (vm *VM) callValueClosure(cl *closure, args []Value) (Value, error) {
	if vm.callDepth >= vm.maxCallDepth {
		return nil, ErrStackOverflow
	}

	// 保存当前执行状态
	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedCallDepth := vm.callDepth
	savedCallStack := vm.callStack
	savedUpvals := vm.upvals

	// 设置新帧
	vm.function = cl.function
	vm.ip = 0
	vm.callDepth = 0
	vm.callStack = nil

	// 创建寄存器窗口
	regCount := max(cl.function.Registers, cl.function.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}
	for i := 0; i < len(args) && i < cl.function.Params; i++ {
		vm.registers[i] = args[i]
	}

	// 设置 upvalue 列表
	vm.upvals = cl.upvals

	// 执行函数
	err := vm.run()
	result := vm.result

	// 恢复执行状态
	vm.ip = savedIP
	vm.function = savedFunc
	vm.registers = savedRegs
	vm.registerBase = savedRegBase
	vm.callDepth = savedCallDepth
	vm.callStack = savedCallStack
	vm.upvals = savedUpvals

	if err != nil {
		return nil, err
	}
	return result, nil
}

// callValueCompiled 调用编译后的 JPL 函数（保存/恢复 VM 状态）
func (vm *VM) callValueCompiled(fn *CompiledFunction, args []Value) (Value, error) {
	if vm.callDepth >= vm.maxCallDepth {
		return nil, ErrStackOverflow
	}

	// 保存当前执行状态
	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedCallDepth := vm.callDepth
	savedCallStack := vm.callStack

	// 设置新帧
	vm.function = fn
	vm.ip = 0
	vm.callDepth = 0
	vm.callStack = nil

	// 创建寄存器窗口
	regCount := max(fn.Registers, fn.Params)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}
	for i := 0; i < len(args) && i < fn.Params; i++ {
		vm.registers[i] = args[i]
	}

	// 执行函数
	err := vm.run()
	result := vm.result

	// 恢复执行状态
	vm.ip = savedIP
	vm.function = savedFunc
	vm.registers = savedRegs
	vm.registerBase = savedRegBase
	vm.callDepth = savedCallDepth
	vm.callStack = savedCallStack

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Close 关闭虚拟机
// Close 关闭虚拟机实例，释放所有内部资源。
//
// 此方法清理 VM 占用的内存，包括：
//   - 清除执行结果
//   - 释放寄存器
//   - 释放调用栈
//   - 标记 VM 为已关闭状态
//
// 关闭后的 VM 不能再使用，任何方法调用都会返回 ErrVMClosed。
//
// 线程安全：不是线程安全的，应在 Execute() 完成后调用。
//
// 返回值：
//   - nil: 成功关闭
//   - ErrVMClosed: VM 已经关闭
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$x = 100`)
//	vm.Execute()
//
//	// 使用完毕后关闭
//	if err := vm.Close(); err != nil {
//	    log.Printf("关闭 VM 失败: %v", err)
//	}
//
//	// 关闭后不可再使用
//	// vm.Execute() // 错误: ErrVMClosed
func (vm *VM) Close() error {
	if vm.closed {
		return ErrVMClosed
	}
	vm.closed = true
	vm.result = nil
	vm.registers = nil
	vm.callStack = nil
	return nil
}

// IsClosed 检查虚拟机是否已关闭。
//
// 返回 true 表示 VM 已经关闭，此时再调用 Execute() 等方法会返回 ErrVMClosed 错误。
//
// 返回值：
//   - true: VM 已关闭
//   - false: VM 处于活动状态
func (vm *VM) IsClosed() bool {
	return vm.closed
}

// GetGlobalNames 返回所有全局变量名列表
// GetGlobalNames 返回所有全局变量的名称列表。
//
// 此方法用于获取当前 VM 中所有已定义的全局变量名。
// 返回的列表包含所有通过脚本或 SetGlobal 设置的全局变量。
//
// 返回值：
//   - []string: 全局变量名列表（包含 $ 前缀）
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$a = 1; $b = 2`)
//	vm.Execute()
//
//	names := vm.GetGlobalNames()
//	for _, name := range names {
//	    fmt.Println(name) // 输出: $a, $b
//	}
func (vm *VM) GetGlobalNames() []string {
	names := make([]string, 0, len(vm.globalNames))
	for name := range vm.globalNames {
		names = append(names, name)
	}
	return names
}

// GetAllFunctionNames 返回所有可用的函数名称列表（包括编译的函数和引擎注册的函数）。
//
// 此方法合并了：
//   - 当前程序中编译的函数
//   - 引擎中注册的 Go 函数
//
// 返回的列表是去重的，不包含重复项。
//
// 返回值：
//   - []string: 所有函数名列表
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    function add(a, b) { return a + b }
//	    function sub(a, b) { return a - b }
//	`)
//	vm.Execute()
//
//	// 注册 Go 函数
//	eng.RegisterFunc("multiply", myMultiplyFunc)
//
//	names := vm.GetAllFunctionNames()
//	fmt.Printf("可用函数: %v\n", names) // 输出: [add, sub, multiply]
func (vm *VM) GetAllFunctionNames() []string {
	nameSet := make(map[string]bool)
	// 编译后的函数
	for name := range vm.funcMap {
		nameSet[name] = true
	}
	// 引擎注册的 Go 函数
	if vm.engine != nil {
		vm.engine.mu.RLock()
		for name := range vm.engine.functions {
			nameSet[name] = true
		}
		vm.engine.mu.RUnlock()
	}
	names := make([]string, 0, len(nameSet))
	for name := range nameSet {
		names = append(names, name)
	}
	return names
}

// FunctionExists 检查指定名称的函数是否存在。
//
// 此方法检查函数是否存在于：
//   - 当前程序编译的函数表中
//   - 引擎注册的 Go 函数中
//
// 参数：
//   - name: 要检查的函数名
//
// 返回值：
//   - true: 函数存在
//   - false: 函数不存在
//
// 使用示例：
//
//	vm, _ := eng.Compile(`function greet(name) { return "Hello, " + name }`)
//	vm.Execute()
//
//	if vm.FunctionExists("greet") {
//	    result, _ := vm.CallByName("greet", engine.NewString("World"))
//	    fmt.Println(result.String()) // 输出: Hello, World
//	}
func (vm *VM) FunctionExists(name string) bool {
	if _, ok := vm.funcMap[name]; ok {
		return true
	}
	if vm.engine != nil {
		vm.engine.mu.RLock()
		_, exists := vm.engine.functions[name]
		vm.engine.mu.RUnlock()
		return exists
	}
	return false
}

// GetEngineFunctionNames 返回引擎中注册的 Go 函数名称列表。
//
// 此方法仅返回通过 Engine.RegisterFunc 注册的 Go 函数名，
// 不包含脚本中定义的编译函数。
//
// 返回值：
//   - []string: 引擎注册的 Go 函数名列表
//   - nil: 如果引擎为空
//
// 使用示例：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 注册 Go 函数
//	eng.RegisterFunc("add", addFunc)
//	eng.RegisterFunc("sub", subFunc)
//
//	vm, _ := eng.Compile(`$x = 10`)
//	goFuncs := vm.GetEngineFunctionNames()
//	fmt.Printf("Go 函数: %v\n", goFuncs) // 输出: [add, sub]
func (vm *VM) GetEngineFunctionNames() []string {
	if vm.engine == nil {
		return nil
	}
	vm.engine.mu.RLock()
	defer vm.engine.mu.RUnlock()
	names := make([]string, 0, len(vm.engine.functions))
	for name := range vm.engine.functions {
		names = append(names, name)
	}
	return names
}

// Disassemble 反编译当前程序的字节码，返回可读的汇编格式文本。
//
// 此方法用于调试，将编译后的字节码转换为人类可读的汇编指令格式。
// 每行显示一条指令及其参数，便于理解程序的执行流程。
//
// 返回值：
//   - string: 反编译后的文本，如果程序为空则返回空字符串
//
// 使用示例：
//
//	vm, _ := eng.Compile(`
//	    $x = 10 + 20
//	    print($x)
//	`)
//
//	asm := vm.Disassemble()
//	fmt.Println(asm)
//	// 输出示例:
//	// == main ==
//	//   params: 0, registers: 2, upvals: 0, constants: 3, instructions: 4
//	//   0000: LOADK       R0, K0 (10)
//	//   0001: LOADK       R1, K1 (20)
//	//   0002: ADD         R0, R0, R1
//	//   0003: SETGLOBAL   R0, K2 ($x)
func (vm *VM) Disassemble() string {
	if vm.program == nil {
		return ""
	}
	return DisassembleProgram(vm.program)
}

// Reset 重置虚拟机的执行状态。
//
// 此方法清理 VM 的执行状态，包括：
//   - 清除执行结果和错误
//   - 释放寄存器和调用栈
//   - 重置指令指针和当前函数
//   - 清空异常处理器栈
//   - 重置中断状态
//
// 注意：此方法不会释放全局变量、函数表和程序。
// 适用于 REPL 场景，允许 VM 实例重复使用。
//
// 使用示例：
//
//	vm, _ := eng.Compile(`$x = 100`)
//	vm.Execute()
//
//	// 重置 VM 准备下一轮执行
//	vm.Reset()
//
//	// 重新编译并执行新程序
//	newVm, _ := eng.Compile(`$y = 200`)
//	vm.SetProgram(newVm.program)
//	vm.Execute()
func (vm *VM) Reset() {
	vm.result = NewNull()
	vm.err = nil
	vm.registers = nil
	vm.callStack = nil
	vm.callDepth = 0
	vm.ip = 0
	vm.function = nil
	vm.upvals = nil
	vm.tryHandlers = nil
	// 重置中断状态
	vm.ResetInterrupt()
	vm.running.Store(false)
}

// SetProgram 更新 VM 执行的程序（主要用于 REPL 场景）。
//
// 此方法允许动态更换 VM 的程序，新程序中的函数会追加到现有的函数表中，
// 保留之前定义的所有函数。这支持 REPL 中增量式定义函数和变量。
//
// 注意：调用此方法前建议先执行 Reset() 清理执行状态。
//
// 参数：
//   - prog: 新的程序实例
//
// 使用示例（REPL 场景）：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//
//	// 第一轮：定义函数
//	vm, _ := eng.Compile(`function add(a, b) { return a + b }`)
//	vm.Execute()
//
//	// 第二轮：添加更多代码（复用同一 VM）
//	vm.Reset()
//	prog2, _ := engine.CompileString(`function sub(a, b) { return a - b }`)
//	vm.SetProgram(prog2)
//	vm.Execute()
//
//	// 现在两个函数都可用
//	result, _ := vm.CallByName("add", engine.NewInt(10), engine.NewInt(20))
//	fmt.Println(result.Int()) // 输出: 30
func (vm *VM) SetProgram(prog *Program) {
	vm.program = prog
	vm.buildFuncMap()
}

// Interrupt 发送中断信号，中止当前正在执行的脚本。
//
// 此方法用于强制终止长时间运行的脚本（如死循环）。
// 中断信号会触发 ErrInterrupted 错误，VM 执行会立即停止。
//
// 特点：
//   - 线程安全，可从任意 goroutine 调用
//   - 幂等设计，可多次调用（使用 sync.Once 保护）
//   - 仅当 VM 正在执行时生效
//
// 使用场景：
//   - REPL 超时控制
//   - 用户取消长时间运行的计算
//   - 服务器关闭时终止所有脚本
//
// 使用示例：
//
//	// 启动一个可能长时间运行的脚本
//	vm, _ := eng.Compile(`while(true) {}`)
//
//	go func() {
//	    // 5秒后中断
//	    time.Sleep(5 * time.Second)
//	    vm.Interrupt()
//	}()
//
//	err := vm.Execute() // 5秒后被中断
//	if err == engine.ErrInterrupted {
//	    fmt.Println("脚本被中断")
//	}
func (vm *VM) Interrupt() {
	if vm.running.Load() {
		vm.interruptOnce.Do(func() {
			close(vm.interrupt)
		})
	}
}

// ResetInterrupt 重置中断状态，供下次执行使用
// ResetInterrupt 重置中断状态，为下一次执行做准备。
//
// 此方法在以下场景调用：
//   - 中断后重新执行脚本前
//   - Execute() 正常完成后
//   - Reset() 重置 VM 时
//
// 它会关闭旧的中断 channel（如果未关闭），并创建新的 channel。
//
// 注意：此方法确保 Interrupt() 在每次执行周期只生效一次。
//
// 使用示例：
//
//	vm.Interrupt() // 中断当前执行
//	vm.Execute()   // 脚本被中断，返回 ErrInterrupted
//
//	// 重置中断状态，准备下一轮
//	vm.ResetInterrupt()
//	vm.Reset() // 同时重置执行状态
//
//	// 现在可以重新执行
//	err := vm.Execute()
func (vm *VM) ResetInterrupt() {
	vm.interruptOnce.Do(func() {
		// 确保旧 channel 已关闭（若未关闭则关闭）
		close(vm.interrupt)
	})
	vm.interrupt = make(chan struct{})
	vm.interruptOnce = sync.Once{}
}

// SetInterruptCheckInterval 设置中断检查的频率。
//
// VM 在执行过程中每执行 N 条指令后检查一次中断信号。
// 默认值为 1000（每 1000 条指令检查一次）。
//
// 调整建议：
//   - 值越小，响应中断越快，但性能开销越大
//   - 值越大，性能越好，但中断响应延迟越大
//   - 默认值 1000 在响应速度和性能之间取得平衡
//
// 参数：
//   - n: 每 N 条指令检查一次中断，必须大于 0
//
// 使用示例：
//
//	vm, _ := eng.Compile(`while(true) {}`)
//
//	// 降低检查间隔，更快响应中断
//	vm.SetInterruptCheckInterval(100)
//
//	go func() {
//	    time.Sleep(100 * time.Millisecond)
//	    vm.Interrupt()
//	}()
//
//	vm.Execute() // 大约 100ms 后被中断
func (vm *VM) SetInterruptCheckInterval(n int) {
	if n > 0 {
		vm.interruptCheckInterval = n
	}
}

// InterruptChannel 返回中断信号 channel，供高级场景使用。
//
// 此方法返回只读的中断 channel，主要用于以下场景：
//   - 内置函数（如 sleep()）需要响应中断
//   - 自定义的外部操作需要与 VM 中断协同
//   - 实现复杂的超时控制逻辑
//
// 返回值：
//   - <-chan struct{}: 只读的中断信号 channel
//
// 使用示例（自定义 sleep 函数）：
//
//	eng.RegisterFunc("mySleep", func(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
//	    duration := time.Duration(args[0].Int()) * time.Millisecond
//
//	    select {
//	    case <-time.After(duration):
//	        return engine.NewNull(), nil
//	    case <-ctx.VM.InterruptChannel():
//	        return nil, engine.ErrInterrupted
//	    }
//	})
func (vm *VM) InterruptChannel() <-chan struct{} {
	return vm.interrupt
}

// IsRunning 检查 VM 当前是否正在执行脚本。
//
// 此方法返回 VM 的执行状态，主要用于：
//   - 判断是否可以安全调用 Interrupt()
//   - 监控脚本执行状态
//   - 防止重复执行（同一时间只能有一个 Execute() 运行）
//
// 返回值：
//   - true: VM 正在执行中
//   - false: VM 空闲或未初始化
//
// 使用示例：
//
//	if vm.IsRunning() {
//	    fmt.Println("脚本正在执行...")
//
//	    // 发送中断信号
//	    vm.Interrupt()
//
//	    // 等待执行结束
//	    for vm.IsRunning() {
//	        time.Sleep(10 * time.Millisecond)
//	    }
//	}
func (vm *VM) IsRunning() bool {
	return vm.running.Load()
}

// ============================================================================
// GC 方法
// ============================================================================

// SetGC 设置 GC 实例
func (vm *VM) SetGC(g *gc.GC) {
	vm.gcField = g
}

// GetGC 获取 GC 实例
func (vm *VM) GetGC() *gc.GC {
	return vm.gcField
}

// GCRoots 实现 gc.RootProvider 接口，返回所有 GC 根对象
func (vm *VM) GCRoots() []any {
	var roots []any

	// 全局变量
	for _, v := range vm.globals {
		if mo := AsManagedObject(v); mo != nil {
			roots = append(roots, mo)
		}
	}

	// 当前寄存器
	for _, v := range vm.registers {
		if mo := AsManagedObject(v); mo != nil {
			roots = append(roots, mo)
		}
	}

	// 调用栈中的寄存器
	for _, frame := range vm.callStack {
		for _, v := range frame.registers {
			if mo := AsManagedObject(v); mo != nil {
				roots = append(roots, mo)
			}
		}
	}

	// 执行结果
	if mo := AsManagedObject(vm.result); mo != nil {
		roots = append(roots, mo)
	}

	// 函数常量（通常不参与 GC，但闭包常量可能）
	if vm.function != nil {
		for _, c := range vm.function.Constants {
			if mo := AsManagedObject(c); mo != nil {
				roots = append(roots, mo)
			}
		}
	}

	return roots
}

// gcIncRef 增加 Value 的引用计数（如果它是托管对象）
func (vm *VM) gcIncRef(v Value) {
	if vm.gcField == nil || v == nil {
		return
	}
	if mo := AsManagedObject(v); mo != nil {
		vm.gcField.IncRef(mo)
	}
}

// gcDecRef 减少 Value 的引用计数（如果它是托管对象）
func (vm *VM) gcDecRef(v Value) {
	if vm.gcField == nil || v == nil {
		return
	}
	if mo := AsManagedObject(v); mo != nil {
		vm.gcField.DecRef(mo)
	}
}

// gcSetRegister 设置寄存器值并管理引用计数
func (vm *VM) gcSetRegister(reg int, val Value) {
	if reg < 0 || reg >= len(vm.registers) {
		return
	}
	// 检测运行时错误值，记录到引擎错误日志
	if re, ok := val.(*runtimeError); ok && vm.engine != nil {
		vm.engine.logError(NewRuntimeError(re.msg))
	}
	old := vm.registers[reg]
	vm.gcIncRef(val)
	vm.gcDecRef(old)
	vm.registers[reg] = val
}

// gcSetGlobal 设置全局变量并管理引用计数（兼容旧代码）
func (vm *VM) gcSetGlobal(name string, val Value) {
	// 检测运行时错误值，记录到引擎错误日志
	if re, ok := val.(*runtimeError); ok && vm.engine != nil {
		vm.engine.logError(NewRuntimeError(re.msg))
	}
	// 使用新的基于索引的方法
	vm.SetGlobal(name, val)
}

// ============================================================================
// 模块系统操作码
// ============================================================================

// opImport 执行 OP_IMPORT 指令
// A = 0: import "x"（自动推导命名空间名）
// A > 0 且 Constants[A] 是数组: from "x" import ...（A 是 names 常量索引）
// A > 0 且 Constants[A] 是字符串: import "x" as y（A 是 alias 常量索引）
// Bx = 源路径常量索引
func (vm *VM) opImport(ins Instruction) error {
	srcIdx := ins.Bx()
	param := ins.A()

	source := vm.function.Constants[srcIdx].String()

	// 检查缓存
	cache, ok := vm.engine.GetModule(source)
	if !ok {
		// 尝试加载
		var err error
		cache, err = vm.engine.LoadModule(source)
		if err != nil {
			return NewRuntimeError(fmt.Sprintf("import %q failed: %v", source, err))
		}
		// 缓存
		vm.engine.CacheModule(source, cache)
	}

	if param == 0 {
		// import "x" — 创建命名空间对象，变量名从路径自动推导
		nsName := moduleName(source)
		nsObj := make(map[string]Value)
		maps.Copy(nsObj, cache.Exports)
		vm.SetGlobal(nsName, NewObject(nsObj))
	} else {
		// A > 0，检查常量类型区分 from...import 和 import...as
		constVal := vm.function.Constants[param]
		if constVal.Type() == TypeArray {
			// from "x" import a, b, c — 选择性导入到全局
			names := constVal.Array()
			for _, nameVal := range names {
				name := nameVal.String()
				if val, ok := cache.Exports[name]; ok {
					vm.SetGlobal(name, val)
				} else {
					return NewRuntimeError(fmt.Sprintf(
						"import %q: symbol %q not found", source, name))
				}
			}
		} else {
			// import "x" as foo — 使用指定的命名空间名
			nsName := constVal.String()
			nsObj := make(map[string]Value)
			maps.Copy(nsObj, cache.Exports)
			vm.SetGlobal(nsName, NewObject(nsObj))
		}
	}

	return nil
}

// opInclude 执行 OP_INCLUDE 指令
// A = 1 表示 include_once，0 表示 include
// Bx = 源路径常量索引
func (vm *VM) opInclude(ins Instruction) error {
	srcIdx := ins.Bx()
	once := ins.A() == 1

	source := vm.function.Constants[srcIdx].String()

	// include_once：检查缓存（用特殊前缀区分）
	if once {
		cacheKey := "_once:" + source
		if _, ok := vm.engine.GetModule(cacheKey); ok {
			return nil // 已加载，跳过
		}
	}

	// 解析文件路径并读取内容
	resolvedPath, content, err := vm.loadIncludeSource(source)
	if err != nil {
		return NewRuntimeError(fmt.Sprintf("include %q failed: %v", source, err))
	}

	// 编译模块代码
	prog, err := CompileStringWithName(string(content), resolvedPath)
	if err != nil {
		return NewRuntimeError(fmt.Sprintf("include %q compile failed: %v", source, err))
	}

	// 在当前 VM 上下文中执行（共享 globals 和 funcMap）
	// 保存当前执行状态
	savedIP := vm.ip
	savedFunc := vm.function
	savedRegs := vm.registers
	savedRegBase := vm.registerBase
	savedProgram := vm.program
	savedGlobals := vm.globals // 保存原始 globals 引用
	savedGlobalNames := vm.globalNames
	savedGlobalIndexToName := vm.globalIndexToName

	// 将模块的全局变量名合并到当前 VM，并构建索引映射表
	moduleToVMIndex := make(map[int]int) // 模块索引 -> VM 索引
	for modIdx, name := range prog.GlobalNames {
		if vmIdx, exists := vm.globalNames[name]; exists {
			// 已存在，使用现有索引
			moduleToVMIndex[modIdx] = vmIdx
		} else {
			// 新全局变量，添加到当前 VM 的 globals
			vmIdx := len(vm.globals)
			vm.globalNames[name] = vmIdx
			vm.globals = append(vm.globals, NewNull())
			vm.globalIndexToName = append(vm.globalIndexToName, name)
			moduleToVMIndex[modIdx] = vmIdx
		}
	}
	// 关键：更新 savedGlobals 以反映可能的变化（append 可能重新分配数组）
	savedGlobals = vm.globals

	// 创建模块专用的 globals 数组（模块执行时使用）
	moduleGlobals := make([]Value, len(prog.GlobalNames))
	for i := range moduleGlobals {
		moduleGlobals[i] = NewNull()
	}
	// 复制已存在的全局变量值到模块 globals（从 vm.globals，它可能已经被扩展）
	for modIdx, vmIdx := range moduleToVMIndex {
		if modIdx < len(moduleGlobals) && vmIdx < len(vm.globals) {
			moduleGlobals[modIdx] = vm.globals[vmIdx]
		}
	}

	// 临时切换为模块的 globals，这样模块指令的索引就能正确工作
	vm.globals = moduleGlobals
	// 同时更新 globalNames 和 globalIndexToName 为模块的（临时）
	moduleGlobalNames := make(map[string]int)
	moduleGlobalIndexToName := make([]string, len(prog.GlobalNames))
	for i, name := range prog.GlobalNames {
		moduleGlobalNames[name] = i
		moduleGlobalIndexToName[i] = name
	}
	vm.globalNames = moduleGlobalNames
	vm.globalIndexToName = moduleGlobalIndexToName

	// 执行模块的主函数
	vm.program = prog
	vm.function = prog.Main
	vm.ip = 0
	vm.registerBase = 0
	regCount := max(prog.Main.Registers, 1)
	vm.registers = make([]Value, regCount)
	for i := range vm.registers {
		vm.registers[i] = NewNull()
	}

	// 将模块的函数合并到 funcMap
	for _, fn := range prog.Functions {
		if fn.Name != "<main>" {
			vm.funcMap[fn.Name] = append(vm.funcMap[fn.Name], fn)
		}
	}

	// 执行
	runErr := vm.run()

	// 将模块的 globals 复制回 VM 的 globals（根据索引映射）
	for modIdx, vmIdx := range moduleToVMIndex {
		if modIdx < len(moduleGlobals) && vmIdx < len(savedGlobals) {
			savedGlobals[vmIdx] = moduleGlobals[modIdx]
		}
	}

	// 恢复执行状态
	vm.ip = savedIP
	vm.function = savedFunc
	vm.registers = savedRegs
	vm.registerBase = savedRegBase
	vm.program = savedProgram
	vm.globals = savedGlobals
	vm.globalNames = savedGlobalNames
	vm.globalIndexToName = savedGlobalIndexToName

	if runErr != nil {
		return NewRuntimeError(fmt.Sprintf("include %q execution failed: %v", source, runErr))
	}

	// include_once：缓存（仅标记已加载）
	if once {
		cacheKey := "_once:" + source
		vm.engine.CacheModule(cacheKey, &ModuleCache{Exports: map[string]Value{}})
	}

	return nil
}

// loadIncludeSource 加载 include 的源文件内容
func (vm *VM) loadIncludeSource(source string) (string, []byte, error) {
	if vm.engine == nil {
		return "", nil, fmt.Errorf("no engine configured")
	}
	vm.engine.mu.RLock()
	loader := vm.engine.moduleLoader
	vm.engine.mu.RUnlock()

	if loader == nil {
		return "", nil, fmt.Errorf("no module loader configured")
	}

	// 使用加载器解析路径
	if fml, ok := loader.(*FileModuleLoader); ok {
		resolvedPath, err := fml.resolvePath(source)
		if err != nil {
			return "", nil, err
		}
		content, err := os.ReadFile(resolvedPath)
		if err != nil {
			return "", nil, err
		}
		return resolvedPath, content, nil
	}

	return "", nil, fmt.Errorf("unsupported module loader type")
}

// ============================================================================
// 调试方法
// ============================================================================

// SetDebugMode 设置调试模式。
// 当开启时，VM 会打印每条指令的执行信息和寄存器状态。
func (vm *VM) SetDebugMode(enabled bool) {
	vm.debugMode = enabled
	if enabled {
		vm.traceConfig = &TraceConfig{
			Enabled:  true,
			Writer:   os.Stdout,
			ShowRegs: true,
		}
	} else {
		vm.traceConfig = nil
	}
}

// GetDebugMode 获取当前调试模式状态。
func (vm *VM) GetDebugMode() bool {
	return vm.debugMode
}

// moduleName 从模块路径提取模块名
// "math" → "math"
// "math.jpl" → "math"
// "utils/helpers.jpl" → "helpers"
// "https://example.com/lib.jpl" → "lib"
func moduleName(source string) string {
	// 去掉 URL scheme
	if idx := len(source) - 1; idx >= 0 {
		for i := len(source) - 1; i >= 0; i-- {
			if source[i] == '/' {
				source = source[i+1:]
				break
			}
		}
	}
	// 去掉 .jpl 后缀
	if len(source) > 4 && source[len(source)-4:] == ".jpl" {
		source = source[:len(source)-4]
	}
	// 如果包含路径分隔符，取最后一段
	for i := len(source) - 1; i >= 0; i-- {
		if source[i] == '/' || source[i] == '\\' {
			source = source[i+1:]
			break
		}
	}
	return source
}
