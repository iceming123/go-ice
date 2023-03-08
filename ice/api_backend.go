// Copyright 2015 The go-ethereum Authors
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

package ice

import (
	"context"
	"errors"
	"math/big"

	"github.com/iceming123/go-ice/accounts"
	"github.com/iceming123/go-ice/common"
	"github.com/iceming123/go-ice/common/math"
	"github.com/iceming123/go-ice/core"
	"github.com/iceming123/go-ice/core/bloombits"
	"github.com/iceming123/go-ice/core/rawdb"
	"github.com/iceming123/go-ice/core/state"
	"github.com/iceming123/go-ice/core/types"
	"github.com/iceming123/go-ice/core/vm"
	"github.com/iceming123/go-ice/event"
	"github.com/iceming123/go-ice/ice/downloader"
	"github.com/iceming123/go-ice/ice/gasprice"
	"github.com/iceming123/go-ice/icedb"
	"github.com/iceming123/go-ice/params"
	"github.com/iceming123/go-ice/rpc"
)

// ICEAPIBackend implements ethapi.Backend for full nodes
type ICEAPIBackend struct {
	ice *Icechain
	gpo *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *ICEAPIBackend) ChainConfig() *params.ChainConfig {
	return b.ice.chainConfig
}

// CurrentBlock return the fast chain current Block
func (b *ICEAPIBackend) CurrentBlock() *types.Block {
	return b.ice.blockchain.CurrentBlock()
}

// CurrentSnailBlock return the Snail chain current Block
func (b *ICEAPIBackend) CurrentSnailBlock() *types.SnailBlock {
	return b.ice.snailblockchain.CurrentBlock()
}

// SetHead Set the newest position of Fast Chain, that will reset the fast blockchain comment
func (b *ICEAPIBackend) SetHead(number uint64) {
	b.ice.protocolManager.downloader.Cancel()
	b.ice.blockchain.SetHead(number)
}

// SetSnailHead Set the newest position of snail chain
func (b *ICEAPIBackend) SetSnailHead(number uint64) {
	b.ice.protocolManager.downloader.Cancel()
	b.ice.snailblockchain.SetHead(number)
}

// HeaderByNumber returns Header of fast chain by the number
// rpc.PendingBlockNumber == "pending"; rpc.LatestBlockNumber == "latest" ; rpc.LatestBlockNumber == "earliest"
func (b *ICEAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.ice.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ice.blockchain.CurrentBlock().Header(), nil
	}
	return b.ice.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

// HeaderByHash returns header of fast chain by the hash
func (b *ICEAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.ice.blockchain.GetHeaderByHash(hash), nil
}

// SnailHeaderByNumber returns Header of snail chain by the number
// rpc.PendingBlockNumber == "pending"; rpc.LatestBlockNumber == "latest" ; rpc.LatestBlockNumber == "earliest"
func (b *ICEAPIBackend) SnailHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailHeader, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.ice.miner.PendingSnailBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ice.snailblockchain.CurrentBlock().Header(), nil
	}
	return b.ice.snailblockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

// BlockByNumber returns block of fast chain by the number
func (b *ICEAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Only snailchain has miner, also return current block here for fastchain
	if blockNr == rpc.PendingBlockNumber {
		block := b.ice.blockchain.CurrentBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ice.blockchain.CurrentBlock(), nil
	}
	return b.ice.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

// SnailBlockByNumber returns block of snial chain by the number
func (b *ICEAPIBackend) SnailBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailBlock, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.ice.miner.PendingSnailBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ice.snailblockchain.CurrentBlock(), nil
	}
	return b.ice.snailblockchain.GetBlockByNumber(uint64(blockNr)), nil
}
func (b *ICEAPIBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	if blockNr, ok := blockNrOrHash.Number(); ok {
		return b.StateAndHeaderByNumber(ctx, blockNr)
	}
	if hash, ok := blockNrOrHash.Hash(); ok {
		return b.StateAndHeaderByHash(ctx, hash)
	}
	return nil, nil, errors.New("invalid arguments; neither block nor hash specified")
}

// StateAndHeaderByNumber returns the state of block by the number
func (b *ICEAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		state, _ := b.ice.blockchain.State()
		block := b.ice.blockchain.CurrentBlock()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.ice.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}
