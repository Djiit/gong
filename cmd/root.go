package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/Djiit/gong/cmd/ping"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	logLevel    string
	dryRun      bool
	githubToken string
	rootCmd     = &cobra.Command{
		Use:     "gong",
		Long:    "gong is a CLI tool to ping reviewers.",
		Example: "gong",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			levelMap := map[string]zerolog.Level{
				"panic": 5,
				"fatal": 4,
				"error": 3,
				"warn":  2,
				"info":  1,
				"debug": 0,
				"trace": -1,
			}
			zerolog.SetGlobalLevel(levelMap[logLevel])

			log.Debug().Msg("Using config file: " + viper.ConfigFileUsed())
			log.Trace().Msgf("Config: %+v", viper.AllSettings())
		},
	}
)

func SetVersion(version string) {
	rootCmd.Version = version
}

func Execute() error {
	return rootCmd.ExecuteContext(context.Background())
}

func init() {
	// Init logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is $HOME/.gong.yaml)")
	rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "GitHub token")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level. (default: info)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run in dry-run mode. (default: false)")
	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		log.Fatal().Msgf("Error binding flags: %v", err)
	}

	// Initialize cobra
	cobra.OnInitialize(initConfig)

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

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gong")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GONG")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		log.Warn().Msg("No config file found or error reading config: " + err.Error())
	}
}
