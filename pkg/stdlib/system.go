package stdlib

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/gnuos/jpl/engine"
)

// RegisterSystem 注册系统相关函数到引擎。
//
// 注册的函数：
//   - 磁盘空间: disk_free_space, disk_total_space
//   - 文件时间: fileatime, filemtime, filectime, touch
//   - 进程信息: getpid, getuid, getgid, umask
//   - 系统信息: uname
//
// 参数：
//   - e: 引擎实例
func RegisterSystem(e *engine.Engine) {
	e.RegisterFunc("disk_free_space", builtinDiskFreeSpace)
	e.RegisterFunc("disk_total_space", builtinDiskTotalSpace)
	e.RegisterFunc("fileatime", builtinFileatime)
	e.RegisterFunc("filemtime", builtinFilemtime)
	e.RegisterFunc("filectime", builtinFilectime)
	e.RegisterFunc("touch", builtinTouch)
	e.RegisterFunc("getpid", builtinGetpid)
	e.RegisterFunc("getuid", builtinGetuid)
	e.RegisterFunc("getgid", builtinGetgid)
	e.RegisterFunc("umask", builtinUmask)
	e.RegisterFunc("uname", builtinUname)

	// 模块注册 — import "sys" 可用
	e.RegisterModule("sys", map[string]engine.GoFunction{
		"disk_free_space": builtinDiskFreeSpace, "disk_total_space": builtinDiskTotalSpace,
		"fileatime": builtinFileatime, "filemtime": builtinFilemtime, "filectime": builtinFilectime,
		"touch":  builtinTouch,
		"getpid": builtinGetpid, "getuid": builtinGetuid, "getgid": builtinGetgid,
		"umask": builtinUmask,
		"uname": builtinUname,
	})
}

// SystemNames 返回系统函数名称列表。
func SystemNames() []string {
	return []string{
		"disk_free_space", "disk_total_space",
		"fileatime", "filemtime", "filectime",
		"touch",
		"getpid", "getuid", "getgid",
		"umask",
		"uname",
	}
}

// ============================================================================
// 磁盘空间函数
// ============================================================================

// builtinDiskFreeSpace 返回指定路径所在磁盘的可用空间（字节）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 路径（字符串）
//
// 返回值：
//   - int: 可用空间（字节）
//   - error: 参数错误或获取失败
//
// 使用示例：
//
//	disk_free_space("/")        // → 可用字节数
//	disk_free_space("/home")
func builtinDiskFreeSpace(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("disk_free_space() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("disk_free_space() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("disk_free_space(%q): %w", path, err)
	}

	// Available space = Bavail * Bsize
	available := int64(stat.Bavail) * int64(stat.Bsize)
	return engine.NewInt(available), nil
}

// builtinDiskTotalSpace 返回指定路径所在磁盘的总空间（字节）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 路径（字符串）
//
// 返回值：
//   - int: 总空间（字节）
//   - error: 参数错误或获取失败
//
// 使用示例：
//
//	disk_total_space("/")       // → 总字节数
func builtinDiskTotalSpace(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("disk_total_space() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("disk_total_space() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("disk_total_space(%q): %w", path, err)
	}

	// Total space = Blocks * Bsize
	total := int64(stat.Blocks) * int64(stat.Bsize)
	return engine.NewInt(total), nil
}

// ============================================================================
// 文件时间函数
// ============================================================================

// builtinFileatime 返回文件的最后访问时间（Unix 时间戳）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - int: Unix 时间戳（秒）
//   - error: 文件不存在或获取失败
//
// 使用示例：
//
//	fileatime("/etc/passwd")    // → 时间戳
func builtinFileatime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("fileatime() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("fileatime() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return nil, fmt.Errorf("fileatime(%q): %w", path, err)
	}

	// Atim is the access time
	atime := stat.Atim.Sec
	return engine.NewInt(atime), nil
}

// builtinFilemtime 返回文件的最后修改时间（Unix 时间戳）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - int: Unix 时间戳（秒）
//   - error: 文件不存在或获取失败
//
// 使用示例：
//
//	filemtime("/etc/passwd")    // → 时间戳
func builtinFilemtime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("filemtime() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("filemtime() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("filemtime(%q): %w", path, err)
	}

	return engine.NewInt(info.ModTime().Unix()), nil
}

