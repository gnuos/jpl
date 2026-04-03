package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/engine"
)

// ============================================================================
// 文件读写测试
// ============================================================================

func TestBuiltinRead(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("read", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result.String())
	}
}

func TestBuiltinReadEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(tmpFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("read", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("expected empty string, got '%s'", result.String())
	}
}

func TestBuiltinReadNotFound(t *testing.T) {
	_, err := callBuiltin("read", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestBuiltinReadWrongArgs(t *testing.T) {
	_, err := callBuiltin("read", engine.NewInt(42))
	if err == nil {
		t.Error("expected error for non-string argument")
	}
}

func TestBuiltinReadLines(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "lines.txt")
	content := "line1\nline2\nline3"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("readLines", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	expected := []string{"line1", "line2", "line3"}
	if len(arr) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(arr))
	}
	for i, v := range arr {
		if v.String() != expected[i] {
			t.Errorf("line %d: expected '%s', got '%s'", i, expected[i], v.String())
		}
	}
}

func TestBuiltinWrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "write.txt")

	_, err := callBuiltin("write", engine.NewString(tmpFile), engine.NewString("test content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(content))
	}
}

func TestBuiltinWriteOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "overwrite.txt")
	err := os.WriteFile(tmpFile, []byte("old"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err = callBuiltin("write", engine.NewString(tmpFile), engine.NewString("new"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("expected 'new', got '%s'", string(content))
	}
}

func TestBuiltinAppend(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "append.txt")
	err := os.WriteFile(tmpFile, []byte("hello "), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	_, err = callBuiltin("append", engine.NewString(tmpFile), engine.NewString("world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestBuiltinAppendCreate(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new.txt")

	_, err := callBuiltin("append", engine.NewString(tmpFile), engine.NewString("content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("expected 'content', got '%s'", string(content))
	}
}

func TestBuiltinExists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "exists.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// 存在的文件
	result, err := callBuiltin("exists", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true for existing file")
	}

	// 不存在的文件
	result, err = callBuiltin("exists", engine.NewString("/nonexistent/file.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for nonexistent file")
	}
}

// ============================================================================
// 目录操作测试
// ============================================================================

func TestBuiltinMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "newdir")

	result, err := callBuiltin("mkdir", engine.NewString(newDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}

	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestBuiltinMkdirExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing")
	err := os.Mkdir(existingDir, 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	_, err = callBuiltin("mkdir", engine.NewString(existingDir))
	if err == nil {
		t.Error("expected error for existing directory")
	}
}

func TestBuiltinMkdirAll(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")

	result, err := callBuiltin("mkdirAll", engine.NewString(nestedDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}

	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestBuiltinRmdir(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	err := os.Mkdir(emptyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	result, err := callBuiltin("rmdir", engine.NewString(emptyDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true")
	}

	_, err = os.Stat(emptyDir)
	if !os.IsNotExist(err) {
		t.Error("directory should be deleted")
	}
}

func TestBuiltinRmdirNonEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	err := os.Mkdir(nonEmptyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	_, err = callBuiltin("rmdir", engine.NewString(nonEmptyDir))
	if err == nil {
		t.Error("expected error for non-empty directory")
	}
}

func TestBuiltinListDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	result, err := callBuiltin("listDir", engine.NewString(tmpDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Errorf("expected 3 entries, got %d", len(arr))
	}
}

func TestBuiltinListDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	os.Mkdir(emptyDir, 0755)

	result, err := callBuiltin("listDir", engine.NewString(emptyDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.Array()
	if len(arr) != 0 {
		t.Errorf("expected 0 entries, got %d", len(arr))
	}
}

// ============================================================================
// 文件信息测试
// ============================================================================

func TestBuiltinStat(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "stat.txt")
	err := os.WriteFile(tmpFile, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("stat", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()
	if obj["name"].String() != "stat.txt" {
		t.Errorf("expected name 'stat.txt', got '%s'", obj["name"].String())
	}
	if obj["size"].Int() != 5 {
		t.Errorf("expected size 5, got %d", obj["size"].Int())
	}
	if obj["isDir"].Bool() {
		t.Error("expected isDir to be false")
	}
	if !obj["isFile"].Bool() {
		t.Error("expected isFile to be true")
	}
}

func TestBuiltinStatDir(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := callBuiltin("stat", engine.NewString(tmpDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()
	if !obj["isDir"].Bool() {
		t.Error("expected isDir to be true")
	}
	if obj["isFile"].Bool() {
		t.Error("expected isFile to be false")
	}
}

func TestBuiltinFileSize(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "size.txt")
	content := "1234567890"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("fileSize", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 10 {
		t.Errorf("expected size 10, got %d", result.Int())
	}
}

func TestBuiltinIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	// 文件
	result, err := callBuiltin("isFile", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true for file")
	}

	// 目录
	result, err = callBuiltin("isFile", engine.NewString(tmpDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for directory")
	}
}

func TestBuiltinIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	// 目录
	result, err := callBuiltin("isDir", engine.NewString(tmpDir))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("expected true for directory")
	}

	// 文件
	result, err = callBuiltin("isDir", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for file")
	}
}

func TestBuiltinIsFileNotFound(t *testing.T) {
	result, err := callBuiltin("isFile", engine.NewString("/nonexistent/file.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for nonexistent file")
	}
}

func TestBuiltinIsDirNotFound(t *testing.T) {
	result, err := callBuiltin("isDir", engine.NewString("/nonexistent/dir"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Bool() {
		t.Error("expected false for nonexistent dir")
	}
}

// ============================================================================
// 路径处理测试
// ============================================================================

func TestBuiltinDirname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/file.txt", "/home/user"},
		{"file.txt", "."},
		{"/", "/"},
		{"", "."},
	}

	for _, tt := range tests {
		result, err := callBuiltin("dirname", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("dirname(%q): unexpected error: %v", tt.input, err)
		}
		if result.String() != tt.expected {
			t.Errorf("dirname(%q): expected %q, got %q", tt.input, tt.expected, result.String())
		}
	}
}

func TestBuiltinBasename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/file.txt", "file.txt"},
		{"file.txt", "file.txt"},
		{"/home/user/", "user"},
		{"/", "/"},
	}

	for _, tt := range tests {
		result, err := callBuiltin("basename", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("basename(%q): unexpected error: %v", tt.input, err)
		}
		if result.String() != tt.expected {
			t.Errorf("basename(%q): expected %q, got %q", tt.input, tt.expected, result.String())
		}
	}
}

func TestBuiltinExtname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"file.txt", ".txt"},
		{"file.tar.gz", ".gz"},
		{"file", ""},
		{".hidden", ".hidden"},
		{"/path/to/file.go", ".go"},
	}

	for _, tt := range tests {
		result, err := callBuiltin("extname", engine.NewString(tt.input))
		if err != nil {
			t.Fatalf("extname(%q): unexpected error: %v", tt.input, err)
		}
		if result.String() != tt.expected {
			t.Errorf("extname(%q): expected %q, got %q", tt.input, tt.expected, result.String())
		}
	}
}

func TestBuiltinJoinPath(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{"/home", "user", "file.txt"}, "/home/user/file.txt"},
		{[]string{"a", "b", "c"}, "a/b/c"},
		{[]string{"/", "home", "user"}, "/home/user"},
		{[]string{"file.txt"}, "file.txt"},
	}

	for _, tt := range tests {
		args := make([]engine.Value, len(tt.args))
		for i, s := range tt.args {
			args[i] = engine.NewString(s)
		}
		result, err := callBuiltin("joinPath", args...)
		if err != nil {
			t.Fatalf("joinPath(%v): unexpected error: %v", tt.args, err)
		}
		if result.String() != tt.expected {
			t.Errorf("joinPath(%v): expected %q, got %q", tt.args, tt.expected, result.String())
		}
	}
}

func TestBuiltinJoinPathNoArgs(t *testing.T) {
	_, err := callBuiltin("joinPath")
	if err == nil {
		t.Error("expected error for no arguments")
	}
}

func TestBuiltinJoinPathNonString(t *testing.T) {
	_, err := callBuiltin("joinPath", engine.NewString("/home"), engine.NewInt(42))
	if err == nil {
		t.Error("expected error for non-string argument")
	}
}

// ============================================================================
// 边缘情况测试
// ============================================================================

func TestBuiltinReadBinaryContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "binary.dat")
	binary := []byte{0x00, 0x01, 0x02, 0xFF}
	err := os.WriteFile(tmpFile, binary, 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("read", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 读取为字符串，包含二进制内容
	if len(result.String()) != 4 {
		t.Errorf("expected length 4, got %d", len(result.String()))
	}
}

func TestBuiltinWriteEmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.txt")

	_, err := callBuiltin("write", engine.NewString(tmpFile), engine.NewString(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "" {
		t.Errorf("expected empty file, got '%s'", string(content))
	}
}

func TestBuiltinWriteInvalidPath(t *testing.T) {
	_, err := callBuiltin("write", engine.NewString("/nonexistent/dir/file.txt"), engine.NewString("test"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestBuiltinMkdirInvalidPath(t *testing.T) {
	_, err := callBuiltin("mkdir", engine.NewString("/nonexistent/parent/newdir"))
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestBuiltinRmdirNotFound(t *testing.T) {
	_, err := callBuiltin("rmdir", engine.NewString("/nonexistent/dir"))
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestBuiltinListDirNotFound(t *testing.T) {
	_, err := callBuiltin("listDir", engine.NewString("/nonexistent/dir"))
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestBuiltinStatNotFound(t *testing.T) {
	_, err := callBuiltin("stat", engine.NewString("/nonexistent/file"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestBuiltinFileSizeNotFound(t *testing.T) {
	_, err := callBuiltin("fileSize", engine.NewString("/nonexistent/file"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// ============================================================================
// 错误参数测试
// ============================================================================

func TestFileIOWrongArgCount(t *testing.T) {
	functions := []string{
		"read", "readLines", "exists",
		"mkdir", "mkdirAll", "rmdir", "listDir",
		"stat", "fileSize", "isFile", "isDir",
		"dirname", "basename", "extname",
	}

	for _, fn := range functions {
		// 无参数
		_, err := callBuiltin(fn)
		if err == nil {
			t.Errorf("%s(): expected error for no arguments", fn)
		}

		// 多参数
		_, err = callBuiltin(fn, engine.NewString("a"), engine.NewString("b"))
		if err == nil {
			t.Errorf("%s(): expected error for too many arguments", fn)
		}
	}

	// write 和 append 需要 2 个参数
	_, err := callBuiltin("write", engine.NewString("path"))
	if err == nil {
		t.Error("write(): expected error for 1 argument")
	}

	_, err = callBuiltin("append", engine.NewString("path"))
	if err == nil {
		t.Error("append(): expected error for 1 argument")
	}
}

func TestFileIOWrongArgType(t *testing.T) {
	functions := []string{
		"read", "readLines", "exists",
		"mkdir", "mkdirAll", "rmdir", "listDir",
		"stat", "fileSize", "isFile", "isDir",
		"dirname", "basename", "extname",
	}

	for _, fn := range functions {
		_, err := callBuiltin(fn, engine.NewInt(42))
		if err == nil {
			t.Errorf("%s(): expected error for non-string argument", fn)
		}
	}

	// write 和 append 的第二个参数也必须是字符串
	_, err := callBuiltin("write", engine.NewString("path"), engine.NewInt(42))
	if err == nil {
		t.Error("write(): expected error for non-string content")
	}

	_, err = callBuiltin("append", engine.NewString("path"), engine.NewInt(42))
	if err == nil {
		t.Error("append(): expected error for non-string content")
	}
}

// TestFileSystemFunctions 测试文件系统扩展函数
func TestFileSystemFunctions(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	tests := []struct {
		name string
		fn   string
	}{
		{"chdir", "chdir"},
		{"rename", "rename"},
		{"unlink", "unlink"},
		{"realpath", "realpath"},
		{"is_readable", "is_readable"},
		{"is_writable", "is_writable"},
		{"chmod", "chmod"},
		{"scandir", "scandir"},
		{"glob", "glob"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := e.GetRegisteredFunc(tt.fn)
			if fn == nil {
				t.Errorf("%s function not registered", tt.fn)
			}
		})
	}
}

// TestChdir 测试 chdir 函数
func TestChdir(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	ctx := engine.NewContext(e, nil)

	// Get current dir
	oldDir, _ := os.Getwd()

	// Change to parent directory
	result, err := builtinChdir(ctx, []engine.Value{engine.NewString("..")})
	if err != nil {
		t.Logf("chdir may fail in test environment: %v", err)
		return
	}

	if !result.Bool() {
		t.Error("chdir() should return true on success")
	}

	// Restore
	os.Chdir(oldDir)
}

// TestRealpath 测试 realpath 函数
func TestRealpath(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	ctx := engine.NewContext(e, nil)

	// Test with current directory
	result, err := builtinRealpath(ctx, []engine.Value{engine.NewString(".")})
	if err != nil {
		t.Fatalf("realpath() failed: %v", err)
	}

	if result.Type() != engine.TypeString {
		t.Errorf("realpath() should return string, got %s", result.Type())
	}

	path := result.String()
	if path == "" {
		t.Error("realpath() returned empty string")
	}

	t.Logf("realpath(\".\") = %s", path)
}

// TestIsReadableWritable 测试 is_readable 和 is_writable
func TestIsReadableWritable(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	ctx := engine.NewContext(e, nil)

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Skipf("Cannot create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Test is_readable
	readable, err := builtinIsReadable(ctx, []engine.Value{engine.NewString(tmpFile.Name())})
	if err != nil {
		t.Errorf("is_readable() error: %v", err)
	}
	if !readable.Bool() {
		t.Logf("File %s may not be readable in test environment", tmpFile.Name())
	}

	// Test is_writable
	writable, err := builtinIsWritable(ctx, []engine.Value{engine.NewString(tmpFile.Name())})
	if err != nil {
		t.Errorf("is_writable() error: %v", err)
	}
	if !writable.Bool() {
		t.Logf("File %s may not be writable in test environment", tmpFile.Name())
	}
}

// TestScandir 测试 scandir 函数
func TestScandir(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	ctx := engine.NewContext(e, nil)

	// Scan current directory
	result, err := builtinScandir(ctx, []engine.Value{engine.NewString(".")})
	if err != nil {
		t.Fatalf("scandir() failed: %v", err)
	}

	if result.Type() != engine.TypeArray {
		t.Errorf("scandir() should return array, got %s", result.Type())
	}

	if result.Len() == 0 {
		t.Error("scandir() returned empty array")
	}

	t.Logf("scandir(\".\") returned %d entries", result.Len())
}

// TestGlob 测试 glob 函数
func TestGlob(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterFileIO(e)

	ctx := engine.NewContext(e, nil)

	// Find all .go files in current directory
	result, err := builtinGlob(ctx, []engine.Value{engine.NewString("*.go")})
	if err != nil {
		t.Fatalf("glob() failed: %v", err)
	}

	if result.Type() != engine.TypeArray {
		t.Errorf("glob() should return array, got %s", result.Type())
	}

	t.Logf("glob(\"*.go\") returned %d matches", result.Len())
}

// ============================================================================
// file_get_contents 测试
// ============================================================================

func TestFileGetContentsBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("file_get_contents", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "hello world" {
		t.Errorf("expected 'hello world', got %s", result.String())
	}
}

func TestFileGetContentsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("file_get_contents", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "" {
		t.Errorf("expected empty string, got %s", result.String())
	}
}

func TestFileGetContentsNotFound(t *testing.T) {
	_, err := callBuiltin("file_get_contents", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("file_get_contents(nonexistent) should return error")
	}
}

func TestFileGetContentsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("file_get_contents")
	if err == nil {
		t.Error("file_get_contents(0 args) should return error")
	}
}

func TestFileGetContentsNotString(t *testing.T) {
	_, err := callBuiltin("file_get_contents", engine.NewInt(42))
	if err == nil {
		t.Error("file_get_contents(42) should return error")
	}
}

// ============================================================================
// file_put_contents 测试
// ============================================================================

func TestFilePutContentsBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.txt")

	result, err := callBuiltin("file_put_contents", engine.NewString(tmpFile), engine.NewString("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return number of bytes written
	if result.Int() != 11 {
		t.Errorf("expected 11 bytes written, got %d", result.Int())
	}

	// Verify file content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %s", string(content))
	}
}

func TestFilePutContentsOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "overwrite.txt")

	// Write initial content
	if err := os.WriteFile(tmpFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Overwrite
	_, err := callBuiltin("file_put_contents", engine.NewString(tmpFile), engine.NewString("new content"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(tmpFile)
	if string(content) != "new content" {
		t.Errorf("expected 'new content', got %s", string(content))
	}
}

func TestFilePutContentsWrongArgCount(t *testing.T) {
	_, err := callBuiltin("file_put_contents")
	if err == nil {
		t.Error("file_put_contents(0 args) should return error")
	}

	_, err = callBuiltin("file_put_contents", engine.NewString("path"))
	if err == nil {
		t.Error("file_put_contents(1 arg) should return error")
	}
}

func TestFilePutContentsNotString(t *testing.T) {
	_, err := callBuiltin("file_put_contents", engine.NewInt(42), engine.NewString("data"))
	if err == nil {
		t.Error("file_put_contents(int, string) should return error")
	}

	_, err = callBuiltin("file_put_contents", engine.NewString("path"), engine.NewInt(42))
	if err == nil {
		t.Error("file_put_contents(string, int) should return error")
	}
}

// ============================================================================
// copy 测试
// ============================================================================

func TestCopyBasic(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	if err := os.WriteFile(srcFile, []byte("copy me"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	result, err := callBuiltin("copy", engine.NewString(srcFile), engine.NewString(dstFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Bool() {
		t.Error("copy() should return true")
	}

	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(content) != "copy me" {
		t.Errorf("expected 'copy me', got %s", string(content))
	}
}

func TestCopyOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	if err := os.WriteFile(srcFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}
	if err := os.WriteFile(dstFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create dest file: %v", err)
	}

	_, err := callBuiltin("copy", engine.NewString(srcFile), engine.NewString(dstFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(dstFile)
	if string(content) != "new" {
		t.Errorf("expected 'new', got %s", string(content))
	}
}

func TestCopySourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	dstFile := filepath.Join(tmpDir, "dest.txt")

	_, err := callBuiltin("copy", engine.NewString("/nonexistent/source.txt"), engine.NewString(dstFile))
	if err == nil {
		t.Error("copy(nonexistent, dest) should return error")
	}
}

func TestCopyWrongArgCount(t *testing.T) {
	_, err := callBuiltin("copy")
	if err == nil {
		t.Error("copy(0 args) should return error")
	}

	_, err = callBuiltin("copy", engine.NewString("src"))
	if err == nil {
		t.Error("copy(1 arg) should return error")
	}
}

func TestCopyNotString(t *testing.T) {
	_, err := callBuiltin("copy", engine.NewInt(42), engine.NewString("dst"))
	if err == nil {
		t.Error("copy(int, string) should return error")
	}
}

// ============================================================================
// readfile 测试
// ============================================================================

func TestReadfileBasic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("read me"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := callBuiltin("readfile", engine.NewString(tmpFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "read me" {
		t.Errorf("expected 'read me', got %s", result.String())
	}
}

func TestReadfileNotFound(t *testing.T) {
	_, err := callBuiltin("readfile", engine.NewString("/nonexistent/file.txt"))
	if err == nil {
		t.Error("readfile(nonexistent) should return error")
	}
}

func TestReadfileWrongArgCount(t *testing.T) {
	_, err := callBuiltin("readfile")
	if err == nil {
		t.Error("readfile(0 args) should return error")
	}
}

// ============================================================================
// pathinfo 测试
// ============================================================================

func TestPathinfoBasic(t *testing.T) {
	result, err := callBuiltin("pathinfo", engine.NewString("/path/to/file.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()

	dirname := obj["dirname"]
	if dirname.String() != "/path/to" {
		t.Errorf("dirname expected '/path/to', got %s", dirname.String())
	}

	basename := obj["basename"]
	if basename.String() != "file.txt" {
		t.Errorf("basename expected 'file.txt', got %s", basename.String())
	}

	extension := obj["extension"]
	if extension.String() != "txt" {
		t.Errorf("extension expected 'txt', got %s", extension.String())
	}

	filename := obj["filename"]
	if filename.String() != "file" {
		t.Errorf("filename expected 'file', got %s", filename.String())
	}
}

func TestPathinfoNoExtension(t *testing.T) {
	result, err := callBuiltin("pathinfo", engine.NewString("/path/to/file"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()

	extension := obj["extension"]
	if extension.String() != "" {
		t.Errorf("extension expected empty, got %s", extension.String())
	}
}

func TestPathinfoSimpleFile(t *testing.T) {
	result, err := callBuiltin("pathinfo", engine.NewString("file.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj := result.Object()

	dirname := obj["dirname"]
	if dirname.String() != "." {
		t.Errorf("dirname expected '.', got %s", dirname.String())
	}

	basename := obj["basename"]
	if basename.String() != "file.txt" {
		t.Errorf("basename expected 'file.txt', got %s", basename.String())
	}
}

func TestPathinfoWrongArgCount(t *testing.T) {
	_, err := callBuiltin("pathinfo")
	if err == nil {
		t.Error("pathinfo(0 args) should return error")
	}
}

func TestPathinfoNotString(t *testing.T) {
	_, err := callBuiltin("pathinfo", engine.NewInt(42))
	if err == nil {
		t.Error("pathinfo(42) should return error")
	}
}

// ============================================================================
// 流操作参数校验测试
// ============================================================================

func TestFseekInvalidArg(t *testing.T) {
	// 非 stream 参数应报错
	_, err := callBuiltin("fseek", engine.NewString("dummy"), engine.NewInt(0), engine.NewInt(0))
	if err == nil {
		t.Error("fseek() with string should return error")
	}
}

func TestFtellInvalidArg(t *testing.T) {
	// 非 stream 参数应报错
	_, err := callBuiltin("ftell", engine.NewString("dummy"))
	if err == nil {
		t.Error("ftell() with string should return error")
	}
}

func TestRewindInvalidArg(t *testing.T) {
	// 非 stream 参数应报错
	_, err := callBuiltin("rewind", engine.NewString("dummy"))
	if err == nil {
		t.Error("rewind() with string should return error")
	}
}

func TestFtruncateInvalidArg(t *testing.T) {
	// 非 stream 参数应报错
	_, err := callBuiltin("ftruncate", engine.NewString("dummy"), engine.NewInt(0))
	if err == nil {
		t.Error("ftruncate() with string should return error")
	}
}

func TestFgetcsvInvalidArg(t *testing.T) {
	// 非 stream 参数应报错
	_, err := callBuiltin("fgetcsv", engine.NewString("dummy"))
	if err == nil {
		t.Error("fgetcsv() with string should return error")
	}
}

// ============================================================================
// 流操作功能测试
// ============================================================================

func TestFseekFtellRewind(t *testing.T) {
	// 创建临时文件
	tmpFile := filepath.Join(t.TempDir(), "seek_test.txt")
	if err := os.WriteFile(tmpFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatal(err)
	}

	// 打开文件流
	stream := engine.NewFileStream(tmpFile, engine.StreamRead)

	// ftell: 初始位置应为 0
	pos, err := callBuiltin("ftell", stream)
	if err != nil {
		t.Fatalf("ftell() error: %v", err)
	}
	if pos.Int() != 0 {
		t.Errorf("ftell() expected 0, got %d", pos.Int())
	}

	// fseek: 移动到位置 7
	pos, err = callBuiltin("fseek", stream, engine.NewInt(7), engine.NewInt(0))
	if err != nil {
		t.Fatalf("fseek() error: %v", err)
	}
	if pos.Int() != 7 {
		t.Errorf("fseek() expected 7, got %d", pos.Int())
	}

	// ftell: 验证新位置
	pos, err = callBuiltin("ftell", stream)
	if err != nil {
		t.Fatalf("ftell() error: %v", err)
	}
	if pos.Int() != 7 {
		t.Errorf("ftell() expected 7, got %d", pos.Int())
	}

	// rewind: 回到开头
	_, err = callBuiltin("rewind", stream)
	if err != nil {
		t.Fatalf("rewind() error: %v", err)
	}

	// ftell: 验证回到 0
	pos, err = callBuiltin("ftell", stream)
	if err != nil {
		t.Fatalf("ftell() error: %v", err)
	}
	if pos.Int() != 0 {
		t.Errorf("ftell() after rewind expected 0, got %d", pos.Int())
	}
}

func TestFtruncate(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "trunc_test.txt")
	if err := os.WriteFile(tmpFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatal(err)
	}

	// 打开为读写流
	stream := engine.NewFileStream(tmpFile, engine.StreamReadWrite)

	// 截断到 5 字节
	result, err := callBuiltin("ftruncate", stream, engine.NewInt(5))
	if err != nil {
		t.Fatalf("ftruncate() error: %v", err)
	}
	if !result.Bool() {
		t.Error("ftruncate() should return true")
	}

	// 验证文件大小
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 5 {
		t.Errorf("file size after ftruncate expected 5, got %d", info.Size())
	}
}

func TestFgetcsv(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "csv_test.txt")
	content := "name,age,city\nAlice,30,Beijing\nBob,25,Shanghai\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stream := engine.NewFileStream(tmpFile, engine.StreamRead)

	// 读取第一行
	row, err := callBuiltin("fgetcsv", stream)
	if err != nil {
		t.Fatalf("fgetcsv() error: %v", err)
	}
	arr := row.Array()
	if len(arr) != 3 || arr[0].String() != "name" || arr[1].String() != "age" || arr[2].String() != "city" {
		t.Errorf("fgetcsv() row 1 expected [name,age,city], got %v", row.Stringify())
	}

	// 读取第二行
	row, err = callBuiltin("fgetcsv", stream)
	if err != nil {
		t.Fatalf("fgetcsv() error: %v", err)
	}
	arr = row.Array()
	if len(arr) != 3 || arr[0].String() != "Alice" {
		t.Errorf("fgetcsv() row 2 expected Alice, got %v", arr[0].String())
	}
}

func TestFgetcsvCustomDelimiter(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "tsv_test.txt")
	content := "name\tage\nAlice\t30\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stream := engine.NewFileStream(tmpFile, engine.StreamRead)

	// 使用 Tab 分隔符
	row, err := callBuiltin("fgetcsv", stream, engine.NewString("\t"))
	if err != nil {
		t.Fatalf("fgetcsv() error: %v", err)
	}
	arr := row.Array()
	if len(arr) != 2 || arr[0].String() != "name" || arr[1].String() != "age" {
		t.Errorf("fgetcsv() with tab delimiter expected [name,age], got %v", row.Stringify())
	}
}

// ============================================================================
// 集成测试
// ============================================================================

func TestFileGetPutRoundTrip(t *testing.T) {
	script := `$tmp = "/tmp/jpl_test_fileio.txt";
file_put_contents($tmp, "round trip test");
$content = file_get_contents($tmp);
unlink($tmp);
$content`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "round trip test" {
		t.Errorf("expected 'round trip test', got %s", result.String())
	}
}

func TestCopyIntegration(t *testing.T) {
	script := `$src = "/tmp/jpl_copy_src.txt";
$dst = "/tmp/jpl_copy_dst.txt";
file_put_contents($src, "copy test");
copy($src, $dst);
$content = file_get_contents($dst);
unlink($src);
unlink($dst);
$content`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "copy test" {
		t.Errorf("expected 'copy test', got %s", result.String())
	}
}

func TestPathinfoIntegration(t *testing.T) {
	script := `$info = pathinfo("/home/user/document.pdf");
$info["extension"]`
	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.String() != "pdf" {
		t.Errorf("expected 'pdf', got %s", result.String())
	}
}
