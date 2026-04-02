package stdlib

import (
	"os"
	"testing"
	"time"

	"github.com/gnuos/jpl/engine"
)

// TestNetUnixSocket 测试 Unix Domain Socket
func TestNetUnixSocket(t *testing.T) {
	socketPath := "/tmp/test_jpl_unix_" + time.Now().Format("20060102150405")
	defer os.Remove(socketPath)

	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	// 创建监听
	listenArgs := []engine.Value{
		engine.NewString(socketPath),
	}
	server, err := builtinNetUnixListen(nil, listenArgs)
	if err != nil {
		t.Fatalf("net_unix_listen() error = %v", err)
	}

	if server.Type() != engine.TypeStream {
		t.Errorf("server type = %v, expected TypeStream", server.Type())
	}

	// 检查是否为 Unix socket
	isUnixArgs := []engine.Value{server}
	isUnix, _ := builtinNetIsUnix(nil, isUnixArgs)
	if !isUnix.Bool() {
		t.Error("net_is_unix() should return true for Unix socket")
	}

	// 获取地址
	getsockArgs := []engine.Value{server}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)
	if sockname.Type() != engine.TypeObject {
		t.Errorf("getsockname type = %v, expected TypeObject", sockname.Type())
	}
}

// TestNetIsUnix 测试 is_unix 检查
func TestNetIsUnix(t *testing.T) {
	socketPath := "/tmp/test_jpl_isunix_" + time.Now().Format("20060102150405")
	defer os.Remove(socketPath)

	// 创建 Unix socket
	listenArgs := []engine.Value{engine.NewString(socketPath)}
	unixSock, _ := builtinNetUnixListen(nil, listenArgs)

	// 检查 Unix socket
	args := []engine.Value{unixSock}
	result, _ := builtinNetIsUnix(nil, args)
	if !result.Bool() {
		t.Error("net_is_unix() should return true for Unix socket")
	}

	// 非 socket 应该返回 false
	result2, _ := builtinNetIsUnix(nil, []engine.Value{engine.NewInt(1)})
	if result2.Bool() {
		t.Error("net_is_unix() should return false for non-socket")
	}
}

