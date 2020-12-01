package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"runtime/trace"

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
		// TODO: Ensure we always close debugOutput to prevent resource
		// leaks. This probably requires all sub-commands not to call
		// logger.Fatal() and to leave all exit paths to the root command.
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if doProfile && doTrace {
				logger.Fatal(errors.New("cannot enable both profiling and tracing"))
			}
			if doProfile {
				logger.Println("enabled profiling")
				debugOutput, err := os.Create("dud.pprof")
				if err != nil {
					logger.Fatal(err)
				}
				stopProfiling = fgprof.Start(debugOutput, fgprof.FormatPprof)
				fgprof.Start(debugOutput, fgprof.FormatPprof)
			} else if doTrace {
				logger.Println("enabled tracing")
				debugOutput, err := os.Create("dud.trace")
				if err != nil {
					logger.Fatal(err)
				}
				trace.Start(debugOutput)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if doProfile {
				defer debugOutput.Close()
				logger.Println("writing profiling output to dud.pprof")
				if err := stopProfiling(); err != nil {
					logger.Fatal(err)
				}
			} else if doTrace {
				defer debugOutput.Close()
				logger.Println("writing tracing output to dud.trace")
				trace.Stop()
			}
		},
	}

	// This is the Logger for the entire application.
	logger *log.Logger

	// This is the project root directory.
	rootDir string

	doProfile, doTrace bool
	debugOutput        *os.File
	stopProfiling      func() error
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
	pflag.BoolVar(&doProfile, "profile", false, "enable profiling")
	pflag.BoolVar(&doTrace, "trace", false, "enable tracing")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Println(Version)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:    "gen-docs",
		Short:  "Generate Markdown documentation for this command",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			linkHandler := func(name string) string {
				// See: https://gohugo.io/content-management/cross-references/
				return fmt.Sprintf(`{{< relref "%s" >}}`, name)
			}

			filePrepender := func(filename string) string {
				name := filepath.Base(filename)
				base := strings.TrimSuffix(name, filepath.Ext(name))
				return fmt.Sprintf("---\ntitle: %s\n---\n", strings.Replace(base, "_", " ", -1))
			}

			dir := args[0]
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			return doc.GenMarkdownTreeCustom(rootCmd, dir, filePrepender, linkHandler)
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
