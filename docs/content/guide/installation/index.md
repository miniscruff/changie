---
title: "Installation"
date: 2021-01-30T23:20:31-08:00
draft: false
weight: 1
summary: How to install changie under different operating systems.
---

## deb/rpm
Download a `.deb` or `.rpm` file from the [releases page](https://github.com/miniscruff/changie/releases)
and install with `dpkg -i` and `rpm -i` respectively.

## Windows Scoop
On windows you can use [scoop](https://scoop.sh/) by first adding the repo and then installing.
```sh
scoop bucket add repo https://github.com/miniscruff/changie
scoop install changie
```

## macOS with Homebrew

On macOS, you can use [Homebrew](https://brew.sh/) to install by first tapping
the repository.

```sh
brew tap miniscruff/changie https://github.com/miniscruff/changie
brew install changie
```

## ArchLinux
An [AUR package](https://aur.archlinux.org/packages/changie/) is available.

```sh
trizen -S changie
```

## GitHub action
This [GitHub action](https://github.com/miniscruff/changie-action) can be used.

```yaml
- name: Batch a new minor version
  uses: miniscruff/changie-action@version # view action repo for latest version
  with:
    version: latest # download the latest changie version
    args: batch minor
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
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

**NOTE** : The `it` option is not required in batch mode when you specify all required flags for the `new` command.

## Manual
* Download from [here](https://github.com/miniscruff/changie/releases).
* Add executable somewhere in your path depending on your platform.

## From source
Go install can be used to download changie from the main branch.

```
go install github.com/miniscruff/changie@latest
```
