<p align="center">
  <a href="https://changie.dev">
    <img alt="Changie Logo" src="./docs/themes/hugo-whisper-theme/static/images/logo.svg" height="256" />
  </a>
  <h3 align="center">Changie</h3>
  <p align="center">Separate your changelog from commit messages without conflicts.</p>
</p>

[![Codecov](https://img.shields.io/codecov/c/github/miniscruff/changie?style=for-the-badge&logo=codecov)](https://codecov.io/gh/miniscruff/changie)
[![GitHub release](https://img.shields.io/github/v/release/miniscruff/changie?style=for-the-badge&logo=github)](https://github.com/miniscruff/changie/releases)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/miniscruff/changie/test.yml?event=push&style=for-the-badge&logo=github)](https://github.com/miniscruff/changie/actions/workflows/test.yml)
[![Go Packge](https://img.shields.io/badge/Go-Reference-grey?style=for-the-badge&logo=go&logoColor=white&label=%20&labelColor=007D9C)](https://pkg.go.dev/github.com/miniscruff/changie)
[![Awesome Go](https://img.shields.io/badge/awesome-awesome?style=for-the-badge&logo=awesomelists&logoColor=white&label=%20&labelColor=CCA6C4&color=494368)](https://github.com/avelino/awesome-go#utilities)

![getting_started](./examples/getting_started.gif)

## Features
* File based changelog management keeps your commit history and release notes separate.
* Track changes while you work while the knowledge is fresh.
* Extensive [configuration options](https://changie.dev/config) to fit your project.
* Language and framework agnostic using a single go binary.

## Getting Started
* User documentation is available on the [website](https://changie.dev/).
* Specifically, the [guide](https://changie.dev/guide/) is a good place to start.
* There is also a [Changie GitHub Action](https://github.com/miniscruff/changie-action) you can use
* View Changie's [Changelog](CHANGELOG.md) for an example.

## Need help?
Use the [discussions page](https://github.com/miniscruff/changie/discussions) for help requests and how-to questions.

Please open [GitHub issues](https://github.com/miniscruff/changie/issues) for bugs and feature requests.
File an issue before creating a pull request, unless it is something simple like a typo.

## Media
* [Introduction to Changelog Management](https://dev.to/miniscruff/changie-automated-changelog-tool-11ed)
* [Get help with Changie](https://dev.to/miniscruff/get-help-automating-your-releases-21ig)
* [Headers, footers, bumping, latest, dry run](https://dev.to/miniscruff/changie-automated-changelog-generation-for-any-project-1b52)
* [Prereleases and metadata](https://dev.to/miniscruff/changie-automated-changelog-generation-for-large-projects-41hm)
* [Choices and Replacements](https://dev.to/miniscruff/changie-choices-and-replacements-40p5)
* [Auto mode and GitHub action](https://dev.to/miniscruff/changie-auto-mode-and-github-action-1279)

_pull requests encouraged to add your own media here_

## Want to Contribute?
If you want to contribute through code or documentation, the [Contributing guide](CONTRIBUTING.md) is the place to start.
If you need additional help create an issue or post on discussions.

## Semantic Version Compatibility
Changie is focused around the CLI and its configuration options and aims to keep existing CLI commands and options suported in major versions.
It is possible to use Changie as a dependent package but no support or compability is guaranteed.

## Tasks
Below is a list of common development tasks, these can easily be run using [xc](https://xcfile.dev/).
For example `xc test` will run the test suite.

### test
Run unit test suite with code coverage enabled
```
go test -coverprofile=c.out ./...
```

### coverage
requires: test
```
go tool cover -html=c.out
```

### lint
```
goimports -w -local github.com/miniscruff/changie .
golangci-lint run ./...
```

### gen
Generate config and CLI docs

```
go run main.go gen
```

### docs-serve
Serve a locally running hugo instance for documentation

requires: gen
```
hugo serve -s docs
```

## License
Distributed under the [MIT License](LICENSE).
