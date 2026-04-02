package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// Error 函数测试
// ============================================================================

func getErrorField(t *testing.T, val engine.Value, field string) engine.Value {
	errObj := val.Object()
	if errObj == nil {
		t.Fatalf("expected error to have Object() method returning fields")
	}
	v, ok := errObj[field]
	if !ok {
		t.Fatalf("error field %q not found", field)
	}
	return v
}

func TestBuiltinErrorWithMessageOnly(t *testing.T) {
	result, err := callBuiltin("error", engine.NewString("something went wrong"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return an error object
	if result.Type() != engine.TypeError {
		t.Errorf("expected TypeError, got %v", result.Type())
	}

	// Check message
	msg := getErrorField(t, result, "message")
	if msg.String() != "something went wrong" {
		t.Errorf("expected message 'something went wrong', got '%s'", msg.String())
	}

	// Default code should be 0
	code := getErrorField(t, result, "code")
	if code.Int() != 0 {
		t.Errorf("expected code 0, got %d", code.Int())
	}

	// Default type should be "Error"
	errType := getErrorField(t, result, "type")
	if errType.String() != "Error" {
		t.Errorf("expected type 'Error', got '%s'", errType.String())
	}
}

func TestBuiltinErrorWithMessageAndCode(t *testing.T) {
	result, err := callBuiltin("error", engine.NewString("not found"), engine.NewInt(404))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "not found" {
		t.Errorf("expected message 'not found', got '%s'", msg.String())
	}

	code := getErrorField(t, result, "code")
	if code.Int() != 404 {
		t.Errorf("expected code 404, got %d", code.Int())
	}
}

func TestBuiltinErrorWithAllArgs(t *testing.T) {
	result, err := callBuiltin("error",
		engine.NewString("validation failed"),
		engine.NewInt(400),
		engine.NewString("ValidationError"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "validation failed" {
		t.Errorf("expected message 'validation failed', got '%s'", msg.String())
	}

	code := getErrorField(t, result, "code")
	if code.Int() != 400 {
		t.Errorf("expected code 400, got %d", code.Int())
	}

	errType := getErrorField(t, result, "type")
	if errType.String() != "ValidationError" {
		t.Errorf("expected type 'ValidationError', got '%s'", errType.String())
	}
}

func TestBuiltinErrorWithZeroCode(t *testing.T) {
	result, err := callBuiltin("error", engine.NewString("info"), engine.NewInt(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code := getErrorField(t, result, "code")
	if code.Int() != 0 {
		t.Errorf("expected code 0, got %d", code.Int())
	}
}

func TestBuiltinErrorWithNegativeCode(t *testing.T) {
	result, err := callBuiltin("error", engine.NewString("system error"), engine.NewInt(-1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code := getErrorField(t, result, "code")
	if code.Int() != -1 {
		t.Errorf("expected code -1, got %d", code.Int())
	}
}

func TestBuiltinErrorEmptyMessage(t *testing.T) {
	result, err := callBuiltin("error", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "" {
		t.Errorf("expected empty message, got '%s'", msg.String())
	}
}

func TestBuiltinErrorNoArgs(t *testing.T) {
	_, err := callBuiltin("error")
	if err == nil {
		t.Error("error() with no args should return error")
	}
}

func TestBuiltinErrorTooManyArgs(t *testing.T) {
	_, err := callBuiltin("error",
		engine.NewString("msg"),
		engine.NewInt(1),
		engine.NewString("type"),
		engine.NewString("extra"),
	)
	if err == nil {
		t.Error("error() with 4 args should return error")
	}
}

func TestBuiltinErrorTypeStringConversion(t *testing.T) {
	// Test that non-string args get converted via String()
	result, err := callBuiltin("error", engine.NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "42" {
		t.Errorf("expected message '42', got '%s'", msg.String())
	}
}

func TestBuiltinErrorCodeFloatConversion(t *testing.T) {
	// Float should be converted to int (truncated)
	result, err := callBuiltin("error",
		engine.NewString("test"),
		engine.NewFloat(500.7),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	code := getErrorField(t, result, "code")
	if code.Int() != 500 {
		t.Errorf("expected code 500 (truncated from 500.7), got %d", code.Int())
	}
}

func TestBuiltinErrorIntMessage(t *testing.T) {
	// Test with integer message
	result, err := callBuiltin("error", engine.NewInt(123))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "123" {
		t.Errorf("expected message '123', got '%s'", msg.String())
	}
}

func TestBuiltinErrorBoolMessage(t *testing.T) {
	// Test with boolean message
	result, err := callBuiltin("error", engine.NewBool(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := getErrorField(t, result, "message")
	if msg.String() != "true" {
		t.Errorf("expected message 'true', got '%s'", msg.String())
	}
}

func TestBuiltinErrorStringRepresentation(t *testing.T) {
	// Test String() output
	result, err := callBuiltin("error",
		engine.NewString("test error"),
		engine.NewInt(500),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	str := result.String()
	expected := "Error: test error (code: 500)"
	if str != expected {
		t.Errorf("expected string '%s', got '%s'", expected, str)
	}
}

func TestBuiltinErrorStringWithoutCode(t *testing.T) {
	// Test String() output without code
	result, err := callBuiltin("error", engine.NewString("simple error"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	str := result.String()
	expected := "Error: simple error"
	if str != expected {
		t.Errorf("expected string '%s', got '%s'", expected, str)
	}
}
