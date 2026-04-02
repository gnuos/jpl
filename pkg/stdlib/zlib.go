package stdlib

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/gnuos/jpl/engine"
)

func RegisterZlib(e *engine.Engine) {
	e.RegisterFunc("zlib_encode", builtinZlibEncode)
	e.RegisterFunc("zlib_decode", builtinZlibDecode)
	e.RegisterFunc("deflate", builtinDeflate)
	e.RegisterFunc("inflate", builtinInflate)

	e.RegisterModule("zlib", map[string]engine.GoFunction{
		"zlib_encode": builtinZlibEncode,
		"zlib_decode": builtinZlibDecode,
		"deflate":     builtinDeflate,
		"inflate":     builtinInflate,
	})
}

func ZlibNames() []string {
	return []string{
		"zlib_encode",
		"zlib_decode",
		"deflate",
		"inflate",
	}
}

func builtinZlibEncode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zlib_encode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()

	var buf bytes.Buffer
	writer := zlib.NewWriter(&buf)
	_, err := writer.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("zlib_encode() compression failed: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("zlib_encode() compression failed: %v", err)
	}

	return engine.NewString(buf.String()), nil
}

func builtinZlibDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("zlib_decode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	reader, err := zlib.NewReader(bytes.NewReader([]byte(data)))
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

func builtinDeflate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("deflate() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()

	var buf bytes.Buffer
	writer, err := zlib.NewWriterLevel(&buf, zlib.DefaultCompression)
	if err != nil {
		return nil, fmt.Errorf("deflate() failed: %v", err)
	}
	_, err = writer.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("deflate() failed: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("deflate() failed: %v", err)
	}

	return engine.NewString(buf.String()), nil
}

func builtinInflate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("inflate() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	reader, err := zlib.NewReader(bytes.NewReader([]byte(data)))
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
