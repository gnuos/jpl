package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestCwd 测试 cwd 函数
func TestCwd(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	fn := e.GetRegisteredFunc("cwd")
	if fn == nil {
		t.Fatal("cwd function not registered")
	}

	// Test calling cwd
	ctx := engine.NewContext(e, nil)
	result, err := fn(ctx, []engine.Value{})
	if err != nil {
		t.Fatalf("cwd() failed: %v", err)
	}

	if result.Type() != engine.TypeString {
		t.Errorf("cwd() should return string, got %s", result.Type())
	}

	// Should return a non-empty path
	if result.String() == "" {
		t.Error("cwd() returned empty string")
	}

	t.Logf("cwd() returned: %s", result.String())
}

// TestNowEnhanced 测试增强版 now 函数
func TestNowEnhanced(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	ctx := engine.NewContext(e, nil)

	// Test now() without args
	result, err := builtinNow(ctx, []engine.Value{})
	if err != nil {
		t.Fatalf("now() failed: %v", err)
	}

	if result.Type() != engine.TypeObject {
		t.Errorf("now() should return object, got %s", result.Type())
	}

	// Check that new fields exist
	obj := result.Object()
	requiredFields := []string{"year", "month", "day", "hour", "minute", "second",
		"millisecond", "weekday", "timezone", "timestamp"}
	for _, field := range requiredFields {
		if _, ok := obj[field]; !ok {
			t.Errorf("now() result missing field: %s", field)
		}
	}

	// Test now() with format string
	formatResult, err := builtinNow(ctx, []engine.Value{engine.NewString("Y-m-d")})
	if err != nil {
		t.Fatalf("now(\"Y-m-d\") failed: %v", err)
	}

	if formatResult.Type() != engine.TypeString {
		t.Errorf("now(\"Y-m-d\") should return string, got %s", formatResult.Type())
	}

	// Check format contains dashes (date format)
	formatted := formatResult.String()
	if len(formatted) < 8 {
		t.Errorf("Formatted date too short: %s", formatted)
	}

	t.Logf("now() object has %d fields", len(obj))
	t.Logf("now(\"Y-m-d\") = %s", formatted)
}
