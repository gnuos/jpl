package engine

import (
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/parser"
	"github.com/gnuos/jpl/token"
)

// ============================================================================
// 符号表
// ============================================================================

// Symbol 变量符号
type Symbol struct {
	Name     string // 变量名
	Index    int    // 寄存器索引或 upvalue 索引
	Scope    int    // 声明时的作用域深度
	Captured bool   // 是否被内层闭包捕获
}

// UpvalueDesc upvalue 描述
type UpvalueDesc struct {
	Index   int
	IsLocal bool
}

// ============================================================================
// 循环上下文
// ============================================================================

// loopContext 循环编译上下文
type loopContext struct {
	breakJumps    []int // break 跳转待修补位置
	continueJumps []int // continue 跳转待修补位置
	continuePC    int   // continue 目标 PC
}

// ============================================================================
// 编译器
// ============================================================================

// Compiler 是 JPL 字节码编译器，将 AST 转换为字节码。
//
// 编译器实现了单遍编译（Single-Pass Compilation），直接遍历 AST 生成字节码，
// 无需中间表示（IR）。主要特性包括：
//   - 基于寄存器的虚拟机代码生成
//   - 线性扫描寄存器分配
//   - 词法作用域管理（块级作用域、函数作用域）
//   - 闭包支持（Upvalue 捕获）
//   - 函数重载（按参数数量）
//   - 全局变量索引优化（编译期分配）
//   - global/static 关键字支持
//
// 编译流程：
//  1. NewCompiler() 创建编译器
//  2. 遍历 AST 节点，调用对应的 compileXXX 方法
//  3. emit 指令到字节码缓冲区
//  4. 生成 CompiledFunction
//  5. Compile() / CompileString() 完成编译并返回 Program
//
// 寄存器分配策略：
//   - 表达式求值使用临时寄存器
//   - 变量存储使用专用寄存器
//   - 函数调用时保存/恢复寄存器窗口
//
// 线程安全：Compiler 不是线程安全的，同一时间只能编译一个程序。
type Compiler struct {
	parent *Compiler // 父编译器（用于嵌套函数编译）

	// 编译输出
	bytecode  []Instruction // 生成的字节码
	constants []Value       // 常量池

	// 寄存器管理
	nextReg int // 下一个可用寄存器编号
	maxReg  int // 峰值寄存器数（用于计算 maxStackSize）

	// 作用域管理
	scopes      []map[string]*Symbol // 作用域栈（0 为全局作用域）
	scopeDepth  int                  // 当前作用域深度（0 = 全局）
	globalScope bool                 // 是否为全局作用域

	// 函数信息
	funcName   string        // 当前函数名（调试用）
	funcParams int           // 参数数量
	upvals     []UpvalueDesc // Upvalue 描述列表

	// 循环上下文栈（支持 break/continue）
	loops []*loopContext

	// 所有编译的函数（包括嵌套函数）
	functions []*CompiledFunction

	// global 声明跟踪（PHP 风格的全局变量）
	globalVars map[string]bool // 标记为 global 的变量名

	// static 声明跟踪（持久化变量）
	staticVars map[string]*staticVarInfo // 标记为 static 的变量

	// 全局变量索引分配（P1-2 优化：编译期确定索引，运行时 O(1) 访问）
	globalIndices map[string]int // 全局变量名 -> 数组索引
	globalNames   *[]string      // 指针，确保所有子编译器共享同一个底层数组

	// Phase 7.8: 魔术常量支持
	filename    string // 源文件名（用于 __FILE__）
	dirname     string // 源文件目录（用于 __DIR__）
	compileTime string // 编译时间（用于 __TIME__）
	compileDate string // 编译日期（用于 __DATE__）

	// 对象字面量上下文（用于 @member 语法）
	objectDepth   int             // 当前对象字面量嵌套深度
	objectSelfSym map[int]*Symbol // 每层对象字面量的 self 符号 (depth -> symbol)

	// 源码行号追踪（用于运行时错误定位）
	sourceLines []int // 每条指令对应的源码行号
	currentLine int   // 当前正在编译的源码行号
}

// NewCompiler 创建新的字节码编译器实例。
//
// 此方法初始化编译器，准备进行 AST 到字节码的转换。
// 编译器默认处于全局作用域，等待编译程序入口（main 函数）。
//
// 返回值：
//   - *Compiler: 新创建的编译器实例
//
// 使用示例：
//
//	// 创建编译器
//	c := engine.NewCompiler()
//
//	// 解析源代码为 AST
//	program, err := parser.Parse(source)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 编译 AST
//	compiled, err := engine.Compile(program)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// NewCompiler 创建新的编译器实例。
//
// 编译器用于将 AST 转换为字节码。每次调用创建一个独立的编译器，
// 可以编译多个程序。编译器本身不是线程安全的。
func NewCompiler() *Compiler {
	globalNames := make([]string, 0, 256)
	c := &Compiler{
		scopes:        []map[string]*Symbol{make(map[string]*Symbol)},
		globalScope:   true,
		globalVars:    make(map[string]bool),
		staticVars:    make(map[string]*staticVarInfo),
		globalIndices: make(map[string]int),
		globalNames:   &globalNames,
		objectSelfSym: make(map[int]*Symbol),
	}
	return c
}

// NewCompilerWithGlobals 使用已存在的全局变量名创建编译器（用于 REPL 等场景）。
//
// 此方法创建的编译器会继承现有的全局变量命名空间，确保同一 VM 中
// 多次编译的代码能正确访问和修改已存在的全局变量。
//
// 参数：
//   - existingGlobals: 已存在的全局变量名列表，索引对应 VM 中 globals 切片的索引
//
// 返回值：
//   - *Compiler: 配置好的编译器实例
//
// 使用示例（REPL 场景）：
//
//	// 第一轮编译创建全局变量
//	prog1, _ := engine.CompileString("a = 10")
//	vm.Execute()
//
//	// 第二轮编译需要知道 a 的索引是 0
//	compiler := engine.NewCompilerWithGlobals([]string{"a"})
//	prog2, _ := compiler.CompileSource("b = a + 5")
func NewCompilerWithGlobals(existingGlobals []string) *Compiler {
	// 创建共享的 globalNames 切片，预填充已存在的名称
	// 容量至少为 256 或现有全局变量数量的较大值
	capacity := len(existingGlobals)
	if capacity < 256 {
		capacity = 256
	}
	globalNames := make([]string, len(existingGlobals), capacity)
	copy(globalNames, existingGlobals)

	// 构建 globalIndices 映射
	globalIndices := make(map[string]int, len(existingGlobals))
	for i, name := range existingGlobals {
		globalIndices[name] = i
	}

	c := &Compiler{
		scopes:        []map[string]*Symbol{make(map[string]*Symbol)},
		globalScope:   true,
		globalVars:    make(map[string]bool),
		staticVars:    make(map[string]*staticVarInfo),
		globalIndices: globalIndices,
		globalNames:   &globalNames,
		objectSelfSym: make(map[int]*Symbol),
	}
	return c
}

// allocateGlobalIndex 分配全局变量索引（编译期优化）
// 返回现有索引或分配新索引
func (c *Compiler) allocateGlobalIndex(name string) int {
	if idx, ok := c.globalIndices[name]; ok {
		return idx
	}
	idx := len(*c.globalNames)
	c.globalIndices[name] = idx
	*c.globalNames = append(*c.globalNames, name)
	return idx
}

// getGlobalIndex 获取已分配的全局变量索引（如果不存在返回 -1）
func (c *Compiler) getGlobalIndex(name string) int {
	if idx, ok := c.globalIndices[name]; ok {
		return idx
	}
	return -1
}

// staticVarInfo 静态变量信息
type staticVarInfo struct {
	name     string // 变量名
	initIdx  int    // 初始化值的常量索引（-1 表示无初始值）
	initReg  int    // 初始化值所在的寄存器（编译时用）
	hasValue bool   // 是否有初始值
}

// ============================================================================
// 指令发射
// ============================================================================

func (c *Compiler) emit(ins Instruction) int {
	pos := len(c.bytecode)
	c.bytecode = append(c.bytecode, ins)
	// 记录当前源码行号
	c.sourceLines = append(c.sourceLines, c.currentLine)
	return pos
}

func (c *Compiler) emitABC(op Opcode, a, b, c_ int) int {
	return c.emit(NewABC(op, a, b, c_))
}

func (c *Compiler) emitABx(op Opcode, a, bx int) int {
	return c.emit(NewABx(op, a, bx))
}

func (c *Compiler) emitAsBx(op Opcode, a, sbx int) int {
	return c.emit(NewAsBx(op, a, sbx))
}

func (c *Compiler) currentPC() int {
	return len(c.bytecode)
}

func (c *Compiler) patchJump(pos int) {
	offset := c.currentPC() - pos - 1
	ins := c.bytecode[pos]
	c.bytecode[pos] = NewAsBx(ins.OP(), ins.A(), offset)
}

// ============================================================================
// 常量池
// ============================================================================

func (c *Compiler) addConstant(v Value) int {
	for i, k := range c.constants {
		if k.Equals(v) {
			return i
		}
	}
	idx := len(c.constants)
	c.constants = append(c.constants, v)
	return idx
}

// ============================================================================
// 寄存器管理
// ============================================================================

func (c *Compiler) allocReg() int {
	r := c.nextReg
	c.nextReg++
	if c.nextReg > c.maxReg {
		c.maxReg = c.nextReg
	}
	return r
}

func (c *Compiler) freeReg() {
	if c.nextReg > 0 {
		c.nextReg--
	}
}

// ============================================================================
// 作用域管理
// ============================================================================

func (c *Compiler) pushScope() {
	c.scopes = append(c.scopes, make(map[string]*Symbol))
	c.scopeDepth++
}

func (c *Compiler) popScope() {
	if len(c.scopes) <= 1 {
		return
	}
	scope := c.scopes[len(c.scopes)-1]
	for _, sym := range scope {
		if !sym.Captured && sym.Index+1 == c.nextReg {
			c.nextReg--
		}
	}
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeDepth--
}

func (c *Compiler) declareVar(name string) *Symbol {
	sym := &Symbol{
		Name:  name,
		Index: c.nextReg,
		Scope: c.scopeDepth,
	}
	c.scopes[len(c.scopes)-1][name] = sym
	c.allocReg()
	return sym
}

func (c *Compiler) resolveVar(name string) *Symbol {
	// _ 前缀变量：仅在当前（最内层）作用域可访问
	if strings.HasPrefix(name, "_") {
		if sym, ok := c.scopes[len(c.scopes)-1][name]; ok {
			return sym
		}
		return nil
	}
	// 普通变量：沿作用域链向上查找
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if sym, ok := c.scopes[i][name]; ok {
			return sym
		}
	}
	return nil
}

// resolveGlobal 检查变量是否存在于全局作用域（沿父编译器链查找）
func (c *Compiler) resolveGlobal(name string) bool {
	// 先检查当前编译器的全局作用域（scopes[0]）
	if len(c.scopes) > 0 {
		if _, ok := c.scopes[0][name]; ok {
			return true
		}
	}
	// 沿父编译器链查找
	if c.parent != nil {
		return c.parent.resolveGlobal(name)
	}
	return false
}

// resolveUpvalue 解析 upvalue（从父编译器的作用域中捕获变量）
// 返回 upvalue 索引，如果变量不在任何父作用域中返回 -1
func (c *Compiler) resolveUpvalue(name string) int {
	if c.parent == nil {
		return -1
	}

	// 在父编译器的局部变量中查找
	sym := c.parent.resolveVar(name)
	if sym != nil {
		// 标记为被捕获
		sym.Captured = true
		// 添加为 upvalue（直接引用父编译器的局部变量）
		return c.addUpvalue(sym.Index, true)
	}

	// 递归查找更外层的 upvalue
	upvalIdx := c.parent.resolveUpvalue(name)
	if upvalIdx >= 0 {
		// 添加为 upvalue（引用父编译器的 upvalue）
		return c.addUpvalue(upvalIdx, false)
	}

	return -1
}

// addUpvalue 添加 upvalue 到当前函数
func (c *Compiler) addUpvalue(index int, isLocal bool) int {
	// 检查是否已存在相同的 upvalue
	for i, uv := range c.upvals {
		if uv.Index == index && uv.IsLocal == isLocal {
			return i
		}
	}

	// 添加新的 upvalue
	idx := len(c.upvals)
	c.upvals = append(c.upvals, UpvalueDesc{Index: index, IsLocal: isLocal})
	return idx
}

func (c *Compiler) findOrCreateLocal(name string) *Symbol {
	sym := c.resolveVar(name)
	if sym != nil {
		return sym
	}
	return c.declareVar(name)
}

// buildVarNames 从当前作用域构建寄存器→变量名映射
func (c *Compiler) buildVarNames() []string {
	if len(c.scopes) == 0 {
		return nil
	}
	scope := c.scopes[0]
	if len(scope) == 0 {
		return nil
	}
	names := make([]string, c.maxReg)
	for _, sym := range scope {
		if sym.Index >= 0 && sym.Index < len(names) {
			names[sym.Index] = sym.Name
		}
	}
	return names
}

