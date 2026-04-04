package stdlib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterBinary 注册二进制处理函数到引擎
// 注册的函数包括：pack/unpack（二进制打包/解包）和 buffer_xxx（缓冲区操作）系列函数
func RegisterBinary(e *engine.Engine) {
	e.RegisterFunc("pack", builtinPack)
	e.RegisterFunc("unpack", builtinUnpack)
	e.RegisterFunc("buffer_new", builtinBufferNew)
	e.RegisterFunc("buffer_new_from", builtinBufferNewFrom)
	e.RegisterFunc("buffer_set_endian", builtinBufferSetEndian)
	e.RegisterFunc("buffer_write_int8", builtinBufferWriteInt8)
	e.RegisterFunc("buffer_write_int16", builtinBufferWriteInt16)
	e.RegisterFunc("buffer_write_int32", builtinBufferWriteInt32)
	e.RegisterFunc("buffer_write_uint8", builtinBufferWriteUint8)
	e.RegisterFunc("buffer_write_uint16", builtinBufferWriteUint16)
	e.RegisterFunc("buffer_write_uint32", builtinBufferWriteUint32)
	e.RegisterFunc("buffer_write_float32", builtinBufferWriteFloat32)
	e.RegisterFunc("buffer_write_float64", builtinBufferWriteFloat64)
	e.RegisterFunc("buffer_write_string", builtinBufferWriteString)
	e.RegisterFunc("buffer_write_bytes", builtinBufferWriteBytes)
	e.RegisterFunc("buffer_read_int8", builtinBufferReadInt8)
	e.RegisterFunc("buffer_read_int16", builtinBufferReadInt16)
	e.RegisterFunc("buffer_read_int32", builtinBufferReadInt32)
	e.RegisterFunc("buffer_read_uint8", builtinBufferReadUint8)
	e.RegisterFunc("buffer_read_uint16", builtinBufferReadUint16)
	e.RegisterFunc("buffer_read_uint32", builtinBufferReadUint32)
	e.RegisterFunc("buffer_read_float32", builtinBufferReadFloat32)
	e.RegisterFunc("buffer_read_float64", builtinBufferReadFloat64)
	e.RegisterFunc("buffer_read_string", builtinBufferReadString)
	e.RegisterFunc("buffer_read_bytes", builtinBufferReadBytes)
	e.RegisterFunc("buffer_seek", builtinBufferSeek)
	e.RegisterFunc("buffer_tell", builtinBufferTell)
	e.RegisterFunc("buffer_length", builtinBufferLength)
	e.RegisterFunc("buffer_reset", builtinBufferReset)
	e.RegisterFunc("buffer_to_bytes", builtinBufferToBytes)
	e.RegisterFunc("buffer_to_string", builtinBufferToString)
	e.RegisterFunc("is_buffer", builtinIsBuffer)

	// 模块注册 — import "binary" 可用
	e.RegisterModule("binary", map[string]engine.GoFunction{
		"pack":                 builtinPack,
		"unpack":               builtinUnpack,
		"buffer_new":           builtinBufferNew,
		"buffer_new_from":      builtinBufferNewFrom,
		"buffer_set_endian":    builtinBufferSetEndian,
		"buffer_write_int8":    builtinBufferWriteInt8,
		"buffer_write_int16":   builtinBufferWriteInt16,
		"buffer_write_int32":   builtinBufferWriteInt32,
		"buffer_write_uint8":   builtinBufferWriteUint8,
		"buffer_write_uint16":  builtinBufferWriteUint16,
		"buffer_write_uint32":  builtinBufferWriteUint32,
		"buffer_write_float32": builtinBufferWriteFloat32,
		"buffer_write_float64": builtinBufferWriteFloat64,
		"buffer_write_string":  builtinBufferWriteString,
		"buffer_write_bytes":   builtinBufferWriteBytes,
		"buffer_read_int8":     builtinBufferReadInt8,
		"buffer_read_int16":    builtinBufferReadInt16,
		"buffer_read_int32":    builtinBufferReadInt32,
		"buffer_read_uint8":    builtinBufferReadUint8,
		"buffer_read_uint16":   builtinBufferReadUint16,
		"buffer_read_uint32":   builtinBufferReadUint32,
		"buffer_read_float32":  builtinBufferReadFloat32,
		"buffer_read_float64":  builtinBufferReadFloat64,
		"buffer_read_string":   builtinBufferReadString,
		"buffer_read_bytes":    builtinBufferReadBytes,
		"buffer_seek":          builtinBufferSeek,
		"buffer_tell":          builtinBufferTell,
		"buffer_length":        builtinBufferLength,
		"buffer_reset":         builtinBufferReset,
		"buffer_to_bytes":      builtinBufferToBytes,
		"buffer_to_string":     builtinBufferToString,
		"is_buffer":            builtinIsBuffer,
	})
}

