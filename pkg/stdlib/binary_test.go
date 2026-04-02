package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestPackBasic 测试基础 pack 功能
func TestPackBasic(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterBinary(eng)

	tests := []struct {
		name     string
		format   string
		args     []int64
		expected []byte
	}{
		{"uint8 (C)", "C", []int64{255}, []byte{255}},
		{"uint16 big (S)", "S", []int64{258}, []byte{0x01, 0x02}},    // 258 = 0x0102
		{"uint16 little (s)", "s", []int64{258}, []byte{0x02, 0x01}}, // 258 = 0x0201
		{"uint32 big (N)", "N", []int64{16909060}, []byte{0x01, 0x02, 0x03, 0x04}},
		{"uint32 little (V)", "V", []int64{67305985}, []byte{0x01, 0x02, 0x03, 0x04}},
		{"uint64 big (Q)", "Q", []int64{72623859790382856}, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []engine.Value{engine.NewString(tt.format)}
			for _, v := range tt.args {
				args = append(args, engine.NewInt(v))
			}

			result, err := builtinPack(nil, args)
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}

			arr := result.Array()
			if len(arr) != len(tt.expected) {
				t.Fatalf("pack() returned %d bytes, expected %d", len(arr), len(tt.expected))
			}

			for i, expected := range tt.expected {
				if arr[i].Int() != int64(expected) {
					t.Errorf("pack() byte[%d] = 0x%02x, expected 0x%02x", i, arr[i].Int(), expected)
				}
			}
		})
	}
}

// TestPackMultiValue 测试多值 pack
func TestPackMultiValue(t *testing.T) {
	args := []engine.Value{
		engine.NewString("NNS"),
		engine.NewInt(1),
		engine.NewInt(2),
		engine.NewInt(258),
	}

	result, err := builtinPack(nil, args)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	arr := result.Array()
	expected := []byte{0, 0, 0, 1, 0, 0, 0, 2, 0x01, 0x02}

	if len(arr) != len(expected) {
		t.Fatalf("pack() returned %d bytes, expected %d", len(arr), len(expected))
	}

	for i, expectedByte := range expected {
		if arr[i].Int() != int64(expectedByte) {
			t.Errorf("pack() byte[%d] = 0x%02x, expected 0x%02x", i, arr[i].Int(), expectedByte)
		}
	}
}

// TestPackString 测试字符串 pack
func TestPackString(t *testing.T) {
	// 测试 'a' 格式（空填充字符串）
	args := []engine.Value{
		engine.NewString("a"),
		engine.NewString("hello"),
	}

	result, err := builtinPack(nil, args)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	arr := result.Array()
	expected := []byte("hello")

	if len(arr) != len(expected) {
		t.Fatalf("pack() returned %d bytes, expected %d", len(arr), len(expected))
	}
}

// TestPackZeroTerm 测试零结尾字符串
func TestPackZeroTerm(t *testing.T) {
	args := []engine.Value{
		engine.NewString("Z"),
		engine.NewString("hello"),
	}

	result, err := builtinPack(nil, args)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	arr := result.Array()
	expected := []byte("hello\x00")

	if len(arr) != len(expected) {
		t.Fatalf("pack() returned %d bytes, expected %d", len(arr), len(expected))
	}
}

// TestPackPadding 测试填充字节
func TestPackPadding(t *testing.T) {
	args := []engine.Value{
		engine.NewString("CxC"),
		engine.NewInt(255),
		engine.NewInt(128),
	}

	result, err := builtinPack(nil, args)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	arr := result.Array()
	expected := []byte{255, 0, 128}

	if len(arr) != len(expected) {
		t.Fatalf("pack() returned %d bytes, expected %d", len(arr), len(expected))
	}

	for i, expectedByte := range expected {
		if arr[i].Int() != int64(expectedByte) {
			t.Errorf("pack() byte[%d] = 0x%02x, expected 0x%02x", i, arr[i].Int(), expectedByte)
		}
	}
}

