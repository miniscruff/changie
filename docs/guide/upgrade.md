---
title: 'Upgrade Guide'
description: How to upgrade versions that are backwards incompatible
---

## From v0.5.0
Kind configuration moved from a string array to an array of objects.
These objects allow you to customize each kind but does cause a backwards
incompatibility issue.

In order to resolve this issue you will need to specify kinds as an object.
The old string value is now the label.

```yaml
# Old
kinds:
  - Added
  - Changed
  - Deprecated

# New
kinds:
  - label: Added
  - label: Changed
  - label: Deprecated
```
