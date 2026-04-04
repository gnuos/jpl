package stdlib

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/gnuos/jpl/engine"
)

// RegisterIP 注册 IP 地址处理函数到引擎
// 提供 IPv4/IPv6 地址解析、转换、验证等功能
func RegisterIP(e *engine.Engine) {
	e.RegisterFunc("ip2long", builtinIP2Long)
	e.RegisterFunc("long2ip", builtinLong2IP)
	e.RegisterFunc("ip_parse", builtinIPParse)
	e.RegisterFunc("ip_to_hex", builtinIPToHex)
	e.RegisterFunc("ip_to_bin", builtinIPToBin)
	e.RegisterFunc("ip_from_hex", builtinIPFromHex)
	e.RegisterFunc("ip_from_bin", builtinIPFromBin)
	e.RegisterFunc("ip_version", builtinIPVersion)
	e.RegisterFunc("ip_valid", builtinIPValid)

	// 模块注册 — import "ip" 可用
	e.RegisterModule("ip", map[string]engine.GoFunction{
		"ip2long":  builtinIP2Long,
		"long2ip":  builtinLong2IP,
		"parse":    builtinIPParse,
		"to_hex":   builtinIPToHex,
		"to_bin":   builtinIPToBin,
		"from_hex": builtinIPFromHex,
		"from_bin": builtinIPFromBin,
		"version":  builtinIPVersion,
		"valid":    builtinIPValid,
	})
}

// IPNames 返回 IP 处理函数名称列表
func IPNames() []string {
	return []string{
		"ip2long", "long2ip", "ip_parse",
		"ip_to_hex", "ip_to_bin", "ip_from_hex", "ip_from_bin",
		"ip_version", "ip_valid",
	}
}

// builtinIP2Long 将 IPv4 地址转换为长整型数字
// ip2long($ip) → int
//
// 参数：
//   - args[0]: IPv4 地址字符串，如 "192.168.1.1"
//
// 返回值：
//   - 32位无符号整数（对于 IPv6 返回大整数）
//   - null: 无效 IP 地址
//
// 示例：
//
//	ip2long("192.168.1.1") → 3232235777
//	ip2long("127.0.0.1") → 2130706433
func builtinIP2Long(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip2long() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip2long() expects string, got %s", args[0].Type())
	}

	ip := args[0].String()
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return engine.NewNull(), nil
	}

	// 转换为 IPv4 格式
	parsedIP = parsedIP.To4()
	if parsedIP == nil {
		// IPv6 地址
		// 将 16 字节 IPv6 转换为两个 64 位整数（大整数）
		high := binary.BigEndian.Uint64(parsedIP[:8])
		low := binary.BigEndian.Uint64(parsedIP[8:])

		// 组合成 128 位整数（用字符串表示）
		result := fmt.Sprintf("%x%016x", high, low)
		// 转换为十进制字符串
		val, _ := strconv.ParseUint(result, 16, 64)
		return engine.NewInt(int64(val)), nil
	}

	// IPv4: 将 4 字节转换为 32 位整数
	val := binary.BigEndian.Uint32(parsedIP)
	return engine.NewInt(int64(val)), nil
}

// builtinLong2IP 将长整型数字转换为 IPv4 地址
// long2ip($long) → string
//
// 参数：
//   - args[0]: 32位无符号整数
//
// 返回值：
//   - IPv4 地址字符串
//   - null: 无效输入
//
// 示例：
//
//	long2ip(3232235777) → "192.168.1.1"
//	long2ip(2130706433) → "127.0.0.1"
func builtinLong2IP(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("long2ip() expects 1 argument, got %d", len(args))
	}

	var val uint32
	if args[0].Type() == engine.TypeInt {
		val = uint32(args[0].Int())
	} else if args[0].Type() == engine.TypeFloat {
		val = uint32(args[0].Float())
	} else {
		return engine.NewNull(), nil
	}

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, val)

	return engine.NewString(ip.String()), nil
}