func (b *ICEAPIBackend) StateAndHeaderByHash(ctx context.Context, hash common.Hash) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByHash(ctx, hash)
	if err != nil {
		return nil, nil, err
	}
	if header == nil {
		return nil, nil, errors.New("header for hash not found")
	}
	stateDb, err := b.ice.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

// GetBlock returns the block by the block's hash
func (b *ICEAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.ice.blockchain.GetBlockByHash(hash), nil
}

// GetSnailBlock returns the snail block by the block's hash
func (b *ICEAPIBackend) GetSnailBlock(ctx context.Context, hash common.Hash) (*types.SnailBlock, error) {
	return b.ice.snailblockchain.GetBlockByHash(hash), nil
}

// GetFruit returns the fruit by the block's hash
func (b *ICEAPIBackend) GetFruit(ctx context.Context, fastblockHash common.Hash) (*types.SnailBlock, error) {
	return b.ice.snailblockchain.GetFruit(fastblockHash), nil
}

// GetReceipts returns the Receipt details by txhash
func (b *ICEAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.ice.chainDb, hash); number != nil {
		return rawdb.ReadReceipts(b.ice.chainDb, hash, *number), nil
	}
	return nil, nil
}

// GetLogs returns the logs by txhash
func (b *ICEAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	number := rawdb.ReadHeaderNumber(b.ice.chainDb, hash)
	if number == nil {
		return nil, nil
	}
	receipts := rawdb.ReadReceipts(b.ice.chainDb, hash, *number)
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

// GetTd returns the total diffcult with block height by blockhash
func (b *ICEAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.ice.snailblockchain.GetTdByHash(blockHash)
}

// GetEVM returns the EVM
func (b *ICEAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.ice.BlockChain(), nil, nil)
	return vm.NewEVM(context, state, b.ice.chainConfig, vmCfg), vmError, nil
}

// SubscribeRemovedLogsEvent registers a subscription of RemovedLogsEvent in fast blockchain
func (b *ICEAPIBackend) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return b.ice.BlockChain().SubscribeRemovedLogsEvent(ch)
}

// SubscribeChainEvent registers a subscription of chainEvnet in fast blockchain
func (b *ICEAPIBackend) SubscribeChainEvent(ch chan<- types.FastChainEvent) event.Subscription {
	return b.ice.BlockChain().SubscribeChainEvent(ch)
}

// SubscribeChainHeadEvent registers a subscription of chainHeadEvnet in fast blockchain
func (b *ICEAPIBackend) SubscribeChainHeadEvent(ch chan<- types.FastChainHeadEvent) event.Subscription {
	return b.ice.BlockChain().SubscribeChainHeadEvent(ch)
}

// SubscribeChainSideEvent registers a subscription of chainSideEvnet in fast blockchain,deprecated
func (b *ICEAPIBackend) SubscribeChainSideEvent(ch chan<- types.FastChainSideEvent) event.Subscription {
	return b.ice.BlockChain().SubscribeChainSideEvent(ch)
}

// SubscribeLogsEvent registers a subscription of log in fast blockchain
func (b *ICEAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.ice.BlockChain().SubscribeLogsEvent(ch)
}

// GetReward returns the Reward info by number in fastchain
func (b *ICEAPIBackend) GetReward(number int64) *types.BlockReward {
	if number < 0 {
		return b.ice.blockchain.CurrentReward()
	}
	return b.ice.blockchain.GetBlockReward(uint64(number))
}

// GetSnailRewardContent returns the Reward content by number in Snailchain
func (b *ICEAPIBackend) GetSnailRewardContent(snailNumber rpc.BlockNumber) *types.SnailRewardContenet {
	return b.ice.agent.GetSnailRewardContent(uint64(snailNumber))
}

func (b *ICEAPIBackend) GetChainRewardContent(blockNr rpc.BlockNumber) *types.ChainReward {
	sheight := uint64(blockNr)
	return b.ice.blockchain.GetRewardInfos(sheight)
}

// GetStateChangeByFastNumber returns the Committee info by committee number
func (b *ICEAPIBackend) GetStateChangeByFastNumber(fastNumber rpc.BlockNumber) *types.BlockBalance {
	return b.ice.blockchain.GetBalanceInfos(uint64(fastNumber))
}

