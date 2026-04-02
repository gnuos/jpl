package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// 文件模块加载器
// ============================================================================

// FileModuleLoader 文件系统模块加载器
// 搜索顺序：
//  1. 绝对路径 → 直接加载
//  2. URL（http/https） → 远程加载（带磁盘缓存和锁文件校验）
//  3. 相对路径 → 基于 baseDir 查找
//  4. 裸名称 → 搜索路径查找（scriptDir → jpl_modules/ → ~/.jpl/modules/）
type FileModuleLoader struct {
	baseDir      string   // 基准目录（当前脚本所在目录）
	searchPaths  []string // 额外搜索路径
	cacheDir     string   // URL 缓存目录（默认 ~/.jpl/cache/）
	lockFile     *LockFile
	lockFilePath string // 锁文件路径
	frozen       bool   // frozen 模式：锁文件 hash 不匹配则报错
}

// NewFileModuleLoader 创建文件模块加载器
func NewFileModuleLoader(baseDir string) *FileModuleLoader {
	cacheDir, _ := DefaultCacheDir()
	return &FileModuleLoader{
		baseDir:  baseDir,
		cacheDir: cacheDir,
	}
}

// SetLockFile 设置锁文件路径
func (l *FileModuleLoader) SetLockFile(path string) error {
	l.lockFilePath = path
	lf, err := LoadLockFile(path)
	if err != nil {
		return err
	}
	l.lockFile = lf
	return nil
}

// SetFrozen 设置 frozen 模式
func (l *FileModuleLoader) SetFrozen(frozen bool) {
	l.frozen = frozen
}

// SetCacheDir 设置缓存目录
func (l *FileModuleLoader) SetCacheDir(dir string) {
	l.cacheDir = dir
}

// SaveLockFile 保存锁文件（如有变更）
func (l *FileModuleLoader) SaveLockFile() error {
	if l.lockFile == nil || l.lockFilePath == "" {
		return nil
	}
	return SaveLockFile(l.lockFilePath, l.lockFile)
}

// AddSearchPath 添加搜索路径
func (l *FileModuleLoader) AddSearchPath(path string) {
	l.searchPaths = append(l.searchPaths, path)
}

// LoadModule 加载模块
func (l *FileModuleLoader) LoadModule(source string, engine *Engine) (*ModuleCache, error) {
	// URL 加载
	if isURL(source) {
		return l.loadFromURL(source, engine)
	}

	// 解析文件路径
	resolvedPath, err := l.resolvePath(source)
	if err != nil {
		return nil, err
	}

	return l.loadFromFile(resolvedPath, engine)
}

// resolvePath 解析模块路径
func (l *FileModuleLoader) resolvePath(source string) (string, error) {
	// 1. 绝对路径
	if filepath.IsAbs(source) {
		if fileExists(source) {
			return source, nil
		}
		return "", fmt.Errorf("module not found: %s", source)
	}

	// 2. 相对路径（含 /）→ 基于 baseDir
	if strings.Contains(source, "/") || strings.Contains(source, "\\") {
		candidates := []string{}
		if l.baseDir != "" {
			candidates = append(candidates, filepath.Join(l.baseDir, source))
		}
		// 添加 .jpl 后缀
		if !strings.HasSuffix(source, ".jpl") {
			for i, c := range candidates {
				candidates[i] = c + ".jpl"
			}
		}
		for _, path := range candidates {
			if fileExists(path) {
				return path, nil
			}
		}
		return "", fmt.Errorf("module not found: %s (searched relative to %s)", source, l.baseDir)
	}

	// 3. 裸名称 → 搜索路径查找
	return l.searchForModule(source)
}

