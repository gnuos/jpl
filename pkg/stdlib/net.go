package stdlib

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/gnuos/jpl/engine"
)

// RegisterNet 注册网络函数到引擎
// 提供 TCP、Unix Domain Socket、UDP 三种传输层的统一网络编程接口
func RegisterNet(e *engine.Engine) {
	// TCP
	e.RegisterFunc("net_tcp_listen", builtinNetTcpListen)
	e.RegisterFunc("net_tcp_connect", builtinNetTcpConnect)
	e.RegisterFunc("net_tcp_accept", builtinNetTcpAccept)

	// Unix Domain
	e.RegisterFunc("net_unix_listen", builtinNetUnixListen)
	e.RegisterFunc("net_unix_connect", builtinNetUnixConnect)
	e.RegisterFunc("net_unix_accept", builtinNetUnixAccept)

	// UDP
	e.RegisterFunc("net_udp_bind", builtinNetUdpBind)
	e.RegisterFunc("net_udp_sendto", builtinNetUdpSendto)
	e.RegisterFunc("net_udp_recvfrom", builtinNetUdpRecvfrom)

	// 通用
	e.RegisterFunc("net_send", builtinNetSend)
	e.RegisterFunc("net_recv", builtinNetRecv)
	e.RegisterFunc("net_close", builtinNetClose)
	e.RegisterFunc("net_getsockname", builtinNetGetsockname)
	e.RegisterFunc("net_getpeername", builtinNetGetpeername)
	e.RegisterFunc("net_set_nonblock", builtinNetSetNonblock)
	e.RegisterFunc("net_is_unix", builtinNetIsUnix)

	// 网络事件注册（使用通用接口）
	RegisterNetEvents(e)

	// 模块注册 — import "net" 可用
	e.RegisterModule("net", map[string]engine.GoFunction{
		"tcp_listen":   builtinNetTcpListen,
		"tcp_connect":  builtinNetTcpConnect,
		"tcp_accept":   builtinNetTcpAccept,
		"unix_listen":  builtinNetUnixListen,
		"unix_connect": builtinNetUnixConnect,
		"unix_accept":  builtinNetUnixAccept,
		"udp_bind":     builtinNetUdpBind,
		"udp_sendto":   builtinNetUdpSendto,
		"udp_recvfrom": builtinNetUdpRecvfrom,
		"send":         builtinNetSend,
		"recv":         builtinNetRecv,
		"close":        builtinNetClose,
		"getsockname":  builtinNetGetsockname,
		"getpeername":  builtinNetGetpeername,
		"set_nonblock": builtinNetSetNonblock,
		"is_unix":      builtinNetIsUnix,
	})
}

// NetNames 返回网络函数名称列表
func NetNames() []string {
	return []string{
		// TCP
		"net_tcp_listen", "net_tcp_connect", "net_tcp_accept",
		// Unix Domain
		"net_unix_listen", "net_unix_connect", "net_unix_accept",
		// UDP
		"net_udp_bind", "net_udp_sendto", "net_udp_recvfrom",
		// 通用
		"net_send", "net_recv", "net_close",
		"net_getsockname", "net_getpeername", "net_set_nonblock", "net_is_unix",
		// 网络事件（通用接口封装）
		"net_on_accept", "net_on_read", "net_on_write",
		"net_off", "net_off_read", "net_off_write",
	}
}

// NetSocketValue 表示网络 socket 对象
// 统一封装 TCP、Unix Domain、UDP 三种 socket 类型
//
// 设计说明：
// - 通过 isUnix 和 isUDP 标志区分 socket 类型
// - TCP/Unix 使用 conn (net.Conn) 进行连接通信
// - 服务器使用 listener (net.Listener) 接受连接
// - UDP 使用 udpConn (*net.UDPConn) 进行无连接通信
// - fd 字段用于与事件循环集成（通过 syscall 操作）
//
// 使用场景：
// 1. 同步阻塞模式：直接调用 net_send/net_recv
// 2. 异步非阻塞模式：与 ev 模块配合，通过 fd 注册事件处理器
//
// 注意：socket 对象实现了 engine.Value 接口，可以在 JPL 中作为值传递
type NetSocketValue struct {
	fd       int
	isUnix   bool
	isUDP    bool
	conn     net.Conn     // TCP/Unix client
	listener net.Listener // TCP/Unix server
	udpConn  *net.UDPConn // UDP
	udpAddr  *net.UDPAddr // UDP 绑定地址
}

// Type 返回类型标识
func (s *NetSocketValue) Type() engine.ValueType { return engine.TypeStream }