func (b *ICEAPIBackend) GetBalanceChangeBySnailNumber(snailNumber rpc.BlockNumber) *types.BalanceChangeContent {
	var sBlock = b.ice.SnailBlockChain().GetBlockByNumber(uint64(snailNumber))
	state, _ := b.ice.BlockChain().State()
	var (
		addrWithBalance          = make(map[common.Address]*big.Int)
		committeeAddrWithBalance = make(map[common.Address]*big.Int)
		blockFruits              = sBlock.Body().Fruits
		blockFruitsLen           = big.NewInt(int64(len(blockFruits)))
	)
	if blockFruitsLen.Uint64() == 0 {
		return nil
	}
	//snailBlock miner's award
	var balance = state.GetBalance(sBlock.Coinbase())
	addrWithBalance[sBlock.Coinbase()] = balance

	for _, fruit := range blockFruits {
		if addrWithBalance[fruit.Coinbase()] == nil {
			addrWithBalance[fruit.Coinbase()] = state.GetBalance(fruit.Coinbase())
		}
		var committeeMembers = b.ice.election.GetCommittee(fruit.FastNumber())

		for _, cm := range committeeMembers {
			if committeeAddrWithBalance[cm.Coinbase] == nil {
				committeeAddrWithBalance[cm.Coinbase] = state.GetBalance(cm.Coinbase)
			}
		}
	}
	for addr, balance := range committeeAddrWithBalance {
		if addrWithBalance[addr] == nil {
			addrWithBalance[addr] = balance
		}
	}
	return &types.BalanceChangeContent{addrWithBalance}
}

func (b *ICEAPIBackend) GetCommittee(number rpc.BlockNumber) (map[string]interface{}, error) {
	if number == rpc.LatestBlockNumber {
		return b.ice.election.GetCommitteeById(new(big.Int).SetUint64(b.ice.agent.CommitteeNumber())), nil
	}
	return b.ice.election.GetCommitteeById(big.NewInt(number.Int64())), nil
}

func (b *ICEAPIBackend) GetCurrentCommitteeNumber() *big.Int {
	return b.ice.election.GetCurrentCommitteeNumber()
}

// SendTx returns nil by success to add local txpool
func (b *ICEAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.ice.txPool.AddLocal(signedTx)
}

// GetPoolTransactions returns Transactions by pending state in txpool
func (b *ICEAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.ice.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

// GetPoolTransaction returns Transaction by txHash in txpool
func (b *ICEAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.ice.txPool.Get(hash)
}

// GetPoolNonce returns user nonce by user address in txpool
func (b *ICEAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.ice.txPool.State().GetNonce(addr), nil
}

// Stats returns the count tx in txpool
func (b *ICEAPIBackend) Stats() (pending int, queued int) {
	return b.ice.txPool.Stats()
}

func (b *ICEAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.ice.TxPool().Content()
}

// SubscribeNewTxsEvent returns the subscript event of new tx
func (b *ICEAPIBackend) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return b.ice.TxPool().SubscribeNewTxsEvent(ch)
}

// Downloader returns the fast downloader
func (b *ICEAPIBackend) Downloader() *downloader.Downloader {
	return b.ice.Downloader()
}

// ProtocolVersion returns the version of protocol
func (b *ICEAPIBackend) ProtocolVersion() int {
	return b.ice.EthVersion()
}

// SuggestPrice returns tht suggest gas price
func (b *ICEAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

// ChainDb returns tht database of fastchain
func (b *ICEAPIBackend) ChainDb() icedb.Database {
	return b.ice.ChainDb()
}

// EventMux returns Event locker
func (b *ICEAPIBackend) EventMux() *event.TypeMux {
	return b.ice.EventMux()
}

// AccountManager returns Account Manager
func (b *ICEAPIBackend) AccountManager() *accounts.Manager {
	return b.ice.AccountManager()
}

// SnailPoolContent returns snail pool content
func (b *ICEAPIBackend) SnailPoolContent() []*types.SnailBlock {
	return b.ice.SnailPool().Content()
}

// SnailPoolInspect returns snail pool Inspect
func (b *ICEAPIBackend) SnailPoolInspect() []*types.SnailBlock {
	return b.ice.SnailPool().Inspect()
}

// SnailPoolStats returns snail pool Stats
func (b *ICEAPIBackend) SnailPoolStats() (pending int, unVerified int) {
	return b.ice.SnailPool().Stats()
}

// BloomStatus returns Bloom Status
func (b *ICEAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.ice.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

// ServiceFilter make the Filter for the truechian
func (b *ICEAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ice.bloomRequests)
	}
}
