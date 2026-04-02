package engine

import (
	"testing"
)

// ============================================================================
// 测试辅助函数
// ============================================================================

func compileOrFail(t *testing.T, script string) *Program {
	t.Helper()
	prog, err := CompileString(script)
	if err != nil {
		t.Fatalf("编译失败: %v\n脚本: %s", err, script)
	}
	return prog
}

func getMain(t *testing.T, prog *Program) *CompiledFunction {
	t.Helper()
	return prog.Main
}

// 验证指令
func assertInstruction(t *testing.T, fn *CompiledFunction, pc int, expectedOP Opcode, desc string) {
	t.Helper()
	if pc >= len(fn.Bytecode) {
		t.Fatalf("%s: PC %d 越界（共 %d 条指令）", desc, pc, len(fn.Bytecode))
	}
	ins := fn.Bytecode[pc]
	if ins.OP() != expectedOP {
		t.Errorf("%s: PC %d 期望 %s, 实际 %s", desc, pc, expectedOP, ins.OP())
	}
}

func assertInstructionABC(t *testing.T, fn *CompiledFunction, pc int, op Opcode, a, b, c int, desc string) {
	t.Helper()
	assertInstruction(t, fn, pc, op, desc)
	ins := fn.Bytecode[pc]
	if ins.A() != a || ins.B() != b || ins.C() != c {
		t.Errorf("%s: PC %d 期望 A=%d B=%d C=%d, 实际 A=%d B=%d C=%d",
			desc, pc, a, b, c, ins.A(), ins.B(), ins.C())
	}
}

func assertInstructionABx(t *testing.T, fn *CompiledFunction, pc int, op Opcode, a, bx int, desc string) {
	t.Helper()
	assertInstruction(t, fn, pc, op, desc)
	ins := fn.Bytecode[pc]
	if ins.A() != a || ins.Bx() != bx {
		t.Errorf("%s: PC %d 期望 A=%d Bx=%d, 实际 A=%d Bx=%d",
			desc, pc, a, bx, ins.A(), ins.Bx())
	}
}

func assertConstantInt(t *testing.T, fn *CompiledFunction, idx int, expected int64, desc string) {
	t.Helper()
	if idx >= len(fn.Constants) {
		t.Fatalf("%s: 常量索引 %d 越界", desc, idx)
	}
	v := fn.Constants[idx]
	if v.Type() != TypeInt || v.Int() != expected {
		t.Errorf("%s: 期望 int(%d), 实际 %v", desc, expected, v)
	}
}

func assertConstantString(t *testing.T, fn *CompiledFunction, idx int, expected string, desc string) {
	t.Helper()
	if idx >= len(fn.Constants) {
		t.Fatalf("%s: 常量索引 %d 越界", desc, idx)
	}
	v := fn.Constants[idx]
	if v.Type() != TypeString || v.String() != expected {
		t.Errorf("%s: 期望 string(%q), 实际 %v", desc, expected, v)
	}
}

func assertRegisters(t *testing.T, fn *CompiledFunction, expected int, desc string) {
	t.Helper()
	if fn.Registers != expected {
		t.Errorf("%s: 期望 %d 个寄存器, 实际 %d", desc, expected, fn.Registers)
	}
}

// ============================================================================
// 数字字面量
// =============================================================================

func TestCompileNumberLiteral(t *testing.T) {
	prog := compileOrFail(t, "42;")
	fn := getMain(t, prog)

	// LOADK R0, <42>
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "数字字面量")
	assertConstantInt(t, fn, 0, 42, "常量值")
	assertRegisters(t, fn, 1, "寄存器数")
}

func TestCompileFloatLiteral(t *testing.T) {
	prog := compileOrFail(t, "3.14;")
	fn := getMain(t, prog)

	assertInstruction(t, fn, 0, OP_LOADK, "浮点字面量")
	if fn.Constants[0].Float() != 3.14 {
		t.Errorf("期望 3.14, 实际 %f", fn.Constants[0].Float())
	}
}

// ============================================================================
// 字符串字面量
// =============================================================================

func TestCompileStringLiteral(t *testing.T) {
	prog := compileOrFail(t, `"hello";`)
	fn := getMain(t, prog)

	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "字符串字面量")
	assertConstantString(t, fn, 0, "hello", "字符串常量")
}

