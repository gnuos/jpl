package engine

import (
	"os"
	"strings"
	"testing"
)

func TestStreamType(t *testing.T) {
	stdin := NewStdinStream()
	if stdin.Type() != TypeStream {
		t.Errorf("expected TypeStream, got %v", stdin.Type())
	}
}

func TestStreamStringify(t *testing.T) {
	tests := []struct {
		stream   Value
		expected string
	}{
		{NewStdinStream(), "stream:stdin"},
		{NewStdoutStream(), "stream:stdout"},
		{NewStderrStream(), "stream:stderr"},
	}

	for _, tt := range tests {
		if got := tt.stream.Stringify(); got != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, got)
		}
	}
}

func TestStreamEquals(t *testing.T) {
	stdin1 := NewStdinStream()
	stdin2 := NewStdinStream()

	if !stdin1.Equals(stdin2) {
		t.Error("stdin streams should be equal")
	}

	stdout := NewStdoutStream()
	if stdin1.Equals(stdout) {
		t.Error("stdin and stdout should not be equal")
	}

	if stdin1.Equals(NewInt(0)) {
		t.Error("stream should not equal int")
	}
}

func TestStreamBool(t *testing.T) {
	stdin := NewStdinStream()
	if !stdin.Bool() {
		t.Error("stream should be true when not closed")
	}

	stdin.(*streamValue).closed = true
	if stdin.Bool() {
		t.Error("stream should be false when closed")
	}
}

func TestStreamRead(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_test.txt"
	content := "test content\n"

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	file.WriteString(content)
	file.Close()
	defer os.Remove(tmpFile)

	stream := NewFileStream(tmpFile, StreamRead)
	if stream.Type() != TypeStream {
		t.Errorf("expected TypeStream, got %v", stream.Type())
	}

	buf := make([]byte, 100)
	n, err := stream.(*streamValue).Read(buf)
	if err != nil {
		t.Errorf("read error: %v", err)
	}

	result := string(buf[:n])
	if result != content {
		t.Errorf("expected %q, got %q", content, result)
	}

	stream.(*streamValue).Close()
}

func TestStreamWrite(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_write_test.txt"
	content := "hello world"

	stream := NewFileStream(tmpFile, StreamWrite)
	n, err := stream.(*streamValue).Write([]byte(content))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n != len(content) {
		t.Errorf("expected write %d bytes, got %d", len(content), n)
	}

	stream.(*streamValue).Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}

	os.Remove(tmpFile)
}

func TestStreamClose(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_close_test.txt"

	stream := NewFileStream(tmpFile, StreamWrite)
	stream.(*streamValue).Write([]byte("test"))
	stream.(*streamValue).Close()

	err := stream.(*streamValue).Close()
	if err != nil {
		t.Error("second close should not return error")
	}

	os.Remove(tmpFile)
}

func TestStreamClosed(t *testing.T) {
	stream := NewStdinStream()
	if stream.(*streamValue).IsClosed() {
		t.Error("stream should not be closed initially")
	}

	stream.(*streamValue).Close()
	if !stream.(*streamValue).IsClosed() {
		t.Error("stream should be closed after Close()")
	}
}

func TestStreamMode(t *testing.T) {
	stdin := NewStdinStream()
	if stdin.(*streamValue).Mode() != StreamRead {
		t.Error("stdin should be in read mode")
	}

	stdout := NewStdoutStream()
	if stdout.(*streamValue).Mode() != StreamWrite {
		t.Error("stdout should be in write mode")
	}

	stderr := NewStderrStream()
	if stderr.(*streamValue).Mode() != StreamWrite {
		t.Error("stderr should be in write mode")
	}
}

