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
```
scoop bucket add repo https://github.com/miniscruff/changie
scoop install changie
```

## Manual
* Download from [here](https://github.com/miniscruff/changie/releases).
* Add executable somewhere in your path depending on your platform.

## From source
Go get can be used to download

```
go get -u github.com/miniscruff/changie
```
