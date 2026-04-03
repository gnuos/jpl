// Package engine 实现 JPL 脚本语言核心引擎
package engine

import "fmt"

// ============================================================================
// 操作码定义
// ============================================================================

// Opcode 字节码操作码
type Opcode byte

const (
	// 加载/存储
	OP_NOP       Opcode = iota // 空操作
	OP_LOAD                    // R[A] = R[B]（寄存器拷贝）
	OP_LOADK                   // R[A] = K[Bx]（加载常量）
	OP_LOADNULL                // R[A] = null
	OP_LOADBOOL                // R[A] = bool(B)
	OP_GETGLOBAL               // R[A] = Globals[K[Bx]]
	OP_SETGLOBAL               // Globals[K[Bx]] = R[A]
	OP_GETVAR                  // R[A] = lookup(R[B].String())（动态变量读取）
	OP_SETVAR                  // assign(R[B].String(), R[C])（动态变量写入）

	// 算术运算 R[A] = R[B] op R[C]
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_NEG // R[A] = -R[B]

	// 比较运算 R[A] = R[B] op R[C]
	OP_EQ
	OP_NEQ
	OP_LT
	OP_GT
	OP_LTE
	OP_GTE

	// 字符串连接
	OP_CONCAT // R[A] = R[B] .. R[C]

	// 范围
	OP_RANGE           // R[A] = range(R[B], R[C]) 半开区间
	OP_RANGE_INCLUSIVE // R[A] = range(R[B], R[C]) 闭区间

	// 逻辑运算
	OP_AND // R[A] = R[B] && R[C]
	OP_OR  // R[A] = R[B] || R[C]
	OP_NOT // R[A] = !R[B]

	// 数组和对象
	OP_NEWARRAY  // R[A] = [R[B], R[B+1], ..., R[B+C-1]]
	OP_NEWOBJECT // R[A] = {K[Bx]: R[B], ...}（N 个键值对）
	OP_GETINDEX  // R[A] = R[B][R[C]]
	OP_SETINDEX  // R[B][R[C]] = R[A]
	OP_GETMEMBER // R[A] = R[B].K[Cx]
	OP_SETMEMBER // R[B].K[Cx] = R[A]

	// 控制流
	OP_JMP      // PC += sBx
	OP_JMPIF    // if R[A] then PC += sBx
	OP_JMPIFNOT // if !R[A] then PC += sBx

	// 函数调用
	OP_CALL      // R[A](R[A+1], ..., R[A+C-1])，结果存 R[A]
	OP_TAIL_CALL // 尾调用：复用当前栈帧
	OP_RETURN    // return R[A]

	// 闭包（Phase 3 实现）
	OP_CLOSURE      // R[A] = Closure(K[Bx])，后跟 Cx 条捕获描述
	OP_GETUPVAL     // R[A] = Upval[B]
	OP_SETUPVAL     // Upval[A] = R[B]
	OP_CLOSE_UPVALS // 关闭 R[A] 及以上的 upvalue

	// 异常处理
	OP_THROW     // throw R[A]
	OP_TRY_BEGIN // try 块开始，sBx = catch 跳转偏移
	OP_TRY_END   // try 块结束

	// 其他
	OP_POP    // 弹出栈顶（丢弃值）
	OP_DUP    // 复制栈顶
	OP_TYPEOF // R[A] = type(R[B]) 的字符串表示
	OP_CAST   // R[A] = cast(R[B], C) 类型转换指令，实现 Go 风格类型转换语法 int(x)/float(x)/string(x)/bool(x)
	// 参数: A=目标寄存器(存储结果), B=源寄存器(被转换值), C=转换类型
	// C 值含义: 0=int, 1=float, 2=string, 3=bool
	// 转换规则参见 vm.go 中的 castToInt/castToFloat 函数
	OP_FORMAT // R[A] = sprintf(R[B], R[C]) 字符串格式化（用于插值格式化）

	// 模块系统
	OP_IMPORT  // 加载模块，Bx=源路径常量索引，B=1表示选择性导入
	OP_INCLUDE // 执行并注入，Bx=源路径常量索引，B=1表示 include_once

	// 位运算（新增）
	OP_BITAND // R[A] = R[B] & R[C]（按位与）
	OP_BITOR  // R[A] = R[B] | R[C]（按位或）
	OP_BITXOR // R[A] = R[B] ^ R[C]（按位异或）
	OP_BITNOT // R[A] = ~R[B]（按位取反）
	OP_SHL    // R[A] = R[B] << R[C]（左移）
	OP_SHR    // R[A] = R[B] >> R[C]（右移）

	// 迭代器（foreach 支持）
	OP_ITERINIT // R[A] = iter(R[B])，初始化迭代器，A=迭代器寄存器，B=被迭代对象
	OP_ITERNEXT // if (key, val = next(R[A])) { PC += sBx } else { 结束迭代 }，C=0表示只取值
	OP_ITEREND  // 清理迭代器资源（可选，用于显式关闭迭代器）

	// 运行时魔术常量（命令行参数）
	OP_GETARGV // R[A] = script_argv (命令行参数数组)
	OP_GETARGC // R[A] = script_argc (命令行参数数量)

	// 正则匹配
	OP_REGEX_MATCH // R[A] = R[B] =~ R[C] (字符串匹配正则，返回 bool)

	OP_COUNT // 操作码总数（哨兵）
)