// BinaryNames 返回二进制处理函数名称列表
func BinaryNames() []string {
	return []string{
		"pack", "unpack",
		"buffer_new", "buffer_new_from", "buffer_set_endian",
		"buffer_write_int8", "buffer_write_int16", "buffer_write_int32",
		"buffer_write_uint8", "buffer_write_uint16", "buffer_write_uint32",
		"buffer_write_float32", "buffer_write_float64",
		"buffer_write_string", "buffer_write_bytes",
		"buffer_read_int8", "buffer_read_int16", "buffer_read_int32",
		"buffer_read_uint8", "buffer_read_uint16", "buffer_read_uint32",
		"buffer_read_float32", "buffer_read_float64",
		"buffer_read_string", "buffer_read_bytes",
		"buffer_seek", "buffer_tell", "buffer_length", "buffer_reset",
		"buffer_to_bytes", "buffer_to_string",
		"is_buffer",
	}
}

// packFormat 定义 pack/unpack 的格式字符
// 支持以下格式代码：
//
//	C: 无符号字节(1字节)
//	S: 无符号短整型大端序(2字节)
//	s: 无符号短整型小端序(2字节)
//	N: 无符号整型大端序(4字节)
//	V: 无符号整型小端序(4字节)
//	Q: 无符号长整型大端序(8字节)
//	q: 无符号长整型小端序(8字节)
//	f: 单精度浮点数(4字节，大端序)
//	d: 双精度浮点数(8字节，大端序)
//	a: 字符串（非零结尾）
//	Z: 零结尾字符串
//	x: 填充字节(1字节，值为0)
type packFormat struct {
	code       byte
	size       int
	endian     binary.ByteOrder
	isString   bool
	isZeroTerm bool
}

// parsePackFormat 解析单个格式字符
// 根据格式代码返回对应的格式定义，包括数据大小、字节序等信息
func parsePackFormat(code byte, defaultEndian binary.ByteOrder) (packFormat, error) {
	switch code {
	case 'C':
		return packFormat{code: code, size: 1, endian: defaultEndian}, nil
	case 'S':
		return packFormat{code: code, size: 2, endian: binary.BigEndian}, nil
	case 's':
		return packFormat{code: code, size: 2, endian: binary.LittleEndian}, nil
	case 'N':
		return packFormat{code: code, size: 4, endian: binary.BigEndian}, nil
	case 'V':
		return packFormat{code: code, size: 4, endian: binary.LittleEndian}, nil
	case 'Q':
		return packFormat{code: code, size: 8, endian: binary.BigEndian}, nil
	case 'q':
		return packFormat{code: code, size: 8, endian: binary.LittleEndian}, nil
	case 'f':
		return packFormat{code: code, size: 4, endian: binary.BigEndian}, nil
	case 'd':
		return packFormat{code: code, size: 8, endian: binary.BigEndian}, nil
	case 'a', 'Z':
		return packFormat{code: code, size: 0, endian: nil, isString: true, isZeroTerm: code == 'Z'}, nil
	case 'x':
		return packFormat{code: code, size: 1, endian: nil}, nil
	default:
		return packFormat{}, fmt.Errorf("unknown format code: %c", code)
	}
}