// searchForModule 在搜索路径中查找模块
func (l *FileModuleLoader) searchForModule(name string) (string, error) {
	// 构建文件名（自动添加 .jpl 后缀）
	filename := name
	if !strings.HasSuffix(filename, ".jpl") {
		filename = name + ".jpl"
	}

	// 搜索路径列表
	var searchDirs []string

	// 1. 脚本同目录
	if l.baseDir != "" {
		searchDirs = append(searchDirs, l.baseDir)
	}

	// 2. 额外搜索路径
	searchDirs = append(searchDirs, l.searchPaths...)

	// 3. 项目根目录 jpl_modules/
	if l.baseDir != "" {
		// 向上查找包含 jpl_modules 的目录
		dir := l.baseDir
		for {
			modDir := filepath.Join(dir, "jpl_modules")
			if dirExists(modDir) {
				searchDirs = append(searchDirs, modDir)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// 4. 用户目录 ~/.jpl/modules/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userModDir := filepath.Join(homeDir, ".jpl", "modules")
		searchDirs = append(searchDirs, userModDir)
	}

	// 搜索
	for _, dir := range searchDirs {
		path := filepath.Join(dir, filename)
		if fileExists(path) {
			return path, nil
		}
		// 也尝试子目录/index.jpl
		indexPath := filepath.Join(dir, name, "index.jpl")
		if fileExists(indexPath) {
			return indexPath, nil
		}
	}

	return "", fmt.Errorf("module %q not found in search paths", name)
}

// loadFromFile 从文件加载模块
func (l *FileModuleLoader) loadFromFile(path string, engine *Engine) (*ModuleCache, error) {
	script, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %q: %w", path, err)
	}

	// 创建新的加载器（基于模块所在目录）
	absPath, _ := filepath.Abs(path)
	moduleDir := filepath.Dir(absPath)
	subLoader := NewFileModuleLoader(moduleDir)
	subLoader.searchPaths = l.searchPaths
	subLoader.cacheDir = l.cacheDir
	subLoader.lockFile = l.lockFile
	subLoader.lockFilePath = l.lockFilePath
	subLoader.frozen = l.frozen

	// 临时设置加载器
	oldLoader := engine.moduleLoader
	engine.SetModuleLoader(subLoader)
	defer func() {
		engine.SetModuleLoader(oldLoader)
		// 传播锁文件变更
		l.lockFile = subLoader.lockFile
	}()

	// 编译并执行
	exports, err := engine.RunScriptString(string(script), absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to execute module %q: %w", path, err)
	}

	return &ModuleCache{Exports: exports}, nil
}

// loadFromURL 从 URL 加载模块（带磁盘缓存和锁文件校验）
func (l *FileModuleLoader) loadFromURL(url string, engine *Engine) (*ModuleCache, error) {
	var body []byte
	var hash string
	var fromCache bool

	// 1. 尝试内存缓存（Engine 模块缓存已在调用前检查，这里跳过）

	// 2. 尝试磁盘缓存
	if l.cacheDir != "" {
		// 如果有锁文件记录，用其 hash 查缓存
		if l.lockFile != nil {
			if entry, ok := l.lockFile.Remote[url]; ok {
				cached, err := ReadCache(l.cacheDir, entry.Hash)
				if err == nil {
					body = cached
					hash = entry.Hash
					fromCache = true
				}
			}
		}
	}

	// 3. 缓存未命中，从网络下载
	if body == nil {
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			// 网络失败时，如果有磁盘缓存可降级使用
			return nil, fmt.Errorf("failed to fetch module from %q: %w", url, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch module from %q: HTTP %d", url, resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read module response from %q: %w", url, err)
		}

		hash = HashContent(body)
	}

	// 4. 锁文件校验
	if l.lockFile != nil {
		if err := l.lockFile.VerifyHash(url, hash, l.frozen); err != nil {
			return nil, err
		}
	}

	// 5. 写入磁盘缓存（新下载的内容）
	if !fromCache && l.cacheDir != "" {
		if err := WriteCache(l.cacheDir, hash, body); err != nil {
			// 缓存写入失败不影响加载，仅记录
		}
	}

	// 6. 更新锁文件
	if l.lockFile != nil {
		l.lockFile.UpdateEntry(url, hash, int64(len(body)))
	}

	// 7. 创建子加载器（URL 模块中的相对路径基于 URL 目录解析）
	subLoader := NewFileModuleLoader(l.baseDir)
	subLoader.searchPaths = l.searchPaths
	subLoader.cacheDir = l.cacheDir
	subLoader.lockFile = l.lockFile
	subLoader.lockFilePath = l.lockFilePath
	subLoader.frozen = l.frozen

	oldLoader := engine.moduleLoader
	engine.SetModuleLoader(subLoader)
	defer func() {
		engine.SetModuleLoader(oldLoader)
		l.lockFile = subLoader.lockFile
	}()

	exports, err := engine.RunScriptString(string(body), url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute module from %q: %w", url, err)
	}

	return &ModuleCache{Exports: exports}, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// isURL 检查是否为 URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists 检查目录是否存在
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
