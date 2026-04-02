package pm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/pkg/pm"
)

// ============================================================================
// 集成测试：完整的 add → install → remove 流程
// ============================================================================

// setupTestProject 创建测试项目目录
func setupTestProject(t *testing.T) string {
	t.Helper()
	projectDir := t.TempDir()

	// 创建清单
	manifest := &pm.Manifest{
		Name:         "test-project",
		Version:      "1.0.0",
		Dependencies: make(map[string]string),
	}
	if err := manifest.Save(projectDir); err != nil {
		t.Fatalf("failed to save manifest: %v", err)
	}

	// 创建 jpl_modules 目录
	modulesDir := filepath.Join(projectDir, "jpl_modules")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		t.Fatalf("failed to create modules dir: %v", err)
	}

	return projectDir
}

// createMockPackage 创建模拟的包目录（模拟 git clone 结果）
func createMockPackage(t *testing.T, dir, name, version string) string {
	t.Helper()
	pkgDir := filepath.Join(dir, name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}

	// 创建 index.jpl
	indexContent := "// " + name + " package\nputs \"Hello from " + name + "\"\n"
	if err := os.WriteFile(filepath.Join(pkgDir, "index.jpl"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("failed to create index.jpl: %v", err)
	}

	// 创建 jpl.json（如果有版本）
	if version != "" {
		pkgManifest := &pm.Manifest{
			Name:         name,
			Version:      version,
			Dependencies: make(map[string]string),
		}
		if err := pkgManifest.Save(pkgDir); err != nil {
			t.Fatalf("failed to save package manifest: %v", err)
		}
	}

	return pkgDir
}

// ============================================================================
// 测试清单文件流程
// ============================================================================

func TestIntegration_ManifestReadWrite(t *testing.T) {
	projectDir := setupTestProject(t)

	// 加载清单
	manifest, err := pm.LoadManifest(filepath.Join(projectDir, pm.ManifestFileName))
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// 验证初始状态
	if manifest.Name != "test-project" {
		t.Errorf("Name = %q, want %q", manifest.Name, "test-project")
	}
	if len(manifest.Dependencies) != 0 {
		t.Errorf("Dependencies count = %d, want 0", len(manifest.Dependencies))
	}

	// 添加依赖
	manifest.AddDependency("utils", "https://github.com/user/utils.git")
	manifest.AddDependency("http", "https://github.com/user/http.git@v1.0.0")

	// 保存
	if err := manifest.Save(projectDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 重新加载验证
	loaded, err := pm.LoadManifest(filepath.Join(projectDir, pm.ManifestFileName))
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if !loaded.HasDependency("utils") {
		t.Error("missing dependency: utils")
	}
	if !loaded.HasDependency("http") {
		t.Error("missing dependency: http")
	}
	if loaded.Dependencies["utils"] != "https://github.com/user/utils.git" {
		t.Errorf("utils source = %q, want %q", loaded.Dependencies["utils"], "https://github.com/user/utils.git")
	}

	// 移除依赖
	loaded.RemoveDependency("utils")
	if loaded.HasDependency("utils") {
		t.Error("utils should be removed")
	}
}

// ============================================================================
// 测试包安装流程
// ============================================================================

func TestIntegration_InstallAndRemovePackage(t *testing.T) {
	projectDir := setupTestProject(t)
	modulesDir := filepath.Join(projectDir, "jpl_modules")

	// 创建模拟的源目录
	srcDir := t.TempDir()
	mockPkg := createMockPackage(t, srcDir, "utils", "1.2.3")

	// 安装包
	if err := pm.InstallPackage(mockPkg, modulesDir, "utils"); err != nil {
		t.Fatalf("InstallPackage failed: %v", err)
	}

	// 验证安装
	installedDir := filepath.Join(modulesDir, "utils")
	if _, err := os.Stat(installedDir); os.IsNotExist(err) {
		t.Fatal("package not installed")
	}

	// 验证文件存在
	if _, err := os.Stat(filepath.Join(installedDir, "index.jpl")); os.IsNotExist(err) {
		t.Error("index.jpl not found")
	}
	if _, err := os.Stat(filepath.Join(installedDir, "jpl.json")); os.IsNotExist(err) {
		t.Error("jpl.json not found")
	}

	// 验证 .git 目录被删除
	if _, err := os.Stat(filepath.Join(installedDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should be removed")
	}

	// 移除包
	if err := pm.RemovePackage(modulesDir, "utils"); err != nil {
		t.Fatalf("RemovePackage failed: %v", err)
	}

	// 验证移除
	if _, err := os.Stat(installedDir); !os.IsNotExist(err) {
		t.Error("package should be removed")
	}
}

// ============================================================================
// 测试完整流程：init → add → install → list → remove
// ============================================================================

func TestIntegration_FullWorkflow(t *testing.T) {
	projectDir := t.TempDir()

	// 1. 初始化项目
	initOpts := &pm.InitOptions{
		Dir:  projectDir,
		Name: "my-project",
	}
	result, err := pm.InitProject(initOpts)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// 验证初始化结果
	if result.Dir != projectDir {
		t.Errorf("Dir = %q, want %q", result.Dir, projectDir)
	}

	// 验证文件存在
	if _, err := os.Stat(filepath.Join(projectDir, pm.ManifestFileName)); os.IsNotExist(err) {
		t.Error("jpl.json not created")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "main.jpl")); os.IsNotExist(err) {
		t.Error("main.jpl not created")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "jpl_modules")); os.IsNotExist(err) {
		t.Error("jpl_modules not created")
	}

	// 2. 加载清单并添加依赖
	manifest, err := pm.LoadManifest(filepath.Join(projectDir, pm.ManifestFileName))
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// 创建模拟的源包
	srcDir := t.TempDir()
	createMockPackage(t, srcDir, "utils", "1.0.0")
	createMockPackage(t, srcDir, "http", "2.0.0")

	// 添加依赖到清单
	manifest.AddDependency("utils", "file://"+filepath.Join(srcDir, "utils"))
	manifest.AddDependency("http", "file://"+filepath.Join(srcDir, "http"))
	if err := manifest.Save(projectDir); err != nil {
		t.Fatalf("Save manifest failed: %v", err)
	}

	// 3. 安装依赖
	modulesDir := filepath.Join(projectDir, "jpl_modules")
	if err := pm.InstallPackage(filepath.Join(srcDir, "utils"), modulesDir, "utils"); err != nil {
		t.Fatalf("Install utils failed: %v", err)
	}
	if err := pm.InstallPackage(filepath.Join(srcDir, "http"), modulesDir, "http"); err != nil {
		t.Fatalf("Install http failed: %v", err)
	}

	// 4. 列出依赖
	packages, err := pm.ListInstalled(manifest, modulesDir)
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}
	if len(packages) != 2 {
		t.Errorf("packages count = %d, want 2", len(packages))
	}

	// 5. 移除依赖
	if err := pm.RemovePackage(modulesDir, "utils"); err != nil {
		t.Fatalf("RemovePackage failed: %v", err)
	}
	manifest.RemoveDependency("utils")
	if err := manifest.Save(projectDir); err != nil {
		t.Fatalf("Save manifest failed: %v", err)
	}

	// 验证移除后的状态
	packages, err = pm.ListInstalled(manifest, modulesDir)
	if err != nil {
		t.Fatalf("ListInstalled failed: %v", err)
	}
	if len(packages) != 1 {
		t.Errorf("packages count = %d, want 1", len(packages))
	}
	if packages[0].Name != "http" {
		t.Errorf("remaining package = %q, want %q", packages[0].Name, "http")
	}
}

