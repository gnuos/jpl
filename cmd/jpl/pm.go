package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/pkg/pm"
	"github.com/spf13/cobra"
)

// ============================================================================
// jpl init - 初始化项目
// ============================================================================

var (
	initName      string
	initDesc      string
	initNoExample bool
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "初始化 JPL 项目",
	Long: `在指定目录中初始化 JPL 项目，创建 jpl.json 清单文件和示例文件。

示例：
  jpl init                      # 在当前目录初始化
  jpl init my-project           # 创建 my-project/ 目录并初始化
  jpl init --name my-app        # 指定项目名称
  jpl init --desc "My app"      # 指定描述
  jpl init --no-example         # 不创建示例文件`,
	Args: cobra.MaximumNArgs(1),
	Run:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "项目名称（默认使用目录名）")
	initCmd.Flags().StringVar(&initDesc, "desc", "", "项目描述")
	initCmd.Flags().BoolVar(&initNoExample, "no-example", false, "不创建示例文件")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	opts := &pm.InitOptions{
		Name:        initName,
		Description: initDesc,
		NoExample:   initNoExample,
	}

	if len(args) > 0 {
		opts.Dir = args[0]
	}

	result, err := pm.InitProject(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("项目已初始化: %s\n", result.Dir)
	fmt.Printf("  清单文件: %s\n", filepath.Base(result.ManifestPath))
	if result.ExamplePath != "" {
		fmt.Printf("  示例文件: %s\n", filepath.Base(result.ExamplePath))
	}
	fmt.Printf("  依赖目录: jpl_modules/\n")
	fmt.Printf("\n使用 'jpl run main.jpl' 运行示例\n")
}

// ============================================================================
// jpl add - 添加依赖
// ============================================================================

var (
	addName    string
	addVersion string
	addNoCache bool
)

var addCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "添加依赖到项目",
	Long: `从 git 仓库或本地路径添加依赖到项目。

支持的源地址格式：
  https://github.com/user/repo.git
  https://github.com/user/repo.git@v1.0.0
  https://github.com/user/repo.git#main
  ../local-path

示例：
  jpl add https://github.com/user/jpl-utils.git
  jpl add https://github.com/user/jpl-http.git@v1.2.0 --name http
  jpl add ../my-lib`,
	Args: cobra.ExactArgs(1),
	Run:  runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addName, "name", "", "自定义导入名称")
	addCmd.Flags().StringVar(&addVersion, "version", "", "指定版本（tag）")
	addCmd.Flags().BoolVar(&addNoCache, "no-cache", false, "禁用全局缓存")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	source := args[0]
	projectDir, _ := os.Getwd()

	// 解析源地址（支持版本约束）
	sourceInfo, err := pm.ParseSourceWithConstraint(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的源地址: %v\n", err)
		os.Exit(1)
	}

	// 如果指定了 --version，覆盖源地址中的约束/tag
	if addVersion != "" {
		if pm.IsVersionConstraint(addVersion) {
			sourceInfo.Constraint = addVersion
		} else {
			sourceInfo.Tag = addVersion
		}
		sourceInfo.Constraint = ""
	}

	// 确定导入名称
	name := pm.ResolveName(&pm.SourceInfo{URL: sourceInfo.URL}, addName)
	if name == "" {
		fmt.Fprintf(os.Stderr, "错误: 无法推断导入名称，请使用 --name 指定\n")
		os.Exit(1)
	}

	// 加载或创建清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 检查依赖是否已存在
	if manifest.HasDependency(name) {
		fmt.Fprintf(os.Stderr, "错误: 依赖 %q 已存在\n", name)
		os.Exit(1)
	}

	// 克隆仓库
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] 克隆仓库: %s\n", sourceInfo.URL)
	}

	var result *pm.CloneResult
	var tmpDir string

	// 根据是否有限制选择克隆方式
	if sourceInfo.Constraint != "" {
		// 使用版本约束克隆
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] 使用版本约束: %s\n", sourceInfo.Constraint)
		}
		result, tmpDir, err = pm.CloneWithConstraint(sourceInfo.URL, sourceInfo.Constraint)
	} else {
		// 直接克隆指定 tag/branch
		result, tmpDir, err = pm.CloneAndCheckout(sourceInfo.URL, sourceInfo.Tag, sourceInfo.Branch)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 克隆失败: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// 安装到 jpl_modules/
	modulesDir := filepath.Join(projectDir, "jpl_modules")
	if err := pm.InstallPackage(tmpDir, modulesDir, name); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 安装失败: %v\n", err)
		os.Exit(1)
	}

	// 尝试安装传递依赖
	transitiveCount := installTransitiveDeps(tmpDir, modulesDir, projectDir)

	// 尝试放入全局缓存
	if !addNoCache {
		cache, err := pm.NewPackageCache()
		if err == nil {
			cache.Put(sourceInfo.URL, result.CommitHash, tmpDir)
		}
	}

	// 更新清单（保留原始源地址，包括版本约束）
	manifest.AddDependency(name, source)
	if err := manifest.Save(projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存清单失败: %v\n", err)
		os.Exit(1)
	}

	// 更新锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, err := engine.LoadLockFile(lockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 加载锁文件失败: %v\n", err)
	} else {
		lf.UpdatePkgEntry(name, source, sourceInfo.URL, result.Version, result.CommitHash, "", nil)
		if err := engine.SaveLockFile(lockPath, lf); err != nil {
			fmt.Fprintf(os.Stderr, "警告: 保存锁文件失败: %v\n", err)
		}
	}

	fmt.Printf("添加依赖: %s (%s @ %s)\n", name, sourceInfo.URL, result.CommitHash[:7])
	if result.Version != "" {
		fmt.Printf("版本: %s\n", result.Version)
	}
	fmt.Printf("安装到: jpl_modules/%s/\n", name)
	if transitiveCount > 0 {
		fmt.Printf("传递依赖: %d 个\n", transitiveCount)
	}
}

