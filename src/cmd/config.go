package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config {get|set}",
		Short: "Print or modify fields in the config file",
		Long:  "Config prints or modifies fields in the config file",
	}

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "get config_field",
			Short: "Get the value of a field in the config file",
			Long:  "Get the value of a field in the config file",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(viper.Get(args[0]))
			},
		},
	)

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "set config_field new_value",
			Short: "Set the value of a field in the config file",
			Long:  "Set the value of a field in the config file",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				viper.Set(args[0], args[1])
				if err := viper.WriteConfig(); err != nil {
					logger.Fatal(err)
				}
			},
		},
	)

	rootCmd.AddCommand(configCmd)
}