// ============================================================================
// 布尔和 null 字面量
// =============================================================================

func TestCompileBoolLiteral(t *testing.T) {
	prog := compileOrFail(t, "true;")
	fn := getMain(t, prog)

	assertInstructionABC(t, fn, 0, OP_LOADBOOL, 0, 1, 0, "true")
}

func TestCompileNullLiteral(t *testing.T) {
	prog := compileOrFail(t, "null;")
	fn := getMain(t, prog)

	assertInstructionABC(t, fn, 0, OP_LOADNULL, 0, 0, 0, "null")
}

// ============================================================================
// 算术运算
// =============================================================================

func TestCompileAddition(t *testing.T) {
	// 1 + 2 = 3，常量折叠
	prog := compileOrFail(t, "1 + 2;")
	fn := getMain(t, prog)
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 3")
}

func TestCompileSubtraction(t *testing.T) {
	// 10 - 3 = 7，常量折叠
	prog := compileOrFail(t, "10 - 3;")
	fn := getMain(t, prog)
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 7")
}

func TestCompileMultiplication(t *testing.T) {
	// 3 * 4 = 12，常量折叠
	prog := compileOrFail(t, "3 * 4;")
	fn := getMain(t, prog)
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 12")
}

func TestCompileDivision(t *testing.T) {
	// 10 / 2 = 5，常量折叠
	prog := compileOrFail(t, "10 / 2;")
	fn := getMain(t, prog)
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 5")
}

func TestCompileModulo(t *testing.T) {
	// 10 % 3 = 1，常量折叠为单条 LOADK
	prog := compileOrFail(t, "10 % 3;")
	fn := getMain(t, prog)

	// 常量池: [0]=1
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 1")
}

func TestCompileNestedArithmetic(t *testing.T) {
	// 使用变量避免常量折叠，测试实际的 MUL 指令生成
	prog := compileOrFail(t, "($a + $b) * $c;")
	fn := getMain(t, prog)

	// R1=$a, R2=$b, R3=R1+R2, R4=$c, R0=R3*R4
	// 指令: LOAD a, LOAD b, ADD, LOAD c, MUL, RETURN
	foundMul := false
	for _, inst := range fn.Bytecode {
		if inst.OP() == OP_MUL {
			foundMul = true
		}
	}
	if !foundMul {
		t.Error("嵌套算术表达式缺少 MUL 指令")
	}
}

// ============================================================================
// 比较运算
// =============================================================================

func TestCompileComparison(t *testing.T) {
	tests := []struct {
		script string
		op     Opcode
		desc   string
	}{
		// 使用变量避免常量折叠
		{"$a == $b;", OP_EQ, "等于"},
		{"$a != $b;", OP_NEQ, "不等于"},
		{"$a < $b;", OP_LT, "小于"},
		{"$a > $b;", OP_GT, "大于"},
		{"$a <= $b;", OP_LTE, "小于等于"},
		{"$a >= $b;", OP_GTE, "大于等于"},
	}

	for _, tt := range tests {
		prog := compileOrFail(t, tt.script)
		fn := getMain(t, prog)
		found := false
		for _, inst := range fn.Bytecode {
			if inst.OP() == tt.op {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: 缺少 %s 指令", tt.desc, tt.op)
		}
	}
}

// ============================================================================
// 一元运算
// =============================================================================

func TestCompileUnaryNegate(t *testing.T) {
	// 使用变量避免常量折叠
	prog := compileOrFail(t, "-$a;")
	fn := getMain(t, prog)

	foundNeg := false
	for _, inst := range fn.Bytecode {
		if inst.OP() == OP_NEG {
			foundNeg = true
		}
	}
	if !foundNeg {
		t.Error("取反表达式缺少 NEG 指令")
	}
}

func TestCompileUnaryNot(t *testing.T) {
	// 使用变量避免常量折叠
	prog := compileOrFail(t, "!$a;")
	fn := getMain(t, prog)

	foundNot := false
	for _, inst := range fn.Bytecode {
		if inst.OP() == OP_NOT {
			foundNot = true
		}
	}
	if !foundNot {
		t.Error("逻辑非表达式缺少 NOT 指令")
	}
}

// ============================================================================
// 变量声明
// =============================================================================

func TestCompileVarDecl(t *testing.T) {
	prog := compileOrFail(t, "$x = 42;")
	fn := getMain(t, prog)

	// R0 = 42
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "变量声明加载常量")
	assertConstantInt(t, fn, 0, 42, "常量 42")
	// 注意：隐式 RETURN 会增加寄存器
	assertRegisters(t, fn, 2, "一个局部变量 + 隐式 RETURN")
}