// IsNull 检查是否为 null
func (s *NetSocketValue) IsNull() bool { return s == nil }

// Bool 返回布尔值
func (s *NetSocketValue) Bool() bool { return s != nil }

// Int 返回 fd
func (s *NetSocketValue) Int() int64 { return int64(s.fd) }

// Float 返回浮点数
func (s *NetSocketValue) Float() float64 { return float64(s.fd) }

// String 返回字符串表示
func (s *NetSocketValue) String() string {
	if s.isUnix {
		return fmt.Sprintf("UnixSocket(%d)", s.fd)
	} else if s.isUDP {
		return fmt.Sprintf("UDPSocket(%d)", s.fd)
	}
	return fmt.Sprintf("TCPSocket(%d)", s.fd)
}

// Stringify 返回 JSON 序列化字符串
func (s *NetSocketValue) Stringify() string { return s.String() }

// Array 返回数组值
func (s *NetSocketValue) Array() []engine.Value { return nil }

// Object 返回对象值
func (s *NetSocketValue) Object() map[string]engine.Value { return nil }

// Len 返回长度
func (s *NetSocketValue) Len() int { return 0 }

// Equals 等于
func (s *NetSocketValue) Equals(v engine.Value) bool { return false }

// Less 小于
func (s *NetSocketValue) Less(v engine.Value) bool { return false }

// Greater 大于
func (s *NetSocketValue) Greater(v engine.Value) bool { return false }

// LessEqual 小于等于
func (s *NetSocketValue) LessEqual(v engine.Value) bool { return false }

// GreaterEqual 大于等于
func (s *NetSocketValue) GreaterEqual(v engine.Value) bool { return false }

// ToBigInt 转换为大整数
func (s *NetSocketValue) ToBigInt() engine.Value { return engine.NewInt(0) }

// ToBigDecimal 转换为大十进制数
func (s *NetSocketValue) ToBigDecimal() engine.Value { return engine.NewFloat(0) }

// Add 添加值
func (s *NetSocketValue) Add(v engine.Value) engine.Value { return s }

// Sub 减去值
func (s *NetSocketValue) Sub(v engine.Value) engine.Value { return s }

// Mul 乘以值
func (s *NetSocketValue) Mul(v engine.Value) engine.Value { return s }

// Div 除以值
func (s *NetSocketValue) Div(v engine.Value) engine.Value { return s }

// Mod 取模
func (s *NetSocketValue) Mod(v engine.Value) engine.Value { return s }

// Negate 取反
func (s *NetSocketValue) Negate() engine.Value { return s }

// getFdFromConn 从 net.Conn 获取底层文件描述符
//
// 用于：
// - 与事件循环集成（ev_attach 需要 fd）
// - 设置非阻塞模式
// - 其他底层 socket 操作
//
// 支持类型：TCPConn、UnixConn
// 返回 -1 表示获取失败
func getFdFromConn(conn net.Conn) int {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		file, err := tcpConn.File()
		if err != nil {
			return -1
		}
		fd := int(file.Fd())
		return fd
	}
	if unixConn, ok := conn.(*net.UnixConn); ok {
		file, err := unixConn.File()
		if err != nil {
			return -1
		}
		fd := int(file.Fd())
		return fd
	}
	return -1
}

// getFdFromListener 从 net.Listener 获取底层文件描述符
//
// 用于服务器 socket 的事件循环集成
// 支持类型：TCPListener、UnixListener
// 返回 -1 表示获取失败
func getFdFromListener(ln net.Listener) int {
	if tcpListener, ok := ln.(*net.TCPListener); ok {
		file, err := tcpListener.File()
		if err != nil {
			return -1
		}
		// 不关闭文件，因为 fd 会被 epoll 使用
		// file.Close() 会导致 epoll 注册的 fd 无效
		fd := int(file.Fd())
		return fd
	}
	if unixListener, ok := ln.(*net.UnixListener); ok {
		file, err := unixListener.File()
		if err != nil {
			return -1
		}
		fd := int(file.Fd())
		return fd
	}
	return -1
}

// builtinNetTcpListen 创建 TCP 监听 socket（服务器）
// net_tcp_listen(host, port) → socket
//
// 参数：
//   - args[0]: 主机地址（字符串），如 "0.0.0.0" 或 "127.0.0.1"
//   - args[1]: 端口号（整数），如 8080
//
// 返回值：
//   - NetSocketValue 对象，类型为服务器监听 socket
//
// 使用流程：
//
//	$server = net_tcp_listen("0.0.0.0", 8080)
//	$client = net_tcp_accept($server)
//
// 错误：地址已被占用、权限不足（端口 < 1024）等
func builtinNetTcpListen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_tcp_listen() expects 2 arguments, got %d", len(args))
	}

	host := args[0].String()
	var port int
	if args[1].Type() == engine.TypeInt {
		port = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		port = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("net_tcp_listen() expects int port, got %s", args[1].Type())
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("net_tcp_listen() failed: %v", err)
	}

	fd := getFdFromListener(ln)

	return &NetSocketValue{
		fd:       fd,
		listener: ln,
	}, nil
}

