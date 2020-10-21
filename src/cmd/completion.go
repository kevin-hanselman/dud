package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion script",
	Long: `To load completions:

Bash:

$ source <(dud completion bash)

# To load completions for each session, execute once:
Linux:
$ dud completion bash > /etc/bash_completion.d/dud
MacOS:
$ dud completion bash > /usr/local/etc/bash_completion.d/dud

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ dud completion zsh > "${fpath[1]}/_dud"

# You will need to start a new shell for this setup to take effect.

Fish:

$ dud completion fish | source

# To load completions for each session, execute once:
$ dud completion fish > ~/.config/fish/completions/dud.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		}
	},
	// Override rootCmd's PersistentPreRun which changes dir to the project
	// root.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {},
}
