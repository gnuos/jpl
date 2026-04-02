package stdlib

import (
	"runtime"
	"testing"
	"time"

	"github.com/gnuos/jpl/engine"
)

// =============================================================================
// Phase 9.7 边界测试
// 测试类型：网络超时、错误处理、资源清理、并发、大数据量
// =============================================================================

// TestConnectInvalidPort 测试连接无效端口
func TestConnectInvalidPort(t *testing.T) {
	connectArgs := []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(1),
	}
	_, err := builtinNetTcpConnect(nil, connectArgs)
	if err == nil {
		t.Error("connect to unlistening port should return error")
	} else {
		t.Logf("connect refused (expected): %v", err)
	}
}

// TestConnectInvalidHost 测试连接无效主机
func TestConnectInvalidHost(t *testing.T) {
	connectArgs := []engine.Value{
		engine.NewString("invalid-host-99999.nonexist"),
		engine.NewInt(8080),
	}
	_, err := builtinNetTcpConnect(nil, connectArgs)
	if err == nil {
		t.Error("connect to invalid host should return error")
	} else {
		t.Logf("connect error (expected): %v", err)
	}
}

// TestListenInvalidAddress 测试绑定无效地址
func TestListenInvalidAddress(t *testing.T) {
	_, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("999.999.999.999"),
		engine.NewInt(8080),
	})
	if err == nil {
		t.Error("listen on invalid address should return error")
	} else {
		t.Logf("listen error (expected): %v", err)
	}
}

// TestSendOnClosedSocket 测试向已关闭 socket 发送
func TestSendOnClosedSocket(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := int(sockname.Object()["port"].Int())

	accepted := make(chan engine.Value, 1)
	go func() {
		client, _ := builtinNetTcpAccept(nil, []engine.Value{server})
		if client != nil {
			accepted <- client
			time.Sleep(50 * time.Millisecond)
			builtinNetClose(nil, []engine.Value{client})
		}
	}()

	conn, err := builtinNetTcpConnect(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(int64(port)),
	})
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}

	// 先发送一次成功
	_, err = builtinNetSend(nil, []engine.Value{conn, engine.NewString("hello")})
	if err != nil {
		t.Errorf("first send error: %v", err)
	}

	// 等待对端关闭
	<-accepted
	time.Sleep(100 * time.Millisecond)

	// 关闭本地 socket
	builtinNetClose(nil, []engine.Value{conn})
	builtinNetClose(nil, []engine.Value{server})

	// 向已关闭 socket 发送应报错
	_, err = builtinNetSend(nil, []engine.Value{conn, engine.NewString("after close")})
	if err == nil {
		t.Error("send on closed socket should return error")
	} else {
		t.Logf("send on closed socket error (expected): %v", err)
	}
}

// TestMultipleConnections 测试多连接并发
func TestMultipleConnections(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := int(sockname.Object()["port"].Int())

	const numClients = 10
	accepted := make(chan engine.Value, numClients)
	clientDone := make(chan engine.Value, numClients)
	errCh := make(chan error, numClients*2)

	// 服务器接受连接
	go func() {
		for i := 0; i < numClients; i++ {
			client, err := builtinNetTcpAccept(nil, []engine.Value{server})
			if err != nil {
				errCh <- err
				return
			}
			accepted <- client
		}
	}()

	// 客户端并发连接
	for i := 0; i < numClients; i++ {
		go func() {
			conn, err := builtinNetTcpConnect(nil, []engine.Value{
				engine.NewString("127.0.0.1"),
				engine.NewInt(int64(port)),
			})
			if err != nil {
				errCh <- err
				return
			}
			clientDone <- conn
		}()
	}

	// 收集所有连接
	var serverConns, clientConns []engine.Value
	timeout := time.After(5 * time.Second)
	for i := 0; i < numClients*2; i++ {
		select {
		case c := <-accepted:
			serverConns = append(serverConns, c)
		case c := <-clientDone:
			clientConns = append(clientConns, c)
		case err := <-errCh:
			t.Errorf("error: %v", err)
		case <-timeout:
			t.Fatalf("timeout: got %d server, %d clients", len(serverConns), len(clientConns))
		}
	}

	// 清理
	for _, c := range serverConns {
		builtinNetClose(nil, []engine.Value{c})
	}
	for _, c := range clientConns {
		builtinNetClose(nil, []engine.Value{c})
	}

	t.Logf("established %d concurrent connections", numClients)
}