// builtinNetTcpConnect 创建 TCP 连接（客户端）
// net_tcp_connect(host, port) → socket
//
// 参数：
//   - args[0]: 目标主机（字符串），如 "example.com" 或 "192.168.1.1"
//   - args[1]: 目标端口（整数），如 80
//
// 返回值：
//   - NetSocketValue 对象，类型为已连接的客户端 socket
//
// 注意：
//   - 这是一个同步阻塞调用，连接建立后返回
//   - 超时由操作系统控制（通常约 75 秒）
//   - 如需非阻塞连接，先用 net_tcp_listen 创建非阻塞 socket 再连接
func builtinNetTcpConnect(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_tcp_connect() expects 2 arguments, got %d", len(args))
	}

	host := args[0].String()
	var port int
	if args[1].Type() == engine.TypeInt {
		port = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		port = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("net_tcp_connect() expects int port, got %s", args[1].Type())
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("net_tcp_connect() failed: %v", err)
	}

	fd := getFdFromConn(conn)

	return &NetSocketValue{
		fd:   fd,
		conn: conn,
	}, nil
}

// builtinNetTcpAccept 接受 TCP 连接
// net_tcp_accept(server_fd) → client_socket
//
// 参数：
//   - args[0]: TCP 服务器 socket（由 net_tcp_listen 创建）
//
// 返回值：
//   - 新的 NetSocketValue 对象，代表客户端连接
//
// 阻塞行为：
//   - 同步模式：如果没有待接受连接，阻塞等待
//   - 异步模式：与事件循环配合，在 on_accept 回调中调用（非阻塞）
//
// 典型用法（配合事件循环）：
//
//	$registry.on_accept($server, fn($client) {
//	    // 处理新连接
//	    $registry.on_read($client, fn($fd) {
//	        $data = net_recv($client, 1024)
//	        net_send($client, "Echo: " + $data)
//	    })
//	})
func builtinNetTcpAccept(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_tcp_accept() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_tcp_accept() expects socket, got %s", args[0].Type())
	}

	if sock.listener == nil {
		return nil, fmt.Errorf("net_tcp_accept() socket is not a listener")
	}

	conn, err := sock.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("net_tcp_accept() failed: %v", err)
	}

	fd := getFdFromConn(conn)

	return &NetSocketValue{
		fd:   fd,
		conn: conn,
	}, nil
}

// builtinNetUnixListen 创建 Unix Domain Socket 监听（服务器）
// net_unix_listen(path) → socket
//
// 参数：
//   - args[0]: Socket 文件路径（字符串），如 "/tmp/server.sock"
//
// 返回值：
//   - NetSocketValue 对象，类型为 Unix Domain 服务器
//
// 优势：
//   - 本地进程间通信，效率高于 TCP
//   - 基于文件权限控制访问
//   - 无网络协议开销
//
// 注意：
//   - Socket 文件会自动创建
//   - 如果文件已存在，会报错（需手动删除）
//   - 程序退出后建议删除 socket 文件
func builtinNetUnixListen(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_unix_listen() expects 1 argument, got %d", len(args))
	}

	path := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("net_unix_listen() expects string path, got %s", args[0].Type())
	}

	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("net_unix_listen() failed: %v", err)
	}

	fd := getFdFromListener(ln)

	return &NetSocketValue{
		fd:       fd,
		isUnix:   true,
		listener: ln,
	}, nil
}

// builtinNetUnixConnect 连接 Unix Domain Socket（客户端）
// net_unix_connect(path) → socket
//
// 参数：
//   - args[0]: Socket 文件路径（字符串）
//
// 返回值：
//   - NetSocketValue 对象，类型为已连接的客户端
//
// 典型场景：
//   - 与本地服务（如数据库、缓存）通信
//   - Docker 与宿主进程通信
//   - 需要高吞吐量的本地 IPC
func builtinNetUnixConnect(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_unix_connect() expects 1 argument, got %d", len(args))
	}

	path := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("net_unix_connect() expects string path, got %s", args[0].Type())
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("net_unix_connect() failed: %v", err)
	}

	fd := getFdFromConn(conn)

	return &NetSocketValue{
		fd:     fd,
		isUnix: true,
		conn:   conn,
	}, nil
}

