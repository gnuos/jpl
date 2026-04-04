package stdlib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/gnuos/jpl/engine"
)

// =============================================================================
// 加密模块 - 提供 Hash、HMAC、编码、AES 加密功能
// =============================================================================

// RegisterCrypto 注册加密函数到引擎
func RegisterCrypto(e *engine.Engine) {
	// Hash 函数（与 hash.go 中的不重复）
	// hash.go 已有: md5, sha1, sha1_file, md5_file, crc32
	// 这里添加: sha256, sha512, hmac
	e.RegisterFunc("sha256", builtinSHA256)
	e.RegisterFunc("sha512", builtinSHA512)
	e.RegisterFunc("hmac_sha256", builtinHMACSHA256)
	e.RegisterFunc("hmac_sha512", builtinHMACSHA512)

	// Hex 编码（hash.go 没有）
	e.RegisterFunc("hex_encode", builtinHexEncode)
	e.RegisterFunc("hex_decode", builtinHexDecode)

	// AES 加密
	e.RegisterFunc("aes_encrypt", builtinAESEncrypt)
	e.RegisterFunc("aes_decrypt", builtinAESDecrypt)

	// bcrypt 密码哈希
	RegisterBcrypt(e)

	// ECC 椭圆曲线（Ed25519 + X25519）
	RegisterEcc(e)

	// RSA 加密/签名
	RegisterRsa(e)

	// 模块注册 - import "crypto" 可用
	// 注意：base64 函数已在 hash 模块中注册，这里通过 crypto 模块重新导出
	e.RegisterModule("crypto", map[string]engine.GoFunction{
		"sha256":      builtinSHA256,
		"sha512":      builtinSHA512,
		"hmac_sha256": builtinHMACSHA256,
		"hmac_sha512": builtinHMACSHA512,
		"hex_encode":  builtinHexEncode,
		"hex_decode":  builtinHexDecode,
		"aes_encrypt": builtinAESEncrypt,
		"aes_decrypt": builtinAESDecrypt,
		// base64 函数引用 hash.go 中的实现（通过 RegisterHash 注册）
		"base64_encode": builtinBase64Encode,
		"base64_decode": builtinBase64Decode,
		// bcrypt
		"bcrypt_hash":   builtinBcryptHash,
		"bcrypt_verify": builtinBcryptVerify,
		"bcrypt_cost":   builtinBcryptCost,
		// ECC
		"ed25519_generate_key": builtinEd25519GenerateKey,
		"ed25519_sign":         builtinEd25519Sign,
		"ed25519_verify":       builtinEd25519Verify,
		"ed25519_public_key":   builtinEd25519PublicKey,
		"x25519_generate_key":  builtinX25519GenerateKey,
		"x25519_shared_secret": builtinX25519SharedSecret,
		"x25519_public_key":    builtinX25519PublicKey,
		// RSA
		"rsa_generate_key": builtinRSAGenerateKey,
		"rsa_encrypt":      builtinRSAEncrypt,
		"rsa_decrypt":      builtinRSADecrypt,
		"rsa_sign":         builtinRSASign,
		"rsa_verify":       builtinRSAVerify,
		"rsa_public_key":   builtinRSAPublicKey,
	})
}

// CryptoNames 返回加密函数名称列表
func CryptoNames() []string {
	return []string{
		"sha256", "sha512", "sha1",
		"hmac_sha256", "hmac_sha512",
		"base64_encode", "base64_decode",
		"hex_encode", "hex_decode",
		"aes_encrypt", "aes_decrypt",
	}
}

// builtinSHA256 计算 SHA-256 哈希
// sha256(data) → hex_string
//
// 示例：
//
//	sha256("Hello World")  // → "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e"
func builtinSHA256(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha256() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	hash := sha256.Sum256([]byte(data))
	return engine.NewString(hex.EncodeToString(hash[:])), nil
}

// builtinSHA512 计算 SHA-512 哈希
// sha512(data) → hex_string
func builtinSHA512(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha512() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	hash := sha512.Sum512([]byte(data))
	return engine.NewString(hex.EncodeToString(hash[:])), nil
}

