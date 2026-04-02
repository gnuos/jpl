package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestGzencode(t *testing.T) {
	result, err := builtinGzencode(nil, []engine.Value{
		engine.NewString("hello world"),
	})
	if err != nil {
		t.Fatalf("gzencode error: %v", err)
	}
	if result.Type() != engine.TypeString {
		t.Fatalf("expected string, got %s", result.Type())
	}
	if result.String() == "hello world" {
		t.Fatal("gzencode should produce compressed output")
	}
	t.Log("gzencode: test passed")
}

func TestGzdecode(t *testing.T) {
	compressed, _ := builtinGzencode(nil, []engine.Value{
		engine.NewString("hello world"),
	})

	result, err := builtinGzdecode(nil, []engine.Value{compressed})
	if err != nil {
		t.Fatalf("gzdecode error: %v", err)
	}
	if result.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %s", result.String())
	}
	t.Log("gzdecode: test passed")
}

func TestGzfile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt.gz")

	_, err := builtinWriteGzfile(nil, []engine.Value{
		engine.NewString(testFile),
		engine.NewString("line1\nline2\nline3"),
	})
	if err != nil {
		t.Fatalf("writegzfile error: %v", err)
	}

	result, err := builtinGzfile(nil, []engine.Value{
		engine.NewString(testFile),
	})
	if err != nil {
		t.Fatalf("gzfile error: %v", err)
	}
	if result.Type() != engine.TypeArray {
		t.Fatalf("expected array, got %s", result.Type())
	}
	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(arr))
	}
	t.Log("gzfile: test passed")
}

func TestZlibEncode(t *testing.T) {
	result, err := builtinZlibEncode(nil, []engine.Value{
		engine.NewString("hello world"),
	})
	if err != nil {
		t.Fatalf("zlib_encode error: %v", err)
	}
	if result.Type() != engine.TypeString {
		t.Fatalf("expected string, got %s", result.Type())
	}
	if result.String() == "hello world" {
		t.Fatal("zlib_encode should produce compressed output")
	}
	t.Log("zlib_encode: test passed")
}

func TestZlibDecode(t *testing.T) {
	compressed, _ := builtinZlibEncode(nil, []engine.Value{
		engine.NewString("hello world"),
	})

	result, err := builtinZlibDecode(nil, []engine.Value{compressed})
	if err != nil {
		t.Fatalf("zlib_decode error: %v", err)
	}
	if result.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %s", result.String())
	}
	t.Log("zlib_decode: test passed")
}

func TestDeflate(t *testing.T) {
	result, err := builtinDeflate(nil, []engine.Value{
		engine.NewString("hello world"),
	})
	if err != nil {
		t.Fatalf("deflate error: %v", err)
	}
	if result.Type() != engine.TypeString {
		t.Fatalf("expected string, got %s", result.Type())
	}
	t.Log("deflate: test passed")
}

func TestInflate(t *testing.T) {
	compressed, _ := builtinDeflate(nil, []engine.Value{
		engine.NewString("hello world"),
	})

	result, err := builtinInflate(nil, []engine.Value{compressed})
	if err != nil {
		t.Fatalf("inflate error: %v", err)
	}
	if result.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %s", result.String())
	}
	t.Log("inflate: test passed")
}

func TestZipCreateAndOpen(t *testing.T) {
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")

	entries := []engine.Value{
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("file1.txt"),
			"content": engine.NewString("content1"),
		}),
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("file2.txt"),
			"content": engine.NewString("content2"),
		}),
	}

	_, err := builtinZipCreate(nil, []engine.Value{
		engine.NewString(zipFile),
		engine.NewArray(entries),
	})
	if err != nil {
		t.Fatalf("zip_create error: %v", err)
	}

	zipHandle, err := builtinZipOpen(nil, []engine.Value{
		engine.NewString(zipFile),
	})
	if err != nil || zipHandle.IsNull() {
		t.Fatalf("zip_open error: %v", err)
	}

	entry, err := builtinZipRead(nil, []engine.Value{zipHandle})
	if err != nil || entry.IsNull() {
		t.Fatalf("zip_read error: %v", err)
	}

	name, err := builtinZipEntryName(nil, []engine.Value{entry})
	if err != nil {
		t.Fatalf("zip_entry_name error: %v", err)
	}
	if name.String() != "file1.txt" {
		t.Fatalf("expected 'file1.txt', got %s", name.String())
	}

	builtinZipClose(nil, []engine.Value{zipHandle})
	t.Log("zip: create/open/read test passed")
}

