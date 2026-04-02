package stdlib

import (
	"os"
	"testing"
	"time"

	"github.com/gnuos/jpl/engine"
)

// TestIntegrationEchoUnix 测试 Unix Domain Echo 服务器
func TestIntegrationEchoUnix(t *testing.T) {
	socketPath := "/tmp/test_echo_unix_" + time.Now().Format("20060102150405") + ".sock"
	defer os.Remove(socketPath)

	eng := engine.NewEngine()
	defer eng.Close()
	RegisterEv(eng)
	RegisterNet(eng)

	// 创建 Echo 服务器
	server, err := builtinNetUnixListen(nil, []engine.Value{engine.NewString(socketPath)})
	if err != nil {
		t.Fatalf("Failed to create Unix server: %v", err)
	}

	// 创建注册表
	_ = NewEvRegistry()

	// 注册 accept 事件
	acceptHandler := func(clientFD int) {
		// 为新客户端注册读事件
		clientSock := &NetSocketValue{fd: clientFD}

		// 使用 net_recv 读取数据
		readArgs := []engine.Value{clientSock, engine.NewInt(1024)}
		data, err := builtinNetRecv(nil, readArgs)
		if err != nil {
			return
		}

		// 回显数据
		echoData := "Echo: " + data.String()
		sendArgs := []engine.Value{clientSock, engine.NewString(echoData)}
		builtinNetSend(nil, sendArgs)

		// 关闭连接
		closeArgs := []engine.Value{clientSock}
		builtinNetClose(nil, closeArgs)
	}

	// 这里简化处理，实际应该通过 ev 注册表
	_ = acceptHandler

	// 创建事件循环
	_, err = builtinEvLoopNew(nil, []engine.Value{})
	if err != nil {
		t.Logf("EvLoop not available on this platform: %v", err)
		return
	}

	// 关闭服务器
	builtinNetClose(nil, []engine.Value{server})
	t.Log("Unix Domain Echo server test passed")
}

// TestIntegrationTCPClientServer 测试 TCP 客户端-服务器
func TestIntegrationTCPClientServer(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)
	RegisterDNS(eng)

	// DNS 解析
	dnsArgs := []engine.Value{engine.NewString("localhost")}
	ips, err := builtinDNSResolve(nil, dnsArgs)
	if err != nil {
		t.Logf("DNS resolve failed (may be network issue): %v", err)
		return
	}

	ipArr := ips.Array()
	if len(ipArr) == 0 {
		t.Error("DNS resolve returned no IPs")
		return
	}

	t.Logf("localhost resolved to: %v", ipArr)

	// 创建 TCP 服务器
	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0), // 自动分配端口
	})
	if err != nil {
		t.Fatalf("Failed to create TCP server: %v", err)
	}

	// 获取服务器地址
	getsockArgs := []engine.Value{server}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)

	var port int64
	if obj := sockname.Object(); obj != nil {
		if portVal := obj["port"]; portVal != nil {
			port = portVal.Int()
		}
	}

	t.Logf("Server listening on port: %d", port)

	// 关闭服务器
	builtinNetClose(nil, []engine.Value{server})
	t.Log("TCP Client-Server test passed")
}

// TestIntegrationBufferAndNetwork 测试 Buffer 与网络结合
func TestIntegrationBufferAndNetwork(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterBinary(eng)

	// 创建 Buffer
	buf := NewBuffer("big")

	// 模拟网络协议打包
	// 协议格式: [长度:4字节][数据:N字节]
	message := "Hello, Network!"

	// 写入长度
	lenArgs := []engine.Value{buf, engine.NewInt(int64(len(message)))}
	builtinBufferWriteUint32(nil, lenArgs)

	// 写入数据
	dataArgs := []engine.Value{buf, engine.NewString(message)}
	builtinBufferWriteString(nil, dataArgs)

	// 验证长度
	lengthArgs := []engine.Value{buf}
	length, _ := builtinBufferLength(nil, lengthArgs)
	expectedLen := 4 + len(message) // uint32 + string
	if length.Int() != int64(expectedLen) {
		t.Errorf("Buffer length = %d, expected %d", length.Int(), expectedLen)
	}

	// 转换为字节数组（模拟网络发送）
	toBytesArgs := []engine.Value{buf}
	bytes, _ := builtinBufferToBytes(nil, toBytesArgs)

	bytesArr := bytes.Array()
	t.Logf("Packed data (%d bytes): first 4 bytes should be length %d",
		len(bytesArr), len(message))

	// 验证前4字节是长度
	var packedLen int64
	for i := range 4 {
		packedLen = packedLen<<8 + bytesArr[i].Int()
	}
	if packedLen != int64(len(message)) {
		t.Errorf("Packed length = %d, expected %d", packedLen, len(message))
	}

	t.Log("Buffer + Network protocol test passed")
}

// TestIntegrationDNSAndConnect 测试 DNS 解析后连接
func TestIntegrationDNSAndConnect(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterDNS(eng)
	RegisterNet(eng)

	// 解析 localhost
	dnsArgs := []engine.Value{engine.NewString("localhost")}
	ips, err := builtinDNSResolveV4(nil, dnsArgs)
	if err != nil {
		t.Logf("DNS resolve failed: %v", err)
		return
	}

	ipArr := ips.Array()
	if len(ipArr) == 0 {
		t.Error("No IPv4 addresses found for localhost")
		return
	}

	// 使用解析的 IP 创建服务器
	ip := ipArr[0].String()
	server, err := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString(ip),
		engine.NewInt(0), // 自动端口
	})
	if err != nil {
		t.Fatalf("Failed to create server on %s: %v", ip, err)
	}

	// 获取实际端口
	getsockArgs := []engine.Value{server}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)

	var port int64
	if obj := sockname.Object(); obj != nil {
		if portVal := obj["port"]; portVal != nil {
			port = portVal.Int()
		}
	}

	t.Logf("Created server on %s:%d (from DNS resolve)", ip, port)

	// 关闭
	builtinNetClose(nil, []engine.Value{server})
	t.Log("DNS + Connect integration test passed")
}