// TestUnpackBasic 测试基础 unpack 功能
func TestUnpackBasic(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     []byte
		expected int64
	}{
		{"uint8 (C)", "C", []byte{255}, 255},
		{"uint16 big (S)", "S", []byte{0x01, 0x02}, 258},
		{"uint16 little (s)", "s", []byte{0x02, 0x01}, 258},
		{"uint32 big (N)", "N", []byte{0x01, 0x02, 0x03, 0x04}, 16909060},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建字节数组
			bytesArr := make([]engine.Value, len(tt.data))
			for i, b := range tt.data {
				bytesArr[i] = engine.NewInt(int64(b))
			}

			args := []engine.Value{
				engine.NewString(tt.format),
				engine.NewArray(bytesArr),
			}

			result, err := builtinUnpack(nil, args)
			if err != nil {
				t.Fatalf("unpack() error = %v", err)
			}

			if result.Int() != tt.expected {
				t.Errorf("unpack() = %d, expected %d", result.Int(), tt.expected)
			}
		})
	}
}

// TestUnpackMultiValue 测试多值 unpack
func TestUnpackMultiValue(t *testing.T) {
	data := []byte{0, 0, 0, 1, 0, 0, 0, 2, 0x01, 0x02}
	bytesArr := make([]engine.Value, len(data))
	for i, b := range data {
		bytesArr[i] = engine.NewInt(int64(b))
	}

	args := []engine.Value{
		engine.NewString("NNS"),
		engine.NewArray(bytesArr),
	}

	result, err := builtinUnpack(nil, args)
	if err != nil {
		t.Fatalf("unpack() error = %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("unpack() returned %d values, expected 3", len(arr))
	}

	if arr[0].Int() != 1 {
		t.Errorf("unpack()[0] = %d, expected 1", arr[0].Int())
	}
	if arr[1].Int() != 2 {
		t.Errorf("unpack()[1] = %d, expected 2", arr[1].Int())
	}
	if arr[2].Int() != 258 {
		t.Errorf("unpack()[2] = %d, expected 258", arr[2].Int())
	}
}

// TestBufferNew 测试 Buffer 创建
func TestBufferNew(t *testing.T) {
	// 测试默认大端
	buf, err := builtinBufferNew(nil, []engine.Value{})
	if err != nil {
		t.Fatalf("buffer_new() error = %v", err)
	}

	if buf.Type() != engine.TypeObject {
		t.Errorf("buffer_new() type = %v, expected TypeObject", buf.Type())
	}

	if buf.Int() != 0 {
		t.Errorf("buffer_new() length = %d, expected 0", buf.Int())
	}

	// 测试显式小端
	buf2, err := builtinBufferNew(nil, []engine.Value{engine.NewString("little")})
	if err != nil {
		t.Fatalf("buffer_new() error = %v", err)
	}

	if buf2.Type() != engine.TypeObject {
		t.Errorf("buffer_new() type = %v, expected TypeObject", buf2.Type())
	}
}

// TestBufferWriteRead 测试 Buffer 读写
func TestBufferWriteRead(t *testing.T) {
	buf := NewBuffer("big")

	// 写入数据
	args := []engine.Value{buf, engine.NewInt(255)}
	_, err := builtinBufferWriteUint8(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_uint8() error = %v", err)
	}

	args = []engine.Value{buf, engine.NewInt(1000)}
	_, err = builtinBufferWriteUint16(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_uint16() error = %v", err)
	}

	args = []engine.Value{buf, engine.NewInt(123456)}
	_, err = builtinBufferWriteUint32(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_uint32() error = %v", err)
	}

	// 验证长度
	lenArgs := []engine.Value{buf}
	lenResult, err := builtinBufferLength(nil, lenArgs)
	if err != nil {
		t.Fatalf("buffer_length() error = %v", err)
	}

	if lenResult.Int() != 7 { // 1 + 2 + 4
		t.Errorf("buffer_length() = %d, expected 7", lenResult.Int())
	}

	// 读取数据
	resetArgs := []engine.Value{buf}
	_, err = builtinBufferReset(nil, resetArgs)
	if err != nil {
		t.Fatalf("buffer_reset() error = %v", err)
	}

	// 重新写入用于读取测试
	args = []engine.Value{buf, engine.NewInt(255)}
	builtinBufferWriteUint8(nil, args)
	args = []engine.Value{buf, engine.NewInt(1000)}
	builtinBufferWriteUint16(nil, args)

	// 读取
	readArgs := []engine.Value{buf}
	val8, _ := builtinBufferReadUint8(nil, readArgs)
	if val8.Int() != 255 {
		t.Errorf("buffer_read_uint8() = %d, expected 255", val8.Int())
	}

	val16, _ := builtinBufferReadUint16(nil, readArgs)
	if val16.Int() != 1000 {
		t.Errorf("buffer_read_uint16() = %d, expected 1000", val16.Int())
	}
}

// TestBufferEndian 测试 Buffer 字节序切换
func TestBufferEndian(t *testing.T) {
	buf := NewBuffer("big")

	// 设置为小端
	args := []engine.Value{buf, engine.NewString("little")}
	result, err := builtinBufferSetEndian(nil, args)
	if err != nil {
		t.Fatalf("buffer_set_endian() error = %v", err)
	}

	if !result.Bool() {
		t.Error("buffer_set_endian() should return true for valid endian")
	}

	// 测试无效字节序
	args = []engine.Value{buf, engine.NewString("invalid")}
	result, err = builtinBufferSetEndian(nil, args)
	if err != nil {
		t.Fatalf("buffer_set_endian() error = %v", err)
	}

	if result.Bool() {
		t.Error("buffer_set_endian() should return false for invalid endian")
	}
}

// TestBufferString 测试 Buffer 字符串操作
func TestBufferString(t *testing.T) {
	buf := NewBuffer("big")

	// 写入字符串
	args := []engine.Value{buf, engine.NewString("hello world")}
	_, err := builtinBufferWriteString(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_string() error = %v", err)
	}

	// 验证长度
	lenArgs := []engine.Value{buf}
	lenResult, _ := builtinBufferLength(nil, lenArgs)
	if lenResult.Int() != 11 {
		t.Errorf("buffer_length() = %d, expected 11", lenResult.Int())
	}

	// 读取字符串
	readArgs := []engine.Value{buf, engine.NewInt(5)}
	str, _ := builtinBufferReadString(nil, readArgs)
	if str.String() != "hello" {
		t.Errorf("buffer_read_string() = %q, expected \"hello\"", str.String())
	}
}

// TestBufferSeek 测试 Buffer 游标操作
func TestBufferSeek(t *testing.T) {
	buf := NewBuffer("big")

	// 写入一些数据
	args := []engine.Value{buf, engine.NewInt(100)}
	builtinBufferWriteUint32(nil, args)

	// 重置游标
	seekArgs := []engine.Value{buf, engine.NewInt(0)}
	pos, err := builtinBufferSeek(nil, seekArgs)
	if err != nil {
		t.Fatalf("buffer_seek() error = %v", err)
	}

	if pos.Int() != 0 {
		t.Errorf("buffer_seek() position = %d, expected 0", pos.Int())
	}

	// 检查 tell
	tellArgs := []engine.Value{buf}
	tellPos, _ := builtinBufferTell(nil, tellArgs)
	if tellPos.Int() != 0 {
		t.Errorf("buffer_tell() = %d, expected 0", tellPos.Int())
	}
}

// TestBufferToBytes 测试 Buffer 转字节数组
func TestBufferToBytes(t *testing.T) {
	buf := NewBuffer("big")

	// 写入数据
	args := []engine.Value{buf, engine.NewInt(1)}
	builtinBufferWriteUint8(nil, args)
	args = []engine.Value{buf, engine.NewInt(2)}
	builtinBufferWriteUint8(nil, args)
	args = []engine.Value{buf, engine.NewInt(3)}
	builtinBufferWriteUint8(nil, args)

	// 转换为字节数组
	toBytesArgs := []engine.Value{buf}
	result, err := builtinBufferToBytes(nil, toBytesArgs)
	if err != nil {
		t.Fatalf("buffer_to_bytes() error = %v", err)
	}

	arr := result.Array()
	if len(arr) != 3 {
		t.Fatalf("buffer_to_bytes() returned %d bytes, expected 3", len(arr))
	}

	for i, expected := range []int64{1, 2, 3} {
		if arr[i].Int() != expected {
			t.Errorf("buffer_to_bytes()[%d] = %d, expected %d", i, arr[i].Int(), expected)
		}
	}
}

// TestBufferToString 测试 Buffer 转字符串
func TestBufferToString(t *testing.T) {
	buf := NewBuffer("big")

	// 写入字符串
	args := []engine.Value{buf, engine.NewString("hello")}
	builtinBufferWriteString(nil, args)

	// 转换为字符串
	toStringArgs := []engine.Value{buf}
	result, err := builtinBufferToString(nil, toStringArgs)
	if err != nil {
		t.Fatalf("buffer_to_string() error = %v", err)
	}

	if result.String() != "hello" {
		t.Errorf("buffer_to_string() = %q, expected \"hello\"", result.String())
	}
}

// TestBufferFloat 测试 Buffer 浮点数
func TestBufferFloat(t *testing.T) {
	buf := NewBuffer("big")

	// 写入 float32
	args := []engine.Value{buf, engine.NewFloat(3.14)}
	_, err := builtinBufferWriteFloat32(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_float32() error = %v", err)
	}

	// 读取 float32
	readArgs := []engine.Value{buf}
	val, _ := builtinBufferReadFloat32(nil, readArgs)
	// 允许小的精度误差
	if val.Float() < 3.13 || val.Float() > 3.15 {
		t.Errorf("buffer_read_float32() = %f, expected ~3.14", val.Float())
	}
}

// TestBufferWriteReadBytes 测试 Buffer 字节数组操作
func TestBufferWriteReadBytes(t *testing.T) {
	buf := NewBuffer("big")

	// 写入字节数组
	bytesArr := []engine.Value{
		engine.NewInt(1),
		engine.NewInt(2),
		engine.NewInt(3),
		engine.NewInt(4),
		engine.NewInt(5),
	}
	args := []engine.Value{buf, engine.NewArray(bytesArr)}
	_, err := builtinBufferWriteBytes(nil, args)
	if err != nil {
		t.Fatalf("buffer_write_bytes() error = %v", err)
	}

	// 读取字节数组
	readArgs := []engine.Value{buf, engine.NewInt(3)}
	result, _ := builtinBufferReadBytes(nil, readArgs)
	arr := result.Array()

	if len(arr) != 3 {
		t.Fatalf("buffer_read_bytes() returned %d bytes, expected 3", len(arr))
	}

	for i, expected := range []int64{1, 2, 3} {
		if arr[i].Int() != expected {
			t.Errorf("buffer_read_bytes()[%d] = %d, expected %d", i, arr[i].Int(), expected)
		}
	}
}

// TestBufferNewFrom 测试 buffer_new_from 函数
func TestBufferNewFrom(t *testing.T) {
	// 测试从字节数组创建
	t.Run("from array", func(t *testing.T) {
		bytes := []engine.Value{
			engine.NewInt(0x48), // H
			engine.NewInt(0x65), // e
			engine.NewInt(0x6C), // l
			engine.NewInt(0x6C), // l
			engine.NewInt(0x6F), // o
		}
		arr := engine.NewArray(bytes)
		buf, err := builtinBufferNewFrom(nil, []engine.Value{arr})
		if err != nil {
			t.Fatalf("buffer_new_from() error = %v", err)
		}

		if buf.Type() != engine.TypeObject {
			t.Errorf("buffer_new_from() type = %v, expected TypeObject", buf.Type())
		}

		// 验证内容长度
		lenArgs := []engine.Value{buf}
		lenResult, _ := builtinBufferLength(nil, lenArgs)
		if lenResult.Int() != 5 {
			t.Errorf("buffer_new_from() length = %d, expected 5", lenResult.Int())
		}

		// 验证内容
		strResult, _ := builtinBufferToString(nil, lenArgs)
		if strResult.String() != "Hello" {
			t.Errorf("buffer_new_from() content = %q, expected 'Hello'", strResult.String())
		}
	})

	// 测试从字符串创建
	t.Run("from string", func(t *testing.T) {
		buf, err := builtinBufferNewFrom(nil, []engine.Value{engine.NewString("World"), engine.NewString("little")})
		if err != nil {
			t.Fatalf("buffer_new_from() error = %v", err)
		}

		lenArgs := []engine.Value{buf}
		lenResult, _ := builtinBufferLength(nil, lenArgs)
		if lenResult.Int() != 5 {
			t.Errorf("buffer_new_from() length = %d, expected 5", lenResult.Int())
		}

		strResult, _ := builtinBufferToString(nil, lenArgs)
		if strResult.String() != "World" {
			t.Errorf("buffer_new_from() content = %q, expected 'World'", strResult.String())
		}
	})

	// 测试错误参数
	t.Run("invalid args", func(t *testing.T) {
		// 缺少参数
		_, err := builtinBufferNewFrom(nil, []engine.Value{})
		if err == nil {
			t.Error("buffer_new_from() should error with no arguments")
		}

		// 无效类型
		_, err = builtinBufferNewFrom(nil, []engine.Value{engine.NewInt(123)})
		if err == nil {
			t.Error("buffer_new_from() should error with invalid type")
		}
	})

	// 测试包含非整数元素的数组（应跳过非整数元素）
	t.Run("mixed array", func(t *testing.T) {
		bytes := []engine.Value{
			engine.NewInt(0x41),
			engine.NewString("skip"), // 应被跳过
			engine.NewInt(0x42),
		}
		arr := engine.NewArray(bytes)
		buf, err := builtinBufferNewFrom(nil, []engine.Value{arr})
		if err != nil {
			t.Fatalf("buffer_new_from() error = %v", err)
		}

		lenArgs := []engine.Value{buf}
		lenResult, _ := builtinBufferLength(nil, lenArgs)
		if lenResult.Int() != 2 { // 只有 0x41 和 0x42
			t.Errorf("buffer_new_from() length = %d, expected 2", lenResult.Int())
		}
	})
}

// TestBufferSignedInt 测试有符号整数读写
func TestBufferSignedInt(t *testing.T) {
	// 测试 int8
	t.Run("int8", func(t *testing.T) {
		buf := NewBuffer("big")

		// 写入正负值
		_, err := builtinBufferWriteInt8(nil, []engine.Value{buf, engine.NewInt(-128)})
		if err != nil {
			t.Fatalf("buffer_write_int8() error = %v", err)
		}

		_, err = builtinBufferWriteInt8(nil, []engine.Value{buf, engine.NewInt(127)})
		if err != nil {
			t.Fatalf("buffer_write_int8() error = %v", err)
		}

		// 验证长度
		lenArgs := []engine.Value{buf}
		lenResult, _ := builtinBufferLength(nil, lenArgs)
		if lenResult.Int() != 2 {
			t.Errorf("buffer length = %d, expected 2", lenResult.Int())
		}

		// 重置并重新写入用于读取测试
		_, _ = builtinBufferReset(nil, lenArgs)
		_, _ = builtinBufferWriteInt8(nil, []engine.Value{buf, engine.NewInt(-128)})
		_, _ = builtinBufferWriteInt8(nil, []engine.Value{buf, engine.NewInt(127)})

		// 读取并验证
		val1, _ := builtinBufferReadInt8(nil, []engine.Value{buf})
		if val1.Int() != -128 {
			t.Errorf("buffer_read_int8() = %d, expected -128", val1.Int())
		}

		val2, _ := builtinBufferReadInt8(nil, []engine.Value{buf})
		if val2.Int() != 127 {
			t.Errorf("buffer_read_int8() = %d, expected 127", val2.Int())
		}

		// 测试溢出读取返回 null
		val3, _ := builtinBufferReadInt8(nil, []engine.Value{buf})
		if !val3.IsNull() {
			t.Errorf("buffer_read_int8() on empty should return null, got %v", val3)
		}
	})

	// 测试 int16
	t.Run("int16", func(t *testing.T) {
		buf := NewBuffer("little")

		// 写入正负值
		_, _ = builtinBufferWriteInt16(nil, []engine.Value{buf, engine.NewInt(-32768)})
		_, _ = builtinBufferWriteInt16(nil, []engine.Value{buf, engine.NewInt(32767)})

		// 重置并重新写入用于读取测试
		_, _ = builtinBufferReset(nil, []engine.Value{buf})
		_, _ = builtinBufferWriteInt16(nil, []engine.Value{buf, engine.NewInt(-32768)})
		_, _ = builtinBufferWriteInt16(nil, []engine.Value{buf, engine.NewInt(32767)})

		val1, _ := builtinBufferReadInt16(nil, []engine.Value{buf})
		if val1.Int() != -32768 {
			t.Errorf("buffer_read_int16() = %d, expected -32768", val1.Int())
		}

		val2, _ := builtinBufferReadInt16(nil, []engine.Value{buf})
		if val2.Int() != 32767 {
			t.Errorf("buffer_read_int16() = %d, expected 32767", val2.Int())
		}
	})

	// 测试 int32
	t.Run("int32", func(t *testing.T) {
		buf := NewBuffer("big")

		// 写入正负值
		_, _ = builtinBufferWriteInt32(nil, []engine.Value{buf, engine.NewInt(-2147483648)})
		_, _ = builtinBufferWriteInt32(nil, []engine.Value{buf, engine.NewInt(2147483647)})

		// 重置并重新写入用于读取测试
		_, _ = builtinBufferReset(nil, []engine.Value{buf})
		_, _ = builtinBufferWriteInt32(nil, []engine.Value{buf, engine.NewInt(-2147483648)})
		_, _ = builtinBufferWriteInt32(nil, []engine.Value{buf, engine.NewInt(2147483647)})

		val1, _ := builtinBufferReadInt32(nil, []engine.Value{buf})
		if val1.Int() != -2147483648 {
			t.Errorf("buffer_read_int32() = %d, expected -2147483648", val1.Int())
		}

		val2, _ := builtinBufferReadInt32(nil, []engine.Value{buf})
		if val2.Int() != 2147483647 {
			t.Errorf("buffer_read_int32() = %d, expected 2147483647", val2.Int())
		}
	})
}

// TestBufferSignedIntErrors 测试有符号整数错误处理
func TestBufferSignedIntErrors(t *testing.T) {
	// 测试参数错误
	t.Run("write errors", func(t *testing.T) {
		// 缺少参数
		_, err := builtinBufferWriteInt8(nil, []engine.Value{})
		if err == nil {
			t.Error("buffer_write_int8() should error with no arguments")
		}

		// 无效 buffer 类型
		_, err = builtinBufferWriteInt16(nil, []engine.Value{engine.NewInt(123), engine.NewInt(1)})
		if err == nil {
			t.Error("buffer_write_int16() should error with non-buffer first arg")
		}

		// 无效 buffer 类型 - int32
		_, err = builtinBufferWriteInt32(nil, []engine.Value{engine.NewString("not buffer"), engine.NewInt(1)})
		if err == nil {
			t.Error("buffer_write_int32() should error with non-buffer first arg")
		}
	})

	t.Run("read errors", func(t *testing.T) {
		// 缺少参数
		_, err := builtinBufferReadInt8(nil, []engine.Value{})
		if err == nil {
			t.Error("buffer_read_int8() should error with no arguments")
		}

		// 无效 buffer 类型
		_, err = builtinBufferReadInt16(nil, []engine.Value{engine.NewInt(123)})
		if err == nil {
			t.Error("buffer_read_int16() should error with non-buffer arg")
		}

		// 无效 buffer 类型 - int32
		_, err = builtinBufferReadInt32(nil, []engine.Value{engine.NewString("not buffer")})
		if err == nil {
			t.Error("buffer_read_int32() should error with non-buffer arg")
		}
	})
}

// TestBufferSignedIntFromExisting 测试从现有数据读取有符号整数
func TestBufferSignedIntFromExisting(t *testing.T) {
	// 使用 buffer_new_from 创建包含特定字节的 buffer，然后读取有符号整数
	t.Run("read negative from bytes", func(t *testing.T) {
		// -1 的补码表示：0xFF (int8), 0xFFFF (int16), 0xFFFFFFFF (int32)
		bytes := []engine.Value{
			engine.NewInt(0xFF),                      // -1 as int8
			engine.NewInt(0xFF), engine.NewInt(0xFF), // -1 as int16 (big endian)
		}
		arr := engine.NewArray(bytes)
		buf, _ := builtinBufferNewFrom(nil, []engine.Value{arr})

		val8, _ := builtinBufferReadInt8(nil, []engine.Value{buf})
		if val8.Int() != -1 {
			t.Errorf("int8 -1 = %d, expected -1", val8.Int())
		}

		val16, _ := builtinBufferReadInt16(nil, []engine.Value{buf})
		if val16.Int() != -1 {
			t.Errorf("int16 -1 = %d, expected -1", val16.Int())
		}
	})

	// 测试小端序读取
	t.Run("little endian", func(t *testing.T) {
		// 小端序：0x0102 存储为 [0x02, 0x01]
		// 对于有符号数：0xFFFE = -2 (小端序存储为 [0xFE, 0xFF])
		bytes := []engine.Value{
			engine.NewInt(0xFE),
			engine.NewInt(0xFF),
		}
		arr := engine.NewArray(bytes)
		buf, _ := builtinBufferNewFrom(nil, []engine.Value{arr, engine.NewString("little")})

		val16, _ := builtinBufferReadInt16(nil, []engine.Value{buf})
		if val16.Int() != -2 {
			t.Errorf("int16 little endian -2 = %d, expected -2", val16.Int())
		}
	})
}

// TestIsBuffer 测试 is_buffer 函数
func TestIsBuffer(t *testing.T) {
	// 测试 Buffer 对象
	t.Run("with buffer", func(t *testing.T) {
		buf := NewBuffer("big")
		result, err := builtinIsBuffer(nil, []engine.Value{buf})
		if err != nil {
			t.Fatalf("is_buffer() error = %v", err)
		}
		if !result.Bool() {
			t.Error("is_buffer(buffer) should return true")
		}
	})

	// 测试非 Buffer 类型
	t.Run("with non-buffer types", func(t *testing.T) {
		testCases := []struct {
			name  string
			value engine.Value
		}{
			{"int", engine.NewInt(123)},
			{"float", engine.NewFloat(3.14)},
			{"string", engine.NewString("hello")},
			{"array", engine.NewArray([]engine.Value{engine.NewInt(1)})},
			{"bool true", engine.NewBool(true)},
			{"bool false", engine.NewBool(false)},
			{"null", engine.NewNull()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := builtinIsBuffer(nil, []engine.Value{tc.value})
				if err != nil {
					t.Fatalf("is_buffer() error = %v", err)
				}
				if result.Bool() {
					t.Errorf("is_buffer(%s) should return false", tc.name)
				}
			})
		}
	})

	// 测试错误参数
	t.Run("invalid args", func(t *testing.T) {
		// 缺少参数
		_, err := builtinIsBuffer(nil, []engine.Value{})
		if err == nil {
			t.Error("is_buffer() should error with no arguments")
		}

		// 参数过多
		_, err = builtinIsBuffer(nil, []engine.Value{
			NewBuffer("big"),
			engine.NewInt(123),
		})
		if err == nil {
			t.Error("is_buffer() should error with too many arguments")
		}
	})
}
