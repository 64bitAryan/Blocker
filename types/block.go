package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/cbergoon/merkletree"
	pb "google.golang.org/protobuf/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{
		hash: hash,
	}
}

func (h TxHash) CalculateHash() ([]byte, error) {
	return h.hash, nil
}

func (h TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(h.hash, other.(TxHash).hash)
	return equals, nil
}

func SignBlock(pk *crypto.PrivateKeys, b *proto.Block) *crypto.Signature {
	hash := HashBlock(b)
	sig := pk.Sign(hash)
	b.PublicKey = pk.Public().Bytes()
	b.Signature = sig.Bytes()

	if len(b.Transactions) > 0 {

		tree, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}

		b.Header.RootHash = tree.MerkleRoot()
	}

	return sig
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.Transactions) > 0 {
		if !VerifyRootHash(b) {
			return false
		}
	}

	if len(b.PublicKey) != crypto.PubKeyLen {
		return false
	}
	if len(b.Signature) != crypto.SignatureLen {
		return false
	}
	sig := crypto.SignatureFromBytes(b.Signature)
	pubKey := crypto.PublicKeyFromBytes(b.PublicKey)
	hash := HashBlock(b)
	return sig.Verify(pubKey, hash)
}

func VerifyRootHash(b *proto.Block) bool {
	merkletree, err := GetMerkleTree(b)
	if err != nil {
		return false
	}

	valid, err := merkletree.VerifyTree()
	if err != nil {
		return false
	}

	if !valid {
		return false
	}

	return bytes.Equal(b.Header.RootHash, merkletree.MerkleRoot())
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	nTransaction := len(b.Transactions)
	list := make([]merkletree.Content, nTransaction)
	for i := 0; i < nTransaction; i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}

	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	b.Header.RootHash = t.MerkleRoot()
	return t, nil
}

// HashBlock returns SHA256 of the header
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	// hash block
	hash := sha256.Sum256(b)
	return hash[:]
}
