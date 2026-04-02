package pm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// 依赖解析器
// ============================================================================

// 三色标记（用于循环检测）
const (
	colorWhite = 0 // 未访问
	colorGray  = 1 // 正在访问（祖先节点）
	colorBlack = 2 // 已完成访问
)

// ResolvedPkg 已解析的包
type ResolvedPkg struct {
	Name         string   // 包名
	Source       string   // 源地址
	Resolved     string   // 解析后的 URL
	Constraint   string   // 版本约束（如 "^1.2.0"）
	Version      string   // 实际版本（tag/branch）
	Commit       string   // commit hash
	Dependencies []string // 传递依赖列表
	IsTransitive bool     // 是否为传递依赖
}

// ResolveResult 依赖解析结果
type ResolveResult struct {
	Packages  map[string]*ResolvedPkg // 已解析的包映射
	Conflicts []VersionConflict       // 版本冲突警告
	Order     []string                // 安装顺序（拓扑排序）
}

// VersionConflict 版本冲突
type VersionConflict struct {
	Name       string // 包名
	Existing   string // 已存在的源
	Incoming   string // 新来的源
	Resolution string // 解决方式
}

// Resolver 依赖解析器
type Resolver struct {
	manifest   *Manifest        // 项目清单
	modulesDir string           // jpl_modules 目录
	lockFile   *LockFileWrapper // 锁文件包装

	// 解析状态
	colors    map[string]int    // 节点颜色
	parents   map[string]string // 父节点（用于输出循环路径）
	resolved  map[string]*ResolvedPkg
	conflicts []VersionConflict
	order     []string
}

// LockFileWrapper 锁文件包装（避免循环导入 engine 包）
type LockFileWrapper struct {
	Entries map[string]LockFileEntry
}

// LockFileEntry 锁文件条目
type LockFileEntry struct {
	Source   string
	Version  string
	Commit   string
	Resolved string
}

// NewResolver 创建依赖解析器
func NewResolver(manifest *Manifest, modulesDir string) *Resolver {
	return &Resolver{
		manifest:   manifest,
		modulesDir: modulesDir,
		colors:     make(map[string]int),
		parents:    make(map[string]string),
		resolved:   make(map[string]*ResolvedPkg),
	}
}

// SetLockFile 设置锁文件
func (r *Resolver) SetLockFile(lf *LockFileWrapper) {
	r.lockFile = lf
}

// Resolve 解析所有依赖（包括传递依赖）
func (r *Resolver) Resolve() (*ResolveResult, error) {
	// 先解析顶层依赖
	for name, source := range r.manifest.Dependencies {
		if err := r.resolvePackage(name, source, false, ""); err != nil {
			return nil, err
		}
	}

	// 构建拓扑排序
	r.buildOrder()

	return &ResolveResult{
		Packages:  r.resolved,
		Conflicts: r.conflicts,
		Order:     r.order,
	}, nil
}

// resolvePackage 解析单个包及其传递依赖
func (r *Resolver) resolvePackage(name, source string, isTransitive bool, fromPkg string) error {
	// 检查是否已解析
	if _, exists := r.resolved[name]; exists {
		// 检查版本冲突
		return r.checkConflict(name, source, fromPkg)
	}

	// 解析源地址（支持版本约束）
	sourceInfo, err := ParseSourceWithConstraint(source)
	if err != nil {
		return fmt.Errorf("package %q: invalid source: %w", name, err)
	}

	// 三色标记：标记为灰色（正在访问）
	r.colors[name] = colorGray
	r.parents[name] = fromPkg

	// 克隆仓库获取传递依赖
	var result *CloneResult
	var tmpDir string

	if sourceInfo.Constraint != "" {
		// 使用版本约束克隆
		result, tmpDir, err = CloneWithConstraint(sourceInfo.URL, sourceInfo.Constraint)
	} else {
		// 直接克隆指定 tag/branch
		result, tmpDir, err = CloneAndCheckout(sourceInfo.URL, sourceInfo.Tag, sourceInfo.Branch)
	}

	if err != nil {
		r.colors[name] = colorWhite // 回退状态
		return fmt.Errorf("package %q: clone failed: %w", name, err)
	}
	defer os.RemoveAll(tmpDir)

	// 读取包的清单文件（获取传递依赖）
	var transitiveDeps []string
	pkgManifestPath := filepath.Join(tmpDir, ManifestFileName)
	if _, err := os.Stat(pkgManifestPath); err == nil {
		pkgManifest, err := LoadManifest(pkgManifestPath)
		if err == nil && len(pkgManifest.Dependencies) > 0 {
			transitiveDeps = make([]string, 0, len(pkgManifest.Dependencies))
			for depName, depSource := range pkgManifest.Dependencies {
				transitiveDeps = append(transitiveDeps, depName)

				// 检查循环依赖
				if r.colors[depName] == colorGray {
					// 检测到循环依赖
					cycle := r.buildCyclePath(name, depName)
					return fmt.Errorf("circular dependency detected:\n%s", cycle)
				}

				// 递归解析传递依赖
				if err := r.resolvePackage(depName, depSource, true, name); err != nil {
					return err
				}
			}
		}
	}

	// 记录解析结果
	r.resolved[name] = &ResolvedPkg{
		Name:         name,
		Source:       source,
		Resolved:     sourceInfo.URL,
		Constraint:   sourceInfo.Constraint,
		Version:      result.Version,
		Commit:       result.CommitHash,
		Dependencies: transitiveDeps,
		IsTransitive: isTransitive,
	}

	// 三色标记：标记为黑色（已完成）
	r.colors[name] = colorBlack

	return nil
}

