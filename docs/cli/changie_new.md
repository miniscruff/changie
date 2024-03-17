---
title: "changie new"
description: "Help for using the 'changie new' command"
---
## changie new

Create a new change file

### Synopsis

Creates a new change file.
Change files are processed when batching a new release.
Each version is merged together for the overall project changelog.

```
changie new [flags]
```

### Options

```
  -b, --body string        Set the change body without a prompt
  -c, --component string   Set the change component without a prompt
  -m, --custom strings     Set custom values without a prompt
  -d, --dry-run            Print new fragment instead of writing to disk
  -e, --editor             Edit body message using your text editor defined by 'EDITOR' env variable
  -h, --help               help for new
  -k, --kind string        Set the change kind without a prompt
  -j, --projects strings   Set the change projects without a prompt
```

### SEE ALSO

* [changie](changie.md)	 - changie handles conflict-free changelog management

