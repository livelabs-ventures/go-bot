package main

import (
	"fmt"
	"os"

	"github.com/standup-bot/standup-bot/internal/cli"
)

// Variables set by goreleaser at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Pass version info to CLI
	cli.SetVersionInfo(version, commit, date)
	
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}