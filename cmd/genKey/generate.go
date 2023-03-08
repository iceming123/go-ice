package main

import (
	"encoding/hex"
	"fmt"

	"github.com/iceming123/go-ice/cmd/utils"
	"github.com/iceming123/go-ice/crypto"

	"gopkg.in/urfave/cli.v1"
)

var (
	generateCommand = cli.Command{
		Name:      "generate",
		Usage:     "Generate new key item",
		ArgsUsage: "",
		Description: `
Generate a new key item.
`,
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "sum",
				Usage: "key info count",
				Value: 1,
			},
		},
		Action: func(ctx *cli.Context) error {
			count := ctx.Int("sum")
			if count <= 0 || count > 100 {
				count = 100
			}
			makeAddress(count)

			return nil
		},
	}
)

func makeAddress(count int) {
	for i := 0; i < count; i++ {
		if privateKey, err := crypto.GenerateKey(); err != nil {
			utils.Fatalf("Error GenerateKey: %v", err)
		} else {
			fmt.Println("private key:", hex.EncodeToString(crypto.FromECDSA(privateKey)))
			fmt.Println("public key:", hex.EncodeToString(crypto.FromECDSAPub(&privateKey.PublicKey)))
			addr := crypto.PubkeyToAddress(privateKey.PublicKey)
			//fmt.Println("address-CV:", addr.String())
			//fmt.Println("address-0x:", addr.StringToVC())
			fmt.Println("address-0x: ", addr.String())
			fmt.Println("-------------------------------------------------------")
		}
	}
}
