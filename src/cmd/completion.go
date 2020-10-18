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

$ source <(duc completion bash)

# To load completions for each session, execute once:
Linux:
$ duc completion bash > /etc/bash_completion.d/duc
MacOS:
$ duc completion bash > /usr/local/etc/bash_completion.d/duc

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ duc completion zsh > "${fpath[1]}/_duc"

# You will need to start a new shell for this setup to take effect.

Fish:

$ duc completion fish | source

# To load completions for each session, execute once:
$ duc completion fish > ~/.config/fish/completions/duc.fish
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
}