// builtinNetUnixAccept 接受 Unix Domain Socket 连接
// net_unix_accept(server_fd) → client_socket
//
// 参数：
//   - args[0]: Unix Domain 服务器 socket（由 net_unix_listen 创建）
//
// 返回值：
//   - 新的 NetSocketValue 对象，代表客户端连接
//
// 注意：与 net_tcp_accept 行为相同，只是作用于 Unix Domain socket
func builtinNetUnixAccept(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_unix_accept() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_unix_accept() expects socket, got %s", args[0].Type())
	}

	if !sock.isUnix || sock.listener == nil {
		return nil, fmt.Errorf("net_unix_accept() socket is not a Unix listener")
	}

	conn, err := sock.listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("net_unix_accept() failed: %v", err)
	}

	fd := getFdFromConn(conn)

	return &NetSocketValue{
		fd:     fd,
		isUnix: true,
		conn:   conn,
	}, nil
}

// builtinNetUdpBind 创建 UDP socket 并绑定地址
// net_udp_bind(host, port) → socket
//
// 参数：
//   - args[0]: 绑定地址（字符串），如 "0.0.0.0" 或特定 IP
//   - args[1]: 端口号（整数）
//
// 返回值：
//   - NetSocketValue 对象，类型为 UDP socket
//
// UDP 特性：
//   - 无连接：不需要先建立连接就可以收发数据
//   - 数据报：每条消息独立，可能丢失、重复、乱序
//   - 效率高：无连接建立开销，适合实时应用
//
// 使用场景：
//   - DNS 查询、视频流、游戏同步
//   - 广播/组播
//   - 简单请求-响应协议
func builtinNetUdpBind(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_udp_bind() expects 2 arguments, got %d", len(args))
	}

	host := args[0].String()
	var port int
	if args[1].Type() == engine.TypeInt {
		port = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		port = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("net_udp_bind() expects int port, got %s", args[1].Type())
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("net_udp_bind() resolve failed: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("net_udp_bind() failed: %v", err)
	}

	file, err := conn.File()
	if err != nil {
		return nil, fmt.Errorf("net_udp_bind() get fd failed: %v", err)
	}
	defer file.Close()
	fd := int(file.Fd())

	return &NetSocketValue{
		fd:      fd,
		isUDP:   true,
		udpConn: conn,
		udpAddr: addr,
	}, nil
}

// builtinNetUdpSendto 发送 UDP 数据报
// net_udp_sendto(fd, data, host, port) → bytes_sent
//
// 参数：
//   - args[0]: UDP socket（由 net_udp_bind 创建）
//   - args[1]: 要发送的数据（字符串）
//   - args[2]: 目标主机地址
//   - args[3]: 目标端口
//
// 返回值：
//   - 实际发送的字节数
//
// 注意：
//   - 数据报大小有限制（通常 512-65,507 字节）
//   - 超过 MTU 的数据报可能被分片或丢弃
//   - 发送失败不保证通知（UDP 无确认机制）
func builtinNetUdpSendto(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("net_udp_sendto() expects 4 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_udp_sendto() expects socket, got %s", args[0].Type())
	}

	if !sock.isUDP {
		return nil, fmt.Errorf("net_udp_sendto() socket is not UDP")
	}

	data := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("net_udp_sendto() expects string data, got %s", args[1].Type())
	}

	host := args[2].String()
	var port int
	if args[3].Type() == engine.TypeInt {
		port = int(args[3].Int())
	} else if args[3].Type() == engine.TypeFloat {
		port = int(args[3].Float())
	} else {
		return nil, fmt.Errorf("net_udp_sendto() expects int port, got %s", args[3].Type())
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, fmt.Errorf("net_udp_sendto() resolve failed: %v", err)
	}

	n, err := sock.udpConn.WriteToUDP([]byte(data), addr)
	if err != nil {
		return nil, fmt.Errorf("net_udp_sendto() failed: %v", err)
	}

	return engine.NewInt(int64(n)), nil
}

