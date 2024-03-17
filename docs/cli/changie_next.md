---
title: "changie next"
description: "Help for using the 'changie next' command"
---
## changie next

Next echos the next version based on semantic versioning

### Synopsis

Next increments version based on semantic versioning.
Check latest version and increment part (major, minor, patch).
If auto is used, it will try and find the next version based on what kinds of changes are
currently unreleased.
Echo the next release version number to be used by CI tools or other commands like batch.

```
changie next major|minor|patch|auto [flags]
```

### Options

```
  -h, --help                 help for next
  -i, --include strings      Include extra directories to search for change files, relative to change directory
  -m, --metadata strings     Metadata values to append to version
  -p, --prerelease strings   Prerelease values to append to version
  -j, --project string       Specify which project we are interested in
```

### SEE ALSO

* [changie](changie.md)	 - changie handles conflict-free changelog management

