package stdlib

import (
	"github.com/gnuos/jpl/engine"
)

// valueToString 将值转换为用户友好的字符串表示
func valueToString(v engine.Value) string {
	return v.String()
}

// RegisterAll 将所有内置函数注册到引擎。
//
// 此方法注册所有内置函数模块，包括：
//   - I/O 函数（print、println 等）
//   - 工具函数（len、type、echo 等）
//   - 函数式编程（map、filter、reduce 等）
//   - 类型检查（is_int、is_string 等）
//   - 类型转换（intval、floatval、strval 等）
//   - 文件 I/O（read、write、append 等）
//   - 动态常量（define、defined）
//   - 哈希/编码（md5、sha1、base64 等）
//   - URL 处理（urlencode、parse_url 等）
//   - GC 控制（gc、gc_info）
//   - 字符串函数（strlen、substr、split 等）
//   - 数组函数（push、pop、map、filter 等）
//   - 数学函数（abs、sin、cos、sqrt 等）
//   - 调试函数（errors、last_error）
//   - 日期时间（time、date、sleep 等）
//   - 求值函数（eval）
//   - 反射 API（typeof、getvar、setvar 等）
//   - 错误处理（error）
//   - 进程控制（exit、die）
//
// 通常在创建引擎后立即调用：
//
//	eng := engine.NewEngine()
//	defer eng.Close()
//	buildin.RegisterAll(eng)  // 注册所有内置函数
func RegisterAll(e *engine.Engine) {
	RegisterPresetConstants(e)
	RegisterIO(e)
	RegisterUtil(e)
	RegisterFunctional(e)
	RegisterTypeCheck(e)
	RegisterTypeConvert(e)
	RegisterFileIO(e)
	RegisterFileIOAsync(e) // 异步文件 IO
	RegisterConstFunc(e)
	RegisterHash(e)
	RegisterURL(e)
	RegisterGC(e)
	RegisterString(e)
	RegisterArray(e)
	RegisterMath(e)
	RegisterDebug(e)
	RegisterDateTime(e)
	RegisterEval(e)
	RegisterReflect(e)
	RegisterError(e)
	RegisterJSON(e)
	RegisterBitwise(e)
	RegisterProcess(e)
	RegisterProcessExt(e) // 进程扩展函数
	RegisterVMFunc(e)
	RegisterSystem(e)
	RegisterBinary(e)
	RegisterEv(e)
	RegisterNet(e)
	RegisterDNS(e)
	RegisterObjectParse(e) // 注册安全的对象解析函数
	RegisterTLS(e)         // TLS/SSL 加密通信
	RegisterHTTP(e)        // HTTP 客户端
	RegisterRe(e)          // 正则表达式
	RegisterCrypto(e)      // 加密模块
	RegisterIP(e)          // IP 地址处理
	RegisterDelete(e)      // delete/unset 函数
	RegisterGzip(e)        // gzip 压缩
	RegisterZlib(e)        // zlib 压缩
	RegisterBrotli(e)      // brotli 压缩
	RegisterZip(e)         // zip 归档
	RegisterTar(e)         // tar 归档
}

// FunctionNames 返回所有内置函数的名称列表。
//
// 此方法收集所有内置函数模块的函数名，用于：
//   - 代码补全（REPL、IDE）
//   - 文档生成
//   - 函数存在性检查
//
// 返回值：
//   - []string: 所有内置函数名列表（已排序去重）
//
// 使用示例：
//
//	names := buildin.FunctionNames()
//	fmt.Printf("可用内置函数: %d 个\n", len(names))
//	// 输出: [assert date echo filter format gc gc_info ...]
func FunctionNames() []string {
	var names []string
	names = append(names, IONames()...)
	names = append(names, UtilNames()...)
	names = append(names, FunctionalNames()...)
	names = append(names, TypeCheckNames()...)
	names = append(names, TypeConvertNames()...)
	names = append(names, FileIONames()...)
	names = append(names, FileIOAsyncNames()...) // 异步文件 IO
	names = append(names, ConstFuncNames()...)
	names = append(names, HashNames()...)
	names = append(names, UrlNames()...)
	names = append(names, GCNames()...)
	names = append(names, StringNames()...)
	names = append(names, ArrayNames()...)
	names = append(names, MathNames()...)
	names = append(names, DebugNames()...)
	names = append(names, DateTimeNames()...)
	names = append(names, EvalNames()...)
	names = append(names, ReflectNames()...)
	names = append(names, ErrorNames()...)
	names = append(names, JSONNames()...)
	names = append(names, BitwiseNames()...)
	names = append(names, ProcessNames()...)
	names = append(names, ProcessExtNames()...) // 进程扩展函数
	names = append(names, VMFuncNames()...)
	names = append(names, SystemNames()...)
	names = append(names, BinaryNames()...)
	names = append(names, EvNames()...)
	names = append(names, NetNames()...)
	names = append(names, DNSNames()...)
	names = append(names, ObjectParseNames()...)
	names = append(names, TLSNames()...)
	names = append(names, HTTPNames()...)
	names = append(names, ReNames()...)
	names = append(names, CryptoNames()...)
	names = append(names, IPNames()...)
	names = append(names, DeleteNames()...)
	names = append(names, GzipNames()...)
	names = append(names, ZlibNames()...)
	names = append(names, ZipNames()...)
	names = append(names, TarNames()...)
	return names
}
