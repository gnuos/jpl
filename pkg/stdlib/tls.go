package stdlib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/gnuos/jpl/engine"
)

// =============================================================================
// TLS 模块 - 实现 TLS/SSL 加密通信
// =============================================================================

// TLSSocketValue 表示 TLS 连接对象
type TLSSocketValue struct {
	conn     net.Conn    // TLS 连接（实现了 net.Conn 接口）
	config   *tls.Config // TLS 配置
	isServer bool        // 是否为服务端
}

// Type 返回类型标识
func (s *TLSSocketValue) Type() engine.ValueType { return engine.TypeStream }

// IsNull 检查是否为 null
func (s *TLSSocketValue) IsNull() bool { return s == nil }

// Bool 返回布尔值
func (s *TLSSocketValue) Bool() bool { return s != nil && s.conn != nil }

// Int 返回 fd（TLS 不支持直接获取 fd，返回 0）
func (s *TLSSocketValue) Int() int64 { return 0 }

// Float 返回浮点数
func (s *TLSSocketValue) Float() float64 { return 0 }

// String 返回字符串表示
func (s *TLSSocketValue) String() string {
	if s.isServer {
		return "TLSServer"
	}
	return "TLSClient"
}

// Stringify 返回 JSON 序列化字符串
func (s *TLSSocketValue) Stringify() string { return s.String() }

// Array 返回数组值
func (s *TLSSocketValue) Array() []engine.Value { return nil }

// Object 返回对象值
func (s *TLSSocketValue) Object() map[string]engine.Value { return nil }

// Len 返回长度
func (s *TLSSocketValue) Len() int { return 0 }

// Equals 等于
func (s *TLSSocketValue) Equals(v engine.Value) bool { return false }

// Less 小于
func (s *TLSSocketValue) Less(v engine.Value) bool { return false }

// Greater 大于
func (s *TLSSocketValue) Greater(v engine.Value) bool { return false }

// LessEqual 小于等于
func (s *TLSSocketValue) LessEqual(v engine.Value) bool { return false }

// GreaterEqual 大于等于
func (s *TLSSocketValue) GreaterEqual(v engine.Value) bool { return false }

// ToBigInt 转换为大整数
func (s *TLSSocketValue) ToBigInt() engine.Value { return engine.NewInt(0) }

// ToBigDecimal 转换为大浮点数
func (s *TLSSocketValue) ToBigDecimal() engine.Value { return engine.NewFloat(0) }

// Add 加法
func (s *TLSSocketValue) Add(v engine.Value) engine.Value { return s }

// Sub 减法
func (s *TLSSocketValue) Sub(v engine.Value) engine.Value { return s }

// Mul 乘法
func (s *TLSSocketValue) Mul(v engine.Value) engine.Value { return s }

// Div 除法
func (s *TLSSocketValue) Div(v engine.Value) engine.Value { return s }

// Mod 取模
func (s *TLSSocketValue) Mod(v engine.Value) engine.Value { return s }

// Negate 取反
func (s *TLSSocketValue) Negate() engine.Value { return s }

// RegisterTLS 注册 TLS 函数到引擎
func RegisterTLS(e *engine.Engine) {
	// 连接管理
	e.RegisterFunc("tls_connect", builtinTLSConnect)
	e.RegisterFunc("tls_listen", builtinTLSListen)
	e.RegisterFunc("tls_accept", builtinTLSAccept)
	e.RegisterFunc("tls_close", builtinTLSClose)

	// 数据传输
	e.RegisterFunc("tls_send", builtinTLSSend)
	e.RegisterFunc("tls_recv", builtinTLSRecv)

	// 信息获取
	e.RegisterFunc("tls_get_cipher", builtinTLSGetCipher)
	e.RegisterFunc("tls_get_version", builtinTLSGetVersion)
	e.RegisterFunc("tls_get_cert_info", builtinTLSGetCertInfo)
	e.RegisterFunc("tls_set_cert", builtinTLSSetCert)

	// 证书生成
	e.RegisterFunc("tls_gen_cert", builtinTLSGenCert)

	// 模块注册 - import "tls" 可用
	e.RegisterModule("tls", map[string]engine.GoFunction{
		"connect":       builtinTLSConnect,
		"listen":        builtinTLSListen,
		"accept":        builtinTLSAccept,
		"close":         builtinTLSClose,
		"send":          builtinTLSSend,
		"recv":          builtinTLSRecv,
		"get_cipher":    builtinTLSGetCipher,
		"get_version":   builtinTLSGetVersion,
		"get_cert_info": builtinTLSGetCertInfo,
		"set_cert":      builtinTLSSetCert,
		"gen_cert":      builtinTLSGenCert,
	})
}

