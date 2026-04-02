package jpl

import (
	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/gc"
)

// 类型别名导出
type (
	GoFunction = engine.GoFunction
	Value      = engine.Value
	VM         = engine.VM
	Context    = engine.Context
	Engine     = engine.Engine
)

// 错误常量导出
var (
	ErrEngineClosed     = engine.ErrEngineClosed
	ErrVMClosed         = engine.ErrVMClosed
	ErrCompileFailed    = engine.ErrCompileFailed
	ErrRuntimeFailed    = engine.ErrRuntimeFailed
	ErrInvalidArg       = engine.ErrInvalidArg
	ErrTypeMismatch     = engine.ErrTypeMismatch
	ErrDivideByZero     = engine.ErrDivideByZero
	ErrUndefinedVar     = engine.ErrUndefinedVar
	ErrUndefinedFunc    = engine.ErrUndefinedFunc
	ErrIndexOutOfBounds = engine.ErrIndexOutOfBounds
	ErrStackOverflow    = engine.ErrStackOverflow
)

// 构造函数导出
var (
	NewEngine       = engine.NewEngine
	NewNull         = engine.NewNull
	NewBool         = engine.NewBool
	NewInt          = engine.NewInt
	NewFloat        = engine.NewFloat
	NewString       = engine.NewString
	NewArray        = engine.NewArray
	NewObject       = engine.NewObject
	NewContext      = engine.NewContext
	NewEngineError  = engine.NewEngineError
	NewCompileError = engine.NewCompileError
	NewRuntimeError = engine.NewRuntimeError
)

// 类型枚举导出
const (
	TypeNull       = engine.TypeNull
	TypeBool       = engine.TypeBool
	TypeInt        = engine.TypeInt
	TypeFloat      = engine.TypeFloat
	TypeString     = engine.TypeString
	TypeArray      = engine.TypeArray
	TypeObject     = engine.TypeObject
	TypeFunc       = engine.TypeFunc
	TypeBigDecimal = engine.TypeBigDecimal
)

// 编译相关导出
type (
	Opcode           = engine.Opcode
	Instruction      = engine.Instruction
	CompiledFunction = engine.CompiledFunction
	Program          = engine.Program
	Symbol           = engine.Symbol
	UpvalueDesc      = engine.UpvalueDesc
	Compiler         = engine.Compiler
	FunctionInfo     = engine.FunctionInfo
)

// 操作码导出
const (
	OP_NOP          = engine.OP_NOP
	OP_LOAD         = engine.OP_LOAD
	OP_LOADK        = engine.OP_LOADK
	OP_LOADNULL     = engine.OP_LOADNULL
	OP_LOADBOOL     = engine.OP_LOADBOOL
	OP_GETGLOBAL    = engine.OP_GETGLOBAL
	OP_SETGLOBAL    = engine.OP_SETGLOBAL
	OP_GETVAR       = engine.OP_GETVAR
	OP_SETVAR       = engine.OP_SETVAR
	OP_ADD          = engine.OP_ADD
	OP_SUB          = engine.OP_SUB
	OP_MUL          = engine.OP_MUL
	OP_DIV          = engine.OP_DIV
	OP_MOD          = engine.OP_MOD
	OP_NEG          = engine.OP_NEG
	OP_EQ           = engine.OP_EQ
	OP_NEQ          = engine.OP_NEQ
	OP_LT           = engine.OP_LT
	OP_GT           = engine.OP_GT
	OP_LTE          = engine.OP_LTE
	OP_GTE          = engine.OP_GTE
	OP_CONCAT       = engine.OP_CONCAT
	OP_AND          = engine.OP_AND
	OP_OR           = engine.OP_OR
	OP_NOT          = engine.OP_NOT
	OP_NEWARRAY     = engine.OP_NEWARRAY
	OP_NEWOBJECT    = engine.OP_NEWOBJECT
	OP_GETINDEX     = engine.OP_GETINDEX
	OP_SETINDEX     = engine.OP_SETINDEX
	OP_GETMEMBER    = engine.OP_GETMEMBER
	OP_SETMEMBER    = engine.OP_SETMEMBER
	OP_JMP          = engine.OP_JMP
	OP_JMPIF        = engine.OP_JMPIF
	OP_JMPIFNOT     = engine.OP_JMPIFNOT
	OP_CALL         = engine.OP_CALL
	OP_RETURN       = engine.OP_RETURN
	OP_CLOSURE      = engine.OP_CLOSURE
	OP_GETUPVAL     = engine.OP_GETUPVAL
	OP_SETUPVAL     = engine.OP_SETUPVAL
	OP_CLOSE_UPVALS = engine.OP_CLOSE_UPVALS
	OP_THROW        = engine.OP_THROW
	OP_TRY_BEGIN    = engine.OP_TRY_BEGIN
	OP_TRY_END      = engine.OP_TRY_END
	OP_POP          = engine.OP_POP
	OP_DUP          = engine.OP_DUP
	OP_TYPEOF       = engine.OP_TYPEOF
	OP_COUNT        = engine.OP_COUNT
)