// builtinPack 将多个值按指定格式打包成二进制字节数组
//
// 参数：
//   - args[0]: 格式字符串，由格式代码字符组成（如 "CSN" 表示字节+短整+整型）
//   - args[1..]: 要打包的数值，数量和类型必须与格式字符串匹配
//
// 返回值：
//   - 字节数组（Array 类型），每个元素是一个字节值（0-255）
//
// 示例：
//
//	pack("CSN", 1, 2, 3) → 打包 1字节 + 2字节 + 4字节 = 7字节数组
//	pack("Z", "hello") → [104, 101, 108, 108, 111, 0]（字符串+零结尾）
//
// 注意：Q/q（64位整数）使用 int64 直接转换，避免 float64 精度丢失
func builtinPack(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("pack() expects at least 1 argument, got %d", len(args))
	}

	formatStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("pack() expects string format, got %s", args[0].Type())
	}

	buf := new(bytes.Buffer)
	valIdx := 1

	for i := 0; i < len(formatStr); i++ {
		code := formatStr[i]

		// 获取格式定义
		pf, err := parsePackFormat(code, binary.BigEndian)
		if err != nil {
			return nil, err
		}

		// 处理填充字节 'x'
		if code == 'x' {
			buf.WriteByte(0)
			continue
		}

		// 检查参数
		if valIdx >= len(args) {
			return nil, fmt.Errorf("pack(): not enough arguments for format '%c' at position %d", code, i)
		}

		val := args[valIdx]
		valIdx++

		// 处理字符串类型
		if pf.isString {
			str := val.String()
			if val.Type() != engine.TypeString {
				return nil, fmt.Errorf("pack() expects string for format '%c', got %s", code, val.Type())
			}
			buf.WriteString(str)
			if pf.isZeroTerm {
				buf.WriteByte(0)
			}
			continue
		}

		// 处理数值类型
		var num float64
		var intNum int64
		isInt := false
		if val.Type() == engine.TypeInt {
			intNum = val.Int()
			num = float64(intNum)
			isInt = true
		} else if val.Type() == engine.TypeFloat {
			num = val.Float()
		} else {
			return nil, fmt.Errorf("pack() expects number for format '%c', got %s", code, val.Type())
		}

		// 根据格式写入数据
		switch code {
		case 'C':
			buf.WriteByte(byte(num))
		case 'S', 's':
			binary.Write(buf, pf.endian, uint16(num))
		case 'N', 'V':
			binary.Write(buf, pf.endian, uint32(num))
		case 'Q', 'q':
			// 使用 intNum 避免 float64 精度丢失
			if isInt {
				binary.Write(buf, pf.endian, uint64(intNum))
			} else {
				binary.Write(buf, pf.endian, uint64(num))
			}
		case 'f':
			binary.Write(buf, binary.BigEndian, float32(num))
		case 'd':
			binary.Write(buf, binary.BigEndian, float64(num))
		}
	}

	// 返回字节数组
	data := buf.Bytes()
	values := make([]engine.Value, len(data))
	for i, b := range data {
		values[i] = engine.NewInt(int64(b))
	}
	return engine.NewArray(values), nil
}

// builtinUnpack 将二进制字节数组按指定格式解包成多个值
//
// 参数：
//   - args[0]: 格式字符串，由格式代码字符组成
//   - args[1]: 字节数组（Array 类型），每个元素是一个字节值
//
// 返回值：
//   - 单个值：如果格式字符串只有一个代码
//   - 值数组：如果格式字符串有多个代码
//
// 示例：
//
//	unpack("CSN", bytes) → [1, 2, 3]（解包为3个数值）
//	unpack("Z", [104, 101, 0]) → "he"（读取零结尾字符串）
//
// 注意：字符串类型 'a'/'Z' 会读取剩余所有数据
func builtinUnpack(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("unpack() expects exactly 2 arguments, got %d", len(args))
	}

	formatStr := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("unpack() expects string format, got %s", args[0].Type())
	}

	// 获取字节数组
	if args[1].Type() != engine.TypeArray {
		return nil, fmt.Errorf("unpack() expects array of bytes, got %s", args[1].Type())
	}

	arr := args[1].Array()
	if arr == nil {
		return nil, fmt.Errorf("unpack() expects array of bytes")
	}

	// 转换为字节切片
	data := make([]byte, len(arr))
	for i, elem := range arr {
		if elem.Type() == engine.TypeInt {
			data[i] = byte(elem.Int())
		}
	}

	reader := bytes.NewReader(data)
	results := make([]engine.Value, 0)

	for i := 0; i < len(formatStr); i++ {
		code := formatStr[i]

		// 获取格式定义
		pf, err := parsePackFormat(code, binary.BigEndian)
		if err != nil {
			return nil, err
		}

		// 处理填充字节 'x'
		if code == 'x' {
			reader.Seek(1, io.SeekCurrent)
			continue
		}

		// 处理字符串类型
		if pf.isString {
			// 读取剩余所有数据
			remaining := reader.Len()
			strBytes := make([]byte, remaining)
			reader.Read(strBytes)

			if pf.isZeroTerm {
				// 找到零结尾
				for j, b := range strBytes {
					if b == 0 {
						strBytes = strBytes[:j]
						break
					}
				}
			}

			results = append(results, engine.NewString(string(strBytes)))
			continue
		}

		// 读取数值
		var val engine.Value
		switch code {
		case 'C':
			b, _ := reader.ReadByte()
			val = engine.NewInt(int64(b))
		case 'S', 's':
			var v uint16
			binary.Read(reader, pf.endian, &v)
			val = engine.NewInt(int64(v))
		case 'N', 'V':
			var v uint32
			binary.Read(reader, pf.endian, &v)
			val = engine.NewInt(int64(v))
		case 'Q', 'q':
			var v uint64
			binary.Read(reader, pf.endian, &v)
			val = engine.NewInt(int64(v))
		case 'f':
			var v float32
			binary.Read(reader, binary.BigEndian, &v)
			val = engine.NewFloat(float64(v))
		case 'd':
			var v float64
			binary.Read(reader, binary.BigEndian, &v)
			val = engine.NewFloat(v)
		}
		results = append(results, val)
	}

	// 返回结果
	if len(results) == 1 {
		return results[0], nil
	}

	// 多值返回数组
	return engine.NewArray(results), nil
}