// installTransitiveDeps 安装包的传递依赖
func installTransitiveDeps(pkgDir, modulesDir, projectDir string) int {
	// 读取包的清单
	pkgManifestPath := filepath.Join(pkgDir, pm.ManifestFileName)
	if _, err := os.Stat(pkgManifestPath); os.IsNotExist(err) {
		return 0
	}

	pkgManifest, err := pm.LoadManifest(pkgManifestPath)
	if err != nil || len(pkgManifest.Dependencies) == 0 {
		return 0
	}

	// 加载项目清单
	projectManifest, _ := pm.LoadOrCreateManifest(projectDir)

	count := 0
	for depName, depSource := range pkgManifest.Dependencies {
		// 跳过已安装的
		if projectManifest.HasDependency(depName) {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] 传递依赖 %q 已存在，跳过\n", depName)
			}
			continue
		}

		// 跳过已安装在 jpl_modules 中的
		depDir := filepath.Join(modulesDir, depName)
		if _, err := os.Stat(depDir); err == nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] 传递依赖 %q 已安装，跳过\n", depName)
			}
			continue
		}

		// 解析并安装
		depInfo, err := pm.ParseSource(depSource)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  警告: 传递依赖 %q 源地址无效: %v\n", depName, err)
			continue
		}

		result, depTmpDir, err := pm.CloneAndCheckout(depInfo.URL, depInfo.Tag, depInfo.Branch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  警告: 传递依赖 %q 克隆失败: %v\n", depName, err)
			continue
		}

		if err := pm.InstallPackage(depTmpDir, modulesDir, depName); err != nil {
			fmt.Fprintf(os.Stderr, "  警告: 传递依赖 %q 安装失败: %v\n", depName, err)
			os.RemoveAll(depTmpDir)
			continue
		}

		// 更新项目清单（标记为传递依赖来源）
		projectManifest.AddDependency(depName, depSource)
		os.RemoveAll(depTmpDir)

		fmt.Printf("  (传递依赖) %s (%s @ %s)\n", depName, depInfo.URL, result.CommitHash[:7])
		count++
	}

	// 保存更新后的清单
	if count > 0 {
		projectManifest.Save(projectDir)
	}

	return count
}

