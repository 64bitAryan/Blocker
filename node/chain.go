package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/types"
)

const godSeed = "54967bdaf7dacbf0adf004ad2ddb1196073239bb0b83bf587c21edf503a3a90e"

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: []*proto.Header{},
	}
}

func (list *HeaderList) Add(h *proto.Header) {
	list.headers = append(list.headers, h)
}

func (list *HeaderList) Get(index int) *proto.Header {
	if index > list.Height() {
		panic("index too heigh")
	}
	return list.headers[index]
}

func (list *HeaderList) Height() int {

	return list.Len() - 1
}

func (list *HeaderList) Len() int {
	return len(list.headers)
}

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	blockstore BlockStorer
	utxoStore  UTXOStorer
	headers    *HeaderList
}

func NewChain(bs BlockStorer, txStore TXStorer) *Chain {
	chain := &Chain{
		blockstore: bs,
		txStore:    txStore,
		utxoStore:  NewMemoryUTXOStore(),
		headers:    NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())
	return chain

}
func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)

	for _, tx := range b.Transactions {

		if err := c.txStore.Put(tx); err != nil {
			return err
		}
		hash := hex.EncodeToString(types.HashTransaction(tx))

		for it, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: it,
				Amount:   output.Amount,
				Spent:    false,
			}
			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}
	}

	return c.blockstore.Put(b)
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockstore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - height (%d)", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	if !types.VerifyBlock(b) {
		return fmt.Errorf("invalid block signature")
	}

	currBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := types.HashBlock(currBlock)
	if !bytes.Equal(hash, b.Header.PrevHash) {
		return fmt.Errorf("invalid previous block hash")
	}

	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	// Verify the signature
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("invalid tx signature")
	}
	// Check if all the inputs are unspent
	var (
		hash    = hex.EncodeToString(types.HashTransaction(tx))
		nInputs = len(tx.Inputs)
	)
	sumInput := 0
	for i := 0; i < nInputs; i++ {
		prevHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
		key := fmt.Sprintf("%s-%d", prevHash, i)
		utxo, err := c.utxoStore.Get(key)
		sumInput += int(utxo.Amount)
		if err != nil {
			return err
		} else if utxo.Spent {
			return fmt.Errorf("input %d of tx %s is already spent", i, hash)
		}
	}

	sumOutput := 0
	for _, output := range tx.Outputs {
		sumOutput += int(output.Amount)
	}

	if sumInput < sumOutput {
		return fmt.Errorf("invalid tx: insufficient balance :: input sum (%d), output sum (%d)", sumInput, sumOutput)
	}

	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromSeedStr(godSeed)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	return block
}
