package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	pubKey := privKey.Public()
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestPrivateKeySign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	msg := []byte("hello blockchain")
	sig := privKey.Sign(msg)
	assert.True(t, sig.Verify(pubKey, msg))

	// test with invalid message
	assert.False(t, sig.Verify(pubKey, []byte("foo")))

	// test with invalid public key
	invalidPubKey := GeneratePrivateKey()
	invalidPrivKey := invalidPubKey.Public()
	assert.False(t, sig.Verify(invalidPrivKey, msg))
}

func TestPublickeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	address := pubKey.Address()
	assert.Equal(t, AddressLen, len(address.Bytes()))
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		addressStr = "ae57bba3aaf9c09ff974a12454010eb392a2424c"
		seed       = "7bc7e3eb7bd703057cf3d7bd61c8ac277b2167584d9dc3aa94350e07e2f43ae7"
		privKey    = NewPrivateKeyFromString(seed)
	)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	assert.Equal(t, addressStr, address.String())
}
