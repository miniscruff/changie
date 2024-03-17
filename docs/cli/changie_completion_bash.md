---
title: "changie completion bash"
description: "Help for using the 'changie completion bash' command"
---
## changie completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(changie completion bash)

To load completions for every new session, execute once:

#### Linux:

	changie completion bash > /etc/bash_completion.d/changie

#### macOS:

	changie completion bash > $(brew --prefix)/etc/bash_completion.d/changie

You will need to start a new shell for this setup to take effect.


```
changie completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [changie completion](changie_completion.md)	 - Generate the autocompletion script for the specified shell

