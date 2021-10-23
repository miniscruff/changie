---
title: 'GoReleaser'
date: 2021-01-31T18:53:05-08:00
weight: 3
summary: Using Changie for GoReleaser release notes
---

Changie itself uses [GoReleaser](https://goreleaser.com) and can be integrated
with a few steps.

First disable GoReleaser from generating its own release notes by adding this to your config.

```yaml
changelog:
  skip: true
```

By default GoReleaser expects to release the current tag but we can let GitHub
create one for us during the release.
To do this set the goreleaser current tag environment variable using changie latest.

```bash
export GORELEASER_CURRENT_TAG="$(changie latest)"
```

Finally we can run GoReleaser, you will need to add two parameters, release notes and skip validate.
We need to use skip validate because we skip the git tag.
If you choose to tag the commit instead you do not need to use this.

```bash
goreleaser --release-notes="changes/$(changie latest)" --skip-validate
```

If you would like to use the goreleaser github action you can reference [release.yaml](https://github.com/miniscruff/changie/blob/main/.github/workflows/release.yml).
