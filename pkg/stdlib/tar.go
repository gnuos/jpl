package stdlib

import (
	"archive/tar"
	"fmt"
	"io"
	"os"

	"github.com/gnuos/jpl/engine"
)

func RegisterTar(e *engine.Engine) {
	e.RegisterFunc("tar_open", builtinTarOpen)
	e.RegisterFunc("tar_read", builtinTarRead)
	e.RegisterFunc("tar_entry_name", builtinTarEntryName)
	e.RegisterFunc("tar_entry_size", builtinTarEntrySize)
	e.RegisterFunc("tar_entry_isdir", builtinTarEntryIsdir)
	e.RegisterFunc("tar_entry_read", builtinTarEntryRead)
	e.RegisterFunc("tar_entry_close", builtinTarEntryClose)
	e.RegisterFunc("tar_close", builtinTarClose)
	e.RegisterFunc("tar_create", builtinTarCreate)

	e.RegisterModule("tar", map[string]engine.GoFunction{
		"tar_open":        builtinTarOpen,
		"tar_read":        builtinTarRead,
		"tar_entry_name":  builtinTarEntryName,
		"tar_entry_size":  builtinTarEntrySize,
		"tar_entry_isdir": builtinTarEntryIsdir,
		"tar_entry_read":  builtinTarEntryRead,
		"tar_entry_close": builtinTarEntryClose,
		"tar_close":       builtinTarClose,
		"tar_create":      builtinTarCreate,
	})
}

func TarNames() []string {
	return []string{
		"tar_open",
		"tar_read",
		"tar_entry_name",
		"tar_entry_size",
		"tar_entry_isdir",
		"tar_entry_read",
		"tar_entry_close",
		"tar_close",
		"tar_create",
	}
}

type TarHandle struct {
	filename string
	reader   *tar.Reader
	file     *os.File
}

type TarEntry struct {
	header *tar.Header
	name   string
	size   int64
	isdir  bool
}

func (t *TarHandle) Type() engine.ValueType               { return engine.TypeObject }
func (t *TarHandle) IsNull() bool                         { return t == nil || t.reader == nil }
func (t *TarHandle) Bool() bool                           { return t != nil && t.reader != nil }
func (t *TarHandle) Int() int64                           { return 0 }
func (t *TarHandle) Float() float64                       { return 0.0 }
func (t *TarHandle) String() string                       { return "tar:" + t.filename }
func (t *TarHandle) Stringify() string                    { return t.String() }
func (t *TarHandle) Array() []engine.Value                { return nil }
func (t *TarHandle) Object() map[string]engine.Value      { return nil }
func (t *TarHandle) Len() int                             { return 0 }
func (t *TarHandle) Equals(other engine.Value) bool       { return false }
func (t *TarHandle) Less(other engine.Value) bool         { return false }
func (t *TarHandle) Greater(other engine.Value) bool      { return false }
func (t *TarHandle) LessEqual(other engine.Value) bool    { return false }
func (t *TarHandle) GreaterEqual(other engine.Value) bool { return false }
func (t *TarHandle) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (t *TarHandle) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (t *TarHandle) Add(other engine.Value) engine.Value  { return t }
func (t *TarHandle) Sub(other engine.Value) engine.Value  { return t }
func (t *TarHandle) Mul(other engine.Value) engine.Value  { return t }
func (t *TarHandle) Div(other engine.Value) engine.Value  { return t }
func (t *TarHandle) Mod(other engine.Value) engine.Value  { return t }
func (t *TarHandle) Negate() engine.Value                 { return t }

func (e *TarEntry) Type() engine.ValueType               { return engine.TypeObject }
func (e *TarEntry) IsNull() bool                         { return e == nil || e.header == nil }
func (e *TarEntry) Bool() bool                           { return e != nil && e.header != nil }
func (e *TarEntry) Int() int64                           { return 0 }
func (e *TarEntry) Float() float64                       { return 0.0 }
func (e *TarEntry) String() string                       { return "tar_entry:" + e.name }
func (e *TarEntry) Stringify() string                    { return e.String() }
func (e *TarEntry) Array() []engine.Value                { return nil }
func (e *TarEntry) Object() map[string]engine.Value      { return nil }
func (e *TarEntry) Len() int                             { return 0 }
func (e *TarEntry) Equals(other engine.Value) bool       { return false }
func (e *TarEntry) Less(other engine.Value) bool         { return false }
func (e *TarEntry) Greater(other engine.Value) bool      { return false }
func (e *TarEntry) LessEqual(other engine.Value) bool    { return false }
func (e *TarEntry) GreaterEqual(other engine.Value) bool { return false }
func (e *TarEntry) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (e *TarEntry) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (e *TarEntry) Add(other engine.Value) engine.Value  { return e }
func (e *TarEntry) Sub(other engine.Value) engine.Value  { return e }
func (e *TarEntry) Mul(other engine.Value) engine.Value  { return e }
func (e *TarEntry) Div(other engine.Value) engine.Value  { return e }
func (e *TarEntry) Mod(other engine.Value) engine.Value  { return e }
func (e *TarEntry) Negate() engine.Value                 { return e }

