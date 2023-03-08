package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/iceming123/go-ice/cmd/utils"

	"gopkg.in/urfave/cli.v1"
)

var (
	gitCommit = ""
	gitData   = ""
	app       *cli.App
)

func init() {
	app = utils.NewApp(gitCommit, gitData, "an common generate and convert address tool")
	app.Commands = []cli.Command{
		generateCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
