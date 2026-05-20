package cmd

import (
	"fmt"
	"os"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// cfgFile stores the path to the configuration file provided via flags.
	cfgFile string
	// logger is the global structured logger for the application.
	logger *zap.Logger
	// dryRun indicates whether the application should perform a live cleanup or just simulate it.
	dryRun bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mls",
	Short: "A high-performance storage cleanup tool for macOS",
	Long: `MrLeanStorage (mls) is a tool designed to safely and efficiently clean up
large cache and temporary files on macOS.

It focuses on performance, safety (dry-run by default), and multi-profile support.`,
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init initializes the root command's flags and configures the initial setup.
func init() {
	cobra.OnInitialize(initLogger, initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.MrLeanStorage.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", true, "Enable dry-run mode (no files deleted)")

	viper.BindPFlag("dry_run", rootCmd.PersistentFlags().Lookup("dry-run"))
}

// initLogger initializes the zap logger based on the verbose flag.
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

// initConfig reads in the config file and ENV variables if set.
// If the config file does not exist, it creates a default one.
func initConfig() {
	if cfgFile == "" {
		defaultPath, err := config.GetDefaultConfigPath()
		cobra.CheckErr(err)
		cfgFile = defaultPath
	}

	// Create default config if it doesn't exist
	err := config.CreateDefaultConfig(cfgFile)
	if err != nil {
		logger.Warn("Failed to create default config", zap.String("path", cfgFile), zap.Error(err))
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("Failed to read config file", zap.Error(err))
	} else {
		logger.Debug("Using config file", zap.String("path", viper.ConfigFileUsed()))
	}
}
