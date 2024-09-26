package node

import (
	"testing"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/types"
	"github.com/64bitAryan/blocker/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	b := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)
	b.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, b)
	return b
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	require.Equal(t, chain.Height(), 0)
	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	for i := 0; i < 100; i++ {
		b := randomBlock(t, chain)
		require.Nil(t, chain.AddBlock(b))
		require.Equal(t, chain.Height(), i+1)
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	for i := 0; i < 100; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)

		require.Nil(t, chain.AddBlock(block))
		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}

}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("0ff6af1c9aa971ef969f2cf72b2cfb8bf21d52c80c131d959de4d8edc6687e21")
	assert.Nil(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(prevTx),
			PrevOutIndex: 0,
			PublicKey:    privKey.Public().Bytes(),
		},
	}
	outpus := []*proto.TxOutput{
		{
			Amount:  100,
			Address: recipient,
		},
		{
			Amount:  900,
			Address: privKey.Public().Address().Bytes(),
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outpus,
	}

	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)
	require.Nil(t, chain.AddBlock(block))
}

func TestAddBlocksWithTxInsuffucientFunds(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromSeedStr(godSeed)
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("0ff6af1c9aa971ef969f2cf72b2cfb8bf21d52c80c131d959de4d8edc6687e21")
	assert.Nil(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(prevTx),
			PrevOutIndex: 0,
			PublicKey:    privKey.Public().Bytes(),
		},
	}
	outpus := []*proto.TxOutput{
		{
			Amount:  1001,
			Address: recipient,
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outpus,
	}

	sig := types.SignTransaction(privKey, tx)

	block.Transactions = append(block.Transactions, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	require.NotNil(t, chain.AddBlock(block))
}