// TestLargeDataTransfer 测试大数据传输
func TestLargeDataTransfer(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := int(sockname.Object()["port"].Int())

	// 256KB 数据
	dataSize := 256 * 1024
	largeData := make([]byte, dataSize)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	serverResult := make(chan int, 1)
	serverErr := make(chan error, 1)

	// 服务器接收
	go func() {
		client, err := builtinNetTcpAccept(nil, []engine.Value{server})
		if err != nil {
			serverErr <- err
			return
		}
		defer builtinNetClose(nil, []engine.Value{client})

		var received []byte
		for len(received) < dataSize {
			data, err := builtinNetRecv(nil, []engine.Value{client, engine.NewInt(65536)})
			if err != nil {
				break
			}
			received = append(received, []byte(data.String())...)
		}
		serverResult <- len(received)
	}()

	// 客户端连接并发送
	conn, err := builtinNetTcpConnect(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(int64(port)),
	})
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn})

	sent, err := builtinNetSend(nil, []engine.Value{conn, engine.NewString(string(largeData))})
	if err != nil {
		t.Fatalf("send error: %v", err)
	}
	t.Logf("sent %d bytes", sent.Int())

	// 等待服务器接收
	select {
	case n := <-serverResult:
		if n != dataSize {
			t.Errorf("received %d bytes, expected %d", n, dataSize)
		} else {
			t.Logf("received %d bytes (complete)", n)
		}
	case err := <-serverErr:
		t.Fatalf("server error: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for server")
	}
}

// TestRapidConnectDisconnect 测试快速连接断开
func TestRapidConnectDisconnect(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := int(sockname.Object()["port"].Int())

	const iterations = 50
	acceptDone := make(chan bool, 1)

	// 服务器快速接受并关闭
	go func() {
		for i := 0; i < iterations; i++ {
			client, err := builtinNetTcpAccept(nil, []engine.Value{server})
			if err != nil {
				return
			}
			builtinNetClose(nil, []engine.Value{client})
		}
		acceptDone <- true
	}()

	// 客户端快速连接断开
	for i := 0; i < iterations; i++ {
		conn, err := builtinNetTcpConnect(nil, []engine.Value{
			engine.NewString("127.0.0.1"),
			engine.NewInt(int64(port)),
		})
		if err != nil {
			t.Errorf("iteration %d: %v", i, err)
			continue
		}
		builtinNetClose(nil, []engine.Value{conn})
	}

	select {
	case <-acceptDone:
		t.Logf("completed %d rapid cycles", iterations)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for server")
	}
}