// ============================================================================
// 测试依赖解析
// ============================================================================

func TestIntegration_ResolverEmpty(t *testing.T) {
	projectDir := setupTestProject(t)
	modulesDir := filepath.Join(projectDir, "jpl_modules")

	// 加载清单
	manifest, err := pm.LoadManifest(filepath.Join(projectDir, pm.ManifestFileName))
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	// 创建解析器
	resolver := pm.NewResolver(manifest, modulesDir)

	// 解析空依赖
	result, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if len(result.Packages) != 0 {
		t.Errorf("packages count = %d, want 0", len(result.Packages))
	}
	if len(result.Conflicts) != 0 {
		t.Errorf("conflicts count = %d, want 0", len(result.Conflicts))
	}
	if len(result.Order) != 0 {
		t.Errorf("order count = %d, want 0", len(result.Order))
	}
}

// ============================================================================
// 测试源地址解析
// ============================================================================

func TestIntegration_SourceParsing(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantURL    string
		wantConst  string
		wantTag    string
		wantBranch string
	}{
		{
			name:      "git url with constraint",
			source:    "https://github.com/user/lib.git@^1.2.3",
			wantURL:   "https://github.com/user/lib.git",
			wantConst: "^1.2.3",
		},
		{
			name:    "git url with tag",
			source:  "https://github.com/user/lib.git@v1.0.0",
			wantURL: "https://github.com/user/lib.git",
			wantTag: "v1.0.0",
		},
		{
			name:       "git url with branch",
			source:     "https://github.com/user/lib.git#main",
			wantURL:    "https://github.com/user/lib.git",
			wantBranch: "main",
		},
		{
			name:    "local path",
			source:  "../my-lib",
			wantURL: "../my-lib",
		},
		{
			name:    "git url plain",
			source:  "https://github.com/user/lib.git",
			wantURL: "https://github.com/user/lib.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := pm.ParseSourceWithConstraint(tt.source)
			if err != nil {
				t.Fatalf("ParseSourceWithConstraint failed: %v", err)
			}
			if info.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", info.URL, tt.wantURL)
			}
			if info.Constraint != tt.wantConst {
				t.Errorf("Constraint = %q, want %q", info.Constraint, tt.wantConst)
			}
			if info.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", info.Tag, tt.wantTag)
			}
			if info.Branch != tt.wantBranch {
				t.Errorf("Branch = %q, want %q", info.Branch, tt.wantBranch)
			}
		})
	}
}

// ============================================================================
// 测试缓存
// ============================================================================

func TestIntegration_Cache(t *testing.T) {
	cacheDir := t.TempDir()
	cache := pm.NewPackageCacheWithDir(cacheDir)

	// 创建源目录
	srcDir := t.TempDir()
	testFile := filepath.Join(srcDir, "test.jpl")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	url := "https://github.com/user/lib.git"
	commit := "abc123def456"

	// 验证缓存未命中
	if cache.Has(url, commit) {
		t.Error("cache should be empty")
	}

	// 放入缓存
	if err := cache.Put(url, commit, srcDir); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// 验证缓存命中
	if !cache.Has(url, commit) {
		t.Error("cache should have the package")
	}

	// 从缓存获取
	dstDir := t.TempDir()
	if err := cache.Get(url, commit, dstDir); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(filepath.Join(dstDir, "test.jpl"))
	if err != nil {
		t.Fatalf("failed to read cached file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("content = %q, want %q", string(content), "test content")
	}
}
