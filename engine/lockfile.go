package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

// ============================================================================
// 锁文件 (jpl.lock.yaml)
// ============================================================================

// LockFile JPL 模块锁文件
type LockFile struct {
	Version   int                  `yaml:"version"`
	Generated string               `yaml:"generated"`
	Remote    map[string]LockEntry `yaml:"remote,omitempty"`
	Packages  map[string]PkgEntry  `yaml:"packages,omitempty"`
}

// LockEntry 单个远程模块的锁记录
type LockEntry struct {
	Hash       string `yaml:"hash"`                 // sha256:<hex>
	Downloaded string `yaml:"downloaded,omitempty"` // RFC3339 时间
	Size       int64  `yaml:"size,omitempty"`       // 文件大小（字节）
}

// PkgEntry 包管理器依赖的锁记录
type PkgEntry struct {
	Source       string   `yaml:"source"`                 // 源地址
	Resolved     string   `yaml:"resolved"`               // 解析后的 URL
	Version      string   `yaml:"version,omitempty"`      // 版本（tag/branch）
	Commit       string   `yaml:"commit"`                 // commit hash
	Hash         string   `yaml:"hash"`                   // sha256:<hex>
	Dependencies []string `yaml:"dependencies,omitempty"` // 传递依赖列表
}

// LoadLockFile 加载锁文件
func LoadLockFile(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LockFile{
				Version:  1,
				Remote:   make(map[string]LockEntry),
				Packages: make(map[string]PkgEntry),
			}, nil
		}
		return nil, err
	}

	var lf LockFile
	if err := yaml.Unmarshal(data, &lf); err != nil {
		return nil, fmt.Errorf("failed to parse lock file %q: %w", path, err)
	}

	if lf.Remote == nil {
		lf.Remote = make(map[string]LockEntry)
	}
	if lf.Packages == nil {
		lf.Packages = make(map[string]PkgEntry)
	}

	return &lf, nil
}

// SaveLockFile 保存锁文件
func SaveLockFile(path string, lf *LockFile) error {
	lf.Generated = time.Now().UTC().Format(time.RFC3339)

	data, err := yaml.Marshal(lf)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create lock file directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// VerifyHash 校验 URL 的 hash 是否匹配锁文件记录
// 返回 nil 表示匹配，error 表示不匹配或 frozen 模式下未找到
func (lf *LockFile) VerifyHash(url, hash string, frozen bool) error {
	entry, exists := lf.Remote[url]
	if !exists {
		if frozen {
			return fmt.Errorf("frozen lockfile: URL %q not in lock file", url)
		}
		return nil // 非 frozen 模式，新 URL 允许
	}

	if entry.Hash != hash {
		if frozen {
			return fmt.Errorf("frozen lockfile: hash mismatch for %q\n  expected: %s\n  got:      %s",
				url, entry.Hash, hash)
		}
		// 非 frozen 模式，hash 变更允许（后续会更新锁文件）
	}

	return nil
}

// UpdateEntry 更新锁文件中的条目
func (lf *LockFile) UpdateEntry(url, hash string, size int64) {
	lf.Remote[url] = LockEntry{
		Hash:       hash,
		Downloaded: time.Now().UTC().Format(time.RFC3339),
		Size:       size,
	}
}

// UpdatePkgEntry 更新包管理器依赖条目
func (lf *LockFile) UpdatePkgEntry(name, source, resolved, version, commit, hash string, deps []string) {
	lf.Packages[name] = PkgEntry{
		Source:       source,
		Resolved:     resolved,
		Version:      version,
		Commit:       commit,
		Hash:         hash,
		Dependencies: deps,
	}
}

// RemovePkgEntry 移除包管理器依赖条目
func (lf *LockFile) RemovePkgEntry(name string) {
	delete(lf.Packages, name)
}

// GetPkgEntry 获取包管理器依赖条目
func (lf *LockFile) GetPkgEntry(name string) (PkgEntry, bool) {
	entry, exists := lf.Packages[name]
	return entry, exists
}

// ============================================================================
// 磁盘缓存 (~/.jpl/cache/)
// ============================================================================

// DefaultCacheDir 返回默认缓存目录 ~/.jpl/cache/
func DefaultCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".jpl", "cache"), nil
}

// HashContent 计算内容的 SHA256 哈希
func HashContent(data []byte) string {
	h := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(h[:])
}

// CacheFilePath 根据 hash 生成缓存文件路径
// hash: "sha256:abc123..." → cacheDir/abc123...jpl
func CacheFilePath(cacheDir, hash string) string {
	// 去掉 "sha256:" 前缀
	h := hash
	if len(h) > 7 && h[:7] == "sha256:" {
		h = h[7:]
	}
	return filepath.Join(cacheDir, h+".jpl")
}

// ReadCache 从磁盘缓存读取内容
func ReadCache(cacheDir, hash string) ([]byte, error) {
	path := CacheFilePath(cacheDir, hash)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err // os.ErrNotExist 表示缓存未命中
	}

	// 校验缓存完整性
	actualHash := HashContent(data)
	if actualHash != hash {
		// 缓存损坏，删除并返回未命中
		os.Remove(path)
		return nil, os.ErrNotExist
	}

	return data, nil
}

// WriteCache 写入磁盘缓存
func WriteCache(cacheDir, hash string, data []byte) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	path := CacheFilePath(cacheDir, hash)
	return os.WriteFile(path, data, 0644)
}

// ClearCache 清空缓存目录
func ClearCache(cacheDir string) error {
	if !dirExists(cacheDir) {
		return nil
	}
	return os.RemoveAll(cacheDir)
}
