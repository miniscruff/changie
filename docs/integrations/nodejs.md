During the release process it is likely you want to update your package json to
use the new version.
This can be handled automatically by Changie using the
[replacements configuration](../config/index.md#config-replacements), which occur when you
run `changie merge`.

Below is how you could configure it for NodeJS.

```yaml
replacements:
  - path: package.json
    find: '  "version": ".*",'
    replace: '  "version": "{{.VersionNoPrefix}}",'
```

Note: If you do not use any `v` prefixes on your versions ( `1.3.4` instead of `v1.3.4` )
you can just use `{{.Version}}`.
