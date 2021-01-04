# Changie

_changie is in early development and is subject to change_

[![codecov](https://codecov.io/gh/miniscruff/changie/branch/main/graph/badge.svg?token=7HT2E32FMB)](https://codecov.io/gh/miniscruff/changie)
[![Go Report Card](https://goreportcard.com/badge/github.com/miniscruff/changie)](https://goreportcard.com/report/github.com/miniscruff/changie)
[![Release](https://img.shields.io/github/v/release/miniscruff/changie?sort=semver)](https://github.com/miniscruff/changie/releases)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/miniscruff/changie/test)](https://github.com/miniscruff/changie/actions?query=workflow%3Atest)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/miniscruff/changie)](https://pkg.go.dev/github.com/miniscruff/changie)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Automated changelog tool for preparing releases with lots of customization options.
Changie aims to be a universal tool for any project language or style but limiting itself to changelogs and version management.

## Installation

### From releases
Download from [here](https://github.com/miniscruff/changie/releases)

### From scoop
On windows you can use [scoop](https://scoop.sh/) by first adding this repo and then installing.
```
scoop bucket add repo https://github.com/miniscruff/changie
scoop install changie
```

### From source
Go get can be used to download

```
go get -u github.com/miniscruff/changie
```

## Quick Start

1. Use `changie init` command to bootstrap your project.
This will create everything you need with some default values.

    Note: if you already have a `CHANGELOG.md` file you should create a backup.
    You can save all your old changelogs by moving it to the "changes" directory renaming it to the latest release.
    For example, if your latest release was v0.3.2, `mv CHANGELOG_copy.md changes/v0.3.2.md`.
    This will include all the existing changelogs when generating a new one with changie.

1. Take a look at the generated `.changie.yaml` to see if there are any adjustments you need to make.
More details on configuration below.

1. Use `changie new` when a change is complete.
This will create a `.yaml` file in your "unreleased" directory.
Repeat this step for all changes until it is time to release.

1. Use `changie batch "version"` to combine all unreleased changes into a single version change.
If you need to make some manual adjustments to this version like high level overviews or similar you can do so.

    Note: Since this is a complete changelog for a single release, it works as github release notes.

1. For CI/CD usage you can use `changie latest` to print out the latest version batched.
See the release workflow [here](/.github/workflows/release.yml) for an example

1. Finally use `changie merge` to merge all version changelogs into one `CHANGELOG.md` file.
This file is the default standard for where to find all of a project changes and is also the default for changie to output.

## Configuration
See [config.go](/cmd/config.go).

All configurations are made in the `.changie.yaml` config file that is generated in the init command.
The default values represent a [keep a changelog](https://keepachangelog.com/en/1.0.0/) style setup.
This is just the default and many styles can be achieved through configuration changes.

### Directories and Files
Directories and files can have there paths adjusted from the config.

```yaml
# relative to project root
changesDir: changes
# relative to changes directory
unreleasedDir: unreleased
# filepath or name relative to changes directory
headerPath: header.tpl.md
# relative to project root
changelogPath: CHANGELOG.md
# plain extension for version files
versionExt: md
# changelog path and version ext should probably match.
```

### Formatting
Changie utilizes [go template](https://golang.org/pkg/text/template/) for formatting version, kind and change lines.
Each one can be customized through the config file and change formats can be extended with custom choices.
See [change.go](cmd/change.go) for the change structure used in the change format.

    Due to the ordering of commands you must add custom choices before
    you added any change files in order to use the custom values in your format.

```yaml
# Version structure:
# .Version: semantic version
# .Time: go time.Time of the batch command
versionFormat: "## {{.Version}}"
# Kind structure:
# .Kind: name of the kind
kindFormat: "### {{.Kind}}"
# Change include the cmd/Change type
changeFormat: "* {{.Body}}"
```

### Kinds
Every change documented for end users fall under a kind.
The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed and security.
When creating a new change you must select which of these changes fits your change.

When batching changes into a version, changes are sorted by kind in the order listed, so if you want new features listed on top place them on top of the list.

```yaml
kinds:
- Added
- Removed
- Hotfix
- etc
```

### Custom Choices
If your project wants to include more data along with each change you can add additional requirements to change files by creating custom choices.
These choices are asked when creating new change files and can be used when formatting changes.
A simple one could be the issue number or authors github name.
Currently there are three types of choices; strings, ints and enums.

```yaml
custom:
- key: Issue # used to reference the value in change format
  label: Issue Number # optional label when asking for changes
  type: int
  minInt: 1
  # maxInt: is also possible
- key: Author
  label: Author Github Name
  type: string
- key: Emoji
  label: Flair emoji
  type: enum
  enumOptions:
  - bug
  - pencil2
  - smile
  - dog
changeFormat: "* :{{.Custom.Emoji}}: {{.Body}} (#{{.Custom.Issue}} by {{.Custom.Author}})"
# Can create markdown links as well:
# [#{{.Custom.Issue}}](github.com/project/issues/{{.Custom.Issue}})
```

### Replacements
When working in projects that include the version directly in the source code a replacement option exists to find and replace those values.
This works similar to the find and replace from IDE tools but also includes the file path of the file.
An example for a Node package.json is below.

```json
{
    "name": "my project",
    "version": "0.1.0",
    "main": "main.js"
}
```

```yaml
replacements:
- path: package.json
  find: '  "version": ".*"'
  replace: '  "version": "{{.VersionNoPrefix}}",'
```

### Header
When merging all versions into one changelog file a header is added at the top.
The path is defined in the config file and a default one is created when initializing.