// ============================================================================
// jpl remove - 移除依赖
// ============================================================================

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "移除项目依赖",
	Long: `从项目中移除指定的依赖。

示例：
  jpl remove utils`,
	Args: cobra.ExactArgs(1),
	Run:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) {
	name := args[0]
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// 检查依赖是否存在
	if !manifest.HasDependency(name) {
		fmt.Fprintf(os.Stderr, "错误: 依赖 %q 不存在\n", name)
		os.Exit(1)
	}

	// 删除 jpl_modules/<name>/ 目录
	modulesDir := filepath.Join(projectDir, "jpl_modules")
	if err := pm.RemovePackage(modulesDir, name); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] 警告: %v\n", err)
		}
	}

	// 从清单中移除
	manifest.RemoveDependency(name)
	if err := manifest.Save(projectDir); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 保存清单失败: %v\n", err)
		os.Exit(1)
	}

	// 从锁文件中移除
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, err := engine.LoadLockFile(lockPath)
	if err == nil {
		lf.RemovePkgEntry(name)
		engine.SaveLockFile(lockPath, lf)
	}

	fmt.Printf("移除依赖: %s\n", name)
	fmt.Printf("已删除: jpl_modules/%s/\n", name)
}

// ============================================================================
// jpl install - 安装全部依赖
// ============================================================================

var (
	installNoCache bool
	installResolve bool
	installJobs    int
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "安装项目全部依赖",
	Long: `根据 jpl.json 清单文件安装项目的所有依赖。

示例：
  jpl install
  jpl install --resolve    # 使用依赖解析器（检测循环依赖和冲突）`,
	Args: cobra.NoArgs,
	Run:  runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&installNoCache, "no-cache", false, "禁用全局缓存")
	installCmd.Flags().BoolVar(&installResolve, "resolve", false, "使用依赖解析器（检测循环依赖和冲突）")
	installCmd.Flags().IntVarP(&installJobs, "jobs", "j", 4, "并行克隆的最大并发数")
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.Dependencies) == 0 {
		fmt.Println("没有需要安装的依赖。")
		return
	}

	modulesDir := filepath.Join(projectDir, "jpl_modules")

	// 如果使用 --resolve 模式，使用 Resolver 进行完整的依赖解析
	if installResolve {
		runInstallWithResolver(manifest, modulesDir, projectDir)
		return
	}

	// 标准安装模式
	runInstallStandard(manifest, modulesDir, projectDir)
}