// ============================================================================
// 子编译器（函数编译）
// ============================================================================

func (c *Compiler) newChild(name string) *Compiler {
	child := &Compiler{
		parent:        c,
		scopes:        []map[string]*Symbol{make(map[string]*Symbol)},
		globalScope:   false,
		funcName:      name,
		functions:     c.functions,
		globalVars:    make(map[string]bool),
		staticVars:    make(map[string]*staticVarInfo),
		globalIndices: c.globalIndices, // 共享全局索引映射
		globalNames:   c.globalNames,   // 共享全局名称切片
		objectSelfSym: c.objectSelfSym, // 共享 self 符号映射
		sourceLines:   make([]int, 0),  // 每个函数有自己的行号表
	}
	return child
}

// ============================================================================
// 主编译入口
// ============================================================================

// Compile 将 AST 程序编译为字节码程序。
//
// 这是主要的编译入口，将解析后的 AST 转换为可执行的 Program。
// 编译过程包括：
//   - 创建新编译器实例
//   - 遍历 AST 生成字节码
//   - 构建主函数（<main>）
//   - 收集所有编译的函数
//   - 分配全局变量索引
//
// 返回值：
//   - *Program: 编译后的程序，包含主函数、所有函数、常量池和全局变量名
//   - nil: 编译成功
//   - error: 编译失败（语法错误、语义错误等）
//
// 使用示例：
//
//	// 1. 解析源代码
//	source := `$x = 10 + 20`
//	program, err := parser.Parse(source)
//	if err != nil {
//	    log.Fatal("解析错误:", err)
//	}
//
//	// 2. 编译为字节码
//	compiled, err := engine.Compile(program)
//	if err != nil {
//	    log.Fatal("编译错误:", err)
//	}
//
//	// 3. 创建 VM 并执行
//	vm := engine.NewVMWithProgram(eng, compiled)
//	if err := vm.Execute(); err != nil {
//	    log.Fatal("执行错误:", err)
//	}
func Compile(program *parser.Program) (*Program, error) {
	c := NewCompiler()

	// 捕获编译 panic 并转换为 error
	var compileErr error

	// 使用 recover 捕获 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case *CompileError:
					compileErr = e
				case error:
					compileErr = &CompileError{Message: e.Error()}
				default:
					compileErr = &CompileError{Message: fmt.Sprintf("%v", r)}
				}
			}
		}()
		c.compileProgram(program)
	}()

	if compileErr != nil {
		return nil, compileErr
	}

	main := &CompiledFunction{
		Name:       "<main>",
		Params:     0,
		Registers:  c.maxReg,
		Bytecode:   c.bytecode,
		Constants:  c.constants,
		SourceLine: 0,
		VarNames:   c.buildVarNames(),
	}

	allFuncs := []*CompiledFunction{main}
	allFuncs = append(allFuncs, c.functions...)

	return &Program{
		Main:        main,
		Functions:   allFuncs,
		Constants:   c.constants,
		GlobalNames: *c.globalNames,
	}, nil
}

func (c *Compiler) compileProgram(program *parser.Program) {
	for i, stmt := range program.Statements {
		// 如果是最后一条语句且是表达式语句，将其作为返回值处理
		if i == len(program.Statements)-1 {
			if exprStmt, ok := stmt.(*parser.ExprStmt); ok {
				r := c.allocReg()
				c.compileExprToReg(exprStmt.Expression, r)
				c.emitABC(OP_RETURN, r, 0, 0)
				c.freeReg()
				continue
			}
		}
		c.compileStmt(stmt)
	}

	// 如果主函数没有显式 RETURN，添加隐式 return null
	if len(c.bytecode) == 0 || c.bytecode[len(c.bytecode)-1].OP() != OP_RETURN {
		r := c.allocReg()
		c.emitABC(OP_LOADNULL, r, 0, 0)
		c.emitABC(OP_RETURN, r, 0, 0)
		c.freeReg()
	}
}

// ============================================================================
// 语句编译
// ============================================================================

func (c *Compiler) compileStmt(stmt parser.Statement) {
	// 设置当前源码行号
	c.currentLine = stmt.Pos().Line

	switch s := stmt.(type) {
	case *parser.VarDecl:
		c.compileVarDecl(s)
	case *parser.ConstDecl:
		c.compileConstDecl(s)
	case *parser.ExprStmt:
		c.compileExprStmt(s)
	case *parser.BlockStmt:
		c.compileBlockStmt(s)
	case *parser.IfStmt:
		c.compileIfStmt(s)
	case *parser.WhileStmt:
		c.compileWhileStmt(s)
	case *parser.ForStmt:
		c.compileForStmt(s)
	case *parser.ForeachStmt:
		c.compileForeachStmt(s)
	case *parser.ReturnStmt:
		c.compileReturnStmt(s)
	case *parser.BreakStmt:
		c.compileBreakStmt()
	case *parser.ContinueStmt:
		c.compileContinueStmt()
	case *parser.FuncDecl:
		c.compileFuncDecl(s)
	case *parser.ImportStmt:
		c.compileImportStmt(s)
	case *parser.IncludeStmt:
		c.compileIncludeStmt(s)
	case *parser.TryCatchStmt:
		c.compileTryCatchStmt(s)
	case *parser.ThrowStmt:
		c.compileThrowStmt(s)
	case *parser.GlobalDecl:
		c.compileGlobalDecl(s)
	case *parser.StaticDecl:
		c.compileStaticDecl(s)
	case *parser.MatchStmt:
		c.compileMatchStmt(s)
	default:
		// 未知语句类型，跳过
	}
}

func (c *Compiler) compileVarDecl(decl *parser.VarDecl) {
	name := decl.Name.Value

	// 检查是否为 global 声明的变量
	if c.isGlobalVar(name) {
		if decl.Value != nil {
			r := c.allocReg()
			c.compileExprToReg(decl.Value, r)
			gIdx := c.allocateGlobalIndex(name)
			c.emitABx(OP_SETGLOBAL, r, gIdx)
			c.freeReg()
		}
		return
	}

	if decl.Value != nil {
		// $name = expr

		// 检查是否为 static 声明的变量
		if c.isStaticVar(name) {
			r := c.allocReg()
			c.compileExprToReg(decl.Value, r)
			gIdx := c.allocateGlobalIndex(c.staticStorageKey(name))
			c.emitABx(OP_SETGLOBAL, r, gIdx)
			c.freeReg()
			return
		}

		sym := c.resolveVar(name)
		if sym == nil {
			// 先检查是否为 upvalue
			upvalIdx := c.resolveUpvalue(name)
			if upvalIdx >= 0 {
				r := c.allocReg()
				c.compileExprToReg(decl.Value, r)
				c.emitABC(OP_SETUPVAL, upvalIdx, r, 0)
				c.freeReg()
				return
			}
			// 检查是否为全局变量（沿父编译器链查找）
			if !c.globalScope && c.resolveGlobal(name) {
				// 函数内修改全局变量：使用 SETGLOBAL
				r := c.allocReg()
				c.compileExprToReg(decl.Value, r)
				gIdx := c.allocateGlobalIndex(name)
				c.emitABx(OP_SETGLOBAL, r, gIdx)
				c.freeReg()
				return
			}
			// 新变量：声明为局部变量
			sym = c.declareVar(name)
		}
		c.compileExprToReg(decl.Value, sym.Index)
		// 全局作用域：同时写入 globals map，供函数内 OP_GETGLOBAL 访问
		if c.globalScope {
			gIdx := c.allocateGlobalIndex(name)
			c.emitABx(OP_SETGLOBAL, sym.Index, gIdx)
		}
	} else {
		// $name （无初始值）
		if c.resolveVar(name) == nil {
			// 先检查是否为 upvalue
			upvalIdx := c.resolveUpvalue(name)
			if upvalIdx >= 0 {
				return
			}
			// 检查是否为全局变量
			if !c.globalScope && c.resolveGlobal(name) {
				return
			}
			sym := c.declareVar(name)
			c.emitABC(OP_LOADNULL, sym.Index, 0, 0)
			if c.globalScope {
				gIdx := c.allocateGlobalIndex(name)
				c.emitABx(OP_SETGLOBAL, sym.Index, gIdx)
			}
		}
	}
}

func (c *Compiler) compileConstDecl(decl *parser.ConstDecl) {
	name := decl.Name.Value
	sym := c.declareVar(name)
	c.compileExprToReg(decl.Value, sym.Index)
}

func (c *Compiler) compileExprStmt(stmt *parser.ExprStmt) {
	r := c.allocReg()
	c.compileExprToReg(stmt.Expression, r)
	c.freeReg()
}

func (c *Compiler) compileBlockStmt(block *parser.BlockStmt) {
	c.pushScope()
	for _, stmt := range block.Statements {
		c.compileStmt(stmt)
	}
	c.popScope()
}

func (c *Compiler) compileIfStmt(stmt *parser.IfStmt) {
	// 编译条件
	condReg := c.allocReg()
	c.compileExprToReg(stmt.Cond, condReg)

	// 条件跳转（跳过 if 体）
	jumpIfNot := c.currentPC()
	c.emitAsBx(OP_JMPIFNOT, condReg, 0)
	c.freeReg()

	// if 体
	c.pushScope()
	c.compileBlockStmtWithoutScope(stmt.Body)
	c.popScope()

	if stmt.Else != nil {
		// 有 else 分支：if 体结束跳过 else
		jumpOverElse := c.currentPC()
		c.emitAsBx(OP_JMP, 0, 0)

		// 修补条件跳转
		c.patchJump(jumpIfNot)

		// else 体
		c.pushScope()
		if block, ok := stmt.Else.(*parser.BlockStmt); ok {
			c.compileBlockStmtWithoutScope(block)
		} else {
			c.compileStmt(stmt.Else)
		}
		c.popScope()

		// 修补跳过 else 的跳转
		c.patchJump(jumpOverElse)
	} else {
		// 无 else 分支
		c.patchJump(jumpIfNot)
	}
}

func (c *Compiler) compileWhileStmt(stmt *parser.WhileStmt) {
	// 条件开始位置
	condPC := c.currentPC()

	// 编译条件
	condReg := c.allocReg()
	c.compileExprToReg(stmt.Cond, condReg)

	// 条件跳转（跳过循环体）
	jumpIfNot := c.currentPC()
	c.emitAsBx(OP_JMPIFNOT, condReg, 0)
	c.freeReg()

	// 循环体
	loop := &loopContext{
		continuePC: condPC,
	}
	c.loops = append(c.loops, loop)

	c.pushScope()
	c.compileBlockStmtWithoutScope(stmt.Body)
	c.popScope()

	// 跳回条件检查
	c.emitAsBx(OP_JMP, 0, condPC-c.currentPC()-1)

	// 修补条件跳转
	c.patchJump(jumpIfNot)

	// 修补 break/continue 跳转
	for _, pos := range loop.breakJumps {
		c.patchJumpTo(pos, c.currentPC())
	}
	for _, pos := range loop.continueJumps {
		c.patchJumpTo(pos, condPC)
	}

	c.loops = c.loops[:len(c.loops)-1]
}

func (c *Compiler) compileForStmt(stmt *parser.ForStmt) {
	c.pushScope()

	// 初始化
	if stmt.Init != nil {
		c.compileStmt(stmt.Init)
	}

	// 条件开始位置
	condPC := c.currentPC()

	// 编译条件
	var jumpIfNot int
	if stmt.Cond != nil {
		condReg := c.allocReg()
		c.compileExprToReg(stmt.Cond, condReg)
		jumpIfNot = c.currentPC()
		c.emitAsBx(OP_JMPIFNOT, condReg, 0)
		c.freeReg()
	}

	// 循环体
	loop := &loopContext{}
	c.loops = append(c.loops, loop)

	c.pushScope()
	c.compileBlockStmtWithoutScope(stmt.Body)
	c.popScope()

	// 后置表达式
	loop.continuePC = c.currentPC() // 设置 continue 跳转到后置表达式开始位置
	if stmt.Post != nil {
		r := c.allocReg()
		c.compileExprToReg(stmt.Post, r)
		c.freeReg()
	}

	// 跳回条件检查
	c.emitAsBx(OP_JMP, 0, condPC-c.currentPC()-1)

	// 修补条件跳转
	if stmt.Cond != nil {
		c.patchJump(jumpIfNot)
	}

	// 修补 break/continue 跳转
	for _, pos := range loop.breakJumps {
		c.patchJumpTo(pos, c.currentPC())
	}
	for _, pos := range loop.continueJumps {
		c.patchJumpTo(pos, loop.continuePC)
	}

	c.loops = c.loops[:len(c.loops)-1]
	c.popScope()
}

