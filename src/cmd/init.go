package cmd

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize a Dud project",
		Long:  `Init initializes a Dud project in the current directory.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := os.MkdirAll(".dud/cache", 0o755); err != nil {
				fatal(err)
			}

			dudConf := `# Dud project-level config file

# This config file overrides any user configuration from
# "$XDG_CONFIG_HOME/dud/config.yaml" or "~/.config/dud/config.yaml".

# This is the default cache location. Uncomment here or add to your user-level
# config to override.
# cache: .dud/cache

# To enable push and fetch, set 'remote' to a valid rclone remote path. For
# example, if you have a remote called "s3" in your .dud/rclone.conf, and you
# want your remote cache to live in a bucket called 'dud', you would write:
#
# remote: s3:dud
#
# For more info, see the rclone docs:
# https://rclone.org/docs/#syntax-of-remote-paths
`

			if err := ioutil.WriteFile(".dud/config.yaml", []byte(dudConf), 0o644); err != nil {
				fatal(err)
			}

			if err := ioutil.WriteFile(indexPath, []byte{}, 0o644); err != nil {
				fatal(err)
			}

			if err := ioutil.WriteFile(".dud/.gitignore", []byte("/cache/"), 0o644); err != nil {
				fatal(err)
			}

			rcloneConf := `# rclone config file
# Run 'rclone --config .dud/rclone.conf config' to setup a remote Dud cache,
# and then set the 'remote' value in .dud/config.yaml to a valid rclone remote
# path.
# See: https://rclone.org/docs/#syntax-of-remote-paths
`
			if err := ioutil.WriteFile(".dud/rclone.conf", []byte(rcloneConf), 0o644); err != nil {
				fatal(err)
			}

			logger.Info.Println(`Dud project initialized.
See .dud/config.yaml and .dud/rclone.conf to customize the project.`)
		},
	})
}
