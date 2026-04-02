# JPL 的内嵌引擎 API 文档

脚本引擎 API 使用示范：

```go
// 编译便捷函数示例
prog, err := jpl.CompileString("$x = 1 + 2")
if err != nil {
    log.Fatal(err)
}

// 带文件名的编译（更好的错误报告）
prog, err = jpl.CompileStringWithName("$y = $x * 2", "script.jpl")

// REPL场景：保持全局变量索引一致性
globals := []string{"x", "y"}
prog, err = jpl.CompileStringWithGlobals("$z = $x + $y", "repl", globals)

// 错误值处理
errVal := jpl.NewError("not found", 404, "HttpError")
if jpl.IsError(errVal) {
    if msg, ok := jpl.GetErrorField(errVal, "message"); ok {
        fmt.Println(msg.String()) // "not found"
    }
    if code, ok := jpl.GetErrorField(errVal, "code"); ok {
        fmt.Println(code.Int()) // 404
    }
}

// GC托管对象（高级场景）
gc := gc.New()
arr := jpl.NewArrayGC([]jpl.Value{jpl.NewInt(1), jpl.NewInt(2)}, gc)
obj := jpl.NewObjectGC(map[string]jpl.Value{"key": jpl.NewString("value")}, gc)

// 检查是否为GC管理对象
if jpl.IsManagedObject(arr) {
    mo := jpl.AsManagedObject(arr)
    fmt.Printf("引用计数: %d\n", mo.GetRefCount())
}

// 值工具函数
a := jpl.NewInt(10)
b := jpl.NewInt(20)
result := jpl.ValueAdd(a, b) // 等价于 a.Add(b)

// 类型强制转换
str := jpl.NewString("42")
num := jpl.CoerceToInt(str) // 42

// 模块加载器
loader := jpl.NewFileModuleLoader("./scripts")
loader.AddSearchPath("./modules")
eng.SetModuleLoader(loader)
```

---

## 上下文

```go
// Context 注册函数的上下文
type Context struct { ... }

// NewContext 创建新的上下文
func NewContext(engine *Engine, vm *VM) *Context

// Engine 返回关联的引擎实例
func (c *Context) Engine() *Engine

// VM 返回关联的虚拟机实例
func (c *Context) VM() *VM

// 结果构造函数
func (c *Context) NewResult(value Value) Value
func (c *Context) ResultNull() Value
func (c *Context) ResultBool(v bool) Value
func (c *Context) ResultInt(v int64) Value
func (c *Context) ResultFloat(v float64) Value
func (c *Context) ResultString(v string) Value
func (c *Context) ResultArray(v []Value) Value
func (c *Context) ResultObject(v map[string]Value) Value
```

---

## 引擎管理

```go
// Engine JPL 脚本引擎主结构体
type Engine struct { ... }

// NewEngine 创建新的引擎实例
func NewEngine() *Engine

// Close 关闭引擎，释放资源
func (e *Engine) Close() error

// IsClosed 检查引擎是否已关闭
func (e *Engine) IsClosed() bool
```

---

## 变量管理

```go
// Set 设置变量
func (e *Engine) Set(name string, value Value) error

// Get 获取变量
func (e *Engine) Get(name string) (Value, error)

// RegisterConst 注册常量
func (e *Engine) RegisterConst(name string, value Value) error

// GetConst 查询常量是否存在
func (e *Engine) GetConst(name string) (Value, bool)

// GetConstantNames 返回所有已注册的常量名称列表
func (e *Engine) GetConstantNames() []string
```

---

## 函数注册

```go
// RegisterFunc 注册 Go 函数供脚本调用
func (e *Engine) RegisterFunc(name string, fn GoFunction) error

// GetRegisteredFunc 获取已注册的 Go 函数
func (e *Engine) GetRegisteredFunc(name string) GoFunction

// GoFunction Go 语言注册函数类型
type GoFunction func(ctx *Context, args []Value) (Value, error)
```

---

## 内置函数

