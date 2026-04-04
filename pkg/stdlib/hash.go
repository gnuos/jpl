package stdlib

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"os"

	"github.com/gnuos/jpl/engine"
)

// RegisterHash 注册 Hash/编码函数
func RegisterHash(e *engine.Engine) {
	// 全局注册
	e.RegisterFunc("md5", builtinMd5)
	e.RegisterFunc("sha1", builtinSha1)
	e.RegisterFunc("md5_file", builtinMd5File)
	e.RegisterFunc("sha1_file", builtinSha1File)
	e.RegisterFunc("base64_encode", builtinBase64Encode)
	e.RegisterFunc("base64_decode", builtinBase64Decode)
	e.RegisterFunc("crc32", builtinCrc32)

	// 模块注册 — import "hash" 可用
	e.RegisterModule("hash", map[string]engine.GoFunction{
		"md5": builtinMd5, "sha1": builtinSha1,
		"md5_file": builtinMd5File, "sha1_file": builtinSha1File,
		"base64_encode": builtinBase64Encode, "base64_decode": builtinBase64Decode,
		"crc32": builtinCrc32,
	})
}

// HashNames 返回 Hash/编码函数名称列表
func HashNames() []string {
	return []string{"md5", "sha1", "md5_file", "sha1_file", "base64_encode", "base64_decode", "crc32"}
}

// builtinMd5 计算字符串的 MD5 哈希值
//
// ⚠️ 安全警告：MD5 已被破解用于安全用途，仅用于兼容性或非安全场景
//
// MD5 产生 128 位（32 个十六进制字符）哈希值，常用于：
//   - 文件完整性校验（非安全场景）
//   - 旧系统兼容性
//   - 数据去重（低冲突场景）
//
// md5(data) → hex_string
//
// 参数：
//   - data: 要计算哈希的字符串
//
// 返回值：
//   - string: 32 个十六进制字符的哈希值
//
// 使用示例：
//
//	md5("hello")           // → "5d41402abc4b2a76b9719d911017c592"
//	md5("")                // → "d41d8cd98f00b204e9800998ecf8427e"
//
// 常见用途：
//   - 文件校验：md5(read("file.bin"))
//   - 缓存键：md5(serialize($data))
func builtinMd5(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("md5() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("md5() argument must be a string, got %s", args[0].Type())
	}
	h := md5.Sum([]byte(args[0].String()))
	return engine.NewString(hex.EncodeToString(h[:])), nil
}

// builtinSha1 计算字符串的 SHA-1 哈希值
//
// ⚠️ 安全警告：SHA-1 已被破解用于安全用途，仅用于兼容性或非安全场景
//
// SHA-1 产生 160 位（40 个十六进制字符）哈希值，比 MD5 更安全但仍不推荐用于安全场景。
//
// sha1(data) → hex_string
//
// 参数：
//   - data: 要计算哈希的字符串
//
// 返回值：
//   - string: 40 个十六进制字符的哈希值
//
// 使用示例：
//
//	sha1("hello")          // → "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"
func builtinSha1(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha1() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("sha1() argument must be a string, got %s", args[0].Type())
	}
	h := sha1.Sum([]byte(args[0].String()))
	return engine.NewString(hex.EncodeToString(h[:])), nil
}

// builtinBase64Encode 将字符串编码为 Base64 格式
//
// Base64 将二进制数据转换为 ASCII 字符，便于在 JSON、XML、URL 等纯文本格式中传输。
// 编码后数据量增加约 33%，每 3 字节数据编码为 4 个字符。
//
// base64_encode(data) → string
//
// 参数：
//   - data: 要编码的字符串（任意二进制数据）
//
// 返回值：
//   - string: Base64 编码后的字符串
//
// 使用示例：
//
//	base64_encode("Hello")     // → "SGVsbG8="
//	base64_encode("123")       // → "MTIz"
//	base64_encode("\x00\x01")  // → "AAE="
//
// 常见用途：
//   - URL 参数：base64_encode($binary)
//   - JSON 数据：json_encode({data: base64_encode($binary)})
//   - 认证头：base64_encode($username . ":" . $password)
func builtinBase64Encode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64_encode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("base64_encode() argument must be a string, got %s", args[0].Type())
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(args[0].String()))
	return engine.NewString(encoded), nil
}

