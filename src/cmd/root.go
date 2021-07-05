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
)

type emptyIndexError struct{}

func (e emptyIndexError) Error() string {
	return "index is empty"
}

var (
	// Version is the version of the app.
	Version string

	rootCmd = &cobra.Command{
		Use: "dud",
		Long: `Dud is a lightweight tool for versioning data alongside source code and
building data pipelines.`,
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

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number and exit",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	})
}

// Main is the entry point to the cobra CLI.
func Main() {
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

// This function also converts any paths passed in to be relative to the
// project root.
// TODO: This function does a lot. Find a better abstraction (or group of
// abstractions).
func cdToProjectRootAndReadConfig(paths []string) (
	rootDir string,
	relPaths []string,
	err error,
) {
	rootDir, err = getProjectRootDir()
	if err != nil {
		return
	}
	relPaths = make([]string, len(paths))
	for i, path := range paths {
		relPaths[i], err = pathAbsThenRel(rootDir, path)
		if err != nil {
			return
		}
	}
	if err = os.Chdir(rootDir); err != nil {
		return
	}
	viper.SetConfigFile(filepath.Join(rootDir, ".dud", "config.yaml"))
	err = viper.ReadInConfig()
	return
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

// pathAbsThenRel ensures target is absolute before calling
// filepath.Rel(base, target).
func pathAbsThenRel(base, target string) (string, error) {
	target, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	return filepath.Rel(base, target)
}
