---
title: "Choices"
date: 2021-01-31T14:14:04-08:00
draft: false
weight: 4
summary: Add additional custom choices to provide extra details to generated changelog.
---

### custom
type: _[]Choice_

Changie will only ask developers for a kind and body message by default.
You can add more metadata to change files by including custom choices.
These choices add key value pairs to the change format, see [changeFormat](/config/formatting/#changeFormat) for more details.

## Choice
type: _struct_

Choice defines a custom choice that is asked when using `changie new`.
The result is an additional custom value in the change file for including in the change line.

A simple one could be the issue number or authors github name.
See [type](#type) for possible types.

### key
type: _string_

Value used as the key in the custom map for the change format.
This should only contain alpha numeric characters, usually starting with a capital.

Example: `Issue`

### label
type: _string_

Description used in the prompt when asking for the choice.
If empty `key` is used instead.

### type
type: _string_

Specifies the type of choice which changes the prompt.

| type | description | options |
| --- | --- | --- |
| **string** | Freeform text option | No other options |
| **int** | Whole numbers | Min and max int values to limit value |
| **enum** | Limited set of strings | enumOptions is used to specify possible values |

### minInt
type: _int_

If specified the input value must be greater than or equal to minInt.

### maxInt
type: _int_

If specified the input value must be less than or equal to maxInt.

### enumOptions
type: _[]string_

When using the enum type, you must also specify what possible options to allow.
Developers will be given a selection list to select the value they choose.
