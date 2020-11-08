package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/felixge/fgprof"
)

var (
	// Version is the application version. It is set during compilation using
	// ldflags.
	Version string

	rootCmd = &cobra.Command{
		Use: "dud",
		Long: `Dud is a tool to for storing, versioning, and reproducing large files alongside
source code.`,
		// TODO: Ensure we always close profilingOutput to prevent resource
		// leaks. This probably requires all sub-commands not to call
		// logger.Fatal() and to leave all exit paths to the root command.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if profile {
				logger.Println("enabled CPU profiling")
				profilingOutput, err := os.Create("dud.prof")
				if err != nil {
					logger.Fatal(err)
				}
				stopProfiling = fgprof.Start(profilingOutput, fgprof.FormatPprof)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if profile {
				defer profilingOutput.Close()
				logger.Println("writing CPU profiling to dud.prof")
				if err := stopProfiling(); err != nil {
					logger.Fatal(err)
				}
			}
		},
	}

	// This is the Logger for the entire application.
	logger *log.Logger

	// This is the project root directory.
	rootDir string

	profile         bool
	profilingOutput *os.File
	stopProfiling   func() error
)

// Main is the entry point to the cobra CLI.
func Main() {
	logger = log.New(os.Stderr, "", 0)
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err)
	}
}

func init() {
	// This must be a global flag, not one associated with the root Command.
	pflag.BoolVar(&profile, "profile", false, "enable profiling")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Println(Version)
		},
	})
}

func requireInitializedProject(cmd *cobra.Command, args []string) {
	var err error
	rootDir, err = getProjectRootDir()
	if err != nil {
		logger.Fatal(err)
	}
	if err := os.Chdir(rootDir); err != nil {
		logger.Fatal(err)
	}
	viper.SetConfigFile(filepath.Join(rootDir, ".dud", "config.yaml"))

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal(err)
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
