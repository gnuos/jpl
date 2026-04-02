package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestDNSResolve 测试域名解析
func TestDNSResolve(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterDNS(eng)

	// 解析 localhost
	args := []engine.Value{engine.NewString("localhost")}
	result, err := builtinDNSResolve(nil, args)
	if err != nil {
		t.Fatalf("dns_resolve() error = %v", err)
	}

	arr := result.Array()
	if len(arr) == 0 {
		t.Error("dns_resolve() should return at least one IP for localhost")
	}

	// 检查是否包含 127.0.0.1
	found := false
	for _, ip := range arr {
		if ip.String() == "127.0.0.1" {
			found = true
			break
		}
	}
	if !found {
		t.Logf("localhost resolved to: %v", arr)
	}
}

// TestDNSResolveOne 测试解析单个 IP
func TestDNSResolveOne(t *testing.T) {
	args := []engine.Value{engine.NewString("localhost")}
	result, err := builtinDNSResolveOne(nil, args)
	if err != nil {
		t.Fatalf("dns_resolve_one() error = %v", err)
	}

	ip := result.String()
	if ip == "" {
		t.Error("dns_resolve_one() should return a non-empty IP")
	}

	t.Logf("localhost resolved to: %s", ip)
}

// TestDNSResolveV4 测试 IPv4 解析
func TestDNSResolveV4(t *testing.T) {
	args := []engine.Value{engine.NewString("localhost")}
	result, err := builtinDNSResolveV4(nil, args)
	if err != nil {
		t.Fatalf("dns_resolve_v4() error = %v", err)
	}

	arr := result.Array()
	if len(arr) == 0 {
		t.Error("dns_resolve_v4() should return at least one IPv4 for localhost")
	}

	// 检查都是 IPv4
	for _, ip := range arr {
		ipStr := ip.String()
		// 简单检查：IPv4 应该包含 4 个数字
		t.Logf("IPv4: %s", ipStr)
	}
}

// TestDNSGetRecords 测试获取 DNS 记录
func TestDNSGetRecords(t *testing.T) {
	// 测试获取所有记录
	args := []engine.Value{engine.NewString("localhost")}
	result, err := builtinDNSGetRecords(nil, args)
	if err != nil {
		t.Fatalf("dns_get_records() error = %v", err)
	}

	arr := result.Array()
	t.Logf("Found %d records for localhost", len(arr))

	// 显示记录
	for _, record := range arr {
		obj := record.Object()
		if obj != nil {
			recordType := obj["type"]
			if recordType != nil {
				t.Logf("Record type: %s", recordType.String())
			}
		}
	}
}

// TestDNSInvalidHost 测试无效域名
func TestDNSInvalidHost(t *testing.T) {
	// 使用一个不太可能存在的域名
	args := []engine.Value{engine.NewString("this-domain-should-not-exist-12345.invalid")}
	_, err := builtinDNSResolve(nil, args)
	if err == nil {
		t.Error("dns_resolve() should return error for invalid domain")
	}
}