// builtinIPParse 解析 IP 地址并返回详细信息
// ip_parse($ip) → object | null
//
// 参数：
//   - args[0]: IP 地址字符串
//
// 返回值：
//   - 对象包含以下字段：
//   - ip: 原始 IP 字符串
//   - version: 4 或 6
//   - parts: 数组，IPv4 为 4 个数字，IPv6 为 8 组 16进制
//   - type: "unicast", "multicast", "loopback", "private" 等
//
// 示例：
//
//	ip_parse("192.168.1.1")
//	→ {ip: "192.168.1.1", version: 4, parts: [192, 168, 1, 1], type: "private"}
func builtinIPParse(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_parse() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_parse() expects string, got %s", args[0].Type())
	}

	ipStr := args[0].String()
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return engine.NewNull(), nil
	}

	result := make(map[string]engine.Value)
	result["ip"] = engine.NewString(ipStr)

	// 检查 IPv4
	ip4 := parsedIP.To4()
	if ip4 != nil {
		result["version"] = engine.NewInt(4)

		// 解析各部分
		parts := make([]engine.Value, 4)
		for i := 0; i < 4; i++ {
			parts[i] = engine.NewInt(int64(ip4[i]))
		}
		result["parts"] = engine.NewArray(parts)

		// 判断类型
		ipType := classifyIPv4(ip4)
		result["type"] = engine.NewString(ipType)
	} else {
		// IPv6
		result["version"] = engine.NewInt(6)

		// 解析各部分（8组，每组4位十六进制）
		parts := make([]engine.Value, 8)
		for i := 0; i < 8; i++ {
			val := binary.BigEndian.Uint16(parsedIP[i*2 : (i+1)*2])
			parts[i] = engine.NewInt(int64(val))
		}
		result["parts"] = engine.NewArray(parts)

		// 判断类型
		ipType := classifyIPv6(parsedIP)
		result["type"] = engine.NewString(ipType)
	}

	return engine.NewObject(result), nil
}

// builtinIPToHex 将 IP 地址转换为十六进制字符串
// ip_to_hex($ip) → string
//
// 参数：
//   - args[0]: IP 地址字符串
//
// 返回值：
//   - 十六进制字符串（IPv4 为 8 位，IPv6 为 32 位，不带分隔符）
//   - null: 无效 IP
//
// 示例：
//
//	ip_to_hex("192.168.1.1") → "c0a80101"
//	ip_to_hex("2001:db8::1") → "20010db8000000000000000000000001"
func builtinIPToHex(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_to_hex() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_to_hex() expects string, got %s", args[0].Type())
	}

	ipStr := args[0].String()
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return engine.NewNull(), nil
	}

	// 转换为字符串形式的十六进制
	var hexStr string
	if ip4 := parsedIP.To4(); ip4 != nil {
		// IPv4: 4字节
		hexStr = fmt.Sprintf("%02x%02x%02x%02x", ip4[0], ip4[1], ip4[2], ip4[3])
	} else {
		// IPv6: 16字节
		hexStr = fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x",
			parsedIP[0], parsedIP[1], parsedIP[2], parsedIP[3],
			parsedIP[4], parsedIP[5], parsedIP[6], parsedIP[7],
			parsedIP[8], parsedIP[9], parsedIP[10], parsedIP[11],
			parsedIP[12], parsedIP[13], parsedIP[14], parsedIP[15])
	}

	return engine.NewString(hexStr), nil
}

// builtinIPToBin 将 IP 地址转换为二进制字符串
// ip_to_bin($ip) → string
//
// 参数：
//   - args[0]: IP 地址字符串
//
// 返回值：
//   - 二进制字符串（IPv4 为 32 位，IPv6 为 128 位）
//   - null: 无效 IP
//
// 示例：
//
//	ip_to_bin("192.168.1.1") → "11000000101010000000000100000001"
//	ip_to_bin("127.0.0.1") → "01111111000000000000000000000001"
func builtinIPToBin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_to_bin() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_to_bin() expects string, got %s", args[0].Type())
	}

	ipStr := args[0].String()
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return engine.NewNull(), nil
	}

	// 获取字节数据
	var data []byte
	if ip4 := parsedIP.To4(); ip4 != nil {
		data = ip4
	} else {
		data = parsedIP
	}

	// 转换为二进制字符串
	var result strings.Builder
	for _, b := range data {
		result.WriteString(fmt.Sprintf("%08b", b))
	}

	return engine.NewString(result.String()), nil
}

// builtinIPFromHex 将十六进制字符串转换为 IP 地址
// ip_from_hex($hex) → string | null
//
// 参数：
//   - args[0]: 十六进制字符串（8 位为 IPv4，32 位为 IPv6）
//
// 返回值：
//   - IP 地址字符串
//   - null: 无效十六进制或长度错误
//
// 示例：
//
//	ip_from_hex("c0a80101") → "192.168.1.1"
//	ip_from_hex("20010db8000000000000000000000001") → "2001:db8::1"
func builtinIPFromHex(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_from_hex() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_from_hex() expects string, got %s", args[0].Type())
	}

	hexStr := strings.ToLower(args[0].String())

	// 去除可能的 0x 前缀
	hexStr = strings.TrimPrefix(hexStr, "0x")
	hexStr = strings.TrimPrefix(hexStr, "0X")

	switch len(hexStr) {
	case 8:
		// IPv4
		data, err := parseHex(hexStr)
		if err != nil {
			return engine.NewNull(), nil
		}
		ip := net.IP(data)
		return engine.NewString(ip.String()), nil

	case 32:
		// IPv6
		data, err := parseHex(hexStr)
		if err != nil {
			return engine.NewNull(), nil
		}
		ip := net.IP(data)
		return engine.NewString(ip.String()), nil

	default:
		return engine.NewNull(), nil
	}
}

