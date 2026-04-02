package pm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ============================================================================
// Git 操作
// ============================================================================

// GitRepo git 仓库信息
type GitRepo struct {
	URL    string // 仓库 URL
	Tag    string // 指定的 tag
	Branch string // 指定的分支
	Dir    string // 本地克隆目录
}

// CloneResult 克隆结果
type CloneResult struct {
	CommitHash string // 最终的 commit hash
	Version    string // 版本（tag 或 branch 名称）
}

// CloneAndCheckout 克隆仓库并 checkout 到指定版本
func CloneAndCheckout(url, tag, branch string) (*CloneResult, string, error) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "jpl-pm-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 克隆仓库
	repo := &GitRepo{
		URL:    url,
		Tag:    tag,
		Branch: branch,
		Dir:    tmpDir,
	}

	if err := repo.Clone(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, "", err
	}

	// checkout 到指定版本
	if tag != "" {
		if err := repo.CheckoutTag(tag); err != nil {
			os.RemoveAll(tmpDir)
			return nil, "", err
		}
	} else if branch != "" {
		if err := repo.CheckoutBranch(branch); err != nil {
			os.RemoveAll(tmpDir)
			return nil, "", err
		}
	}

	// 获取 commit hash
	commitHash, err := repo.GetCommitHash()
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, "", err
	}

	result := &CloneResult{
		CommitHash: commitHash,
		Version:    tag,
	}
	if branch != "" && tag == "" {
		result.Version = branch
	}

	return result, tmpDir, nil
}

// CloneWithConstraint 根据版本约束克隆仓库并 checkout 到最佳匹配版本
func CloneWithConstraint(url, constraintStr string) (*CloneResult, string, error) {
	// 获取远程标签列表
	tags, err := ListRemoteTags(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tags) == 0 {
		// 无标签，克隆最新 commit
		return CloneAndCheckout(url, "", "")
	}

	// 选择满足约束的最佳版本
	bestTag, err := SelectBestVersion(tags, constraintStr)
	if err != nil {
		return nil, "", fmt.Errorf("no version satisfies constraint %q: %w", constraintStr, err)
	}

	// 克隆并 checkout 到选定的 tag
	return CloneAndCheckout(url, bestTag, "")
}

// Clone 克隆仓库
func (r *GitRepo) Clone() error {
	args := []string{"clone", "--depth", "1"}

	// 如果指定了分支
	if r.Branch != "" {
		args = append(args, "--branch", r.Branch)
	}

	args = append(args, r.URL, r.Dir)

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, string(output))
	}

	return nil
}

// CheckoutTag checkout 到指定 tag
func (r *GitRepo) CheckoutTag(tag string) error {
	cmd := exec.Command("git", "checkout", "tags/"+tag)
	cmd.Dir = r.Dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout tag %q failed: %w\n%s", tag, err, string(output))
	}
	return nil
}

// CheckoutBranch checkout 到指定分支
func (r *GitRepo) CheckoutBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = r.Dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout branch %q failed: %w\n%s", branch, err, string(output))
	}
	return nil
}

// GetCommitHash 获取当前 HEAD 的 commit hash
func (r *GitRepo) GetCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.Dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetShortCommitHash 获取短 commit hash（7 位）
func (r *GitRepo) GetShortCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = r.Dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ListTags 列出仓库的所有标签（按版本排序）
func (r *GitRepo) ListTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname")
	cmd.Dir = r.Dir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git tag failed: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result, nil
}

// ListRemoteTags 列出远程仓库的标签（无需克隆）
func ListRemoteTags(url string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", url)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote failed: %w", err)
	}

	var tags []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		// 格式: <hash>	refs/tags/<tag>
		if idx := strings.Index(line, "refs/tags/"); idx >= 0 {
			tag := line[idx+10:]
			// 跳过 ^{} 后缀的 peeled tag
			if !strings.HasSuffix(tag, "^{}") {
				tags = append(tags, tag)
			}
		}
	}
	return tags, nil
}

// ============================================================================
// 并行克隆
// ============================================================================

// CloneJob 单个克隆任务
type CloneJob struct {
	Name       string // 包名
	URL        string // 仓库 URL
	Tag        string // 指定的 tag
	Branch     string // 指定的分支
	Constraint string // 版本约束
}

// CloneJobResult 单个克隆任务结果
type CloneJobResult struct {
	Name        string       // 包名
	Result      *CloneResult // 克隆结果
	TmpDir      string       // 临时目录
	Err         error        // 错误
	IsFromCache bool         // 是否来自缓存
}

// ParallelClone 并行克隆多个仓库
// jobs: 最大并发数（<=0 使用默认值 4）
func ParallelClone(jobs []CloneJob, maxConcurrency int, cache *PackageCache, verbose bool) []CloneJobResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}

	results := make([]CloneJobResult, len(jobs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // 信号量控制并发

	for i, job := range jobs {
		wg.Add(1)
		go func(idx int, j CloneJob) {
			defer wg.Done()
			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			res := CloneJobResult{Name: j.Name}

			// 尝试从缓存获取
			if cache != nil {
				// 如果有约束，需要先获取 tags 再确定版本
				if j.Constraint != "" {
					tags, err := ListRemoteTags(j.URL)
					if err == nil && len(tags) > 0 {
						bestTag, err := SelectBestVersion(tags, j.Constraint)
						if err == nil {
							j.Tag = bestTag
							j.Branch = ""
						}
					}
				}

				// 获取 commit hash 需要先克隆，这里只检查已缓存的
				// 缓存查找需要 commit hash，但此时还不知道
				// 所以并行克隆阶段不使用缓存，由调用方在安装阶段处理
			}

			// 克隆
			var cloneResult *CloneResult
			var tmpDir string
			var err error

			if j.Constraint != "" {
				cloneResult, tmpDir, err = CloneWithConstraint(j.URL, j.Constraint)
			} else {
				cloneResult, tmpDir, err = CloneAndCheckout(j.URL, j.Tag, j.Branch)
			}

			res.Result = cloneResult
			res.TmpDir = tmpDir
			res.Err = err

			if verbose && err == nil {
				fmt.Fprintf(os.Stderr, "[verbose] %s: cloned %s @ %s\n", j.Name, j.URL, cloneResult.CommitHash[:7])
			}

			results[idx] = res
		}(i, job)
	}

	wg.Wait()
	return results
}

// ============================================================================
// 包安装操作
// ============================================================================

// InstallPackage 将克隆的仓库安装到 jpl_modules/
func InstallPackage(sourceDir, modulesDir, name string) error {
	targetDir := filepath.Join(modulesDir, name)

	// 如果目标目录已存在，先删除
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("failed to remove existing package %q: %w", name, err)
		}
	}

	// 确保 modules 目录存在
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create modules directory: %w", err)
	}

	// 复制仓库内容到目标目录（排除 .git）
	if err := copyDir(sourceDir, targetDir); err != nil {
		return fmt.Errorf("failed to install package %q: %w", name, err)
	}

	// 删除 .git 目录
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		os.RemoveAll(gitDir)
	}

	return nil
}

// RemovePackage 从 jpl_modules/ 移除包
func RemovePackage(modulesDir, name string) error {
	targetDir := filepath.Join(modulesDir, name)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("package %q not installed", name)
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to remove package %q: %w", name, err)
	}

	return nil
}

// copyDir 递归复制目录（跳过 .git）
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 跳过 .git 目录
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}
