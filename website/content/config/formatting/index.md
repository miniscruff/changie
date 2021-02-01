---
title: "Formatting"
date: 2021-01-31T14:13:51-08:00
draft: false
weight: 3
summary: Customize how version and changelog files are generated.
---

Changie utilizes [go template](https://golang.org/pkg/text/template/) for formatting version, kind, change and replacement lines.
Additional fields can be added to change lines by adding [custom choices](/config/choices).

> Due to the ordering of commands you must add custom choices before
you added any change files in order to use the custom values in your format.

### kinds
type: _[]string_

Every change documented for end users fall under a kind.
The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed and security.
When creating a new change you must select which of these changes fits your change.

When batching changes into a version, changes are sorted by kind in the order listed, so if you want new features listed on top place them on top of the list.

### versionFormat
type: _string_

Template used to generate version headers in version files and changelog.

**Version Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Version** | _string_ | Semantic version of the changes. |
| **Time** | _time.Time_ | Time of generated version. |

### kindFormat
type: _string_

Template used to generate kind headers for version files and changelog.

**Kind Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | _string_ | Name of the kind |

### changeFormat

**Change Arguments**

Template used to generate change lines in version files and changelog.
Custom values are created through [custom choices](/config/choices) and can be accessible through the Custom argument.

For example, if you had a custom value named `Issue` you can include that in your change using `{{.Custom.Issue}}`.

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | _string_ | What kind of change this is |
| **Body** | _string_ | Body message of the change |
| **Custom** | _map[string]string_ | Map of custom values if any exist |