func TestCompileVarDeclExpression(t *testing.T) {
	prog := compileOrFail(t, "$x = 1 + 2;")
	fn := getMain(t, prog)

	// 常量折叠：1 + 2 = 3，编译为单条 LOADK
	// 常量池: [0]=3, [1]="$x"
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "加载折叠常量 3")
	// 全局索引：$x 是第一个全局变量，索引为 0
	assertInstructionABx(t, fn, 1, OP_SETGLOBAL, 0, 0, "SETGLOBAL $x")
}

// ============================================================================
// 变量引用
// =============================================================================

func TestCompileVarReference(t *testing.T) {
	prog := compileOrFail(t, "$x = 10; $y = $x;")
	fn := getMain(t, prog)

	// 常量池: [0]=10
	// 全局索引: $x=0, $y=1
	// $x = 10: R0 = 10, SETGLOBAL $x (索引 0)
	assertInstructionABx(t, fn, 0, OP_LOADK, 0, 0, "$x = 10")
	assertInstructionABx(t, fn, 1, OP_SETGLOBAL, 0, 0, "SETGLOBAL $x")
	// $y = $x: R1 = R0, SETGLOBAL $y (索引 1)
	assertInstructionABC(t, fn, 2, OP_LOAD, 1, 0, 0, "拷贝 $x 到 $y")
	assertInstructionABx(t, fn, 3, OP_SETGLOBAL, 1, 1, "SETGLOBAL $y")
	// 注意：隐式 RETURN 会增加寄存器
	assertRegisters(t, fn, 3, "两个局部变量 + 隐式 RETURN")
}

func TestCompileVarInExpression(t *testing.T) {
	prog := compileOrFail(t, "$x = 10; $y = $x + 5;")
	fn := getMain(t, prog)

	// 常量池: [0]=10, [1]="$x", [2]=5, [3]="$y"
	// $x = 10
	assertInstruction(t, fn, 0, OP_LOADK, "$x = 10")
	assertInstruction(t, fn, 1, OP_SETGLOBAL, "SETGLOBAL $x")
	// $y = $x + 5
	assertInstruction(t, fn, 2, OP_LOAD, "加载 $x")
	assertInstruction(t, fn, 3, OP_LOADK, "加载 5")
	assertInstruction(t, fn, 4, OP_ADD, "$x + 5")
	assertInstruction(t, fn, 5, OP_SETGLOBAL, "SETGLOBAL $y")
}

// ============================================================================
// if/else
// =============================================================================

func TestCompileIfStatement(t *testing.T) {
	prog := compileOrFail(t, "if (true) { $x = 1; }")
	fn := getMain(t, prog)

	assertInstruction(t, fn, 0, OP_LOADBOOL, "条件 true")
	assertInstruction(t, fn, 1, OP_JMPIFNOT, "条件跳转")
	assertInstruction(t, fn, 2, OP_LOADK, "if 体: $x = 1")
	assertInstruction(t, fn, 3, OP_SETGLOBAL, "SETGLOBAL $x")

	// 跳转偏移应指向 if 体之后（跳过 LOADK + SETGLOBAL 共 2 条）
	ins := fn.Bytecode[1]
	if ins.AsBx() != 2 {
		t.Errorf("JMPIFNOT 偏移期望 2, 实际 %d", ins.AsBx())
	}
}

func TestCompileIfElseStatement(t *testing.T) {
	prog := compileOrFail(t, "if (true) { $x = 1; } else { $x = 2; }")
	fn := getMain(t, prog)

	assertInstruction(t, fn, 0, OP_LOADBOOL, "条件")
	assertInstruction(t, fn, 1, OP_JMPIFNOT, "条件跳转")
	assertInstruction(t, fn, 2, OP_LOADK, "if 体")
	assertInstruction(t, fn, 3, OP_SETGLOBAL, "SETGLOBAL if 体 $x")
	assertInstruction(t, fn, 4, OP_JMP, "跳过 else")
	assertInstruction(t, fn, 5, OP_LOADK, "else 体")
	assertInstruction(t, fn, 6, OP_SETGLOBAL, "SETGLOBAL else 体 $x")
}