// checkConflict 检查版本冲突
func (r *Resolver) checkConflict(name, source, fromPkg string) error {
	existing := r.resolved[name]

	// 解析新源地址（支持版本约束）
	newInfo, err := ParseSourceWithConstraint(source)
	if err != nil {
		return nil // 解析失败，跳过冲突检查
	}

	// URL 不同则冲突
	if existing.Resolved != newInfo.URL {
		conflict := VersionConflict{
			Name:       name,
			Existing:   existing.Source,
			Incoming:   source,
			Resolution: "using existing",
		}
		r.conflicts = append(r.conflicts, conflict)
		return nil
	}

	// URL 相同，检查版本约束是否兼容
	if existing.Constraint != "" && newInfo.Constraint != "" {
		// 验证约束可解析
		_, err1 := ParseConstraint(existing.Constraint)
		_, err2 := ParseConstraint(newInfo.Constraint)

		if err1 == nil && err2 == nil {
			// 约束不同则记录警告
			if existing.Constraint != newInfo.Constraint {
				conflict := VersionConflict{
					Name:       name,
					Existing:   existing.Constraint,
					Incoming:   newInfo.Constraint,
					Resolution: "constraints differ",
				}
				r.conflicts = append(r.conflicts, conflict)
			}
		}
	}

	return nil
}

// buildCyclePath 构建循环依赖路径
func (r *Resolver) buildCyclePath(current, target string) string {
	// 从 current 回溯到 target
	path := []string{current}
	node := current
	for node != target && node != "" {
		node = r.parents[node]
		if node != "" {
			path = append(path, node)
		}
	}
	path = append(path, current) // 闭合循环

	// 反转路径
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return "  " + strings.Join(path, " → ")
}

// buildOrder 构建拓扑排序（安装顺序）
func (r *Resolver) buildOrder() {
	visited := make(map[string]bool)
	order := []string{}

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		pkg, exists := r.resolved[name]
		if !exists {
			return
		}

		// 先访问依赖
		for _, dep := range pkg.Dependencies {
			visit(dep)
		}

		order = append(order, name)
	}

	for name := range r.resolved {
		visit(name)
	}

	r.order = order
}

// ============================================================================
// 已安装包查询
// ============================================================================

// InstalledPackage 已安装的包信息
type InstalledPackage struct {
	Name    string
	Version string
	Source  string
}

// ListInstalled 列出已安装的包
func ListInstalled(manifest *Manifest, modulesDir string) ([]InstalledPackage, error) {
	var packages []InstalledPackage

	for name, source := range manifest.Dependencies {
		pkgDir := filepath.Join(modulesDir, name)
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			continue
		}

		pkg := InstalledPackage{
			Name:   name,
			Source: source,
		}

		// 尝试读取包的清单获取版本
		pkgManifestPath := filepath.Join(pkgDir, ManifestFileName)
		if pkgManifest, err := LoadManifest(pkgManifestPath); err == nil {
			pkg.Version = pkgManifest.Version
		}

		packages = append(packages, pkg)
	}

	return packages, nil
}
