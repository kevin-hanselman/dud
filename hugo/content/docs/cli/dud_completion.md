---
title: dud completion
---
## dud completion

Generate shell completion script

### Synopsis

Completion generates a completion script for the given shell.

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


```
dud completion {bash|zsh|fish}
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --profile   enable profiling
      --trace     enable tracing
```

### SEE ALSO

* [dud]({{< relref "dud.md" >}})	 - 