// BufferValue 表示一个二进制缓冲区对象
// 用于构建和读取二进制协议数据，支持多种数据类型的读写
// 默认使用大端序（Big Endian），可切换为小端序（Little Endian）
type BufferValue struct {
	data   *bytes.Buffer    // 底层字节缓冲区
	order  binary.ByteOrder // 当前字节序（BigEndian 或 LittleEndian）
	reader *bytes.Reader    // 用于读取的 Reader 实例（延迟创建）
}

// Type 返回类型标识
func (b *BufferValue) Type() engine.ValueType { return engine.TypeObject }

// IsNull 检查是否为 null
func (b *BufferValue) IsNull() bool { return false }

// Bool 返回布尔值
func (b *BufferValue) Bool() bool { return true }

// Int 返回整数（缓冲区长度）
func (b *BufferValue) Int() int64 { return int64(b.data.Len()) }

// Float 返回浮点数
func (b *BufferValue) Float() float64 { return float64(b.data.Len()) }

// String 返回字符串表示
func (b *BufferValue) String() string {
	return fmt.Sprintf("Buffer(%d bytes)", b.data.Len())
}

// Stringify 返回 JSON 序列化字符串
func (b *BufferValue) Stringify() string {
	return b.String()
}

// Array 返回数组值
func (b *BufferValue) Array() []engine.Value { return nil }

// Object 返回对象值
func (b *BufferValue) Object() map[string]engine.Value { return nil }

// Len 返回长度
func (b *BufferValue) Len() int { return b.data.Len() }

// Add 添加值
func (b *BufferValue) Add(v engine.Value) engine.Value { return b }

// Sub 减去值
func (b *BufferValue) Sub(v engine.Value) engine.Value { return b }

// Mul 乘以值
func (b *BufferValue) Mul(v engine.Value) engine.Value { return b }

// Div 除以值
func (b *BufferValue) Div(v engine.Value) engine.Value { return b }

// Mod 取模
func (b *BufferValue) Mod(v engine.Value) engine.Value { return b }

// Negate 取反
func (b *BufferValue) Negate() engine.Value { return b }

// Equals 等于
func (b *BufferValue) Equals(v engine.Value) bool {
	if other, ok := v.(*BufferValue); ok {
		return bytes.Equal(b.data.Bytes(), other.data.Bytes())
	}
	return false
}

// Less 小于
func (b *BufferValue) Less(v engine.Value) bool { return false }

// Greater 大于
func (b *BufferValue) Greater(v engine.Value) bool { return false }

// LessEqual 小于等于
func (b *BufferValue) LessEqual(v engine.Value) bool { return false }

// GreaterEqual 大于等于
func (b *BufferValue) GreaterEqual(v engine.Value) bool { return false }

// ToBigInt 转换为大整数
func (b *BufferValue) ToBigInt() engine.Value { return engine.NewInt(0) }

