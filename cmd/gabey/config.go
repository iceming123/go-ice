package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"unicode"

	"gopkg.in/urfave/cli.v1"

	"github.com/iceming123/go-ice/cmd/utils"
	"github.com/iceming123/go-ice/crypto"
	"github.com/iceming123/go-ice/ice"
	"github.com/iceming123/go-ice/node"
	"github.com/iceming123/go-ice/params"
	"github.com/naoina/toml"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...)),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type IcestatsConfig struct {
	URL string `toml:",omitempty"`
}

type gethConfig struct {
	Ice      ice.Config
	Node     node.Config
	Icestats IcestatsConfig
}

func loadConfig(file string, cfg *gethConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "ice", "eth", "impawn", "shh")
	cfg.WSModules = append(cfg.WSModules, "ice")
	cfg.IPCPath = "gice.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, gethConfig) {
	// Load defaults.
	cfg := gethConfig{
		Ice:  ice.DefaultConfig,
		Node: defaultNodeConfig(),
	}
	if ctx.GlobalBool(utils.SingleNodeFlag.Name) {
		// set iceconfig
		prikey, _ := crypto.HexToECDSA("229ca04fb83ec698296037c7d2b04a731905df53b96c260555cbeed9e4c64036")
		cfg.Ice.PrivateKey = prikey
		cfg.Ice.CommitteeKey = crypto.FromECDSA(prikey)

		//cfg.Ice.MineFruit = true
		cfg.Ice.Mine = true
		cfg.Ice.Etherbase = crypto.PubkeyToAddress(prikey.PublicKey)
		//cfg.Ice.NetworkId =400
		//set node config
		cfg.Node.HTTPPort = 8888
		cfg.Node.HTTPHost = "127.0.0.1"
		cfg.Node.HTTPModules = []string{"db", "ice", "net", "web3", "personal", "admin", "miner", "eth"}

		ctx.GlobalSet("datadir", "./data")
	}
	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetIcechainConfig(ctx, stack, &cfg.Ice)
	if ctx.GlobalIsSet(utils.IcestatsURLFlag.Name) {
		cfg.Icestats.URL = ctx.GlobalString(utils.IcestatsURLFlag.Name)
	}

	return stack, cfg
}

func makeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)

	utils.RegisterIceService(stack, &cfg.Ice)

	// Add the Icechain Stats daemon if requested.
	if cfg.Icestats.URL != "" {
		utils.RegisterIceStatsService(stack, cfg.Icestats.URL)
	}
	return stack
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.Ice.Genesis != nil {
		cfg.Ice.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