// ============================================================================
// while 循环
// =============================================================================

func TestCompileWhileLoop(t *testing.T) {
	prog := compileOrFail(t, "$i = 0; while ($i < 10) { $i = $i + 1; }")
	fn := getMain(t, prog)

	// 验证存在 JMP 指令跳回循环开始
	foundJMP := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMP && ins.AsBx() < 0 {
			foundJMP = true
			break
		}
	}
	if !foundJMP {
		t.Error("未找到跳回循环开始的 JMP 指令")
	}
}

// ============================================================================
// for 循环
// =============================================================================

func TestCompileForLoop(t *testing.T) {
	prog := compileOrFail(t, "for ($i = 0; $i < 10; $i = $i + 1) { }")
	fn := getMain(t, prog)

	// 验证存在条件跳转和回跳
	hasCondJump := false
	hasBackJump := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMPIFNOT {
			hasCondJump = true
		}
		if ins.OP() == OP_JMP && ins.AsBx() < 0 {
			hasBackJump = true
		}
	}
	if !hasCondJump {
		t.Error("未找到条件跳转指令")
	}
	if !hasBackJump {
		t.Error("未找到回跳指令")
	}
}

// ============================================================================
// break/continue
// =============================================================================

func TestCompileBreakContinue(t *testing.T) {
	prog := compileOrFail(t, `
		$i = 0;
		while ($i < 10) {
			if ($i == 5) {
				break;
			}
			$i = $i + 1;
		}
	`)
	fn := getMain(t, prog)

	// 验证存在多个 JMP 指令（break 和循环回跳）
	jmpCount := 0
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMP {
			jmpCount++
		}
	}
	if jmpCount < 2 {
		t.Errorf("期望至少 2 条 JMP, 实际 %d", jmpCount)
	}
}

// ============================================================================
// 字符串连接
// =============================================================================

func TestCompileStringConcat(t *testing.T) {
	// 使用变量避免常量折叠
	prog := compileOrFail(t, `$a .. $b;`)
	fn := getMain(t, prog)

	foundConcat := false
	for _, inst := range fn.Bytecode {
		if inst.OP() == OP_CONCAT {
			foundConcat = true
		}
	}
	if !foundConcat {
		t.Error("字符串连接缺少 CONCAT 指令")
	}
}

// ============================================================================
// 三元表达式
// =============================================================================

func TestCompileTernary(t *testing.T) {
	// 使用变量条件避免常量折叠
	prog := compileOrFail(t, "$x = $cond ? 1 : 2;")
	fn := getMain(t, prog)

	// 验证存在 JMPIFNOT 和 JMP
	hasCondJump := false
	hasJump := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMPIFNOT {
			hasCondJump = true
		}
		if ins.OP() == OP_JMP {
			hasJump = true
		}
	}
	if !hasCondJump {
		t.Error("三元表达式缺少条件跳转")
	}
	if !hasJump {
		t.Error("三元表达式缺少跳过跳转")
	}
}

// ============================================================================
// 函数声明
// =============================================================================

func TestCompileFuncDecl(t *testing.T) {
	prog := compileOrFail(t, `
		fn add($a, $b) {
			return $a + $b;
		}
	`)
	fn := getMain(t, prog)

	// 验证主函数加载了函数（使用 CLOSURE 指令）
	assertInstruction(t, fn, 0, OP_CLOSURE, "加载函数闭包")

	// 验证函数列表
	if len(prog.Functions) < 2 {
		t.Fatalf("期望至少 2 个函数（main + add），实际 %d", len(prog.Functions))
	}

	addFn := prog.Functions[1]
	if addFn.Name != "add" {
		t.Errorf("期望函数名 add, 实际 %s", addFn.Name)
	}
	if addFn.Params != 2 {
		t.Errorf("期望 2 个参数, 实际 %d", addFn.Params)
	}
	if addFn.Registers < 2 {
		t.Errorf("期望至少 2 个寄存器, 实际 %d", addFn.Registers)
	}

	// 验证函数体：ADD R2, R0, R1; RETURN R2（或类似）
	hasAdd := false
	hasReturn := false
	for _, ins := range addFn.Bytecode {
		if ins.OP() == OP_ADD {
			hasAdd = true
		}
		if ins.OP() == OP_RETURN {
			hasReturn = true
		}
	}
	if !hasAdd {
		t.Error("函数体缺少 ADD 指令")
	}
	if !hasReturn {
		t.Error("函数体缺少 RETURN 指令")
	}
}

