package stdlib

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/gnuos/jpl/engine"
)

func RegisterGzip(e *engine.Engine) {
	e.RegisterFunc("gzencode", builtinGzencode)
	e.RegisterFunc("gzdecode", builtinGzdecode)
	e.RegisterFunc("gzfile", builtinGzfile)
	e.RegisterFunc("writegzfile", builtinWriteGzfile)
	e.RegisterFunc("gzopen", builtinGzopen)
	e.RegisterFunc("gzread", builtinGzread)
	e.RegisterFunc("gzwrite", builtinGzwrite)
	e.RegisterFunc("gzclose", builtinGzclose)
	e.RegisterFunc("gzgets", builtinGzgets)
	e.RegisterFunc("gzeof", builtinGzeof)

	e.RegisterModule("gzip", map[string]engine.GoFunction{
		"gzencode":    builtinGzencode,
		"gzdecode":    builtinGzdecode,
		"gzfile":      builtinGzfile,
		"writegzfile": builtinWriteGzfile,
		"gzopen":      builtinGzopen,
		"gzread":      builtinGzread,
		"gzwrite":     builtinGzwrite,
		"gzclose":     builtinGzclose,
		"gzgets":      builtinGzgets,
		"gzeof":       builtinGzeof,
	})
}

func GzipNames() []string {
	return []string{
		"gzencode",
		"gzdecode",
		"gzfile",
		"writegzfile",
		"gzopen",
		"gzread",
		"gzwrite",
		"gzclose",
		"gzgets",
		"gzeof",
	}
}

type GzipValue struct {
	file   *os.File
	reader *gzip.Reader
	writer *gzip.Writer
	mode   string
	path   string
	eof    bool
}

func (g *GzipValue) Type() engine.ValueType { return engine.TypeObject }
func (g *GzipValue) IsNull() bool           { return false }
func (g *GzipValue) Bool() bool             { return g != nil && g.file != nil }
func (g *GzipValue) Int() int64             { return 0 }
func (g *GzipValue) Float() float64         { return 0.0 }
func (g *GzipValue) String() string         { return "gzip:" + g.path }
func (g *GzipValue) Stringify() string      { return g.String() }
func (g *GzipValue) Array() []engine.Value  { return nil }
func (g *GzipValue) Object() map[string]engine.Value {
	return map[string]engine.Value{
		"path": engine.NewString(g.path),
		"mode": engine.NewString(g.mode),
	}
}
func (g *GzipValue) Len() int                             { return 0 }
func (g *GzipValue) Equals(other engine.Value) bool       { return false }
func (g *GzipValue) Less(other engine.Value) bool         { return false }
func (g *GzipValue) Greater(other engine.Value) bool      { return false }
func (g *GzipValue) LessEqual(other engine.Value) bool    { return false }
func (g *GzipValue) GreaterEqual(other engine.Value) bool { return false }
func (g *GzipValue) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (g *GzipValue) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (g *GzipValue) Add(other engine.Value) engine.Value  { return g }
func (g *GzipValue) Sub(other engine.Value) engine.Value  { return g }
func (g *GzipValue) Mul(other engine.Value) engine.Value  { return g }
func (g *GzipValue) Div(other engine.Value) engine.Value  { return g }
func (g *GzipValue) Mod(other engine.Value) engine.Value  { return g }
func (g *GzipValue) Negate() engine.Value                 { return g }

func builtinGzencode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gzencode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()

	writerBuf.Reset()
	writer := gzip.NewWriter(&writerBuf)
	_, err := writer.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("gzencode() compression failed: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("gzencode() compression failed: %v", err)
	}

	return engine.NewString(writerBuf.String()), nil
}

var writerBuf bytes.Buffer

func builtinGzdecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gzdecode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	reader, err := gzip.NewReader(bytes.NewReader([]byte(data)))
	if err != nil {
		return engine.NewNull(), nil
	}
	defer reader.Close()

	result, err := io.ReadAll(reader)
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(string(result)), nil
}

func builtinGzfile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gzfile() expects 1 argument, got %d", len(args))
	}

	filename := args[0].String()

	file, err := os.Open(filename)
	if err != nil {
		return engine.NewNull(), nil
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return engine.NewNull(), nil
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return engine.NewNull(), nil
	}

	lines := splitLines(string(content))
	return engine.NewArray(lines), nil
}

