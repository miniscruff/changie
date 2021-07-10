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

### label
type: _string_

Label is the only required element for kinds.
This value is shown to users as part of the Kind selection prompt.
It is also the value saved in change files.

### header
type: _string_

Header allows you to override the header output.
Overrides the shared [kind format](/config/shared-formatting#kindformat) when specified.

**Header Arguments**

| Field | Type | Description |
| --- | --- | --- |
| **Kind** | _string_ | Kind label value |

### changeFormat
type: _string_

Change format allows you to override the output of a change for each kind.
Overrides the shared [change format](/config/shared-formatting#changeformat) when specified.
