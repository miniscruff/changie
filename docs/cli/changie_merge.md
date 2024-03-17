---
title: "changie merge"
description: "Help for using the 'changie merge' command"
---
## changie merge

Merge all versions into one changelog

### Synopsis

Merge all version files into one changelog file and run any replacement commands.

Note that a newline is added between each version file.

```
changie merge [flags]
```

### Options

```
  -d, --dry-run                     Print merged changelog instead of writing to disk, will not run replacements
  -h, --help                        help for merge
  -u, --include-unreleased string   Include unreleased changes with this value as the header
```

### SEE ALSO

* [changie](changie.md)	 - changie handles conflict-free changelog management

