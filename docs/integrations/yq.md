If you are familiar with [jq](https://github.com/jqlang/jq) you may be interested to know that
there is a similar tool specifically built for yaml files named
[yq](https://mikefarah.gitbook.io/yq).

This tool comes in handy if you are every looking to combine a yaml configuration file
with any sort of script or tooling.
For instance, if you ever want to pull your changie kind keys out into a list and process.

Below is a small sample of commands you can run.

## Kind keys

```sh
yq '.kinds[].key' .changie.yaml
```

Output
```
added
changed
deprecated
removed
fixed
security
```

## Kind labels

```sh
yq '.kinds[].label' .changie.yaml
```

Output
```
✨ Added
🔥 Changed
⚰️ Deprecated
🗑️ Removed
🪲 Fixed
🦺 Security
```

## Custom keys

```sh
yq '.custom[].key' .changie.yaml
```

Output
```
Issue
```
