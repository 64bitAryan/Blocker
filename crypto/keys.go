package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
)

const (
	PrivKeyLen   = 64
	PubKeyLen    = 32
	SeedLen      = 32
	AddressLen   = 20
	SignatureLen = 64
)

type PrivateKeys struct {
	key ed25519.PrivateKey
}

type PublicKeys struct {
	key ed25519.PublicKey
}

type Signature struct {
	value []byte
}

type Address struct {
	value []byte
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKeys {
	if len(seed) != SeedLen {
		panic("invalid seed length, must be 32")
	}
	return &PrivateKeys{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func NewPrivateKeyFromSeedStr(seed string) *PrivateKeys {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(seedBytes)
}

func NewPrivateKeyFromString(s string) *PrivateKeys {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func GeneratePrivateKey() *PrivateKeys {
	seed := make([]byte, SeedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return &PrivateKeys{
		key: ed25519.NewKeyFromSeed(seed),
	}
}

func (p *PrivateKeys) Bytes() []byte {
	return p.key
}

func (p *PrivateKeys) Public() *PublicKeys {
	b := make([]byte, PubKeyLen)
	copy(b, p.key[32:])
	return &PublicKeys{
		key: b,
	}
}

func SignatureFromBytes(b []byte) *Signature {
	if len(b) != SignatureLen {
		panic("length of bytes not equal to 64")
	}
	return &Signature{
		value: b,
	}
}

func (p *PrivateKeys) Sign(msg []byte) *Signature {
	s := ed25519.Sign(p.key, msg)
	return &Signature{
		value: s,
	}
}

func PublicKeyFromBytes(b []byte) *PublicKeys {
	if len(b) != PubKeyLen {
		panic("invalid key length")
	}
	return &PublicKeys{
		key: ed25519.PublicKey(b),
	}
}

func (p *PublicKeys) Bytes() []byte {
	return p.key
}

func (p *PublicKeys) Address() Address {
	return Address{
		value: p.key[len(p.key)-AddressLen:],
	}
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) Verify(pubKey *PublicKeys, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

func (a Address) Bytes() []byte {
	return a.value
}

func AddressFromBytes(b []byte) Address {
	if len(b) != AddressLen {
		panic("length of (address) bytes not equal to 20")
	}
	return Address{
		value: b,
	}
}

// implementing string interface for address
func (a Address) String() string {
	return hex.EncodeToString(a.value)
}