// ToBigDecimal 转换为大十进制数
func (b *BufferValue) ToBigDecimal() engine.Value { return engine.NewFloat(0) }

// Bytes 返回字节数组
func (b *BufferValue) Bytes() []byte {
	return b.data.Bytes()
}

// ensureReader 确保 reader 存在
// 如果 reader 为 nil，则使用当前缓冲区数据创建新的 reader
// 用于延迟初始化，避免在不需要读取时创建不必要的对象
func (b *BufferValue) ensureReader() {
	if b.reader == nil {
		b.reader = bytes.NewReader(b.data.Bytes())
	}
}

// NewBuffer 创建新的 Buffer 对象
//
// 参数：
//   - endian: 字节序，"big"/"be" 表示大端序（默认），"little"/"le" 表示小端序
//
// 返回值：
//   - 新创建的 BufferValue 实例
//
// 示例：
//
//	NewBuffer("big") → 大端序缓冲区
//	NewBuffer("little") → 小端序缓冲区
func NewBuffer(endian string) *BufferValue {
	var order binary.ByteOrder = binary.BigEndian
	if strings.ToLower(endian) == "little" {
		order = binary.LittleEndian
	}
	return &BufferValue{
		data:  new(bytes.Buffer),
		order: order,
	}
}

// builtinBufferNew 创建 Buffer 对象
// buffer_new([endian]) → Buffer
//
// 参数：
//   - args[0]: 可选，字节序字符串（"big"/"little"），默认为 "big"
//
// 返回值：
//   - Buffer 对象
func builtinBufferNew(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	endian := "big"
	if len(args) > 0 {
		endian = strings.ToLower(args[0].String())
	}
	return NewBuffer(endian), nil
}

// builtinBufferSetEndian 设置 Buffer 的字节序
// buffer_set_endian($buf, $endian) → bool
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 字节序字符串（"big"/"be"/"little"/"le"）
//
// 返回值：
//   - true: 设置成功
//   - false: 无效的字节序
func builtinBufferSetEndian(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_set_endian() expects 2 arguments, got %d", len(args))
	}

	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_set_endian() expects Buffer, got %s", args[0].Type())
	}

	endian := args[1].String()
	switch strings.ToLower(endian) {
	case "big", "be":
		buf.order = binary.BigEndian
		return engine.NewBool(true), nil
	case "little", "le":
		buf.order = binary.LittleEndian
		return engine.NewBool(true), nil
	default:
		return engine.NewBool(false), nil
	}
}

