---
title: "changie batch"
description: "Help for using the 'changie batch' command"
---
## changie batch

Batch unreleased changes into a single changelog

### Synopsis

Merges all unreleased changes into one version changelog.

Batch takes one argument for the next version to use, below are possible options.
* A specific semantic version value, with optional prefix
* Major, minor or patch to bump one level by one
* Auto which will automatically bump based on what changes were found

The new version changelog can then be modified with extra descriptions,
context or with custom tweaks before merging into the main file.
Line breaks are added before each formatted line except the first, if you wish to
add more line breaks include them in your format configurations.

Changes are sorted in the following order:
* Components if enabled, in order specified by config.components
* Kinds if enabled, in order specified by config.kinds
* Timestamp oldest first

```
changie batch version|major|minor|patch|auto [flags]
```

### Options

```
  -d, --dry-run              Print batched changes instead of writing to disk, does not delete fragments
      --footer-path string   Path to version footer file in unreleased directory
  -f, --force                Force a new version file even if one already exists
      --header-path string   Path to version header file in unreleased directory
  -h, --help                 help for batch
  -i, --include strings      Include extra directories to search for change files, relative to change directory
  -k, --keep                 Keep change fragments instead of deleting them
  -m, --metadata strings     Metadata values to append to version
      --move-dir string      Path to move unreleased changes
  -p, --prerelease strings   Prerelease values to append to version
  -j, --project string       Specify which project version we are batching
      --remove-prereleases   Remove existing prerelease versions
```

### SEE ALSO

* [changie](changie.md)	 - changie handles conflict-free changelog management

