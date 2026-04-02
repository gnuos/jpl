package jpl

import "runtime/debug"

// 版本信息
const (
	Version      = "0.9.0"
	VersionMajor = 0
	VersionMinor = 9
	VersionPatch = 0
	ReleaseDate  = "2026-04-02"
)

// BuildInfo 返回构建信息
func BuildInfo() map[string]string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return map[string]string{
			"version": Version,
			"date":    ReleaseDate,
		}
	}

	result := map[string]string{
		"version": Version,
		"date":    ReleaseDate,
		"go":      info.GoVersion,
		"path":    info.Path,
	}

	// 提取 Git 提交哈希和构建时间
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			result["commit"] = setting.Value
		case "vcs.time":
			result["build_time"] = setting.Value
		}
	}

	return result
}

// FullVersion 返回完整版本字符串
func FullVersion() string {
	return "JPL v" + Version + " (" + ReleaseDate + ")"
}

// LibVersion 返回库版本字符串
func LibVersion() string {
	return Version
}

// LibSignature 返回库签名
func LibSignature() string {
	return "JPL/" + Version
}