// builtinBufferWriteUint8 写入无符号8位整数
// buffer_write_uint8($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的8位整数值（0-255）
func builtinBufferWriteUint8(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_uint8() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_uint8() expects Buffer, got %s", args[0].Type())
	}
	var val uint8
	if args[1].Type() == engine.TypeInt {
		val = uint8(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = uint8(args[1].Float())
	}
	buf.data.WriteByte(val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteUint16 写入无符号16位整数
// buffer_write_uint16($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的16位整数值（0-65535）
//
// 注意：按照 Buffer 当前的字节序写入
func builtinBufferWriteUint16(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_uint16() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_uint16() expects Buffer, got %s", args[0].Type())
	}
	var val uint16
	if args[1].Type() == engine.TypeInt {
		val = uint16(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = uint16(args[1].Float())
	}
	binary.Write(buf.data, buf.order, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteUint32 写入无符号32位整数
// buffer_write_uint32($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的32位整数值
//
// 注意：按照 Buffer 当前的字节序写入
func builtinBufferWriteUint32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_uint32() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_uint32() expects Buffer, got %s", args[0].Type())
	}
	var val uint32
	if args[1].Type() == engine.TypeInt {
		val = uint32(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = uint32(args[1].Float())
	}
	binary.Write(buf.data, buf.order, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteFloat32 写入单精度浮点数
// buffer_write_float32($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的32位浮点数值
//
// 注意：固定使用大端序写入
func builtinBufferWriteFloat32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_float32() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_float32() expects Buffer, got %s", args[0].Type())
	}
	var val float32
	if args[1].Type() == engine.TypeFloat {
		val = float32(args[1].Float())
	} else if args[1].Type() == engine.TypeInt {
		val = float32(args[1].Int())
	}
	binary.Write(buf.data, binary.BigEndian, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteFloat64 写入双精度浮点数
// buffer_write_float64($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的64位浮点数值
//
// 注意：固定使用大端序写入
func builtinBufferWriteFloat64(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_float64() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_float64() expects Buffer, got %s", args[0].Type())
	}
	var val float64
	if args[1].Type() == engine.TypeFloat {
		val = args[1].Float()
	} else if args[1].Type() == engine.TypeInt {
		val = float64(args[1].Int())
	}
	binary.Write(buf.data, binary.BigEndian, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteString 写入字符串
// buffer_write_string($buf, $string) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的字符串
//
// 注意：直接写入字符串字节，不添加长度前缀或零结尾
func builtinBufferWriteString(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_string() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_string() expects Buffer, got %s", args[0].Type())
	}
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("buffer_write_string() expects string, got %s", args[1].Type())
	}
	buf.data.WriteString(args[1].String())
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteBytes 写入字节数组
// buffer_write_bytes($buf, $bytes_array) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 字节数组，每个元素是0-255的整数值
//
// 示例：
//
//	buffer_write_bytes($buf, [1, 2, 3, 4])
func builtinBufferWriteBytes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_bytes() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_bytes() expects Buffer, got %s", args[0].Type())
	}

	if args[1].Type() != engine.TypeArray {
		return nil, fmt.Errorf("buffer_write_bytes() expects array of bytes, got %s", args[1].Type())
	}

	arr := args[1].Array()
	for _, elem := range arr {
		if elem.Type() == engine.TypeInt {
			buf.data.WriteByte(byte(elem.Int()))
		}
	}
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferReadUint8 读取无符号8位整数
// buffer_read_uint8($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的8位整数值（0-255）
//   - null: 缓冲区已读完
func builtinBufferReadUint8(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_uint8() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_uint8() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	b, err := buf.reader.ReadByte()
	if err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewInt(int64(b)), nil
}

// builtinBufferReadUint16 读取无符号16位整数
// buffer_read_uint16($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的16位整数值（0-65535）
//   - null: 缓冲区不足或读取失败
//
// 注意：按照 Buffer 当前的字节序读取
func builtinBufferReadUint16(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_uint16() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_uint16() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v uint16
	binary.Read(buf.reader, buf.order, &v)
	return engine.NewInt(int64(v)), nil
}

// builtinBufferReadUint32 读取无符号32位整数
// buffer_read_uint32($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的32位整数值
//   - null: 缓冲区不足或读取失败
//
// 注意：按照 Buffer 当前的字节序读取
func builtinBufferReadUint32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_uint32() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_uint32() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v uint32
	binary.Read(buf.reader, buf.order, &v)
	return engine.NewInt(int64(v)), nil
}

// builtinBufferReadFloat32 读取单精度浮点数
// buffer_read_float32($buf) → float | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的32位浮点数值
//   - null: 缓冲区不足或读取失败
//
// 注意：固定使用大端序读取
func builtinBufferReadFloat32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_float32() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_float32() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v float32
	binary.Read(buf.reader, binary.BigEndian, &v)
	return engine.NewFloat(float64(v)), nil
}

// builtinBufferReadFloat64 读取双精度浮点数
// buffer_read_float64($buf) → float | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的64位浮点数值
//   - null: 缓冲区不足或读取失败
//
// 注意：固定使用大端序读取
func builtinBufferReadFloat64(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_float64() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_float64() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v float64
	binary.Read(buf.reader, binary.BigEndian, &v)
	return engine.NewFloat(v), nil
}

// builtinBufferReadString 读取指定长度的字符串
// buffer_read_string($buf, $length) → string
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要读取的字节长度
//
// 返回值：
//   - 读取的字符串（可能短于请求长度，如果缓冲区不足）
//
// 注意：不会自动处理零结尾或长度前缀，纯字节到字符串转换
func builtinBufferReadString(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_read_string() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_string() expects Buffer, got %s", args[0].Type())
	}
	var length int
	if args[1].Type() == engine.TypeInt {
		length = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		length = int(args[1].Float())
	}
	buf.ensureReader()
	data := make([]byte, length)
	n, _ := buf.reader.Read(data)
	return engine.NewString(string(data[:n])), nil
}

// builtinBufferReadBytes 读取指定长度的字节数组
// buffer_read_bytes($buf, $length) → array
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要读取的字节长度
//
// 返回值：
//   - 字节数组，每个元素是0-255的整数值
//   - 实际长度可能短于请求长度，如果缓冲区不足
func builtinBufferReadBytes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_read_bytes() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_bytes() expects Buffer, got %s", args[0].Type())
	}
	var length int
	if args[1].Type() == engine.TypeInt {
		length = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		length = int(args[1].Float())
	}
	buf.ensureReader()
	data := make([]byte, length)
	n, _ := buf.reader.Read(data)

	values := make([]engine.Value, n)
	for i := range n {
		values[i] = engine.NewInt(int64(data[i]))
	}
	return engine.NewArray(values), nil
}