```go
// RegisterAll 注册所有内置函数到引擎
func RegisterAll(e *Engine)

// FunctionNames 返回所有内置函数名称列表（用于 REPL 自动补全）
func FunctionNames() []string
```

---

## Token

```go
// Keyword 查询标识符是否为关键字
func Keyword(name string) TokenType

// Keywords 返回所有关键字名称列表（用于 REPL 自动补全、语法高亮）
func Keywords() []string
```

---

## 编译与执行

```go
// Compile 编译脚本，返回虚拟机实例
func (e *Engine) Compile(script string) (*VM, error)

// CompileFile 编译脚本文件
func (e *Engine) CompileFile(filename string) (*VM, error)

// CompileString 编译脚本字符串（便捷入口）
func CompileString(script string) (*Program, error)

// CompileStringWithName 编译脚本字符串（指定文件名）
// 用于错误报告中显示正确的文件名
func CompileStringWithName(script string, filename string) (*Program, error)

// CompileStringWithGlobals 编译脚本字符串（指定文件名和现有全局变量）
// 用于 REPL 等需要保持全局变量索引一致性的场景
func CompileStringWithGlobals(script string, filename string, existingGlobals []string) (*Program, error)
```

---

## 编译器（高级）

```go
// NewCompiler 创建新的编译器实例
func NewCompiler() *Compiler

// NewCompilerWithGlobals 使用已存在的全局变量名创建编译器
// 用于 REPL 等场景，确保多次编译的代码能正确访问共享的全局变量
func NewCompilerWithGlobals(existingGlobals []string) *Compiler
```

---

## 虚拟机

```go
// VM 虚拟机实例
type VM struct { ... }

// NewVMWithProgram 使用程序创建虚拟机
func NewVMWithProgram(engine *Engine, prog *Program) *VM

// Execute 执行脚本
func (vm *VM) Execute() error

// ExecuteJSON 执行脚本并返回 JSON 序列化结果
func (vm *VM) ExecuteJSON() (string, error)

// GetResult 获取执行结果
func (vm *VM) GetResult() Value

// SetResult 设置执行结果
func (vm *VM) SetResult(value Value)

// Close 关闭虚拟机
func (vm *VM) Close() error

// IsClosed 检查虚拟机是否已关闭
func (vm *VM) IsClosed() bool

// Reset 重置虚拟机状态
func (vm *VM) Reset()

// SetProgram 更新虚拟机的程序（用于 REPL 场景）
// 新程序中的函数会追加到函数表，保留之前定义的函数
func (vm *VM) SetProgram(prog *Program)
```

---

## 反射 API

```go
// ListFunctions 返回所有已注册函数名列表
func (vm *VM) ListFunctions() []string

// GetFunctionInfo 获取指定函数的详细信息（返回所有重载版本）
func (vm *VM) GetFunctionInfo(name string) ([]FunctionInfo, bool)

// CallByName 按函数名动态调用函数
func (vm *VM) CallByName(name string, args ...Value) (Value, error)

// CallValue 调用任意函数值（闭包、Go 函数、字符串名函数）
// 供 Go 注册函数回调闭包/lambda 时使用
func (vm *VM) CallValue(funcVal Value, args ...Value) (Value, error)

// CurrentFunction 返回当前正在执行的编译函数
// 用于内置函数获取调用者信息，支持 func_num_args() 等反射函数
func (vm *VM) CurrentFunction() *CompiledFunction

// CurrentRegisters 返回当前函数的寄存器窗口
// 用于内置函数获取调用者的参数值，支持 func_get_arg() 等反射函数
func (vm *VM) CurrentRegisters() []Value

// FunctionInfo 函数信息
type FunctionInfo struct {
    Name       string   // 函数名
    ParamNames []string // 参数名列表
    ParamCount int      // 参数数量
}
```

---

## 调试 API