// TestNetIPv6 测试 IPv6 地址处理
func TestNetIPv6(t *testing.T) {
	// 测试 net.JoinHostPort 修复 IPv6 地址格式
	// 在修复前，IPv6 地址如 "::1" 直接拼接会导致错误格式 "::1:PORT"
	// 修复后应正确格式化为 "[::1]:PORT"

	// 测试 IPv6 回环地址监听
	listenArgs := []engine.Value{
		engine.NewString("::1"),
		engine.NewInt(0), // 使用端口 0 让系统分配可用端口
	}
	server, err := builtinNetTcpListen(nil, listenArgs)
	if err != nil {
		t.Fatalf("net_tcp_listen() with IPv6 loopback error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	// 获取实际监听的端口
	getsockArgs := []engine.Value{server}
	sockname, err := builtinNetGetsockname(nil, getsockArgs)
	if err != nil {
		t.Fatalf("net_getsockname() error = %v", err)
	}

	sockObj := sockname.Object()
	portVal := sockObj["port"]
	if portVal == nil {
		t.Fatal("getsockname did not return port")
	}
	port := int(portVal.Int())

	// 测试 IPv6 连接
	connectArgs := []engine.Value{
		engine.NewString("::1"),
		engine.NewInt(int64(port)),
	}

	// 在 goroutine 中先接受连接
	acceptCh := make(chan engine.Value, 1)
	acceptErrCh := make(chan error, 1)
	go func() {
		client, err := builtinNetTcpAccept(nil, []engine.Value{server})
		if err != nil {
			acceptErrCh <- err
			return
		}
		acceptCh <- client
	}()

	// 连接服务器
	conn, err := builtinNetTcpConnect(nil, connectArgs)
	if err != nil {
		t.Fatalf("net_tcp_connect() with IPv6 loopback error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn})

	// 等待接受连接
	select {
	case client := <-acceptCh:
		defer builtinNetClose(nil, []engine.Value{client})
		// 连接成功建立，测试通过
	case err := <-acceptErrCh:
		t.Fatalf("net_tcp_accept() error = %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection accept")
	}
}

// TestNetIPv6AllInterfaces 测试 IPv6 全接口绑定 (::)
func TestNetIPv6AllInterfaces(t *testing.T) {
	// 测试绑定所有 IPv6 接口
	listenArgs := []engine.Value{
		engine.NewString("::"),
		engine.NewInt(0),
	}
	server, err := builtinNetTcpListen(nil, listenArgs)
	if err != nil {
		t.Fatalf("net_tcp_listen() with IPv6 all interfaces error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	// 获取实际端口
	getsockArgs := []engine.Value{server}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)
	sockObj := sockname.Object()
	port := int(sockObj["port"].Int())

	// 连接测试 - 使用 IPv6 回环地址
	acceptCh := make(chan engine.Value, 1)
	go func() {
		client, _ := builtinNetTcpAccept(nil, []engine.Value{server})
		acceptCh <- client
	}()

	conn, err := builtinNetTcpConnect(nil, []engine.Value{
		engine.NewString("::1"),
		engine.NewInt(int64(port)),
	})
	if err != nil {
		t.Fatalf("net_tcp_connect() with IPv6 all interfaces error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn})

	select {
	case <-acceptCh:
		// 成功
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for connection")
	}
}

// TestNetIPv6DualStack 测试 IPv4/IPv6 双栈（如果系统支持）
func TestNetIPv6DualStack(t *testing.T) {
	// 先尝试 IPv4
	listenArgs4 := []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	}
	server4, err := builtinNetTcpListen(nil, listenArgs4)
	if err != nil {
		t.Fatalf("net_tcp_listen() with IPv4 error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server4})

	// 获取 IPv4 端口
	sockname4, _ := builtinNetGetsockname(nil, []engine.Value{server4})
	port4 := int(sockname4.Object()["port"].Int())

	// 测试 IPv4 连接
	go func() {
		builtinNetTcpAccept(nil, []engine.Value{server4})
	}()

	conn4, err := builtinNetTcpConnect(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(int64(port4)),
	})
	if err != nil {
		t.Fatalf("net_tcp_connect() with IPv4 error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn4})

	// 然后测试 IPv6（确保不会相互干扰）
	listenArgs6 := []engine.Value{
		engine.NewString("::1"),
		engine.NewInt(0),
	}
	server6, err := builtinNetTcpListen(nil, listenArgs6)
	if err != nil {
		t.Fatalf("net_tcp_listen() with IPv6 error = %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server6})

	t.Log("IPv4 and IPv6 can work simultaneously")
}

// TestNetInvalidAddress 测试无效地址的错误处理
func TestNetInvalidAddress(t *testing.T) {
	// 测试无效的主机名
	listenArgs := []engine.Value{
		engine.NewString("not-a-valid-hostname-!!!"),
		engine.NewInt(8080),
	}
	_, err := builtinNetTcpListen(nil, listenArgs)
	if err == nil {
		t.Error("net_tcp_listen() with invalid host should return error")
	}

	// 测试连接到无效地址
	connectArgs := []engine.Value{
		engine.NewString("invalid-host-that-does-not-exist.example"),
		engine.NewInt(8080),
	}
	_, err = builtinNetTcpConnect(nil, connectArgs)
	if err == nil {
		t.Error("net_tcp_connect() with invalid host should return error")
	}
}

// TestNetPortBoundaries 测试端口边界值
func TestNetPortBoundaries(t *testing.T) {
	// 测试端口 0（自动分配）
	listenArgs := []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	}
	server, err := builtinNetTcpListen(nil, listenArgs)
	if err != nil {
		t.Fatalf("net_tcp_listen() with port 0 error = %v", err)
	}

	// 获取分配的端口
	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := sockname.Object()["port"].Int()
	if port <= 0 || port > 65535 {
		t.Errorf("port 0 should be assigned a valid ephemeral port, got %d", port)
	}
	builtinNetClose(nil, []engine.Value{server})

	// 测试有效端口 1
	listenArgs2 := []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(1),
	}
	_, err = builtinNetTcpListen(nil, listenArgs2)
	// 端口 1 需要 root 权限，可能会失败
	if err != nil {
		t.Logf("Port 1 requires root privileges: %v", err)
	} else {
		t.Log("Port 1 binding succeeded (running as root?)")
	}

	// 测试端口 65535
	listenArgs3 := []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(65535),
	}
	server3, err := builtinNetTcpListen(nil, listenArgs3)
	if err != nil {
		t.Logf("Port 65535 may be in use: %v", err)
	} else {
		builtinNetClose(nil, []engine.Value{server3})
	}
}
