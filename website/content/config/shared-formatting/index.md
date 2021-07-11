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

### body
type: [BodyConfig](/config/shared-formatting/#body-config) | default: `empty` | optional

Customize the requirements of the body prompt.

### components
type: `[]string` | default: `empty` | optional

Components are an optional layer of changelogs suited for projects that want to
split change fragments by an area of the project.
An example could be splitting your changelogs by packages for a monorepo.

If no components are listed then the component prompt will be skipped and no
component header included.
By default no components are configured.

### kinds
type: `[]`[KindConfig](/config/kind-formatting) | default: `empty` | optional

Kinds are another optional layer of changelogs suited for specifying what type
of change we are making.
If configured, developers will be prompted to select a kind.
See [kind formatting](/config/kind-formatting) for how to further customize kinds.

The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed, and security.

### versionFormat
type: `string` | default: `""` | required

Template used to generate version headers in version files and changelog.

**Version Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Version** | `string` | Semantic version of the changes |
| **Time** | `time.Time` | Time of generated version |

### componentFormat
type: `string` | default: `""` | optional

Template used to generate component headers.
If format is empty no header will be included.
If components are disabled, the format is unused.

**Component Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Component** | `string` | Name of the component |

### kindFormat
type: `string` | default: `""` | optional

Template used to generate kind headers.
If format is empty no header will be included.
If kinds are disabled, the format is unused.

**Kind Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | `string` | Name of the kind |

### changeFormat
type: `string` | default: `""` | required

Template used to generate change lines in version files and changelog.
Custom values are created through [custom choices](/config/choices) and can be accessible through the Custom argument.

For example, if you had a custom value named `Issue` you can include that in your change using `{{.Custom.Issue}}`.

**Change Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Component** | `string` | What kind of component we are changing, only included if enabled |
| **Kind** | `string` | What kind of change this is, only included if enabled |
| **Body** | `string` | Body message of the change |
| **Time** | `time.Time` | Time of generated change |
| **Custom** | `map[string]string` | Map of custom values if any exist |

## Body Config
type: `struct`

### minLength
type: `int` | default: `nil` | optional

If specified the body must be at least this long.

### maxLength
type: `int` | default: `nil` | optional

If specified the body must be no more than this long.