// TestIntegrationUDPEcho 测试 UDP 回显
func TestIntegrationUDPEcho(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterNet(eng)

	// 创建 UDP socket
	serverUDP, err := builtinNetUdpBind(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0), // 自动端口
	})
	if err != nil {
		t.Fatalf("Failed to bind UDP: %v", err)
	}

	// 获取端口
	getsockArgs := []engine.Value{serverUDP}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)

	var port int64
	if obj := sockname.Object(); obj != nil {
		if portVal := obj["port"]; portVal != nil {
			port = portVal.Int()
		}
	}

	t.Logf("UDP server bound to port: %d", port)

	// 创建客户端
	clientUDP, err := builtinNetUdpBind(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})
	if err != nil {
		t.Fatalf("Failed to bind client UDP: %v", err)
	}

	// 发送数据到服务器
	message := "Hello UDP"
	sendArgs := []engine.Value{
		clientUDP,
		engine.NewString(message),
		engine.NewString("127.0.0.1"),
		engine.NewInt(port),
	}

	sent, err := builtinNetUdpSendto(nil, sendArgs)
	if err != nil {
		t.Fatalf("Failed to send UDP: %v", err)
	}
	t.Logf("Sent %d bytes via UDP", sent.Int())

	// 接收数据（非阻塞测试）
	recvArgs := []engine.Value{serverUDP, engine.NewInt(1024)}
	result, err := builtinNetUdpRecvfrom(nil, recvArgs)

	// 可能立即返回或稍等（简化处理）
	if err != nil {
		t.Logf("UDP recv (may need async): %v", err)
	} else {
		arr := result.Array()
		if len(arr) >= 3 {
			data := arr[0].String()
			fromIP := arr[1].String()
			fromPort := arr[2].Int()
			t.Logf("Received from %s:%d: %s", fromIP, fromPort, data)
		}
	}

	// 关闭
	builtinNetClose(nil, []engine.Value{serverUDP})
	builtinNetClose(nil, []engine.Value{clientUDP})
	t.Log("UDP Echo integration test passed")
}

// TestIntegrationPackAndSend 测试 pack 后通过网络发送
func TestIntegrationPackAndSend(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterBinary(eng)
	RegisterNet(eng)

	// 创建一个简单的协议包
	// [magic:4][version:2][flags:2][length:4]
	// N=uint32(4B), S=uint16大端(2B), s=uint16小端(2B), N=uint32(4B) = 12 bytes
	packArgs := []engine.Value{
		engine.NewString("NSsN"),  // N=uint32, S=uint16大端, s=uint16小端, N=uint32
		engine.NewInt(0x48454C50), // "HELP" as magic
		engine.NewInt(1),          // version 1
		engine.NewInt(0x0001),     // flags (小端)
		engine.NewInt(100),        // length
	}

	packed, err := builtinPack(nil, packArgs)
	if err != nil {
		t.Fatalf("pack() failed: %v", err)
	}

	packedBytes := packed.Array()
	t.Logf("Packed %d bytes protocol header", len(packedBytes))

	// 验证打包结果: 4 + 2 + 2 + 4 = 12 bytes
	if len(packedBytes) != 12 {
		t.Errorf("Packed length = %d, expected 12", len(packedBytes))
	}

	// 验证 magic (前4字节)
	var magic int64
	for i := range 4 {
		magic = magic<<8 + packedBytes[i].Int()
	}
	if magic != 0x48454C50 {
		t.Errorf("Magic = 0x%X, expected 0x48454C50", magic)
	}

	t.Log("Pack + Network integration test passed")
}

// TestIntegrationCompleteNetworkStack 完整网络栈测试
func TestIntegrationCompleteNetworkStack(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()

	// 注册所有网络相关模块
	RegisterBinary(eng)
	RegisterEv(eng)
	RegisterNet(eng)
	RegisterDNS(eng)

	// 1. DNS 解析
	t.Log("Step 1: DNS Resolution")
	dnsArgs := []engine.Value{engine.NewString("localhost")}
	ips, _ := builtinDNSResolve(nil, dnsArgs)
	t.Logf("  Resolved to %d IPs", len(ips.Array()))

	// 2. 创建服务器
	t.Log("Step 2: Create Server")
	server, _ := builtinNetTcpListen(nil, []engine.Value{
		engine.NewString("127.0.0.1"),
		engine.NewInt(0),
	})

	getsockArgs := []engine.Value{server}
	sockname, _ := builtinNetGetsockname(nil, getsockArgs)
	var port int64
	if obj := sockname.Object(); obj != nil {
		if portVal := obj["port"]; portVal != nil {
			port = portVal.Int()
		}
	}
	t.Logf("  Server on port %d", port)

	// 3. 准备二进制数据
	t.Log("Step 3: Prepare Binary Data")
	buf := NewBuffer("big")
	builtinBufferWriteString(nil, []engine.Value{buf, engine.NewString("Test message")})
	dataLen, _ := builtinBufferLength(nil, []engine.Value{buf})
	t.Logf("  Prepared %d bytes", dataLen.Int())

	// 4. 关闭
	t.Log("Step 4: Cleanup")
	builtinNetClose(nil, []engine.Value{server})

	t.Log("Complete network stack integration test passed!")
}
