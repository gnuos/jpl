package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/lexer"
	"github.com/gnuos/jpl/parser"
)

// TestTypeCastSyntax 测试类型转换语法
func TestTypeCastSyntax(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"int from string", `$x = int("42");`, "42"},
		{"int from float truncates", `$x = int(3.7);`, "3"},
		{"int from negative float", `$x = int(-3.7);`, "-3"},
		{"int from true", `$x = int(true);`, "1"},
		{"int from false", `$x = int(false);`, "0"},
		{"int from null", `$x = int(null);`, "0"},
		{"int from invalid string", `$x = int("abc");`, "0"},
		{"int from empty string", `$x = int("");`, "0"},

		{"float from string", `$x = float("3.14");`, "3.14"},
		{"float from int", `$x = float(42);`, "42"},
		{"float from true", `$x = float(true);`, "1"},
		{"float from false", `$x = float(false);`, "0"},
		{"float from null", `$x = float(null);`, "0"},

		{"string from int", `$x = string(123);`, "123"},
		{"string from float", `$x = string(3.14);`, "3.14"},
		{"string from true", `$x = string(true);`, "true"},
		{"string from false", `$x = string(false);`, "false"},
		{"string from null", `$x = string(null);`, ""},

		{"bool from non-zero int", `$x = bool(42);`, "true"},
		{"bool from zero int", `$x = bool(0);`, "false"},
		{"bool from non-empty string", `$x = bool("hello");`, "true"},
		{"bool from empty string", `$x = bool("");`, "false"},
		{"bool from non-empty array", `$x = bool([1, 2]);`, "true"},
		{"bool from empty array", `$x = bool([]);`, "false"},
		{"bool from null", `$x = bool(null);`, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析代码
			l := lexer.NewLexer(tt.code, "test.jpl")
			p := parser.NewParser(l)
			program := p.Parse()
			if len(p.Errors()) > 0 {
				t.Fatalf("Parse errors: %v", p.Errors())
			}

			// 编译
			compiled, err := engine.Compile(program)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			// 执行
			e := engine.NewEngine()
			RegisterAll(e)
			defer e.Close()
			vm := engine.NewVMWithProgram(e, compiled)
			err = vm.Execute()
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			// 获取变量值（变量名带$前缀）
			varValue, ok := vm.GetGlobal("$x")
			if !ok {
				t.Fatalf("Variable $x not found")
			}

			got := varValue.String()
			if got != tt.expected {
				t.Errorf("TypeCast result = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestTypeCastInExpression 测试在表达式中使用类型转换
func TestTypeCastInExpression(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int64
	}{
		{"int in arithmetic", `$x = int("10") + 5;`, 15},
		{"int in comparison", `$x = int("5") == 5 ? 1 : 0;`, 1},
		{"nested cast", `$x = int(float("3.7"));`, 3},
		{"string concat after int cast", `$x = int("10") + 20;`, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析代码
			l := lexer.NewLexer(tt.code, "test.jpl")
			p := parser.NewParser(l)
			program := p.Parse()
			if len(p.Errors()) > 0 {
				t.Fatalf("Parse errors: %v", p.Errors())
			}

			// 编译
			compiled, err := engine.Compile(program)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}

			// 执行
			e := engine.NewEngine()
			RegisterAll(e)
			defer e.Close()
			vm := engine.NewVMWithProgram(e, compiled)
			err = vm.Execute()
			if err != nil {
				t.Fatalf("Runtime error: %v", err)
			}

			// 获取变量值（变量名带$前缀）
			varValue, ok := vm.GetGlobal("$x")
			if !ok {
				t.Fatalf("Variable $x not found")
			}

			if varValue.Type() != engine.TypeInt {
				t.Errorf("Expected int type, got %s", varValue.Type())
			}

			got := varValue.Int()
			if got != tt.expected {
				t.Errorf("TypeCast result = %d, expected %d", got, tt.expected)
			}
		})
	}
}

// TestTypeCastAST 测试 TypeCast AST 节点
func TestTypeCastAST(t *testing.T) {
	code := `$x = int("42");`

	l := lexer.NewLexer(code, "test.jpl")
	p := parser.NewParser(l)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse errors: %v", p.Errors())
	}

	// 检查 AST 结构
	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	// 获取变量声明语句
	varDecl, ok := program.Statements[0].(*parser.VarDecl)
	if !ok {
		t.Fatalf("Expected VarDecl, got %T", program.Statements[0])
	}

	// 检查初始值是 TypeCast
	typeCast, ok := varDecl.Value.(*parser.TypeCast)
	if !ok {
		t.Fatalf("Expected TypeCast, got %T", varDecl.Value)
	}

	if typeCast.Type != "int" {
		t.Errorf("Expected cast type 'int', got '%s'", typeCast.Type)
	}

	// 检查被转换的表达式
	strLit, ok := typeCast.Expr.(*parser.StringLiteral)
	if !ok {
		t.Fatalf("Expected StringLiteral in TypeCast.Expr, got %T", typeCast.Expr)
	}

	if strLit.Value != "42" {
		t.Errorf("Expected string '42', got '%s'", strLit.Value)
	}
}
