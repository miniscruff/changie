---
title: "Directories and Files"
date: 2021-01-31T14:13:35-08:00
draft: false
weight: 1
summary: Configure paths and files to fit your project.
---

Directories and files can have there paths adjusted from the config.

### changesDir
type: *string*

Directory for change files, header file and unreleased files.
Relative to project root.

### unreleasedDir
type: *string*

Directory for all unreleased change files.
Relative to `changesDir`.

### headerPath
type: *string*

When merging all versions into one changelog file a header is added at the top.
A default header is created when initializing that follows "Keep a Changelog".

Filepath for your changelog header file.
Relative to `changesDir`.

### changelogPath
type: *string*

Filepath for the generated changelog file.
Relative to project root.

### versionExt
type: *string*

File extension for generated version files.
This should probably match your changelog path file.
Must not include the period.
