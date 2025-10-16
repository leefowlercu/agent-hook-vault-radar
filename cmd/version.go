package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - set by main package
var (
	version   string
	buildTime string
	commit    string
)

// SetVersionInfo sets the version information from main package
func SetVersionInfo(v, bt, c string) {
	version = v
	buildTime = bt
	commit = c
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Display version, build time, commit, and Go version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hook-vault-radar version %s\n", version)
		fmt.Printf("  Build Time: %s\n", buildTime)
		fmt.Printf("  Commit:     %s\n", commit)
		fmt.Printf("  Go Version: %s\n", runtime.Version())
	},
}
