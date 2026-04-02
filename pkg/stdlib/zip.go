package stdlib

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/gnuos/jpl/engine"
)

// RegisterZip 将 zip 压缩函数注册到引擎。
//
// 注册的函数：
//   - zip_open: 打开 zip 文件
//   - zip_read: 读取 zip 条目
//   - zip_entry_name: 获取 zip 条目名称
//   - zip_entry_filesize: 获取 zip 条目原始大小
//   - zip_entry_compressedsize: 获取 zip 条目压缩后大小
//   - zip_entry_read: 读取 zip 条目内容
//   - zip_entry_close: 关闭当前 zip 条目
//   - zip_close: 关闭 zip 文件
//   - zip_create: 创建 zip 文件
//
// 参数：
//   - e: 引擎实例
func RegisterZip(e *engine.Engine) {
	e.RegisterFunc("zip_open", builtinZipOpen)
	e.RegisterFunc("zip_read", builtinZipRead)
	e.RegisterFunc("zip_entry_name", builtinZipEntryName)
	e.RegisterFunc("zip_entry_filesize", builtinZipEntryFilesize)
	e.RegisterFunc("zip_entry_compressedsize", builtinZipEntryCompressedSize)
	e.RegisterFunc("zip_entry_read", builtinZipEntryRead)
	e.RegisterFunc("zip_entry_close", builtinZipEntryClose)
	e.RegisterFunc("zip_close", builtinZipClose)
	e.RegisterFunc("zip_create", builtinZipCreate)

	e.RegisterModule("zip", map[string]engine.GoFunction{
		"zip_open":                 builtinZipOpen,
		"zip_read":                 builtinZipRead,
		"zip_entry_name":           builtinZipEntryName,
		"zip_entry_filesize":       builtinZipEntryFilesize,
		"zip_entry_compressedsize": builtinZipEntryCompressedSize,
		"zip_entry_read":           builtinZipEntryRead,
		"zip_entry_close":          builtinZipEntryClose,
		"zip_close":                builtinZipClose,
		"zip_create":               builtinZipCreate,
	})
}

// ZipNames 返回 zip 函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func ZipNames() []string {
	return []string{
		"zip_open",
		"zip_read",
		"zip_entry_name",
		"zip_entry_filesize",
		"zip_entry_compressedsize",
		"zip_entry_read",
		"zip_entry_close",
		"zip_close",
		"zip_create",
	}
}

type ZipHandle struct {
	filename string
	reader   *zip.ReadCloser
	entries  []*zip.File
	current  int
}

type ZipEntry struct {
	file *zip.File
	name string
}

func (z *ZipHandle) Type() engine.ValueType               { return engine.TypeObject }
func (z *ZipHandle) IsNull() bool                         { return z == nil || z.reader == nil }
func (z *ZipHandle) Bool() bool                           { return z != nil && z.reader != nil }
func (z *ZipHandle) Int() int64                           { return 0 }
func (z *ZipHandle) Float() float64                       { return 0.0 }
func (z *ZipHandle) String() string                       { return "zip:" + z.filename }
func (z *ZipHandle) Stringify() string                    { return z.String() }
func (z *ZipHandle) Array() []engine.Value                { return nil }
func (z *ZipHandle) Object() map[string]engine.Value      { return nil }
func (z *ZipHandle) Len() int                             { return 0 }
func (z *ZipHandle) Equals(other engine.Value) bool       { return false }
func (z *ZipHandle) Less(other engine.Value) bool         { return false }
func (z *ZipHandle) Greater(other engine.Value) bool      { return false }
func (z *ZipHandle) LessEqual(other engine.Value) bool    { return false }
func (z *ZipHandle) GreaterEqual(other engine.Value) bool { return false }
func (z *ZipHandle) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (z *ZipHandle) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (z *ZipHandle) Add(other engine.Value) engine.Value  { return z }
func (z *ZipHandle) Sub(other engine.Value) engine.Value  { return z }
func (z *ZipHandle) Mul(other engine.Value) engine.Value  { return z }
func (z *ZipHandle) Div(other engine.Value) engine.Value  { return z }
func (z *ZipHandle) Mod(other engine.Value) engine.Value  { return z }
func (z *ZipHandle) Negate() engine.Value                 { return z }

