package stdlib

import (
	"strings"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestReMatch 测试正则匹配
func TestReMatch(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReMatch(nil, []engine.Value{
		engine.NewString(`\d+`),
		engine.NewString("abc123"),
	})
	if err != nil {
		t.Fatalf("re_match() error = %v", err)
	}
	if !result.Bool() {
		t.Error("re_match(`\\d+`, `abc123`) should return true")
	}

	result, err = builtinReMatch(nil, []engine.Value{
		engine.NewString(`^\d+$`),
		engine.NewString("abc123"),
	})
	if err != nil {
		t.Fatalf("re_match() error = %v", err)
	}
	if result.Bool() {
		t.Error("re_match(`^\\d+$`, `abc123`) should return false")
	}

	t.Log("re_match: basic tests passed")
}

// TestReMatchInvalidPattern 测试无效正则
func TestReMatchInvalidPattern(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	_, err := builtinReMatch(nil, []engine.Value{
		engine.NewString(`[`),
		engine.NewString("test"),
	})
	if err == nil {
		t.Error("re_match() with invalid pattern should return error")
	} else {
		t.Logf("re_match() invalid pattern error (expected): %v", err)
	}
}

// TestReSearch 测试查找第一个匹配
func TestReSearch(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReSearch(nil, []engine.Value{
		engine.NewString(`[\w.-]+@[\w.-]+\.\w+`),
		engine.NewString("Contact: john@example.com or jane@test.org"),
	})
	if err != nil {
		t.Fatalf("re_search() error = %v", err)
	}
	if result.String() != "john@example.com" {
		t.Errorf("re_search() = %s, expected john@example.com", result.String())
	}

	result, err = builtinReSearch(nil, []engine.Value{
		engine.NewString(`\d+`),
		engine.NewString("no numbers here"),
	})
	if err != nil {
		t.Fatalf("re_search() error = %v", err)
	}
	if result.String() != "" {
		t.Errorf("re_search() no match = %s, expected empty string", result.String())
	}

	t.Log("re_search: tests passed")
}

// TestReFindall 测试查找所有匹配
func TestReFindall(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReFindall(nil, []engine.Value{
		engine.NewString(`\d+`),
		engine.NewString("Room 101, Floor 5, Building 3"),
	})
	if err != nil {
		t.Fatalf("re_findall() error = %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Errorf("re_findall() returned %d matches, expected 3", len(arr))
	}

	expected := []string{"101", "5", "3"}
	for i, exp := range expected {
		if i < len(arr) && arr[i].String() != exp {
			t.Errorf("re_findall()[%d] = %s, expected %s", i, arr[i].String(), exp)
		}
	}

	t.Log("re_findall: tests passed")
}

// TestReSub 测试替换
func TestReSub(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReSub(nil, []engine.Value{
		engine.NewString(`\d+`),
		engine.NewString("[NUM]"),
		engine.NewString("Room 101, Floor 5"),
	})
	if err != nil {
		t.Fatalf("re_sub() error = %v", err)
	}

	if !strings.Contains(result.String(), "[NUM]") {
		t.Errorf("re_sub() = %s, should contain [NUM]", result.String())
	}

	t.Log("re_sub: tests passed")
}

// TestReSplit 测试分割
func TestReSplit(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReSplit(nil, []engine.Value{
		engine.NewString(`\s*,\s*`),
		engine.NewString("apple, banana ,orange"),
	})
	if err != nil {
		t.Fatalf("re_split() error = %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Errorf("re_split() returned %d parts, expected 3", len(arr))
	}

	expected := []string{"apple", "banana", "orange"}
	for i, exp := range expected {
		if i < len(arr) && arr[i].String() != exp {
			t.Errorf("re_split()[%d] = %s, expected %s", i, arr[i].String(), exp)
		}
	}

	t.Log("re_split: tests passed")
}

// TestReGroups 测试捕获组
func TestReGroups(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterRe(eng)

	result, err := builtinReGroups(nil, []engine.Value{
		engine.NewString(`(\d{4})-(\d{2})`),
		engine.NewString("Date: 2024-03-15"),
	})
	if err != nil {
		t.Fatalf("re_groups() error = %v", err)
	}

	obj := result.Object()
	if obj["0"].String() != "2024-03" {
		t.Errorf("Full match = %s, expected 2024-03", obj["0"].String())
	}
	if obj["1"].String() != "2024" {
		t.Errorf("Group 1 = %s, expected 2024", obj["1"].String())
	}
	if obj["2"].String() != "03" {
		t.Errorf("Group 2 = %s, expected 03", obj["2"].String())
	}

	result, err = builtinReGroups(nil, []engine.Value{
		engine.NewString(`(?P<year>\d{4})-(?P<month>\d{2})`),
		engine.NewString("2024-03"),
	})
	if err != nil {
		t.Fatalf("re_groups() with named groups error = %v", err)
	}

	obj = result.Object()
	if obj["year"].String() != "2024" {
		t.Errorf("Named group 'year' = %s, expected 2024", obj["year"].String())
	}
	if obj["month"].String() != "03" {
		t.Errorf("Named group 'month' = %s, expected 03", obj["month"].String())
	}

	result, err = builtinReGroups(nil, []engine.Value{
		engine.NewString(`\d+`),
		engine.NewString("no numbers"),
	})
	if err != nil {
		t.Fatalf("re_groups() error = %v", err)
	}
	if !result.IsNull() {
		t.Error("re_groups() with no match should return null")
	}

	t.Log("re_groups: tests passed")
}