func TestCompileFuncParamNames(t *testing.T) {
	prog := compileOrFail(t, `
		fn greet($name, $greeting) {
			return $greeting + " " + $name;
		}
	`)

	// 验证函数列表
	if len(prog.Functions) < 2 {
		t.Fatalf("期望至少 2 个函数（main + greet），实际 %d", len(prog.Functions))
	}

	greetFn := prog.Functions[1]
	if greetFn.Name != "greet" {
		t.Errorf("期望函数名 greet, 实际 %s", greetFn.Name)
	}

	// 验证 ParamNames（包含 $ 前缀）
	if len(greetFn.ParamNames) != 2 {
		t.Fatalf("期望 2 个参数名, 实际 %d", len(greetFn.ParamNames))
	}
	if greetFn.ParamNames[0] != "$name" {
		t.Errorf("期望第一个参数名 $name, 实际 %s", greetFn.ParamNames[0])
	}
	if greetFn.ParamNames[1] != "$greeting" {
		t.Errorf("期望第二个参数名 $greeting, 实际 %s", greetFn.ParamNames[1])
	}
}

// ============================================================================
// 函数调用
// =============================================================================

func TestCompileFunctionCall(t *testing.T) {
	prog := compileOrFail(t, `
		fn double($x) {
			return $x * 2;
		}
		$result = double(5);
	`)
	fn := getMain(t, prog)

	// 验证存在 CALL 指令
	hasCall := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_CALL {
			hasCall = true
			break
		}
	}
	if !hasCall {
		t.Error("未找到 CALL 指令")
	}
}

// ============================================================================
// 数组字面量
// =============================================================================

func TestCompileArrayLiteral(t *testing.T) {
	prog := compileOrFail(t, "$arr = [1, 2, 3];")
	fn := getMain(t, prog)

	hasNewArray := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_NEWARRAY {
			hasNewArray = true
			break
		}
	}
	if !hasNewArray {
		t.Error("未找到 NEWARRAY 指令")
	}
}

func TestCompileEmptyArray(t *testing.T) {
	prog := compileOrFail(t, "$arr = [];")
	fn := getMain(t, prog)

	assertInstructionABC(t, fn, 0, OP_NEWARRAY, 0, 0, 0, "空数组")
}

// ============================================================================
// 对象字面量
// =============================================================================

func TestCompileObjectLiteral(t *testing.T) {
	prog := compileOrFail(t, `$obj = {"name": "test", "value": 42};`)
	fn := getMain(t, prog)

	hasNewObj := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_NEWOBJECT {
			hasNewObj = true
			break
		}
	}
	if !hasNewObj {
		t.Error("未找到 NEWOBJECT 指令")
	}
}

func TestCompileEmptyObject(t *testing.T) {
	prog := compileOrFail(t, "$obj = {};")
	fn := getMain(t, prog)

	assertInstructionABC(t, fn, 0, OP_NEWOBJECT, 0, 0, 0, "空对象")
}

// ============================================================================
// 索引访问
// =============================================================================

func TestCompileIndexAccess(t *testing.T) {
	prog := compileOrFail(t, "$arr = [1, 2, 3]; $x = $arr[0];")
	fn := getMain(t, prog)

	hasGetIndex := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_GETINDEX {
			hasGetIndex = true
			break
		}
	}
	if !hasGetIndex {
		t.Error("未找到 GETINDEX 指令")
	}
}

func TestCompileIndexAssign(t *testing.T) {
	prog := compileOrFail(t, "$arr = [1, 2, 3]; $arr[0] = 99;")
	fn := getMain(t, prog)

	hasSetIndex := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_SETINDEX {
			hasSetIndex = true
			break
		}
	}
	if !hasSetIndex {
		t.Error("未找到 SETINDEX 指令")
	}
}

