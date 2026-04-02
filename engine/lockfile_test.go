package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// 哈希计算测试
// ============================================================================

func TestHashContent(t *testing.T) {
	data := []byte("hello world")
	hash := HashContent(data)

	// SHA256("hello world") = b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
	expected := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("期望 %q，得到 %q", expected, hash)
	}
}

func TestHashContentDifferent(t *testing.T) {
	hash1 := HashContent([]byte("hello"))
	hash2 := HashContent([]byte("world"))
	if hash1 == hash2 {
		t.Error("不同内容应有不同哈希")
	}
}

// ============================================================================
// 缓存读写测试
// ============================================================================

func TestCacheReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-cache-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	data := []byte("fn greet(name) { return \"hello \" .. name; }")
	hash := HashContent(data)

	// 写入缓存
	if err := WriteCache(tmpDir, hash, data); err != nil {
		t.Fatalf("写入缓存失败: %v", err)
	}

	// 读取缓存
	cached, err := ReadCache(tmpDir, hash)
	if err != nil {
		t.Fatalf("读取缓存失败: %v", err)
	}

	if string(cached) != string(data) {
		t.Errorf("缓存内容不匹配")
	}
}

func TestCacheMiss(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-cache-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = ReadCache(tmpDir, "sha256:nonexistent")
	if !os.IsNotExist(err) {
		t.Errorf("期望 ErrNotExist，得到 %v", err)
	}
}

func TestCacheIntegrityCheck(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-cache-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	data := []byte("original content")
	hash := HashContent(data)

	// 写入正确内容
	if err := WriteCache(tmpDir, hash, data); err != nil {
		t.Fatalf("写入缓存失败: %v", err)
	}

	// 篡改缓存文件
	path := CacheFilePath(tmpDir, hash)
	os.WriteFile(path, []byte("tampered content"), 0644)

	// 读取应该失败（完整性校验）
	_, err = ReadCache(tmpDir, hash)
	if !os.IsNotExist(err) {
		t.Errorf("篡改的缓存应返回 ErrNotExist，得到 %v", err)
	}
}

func TestClearCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-cache-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 写入一些缓存
	WriteCache(tmpDir, HashContent([]byte("a")), []byte("a"))
	WriteCache(tmpDir, HashContent([]byte("b")), []byte("b"))

	// 清空
	if err := ClearCache(tmpDir); err != nil {
		t.Fatalf("清空缓存失败: %v", err)
	}

	if dirExists(tmpDir) {
		t.Error("缓存目录应已被删除")
	}
}

// ============================================================================
// 锁文件测试
// ============================================================================

func TestLockFileCreateAndSave(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-lock-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "jpl.lock.yaml")

	// 加载不存在的锁文件（应创建空的）
	lf, err := LoadLockFile(lockPath)
	if err != nil {
		t.Fatalf("加载锁文件失败: %v", err)
	}
	if len(lf.Remote) != 0 {
		t.Error("新锁文件应为空")
	}

	// 添加条目
	lf.UpdateEntry("https://example.com/lib.jpl", "sha256:abc123", 1024)

	// 保存
	if err := SaveLockFile(lockPath, lf); err != nil {
		t.Fatalf("保存锁文件失败: %v", err)
	}

	// 重新加载
	lf2, err := LoadLockFile(lockPath)
	if err != nil {
		t.Fatalf("重新加载锁文件失败: %v", err)
	}

	entry, ok := lf2.Remote["https://example.com/lib.jpl"]
	if !ok {
		t.Fatal("条目未保存")
	}
	if entry.Hash != "sha256:abc123" {
		t.Errorf("hash 期望 'sha256:abc123'，得到 %q", entry.Hash)
	}
	if entry.Size != 1024 {
		t.Errorf("size 期望 1024，得到 %d", entry.Size)
	}
}

func TestLockFileVerifyHash(t *testing.T) {
	lf := &LockFile{
		Version: 1,
		Remote: map[string]LockEntry{
			"https://example.com/lib.jpl": {Hash: "sha256:abc123"},
		},
	}

	// 匹配
	if err := lf.VerifyHash("https://example.com/lib.jpl", "sha256:abc123", false); err != nil {
		t.Errorf("匹配的 hash 不应报错: %v", err)
	}

	// 不匹配（非 frozen）
	if err := lf.VerifyHash("https://example.com/lib.jpl", "sha256:different", false); err != nil {
		t.Errorf("非 frozen 模式不匹配的 hash 不应报错: %v", err)
	}

	// 不匹配（frozen）
	if err := lf.VerifyHash("https://example.com/lib.jpl", "sha256:different", true); err == nil {
		t.Error("frozen 模式不匹配的 hash 应报错")
	}

	// 新 URL（frozen）
	if err := lf.VerifyHash("https://example.com/new.jpl", "sha256:new", true); err == nil {
		t.Error("frozen 模式新 URL 应报错")
	}

	// 新 URL（非 frozen）
	if err := lf.VerifyHash("https://example.com/new.jpl", "sha256:new", false); err != nil {
		t.Errorf("非 frozen 模式新 URL 不应报错: %v", err)
	}
}

func TestLockFileYAMLFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jpl-lock-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	lockPath := filepath.Join(tmpDir, "jpl.lock.yaml")

	lf := &LockFile{Version: 1, Remote: make(map[string]LockEntry)}
	lf.UpdateEntry("https://example.com/a.jpl", "sha256:aaa111", 512)
	lf.UpdateEntry("https://example.com/b.jpl", "sha256:bbb222", 256)

	if err := SaveLockFile(lockPath, lf); err != nil {
		t.Fatalf("保存失败: %v", err)
	}

	// 读取 YAML 内容验证格式
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}

	content := string(data)
	if !contains(content, "version: 1") {
		t.Error("YAML 应包含 version")
	}
	if !contains(content, "https://example.com/a.jpl") {
		t.Error("YAML 应包含 URL")
	}
	if !contains(content, "sha256:aaa111") {
		t.Error("YAML 应包含 hash")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
