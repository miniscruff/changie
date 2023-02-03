---
title: "Quick Start"
date: 2021-01-30T23:38:40-08:00
draft: false
weight: 2
asciinema: true
summary: Quick guide on how to get started with changie.
---

> Before starting, if you already have a `CHANGELOG.md` read the
> [backup guide](/guide/backup) first.

Run `init` to bootstrap your project with a sample config, header and empty changelog.

```shell
changie init
```

You can [configure](/config) changie by editing the generated `.changie.yaml` file.

When completing work on a feature, bugfix or user impacting change use the new command
to generate your change file.

```shell
changie new
```

When it is time to prepare your next release, batch all unreleased changes into one using the batch command.

```shell
# changie supports semver bump values
changie batch <major|minor|patch>
# using an explicit version
changie batch <version>
# or using auto if you have kinds configured for auto bumps
changie batch auto
```

After you have batched a new version you can merge it into the parent changelog using the merge command.

```shell
changie merge
```

{{< asciinema
  key="quick-start"
  rows="10"
  preload="1"
  author="miniscruff"
>}}