// builtinNetUdpRecvfrom 接收 UDP 数据报
// net_udp_recvfrom(fd, len) → [data, from_ip, from_port]
//
// 参数：
//   - args[0]: UDP socket
//   - args[1]: 缓冲区大小（整数）
//
// 返回值：
//   - 数组：[接收的数据字符串, 来源IP, 来源端口]
//
// 阻塞行为：
//   - 同步模式：如果没有数据，阻塞等待
//   - 异步模式：在 on_read 回调中使用（推荐）
//
// 注意：
//   - 如果缓冲区太小，多余数据会被丢弃
//   - 返回的数据可能比请求的 len 短
//   - UDP 是无连接的，每次接收的数据可能来自不同来源
func builtinNetUdpRecvfrom(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_udp_recvfrom() expects 2 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_udp_recvfrom() expects socket, got %s", args[0].Type())
	}

	if !sock.isUDP {
		return nil, fmt.Errorf("net_udp_recvfrom() socket is not UDP")
	}

	var length int
	if args[1].Type() == engine.TypeInt {
		length = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		length = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("net_udp_recvfrom() expects int length, got %s", args[1].Type())
	}

	buffer := make([]byte, length)
	n, addr, err := sock.udpConn.ReadFromUDP(buffer)
	if err != nil {
		return nil, fmt.Errorf("net_udp_recvfrom() failed: %v", err)
	}

	// 返回 [data, from_ip, from_port]
	result := make([]engine.Value, 3)
	result[0] = engine.NewString(string(buffer[:n]))
	if addr != nil {
		result[1] = engine.NewString(addr.IP.String())
		result[2] = engine.NewInt(int64(addr.Port))
	} else {
		result[1] = engine.NewString("")
		result[2] = engine.NewInt(0)
	}

	return engine.NewArray(result), nil
}

// builtinNetSend 发送数据（TCP 或 Unix Domain Socket）
// net_send(fd, data) → bytes_sent
//
// 参数：
//   - args[0]: 已连接的 socket（TCP 或 Unix）
//   - args[1]: 要发送的数据（字符串）
//
// 返回值：
//   - 实际发送的字节数
//
// 阻塞行为：
//   - 同步模式：如果发送缓冲区满，阻塞等待
//   - 非阻塞模式：可能返回少于请求的字节数（配合事件循环使用）
//
// 错误：连接断开、权限不足等
//
// 注意：不适用于 UDP，UDP 请使用 net_udp_sendto
func builtinNetSend(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_send() expects 2 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_send() expects socket, got %s", args[0].Type())
	}

	data := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("net_send() expects string data, got %s", args[1].Type())
	}

	if sock.isUDP {
		return nil, fmt.Errorf("net_send() cannot use with UDP, use net_udp_sendto()")
	}

	if sock.conn == nil {
		return nil, fmt.Errorf("net_send() socket is not connected")
	}

	n, err := sock.conn.Write([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("net_send() failed: %v", err)
	}

	return engine.NewInt(int64(n)), nil
}

// builtinNetRecv 接收数据（TCP 或 Unix Domain Socket）
// net_recv(fd, len) → data
//
// 参数：
//   - args[0]: 已连接的 socket（TCP 或 Unix）
//   - args[1]: 最大接收字节数（整数）
//
// 返回值：
//   - 接收的数据（字符串）
//   - 如果连接关闭，可能返回空字符串
//
// 阻塞行为：
//   - 同步模式：如果没有数据，阻塞等待
//   - 异步模式：在 on_read 回调中使用（推荐）
//
// 注意：
//   - 返回的数据可能比请求的 len 短（流式协议特性）
//   - 需要应用层协议（如 HTTP）来确定消息边界
//   - 不适用于 UDP，UDP 请使用 net_udp_recvfrom
func builtinNetRecv(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_recv() expects 2 arguments, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_recv() expects socket, got %s", args[0].Type())
	}

	var length int
	if args[1].Type() == engine.TypeInt {
		length = int(args[1].Int())
	} else if args[1].Type() == engine.TypeFloat {
		length = int(args[1].Float())
	} else {
		return nil, fmt.Errorf("net_recv() expects int length, got %s", args[1].Type())
	}

	if sock.isUDP {
		return nil, fmt.Errorf("net_recv() cannot use with UDP, use net_udp_recvfrom()")
	}

	if sock.conn == nil {
		return nil, fmt.Errorf("net_recv() socket is not connected")
	}

	buffer := make([]byte, length)
	n, err := sock.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("net_recv() failed: %v", err)
	}

	return engine.NewString(string(buffer[:n])), nil
}

// builtinNetClose 关闭 socket
// net_close(fd) → bool
//
// 参数：
//   - args[0]: socket 对象
//
// 返回值：
//   - true: 关闭成功
//   - false: 关闭失败
//
// 关闭操作：
//   - 服务器 socket：停止监听，关闭所有资源
//   - 客户端 socket：关闭连接，发送 FIN 包
//   - UDP socket：关闭 socket
//
// 注意：
//   - 关闭后 socket 对象不可再用
//   - 建议配合事件循环使用，先注销事件处理器再关闭
//   - Unix Domain socket 不会自动删除文件，需手动删除
func builtinNetClose(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_close() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_close() expects socket, got %s", args[0].Type())
	}

	var err error
	if sock.conn != nil {
		err = sock.conn.Close()
	} else if sock.listener != nil {
		err = sock.listener.Close()
	} else if sock.udpConn != nil {
		err = sock.udpConn.Close()
	}

	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(true), nil
}