// TLSNames 返回 TLS 函数名称列表
func TLSNames() []string {
	return []string{
		"tls_connect", "tls_listen", "tls_accept", "tls_close",
		"tls_send", "tls_recv",
		"tls_get_cipher", "tls_get_version", "tls_get_cert_info", "tls_set_cert",
		"tls_gen_cert",
	}
}

// parseTLSOptions 解析 TLS 选项参数
func parseTLSOptions(args engine.Value) (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, // 最低 TLS 1.2
	}

	if args == nil || args.Type() != engine.TypeObject {
		return config, nil
	}

	obj := args.Object()

	// verify 选项
	if verify, ok := obj["verify"]; ok {
		config.InsecureSkipVerify = !verify.Bool()
	}

	// ca_file 选项
	if caFile, ok := obj["ca_file"]; ok && caFile.Type() == engine.TypeString {
		caData, err := os.ReadFile(caFile.String())
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %v", err)
		}
		config.RootCAs = x509.NewCertPool()
		config.RootCAs.AppendCertsFromPEM(caData)
	}

	// cert_file 和 key_file 选项（客户端证书）
	if certFile, ok := obj["cert_file"]; ok && certFile.Type() == engine.TypeString {
		keyFile := obj["key_file"]
		if keyFile == nil || keyFile.Type() != engine.TypeString {
			return nil, fmt.Errorf("key_file required when cert_file is provided")
		}
		cert, err := tls.LoadX509KeyPair(certFile.String(), keyFile.String())
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// server_name 选项（SNI）
	if serverName, ok := obj["server_name"]; ok && serverName.Type() == engine.TypeString {
		config.ServerName = serverName.String()
	}

	return config, nil
}

// builtinTLSConnect 建立 TLS 客户端连接
// tls_connect(host, port, options?) → TLSSocketValue
//
// 参数：
//   - args[0]: 主机地址（字符串）
//   - args[1]: 端口号（整数）
//   - args[2]: 选项对象（可选）
//   - verify: 是否验证证书（默认 true）
//   - ca_file: CA 证书文件路径
//   - cert_file: 客户端证书文件路径（mTLS）
//   - key_file: 客户端私钥文件路径（mTLS）
//   - server_name: SNI 主机名
//
// 返回值：
//   - TLSSocketValue 对象
//
// 示例：
//
//	$conn = tls_connect("api.example.com", 443)
//	$conn = tls_connect("api.example.com", 443, {verify: true, ca_file: "/path/to/ca.crt"})
func builtinTLSConnect(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("tls_connect() expects at least 2 arguments, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("tls_connect() expects string host, got %s", args[0].Type())
	}

	var port int
	if args[1].Type() == engine.TypeInt {
		port = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		port = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("tls_connect() expects int port, got %s", args[1].Type())
	}

	// 解析选项
	var config *tls.Config
	if len(args) >= 3 {
		var err error
		config, err = parseTLSOptions(args[2])
		if err != nil {
			return nil, err
		}
	} else {
		config = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		}
	}

	// 建立连接
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := tls.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("tls_connect() failed: %v", err)
	}

	return &TLSSocketValue{
		conn:     conn,
		config:   config,
		isServer: false,
	}, nil
}

