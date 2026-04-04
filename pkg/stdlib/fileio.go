package stdlib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterFileIO 注册文件 I/O 相关函数到引擎。
//
// 注册的函数：
//   - 文件读写: read, readLines, write, append, exists, file_get_contents, file_put_contents, copy, readfile
//   - 目录操作: mkdir, mkdirAll, rmdir, listDir
//   - 文件信息: stat, fileSize, isFile, isDir
//   - 路径处理: dirname, basename, extname, joinPath, absPath, relPath, cwd, realpath, pathinfo
//   - 文件系统: chdir, rename, unlink, is_readable, is_writable, chmod, scandir, glob
//   - 流操作: fseek, ftell, rewind, ftruncate, fgetcsv
//
// 参数：
//   - e: 引擎实例
func RegisterFileIO(e *engine.Engine) {
	// 文件读写
	e.RegisterFunc("read", builtinRead)
	e.RegisterFunc("readLines", builtinReadLines)
	e.RegisterFunc("write", builtinWrite)
	e.RegisterFunc("append", builtinAppend)
	e.RegisterFunc("exists", builtinExists)
	e.RegisterFunc("file_get_contents", builtinFileGetContents)
	e.RegisterFunc("file_put_contents", builtinFilePutContents)
	e.RegisterFunc("copy", builtinCopy)
	e.RegisterFunc("readfile", builtinReadfile)

	// 目录操作
	e.RegisterFunc("mkdir", builtinMkdir)
	e.RegisterFunc("mkdirAll", builtinMkdirAll)
	e.RegisterFunc("rmdir", builtinRmdir)
	e.RegisterFunc("listDir", builtinListDir)

	// 文件信息
	e.RegisterFunc("stat", builtinStat)
	e.RegisterFunc("fileSize", builtinFileSize)
	e.RegisterFunc("file_size", builtinFileSize) // snake_case alias
	e.RegisterFunc("isFile", builtinIsFile)
	e.RegisterFunc("is_file", builtinIsFile) // snake_case alias
	e.RegisterFunc("isDir", builtinIsDir)
	e.RegisterFunc("is_dir", builtinIsDir)       // snake_case alias
	e.RegisterFunc("file_exists", builtinExists) // PHP-style alias

	// 路径处理
	e.RegisterFunc("dirname", builtinDirname)
	e.RegisterFunc("basename", builtinBasename)
	e.RegisterFunc("extname", builtinExtname)
	e.RegisterFunc("joinPath", builtinJoinPath)
	e.RegisterFunc("absPath", builtinAbsPath)
	e.RegisterFunc("relPath", builtinRelPath)
	e.RegisterFunc("cwd", builtinCwd)
	e.RegisterFunc("realpath", builtinRealpath)
	e.RegisterFunc("pathinfo", builtinPathinfo)

	// Phase 7.6: 文件系统操作
	e.RegisterFunc("chdir", builtinChdir)
	e.RegisterFunc("rename", builtinRename)
	e.RegisterFunc("unlink", builtinUnlink)
	e.RegisterFunc("is_readable", builtinIsReadable)
	e.RegisterFunc("is_writable", builtinIsWritable)
	e.RegisterFunc("chmod", builtinChmod)
	e.RegisterFunc("scandir", builtinScandir)
	e.RegisterFunc("glob", builtinGlob)

	// Phase 11.4: 流操作（Phase 8 流类型已实现）
	e.RegisterFunc("fseek", builtinFseek)
	e.RegisterFunc("ftell", builtinFtell)
	e.RegisterFunc("rewind", builtinRewind)
	e.RegisterFunc("ftruncate", builtinFtruncate)
	e.RegisterFunc("fgetcsv", builtinFgetcsv)

	// P1: 文件 I/O 增强
	e.RegisterFunc("tempfile", builtinTempfile)
	e.RegisterFunc("read_json", builtinReadJSON)
	e.RegisterFunc("write_json", builtinWriteJSON)
	e.RegisterFunc("walk", builtinWalk)

	// 模块注册 — import "file" 可用
	e.RegisterModule("file", map[string]engine.GoFunction{
		"read":              builtinRead,
		"readLines":         builtinReadLines,
		"write":             builtinWrite,
		"append":            builtinAppend,
		"exists":            builtinExists,
		"file_get_contents": builtinFileGetContents,
		"file_put_contents": builtinFilePutContents,
		"copy":              builtinCopy,
		"readfile":          builtinReadfile,
		"mkdir":             builtinMkdir,
		"mkdirAll":          builtinMkdirAll,
		"rmdir":             builtinRmdir,
		"listDir":           builtinListDir,
		"stat":              builtinStat,
		"fileSize":          builtinFileSize,
		"isFile":            builtinIsFile,
		"isDir":             builtinIsDir,
		"dirname":           builtinDirname,
		"basename":          builtinBasename,
		"extname":           builtinExtname,
		"joinPath":          builtinJoinPath,
		"absPath":           builtinAbsPath,
		"relPath":           builtinRelPath,
		"cwd":               builtinCwd,
		"realpath":          builtinRealpath,
		"pathinfo":          builtinPathinfo,
		"chdir":             builtinChdir,
		"rename":            builtinRename,
		"unlink":            builtinUnlink,
		"is_readable":       builtinIsReadable,
		"is_writable":       builtinIsWritable,
		"chmod":             builtinChmod,
		"scandir":           builtinScandir,
		"glob":              builtinGlob,
		"tempfile":          builtinTempfile,
		"read_json":         builtinReadJSON,
		"write_json":        builtinWriteJSON,
		"walk":              builtinWalk,
	})
}

