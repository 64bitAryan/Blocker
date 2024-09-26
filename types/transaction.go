package types

import (
	"crypto/sha256"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignTransaction(pk *crypto.PrivateKeys, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, inputs := range tx.Inputs {
		if len(inputs.Signature) == 0 {
			panic("the transaction has no signature")
		}

		sig := crypto.SignatureFromBytes(inputs.Signature)
		pubKey := crypto.PublicKeyFromBytes(inputs.PublicKey)
		// TODO: we have set the signature to nil,fix this
		inputs.Signature = nil
		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
	}
	return true
}