// builtinBufferSeek 设置读取位置
// buffer_seek($buf, $offset) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 偏移量（从0开始）
//
// 返回值：
//   - 新的读取位置
//   - null: seek失败
//
// 注意：seek 操作只影响读取位置，不影响写入位置
func builtinBufferSeek(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_seek() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_seek() expects Buffer, got %s", args[0].Type())
	}
	var offset int64
	if args[1].Type() == engine.TypeInt {
		offset = args[1].Int()
	} else if args[1].Type() == engine.TypeFloat {
		offset = int64(args[1].Float())
	}
	buf.ensureReader()
	pos, err := buf.reader.Seek(offset, io.SeekStart)
	if err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewInt(pos), nil
}

// builtinBufferTell 获取当前读取位置
// buffer_tell($buf) → int
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 当前读取位置（从0开始的字节偏移）
func builtinBufferTell(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_tell() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_tell() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	// 通过 Seek(0, Current) 获取当前位置
	pos, _ := buf.reader.Seek(0, io.SeekCurrent)
	return engine.NewInt(pos), nil
}

// builtinBufferLength 获取缓冲区当前长度
// buffer_length($buf) → int
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 缓冲区中的字节数
func builtinBufferLength(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_length() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_length() expects Buffer, got %s", args[0].Type())
	}
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferReset 重置缓冲区
// buffer_reset($buf) → bool
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - true: 重置成功
//
// 注意：清空所有数据，读取位置归零，但字节序设置保持不变
func builtinBufferReset(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_reset() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_reset() expects Buffer, got %s", args[0].Type())
	}
	buf.data.Reset()
	buf.reader = nil
	return engine.NewBool(true), nil
}

// builtinBufferToBytes 将缓冲区内容转换为字节数组
// buffer_to_bytes($buf) → array
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 字节数组，每个元素是0-255的整数值
//
// 注意：返回的是缓冲区内容的副本，不影响当前读取位置
func builtinBufferToBytes(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_to_bytes() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_to_bytes() expects Buffer, got %s", args[0].Type())
	}
	data := buf.data.Bytes()
	values := make([]engine.Value, len(data))
	for i, b := range data {
		values[i] = engine.NewInt(int64(b))
	}
	return engine.NewArray(values), nil
}

// builtinBufferToString 将缓冲区内容转换为字符串
// buffer_to_string($buf) → string
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 缓冲区内容的字符串表示
//
// 注意：直接字节到字符串转换，不做任何编码处理
func builtinBufferToString(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_to_string() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_to_string() expects Buffer, got %s", args[0].Type())
	}
	return engine.NewString(buf.data.String()), nil
}

// builtinBufferNewFrom 从字节数据创建 Buffer 对象
// buffer_new_from($bytes, [endian]) → Buffer
//
// 参数：
//   - args[0]: 字节数组（整数数组）或字符串
//   - args[1]: 可选，字节序字符串（"big"/"little"），默认为 "big"
//
// 返回值：
//   - Buffer 对象，预填充了输入的字节数据
//
// 示例：
//
//	$buf = buffer_new_from([0x48, 0x65, 0x6C, 0x6C, 0x6F])
//	$buf = buffer_new_from("Hello", "little")
func builtinBufferNewFrom(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("buffer_new_from() expects at least 1 argument, got %d", len(args))
	}

	endian := "big"
	if len(args) > 1 {
		endian = strings.ToLower(args[1].String())
	}

	buf := NewBuffer(endian)

	switch args[0].Type() {
	case engine.TypeArray:
		arr := args[0].Array()
		for _, elem := range arr {
			if elem.Type() == engine.TypeInt {
				buf.data.WriteByte(byte(elem.Int()))
			}
		}
	case engine.TypeString:
		buf.data.WriteString(args[0].String())
	default:
		return nil, fmt.Errorf("buffer_new_from() expects array or string, got %s", args[0].Type())
	}

	return buf, nil
}

