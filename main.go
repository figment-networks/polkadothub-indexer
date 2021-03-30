package main

import (
	"github.com/figment-networks/polkadothub-indexer/cli"
)

//go:generate swagger generate spec

func main() {
	cli.Run()
}