// builtinTLSListen 创建 TLS 服务端监听
// tls_listen(port, cert, key, options?) → TLSSocketValue
//
// 参数：
//   - args[0]: 端口号（整数）
//   - args[1]: 服务器证书文件路径（字符串）
//   - args[2]: 服务器私钥文件路径（字符串）
//   - args[3]: 选项对象（可选）
//   - host: 绑定地址（默认 "0.0.0.0"）
//
// 返回值：
//   - TLSSocketValue 对象（服务端监听状态）
//
// 示例：
//
//	$server = tls_listen(8443, "/path/to/server.crt", "/path/to/server.key")
func builtinTLSListen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("tls_listen() expects at least 3 arguments, got %d", len(args))
	}

	var port int
	if args[0].Type() == engine.TypeInt {
		port = int(args[0].Int())
	} else if args[0].Type() == engine.TypeFloat {
		port = int(args[0].Float())
	} else {
		return nil, fmt.Errorf("tls_listen() expects int port, got %s", args[0].Type())
	}

	certFile := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("tls_listen() expects string cert_file, got %s", args[1].Type())
	}

	keyFile := args[2].String()
	if args[2].Type() != engine.TypeString {
		return nil, fmt.Errorf("tls_listen() expects string key_file, got %s", args[2].Type())
	}

	// 加载证书
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("tls_listen() failed to load certificate: %v", err)
	}

	// 解析选项
	host := "0.0.0.0"
	if len(args) >= 4 && args[3].Type() == engine.TypeObject {
		obj := args[3].Object()
		if h, ok := obj["host"]; ok && h.Type() == engine.TypeString {
			host = h.String()
		}
	}

	config := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}

	// 创建监听
	address := fmt.Sprintf("%s:%d", host, port)
	listener, err := tls.Listen("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("tls_listen() failed: %v", err)
	}

	// 注意：这里我们返回一个特殊的监听 socket
	// 它不是一个连接，但可以用于 accept
	return &tlsListenerValue{
		listener: listener,
		config:   config,
	}, nil
}

// tlsListenerValue 表示 TLS 监听对象
type tlsListenerValue struct {
	listener net.Listener
	config   *tls.Config
}

// Type 返回类型标识
func (l *tlsListenerValue) Type() engine.ValueType { return engine.TypeStream }

// IsNull 检查是否为 null
func (l *tlsListenerValue) IsNull() bool { return l == nil }

// Bool 返回布尔值
func (l *tlsListenerValue) Bool() bool { return l != nil && l.listener != nil }

// Int 返回 0
func (l *tlsListenerValue) Int() int64 { return 0 }

// Float 返回 0
func (l *tlsListenerValue) Float() float64 { return 0 }

// String 返回字符串表示
func (l *tlsListenerValue) String() string { return "TLSListener" }

// Stringify 返回 JSON 序列化字符串
func (l *tlsListenerValue) Stringify() string { return l.String() }

// Array 返回数组值
func (l *tlsListenerValue) Array() []engine.Value { return nil }

// Object 返回对象值
func (l *tlsListenerValue) Object() map[string]engine.Value { return nil }

// Len 返回长度
func (l *tlsListenerValue) Len() int { return 0 }

// Equals 等于
func (l *tlsListenerValue) Equals(v engine.Value) bool { return false }

// Less 小于
func (l *tlsListenerValue) Less(v engine.Value) bool { return false }

// Greater 大于
func (l *tlsListenerValue) Greater(v engine.Value) bool { return false }

// LessEqual 小于等于
func (l *tlsListenerValue) LessEqual(v engine.Value) bool { return false }

// GreaterEqual 大于等于
func (l *tlsListenerValue) GreaterEqual(v engine.Value) bool { return false }

// ToBigInt 转换为大整数
func (l *tlsListenerValue) ToBigInt() engine.Value { return engine.NewInt(0) }

// ToBigDecimal 转换为大浮点数
func (l *tlsListenerValue) ToBigDecimal() engine.Value { return engine.NewFloat(0) }

// Add 加法
func (l *tlsListenerValue) Add(v engine.Value) engine.Value { return l }

// Sub 减法
func (l *tlsListenerValue) Sub(v engine.Value) engine.Value { return l }

// Mul 乘法
func (l *tlsListenerValue) Mul(v engine.Value) engine.Value { return l }

// Div 除法
func (l *tlsListenerValue) Div(v engine.Value) engine.Value { return l }

// Mod 取模
func (l *tlsListenerValue) Mod(v engine.Value) engine.Value { return l }

// Negate 取反
func (l *tlsListenerValue) Negate() engine.Value { return l }

// builtinTLSAccept 接受 TLS 连接
// tls_accept(server) → TLSSocketValue
//
// 参数：
//   - args[0]: TLS 监听对象（由 tls_listen 创建）
//
// 返回值：
//   - TLSSocketValue 对象（新的客户端连接）
func builtinTLSAccept(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tls_accept() expects 1 argument, got %d", len(args))
	}

	listener, ok := args[0].(*tlsListenerValue)
	if !ok {
		return nil, fmt.Errorf("tls_accept() expects TLS listener, got %s", args[0].Type())
	}

	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("tls_accept() failed: %v", err)
	}

	return &TLSSocketValue{
		conn:     conn,
		config:   listener.config,
		isServer: true,
	}, nil
}

