package pm

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ============================================================================
// ParallelClone 测试
// ============================================================================

func TestParallelClone_EmptyJobs(t *testing.T) {
	results := ParallelClone(nil, 4, nil, false)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestParallelClone_SingleJob(t *testing.T) {
	// 创建一个本地 git 仓库作为测试源
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, repoDir)

	jobs := []CloneJob{
		{Name: "test-pkg", URL: repoDir},
	}

	results := ParallelClone(jobs, 4, nil, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if results[0].Result == nil {
		t.Fatal("expected CloneResult, got nil")
	}
	if results[0].TmpDir == "" {
		t.Fatal("expected TmpDir, got empty")
	}

	os.RemoveAll(results[0].TmpDir)
}

func TestParallelClone_MultipleJobs(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个本地 git 仓库
	numRepos := 3
	for i := 0; i < numRepos; i++ {
		repoDir := filepath.Join(tmpDir, "repo"+string(rune('0'+i)))
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatal(err)
		}
		initTestGitRepo(t, repoDir)
	}

	jobs := []CloneJob{
		{Name: "pkg-a", URL: filepath.Join(tmpDir, "repo0")},
		{Name: "pkg-b", URL: filepath.Join(tmpDir, "repo1")},
		{Name: "pkg-c", URL: filepath.Join(tmpDir, "repo2")},
	}

	results := ParallelClone(jobs, 4, nil, false)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("job %s: unexpected error: %v", r.Name, r.Err)
		}
		if r.Result == nil {
			t.Errorf("job %s: expected CloneResult, got nil", r.Name)
		}
		os.RemoveAll(r.TmpDir)
	}
}

func TestParallelClone_ConcurrencyLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个本地仓库
	numRepos := 6
	for i := 0; i < numRepos; i++ {
		repoDir := filepath.Join(tmpDir, "repo"+string(rune('0'+i)))
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatal(err)
		}
		initTestGitRepo(t, repoDir)
	}

	jobs := make([]CloneJob, numRepos)
	for i := 0; i < numRepos; i++ {
		jobs[i] = CloneJob{
			Name: "pkg" + string(rune('0'+i)),
			URL:  filepath.Join(tmpDir, "repo"+string(rune('0'+i))),
		}
	}

	// 并发限制为 2
	results := ParallelClone(jobs, 2, nil, false)

	// 验证所有任务都完成了
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("job %s: unexpected error: %v", r.Name, r.Err)
		}
		os.RemoveAll(r.TmpDir)
	}
}

func TestParallelClone_InvalidURL(t *testing.T) {
	jobs := []CloneJob{
		{Name: "bad-pkg", URL: "https://invalid.example.com/nonexistent.git"},
	}

	results := ParallelClone(jobs, 4, nil, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestParallelClone_MixedResults(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建一个有效的仓库
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, repoDir)

	jobs := []CloneJob{
		{Name: "good-pkg", URL: repoDir},
		{Name: "bad-pkg", URL: "https://invalid.example.com/nonexistent.git"},
	}

	results := ParallelClone(jobs, 4, nil, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// 找到 good-pkg 和 bad-pkg 的结果
	for _, r := range results {
		if r.Name == "good-pkg" {
			if r.Err != nil {
				t.Errorf("good-pkg: unexpected error: %v", r.Err)
			}
			os.RemoveAll(r.TmpDir)
		} else if r.Name == "bad-pkg" {
			if r.Err == nil {
				t.Error("bad-pkg: expected error, got nil")
			}
		}
	}
}

func TestParallelClone_DefaultConcurrency(t *testing.T) {
	// 测试并发数 <= 0 时使用默认值 4
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}
	initTestGitRepo(t, repoDir)

	jobs := []CloneJob{{Name: "pkg", URL: repoDir}}

	// 并发数为 0，应使用默认值
	results := ParallelClone(jobs, 0, nil, false)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatal("default concurrency failed")
	}
	os.RemoveAll(results[0].TmpDir)

	// 并发数为负数，应使用默认值
	results = ParallelClone(jobs, -1, nil, false)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatal("negative concurrency failed")
	}
	os.RemoveAll(results[0].TmpDir)
}

// ============================================================================
// 辅助函数
// ============================================================================

// initTestGitRepo 初始化一个测试用的 git 仓库
func initTestGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command failed: %v\n%s", err, out)
		}
	}

	// 创建一个文件并提交
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()

	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}
}
