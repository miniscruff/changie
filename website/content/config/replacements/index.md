---
title: "Replacements"
date: 2021-01-31T14:14:11-08:00
draft: false
weight: 5
summary: Automate version values in your entire project with find and replace options.
---

### replacements
type: _[]Replacement_

When working in projects that include the version directly in the source code a replacement option exists to find and replace those values.
This works similar to the find and replace from IDE tools but also includes the file path of the file.

## Replacement
type: _struct_

### path
type: _string_

Path of the file to find and replace in.

### find
type: _string_

Regular expression to search for in the file.

### replace
type: _string_

Template string to replace the line with.

**Replace Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Version** | _string_ | Semantic version of the release, includes `v` suffix if used |
| **VersionNoPrefix** | _string_ | Semantic version of the release without the suffix if used |
