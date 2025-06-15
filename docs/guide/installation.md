---
title: 'Installation'
---

## ArchLinux

An [AUR package](https://aur.archlinux.org/packages/changie/) is available.

```sh
trizen -S changie
```

## Docker

Docker images are uploaded to [GitHub Packages](https://github.com/miniscruff/changie/pkgs/container/changie).

```sh
# Replace latest with any changie command
docker run \
    --mount type=bind,source=$PWD,target=/src \
    -w /src \
    ghcr.io/miniscruff/changie \
    latest
```

**Notes**
1. In order to complete prompts with docker you will need to use an [interactive terminal](https://docs.docker.com/engine/reference/commandline/run/#assign-name-and-allocate-pseudo-tty---name--it)
1. You may also want to include your own user and group ID if any files would be created using the [user option](https://docs.docker.com/engine/reference/run/#user).

```sh
docker run \
    --mount type=bind,source=$PWD,target=/src \
    -w /src \
    -it \
    --user $(id -u ${USER}):$(id -g ${USER}) \
    ghcr.io/miniscruff/changie \
    new
```

## GitHub action

This [GitHub action](https://github.com/miniscruff/changie-action) can be used.

```yaml
- name: Batch a new minor version
  uses: miniscruff/changie-action@VERSION # view action repo for latest version
  with:
    version: latest # use the latest changie version
    args: batch minor
```

## macOS with Homebrew

On macOS, you can use [Homebrew](https://brew.sh/) to install changie from homebrew core.

```sh
brew install changie
```

## Manual

* Download from [here](https://github.com/miniscruff/changie/releases).
* Add executable somewhere in your path depending on your platform.

## Manual deb/rpm packages

Download a `.deb` or `.rpm` file from the [releases page](https://github.com/miniscruff/changie/releases)
and install with `dpkg -i` and `rpm -i` respectively.

## Mise

Changie is included in the [Mise](https://mise.jdx.dev/) registry.
It's recommended to use `mise use` for tools.

```sh
mise use changie
```

This will add changie to the `mise.toml` file.
```toml
[tools]
changie = "latest"
```

Or if you only want to use changie for a single mise task, such as `changie new`.

```toml
[tasks.fragment]
tools.changie = "latest"
run = "changie new"
```

## NodeJS

Changie is available as an [NPM package](https://www.npmjs.com/package/changie).

To add as a dependency of your project

```sh
npm i -D changie
```

To install globally

```
npm i -g changie
```

To run without adding a dependency

```
npx changie
```

## Source

Go install can be used to download changie from the main branch.

```sh
go install github.com/miniscruff/changie@latest
```

## UBI ( universal binary installer )

[UBI](https://github.com/houseabsolute/ubi) can be used to install Changie binaries directly from
GitHub.

```sh
ubi --project miniscruff/changie --in /binary/path
```

## Windows Scoop

On Windows you can use [scoop](https://scoop.sh/) by first adding the repo and then installing.
```sh
scoop bucket add changie https://github.com/miniscruff/changie
scoop install changie
```

## Winget

On Windows you can also use the [winget](https://github.com/microsoft/winget-pkgs) package manager.

```sh
winget install miniscruff.changie
```
