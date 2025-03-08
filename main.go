package main

import (
	"log"

	"github.com/Djiit/gong/cmd"
)

var (
	version = "dev"
	// nolint:unused
	commit = "none"
	// nolint:unused
	date = "unknown"
)

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