// ============================================================================
// 成员访问
// =============================================================================

func TestCompileMemberAccess(t *testing.T) {
	prog := compileOrFail(t, `$obj = {"x": 10}; $y = $obj.x;`)
	fn := getMain(t, prog)

	hasGetMember := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_GETMEMBER {
			hasGetMember = true
			break
		}
	}
	if !hasGetMember {
		t.Error("未找到 GETMEMBER 指令")
	}
}

// ============================================================================
// 逻辑运算（短路求值）
// =============================================================================

func TestCompileShortCircuitAnd(t *testing.T) {
	prog := compileOrFail(t, "$x = true && false;")
	fn := getMain(t, prog)

	hasJMPIFNOT := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMPIFNOT {
			hasJMPIFNOT = true
			break
		}
	}
	if !hasJMPIFNOT {
		t.Error("&& 短路求值缺少 JMPIFNOT")
	}
}

func TestCompileShortCircuitOr(t *testing.T) {
	prog := compileOrFail(t, "$x = true || false;")
	fn := getMain(t, prog)

	hasJMPIF := false
	for _, ins := range fn.Bytecode {
		if ins.OP() == OP_JMPIF {
			hasJMPIF = true
			break
		}
	}
	if !hasJMPIF {
		t.Error("|| 短路求值缺少 JMPIF")
	}
}

// ============================================================================
// 多语句程序
// =============================================================================

func TestCompileMultipleStatements(t *testing.T) {
	prog := compileOrFail(t, `
		$a = 1;
		$b = 2;
		$c = $a + $b;
		$d = $c * 3;
	`)
	fn := getMain(t, prog)

	// 4 个局部变量 + 表达式求值的临时寄存器
	if fn.Registers < 4 {
		t.Errorf("期望至少 4 个寄存器, 实际 %d", fn.Registers)
	}
}

// ============================================================================
// 作用域
// =============================================================================

func TestCompileBlockScope(t *testing.T) {
	prog := compileOrFail(t, `
		$a = 1;
		{
			$b = 2;
		}
		$c = $a;
	`)
	fn := getMain(t, prog)

	// 块级作用域退出后，$b 的寄存器应该被回收
	// $a 在 R0, $c 应复用 R1（或类似的）
	if fn.Registers > 3 {
		t.Errorf("作用域回收后期望 <=3 寄存器, 实际 %d", fn.Registers)
	}
}

// ============================================================================
// 反编译器
// =============================================================================

func TestDisassemble(t *testing.T) {
	prog := compileOrFail(t, `$x = 1 + 2;`)
	output := Disassemble(prog.Main)

	if output == "" {
		t.Error("反编译输出为空")
	}
	t.Log(output)
}

func TestDisassembleProgram(t *testing.T) {
	prog := compileOrFail(t, `
		fn greet($name) {
			return "Hello, " .. $name;
		}
		$msg = greet("World");
	`)
	output := DisassembleProgram(prog)

	if output == "" {
		t.Error("程序反编译输出为空")
	}
	t.Log(output)
}

// ============================================================================
// Opcode.String()
// =============================================================================

func TestOpcodeString(t *testing.T) {
	if OP_ADD.String() != "ADD" {
		t.Errorf("OP_ADD.String() = %q, 期望 ADD", OP_ADD.String())
	}
	if OP_LOADK.String() != "LOADK" {
		t.Errorf("OP_LOADK.String() = %q, 期望 LOADK", OP_LOADK.String())
	}
}

// ============================================================================
// 指令格式
// =============================================================================

func TestInstructionFormat(t *testing.T) {
	// ABC 格式
	ins := NewABC(OP_ADD, 3, 5, 7)
	if ins.OP() != OP_ADD {
		t.Errorf("OP = %v, 期望 ADD", ins.OP())
	}
	if ins.A() != 3 || ins.B() != 5 || ins.C() != 7 {
		t.Errorf("A=%d B=%d C=%d, 期望 3 5 7", ins.A(), ins.B(), ins.C())
	}

	// ABx 格式
	ins = NewABx(OP_LOADK, 2, 100)
	if ins.A() != 2 || ins.Bx() != 100 {
		t.Errorf("A=%d Bx=%d, 期望 2 100", ins.A(), ins.Bx())
	}

	// AsBx 格式
	ins = NewAsBx(OP_JMP, 0, -5)
	if ins.AsBx() != -5 {
		t.Errorf("AsBx=%d, 期望 -5", ins.AsBx())
	}

	ins = NewAsBx(OP_JMPIFNOT, 1, 10)
	if ins.AsBx() != 10 {
		t.Errorf("AsBx=%d, 期望 10", ins.AsBx())
	}
}