// builtinNetGetsockname 获取本地地址信息
// net_getsockname(fd) → {ip, port} 或 {path}
//
// 参数：
//   - args[0]: socket 对象
//
// 返回值（对象）：
//   - TCP/UDP: {ip: "本地IP", port: 本地端口}
//   - Unix Domain: {path: "socket文件路径"}
//
// 用途：
//   - 确认绑定的端口（如绑定 0 后查看分配的端口）
//   - 日志记录
//   - 网络诊断
func builtinNetGetsockname(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_getsockname() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_getsockname() expects socket, got %s", args[0].Type())
	}

	result := make(map[string]engine.Value)

	if sock.isUnix && sock.listener != nil {
		// Unix listener
		if unixLn, ok := sock.listener.(*net.UnixListener); ok {
			addr := unixLn.Addr()
			if unixAddr, ok := addr.(*net.UnixAddr); ok {
				result["path"] = engine.NewString(unixAddr.String())
			}
		}
	} else if sock.listener != nil {
		// TCP listener
		if tcpLn, ok := sock.listener.(*net.TCPListener); ok {
			addr := tcpLn.Addr()
			if tcpAddr, ok := addr.(*net.TCPAddr); ok {
				result["ip"] = engine.NewString(tcpAddr.IP.String())
				result["port"] = engine.NewInt(int64(tcpAddr.Port))
			}
		}
	} else if sock.conn != nil {
		// TCP/Unix client
		addr := sock.conn.LocalAddr()
		if tcpAddr, ok := addr.(*net.TCPAddr); ok {
			result["ip"] = engine.NewString(tcpAddr.IP.String())
			result["port"] = engine.NewInt(int64(tcpAddr.Port))
		} else if unixAddr, ok := addr.(*net.UnixAddr); ok {
			result["path"] = engine.NewString(unixAddr.String())
		}
	} else if sock.udpConn != nil {
		// UDP
		addr := sock.udpConn.LocalAddr()
		if udpAddr, ok := addr.(*net.UDPAddr); ok {
			result["ip"] = engine.NewString(udpAddr.IP.String())
			result["port"] = engine.NewInt(int64(udpAddr.Port))
		}
	}

	obj := engine.NewObject(result)
	return obj, nil
}

// builtinNetGetpeername 获取对端地址信息
// net_getpeername(fd) → {ip, port} 或 {path}
//
// 参数：
//   - args[0]: 已连接的 socket 对象
//
// 返回值（对象）：
//   - TCP: {ip: "对端IP", port: 对端端口}
//   - Unix Domain: {path: "对端socket路径"}
//
// 错误：
//   - 未连接的 socket 返回错误
//
// 用途：
//   - 记录客户端信息
//   - 访问控制（基于 IP 的黑白名单）
//   - 连接追踪
func builtinNetGetpeername(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_getpeername() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_getpeername() expects socket, got %s", args[0].Type())
	}

	if sock.conn == nil {
		return nil, fmt.Errorf("net_getpeername() socket is not connected")
	}

	result := make(map[string]engine.Value)

	addr := sock.conn.RemoteAddr()
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		result["ip"] = engine.NewString(tcpAddr.IP.String())
		result["port"] = engine.NewInt(int64(tcpAddr.Port))
	} else if unixAddr, ok := addr.(*net.UnixAddr); ok {
		result["path"] = engine.NewString(unixAddr.String())
	}

	obj := engine.NewObject(result)
	return obj, nil
}

// builtinNetSetNonblock 设置 socket 为非阻塞模式
// net_set_nonblock(fd) → bool
//
// 参数：
//   - args[0]: socket 对象
//
// 返回值：
//   - true: 设置成功
//   - false: 设置失败
//
// 阻塞 vs 非阻塞：
//   - 阻塞模式：IO 操作未完成时会阻塞等待
//   - 非阻塞模式：IO 操作会立即返回，可能返回 EAGAIN 错误
//
// 推荐用法：
//
//	与事件循环配合，设置非阻塞后使用 ev_on_read/ev_on_write：
//	net_set_nonblock($socket)
//	$registry.on_read($socket, fn($fd) {
//	    $data = net_recv($socket, 1024)
//	    // 处理数据
//	})
//
// 注意：设置非阻塞后，直接调用 net_recv/net_send 可能返回错误，建议配合事件循环使用
func builtinNetSetNonblock(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_set_nonblock() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_set_nonblock() expects socket, got %s", args[0].Type())
	}

	// 使用 syscall 设置非阻塞
	err := syscall.SetNonblock(sock.fd, true)
	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(true), nil
}

