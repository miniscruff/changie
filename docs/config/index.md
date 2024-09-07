---
title: "Configuration"
hide:
  - navigation
---
### body [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L306) [:octicons-link-24:](#config-body) {: #config-body}
type: [BodyConfig](#bodyconfig-type) | optional

Options to customize the body prompt

### changeFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L289) [:octicons-link-24:](#config-changeformat) {: #config-changeformat}
type: `string` | optional | template type: [Change](#change-type)

Template used to generate change lines in version files and changelog.
Custom values are created through custom choices and can be accessible through the Custom argument.
??? Example
    ```yaml
    changeFormat: '* [#{{.Custom.Issue}}](https://github.com/miniscruff/changie/issues/{{.Custom.Issue}}) {{.Body}}'
    ```

### changelogPath [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L253) [:octicons-link-24:](#config-changelogpath) {: #config-changelogpath}
type: `string` | optional

Filepath for the generated changelog file.
Relative to project root.
ChangelogPath is not required if you are using projects.
??? Example
    ```yaml
    changelogPath: CHANGELOG.md
    ```

### changesDir [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L234) [:octicons-link-24:](#config-changesdir) {: #config-changesdir}
type: `string` | required

Directory for change files, header file and unreleased files.
Relative to project root.
??? Example
    ```yaml
    changesDir: .changes
    ```

### componentFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L280) [:octicons-link-24:](#config-componentformat) {: #config-componentformat}
type: `string` | optional | template type: [ComponentData](#componentdata-type)

Template used to generate component headers.
If format is empty no header will be included.
If components are disabled, the format is unused.

### components [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L317) [:octicons-link-24:](#config-components) {: #config-components}
type: `[]string` | optional

Components are an additional layer of organization suited for projects that want to split
change fragments by an area or tag of the project.
An example could be splitting your changelogs by packages for a monorepo.
If no components are listed then the component prompt will be skipped and no component header included.
By default no components are configured.
??? Example
    ```yaml
    components:
    - API
    - CLI
    - Frontend
    ```

### custom [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L344) [:octicons-link-24:](#config-custom) {: #config-custom}
type: [[]Custom](#custom-type) | optional

Custom choices allow you to ask for additional information when creating a new change fragment.
These custom choices are included in the [change custom](#change-custom) value.
??? Example
    ```yaml
    # github issue and author name
    custom:
    - key: Issue
      type: int
      minInt: 1
    - key: Author
      label: GitHub Name
      type: string
      minLength: 3
    ```

### envPrefix [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L371) [:octicons-link-24:](#config-envprefix) {: #config-envprefix}
type: `string` | optional

Prefix of environment variables to load for templates.
The prefix is removed from resulting key map.
??? Example
    ```yaml
    # if we have an environment variable like so:
    # export CHANGIE_PROJECT=changie
    # we can use that in our templates if we set the prefix
    envPrefix: "CHANGIE_"
    versionFormat: "New release for {{.Env.PROJECT}}"
    ```

### footerFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L304) [:octicons-link-24:](#config-footerformat) {: #config-footerformat}
type: `string` | optional | template type: [BatchData](#batchdata-type)

Template used to generate a version footer.
??? Example
    ```yaml
    # config yaml
    custom:
    - key: Author
      type: string
      minLength: 3
    footerFormat: |
      ### Contributors
      {{- range (customs .Changes "Author" | uniq) }}
      * [{{.}}](https://github.com/{{.}})
      {{- end}}
    ```

### fragmentFileFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L274) [:octicons-link-24:](#config-fragmentfileformat) {: #config-fragmentfileformat}
type: `string` | optional | template type: [Change](#change-type)

Customize the file name generated for new fragments.
The default uses the component and kind only if configured for your project.
The file is placed in the unreleased directory, so the full path is:

"{{.ChangesDir}}/{{.UnreleasedDir}}/{{.FragmentFileFormat}}.yaml"
??? Example
    ```yaml
    fragmentFileFormat: "{{.Kind}}-{{.Custom.Issue}}"
    ```

### headerFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L291) [:octicons-link-24:](#config-headerformat) {: #config-headerformat}
type: `string` | optional | template type: [BatchData](#batchdata-type)

Template used to generate a version header.

### headerPath [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L247) [:octicons-link-24:](#config-headerpath) {: #config-headerpath}
type: `string` | optional

Header content included at the top of the merged changelog.
A default header file is created when initializing that follows "Keep a Changelog".

Filepath for your changelog header file.
Relative to [changesDir](#config-changesdir).
??? Example
    ```yaml
    headerPath: header.tpl.md
    ```

### kindFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L284) [:octicons-link-24:](#config-kindformat) {: #config-kindformat}
type: `string` | optional | template type: [KindData](#kinddata-type)

Template used to generate kind headers.
If format is empty no header will be included.
If kinds are disabled, the format is unused.

### kinds [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L331) [:octicons-link-24:](#config-kinds) {: #config-kinds}
type: [[]KindConfig](#kindconfig-type) | optional

Kinds are another optional layer of changelogs suited for specifying what type of change we are
making.
If configured, developers will be prompted to select a kind.

The default list comes from keep a changelog and includes; added, changed, removed, deprecated, fixed, and security.
??? Example
    ```yaml
    kinds:
    - label: Added
    - label: Changed
    - label: Deprecated
    - label: Removed
    - label: Fixed
    - label: Security
    ```

### newlines [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L354) [:octicons-link-24:](#config-newlines) {: #config-newlines}
type: [NewlinesConfig](#newlinesconfig-type) | optional

Newline options allow you to add extra lines between elements written by changie.

### post [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L362) [:octicons-link-24:](#config-post) {: #config-post}
type: [[]PostProcessConfig](#postprocessconfig-type) | optional

Post process options when saving a new change fragment.
??? Example
    ```yaml
    # build a GitHub link from author choice
    post:
    - key: AuthorLink
      value: "https://github.com/{{.Custom.Author}}
    changeFormat: "* {{.Body}} by [{{.Custom.Author}}]({{.Custom.AuthorLink}})"
    ```

### projects [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L381) [:octicons-link-24:](#config-projects) {: #config-projects}
type: [[]ProjectConfig](#projectconfig-type) | optional

Projects allow you to specify child projects as part of a monorepo setup.
??? Example
    ```yaml
    projects:
      - label: UI
        key: ui
        changelog: ui/CHANGELOG.md
      - label: Email Sender
        key: email_sender
        changelog: services/email/CHANGELOG.md
    ```

### projectsVersionSeparator [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L386) [:octicons-link-24:](#config-projectsversionseparator) {: #config-projectsversionseparator}
type: `string` | optional

ProjectsVersionSeparator is used to determine the final version when using projects.
The result is: project key + projectVersionSeparator + latest/next version.
??? Example
    ```yaml
    projectsVersionSeparator: "_"
    ```

### replacements [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L352) [:octicons-link-24:](#config-replacements) {: #config-replacements}
type: [[]Replacement](#replacement-type) | optional

Replacements to run when merging a changelog.
??? Example
    ```yaml
    # nodejs package.json replacement
    replacements:
    - path: package.json
      find: '  "version": ".*",'
      replace: '  "version": "{{.VersionNoPrefix}}",'
    ```

### unreleasedDir [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L239) [:octicons-link-24:](#config-unreleaseddir) {: #config-unreleaseddir}
type: `string` | required

Directory for all unreleased change files.
Relative to [changesDir](#config-changesdir).
??? Example
    ```yaml
    unreleasedDir: unreleased
    ```

### versionExt [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L260) [:octicons-link-24:](#config-versionext) {: #config-versionext}
type: `string` | required

File extension for generated version files.
This should probably match your changelog path file.
Must not include the period.
??? Example
    ```yaml
    # for markdown changelogs
    versionExt: md
    ```

### versionFooterPath [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L266) [:octicons-link-24:](#config-versionfooterpath) {: #config-versionfooterpath}
type: `string` | optional

Filepath for your version footer file relative to [unreleasedDir](#config-unreleaseddir).
It is also possible to use the '--footer-path' parameter when using the [batch command](../cli/changie_batch.md).

### versionFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L276) [:octicons-link-24:](#config-versionformat) {: #config-versionformat}
type: `string` | optional | template type: [BatchData](#batchdata-type)

Template used to generate version headers.

### versionHeaderPath [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L263) [:octicons-link-24:](#config-versionheaderpath) {: #config-versionheaderpath}
type: `string` | optional

Filepath for your version header file relative to [unreleasedDir](#config-unreleaseddir).
It is also possible to use the '--header-path' parameter when using the [batch command](../cli/changie_batch.md).


---
## BatchData [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L12) [:octicons-link-24:](#batchdata-type) {: #batchdata-type}
Batch data is a common structure for templates when generating change fragments.

### Changes [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L32) [:octicons-link-24:](#batchdata-changes) {: #batchdata-changes}
type: [[]Change](#change-type) | optional

Changes included in the batch

### Env [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L35) [:octicons-link-24:](#batchdata-env) {: #batchdata-env}
type: map [ `string` ] `string` | optional

Env vars configured by the system.
See [envPrefix](#config-envprefix) for configuration.

### Major [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L22) [:octicons-link-24:](#batchdata-major) {: #batchdata-major}
type: `int` | optional

Major value of the version

### Metadata [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L30) [:octicons-link-24:](#batchdata-metadata) {: #batchdata-metadata}
type: `string` | optional

Metadata value of the version

### Minor [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L24) [:octicons-link-24:](#batchdata-minor) {: #batchdata-minor}
type: `int` | optional

Minor value of the version

### Patch [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L26) [:octicons-link-24:](#batchdata-patch) {: #batchdata-patch}
type: `int` | optional

Patch value of the version

### Prerelease [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L28) [:octicons-link-24:](#batchdata-prerelease) {: #batchdata-prerelease}
type: `string` | optional

Prerelease value of the version

### PreviousVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L20) [:octicons-link-24:](#batchdata-previousversion) {: #batchdata-previousversion}
type: `string` | optional

Previous released version

### Time [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L14) [:octicons-link-24:](#batchdata-time) {: #batchdata-time}
type: `Time` | optional

Time of the change

### Version [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L16) [:octicons-link-24:](#batchdata-version) {: #batchdata-version}
type: `string` | optional

Version of the change, will include "v" prefix if used

### VersionNoPrefix [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L18) [:octicons-link-24:](#batchdata-versionnoprefix) {: #batchdata-versionnoprefix}
type: `string` | optional

Version of the release without the "v" prefix if used


---
## BodyConfig [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L87) [:octicons-link-24:](#bodyconfig-type) {: #bodyconfig-type}
Body config allows you to customize the default body prompt

### maxLength [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L91) [:octicons-link-24:](#bodyconfig-maxlength) {: #bodyconfig-maxlength}
type: `int64` | optional

Max length specifies the maximum body length

### minLength [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L89) [:octicons-link-24:](#bodyconfig-minlength) {: #bodyconfig-minlength}
type: `int64` | optional

Min length specifies the minimum body length

### block [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L93) [:octicons-link-24:](#bodyconfig-block) {: #bodyconfig-block}
type: `bool` | optional

Block allows multiline text inputs for body messages


---
## Change [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L50) [:octicons-link-24:](#change-type) {: #change-type}
Change represents an atomic change to a project.

### body [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L58) [:octicons-link-24:](#change-body) {: #change-body}
type: `string` | optional

Body message of our change, if one was provided.

### component [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L54) [:octicons-link-24:](#change-component) {: #change-component}
type: `string` | optional

Component of our change, if one was provided.

### custom [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L68) [:octicons-link-24:](#change-custom) {: #change-custom}
type: map [ `string` ] `string` | optional

Custom values corresponding to our options where each key-value pair is the key of the custom option
and value the one provided in the change.
??? Example
    ```yaml
    custom:
    - key: Issue
      type: int
    changeFormat: "{{.Body}} from #{{.Custom.Issue}}"
    ```

### env [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L74) [:octicons-link-24:](#change-env) {: #change-env}
type: map [ `string` ] `string` | optional

Env vars configured by the system.
This is not written in change fragments but instead loaded by the system and accessible for templates.
For example if you want to use an env var in [change format](#config-changeformat) you can,
but env vars configured when executing `changie new` will not be saved.
See [envPrefix](#config-envprefix) for configuration.

### filename [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L76) [:octicons-link-24:](#change-filename) {: #change-filename}
type: `string` | optional

Filename the change was saved to.

### kind [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L56) [:octicons-link-24:](#change-kind) {: #change-kind}
type: `string` | optional

Kind of our change, if one was provided.

### project [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L52) [:octicons-link-24:](#change-project) {: #change-project}
type: `string` | optional

Project of our change, if one was provided.

### time [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/change.go#L60) [:octicons-link-24:](#change-time) {: #change-time}
type: `Time` | required

When our change was made.


---
## ComponentData [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L39) [:octicons-link-24:](#componentdata-type) {: #componentdata-type}
Component data stores data related to writing component headers.

### Component [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L41) [:octicons-link-24:](#componentdata-component) {: #componentdata-component}
type: `string` | required

Name of the component

### Env [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L44) [:octicons-link-24:](#componentdata-env) {: #componentdata-env}
type: map [ `string` ] `string` | optional

Env vars configured by the system.
See [envPrefix](#config-envprefix) for configuration.


---
## Custom [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L51) [:octicons-link-24:](#custom-type) {: #custom-type}
Custom defines a custom choice that is asked when using 'changie new'.
The result is an additional custom value in the change file for including in the change line.

A simple one could be the issue number or authors github name.

??? Example
    ```yaml
    - key: Author
        label: GitHub Name
        type: string
        minLength: 3
    ```
### enumOptions [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L101) [:octicons-link-24:](#custom-enumoptions) {: #custom-enumoptions}
type: `[]string` | optional

When using the enum type, you must also specify what possible options to allow.
Users will be given a selection list to select the value they want.

### key [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L56) [:octicons-link-24:](#custom-key) {: #custom-key}
type: `string` | required

Value used as the key in the custom map for the change format.
This should only contain alpha numeric characters, usually starting with a capital.
??? Example
    ```yaml
    key: Issue
    ```

### label [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L90) [:octicons-link-24:](#custom-label) {: #custom-label}
type: `string` | optional

Description used in the prompt when asking for the choice.
If empty key is used instead.
??? Example
    ```yaml
    label: GitHub Username
    ```

### maxInt [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L94) [:octicons-link-24:](#custom-maxint) {: #custom-maxint}
type: `int64` | optional

If specified the input value must be less than or equal to maxInt.

### maxLength [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L98) [:octicons-link-24:](#custom-maxlength) {: #custom-maxlength}
type: `int64` | optional

If specified string input must be no more than this long

### minInt [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L92) [:octicons-link-24:](#custom-minint) {: #custom-minint}
type: `int64` | optional

If specified the input value must be greater than or equal to minInt.

### minLength [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L96) [:octicons-link-24:](#custom-minlength) {: #custom-minlength}
type: `int64` | optional

If specified the string input must be at least this long

### optional [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L85) [:octicons-link-24:](#custom-optional) {: #custom-optional}
type: `bool` | optional

If true, an empty value will not fail validation.
The optional check is handled before min so you can specify that the value is optional but if it
is used it must have a minimum length or value depending on type.

When building templates that allow for optional values you can compare the custom choice to an
empty string to check for a value or empty.

??? Example
    ```yaml
    custom:
    - key: TicketNumber
      type: int
      optional: true
    changeFormat: >-
    {{- if not (eq .Custom.TicketNumber "")}}
    PROJ-{{.Custom.TicketNumber}}
    {{- end}}
    {{.Body}}
    ```

### type [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L66) [:octicons-link-24:](#custom-type) {: #custom-type}
type: [CustomType](#customtype-type) | required

Specifies the type of choice which changes the prompt.

| value | description | options
| -- | -- | -- |
string | Freeform text | [minLength](#custom-minlength) and [maxLength](#custom-maxlength)
block | Multiline text | [minLength](#custom-minlength) and [maxLength](#custom-maxlength)
int | Whole numbers | [minInt](#custom-minint) and [maxInt](#custom-maxint)
enum | Limited set of strings | [enumOptions](#custom-enumoptions) is used to specify values


---
## CustomType [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/custom.go#L20) [:octicons-link-24:](#customtype-type) {: #customtype-type}
CustomType determines the possible custom choice types.
Current values are: `string`, `block`, `int` and `enum`.


---
## KindConfig [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L40) [:octicons-link-24:](#kindconfig-type) {: #kindconfig-type}
Kind config allows you to customize the options depending on what kind was selected.

### additionalChoices [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L59) [:octicons-link-24:](#kindconfig-additionalchoices) {: #kindconfig-additionalchoices}
type: [[]Custom](#custom-type) | optional

Additional choices allows adding choices per kind

### auto [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L74) [:octicons-link-24:](#kindconfig-auto) {: #kindconfig-auto}
type: `string` | optional

Auto determines what value to bump when using `batch auto` or `next auto`.
Possible values are major, minor, patch or none and the highest one is used if
multiple changes are found. none will not bump the version.
Only none changes is not a valid bump and will fail to batch.
??? Example
    ```yaml
    auto: minor
    ```

### changeFormat [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L57) [:octicons-link-24:](#kindconfig-changeformat) {: #kindconfig-changeformat}
type: `string` | optional

Change format will override the root change format when building changes specific to this kind.
??? Example
    ```yaml
    changeFormat: 'Breaking: {{.Custom.Body}}
    ```

### format [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L53) [:octicons-link-24:](#kindconfig-format) {: #kindconfig-format}
type: `string` | optional

Format will override the root kind format when building the kind header.
??? Example
    ```yaml
    format: '### {{.Kind}} **Breaking Changes**'
    ```

### key [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L45) [:octicons-link-24:](#kindconfig-key) {: #kindconfig-key}
type: `string` | optional

Key is the value used for lookups and file names for kinds.
By default it will use label if no key is provided.
??? Example
    ```yaml
    key: feature
    ```

### label [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L49) [:octicons-link-24:](#kindconfig-label) {: #kindconfig-label}
type: `string` | required

Label is the value used in the prompt when selecting a kind.
??? Example
    ```yaml
    label: Feature
    ```

### post [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L61) [:octicons-link-24:](#kindconfig-post) {: #kindconfig-post}
type: [[]PostProcessConfig](#postprocessconfig-type) | optional

Post process options when saving a new change fragment specific to this kind.

### skipBody [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L65) [:octicons-link-24:](#kindconfig-skipbody) {: #kindconfig-skipbody}
type: `bool` | optional

Skip body allows skipping the parent body prompt.

### skipGlobalChoices [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L63) [:octicons-link-24:](#kindconfig-skipglobalchoices) {: #kindconfig-skipglobalchoices}
type: `bool` | optional

Skip global choices allows skipping the parent choices options.

### skipGlobalPost [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L67) [:octicons-link-24:](#kindconfig-skipglobalpost) {: #kindconfig-skipglobalpost}
type: `bool` | optional

Skip global post allows skipping the parent post processing.


---
## KindData [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L48) [:octicons-link-24:](#kinddata-type) {: #kinddata-type}
Kind data stores data related to writing kind headers.

### Env [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L53) [:octicons-link-24:](#kinddata-env) {: #kinddata-env}
type: map [ `string` ] `string` | optional

Env vars configured by the system.
See [envPrefix](#config-envprefix) for configuration.

### Kind [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L50) [:octicons-link-24:](#kinddata-kind) {: #kinddata-kind}
type: `string` | required

Name of the kind


---
## NewlinesConfig [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L122) [:octicons-link-24:](#newlinesconfig-type) {: #newlinesconfig-type}
Configuration options for newlines before and after different elements.

### afterChange [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L124) [:octicons-link-24:](#newlinesconfig-afterchange) {: #newlinesconfig-afterchange}
type: `int` | optional

Add newlines after change fragment

### afterChangelogHeader [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L126) [:octicons-link-24:](#newlinesconfig-afterchangelogheader) {: #newlinesconfig-afterchangelogheader}
type: `int` | optional

Add newlines after the header file in the merged changelog

### afterChangelogVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L128) [:octicons-link-24:](#newlinesconfig-afterchangelogversion) {: #newlinesconfig-afterchangelogversion}
type: `int` | optional

Add newlines after adding a version to the changelog

### afterComponent [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L130) [:octicons-link-24:](#newlinesconfig-aftercomponent) {: #newlinesconfig-aftercomponent}
type: `int` | optional

Add newlines after component

### afterFooterFile [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L132) [:octicons-link-24:](#newlinesconfig-afterfooterfile) {: #newlinesconfig-afterfooterfile}
type: `int` | optional

Add newlines after footer file

### afterFooter [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L134) [:octicons-link-24:](#newlinesconfig-afterfooter) {: #newlinesconfig-afterfooter}
type: `int` | optional

Add newlines after footer template

### afterHeaderFile [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L136) [:octicons-link-24:](#newlinesconfig-afterheaderfile) {: #newlinesconfig-afterheaderfile}
type: `int` | optional

Add newlines after header file

### afterHeaderTemplate [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L138) [:octicons-link-24:](#newlinesconfig-afterheadertemplate) {: #newlinesconfig-afterheadertemplate}
type: `int` | optional

Add newlines after header template

### afterKind [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L140) [:octicons-link-24:](#newlinesconfig-afterkind) {: #newlinesconfig-afterkind}
type: `int` | optional

Add newlines after kind

### afterVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L142) [:octicons-link-24:](#newlinesconfig-afterversion) {: #newlinesconfig-afterversion}
type: `int` | optional

Add newlines after version

### beforeChange [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L144) [:octicons-link-24:](#newlinesconfig-beforechange) {: #newlinesconfig-beforechange}
type: `int` | optional

Add newlines before change fragment

### beforeChangelogVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L146) [:octicons-link-24:](#newlinesconfig-beforechangelogversion) {: #newlinesconfig-beforechangelogversion}
type: `int` | optional

Add newlines before adding a version to the changelog

### beforeComponent [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L148) [:octicons-link-24:](#newlinesconfig-beforecomponent) {: #newlinesconfig-beforecomponent}
type: `int` | optional

Add newlines before component

### beforeFooterFile [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L150) [:octicons-link-24:](#newlinesconfig-beforefooterfile) {: #newlinesconfig-beforefooterfile}
type: `int` | optional

Add newlines before footer file

### beforeFooterTemplate [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L152) [:octicons-link-24:](#newlinesconfig-beforefootertemplate) {: #newlinesconfig-beforefootertemplate}
type: `int` | optional

Add newlines before footer template

### beforeHeaderFile [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L154) [:octicons-link-24:](#newlinesconfig-beforeheaderfile) {: #newlinesconfig-beforeheaderfile}
type: `int` | optional

Add newlines before header file

### beforeHeaderTemplate [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L156) [:octicons-link-24:](#newlinesconfig-beforeheadertemplate) {: #newlinesconfig-beforeheadertemplate}
type: `int` | optional

Add newlines before header template

### beforeKind [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L158) [:octicons-link-24:](#newlinesconfig-beforekind) {: #newlinesconfig-beforekind}
type: `int` | optional

Add newlines before kind

### beforeVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L160) [:octicons-link-24:](#newlinesconfig-beforeversion) {: #newlinesconfig-beforeversion}
type: `int` | optional

Add newlines before version

### endOfVersion [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L162) [:octicons-link-24:](#newlinesconfig-endofversion) {: #newlinesconfig-endofversion}
type: `int` | optional

Add newlines at the end of the version file


---
## PostProcessConfig [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L169) [:octicons-link-24:](#postprocessconfig-type) {: #postprocessconfig-type}
PostProcessConfig allows adding additional custom values to a change fragment
after all the other inputs are complete.
This will add additional keys to the `custom` section of the fragment.
If the key already exists as part of a custom choice the value will be overridden.

### key [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L171) [:octicons-link-24:](#postprocessconfig-key) {: #postprocessconfig-key}
type: `string` | optional

Key to save the custom value with

### value [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L173) [:octicons-link-24:](#postprocessconfig-value) {: #postprocessconfig-value}
type: `string` | optional | template type: [Change](#change-type)

Value of the custom value as a go template


---
## ProjectConfig [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L186) [:octicons-link-24:](#projectconfig-type) {: #projectconfig-type}
ProjectConfig extends changie to support multiple changelog files for different projects
inside one repository.

??? Example
    ```yaml
    projects:
      - label: UI
        key: ui
        changelog: ui/CHANGELOG.md
      - label: Email Sender
        key: email_sender
        changelog: services/email/CHANGELOG.md
    ```
### changelog [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L198) [:octicons-link-24:](#projectconfig-changelog) {: #projectconfig-changelog}
type: `string` | optional

ChangelogPath is the path to the changelog for this project.
??? Example
    ```yaml
    changelog: src/frontend/CHANGELOG.md
    ```

### key [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L194) [:octicons-link-24:](#projectconfig-key) {: #projectconfig-key}
type: `string` | optional

Key is the value used for unreleased and version output paths.
??? Example
    ```yaml
    key: frontend
    ```

### label [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L190) [:octicons-link-24:](#projectconfig-label) {: #projectconfig-label}
type: `string` | optional

Label is the value used in the prompt when selecting a project.
??? Example
    ```yaml
    label: Frontend
    ```

### replacements [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/config.go#L206) [:octicons-link-24:](#projectconfig-replacements) {: #projectconfig-replacements}
type: [[]Replacement](#replacement-type) | optional

Replacements to run when merging a changelog for our project.
??? Example
    ```yaml
    # nodejs package.json replacement
    replacements:
    - path: ui/package.json
      find: '  "version": ".*",'
      replace: '  "version": "{{.VersionNoPrefix}}",'
    ```


---
## ReplaceData [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L12) [:octicons-link-24:](#replacedata-type) {: #replacedata-type}
Template data used for replacing version values.

### Major [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L18) [:octicons-link-24:](#replacedata-major) {: #replacedata-major}
type: `int` | optional

Major value of the version

### Metadata [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L26) [:octicons-link-24:](#replacedata-metadata) {: #replacedata-metadata}
type: `string` | optional

Metadata value of the version

### Minor [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L20) [:octicons-link-24:](#replacedata-minor) {: #replacedata-minor}
type: `int` | optional

Minor value of the version

### Patch [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L22) [:octicons-link-24:](#replacedata-patch) {: #replacedata-patch}
type: `int` | optional

Patch value of the version

### Prerelease [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L24) [:octicons-link-24:](#replacedata-prerelease) {: #replacedata-prerelease}
type: `string` | optional

Prerelease value of the version

### Version [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L14) [:octicons-link-24:](#replacedata-version) {: #replacedata-version}
type: `string` | optional

Version of the release, will include "v" prefix if used

### VersionNoPrefix [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L16) [:octicons-link-24:](#replacedata-versionnoprefix) {: #replacedata-versionnoprefix}
type: `string` | optional

Version of the release without the "v" prefix if used


---
## Replacement [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L40) [:octicons-link-24:](#replacement-type) {: #replacement-type}
Replacement handles the finding and replacing values when merging the changelog.
This can be used to keep version strings in-sync when preparing a release without having to
manually update them.
This works similar to the find and replace from IDE tools but also includes the file path of the
file.

??? Example
    ```yaml
    # NodeJS package.json
    replacements:
      - path: package.json
        find: '  "version": ".*",'
        replace: '  "version": "{{.VersionNoPrefix}}",'
    ```
### find [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L44) [:octicons-link-24:](#replacement-find) {: #replacement-find}
type: `string` | required

Regular expression to search for in the file.

### flags [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L53) [:octicons-link-24:](#replacement-flags) {: #replacement-flags}
type: `string` | optional

Optional regular expression mode flags.
Defaults to the m flag for multiline such that ^ and $ will match the start and end of each line
and not just the start and end of the string.

For more details on regular expression flags in Go view the
[regexp/syntax](https://pkg.go.dev/regexp/syntax).

### path [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L42) [:octicons-link-24:](#replacement-path) {: #replacement-path}
type: `string` | required

Path of the file to find and replace in.

### replace [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/replacement.go#L46) [:octicons-link-24:](#replacement-replace) {: #replacement-replace}
type: `string` | required | template type: [ReplaceData](#replacedata-type)

Template string to replace the line with.


---
## TemplateCache [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L66) [:octicons-link-24:](#templatecache-type) {: #templatecache-type}
Template cache handles running all the templates for change fragments.
Included options include the default [go template](https://golang.org/pkg/text/template/)
and [sprig functions](https://masterminds.github.io/sprig/) for formatting.
Additionally, custom template functions are listed below for working with changes.

??? Example
    ```yaml
    format: |
    ### Contributors
    {{- range (customs .Changes "Author" | uniq) }}
    * [{{.}}](https://github.com/{{.}})
    {{- end}}
    ```
### bodies [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L159) [:octicons-link-24:](#templatecache-bodies) {: #templatecache-bodies}


Bodies will return all the bodies from the provided changes.
??? Example
    ```yaml
    format: "{{ bodies .Changes }} bodies"
    ```

### components [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L133) [:octicons-link-24:](#templatecache-components) {: #templatecache-components}


Components will return all the components from the provided changes.
??? Example
    ```yaml
    format: "{{components .Changes }} components"
    ```

### count [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L118) [:octicons-link-24:](#templatecache-count) {: #templatecache-count}


Count will return the number of occurrences of a string in a slice.
??? Example
    ```yaml
    format: "{{ kinds .Changes | count \"added\" }} kinds"
    ```

### customs [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L186) [:octicons-link-24:](#templatecache-customs) {: #templatecache-customs}


Customs will return all the values from the custom map by a key.
If a key is missing from a change, it will be an empty string.
??? Example
    ```yaml
    format: "{{ customs .Changes \"Author\" }} authors"
    ```

### kinds [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L146) [:octicons-link-24:](#templatecache-kinds) {: #templatecache-kinds}


Kinds will return all the kindsi from the provided changes.
??? Example
    ```yaml
    format: "{{ kinds .Changes }} kinds"
    ```

### times [:octicons-code-24:](https://github.com/miniscruff/changie/blob/<< current_version >>/core/templatecache.go#L172) [:octicons-link-24:](#templatecache-times) {: #templatecache-times}


Times will return all the times from the provided changes.
??? Example
    ```yaml
    format: "{{ times .Changes }} times"
    ```


---
