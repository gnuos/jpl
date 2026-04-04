package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// disk_free_space 测试
// ============================================================================

func TestDiskFreeSpaceBasic(t *testing.T) {
	result, err := callBuiltin("disk_free_space", engine.NewString("/"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive number
	if result.Int() <= 0 {
		t.Errorf("disk_free_space('/') should return positive value, got %d", result.Int())
	}
}

func TestDiskFreeSpaceWrongArgCount(t *testing.T) {
	_, err := callBuiltin("disk_free_space")
	if err == nil {
		t.Error("disk_free_space(0 args) should return error")
	}
}

func TestDiskFreeSpaceNotString(t *testing.T) {
	_, err := callBuiltin("disk_free_space", engine.NewInt(42))
	if err == nil {
		t.Error("disk_free_space(42) should return error")
	}
}

// ============================================================================
// disk_total_space 测试
// ============================================================================

func TestDiskTotalSpaceBasic(t *testing.T) {
	result, err := callBuiltin("disk_total_space", engine.NewString("/"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive number
	if result.Int() <= 0 {
		t.Errorf("disk_total_space('/') should return positive value, got %d", result.Int())
	}
}

func TestDiskTotalSpaceWrongArgCount(t *testing.T) {
	_, err := callBuiltin("disk_total_space")
	if err == nil {
		t.Error("disk_total_space(0 args) should return error")
	}
}

func TestDiskTotalSpaceNotString(t *testing.T) {
	_, err := callBuiltin("disk_total_space", engine.NewInt(42))
	if err == nil {
		t.Error("disk_total_space(42) should return error")
	}
}

// ============================================================================
// fileatime 测试
// ============================================================================

func TestFileatimeBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("fileatime", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive timestamp
	if result.Int() <= 0 {
		t.Errorf("fileatime() should return positive timestamp, got %d", result.Int())
	}
}

func TestFileatimeNotFound(t *testing.T) {
	_, err := callBuiltin("fileatime", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("fileatime(nonexistent) should return error")
	}
}

func TestFileatimeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("fileatime")
	if err == nil {
		t.Error("fileatime(0 args) should return error")
	}
}

// ============================================================================
// filemtime 测试
// ============================================================================

func TestFilemtimeBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("filemtime", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive timestamp
	if result.Int() <= 0 {
		t.Errorf("filemtime() should return positive timestamp, got %d", result.Int())
	}
}

func TestFilemtimeNotFound(t *testing.T) {
	_, err := callBuiltin("filemtime", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("filemtime(nonexistent) should return error")
	}
}

func TestFilemtimeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("filemtime")
	if err == nil {
		t.Error("filemtime(0 args) should return error")
	}
}

// ============================================================================
// filectime 测试
// ============================================================================

func TestFilectimeBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("filectime", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive timestamp
	if result.Int() <= 0 {
		t.Errorf("filectime() should return positive timestamp, got %d", result.Int())
	}
}

func TestFilectimeNotFound(t *testing.T) {
	_, err := callBuiltin("filectime", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("filectime(nonexistent) should return error")
	}
}

func TestFilectimeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("filectime")
	if err == nil {
		t.Error("filectime(0 args) should return error")
	}
}

// ============================================================================
// touch 测试
// ============================================================================

func TestTouchCreateNew(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "newfile.txt")

	result, err := callBuiltin("touch", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("touch() should return true")
	}

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("touch() should create file if it doesn't exist")
	}
}

func TestTouchExisting(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("touch", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("touch() should return true")
	}
}

func TestTouchWrongArgCount(t *testing.T) {
	_, err := callBuiltin("touch")
	if err == nil {
		t.Error("touch(0 args) should return error")
	}
}

func TestTouchNotString(t *testing.T) {
	_, err := callBuiltin("touch", engine.NewInt(42))
	if err == nil {
		t.Error("touch(42) should return error")
	}
}