func (e *ZipEntry) Type() engine.ValueType               { return engine.TypeObject }
func (e *ZipEntry) IsNull() bool                         { return e == nil || e.file == nil }
func (e *ZipEntry) Bool() bool                           { return e != nil && e.file != nil }
func (e *ZipEntry) Int() int64                           { return 0 }
func (e *ZipEntry) Float() float64                       { return 0.0 }
func (e *ZipEntry) String() string                       { return "zip_entry:" + e.name }
func (e *ZipEntry) Stringify() string                    { return e.String() }
func (e *ZipEntry) Array() []engine.Value                { return nil }
func (e *ZipEntry) Object() map[string]engine.Value      { return nil }
func (e *ZipEntry) Len() int                             { return 0 }
func (e *ZipEntry) Equals(other engine.Value) bool       { return false }
func (e *ZipEntry) Less(other engine.Value) bool         { return false }
func (e *ZipEntry) Greater(other engine.Value) bool      { return false }
func (e *ZipEntry) LessEqual(other engine.Value) bool    { return false }
func (e *ZipEntry) GreaterEqual(other engine.Value) bool { return false }
func (e *ZipEntry) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (e *ZipEntry) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (e *ZipEntry) Add(other engine.Value) engine.Value  { return e }
func (e *ZipEntry) Sub(other engine.Value) engine.Value  { return e }
func (e *ZipEntry) Mul(other engine.Value) engine.Value  { return e }
func (e *ZipEntry) Div(other engine.Value) engine.Value  { return e }
func (e *ZipEntry) Mod(other engine.Value) engine.Value  { return e }
func (e *ZipEntry) Negate() engine.Value                 { return e }

func builtinZipOpen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_open() expects 1 argument, got %d", len(args))
	}

	filename := args[0].String()

	reader, err := zip.OpenReader(filename)
	if err != nil {
		return engine.NewNull(), nil
	}

	handle := &ZipHandle{
		filename: filename,
		reader:   reader,
		entries:  reader.File,
		current:  0,
	}

	return handle, nil
}

func builtinZipRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_read() expects 1 argument, got %d", len(args))
	}

	handle, ok := args[0].(*ZipHandle)
	if !ok || handle == nil || handle.reader == nil {
		return engine.NewNull(), nil
	}

	if handle.current >= len(handle.entries) {
		return engine.NewBool(false), nil
	}

	file := handle.entries[handle.current]
	handle.current++

	entry := &ZipEntry{
		file: file,
		name: file.Name,
	}

	return entry, nil
}

func builtinZipEntryName(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_entry_name() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*ZipEntry)
	if !ok || entry == nil || entry.file == nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(entry.file.Name), nil
}

func builtinZipEntryFilesize(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_entry_filesize() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*ZipEntry)
	if !ok || entry == nil || entry.file == nil {
		return engine.NewNull(), nil
	}

	return engine.NewInt(int64(entry.file.UncompressedSize64)), nil
}

func builtinZipEntryCompressedSize(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_entry_compressedsize() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*ZipEntry)
	if !ok || entry == nil || entry.file == nil {
		return engine.NewNull(), nil
	}

	return engine.NewInt(int64(entry.file.CompressedSize64)), nil
}

func builtinZipEntryRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("zip_entry_read() expects 1-2 arguments, got %d", len(args))
	}

	entry, ok := args[0].(*ZipEntry)
	if !ok || entry == nil || entry.file == nil {
		return engine.NewNull(), nil
	}

	length := 8192
	if len(args) == 2 {
		length = int(args[1].Int())
	}

	rc, err := entry.file.Open()
	if err != nil {
		return engine.NewNull(), nil
	}
	defer rc.Close()

	buf := make([]byte, length)
	n, err := rc.Read(buf)
	if err != nil && err != io.EOF {
		return engine.NewNull(), nil
	}

	return engine.NewString(string(buf[:n])), nil
}

func builtinZipEntryClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_entry_close() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*ZipEntry)
	if !ok || entry == nil {
		return engine.NewInt(0), nil
	}

	entry.file = nil
	entry.name = ""

	return engine.NewInt(0), nil
}

func builtinZipClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zip_close() expects 1 argument, got %d", len(args))
	}

	handle, ok := args[0].(*ZipHandle)
	if !ok || handle == nil || handle.reader == nil {
		return engine.NewInt(0), nil
	}

	err := handle.reader.Close()
	if err != nil {
		return engine.NewInt(0), nil
	}

	handle.reader = nil

	return engine.NewInt(0), nil
}

func builtinZipCreate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("zip_create() expects 2 arguments, got %d", len(args))
	}

	filename := args[0].String()
	entriesArg := args[1]

	if entriesArg.Type() != engine.TypeArray {
		return nil, fmt.Errorf("zip_create() expects array of entries, got %s", entriesArg.Type())
	}

	entries := entriesArg.Array()

	file, err := os.Create(filename)
	if err != nil {
		return engine.NewNull(), nil
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	for _, entry := range entries {
		if entry.Type() != engine.TypeObject {
			continue
		}

		obj := entry.Object()
		name := obj["name"]
		content := obj["content"]

		entryName := name.String()
		entryContent := content.String()

		w, err := zipWriter.Create(entryName)
		if err != nil {
			continue
		}
		_, err = w.Write([]byte(entryContent))
		if err != nil {
			continue
		}
	}

	return engine.NewInt(1), nil
}