// 指令构造函数导出
var (
	NewABC  = engine.NewABC
	NewABx  = engine.NewABx
	NewAsBx = engine.NewAsBx
)

// 工具函数导出
var (
	Disassemble        = engine.Disassemble
	DisassembleProgram = engine.DisassembleProgram
	IsNumeric          = engine.IsNumeric
)

// VM 追踪相关导出
type (
	TraceConfig = engine.TraceConfig
	TraceHook   = engine.TraceHook
	TraceEvent  = engine.TraceEvent
)

// VM 方法导出（内部函数不导出）

// ============================================================================
// 新增API导出
// ============================================================================

// 编译便捷函数导出
var (
	CompileString            = engine.CompileString
	CompileStringWithName    = engine.CompileStringWithName
	CompileStringWithGlobals = engine.CompileStringWithGlobals
)

// 错误常量补充导出
var (
	ErrInterrupted = engine.ErrInterrupted
)

// 错误类型导出
type (
	EngineError  = engine.EngineError
	CompileError = engine.CompileError
	RuntimeError = engine.RuntimeError
)

// 值工具函数导出
var (
	IsTruthy          = engine.IsTruthy
	IsComparable      = engine.IsComparable
	CoerceToInt       = engine.CoerceToInt
	CoerceToFloat     = engine.CoerceToFloat
	CoerceToString    = engine.CoerceToString
	CoerceToBool      = engine.CoerceToBool
	ValueAdd          = engine.ValueAdd
	ValueSub          = engine.ValueSub
	ValueMul          = engine.ValueMul
	ValueDiv          = engine.ValueDiv
	ValueMod          = engine.ValueMod
	ValueNegate       = engine.ValueNegate
	ValueLess         = engine.ValueLess
	ValueGreater      = engine.ValueGreater
	ValueLessEqual    = engine.ValueLessEqual
	ValueGreaterEqual = engine.ValueGreaterEqual
	ConcatValues      = engine.ConcatValues
)

// 错误值处理函数导出
var (
	NewError      = engine.NewError
	IsError       = engine.IsError
	GetErrorField = engine.GetErrorField
)

// GC相关函数导出
var (
	IsManagedObject = engine.IsManagedObject
	AsManagedObject = engine.AsManagedObject
	SetupGCValue    = engine.SetupGCValue
	NewArrayGC      = engine.NewArrayGC
	NewObjectGC     = engine.NewObjectGC
)

// 追踪配置构造函数导出
var (
	NewTraceConfig           = engine.NewTraceConfig
	NewTraceConfigWithWriter = engine.NewTraceConfigWithWriter
)

// 模块加载器导出
type (
	ModuleCache      = engine.ModuleCache
	ModuleLoader     = engine.ModuleLoader
	FileModuleLoader = engine.FileModuleLoader
)

var (
	NewFileModuleLoader = engine.NewFileModuleLoader
)

// GC接口导出
type (
	ManagedObject = gc.ManagedObject
)
