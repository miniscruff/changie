---
title: "Quick Start"
description: Quick guide on how to get started with changie.
---

> Before starting, if you already have a `CHANGELOG.md` read the
> [backup guide](backup.md) first.

Run `init` to bootstrap your project with a sample config, header and empty changelog.

```sh
changie init
```

You can [configure](../config/index.md) changie by editing the generated `.changie.yaml` file.

When completing work on a feature, bugfix or user impacting change use the new command
to generate your change file.

```sh
changie new
```

When it is time to prepare your next release, batch all unreleased changes into one using the batch command.

```sh
# changie supports semver bump values
changie batch <major|minor|patch>
# using an explicit version
changie batch <version>
# or using auto if you have kinds configured for auto bumps
changie batch auto
```

After you have batched a new version you can merge it into the parent changelog using the merge command.

```sh
changie merge
```

<video controls>
<source src="/static/quick_start.webm" type="video/webm">
</video>
