package types

import (
	"github.com/iceming123/go-ice/consensus/tbft/crypto/cryptoamino"
	"github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	RegisterBlockAmino(cdc)
}

//RegisterBlockAmino is register for block amino
func RegisterBlockAmino(cdc *amino.Codec) {
	cryptoAmino.RegisterAmino(cdc)
}
