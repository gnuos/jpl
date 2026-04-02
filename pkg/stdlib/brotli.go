package stdlib

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/gnuos/jpl/engine"
)

type BrotliValue struct {
	path      string
	file      *os.File
	reader    *brotli.Reader
	writer    *brotli.Writer
	writerBuf *bytes.Buffer
	isClosed  bool
}

func (b *BrotliValue) Type() engine.ValueType               { return engine.TypeObject }
func (b *BrotliValue) IsNull() bool                         { return false }
func (b *BrotliValue) Bool() bool                           { return b != nil && !b.isClosed }
func (b *BrotliValue) Int() int64                           { return 0 }
func (b *BrotliValue) Float() float64                       { return 0 }
func (b *BrotliValue) String() string                       { return "brotli:" + b.path }
func (b *BrotliValue) Stringify() string                    { return b.String() }
func (b *BrotliValue) Array() []engine.Value                { return nil }
func (b *BrotliValue) Object() map[string]engine.Value      { return nil }
func (b *BrotliValue) Len() int                             { return 0 }
func (b *BrotliValue) Equals(other engine.Value) bool       { return false }
func (b *BrotliValue) Less(other engine.Value) bool         { return false }
func (b *BrotliValue) Greater(other engine.Value) bool      { return false }
func (b *BrotliValue) LessEqual(other engine.Value) bool    { return false }
func (b *BrotliValue) GreaterEqual(other engine.Value) bool { return false }
func (b *BrotliValue) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (b *BrotliValue) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (b *BrotliValue) Add(other engine.Value) engine.Value  { return b }
func (b *BrotliValue) Sub(other engine.Value) engine.Value  { return b }
func (b *BrotliValue) Mul(other engine.Value) engine.Value  { return b }
func (b *BrotliValue) Div(other engine.Value) engine.Value  { return b }
func (b *BrotliValue) Mod(other engine.Value) engine.Value  { return b }
func (b *BrotliValue) Negate() engine.Value                 { return b }

var brotliWriterBuf bytes.Buffer

func RegisterBrotli(e *engine.Engine) {
	e.RegisterFunc("brotli_encode", builtinBrotliEncode)
	e.RegisterFunc("brotli_decode", builtinBrotliDecode)
	e.RegisterFunc("brotli_compress_file", builtinBrotliCompressFile)
	e.RegisterFunc("brotli_decompress_file", builtinBrotliDecompressFile)
	e.RegisterFunc("brotli_open", builtinBrotliOpen)
	e.RegisterFunc("brotli_read", builtinBrotliRead)
	e.RegisterFunc("brotli_write", builtinBrotliWrite)
	e.RegisterFunc("brotli_close", builtinBrotliClose)

	e.RegisterModule("brotli", map[string]engine.GoFunction{
		"encode":          builtinBrotliEncode,
		"decode":          builtinBrotliDecode,
		"compress_file":   builtinBrotliCompressFile,
		"decompress_file": builtinBrotliDecompressFile,
		"open":            builtinBrotliOpen,
		"read":            builtinBrotliRead,
		"write":           builtinBrotliWrite,
		"close":           builtinBrotliClose,
	})
}

func BrotliNames() []string {
	return []string{
		"brotli_encode", "brotli_decode",
		"brotli_compress_file", "brotli_decompress_file",
		"brotli_open", "brotli_read", "brotli_write", "brotli_close",
	}
}

func builtinBrotliEncode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("brotli_encode() expects 1 argument, got %d", len(args))
	}
	data := args[0].String()
	brotliWriterBuf.Reset()
	writer := brotli.NewWriter(&brotliWriterBuf)
	_, err := writer.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("brotli_encode() failed: %v", err)
	}
	writer.Close()
	return engine.NewString(brotliWriterBuf.String()), nil
}

func builtinBrotliDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("brotli_decode() expects 1 argument, got %d", len(args))
	}
	data := args[0].String()
	reader := brotli.NewReader(bytes.NewReader([]byte(data)))
	out, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("brotli_decode() failed: %v", err)
	}
	return engine.NewString(string(out)), nil
}

