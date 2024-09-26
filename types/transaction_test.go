package types

import (
	"testing"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/util"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.Public().Address().Bytes()

	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivKey.Public().Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}

	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	sign := SignTransaction(fromPrivKey, tx)
	input.Signature = sign.Bytes()
	assert.True(t, VerifyTransaction(tx))
	// fmt.Printf("%+v\n", tx)

}
