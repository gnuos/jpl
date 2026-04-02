package stdlib

import (
	"fmt"
	"net"

	"github.com/gnuos/jpl/engine"
)

// RegisterDNS 注册 DNS 解析函数到引擎
// 提供域名解析能力，支持 IPv4/IPv6 地址查询和 DNS 记录获取
//
// 解析策略：
// - 优先使用系统 DNS 配置（/etc/resolv.conf）
// - 支持 A、AAAA、CNAME、MX、NS、TXT 等多种记录类型
// - 自动处理 CNAME 链式解析
//
// 使用场景：
// - 网络编程中的地址解析
// - 服务发现和健康检查
// - DNS 记录监控和分析
func RegisterDNS(e *engine.Engine) {
	e.RegisterFunc("dns_resolve", builtinDNSResolve)
	e.RegisterFunc("dns_resolve_one", builtinDNSResolveOne)
	e.RegisterFunc("dns_resolve_v4", builtinDNSResolveV4)
	e.RegisterFunc("dns_resolve_v6", builtinDNSResolveV6)
	e.RegisterFunc("dns_get_records", builtinDNSGetRecords)

	// 模块注册 — import "dns" 可用
	e.RegisterModule("dns", map[string]engine.GoFunction{
		"resolve":     builtinDNSResolve,
		"resolve_one": builtinDNSResolveOne,
		"resolve_v4":  builtinDNSResolveV4,
		"resolve_v6":  builtinDNSResolveV6,
		"get_records": builtinDNSGetRecords,
	})
}

// DNSNames 返回 DNS 函数名称列表
func DNSNames() []string {
	return []string{
		"dns_resolve", "dns_resolve_one",
		"dns_resolve_v4", "dns_resolve_v6",
		"dns_get_records",
	}
}

// builtinDNSResolve 解析域名返回所有 IP 地址（IPv4 和 IPv6）
// dns_resolve($host) → [ip1, ip2, ...]
//
// 参数：
//   - args[0]: 域名（字符串），如 "example.com"
//
// 返回值：
//   - 字符串数组，包含所有解析到的 IP 地址
//   - IPv4 格式：如 "93.184.216.34"
//   - IPv6 格式：如 "2606:2800:220:1:248:1893:25c8:1946"
//
// 错误：域名不存在、DNS 服务器不可达、超时等
//
// 注意：返回的顺序不保证，优先返回 IPv4 还是 IPv6 取决于系统配置
func builtinDNSResolve(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dns_resolve() expects 1 argument, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dns_resolve() expects string hostname, got %s", args[0].Type())
	}

	// 使用 net.LookupHost 获取所有 IP（IPv4 和 IPv6）
	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, fmt.Errorf("dns_resolve() failed: %v", err)
	}

	// 转换为数组
	result := make([]engine.Value, len(ips))
	for i, ip := range ips {
		result[i] = engine.NewString(ip)
	}

	return engine.NewArray(result), nil
}

// builtinDNSResolveOne 解析域名返回单个 IP 地址
// dns_resolve_one($host) → ip | null
//
// 参数：
//   - args[0]: 域名（字符串）
//
// 返回值：
//   - 第一个解析到的 IP 地址（字符串）
//   - null: 解析失败或无结果
//
// 用途：
//   - 快速获取可用 IP，不需要关心所有地址
//   - 简单的网络连接场景
//
// 注意：返回的是第一个可用地址，可能是 IPv4 或 IPv6，不保证一致性
func builtinDNSResolveOne(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dns_resolve_one() expects 1 argument, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dns_resolve_one() expects string hostname, got %s", args[0].Type())
	}

	// 使用 net.LookupHost 获取所有 IP，返回第一个
	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, fmt.Errorf("dns_resolve_one() failed: %v", err)
	}

	if len(ips) == 0 {
		return engine.NewNull(), nil
	}

	return engine.NewString(ips[0]), nil
}

// builtinDNSResolveV4 解析域名返回所有 IPv4 地址
// dns_resolve_v4($host) → [ip1, ip2, ...]
//
// 参数：
//   - args[0]: 域名（字符串）
//
// 返回值：
//   - IPv4 地址数组，如 ["93.184.216.34", "93.184.216.35"]
//   - 如果没有 IPv4 记录，返回空数组
//
// 用途：
//   - IPv4 专用网络环境
//   - 需要确定使用 IPv4 的场景
//   - 与仅支持 IPv4 的服务通信
func builtinDNSResolveV4(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dns_resolve_v4() expects 1 argument, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dns_resolve_v4() expects string hostname, got %s", args[0].Type())
	}

	// 使用 net.LookupIP 获取所有 IP，然后过滤 IPv4
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("dns_resolve_v4() failed: %v", err)
	}

	// 过滤 IPv4 地址
	var ipv4s []engine.Value
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, engine.NewString(ip.String()))
		}
	}

	return engine.NewArray(ipv4s), nil
}

