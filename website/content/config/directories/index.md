---
title: "Directories and Files"
date: 2021-01-31T14:13:35-08:00
draft: false
weight: 1
summary: Configure paths and files to fit your project.
---

Directories and files can have there paths adjusted from the config.

### CHANGIE_CONFIG_PATH
By default Changie will try and load `.changie.yaml` or `.changie.yml` before running
commands.
If you want to change this path set the environment variable `CHANGIE_CONFIG_PATH`
to the desired file.

```sh
export CHANGIE_CONFIG_PATH=./tools/changie.yaml
changie latest
```

### changesDir
type: `string` | default: `""` | required

Directory for change files, header file and unreleased files.
Relative to project root.

### unreleasedDir
type: `string` | default: `""` | required

Directory for all unreleased change files.
Relative to `changesDir`.

### headerPath
type: `string` | default: `""` | required

When merging all versions into one changelog file a header is added at the top.
A default header is created when initializing that follows "Keep a Changelog".

Filepath for your changelog header file.
Relative to `changesDir`.

### changelogPath
type: `string` | default: `""` | required

Filepath for the generated changelog file.
Relative to project root.

### versionExt
type: `string` | default: `""` | required

File extension for generated version files.
This should probably match your changelog path file.
Must not include the period.