var opcodeNames = [...]string{
	OP_NOP:             "NOP",
	OP_LOAD:            "LOAD",
	OP_LOADK:           "LOADK",
	OP_LOADNULL:        "LOADNULL",
	OP_LOADBOOL:        "LOADBOOL",
	OP_GETGLOBAL:       "GETGLOBAL",
	OP_SETGLOBAL:       "SETGLOBAL",
	OP_GETVAR:          "GETVAR",
	OP_SETVAR:          "SETVAR",
	OP_ADD:             "ADD",
	OP_SUB:             "SUB",
	OP_MUL:             "MUL",
	OP_DIV:             "DIV",
	OP_MOD:             "MOD",
	OP_NEG:             "NEG",
	OP_EQ:              "EQ",
	OP_NEQ:             "NEQ",
	OP_LT:              "LT",
	OP_GT:              "GT",
	OP_LTE:             "LTE",
	OP_GTE:             "GTE",
	OP_CONCAT:          "CONCAT",
	OP_RANGE:           "RANGE",
	OP_RANGE_INCLUSIVE: "RANGE_INCLUSIVE",
	OP_AND:             "AND",
	OP_OR:              "OR",
	OP_NOT:             "NOT",
	OP_NEWARRAY:        "NEWARRAY",
	OP_NEWOBJECT:       "NEWOBJECT",
	OP_GETINDEX:        "GETINDEX",
	OP_SETINDEX:        "SETINDEX",
	OP_GETMEMBER:       "GETMEMBER",
	OP_SETMEMBER:       "SETMEMBER",
	OP_JMP:             "JMP",
	OP_JMPIF:           "JMPIF",
	OP_JMPIFNOT:        "JMPIFNOT",
	OP_CALL:            "CALL",
	OP_TAIL_CALL:       "TAIL_CALL",
	OP_RETURN:          "RETURN",
	OP_CLOSURE:         "CLOSURE",
	OP_GETUPVAL:        "GETUPVAL",
	OP_SETUPVAL:        "SETUPVAL",
	OP_CLOSE_UPVALS:    "CLOSE_UPVALS",
	OP_THROW:           "THROW",
	OP_TRY_BEGIN:       "TRY_BEGIN",
	OP_TRY_END:         "TRY_END",
	OP_POP:             "POP",
	OP_DUP:             "DUP",
	OP_TYPEOF:          "TYPEOF",
	OP_CAST:            "CAST",
	OP_FORMAT:          "FORMAT",
	OP_IMPORT:          "IMPORT",
	OP_INCLUDE:         "INCLUDE",
	OP_BITAND:          "BITAND",
	OP_BITOR:           "BITOR",
	OP_BITXOR:          "BITXOR",
	OP_BITNOT:          "BITNOT",
	OP_SHL:             "SHL",
	OP_SHR:             "SHR",
	OP_ITERINIT:        "ITERINIT",
	OP_ITERNEXT:        "ITERNEXT",
	OP_ITEREND:         "ITEREND",
	OP_GETARGV:         "GETARGV",
	OP_GETARGC:         "GETARGC",
	OP_REGEX_MATCH:     "REGEX_MATCH",
}

