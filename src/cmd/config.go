package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Modify the config file",
		Long:  "Modify the config file at the project scope",
	}

	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "get <config field>",
			Short: "Get the value of a config field",
			Long:  "Get the value of a config field",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(viper.Get(args[0]))
			},
		},
	)

	configCmd.AddCommand(
		&cobra.Command{
			Use:   "set <config field> <new value>",
			Short: "Set the value of a config field",
			Long:  "Set the value of a config field",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				viper.Set(args[0], args[1])
				if err := viper.WriteConfig(); err != nil {
					log.Fatal(err)
				}
			},
		},
	)
}