func builtinTarOpen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_open() expects 1 argument, got %d", len(args))
	}

	filename := args[0].String()

	file, err := os.Open(filename)
	if err != nil {
		return engine.NewNull(), nil
	}

	handle := &TarHandle{
		filename: filename,
		reader:   tar.NewReader(file),
		file:     file,
	}

	return handle, nil
}

func builtinTarRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_read() expects 1 argument, got %d", len(args))
	}

	handle, ok := args[0].(*TarHandle)
	if !ok || handle == nil || handle.reader == nil {
		return engine.NewNull(), nil
	}

	header, err := handle.reader.Next()
	if err != nil {
		if err == io.EOF {
			return engine.NewBool(false), nil
		}
		return engine.NewNull(), nil
	}

	entry := &TarEntry{
		header: header,
		name:   header.Name,
		size:   header.Size,
		isdir:  header.Typeflag == tar.TypeDir,
	}

	return entry, nil
}

func builtinTarEntryName(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_entry_name() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*TarEntry)
	if !ok || entry == nil || entry.header == nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(entry.header.Name), nil
}

func builtinTarEntrySize(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_entry_size() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*TarEntry)
	if !ok || entry == nil || entry.header == nil {
		return engine.NewNull(), nil
	}

	return engine.NewInt(entry.header.Size), nil
}

func builtinTarEntryIsdir(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_entry_isdir() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*TarEntry)
	if !ok || entry == nil || entry.header == nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(entry.header.Typeflag == tar.TypeDir), nil
}

func builtinTarEntryRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("tar_entry_read() expects 1-2 arguments, got %d", len(args))
	}

	handle, ok := args[0].(*TarHandle)
	if !ok || handle == nil || handle.reader == nil {
		return engine.NewNull(), nil
	}

	length := 8192
	if len(args) == 2 {
		length = int(args[1].Int())
	}

	buf := make([]byte, length)
	n, err := handle.reader.Read(buf)
	if err != nil && err != io.EOF {
		return engine.NewNull(), nil
	}

	return engine.NewString(string(buf[:n])), nil
}

func builtinTarEntryClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_entry_close() expects 1 argument, got %d", len(args))
	}

	entry, ok := args[0].(*TarEntry)
	if !ok || entry == nil {
		return engine.NewInt(0), nil
	}

	entry.header = nil
	entry.name = ""

	return engine.NewInt(0), nil
}

func builtinTarClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tar_close() expects 1 argument, got %d", len(args))
	}

	handle, ok := args[0].(*TarHandle)
	if !ok || handle == nil {
		return engine.NewInt(0), nil
	}

	if handle.file != nil {
		handle.file.Close()
		handle.reader = nil
	}

	return engine.NewInt(0), nil
}

func builtinTarCreate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("tar_create() expects 2 arguments, got %d", len(args))
	}

	filename := args[0].String()
	entriesArg := args[1]

	if entriesArg.Type() != engine.TypeArray {
		return nil, fmt.Errorf("tar_create() expects array of entries, got %s", entriesArg.Type())
	}

	entries := entriesArg.Array()

	file, err := os.Create(filename)
	if err != nil {
		return engine.NewNull(), nil
	}
	defer file.Close()

	tarWriter := tar.NewWriter(file)
	defer tarWriter.Close()

	for _, entry := range entries {
		if entry.Type() != engine.TypeObject {
			continue
		}

		obj := entry.Object()
		name, _ := obj["name"]
		content, _ := obj["content"]

		entryName := name.String()
		entryContent := content.String()

		header := &tar.Header{
			Name: entryName,
			Mode: 0644,
			Size: int64(len(entryContent)),
		}

		err := tarWriter.WriteHeader(header)
		if err != nil {
			continue
		}

		_, err = tarWriter.Write([]byte(entryContent))
		if err != nil {
			continue
		}
	}

	return engine.NewInt(1), nil
}