// runInstallWithResolver 使用 Resolver 进行依赖解析（并行克隆）
func runInstallWithResolver(manifest *pm.Manifest, modulesDir, projectDir string) {
	fmt.Println("解析依赖中...")

	// 创建解析器
	resolver := pm.NewResolver(manifest, modulesDir)

	// 执行解析
	result, err := resolver.Resolve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 依赖解析失败: %v\n", err)
		os.Exit(1)
	}

	// 输出冲突警告
	if len(result.Conflicts) > 0 {
		fmt.Fprintln(os.Stderr, "警告: 检测到版本冲突：")
		for _, conflict := range result.Conflicts {
			fmt.Fprintf(os.Stderr, "  %s: %s vs %s (%s)\n",
				conflict.Name, conflict.Existing, conflict.Incoming, conflict.Resolution)
		}
	}

	// 加载锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, err := engine.LoadLockFile(lockPath)
	if err != nil {
		lf = &engine.LockFile{
			Version:  1,
			Remote:   make(map[string]engine.LockEntry),
			Packages: make(map[string]engine.PkgEntry),
		}
	}

	// 初始化缓存
	var cache *pm.PackageCache
	if !installNoCache {
		cache, _ = pm.NewPackageCache()
	}

	fmt.Println("安装依赖中...")

	// Phase 1: 收集需要克隆的任务，检查已安装和缓存
	type resolvedInstallItem struct {
		name   string
		pkg    *pm.ResolvedPkg
		result *pm.CloneResult
		tmpDir string
	}

	var items []resolvedInstallItem
	var cloneJobs []pm.CloneJob

	for _, name := range result.Order {
		pkg := result.Packages[name]

		// 检查是否已安装
		if _, err := os.Stat(filepath.Join(modulesDir, name)); err == nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] %s: 已安装，跳过\n", name)
			}
			continue
		}

		item := resolvedInstallItem{name: name, pkg: pkg}

		// 尝试从缓存获取
		cacheHit := false
		if cache != nil && pkg.Commit != "" && cache.Has(pkg.Resolved, pkg.Commit) {
			tmpDir, err := os.MkdirTemp("", "jpl-pm-*")
			if err == nil {
				err = cache.Get(pkg.Resolved, pkg.Commit, tmpDir)
				if err == nil {
					item.result = &pm.CloneResult{
						CommitHash: pkg.Commit,
						Version:    pkg.Version,
					}
					item.tmpDir = tmpDir
					cacheHit = true
					if verbose {
						fmt.Fprintf(os.Stderr, "[verbose] %s: 从缓存加载\n", name)
					}
				}
			}
		}

		items = append(items, item)

		// 缓存未命中，加入克隆队列
		if !cacheHit {
			cloneJobs = append(cloneJobs, pm.CloneJob{
				Name:       name,
				URL:        pkg.Resolved,
				Tag:        pkg.Version,
				Branch:     "",
				Constraint: pkg.Constraint,
			})
		}
	}

	// Phase 2: 并行克隆
	if len(cloneJobs) > 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] 并行克隆 %d 个依赖 (并发数: %d)\n", len(cloneJobs), installJobs)
		}

		results := pm.ParallelClone(cloneJobs, installJobs, cache, verbose)

		resultMap := make(map[string]*pm.CloneJobResult)
		for i := range results {
			r := &results[i]
			resultMap[r.Name] = r
		}

		for i := range items {
			if items[i].result == nil {
				if r, ok := resultMap[items[i].name]; ok {
					if r.Err != nil {
						fmt.Fprintf(os.Stderr, "  %s: 克隆失败: %v\n", items[i].name, r.Err)
						continue
					}
					items[i].result = r.Result
					items[i].tmpDir = r.TmpDir

					// 放入缓存
					if cache != nil {
						cache.Put(items[i].pkg.Resolved, r.Result.CommitHash, r.TmpDir)
					}
				}
			}
		}
	}

	// Phase 3: 按拓扑顺序串行安装
	installed := 0
	for _, item := range items {
		if item.result == nil {
			continue
		}

		// 安装
		if err := pm.InstallPackage(item.tmpDir, modulesDir, item.name); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 安装失败: %v\n", item.name, err)
			os.RemoveAll(item.tmpDir)
			continue
		}
		os.RemoveAll(item.tmpDir)

		// 更新锁文件
		lf.UpdatePkgEntry(item.name, item.pkg.Source, item.pkg.Resolved, item.result.Version, item.result.CommitHash, "", item.pkg.Dependencies)

		transitiveStr := ""
		if item.pkg.IsTransitive {
			transitiveStr = " (传递依赖)"
		}
		fmt.Printf("  %s (%s @ %s)%s - OK\n", item.name, item.pkg.Resolved, item.result.CommitHash[:7], transitiveStr)
		installed++
	}

	// 保存锁文件
	engine.SaveLockFile(lockPath, lf)

	fmt.Printf("\n%d 个依赖已安装。\n", installed)
}