// builtinDNSResolveV6 解析域名返回所有 IPv6 地址
// dns_resolve_v6($host) → [ip1, ip2, ...]
//
// 参数：
//   - args[0]: 域名（字符串）
//
// 返回值：
//   - IPv6 地址数组，如 ["2606:2800:220:1:248:1893:25c8:1946"]
//   - 如果没有 IPv6 记录，返回空数组
//
// 用途：
//   - IPv6 优先网络环境
//   - 需要确定使用 IPv6 的场景
//   - 现代云原生应用（Kubernetes 等）
func builtinDNSResolveV6(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("dns_resolve_v6() expects 1 argument, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dns_resolve_v6() expects string hostname, got %s", args[0].Type())
	}

	// 使用 net.LookupIP 获取所有 IP，然后过滤 IPv6
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("dns_resolve_v6() failed: %v", err)
	}

	// 过滤 IPv6 地址（To4() == nil 表示是 IPv6）
	var ipv6s []engine.Value
	for _, ip := range ips {
		if ip.To4() == nil {
			ipv6s = append(ipv6s, engine.NewString(ip.String()))
		}
	}

	return engine.NewArray(ipv6s), nil
}

// builtinDNSGetRecords 获取域名的详细 DNS 记录
// dns_get_records($host [, $type]) → [{type, value, ...}, ...]
//
// 参数：
//   - args[0]: 域名（字符串）
//   - args[1]: 可选，记录类型（字符串），默认为 "ALL"
//
// 支持的记录类型：
//   - "A": IPv4 地址记录
//   - "AAAA": IPv6 地址记录
//   - "CNAME": 别名记录
//   - "MX": 邮件交换记录
//   - "NS": 域名服务器记录
//   - "TXT": 文本记录
//   - "ALL": 所有类型（默认）
//
// 返回值（对象数组）：
//
//	A 记录: {type: "A", ip: "x.x.x.x", host: "example.com"}
//	AAAA 记录: {type: "AAAA", ip: "xxxx:xxxx:...", host: "example.com"}
//	CNAME 记录: {type: "CNAME", target: "canonical.name", host: "example.com"}
//	MX 记录: {type: "MX", priority: 10, target: "mail.example.com", host: "example.com"}
//	NS 记录: {type: "NS", target: "ns1.example.com", host: "example.com"}
//	TXT 记录: {type: "TXT", text: "some text", host: "example.com"}
//
// 用途：
//   - DNS 记录监控和验证
//   - 邮件服务器配置检查
//   - SPF/DKIM/DMARC 记录查询（TXT 记录）
//   - 域名迁移验证
//
// 注意：不同记录类型的查询可能依赖不同的 DNS 服务器，部分记录可能查询失败
func builtinDNSGetRecords(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("dns_get_records() expects 1 or 2 arguments, got %d", len(args))
	}

	host := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("dns_get_records() expects string hostname, got %s", args[0].Type())
	}

	// 记录类型参数（可选）
	recordType := "ALL"
	if len(args) == 2 {
		recordType = args[1].String()
	}

	results := make([]engine.Value, 0)

	// 查询 A 记录（IPv4）
	if recordType == "ALL" || recordType == "A" {
		ips, err := net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				if ip.To4() != nil {
					record := map[string]engine.Value{
						"type": engine.NewString("A"),
						"ip":   engine.NewString(ip.String()),
						"host": engine.NewString(host),
					}
					results = append(results, engine.NewObject(record))
				}
			}
		}
	}

	// 查询 AAAA 记录（IPv6）
	if recordType == "ALL" || recordType == "AAAA" {
		ips, err := net.LookupIP(host)
		if err == nil {
			for _, ip := range ips {
				if ip.To4() == nil {
					record := map[string]engine.Value{
						"type": engine.NewString("AAAA"),
						"ip":   engine.NewString(ip.String()),
						"host": engine.NewString(host),
					}
					results = append(results, engine.NewObject(record))
				}
			}
		}
	}

	// 查询 CNAME 记录
	if recordType == "ALL" || recordType == "CNAME" {
		cname, err := net.LookupCNAME(host)
		if err == nil && cname != "" && cname != host+"." {
			record := map[string]engine.Value{
				"type":   engine.NewString("CNAME"),
				"target": engine.NewString(cname),
				"host":   engine.NewString(host),
			}
			results = append(results, engine.NewObject(record))
		}
	}

	// 查询 MX 记录
	if recordType == "ALL" || recordType == "MX" {
		mxs, err := net.LookupMX(host)
		if err == nil {
			for _, mx := range mxs {
				record := map[string]engine.Value{
					"type":     engine.NewString("MX"),
					"host":     engine.NewString(host),
					"priority": engine.NewInt(int64(mx.Pref)),
					"target":   engine.NewString(mx.Host),
				}
				results = append(results, engine.NewObject(record))
			}
		}
	}

	// 查询 NS 记录
	if recordType == "ALL" || recordType == "NS" {
		nss, err := net.LookupNS(host)
		if err == nil {
			for _, ns := range nss {
				record := map[string]engine.Value{
					"type":   engine.NewString("NS"),
					"host":   engine.NewString(host),
					"target": engine.NewString(ns.Host),
				}
				results = append(results, engine.NewObject(record))
			}
		}
	}

	// 查询 TXT 记录
	if recordType == "ALL" || recordType == "TXT" {
		txts, err := net.LookupTXT(host)
		if err == nil {
			for _, txt := range txts {
				record := map[string]engine.Value{
					"type": engine.NewString("TXT"),
					"host": engine.NewString(host),
					"text": engine.NewString(txt),
				}
				results = append(results, engine.NewObject(record))
			}
		}
	}

	return engine.NewArray(results), nil
}