func builtinBrotliCompressFile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("brotli_compress_file() expects 2 arguments, got %d", len(args))
	}
	srcPath := args[0].String()
	destPath := args[1].String()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("brotli_compress_file() failed to open source: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("brotli_compress_file() failed to create dest: %v", err)
	}
	defer destFile.Close()

	writer := brotli.NewWriter(destFile)
	defer writer.Close()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		return nil, fmt.Errorf("brotli_compress_file() failed: %v", err)
	}
	return engine.NewInt(1), nil
}

func builtinBrotliDecompressFile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("brotli_decompress_file() expects 2 arguments, got %d", len(args))
	}
	srcPath := args[0].String()
	destPath := args[1].String()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("brotli_decompress_file() failed to open source: %v", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("brotli_decompress_file() failed to create dest: %v", err)
	}
	defer destFile.Close()

	reader := brotli.NewReader(srcFile)
	_, err = io.Copy(destFile, reader)
	if err != nil {
		return nil, fmt.Errorf("brotli_decompress_file() failed: %v", err)
	}
	return engine.NewInt(1), nil
}

func builtinBrotliOpen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("brotli_open() expects 1-2 arguments, got %d", len(args))
	}
	path := args[0].String()
	mode := "r"
	if len(args) >= 2 {
		mode = args[1].String()
	}

	var file *os.File
	var err error
	var reader *brotli.Reader
	var writer *brotli.Writer
	var writerBuf *bytes.Buffer

	switch mode {
	case "r", "rb":
		file, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("brotli_open() failed to open: %v", err)
		}
		reader = brotli.NewReader(file)
	case "w", "wb":
		file, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("brotli_open() failed to create: %v", err)
		}
		writerBuf = &bytes.Buffer{}
		writer = brotli.NewWriter(writerBuf)
	case "a", "ab":
		file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("brotli_open() failed to open: %v", err)
		}
		writerBuf = &bytes.Buffer{}
		writer = brotli.NewWriter(writerBuf)
	default:
		return nil, fmt.Errorf("brotli_open() invalid mode: %s (use 'r', 'w', or 'a')", mode)
	}

	return &BrotliValue{
		path:      path,
		file:      file,
		reader:    reader,
		writer:    writer,
		writerBuf: writerBuf,
	}, nil
}

func builtinBrotliRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("brotli_read() expects 1-2 arguments, got %d", len(args))
	}
	handle, ok := args[0].(*BrotliValue)
	if !ok {
		return nil, fmt.Errorf("brotli_read() expects brotli handle, got %s", args[0].Type())
	}
	if handle.reader == nil {
		return nil, fmt.Errorf("brotli_read() handle not opened for reading")
	}
	if handle.isClosed {
		return nil, fmt.Errorf("brotli_read() handle is closed")
	}

	length := 4096
	if len(args) >= 2 {
		length = int(args[1].Int())
		if length <= 0 {
			length = 4096
		}
	}

	buf := make([]byte, length)
	n, err := handle.reader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("brotli_read() failed: %v", err)
	}
	return engine.NewString(string(buf[:n])), nil
}

func builtinBrotliWrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("brotli_write() expects at least 2 arguments, got %d", len(args))
	}
	handle, ok := args[0].(*BrotliValue)
	if !ok {
		return nil, fmt.Errorf("brotli_write() expects brotli handle, got %s", args[0].Type())
	}
	if handle.writer == nil {
		return nil, fmt.Errorf("brotli_write() handle not opened for writing")
	}
	if handle.isClosed {
		return nil, fmt.Errorf("brotli_write() handle is closed")
	}

	data := args[1].String()
	n, err := handle.writer.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("brotli_write() failed: %v", err)
	}
	return engine.NewInt(int64(n)), nil
}

func builtinBrotliClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("brotli_close() expects 1 argument, got %d", len(args))
	}
	handle, ok := args[0].(*BrotliValue)
	if !ok {
		return nil, fmt.Errorf("brotli_close() expects brotli handle, got %s", args[0].Type())
	}
	if handle.isClosed {
		return engine.NewNull(), nil
	}

	if handle.writer != nil {
		handle.writer.Close()
		if handle.writerBuf != nil && handle.file != nil {
			handle.file.Write(handle.writerBuf.Bytes())
		}
	}
	if handle.file != nil {
		handle.file.Close()
	}
	handle.isClosed = true
	return engine.NewNull(), nil
}
