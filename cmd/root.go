package cmd

import (
	"fmt"
	"os"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	logger  *zap.Logger
	dryRun  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mls",
	Short: "A high-performance storage cleanup tool for macOS",
	Long: `MacosLeanStorage (mls) is a tool designed to safely and efficiently clean up
large cache and temporary files on macOS.

It focuses on performance, safety (dry-run by default), and multi-profile support.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger, initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.MacosLeanStorage.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", true, "Enable dry-run mode (no files deleted)")

	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
}

func initLogger() {
	var err error
	if verbose, _ := rootCmd.Flags().GetBool("verbose"); verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		defaultPath, err := config.GetDefaultConfigPath()
		cobra.CheckErr(err)

		// Create default config if it doesn't exist
		err = config.CreateDefaultConfig(defaultPath)
		cobra.CheckErr(err)

		viper.SetConfigFile(defaultPath)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("Failed to read config file", zap.Error(err))
	} else {
		logger.Debug("Using config file", zap.String("path", viper.ConfigFileUsed()))
	}
}