// builtinFilectime 返回文件的状态改变时间（Unix 时间戳）。
//
// 注意：在 Linux 上，ctime 是 inode 状态改变时间（如权限修改），
// 不是文件创建时间。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - int: Unix 时间戳（秒）
//   - error: 文件不存在或获取失败
//
// 使用示例：
//
//	filectime("/etc/passwd")    // → 时间戳
func builtinFilectime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("filectime() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("filectime() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return nil, fmt.Errorf("filectime(%q): %w", path, err)
	}

	// Ctim is the status change time
	ctime := stat.Ctim.Sec
	return engine.NewInt(ctime), nil
}

// builtinTouch 修改文件的访问和修改时间。
//
// 如果文件不存在则创建空文件。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 操作失败
//
// 使用示例：
//
//	touch("/tmp/newfile.txt")   // 创建空文件或更新时间
func builtinTouch(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("touch() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("touch() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()

	// Check if file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Create empty file
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("touch(%q): %w", path, err)
		}
		f.Close()
		return engine.NewBool(true), nil
	}
	if err != nil {
		return nil, fmt.Errorf("touch(%q): %w", path, err)
	}

	// Update access and modification times to now
	now := time.Now()
	if err := os.Chtimes(path, now, now); err != nil {
		// Fallback: just open and close to update access time
		f, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("touch(%q): %w", path, err)
		}
		f.Close()
	}

	return engine.NewBool(true), nil
}

// ============================================================================
// 进程信息函数
// ============================================================================

// builtinGetpid 返回当前进程 ID。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - int: 进程 ID
//
// 使用示例：
//
//	getpid()                    // → 12345
func builtinGetpid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("getpid() expects 0 arguments, got %d", len(args))
	}
	return engine.NewInt(int64(os.Getpid())), nil
}

// builtinGetuid 返回当前用户 ID（仅 Unix/Linux）。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - int: 用户 ID
//
// 使用示例：
//
//	getuid()                    // → 1000
func builtinGetuid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("getuid() expects 0 arguments, got %d", len(args))
	}
	return engine.NewInt(int64(os.Getuid())), nil
}

// builtinGetgid 返回当前组 ID（仅 Unix/Linux）。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - int: 组 ID
//
// 使用示例：
//
//	getgid()                    // → 1000
func builtinGetgid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("getgid() expects 0 arguments, got %d", len(args))
	}
	return engine.NewInt(int64(os.Getgid())), nil
}

// builtinUmask 设置或获取文件创建掩码。
//
// 无参数时返回当前掩码值，有参数时设置新掩码并返回旧值。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数（获取）或 1 个参数（设置，整数）
//
// 返回值：
//   - int: 当前/旧的掩码值
//
// 使用示例：
//
//	umask()                     // → 当前掩码
//	umask(0022)                 // → 设置新掩码，返回旧值
func builtinUmask(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) == 0 {
		// Get current umask (read-only, set to 0 then restore)
		mask := syscall.Umask(0)
		syscall.Umask(mask)
		return engine.NewInt(int64(mask)), nil
	}
	if len(args) == 1 {
		if args[0].Type() != engine.TypeInt {
			return nil, fmt.Errorf("umask() argument must be int, got %s", args[0].Type())
		}
		newMask := int(args[0].Int())
		oldMask := syscall.Umask(newMask)
		return engine.NewInt(int64(oldMask)), nil
	}
	return nil, fmt.Errorf("umask() expects 0 or 1 arguments, got %d", len(args))
}

// ============================================================================
// 系统信息函数
// ============================================================================

// builtinUname 返回系统信息对象。
//
// 返回包含 sysname, nodename, release, version, machine 的对象。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - object: 系统信息对象
//
// 使用示例：
//
//	$info = uname()
//	$info["sysname"]            // → "Linux"
//	$info["machine"]            // → "x86_64"
func builtinUname(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("uname() expects 0 arguments, got %d", len(args))
	}

	obj := map[string]engine.Value{
		"sysname":  engine.NewString(runtime.GOOS),
		"nodename": engine.NewString(getHostname()),
		"release":  engine.NewString(""),
		"version":  engine.NewString(runtime.Version()),
		"machine":  engine.NewString(runtime.GOARCH),
	}

	return engine.NewObject(obj), nil
}

// getHostname 获取主机名
func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}