// builtinTLSClose 关闭 TLS 连接
// tls_close(conn) → bool
//
// 参数：
//   - args[0]: TLS 连接对象
//
// 返回值：
//   - true: 关闭成功
func builtinTLSClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tls_close() expects 1 argument, got %d", len(args))
	}

	// 关闭 TLS 连接
	if sock, ok := args[0].(*TLSSocketValue); ok {
		if sock.conn != nil {
			sock.conn.Close()
		}
		return engine.NewBool(true), nil
	}

	// 关闭 TLS 监听
	if listener, ok := args[0].(*tlsListenerValue); ok {
		if listener.listener != nil {
			listener.listener.Close()
		}
		return engine.NewBool(true), nil
	}

	return nil, fmt.Errorf("tls_close() expects TLS socket or listener, got %s", args[0].Type())
}

// builtinTLSSend 发送加密数据
// tls_send(conn, data) → bytes_sent
//
// 参数：
//   - args[0]: TLS 连接对象
//   - args[1]: 要发送的数据（字符串）
//
// 返回值：
//   - 实际发送的字节数
func builtinTLSSend(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("tls_send() expects 2 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*TLSSocketValue)
	if !ok {
		return nil, fmt.Errorf("tls_send() expects TLS socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return nil, fmt.Errorf("tls_send() socket is not connected")
	}

	data := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("tls_send() expects string data, got %s", args[1].Type())
	}

	n, err := sock.conn.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("tls_send() failed: %v", err)
	}

	return engine.NewInt(int64(n)), nil
}

// builtinTLSRecv 接收解密数据
// tls_recv(conn, len) → data
//
// 参数：
//   - args[0]: TLS 连接对象
//   - args[1]: 最大接收字节数（整数）
//
// 返回值：
//   - 接收的数据（字符串）
func builtinTLSRecv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("tls_recv() expects 2 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*TLSSocketValue)
	if !ok {
		return nil, fmt.Errorf("tls_recv() expects TLS socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return nil, fmt.Errorf("tls_recv() socket is not connected")
	}

	var length int
	if args[1].Type() == engine.TypeInt {
		length = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		length = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("tls_recv() expects int length, got %s", args[1].Type())
	}

	buffer := make([]byte, length)
	n, err := sock.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("tls_recv() failed: %v", err)
	}

	return engine.NewString(string(buffer[:n])), nil
}

// builtinTLSGetCipher 获取协商的加密套件
// tls_get_cipher(conn) → cipher_suite
//
// 参数：
//   - args[0]: TLS 连接对象
//
// 返回值：
//   - 加密套件名称（字符串）
func builtinTLSGetCipher(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tls_get_cipher() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*TLSSocketValue)
	if !ok {
		return nil, fmt.Errorf("tls_get_cipher() expects TLS socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return engine.NewString(""), nil
	}

	// 获取 TLS 连接状态
	if tlsConn, ok := sock.conn.(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		return engine.NewString(tls.CipherSuiteName(state.CipherSuite)), nil
	}

	return engine.NewString(""), nil
}

// builtinTLSGetVersion 获取 TLS 版本
// tls_get_version(conn) → version
//
// 参数：
//   - args[0]: TLS 连接对象
//
// 返回值：
//   - TLS 版本字符串，如 "TLS 1.2", "TLS 1.3"
func builtinTLSGetVersion(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tls_get_version() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*TLSSocketValue)
	if !ok {
		return nil, fmt.Errorf("tls_get_version() expects TLS socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return engine.NewString(""), nil
	}

	if tlsConn, ok := sock.conn.(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		version := "Unknown"
		switch state.Version {
		case tls.VersionTLS10:
			version = "TLS 1.0"
		case tls.VersionTLS11:
			version = "TLS 1.1"
		case tls.VersionTLS12:
			version = "TLS 1.2"
		case tls.VersionTLS13:
			version = "TLS 1.3"
		}
		return engine.NewString(version), nil
	}

	return engine.NewString(""), nil
}