// FileIONames 返回文件 I/O 函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func FileIONames() []string {
	return []string{
		// 文件读写
		"read", "readLines", "write", "append", "exists",
		"file_get_contents", "file_put_contents", "copy", "readfile",
		// 目录操作
		"mkdir", "mkdirAll", "rmdir", "listDir",
		// 文件信息
		"stat", "fileSize", "isFile", "isDir",
		// 路径处理
		"dirname", "basename", "extname", "joinPath", "absPath", "relPath", "cwd", "realpath", "pathinfo",
		// Phase 7.6: 文件系统操作
		"chdir", "rename", "unlink", "is_readable", "is_writable", "chmod", "scandir", "glob",
		// Phase 11.4: 流操作 stub
		"fseek", "ftell", "rewind", "ftruncate", "fgetcsv",
		// P1: 文件 I/O 增强
		"tempfile", "read_json", "write_json", "walk",
	}
}

// ============================================================================
// 文件读写函数
// ============================================================================

// builtinRead 读取文件内容为字符串。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - string: 文件内容
//   - error: 文件不存在或读取失败
//
// 使用示例：
//
//	$content = read("config.txt")
//	$json = read("data.json")
func builtinRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("read() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("read() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read(%q): %w", path, err)
	}

	return engine.NewString(string(content)), nil
}

// builtinReadLines 按行读取文件，返回字符串数组。
//
// 以换行符分割文件内容，返回每行作为数组元素。
// 注意：空行也会被包含在数组中。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - array: 行字符串数组
//   - error: 文件不存在或读取失败
//
// 使用示例：
//
//	$lines = readLines("data.txt")
//	$first = $lines[0]
func builtinReadLines(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("readLines() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("readLines() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("readLines(%q): %w", path, err)
	}

	lines := strings.Split(string(content), "\n")
	values := make([]engine.Value, len(lines))
	for i, line := range lines {
		values[i] = engine.NewString(line)
	}

	return engine.NewArray(values), nil
}

// builtinWrite 写入字符串到文件（覆盖模式）。
//
// 如果文件不存在则创建，存在则覆盖。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//   - args[1]: 要写入的内容（字符串）
//
// 返回值：
//   - null: 成功返回 null
//   - error: 写入失败
//
// 使用示例：
//
//	write("output.txt", "Hello, World!")
//	write("config.json", json_encode({key: "value"}))
func builtinWrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("write() expects 2 arguments (path, content), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("write() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("write() second argument must be string, got %s", args[1].Type())
	}

	path := args[0].String()
	content := args[1].String()

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("write(%q): %w", path, err)
	}

	return engine.NewNull(), nil
}