func (op Opcode) String() string {
	if int(op) < len(opcodeNames) {
		return opcodeNames[op]
	}
	return fmt.Sprintf("OP(%d)", op)
}

// ============================================================================
// 指令格式
// ============================================================================

// Instruction 32 位字节码指令
//
//	ABC 格式: [OP:8][A:8][B:8][C:8]
//	ABx 格式: [OP:8][A:8][ Bx:16  ]
//	AsBx 格式:[OP:8][A:8][ sBx:16 ]
type Instruction uint32

const (
	maskOP = 0xFF
	maskA  = 0xFF
	maskB  = 0xFF
	maskC  = 0xFF
	maskBx = 0xFFFF

	shiftOP = 24
	shiftA  = 16
	shiftB  = 8
	shiftC  = 0
	shiftBx = 0

	maxSBx = 32767 // 2^15 - 1
)

// NewABC 创建 ABC 格式指令
func NewABC(op Opcode, a, b, c int) Instruction {
	return Instruction(uint32(op)<<shiftOP |
		uint32(byte(a))<<shiftA |
		uint32(byte(b))<<shiftB |
		uint32(byte(c))<<shiftC)
}

// NewABx 创建 ABx 格式指令
func NewABx(op Opcode, a, bx int) Instruction {
	return Instruction(uint32(op)<<shiftOP |
		uint32(byte(a))<<shiftA |
		uint32(uint16(bx))<<shiftBx)
}

// NewAsBx 创建 AsBx 格式指令（有符号偏移）
func NewAsBx(op Opcode, a, sbx int) Instruction {
	return Instruction(uint32(op)<<shiftOP |
		uint32(byte(a))<<shiftA |
		uint32(uint16(sbx+maxSBx))<<shiftBx)
}

func (ins Instruction) OP() Opcode { return Opcode((ins >> shiftOP) & maskOP) }
func (ins Instruction) A() int     { return int((ins >> shiftA) & maskA) }
func (ins Instruction) B() int     { return int((ins >> shiftB) & maskB) }
func (ins Instruction) C() int     { return int((ins >> shiftC) & maskC) }
func (ins Instruction) Bx() int    { return int((ins >> shiftBx) & maskBx) }
func (ins Instruction) AsBx() int  { return int((ins>>shiftBx)&maskBx) - maxSBx }

// ============================================================================
// 编译输出
// ============================================================================

// CompiledFunction 编译后的函数
type CompiledFunction struct {
	Name        string        // 函数名
	Params      int           // 参数数量
	ParamNames  []string      // 参数名列表（反射用）
	Registers   int           // 寄存器数量
	Bytecode    []Instruction // 字节码
	Constants   []Value       // 常量池
	NumUpvals   int           // upvalue 数量
	Upvals      []UpvalueDesc // upvalue 描述
	SourceLine  int           // 源码行号（函数起始行）
	SourceLines []int         // 源码行号表（每条指令对应的行号）
	VarNames    []string      // 调试用：寄存器索引 → 变量名
}

// Program 编译后的完整程序
type Program struct {
	Main        *CompiledFunction   // 主函数（入口点）
	Functions   []*CompiledFunction // 所有函数（包括主函数）
	Constants   []Value             // 全局常量池
	GlobalNames []string            // 全局变量名列表（索引 -> 名称）
	Source      string              // 原始源代码
	SourceLines []string            // 按行分割的源代码
}

// ============================================================================
// 反编译器
// ============================================================================

// Disassemble 反编译函数字节码（调试用）
func Disassemble(fn *CompiledFunction) string {
	if fn == nil {
		return "<nil>"
	}
	var result string
	result += fmt.Sprintf("== %s ==\n", fn.Name)
	result += fmt.Sprintf("  params: %d, registers: %d, upvals: %d, constants: %d, instructions: %d\n",
		fn.Params, fn.Registers, fn.NumUpvals, len(fn.Constants), len(fn.Bytecode))

	for i, ins := range fn.Bytecode {
		result += fmt.Sprintf("  %04d: %s\n", i, disassembleInstruction(ins, fn.Constants))
	}
	return result
}

