package stdlib

import (
	"strings"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestJSONFunctionsExist 测试 JSON 函数已注册
func TestJSONFunctionsExist(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)

	// 检查函数名列表
	names := FunctionNames()
	hasJSONEncode := false
	hasJSONDecode := false
	hasJSONPretty := false

	for _, name := range names {
		if name == "json_encode" {
			hasJSONEncode = true
		}
		if name == "json_decode" {
			hasJSONDecode = true
		}
		if name == "json_pretty" {
			hasJSONPretty = true
		}
	}

	if !hasJSONEncode {
		t.Error("Missing json_encode function")
	}
	if !hasJSONDecode {
		t.Error("Missing json_decode function")
	}
	if !hasJSONPretty {
		t.Error("Missing json_pretty function")
	}
}

// TestJSONEncode 测试编码功能
func TestJSONEncode(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)

	// 编译并执行
	compiled, err := engine.CompileString(`json_encode([1, 2, 3])`)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(e, compiled)
	if err := vm.Execute(); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result := vm.GetResult()
	str := result.Stringify()

	// 应该返回 JSON 格式的字符串
	if !strings.Contains(str, "1") || !strings.Contains(str, "2") || !strings.Contains(str, "3") {
		t.Errorf("json_encode result should contain array elements: %s", str)
	}
}

// TestJSONPretty 测试美化输出
func TestJSONPretty(t *testing.T) {
	e := engine.NewEngine()
	RegisterAll(e)

	compiled, err := engine.CompileString(`json_pretty({"a": 1})`)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(e, compiled)
	if err := vm.Execute(); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	result := vm.GetResult()
	str := result.Stringify()

	// 美化输出应该更长（包含换行和空格）
	if len(str) < 10 {
		t.Errorf("json_pretty result too short: %s", str)
	}
}

// TestJSONNumberParsing 测试数字解析（包括 BigInt 和 BigDecimal）
func TestJSONNumberParsing(t *testing.T) {
	// 直接测试 parseJSONNumber 函数
	testCases := []struct {
		name         string
		input        string
		expectedType engine.ValueType
		checkValue   func(engine.Value) bool
	}{
		{
			name:         "普通整数",
			input:        "123",
			expectedType: engine.TypeInt,
			checkValue:   func(v engine.Value) bool { return v.Int() == 123 },
		},
		{
			name:         "负数",
			input:        "-456",
			expectedType: engine.TypeInt,
			checkValue:   func(v engine.Value) bool { return v.Int() == -456 },
		},
		{
			name:         "零",
			input:        "0",
			expectedType: engine.TypeInt,
			checkValue:   func(v engine.Value) bool { return v.Int() == 0 },
		},
		{
			name:         "大整数（超出 int64 范围）",
			input:        "999999999999999999999999999999",
			expectedType: engine.TypeBigInt,
			checkValue: func(v engine.Value) bool {
				str := v.String()
				return len(str) > 19
			},
		},
		{
			name:         "普通小数",
			input:        "1.5",
			expectedType: engine.TypeFloat,
			checkValue:   func(v engine.Value) bool { return v.Float() == 1.5 },
		},
		{
			name:         "科学计数法整数",
			input:        "1e10",
			expectedType: engine.TypeInt,
			checkValue:   func(v engine.Value) bool { return v.Int() == 10000000000 },
		},
		{
			name:         "科学计数法大整数",
			input:        "1e20",
			expectedType: engine.TypeBigInt,
			checkValue: func(v engine.Value) bool {
				str := v.String()
				return str == "100000000000000000000"
			},
		},
		{
			name:         "科学计数法负数",
			input:        "-1e10",
			expectedType: engine.TypeInt,
			checkValue:   func(v engine.Value) bool { return v.Int() == -10000000000 },
		},
		{
			name:         "科学计数法带小数",
			input:        "1.5e-5",
			expectedType: engine.TypeFloat,
			checkValue:   func(v engine.Value) bool { return v.Float() == 0.000015 },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseJSONNumber(tc.input)
			if result.Type() != tc.expectedType {
				t.Errorf("Expected type %v, got %v", tc.expectedType, result.Type())
			}
			if !tc.checkValue(result) {
				t.Errorf("Value check failed for input: %s, got value: %v, string: %s",
					tc.input, result, result.String())
			}
		})
	}
}

// TestJSONDecodeObject 测试 json_decode 是否正确返回对象类型
func TestJSONDecodeObject(t *testing.T) {
	// 测试直接调用 builtinJSONDecode
	result := goValueToJPL(map[string]any{"name": "Atom"})

	if result.Type() != engine.TypeObject {
		t.Errorf("Expected type object, got %v", result.Type())
	}

	// 检查是否能访问属性
	if result.Type() == engine.TypeObject {
		obj := result.Object()
		if val, ok := obj["name"]; !ok {
			t.Error("Object should have 'name' key")
		} else if val.String() != "Atom" {
			t.Errorf("Expected 'Atom', got %q", val.String())
		}
	}
}