// builtinAppend 追加字符串到文件末尾。
//
// 如果文件不存在则创建，存在则在末尾追加内容。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//   - args[1]: 要追加的内容（字符串）
//
// 返回值：
//   - null: 成功返回 null
//   - error: 写入失败
//
// 使用示例：
//
//	append("log.txt", "New log entry\n")
func builtinAppend(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("append() expects 2 arguments (path, content), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("append() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("append() second argument must be string, got %s", args[1].Type())
	}

	path := args[0].String()
	content := args[1].String()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("append(%q): %w", path, err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return nil, fmt.Errorf("append(%q): %w", path, err)
	}

	return engine.NewNull(), nil
}

// builtinExists 检查文件或目录是否存在。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 路径（字符串）
//
// 返回值：
//   - bool: 存在返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	exists("config.txt")     // → true/false
//	exists("/tmp")           // → true
func builtinExists(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exists() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("exists() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(err == nil), nil
}

// ============================================================================
// 目录操作函数
// ============================================================================

// builtinMkdir 创建目录。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目录路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 创建失败
//
// 使用示例：
//
//	mkdir("newdir")
func builtinMkdir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mkdir() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("mkdir() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	err := os.Mkdir(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("mkdir(%q): %w", path, err)
	}

	return engine.NewBool(true), nil
}

// builtinMkdirAll 递归创建目录。
//
// 创建路径中所有不存在的目录，类似 mkdir -p。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目录路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 创建失败
//
// 使用示例：
//
//	mkdirAll("a/b/c")
func builtinMkdirAll(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("mkdirAll() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("mkdirAll() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, fmt.Errorf("mkdirAll(%q): %w", path, err)
	}

	return engine.NewBool(true), nil
}

// builtinRmdir 删除空目录
func builtinRmdir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rmdir() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("rmdir() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	err := os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("rmdir(%q): %w", path, err)
	}

	return engine.NewBool(true), nil
}

// builtinListDir 列出目录内容，返回文件名数组
func builtinListDir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("listDir() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("listDir() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("listDir(%q): %w", path, err)
	}

	values := make([]engine.Value, len(entries))
	for i, entry := range entries {
		values[i] = engine.NewString(entry.Name())
	}

	return engine.NewArray(values), nil
}

// ============================================================================
// 文件信息函数
// ============================================================================

// builtinStat 获取文件信息，返回对象
func builtinStat(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("stat() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("stat() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat(%q): %w", path, err)
	}

	obj := map[string]engine.Value{
		"name":    engine.NewString(info.Name()),
		"size":    engine.NewInt(info.Size()),
		"isDir":   engine.NewBool(info.IsDir()),
		"isFile":  engine.NewBool(!info.IsDir()),
		"modTime": engine.NewString(info.ModTime().Format("2006-01-02 15:04:05")),
	}

	return engine.NewObject(obj), nil
}

// builtinFileSize 获取文件大小（字节数）
func builtinFileSize(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("fileSize() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("fileSize() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("fileSize(%q): %w", path, err)
	}

	return engine.NewInt(info.Size()), nil
}

// builtinIsFile 检查路径是否为文件。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 路径（字符串）
//
// 返回值：
//   - bool: 是文件返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	isFile("config.txt")     // → true
//	isFile("mydir")          // → false
func builtinIsFile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isFile() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("isFile() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(!info.IsDir()), nil
}

// builtinIsDir 检查路径是否为目录。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 路径（字符串）
//
// 返回值：
//   - bool: 是目录返回 true
//   - error: 参数错误
//
// 使用示例：
//
//	isDir("mydir")           // → true
//	isDir("config.txt")      // → false
func builtinIsDir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isDir() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("isDir() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(info.IsDir()), nil
}

// ============================================================================
// 路径处理函数
// ============================================================================

// builtinDirname 获取路径的目录部分
func builtinDirname(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dirname() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dirname() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	return engine.NewString(filepath.Dir(path)), nil
}

