package pm

import (
	"fmt"
	"os"
	"path/filepath"
)

// ============================================================================
// 项目初始化
// ============================================================================

// InitOptions 初始化选项
type InitOptions struct {
	Dir         string // 目标目录（空则使用当前目录）
	Name        string // 项目名称（空则使用目录名）
	Description string // 项目描述
	NoExample   bool   // 不创建示例文件
}

// InitResult 初始化结果
type InitResult struct {
	Dir          string // 项目目录
	ManifestPath string // 清单文件路径
	ExamplePath  string // 示例文件路径
}

// InitProject 初始化项目
func InitProject(opts *InitOptions) (*InitResult, error) {
	// 确定目标目录
	dir := opts.Dir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// 如果目录不存在，创建它
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 确定项目名称
	name := opts.Name
	if name == "" {
		name = filepath.Base(dir)
	}

	// 检查是否已存在 jpl.json
	manifestPath := filepath.Join(dir, ManifestFileName)
	if _, err := os.Stat(manifestPath); err == nil {
		return nil, fmt.Errorf("jpl.json already exists in %s", dir)
	}

	// 创建清单文件
	manifest := &Manifest{
		Name:         name,
		Version:      "0.1.0",
		Description:  opts.Description,
		Dependencies: make(map[string]string),
	}

	if err := manifest.Save(dir); err != nil {
		return nil, fmt.Errorf("failed to create manifest: %w", err)
	}

	// 创建 jpl_modules/ 目录
	modulesDir := filepath.Join(dir, "jpl_modules")
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create modules directory: %w", err)
	}

	// 创建示例文件
	var examplePath string
	if !opts.NoExample {
		examplePath = filepath.Join(dir, "main.jpl")
		exampleContent := generateExampleContent(name)
		if err := os.WriteFile(examplePath, []byte(exampleContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to create example file: %w", err)
		}
	}

	return &InitResult{
		Dir:          dir,
		ManifestPath: manifestPath,
		ExamplePath:  examplePath,
	}, nil
}

// generateExampleContent 生成示例文件内容
func generateExampleContent(projectName string) string {
	return fmt.Sprintf(`// %s - 项目入口文件

puts "Hello from %s!"
`, projectName, projectName)
}
