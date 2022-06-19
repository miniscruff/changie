<p align="center">
  <a href="https://changie.dev">
    <img alt="Changie Logo" src="./docs/themes/hugo-whisper-theme/static/images/logo.svg" height="256" />
  </a>
  <h3 align="center">Changie</h3>
  <p align="center">Separate your changelog from commit messages without conflicts.</p>
</p>

[![codecov](https://codecov.io/gh/miniscruff/changie/branch/main/graph/badge.svg?token=7HT2E32FMB)](https://codecov.io/gh/miniscruff/changie)
[![Go Report Card](https://goreportcard.com/badge/github.com/miniscruff/changie)](https://goreportcard.com/report/github.com/miniscruff/changie)
[![Release](https://img.shields.io/github/v/release/miniscruff/changie?sort=semver)](https://github.com/miniscruff/changie/releases)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/miniscruff/changie/test)](https://github.com/miniscruff/changie/actions?query=workflow%3Atest)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/miniscruff/changie)](https://pkg.go.dev/github.com/miniscruff/changie)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

![asciicast](./docs/static/recordings/overview.gif)

## Features
* File based changelog management keeps your commit history and release notes separate.
* Track changes while you work while the knowledge is fresh.
* Extensive [configuration options](https://changie.dev/config) to fit your project.
* Language and framework agnostic using a single go binary.

## Getting Started
* User documentation is available on the [website](https://changie.dev/).
* Specifically, the [guide](https://changie.dev/guide/) is a good place to start.
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

## Want to Contribute?
If you want to contribute through code or documentation, the [Contributing guide](CONTRIBUTING.md) is the place to start.
If you need additional help create an issue or post on discussions.

## Semantic Version Compatibility
Changie is focused around the CLI and its configuration options and aims to keep existing CLI commands and options suported in major versions.
It is possible to use Changie as a dependent package but no support or compability is guaranteed.

## License
Distributed under the [MIT License](LICENSE).