// builtinBasename 获取路径的文件名部分
func builtinBasename(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("basename() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("basename() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	return engine.NewString(filepath.Base(path)), nil
}

// builtinExtname 获取文件扩展名（含点号）
func builtinExtname(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("extname() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("extname() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	return engine.NewString(filepath.Ext(path)), nil
}

// builtinJoinPath 拼接路径组件
func builtinJoinPath(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("joinPath() expects at least 1 argument, got %d", len(args))
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		if arg.Type() != engine.TypeString {
			return nil, fmt.Errorf("joinPath() arguments must be strings, got %s at index %d", arg.Type(), i)
		}
		parts[i] = arg.String()
	}

	return engine.NewString(filepath.Join(parts...)), nil
}

// builtinAbsPath 获取绝对路径
func builtinAbsPath(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("absPath() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("absPath() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("absPath() failed: %v", err)
	}
	return engine.NewString(abs), nil
}

// builtinRelPath 计算相对路径
func builtinRelPath(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("relPath() expects 2 arguments (target, base), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("relPath() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("relPath() second argument must be string, got %s", args[1].Type())
	}

	target := args[0].String()
	base := args[1].String()
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return nil, fmt.Errorf("relPath() failed: %v", err)
	}
	return engine.NewString(rel), nil
}

// builtinCwd 返回当前工作目录路径。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - string: 当前工作目录的绝对路径
//   - error: 获取失败
//
// 使用示例：
//
//	cwd()                    // → "/home/user"
func builtinCwd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("cwd() expects 0 arguments, got %d", len(args))
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cwd() failed: %v", err)
	}
	return engine.NewString(dir), nil
}

// ============================================================================
// Phase 7.6: 文件系统扩展函数
// ============================================================================

// builtinChdir 改变当前工作目录
func builtinChdir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("chdir() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("chdir() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	if err := os.Chdir(path); err != nil {
		return nil, fmt.Errorf("chdir() failed: %v", err)
	}
	return engine.NewBool(true), nil
}

// builtinRename 重命名或移动文件/目录
//
// 使用 os.Rename 实现，等同于 Unix mv 命令。
// 可以用于：
//   - 重命名文件：rename("old.txt", "new.txt")
//   - 移动文件：rename("dir1/file.txt", "dir2/file.txt")
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 原路径（字符串）
//   - args[1]: 新路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 操作失败（如目标存在、权限不足、跨文件系统移动等）
//
// 使用示例：
//
//	rename("old.txt", "new.txt")          // 重命名
//	rename("file.txt", "backup/file.txt")  // 移动到新目录
func builtinRename(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("rename() expects 2 arguments (old, new), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString || args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("rename() arguments must be strings")
	}

	oldPath := args[0].String()
	newPath := args[1].String()
	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("rename() failed: %v", err)
	}
	return engine.NewBool(true), nil
}

// builtinUnlink 删除文件
//
// ⚠️ 注意：此操作不可恢复，文件将被永久删除
//
// 删除指定文件。注意：
//   - 不能删除目录（使用 rmdir）
//   - 文件不存在会返回错误
//
// unlink(path) → bool
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 删除失败（如文件不存在、权限不足）
//
// 使用示例：
//
//	unlink("temp.txt")     // 删除文件
//	unlink("/tmp/cache")   // 删除临时文件
func builtinUnlink(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unlink() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("unlink() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("unlink() failed: %v", err)
	}
	return engine.NewBool(true), nil
}

// builtinRealpath 返回规范化的绝对路径
func builtinRealpath(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("realpath() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("realpath() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		real = path
	}
	abs, err := filepath.Abs(real)
	if err != nil {
		return nil, fmt.Errorf("realpath() failed: %v", err)
	}
	return engine.NewString(abs), nil
}

// builtinIsReadable 检查文件/目录是否可读
func builtinIsReadable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_readable() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("is_readable() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}
	mode := info.Mode()
	return engine.NewBool(mode&0444 != 0), nil
}