// builtinHMACSHA256 计算 HMAC-SHA256
//
// HMAC（Hash-based Message Authentication Code）用于：
//   - 消息完整性验证：确认消息未被篡改
//   - 消息身份验证：确认消息来自声称的发送者
//
// 与普通哈希的区别：
//   - 普通哈希（sha256）：只验证完整性，不验证来源
//   - HMAC：同时验证完整性和来源（需要共享密钥）
//
// hmac_sha256(key, data) → hex_string
//
// 参数：
//   - key: 共享密钥（字符串），长度无限制但建议至少 32 字节
//   - data: 要认证的消息（字符串）
//
// 返回值：
//   - 64 个十六进制字符的 HMAC 值
//
// 使用示例：
//
//	// 发送方计算 HMAC
//	$key = "my_secret_key"
//	$message = "Hello, World!"
//	$hmac = hmac_sha256($key, $message)
//	// → "7e2a7a8c3b5d9f1e4a6c8b2d0f5e7a9c1b3d5e7f8a9c1d3e5f7a8b9c0d1e2"
//
//	// 接收方验证 HMAC
//	$received_hmac = "7e2a7a8c3b5d9f1e4a6c8b2d0f5e7a9c1b3d5e7f8a9c1d3e5f7a8b9c0d1e2"
//	$calculated = hmac_sha256($key, $message)
//	if ($received_hmac == $calculated) {
//	    println "消息验证成功"
//	} else {
//	    println "消息被篡改或来源不可信"
//	}
//
// 常见用途：
//   - API 请求签名：验证请求参数未被篡改
//   - 消息验证：确保消息在传输过程中未被修改
//   - 密码存储辅助：用于密钥派生（但建议使用 bcrypt）
func builtinHMACSHA256(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("hmac_sha256() expects 2 arguments, got %d", len(args))
	}

	key := args[0].String()
	data := args[1].String()

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return engine.NewString(hex.EncodeToString(mac.Sum(nil))), nil
}

// builtinHMACSHA512 计算 HMAC-SHA512
// hmac_sha512(key, data) → hex_string
func builtinHMACSHA512(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("hmac_sha512() expects 2 arguments, got %d", len(args))
	}

	key := args[0].String()
	data := args[1].String()

	mac := hmac.New(sha512.New, []byte(key))
	mac.Write([]byte(data))
	return engine.NewString(hex.EncodeToString(mac.Sum(nil))), nil
}

// builtinHexEncode Hex 编码
// hex_encode(data) → hex_string
//
// 示例：
//
//	hex_encode("Hello")  // → "48656c6c6f"
func builtinHexEncode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("hex_encode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	encoded := hex.EncodeToString([]byte(data))
	return engine.NewString(encoded), nil
}

// builtinHexDecode Hex 解码
// hex_decode(data) → string | null
//
// 示例：
//
//	hex_decode("48656c6c6f")  // → "Hello"
func builtinHexDecode(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("hex_decode() expects 1 argument, got %d", len(args))
	}

	data := args[0].String()
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return engine.NewNull(), nil
	}
	return engine.NewString(string(decoded)), nil
}