// ============================================================================
// 运行时错误检查
// =============================================================================

func TestIsRuntimeError(t *testing.T) {
	err := newRuntimeError("test")
	if !IsRuntimeError(err) {
		t.Error("运行时错误检查失败")
	}
	if IsRuntimeError(NewInt(42)) {
		t.Error("整数不应是运行时错误")
	}
}

// ============================================================================
// 复杂场景
// =============================================================================

func TestCompileComplexProgram(t *testing.T) {
	prog := compileOrFail(t, `
		fn factorial($n) {
			if ($n <= 1) {
				return 1;
			}
			return $n * factorial($n - 1);
		}
		$result = factorial(5);
	`)

	if len(prog.Functions) < 2 {
		t.Fatalf("期望至少 2 个函数, 实际 %d", len(prog.Functions))
	}

	factFn := prog.Functions[1]

	// factorial 函数应该包含条件跳转、乘法、减法、递归调用
	hasCondJump := false
	hasMul := false
	hasSub := false
	hasCall := false
	for _, ins := range factFn.Bytecode {
		switch ins.OP() {
		case OP_JMPIFNOT:
			hasCondJump = true
		case OP_MUL:
			hasMul = true
		case OP_SUB:
			hasSub = true
		case OP_CALL:
			hasCall = true
		}
	}
	if !hasCondJump {
		t.Error("factorial 缺少条件跳转")
	}
	if !hasMul {
		t.Error("factorial 缺少乘法")
	}
	if !hasSub {
		t.Error("factorial 缺少减法")
	}
	if !hasCall {
		t.Error("factorial 缺少递归调用")
	}
}