// compileForeachStmt 编译 foreach 语句
// 支持两种语法：
//  1. foreach ($value in $array) { ... }           // 只遍历值
//  2. foreach ($key => $value in $array) { ... }   // 遍历键值对
func (c *Compiler) compileForeachStmt(stmt *parser.ForeachStmt) {
	c.pushScope()

	// 编译被遍历的数组/对象表达式
	arrReg := c.allocReg()
	c.compileExprToReg(stmt.Array, arrReg)

	// 初始化迭代器：OP_ITERINIT 迭代器寄存器, 数组寄存器
	iterReg := c.allocReg() // 迭代器状态寄存器
	c.emitABC(OP_ITERINIT, iterReg, arrReg, 0)
	// 注意：不释放 arrReg，避免与后续 declareVar 分配的寄存器冲突
	// iterReg 需要在整个循环中保持有效

	// 循环开始位置（ITERNEXT 指令位置）
	loopStart := c.currentPC()

	// 声明键和值变量
	keyReg := -1
	valReg := -1

	if stmt.Key != nil {
		keySym := c.declareVar(stmt.Key.Value)
		keyReg = keySym.Index
	}

	// 声明值变量
	valSym := c.declareVar(stmt.Value.Value)
	valReg = valSym.Index

	// OP_ITERNEXT：获取下一个键值对
	// 如果迭代结束，跳转到循环结束位置（通过跳过中间的 JMP 实现）
	if keyReg >= 0 {
		c.emitABC(OP_ITERNEXT, iterReg, keyReg, valReg)
	} else {
		// 只需要值，不需要键，用 255 表示无键寄存器
		c.emitABC(OP_ITERNEXT, iterReg, 255, valReg)
	}

	// ITERNEXT 后的 JMP：迭代结束时执行（跳出循环）
	// 正常迭代时会跳过这条 JMP，继续执行循环体
	exitJmp := c.currentPC()
	c.emitAsBx(OP_JMP, 0, 0) // 占位，后面修补

	// 循环上下文（用于 break/continue）
	loop := &loopContext{
		continuePC: loopStart,
	}
	c.loops = append(c.loops, loop)

	// 编译循环体
	c.compileBlockStmtWithoutScope(stmt.Body)

	// 跳回 ITERNEXT（继续下一次迭代）
	c.emitAsBx(OP_JMP, 0, loopStart-c.currentPC()-1)

	// 循环结束位置
	loopEnd := c.currentPC()

	// 修补退出跳转（迭代结束时跳出循环）
	c.patchJumpTo(exitJmp, loopEnd)

	// 修补 break 跳转（跳转到循环结束）
	for _, pos := range loop.breakJumps {
		c.patchJumpTo(pos, loopEnd)
	}

	// 修补 continue 跳转（跳回 ITERNEXT）
	for _, pos := range loop.continueJumps {
		c.patchJumpTo(pos, loopStart)
	}

	// 清理迭代器（可选，对于数组/对象可能不需要显式关闭）
	// c.emitABC(OP_ITEREND, iterReg, 0, 0)

	c.loops = c.loops[:len(c.loops)-1]
	c.freeReg() // 释放迭代器寄存器
	c.popScope()
}

// isValidForeachTarget 检查是否是有效的 foreach 省略语法目标类型
// 编译时只能检查字面量类型，变量引用和函数调用需要运行时检查
func isValidForeachTarget(expr parser.Expression) bool {
	switch expr.(type) {
	case *parser.ArrayLiteral:
		return true
	case *parser.RangeExpr:
		return true
	case *parser.Identifier:
		return true // 变量引用，运行时检查
	case *parser.CallExpr:
		return true // 函数调用，运行时检查
	case *parser.IndexExpr:
		return true // 索引访问，运行时检查
	case *parser.StringLiteral:
		return false // 字符串不支持（可考虑未来支持字符遍历）
	case *parser.ObjectLiteral:
		return false // 对象不支持
	default:
		return false
	}
}

func (c *Compiler) compileReturnStmt(stmt *parser.ReturnStmt) {
	if stmt.Value != nil {
		// 检查是否为尾调用（return func(args)）
		if call, ok := stmt.Value.(*parser.CallExpr); ok {
			// 尾调用：编译调用并发出 RETURN
			callee := c.allocReg()
			c.compileExprToReg(call.Function, callee)

			for i, arg := range call.Arguments {
				argReg := callee + 1 + i
				if argReg >= c.nextReg {
					c.allocReg()
				}
				c.compileExprToReg(arg, argReg)
			}

			c.emitABC(OP_CALL, callee, 0, len(call.Arguments)+1)
			c.emitABC(OP_RETURN, callee, 0, 0)

			for i := 0; i < len(call.Arguments); i++ {
				c.freeReg()
			}
			c.freeReg() // callee
			return
		}
		r := c.allocReg()
		c.compileExprToReg(stmt.Value, r)
		c.emitABC(OP_RETURN, r, 0, 0)
		c.freeReg()
	} else {
		r := c.allocReg()
		c.emitABC(OP_LOADNULL, r, 0, 0)
		c.emitABC(OP_RETURN, r, 0, 0)
		c.freeReg()
	}
}

func (c *Compiler) compileBreakStmt() {
	if len(c.loops) == 0 {
		return // 错误：break 不在循环内
	}
	pos := c.emitAsBx(OP_JMP, 0, 0)
	c.loops[len(c.loops)-1].breakJumps = append(c.loops[len(c.loops)-1].breakJumps, pos)
}

func (c *Compiler) compileContinueStmt() {
	if len(c.loops) == 0 {
		return // 错误：continue 不在循环内
	}
	pos := c.emitAsBx(OP_JMP, 0, 0)
	c.loops[len(c.loops)-1].continueJumps = append(c.loops[len(c.loops)-1].continueJumps, pos)
}

func (c *Compiler) compileTryCatchStmt(stmt *parser.TryCatchStmt) {
	if len(stmt.CatchClauses) == 0 {
		return
	}

	// 为第一个 catch 分支声明变量，所有分支复用同一个寄存器
	catchVarReg := c.nextReg
	catchSym := c.declareVar(stmt.CatchClauses[0].CatchVar.Value)

	// OP_TRY_BEGIN sBx, catchVarReg: sBx = 第一个 catch 块跳转偏移
	tryBeginPos := c.currentPC()
	c.emitAsBx(OP_TRY_BEGIN, catchVarReg, 0)

	// 编译 try 块
	c.pushScope()
	c.compileBlockStmtWithoutScope(stmt.TryBody)
	c.popScope()

	// OP_TRY_END: try 块结束
	c.emitABC(OP_TRY_END, 0, 0, 0)

	// 跳过所有 catch 块
	jumpOverAllCatch := c.currentPC()
	c.emitAsBx(OP_JMP, 0, 0)

	// 修补 OP_TRY_BEGIN 的跳转偏移到第一个 catch 块
	c.patchJump(tryBeginPos)

	// 编译所有 catch 块
	var endJumpPositions []int
	var skipToNextJump int = -1 // 条件不满足时跳到下一个 catch 的位置

	for i, clause := range stmt.CatchClauses {
		// 修补跳到当前 catch 块的跳转
		if skipToNextJump >= 0 {
			c.patchJumpTo(skipToNextJump, c.currentPC())
			skipToNextJump = -1
		}

		// 后续 catch 分支在作用域中注册变量，但复用第一个变量的寄存器
		if i > 0 {
			// 在当前作用域中注册 catch 变量，指向第一个变量的寄存器
			c.scopes[len(c.scopes)-1][clause.CatchVar.Value] = catchSym
		}

		c.pushScope()

		// 条件捕获：如果条件为 false 且还有下一个 catch，则跳到下一个 catch
		if clause.Condition != nil {
			condReg := c.allocReg()
			c.compileExprToReg(clause.Condition, condReg)

			// 条件为真时跳过 re-throw/跳到下一个 catch
			skipAction := c.currentPC()
			c.emitAsBx(OP_JMPIF, condReg, 0)
			c.freeReg()

			if i < len(stmt.CatchClauses)-1 {
				// 还有下一个 catch 分支，条件不满足时跳到下一个
				skipToNextJump = c.currentPC()
				c.emitAsBx(OP_JMP, 0, 0)
			} else {
				// 最后一个 catch 分支，条件不满足时 re-throw
				c.emitABC(OP_THROW, catchVarReg, 0, 0)
			}
			// 修补条件跳转（条件为真时执行当前 catch 体）
			c.patchJump(skipAction)
		} else if i < len(stmt.CatchClauses)-1 {
			// 无条件 catch 但不是最后一个
			// 无条件 catch 会捕获所有异常，所以不会执行到后续 catch
			// 直接执行当前 catch 体
		}

		c.compileBlockStmtWithoutScope(clause.Body)
		c.popScope()

		// 跳过后续的 catch 块
		if i < len(stmt.CatchClauses)-1 {
			endJumpPositions = append(endJumpPositions, c.currentPC())
			c.emitAsBx(OP_JMP, 0, 0)
		}
	}

	// 修补所有跳过后续 catch 块的跳转
	for _, pos := range endJumpPositions {
		c.patchJump(pos)
	}

	// 修补跳过所有 catch 块的跳转
	c.patchJump(jumpOverAllCatch)
}

func (c *Compiler) compileMatchExpr(expr *parser.MatchStmt, target int) {
	expr.IsExpr = true
	matchReg := c.allocReg()
	c.compileExprToReg(expr.Value, matchReg)

	var endJumpPositions []int
	resultReg := target

	c.emitABC(OP_LOADNULL, resultReg, 0, 0)

	for _, mc := range expr.Cases {
		c.compileMatchCase(mc, matchReg, true, resultReg, &endJumpPositions)
	}

	for _, pos := range endJumpPositions {
		c.patchJump(pos)
	}

	c.freeReg()
	c.freeReg()
}

func (c *Compiler) compileMatchStmt(stmt *parser.MatchStmt) {
	matchReg := c.allocReg()
	c.compileExprToReg(stmt.Value, matchReg)

	var endJumpPositions []int

	for _, mc := range stmt.Cases {
		c.compileMatchCase(mc, matchReg, false, 0, &endJumpPositions)
	}

	for _, pos := range endJumpPositions {
		c.patchJump(pos)
	}

	c.freeReg()
}

func (c *Compiler) compileMatchCase(mc *parser.MatchCase, matchReg int, isExpr bool, resultReg int, endJumpPositions *[]int) {
	pattern := mc.Pattern
	caseBody := mc.Body

	c.pushScope()

	var matchResultReg int
	var orPatternJumps []int
	var orPatternJumpToNextCase int

	switch pat := pattern.(type) {
	case *parser.WildcardPattern:
	case *parser.IdentifierPattern:
		sym := c.declareVar(pat.Name.Value)
		c.emitABC(OP_LOAD, sym.Index, matchReg, 0)
		matchResultReg = -1
	case *parser.LiteralPattern:
		matchResultReg = c.compilePatternCompare(pat.Value, matchReg)
	case *parser.OrPattern:
		orPatternJumps, orPatternJumpToNextCase = c.compileOrPattern(pat.Patterns, matchReg)
	case *parser.RangePattern:
		matchResultReg = c.compileRangePattern(pat, matchReg)
	case *parser.RegexPattern:
		matchResultReg = c.compileRegexPattern(pat, matchReg)
	case *parser.ArrayPattern:
		matchResultReg = c.compilePatternCompare(&parser.ArrayLiteral{Token: pat.Token, Elements: c.patternElementsToExprs(pat.Elements)}, matchReg)
	case *parser.ObjectPattern:
		matchResultReg = c.compilePatternCompare(&parser.ObjectLiteral{Token: pat.Token, Pairs: c.patternPairsToExprPairs(pat.Pairs)}, matchReg)
	default:
		matchResultReg = -1
	}

	var notMatchJumps []int
	if _, ok := pattern.(*parser.WildcardPattern); !ok {
		if orPatternJumps != nil {
			if orPatternJumpToNextCase > 0 {
				notMatchJumps = append(notMatchJumps, orPatternJumpToNextCase)
			}
		} else if matchResultReg >= 0 {
			notMatchJumps = append(notMatchJumps, c.currentPC())
			c.emitAsBx(OP_JMPIFNOT, matchResultReg, 0)
		} else if _, ok := pattern.(*parser.IdentifierPattern); !ok {
			notMatchJumps = append(notMatchJumps, c.currentPC())
			c.emitAsBx(OP_JMPIFNOT, matchReg, 0)
		}
	}

	if mc.Guard != nil {
		guardReg := c.allocReg()
		c.compileExprToReg(mc.Guard, guardReg)

		notMatchJumps = append(notMatchJumps, c.currentPC())
		c.emitAsBx(OP_JMPIFNOT, guardReg, 0)
		c.freeReg()
	}

	bodyStartPos := c.currentPC()

	if isExpr {
		c.compileExprToReg(getExprFromStmt(caseBody), resultReg)
	} else {
		c.compileStmt(caseBody)
	}

	jumpToEnd := c.currentPC()
	c.emitAsBx(OP_JMP, 0, 0)
	*endJumpPositions = append(*endJumpPositions, jumpToEnd)

	for _, pos := range notMatchJumps {
		c.patchJump(pos)
	}

	for _, pos := range orPatternJumps {
		c.patchJumpTo(pos, bodyStartPos)
	}

	c.popScope()
}

func (c *Compiler) compilePatternCompare(pattern parser.Expression, matchReg int) int {
	tempReg := c.allocReg()
	c.compileExprToReg(pattern, tempReg)
	resultReg := c.allocReg()
	c.emitABC(OP_EQ, resultReg, matchReg, tempReg)
	c.freeReg()
	c.freeReg()
	return resultReg
}

