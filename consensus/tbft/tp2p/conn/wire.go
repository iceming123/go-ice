package conn

import (
	cryptoAmino "github.com/iceming123/go-ice/consensus/tbft/crypto/cryptoamino"
	amino "github.com/tendermint/go-amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	RegisterPacket(cdc)
}
