// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light Icechain Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/iceming123/go-ice/ice/fastdownloader"

	"github.com/iceming123/go-ice/accounts"
	"github.com/iceming123/go-ice/common"
	"github.com/iceming123/go-ice/common/hexutil"
	"github.com/iceming123/go-ice/consensus"
	"github.com/iceming123/go-ice/core"
	"github.com/iceming123/go-ice/core/bloombits"
	"github.com/iceming123/go-ice/core/rawdb"
	"github.com/iceming123/go-ice/core/types"
	"github.com/iceming123/go-ice/event"
	"github.com/iceming123/go-ice/ice"
	"github.com/iceming123/go-ice/ice/filters"
	"github.com/iceming123/go-ice/ice/gasprice"
	"github.com/iceming123/go-ice/internal/iceapi"
	"github.com/iceming123/go-ice/light"
	"github.com/iceming123/go-ice/log"
	"github.com/iceming123/go-ice/node"
	"github.com/iceming123/go-ice/p2p"
	"github.com/iceming123/go-ice/p2p/discv5"
	"github.com/iceming123/go-ice/params"
	"github.com/iceming123/go-ice/rpc"
)

type LightICE struct {
	lesCommons

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool

	// Handlers
	peers      *peerSet
	txPool     *light.TxPool
	election   *Election
	blockchain *light.LightChain
	serverPool *serverPool
	reqDist    *requestDistributor
	retriever  *retrieveManager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *iceapi.PublicNetAPI

	genesisHash common.Hash
	wg          sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *ice.Config) (*LightICE, error) {
	chainDb, err := ice.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlockForLes(chainDb)
	if genesisErr != nil {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lice := &LightICE{
		lesCommons: lesCommons{
			chainDb: chainDb,
			config:  config,
			iConfig: light.DefaultClientIndexerConfig,
		},
		genesisHash:    genesisHash,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		peers:          peers,
		reqDist:        newRequestDistributor(peers, quitSync),
		accountManager: ctx.AccountManager,
		engine:         ice.CreateConsensusEngine(ctx, &config.MinervaHash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   ice.NewBloomIndexer(chainDb, params.BloomBitsBlocksClient, params.HelperTrieConfirmations, true),
	}

	lice.relay = NewLesTxRelay(peers, lice.reqDist)
	lice.serverPool = newServerPool(chainDb, quitSync, &lice.wg, nil)
	lice.retriever = newRetrieveManager(peers, lice.reqDist, lice.serverPool)

	lice.odr = NewLesOdr(chainDb, light.DefaultClientIndexerConfig, lice.retriever)
	lice.chtIndexer = light.NewChtIndexer(chainDb, lice.odr, params.CHTFrequencyClient, params.HelperTrieConfirmations)
	lice.bloomTrieIndexer = light.NewBloomTrieIndexer(chainDb, lice.odr, params.BloomBitsBlocksClient, params.BloomTrieFrequency)
	lice.odr.SetIndexers(lice.chtIndexer, lice.bloomTrieIndexer, lice.bloomIndexer)

	// Note: NewLightChain adds the trusted checkpoint so it needs an ODR with
	// indexers already set but not started yet
	// TODO make the params.MainnetTrustedCheckpoint in the config
	if lice.blockchain, err = light.NewLightChain(lice.odr, lice.chainConfig, lice.engine, nil); err != nil {
		return nil, err
	}

	lice.election = NewLightElection(lice.blockchain)
	lice.engine.SetElection(lice.election)
	// Note: AddChildIndexer starts the update process for the child
	lice.bloomIndexer.AddChildIndexer(lice.bloomTrieIndexer)
	lice.chtIndexer.Start(lice.blockchain)
	lice.bloomIndexer.Start(lice.blockchain)

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lice.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lice.txPool = light.NewTxPool(lice.chainConfig, lice.blockchain, lice.relay)
	if lice.protocolManager, err = NewProtocolManager(lice.chainConfig, light.DefaultClientIndexerConfig, true,
		config.NetworkId, lice.eventMux, lice.engine, lice.peers, lice.blockchain, nil,
		chainDb, lice.odr, lice.relay, lice.serverPool, quitSync, &lice.wg, lice.genesisHash); err != nil {
		return nil, err
	}
	lice.ApiBackend = &LesApiBackend{lice, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lice.ApiBackend.gpo = gasprice.NewOracle(lice.ApiBackend, gpoParams)
	return lice, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Etherbase is the address that mining rewards will be send to
func (s *LightDummyAPI) Etherbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Etherbase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightICE) APIs() []rpc.API {
	return append(iceapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, // {
		//	Namespace: "eth",
		//	Version:   "1.0",
		//	Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
		//	Public:    true,
		//},
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
	//return apis
}

func (s *LightICE) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightICE) BlockChain() *light.LightChain          { return s.blockchain }
func (s *LightICE) TxPool() *light.TxPool                  { return s.txPool }
func (s *LightICE) Engine() consensus.Engine               { return s.engine }
func (s *LightICE) LesVersion() int                        { return int(ClientProtocolVersions[0]) }
func (s *LightICE) Downloader() *fastdownloader.Downloader { return s.protocolManager.downloader }
func (s *LightICE) EventMux() *event.TypeMux               { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightICE) Protocols() []p2p.Protocol {
	return s.makeProtocols(ClientProtocolVersions)
}
func (s *LightICE) GenesisHash() common.Hash {
	return s.genesisHash
}
func GenesisNumber() uint64 {
	return params.LesProtocolGenesisBlock
}

// Start implements node.Service, starting all internal goroutines needed by the
// Icechain protocol implementation.
func (s *LightICE) Start(srvr *p2p.Server) error {
	log.Warn("Light client mode is an experimental feature")
	s.startBloomHandlers(params.BloomBitsBlocksClient)
	s.netRPCService = iceapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.BlockChain().Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	s.election.Start()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Icechain protocol.
func (s *LightICE) Stop() error {
	s.odr.Stop()
	s.bloomIndexer.Close()
	s.chtIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()
	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
