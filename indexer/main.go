package main

import (
	"log"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