// runInstallStandard 标准安装模式（并行克隆）
func runInstallStandard(manifest *pm.Manifest, modulesDir, projectDir string) {
	// 加载锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, err := engine.LoadLockFile(lockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 加载锁文件失败: %v\n", err)
		lf = &engine.LockFile{
			Version:  1,
			Remote:   make(map[string]engine.LockEntry),
			Packages: make(map[string]engine.PkgEntry),
		}
	}

	// 初始化缓存
	var cache *pm.PackageCache
	if !installNoCache {
		cache, err = pm.NewPackageCache()
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] 警告: 无法初始化缓存: %v\n", err)
			}
		}
	}

	fmt.Println("安装依赖中...")

	// Phase 1: 收集需要克隆的任务，检查缓存命中
	type installItem struct {
		name       string
		source     string
		sourceInfo *pm.SourceWithVersion
		result     *pm.CloneResult
		tmpDir     string
	}

	var items []installItem
	var cloneJobs []pm.CloneJob

	for name, source := range manifest.Dependencies {
		// 解析源地址（支持版本约束）
		sourceInfo, err := pm.ParseSourceWithConstraint(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 无效的源地址: %v\n", name, err)
			continue
		}

		item := installItem{name: name, source: source, sourceInfo: sourceInfo}

		// 检查锁文件中是否有记录
		lockedCommit := ""
		if entry, ok := lf.GetPkgEntry(name); ok {
			lockedCommit = entry.Commit
		}

		// 尝试从缓存获取
		cacheHit := false
		if cache != nil && lockedCommit != "" && cache.Has(sourceInfo.URL, lockedCommit) {
			tmpDir, err := os.MkdirTemp("", "jpl-pm-*")
			if err == nil {
				err = cache.Get(sourceInfo.URL, lockedCommit, tmpDir)
				if err == nil {
					item.result = &pm.CloneResult{CommitHash: lockedCommit}
					item.tmpDir = tmpDir
					cacheHit = true
					if verbose {
						fmt.Fprintf(os.Stderr, "[verbose] %s: 从缓存加载\n", name)
					}
				}
			}
		}

		items = append(items, item)

		// 缓存未命中，加入克隆队列
		if !cacheHit {
			cloneJobs = append(cloneJobs, pm.CloneJob{
				Name:       name,
				URL:        sourceInfo.URL,
				Tag:        sourceInfo.Tag,
				Branch:     sourceInfo.Branch,
				Constraint: sourceInfo.Constraint,
			})
		}
	}

	// Phase 2: 并行克隆
	if len(cloneJobs) > 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] 并行克隆 %d 个依赖 (并发数: %d)\n", len(cloneJobs), installJobs)
		}

		results := pm.ParallelClone(cloneJobs, installJobs, cache, verbose)

		// 将克隆结果映射回 installItem
		resultMap := make(map[string]*pm.CloneJobResult)
		for i := range results {
			r := &results[i]
			resultMap[r.Name] = r
		}

		for i := range items {
			if items[i].result == nil {
				if r, ok := resultMap[items[i].name]; ok {
					if r.Err != nil {
						fmt.Fprintf(os.Stderr, "  %s: 克隆失败: %v\n", items[i].name, r.Err)
						continue
					}
					items[i].result = r.Result
					items[i].tmpDir = r.TmpDir

					// 放入缓存
					if cache != nil {
						cache.Put(items[i].sourceInfo.URL, r.Result.CommitHash, r.TmpDir)
					}
				}
			}
		}
	}

	// Phase 3: 串行安装
	installed := 0
	for _, item := range items {
		if item.result == nil {
			continue
		}

		// 安装
		if err := pm.InstallPackage(item.tmpDir, modulesDir, item.name); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 安装失败: %v\n", item.name, err)
			os.RemoveAll(item.tmpDir)
			continue
		}
		os.RemoveAll(item.tmpDir)

		// 更新锁文件
		lf.UpdatePkgEntry(item.name, item.source, item.sourceInfo.URL, item.result.Version, item.result.CommitHash, "", nil)

		versionStr := ""
		if item.result.Version != "" {
			versionStr = " " + item.result.Version
		}
		fmt.Printf("  %s (%s @ %s%s) - OK\n", item.name, item.sourceInfo.URL, item.result.CommitHash[:7], versionStr)
		installed++
	}

	// 保存锁文件
	if err := engine.SaveLockFile(lockPath, lf); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 保存锁文件失败: %v\n", err)
	}

	fmt.Printf("\n%d 个依赖已安装。\n", installed)
}

