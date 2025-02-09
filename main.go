package main

import (
	"log"

	"github.com/Djiit/pingrequest/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
