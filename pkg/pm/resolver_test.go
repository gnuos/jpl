package pm

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Resolver 测试
// ============================================================================

func TestResolver_BasicDependencies(t *testing.T) {
	// 创建临时目录结构
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "jpl_modules")
	os.MkdirAll(modulesDir, 0755)

	// 创建清单
	manifest := &Manifest{
		Name:         "test-project",
		Version:      "1.0.0",
		Dependencies: map[string]string{
			// 注意：这些是本地路径测试，实际需要 git 仓库
			// 这里只测试解析逻辑，不实际克隆
		},
	}

	resolver := NewResolver(manifest, modulesDir)
	if resolver == nil {
		t.Fatal("NewResolver() returned nil")
	}

	// 测试空依赖解析
	result, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	if len(result.Packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(result.Packages))
	}
}

func TestResolver_ConflictDetection(t *testing.T) {
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "jpl_modules")
	os.MkdirAll(modulesDir, 0755)

	manifest := &Manifest{
		Name:         "test",
		Version:      "1.0.0",
		Dependencies: make(map[string]string),
	}

	resolver := NewResolver(manifest, modulesDir)

	// 手动添加已解析的包
	resolver.resolved["utils"] = &ResolvedPkg{
		Name:     "utils",
		Source:   "https://github.com/user1/utils.git",
		Resolved: "https://github.com/user1/utils.git",
	}

	// 检查冲突
	err := resolver.checkConflict("utils", "https://github.com/user2/utils.git", "parent")
	if err != nil {
		t.Fatalf("checkConflict() failed: %v", err)
	}

	// 验证冲突被记录
	if len(resolver.conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(resolver.conflicts))
	}

	conflict := resolver.conflicts[0]
	if conflict.Name != "utils" {
		t.Errorf("conflict.Name = %q, want %q", conflict.Name, "utils")
	}
}

func TestResolver_CycleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "jpl_modules")
	os.MkdirAll(modulesDir, 0755)

	manifest := &Manifest{
		Name:         "test",
		Version:      "1.0.0",
		Dependencies: make(map[string]string),
	}

	resolver := NewResolver(manifest, modulesDir)

	// 模拟循环依赖：A -> B -> A
	resolver.colors["A"] = colorGray
	resolver.parents["A"] = ""

	// 尝试解析 B（A 的依赖）
	resolver.colors["B"] = colorGray
	resolver.parents["B"] = "A"

	// 检测循环：B 依赖 A，但 A 正在访问中
	if resolver.colors["A"] == colorGray {
		// 这是循环依赖的检测逻辑
		cycle := resolver.buildCyclePath("B", "A")
		// buildCyclePath 从 current 回溯到 target，然后闭合
		// B -> A -> B
		expected := "  B → A → B"
		if cycle != expected {
			t.Errorf("cycle path = %q, want %q", cycle, expected)
		}
	}
}

// ============================================================================
// PackageCache 测试
// ============================================================================

func TestPackageCache_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewPackageCacheWithDir(tmpDir)

	if cache.CacheDir() != tmpDir {
		t.Errorf("CacheDir() = %q, want %q", cache.CacheDir(), tmpDir)
	}
}

func TestPackageCache_GetCachePath(t *testing.T) {
	cache := NewPackageCacheWithDir("/cache")

	tests := []struct {
		url    string
		commit string
		want   string
	}{
		{
			url:    "https://github.com/user/repo.git",
			commit: "abc123",
			want:   "/cache/user/repo/abc123",
		},
		{
			url:    "https://github.com/user/repo.git@v1.0.0",
			commit: "def456",
			want:   "/cache/user/repo/def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := cache.GetCachePath(tt.url, tt.commit)
			if got != tt.want {
				t.Errorf("GetCachePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPackageCache_Has(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewPackageCacheWithDir(tmpDir)

	// 不存在
	if cache.Has("https://github.com/user/repo.git", "abc123") {
		t.Error("Has() should return false for non-existent cache")
	}

	// 创建缓存目录
	cachePath := cache.GetCachePath("https://github.com/user/repo.git", "abc123")
	os.MkdirAll(cachePath, 0755)

	// 现在应该存在
	if !cache.Has("https://github.com/user/repo.git", "abc123") {
		t.Error("Has() should return true for existing cache")
	}
}

func TestPackageCache_PutAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewPackageCacheWithDir(tmpDir)

	// 创建源目录
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "test.jpl"), []byte("test content"), 0644)

	// 放入缓存
	err := cache.Put("https://github.com/user/repo.git", "abc123", srcDir)
	if err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	// 验证缓存存在
	if !cache.Has("https://github.com/user/repo.git", "abc123") {
		t.Error("Has() should return true after Put()")
	}

	// 从缓存获取
	dstDir := filepath.Join(tmpDir, "dst")
	os.MkdirAll(dstDir, 0755)
	err = cache.Get("https://github.com/user/repo.git", "abc123", dstDir)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filepath.Join(dstDir, "test.jpl"))
	if err != nil {
		t.Fatalf("failed to read cached file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("cached content = %q, want %q", string(content), "test content")
	}
}

// ============================================================================
// parseRepoURL 测试
// ============================================================================

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		url       string
		wantOwner string
		wantRepo  string
	}{
		{
			url:       "https://github.com/user/repo.git",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			url:       "https://github.com/user/repo",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			url:       "https://github.com/user/repo.git@v1.0.0",
			wantOwner: "user",
			wantRepo:  "repo",
		},
		{
			url:       "https://gitlab.com/org/project.git",
			wantOwner: "org",
			wantRepo:  "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			owner, repo := parseRepoURL(tt.url)
			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}
			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

// ============================================================================
// ListInstalled 测试
// ============================================================================

func TestListInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	modulesDir := filepath.Join(tmpDir, "jpl_modules")

	// 创建清单
	manifest := &Manifest{
		Name:    "test",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"utils":       "https://github.com/user/utils.git",
			"http":        "https://github.com/user/http.git",
			"nonexistent": "https://github.com/user/none.git",
		},
	}

	// 创建已安装的包目录
	os.MkdirAll(filepath.Join(modulesDir, "utils"), 0755)
	os.MkdirAll(filepath.Join(modulesDir, "http"), 0755)

	// 写入包清单
	utilsManifest := &Manifest{Name: "utils", Version: "1.2.3"}
	utilsManifest.Save(filepath.Join(modulesDir, "utils"))

	httpManifest := &Manifest{Name: "http", Version: "2.0.0"}
	httpManifest.Save(filepath.Join(modulesDir, "http"))

	// 列出已安装的包
	packages, err := ListInstalled(manifest, modulesDir)
	if err != nil {
		t.Fatalf("ListInstalled() failed: %v", err)
	}

	// 应该只有 2 个（nonexistent 不存在）
	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	// 验证版本
	for _, pkg := range packages {
		switch pkg.Name {
		case "utils":
			if pkg.Version != "1.2.3" {
				t.Errorf("utils version = %q, want %q", pkg.Version, "1.2.3")
			}
		case "http":
			if pkg.Version != "2.0.0" {
				t.Errorf("http version = %q, want %q", pkg.Version, "2.0.0")
			}
		}
	}
}
