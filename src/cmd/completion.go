package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion {bash|zsh|fish}",
	Short: "Generate shell completion script",
	Long: `Completion generates a completion script for the given shell.

#### Bash

    $ source <(dud completion bash)

To load completions for each session, execute once:

On Linux:

    $ dud completion bash > /etc/bash_completion.d/dud

On MacOS:

    $ dud completion bash > /usr/local/etc/bash_completion.d/dud

#### Zsh

If shell completion is not already enabled in your environment you will need to
enable it. You can execute the following once:

    $ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for each session, execute once:

    $ dud completion zsh > "${fpath[1]}/_dud"

You will need to start a new shell for this setup to take effect.

#### Fish

    $ dud completion fish | source

To load completions for each session, execute once:

    $ dud completion fish > ~/.config/fish/completions/dud.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.Root().GenFishCompletion(os.Stdout, true)
		}
		if err != nil {
			logger.Fatal(err)
		}
	},
}
