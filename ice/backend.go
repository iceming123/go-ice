// Copyright 2014 The go-ethereum Authors
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

// Package ice implements the Icechain protocol.
package ice

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/iceming123/go-ice/consensus/tbft"
	config "github.com/iceming123/go-ice/params"

	"github.com/iceming123/go-ice/accounts"
	"github.com/iceming123/go-ice/common"
	"github.com/iceming123/go-ice/common/hexutil"
	"github.com/iceming123/go-ice/consensus"
	elect "github.com/iceming123/go-ice/consensus/election"
	ethash "github.com/iceming123/go-ice/consensus/minerva"
	"github.com/iceming123/go-ice/core"
	"github.com/iceming123/go-ice/core/bloombits"
	chain "github.com/iceming123/go-ice/core/snailchain"
	"github.com/iceming123/go-ice/core/snailchain/rawdb"
	"github.com/iceming123/go-ice/core/types"
	"github.com/iceming123/go-ice/core/vm"
	"github.com/iceming123/go-ice/crypto"
	"github.com/iceming123/go-ice/event"
	"github.com/iceming123/go-ice/ice/downloader"
	"github.com/iceming123/go-ice/ice/filters"
	"github.com/iceming123/go-ice/ice/gasprice"
	"github.com/iceming123/go-ice/icedb"
	"github.com/iceming123/go-ice/internal/iceapi"
	"github.com/iceming123/go-ice/log"
	"github.com/iceming123/go-ice/miner"
	"github.com/iceming123/go-ice/node"
	"github.com/iceming123/go-ice/p2p"
	"github.com/iceming123/go-ice/params"
	"github.com/iceming123/go-ice/rlp"
	"github.com/iceming123/go-ice/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Icechain implements the Icechain full node service.
type Icechain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the Icechain

	// Handlers
	txPool *core.TxPool

	snailPool *chain.SnailPool

	agent    *PbftAgent
	election *elect.Election

	blockchain      *core.BlockChain
	snailblockchain *chain.SnailBlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb icedb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend *ICEAPIBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	etherbase common.Address

	networkID     uint64
	netRPCService *iceapi.PublicNetAPI

	pbftServer *tbft.Node

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
}