```go
// GetGlobal 获取全局变量
func (vm *VM) GetGlobal(name string) (Value, bool)

// SetGlobal 设置全局变量
func (vm *VM) SetGlobal(name string, value Value)

// GetGlobalNames 返回所有全局变量的名称列表
// 返回的索引对应 globals 切片中的位置
func (vm *VM) GetGlobalNames() []string

// Disassemble 反编译字节码（调试用）
func (vm *VM) Disassemble() string

// DisassembleProgram 反编译整个程序
func DisassembleProgram(prog *Program) string

// Disassemble 反编译函数字节码
func Disassemble(fn *CompiledFunction) string

// 错误日志（引擎自动记录非致命运行时错误）

// GetLastError 获取最后一个错误
func (e *Engine) GetLastError() error

// GetErrorLog 获取所有错误日志
func (e *Engine) GetErrorLog() []error

// ClearErrorLog 清空错误日志
func (e *Engine) ClearErrorLog()
```

---

## 值类型

```go
// Value 脚本引擎的值接口
type Value interface {
    // 类型检查
    Type() ValueType
    IsNull() bool
    Bool() bool

    // 值获取
    Int() int64
    Float() float64
    String() string
    Array() []Value
    Object() map[string]Value

    // 比较
    Equals(other Value) bool
    Stringify() string

    // 大数转换
    ToBigInt() Value
    ToBigDecimal() Value

    // 算术运算
    Add(other Value) Value
    Sub(other Value) Value
    Mul(other Value) Value
    Div(other Value) Value
    Mod(other Value) Value
    Negate() Value

    // 比较运算
    Less(other Value) bool
    Greater(other Value) bool
    LessEqual(other Value) bool
    GreaterEqual(other Value) bool
}

// ValueType 值类型枚举
type ValueType int

const (
    TypeNull       ValueType = iota // null
    TypeBool                        // bool
    TypeInt                         // int64
    TypeFloat                       // float64
    TypeString                      // string
    TypeArray                       // []Value
    TypeObject                      // map[string]Value
    TypeFunc                        // 函数
    TypeBigInt                      // 大整数 (big.Int)
    TypeBigDecimal                  // 大数 (big.Rat 有理数)
    TypeError                       // 错误对象
    TypeStream                      // 流资源 (io.Reader/Writer)
)
```

### 值构造函数

```go
// NewNull 创建 null 值
func NewNull() Value

// NewBool 创建布尔值
func NewBool(v bool) Value

// NewInt 创建整数值
func NewInt(v int64) Value

// NewFloat 创建浮点值
func NewFloat(v float64) Value

// NewString 创建字符串值
func NewString(v string) Value

// NewArray 创建数组值
func NewArray(v []Value) Value

// NewObject 创建对象值
func NewObject(v map[string]Value) Value

// NewBigInt 创建大整数值
func NewBigInt(v *big.Int) Value

// NewBigDecimal 创建大数值（有理数）
func NewBigDecimal(v *big.Rat) Value

// NewFunc 创建函数值（Go 函数）
func NewFunc(name string, fn GoFunction) Value

// NewRuntimeError 创建运行时错误值
func NewRuntimeError(message string) *RuntimeError
```

### 值工具函数

```go
// IsTruthy 检查值是否为真值（truthy）
func IsTruthy(v Value) bool

// IsComparable 检查两个值是否可以比较
func IsComparable(a, b Value) bool

// CoerceToInt 强制转换为 int64
func CoerceToInt(v Value) int64

// CoerceToFloat 强制转换为 float64
func CoerceToFloat(v Value) float64

// CoerceToString 强制转换为字符串
func CoerceToString(v Value) string

// CoerceToBool 强制转换为布尔值
func CoerceToBool(v Value) bool

// ValueAdd 执行加法（处理类型提升）
func ValueAdd(a, b Value) Value

// ValueSub 执行减法
func ValueSub(a, b Value) Value

// ValueMul 执行乘法
func ValueMul(a, b Value) Value

// ValueDiv 执行除法
func ValueDiv(a, b Value) Value

// ValueMod 执行取模
func ValueMod(a, b Value) Value

// ValueNegate 执行取反
func ValueNegate(a Value) Value

// ValueLess 小于比较
func ValueLess(a, b Value) bool

// ValueGreater 大于比较
func ValueGreater(a, b Value) bool

// ValueLessEqual 小于等于比较
func ValueLessEqual(a, b Value) bool

// ValueGreaterEqual 大于等于比较
func ValueGreaterEqual(a, b Value) bool

// ConcatValues 字符串拼接
func ConcatValues(a, b Value) Value
```