// ============================================================================
// jpl list - 列出依赖
// ============================================================================

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出项目依赖",
	Long: `列出项目已安装的依赖及其版本信息。

示例：
  jpl list`,
	Args: cobra.NoArgs,
	Run:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.Dependencies) == 0 {
		fmt.Printf("%s@%s\n", manifest.Name, manifest.Version)
		fmt.Println("└── (没有依赖)")
		return
	}

	// 加载锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, _ := engine.LoadLockFile(lockPath)

	// 列出依赖
	packages, err := pm.ListInstalled(manifest, filepath.Join(projectDir, "jpl_modules"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: %v\n", err)
	}

	fmt.Printf("%s@%s\n", manifest.Name, manifest.Version)

	for i, pkg := range packages {
		isLast := i == len(packages)-1
		prefix := "├── "
		if isLast {
			prefix = "└── "
		}

		versionStr := ""
		if pkg.Version != "" {
			versionStr = "@" + pkg.Version
		}

		// 从锁文件获取 commit
		commitStr := ""
		if lf != nil {
			if entry, ok := lf.GetPkgEntry(pkg.Name); ok {
				commitStr = " (" + entry.Commit[:7] + ")"
			}
		}

		fmt.Printf("%s%s%s%s\n", prefix, pkg.Name, versionStr, commitStr)
	}

	fmt.Printf("\n(%d 个依赖)\n", len(packages))
}

// ============================================================================
// jpl update - 更新依赖
// ============================================================================

var (
	updateAll  bool
	updateName string
)

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "更新依赖到最新版本",
	Long: `更新项目依赖到最新版本。

示例：
  jpl update            # 更新所有依赖
  jpl update utils      # 更新指定依赖
  jpl update --all      # 更新所有依赖（显式）`,
	Args: cobra.MaximumNArgs(1),
	Run:  runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateAll, "all", false, "更新所有依赖")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) {
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.Dependencies) == 0 {
		fmt.Println("没有需要更新的依赖。")
		return
	}

	// 加载锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, err := engine.LoadLockFile(lockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "警告: 加载锁文件失败: %v\n", err)
		lf = &engine.LockFile{
			Version:  1,
			Remote:   make(map[string]engine.LockEntry),
			Packages: make(map[string]engine.PkgEntry),
		}
	}

	// 确定要更新的依赖
	var toUpdate []string
	if len(args) > 0 {
		name := args[0]
		if !manifest.HasDependency(name) {
			fmt.Fprintf(os.Stderr, "错误: 依赖 %q 不存在\n", name)
			os.Exit(1)
		}
		toUpdate = []string{name}
	} else {
		for name := range manifest.Dependencies {
			toUpdate = append(toUpdate, name)
		}
	}

	fmt.Println("检查更新中...")

	modulesDir := filepath.Join(projectDir, "jpl_modules")
	updated := 0

	for _, name := range toUpdate {
		source := manifest.Dependencies[name]

		// 解析源地址
		sourceInfo, err := pm.ParseSourceWithConstraint(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 无效的源地址: %v\n", name, err)
			continue
		}

		// 获取当前锁文件中的 commit
		oldCommit := ""
		if entry, ok := lf.GetPkgEntry(name); ok {
			oldCommit = entry.Commit
		}

		// 克隆最新版本
		var result *pm.CloneResult
		var tmpDir string

		if sourceInfo.Constraint != "" {
			result, tmpDir, err = pm.CloneWithConstraint(sourceInfo.URL, sourceInfo.Constraint)
		} else {
			result, tmpDir, err = pm.CloneAndCheckout(sourceInfo.URL, sourceInfo.Tag, sourceInfo.Branch)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 检查更新失败: %v\n", name, err)
			continue
		}

		// 检查是否有更新
		if result.CommitHash == oldCommit {
			fmt.Printf("  %s: 已是最新 (%s)\n", name, result.CommitHash[:7])
			os.RemoveAll(tmpDir)
			continue
		}

		// 有更新，重新安装
		if err := pm.InstallPackage(tmpDir, modulesDir, name); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: 安装失败: %v\n", name, err)
			os.RemoveAll(tmpDir)
			continue
		}
		os.RemoveAll(tmpDir)

		// 更新锁文件
		lf.UpdatePkgEntry(name, source, sourceInfo.URL, result.Version, result.CommitHash, "", nil)

		fmt.Printf("  %s: %s → %s\n", name, oldCommit[:7], result.CommitHash[:7])
		updated++
	}

	// 保存锁文件
	if err := engine.SaveLockFile(lockPath, lf); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 保存锁文件失败: %v\n", err)
	}

	if updated > 0 {
		fmt.Printf("\n%d 个依赖已更新。\n", updated)
	} else {
		fmt.Println("\n所有依赖已是最新。")
	}
}

