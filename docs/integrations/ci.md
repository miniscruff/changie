When creating changie fragments as part of your team workflow, you may have
suggestions or comments about the kind, body or custom fields that are
changed outside of the changie tool.
As changie fragments are plain yaml files there is nothing wrong with editing
these after creating the fragment.
However, it is possible to create invalid fragments when doing so,
one such example is if you typo a kind or invalid custom prompt answer.

One way to prevent this issue from causing later problems is to run
changie as part of your CI tests.
Below is an example if you are using the github action.

```yaml
- name: Validate changie fragment is valid
  uses: miniscruff/changie-action@VERSION # view action repo for latest version
  with:
    version: latest # use the latest changie version
    # dry run may not be required as you likely aren't
    # committing the changes anyway, but it will print
    # to stdout this way
    args: batch major --dry-run 
```
