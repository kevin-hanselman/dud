package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dud",
		Short: "Dud is a tool for storing, versioning, and reproducing large files",
		Long: `Dud is a tool to store, version, and reproduce big
data files alongside source code.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Change working directory to the project root.
			// This is done here as opposed to in cobra.OnInitialize so `init`
			// and other commands can override this behavior.
			rootDir, err := getProjectRootDir()
			if err != nil {
				logger.Fatal(err)
			}
			if err := os.Chdir(rootDir); err != nil {
				logger.Fatal(err)
			}
		},
	}

	// This is the Logger for the entire application.
	logger *log.Logger
)

// Main is the entry point to the cobra CLI.
func Main() {
	logger = log.New(os.Stdout, "", 0)
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	rootDir, err := getProjectRootDir()
	if err == nil {
		viper.AddConfigPath(filepath.Join(rootDir, ".dud"))
	}
	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.Fatal(err)
		}
	}
}

func getProjectRootDir() (string, error) {
	dirname, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		dudFolderExists, err := fsutil.Exists(filepath.Join(dirname, ".dud"), false)
		if err != nil {
			return "", err
		}

		if dudFolderExists {
			return dirname, nil
		}

		dirname = filepath.Dir(dirname)

		if dirname == "/" {
			return "", errors.New("no project root directory found")
		}
	}
}