---

## 流类型 API

```go
// StreamMode 流模式
type StreamMode int

const (
    StreamRead      StreamMode = iota // 只读
    StreamWrite                       // 只写
    StreamReadWrite                   // 读写
)

// NewStream 创建通用流值
func NewStream(reader io.Reader, writer io.Writer, closer io.Closer, path string) Value

// NewFileStream 创建文件流
func NewFileStream(path string, mode StreamMode) (Value, error)

// NewStdinStream 创建标准输入流（包装 os.Stdin）
func NewStdinStream() Value

// NewStdoutStream 创建标准输出流（包装 os.Stdout）
func NewStdoutStream() Value

// NewStderrStream 创建标准错误流（包装 os.Stderr）
func NewStderrStream() Value

// NewBufferStream 创建内存缓冲流
func NewBufferStream(buf *bytes.Buffer) Value

// IsStream 检查值是否为流类型
func IsStream(v Value) bool

// StreamRead 从流中读取指定字节数
func StreamRead(v Value, length int) ([]byte, error)

// StreamReadLine 从流中读取一行
func StreamReadLine(v Value) (string, error)

// StreamWrite 向流中写入数据
func StreamWrite(v Value, data []byte) (int, error)

// StreamClose 关闭流
func StreamClose(v Value) error

// StreamEOF 检查流是否到达末尾
func StreamEOF(v Value) bool

// StreamFlush 刷新流缓冲区
func StreamFlush(v Value) error
```

---

## 错误类型

```go
// 错误常量
var (
    ErrEngineClosed     // 引擎已关闭
    ErrVMClosed         // 虚拟机已关闭
    ErrCompileFailed    // 编译失败
    ErrRuntimeFailed    // 运行时失败
    ErrInvalidArg       // 无效参数
    ErrTypeMismatch     // 类型不匹配
    ErrDivideByZero     // 除零错误（IEEE 754 模式下返回 Inf/NaN，此错误仅用于 null 除零等特殊情况）
    ErrUndefinedVar     // 未定义变量
    ErrUndefinedFunc    // 未定义函数
    ErrIndexOutOfBounds // 索引越界
    ErrStackOverflow    // 栈溢出
    ErrInterrupted      // 执行被中断（VM.Interrupt()）
)

// EngineError 引擎级别错误
type EngineError struct { Message string }
func NewEngineError(message string) *EngineError

// CompileError 编译错误
type CompileError struct {
    Message string
    Line    int
    Column  int
    File    string
}
func NewCompileError(message string) *CompileError

// RuntimeError 运行时错误
type RuntimeError struct {
    Message string
    Line    int
    Column  int
    File    string
}
func NewRuntimeError(message string) *RuntimeError
```

### 错误值处理

```go
// NewError 创建错误值
func NewError(message string, code int64, errType string) Value

// IsError 检查值是否为错误类型
func IsError(v Value) bool

// GetErrorField 获取错误值的指定字段（"message", "code", "type"）
func GetErrorField(v Value, field string) (Value, bool)
```

---

## 编译相关类型

```go
// Opcode 字节码操作码
type Opcode byte

// Instruction 32 位字节码指令
type Instruction uint32

// CompiledFunction 编译后的函数
type CompiledFunction struct {
    Name       string        // 函数名
    Params     int           // 参数数量
    ParamNames []string      // 参数名列表
    Registers  int           // 寄存器数量
    Bytecode   []Instruction // 字节码
    Constants  []Value       // 常量池
    NumUpvals  int           // upvalue 数量
    SourceLine int           // 源码行号
    VarNames   []string      // 调试用：寄存器索引 → 变量名
}

// Program 编译后的完整程序
type Program struct {
    Main      *CompiledFunction   // 主函数
    Functions []*CompiledFunction // 所有函数
    Constants []Value             // 全局常量池
}
```