func (c *Compiler) compileOrPatternCompare(pattern parser.Expression, matchReg int) (int, int) {
	tempReg := c.allocReg()
	c.compileExprToReg(pattern, tempReg)
	resultReg := c.allocReg()
	c.emitABC(OP_EQ, resultReg, matchReg, tempReg)
	return resultReg, tempReg
}

func (c *Compiler) compileOrPattern(patterns []parser.Pattern, matchReg int) ([]int, int) {
	if len(patterns) == 0 {
		return nil, -1
	}

	jumpToBodyPositions := make([]int, 0, len(patterns))
	var jumpToNextCasePos int

	for i, subPat := range patterns {
		isLast := i == len(patterns)-1
		switch sp := subPat.(type) {
		case *parser.LiteralPattern:
			tmpMatch := c.allocReg()
			tmpLit := c.allocReg()
			resultReg := c.allocReg()
			c.emitABC(OP_LOAD, tmpMatch, matchReg, 0)
			c.compileExprToReg(sp.Value, tmpLit)
			c.emitABC(OP_EQ, resultReg, tmpMatch, tmpLit)
			jumpToBodyPositions = append(jumpToBodyPositions, c.currentPC())
			c.emitAsBx(OP_JMPIF, resultReg, 0)
		case *parser.IdentifierPattern:
			sym := c.declareVar(sp.Name.Value)
			c.emitABC(OP_LOAD, sym.Index, matchReg, 0)
			jumpToBodyPositions = append(jumpToBodyPositions, c.currentPC())
			c.emitAsBx(OP_JMPIF, matchReg, 0)
		}
		if !isLast {
			jumpToNextCasePos = c.currentPC()
			c.emitAsBx(OP_JMP, 0, 0)
		} else {
			jumpToNextCasePos = c.currentPC()
			c.emitAsBx(OP_JMP, 0, 0)
		}
	}
	return jumpToBodyPositions, jumpToNextCasePos
}

func (c *Compiler) compileRangePattern(pat *parser.RangePattern, matchReg int) int {
	startReg := c.allocReg()
	c.compileExprToReg(pat.Start, startReg)

	endReg := c.allocReg()
	c.compileExprToReg(pat.End, endReg)

	resultReg := c.allocReg()
	geReg := c.allocReg()

	c.emitABC(OP_GTE, geReg, matchReg, startReg)

	if pat.Inclusive {
		leReg := c.allocReg()
		c.emitABC(OP_LTE, leReg, matchReg, endReg)
		c.emitABC(OP_AND, resultReg, geReg, leReg)
		c.freeReg()
	} else {
		ltReg := c.allocReg()
		c.emitABC(OP_LT, ltReg, matchReg, endReg)
		c.emitABC(OP_AND, resultReg, geReg, ltReg)
		c.freeReg()
	}

	c.freeReg()
	c.freeReg()

	return resultReg
}

// compileRegexPattern 编译正则模式匹配（match/case 中）
//
// 逻辑：
//  1. 编译正则字面量为常量
//  2. 用 OP_REGEX_MATCH 检查是否匹配
//  3. 如果有 as $var 绑定，匹配成功后提取捕获组并绑定到变量
func (c *Compiler) compileRegexPattern(pat *parser.RegexPattern, matchReg int) int {
	// 1. 创建正则常量
	regexVal, err := NewRegex(pat.Pattern, pat.Flags)
	if err != nil {
		// 已在 lexer 验证，不应出错
		resultReg := c.allocReg()
		c.emitABC(OP_LOADBOOL, resultReg, 0, 0)
		return resultReg
	}
	regexConstIdx := c.addConstant(regexVal)

	// 2. 加载正则到寄存器
	regexReg := c.allocReg()
	c.emitABx(OP_LOADK, regexReg, regexConstIdx)

	// 3. 执行匹配
	resultReg := c.allocReg()
	c.emitABC(OP_REGEX_MATCH, resultReg, matchReg, regexReg)

	// 4. 如果有 as $var 绑定，需要在匹配成功后提取捕获组
	if pat.Binding != nil {
		// 生成条件跳转：匹配失败跳过绑定代码
		jumpOverBinding := c.currentPC()
		c.emitAsBx(OP_JMPIFNOT, resultReg, 0)

		// 匹配成功：调用内部函数提取捕获组
		// re_groups_raw(pattern, subject) → object
		callReg := c.allocReg() // 函数位置
		arg1Reg := c.allocReg() // 正则对象
		arg2Reg := c.allocReg() // subject 字符串

		// 通过全局索引加载 re_groups_raw 函数
		gIdx := c.allocateGlobalIndex("re_groups_raw")
		c.emitABx(OP_GETGLOBAL, callReg, gIdx)
		c.emitABC(OP_LOAD, arg1Reg, regexReg, 0)
		c.emitABC(OP_LOAD, arg2Reg, matchReg, 0)
		c.emitABC(OP_CALL, callReg, 0, 3) // 3 = 1(func) + 2(args)

		// 将结果绑定到变量
		sym := c.declareVar(pat.Binding.Value)
		c.emitABC(OP_LOAD, sym.Index, callReg, 0)

		c.freeReg() // arg2Reg
		c.freeReg() // arg1Reg
		c.freeReg() // callReg

		// 修补跳转：匹配失败时跳到这里
		c.patchJump(jumpOverBinding)
	}

	c.freeReg() // regexReg
	return resultReg
}

func getExprFromStmt(stmt parser.Statement) parser.Expression {
	if es, ok := stmt.(*parser.ExprStmt); ok {
		return es.Expression
	}
	if bs, ok := stmt.(*parser.BlockStmt); ok {
		if len(bs.Statements) > 0 {
			return getExprFromStmt(bs.Statements[len(bs.Statements)-1])
		}
	}
	return &parser.NullLiteral{}
}

func (c *Compiler) patternElementsToExprs(patterns []parser.Pattern) []parser.Expression {
	exprs := make([]parser.Expression, len(patterns))
	for i, p := range patterns {
		switch lit := p.(type) {
		case *parser.LiteralPattern:
			exprs[i] = lit.Value
		default:
			exprs[i] = &parser.NullLiteral{}
		}
	}
	return exprs
}

func (c *Compiler) patternPairsToExprPairs(patPairs map[string]parser.Pattern) map[parser.Expression]parser.Expression {
	exprPairs := make(map[parser.Expression]parser.Expression)
	for k, p := range patPairs {
		key := &parser.StringLiteral{Token: token.Token{Literal: k}, Value: k}
		switch lit := p.(type) {
		case *parser.LiteralPattern:
			exprPairs[key] = lit.Value
		default:
			exprPairs[key] = &parser.NullLiteral{}
		}
	}
	return exprPairs
}

func (c *Compiler) compileThrowStmt(stmt *parser.ThrowStmt) {
	r := c.allocReg()
	c.compileExprToReg(stmt.Value, r)
	c.emitABC(OP_THROW, r, 0, 0)
	c.freeReg()
}

// compileGlobalDecl 编译 global 声明
// 标记变量为全局变量，后续访问该变量时使用 GETGLOBAL/SETGLOBAL
func (c *Compiler) compileGlobalDecl(stmt *parser.GlobalDecl) {
	for _, name := range stmt.Names {
		// 检查是否为私有变量（_ 前缀），私有变量不能声明为全局
		if strings.HasPrefix(name.Value, "_") {
			panic(&CompileError{
				Message: fmt.Sprintf("私有变量 '%s' 不能声明为全局变量（_ 前缀变量只能在当前作用域访问）", name.Value),
				Line:    name.Pos().Line,
				Column:  name.Pos().Column,
			})
		}
		// 标记为全局变量
		c.globalVars[name.Value] = true
		// 在当前作用域声明该变量（指向全局）
		c.declareVar(name.Value)
	}
}

// compileStaticDecl 编译 static 声明
// 静态变量在函数调用之间保持其值
// 实现方式：静态变量存储在 VM 的 globals 中，使用 _static: 前缀
// 每次读写都通过 GETGLOBAL/SETGLOBAL 操作
func (c *Compiler) compileStaticDecl(stmt *parser.StaticDecl) {
	name := stmt.Name.Value
	key := c.funcName + "::" + name
	storageKey := "_static:" + key

	// 声明局部变量（用于后续引用时知道这是一个变量名）
	sym := c.declareVar(name)
	gIdx := c.allocateGlobalIndex(storageKey)

	if stmt.Value != nil {
		// 有初始值：检查是否已初始化
		loadReg := c.allocReg()
		c.emitABx(OP_GETGLOBAL, loadReg, gIdx)

		// 如果非 null，跳过初始化
		jumpOverInit := c.currentPC()
		c.emitAsBx(OP_JMPIF, loadReg, 0)

		// 初始化：计算初始值并存储
		c.compileExprToReg(stmt.Value, loadReg)
		c.emitABx(OP_SETGLOBAL, loadReg, gIdx)

		c.patchJump(jumpOverInit)

		// 将当前值加载到局部变量寄存器
		c.emitABC(OP_LOAD, sym.Index, loadReg, 0)
		c.freeReg()
	} else {
		// 无初始值：检查是否已初始化
		loadReg := c.allocReg()
		c.emitABx(OP_GETGLOBAL, loadReg, gIdx)

		jumpOverInit := c.currentPC()
		c.emitAsBx(OP_JMPIF, loadReg, 0)

		// 首次初始化为 null
		c.emitABC(OP_LOADNULL, loadReg, 0, 0)
		c.emitABx(OP_SETGLOBAL, loadReg, gIdx)

		c.patchJump(jumpOverInit)
		c.emitABC(OP_LOAD, sym.Index, loadReg, 0)
		c.freeReg()
	}

	// 记录静态变量信息
	c.staticVars[key] = &staticVarInfo{
		name:     name,
		hasValue: stmt.Value != nil,
		initReg:  sym.Index,
	}
}

func (c *Compiler) compileImportStmt(stmt *parser.ImportStmt) {
	// 将源路径作为常量
	srcIdx := c.addConstant(NewString(stmt.Source))

	if len(stmt.Names) > 0 {
		// from "x" import a, b, c — 选择性导入
		names := make([]Value, len(stmt.Names))
		for i, n := range stmt.Names {
			names[i] = NewString(n.Value)
		}
		namesIdx := c.addConstant(NewArray(names))
		// A=names常量索引（数组类型），Bx=source常量索引
		c.emitABx(OP_IMPORT, namesIdx, srcIdx)
	} else if stmt.Alias != nil {
		// import "x" as foo — 带别名的命名空间导入
		aliasIdx := c.addConstant(NewString(stmt.Alias.Value))
		// A=alias常量索引（字符串类型），Bx=source常量索引
		c.emitABx(OP_IMPORT, aliasIdx, srcIdx)
	} else {
		// import "x" — 全部导入，命名空间名从路径自动推导
		c.emitABx(OP_IMPORT, 0, srcIdx)
	}
}

func (c *Compiler) compileIncludeStmt(stmt *parser.IncludeStmt) {
	srcIdx := c.addConstant(NewString(stmt.Source))
	once := 0
	if stmt.Once {
		once = 1
	}
	// B=1 表示 include_once
	c.emitABx(OP_INCLUDE, once, srcIdx)
}

func (c *Compiler) compileFuncDecl(decl *parser.FuncDecl) {
	name := decl.Name.Value

	// 预先分配全局变量索引（供函数体内递归调用时查找）
	if c.globalScope {
		c.allocateGlobalIndex(name)
	}

	// 在当前作用域声明函数变量（必须在编译函数体之前，以便函数体内可以解析为 upvalue）
	sym := c.findOrCreateLocal(name)

	// 编译函数体
	fn := c.compileFunction(name, decl.Parameters, decl.Body)

	// 注册到全局函数列表
	c.functions = append(c.functions, fn)

	// 添加函数索引常量（用于 OP_CLOSURE 精确定位重载函数）
	// 函数在 program.Functions 中的索引 = len(c.functions)（因为 main 是 [0]）
	fnIdx := len(c.functions)
	fnConst := c.addConstant(NewInt(int64(fnIdx)))

	// 始终使用 CLOSURE 指令
	c.emitABx(OP_CLOSURE, sym.Index, fnConst)
	// 发出 upvalue 描述
	for _, uv := range fn.Upvals {
		isLocal := 0
		if uv.IsLocal {
			isLocal = 1
		}
		c.emitABC(OP_LOADBOOL, isLocal, uv.Index, 0)
	}

	// 在全局作用域时，同时存储到全局变量（供外部调用时查找）
	if c.globalScope {
		gIdx := c.allocateGlobalIndex(name)
		c.emitABx(OP_SETGLOBAL, sym.Index, gIdx)
	}
}

