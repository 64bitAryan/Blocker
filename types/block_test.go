package types

import (
	"testing"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestCalculateRootHash(t *testing.T) {

	var (
		privKey = crypto.GeneratePrivateKey()
		block   = util.RandomBlock()
		tx      = &proto.Transaction{
			Version: 1,
		}
	)

	block.Transactions = append(block.Transactions, tx)
	SignBlock(privKey, block)
	assert.True(t, VerifyRootHash(block))
	assert.Equal(t, 32, len(block.Header.RootHash))
}

func TestSignVerifyBlock(t *testing.T) {
	var (
		block   = util.RandomBlock()
		privKey = crypto.GeneratePrivateKey()
		pubkey  = privKey.Public()
	)
	sig := SignBlock(privKey, block)
	assert.Equal(t, 64, len(sig.Bytes()))
	assert.True(t, sig.Verify(pubkey, HashBlock(block)))

	assert.Equal(t, block.PublicKey, pubkey.Bytes())
	assert.Equal(t, block.Signature, sig.Bytes())
	assert.True(t, VerifyBlock(block))

	invalidPrivKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivKey.Public().Bytes()
	assert.False(t, VerifyBlock(block))
}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)
	assert.Equal(t, 32, len(hash))
}