// builtinTLSGetCertInfo 获取证书信息
// tls_get_cert_info(conn) → cert_info
//
// 参数：
//   - args[0]: TLS 连接对象
//
// 返回值：
//   - 证书信息对象，包含 subject, issuer, not_before, not_after, dns_names 等
func builtinTLSGetCertInfo(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tls_get_cert_info() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*TLSSocketValue)
	if !ok {
		return nil, fmt.Errorf("tls_get_cert_info() expects TLS socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return engine.NewObject(map[string]engine.Value{}), nil
	}

	if tlsConn, ok := sock.conn.(*tls.Conn); ok {
		state := tlsConn.ConnectionState()
		if len(state.PeerCertificates) > 0 {
			cert := state.PeerCertificates[0]
			info := map[string]engine.Value{
				"subject":       engine.NewString(cert.Subject.String()),
				"issuer":        engine.NewString(cert.Issuer.String()),
				"not_before":    engine.NewString(cert.NotBefore.String()),
				"not_after":     engine.NewString(cert.NotAfter.String()),
				"serial_number": engine.NewString(cert.SerialNumber.String()),
			}

			// DNS 名称
			dnsNames := make([]engine.Value, len(cert.DNSNames))
			for i, name := range cert.DNSNames {
				dnsNames[i] = engine.NewString(name)
			}
			info["dns_names"] = engine.NewArray(dnsNames)

			return engine.NewObject(info), nil
		}
	}

	return engine.NewObject(map[string]engine.Value{}), nil
}

// builtinTLSSetCert 设置客户端证书（用于 mTLS）
// tls_set_cert(conn, cert_file, key_file) → bool
//
// 参数：
//   - args[0]: TLS 连接对象（尚未握手）
//   - args[1]: 客户端证书文件路径
//   - args[2]: 客户端私钥文件路径
//
// 返回值：
//   - true: 设置成功
func builtinTLSSetCert(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("tls_set_cert() expects 3 arguments, got %d", len(args))
	}

	// 注意：Go 的 tls.Conn 不支持动态设置证书
	// 这个函数主要是为了 API 完整性，实际应在连接前通过 options 设置
	// 这里返回错误提示用户使用 options 方式

	return nil, fmt.Errorf("tls_set_cert() should be used via options in tls_connect(): use {cert_file: '...', key_file: '...'}")
}

// builtinTLSGenCert 生成自签名证书
// tls_gen_cert(options?) → {cert_path, key_path}
//
// 参数：
//   - args[0]: 选项对象（可选）
//   - bits: RSA 密钥位数（默认 2048）
//   - days: 证书有效期天数（默认 365）
//   - common_name: CN 字段（默认 "JPL Generated"）
//   - out_dir: 输出目录（默认系统临时目录）
//   - out_prefix: 文件名前缀（默认 "jpl_tls"）
//
// 返回值：
//   - 对象，包含 cert_path 和 key_path
//
// 示例：
//
//	$paths = tls_gen_cert({bits: 4096, days: 730, common_name: "My Server"})
//	println($paths.cert_path)  # → /tmp/jpl_tls_xxx.crt
//	println($paths.key_path)   # → /tmp/jpl_tls_xxx.key
func builtinTLSGenCert(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	// 默认选项
	bits := 2048
	days := 365
	commonName := "JPL Generated"
	outDir := os.TempDir()
	outPrefix := "jpl_tls"

	// 解析选项
	if len(args) >= 1 && args[0].Type() == engine.TypeObject {
		obj := args[0].Object()

		if b, ok := obj["bits"]; ok && b.Type() == engine.TypeInt {
			bits = int(b.Int())
			if bits < 1024 {
				bits = 1024
			}
		}

		if d, ok := obj["days"]; ok && d.Type() == engine.TypeInt {
			days = int(d.Int())
		}

		if cn, ok := obj["common_name"]; ok && cn.Type() == engine.TypeString {
			commonName = cn.String()
		}

		if dir, ok := obj["out_dir"]; ok && dir.Type() == engine.TypeString {
			outDir = dir.String()
		}

		if prefix, ok := obj["out_prefix"]; ok && prefix.Type() == engine.TypeString {
			outPrefix = prefix.String()
		}
	}

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to generate private key: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Duration(days) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 生成证书
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to create certificate: %v", err)
	}

	// 生成文件名
	timestamp := time.Now().Unix()
	certPath := fmt.Sprintf("%s/%s_%d.crt", outDir, outPrefix, timestamp)
	keyPath := fmt.Sprintf("%s/%s_%d.key", outDir, outPrefix, timestamp)

	// 写入证书文件
	certFile, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to create cert file: %v", err)
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to write certificate: %v", err)
	}

	// 写入私钥文件
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to create key file: %v", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return nil, fmt.Errorf("tls_gen_cert() failed to write private key: %v", err)
	}

	// 返回路径
	return engine.NewObject(map[string]engine.Value{
		"cert_path": engine.NewString(certPath),
		"key_path":  engine.NewString(keyPath),
	}), nil
}