// builtinBufferWriteInt8 写入有符号8位整数
// buffer_write_int8($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的有符号8位整数值（-128~127）
func builtinBufferWriteInt8(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_int8() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_int8() expects Buffer, got %s", args[0].Type())
	}
	var val int8
	if args[1].Type() == engine.TypeInt {
		val = int8(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = int8(args[1].Float())
	}
	binary.Write(buf.data, buf.order, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteInt16 写入有符号16位整数
// buffer_write_int16($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的有符号16位整数值（-32768~32767）
//
// 注意：按照 Buffer 当前的字节序写入
func builtinBufferWriteInt16(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_int16() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_int16() expects Buffer, got %s", args[0].Type())
	}
	var val int16
	if args[1].Type() == engine.TypeInt {
		val = int16(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = int16(args[1].Float())
	}
	binary.Write(buf.data, buf.order, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferWriteInt32 写入有符号32位整数
// buffer_write_int32($buf, $value) → int (新的缓冲区长度)
//
// 参数：
//   - args[0]: Buffer 对象
//   - args[1]: 要写入的有符号32位整数值
//
// 注意：按照 Buffer 当前的字节序写入
func builtinBufferWriteInt32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("buffer_write_int32() expects 2 arguments, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_write_int32() expects Buffer, got %s", args[0].Type())
	}
	var val int32
	if args[1].Type() == engine.TypeInt {
		val = int32(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		val = int32(args[1].Float())
	}
	binary.Write(buf.data, buf.order, val)
	return engine.NewInt(int64(buf.data.Len())), nil
}

// builtinBufferReadInt8 读取有符号8位整数
// buffer_read_int8($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的有符号8位整数值（-128~127）
//   - null: 缓冲区已读完
func builtinBufferReadInt8(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_int8() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_int8() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v int8
	if err := binary.Read(buf.reader, buf.order, &v); err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewInt(int64(v)), nil
}

// builtinBufferReadInt16 读取有符号16位整数
// buffer_read_int16($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的有符号16位整数值（-32768~32767）
//   - null: 缓冲区不足或读取失败
//
// 注意：按照 Buffer 当前的字节序读取
func builtinBufferReadInt16(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_int16() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_int16() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v int16
	if err := binary.Read(buf.reader, buf.order, &v); err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewInt(int64(v)), nil
}

// builtinBufferReadInt32 读取有符号32位整数
// buffer_read_int32($buf) → int | null
//
// 参数：
//   - args[0]: Buffer 对象
//
// 返回值：
//   - 读取的有符号32位整数值
//   - null: 缓冲区不足或读取失败
//
// 注意：按照 Buffer 当前的字节序读取
func builtinBufferReadInt32(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("buffer_read_int32() expects 1 argument, got %d", len(args))
	}
	buf, ok := args[0].(*BufferValue)
	if !ok {
		return nil, fmt.Errorf("buffer_read_int32() expects Buffer, got %s", args[0].Type())
	}
	buf.ensureReader()
	var v int32
	if err := binary.Read(buf.reader, buf.order, &v); err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewInt(int64(v)), nil
}

// builtinIsBuffer 检查值是否为 Buffer 类型
// is_buffer($value) → bool
//
// 参数：
//   - args[0]: 要检查的值
//
// 返回值：
//   - true: 是 Buffer 对象
//   - false: 不是 Buffer 对象
//
// 示例：
//
//	is_buffer(buffer_new())        → true
//	is_buffer("string")            → false
//	is_buffer([1, 2, 3])           → false
func builtinIsBuffer(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("is_buffer() expects 1 argument, got %d", len(args))
	}
	_, ok := args[0].(*BufferValue)
	return engine.NewBool(ok), nil
}

// BinarySigs returns function signatures for REPL :doc command.
func BinarySigs() map[string]string {
	return map[string]string{
		"pack":   "pack(format, values...) → array  — Pack values into binary bytes",
		"unpack": "unpack(format, bytes) → value  — Unpack binary bytes to values",
		"ord":    "ord(str) → int  — Get ASCII value of first character",
		"chr":    "chr(ascii) → string  — Convert ASCII to character",
	}
}
