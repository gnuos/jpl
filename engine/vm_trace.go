package engine

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ============================================================================
// 追踪事件类型
// ============================================================================

// TraceEvent 追踪事件类型
type TraceEvent int

const (
	TraceEventInstruction TraceEvent = iota // 指令执行前
	TraceEventCall                          // 函数调用
	TraceEventReturn                        // 函数返回
	TraceEventJump                          // 跳转指令
)

// String 返回事件类型名称
func (e TraceEvent) String() string {
	switch e {
	case TraceEventInstruction:
		return "EXEC"
	case TraceEventCall:
		return "CALL"
	case TraceEventReturn:
		return "RET"
	case TraceEventJump:
		return "JUMP"
	default:
		return "UNKNOWN"
	}
}

// ============================================================================
// 追踪上下文
// ============================================================================

// TraceContext 追踪上下文，包含当前执行状态信息
type TraceContext struct {
	Event       TraceEvent  // 事件类型
	VM          *VM         // 虚拟机实例
	IP          int         // 当前指令指针
	Instruction Instruction // 当前指令
	Opcode      Opcode      // 操作码
}

// ============================================================================
// 追踪钩子
// ============================================================================

// TraceHook 追踪钩子函数类型
// 返回 true 继续执行，false 暂停执行（用于断点等）
type TraceHook func(ctx *TraceContext) bool

// ============================================================================
// 追踪配置
// ============================================================================

// TraceConfig 追踪配置
type TraceConfig struct {
	Enabled   bool      // 是否启用追踪
	Writer    io.Writer // 输出目标（默认 os.Stdout）
	ShowRegs  bool      // 显示寄存器状态
	ShowStack bool      // 显示调用栈
	Hook      TraceHook // 自定义钩子（可选）
	MaxLines  int       // 最大输出行数（0=无限制）
	lineCount int       // 内部计数器
}

// NewTraceConfig 创建默认追踪配置
func NewTraceConfig() *TraceConfig {
	return &TraceConfig{
		Enabled:  true,
		Writer:   os.Stdout,
		ShowRegs: true,
	}
}

// NewTraceConfigWithWriter 创建带输出目标的追踪配置
func NewTraceConfigWithWriter(w io.Writer) *TraceConfig {
	return &TraceConfig{
		Enabled:  true,
		Writer:   w,
		ShowRegs: true,
	}
}

// ============================================================================
// VM 追踪方法
// ============================================================================

// SetTraceConfig 设置追踪配置
func (vm *VM) SetTraceConfig(config *TraceConfig) {
	// 直接在 VM 上添加字段，不影响现有逻辑
	vm.traceConfig = config
}

// GetTraceConfig 获取追踪配置
func (vm *VM) GetTraceConfig() *TraceConfig {
	return vm.traceConfig
}

// EnableTrace 启用追踪（使用默认配置）
func (vm *VM) EnableTrace() {
	if vm.traceConfig == nil {
		vm.traceConfig = NewTraceConfig()
	}
	vm.traceConfig.Enabled = true
}

// DisableTrace 禁用追踪
func (vm *VM) DisableTrace() {
	if vm.traceConfig != nil {
		vm.traceConfig.Enabled = false
	}
}

// SetTraceHook 设置自定义追踪钩子
func (vm *VM) SetTraceHook(hook TraceHook) {
	if vm.traceConfig == nil {
		vm.traceConfig = NewTraceConfig()
	}
	vm.traceConfig.Hook = hook
}

// ============================================================================
// 内部追踪方法
// ============================================================================

// trace 执行追踪（在指令执行前调用）
func (vm *VM) trace(ins Instruction, op Opcode) {
	if vm.traceConfig == nil || !vm.traceConfig.Enabled {
		return
	}

	// 检查最大行数限制
	if vm.traceConfig.MaxLines > 0 {
		vm.traceConfig.lineCount++
		if vm.traceConfig.lineCount > vm.traceConfig.MaxLines {
			return
		}
	}

	// 构建追踪上下文
	ctx := &TraceContext{
		Event:       TraceEventInstruction,
		VM:          vm,
		IP:          vm.ip - 1, // ip 已经 +1，所以减 1
		Instruction: ins,
		Opcode:      op,
	}

	// 调用自定义钩子
	if vm.traceConfig.Hook != nil {
		if !vm.traceConfig.Hook(ctx) {
			return
		}
	}

	// 默认输出
	w := vm.traceConfig.Writer
	if w == nil {
		w = os.Stdout
	}

	// 输出指令信息
	disasm := disassembleInstruction(ins, vm.function.Constants)
	fmt.Fprintf(w, "[TRACE] ip=%04d | %-30s", ctx.IP, disasm)

	// 显示寄存器状态
	if vm.traceConfig.ShowRegs {
		fmt.Fprintf(w, " | %s", vm.dumpRegisters())
	}

	fmt.Fprintln(w)
}

// traceCall 追踪函数调用
func (vm *VM) traceCall(funcName string, argCount int) {
	if vm.traceConfig == nil || !vm.traceConfig.Enabled {
		return
	}

	w := vm.traceConfig.Writer
	if w == nil {
		w = os.Stdout
	}

	fmt.Fprintf(w, "[TRACE] CALL %s(%d args)\n", funcName, argCount)
}

// traceReturn 追踪函数返回
func (vm *VM) traceReturn(value Value) {
	if vm.traceConfig == nil || !vm.traceConfig.Enabled {
		return
	}

	w := vm.traceConfig.Writer
	if w == nil {
		w = os.Stdout
	}

	fmt.Fprintf(w, "[TRACE] RETURN %s\n", value.Stringify())
}

// ============================================================================
// 状态转储方法
// ============================================================================

