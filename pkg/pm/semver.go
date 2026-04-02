package pm

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ============================================================================
// 语义化版本（基于 Masterminds/semver）
// ============================================================================

// Version 封装 semver.Version
type Version = semver.Version

// Constraints 封装 semver.Constraints
type Constraints = semver.Constraints

// ParseVersion 解析版本字符串
// 支持格式: "1.2.3", "v1.2.3", "1.2.3-alpha.1", "1.2.3+build"
func ParseVersion(s string) (*Version, error) {
	// 移除 v 前缀
	s = strings.TrimPrefix(s, "v")
	return semver.NewVersion(s)
}

// MustParseVersion 解析版本字符串（ panic 版本，仅用于测试）
func MustParseVersion(s string) *Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

// ParseConstraint 解析版本约束字符串
// 支持格式: "^1.2.3", "~1.2.3", ">=1.2.3", "1.2.3", "*"
func ParseConstraint(s string) (*Constraints, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" {
		s = ">=0.0.0"
	}
	return semver.NewConstraint(s)
}

// VersionSatisfies 检查版本是否满足约束
func VersionSatisfies(version *Version, constraint *Constraints) bool {
	return constraint.Check(version)
}

// CompareVersions 比较两个版本
// 返回: -1 (a < b), 0 (a == b), 1 (a > b)
func CompareVersions(a, b *Version) int {
	return a.Compare(b)
}

// ============================================================================
// 源地址版本约束解析
// ============================================================================

// SourceWithVersion 带版本约束的源地址
type SourceWithVersion struct {
	URL        string // 仓库 URL
	Constraint string // 版本约束（如 "^1.2.0"）
	Tag        string // 精确 tag（向后兼容）
	Branch     string // 分支（向后兼容）
}

// ParseSourceWithConstraint 解析带版本约束的源地址
// 支持格式：
//   - https://github.com/user/repo.git@^1.2.0
//   - https://github.com/user/repo.git@~2.0.0
//   - https://github.com/user/repo.git@>=1.0.0
//   - https://github.com/user/repo.git@v1.0.0（精确 tag，向后兼容）
//   - https://github.com/user/repo.git#main（分支，向后兼容）
func ParseSourceWithConstraint(source string) (*SourceWithVersion, error) {
	if source == "" {
		return nil, fmt.Errorf("empty source")
	}

	info := &SourceWithVersion{}

	// 本地路径
	if source[0] == '.' || source[0] == '/' {
		info.URL = source
		return info, nil
	}

	// 检查是否有 @version
	if idx := rfind(source, '@'); idx > 0 {
		info.URL = source[:idx]
		versionPart := source[idx+1:]

		// 判断是版本约束还是精确 tag
		if isVersionConstraint(versionPart) {
			info.Constraint = versionPart
		} else {
			info.Tag = versionPart
		}
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

// isVersionConstraint 检查字符串是否为版本约束
func isVersionConstraint(s string) bool {
	// 版本约束通常以 ^, ~, >=, >, <, <=, = 开头
	if len(s) == 0 {
		return false
	}

	// 检查运算符前缀
	switch s[0] {
	case '^', '~', '>', '<', '=':
		return true
	}

	// 检查是否为 x-range（如 "1.x", "1.2.x"）
	if strings.Contains(s, "x") || strings.Contains(s, "*") {
		return true
	}

	// 检查是否为 hyphen range（如 "1.0.0 - 2.0.0"）
	if strings.Contains(s, " - ") {
		return true
	}

	return false
}

// IsVersionConstraint 检查字符串是否为版本约束（导出版）
func IsVersionConstraint(s string) bool {
	return isVersionConstraint(s)
}

// SelectBestVersion 从可用版本列表中选择满足约束的最佳版本
func SelectBestVersion(versions []string, constraintStr string) (string, error) {
	if constraintStr == "" {
		// 无约束，返回最新版本
		if len(versions) == 0 {
			return "", fmt.Errorf("no versions available")
		}
		return versions[len(versions)-1], nil
	}

	constraint, err := ParseConstraint(constraintStr)
	if err != nil {
		return "", fmt.Errorf("invalid constraint %q: %w", constraintStr, err)
	}

	// 解析所有版本并排序
	var parsed []*Version
	for _, v := range versions {
		ver, err := ParseVersion(v)
		if err != nil {
			continue // 跳过无效版本
		}
		parsed = append(parsed, ver)
	}

	// 从最新版本开始查找满足约束的版本
	for i := len(parsed) - 1; i >= 0; i-- {
		if constraint.Check(parsed[i]) {
			return "v" + parsed[i].String(), nil
		}
	}

	return "", fmt.Errorf("no version satisfies constraint %q", constraintStr)
}
