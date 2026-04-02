package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestBuiltinExit(t *testing.T) {
	tests := []struct {
		name     string
		args     []engine.Value
		wantCode int
	}{
		{"exit() - default code 0", []engine.Value{}, 0},
		{"exit(0)", []engine.Value{engine.NewInt(0)}, 0},
		{"exit(1)", []engine.Value{engine.NewInt(1)}, 1},
		{"exit(255)", []engine.Value{engine.NewInt(255)}, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &engine.Context{}
			_, err := builtinExit(ctx, tt.args)

			if err == nil {
				t.Error("expected ExitError, got nil")
				return
			}

			exitErr, ok := err.(*engine.ExitError)
			if !ok {
				t.Errorf("expected *ExitError, got %T: %v", err, err)
				return
			}

			if exitErr.Code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, exitErr.Code)
			}
		})
	}
}

func TestBuiltinDie(t *testing.T) {
	tests := []struct {
		name        string
		args        []engine.Value
		wantCode    int
		wantMessage string
	}{
		{"die() - no args", []engine.Value{}, 0, ""},
		{"die(msg)", []engine.Value{engine.NewString("Fatal error")}, 0, "Fatal error"},
		{"die(msg, code)", []engine.Value{engine.NewString("Fatal error"), engine.NewInt(2)}, 2, "Fatal error"},
		{"die(msg, 255)", []engine.Value{engine.NewString("Max code"), engine.NewInt(255)}, 255, "Max code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &engine.Context{}
			_, err := builtinDie(ctx, tt.args)

			if err == nil {
				t.Error("expected ExitError, got nil")
				return
			}

			exitErr, ok := err.(*engine.ExitError)
			if !ok {
				t.Errorf("expected *ExitError, got %T: %v", err, err)
				return
			}

			if exitErr.Code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, exitErr.Code)
			}

			if exitErr.Message != tt.wantMessage {
				t.Errorf("expected message %q, got %q", tt.wantMessage, exitErr.Message)
			}
		})
	}
}

func TestExitInScript(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterAll(e)

	// Test exit stops execution
	prog, err := engine.CompileStringWithName(`
		$x = 1
		exit(5)
		$x = 2  // This should not execute
	`, "<test>")
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	execErr := vm.Execute()

	if execErr != nil {
		t.Errorf("expected nil error (ExitError is not an error), got: %v", execErr)
	}

	exitCode := vm.GetExitCode()
	if exitCode != 5 {
		t.Errorf("expected exit code 5, got %d", exitCode)
	}
}

func TestExitIgnoresTryCatch(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterAll(e)

	// Test exit bypasses try/catch
	prog, err := engine.CompileStringWithName(`
		$caught = false
		try {
			exit(42)
		} catch ($e) {
			$caught = true
		}
	`, "<test>")
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	vm.Execute()

	exitCode := vm.GetExitCode()
	if exitCode != 42 {
		t.Errorf("expected exit code 42, got %d", exitCode)
	}

	// Check that catch block was NOT executed
	// (This would require more complex testing with global variable access)
}

func TestExitWithCondition(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterAll(e)

	// Test conditional exit
	prog, err := engine.CompileStringWithName(`
		$error = true
		if ($error) {
			exit(1)
		}
		$error = false
	`, "<test>")
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	vm := engine.NewVMWithProgram(e, prog)
	vm.Execute()

	exitCode := vm.GetExitCode()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
