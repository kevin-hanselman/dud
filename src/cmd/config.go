package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(cacheConfig)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Modify the config file",
	Long:  "Modify the config file at the project scope",
}

var cacheConfig = &cobra.Command{
	Use:   "cache [cache-location]",
	Short: "Set the cache location",
	Long:  "Set the cache location",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		viper.Set("cache", args[0])
		viper.WriteConfigAs(".dud/config")
	},
}