func (c *Compiler) compileFunction(name string, params []*parser.Identifier, body *parser.BlockStmt) *CompiledFunction {
	child := c.newChild(name)
	child.funcParams = len(params)

	// 提取参数名
	paramNames := make([]string, len(params))
	for i, p := range params {
		paramNames[i] = p.Value
		child.declareVar(p.Value)
	}

	// 编译函数体
	for _, stmt := range body.Statements {
		child.compileStmt(stmt)
	}

	// 如果没有显式 return，添加隐式 return null
	if len(child.bytecode) == 0 || child.bytecode[len(child.bytecode)-1].OP() != OP_RETURN {
		r := child.allocReg()
		child.emitABC(OP_LOADNULL, r, 0, 0)
		child.emitABC(OP_RETURN, r, 0, 0)
	}

	fn := &CompiledFunction{
		Name:        name,
		Params:      len(params),
		ParamNames:  paramNames,
		Registers:   child.maxReg,
		Bytecode:    child.bytecode,
		Constants:   child.constants,
		NumUpvals:   len(child.upvals),
		Upvals:      child.upvals,
		SourceLines: child.sourceLines,
		VarNames:    child.buildVarNames(),
	}

	// 更新父编译器的函数列表引用
	c.functions = child.functions

	return fn
}

func (c *Compiler) compileBlockStmtWithoutScope(block *parser.BlockStmt) {
	for _, stmt := range block.Statements {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) patchJumpTo(pos int, targetPC int) {
	ins := c.bytecode[pos]
	offset := targetPC - pos - 1
	c.bytecode[pos] = NewAsBx(ins.OP(), ins.A(), offset)
}

// ============================================================================
// 表达式编译
// ============================================================================

func (c *Compiler) compileExpr(expr parser.Expression, target int) int {
	// 设置当前源码行号
	if expr != nil {
		c.currentLine = expr.Pos().Line
	}

	switch e := expr.(type) {
	case *parser.NumberLiteral:
		c.compileNumberLiteral(e, target)
	case *parser.StringLiteral:
		c.compileStringLiteral(e, target)
	case *parser.BoolLiteral:
		c.compileBoolLiteral(e, target)
	case *parser.NullLiteral:
		c.emitABC(OP_LOADNULL, target, 0, 0)
	case *parser.Identifier:
		c.compileIdentifier(e, target)
	case *parser.BinaryExpr:
		c.compileBinaryExpr(e, target)
	case *parser.UnaryExpr:
		c.compileUnaryExpr(e, target)
	case *parser.CallExpr:
		c.compileCallExpr(e, target)
	case *parser.ArrayLiteral:
		c.compileArrayLiteral(e, target)
	case *parser.ObjectLiteral:
		c.compileObjectLiteral(e, target)
	case *parser.IndexExpr:
		c.compileIndexExpr(e, target, false)
	case *parser.MemberExpr:
		c.compileMemberExpr(e, target, false)
	case *parser.ConcatExpr:
		c.compileConcatExpr(e, target)
	case *parser.RangeExpr:
		c.compileRangeExpr(e, target)
	case *parser.AssignExpr:
		c.compileAssignExpr(e, target)
	case *parser.TernaryExpr:
		c.compileTernaryExpr(e, target)
	case *parser.LambdaExpr:
		c.compileLambdaExpr(e, target)
	case *parser.ArrowExpr:
		c.compileArrowExpr(e, target)
	case *parser.PipeExpr:
		c.compilePipeExpr(e, target)
	case *parser.TypeCast:
		c.compileTypeCast(e, target)
	case *parser.MatchStmt:
		c.compileMatchExpr(e, target)
	case *parser.RegexLiteral:
		c.compileRegexLiteral(e, target)
	default:
		c.emitABC(OP_LOADNULL, target, 0, 0)
	}

	return target
}

func (c *Compiler) compileExprToReg(expr parser.Expression, target int) {
	c.compileExpr(expr, target)
}

func (c *Compiler) compileNumberLiteral(num *parser.NumberLiteral, target int) {
	val := num.Value
	tokenType := num.Token.Type

	// 根据 token 类型处理
	switch tokenType {
	case token.BIGINT:
		// 大整数：使用 big.Int
		bi, ok := new(big.Int).SetString(val, 0)
		if ok {
			kIdx := c.addConstant(NewBigInt(bi))
			c.emitABx(OP_LOADK, target, kIdx)
			return
		}
		// 解析失败，回退到普通整数处理
	case token.BIGDECIMAL:
		// 大浮点数：使用 big.Rat
		br, ok := new(big.Rat).SetString(val)
		if ok {
			kIdx := c.addConstant(NewBigDecimal(br))
			c.emitABx(OP_LOADK, target, kIdx)
			return
		}
		// 解析失败，回退到普通浮点数处理
	}

	// 普通整数处理
	if n, err := strconv.ParseInt(val, 0, 64); err == nil {
		kIdx := c.addConstant(NewInt(n))
		c.emitABx(OP_LOADK, target, kIdx)
	} else if f, err := strconv.ParseFloat(val, 64); err == nil {
		// 浮点数处理
		kIdx := c.addConstant(NewFloat(f))
		c.emitABx(OP_LOADK, target, kIdx)
	} else {
		// 无法解析，默认为 0
		kIdx := c.addConstant(NewInt(0))
		c.emitABx(OP_LOADK, target, kIdx)
	}
}

func (c *Compiler) compileStringLiteral(str *parser.StringLiteral, target int) {
	kIdx := c.addConstant(NewString(str.Value))
	c.emitABx(OP_LOADK, target, kIdx)
}

func (c *Compiler) compileBoolLiteral(b *parser.BoolLiteral, target int) {
	if b.Value {
		c.emitABC(OP_LOADBOOL, target, 1, 0)
	} else {
		c.emitABC(OP_LOADBOOL, target, 0, 0)
	}
}

// compileRegexLiteral 编译正则字面量 #/pattern/flags#
//
// 编译为常量加载，运行时返回 regexValue
func (c *Compiler) compileRegexLiteral(re *parser.RegexLiteral, target int) {
	// 创建 regexValue 并添加到常量池
	regexVal, err := NewRegex(re.Pattern, re.Flags)
	if err != nil {
		// 正则已在 lexer 验证，这里不应该出错
		c.emitABC(OP_LOADNULL, target, 0, 0)
		return
	}
	kIdx := c.addConstant(regexVal)
	c.emitABx(OP_LOADK, target, kIdx)
}

func (c *Compiler) compileIdentifier(id *parser.Identifier, target int) {
	name := id.Value

	// 处理 @member 实例变量语法
	if len(name) > 0 && name[0] == '@' {
		c.compileInstanceVar(id, target)
		return
	}

	// Phase 7.8: 处理魔术常量
	if magicValue := c.getMagicConstant(name, id.Pos()); magicValue != nil {
		idx := c.addConstant(magicValue)
		c.emitABx(OP_LOADK, target, idx)
		return
	}

	// 处理运行时魔术常量 ARGV/ARGC（命令行参数）
	if name == "ARGV" {
		c.emitABC(OP_GETARGV, target, 0, 0)
		return
	}
	if name == "ARGC" {
		c.emitABC(OP_GETARGC, target, 0, 0)
		return
	}

	// 检查是否为 global 声明的变量
	if c.isGlobalVar(name) {
		gIdx := c.allocateGlobalIndex(name)
		c.emitABx(OP_GETGLOBAL, target, gIdx)
		return
	}

	// 检查是否为 static 声明的变量
	if c.isStaticVar(name) {
		gIdx := c.allocateGlobalIndex(c.staticStorageKey(name))
		c.emitABx(OP_GETGLOBAL, target, gIdx)
		return
	}

	sym := c.resolveVar(name)
	if sym != nil {
		if sym.Index != target {
			c.emitABC(OP_LOAD, target, sym.Index, 0)
		}
	} else {
		// 尝试解析 upvalue
		upvalIdx := c.resolveUpvalue(name)
		if upvalIdx >= 0 {
			c.emitABC(OP_GETUPVAL, target, upvalIdx, 0)
		} else {
			// 回退到全局变量（运行时可能由 C 代码或引擎注册）
			gIdx := c.allocateGlobalIndex(name)
			c.emitABx(OP_GETGLOBAL, target, gIdx)
		}
	}
}

// compileInstanceVar 编译 @member 实例变量访问
func (c *Compiler) compileInstanceVar(id *parser.Identifier, target int) {
	memberName := id.Value[1:] // 去掉 @ 前缀

	// 查找当前层级的 self 符号
	sym := c.resolveSelf(id.Pos())
	if sym == nil {
		panic(&CompileError{
			Message: "undefined member '" + id.Value + "' in closure: @" + memberName + " can only be used inside an object literal",
			Line:    id.Pos().Line,
			Column:  id.Pos().Column,
		})
	}

	// 编译为 self.member 访问
	// 先获取 self - 需要通过 upvalue 机制访问父作用域的变量
	selfReg := c.allocReg()

	// 尝试通过 upvalue 机制获取 self（这会自动处理捕获）
	upvalIdx := c.resolveUpvalue("__self__")
	if upvalIdx >= 0 {
		// 通过 upvalue 获取
		c.emitABC(OP_GETUPVAL, selfReg, upvalIdx, 0)
	} else {
		// 同一作用域，直接通过寄存器获取
		c.emitABC(OP_LOAD, selfReg, sym.Index, 0)
	}

	// 然后获取成员
	propIdx := c.addConstant(NewString(memberName))
	c.emitABC(OP_GETMEMBER, target, selfReg, propIdx)

	c.freeReg()
}

// resolveSelf 解析当前层级的 self 符号
func (c *Compiler) resolveSelf(pos token.Position) *Symbol {
	// 从当前层级向上查找
	for depth := c.objectDepth; depth > 0; depth-- {
		if sym, ok := c.objectSelfSym[depth]; ok {
			return sym
		}
	}

	// 检查父编译器（闭包在子编译器中）
	if c.parent != nil {
		// 父编译器的 objectSelfSym 与当前共享，直接查找任意层级的 self
		for depth := len(c.objectSelfSym); depth > 0; depth-- {
			if sym, ok := c.objectSelfSym[depth]; ok {
				return sym
			}
		}
		return c.parent.resolveSelf(pos)
	}
	return nil
}

// isGlobalVar 检查变量是否被标记为 global
func (c *Compiler) isGlobalVar(name string) bool {
	if c.globalVars[name] {
		return true
	}
	// 检查父编译器
	if c.parent != nil {
		return c.parent.isGlobalVar(name)
	}
	return false
}

// isStaticVar 检查变量是否被标记为 static
func (c *Compiler) isStaticVar(name string) bool {
	key := c.funcName + "::" + name
	if _, ok := c.staticVars[key]; ok {
		return true
	}
	return false
}

// staticStorageKey 返回静态变量的全局存储键
func (c *Compiler) staticStorageKey(name string) string {
	return "_static:" + c.funcName + "::" + name
}

func (c *Compiler) compileBinaryExpr(expr *parser.BinaryExpr, target int) {
	// 赋值
	if expr.Operator == "=" {
		c.compileAssign(expr.Left, expr.Right, target)
		return
	}

	// 短路求值 &&, ||
	if expr.Operator == "&&" || expr.Operator == "||" {
		c.compileShortCircuit(expr, target)
		return
	}

	// 常量折叠：如果左右操作数都是数字字面量，编译时计算
	if folded := c.tryFoldBinary(expr); folded != nil {
		kIdx := c.addConstant(folded)
		c.emitABx(OP_LOADK, target, kIdx)
		return
	}

	// 编译左右操作数
	left := c.allocReg()
	c.compileExprToReg(expr.Left, left)
	right := c.allocReg()
	c.compileExprToReg(expr.Right, right)

	// 生成运算指令
	switch expr.Operator {
	case "+":
		c.emitABC(OP_ADD, target, left, right)
	case "-":
		c.emitABC(OP_SUB, target, left, right)
	case "*":
		c.emitABC(OP_MUL, target, left, right)
	case "/":
		c.emitABC(OP_DIV, target, left, right)
	case "%":
		c.emitABC(OP_MOD, target, left, right)
	case "==":
		c.emitABC(OP_EQ, target, left, right)
	case "!=":
		c.emitABC(OP_NEQ, target, left, right)
	case "<":
		c.emitABC(OP_LT, target, left, right)
	case ">":
		c.emitABC(OP_GT, target, left, right)
	case "<=":
		c.emitABC(OP_LTE, target, left, right)
	case ">=":
		c.emitABC(OP_GTE, target, left, right)
	case "&&":
		c.emitABC(OP_AND, target, left, right)
	case "||":
		c.emitABC(OP_OR, target, left, right)
	case "&":
		c.emitABC(OP_BITAND, target, left, right)
	case "|":
		c.emitABC(OP_BITOR, target, left, right)
	case "^":
		c.emitABC(OP_BITXOR, target, left, right)
	case "<<":
		c.emitABC(OP_SHL, target, left, right)
	case ">>":
		c.emitABC(OP_SHR, target, left, right)
	case "..":
		c.emitABC(OP_CONCAT, target, left, right)
	case "=~":
		c.emitABC(OP_REGEX_MATCH, target, left, right)
	default:
		c.emitABC(OP_LOADNULL, target, 0, 0)
	}

	c.freeReg() // right
	c.freeReg() // left
}

// tryFoldBinary 尝试常量折叠，返回折叠后的值，如果不能折叠则返回 nil
func (c *Compiler) tryFoldBinary(expr *parser.BinaryExpr) Value {
	// 获取左右操作数的常量值（递归折叠）
	leftVal := c.tryEvalConstant(expr.Left)
	rightVal := c.tryEvalConstant(expr.Right)
	if leftVal == nil || rightVal == nil {
		return nil
	}

	// 根据运算符计算结果
	switch expr.Operator {
	// 算术
	case "+":
		return tryFoldAdd(leftVal, rightVal)
	case "-":
		return tryFoldArith(leftVal, rightVal, func(a, b float64) float64 { return a - b })
	case "*":
		return tryFoldArith(leftVal, rightVal, func(a, b float64) float64 { return a * b })
	case "/":
		if isZero(rightVal) {
			return nil // 除零不折叠，留到运行时
		}
		return tryFoldArith(leftVal, rightVal, func(a, b float64) float64 { return a / b })
	case "%":
		if isZero(rightVal) {
			return nil
		}
		if leftVal.Type() == TypeInt && rightVal.Type() == TypeInt {
			return NewInt(leftVal.Int() % rightVal.Int())
		}
		return nil

	// 比较
	case "==":
		return NewBool(compareValues(leftVal, rightVal) == 0)
	case "!=":
		return NewBool(compareValues(leftVal, rightVal) != 0)
	case "<":
		return NewBool(compareValues(leftVal, rightVal) < 0)
	case ">":
		return NewBool(compareValues(leftVal, rightVal) > 0)
	case "<=":
		return NewBool(compareValues(leftVal, rightVal) <= 0)
	case ">=":
		return NewBool(compareValues(leftVal, rightVal) >= 0)

	// 逻辑（仅当左操作数为常量时处理短路）
	case "&&":
		if leftVal.Type() == TypeBool {
			if !leftVal.Bool() {
				return NewBool(false) // false && x = false
			}
			if rightVal.Type() == TypeBool {
				return NewBool(rightVal.Bool()) // true && true = true
			}
		}
		return nil
	case "||":
		if leftVal.Type() == TypeBool {
			if leftVal.Bool() {
				return NewBool(true) // true || x = true
			}
			if rightVal.Type() == TypeBool {
				return NewBool(rightVal.Bool()) // false || false = false
			}
		}
		return nil

	// 位运算
	case "&":
		return tryFoldBitwise(leftVal, rightVal, func(a, b int64) int64 { return a & b })
	case "|":
		return tryFoldBitwise(leftVal, rightVal, func(a, b int64) int64 { return a | b })
	case "^":
		return tryFoldBitwise(leftVal, rightVal, func(a, b int64) int64 { return a ^ b })
	case "<<":
		return tryFoldBitwise(leftVal, rightVal, func(a, b int64) int64 { return a << uint(b) })
	case ">>":
		return tryFoldBitwise(leftVal, rightVal, func(a, b int64) int64 { return a >> uint(b) })

	// 字符串拼接
	case "..":
		if leftVal.Type() == TypeString && rightVal.Type() == TypeString {
			return NewString(leftVal.String() + rightVal.String())
		}
		return nil
	}

	return nil
}

// tryEvalConstant 尝试将表达式求值为编译期常量
func (c *Compiler) tryEvalConstant(expr parser.Expression) Value {
	switch e := expr.(type) {
	case *parser.NumberLiteral:
		if n, err := strconv.ParseInt(e.Value, 0, 64); err == nil {
			return NewInt(n)
		}
		if f, err := strconv.ParseFloat(e.Value, 64); err == nil {
			return NewFloat(f)
		}
		return nil
	case *parser.StringLiteral:
		return NewString(e.Value)
	case *parser.BoolLiteral:
		return NewBool(e.Value)
	case *parser.NullLiteral:
		return NewNull()
	case *parser.BinaryExpr:
		return c.tryFoldBinary(e)
	case *parser.ConcatExpr:
		return c.tryFoldConcat(e)
	case *parser.UnaryExpr:
		return c.tryFoldUnary(e)
	case *parser.TernaryExpr:
		return c.tryFoldTernary(e)
	}
	return nil
}

// tryFoldUnary 尝试折叠一元表达式
func (c *Compiler) tryFoldUnary(expr *parser.UnaryExpr) Value {
	val := c.tryEvalConstant(expr.Operand)
	if val == nil {
		return nil
	}
	switch expr.Operator {
	case "-":
		if val.Type() == TypeInt {
			return NewInt(-val.Int())
		}
		if val.Type() == TypeFloat {
			return NewFloat(-val.Float())
		}
	case "!":
		if val.Type() == TypeBool {
			return NewBool(!val.Bool())
		}
	case "~":
		if val.Type() == TypeInt {
			return NewInt(^val.Int())
		}
	}
	return nil
}

// tryFoldTernary 尝试折叠三元表达式
func (c *Compiler) tryFoldTernary(expr *parser.TernaryExpr) Value {
	cond := c.tryEvalConstant(expr.Condition)
	if cond == nil || cond.Type() != TypeBool {
		return nil
	}
	if cond.Bool() {
		return c.tryEvalConstant(expr.TrueExpr)
	}
	return c.tryEvalConstant(expr.FalseExpr)
}

// tryFoldConcat 尝试折叠字符串拼接
func (c *Compiler) tryFoldConcat(expr *parser.ConcatExpr) Value {
	left := c.tryEvalConstant(expr.Left)
	right := c.tryEvalConstant(expr.Right)
	if left != nil && right != nil && left.Type() == TypeString && right.Type() == TypeString {
		return NewString(left.String() + right.String())
	}
	return nil
}

// ============================================================================
// 常量折叠辅助函数
// ============================================================================

// tryFoldAdd 处理加法（支持整数和字符串）
func tryFoldAdd(left, right Value) Value {
	if left.Type() == TypeInt && right.Type() == TypeInt {
		return NewInt(left.Int() + right.Int())
	}
	if left.Type() == TypeFloat || right.Type() == TypeFloat {
		return NewFloat(toFloat(left) + toFloat(right))
	}
	return nil
}

// tryFoldArith 处理算术运算（-, *, /）
func tryFoldArith(left, right Value, op func(float64, float64) float64) Value {
	if left.Type() == TypeInt && right.Type() == TypeInt {
		return NewInt(int64(op(float64(left.Int()), float64(right.Int()))))
	}
	return NewFloat(op(toFloat(left), toFloat(right)))
}

// tryFoldBitwise 处理位运算
func tryFoldBitwise(left, right Value, op func(int64, int64) int64) Value {
	if left.Type() == TypeInt && right.Type() == TypeInt {
		return NewInt(op(left.Int(), right.Int()))
	}
	return nil
}

// compareValues 比较两个值，返回 -1/0/1
func compareValues(a, b Value) int {
	// 同类型比较
	if a.Type() == TypeInt && b.Type() == TypeInt {
		ai, bi := a.Int(), b.Int()
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
		return 0
	}
	if a.Type() == TypeFloat || b.Type() == TypeFloat {
		af, bf := toFloat(a), toFloat(b)
		if af < bf {
			return -1
		}
		if af > bf {
			return 1
		}
		return 0
	}
	if a.Type() == TypeString && b.Type() == TypeString {
		if a.String() < b.String() {
			return -1
		}
		if a.String() > b.String() {
			return 1
		}
		return 0
	}
	if a.Type() == TypeBool && b.Type() == TypeBool {
		ab, bb := a.Bool(), b.Bool()
		if ab == bb {
			return 0
		}
		if !ab && bb {
			return -1
		}
		return 1
	}
	// null == null
	if a.Type() == TypeNull && b.Type() == TypeNull {
		return 0
	}
	return -2 // 无法比较
}

// isZero 检查值是否为零（用于除零检查）
func isZero(v Value) bool {
	if v.Type() == TypeInt {
		return v.Int() == 0
	}
	if v.Type() == TypeFloat {
		return v.Float() == 0.0
	}
	return false
}

// toFloat 将值转为 float64
func toFloat(v Value) float64 {
	if v.Type() == TypeInt {
		return float64(v.Int())
	}
	if v.Type() == TypeFloat {
		return v.Float()
	}
	return 0
}

func (c *Compiler) compileShortCircuit(expr *parser.BinaryExpr, target int) {
	left := c.allocReg()
	c.compileExprToReg(expr.Left, left)

	if expr.Operator == "&&" {
		// left && right：left 为假则跳过 right
		// 先将 left 存到 target（如果 left 为假，target 应该是 left 的值）
		if left != target {
			c.emitABC(OP_LOAD, target, left, 0)
		}
		jump := c.currentPC()
		c.emitAsBx(OP_JMPIFNOT, left, 0)
		c.freeReg()

		right := c.allocReg()
		c.compileExprToReg(expr.Right, right)
		if right != target {
			c.emitABC(OP_LOAD, target, right, 0)
		}
		c.freeReg()

		c.patchJump(jump)
	} else {
		// left || right：left 为真则跳过 right
		// 先将 left 存到 target（如果 left 为真，target 应该是 left 的值）
		if left != target {
			c.emitABC(OP_LOAD, target, left, 0)
		}
		jump := c.currentPC()
		c.emitAsBx(OP_JMPIF, left, 0)
		c.freeReg()

		right := c.allocReg()
		c.compileExprToReg(expr.Right, right)
		if right != target {
			c.emitABC(OP_LOAD, target, right, 0)
		}
		c.freeReg()

		c.patchJump(jump)
	}
}

func (c *Compiler) compileUnaryExpr(expr *parser.UnaryExpr, target int) {
	// 常量折叠
	if folded := c.tryFoldUnary(expr); folded != nil {
		kIdx := c.addConstant(folded)
		c.emitABx(OP_LOADK, target, kIdx)
		return
	}

	operand := c.allocReg()
	c.compileExprToReg(expr.Operand, operand)

	switch expr.Operator {
	case "-":
		c.emitABC(OP_NEG, target, operand, 0)
	case "!":
		c.emitABC(OP_NOT, target, operand, 0)
	case "~":
		// 按位取反
		c.emitABC(OP_BITNOT, target, operand, 0)
	default:
		c.emitABC(OP_LOAD, target, operand, 0)
	}

	c.freeReg()
}

func (c *Compiler) compileCallExpr(expr *parser.CallExpr, target int) {
	// 检查是否为 typeof 内置函数
	if ident, ok := expr.Function.(*parser.Identifier); ok && ident.Value == "typeof" {
		if len(expr.Arguments) != 1 {
			// typeof 需要恰好一个参数
			c.emitABC(OP_LOADNULL, target, 0, 0)
			return
		}
		arg := c.allocReg()
		c.compileExprToReg(expr.Arguments[0], arg)
		c.emitABC(OP_TYPEOF, target, arg, 0)
		c.freeReg()
		return
	}

	// 编译被调用者
	callee := c.allocReg()
	c.compileExprToReg(expr.Function, callee)

	// 编译参数（从 callee+1 开始）
	for i, arg := range expr.Arguments {
		argReg := callee + 1 + i
		if argReg >= c.nextReg {
			c.allocReg()
		}
		c.compileExprToReg(arg, argReg)
	}

	// 调用指令
	c.emitABC(OP_CALL, callee, 0, len(expr.Arguments)+1)

	// 结果移到 target
	if callee != target {
		c.emitABC(OP_LOAD, target, callee, 0)
	}

	// 释放参数寄存器
	for i := 0; i < len(expr.Arguments); i++ {
		c.freeReg()
	}
	c.freeReg() // callee
}

func (c *Compiler) compileArrayLiteral(expr *parser.ArrayLiteral, target int) {
	if len(expr.Elements) == 0 {
		c.emitABC(OP_NEWARRAY, target, 0, 0)
		return
	}

	// 编译元素到连续寄存器
	baseReg := c.nextReg
	for _, elem := range expr.Elements {
		r := c.allocReg()
		c.compileExprToReg(elem, r)
	}

	c.emitABC(OP_NEWARRAY, target, baseReg, len(expr.Elements))

	// 释放元素寄存器
	for i := 0; i < len(expr.Elements); i++ {
		c.freeReg()
	}
}

func (c *Compiler) compileObjectLiteral(expr *parser.ObjectLiteral, target int) {
	if len(expr.Pairs) == 0 {
		c.emitABC(OP_NEWOBJECT, target, 0, 0)
		return
	}

	// 为 self 创建局部变量（在当前作用域）
	// 这样闭包可以通过 upvalue 访问它
	c.declareVar("__self__")
	selfSym := c.scopes[len(c.scopes)-1]["__self__"]
	selfReg := selfSym.Index

	// 创建对象
	c.emitABC(OP_NEWOBJECT, target, 0, 0)

	// 将对象存储到 self 寄存器
	c.emitABC(OP_LOAD, selfReg, target, 0)

	// 进入对象上下文
	c.objectDepth++
	c.objectSelfSym[c.objectDepth] = selfSym

	// 收集键值对
	type kv struct {
		keyIdx   int
		valueReg int
	}
	pairs := make([]kv, 0, len(expr.Pairs))

	for key, val := range expr.Pairs {
		// 处理键
		var keyConst int
		switch k := key.(type) {
		case *parser.StringLiteral:
			keyConst = c.addConstant(NewString(k.Value))
		case *parser.Identifier:
			keyConst = c.addConstant(NewString(k.Value))
		default:
			keyConst = c.addConstant(NewString(key.String()))
		}

		// 编译值
		vr := c.allocReg()
		c.compileExprToReg(val, vr)

		pairs = append(pairs, kv{keyIdx: keyConst, valueReg: vr})
	}

	// 设置成员
	for _, p := range pairs {
		c.emitABC(OP_SETMEMBER, p.valueReg, target, p.keyIdx)
		c.freeReg()
	}

	// 释放 __self__ 寄存器
	c.freeReg()

	// 退出对象上下文
	delete(c.objectSelfSym, c.objectDepth)
	c.objectDepth--
}

func (c *Compiler) compileIndexExpr(expr *parser.IndexExpr, target int, asStore bool) int {
	obj := c.allocReg()
	c.compileExprToReg(expr.Left, obj)
	idx := c.allocReg()
	c.compileExprToReg(expr.Index, idx)

	if !asStore {
		c.emitABC(OP_GETINDEX, target, obj, idx)
	}

	c.freeReg() // idx
	c.freeReg() // obj
	return obj
}

func (c *Compiler) compileMemberExpr(expr *parser.MemberExpr, target int, asStore bool) int {
	obj := c.allocReg()
	c.compileExprToReg(expr.Object, obj)

	propName := expr.Member.Value
	propIdx := c.addConstant(NewString(propName))

	if !asStore {
		c.emitABC(OP_GETMEMBER, target, obj, propIdx)
	}

	c.freeReg()
	return obj
}

func (c *Compiler) compileConcatExpr(expr *parser.ConcatExpr, target int) {
	// 常量折叠：字符串拼接
	leftVal := c.tryEvalConstant(expr.Left)
	rightVal := c.tryEvalConstant(expr.Right)
	if leftVal != nil && rightVal != nil && leftVal.Type() == TypeString && rightVal.Type() == TypeString {
		kIdx := c.addConstant(NewString(leftVal.String() + rightVal.String()))
		c.emitABx(OP_LOADK, target, kIdx)
		return
	}

	left := c.allocReg()
	c.compileExprToReg(expr.Left, left)
	right := c.allocReg()
	c.compileExprToReg(expr.Right, right)

	c.emitABC(OP_CONCAT, target, left, right)

	c.freeReg() // right
	c.freeReg() // left
}

func (c *Compiler) compileRangeExpr(expr *parser.RangeExpr, target int) {
	startReg := c.allocReg()
	c.compileExprToReg(expr.Start, startReg)

	endReg := c.allocReg()
	c.compileExprToReg(expr.End, endReg)

	op := OP_RANGE
	if expr.Inclusive {
		op = OP_RANGE_INCLUSIVE
	}
	c.emitABC(op, target, startReg, endReg)

	c.freeReg() // endReg
	c.freeReg() // startReg
}

func (c *Compiler) compileAssignExpr(expr *parser.AssignExpr, target int) {
	c.compileAssign(expr.Left, expr.Value, target)
}

func (c *Compiler) compileAssign(left, right parser.Expression, target int) {
	switch l := left.(type) {
	case *parser.Identifier:
		// 变量赋值
		name := l.Value

		// 禁止修改运行时魔术常量
		if name == "ARGV" || name == "ARGC" {
			panic(fmt.Sprintf("cannot assign to constant %s", name))
		}

		// 检查是否为 global 声明的变量
		if c.isGlobalVar(name) {
			r := c.allocReg()
			c.compileExprToReg(right, r)
			gIdx := c.allocateGlobalIndex(name)
			c.emitABx(OP_SETGLOBAL, r, gIdx)
			if r != target {
				c.emitABC(OP_LOAD, target, r, 0)
			}
			c.freeReg()
			return
		}

		// 检查是否为 static 声明的变量
		if c.isStaticVar(name) {
			r := c.allocReg()
			c.compileExprToReg(right, r)
			gIdx := c.allocateGlobalIndex(c.staticStorageKey(name))
			c.emitABx(OP_SETGLOBAL, r, gIdx)
			if r != target {
				c.emitABC(OP_LOAD, target, r, 0)
			}
			c.freeReg()
			return
		}

		sym := c.resolveVar(name)
		if sym == nil {
			// 尝试解析 upvalue
			upvalIdx := c.resolveUpvalue(name)
			if upvalIdx >= 0 {
				r := c.allocReg()
				c.compileExprToReg(right, r)
				c.emitABC(OP_SETUPVAL, upvalIdx, r, 0)
				if r != target {
					c.emitABC(OP_LOAD, target, r, 0)
				}
				c.freeReg()
				return
			}
			// 检查是否为全局变量（沿父编译器链查找）
			if c.resolveGlobal(name) {
				// 全局变量：使用 SETGLOBAL
				r := c.allocReg()
				c.compileExprToReg(right, r)
				gIdx := c.allocateGlobalIndex(name)
				c.emitABx(OP_SETGLOBAL, r, gIdx)
				if r != target {
					c.emitABC(OP_LOAD, target, r, 0)
				}
				c.freeReg()
				return
			}
			// 新变量：声明为局部变量
			sym = c.declareVar(name)
		}
		r := c.allocReg()
		c.compileExprToReg(right, r)
		// LOAD A, B: A = B, 所以 LOAD sym.Index, r 将 r 的值存到 sym.Index
		c.emitABC(OP_LOAD, sym.Index, r, 0)
		c.freeReg()
		// 如果 target 与 sym.Index 不同，将值复制到 target
		if sym.Index != target {
			c.emitABC(OP_LOAD, target, sym.Index, 0)
		}
		// 全局作用域：同时更新 globals map
		if c.globalScope {
			gIdx := c.allocateGlobalIndex(name)
			c.emitABx(OP_SETGLOBAL, sym.Index, gIdx)
		}
	case *parser.IndexExpr:
		// $arr[i] = expr
		obj := c.allocReg()
		c.compileExprToReg(l.Left, obj)
		idx := c.allocReg()
		c.compileExprToReg(l.Index, idx)
		val := c.allocReg()
		c.compileExprToReg(right, val)
		c.emitABC(OP_SETINDEX, val, obj, idx)
		if val != target {
			c.emitABC(OP_LOAD, target, val, 0)
		}
		c.freeReg() // val
		c.freeReg() // idx
		c.freeReg() // obj
	case *parser.MemberExpr:
		// $obj.field = expr
		obj := c.allocReg()
		c.compileExprToReg(l.Object, obj)
		propName := l.Member.Value
		propIdx := c.addConstant(NewString(propName))
		val := c.allocReg()
		c.compileExprToReg(right, val)
		c.emitABC(OP_SETMEMBER, val, obj, propIdx)
		if val != target {
			c.emitABC(OP_LOAD, target, val, 0)
		}
		c.freeReg() // val
		c.freeReg() // obj
	default:
		// 未知左值
		r := c.allocReg()
		c.compileExprToReg(right, r)
		if r != target {
			c.emitABC(OP_LOAD, target, r, 0)
		}
		c.freeReg()
	}
}

func (c *Compiler) compileTernaryExpr(expr *parser.TernaryExpr, target int) {
	// 常量折叠
	if folded := c.tryFoldTernary(expr); folded != nil {
		kIdx := c.addConstant(folded)
		c.emitABx(OP_LOADK, target, kIdx)
		return
	}

	// 编译条件
	cond := c.allocReg()
	c.compileExprToReg(expr.Condition, cond)

	// 条件跳转（跳过真值表达式）
	jumpIfNot := c.currentPC()
	c.emitAsBx(OP_JMPIFNOT, cond, 0)
	c.freeReg()

	// 真值表达式
	c.compileExprToReg(expr.TrueExpr, target)

	// 跳过假值表达式
	jumpOver := c.currentPC()
	c.emitAsBx(OP_JMP, 0, 0)

	// 修补条件跳转
	c.patchJump(jumpIfNot)

	// 假值表达式
	c.compileExprToReg(expr.FalseExpr, target)

	// 修补跳过跳转
	c.patchJump(jumpOver)
}

// ============================================================================
// Lambda 和箭头函数
// ============================================================================

func (c *Compiler) compileLambdaExpr(expr *parser.LambdaExpr, target int) {
	// 生成唯一的函数名
	fnName := fmt.Sprintf("<lambda:%d>", expr.Pos().Line)

	// 编译函数体
	fn := c.compileFunction(fnName, expr.Parameters, expr.Body)

	// 注册到函数列表
	c.functions = append(c.functions, fn)

	// 添加函数索引常量
	fnIdx := len(c.functions)
	fnConst := c.addConstant(NewInt(int64(fnIdx)))

	// 始终使用 CLOSURE 指令
	c.emitABx(OP_CLOSURE, target, fnConst)
	// 发出 upvalue 描述
	for _, uv := range fn.Upvals {
		isLocal := 0
		if uv.IsLocal {
			isLocal = 1
		}
		c.emitABC(OP_LOADBOOL, isLocal, uv.Index, 0)
	}
}

func (c *Compiler) compileArrowExpr(expr *parser.ArrowExpr, target int) {
	// 生成唯一的函数名
	fnName := fmt.Sprintf("<arrow:%d>", expr.Pos().Line)

	// 构建函数体
	var body *parser.BlockStmt
	if expr.Body != nil {
		// 单行表达式：return expr
		body = &parser.BlockStmt{
			Statements: []parser.Statement{
				&parser.ReturnStmt{
					Value: expr.Body,
				},
			},
		}
	} else {
		body = expr.BlockBody
	}

	// 编译函数体
	fn := c.compileFunction(fnName, expr.Parameters, body)

	// 注册到函数列表
	c.functions = append(c.functions, fn)

	// 添加函数索引常量
	fnIdx := len(c.functions)
	fnConst := c.addConstant(NewInt(int64(fnIdx)))

	// 始终使用 CLOSURE 指令
	c.emitABx(OP_CLOSURE, target, fnConst)
	// 发出 upvalue 描述
	for _, uv := range fn.Upvals {
		isLocal := 0
		if uv.IsLocal {
			isLocal = 1
		}
		c.emitABC(OP_LOADBOOL, isLocal, uv.Index, 0)
	}
}

// compilePipeExpr 编译管道表达式
// |> 正向管道: a |> f(b,c) = f(a, b, c)，左侧值作为首个参数
// <| 反向管道: f(b,c) <| a = f(b, c, a)，右侧值作为末尾参数
func (c *Compiler) compilePipeExpr(expr *parser.PipeExpr, target int) {
	if expr.Forward {
		// 正向管道: Left |> Right
		// 如果 Right 是 CallExpr: f(b,c) -> f(Left, b, c)
		// 如果 Right 是标识符: f -> f(Left)
		if call, ok := expr.Right.(*parser.CallExpr); ok {
			c.compilePipeCall(expr.Left, call, true, target)
		} else {
			// 简单调用: a |> f => f(a)
			callee := c.allocReg()
			c.compileExprToReg(expr.Right, callee)

			argReg := c.allocReg()
			c.compileExprToReg(expr.Left, argReg)

			c.emitABC(OP_CALL, callee, 0, 2)

			if callee != target {
				c.emitABC(OP_LOAD, target, callee, 0)
			}

			c.freeReg() // argReg
			c.freeReg() // callee
		}
	} else {
		// 反向管道: Left <| Right
		// 如果 Left 是 CallExpr: f(b,c) -> f(b, c, Right)
		// 如果 Left 是标识符: f -> f(Right)
		if call, ok := expr.Left.(*parser.CallExpr); ok {
			c.compilePipeCall(expr.Right, call, false, target)
		} else {
			// 简单调用: f <| a => f(a)
			callee := c.allocReg()
			c.compileExprToReg(expr.Left, callee)

			argReg := c.allocReg()
			c.compileExprToReg(expr.Right, argReg)

			c.emitABC(OP_CALL, callee, 0, 2)

			if callee != target {
				c.emitABC(OP_LOAD, target, callee, 0)
			}

			c.freeReg() // argReg
			c.freeReg() // callee
		}
	}
}

// compilePipeCall 编译带参数的管道调用
// forward=true: value |> f(args...) => f(value, args...)
// forward=false: f(args...) <| value => f(args..., value)
func (c *Compiler) compilePipeCall(value parser.Expression, call *parser.CallExpr, forward bool, target int) {
	// 编译被调用者
	callee := c.allocReg()
	c.compileExprToReg(call.Function, callee)

	totalArgs := len(call.Arguments) + 1

	if forward {
		// 正向: value 作为首个参数
		// 寄存器布局: [callee, value, arg1, arg2, ...]
		// 编译 value 到 callee+1
		argReg := callee + 1
		if argReg >= c.nextReg {
			c.allocReg()
		}
		c.compileExprToReg(value, argReg)

		// 编译原有参数到 callee+2, callee+3, ...
		for i, arg := range call.Arguments {
			argReg := callee + 2 + i
			if argReg >= c.nextReg {
				c.allocReg()
			}
			c.compileExprToReg(arg, argReg)
		}
	} else {
		// 反向: value 作为末尾参数
		// 寄存器布局: [callee, arg1, arg2, ..., value]
		// 编译原有参数
		for i, arg := range call.Arguments {
			argReg := callee + 1 + i
			if argReg >= c.nextReg {
				c.allocReg()
			}
			c.compileExprToReg(arg, argReg)
		}

		// 编译 value 到末尾
		lastArgReg := callee + 1 + len(call.Arguments)
		if lastArgReg >= c.nextReg {
			c.allocReg()
		}
		c.compileExprToReg(value, lastArgReg)
	}

	// 调用
	c.emitABC(OP_CALL, callee, 0, totalArgs+1)

	// 结果移到 target
	if callee != target {
		c.emitABC(OP_LOAD, target, callee, 0)
	}

	// 释放参数寄存器
	for i := 0; i < totalArgs; i++ {
		c.freeReg()
	}
	c.freeReg() // callee
}

// compileTypeCast 编译类型转换表达式
// 将 Go 风格的类型转换语法 int(x), float(x), string(x), bool(x) 编译为 OP_CAST 字节码
//
// 参数：
//   - expr: TypeCast AST 节点，包含目标类型和被转换的表达式
//   - target: 目标寄存器索引，用于存储转换后的结果
//
// 转换类型编码：
//   - 0: int - 转换为整数（截断小数、解析字符串、布尔转 0/1）
//   - 1: float - 转换为浮点数（整数转浮点、解析字符串、布尔转 0.0/1.0）
//   - 2: string - 转换为字符串（调用 String() 方法）
//   - 3: bool - 转换为布尔值（使用 IsTruthy 语义）
//
// 字节码生成：
//  1. 编译被转换的表达式到临时寄存器
//  2. 发出 OP_CAST 指令：OP_CAST target, exprReg, castType
//  3. 释放临时寄存器
//
// 示例：
//
//	int("42") 编译为：
//	  LOADK exprReg, "42"    ; 加载字符串
//	  CAST target, exprReg, 0 ; 转换为 int
func (c *Compiler) compileTypeCast(expr *parser.TypeCast, target int) {
	// 编译被转换的表达式
	exprReg := c.allocReg()
	c.compileExpr(expr.Expr, exprReg)

	// 根据目标类型选择转换操作
	var castType int
	switch expr.Type {
	case "int":
		castType = 0
	case "float":
		castType = 1
	case "string":
		castType = 2
	case "bool":
		castType = 3
	default:
		castType = 0 // 默认转 int
	}

	// 发出类型转换指令: OP_CAST target, exprReg, castType
	c.emitABC(OP_CAST, target, exprReg, castType)

	// 释放临时寄存器
	c.freeReg()
}

// ============================================================================
// 辅助函数
// ============================================================================

// IsRuntimeError 检查 Value 是否为运行时错误类型。
//
// 此方法用于在 Go 代码中检查脚本执行产生的错误值，
// 特别适用于从脚本返回值中提取错误信息。
//
// 参数：
//   - v: 要检查的 Value
//
// 返回值：
//   - true: 值是运行时错误
//   - false: 值不是运行时错误
//
// 使用示例：
//
//	result := vm.GetResult()
//	if engine.IsRuntimeError(result) {
//	    fmt.Println("执行出错:", result.Stringify())
//	}
//
// IsRuntimeError 检查给定值是否为运行时错误。
//
// 返回 true 表示该值是一个运行时错误，可通过 vm.GetResult() 获取。
func IsRuntimeError(v Value) bool {
	_, ok := v.(*runtimeError)
	return ok
}

// DisassembleProgram 反编译整个程序，返回可读的汇编格式文本。
//
// 此方法用于调试，将编译后的 Program 中的所有函数字节码
// 转换为人类可读的汇编指令格式。每个函数单独显示，以换行分隔。
//
// 输出格式：
//   - 函数头：== 函数名 ==
//   - 元信息：params、registers、upvals、constants、instructions 数量
//   - 指令列表：每行一条，包含 PC、操作码、操作数
//
// 参数：
//   - prog: 要反编译的程序
//
// 返回值：
//   - string: 反编译后的文本
//
// 使用示例：
//
//	// 编译脚本
//	compiled, _ := engine.CompileString(`$x = 10 + 20`)
//
//	// 查看字节码
//	fmt.Println(engine.DisassembleProgram(compiled))
//	// 输出示例:
//	// == <main> ==
//	//   params: 0, registers: 2, upvals: 0, constants: 3, instructions: 4
//	//   0000: LOADK       R0, K0 (10)
//	//   0001: LOADK       R1, K1 (20)
//	//   0002: ADD         R0, R0, R1
//	//   0003: SETGLOBAL   R0, K2 ($x)
func DisassembleProgram(prog *Program) string {
	var result string
	for _, fn := range prog.Functions {
		result += Disassemble(fn)
		result += "\n"
	}
	return result
}

// CompileString 编译 JPL 源代码字符串（便捷入口）。
//
// 这是一个便捷的编译入口，内部完成：
//  1. 词法分析（source -> tokens）
//  2. 语法分析（tokens -> AST）
//  3. 字节码编译（AST -> Program）
//
// 文件名默认为 "<script>"，用于错误报告。
// 如果需要指定文件名，使用 CompileStringWithName()。
//
// 参数：
//   - script: JPL 源代码字符串
//
// 返回值：
//   - *Program: 编译后的程序
//   - nil: 编译成功
//   - error: 词法/语法/编译错误
//
// 使用示例：
//
//	// 直接编译代码字符串
//	compiled, err := engine.CompileString(`
//	    $x = 10
//	    $y = 20
//	    $result = $x + $y
//	`)
//	if err != nil {
//	    log.Fatal("编译失败:", err)
//	}
//
//	// 创建 VM 执行
//	vm := engine.NewVMWithProgram(eng, compiled)
//	vm.Execute()
func CompileString(script string) (*Program, error) {
	l := lexer.NewLexer(script, "<script>")
	p := parser.NewParser(l)
	program := p.Parse()

	// 检查解析错误
	if len(p.Errors()) > 0 {
		msg := ""
		for _, e := range p.Errors() {
			msg += e + "\n"
		}
		return nil, &CompileError{Message: msg}
	}

	return Compile(program)
}

// CompileStringWithName 编译 JPL 源代码字符串（指定文件名）。
//
// 与 CompileString() 相同，但允许指定文件名，用于：
//   - 错误报告中显示正确的文件名
//   - 调试信息中的源文件定位
//   - __FILE__ 常量的值
//
// 此方法内部同样完成词法分析、语法分析和字节码编译三个步骤。
//
// 参数：
//   - script: JPL 源代码字符串
//   - filename: 源文件名（用于错误报告和调试）
//
// 返回值：
//   - *Program: 编译后的程序
//   - nil: 编译成功
//   - error: 词法/语法/编译错误
//
// 使用示例：
//
//	// 从文件读取并编译
//	source, err := os.ReadFile("script.jpl")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 使用真实文件名编译
//	compiled, err := engine.CompileStringWithName(
//	    string(source),
//	    "script.jpl",
//	)
//	if err != nil {
//	    log.Fatal("编译 script.jpl 失败:", err)
//	}
//
//	// 执行时错误报告会显示文件名
//	vm := engine.NewVMWithProgram(eng, compiled)
//	if err := vm.Execute(); err != nil {
//	    log.Printf("script.jpl 执行失败: %v", err)
//	}
func CompileStringWithName(script string, filename string) (*Program, error) {
	l := lexer.NewLexer(script, filename)
	p := parser.NewParser(l)
	program := p.Parse()

	// 检查解析错误
	if len(p.Errors()) > 0 {
		msg := ""
		for _, e := range p.Errors() {
			msg += e + "\n"
		}
		return nil, &CompileError{Message: msg}
	}

	c := NewCompiler()
	c.filename = filename
	c.dirname = getDirFromFilename(filename)

	// 捕获编译 panic 并转换为 error
	var compileErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case *CompileError:
					compileErr = e
				case error:
					compileErr = &CompileError{Message: e.Error()}
				default:
					compileErr = &CompileError{Message: fmt.Sprintf("%v", r)}
				}
			}
		}()
		c.compileProgram(program)
	}()

	if compileErr != nil {
		return nil, compileErr
	}

	return c.buildProgram(), nil
}

