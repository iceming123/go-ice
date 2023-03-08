package vm

import (
	"math/big"
	"testing"

	"github.com/iceming123/go-ice/common"
	"github.com/iceming123/go-ice/core/state"
	"github.com/iceming123/go-ice/core/types"
	"github.com/iceming123/go-ice/crypto"
	"github.com/iceming123/go-ice/icedb"
	"github.com/iceming123/go-ice/log"
	"github.com/iceming123/go-ice/params"
)

func TestDeposit(t *testing.T) {

	priKey, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(priKey.PublicKey)
	pub := crypto.FromECDSAPub(&priKey.PublicKey)
	value := big.NewInt(1000)

	statedb, _ := state.New(common.Hash{}, state.NewDatabase(icedb.NewMemDatabase()))
	statedb.GetOrNewStateObject(types.StakingAddress)
	evm := NewEVM(Context{}, statedb, params.TestChainConfig, Config{})

	log.Info("Staking deposit", "address", from.String(), "value", value)
	impawn := NewImpawnImpl()
	impawn.Load(evm.StateDB, types.StakingAddress)

	impawn.InsertSAccount2(1000, 0, from, pub, value, big.NewInt(0), true)
	impawn.Save(evm.StateDB, types.StakingAddress)

	impawn1 := NewImpawnImpl()
	impawn1.Load(evm.StateDB, types.StakingAddress)
}
