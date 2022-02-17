---
title: "Version header and footers"
date: 2021-01-31T14:13:35-08:00
draft: false
weight: 2
summary: Headers and footers to version change files
---

Changie utilizes [go template](https://golang.org/pkg/text/template/) and
[sprig](https://masterminds.github.io/sprig/) functions for formatting.
In addition to that a few template functions are available when working with changes.

When batching changes into a version, the headers and footers are placed as such:
1. Header file
1. Header template
1. All changes
1. Footer template
1. Footer file

All elements are optional and will be added together if all are provided.

## Configuration

### versionHeaderPath
type: `string` | default: `""` | optional

When batching change fragments into a single version file you can include a version header
paragraph by creating a file at the path specified.
This file must be placed in your unreleased directory and must not be a `.yaml` file.
This is empty by default and is considered optional and will be skipped if the file is not found.

Filepath for your version header file.
Relative to `unreleasedDir`.
It is also possible to use the `--header-path` parameter when using the [batch command](/cli/changie_batch).
It is possible to use both a CLI parameter and config file.

### versionFooterPath
type: `string` | default: `""` | optional

When batching change fragments into a single version file you can include a version footer
paragraph by creating a file at the path specified.
This file must be placed in your unreleased directory and must not be a `.yaml` file.
This is empty by default and is considered optional and will be skipped if the file is not found.

Filepath for your version header file.
Relative to `unreleasedDir`.
It is also possible to use the `--footer-path` parameter when using the [batch command](/cli/changie_batch).
It is possible to use both a CLI parameter and config file.

## Template Functions
Below are all the custom template functions available for headers and footers.

For functions that return a slice you can use the [range](https://pkg.go.dev/text/template#hdr-Actions)
action to loop through the values.

### count
returns: `int` | requires: value and items

Get the number of occurances of value in items.
Value is a `string`, items is a `[]string`.
Can be used in a few functions below.

Example: `Owner made {{customs .Changes "Author" | count "owner"}} changes`

### components
returns: `[]string` | requires: changes

Get all the components for our changes.

Example: `{{components .Changes | uniq | len}} components updated`

### kinds
returns: `[]string` | requires: changes

Get all the kinds for our changes.

Example: `{{kinds .Changes | count "Fixed"}} fixed issues this release`

### bodies
returns: `[]string` | requires: changes

Get all the bodies for our changes.

Example: `newest change {{bodies .Changes | first}}`

### times
returns: `[]time.Time` | requires: changes

Get all the timestamps for the changes.

Example: `oldest change {{times .Changes | last}}`

### customs
returns: `[]string` | requires: changes and key

Get all values with a key in the changes.
Will return an error if that key is missing from any change.

Example: `{{range (customs .Changes "Issue")}}* {{.}}{{end}}`

## Examples

### Contributors Footer
Display all unique contributors from a custom choice.

```yaml
# config yaml
custom:
- key: Author
  type: string
  minLength: 3
versionFooterFormat: |
  ### Components
  {{- range (customs .Changes "Author" | uniq) }}
  * [{{.}}](https://github.com/{{.}})
  {{- end}}`
```