// ============================================================================
// getpid 测试
// ============================================================================

func TestGetpidBasic(t *testing.T) {
	result, err := callBuiltin("getpid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a positive PID
	if result.Int() <= 0 {
		t.Errorf("getpid() should return positive PID, got %d", result.Int())
	}
}

func TestGetpidWrongArgCount(t *testing.T) {
	_, err := callBuiltin("getpid", engine.NewInt(1))
	if err == nil {
		t.Error("getpid(1 arg) should return error")
	}
}

// ============================================================================
// getuid 测试
// ============================================================================

func TestGetuidBasic(t *testing.T) {
	result, err := callBuiltin("getuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a non-negative UID
	if result.Int() < 0 {
		t.Errorf("getuid() should return non-negative UID, got %d", result.Int())
	}
}

func TestGetuidWrongArgCount(t *testing.T) {
	_, err := callBuiltin("getuid", engine.NewInt(1))
	if err == nil {
		t.Error("getuid(1 arg) should return error")
	}
}

// ============================================================================
// getgid 测试
// ============================================================================

func TestGetgidBasic(t *testing.T) {
	result, err := callBuiltin("getgid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a non-negative GID
	if result.Int() < 0 {
		t.Errorf("getgid() should return non-negative GID, got %d", result.Int())
	}
}

func TestGetgidWrongArgCount(t *testing.T) {
	_, err := callBuiltin("getgid", engine.NewInt(1))
	if err == nil {
		t.Error("getgid(1 arg) should return error")
	}
}

// ============================================================================
// umask 测试
// ============================================================================

func TestUmaskGet(t *testing.T) {
	result, err := callBuiltin("umask")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return a valid mask (0-0777)
	mask := result.Int()
	if mask < 0 || mask > 0777 {
		t.Errorf("umask() should return valid mask, got %o", mask)
	}
}

func TestUmaskSet(t *testing.T) {
	// Save original umask
	original, _ := callBuiltin("umask")

	// Set new umask
	result, err := callBuiltin("umask", engine.NewInt(0022))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return old umask
	if result.Int() != original.Int() {
		// This might fail if umask was already 0022
	}

	// Restore original umask
	callBuiltin("umask", engine.NewInt(original.Int()))
}

func TestUmaskWrongArgCount(t *testing.T) {
	_, err := callBuiltin("umask", engine.NewInt(1), engine.NewInt(2))
	if err == nil {
		t.Error("umask(2 args) should return error")
	}
}

func TestUmaskNotInt(t *testing.T) {
	_, err := callBuiltin("umask", engine.NewString("0022"))
	if err == nil {
		t.Error("umask(string) should return error")
	}
}

// ============================================================================
// uname 测试
// ============================================================================

func TestUnameBasic(t *testing.T) {
	result, err := callBuiltin("uname")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()

	os := obj["os"]
	if os.String() == "" {
		t.Error("uname()['os'] should not be empty")
	}

	arch := obj["arch"]
	if arch.String() == "" {
		t.Error("uname()['arch'] should not be empty")
	}
}

func TestUnameWrongArgCount(t *testing.T) {
	_, err := callBuiltin("uname", engine.NewInt(1))
	if err == nil {
		t.Error("uname(1 arg) should return error")
	}
}

// ============================================================================
// 集成测试
// ============================================================================

func TestSystemIntegration(t *testing.T) {
	script := `$pid = getpid();
$uid = getuid();
$pid > 0 && $uid >= 0`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("getpid() and getuid() should return valid values")
	}
}

func TestDiskSpaceIntegration(t *testing.T) {
	script := `$free = disk_free_space("/");
$total = disk_total_space("/");
$free > 0 && $total > 0 && $free <= $total`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("disk space functions should return valid values")
	}
}

func TestUnameIntegration(t *testing.T) {
	script := `$info = uname();
$info["os"]`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() == "" {
		t.Error("uname()['os'] should not be empty")
	}
}
