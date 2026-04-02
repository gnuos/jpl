package stdlib

import (
	"os"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// MD5 测试
// ============================================================================

func TestMd5Basic(t *testing.T) {
	result, err := callBuiltin("md5", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "5d41402abc4b2a76b9719d911017c592"
	if result.String() != expected {
		t.Errorf("md5('hello') expected %s, got %s", expected, result.String())
	}
}

func TestMd5Empty(t *testing.T) {
	result, err := callBuiltin("md5", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if result.String() != expected {
		t.Errorf("md5('') expected %s, got %s", expected, result.String())
	}
}

func TestMd5WrongArgCount(t *testing.T) {
	_, err := callBuiltin("md5")
	if err == nil {
		t.Error("md5(0 args) should return error")
	}

	_, err = callBuiltin("md5", engine.NewString("a"), engine.NewString("b"))
	if err == nil {
		t.Error("md5(2 args) should return error")
	}
}

func TestMd5NotString(t *testing.T) {
	_, err := callBuiltin("md5", engine.NewInt(42))
	if err == nil {
		t.Error("md5(42) should return error")
	}
}

// ============================================================================
// SHA1 测试
// ============================================================================

func TestSha1Basic(t *testing.T) {
	result, err := callBuiltin("sha1", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if result.String() != expected {
		t.Errorf("sha1('hello') expected %s, got %s", expected, result.String())
	}
}

func TestSha1Empty(t *testing.T) {
	result, err := callBuiltin("sha1", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if result.String() != expected {
		t.Errorf("sha1('') expected %s, got %s", expected, result.String())
	}
}

func TestSha1WrongArgCount(t *testing.T) {
	_, err := callBuiltin("sha1")
	if err == nil {
		t.Error("sha1(0 args) should return error")
	}
}

func TestSha1NotString(t *testing.T) {
	_, err := callBuiltin("sha1", engine.NewInt(42))
	if err == nil {
		t.Error("sha1(42) should return error")
	}
}

// ============================================================================
// Base64 编码测试
// ============================================================================

func TestBase64EncodeBasic(t *testing.T) {
	result, err := callBuiltin("base64_encode", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "aGVsbG8="
	if result.String() != expected {
		t.Errorf("base64_encode('hello') expected %s, got %s", expected, result.String())
	}
}

func TestBase64EncodeEmpty(t *testing.T) {
	result, err := callBuiltin("base64_encode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("base64_encode('') expected empty, got %s", result.String())
	}
}

func TestBase64EncodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("base64_encode")
	if err == nil {
		t.Error("base64_encode(0 args) should return error")
	}
}

func TestBase64EncodeNotString(t *testing.T) {
	_, err := callBuiltin("base64_encode", engine.NewInt(42))
	if err == nil {
		t.Error("base64_encode(42) should return error")
	}
}

// ============================================================================
// Base64 解码测试
// ============================================================================

func TestBase64DecodeBasic(t *testing.T) {
	result, err := callBuiltin("base64_decode", engine.NewString("aGVsbG8="))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello" {
		t.Errorf("base64_decode('aGVsbG8=') expected 'hello', got %s", result.String())
	}
}

func TestBase64DecodeEmpty(t *testing.T) {
	result, err := callBuiltin("base64_decode", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("base64_decode('') expected empty, got %s", result.String())
	}
}

func TestBase64DecodeInvalid(t *testing.T) {
	_, err := callBuiltin("base64_decode", engine.NewString("not!!!valid"))
	if err == nil {
		t.Error("base64_decode(invalid) should return error")
	}
}

func TestBase64DecodeWrongArgCount(t *testing.T) {
	_, err := callBuiltin("base64_decode")
	if err == nil {
		t.Error("base64_decode(0 args) should return error")
	}
}

func TestBase64DecodeNotString(t *testing.T) {
	_, err := callBuiltin("base64_decode", engine.NewInt(42))
	if err == nil {
		t.Error("base64_decode(42) should return error")
	}
}

// ============================================================================
// CRC32 测试
// ============================================================================

func TestCrc32Basic(t *testing.T) {
	result, err := callBuiltin("crc32", engine.NewString("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// CRC32 of "hello" = 907060870
	if result.Int() != 907060870 {
		t.Errorf("crc32('hello') expected 907060870, got %d", result.Int())
	}
}

func TestCrc32Empty(t *testing.T) {
	result, err := callBuiltin("crc32", engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 0 {
		t.Errorf("crc32('') expected 0, got %d", result.Int())
	}
}

func TestCrc32WrongArgCount(t *testing.T) {
	_, err := callBuiltin("crc32")
	if err == nil {
		t.Error("crc32(0 args) should return error")
	}
}

func TestCrc32NotString(t *testing.T) {
	_, err := callBuiltin("crc32", engine.NewInt(42))
	if err == nil {
		t.Error("crc32(42) should return error")
	}
}

// ============================================================================
// 集成测试
// ============================================================================

func TestMd5Integration(t *testing.T) {
	script := `$hash = md5("hello");
$hash`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "5d41402abc4b2a76b9719d911017c592" {
		t.Errorf("expected md5 hash, got %s", result.String())
	}
}

func TestBase64RoundTrip(t *testing.T) {
	script := `$encoded = base64_encode("hello world");
$decoded = base64_decode($encoded);
$decoded`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("expected 'hello world', got %s", result.String())
	}
}

// ============================================================================
// md5_file 测试
// ============================================================================

func TestMd5FileBasic(t *testing.T) {
	// 创建临时文件
	tmpFile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("md5_file", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// md5("hello") = 5d41402abc4b2a76b9719d911017c592
	expected := "5d41402abc4b2a76b9719d911017c592"
	if result.String() != expected {
		t.Errorf("md5_file() expected %s, got %s", expected, result.String())
	}
}

func TestMd5FileEmpty(t *testing.T) {
	tmpFile := t.TempDir() + "/empty.txt"
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("md5_file", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if result.String() != expected {
		t.Errorf("md5_file(empty) expected %s, got %s", expected, result.String())
	}
}

func TestMd5FileNotFound(t *testing.T) {
	_, err := callBuiltin("md5_file", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("md5_file(nonexistent) should return error")
	}
}

func TestMd5FileWrongArgCount(t *testing.T) {
	_, err := callBuiltin("md5_file")
	if err == nil {
		t.Error("md5_file(0 args) should return error")
	}
}

func TestMd5FileNotString(t *testing.T) {
	_, err := callBuiltin("md5_file", engine.NewInt(42))
	if err == nil {
		t.Error("md5_file(42) should return error")
	}
}

// ============================================================================
// sha1_file 测试
// ============================================================================

func TestSha1FileBasic(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("sha1_file", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// sha1("hello") = aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
	if result.String() != expected {
		t.Errorf("sha1_file() expected %s, got %s", expected, result.String())
	}
}

func TestSha1FileEmpty(t *testing.T) {
	tmpFile := t.TempDir() + "/empty.txt"
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("sha1_file", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if result.String() != expected {
		t.Errorf("sha1_file(empty) expected %s, got %s", expected, result.String())
	}
}

func TestSha1FileNotFound(t *testing.T) {
	_, err := callBuiltin("sha1_file", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("sha1_file(nonexistent) should return error")
	}
}

func TestSha1FileWrongArgCount(t *testing.T) {
	_, err := callBuiltin("sha1_file")
	if err == nil {
		t.Error("sha1_file(0 args) should return error")
	}
}

func TestSha1FileNotString(t *testing.T) {
	_, err := callBuiltin("sha1_file", engine.NewInt(42))
	if err == nil {
		t.Error("sha1_file(42) should return error")
	}
}
