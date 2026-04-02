package stdlib

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/gnuos/jpl/engine"
)

func RegisterRsa(e *engine.Engine) {
	e.RegisterFunc("rsa_generate_key", builtinRSAGenerateKey)
	e.RegisterFunc("rsa_encrypt", builtinRSAEncrypt)
	e.RegisterFunc("rsa_decrypt", builtinRSADecrypt)
	e.RegisterFunc("rsa_sign", builtinRSASign)
	e.RegisterFunc("rsa_verify", builtinRSAVerify)
	e.RegisterFunc("rsa_public_key", builtinRSAPublicKey)

	e.RegisterModule("rsa", map[string]engine.GoFunction{
		"generate_key": builtinRSAGenerateKey,
		"encrypt":      builtinRSAEncrypt,
		"decrypt":      builtinRSADecrypt,
		"sign":         builtinRSASign,
		"verify":       builtinRSAVerify,
		"public_key":   builtinRSAPublicKey,
	})
}

func RsaNames() []string {
	return []string{
		"rsa_generate_key", "rsa_encrypt", "rsa_decrypt",
		"rsa_sign", "rsa_verify", "rsa_public_key",
	}
}

func builtinRSAGenerateKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	bits := 2048
	if len(args) >= 1 {
		bits = int(args[0].Int())
		if bits < 512 {
			bits = 512
		}
		if bits > 8192 {
			bits = 8192
		}
	}

	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("rsa_generate_key() failed: %v", err)
	}

	pubKey := &privKey.PublicKey

	return engine.NewObject(map[string]engine.Value{
		"public_key":  engine.NewString(base64.StdEncoding.EncodeToString(encodeRSAPublicKey(pubKey))),
		"private_key": engine.NewString(base64.StdEncoding.EncodeToString(encodeRSAPrivateKey(privKey))),
		"n":           engine.NewString(base64.StdEncoding.EncodeToString(pubKey.N.Bytes())),
		"e":           engine.NewInt(int64(pubKey.E)),
	}), nil
}

func builtinRSAEncrypt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("rsa_encrypt() expects 2 arguments, got %d", len(args))
	}

	msg := args[0].String()
	pubKeyStr := args[1].String()

	pubKey, err := decodeRSAPublicKey(pubKeyStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_encrypt() invalid public key: %v", err)
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(msg))
	if err != nil {
		return nil, fmt.Errorf("rsa_encrypt() failed: %v", err)
	}

	return engine.NewString(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func builtinRSADecrypt(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("rsa_decrypt() expects 2 arguments, got %d", len(args))
	}

	cipherStr := args[0].String()
	privKeyStr := args[1].String()

	ciphertext, err := base64.StdEncoding.DecodeString(cipherStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_decrypt() invalid ciphertext: %v", err)
	}

	privKey, err := decodeRSAPrivateKey(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_decrypt() invalid private key: %v", err)
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("rsa_decrypt() failed: %v", err)
	}

	return engine.NewString(string(plaintext)), nil
}

func builtinRSASign(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("rsa_sign() expects 2 arguments, got %d", len(args))
	}

	msg := args[0].String()
	privKeyStr := args[1].String()

	privKey, err := decodeRSAPrivateKey(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_sign() invalid private key: %v", err)
	}

	h := sha256.New()
	h.Write([]byte(msg))
	digest := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, digest)
	if err != nil {
		return nil, fmt.Errorf("rsa_sign() failed: %v", err)
	}

	return engine.NewString(base64.StdEncoding.EncodeToString(signature)), nil
}

func builtinRSAVerify(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("rsa_verify() expects 3 arguments, got %d", len(args))
	}

	msg := args[0].String()
	sigStr := args[1].String()
	pubKeyStr := args[2].String()

	sig, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_verify() invalid signature: %v", err)
	}

	pubKey, err := decodeRSAPublicKey(pubKeyStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_verify() invalid public key: %v", err)
	}

	h := sha256.New()
	h.Write([]byte(msg))
	digest := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, sig)
	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(true), nil
}

func builtinRSAPublicKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rsa_public_key() expects 1 argument, got %d", len(args))
	}

	privKeyStr := args[0].String()

	privKey, err := decodeRSAPrivateKey(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("rsa_public_key() invalid private key: %v", err)
	}

	pubKey := &privKey.PublicKey
	return engine.NewString(base64.StdEncoding.EncodeToString(encodeRSAPublicKey(pubKey))), nil
}

func encodeRSAPublicKey(pubKey *rsa.PublicKey) []byte {
	nBytes := pubKey.N.Bytes()
	result := make([]byte, len(nBytes)+4)
	copy(result, nBytes)
	result[len(nBytes)] = byte(pubKey.E >> 24)
	result[len(nBytes)+1] = byte(pubKey.E >> 16)
	result[len(nBytes)+2] = byte(pubKey.E >> 8)
	result[len(nBytes)+3] = byte(pubKey.E)
	return result
}

func decodeRSAPublicKey(data string) (*rsa.PublicKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	if len(decoded) < 4 {
		return nil, fmt.Errorf("invalid public key data")
	}

	n := new(rsa.PublicKey)
	n.N.SetBytes(decoded[:len(decoded)-4])
	n.E = int(decoded[len(decoded)-4])<<24 | int(decoded[len(decoded)-3])<<16 | int(decoded[len(decoded)-2])<<8 | int(decoded[len(decoded)-1])

	return n, nil
}

func encodeRSAPrivateKey(privKey *rsa.PrivateKey) []byte {
	return privKey.N.Bytes()
}

func decodeRSAPrivateKey(data string) (*rsa.PrivateKey, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	n := new(rsa.PrivateKey)
	n.PublicKey.N.SetBytes(decoded)
	n.PublicKey.E = 65537
	n.D = new(rsa.PrivateKey).D
	n.Primes = n.Primes[:0]

	return n, nil
}
