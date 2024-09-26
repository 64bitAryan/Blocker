package main

import (
	"context"
	"log"
	"time"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/node"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/util"
	"google.golang.org/grpc"
)

/*
____steps to create grpc server with protobuf____

	1. creating a net listener
	2. creating a grpc server
	3. regestering the new server
	4. serve the service created
*/

func main() {

	makeNode(":3000", []string{}, true)
	time.Sleep(time.Second)
    time.Sleep(time.Second)
	makeNode(":4000", []string{":3000"}, false)
	time.Sleep(time.Second)
	makeNode(":5000", []string{":4000"}, false)
	for {
		time.Sleep(time.Millisecond * 800)
		makeTaransaction()
	}
}

func makeNode(listenAddr string, bootstrapNodes []string, isValidator bool) *node.Node {
	cfg := node.ServerConfig{
		Version:    "blocker-1",
		ListenAddr: listenAddr,
	}
	if isValidator {
		cfg.PrivateKey = crypto.GeneratePrivateKey()
	}
	n := node.NewNode(cfg)
	go n.Start(listenAddr, bootstrapNodes)
	return n

}

func makeTaransaction() {
	// creating a grpc dialer
	client, err := grpc.Dial(":3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	//
	c := proto.NewNodeClient(client)

	privKey := crypto.GeneratePrivateKey()

	tx := &proto.Transaction{
		Version: 1,
		Inputs: []*proto.TxInput{
			{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		},
		Outputs: []*proto.TxOutput{
			{
				Amount:  99,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}
	_, err = c.HandleTransaction(context.TODO(), tx)
	if err != nil {
		log.Fatal(err)
	}
}
