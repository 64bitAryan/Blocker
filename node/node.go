package node

import (
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/64bitAryan/blocker/crypto"
	"github.com/64bitAryan/blocker/proto"
	"github.com/64bitAryan/blocker/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const blockTime = time.Second * 5

type Mempool struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		txx: make(map[string]*proto.Transaction),
	}
}

func (pool *Mempool) Has(tx *proto.Transaction) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()
	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := pool.txx[hash]
	return ok
}

func (pool *Mempool) Add(tx *proto.Transaction) bool {
	if pool.Has(tx) {
		return false
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()
	hash := hex.EncodeToString(types.HashTransaction(tx))
	pool.txx[hash] = tx
	return true
}

func (pool *Mempool) Clear() []*proto.Transaction {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	txx := make([]*proto.Transaction, len(pool.txx))
	it := 0
	for k, v := range pool.txx {
		delete(pool.txx, k)
		txx[it] = v
		it++
	}
	return txx
}

func (pool *Mempool) len() int {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	return len(pool.txx)
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKeys
}

type Node struct {
	ServerConfig
	logger   *zap.SugaredLogger
	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool
	proto.UnimplementedNodeServer
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.Development = true
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()

	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMempool(),
		ServerConfig: cfg,
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	// creating a new grpc server
	var (
		opts       = []grpc.ServerOption{}
		grpcServer = grpc.NewServer(opts...)
	)
	// generated function, regestering new server
	proto.RegisterNodeServer(grpcServer, n)

	n.logger.Infow("node started...", "port:", n.ListenAddr)

	if len(bootstrapNodes) > 0 {
		go n.bootstrapNetwork(bootstrapNodes)
	}

	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	return grpcServer.Serve(ln)
}

func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {

	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)

	return n.getVersion(), nil
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))

	if n.mempool.Add(tx) {
		n.logger.Debugw("received tx", "from", peer.Addr, "hash", hash, "we", n.ListenAddr)
		go func() {
			if err := n.broadcase(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}

	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubKey", n.PrivateKey.Public(), "BlockTime", blockTime)
	ticker := time.NewTicker(blockTime)
	for {
		<-ticker.C
		txx := n.mempool.Clear()
		n.logger.Debugw("time to create a new block", "lenTx", len(txx))
	}
}

func (n *Node) broadcase(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()

	n.peers[c] = v
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}

	n.logger.Debugf("new peer successfully connected", "we", n.ListenAddr, "remote node", v.ListenAddr, "Height", v.Height)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {

		if !n.canConnectWith(addr) {
			continue
		}

		n.logger.Debugw("dialing remote node ", "we", n.ListenAddr, "remote", addr)
		c, v, err := n.dialRemoteNode(addr)
		if err != nil {
			return err
		}
		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) dialRemoteNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "blocker-0.1",
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}

	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}

	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	peers := []string{}
	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}
	return peers
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(c), nil
}