// builtinIsWritable 检查文件/目录是否可写
func builtinIsWritable(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_writable() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("is_writable() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	info, err := os.Stat(path)
	if err != nil {
		return engine.NewBool(false), nil
	}
	mode := info.Mode()
	return engine.NewBool(mode&0222 != 0), nil
}

// builtinChmod 修改文件或目录的权限
//
// 修改文件或目录的访问权限（Unix 风格）。
// 权限使用八进制数表示，如 0755、0644。
//
// 权限位说明：
//   - 0: 无权限
//   - 1: 执行权限 (x)
//   - 2: 写权限 (w)
//   - 4: 读权限 (r)
//   - 组合：4+2+1=7 (rwx)
//
// 常用权限：
//   - 0755: 所有者 rwx，组/其他 rx（目录常见）
//   - 0644: 所有者 rw，组/其他 r（文件常见）
//   - 0600: 只有所有者可读写（敏感文件）
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//   - args[1]: 权限模式（整数，如 0755）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 操作失败
//
// 使用示例：
//
//	chmod("script.sh", 0755)   // 添加执行权限
//	chmod("secret.txt", 0600)  // 仅所有者可读写
func builtinChmod(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("chmod() expects 2 arguments (path, mode), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("chmod() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeInt {
		return nil, fmt.Errorf("chmod() second argument must be int (mode), got %s", args[1].Type())
	}

	path := args[0].String()
	mode := os.FileMode(args[1].Int())
	if err := os.Chmod(path, mode); err != nil {
		return nil, fmt.Errorf("chmod() failed: %v", err)
	}
	return engine.NewBool(true), nil
}

// builtinScandir 扫描目录内容，返回条目数组。
//
// 返回目录中所有文件和子目录的名称列表（不包含 . 和 ..）。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 目录路径（字符串）
//
// 返回值：
//   - array: 文件/目录名字符串数组
//   - error: 目录不存在或读取失败
//
// 使用示例：
//
//	scandir(".")             // → ["file1.txt", "subdir", ...]
//	scandir("/tmp")
func builtinScandir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("scandir() expects 1-2 arguments (path, [sorting_order]), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("scandir() first argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	sorting := 0
	if len(args) >= 2 {
		sorting = int(args[1].Int())
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("scandir() failed: %v", err)
	}

	var result []engine.Value
	for _, entry := range entries {
		result = append(result, engine.NewString(entry.Name()))
	}

	// Simple sorting
	if sorting == 1 {
		// TODO: reverse sort
	}

	return engine.NewArray(result), nil
}

// builtinGlob 按模式查找匹配的文件。
//
// 支持的通配符：
//   - *: 匹配任意字符（不含路径分隔符）
//   - **: 匹配任意字符（包含路径分隔符）
//   - ?: 匹配单个字符
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 匹配模式（字符串）
//
// 返回值：
//   - array: 匹配的文件路径数组
//   - error: 参数错误
//
// 使用示例：
//
//	glob("*.go")             // → ["main.go", "util.go", ...]
//	glob("**/*.txt")         // → 递归查找所有 .txt 文件
func builtinGlob(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("glob() expects 1-2 arguments (pattern, [flags]), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("glob() first argument must be string, got %s", args[0].Type())
	}

	pattern := args[0].String()
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob() failed: %v", err)
	}

	var result []engine.Value
	for _, match := range matches {
		result = append(result, engine.NewString(match))
	}

	return engine.NewArray(result), nil
}

// ============================================================================
// Phase 11.4: 文件 IO 增强函数
// ============================================================================

// builtinFileGetContents 读取文件内容为字符串（PHP 风格别名）。
//
// 与 read() 功能相同，提供 PHP 兼容的函数名。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - string: 文件内容
//   - error: 文件不存在或读取失败
//
// 使用示例：
//
//	$content = file_get_contents("config.txt")
func builtinFileGetContents(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("file_get_contents() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("file_get_contents() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("file_get_contents(%q): %w", path, err)
	}

	return engine.NewString(string(content)), nil
}

// builtinFilePutContents 写入字符串到文件（PHP 风格别名）。
//
// 与 write() 功能相同，提供 PHP 兼容的函数名。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//   - args[1]: 要写入的内容（字符串）
//
// 返回值：
//   - int: 写入的字节数
//   - error: 写入失败
//
// 使用示例：
//
//	file_put_contents("output.txt", "Hello, World!")
func builtinFilePutContents(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("file_put_contents() expects 2 arguments (path, content), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("file_put_contents() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("file_put_contents() second argument must be string, got %s", args[1].Type())
	}

	path := args[0].String()
	content := args[1].String()

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("file_put_contents(%q): %w", path, err)
	}

	return engine.NewInt(int64(len(content))), nil
}

