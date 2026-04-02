package engine

import (
	"bytes"
	"io"
	"os"
)

type StreamMode int

const (
	StreamRead StreamMode = iota
	StreamWrite
	StreamReadWrite
)

func (m StreamMode) String() string {
	switch m {
	case StreamRead:
		return "r"
	case StreamWrite:
		return "w"
	case StreamReadWrite:
		return "rw"
	default:
		return "unknown"
	}
}

type streamValue struct {
	mode   StreamMode
	reader io.Reader
	writer io.Writer
	closer io.Closer
	path   string
	closed bool
}

func (v *streamValue) Type() ValueType          { return TypeStream }
func (v *streamValue) IsNull() bool             { return false }
func (v *streamValue) Bool() bool               { return !v.closed }
func (v *streamValue) Int() int64               { return 0 }
func (v *streamValue) Float() float64           { return 0.0 }
func (v *streamValue) String() string           { return v.path }
func (v *streamValue) Array() []Value           { return nil }
func (v *streamValue) Object() map[string]Value { return nil }
func (v *streamValue) Len() int                 { return 0 }
func (v *streamValue) ToBigInt() Value          { return NewInt(0) }
func (v *streamValue) ToBigDecimal() Value      { return NewFloat(0.0) }

func (v *streamValue) Equals(other Value) bool {
	if other == nil {
		return false
	}
	if other.Type() != TypeStream {
		return false
	}
	otherStream, ok := other.(*streamValue)
	if !ok {
		return false
	}
	return v.path == otherStream.path && v.mode == otherStream.mode
}

func (v *streamValue) Stringify() string {
	return "stream:" + v.path
}

func (v *streamValue) Add(other Value) Value { return v }
func (v *streamValue) Sub(other Value) Value { return v }
func (v *streamValue) Mul(other Value) Value { return v }
func (v *streamValue) Div(other Value) Value { return v }
func (v *streamValue) Mod(other Value) Value { return v }
func (v *streamValue) Negate() Value         { return v }

func (v *streamValue) Less(other Value) bool         { return false }
func (v *streamValue) Greater(other Value) bool      { return false }
func (v *streamValue) LessEqual(other Value) bool    { return false }
func (v *streamValue) GreaterEqual(other Value) bool { return true }

func (v *streamValue) Read(p []byte) (int, error) {
	if v.closed {
		return 0, io.EOF
	}
	if v.reader == nil {
		return 0, io.EOF
	}
	return v.reader.Read(p)
}

func (v *streamValue) Write(p []byte) (int, error) {
	if v.closed {
		return 0, os.ErrClosed
	}
	if v.writer == nil {
		return 0, os.ErrInvalid
	}
	return v.writer.Write(p)
}

func (v *streamValue) Close() error {
	if v.closed {
		return nil
	}
	v.closed = true
	if v.closer != nil {
		return v.closer.Close()
	}
	return nil
}

func (v *streamValue) IsReadable() bool {
	if v.closed {
		return false
	}
	return v.mode == StreamRead || v.mode == StreamReadWrite
}

func (v *streamValue) IsWritable() bool {
	if v.closed {
		return false
	}
	return v.mode == StreamWrite || v.mode == StreamReadWrite
}

func (v *streamValue) IsClosed() bool {
	return v.closed
}

func (v *streamValue) Mode() StreamMode {
	return v.mode
}

// GetReader 返回底层 reader（用于 fseek 等操作）
func (v *streamValue) GetReader() io.Reader {
	return v.reader
}

// GetWriter 返回底层 writer
func (v *streamValue) GetWriter() io.Writer {
	return v.writer
}

// Seek 移动文件指针，返回新位置。仅文件流支持。
func (v *streamValue) Seek(offset int64, whence int) (int64, error) {
	if v.closed {
		return 0, os.ErrClosed
	}
	seeker, ok := v.reader.(io.Seeker)
	if !ok {
		return 0, os.ErrInvalid
	}
	return seeker.Seek(offset, whence)
}