func TestTarCreateAndOpen(t *testing.T) {
	tmpDir := t.TempDir()
	tarFile := filepath.Join(tmpDir, "test.tar")

	entries := []engine.Value{
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("file1.txt"),
			"content": engine.NewString("content1"),
		}),
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("file2.txt"),
			"content": engine.NewString("content2"),
		}),
	}

	_, err := builtinTarCreate(nil, []engine.Value{
		engine.NewString(tarFile),
		engine.NewArray(entries),
	})
	if err != nil {
		t.Fatalf("tar_create error: %v", err)
	}

	tarHandle, err := builtinTarOpen(nil, []engine.Value{
		engine.NewString(tarFile),
	})
	if err != nil || tarHandle.IsNull() {
		t.Fatalf("tar_open error: %v", err)
	}

	entry, err := builtinTarRead(nil, []engine.Value{tarHandle})
	if err != nil || entry.IsNull() {
		t.Fatalf("tar_read error: %v", err)
	}

	name, err := builtinTarEntryName(nil, []engine.Value{entry})
	if err != nil {
		t.Fatalf("tar_entry_name error: %v", err)
	}
	if name.String() != "file1.txt" {
		t.Fatalf("expected 'file1.txt', got %s", name.String())
	}

	builtinTarClose(nil, []engine.Value{tarHandle})
	t.Log("tar: create/open/read test passed")
}

func TestZipEntryRead(t *testing.T) {
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")

	entries := []engine.Value{
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("test.txt"),
			"content": engine.NewString("hello world"),
		}),
	}

	builtinZipCreate(nil, []engine.Value{
		engine.NewString(zipFile),
		engine.NewArray(entries),
	})

	zipHandle, _ := builtinZipOpen(nil, []engine.Value{
		engine.NewString(zipFile),
	})
	entry, _ := builtinZipRead(nil, []engine.Value{zipHandle})

	content, err := builtinZipEntryRead(nil, []engine.Value{entry, engine.NewInt(100)})
	if err != nil {
		t.Fatalf("zip_entry_read error: %v", err)
	}
	if content.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %s", content.String())
	}

	builtinZipEntryClose(nil, []engine.Value{entry})
	builtinZipClose(nil, []engine.Value{zipHandle})
	t.Log("zip_entry_read: test passed")
}

func TestTarEntryRead(t *testing.T) {
	tmpDir := t.TempDir()
	tarFile := filepath.Join(tmpDir, "test.tar")

	entries := []engine.Value{
		engine.NewObject(map[string]engine.Value{
			"name":    engine.NewString("test.txt"),
			"content": engine.NewString("hello world"),
		}),
	}

	builtinTarCreate(nil, []engine.Value{
		engine.NewString(tarFile),
		engine.NewArray(entries),
	})

	tarHandle, _ := builtinTarOpen(nil, []engine.Value{
		engine.NewString(tarFile),
	})
	_, _ = builtinTarRead(nil, []engine.Value{tarHandle})

	content, err := builtinTarEntryRead(nil, []engine.Value{tarHandle, engine.NewInt(100)})
	if err != nil {
		t.Fatalf("tar_entry_read error: %v", err)
	}
	if content.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %s", content.String())
	}

	builtinTarClose(nil, []engine.Value{tarHandle})
	t.Log("tar_entry_read: test passed")
}

func init() {
	tmpDir := os.TempDir()
	_ = os.MkdirAll(filepath.Join(tmpDir, "jpl_test"), 0755)
}
