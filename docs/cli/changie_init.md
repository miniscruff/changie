---
title: "changie init"
description: "Help for using the 'changie init' command"
---
## changie init

Initialize a new changie skeleton

### Synopsis

Initialize a few changie specifics.

* Folder to place all change files
* Subfolder to place all unreleased changes
* Header file to place on top of the changelog
* Output file when generating a changelog
* Unreleased folder includes a .gitkeep file

Values will also be saved in a changie config at .changie.yaml.
Default values follow keep a changelog and semver specs but are customizable.

```
changie init [flags]
```

### Options

```
  -d, --dir string      directory for all changes (default ".changes")
  -f, --force           force initialize even if config already exist
  -h, --help            help for init
  -o, --output string   file path to output our changelog (default "CHANGELOG.md")
```

### SEE ALSO

* [changie](changie.md)	 - changie handles conflict-free changelog management

