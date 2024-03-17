---
title: "changie completion powershell"
description: "Help for using the 'changie completion powershell' command"
---
## changie completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	changie completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
changie completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [changie completion](changie_completion.md)	 - Generate the autocompletion script for the specified shell