func (s *Icechain) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Icechain object (including the
// initialisation of the common Icechain object)
func New(ctx *node.ServiceContext, config *Config) (*Icechain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run ice.Icechain in light sync mode, use les.LightIcechain")
	}
	//if config.SyncMode == downloader.SnapShotSync {
	//	return nil, errors.New("can't run ice.Icechain in SnapShotSync sync mode, use les.LightIcechain")
	//}

	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	//chainDb, err := CreateDB(ctx, config, path)
	if err != nil {
		return nil, err
	}

	chainConfig, genesisHash, _, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	log.Info("Initialised chain configuration", "config", chainConfig)

	/*if config.Genesis != nil {
		config.MinerGasFloor = config.Genesis.GasLimit * 9 / 10
		config.MinerGasCeil = config.Genesis.GasLimit * 11 / 10
	}*/

	ice := &Icechain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.MinervaHash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		networkID:      config.NetworkId,
		gasPrice:       config.GasPrice,
		etherbase:      config.Etherbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks, params.BloomConfirms, false),
	}

	log.Info("Initialising Icechain protocol", "versions", ProtocolVersions, "network", config.NetworkId, "syncmode", config.SyncMode)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gice upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Deleted: config.DeletedState, Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)

	ice.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, ice.chainConfig, ice.engine, vmConfig)
	if err != nil {
		return nil, err
	}

	ice.snailblockchain, err = chain.NewSnailBlockChain(chainDb, ice.chainConfig, ice.engine, ice.blockchain)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ice.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	//  rewind snail if case of incompatible config
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding snail chain to upgrade configuration", "err", compat)
		ice.snailblockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	ice.bloomIndexer.Start(ice.blockchain)

	consensus.InitTIP8(chainConfig, ice.snailblockchain)
	//sv := chain.NewBlockValidator(ice.chainConfig, ice.blockchain, ice.snailblockchain, ice.engine)
	//ice.snailblockchain.SetValidator(sv)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	if config.SnailPool.Journal != "" {
		config.SnailPool.Journal = ctx.ResolvePath(config.SnailPool.Journal)
	}

	ice.txPool = core.NewTxPool(config.TxPool, ice.chainConfig, ice.blockchain)

	//ice.snailPool = chain.NewSnailPool(config.SnailPool, ice.blockchain, ice.snailblockchain, ice.engine, sv)
	ice.snailPool = chain.NewSnailPool(config.SnailPool, ice.blockchain, ice.snailblockchain, ice.engine)

	ice.election = elect.NewElection(ice.chainConfig, ice.blockchain, ice.snailblockchain, ice.config)

	//ice.snailblockchain.Validator().SetElection(ice.election, ice.blockchain)

	ice.engine.SetElection(ice.election)
	ice.engine.SetSnailChainReader(ice.snailblockchain)
	ice.election.SetEngine(ice.engine)

	//coinbase, _ := ice.Etherbase()
	ice.agent = NewPbftAgent(ice, ice.chainConfig, ice.engine, ice.election, config.MinerGasFloor, config.MinerGasCeil)
	if ice.protocolManager, err = NewProtocolManager(
		ice.chainConfig, config.SyncMode, config.NetworkId,
		ice.eventMux, ice.txPool, ice.snailPool, ice.engine,
		ice.blockchain, ice.snailblockchain,
		chainDb, ice.agent); err != nil {
		return nil, err
	}

	ice.miner = miner.New(ice, ice.chainConfig, ice.EventMux(), ice.engine, ice.election, ice.Config().MineFruit, ice.Config().NodeType, ice.Config().RemoteMine, ice.Config().Mine)
	ice.miner.SetExtra(makeExtraData(config.ExtraData))

	committeeKey, err := crypto.ToECDSA(ice.config.CommitteeKey)
	if err == nil {
		ice.miner.SetElection(ice.config.EnableElection, crypto.FromECDSAPub(&committeeKey.PublicKey))
	}

	ice.APIBackend = &ICEAPIBackend{ice, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	ice.APIBackend.gpo = gasprice.NewOracle(ice.APIBackend, gpoParams)
	return ice, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gice",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (icedb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*icedb.LDBDatabase); ok {
		db.Meter("ice/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Icechain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *ethash.Config, chainConfig *params.ChainConfig,
	db icedb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	// snail chain not need clique
	/*
		if chainConfig.Clique != nil {
			return clique.New(chainConfig.Clique, db)
		}*/
	// Otherwise assume proof-of-work
	switch config.PowMode {
	case ethash.ModeFake:
		log.Info("-----Fake mode")
		log.Warn("Ethash used in fake mode")
		return ethash.NewFaker()
	case ethash.ModeTest:
		log.Warn("Ethash used in test mode")
		return ethash.NewTester()
	case ethash.ModeShared:
		log.Warn("Ethash used in shared mode")
		return ethash.NewShared()
	default:
		engine := ethash.New(ethash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs return the collection of RPC services the ice package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Icechain) APIs() []rpc.API {
	apis := iceapi.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append ice	APIs and  Eth APIs
	namespaces := []string{"ice", "eth"}
	for _, name := range namespaces {
		apis = append(apis, []rpc.API{
			{
				Namespace: name,
				Version:   "1.0",
				Service:   NewPublicIcechainAPI(s),
				Public:    true,
			}, {
				Namespace: name,
				Version:   "1.0",
				Service:   NewPublicMinerAPI(s),
				Public:    true,
			}, {
				Namespace: name,
				Version:   "1.0",
				Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
				Public:    true,
			}, {
				Namespace: name,
				Version:   "1.0",
				Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
				Public:    true,
			},
		}...)
	}
	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Icechain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Icechain) ResetWithFastGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Icechain) Etherbase() (eb common.Address, err error) {
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()

	if etherbase != (common.Address{}) {
		return etherbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			etherbase := accounts[0].Address

			s.lock.Lock()
			s.etherbase = etherbase
			s.lock.Unlock()

			log.Info("Coinbase automatically configured", "address", etherbase)
			return etherbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("coinbase must be explicitly specified")
}

// SetEtherbase sets the mining reward address.
func (s *Icechain) SetEtherbase(etherbase common.Address) {
	s.lock.Lock()
	s.etherbase = etherbase
	s.agent.committeeNode.Coinbase = etherbase
	s.lock.Unlock()

	s.miner.SetEtherbase(etherbase)
}

func (s *Icechain) StartMining(local bool) error {
	eb, err := s.Etherbase()
	if err != nil {
		log.Error("Cannot start mining without coinbase", "err", err)
		return fmt.Errorf("coinbase missing: %v", err)
	}

	// snail chain not need clique
	/*
		if clique, ok := s.engine.(*clique.Clique); ok {
			wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
			if wallet == nil || err != nil {
				log.Error("Etherbase account unavailable locally", "err", err)
				return fmt.Errorf("signer missing: %v", err)
			}
			clique.Authorize(eb, wallet.SignHash)
		}*/

	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so none will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptFruits, 1)

	}
	go s.miner.Start(eb)
	return nil
}

func (s *Icechain) StopMining()                       { s.miner.Stop() }
func (s *Icechain) IsMining() bool                    { return s.miner.Mining() }
func (s *Icechain) Miner() *miner.Miner               { return s.miner }
func (s *Icechain) PbftAgent() *PbftAgent             { return s.agent }
func (s *Icechain) AccountManager() *accounts.Manager { return s.accountManager }
func (s *Icechain) BlockChain() *core.BlockChain      { return s.blockchain }
func (s *Icechain) Config() *Config                   { return s.config }

func (s *Icechain) SnailBlockChain() *chain.SnailBlockChain { return s.snailblockchain }
func (s *Icechain) TxPool() *core.TxPool                    { return s.txPool }

func (s *Icechain) SnailPool() *chain.SnailPool { return s.snailPool }

func (s *Icechain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Icechain) Engine() consensus.Engine           { return s.engine }
func (s *Icechain) ChainDb() icedb.Database            { return s.chainDb }
func (s *Icechain) IsListening() bool                  { return true } // Always listening
func (s *Icechain) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Icechain) NetVersion() uint64                 { return s.networkID }
func (s *Icechain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *Icechain) Synced() bool                       { return atomic.LoadUint32(&s.protocolManager.acceptTxs) == 1 }
func (s *Icechain) ArchiveMode() bool                  { return s.config.NoPruning }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Icechain) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Icechain protocol implementation.
func (s *Icechain) Start(srvr *p2p.Server) error {

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = iceapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	s.startPbftServer()
	if s.pbftServer == nil {
		log.Error("start pbft server failed.")
		return errors.New("start pbft server failed.")
	}
	s.agent.server = s.pbftServer
	log.Info("", "server", s.agent.server)
	s.agent.Start()

	s.election.Start()

	//start fruit journal
	s.snailPool.Start()

	// Start the networking layer and the light server if requested
	s.protocolManager.Start2(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}

	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Icechain protocol.
func (s *Icechain) Stop() error {
	s.stopPbftServer()
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.snailblockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.snailPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

func (s *Icechain) startPbftServer() error {
	priv, err := crypto.ToECDSA(s.config.CommitteeKey)
	if err != nil {
		return err
	}

	cfg := config.DefaultConfig()
	cfg.P2P.ListenAddress1 = "tcp://0.0.0.0:" + strconv.Itoa(s.config.Port)
	cfg.P2P.ListenAddress2 = "tcp://0.0.0.0:" + strconv.Itoa(s.config.StandbyPort)

	n1, err := tbft.NewNode(cfg, "1", priv, s.agent)
	if err != nil {
		return err
	}
	s.pbftServer = n1
	return n1.Start()
}

func (s *Icechain) stopPbftServer() error {
	if s.pbftServer != nil {
		s.pbftServer.Stop()
	}
	return nil
}