// TestCompileImport 测试 import 语句编译
func TestCompileImport(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"import 语句", `import "math";`},
		{"from import 语句", `from "math" import sqrt, abs;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := compileOrFail(t, tt.script)
			main := getMain(t, prog)

			// 应有 OP_IMPORT 指令
			hasImport := false
			for _, ins := range main.Bytecode {
				if ins.OP() == OP_IMPORT {
					hasImport = true
					break
				}
			}
			if !hasImport {
				t.Error("缺少 OP_IMPORT 指令")
			}
		})
	}
}

// TestCompileInclude 测试 include 语句编译
func TestCompileInclude(t *testing.T) {
	tests := []struct {
		name   string
		script string
		once   bool
	}{
		{"include", `include "utils.jpl";`, false},
		{"include_once", `include_once "config.jpl";`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := compileOrFail(t, tt.script)
			main := getMain(t, prog)

			hasInclude := false
			for _, ins := range main.Bytecode {
				if ins.OP() == OP_INCLUDE {
					hasInclude = true
					if ins.A() != boolToInt(tt.once) {
						t.Errorf("A 期望 %d，得到 %d", boolToInt(tt.once), ins.A())
					}
					break
				}
			}
			if !hasInclude {
				t.Error("缺少 OP_INCLUDE 指令")
			}
		})
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ============================================================================
// 常量折叠测试
// ============================================================================

// TestConstantFoldingIntArith 整数算术折叠
func TestConstantFoldingIntArith(t *testing.T) {
	tests := []struct {
		expr string
		want int64
	}{
		{"3 + 4", 7},
		{"10 - 3", 7},
		{"3 * 4", 12},
		{"10 / 2", 5},
		{"10 % 3", 1},
		{"-5", -5},
		{"-(3 + 2)", -5},
		{"~0", -1},
		{"~(-1)", 0},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			prog := compileOrFail(t, "$x = "+tt.expr)
			main := getMain(t, prog)
			for _, inst := range main.Bytecode {
				op := inst.OP()
				if op == OP_ADD || op == OP_SUB || op == OP_MUL || op == OP_DIV || op == OP_MOD || op == OP_NEG {
					t.Errorf("表达式 %s 未被折叠，仍包含算术指令 %s", tt.expr, op)
				}
			}
		})
	}
}

// TestConstantFoldingNested 嵌套表达式折叠
func TestConstantFoldingNested(t *testing.T) {
	tests := []struct {
		expr string
		want int64
	}{
		{"3 + 4 * 2", 11},
		{"(3 + 4) * 2", 14},
		{"1 + 2 + 3 + 4", 10},
		{"2 * 3 * 4", 24},
		{"10 - 3 - 2", 5},
		{"2 + 3 * 4 - 1", 13},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			prog := compileOrFail(t, "$x = "+tt.expr)
			main := getMain(t, prog)
			for _, inst := range main.Bytecode {
				op := inst.OP()
				if op == OP_ADD || op == OP_SUB || op == OP_MUL || op == OP_DIV {
					t.Errorf("嵌套表达式 %s 未被折叠，包含 %s", tt.expr, op)
				}
			}
		})
	}
}

// TestConstantFoldingStringConcat 字符串拼接折叠
func TestConstantFoldingStringConcat(t *testing.T) {
	prog := compileOrFail(t, `$x = "hello" .. " " .. "world"`)
	main := getMain(t, prog)
	for _, inst := range main.Bytecode {
		if inst.OP() == OP_CONCAT {
			t.Error("字符串拼接未被折叠")
		}
	}
}

// TestConstantFoldingComparison 比较运算折叠
func TestConstantFoldingComparison(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"5 > 3"},
		{"3 < 5"},
		{"3 == 3"},
		{"3 != 4"},
		{"5 >= 5"},
		{"4 <= 4"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			prog := compileOrFail(t, "$x = "+tt.expr)
			main := getMain(t, prog)
			for _, inst := range main.Bytecode {
				op := inst.OP()
				if op == OP_EQ || op == OP_NEQ || op == OP_LT || op == OP_GT || op == OP_LTE || op == OP_GTE {
					t.Errorf("比较 %s 未被折叠", tt.expr)
				}
			}
		})
	}
}

// TestConstantFoldingBoolLogic 布尔逻辑折叠
func TestConstantFoldingBoolLogic(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"true && true"},
		{"true && false"},
		{"false || true"},
		{"false || false"},
		{"!true"},
		{"!false"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			prog := compileOrFail(t, "$x = "+tt.expr)
			main := getMain(t, prog)
			for _, inst := range main.Bytecode {
				op := inst.OP()
				if op == OP_AND || op == OP_OR || op == OP_NOT {
					t.Errorf("布尔逻辑 %s 未被折叠", tt.expr)
				}
			}
		})
	}
}

// TestConstantFoldingTernary 三元表达式折叠
func TestConstantFoldingTernary(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"true ? 1 : 2"},
		{"false ? 1 : 2"},
		{"(3 > 2) ? \"yes\" : \"no\""},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			prog := compileOrFail(t, "$x = "+tt.expr)
			main := getMain(t, prog)
			for _, inst := range main.Bytecode {
				op := inst.OP()
				if op == OP_JMP || op == OP_JMPIF || op == OP_JMPIFNOT {
					t.Errorf("三元表达式 %s 未被折叠，包含跳转指令", tt.expr)
				}
			}
		})
	}
}

// TestConstantFoldingNoFold 非常量不折叠
func TestConstantFoldingNoFold(t *testing.T) {
	prog := compileOrFail(t, "$x = $y + 3")
	main := getMain(t, prog)
	hasAdd := false
	for _, inst := range main.Bytecode {
		if inst.OP() == OP_ADD {
			hasAdd = true
		}
	}
	if !hasAdd {
		t.Error("含变量的表达式不应被折叠，但缺少 ADD 指令")
	}
}

// TestConstantFoldingDivZero 除零不折叠
func TestConstantFoldingDivZero(t *testing.T) {
	prog := compileOrFail(t, "$x = 5 / 0")
	main := getMain(t, prog)
	hasDiv := false
	for _, inst := range main.Bytecode {
		if inst.OP() == OP_DIV {
			hasDiv = true
		}
	}
	if !hasDiv {
		t.Error("除零不应被折叠，但缺少 DIV 指令")
	}
}
