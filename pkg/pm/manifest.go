package pm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gnuos/jpl/pkg/task"
)

// ============================================================================
// 清单文件 (jpl.json)
// ============================================================================

// Manifest 项目清单文件结构
type Manifest struct {
	Name         string                  `json:"name"`
	Version      string                  `json:"version,omitempty"`
	Description  string                  `json:"description,omitempty"`
	Dependencies map[string]string       `json:"dependencies"`
	Tasks        map[string]task.TaskDef `json:"tasks,omitempty"`
}

// ManifestFileName 清单文件名
const ManifestFileName = "jpl.json"

// LoadManifest 从指定路径加载清单文件
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("manifest not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest %q: %w", path, err)
	}

	if m.Dependencies == nil {
		m.Dependencies = make(map[string]string)
	}

	if m.Tasks == nil {
		m.Tasks = make(map[string]task.TaskDef)
	}

	return &m, nil
}

// LoadOrCreateManifest 加载清单文件，不存在则创建默认清单
func LoadOrCreateManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 创建默认清单
		m := &Manifest{
			Name:         filepath.Base(dir),
			Version:      "0.1.0",
			Dependencies: make(map[string]string),
			Tasks:        make(map[string]task.TaskDef),
		}
		return m, nil
	}
	return LoadManifest(path)
}

// Save 保存清单文件到指定目录
func (m *Manifest) Save(dir string) error {
	path := filepath.Join(dir, ManifestFileName)
	return m.SaveTo(path)
}

// SaveTo 保存清单文件到指定路径
func (m *Manifest) SaveTo(path string) error {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// HasDependency 检查是否已存在指定依赖
func (m *Manifest) HasDependency(name string) bool {
	_, exists := m.Dependencies[name]
	return exists
}

// AddDependency 添加依赖
func (m *Manifest) AddDependency(name, source string) {
	m.Dependencies[name] = source
}

// RemoveDependency 移除依赖
func (m *Manifest) RemoveDependency(name string) bool {
	if _, exists := m.Dependencies[name]; !exists {
		return false
	}
	delete(m.Dependencies, name)
	return true
}

// HasTask 检查是否已存在指定任务
func (m *Manifest) HasTask(name string) bool {
	_, exists := m.Tasks[name]
	return exists
}

// GetTask 获取指定任务定义
func (m *Manifest) GetTask(name string) (task.TaskDef, bool) {
	t, exists := m.Tasks[name]
	return t, exists
}

// ============================================================================
// 依赖源地址解析
// ============================================================================

// SourceInfo 解析后的依赖源信息
type SourceInfo struct {
	URL    string // git 仓库 URL
	Tag    string // 可选的 tag（@v1.0.0）
	Branch string // 可选的分支（#main）
}

// ParseSource 解析依赖源地址
// 支持格式：
//   - https://github.com/user/repo.git
//   - https://github.com/user/repo.git@v1.0.0
//   - https://github.com/user/repo.git#main
//   - ../local-path
func ParseSource(source string) (*SourceInfo, error) {
	if source == "" {
		return nil, fmt.Errorf("empty source")
	}

	info := &SourceInfo{}

	// 本地路径（相对或绝对）
	if source[0] == '.' || source[0] == '/' {
		info.URL = source
		return info, nil
	}

	// 检查是否有 @tag
	if idx := rfind(source, '@'); idx > 0 {
		info.URL = source[:idx]
		info.Tag = source[idx+1:]
		return info, nil
	}

	// 检查是否有 #branch
	if idx := rfind(source, '#'); idx > 0 {
		info.URL = source[:idx]
		info.Branch = source[idx+1:]
		return info, nil
	}

	info.URL = source
	return info, nil
}

// ResolveName 从源地址推断导入名称
// 规则：
//  1. 如果提供了 explicitName，使用它
//  2. 否则从 URL 推断：https://github.com/user/jpl-utils.git → jpl-utils
func ResolveName(source *SourceInfo, explicitName string) string {
	if explicitName != "" {
		return explicitName
	}

	url := source.URL

	// 本地路径
	if url[0] == '.' || url[0] == '/' {
		return filepath.Base(url)
	}

	// 从 URL 提取仓库名
	// 移除协议前缀
	if idx := len(url); idx > 0 {
		// 查找最后一个 /
		lastSlash := -1
		for i := len(url) - 1; i >= 0; i-- {
			if url[i] == '/' {
				lastSlash = i
				break
			}
		}
		if lastSlash >= 0 && lastSlash < len(url)-1 {
			name := url[lastSlash+1:]
			// 移除 .git 后缀
			if len(name) > 4 && name[len(name)-4:] == ".git" {
				name = name[:len(name)-4]
			}
			return name
		}
	}

	return "unknown"
}

// rfind 查找字符串中最后一次出现的字符位置
func rfind(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
