package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/trace"
	"strings"

	"github.com/felixge/fgprof"
	"github.com/kevin-hanselman/dud/src/agglog"
	"github.com/kevin-hanselman/dud/src/cache"
	"github.com/kevin-hanselman/dud/src/fsutil"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

const (
	indexPath = ".dud/index"
	lockPath  = ".dud/lock"
)

type emptyIndexError struct{}

func (e emptyIndexError) Error() string {
	return "index is empty"
}

type projectLockedError struct{}

func (e projectLockedError) Error() string {
	return fmt.Sprintf("project lock file '%s' exists", lockPath)
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
				// TODO: If we stop relying on the project-wide lock file, this
				// should be flocked.
				debugOutput, err := os.Create("dud.pprof")
				if err != nil {
					fatal(err)
				}
				// TODO: Consider replacing this with github.com/pkg/profile
				stopProfiling = fgprof.Start(debugOutput, fgprof.FormatPprof)
				fgprof.Start(debugOutput, fgprof.FormatPprof)
			} else if doTrace {
				logger.Info.Println("enabled tracing")
				// TODO: If we stop relying on the project-wide lock file, this
				// should be flocked.
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

	doProfile, doTrace, verbose, projectLocked bool
	debugOutput                                *os.File
	stopProfiling                              func() error
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
		Debug: log.New(io.Discard, "", 0),
	}
	if err := rootCmd.Execute(); err != nil {
		fatal(err)
	}
	if err := unlockProject(); err != nil {
		fatal(err)
	}
	if err := stopDebugging(); err != nil {
		logger.Error.Fatal(err)
	}
}

// fatal ensures we gracefully stop profiling or tracing before exiting.
func fatal(err error) {
	if !errors.Is(err, projectLockedError{}) {
		if err := unlockProject(); err != nil {
			logger.Error.Println(err)
		}
	}
	if err := stopDebugging(); err != nil {
		logger.Error.Println(err)
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

func readUserConfig() (path string, err error) {
	userConfigDir := os.Getenv("XDG_CONFIG_HOME")
	if userConfigDir == "" {
		userConfigDir, err = homedir.Expand("~/.config")
		if err != nil {
			return
		}
	}
	dir := filepath.Join(userConfigDir, "dud")
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path = filepath.Join(dir, "config.yaml")
	viper.SetConfigFile(path)
	err = viper.MergeInConfig()
	return
}

func readProjectConfig(rootDir string) (path string, err error) {
	path = filepath.Join(rootDir, ".dud", "config.yaml")
	viper.SetConfigFile(path)
	err = viper.MergeInConfig()
	return
}

// Read the project and user config files and merge them. Project config files
// take precedence over user config files.
func readConfig(rootDir string) (err error) {
	viper.SetDefault("cache", ".dud/cache")

	if _, err := readUserConfig(); err != nil {
		isNotExist := os.IsNotExist(err)
		_, isViperNotFound := err.(viper.ConfigFileNotFoundError)
		if !(isViperNotFound || isNotExist) {
			return err
		}
	}
	_, err = readProjectConfig(rootDir)
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

// pathAbsThenRel ensures path is absolute before calling
// filepath.Rel(base, target).
func pathAbsThenRel(base string, path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Rel(base, absPath)
}

// Prevent concurrent invocations of Dud in the current project using a simple
// lock file. This isn't a comprehensive solution for concurrent user errors,
// but it should prevent the most common problems.
func lockProject(rootDir string) error {
	// If we're already in the project root, we technically can use lockPath
	// directly, but this approach explicitly requires we know the project
	// root.
	lockFile, err := os.OpenFile(
		filepath.Join(rootDir, lockPath),
		// O_EXCL is key. If the file already exists, someone else has already
		// claimed the file.
		os.O_CREATE|os.O_RDWR|os.O_EXCL,
		0o600,
	)
	if err == nil {
		projectLocked = true
		return lockFile.Close()
	}
	if os.IsExist(err) {
		logger.Error.Println(`Another invocation of Dud may be running, or Dud may have exited
unexpectedly and orphaned the lock file. If you are certain Dud is not already
running and your project is healthy, you may remove the lock file and try
running Dud again.`)
		return projectLockedError{}
	}
	return err
}

func unlockProject() error {
	if projectLocked {
		// If os.Remove succeeds, we're unlocked. If it fails, we should be calling
		// fatal(), and we don't want try unlocking again.
		projectLocked = false
		return os.Remove(lockPath)
	}
	return nil
}

func preparePaths(rootDir string, paths []string) (err error) {
	for i, path := range paths {
		paths[i], err = pathAbsThenRel(rootDir, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// Do a bunch of bookkeeping to prepare for usual execution of Dud operations.
// The paths argument is updated in-place so each path is relative to the
// project root directory.
func prepare(paths []string) (rootDir string, ch cache.LocalCache, idx index.Index, err error) {
	// The order of operations here is important. Before we cd to the project
	// root directory, we need to adjust the paths, which are relative to the
	// working directory (or absolute paths already).
	rootDir, err = getProjectRootDir()
	if err != nil {
		return
	}

	if err = preparePaths(rootDir, paths); err != nil {
		return
	}

	if err = os.Chdir(rootDir); err != nil {
		return
	}

	if err = lockProject(rootDir); err != nil {
		return
	}

	if err = readConfig(rootDir); err != nil {
		return
	}

	ch, err = cache.NewLocalCache(viper.GetString("cache"))
	if err != nil {
		return
	}

	idx, err = index.FromFile(indexPath)
	return
}
