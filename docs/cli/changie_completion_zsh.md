---
title: "changie completion zsh"
description: "Help for using the 'changie completion zsh' command"
---
## changie completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(changie completion zsh)

To load completions for every new session, execute once:

#### Linux:

	changie completion zsh > "${fpath[1]}/_changie"

#### macOS:

	changie completion zsh > $(brew --prefix)/share/zsh/site-functions/_changie

You will need to start a new shell for this setup to take effect.


```
changie completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [changie completion](changie_completion.md)	 - Generate the autocompletion script for the specified shell