func TestStreamReadWrite(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_rw_test.txt"
	os.Remove(tmpFile)
	defer os.Remove(tmpFile)

	stream := NewFileStream(tmpFile, StreamReadWrite)
	if stream.Type() == TypeError {
		t.Skip("file creation not supported in this environment")
	}

	_, err := stream.(*streamValue).Write([]byte("test"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	stream.(*streamValue).Close()

	stream = NewFileStream(tmpFile, StreamRead)
	buf := make([]byte, 100)
	n, err := stream.(*streamValue).Read(buf)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if string(buf[:n]) != "test" {
		t.Errorf("expected 'test', got %q", string(buf[:n]))
	}

	stream.(*streamValue).Close()
}

func TestStreamNonReadable(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_nr_test.txt"

	stream := NewFileStream(tmpFile, StreamWrite)
	if stream.(*streamValue).IsReadable() {
		t.Error("write-only stream should not be readable")
	}
	if !stream.(*streamValue).IsWritable() {
		t.Error("write-only stream should be writable")
	}
	stream.(*streamValue).Close()
	os.Remove(tmpFile)
}

func TestStreamNonWritable(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_nw_test.txt"

	file, _ := os.Create(tmpFile)
	file.WriteString("test")
	file.Close()
	defer os.Remove(tmpFile)

	stream := NewFileStream(tmpFile, StreamRead)
	if stream.(*streamValue).IsWritable() {
		t.Error("read-only stream should not be writable")
	}
	if !stream.(*streamValue).IsReadable() {
		t.Error("read-only stream should be readable")
	}
	stream.(*streamValue).Close()
}

func TestStreamReadAfterClose(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_rac_test.txt"

	file, _ := os.Create(tmpFile)
	file.WriteString("test")
	file.Close()
	defer os.Remove(tmpFile)

	stream := NewFileStream(tmpFile, StreamRead)
	stream.(*streamValue).Close()

	buf := make([]byte, 10)
	_, err := stream.(*streamValue).Read(buf)
	if err == nil {
		t.Error("should return error after close")
	}
}

func TestStreamWrapStream(t *testing.T) {
	stdin := NewStdinStream()
	wrapper := WrapStream(stdin)

	if !wrapper.IsStream() {
		t.Error("wrapper should recognize stream")
	}

	nullWrapper := WrapStream(nil)
	if nullWrapper.IsStream() {
		t.Error("nil wrapper should not be stream")
	}

	intValue := NewInt(42)
	intWrapper := WrapStream(intValue)
	if intWrapper.IsStream() {
		t.Error("int wrapper should not be stream")
	}
}

func TestStreamValueMethods(t *testing.T) {
	stream := NewStdinStream()

	if stream.Int() != 0 {
		t.Error("Int() should return 0")
	}
	if stream.Float() != 0.0 {
		t.Error("Float() should return 0.0")
	}
	if stream.Array() != nil {
		t.Error("Array() should return nil")
	}
	if stream.Object() != nil {
		t.Error("Object() should return nil")
	}
	if stream.Len() != 0 {
		t.Error("Len() should return 0")
	}
}

func TestStreamAdd(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Add(NewInt(1))

	if result.Type() != TypeStream {
		t.Error("Add should return self")
	}
}

func TestStreamLess(t *testing.T) {
	stream := NewStdinStream()
	if stream.Less(NewInt(1)) {
		t.Error("stream should never be less")
	}
	if !stream.GreaterEqual(NewInt(1)) {
		t.Error("stream should always be greater or equal")
	}
}

func TestStreamBufRead(t *testing.T) {
	r := strings.NewReader("hello world")
	stdin := &streamValue{
		mode:   StreamRead,
		reader: r,
		path:   "buffer",
	}

	buf := make([]byte, 5)
	n, err := stdin.Read(buf)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got %s", string(buf))
	}
}

func TestStreamBufWrite(t *testing.T) {
	var buf strings.Builder
	stdin := &streamValue{
		mode:   StreamWrite,
		writer: &buf,
		path:   "buffer",
	}

	n, err := stdin.Write([]byte("test"))
	if err != nil {
		t.Errorf("write error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes, got %d", n)
	}
	if buf.String() != "test" {
		t.Errorf("expected 'test', got %s", buf.String())
	}
}

func TestStreamModeString(t *testing.T) {
	tests := []struct {
		mode     StreamMode
		expected string
	}{
		{StreamRead, "r"},
		{StreamWrite, "w"},
		{StreamReadWrite, "rw"},
		{StreamMode(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, got)
		}
	}
}

func TestStreamReadWriteMode(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_rwm_test.txt"
	stream := NewFileStream(tmpFile, StreamReadWrite)

	if !stream.(*streamValue).IsReadable() {
		t.Error("read-write stream should be readable")
	}
	if !stream.(*streamValue).IsWritable() {
		t.Error("read-write stream should be writable")
	}

	stream.(*streamValue).Close()
	os.Remove(tmpFile)
}

func TestStreamInvalidFile(t *testing.T) {
	stream := NewFileStream("/nonexistent/path/to/file/that/does/not/exist.txt", StreamRead)

	if stream.Type() == TypeStream {
		t.Error("should return runtime error for invalid file")
	}
}

func TestStreamStdout(t *testing.T) {
	stdout := NewStdoutStream()
	if stdout.Type() != TypeStream {
		t.Errorf("expected TypeStream, got %v", stdout.Type())
	}
	if stdout.String() != "stdout" {
		t.Errorf("expected 'stdout', got %s", stdout.String())
	}
}

func TestStreamStderr(t *testing.T) {
	stderr := NewStderrStream()
	if stderr.Type() != TypeStream {
		t.Errorf("expected TypeStream, got %v", stderr.Type())
	}
	if stderr.String() != "stderr" {
		t.Errorf("expected 'stderr', got %s", stderr.String())
	}
}

func TestStreamObjectReturn(t *testing.T) {
	stream := NewStdinStream()
	obj := stream.Object()
	if obj != nil {
		t.Error("Object() should return nil for stream")
	}
}

func TestStreamArrayReturn(t *testing.T) {
	stream := NewStdinStream()
	arr := stream.Array()
	if arr != nil {
		t.Error("Array() should return nil for stream")
	}
}

func TestStreamToBigInt(t *testing.T) {
	stream := NewStdinStream()
	bigInt := stream.ToBigInt()
	if bigInt.Int() != 0 {
		t.Error("ToBigInt() should return Int(0)")
	}
}

func TestStreamToBigDecimal(t *testing.T) {
	stream := NewStdinStream()
	bigDec := stream.ToBigDecimal()
	if bigDec.Float() != 0.0 {
		t.Error("ToBigDecimal() should return Float(0.0)")
	}
}

func TestStreamNegate(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Negate()
	if result.Type() != TypeStream {
		t.Error("Negate should return self")
	}
}

func TestStreamMul(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Mul(NewInt(2))
	if result.Type() != TypeStream {
		t.Error("Mul should return self")
	}
}

func TestStreamDiv(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Div(NewInt(2))
	if result.Type() != TypeStream {
		t.Error("Div should return self")
	}
}

func TestStreamMod(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Mod(NewInt(2))
	if result.Type() != TypeStream {
		t.Error("Mod should return self")
	}
}

func TestStreamSub(t *testing.T) {
	stream := NewStdinStream()
	result := stream.Sub(NewInt(2))
	if result.Type() != TypeStream {
		t.Error("Sub should return self")
	}
}

func TestStreamWriteOnlyRead(t *testing.T) {
	tmpFile := "/tmp/jpl_stream_wor_test.txt"
	stream := NewFileStream(tmpFile, StreamWrite)

	buf := make([]byte, 10)
	n, err := stream.(*streamValue).Read(buf)

	if n != 0 {
		t.Error("read-only stream should return 0 bytes")
	}
	if err == nil {
		t.Error("should return error when reading from write-only stream")
	}

	stream.(*streamValue).Close()
	os.Remove(tmpFile)
}
