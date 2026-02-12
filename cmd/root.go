package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	wsURL   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "resonite-sh",
	Short: "A shell interface for Resonite VR via ResoLink",
	Long: `resonite-sh provides a REPL/command-line interface for Resonite VR
that treats everything as an object in a filesystem-like hierarchy.

Navigate slots like directories, inspect components like files,
and manipulate the scene graph with familiar shell commands.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.resonite-sh/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&wsURL, "url", "", "ResoLink WebSocket URL (default: ws://localhost:29551)")

	// Bind flags to viper
	viper.BindPFlag("connection.url", rootCmd.PersistentFlags().Lookup("url"))
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

		// Search config in home directory with name ".resonite-sh" (without extension).
		configDir := filepath.Join(home, ".resonite-sh")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("RESONITE")
	viper.AutomaticEnv() // read in environment variables that match

	// Set defaults
	viper.SetDefault("connection.url", "ws://localhost:29551")
	viper.SetDefault("connection.timeout", "30s")
	viper.SetDefault("connection.reconnect", true)
	viper.SetDefault("connection.reconnect_interval", "5s")
	viper.SetDefault("cache.enable", true)
	viper.SetDefault("cache.ttl", "5m")
	viper.SetDefault("editor.format", "yaml")
	viper.SetDefault("ui.prompt", "resonite:%p> ")
	viper.SetDefault("ui.colors", true)
	viper.SetDefault("history.size", 1000)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
