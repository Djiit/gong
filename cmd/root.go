package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Djiit/pingrequest/cmd/ping"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	verbose     bool
	dryRun      bool
	githubToken string
	rootCmd     = &cobra.Command{
		Use:     "pingrequest",
		Long:    "pingrequest is a CLI tool to ping reviewers.",
		Example: "pingrequest",
	}
)

func SetVersion(version string) {
	rootCmd.Version = version
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func init() {
	// Initialize cobra
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.ping-request.yaml)")
	rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "GitHub token")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display more verbose output in console output. (default: false)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode. (default: false)")
	viper.BindPFlags(rootCmd.PersistentFlags())

	// Add subcommands
	rootCmd.AddCommand(ping.PingCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ping-request" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ping-request")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("PING_REQUEST")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