// builtinIPFromBin 将二进制字符串转换为 IP 地址
// ip_from_bin($bin) → string | null
//
// 参数：
//   - args[0]: 二进制字符串（32 位为 IPv4，128 位为 IPv6）
//
// 返回值：
//   - IP 地址字符串
//   - null: 无效二进制字符串或长度错误
//
// 示例：
//
//	ip_from_bin("11000000101010000000000100000001") → "192.168.1.1"
func builtinIPFromBin(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_from_bin() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_from_bin() expects string, got %s", args[0].Type())
	}

	binStr := args[0].String()

	switch len(binStr) {
	case 32:
		// IPv4
		data, err := parseBinary(binStr)
		if err != nil {
			return engine.NewNull(), nil
		}
		ip := net.IP(data)
		return engine.NewString(ip.String()), nil

	case 128:
		// IPv6
		data, err := parseBinary(binStr)
		if err != nil {
			return engine.NewNull(), nil
		}
		ip := net.IP(data)
		return engine.NewString(ip.String()), nil

	default:
		return engine.NewNull(), nil
	}
}

// builtinIPVersion 检测 IP 地址版本
// ip_version($ip) → int | null
//
// 参数：
//   - args[0]: IP 地址字符串
//
// 返回值：
//   - 4: IPv4 地址
//   - 6: IPv6 地址
//   - null: 无效 IP
//
// 示例：
//
//	ip_version("192.168.1.1") → 4
//	ip_version("2001:db8::1") → 6
//	ip_version("invalid") → null
func builtinIPVersion(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_version() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("ip_version() expects string, got %s", args[0].Type())
	}

	ipStr := args[0].String()
	parsedIP := net.ParseIP(ipStr)
	if parsedIP == nil {
		return engine.NewNull(), nil
	}

	if parsedIP.To4() != nil {
		return engine.NewInt(4), nil
	}
	return engine.NewInt(6), nil
}

// builtinIPValid 验证 IP 地址格式是否有效
// ip_valid($ip) → bool
//
// 参数：
//   - args[0]: IP 地址字符串
//
// 返回值：
//   - true: 有效 IP 地址
//   - false: 无效 IP 地址
//
// 示例：
//
//	ip_valid("192.168.1.1") → true
//	ip_valid("256.1.1.1") → false
//	ip_valid("2001:db8::1") → true
//	ip_valid("not an ip") → false
func builtinIPValid(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ip_valid() expects 1 argument, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return engine.NewBool(false), nil
	}

	ipStr := args[0].String()
	parsedIP := net.ParseIP(ipStr)

	return engine.NewBool(parsedIP != nil), nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// classifyIPv4 分类 IPv4 地址类型
func classifyIPv4(ip net.IP) string {
	if ip.Equal(net.IPv4(127, 0, 0, 1)) {
		return "loopback"
	}
	if ip.IsPrivate() {
		return "private"
	}
	if ip.IsMulticast() {
		return "multicast"
	}
	if ip.IsLinkLocalUnicast() {
		return "linklocal"
	}
	if ip.IsLoopback() {
		return "loopback"
	}
	return "unicast"
}

// classifyIPv6 分类 IPv6 地址类型
func classifyIPv6(ip net.IP) string {
	if ip.IsLoopback() {
		return "loopback"
	}
	if ip.IsPrivate() {
		return "private"
	}
	if ip.IsMulticast() {
		return "multicast"
	}
	if ip.IsLinkLocalUnicast() {
		return "linklocal"
	}
	if ip.IsInterfaceLocalMulticast() {
		return "interface_local_multicast"
	}
	if ip.IsLinkLocalMulticast() {
		return "link_local_multicast"
	}
	if ip.IsGlobalUnicast() {
		return "global_unicast"
	}
	return "unicast"
}

// parseHex 将十六进制字符串解析为字节数组
func parseHex(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hex length")
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return nil, err
		}
		result[i/2] = byte(b)
	}
	return result, nil
}

// parseBinary 将二进制字符串解析为字节数组
func parseBinary(s string) ([]byte, error) {
	if len(s)%8 != 0 {
		return nil, fmt.Errorf("invalid binary length")
	}

	result := make([]byte, len(s)/8)
	for i := 0; i < len(s); i += 8 {
		b, err := strconv.ParseUint(s[i:i+8], 2, 8)
		if err != nil {
			return nil, err
		}
		result[i/8] = byte(b)
	}
	return result, nil
}

// IPSigs returns function signatures for REPL :doc command.
func IPSigs() map[string]string {
	return map[string]string{
		"ip_parse":    "ip_parse(ip) → object  — Parse IP address details",
		"ip_format":   "ip_format(ip, format) → string  — Format IP address",
		"ip_validate": "ip_validate(ip) → bool  — Validate IP address format",
		"ip_in_range": "ip_in_range(ip, range) → bool  — Check if IP is in range",
	}
}