func disassembleInstruction(ins Instruction, consts []Value) string {
	op := ins.OP()
	a := ins.A()
	b := ins.B()
	c := ins.C()
	bx := ins.Bx()
	sbx := ins.AsBx()

	switch op {
	case OP_LOADK:
		var kstr string
		if bx < len(consts) {
			kstr = consts[bx].Stringify()
		} else {
			kstr = fmt.Sprintf("K(%d)", bx)
		}
		return fmt.Sprintf("%-12s R%d, %s", op, a, kstr)
	case OP_GETGLOBAL, OP_SETGLOBAL:
		var kstr string
		if bx < len(consts) {
			kstr = consts[bx].Stringify()
		} else {
			kstr = fmt.Sprintf("K(%d)", bx)
		}
		return fmt.Sprintf("%-12s R%d, %s", op, a, kstr)
	case OP_LOADBOOL:
		return fmt.Sprintf("%-12s R%d, %t", op, a, b != 0)
	case OP_LOAD, OP_LOADNULL, OP_NEG, OP_NOT, OP_BITNOT, OP_TYPEOF:
		return fmt.Sprintf("%-12s R%d, R%d", op, a, b)
	case OP_GETVAR:
		return fmt.Sprintf("%-12s R%d, R%d", op, a, b)
	case OP_SETVAR:
		return fmt.Sprintf("%-12s R%d, R%d", op, b, c)
	case OP_ADD, OP_SUB, OP_MUL, OP_DIV, OP_MOD,
		OP_EQ, OP_NEQ, OP_LT, OP_GT, OP_LTE, OP_GTE,
		OP_CONCAT, OP_AND, OP_OR,
		OP_BITAND, OP_BITOR, OP_BITXOR, OP_SHL, OP_SHR,
		OP_GETINDEX, OP_SETINDEX:
		return fmt.Sprintf("%-12s R%d, R%d, R%d", op, a, b, c)
	case OP_JMP:
		return fmt.Sprintf("%-12s %+d", op, sbx)
	case OP_JMPIF, OP_JMPIFNOT:
		return fmt.Sprintf("%-12s R%d, %+d", op, a, sbx)
	case OP_CALL, OP_TAIL_CALL:
		return fmt.Sprintf("%-12s R%d, %d", op, a, c)
	case OP_RETURN:
		return fmt.Sprintf("%-12s R%d", op, a)
	case OP_NEWARRAY:
		return fmt.Sprintf("%-12s R%d, R%d, %d", op, a, b, c)
	case OP_GETMEMBER, OP_SETMEMBER:
		var kstr string
		if c < len(consts) {
			kstr = consts[c].Stringify()
		} else {
			kstr = fmt.Sprintf("K(%d)", c)
		}
		return fmt.Sprintf("%-12s R%d, R%d, %s", op, a, b, kstr)
	case OP_CLOSURE:
		return fmt.Sprintf("%-12s R%d, K(%d), %d", op, a, bx, c)
	case OP_GETUPVAL, OP_SETUPVAL, OP_CLOSE_UPVALS:
		return fmt.Sprintf("%-12s R%d, %d", op, a, b)
	case OP_THROW, OP_POP, OP_DUP:
		return fmt.Sprintf("%-12s R%d", op, a)
	case OP_TRY_BEGIN:
		return fmt.Sprintf("%-12s %+d", op, sbx)
	case OP_TRY_END:
		return op.String()
	case OP_IMPORT, OP_INCLUDE:
		var srcStr string
		if bx < len(consts) {
			srcStr = consts[bx].Stringify()
		} else {
			srcStr = fmt.Sprintf("K(%d)", bx)
		}
		return fmt.Sprintf("%-12s %s, B=%d", op, srcStr, b)
	case OP_NOP:
		return op.String()
	default:
		return fmt.Sprintf("%s R%d, %d, %d", op, a, b, c)
	}
}
