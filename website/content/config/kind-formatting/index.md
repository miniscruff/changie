---
title: "Kind Formatting"
date: 2021-01-31T14:13:51-08:00
draft: false
weight: 4
summary: Customize how kinds are formatted.
---

Configuration can be split per kind of change to allow more flexibility.
Most configurations are optional and will default to the
[shared formatting](/config/shared-formatting) values when omitted.

Example of a two line breaking change:

```yaml
kinds:
  - label: Breaking
    header: "## :warning: Breaking"
    skipGlobalChoices: true
    changeFormat: >-
      * {{.Reason}}
      {{.Body}}
    additionalChoices:
    - key: Reason
      label: Reason for breaking change
      type: string
```

### label
type: `string` | default: `""` | required

Label is the only required element for kinds.
This value is shown to users as part of the Kind selection prompt.
It is also the value saved in change files.

### format
type: `string` | default: `""` | optional

Format allows you to override the header output.
Overrides the shared [kind format](/config/shared-formatting#kindformat) when specified.

**Format Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | _string_ | Kind label value |

### changeFormat
type: `string` | default: `""` | optional

Change format allows you to override the output of a change for each kind.
Overrides the shared [change format](/config/shared-formatting#changeformat)
when specified.

### skipBody
type: `bool` | default: `false` | optional

Whether or not our kind will skip the default body prompt.

### skipGlobalChoices
type: `bool` | default: `false` | optional

Whether or not our kind will skip the global choices configuration.
This allows you to create kinds that have a specific set of custom choices separate
from the other kinds.

### additionalChoices
type: `[]`[Choice](/config/choices) | default: `empty` | optional

Additional choices is a list of custom choices identical to the global
[choices](/config/choices#choice) but for only this one kind.
These choices will be asked after the global choices.
You will need to create a custom change format to use the new choices.