// builtinCopy 复制文件。
//
// 将源文件内容复制到目标文件。如果目标文件存在则覆盖。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 源文件路径（字符串）
//   - args[1]: 目标文件路径（字符串）
//
// 返回值：
//   - bool: 成功返回 true
//   - error: 复制失败
//
// 使用示例：
//
//	copy("source.txt", "backup.txt")
func builtinCopy(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("copy() expects 2 arguments (source, dest), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("copy() first argument must be string, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("copy() second argument must be string, got %s", args[1].Type())
	}

	src := args[0].String()
	dst := args[1].String()

	srcFile, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("copy() cannot open source %q: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("copy() cannot create dest %q: %w", dst, err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, fmt.Errorf("copy() failed: %w", err)
	}

	return engine.NewBool(true), nil
}

// builtinReadfile 读取文件并返回内容字符串。
//
// 与 read() 功能相同，提供 PHP 兼容的函数名。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - string: 文件内容
//   - error: 文件不存在或读取失败
//
// 使用示例：
//
//	$content = readfile("config.txt")
func builtinReadfile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("readfile() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("readfile() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("readfile(%q): %w", path, err)
	}

	return engine.NewString(string(content)), nil
}

// builtinPathinfo 返回路径信息对象。
//
// 返回包含 dirname、basename、extension、filename 的对象。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件路径（字符串）
//
// 返回值：
//   - object: 包含路径组成部分的对象
//   - error: 参数错误
//
// 使用示例：
//
//	pathinfo("/path/to/file.txt")  // → {dirname: "/path/to", basename: "file.txt", extension: "txt", filename: "file"}
func builtinPathinfo(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("pathinfo() expects 1 argument (path), got %d", len(args))
	}
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("pathinfo() argument must be string, got %s", args[0].Type())
	}

	path := args[0].String()
	ext := filepath.Ext(path)
	basename := filepath.Base(path)
	dirname := filepath.Dir(path)
	filename := strings.TrimSuffix(basename, ext)
	// Remove leading dot from extension
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}

	obj := map[string]engine.Value{
		"dirname":   engine.NewString(dirname),
		"basename":  engine.NewString(basename),
		"extension": engine.NewString(ext),
		"filename":  engine.NewString(filename),
	}

	return engine.NewObject(obj), nil
}

// ============================================================================
// Phase 11.4: 流操作（Phase 8 流类型已实现）
// ============================================================================

// builtinFseek 移动文件指针位置。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件流
//   - args[1]: 偏移量（整数）
//   - args[2]: 起始位置（整数，0=SEEK_SET, 1=SEEK_CUR, 2=SEEK_END）
//
// 返回值：
//   - 新位置（整数）
func builtinFseek(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("fseek() expects 3 arguments: stream, offset, whence")
	}

	sv := engine.ToStreamValue(args[0])
	if sv == nil {
		return nil, fmt.Errorf("fseek() argument 1 must be a stream")
	}

	offset := args[1].Int()
	whence := int(args[2].Int())

	switch whence {
	case 0:
		whence = io.SeekStart
	case 1:
		whence = io.SeekCurrent
	case 2:
		whence = io.SeekEnd
	default:
		return nil, fmt.Errorf("fseek(): invalid whence %d (use 0=SET, 1=CUR, 2=END)", whence)
	}

	pos, err := sv.Seek(offset, whence)
	if err != nil {
		return nil, fmt.Errorf("fseek(): %w", err)
	}
	return engine.NewInt(pos), nil
}

// builtinFtell 获取文件指针当前位置。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件流
//
// 返回值：
//   - 当前位置（整数）
func builtinFtell(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ftell() expects 1 argument: stream")
	}

	sv := engine.ToStreamValue(args[0])
	if sv == nil {
		return nil, fmt.Errorf("ftell() argument must be a stream")
	}

	pos, err := sv.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("ftell(): %w", err)
	}
	return engine.NewInt(pos), nil
}

// builtinRewind 重置文件指针到开头。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件流
//
// 返回值：
//   - null
func builtinRewind(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("rewind() expects 1 argument: stream")
	}

	sv := engine.ToStreamValue(args[0])
	if sv == nil {
		return nil, fmt.Errorf("rewind() argument must be a stream")
	}

	_, err := sv.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("rewind(): %w", err)
	}
	return engine.NewNull(), nil
}

