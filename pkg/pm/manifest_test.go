package pm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnuos/jpl/pkg/task"
)

// ============================================================================
// Manifest 测试
// ============================================================================

func TestManifestSaveAndLoad(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建清单
	m := &Manifest{
		Name:    "test-project",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"utils": "https://github.com/user/jpl-utils.git",
		},
	}

	// 保存
	if err := m.Save(tmpDir); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// 验证文件存在
	path := filepath.Join(tmpDir, ManifestFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("manifest file not created")
	}

	// 加载
	loaded, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest() failed: %v", err)
	}

	// 验证
	if loaded.Name != m.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, m.Name)
	}
	if loaded.Version != m.Version {
		t.Errorf("Version = %q, want %q", loaded.Version, m.Version)
	}
	if len(loaded.Dependencies) != 1 {
		t.Fatalf("Dependencies count = %d, want 1", len(loaded.Dependencies))
	}
	if loaded.Dependencies["utils"] != "https://github.com/user/jpl-utils.git" {
		t.Errorf("Dependencies[utils] = %q, want %q",
			loaded.Dependencies["utils"], "https://github.com/user/jpl-utils.git")
	}
}

func TestLoadOrCreateManifest_New(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := LoadOrCreateManifest(tmpDir)
	if err != nil {
		t.Fatalf("LoadOrCreateManifest() failed: %v", err)
	}

	if m.Name != filepath.Base(tmpDir) {
		t.Errorf("Name = %q, want %q", m.Name, filepath.Base(tmpDir))
	}
	if m.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", m.Version, "0.1.0")
	}
	if m.Dependencies == nil {
		t.Error("Dependencies is nil")
	}
}

func TestManifestDependencyOperations(t *testing.T) {
	m := &Manifest{
		Dependencies: make(map[string]string),
	}

	// 添加依赖
	m.AddDependency("utils", "https://github.com/user/utils.git")
	if !m.HasDependency("utils") {
		t.Error("HasDependency(utils) = false, want true")
	}

	// 移除依赖
	if !m.RemoveDependency("utils") {
		t.Error("RemoveDependency(utils) = false, want true")
	}
	if m.HasDependency("utils") {
		t.Error("HasDependency(utils) = true after removal")
	}

	// 移除不存在的依赖
	if m.RemoveDependency("nonexistent") {
		t.Error("RemoveDependency(nonexistent) = true, want false")
	}
}

// ============================================================================
// SourceInfo 测试
// ============================================================================

func TestParseSource_GitURL(t *testing.T) {
	tests := []struct {
		source     string
		wantURL    string
		wantTag    string
		wantBranch string
	}{
		{
			source:  "https://github.com/user/repo.git",
			wantURL: "https://github.com/user/repo.git",
		},
		{
			source:  "https://github.com/user/repo.git@v1.0.0",
			wantURL: "https://github.com/user/repo.git",
			wantTag: "v1.0.0",
		},
		{
			source:     "https://github.com/user/repo.git#main",
			wantURL:    "https://github.com/user/repo.git",
			wantBranch: "main",
		},
		{
			source:  "../my-lib",
			wantURL: "../my-lib",
		},
		{
			source:  "/absolute/path",
			wantURL: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			info, err := ParseSource(tt.source)
			if err != nil {
				t.Fatalf("ParseSource(%q) failed: %v", tt.source, err)
			}
			if info.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", info.URL, tt.wantURL)
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

func TestParseSource_Empty(t *testing.T) {
	_, err := ParseSource("")
	if err == nil {
		t.Error("ParseSource(\"\") should return error")
	}
}

// ============================================================================
// ResolveName 测试
// ============================================================================

func TestResolveName(t *testing.T) {
	tests := []struct {
		source       string
		explicitName string
		want         string
	}{
		{
			source: "https://github.com/user/jpl-utils.git",
			want:   "jpl-utils",
		},
		{
			source:       "https://github.com/user/jpl-utils.git",
			explicitName: "my-utils",
			want:         "my-utils",
		},
		{
			source: "../my-lib",
			want:   "my-lib",
		},
		{
			source: "https://github.com/user/repo",
			want:   "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			info, err := ParseSource(tt.source)
			if err != nil {
				t.Fatalf("ParseSource(%q) failed: %v", tt.source, err)
			}
			got := ResolveName(info, tt.explicitName)
			if got != tt.want {
				t.Errorf("ResolveName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// ============================================================================
// rfind 测试
// ============================================================================

func TestRfind(t *testing.T) {
	tests := []struct {
		s    string
		c    byte
		want int
	}{
		{"hello@world", '@', 5},
		{"hello#world", '#', 5},
		{"hello", '@', -1},
		{"@start", '@', 0},
		{"end@", '@', 3},
		{"", '@', -1},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := rfind(tt.s, tt.c)
			if got != tt.want {
				t.Errorf("rfind(%q, %q) = %d, want %d", tt.s, string(tt.c), got, tt.want)
			}
		})
	}
}

// ============================================================================
// Manifest Tasks 测试
// ============================================================================

func TestManifestSaveAndLoad_WithTasks(t *testing.T) {
	tmpDir := t.TempDir()

	m := &Manifest{
		Name:    "test-project",
		Version: "1.0.0",
		Dependencies: map[string]string{
			"utils": "https://github.com/user/jpl-utils.git",
		},
		Tasks: map[string]task.TaskDef{
			"test":  {Cmd: "jpl run tests/main.jpl"},
			"build": {Cmd: "jpl run build.jpl", Deps: []string{"clean"}},
			"clean": {Cmd: "rm -rf build"},
		},
	}

	// 保存
	if err := m.Save(tmpDir); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// 加载
	path := filepath.Join(tmpDir, ManifestFileName)
	loaded, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("LoadManifest() failed: %v", err)
	}

	// 验证 tasks
	if len(loaded.Tasks) != 3 {
		t.Fatalf("Tasks count = %d, want 3", len(loaded.Tasks))
	}

	if loaded.Tasks["test"].Cmd != "jpl run tests/main.jpl" {
		t.Errorf("test.cmd = %q, want %q", loaded.Tasks["test"].Cmd, "jpl run tests/main.jpl")
	}
	if loaded.Tasks["build"].Cmd != "jpl run build.jpl" {
		t.Errorf("build.cmd = %q, want %q", loaded.Tasks["build"].Cmd, "jpl run build.jpl")
	}
	if len(loaded.Tasks["build"].Deps) != 1 || loaded.Tasks["build"].Deps[0] != "clean" {
		t.Errorf("build.deps = %v, want [clean]", loaded.Tasks["build"].Deps)
	}

	// 验证 HasTask / GetTask
	if !loaded.HasTask("test") {
		t.Error("HasTask(test) = false, want true")
	}
	if loaded.HasTask("nonexistent") {
		t.Error("HasTask(nonexistent) = true, want false")
	}

	taskDef, ok := loaded.GetTask("build")
	if !ok {
		t.Fatal("GetTask(build) not found")
	}
	if len(taskDef.Deps) != 1 || taskDef.Deps[0] != "clean" {
		t.Errorf("GetTask(build).Deps = %v, want [clean]", taskDef.Deps)
	}
}
