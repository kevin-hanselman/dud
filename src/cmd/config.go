package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	validFields      = []string{"cache", "remote"}
	targetUserConfig bool
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config {get|set}",
		Short: "Print or modify fields in the config file",
		Long:  "Config prints or modifies fields in the config file",
	}

	configCmd.AddCommand(
		&cobra.Command{
			Use:       "get <config_field>",
			Short:     "Get the value of a field in the config file",
			Long:      "Get the value of a field in the config file",
			ValidArgs: validFields,
			Args:      cobra.ExactValidArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				rootDir, err := getProjectRootDir()
				if err != nil {
					fatal(err)
				}
				if err := lockProject(rootDir); err != nil {
					fatal(err)
				}
				if err := readConfig(rootDir); err != nil {
					fatal(err)
				}
				logger.Info.Println(viper.Get(args[0]))
			},
		},
	)

	setCmd := &cobra.Command{
		Use:       "set <config_field> <new_value>",
		Short:     "Set the value of a field in the config file",
		Long:      "Set the value of a field in the config file",
		ValidArgs: validFields,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected two arguments, got %d", len(args))
			}

			for _, field := range validFields {
				if args[0] == field {
					return nil
				}
			}

			return fmt.Errorf(
				"invalid argument; expected one of [%s]",
				strings.Join(validFields, ", "),
			)
		},
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if targetUserConfig {
				_, err = readUserConfig()
			} else {
				var rootDir string
				rootDir, err = getProjectRootDir()
				if err != nil {
					fatal(err)
				}
				if err := lockProject(rootDir); err != nil {
					fatal(err)
				}
				_, err = readProjectConfig(rootDir)
			}

			if err != nil && !os.IsNotExist(err) {
				fatal(err)
			}
			viper.Set(args[0], args[1])
			if err := viper.WriteConfig(); err != nil {
				fatal(err)
			}
		},
	}
	setCmd.Flags().BoolVarP(
		&targetUserConfig,
		"user",
		"u",
		false,
		"target the user-level config file",
	)
	configCmd.AddCommand(setCmd)

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Long:  "Print the config file path",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var (
				configPath string
				err        error
			)

			if targetUserConfig {
				configPath, err = readUserConfig()
			} else {
				var rootDir string
				rootDir, err = getProjectRootDir()
				if err != nil {
					fatal(err)
				}
				configPath, err = readProjectConfig(rootDir)
			}

			if err != nil && !os.IsNotExist(err) {
				fatal(err)
			}
			fmt.Println(configPath)
		},
	}
	pathCmd.Flags().BoolVarP(
		&targetUserConfig,
		"user",
		"u",
		false,
		"target the user-level config file",
	)
	configCmd.AddCommand(pathCmd)

	rootCmd.AddCommand(configCmd)
}