// builtinFtruncate 截断文件到指定大小。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件流
//   - args[1]: 新大小（整数）
//
// 返回值：
//   - bool: 是否成功
func builtinFtruncate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ftruncate() expects 2 arguments: stream, size")
	}

	sv := engine.ToStreamValue(args[0])
	if sv == nil {
		return nil, fmt.Errorf("ftruncate() argument 1 must be a stream")
	}

	size := args[1].Int()
	if err := sv.Truncate(size); err != nil {
		return nil, fmt.Errorf("ftruncate(): %w", err)
	}
	return engine.NewBool(true), nil
}

// builtinFgetcsv 从 CSV 文件读取一行。
// 使用 csv.NewReader 读取 CSV 数据。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 文件流
//   - args[1]: 分隔符（可选，默认逗号）
//   - args[2]: 包围符（可选，默认双引号）
//
// 返回值：
//   - 数组: CSV 字段数组
func builtinFgetcsv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("fgetcsv() expects at least 1 argument: stream")
	}

	sv := engine.ToStreamValue(args[0])
	if sv == nil {
		return nil, fmt.Errorf("fgetcsv() argument 1 must be a stream")
	}

	// 读取一行
	var line string
	buf := make([]byte, 4096)
	var lineBuf []byte

	for {
		n, err := sv.Read(buf)
		if n > 0 {
			for i := range n {
				if buf[i] == '\n' {
					line = string(lineBuf)
					// 回退多余读取的部分（将 \n 之后的内容放回）
					if i+1 < n {
						// 需要把剩余字节 seek 回去
						sv.Seek(int64(i+1-n), io.SeekCurrent)
					}
					return parseCSVLine(line, args[1:])
				}
				lineBuf = append(lineBuf, buf[i])
			}
		}
		if err == io.EOF {
			if len(lineBuf) > 0 {
				line = string(lineBuf)
				return parseCSVLine(line, args[1:])
			}
			return engine.NewNull(), nil
		}
		if err != nil {
			return nil, fmt.Errorf("fgetcsv(): %w", err)
		}
	}
}

// parseCSVLine 解析一行 CSV 数据
func parseCSVLine(line string, extraArgs []engine.Value) (engine.Value, error) {
	delimiter := ','
	if len(extraArgs) >= 1 {
		d := extraArgs[0].String()
		if len(d) > 0 {
			delimiter = rune(d[0])
		}
	}

	var fields []string
	var field strings.Builder
	inQuote := false

	for i := 0; i < len(line); i++ {
		ch := rune(line[i])
		if ch == '"' {
			if inQuote && i+1 < len(line) && line[i+1] == '"' {
				field.WriteRune('"')
				i++
			} else {
				inQuote = !inQuote
			}
		} else if ch == delimiter && !inQuote {
			fields = append(fields, field.String())
			field.Reset()
		} else {
			field.WriteRune(ch)
		}
	}
	fields = append(fields, field.String())

	arr := make([]engine.Value, len(fields))
	for i, f := range fields {
		arr[i] = engine.NewString(f)
	}
	return engine.NewArray(arr), nil
}

// ============================================================================
// P1: 文件 I/O 增强函数
// ============================================================================

// builtinTempfile 创建临时文件并返回路径。
func builtinTempfile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	prefix := "jpl_tmp"
	suffix := ""
	if len(args) >= 1 {
		prefix = args[0].String()
	}
	if len(args) >= 2 {
		suffix = args[1].String()
	}
	f, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		return nil, fmt.Errorf("tempfile() failed: %v", err)
	}
	f.Close()
	return engine.NewString(f.Name()), nil
}

// builtinReadJSON 读取并解析 JSON 文件。
func builtinReadJSON(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("read_json() expects 1 argument, got %d", len(args))
	}
	content, err := os.ReadFile(args[0].String())
	if err != nil {
		return nil, fmt.Errorf("read_json() failed: %v", err)
	}
	return builtinJSONDecode(ctx, []engine.Value{engine.NewString(string(content))})
}

