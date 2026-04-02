package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestParseObjectBasic 测试基本对象解析
func TestParseObjectBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "简单对象",
			input:    "{a: 1, b: 2}",
			expected: map[string]string{"a": "1", "b": "2"},
		},
		{
			name:     "字符串值",
			input:    "{name: 'John', age: 30}",
			expected: map[string]string{"name": "John", "age": "30"},
		},
		{
			name:     "空对象",
			input:    "{}",
			expected: map[string]string{},
		},
		{
			name:     "尾随逗号",
			input:    "{a: 1,}",
			expected: map[string]string{"a": "1"},
		},
		{
			name:     "字符串键",
			input:    `{"key with spaces": "value"}`,
			expected: map[string]string{"key with spaces": "value"},
		},
		{
			name:     "布尔值和 null",
			input:    "{active: true, deleted: false, data: null}",
			expected: map[string]string{"active": "true", "deleted": "false", "data": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("parse_object", engine.NewString(tt.input))
			if err != nil {
				t.Fatalf("parse_object() error: %v", err)
			}

			if result.Type() != engine.TypeObject {
				t.Errorf("Expected object type, got %s", result.Type())
				return
			}

			obj := result.Object()
			if len(obj) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(obj))
			}

			for key, expectedVal := range tt.expected {
				val, ok := obj[key]
				if !ok {
					t.Errorf("Key '%s' not found", key)
					continue
				}
				if val.String() != expectedVal {
					t.Errorf("Key '%s': expected %q, got %q", key, expectedVal, val.String())
				}
			}
		})
	}
}

// TestParseObjectNested 测试嵌套对象和数组
func TestParseObjectNested(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result engine.Value)
	}{
		{
			name:  "嵌套对象",
			input: "{user: {name: 'John', age: 30}}",
			check: func(t *testing.T, result engine.Value) {
				obj := result.Object()
				user, ok := obj["user"]
				if !ok {
					t.Error("Key 'user' not found")
					return
				}
				if user.Type() != engine.TypeObject {
					t.Errorf("Expected nested object, got %s", user.Type())
					return
				}
				userObj := user.Object()
				if userObj["name"].String() != "John" {
					t.Errorf("Expected name='John', got %s", userObj["name"].String())
				}
			},
		},
		{
			name:  "包含数组的对象",
			input: "{items: [1, 2, 3]}",
			check: func(t *testing.T, result engine.Value) {
				obj := result.Object()
				items, ok := obj["items"]
				if !ok {
					t.Error("Key 'items' not found")
					return
				}
				if items.Type() != engine.TypeArray {
					t.Errorf("Expected array, got %s", items.Type())
					return
				}
				arr := items.Array()
				if len(arr) != 3 {
					t.Errorf("Expected 3 items, got %d", len(arr))
				}
			},
		},
		{
			name:  "数组中的对象",
			input: "{users: [{name: 'A'}, {name: 'B'}]}",
			check: func(t *testing.T, result engine.Value) {
				obj := result.Object()
				users := obj["users"].Array()
				if len(users) != 2 {
					t.Errorf("Expected 2 users, got %d", len(users))
					return
				}
				firstUser := users[0].Object()
				if firstUser["name"].String() != "A" {
					t.Errorf("Expected first user name='A', got %s", firstUser["name"].String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := callBuiltin("parse_object", engine.NewString(tt.input))
			if err != nil {
				t.Fatalf("parse_object() error: %v", err)
			}
			tt.check(t, result)
		})
	}
}

// TestParseObjectSecurity 测试安全限制（应该拒绝的内容）
func TestParseObjectSecurity(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "函数调用",
			input: "{result: delete_all_files()}",
		},
		{
			name:  "变量引用",
			input: "{value: $x}",
		},
		{
			name:  "表达式",
			input: "{result: 1 + 2}",
		},
		{
			name:  "非对象输入",
			input: "123",
		},
		{
			name:  "代码块",
			input: "{x: {exit(1)}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := callBuiltin("parse_object", engine.NewString(tt.input))
			if err == nil {
				t.Errorf("Expected error for input: %s", tt.input)
			} else {
				t.Logf("Got expected error: %v", err)
			}
		})
	}
}

// TestParseObjectErrors 测试错误处理
func TestParseObjectErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []engine.Value
		expectError bool
	}{
		{
			name:        "无参数",
			args:        []engine.Value{},
			expectError: true,
		},
		{
			name:        "非字符串参数",
			args:        []engine.Value{engine.NewInt(123)},
			expectError: true,
		},
		{
			name:        "多个参数",
			args:        []engine.Value{engine.NewString("{}"), engine.NewString("{}")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := callBuiltin("parse_object", tt.args...)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