// builtinNetIsUnix 检查 socket 是否为 Unix Domain Socket
// net_is_unix(fd) → bool
//
// 参数：
//   - args[0]: socket 对象
//
// 返回值：
//   - true: 是 Unix Domain Socket
//   - false: 是 TCP 或 UDP socket
//
// 用途：
//   - 编写跨协议通用代码时判断 socket 类型
//   - 日志记录时区分协议类型
//   - 某些操作仅适用于 Unix Domain（如获取 socket 文件权限）
func builtinNetIsUnix(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("net_is_unix() expects 1 argument, got %d", len(args))
	}

	sock, ok := args[0].(*NetSocketValue)
	if !ok {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(sock.isUnix), nil
}

// ==============================================================================
// 网络事件注册方法（使用通用事件接口）
// 提供 net_on_accept, net_on_read, net_on_write 等函数
// 内部使用 registry.RegisterHandler() 注册事件
// ==============================================================================

// RegisterNetEvents 注册网络事件函数到引擎
func RegisterNetEvents(e *engine.Engine) {
	e.RegisterFunc("net_on_accept", builtinNetOnAccept)
	e.RegisterFunc("net_on_read", builtinNetOnRead)
	e.RegisterFunc("net_on_write", builtinNetOnWrite)
	e.RegisterFunc("net_off", builtinNetOff)
	e.RegisterFunc("net_off_read", builtinNetOffRead)
	e.RegisterFunc("net_off_write", builtinNetOffWrite)
}

// NetEventNames 返回网络事件函数名称列表
func NetEventNames() []string {
	return []string{
		"net_on_accept", "net_on_read", "net_on_write",
		"net_off", "net_off_read", "net_off_write",
	}
}

// builtinNetOnAccept 注册 accept 事件处理器
// net_on_accept($registry, $server, fn($client) { ... }) → int (handler_id)
func builtinNetOnAccept(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("net_on_accept() expects 3 arguments: registry, server, callback")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_on_accept() expects EvRegistry, got %s", args[0].Type())
	}

	serverSocket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_on_accept() expects NetSocketValue, got %s", args[1].Type())
	}

	fn, ok := args[2].(engine.Value)
	if !ok {
		return nil, fmt.Errorf("net_on_accept() expects function callback, got %s", args[2].Type())
	}

	if serverSocket.listener == nil {
		return nil, fmt.Errorf("net_on_accept() socket has no listener")
	}

	// 获取 registry 的 context
	loopCtx := registry.Context()
	if loopCtx == nil {
		loopCtx = context.Background()
	}

	// 创建独立的 context 用于控制 goroutine
	gCtx, gCancel := context.WithCancel(loopCtx)

	// 注册事件处理器
	handlerID := registry.RegisterHandler("accept", serverSocket, fn, ctx, gCancel)

	// 启动 goroutine 监听 accept 事件
	go acceptLoop(gCtx, serverSocket.listener, ctx, fn)

	return engine.NewInt(int64(handlerID)), nil
}

// acceptLoop 监听服务器 socket 的新连接
func acceptLoop(ctx context.Context, listener net.Listener, callbackCtx *engine.Context, callback engine.Value) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置 Accept 超时，以便定期检查 context
			if tcpListener, ok := listener.(*net.TCPListener); ok {
				tcpListener.SetDeadline(time.Now().Add(100 * time.Millisecond))
			} else if unixListener, ok := listener.(*net.UnixListener); ok {
				unixListener.SetDeadline(time.Now().Add(100 * time.Millisecond))
			}

			conn, err := listener.Accept()
			if err != nil {
				// 检查是否是超时错误
				if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
					continue
				}
				// 其他错误，可能是 listener 被关闭
				return
			}

			// 创建客户端 socket 对象
			clientFd := getFdFromConn(conn)
			clientSocket := &NetSocketValue{
				fd:   clientFd,
				conn: conn,
			}

			// 调用 JPL 回调
			if callbackCtx != nil && callback != nil {
				_, _ = callbackCtx.VM().CallValue(callback, clientSocket)
			}
		}
	}
}

