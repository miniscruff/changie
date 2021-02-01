---
title: "Quick Start"
date: 2021-01-30T23:38:40-08:00
draft: false
weight: 2
summary: Quick guide on how to get started with changie.
---

> Before starting, if you already have a `CHANGELOG.md` read the 
> [backup guide](/guide/backup) first.

Run `changie init` to bootstrap your project with a sample config, header and empty changelog.

`changie init`

You can [configure](/config) changie by editing the generated `.changie.yaml` file.

When completing work on a feature, bugfix or user impacting change use the new command
to generate your change file.

`changie new`

When it is time to prepare your next release, batch all unreleased changes into one using the batch command.

`changie batch <version>`

After you have batched a new version you can merge it into the parent changelog using the merge command.

`changie merge`