// CompileStringWithGlobals 编译 JPL 源代码字符串（指定文件名和现有全局变量）。
//
// 此方法与 CompileStringWithName() 类似，但允许传入已存在的全局变量名列表，
// 确保多次编译的代码能正确访问和修改共享的全局变量（用于 REPL 等场景）。
//
// 参数：
//   - script: JPL 源代码字符串
//   - filename: 源文件名（用于错误报告和调试）
//   - existingGlobals: 已存在的全局变量名列表（可为 nil）
//
// 返回值：
//   - *Program: 编译后的程序
//   - nil: 编译成功
//   - error: 词法/语法/编译错误
//
// 使用示例（REPL 场景）：
//
//	// 第一轮编译
//	prog1, _ := engine.CompileStringWithGlobals("a = 10", "<repl>", nil)
//	vm.Execute()
//
//	// 获取全局变量名
//	names := vm.GetGlobalNames()
//
//	// 第二轮编译（知道 a 的索引）
//	prog2, _ := engine.CompileStringWithGlobals("b = a + 5", "<repl>", names)
//	vm.SetProgram(prog2)
//	vm.Execute() // 现在 b = 15
func CompileStringWithGlobals(script string, filename string, existingGlobals []string) (*Program, error) {
	l := lexer.NewLexer(script, filename)
	p := parser.NewParser(l)
	program := p.Parse()

	// 检查解析错误
	if len(p.Errors()) > 0 {
		msg := ""
		for _, e := range p.Errors() {
			msg += e + "\n"
		}
		return nil, &CompileError{Message: msg}
	}

	var c *Compiler
	if len(existingGlobals) > 0 {
		c = NewCompilerWithGlobals(existingGlobals)
	} else {
		c = NewCompiler()
	}

	// Phase 7.8: 设置魔术常量所需信息
	c.filename = filename
	c.dirname = getDirFromFilename(filename)

	// 捕获编译 panic 并转换为 error
	var compileErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case *CompileError:
					compileErr = e
				case error:
					compileErr = &CompileError{Message: e.Error()}
				default:
					compileErr = &CompileError{Message: fmt.Sprintf("%v", r)}
				}
			}
		}()
		c.compileProgram(program)
	}()

	if compileErr != nil {
		return nil, compileErr
	}

	return c.buildProgram(), nil
}

