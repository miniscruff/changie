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

## Manual
* Download from [here](https://github.com/miniscruff/changie/releases).
* Add executable somewhere in your path depending on your platform.

## From source
Go install can be used to download changie from the main branch.

```
go install github.com/miniscruff/changie@latest
```
