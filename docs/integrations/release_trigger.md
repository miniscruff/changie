---
title: "Release Trigger"
date: 2021-01-30T23:38:40-08:00
weight: 2
summary: Kicking off a release using Changie
---

Changie expects to be the first part of the release process as it modifies files
that are kept in the repository.

A method used by Changie itself is to detect changes to the root CHANGELOG file
as a trigger to begin the release process.
Below is how you can do that in a GitHub action.

```yaml
on:
  push:
    branches: [ main ] # your default branch if different
    paths: [ CHANGELOG.md ] # your changelog file if different

jobs:
  release:
   # do your releasing here
```

Then you can use Changie to update your changelog and let your action do the rest.
