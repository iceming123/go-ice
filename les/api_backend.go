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

package les

import (
	"context"
	"errors"
	"math/big"

	"github.com/iceming123/go-ice/ice/downloader"
	"github.com/iceming123/go-ice/ice/fastdownloader"
	"github.com/iceming123/go-ice/light"

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
	"github.com/iceming123/go-ice/ice/gasprice"
	"github.com/iceming123/go-ice/icedb"
	"github.com/iceming123/go-ice/params"
	"github.com/iceming123/go-ice/rpc"
)

type LesApiBackend struct {
	ice *LightICE
	gpo *gasprice.Oracle
}

var (
	NotSupportOnLes = errors.New("not support on les protocol")
)

//////////////////////////////////////////////////////////////
func (b *LesApiBackend) SetSnailHead(number uint64) {

}
func (b *LesApiBackend) SnailHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailHeader, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) SnailBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetFruit(ctx context.Context, fastblockHash common.Hash) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *types.Header, error) {
	return nil, nil, NotSupportOnLes
}
func (b *LesApiBackend) StateAndHeaderByHash(ctx context.Context, hash common.Hash) (*state.StateDB, *types.Header, error) {
	return nil, nil, NotSupportOnLes
}
func (b *LesApiBackend) GetSnailBlock(ctx context.Context, blockHash common.Hash) (*types.SnailBlock, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetReward(number int64) *types.BlockReward {
	return nil
}
func (b *LesApiBackend) GetCommittee(id rpc.BlockNumber) (map[string]interface{}, error) {
	return nil, NotSupportOnLes
}
func (b *LesApiBackend) GetCurrentCommitteeNumber() *big.Int {
	return nil
}
func (b *LesApiBackend) GetStateChangeByFastNumber(fastNumber rpc.BlockNumber) *types.BlockBalance {
	return nil
}
func (b *LesApiBackend) GetBalanceChangeBySnailNumber(snailNumber rpc.BlockNumber) *types.BalanceChangeContent {
	return nil
}
func (b *LesApiBackend) GetSnailRewardContent(blockNr rpc.BlockNumber) *types.SnailRewardContenet {
	return nil
}
func (b *LesApiBackend) GetChainRewardContent(blockNr rpc.BlockNumber) *types.ChainReward {
	return nil
}
func (b *LesApiBackend) CurrentSnailBlock() *types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolContent() []*types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolInspect() []*types.SnailBlock {
	return nil
}
func (b *LesApiBackend) SnailPoolStats() (pending int, unVerified int) {
	return 0, 0
}
func (b *LesApiBackend) Downloader() *downloader.Downloader {
	return nil
}

//////////////////////////////////////////////////////////////
func (b *LesApiBackend) ChainConfig() *params.ChainConfig {
	return b.ice.chainConfig
}

func (b *LesApiBackend) CurrentBlock() *types.Block {
	return types.NewBlockWithHeader(b.ice.blockchain.CurrentHeader())
}

func (b *LesApiBackend) SetHead(number uint64) {
	b.ice.protocolManager.downloader.Cancel()
	b.ice.blockchain.SetHead(number)
}

func (b *LesApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber || blockNr == rpc.PendingBlockNumber {
		return b.ice.blockchain.CurrentHeader(), nil
	}

	return b.ice.blockchain.GetHeaderByNumberOdr(ctx, uint64(blockNr))
}
func (b *LesApiBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.ice.blockchain.GetHeaderByHash(hash), nil
}

func (b *LesApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, err
	}
	return b.GetBlock(ctx, header.Hash())
}

func (b *LesApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	return light.NewState(ctx, header, b.ice.odr), header, nil
}

func (b *LesApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.ice.blockchain.GetBlockByHash(ctx, blockHash)
}

func (b *LesApiBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	if number := rawdb.ReadHeaderNumber(b.ice.chainDb, hash); number != nil {
		return light.GetBlockReceipts(ctx, b.ice.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	if number := rawdb.ReadHeaderNumber(b.ice.chainDb, hash); number != nil {
		return light.GetBlockLogs(ctx, b.ice.odr, hash, *number)
	}
	return nil, nil
}

func (b *LesApiBackend) GetTd(hash common.Hash) *big.Int {
	return big.NewInt(0)
	//return b.ice.blockchain.GetTdByHash(hash)
}

func (b *LesApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	context := core.NewEVMContext(msg, header, b.ice.blockchain, nil, nil)
	return vm.NewEVM(context, state, b.ice.chainConfig, vmCfg), state.Error, nil
}

func (b *LesApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.ice.txPool.Add(ctx, signedTx)
}

func (b *LesApiBackend) RemoveTx(txHash common.Hash) {
	b.ice.txPool.RemoveTx(txHash)
}

func (b *LesApiBackend) GetPoolTransactions() (types.Transactions, error) {
	return b.ice.txPool.GetTransactions()
}

func (b *LesApiBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return b.ice.txPool.GetTransaction(txHash)
}

func (b *LesApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.ice.txPool.GetNonce(ctx, addr)
}

func (b *LesApiBackend) Stats() (pending int, queued int) {
	return b.ice.txPool.Stats(), 0
}

func (b *LesApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.ice.txPool.Content()
}

func (b *LesApiBackend) SubscribeNewTxsEvent(ch chan<- types.NewTxsEvent) event.Subscription {
	return b.ice.txPool.SubscribeNewTxsEvent(ch)
}

func (b *LesApiBackend) SubscribeChainEvent(ch chan<- types.FastChainEvent) event.Subscription {
	return b.ice.blockchain.SubscribeChainEvent(ch)
}

func (b *LesApiBackend) SubscribeChainHeadEvent(ch chan<- types.FastChainHeadEvent) event.Subscription {
	return b.ice.blockchain.SubscribeChainHeadEvent(ch)
}

func (b *LesApiBackend) SubscribeChainSideEvent(ch chan<- types.FastChainSideEvent) event.Subscription {
	return b.ice.blockchain.SubscribeChainSideEvent(ch)
}

func (b *LesApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.ice.blockchain.SubscribeLogsEvent(ch)
}

func (b *LesApiBackend) SubscribeRemovedLogsEvent(ch chan<- types.RemovedLogsEvent) event.Subscription {
	return b.ice.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (b *LesApiBackend) FastDownloader() *fastdownloader.Downloader {
	return b.ice.Downloader()
}

func (b *LesApiBackend) ProtocolVersion() int {
	return b.ice.LesVersion() + 10000
}

func (b *LesApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LesApiBackend) ChainDb() icedb.Database {
	return b.ice.chainDb
}

func (b *LesApiBackend) EventMux() *event.TypeMux {
	return b.ice.eventMux
}

func (b *LesApiBackend) AccountManager() *accounts.Manager {
	return b.ice.accountManager
}

func (b *LesApiBackend) BloomStatus() (uint64, uint64) {
	if b.ice.bloomIndexer == nil {
		return 0, 0
	}
	sections, _, _ := b.ice.bloomIndexer.Sections()
	return params.BloomBitsBlocksClient, sections
}

func (b *LesApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ice.bloomRequests)
	}
}
