package stdlib

import (
	"encoding/base64"
	"fmt"

	"github.com/gnuos/jpl/engine"
	"golang.org/x/crypto/bcrypt"
)

func RegisterBcrypt(e *engine.Engine) {
	e.RegisterFunc("bcrypt_hash", builtinBcryptHash)
	e.RegisterFunc("bcrypt_verify", builtinBcryptVerify)
	e.RegisterFunc("bcrypt_cost", builtinBcryptCost)

	e.RegisterModule("bcrypt", map[string]engine.GoFunction{
		"hash":   builtinBcryptHash,
		"verify": builtinBcryptVerify,
		"cost":   builtinBcryptCost,
	})
}

func BcryptNames() []string {
	return []string{"bcrypt_hash", "bcrypt_verify", "bcrypt_cost"}
}

func builtinBcryptHash(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bcrypt_hash() expects 1 argument, got %d", len(args))
	}

	password := args[0].String()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt_hash() failed: %v", err)
	}

	return engine.NewString(string(hash)), nil
}

func builtinBcryptVerify(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("bcrypt_verify() expects 2 arguments, got %d", len(args))
	}

	password := args[0].String()
	hash := args[1].String()

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return engine.NewBool(false), nil
	}

	return engine.NewBool(true), nil
}

func builtinBcryptCost(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("bcrypt_cost() expects 1 argument, got %d", len(args))
	}

	hash := args[0].String()
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return engine.NewNull(), nil
	}

	return engine.NewInt(int64(cost)), nil
}

type BcryptValue struct {
	hash []byte
}

func (b *BcryptValue) Type() engine.ValueType { return engine.TypeObject }
func (b *BcryptValue) IsNull() bool           { return false }
func (b *BcryptValue) Bool() bool             { return b.hash != nil }
func (b *BcryptValue) Int() int64             { return 0 }
func (b *BcryptValue) Float() float64         { return 0 }
func (b *BcryptValue) String() string         { return "bcrypt:" + base64.StdEncoding.EncodeToString(b.hash) }
func (b *BcryptValue) Stringify() string      { return b.String() }
func (b *BcryptValue) Array() []engine.Value  { return nil }
func (b *BcryptValue) Object() map[string]engine.Value {
	return map[string]engine.Value{
		"hash": engine.NewString(string(b.hash)),
	}
}
func (b *BcryptValue) Len() int                             { return 0 }
func (b *BcryptValue) Equals(other engine.Value) bool       { return false }
func (b *BcryptValue) Less(other engine.Value) bool         { return false }
func (b *BcryptValue) Greater(other engine.Value) bool      { return false }
func (b *BcryptValue) LessEqual(other engine.Value) bool    { return false }
func (b *BcryptValue) GreaterEqual(other engine.Value) bool { return false }
func (b *BcryptValue) ToBigInt() engine.Value               { return engine.NewInt(0) }
func (b *BcryptValue) ToBigDecimal() engine.Value           { return engine.NewFloat(0) }
func (b *BcryptValue) Add(other engine.Value) engine.Value  { return b }
func (b *BcryptValue) Sub(other engine.Value) engine.Value  { return b }
func (b *BcryptValue) Mul(other engine.Value) engine.Value  { return b }
func (b *BcryptValue) Div(other engine.Value) engine.Value  { return b }
func (b *BcryptValue) Mod(other engine.Value) engine.Value  { return b }
func (b *BcryptValue) Negate() engine.Value                 { return b }