// ============================================================================
// jpl outdated - 检查过时的依赖
// ============================================================================

var outdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "检查过时的依赖",
	Long: `检查项目依赖是否有新版本可用。

示例：
  jpl outdated`,
	Args: cobra.NoArgs,
	Run:  runOutdated,
}

func init() {
	rootCmd.AddCommand(outdatedCmd)
}

func runOutdated(cmd *cobra.Command, args []string) {
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if len(manifest.Dependencies) == 0 {
		fmt.Println("没有依赖。")
		return
	}

	// 加载锁文件
	lockPath := filepath.Join(projectDir, "jpl.lock.yaml")
	lf, _ := engine.LoadLockFile(lockPath)

	fmt.Println("检查过时的依赖...")

	type outdatedInfo struct {
		Name     string
		Current  string
		Latest   string
		UpToDate bool
	}

	var outdated []outdatedInfo

	for name, source := range manifest.Dependencies {
		// 解析源地址
		sourceInfo, err := pm.ParseSourceWithConstraint(source)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] %s: 无效的源地址: %v\n", name, err)
			}
			continue
		}

		// 获取当前 commit
		currentCommit := ""
		if lf != nil {
			if entry, ok := lf.GetPkgEntry(name); ok {
				currentCommit = entry.Commit
			}
		}

		// 获取最新 commit
		var result *pm.CloneResult
		var tmpDir string

		if sourceInfo.Constraint != "" {
			result, tmpDir, err = pm.CloneWithConstraint(sourceInfo.URL, sourceInfo.Constraint)
		} else {
			result, tmpDir, err = pm.CloneAndCheckout(sourceInfo.URL, sourceInfo.Tag, sourceInfo.Branch)
		}

		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[verbose] %s: 获取最新版本失败: %v\n", name, err)
			}
			continue
		}
		os.RemoveAll(tmpDir)

		info := outdatedInfo{
			Name:    name,
			Current: currentCommit,
			Latest:  result.CommitHash,
		}

		if currentCommit == result.CommitHash {
			info.UpToDate = true
		}

		outdated = append(outdated, info)
	}

	// 输出结果
	hasOutdated := false
	for _, info := range outdated {
		if !info.UpToDate {
			hasOutdated = true
			fmt.Printf("  %s: %s → %s\n", info.Name, info.Current[:7], info.Latest[:7])
		} else if verbose {
			fmt.Printf("  %s: 已是最新 (%s)\n", info.Name, info.Current[:7])
		}
	}

	if !hasOutdated {
		fmt.Println("\n所有依赖已是最新。")
	} else {
		fmt.Printf("\n使用 'jpl update' 更新过时的依赖。\n")
	}
}
