package stdlib

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/gnuos/jpl/engine"
	"golang.org/x/crypto/curve25519"
)

func RegisterEcc(e *engine.Engine) {
	e.RegisterFunc("ed25519_generate_key", builtinEd25519GenerateKey)
	e.RegisterFunc("ed25519_sign", builtinEd25519Sign)
	e.RegisterFunc("ed25519_verify", builtinEd25519Verify)
	e.RegisterFunc("ed25519_public_key", builtinEd25519PublicKey)
	e.RegisterFunc("x25519_generate_key", builtinX25519GenerateKey)
	e.RegisterFunc("x25519_shared_secret", builtinX25519SharedSecret)
	e.RegisterFunc("x25519_public_key", builtinX25519PublicKey)

	e.RegisterModule("ecc", map[string]engine.GoFunction{
		"ed25519_generate_key": builtinEd25519GenerateKey,
		"ed25519_sign":         builtinEd25519Sign,
		"ed25519_verify":       builtinEd25519Verify,
		"ed25519_public_key":   builtinEd25519PublicKey,
		"x25519_generate_key":  builtinX25519GenerateKey,
		"x25519_shared_secret": builtinX25519SharedSecret,
		"x25519_public_key":    builtinX25519PublicKey,
	})
}

func EccNames() []string {
	return []string{
		"ed25519_generate_key", "ed25519_sign", "ed25519_verify", "ed25519_public_key",
		"x25519_generate_key", "x25519_shared_secret", "x25519_public_key",
	}
}

func builtinEd25519GenerateKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("ed25519_generate_key() expects 0 arguments, got %d", len(args))
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519_generate_key() failed: %v", err)
	}

	return engine.NewObject(map[string]engine.Value{
		"public_key":  engine.NewString(base64.StdEncoding.EncodeToString(pubKey)),
		"private_key": engine.NewString(base64.StdEncoding.EncodeToString(privKey)),
	}), nil
}

func builtinEd25519Sign(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("ed25519_sign() expects 2 arguments, got %d", len(args))
	}

	msg := args[0].String()
	privKeyStr := args[1].String()

	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("ed25519_sign() invalid private key: %v", err)
	}

	if len(privKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("ed25519_sign() invalid private key size")
	}

	signature := ed25519.Sign(privKey, []byte(msg))
	return engine.NewString(base64.StdEncoding.EncodeToString(signature)), nil
}

func builtinEd25519Verify(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("ed25519_verify() expects 3 arguments, got %d", len(args))
	}

	msg := args[0].String()
	sigStr := args[1].String()
	pubKeyStr := args[2].String()

	sig, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return nil, fmt.Errorf("ed25519_verify() invalid signature: %v", err)
	}

	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, fmt.Errorf("ed25519_verify() invalid public key: %v", err)
	}

	if len(pubKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("ed25519_verify() invalid public key size")
	}

	valid := ed25519.Verify(pubKey, []byte(msg), sig)
	return engine.NewBool(valid), nil
}

func builtinEd25519PublicKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ed25519_public_key() expects 1 argument, got %d", len(args))
	}

	privKeyStr := args[0].String()

	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("ed25519_public_key() invalid private key: %v", err)
	}

	if len(privKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("ed25519_public_key() invalid private key size")
	}

	pubKey := ed25519.PrivateKey(privKey).Public().(ed25519.PublicKey)
	return engine.NewString(base64.StdEncoding.EncodeToString(pubKey)), nil
}

func builtinX25519GenerateKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("x25519_generate_key() expects 0 arguments, got %d", len(args))
	}

	var privKey [32]byte
	if _, err := rand.Read(privKey[:]); err != nil {
		return nil, fmt.Errorf("x25519_generate_key() failed: %v", err)
	}

	privKey[0] &= 248
	privKey[31] &= 127
	privKey[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)

	return engine.NewObject(map[string]engine.Value{
		"public_key":  engine.NewString(base64.StdEncoding.EncodeToString(pubKey[:])),
		"private_key": engine.NewString(base64.StdEncoding.EncodeToString(privKey[:])),
	}), nil
}

func builtinX25519SharedSecret(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("x25519_shared_secret() expects 2 arguments, got %d", len(args))
	}

	privKeyStr := args[0].String()
	pubKeyStr := args[1].String()

	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("x25519_shared_secret() invalid private key: %v", err)
	}

	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, fmt.Errorf("x25519_shared_secret() invalid public key: %v", err)
	}

	if len(privKey) != 32 || len(pubKey) != 32 {
		return nil, fmt.Errorf("x25519_shared_secret() invalid key size")
	}

	var privKeyArr []byte
	var pubKeyArr []byte
	copy(privKeyArr[:31], privKey)
	copy(pubKeyArr[:31], pubKey)

	sharedSecret, err := curve25519.X25519(privKeyArr, pubKeyArr)
	if err != nil {
		return nil, fmt.Errorf("x25519_shared_secret() failed to generate")
	}

	return engine.NewString(base64.StdEncoding.EncodeToString(sharedSecret[:31])), nil
}

func builtinX25519PublicKey(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("x25519_public_key() expects 1 argument, got %d", len(args))
	}

	privKeyStr := args[0].String()

	privKey, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, fmt.Errorf("x25519_public_key() invalid private key: %v", err)
	}

	if len(privKey) != 32 {
		return nil, fmt.Errorf("x25519_public_key() invalid private key size")
	}

	var privKeyArr [32]byte
	copy(privKeyArr[:], privKey)

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKeyArr)

	return engine.NewString(base64.StdEncoding.EncodeToString(pubKey[:])), nil
}