---

## 指令构造函数

```go
// NewABC 创建 ABC 格式指令
func NewABC(op Opcode, a, b, c int) Instruction

// NewABx 创建 ABx 格式指令
func NewABx(op Opcode, a, bx int) Instruction

// NewAsBx 创建 AsBx 格式指令（有符号偏移）
func NewAsBx(op Opcode, a, sbx int) Instruction
```

---

### 垃圾回收

```go
// ManagedObject GC可管理的堆对象接口
type ManagedObject interface {
    ObjID() uint64
    GetRefCount() int
    IncRef()
    DecRef()
    IsAlive() bool
    MarkChildren(marker func(child any))
    OnFree()
}

// IsManagedObject 检查值是否为GC可管理对象
func IsManagedObject(v Value) bool

// AsManagedObject 将值转换为GC管理对象接口
func AsManagedObject(v Value) ManagedObject

// SetupGCValue 为堆类型值设置GC管理器
func SetupGCValue(v Value, g *gc.GC)

// NewArrayGC 创建GC托管的数组值
func NewArrayGC(v []Value, g *gc.GC) Value

// NewObjectGC 创建GC托管的对象值
func NewObjectGC(v map[string]Value, g *gc.GC) Value
```

---

## 追踪调试

```go
// TraceConfig 追踪配置
type TraceConfig struct {
    Enabled   bool      // 是否启用追踪
    Writer    io.Writer // 输出目标
    ShowRegs  bool      // 显示寄存器状态
    ShowStack bool      // 显示调用栈
    Hook      TraceHook // 自定义钩子
    MaxLines  int       // 最大输出行数
}

// NewTraceConfig 创建默认追踪配置
func NewTraceConfig() *TraceConfig

// NewTraceConfigWithWriter 创建带输出目标的追踪配置
func NewTraceConfigWithWriter(w io.Writer) *TraceConfig

// SetTraceConfig 设置追踪配置
func (vm *VM) SetTraceConfig(config *TraceConfig)
```

---

*此文档可参考 Go doc 获取最新信息: `go doc ./...`*

---

## 模块化

```go
// RegisterModule 注册 Go 模块
e.RegisterModule("mymath", map[string]GoFunction{
    "add": addFunc,
    "mul": mulFunc,
})
// 脚本中: import "mymath"; mymath.add(1, 2)

// SetModuleLoader 设置模块加载器（支持文件和 URL）
loader := engine.NewFileModuleLoader("./scripts")
loader.SetLockFile("./jpl.lock.yaml")   // 设置锁文件
loader.SetFrozen(true)                    // frozen 模式（CI 用）
e.SetModuleLoader(loader)

// SaveLockFile 保存锁文件变更
loader.SaveLockFile()
```

### 锁文件 (jpl.lock.yaml)

```yaml
version: 1
generated: "2026-03-23T12:00:00Z"
remote:
  https://example.com/lib.jpl:
    hash: "sha256:a1b2c3..."
    downloaded: "2026-03-23T12:00:00Z"
    size: 2048
```

- 自动更新：URL 导入时检测 hash 变化
- frozen 模式：hash 不匹配直接报错（CI 友好）
- 磁盘缓存：`~/.jpl/cache/<sha256>.jpl`

### 模块加载器

```go
// ModuleCache 模块缓存
type ModuleCache struct {
    Exports map[string]Value
}

// ModuleLoader 模块加载器接口
type ModuleLoader interface {
    LoadModule(source string, engine *Engine) (*ModuleCache, error)
}

// FileModuleLoader 文件系统模块加载器
type FileModuleLoader struct {
    // ... 内部字段
}

// NewFileModuleLoader 创建文件模块加载器
func NewFileModuleLoader(baseDir string) *FileModuleLoader
```