// builtinNetOnRead 注册读事件处理器
// net_on_read($registry, $socket, fn($socket, $data) { ... }) → int (handler_id)
func builtinNetOnRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("net_on_read() expects 3 arguments: registry, socket, callback")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_on_read() expects EvRegistry, got %s", args[0].Type())
	}

	socket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_on_read() expects NetSocketValue, got %s", args[1].Type())
	}

	fn, ok := args[2].(engine.Value)
	if !ok {
		return nil, fmt.Errorf("net_on_read() expects function callback, got %s", args[2].Type())
	}

	if socket.conn == nil {
		return nil, fmt.Errorf("net_on_read() socket has no connection")
	}

	// 获取 registry 的 context
	loopCtx := registry.Context()
	if loopCtx == nil {
		loopCtx = context.Background()
	}

	// 创建独立的 context
	gCtx, gCancel := context.WithCancel(loopCtx)

	// 注册事件处理器
	handlerID := registry.RegisterHandler("read", socket, fn, ctx, gCancel)

	// 启动 goroutine 监听读事件
	go readLoop(gCtx, socket, ctx, fn)

	return engine.NewInt(int64(handlerID)), nil
}

// readLoop 监听 socket 的可读事件
func readLoop(ctx context.Context, socket *NetSocketValue, callbackCtx *engine.Context, callback engine.Value) {
	conn := socket.conn

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置读超时
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if err != nil {
				if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
					continue
				}
				// 连接关闭或错误
				return
			}

			if n > 0 {
				// 调用 JPL 回调，传入 socket 和读取的数据
				if callbackCtx != nil && callback != nil {
					data := engine.NewString(string(buf[:n]))
					_, _ = callbackCtx.VM().CallValue(callback, socket, data)
				}
			}
		}
	}
}

// builtinNetOnWrite 注册写事件处理器
// net_on_write($registry, $socket, fn($socket) { ... }) → int (handler_id)
func builtinNetOnWrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("net_on_write() expects 3 arguments: registry, socket, callback")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_on_write() expects EvRegistry, got %s", args[0].Type())
	}

	socket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_on_write() expects NetSocketValue, got %s", args[1].Type())
	}

	fn, ok := args[2].(engine.Value)
	if !ok {
		return nil, fmt.Errorf("net_on_write() expects function callback, got %s", args[2].Type())
	}

	// 获取 registry 的 context
	loopCtx := registry.Context()
	if loopCtx == nil {
		loopCtx = context.Background()
	}

	// 创建独立的 context
	gCtx, gCancel := context.WithCancel(loopCtx)

	// 注册事件处理器
	handlerID := registry.RegisterHandler("write", socket, fn, ctx, gCancel)

	// 启动 goroutine（写事件通常不需要持续监听）
	go writeLoop(gCtx, socket, ctx, fn)

	return engine.NewInt(int64(handlerID)), nil
}

// writeLoop 监听 socket 的可写事件
func writeLoop(ctx context.Context, socket *NetSocketValue, callbackCtx *engine.Context, callback engine.Value) {
	// 写事件通常不需要持续监听，只在需要时触发
	// 这里简单实现：等待 context 取消后退出
	<-ctx.Done()
}

// builtinNetOff 注销 socket 的所有事件
// net_off($registry, $socket) → bool
func builtinNetOff(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_off() expects 2 arguments: registry, socket")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_off() expects EvRegistry, got %s", args[0].Type())
	}

	socket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_off() expects NetSocketValue, got %s", args[1].Type())
	}

	registry.UnregisterBySource(socket)

	return engine.NewBool(true), nil
}

// builtinNetOffRead 只注销读事件
// net_off_read($registry, $socket) → bool
func builtinNetOffRead(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_off_read() expects 2 arguments: registry, socket")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_off_read() expects EvRegistry, got %s", args[0].Type())
	}

	socket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_off_read() expects NetSocketValue, got %s", args[1].Type())
	}

	registry.UnregisterBySourceAndType(socket, "read")

	return engine.NewBool(true), nil
}

// builtinNetOffWrite 只注销写事件
// net_off_write($registry, $socket) → bool
func builtinNetOffWrite(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("net_off_write() expects 2 arguments: registry, socket")
	}

	registry, ok := args[0].(*EvRegistryValue)
	if !ok {
		return nil, fmt.Errorf("net_off_write() expects EvRegistry, got %s", args[0].Type())
	}

	socket, ok := args[1].(*NetSocketValue)
	if !ok {
		return nil, fmt.Errorf("net_off_write() expects NetSocketValue, got %s", args[1].Type())
	}

	registry.UnregisterBySourceAndType(socket, "write")

	return engine.NewBool(true), nil
}