// Truncate 截断文件到指定大小。仅文件流支持。
func (v *streamValue) Truncate(size int64) error {
	if v.closed {
		return os.ErrClosed
	}
	f, ok := v.reader.(*os.File)
	if !ok {
		return os.ErrInvalid
	}
	return f.Truncate(size)
}

func NewStdinStream() Value {
	return &streamValue{
		mode:   StreamRead,
		reader: os.Stdin,
		closer: nil,
		path:   "stdin",
		closed: false,
	}
}

func NewStdoutStream() Value {
	return &streamValue{
		mode:   StreamWrite,
		writer: os.Stdout,
		closer: nil,
		path:   "stdout",
		closed: false,
	}
}

func NewStderrStream() Value {
	return &streamValue{
		mode:   StreamWrite,
		writer: os.Stderr,
		closer: nil,
		path:   "stderr",
		closed: false,
	}
}

func NewFileStream(path string, mode StreamMode) Value {
	var file *os.File
	var err error

	switch mode {
	case StreamRead:
		file, err = os.Open(path)
	case StreamWrite:
		file, err = os.Create(path)
	case StreamReadWrite:
		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	default:
		return newRuntimeError("invalid stream mode")
	}

	if err != nil {
		return newRuntimeError("cannot open file: " + err.Error())
	}

	return &streamValue{
		mode:   mode,
		reader: file,
		writer: file,
		closer: file,
		path:   path,
		closed: false,
	}
}

func NewBufferStream(buf []byte) Value {
	var r *bytes.Reader
	if buf == nil {
		buf = []byte{}
	}
	r = bytes.NewReader(buf)

	bufWriter := &bytes.Buffer{}
	if buf != nil {
		bufWriter.Write(buf)
	}

	return &streamValue{
		mode:   StreamReadWrite,
		reader: r,
		writer: bufWriter,
		closer: nil,
		path:   "buffer",
		closed: false,
	}
}

func StreamMeta(v Value) map[string]Value {
	sv, ok := v.(*streamValue)
	if !ok {
		return nil
	}

	meta := make(map[string]Value)
	meta["uri"] = NewString(sv.path)
	meta["mode"] = NewString(sv.mode.String())
	meta["is_readable"] = NewBool(sv.IsReadable())
	meta["is_writable"] = NewBool(sv.IsWritable())
	meta["is_closed"] = NewBool(sv.closed)
	meta["stream_type"] = NewString("stream")

	if sv.path == "stdin" || sv.path == "stdout" || sv.path == "stderr" {
		meta["stream_type"] = NewString("std")
	} else if sv.path == "buffer" {
		meta["stream_type"] = NewString("memory")
	} else {
		meta["stream_type"] = NewString("file")
	}

	meta["timetype"] = NewString("stream")
	meta["unparsed_url"] = NewString(sv.path)
	meta["wrapper_type"] = NewString("STDIO")

	return meta
}

type streamWrapper struct {
	stream *streamValue
}

func WrapStream(v Value) *streamWrapper {
	if sv, ok := v.(*streamValue); ok {
		return &streamWrapper{stream: sv}
	}
	return nil
}

func (w *streamWrapper) IsStream() bool {
	return w != nil && w.stream != nil
}

func (w *streamWrapper) Read(p []byte) (int, error) {
	if !w.IsStream() {
		return 0, os.ErrInvalid
	}
	return w.stream.Read(p)
}

func (w *streamWrapper) Write(p []byte) (int, error) {
	if !w.IsStream() {
		return 0, os.ErrInvalid
	}
	return w.stream.Write(p)
}

func (w *streamWrapper) Close() error {
	if !w.IsStream() {
		return os.ErrInvalid
	}
	return w.stream.Close()
}

type StreamValue = streamValue

func IsStream(v Value) bool {
	_, ok := v.(*streamValue)
	return ok
}

func ToStreamValue(v Value) *streamValue {
	if sv, ok := v.(*streamValue); ok {
		return sv
	}
	return nil
}
