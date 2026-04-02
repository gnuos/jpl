package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

func TestIP2Long(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		ip   string
		want int64
	}{
		{"192.168.1.1", 3232235777},
		{"127.0.0.1", 2130706433},
		{"255.255.255.255", 4294967295},
		{"0.0.0.0", 0},
	}

	for _, tt := range tests {
		result, err := builtinIP2Long(ctx, []engine.Value{engine.NewString(tt.ip)})
		if err != nil {
			t.Errorf("ip2long(%q) error: %v", tt.ip, err)
			continue
		}
		if result.Int() != tt.want {
			t.Errorf("ip2long(%q) = %d, want %d", tt.ip, result.Int(), tt.want)
		}
	}

	// 测试无效 IP
	result, _ := builtinIP2Long(ctx, []engine.Value{engine.NewString("invalid")})
	if result.Type() != engine.TypeNull {
		t.Errorf("ip2long(invalid) should return null")
	}
}

func TestLong2IP(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		val  int64
		want string
	}{
		{3232235777, "192.168.1.1"},
		{2130706433, "127.0.0.1"},
		{0, "0.0.0.0"},
	}

	for _, tt := range tests {
		result, err := builtinLong2IP(ctx, []engine.Value{engine.NewInt(tt.val)})
		if err != nil {
			t.Errorf("long2ip(%d) error: %v", tt.val, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("long2ip(%d) = %q, want %q", tt.val, result.String(), tt.want)
		}
	}
}

func TestIP2LongLong2IPRoundTrip(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"255.255.255.255",
	}

	for _, original := range tests {
		// ip2long
		val, err := builtinIP2Long(ctx, []engine.Value{engine.NewString(original)})
		if err != nil {
			t.Errorf("ip2long(%q) error: %v", original, err)
			continue
		}

		// long2ip
		result, err := builtinLong2IP(ctx, []engine.Value{val})
		if err != nil {
			t.Errorf("long2ip() error: %v", err)
			continue
		}

		if result.String() != original {
			t.Errorf("Round trip failed: original=%q, result=%q", original, result.String())
		}
	}
}

func TestIPParse(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	// 测试 IPv4
	result, err := builtinIPParse(ctx, []engine.Value{engine.NewString("192.168.1.1")})
	if err != nil {
		t.Fatalf("ip_parse error: %v", err)
	}

	obj := result.Object()
	if obj["version"].Int() != 4 {
		t.Errorf("Expected version 4, got %d", obj["version"].Int())
	}
	if obj["type"].String() != "private" {
		t.Errorf("Expected type private, got %s", obj["type"].String())
	}

	// 测试无效 IP
	result, _ = builtinIPParse(ctx, []engine.Value{engine.NewString("invalid")})
	if result.Type() != engine.TypeNull {
		t.Errorf("ip_parse(invalid) should return null")
	}
}

func TestIPToHex(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		ip   string
		want string
	}{
		{"192.168.1.1", "c0a80101"},
		{"127.0.0.1", "7f000001"},
		{"0.0.0.0", "00000000"},
	}

	for _, tt := range tests {
		result, err := builtinIPToHex(ctx, []engine.Value{engine.NewString(tt.ip)})
		if err != nil {
			t.Errorf("ip_to_hex(%q) error: %v", tt.ip, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("ip_to_hex(%q) = %q, want %q", tt.ip, result.String(), tt.want)
		}
	}
}

func TestIPToBin(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	result, err := builtinIPToBin(ctx, []engine.Value{engine.NewString("192.168.1.1")})
	if err != nil {
		t.Fatalf("ip_to_bin error: %v", err)
	}

	expected := "11000000101010000000000100000001"
	if result.String() != expected {
		t.Errorf("ip_to_bin(192.168.1.1) = %q, want %q", result.String(), expected)
	}
}

func TestIPFromHex(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		hex  string
		want string
	}{
		{"c0a80101", "192.168.1.1"},
		{"7f000001", "127.0.0.1"},
	}

	for _, tt := range tests {
		result, err := builtinIPFromHex(ctx, []engine.Value{engine.NewString(tt.hex)})
		if err != nil {
			t.Errorf("ip_from_hex(%q) error: %v", tt.hex, err)
			continue
		}
		if result.String() != tt.want {
			t.Errorf("ip_from_hex(%q) = %q, want %q", tt.hex, result.String(), tt.want)
		}
	}
}

func TestIPFromBin(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	bin := "11000000101010000000000100000001"
	result, err := builtinIPFromBin(ctx, []engine.Value{engine.NewString(bin)})
	if err != nil {
		t.Fatalf("ip_from_bin error: %v", err)
	}

	if result.String() != "192.168.1.1" {
		t.Errorf("ip_from_bin() = %q, want 192.168.1.1", result.String())
	}
}

func TestIPVersion(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		ip       string
		expected int64
	}{
		{"192.168.1.1", 4},
		{"2001:db8::1", 6},
	}

	for _, tt := range tests {
		result, err := builtinIPVersion(ctx, []engine.Value{engine.NewString(tt.ip)})
		if err != nil {
			t.Errorf("ip_version(%q) error: %v", tt.ip, err)
			continue
		}
		if result.Int() != tt.expected {
			t.Errorf("ip_version(%q) = %d, want %d", tt.ip, result.Int(), tt.expected)
		}
	}
}

func TestIPValid(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"2001:db8::1", true},
		{"256.1.1.1", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result, err := builtinIPValid(ctx, []engine.Value{engine.NewString(tt.ip)})
		if err != nil {
			t.Errorf("ip_valid(%q) error: %v", tt.ip, err)
			continue
		}
		if result.Bool() != tt.expected {
			t.Errorf("ip_valid(%q) = %v, want %v", tt.ip, result.Bool(), tt.expected)
		}
	}
}

func TestIPHexBinRoundTrip(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterIP(e)
	ctx := engine.NewContext(e, nil)

	original := "192.168.1.1"

	// 转换为 hex
	hex, err := builtinIPToHex(ctx, []engine.Value{engine.NewString(original)})
	if err != nil {
		t.Fatalf("ip_to_hex error: %v", err)
	}

	// 从 hex 转回
	result, err := builtinIPFromHex(ctx, []engine.Value{hex})
	if err != nil {
		t.Fatalf("ip_from_hex error: %v", err)
	}

	if result.String() != original {
		t.Errorf("Hex round trip failed: original=%q, result=%q", original, result.String())
	}

	// 转换为 bin
	bin, err := builtinIPToBin(ctx, []engine.Value{engine.NewString(original)})
	if err != nil {
		t.Fatalf("ip_to_bin error: %v", err)
	}

	// 从 bin 转回
	result2, err := builtinIPFromBin(ctx, []engine.Value{bin})
	if err != nil {
		t.Fatalf("ip_from_bin error: %v", err)
	}

	if result2.String() != original {
		t.Errorf("Bin round trip failed: original=%q, result=%q", original, result2.String())
	}
}