// builtinBase64Decode 将 Base64 字符串解码为原始数据
//
// Base64 解码是编码的逆操作。注意：
//   - 输入必须是有效的 Base64 字符串
//   - 解码失败会返回错误
//   - 输出是原始字节（可能包含空字符）
//
// base64_decode(data) → string
//
// 参数：
//   - data: Base64 编码的字符串
//
// 返回值：
//   - string: 解码后的原始字符串
//   - error: 无效的 Base64 格式
//
// 使用示例：
//
//	base64_decode("SGVsbG8=")  // → "Hello"
//	base64_decode("MTIz")       // → "123"
//
// 常见用途：
//
//   - 解析认证头：
//     $auth = base64_decode(substr($header, 6))  // 去掉 "Basic " 前缀
//     $parts = split($auth, ":")
//
//   - 解析 JSON 二进制数据：
//     $binary = base64_decode($json.data)
func builtinBase64Decode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64_decode() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("base64_decode() argument must be a string, got %s", args[0].Type())
	}
	decoded, err := base64.StdEncoding.DecodeString(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("base64_decode() invalid input: %v", err)
	}
	return engine.NewString(string(decoded)), nil
}

// builtinCrc32 计算字符串的 CRC32 校验和
//
// CRC32 是 32 位循环冗余校验，常用于快速检测数据传输或存储中的错误。
// 与 MD5/SHA 不同，CRC32 速度快但碰撞概率较高，不适合安全性要求高的场景。
//
// crc32(data) → int
//
// 参数：
//   - data: 要计算校验和的字符串
//
// 返回值：
//   - int: 32 位无符号整数校验和（0 - 4294967295）
//
// 使用示例：
//
//	crc32("hello")    // → 3619931586
//	crc32("")         // → 0
//
// 常见用途：
//   - 快速校验：检查文件/数据是否发生变化
//   - 哈希表：作为快速哈希键（注意碰撞）
//   - 压缩算法：zip/gzip 等使用 CRC32 校验数据完整性
func builtinCrc32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("crc32() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("crc32() argument must be a string, got %s", args[0].Type())
	}
	sum := crc32.ChecksumIEEE([]byte(args[0].String()))
	return engine.NewInt(int64(sum)), nil
}

// builtinMd5File 计算文件的 MD5 哈希值
//
// 读取整个文件内容并计算 MD5 哈希。适用于小文件，大文件可能占用较多内存。
//
// ⚠️ 注意：不推荐用于安全场景，仅用于文件完整性校验
//
// md5_file(path) → string | null
//
// 参数：
//   - path: 文件路径（字符串）
//
// 返回值：
//   - string: 32 个十六进制字符的哈希值
//   - null: 文件不存在或读取失败
//
// 使用示例：
//
//	$hash = md5_file("data.zip")     // → 文件的 MD5
//	$hash = md5_file("/path/to/file")
//
// 实际应用：
//   - 下载校验：比较本地哈希与官方哈希
//   - 缓存失效：文件修改后哈希变化触发重新处理
func builtinMd5File(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("md5_file() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("md5_file() argument must be a string, got %s", args[0].Type())
	}
	data, err := os.ReadFile(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("md5_file() %v", err)
	}
	h := md5.Sum(data)
	return engine.NewString(hex.EncodeToString(h[:])), nil
}

// builtinSha1File 计算文件的 SHA-1 哈希值
//
// 读取整个文件内容并计算 SHA-1 哈希。适用于小文件，大文件可能占用较多内存。
//
// ⚠️ 注意：不推荐用于安全场景，仅用于文件完整性校验
//
// sha1_file(path) → string | null
//
// 参数：
//   - path: 文件路径（字符串）
//
// 返回值：
//   - string: 40 个十六进制字符的哈希值
//   - null: 文件不存在或读取失败
//
// 使用示例：
//
//	$hash = sha1_file("data.zip")    // → 文件的 SHA-1
//
// 实际应用：
//   - Git 使用 SHA-1 识别文件版本
//   - 大文件校验和计算
func builtinSha1File(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha1_file() expects 1 argument, got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("sha1_file() argument must be a string, got %s", args[0].Type())
	}
	data, err := os.ReadFile(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("sha1_file() %v", err)
	}
	h := sha1.Sum(data)
	return engine.NewString(hex.EncodeToString(h[:])), nil
}

// HashSigs returns function signatures for REPL :doc command.
func HashSigs() map[string]string {
	return map[string]string{
		"md5":           "md5(data) → string  — Calculate MD5 hash (32 hex chars)",
		"sha1":          "sha1(data) → string  — Calculate SHA-1 hash (40 hex chars)",
		"sha256":        "sha256(data) → string  — Calculate SHA-256 hash (64 hex chars)",
		"sha512":        "sha512(data) → string  — Calculate SHA-512 hash (128 hex chars)",
		"crc32":         "crc32(data) → int  — Calculate CRC32 checksum",
		"base64_encode": "base64_encode(data) → string  — Encode to Base64",
		"base64_decode": "base64_decode(data) → string  — Decode from Base64",
		"hex_encode":    "hex_encode(data) → string  — Encode to hexadecimal",
		"hex_decode":    "hex_decode(data) → string  — Decode from hexadecimal",
		"md5_file":      "md5_file(path) → string  — Calculate MD5 hash of file",
		"sha1_file":     "sha1_file(path) → string  — Calculate SHA-1 hash of file",
	}
}
