package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"

	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"

	"github.com/felixge/fgprof"
)

const (
	indexPath = ".dud/index"

	emptyIndexMessage = "index is empty"
)

var (
	rootCmd = &cobra.Command{
		Use: "dud",
		Long: `Dud is a tool for storing, versioning, and reproducing large files alongside
source code.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				logger.Debug = log.New(os.Stderr, "", 0)
			}
			if doProfile && doTrace {
				fatal(errors.New("cannot enable both profiling and tracing"))
			}
			if doProfile {
				logger.Info.Println("enabled profiling")
				debugOutput, err := os.Create("dud.pprof")
				if err != nil {
					fatal(err)
				}
				stopProfiling = fgprof.Start(debugOutput, fgprof.FormatPprof)
				fgprof.Start(debugOutput, fgprof.FormatPprof)
			} else if doTrace {
				logger.Info.Println("enabled tracing")
				debugOutput, err := os.Create("dud.trace")
				if err != nil {
					fatal(err)
				}
				if err := trace.Start(debugOutput); err != nil {
					fatal(err)
				}
			}
		},
		DisableAutoGenTag: true,
	}

	// This is the Logger for the entire application.
	logger *agglog.AggLogger

	// This is the project root directory.
	rootDir string

	doProfile, doTrace, verbose bool
	debugOutput                 *os.File
	stopProfiling               func() error
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&doProfile, "profile", false, "enable profiling")
	rootCmd.PersistentFlags().BoolVar(&doTrace, "trace", false, "enable tracing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "increase output verbosity")

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

// Main is the entry point to the cobra CLI.
func Main(version string) {
	rootCmd.Version = version
	logger = &agglog.AggLogger{
		Error: log.New(os.Stderr, "Error: ", 0),
		Info:  log.New(os.Stdout, "", 0),
		Debug: log.New(ioutil.Discard, "", 0),
	}
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
	if err := stopDebugging(); err != nil {
		logger.Error.Fatal(err)
	}
}

// fatal ensures we gracefully stop profiling or tracing before exiting.
func fatal(err error) {
	debugErr := stopDebugging()
	if debugErr != nil {
		logger.Error.Println(debugErr)
	}
	logger.Error.Fatal(err)
}

func stopDebugging() error {
	if debugOutput != nil {
		defer debugOutput.Close()
	}
	if doTrace {
		logger.Info.Println("writing tracing output to dud.trace")
		trace.Stop()
	} else if stopProfiling != nil {
		logger.Info.Println("writing profiling output to dud.pprof")
		if err := stopProfiling(); err != nil {
			return err
		}
	}
	return nil
}

func cdToProjectRootAndReadConfig(_ *cobra.Command, _ []string) {
	var err error
	rootDir, err = getProjectRootDir()
	if err != nil {
		fatal(err)
	}
	if err := os.Chdir(rootDir); err != nil {
		fatal(err)
	}
	viper.SetConfigFile(filepath.Join(rootDir, ".dud", "config.yaml"))

	if err := viper.ReadInConfig(); err != nil {
		fatal(err)
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