// builtinAESEncrypt AES-GCM 加密
//
// ⚠️ 安全说明：
//   - 每次加密会自动生成随机 nonce，无需手动管理
//   - 相同数据+相同密钥会产生不同的密文（nonce 随机）
//   - 密钥必须保密，不要硬编码在代码中
//   - 使用安全的随机数生成器创建密钥
//
// aes_encrypt(data, key) → base64_string
//
// 参数：
//   - data: 要加密的明文字符串（任意长度）
//   - key: 64 个十六进制字符（32 字节）的 AES-256 密钥
//
// 返回值：
//   - base64 编码的密文，格式为：nonce(12字节) || 密文 || 认证标签(16字节)
//   - error: 参数错误或加密失败
//
// 密钥生成示例（在 JPL 中）：
//
//	// 生成随机密钥（需要 64 个十六进制字符 = 32 字节）
//	$random_bytes = []
//	for ($i = 0; $i < 32; $i++) {
//	    $random_bytes = push($random_bytes, floor(random() * 256))
//	}
//	$key = ""
//	for ($b in $random_bytes) {
//	    $key = $key . sprintf("%02x", $b)
//	}
//
// 使用示例：
//
//	$key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
//	$encrypted = aes_encrypt("Secret Message", $key)
//	// → base64 字符串
//
//	// 解密
//	$decrypted = aes_decrypt($encrypted, $key)
//	// → "Secret Message"
func builtinAESEncrypt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("aes_encrypt() expects 2 arguments, got %d", len(args))
	}

	data := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("aes_encrypt() expects string data, got %s", args[0].Type())
	}

	keyHex := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("aes_encrypt() expects string key, got %s", args[1].Type())
	}

	// 解码 hex 密钥
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("aes_encrypt() invalid key hex: %v", err)
	}

	// AES-256 需要 32 字节密钥
	if len(key) != 32 {
		return nil, fmt.Errorf("aes_encrypt() key must be 32 bytes (64 hex chars), got %d bytes", len(key))
	}

	// 创建 AES 块
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes_encrypt() failed to create cipher: %v", err)
	}

	// 使用 GCM 模式（自动处理 IV 和认证）
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes_encrypt() failed to create GCM: %v", err)
	}

	// 加密数据
	// GCM.Seal 会自动生成 nonce 并附加到密文前面
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// 返回 base64 编码
	return engine.NewString(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

// builtinAESDecrypt AES-GCM 解密
//
// ⚠️ 安全说明：
//   - 必须使用与加密时相同的密钥才能正确解密
//   - 密钥错误或密文被篡改会导致解密失败（返回 null）
//   - GCM 模式提供认证校验，篡改后的密文无法解密
//   - 不要在错误消息中透露密钥信息
//
// aes_decrypt(data, key) → string | null
//
// 参数：
//   - data: base64 编码的加密数据（由 aes_encrypt 生成）
//   - key: 64 个十六进制字符（32 字节）的 AES-256 密钥
//
// 返回值：
//   - 解密后的明文字符串
//   - null: 解密失败（密钥错误、密文篡改、格式错误等）
//
// 使用示例：
//
//	$key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
//	$encrypted = aes_encrypt("Secret", $key)
//	$decrypted = aes_decrypt($encrypted, $key)  // → "Secret"
//
//	// 错误的密钥返回 null
//	$wrong_key = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
//	$result = aes_decrypt($encrypted, $wrong_key)  // → null
//
//	// 密文被篡改返回 null
//	$tampered = "xxxx" . substr($encrypted, 4)
//	$result = aes_decrypt($tampered, $key)  // → null
func builtinAESDecrypt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("aes_decrypt() expects 2 arguments, got %d", len(args))
	}

	data := args[0].String()
	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("aes_decrypt() expects string data, got %s", args[0].Type())
	}

	keyHex := args[1].String()
	if args[1].Type() != engine.TypeString {
		return nil, fmt.Errorf("aes_decrypt() expects string key, got %s", args[1].Type())
	}

	// 解码 base64
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return engine.NewNull(), nil
	}

	// 解码 hex 密钥
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return engine.NewNull(), nil
	}

	// AES-256 需要 32 字节密钥
	if len(key) != 32 {
		return engine.NewNull(), nil
	}

	// 创建 AES 块
	block, err := aes.NewCipher(key)
	if err != nil {
		return engine.NewNull(), nil
	}

	// 使用 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return engine.NewNull(), nil
	}

	// 提取 nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return engine.NewNull(), nil
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewString(string(plaintext)), nil
}

// CryptoSigs returns function signatures for REPL :doc command.
func CryptoSigs() map[string]string {
	return map[string]string{
		"aes_encrypt": "aes_encrypt(data, key) → string  — AES-256-GCM encrypt (base64 output)",
		"aes_decrypt": "aes_decrypt(data, key) → string  — AES-256-GCM decrypt",
		"hmac_sha256": "hmac_sha256(key, data) → string  — HMAC-SHA256",
		"hmac_sha512": "hmac_sha512(key, data) → string  — HMAC-SHA512",
	}
}