func builtinWriteGzfile(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("writegzfile() expects 2 arguments, got %d", len(args))
	}

	filename := args[0].String()
	data := args[1].String()

	file, err := os.Create(filename)
	if err != nil {
		return engine.NewNull(), nil
	}
	defer file.Close()

	writer := gzip.NewWriter(file)
	defer writer.Close()

	_, err = writer.Write([]byte(data))
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewInt(1), nil
}

func builtinGzopen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("gzopen() expects 2 arguments, got %d", len(args))
	}

	filename := args[0].String()
	mode := args[1].String()

	var file *os.File
	var err error
	var gz *GzipValue

	if mode == "r" || mode == "rb" {
		file, err = os.Open(filename)
		if err != nil {
			return engine.NewNull(), nil
		}
		reader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return engine.NewNull(), nil
		}
		gz = &GzipValue{
			file:   file,
			reader: reader,
			mode:   "r",
			path:   filename,
		}
		return gz, nil
	} else if mode == "w" || mode == "wb" {
		file, err = os.Create(filename)
		if err != nil {
			return engine.NewNull(), nil
		}
		writer := gzip.NewWriter(file)
		gz = &GzipValue{
			file:   file,
			writer: writer,
			mode:   "w",
			path:   filename,
		}
		return gz, nil
	} else if mode == "a" {
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return engine.NewNull(), nil
		}
		writer := gzip.NewWriter(file)
		gz = &GzipValue{
			file:   file,
			writer: writer,
			mode:   "a",
			path:   filename,
		}
		return gz, nil
	}

	return engine.NewNull(), nil
}

func builtinGzread(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("gzread() expects 1-2 arguments, got %d", len(args))
	}

	gz, ok := args[0].(*GzipValue)
	if !ok {
		return engine.NewNull(), nil
	}

	length := 8192
	if len(args) == 2 {
		length = int(args[1].Int())
	}

	if gz.reader != nil {
		buf := make([]byte, length)
		n, err := gz.reader.Read(buf)
		if err != nil && err != io.EOF {
			return engine.NewNull(), nil
		}
		if n == 0 && err == io.EOF {
			gz.eof = true
		}
		return engine.NewString(string(buf[:n])), nil
	}

	return engine.NewNull(), nil
}

func builtinGzwrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("gzwrite() expects 2 arguments, got %d", len(args))
	}

	gz, ok := args[0].(*GzipValue)
	if !ok {
		return engine.NewInt(0), nil
	}
	if gz.writer == nil {
		return engine.NewInt(0), nil
	}

	data := args[1].String()
	n, err := gz.writer.Write([]byte(data))
	if err != nil {
		return engine.NewInt(0), nil
	}

	return engine.NewInt(int64(n)), nil
}

func builtinGzclose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gzclose() expects 1 argument, got %d", len(args))
	}

	gz, ok := args[0].(*GzipValue)
	if !ok {
		return engine.NewInt(0), nil
	}

	var err error
	if gz.writer != nil {
		err = gz.writer.Close()
	}
	if gz.reader != nil {
		gz.reader.Close()
	}
	if gz.file != nil {
		err = gz.file.Close()
	}

	if err != nil {
		return engine.NewInt(0), nil
	}
	return engine.NewInt(0), nil
}

func builtinGzgets(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("gzgets() expects 1-2 arguments, got %d", len(args))
	}

	gz, ok := args[0].(*GzipValue)
	if !ok || gz.reader == nil {
		return engine.NewNull(), nil
	}

	length := 1024
	if len(args) == 2 {
		length = int(args[1].Int())
	}

	buf := make([]byte, length)
	n, err := gz.reader.Read(buf)
	if err != nil && err != io.EOF {
		return engine.NewNull(), nil
	}

	line := string(buf[:n])
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	if n == 0 && err == io.EOF {
		gz.eof = true
	}

	return engine.NewString(line), nil
}

func builtinGzeof(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("gzeof() expects 1 argument, got %d", len(args))
	}

	gz, ok := args[0].(*GzipValue)
	if !ok || gz.reader == nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(gz.eof), nil
}

func splitLines(s string) []engine.Value {
	var lines []engine.Value
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, engine.NewString(line))
			start = i + 1
		}
	}
	if start < len(s) {
		line := s[start:]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		lines = append(lines, engine.NewString(line))
	}
	return lines
}

// GzipSigs returns function signatures for REPL :doc command.
func GzipSigs() map[string]string {
	return map[string]string{
		"gzip_compress":   "gzip_compress(data) → string  — Compress data with gzip",
		"gzip_decompress": "gzip_decompress(data) → string  — Decompress gzip data",
	}
}
