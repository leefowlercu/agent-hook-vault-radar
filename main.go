package main

import (
	"os"

	"github.com/leefowlercu/agent-hook-vault-radar/cmd"
)

// Version information set via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

func main() {
	// Pass version info to cmd package
	cmd.SetVersionInfo(Version, BuildTime, Commit)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