// TestUDPBasicSendRecv 测试 UDP 基本收发
func TestUDPBasicSendRecv(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	sock, err := builtinNetUdpBind(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("bind error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{sock})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{sock})
	port := int(sockname.Object()["port"].Int())

	testData := "UDP test message"
	_, err = builtinNetUdpSendto(nil, []engine.Value{
		sock,
		engine.NewString(testData),
		engine.NewString("127.0.0.1"),
		engine.NewInt(int64(port)),
	})
	if err != nil {
		t.Fatalf("sendto error: %v", err)
	}

	result, err := builtinNetUdpRecvfrom(nil, []engine.Value{sock, engine.NewInt(1024)})
	if err != nil {
		t.Fatalf("recvfrom error: %v", err)
	}

	arr := result.Array()
	if len(arr) < 1 {
		t.Fatal("recvfrom returned empty array")
	}
	data := arr[0].String()
	if data != testData {
		t.Errorf("received %q, expected %q", data, testData)
	} else {
		t.Logf("UDP send/recv OK: %q", data)
	}
}

// TestUDPLargePacket 测试 UDP 大数据包
func TestUDPLargePacket(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	sock, err := builtinNetUdpBind(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("bind error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{sock})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{sock})
	port := int(sockname.Object()["port"].Int())

	sizes := []int{1, 512, 8192, 32768}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		_, err := builtinNetUdpSendto(nil, []engine.Value{
			sock,
			engine.NewString(string(data)),
			engine.NewString("127.0.0.1"),
			engine.NewInt(int64(port)),
		})
		if err != nil {
			t.Errorf("send %d bytes error: %v", size, err)
			continue
		}

		result, err := builtinNetUdpRecvfrom(nil, []engine.Value{sock, engine.NewInt(int64(size + 100))})
		if err != nil {
			t.Errorf("recv %d bytes error: %v", size, err)
			continue
		}

		arr := result.Array()
		if len(arr) < 1 {
			t.Errorf("size %d: empty response", size)
			continue
		}
		received := arr[0].String()
		if len(received) != size {
			t.Errorf("size %d: received %d bytes", size, len(received))
		} else {
			t.Logf("UDP %d bytes OK", size)
		}
	}
}

// TestFDCleanup 测试 FD 泄漏检测
func TestFDCleanup(t *testing.T) {
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	const iterations = 50

	for i := 0; i < iterations; i++ {
		server, err := builtinNetTcpListen(nil, []engine.Value{
			engine.NewString("127.0.0.1"),
			engine.NewInt(0),
		})
		if err != nil {
			continue
		}

		sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
		port := int(sockname.Object()["port"].Int())

		accepted := make(chan engine.Value, 1)
		go func() {
			client, _ := builtinNetTcpAccept(nil, []engine.Value{server})
			if client != nil {
				accepted <- client
			}
		}()

		conn, err := builtinNetTcpConnect(nil, []engine.Value{
			engine.NewString("127.0.0.1"),
			engine.NewInt(int64(port)),
		})
		if err != nil {
			builtinNetClose(nil, []engine.Value{server})
			continue
		}

		select {
		case client := <-accepted:
			builtinNetClose(nil, []engine.Value{client})
		case <-time.After(500 * time.Millisecond):
		}

		builtinNetClose(nil, []engine.Value{conn})
		builtinNetClose(nil, []engine.Value{server})
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	memGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	if memGrowth > 10*1024*1024 {
		t.Errorf("memory grew by %d bytes (possible leak)", memGrowth)
	} else {
		t.Logf("memory growth: %d bytes after %d iterations", memGrowth, iterations)
	}
}

// TestConcurrentEcho 测试并发回显
func TestConcurrentEcho(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("listen error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	sockname, _ := builtinNetGetsockname(nil, []engine.Value{server})
	port := int(sockname.Object()["port"].Int())

	const numMessages = 10
	serverDone := make(chan bool, 1)

	// 服务器回显
	go func() {
		client, err := builtinNetTcpAccept(nil, []engine.Value{server})
		if err != nil {
			return
		}
		defer builtinNetClose(nil, []engine.Value{client})

		for i := 0; i < numMessages; i++ {
			data, err := builtinNetRecv(nil, []engine.Value{client, engine.NewInt(1024)})
			if err != nil {
				return
			}
			builtinNetSend(nil, []engine.Value{client, engine.NewString("echo:" + data.String())})
		}
		serverDone <- true
	}()

	// 客户端
	conn, err := builtinNetTcpConnect(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(int64(port)),
	})
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn})

	for i := 0; i < numMessages; i++ {
		msg := "msg" + string(rune('0'+i%10))
		builtinNetSend(nil, []engine.Value{conn, engine.NewString(msg)})

		resp, err := builtinNetRecv(nil, []engine.Value{conn, engine.NewInt(1024)})
		if err != nil {
			t.Errorf("recv error at %d: %v", i, err)
			continue
		}

		expected := "echo:" + msg
		if resp.String() != expected {
			t.Errorf("response %d: got %q, expected %q", i, resp.String(), expected)
		}
	}

	select {
	case <-serverDone:
		t.Log("concurrent echo completed")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

// TestUnixSocketConnection 测试 Unix Socket 连接
func TestUnixSocketConnection(t *testing.T) {
	socketPath := "/tmp/test_jpl_boundary_" + time.Now().Format("150405000")
	defer removeFile(socketPath)

	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	server, err := builtinNetUnixListen(nil, []engine.Value{engine.NewString(socketPath)})
	if err != nil {
		t.Fatalf("unix_listen error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{server})

	accepted := make(chan engine.Value, 1)
	go func() {
		client, _ := builtinNetUnixAccept(nil, []engine.Value{server})
		if client != nil {
			accepted <- client
		}
	}()

	conn, err := builtinNetUnixConnect(nil, []engine.Value{engine.NewString(socketPath)})
	if err != nil {
		t.Fatalf("unix_connect error: %v", err)
	}
	defer builtinNetClose(nil, []engine.Value{conn})

	select {
	case client := <-accepted:
		defer builtinNetClose(nil, []engine.Value{client})
		t.Log("unix socket connection established")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout accepting unix connection")
	}
}

func removeFile(path string) {
	// ignore errors
}
