package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validFields = []string{
	"cache",
	"remote",
}

func init() {
	configCmd := &cobra.Command{
		Use:              "config {get|set}",
		Short:            "Print or modify fields in the config file",
		Long:             "Config prints or modifies fields in the config file",
		PersistentPreRun: cdToProjectRootAndReadConfig,
	}

	configCmd.AddCommand(
		&cobra.Command{
			Use:       "get config_field",
			Short:     "Get the value of a field in the config file",
			Long:      "Get the value of a field in the config file",
			ValidArgs: validFields,
			Args:      cobra.ExactValidArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(viper.Get(args[0]))
			},
		},
	)

	configCmd.AddCommand(
		&cobra.Command{
			Use:       "set config_field new_value",
			Short:     "Set the value of a field in the config file",
			Long:      "Set the value of a field in the config file",
			ValidArgs: validFields,
			Args: func(cmd *cobra.Command, args []string) error {
				if len(args) != 2 {
					return fmt.Errorf("expected two arguments, got %d", len(args))
				}

				isValid := false
				for _, field := range validFields {
					if args[0] == field {
						isValid = true
						break
					}
				}

				if !isValid {
					return fmt.Errorf(
						"invalid argument; expected one of [%s]",
						strings.Join(validFields, ", "),
					)
				}
				return nil
			},
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
