package pm

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// 全局包缓存 (~/.jpl/packages/)
// ============================================================================

// PackageCache 全局包缓存
type PackageCache struct {
	cacheDir string // ~/.jpl/packages/
}

// NewPackageCache 创建包缓存
func NewPackageCache() (*PackageCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".jpl", "packages")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &PackageCache{cacheDir: cacheDir}, nil
}

// NewPackageCacheWithDir 使用指定目录创建缓存
func NewPackageCacheWithDir(cacheDir string) *PackageCache {
	return &PackageCache{cacheDir: cacheDir}
}

// CacheDir 返回缓存目录
func (c *PackageCache) CacheDir() string {
	return c.cacheDir
}

// GetCachePath 获取包的缓存路径
// 格式: ~/.jpl/packages/<owner>/<repo>/<commit>/
func (c *PackageCache) GetCachePath(url, commit string) string {
	owner, repo := parseRepoURL(url)
	return filepath.Join(c.cacheDir, owner, repo, commit)
}

// Has 检查缓存中是否有指定版本
func (c *PackageCache) Has(url, commit string) bool {
	if commit == "" {
		return false
	}
	path := c.GetCachePath(url, commit)
	return dirExists(path)
}

// Get 从缓存获取包，复制到目标目录
func (c *PackageCache) Get(url, commit, targetDir string) error {
	if !c.Has(url, commit) {
		return fmt.Errorf("cache miss: %s @ %s", url, commit)
	}

	srcDir := c.GetCachePath(url, commit)
	return copyDir(srcDir, targetDir)
}

// Put 将包放入缓存
func (c *PackageCache) Put(url, commit, sourceDir string) error {
	if commit == "" {
		return fmt.Errorf("empty commit hash")
	}

	cachePath := c.GetCachePath(url, commit)

	// 如果已存在，跳过
	if dirExists(cachePath) {
		return nil
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 复制到缓存
	if err := copyDir(sourceDir, cachePath); err != nil {
		// 清理失败的缓存
		os.RemoveAll(cachePath)
		return fmt.Errorf("failed to cache package: %w", err)
	}

	// 删除 .git 目录
	gitDir := filepath.Join(cachePath, ".git")
	os.RemoveAll(gitDir)

	return nil
}

// Clear 清空缓存
func (c *PackageCache) Clear() error {
	if !dirExists(c.cacheDir) {
		return nil
	}
	return os.RemoveAll(c.cacheDir)
}

// Size 计算缓存大小（字节）
func (c *PackageCache) Size() (int64, error) {
	var size int64
	err := filepath.Walk(c.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// List 列出缓存的包
func (c *PackageCache) List() ([]CachedPackage, error) {
	var packages []CachedPackage

	if !dirExists(c.cacheDir) {
		return packages, nil
	}

	// 遍历 ~/.jpl/packages/<owner>/<repo>/<commit>/
	owners, err := os.ReadDir(c.cacheDir)
	if err != nil {
		return nil, err
	}

	for _, owner := range owners {
		if !owner.IsDir() {
			continue
		}
		ownerDir := filepath.Join(c.cacheDir, owner.Name())

		repos, err := os.ReadDir(ownerDir)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}
			repoDir := filepath.Join(ownerDir, repo.Name())

			commits, err := os.ReadDir(repoDir)
			if err != nil {
				continue
			}

			for _, commit := range commits {
				if !commit.IsDir() {
					continue
				}
				packages = append(packages, CachedPackage{
					Owner:  owner.Name(),
					Repo:   repo.Name(),
					Commit: commit.Name(),
					Path:   filepath.Join(repoDir, commit.Name()),
				})
			}
		}
	}

	return packages, nil
}

// CachedPackage 缓存的包信息
type CachedPackage struct {
	Owner  string
	Repo   string
	Commit string
	Path   string
}

// ============================================================================
// 辅助函数
// ============================================================================

// dirExists 检查目录是否存在
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// parseRepoURL 从 URL 解析 owner 和 repo
// https://github.com/user/repo.git → ("user", "repo")
func parseRepoURL(url string) (string, string) {
	// 移除协议
	s := url
	if idx := strings.Index(s, "://"); idx >= 0 {
		s = s[idx+3:]
	}

	// 移除域名后的路径
	if idx := strings.Index(s, "/"); idx >= 0 {
		s = s[idx+1:]
	}

	// 分割 owner/repo
	parts := strings.SplitN(s, "/", 2)
	if len(parts) < 2 {
		// 没有 owner，使用 hash 作为 owner
		h := sha256.Sum256([]byte(url))
		return "unknown", hex.EncodeToString(h[:8])
	}

	owner := parts[0]
	repo := parts[1]

	// 先移除 @tag 或 #branch
	if idx := strings.Index(repo, "@"); idx >= 0 {
		repo = repo[:idx]
	}
	if idx := strings.Index(repo, "#"); idx >= 0 {
		repo = repo[:idx]
	}

	// 再移除 .git 后缀
	if strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-4]
	}

	return owner, repo
}
