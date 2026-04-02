package engine

import (
	"strings"
	"testing"
)

// ============================================================================
// 追踪功能测试
// ============================================================================

func TestVMTraceBasic(t *testing.T) {
	code := `
	$x = 10;
	$y = 20;
	$z = $x + $y;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	// 启用追踪
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[TRACE]") {
		t.Error("追踪输出应包含 [TRACE]")
	}
	if !strings.Contains(output, "ip=") {
		t.Error("追踪输出应包含指令指针")
	}
	if !strings.Contains(output, "LOADK") {
		t.Error("追踪输出应包含 LOADK 指令")
	}

	t.Logf("追踪输出:\n%s", output)
}

func TestVMTraceWithRegisters(t *testing.T) {
	code := `
	$x = 42;
	$y = $x + 1;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	// 启用追踪并显示寄存器
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	config.ShowRegs = true
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "R") {
		t.Error("追踪输出应包含寄存器状态")
	}

	t.Logf("带寄存器的追踪输出:\n%s", output)
}

func TestVMTraceCustomHook(t *testing.T) {
	code := `
	$i = 0;
	while ($i < 3) {
		$i = $i + 1;
	}
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	// 使用自定义钩子计数指令
	instructionCount := 0
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	config.Hook = func(ctx *TraceContext) bool {
		instructionCount++
		// 只记录跳转指令
		if ctx.Opcode == OP_JMPIF || ctx.Opcode == OP_JMPIFNOT || ctx.Opcode == OP_JMP {
			buf.WriteString("  [HOOK] Jump detected!\n")
		}
		return true
	}
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if instructionCount == 0 {
		t.Error("自定义钩子应被执行")
	}

	t.Logf("自定义钩子执行次数: %d", instructionCount)
	t.Logf("钩子输出:\n%s", buf.String())
}

func TestVMTraceMaxLines(t *testing.T) {
	code := `
	$x = 1;
	$y = 2;
	$z = 3;
	$a = 4;
	$b = 5;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	// 限制输出行数
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	config.MaxLines = 3
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// 注意：隐式 RETURN 会增加指令，MaxLines 应该限制输出
	if len(lines) > 4 {
		t.Errorf("最大行数限制应生效，实际输出 %d 行", len(lines))
	}

	t.Logf("限制行数的追踪输出 (最多3行):\n%s", buf.String())
}

func TestVMTraceDisabled(t *testing.T) {
	code := `
	$x = 10;
	$y = 20;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	// 禁用追踪
	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	config.Enabled = false
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	if buf.Len() > 0 {
		t.Error("禁用追踪时不应有输出")
	}
}

// ============================================================================
// 状态转储测试
// ============================================================================

func TestVMDumpRegisters(t *testing.T) {
	code := `
	$x = 10;
	$y = "hello";
	$z = true;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	dump := vm.DumpRegisters()
	if !strings.Contains(dump, "Registers:") {
		t.Error("寄存器转储应包含标题")
	}

	t.Logf("寄存器转储:\n%s", dump)
}

func TestVMDumpCallStack(t *testing.T) {
	code := `
	fn add($a, $b) {
		return $a + $b;
	}
	$result = add(1, 2);
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	dump := vm.DumpCallStack()
	if !strings.Contains(dump, "Call Stack:") {
		t.Error("调用栈转储应包含标题")
	}

	t.Logf("调用栈转储:\n%s", dump)
}

func TestVMDumpGlobals(t *testing.T) {
	code := `
	$global_var = 100;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	dump := vm.DumpGlobals()
	if !strings.Contains(dump, "Globals:") {
		t.Error("全局变量转储应包含标题")
	}

	t.Logf("全局变量转储:\n%s", dump)
}

func TestVMDumpState(t *testing.T) {
	code := `
	$x = 42;
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	dump := vm.DumpState()
	if !strings.Contains(dump, "VM State") {
		t.Error("状态转储应包含标题")
	}
	if !strings.Contains(dump, "Function:") {
		t.Error("状态转储应包含函数信息")
	}

	t.Logf("完整状态转储:\n%s", dump)
}

// ============================================================================
// 追踪函数调用测试
// ============================================================================

func TestVMTraceFunctionCall(t *testing.T) {
	code := `
	fn multiply($a, $b) {
		return $a * $b;
	}
	$result = multiply(6, 7);
	$result
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)

	var buf strings.Builder
	config := NewTraceConfigWithWriter(&buf)
	config.ShowRegs = true
	vm.SetTraceConfig(config)

	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "CALL multiply") {
		t.Error("追踪输出应包含函数调用")
	}
	if !strings.Contains(output, "RETURN") {
		t.Error("追踪输出应包含函数返回")
	}

	t.Logf("函数调用追踪:\n%s", output)
}

// ============================================================================
// 结构化 Dump API 测试
// ============================================================================

func TestVMGetLocals(t *testing.T) {
	code := `$x = 10; $name = "hello"; $flag = true;`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	info := vm.GetLocals()
	if info.Function != "<main>" {
		t.Errorf("期望函数名 <main>，得到 %s", info.Function)
	}
	if len(info.Vars) != 3 {
		t.Fatalf("期望 3 个局部变量，得到 %d", len(info.Vars))
	}

	// 构建名称→值映射方便验证
	varMap := make(map[string]VarInfo)
	for _, v := range info.Vars {
		varMap[v.Name] = v
	}

	if v, ok := varMap["$x"]; !ok || v.Value.Int() != 10 {
		t.Errorf("$x 期望值 10")
	}
	if v, ok := varMap["$name"]; !ok || v.Value.String() != "hello" {
		t.Errorf("$name 期望值 hello")
	}
	if v, ok := varMap["$flag"]; !ok || v.Value.Bool() != true {
		t.Errorf("$flag 期望值 true")
	}

	t.Logf("局部变量:\n%s", vm.FormatLocals())
}

func TestVMGetGlobals(t *testing.T) {
	code := `$g1 = 42; $g2 = "world";`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	err = vm.Execute()
	if err != nil {
		t.Fatalf("执行失败: %v", err)
	}

	info := vm.GetGlobals()
	if len(info.Vars) != 2 {
		t.Fatalf("期望 2 个全局变量，得到 %d", len(info.Vars))
	}

	varMap := make(map[string]VarInfo)
	for _, v := range info.Vars {
		varMap[v.Name] = v
	}

	if v, ok := varMap["$g1"]; !ok || v.Value.Int() != 42 {
		t.Errorf("$g1 期望值 42")
	}
	if v, ok := varMap["$g2"]; !ok || v.Value.String() != "world" {
		t.Errorf("$g2 期望值 world")
	}

	t.Logf("全局变量:\n%s", vm.FormatGlobals())
}

func TestVMFormatLocalsInFunction(t *testing.T) {
	code := `
	fn add($a, $b) {
		_sum = $a + $b;
		return _sum;
	}
	`
	engine := NewEngine()
	prog, err := CompileString(code)
	if err != nil {
		t.Fatalf("编译失败: %v", err)
	}

	vm := newVMWithProgram(engine, prog)
	// 执行前查看 main 的局部变量
	info := vm.GetLocals()
	if info.Function != "<main>" {
		t.Errorf("期望函数名 <main>，得到 %s", info.Function)
	}

	// 检查 add 函数的 VarNames
	addFns := vm.funcMap["add"]
	if len(addFns) == 0 {
		t.Fatal("未找到函数 add")
	}
	addFn := addFns[0]
	if len(addFn.VarNames) == 0 {
		t.Error("函数 add 的 VarNames 不应为空")
	}
	t.Logf("add 函数 VarNames: %v", addFn.VarNames)
}