// ============================================================================
// Phase 7.8: 魔术常量支持
// ============================================================================

// getMagicConstant 返回魔术常量的值，如果不是魔术常量返回 nil
//
// 支持的魔术常量：
//   - __FILE__: 当前编译的源文件名
//   - __DIR__: 当前编译的源文件所在目录
//   - __LINE__: 当前代码行号
//   - __TIME__: 编译时间（格式：HH:MM:SS）
//   - __DATE__: 编译日期（格式：Mon Jan 2 2006）
//   - __OS__: 操作系统（runtime.GOOS）
//   - JPL_VERSION: JPL 版本号
func (c *Compiler) getMagicConstant(name string, pos token.Position) Value {
	switch name {
	case "__FILE__":
		filename := c.filename
		if filename == "" {
			filename = "[stdin]"
		}
		return NewString(filename)
	case "__DIR__":
		dirname := c.dirname
		if dirname == "" {
			dirname = "."
		}
		return NewString(dirname)
	case "__LINE__":
		return NewInt(int64(pos.Line))
	case "__TIME__":
		if c.compileTime == "" {
			c.compileTime = time.Now().Format("15:04:05")
		}
		return NewString(c.compileTime)
	case "__DATE__":
		if c.compileDate == "" {
			c.compileDate = time.Now().Format("Jan _2 2006")
		}
		return NewString(c.compileDate)
	case "__OS__":
		return NewString(runtime.GOOS)
	case "JPL_VERSION":
		return NewString("1.0.0")
	default:
		return nil
	}
}

// buildProgram 从编译器构建最终的 Program
func (c *Compiler) buildProgram() *Program {
	main := &CompiledFunction{
		Name:        "<main>",
		Params:      0,
		Registers:   c.maxReg,
		Bytecode:    c.bytecode,
		Constants:   c.constants,
		SourceLine:  0,
		SourceLines: c.sourceLines,
		VarNames:    c.buildVarNames(),
	}

	allFuncs := []*CompiledFunction{main}
	allFuncs = append(allFuncs, c.functions...)

	return &Program{
		Main:        main,
		Functions:   allFuncs,
		Constants:   c.constants,
		GlobalNames: *c.globalNames,
	}
}

// getDirFromFilename 从文件名提取目录部分
// 如果文件名为空或为 [stdin]，返回 "."
func getDirFromFilename(filename string) string {
	if filename == "" || filename == "[stdin]" {
		return "."
	}
	// 找到最后一个路径分隔符
	lastSlash := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '/' || filename[i] == '\\' {
			lastSlash = i
			break
		}
	}
	if lastSlash < 0 {
		return "."
	}
	if lastSlash == 0 {
		return "/"
	}
	return filename[:lastSlash]
}