// builtinWriteJSON 序列化并写入 JSON 文件。
func builtinWriteJSON(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("write_json() expects 2-3 arguments, got %d", len(args))
	}
	path := args[0].String()
	value := args[1]
	pretty := false
	if len(args) == 3 {
		pretty = engine.IsTruthy(args[2])
	}
	goValue := jplValueToGo(value)
	var bytes []byte
	var err error
	if pretty {
		bytes, err = json.MarshalIndent(goValue, "", "  ")
	} else {
		bytes, err = json.Marshal(goValue)
	}
	if err != nil {
		return nil, fmt.Errorf("write_json() failed: %v", err)
	}
	if err := os.WriteFile(path, append(bytes, '\n'), 0644); err != nil {
		return nil, fmt.Errorf("write_json() failed: %v", err)
	}
	return engine.NewNull(), nil
}

// builtinWalk 递归遍历目录。
func builtinWalk(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("walk() expects 1 argument, got %d", len(args))
	}
	root := args[0].String()
	var result []engine.Value
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		entry := map[string]engine.Value{
			"path":   engine.NewString(path),
			"is_dir": engine.NewBool(info.IsDir()),
			"size":   engine.NewInt(info.Size()),
		}
		result = append(result, engine.NewObject(entry))
		return nil
	})
	return engine.NewArray(result), nil
}

// FileIOSigs returns function signatures for REPL :doc command.
func FileIOSigs() map[string]string {
	return map[string]string{
		"read":                 "read(path) → string  — Read file content as string",
		"write":                "write(path, content) → null  — Write string to file (overwrite)",
		"append":               "append(path, content) → null  — Append string to file",
		"fopen":                "fopen(path, [mode]) → stream  — Open file stream",
		"fread":                "fread(stream, length) → string  — Read bytes from stream",
		"fgets":                "fgets(stream) → string  — Read line from stream",
		"fwrite":               "fwrite(stream, data) → int  — Write data to stream",
		"fclose":               "fclose(stream) → null  — Close stream",
		"feof":                 "feof(stream) → bool  — Check if stream at EOF",
		"fflush":               "fflush(stream) → bool  — Flush stream buffer",
		"file_exists":          "file_exists(path) → bool  — Check if file exists",
		"is_file":              "is_file(path) → bool  — Check if path is a file",
		"is_dir":               "is_dir(path) → bool  — Check if path is a directory",
		"file_size":            "file_size(path) → int  — Get file size in bytes",
		"file_get_contents":    "file_get_contents(path) → string  — Read entire file as string",
		"file_put_contents":    "file_put_contents(path, content) → int  — Write string to file, return bytes written",
		"copy":                 "copy(source, dest) → bool  — Copy file",
		"readfile":             "readfile(path) → string  — Read file and return content",
		"pathinfo":             "pathinfo(path) → object  — Get path components (dirname, basename, extension, filename)",
		"chdir":                "chdir(path) → bool  — Change working directory",
		"rename":               "rename(old, new) → bool  — Rename or move file",
		"unlink":               "unlink(path) → bool  — Delete file",
		"realpath":             "realpath(path) → string  — Get canonical absolute path",
		"is_readable":          "is_readable(path_or_stream) → bool  — Check if readable",
		"is_writable":          "is_writable(path_or_stream) → bool  — Check if writable",
		"chmod":                "chmod(path, mode) → bool  — Change file permissions",
		"scandir":              "scandir(path, [sorting_order]) → array  — List directory contents",
		"glob":                 "glob(pattern, [flags]) → array  — Find files matching pattern",
		"cwd":                  "cwd() → string  — Get current working directory",
		"fseek":                "fseek(stream, offset, whence) → int  — Seek file position",
		"ftell":                "ftell(stream) → int  — Get current file position",
		"rewind":               "rewind(stream) → null  — Reset file pointer to start",
		"ftruncate":            "ftruncate(stream, size) → bool  — Truncate file to size",
		"fgetcsv":              "fgetcsv(stream, [delimiter], [enclosure]) → array  — Read CSV line",
		"stream_get_meta_data": "stream_get_meta_data(stream) → object  — Get stream metadata",
		"tempfile":             "tempfile([prefix], [suffix]) → string  — Create temp file, return path",
		"read_json":            "read_json(path) → value  — Read and parse JSON file",
		"write_json":           "write_json(path, value, [pretty]) → null  — Write value as JSON to file",
		"walk":                 "walk(root) → array  — Recursive directory traversal",
	}
}
