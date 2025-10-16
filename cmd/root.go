package cmd

import (
	"fmt"
	"os"

	"github.com/leefowlercu/agent-hook-vault-radar/internal/config"
	"github.com/leefowlercu/agent-hook-vault-radar/internal/processor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "hook-vault-radar",
	Short: "Hook framework integration for Vault Radar scanning",
	Long: "\nhook-vault-radar is a CLI tool that integrates with AI agent hook frameworks " +
		"to scan content for secrets and sensitive data using HashiCorp Vault Radar.\n\n" +
		"It reads hook data from stdin as JSON and outputs decisions to stdout as JSON. " +
		"Logging is sent to stderr to keep stdout clean for hook framework communication.",
	PersistentPreRunE: runInit,
	RunE:              runHook,
}

func init() {
	rootCmd.Flags().String("framework", "", "Hook framework to use (e.g., 'claude')")
	rootCmd.Flags().String("log-level", "", "Logging level (debug, info, warn, error)")
	rootCmd.Flags().String("log-format", "", "Logging format (json, text)")

	// Mark framework flag as required
	rootCmd.MarkFlagRequired("framework")

	viper.BindPFlag("framework", rootCmd.Flags().Lookup("framework"))
	viper.BindPFlag("logging.level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("logging.format", rootCmd.Flags().Lookup("log-format"))

	// Add version command
	rootCmd.AddCommand(versionCmd)

	// Enable --version flag on root command
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("hook-vault-radar version {{.Version}}\n")
}

func runInit(cmd *cobra.Command, args []string) error {
	err := config.InitConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize configuration; %w", err)
	}

	return nil
}

func runHook(cmd *cobra.Command, args []string) error {
	// Get framework from configuration
	framework := viper.GetString("framework")
	return processor.Process(os.Stdin, os.Stdout, framework)
}

func Execute() error {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	err := rootCmd.Execute()

	if err != nil {
		cmd, _, _ := rootCmd.Find(os.Args[1:])
		if cmd == nil {
			cmd = rootCmd
		}

		fmt.Printf("Error: %v\n", err)
		if !cmd.SilenceUsage {
			fmt.Printf("\n")
			cmd.SetOut(os.Stdout)
			cmd.Usage()
		}

		return err
	}

	return nil
}
