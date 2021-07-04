---
title: "Shared Formatting"
date: 2021-01-31T14:13:51-08:00
draft: false
weight: 3
summary: Customize how version and changelog files are generated.
---

Changie utilizes [go template](https://golang.org/pkg/text/template/) for formatting version, kind, change and replacement lines.
Additional fields can be added to change lines by adding [custom choices](/config/choices).
You can also customize certain formatting options per kind using the [kind formatting](/config/kind-formatting).

> Due to the ordering of commands you must add custom choices before
> you added any change files in order to use the custom values in your format.

When batching changes into a version, changes are sorted by:
1. Component, if enabled, sorted by index in components config
1. Kind, if enabled, sorted by index in kinds config
1. Time sorted newest first

### components
type: _[]string_

Components are an optional layer of changelogs suited for projects that want to
split change fragments by an area of the project.
An example could be splitting your changelogs by packages for a monorepo.

If no components are listed then the component prompt will be skipped and no
component header included.
By default no components are configured.

### kinds
type: _[][KindConfig](/config/kind-formatting)_

Kinds are another optional layer of changelogs suited for specifying what type
of change we are making.
If configured, developers will be prompted to select a kind.
See [kind formatting](/config/kind-formatting) for how to further customize kinds.

The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed, and security.

### versionFormat
type: _string_

Template used to generate version headers in version files and changelog.

**Version Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Version** | _string_ | Semantic version of the changes |
| **Time** | _time.Time_ | Time of generated version |

### componentFormat
type: _string_

Template used to generate component headers.
If format is empty no header will be included.
If components are disabled, the format is unused.

**Component Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Component** | _string_ | Name of the component |

### kindFormat
type: _string_

Template used to generate kind headers.
If format is empty no header will be included.
If kinds are disabled, the format is unused.

**Kind Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | _string_ | Name of the kind |

### changeFormat
type: _string_

Template used to generate change lines in version files and changelog.
Custom values are created through [custom choices](/config/choices) and can be accessible through the Custom argument.

For example, if you had a custom value named `Issue` you can include that in your change using `{{.Custom.Issue}}`.

**Change Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Component** | _string_ | What kind of component we are changing, only included if enabled |
| **Kind** | _string_ | What kind of change this is, only included if enabled |
| **Body** | _string_ | Body message of the change |
| **Time** | _time.Time_ | Time of generated change |
| **Custom** | _map[string]string_ | Map of custom values if any exist |