// dumpRegisters 格式化寄存器状态
func (vm *VM) dumpRegisters() string {
	var parts []string
	for i, val := range vm.registers {
		if !val.IsNull() {
			parts = append(parts, fmt.Sprintf("R%d=%s", i, val.Stringify()))
		}
	}
	if len(parts) == 0 {
		return "R: all null"
	}
	return strings.Join(parts, " ")
}

// DumpRegisters 打印所有寄存器状态（公共方法）
func (vm *VM) DumpRegisters() string {
	var buf strings.Builder
	buf.WriteString("Registers:\n")
	for i, val := range vm.registers {
		if !val.IsNull() {
			buf.WriteString(fmt.Sprintf("  R%d = %s (%s)\n", i, val.Stringify(), val.Type()))
		}
	}
	return buf.String()
}

// DumpCallStack 打印调用栈
func (vm *VM) DumpCallStack() string {
	var buf strings.Builder
	buf.WriteString("Call Stack:\n")
	if len(vm.callStack) == 0 {
		buf.WriteString("  (empty)\n")
		return buf.String()
	}
	for i := len(vm.callStack) - 1; i >= 0; i-- {
		frame := vm.callStack[i]
		fnName := "<main>"
		if frame.function != nil {
			fnName = frame.function.Name
		}
		buf.WriteString(fmt.Sprintf("  #%d %s (ip=%d)\n", i, fnName, frame.ip))
	}
	return buf.String()
}

// DumpGlobals 打印全局变量
func (vm *VM) DumpGlobals() string {
	var buf strings.Builder
	buf.WriteString("Globals:\n")
	if len(vm.globals) == 0 {
		buf.WriteString("  (empty)\n")
		return buf.String()
	}
	for name, idx := range vm.globalNames {
		val := vm.globals[idx]
		buf.WriteString(fmt.Sprintf("  %s = %s (%s)\n", name, val.Stringify(), val.Type()))
	}
	return buf.String()
}

// DumpState 打印完整 VM 状态
func (vm *VM) DumpState() string {
	var buf strings.Builder
	buf.WriteString("=== VM State ===\n")
	buf.WriteString(fmt.Sprintf("Function: %s\n", vm.function.Name))
	buf.WriteString(fmt.Sprintf("IP: %d\n", vm.ip))
	buf.WriteString(fmt.Sprintf("Call Depth: %d\n", vm.callDepth))
	buf.WriteString("\n")
	buf.WriteString(vm.DumpRegisters())
	buf.WriteString("\n")
	buf.WriteString(vm.DumpCallStack())
	buf.WriteString("\n")
	buf.WriteString(vm.DumpGlobals())
	buf.WriteString("================\n")
	return buf.String()
}

// ============================================================================
// 便捷追踪函数
// ============================================================================

// ============================================================================
// 调试 API — 结构化数据类型
// ============================================================================

// VarInfo 单个变量信息
type VarInfo struct {
	Name  string    // 变量名
	Value Value     // 变量值
	Type  ValueType // 值类型
}

// LocalsInfo 局部变量信息
type LocalsInfo struct {
	Function string    // 当前函数名
	Vars     []VarInfo // 变量列表
}

// GlobalsInfo 全局变量信息
type GlobalsInfo struct {
	Vars []VarInfo // 变量列表
}

// GetLocals 获取当前局部变量的结构化数据
func (vm *VM) GetLocals() LocalsInfo {
	info := LocalsInfo{
		Function: "<main>",
	}
	if vm.function != nil {
		info.Function = vm.function.Name
	}
	if vm.function == nil || len(vm.function.VarNames) == 0 {
		return info
	}
	for i, name := range vm.function.VarNames {
		if i >= len(vm.registers) {
			break
		}
		if name == "" {
			continue
		}
		val := vm.registers[i]
		info.Vars = append(info.Vars, VarInfo{
			Name:  name,
			Value: val,
			Type:  val.Type(),
		})
	}
	return info
}

// GetGlobals 获取全局变量的结构化数据
func (vm *VM) GetGlobals() GlobalsInfo {
	info := GlobalsInfo{}
	for name, idx := range vm.globalNames {
		val := vm.globals[idx]
		info.Vars = append(info.Vars, VarInfo{
			Name:  name,
			Value: val,
			Type:  val.Type(),
		})
	}
	return info
}

// FormatLocals 格式化输出局部变量
func (vm *VM) FormatLocals() string {
	info := vm.GetLocals()
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Locals (%s):\n", info.Function))
	if len(info.Vars) == 0 {
		buf.WriteString("  (empty)\n")
		return buf.String()
	}
	for _, v := range info.Vars {
		buf.WriteString(fmt.Sprintf("  %s = %s (%s)\n", v.Name, v.Value.Stringify(), v.Type))
	}
	return buf.String()
}

// FormatGlobals 格式化输出全局变量
func (vm *VM) FormatGlobals() string {
	info := vm.GetGlobals()
	var buf strings.Builder
	buf.WriteString("Globals:\n")
	if len(info.Vars) == 0 {
		buf.WriteString("  (empty)\n")
		return buf.String()
	}
	for _, v := range info.Vars {
		buf.WriteString(fmt.Sprintf("  %s = %s (%s)\n", v.Name, v.Value.Stringify(), v.Type))
	}
	return buf.String()
}

// TraceToBuffer 创建一个输出到 buffer 的追踪配置
func TraceToBuffer() (*TraceConfig, *strings.Builder) {
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	return config, &buf
}

// TraceWithLimit 创建有行数限制的追踪配置
func TraceWithLimit(maxLines int) *TraceConfig {
	config := NewTraceConfig()
	config.MaxLines = maxLines
	return config
}
