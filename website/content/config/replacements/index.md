---
title: "Replacements"
date: 2021-01-31T14:14:11-08:00
draft: false
weight: 6
summary: Automate version values in your entire project with find and replace options.
---

### replacements
type: `[]Replacement` | default: `empty` | optional

List of replacements.

Example for NodeJS package.json

```yaml
replacements:
  - path: package.json
    find: '  "version": ".*",'
    replace: '  "version": "{{.VersionNoPrefix}}",'
```

## Replacement
type: `struct`

When working in projects that include the version directly in the source code
replacements can be used to replace those values.
This works similar to the find and replace from IDE tools but also includes the
file path of the file.

### path
type: `string` | default: `""` | required

Path of the file to find and replace in.

### find
type: `string` | default: `""` | required

Regular expression to search for in the file.

### replace
type: `string` | default: `""` | required

Template string to replace the line with.

**Replace Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Version** | `string` | Semantic version of the release, includes `v` suffix if used |
| **VersionNoPrefix** | `string` | Semantic version of the release without the suffix if used |
