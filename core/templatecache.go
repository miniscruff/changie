package core

import (
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

// Batch data is a common structure for templates when generating change fragments.
type BatchData struct {
	// Time of the change
	Time time.Time
	// Version of the change, will include "v" prefix if used
	Version string
	// Version of the release without the "v" prefix if used
	VersionNoPrefix string
	// Previous released version
	PreviousVersion string
	// Major value of the version
	Major int
	// Minor value of the version
	Minor int
	// Patch value of the version
	Patch int
	// Prerelease value of the version
	Prerelease string
	// Metadata value of the version
	Metadata string
	// Changes included in the batch
	Changes []Change
	// Env vars configured by the system.
	// See [envPrefix](#config-envprefix) for configuration.
	Env map[string]string
}

// Component data stores data related to writing component headers.
type ComponentData struct {
	// Name of the component
	Component string `required:"true"`
	// Env vars configured by the system.
	// See [envPrefix](#config-envprefix) for configuration.
	Env map[string]string
}

// Kind data stores data related to writing kind headers.
type KindData struct {
	// Name of the kind
	Kind string `required:"true"`
	// Env vars configured by the system.
	// See [envPrefix](#config-envprefix) for configuration.
	Env map[string]string
}

// Template cache handles running all the templates for change fragments.
// Included options include the default [go template](https://golang.org/pkg/text/template/)
// and [sprig functions](https://masterminds.github.io/sprig/) for formatting.
// Additionally, custom template functions are listed below for working with changes.
// example: yaml
// format: |
// ### Contributors
// {{- range (customs .Changes "Author" | uniq) }}
// * [{{.}}](https://github.com/{{.}})
// {{- end}}
type TemplateCache struct {
	cache map[string]*template.Template
}

func NewTemplateCache() *TemplateCache {
	return &TemplateCache{
		cache: make(map[string]*template.Template),
	}
}

func (tc *TemplateCache) Load(text string) (*template.Template, error) {
	cachedTemplate, ok := tc.cache[text]
	if ok {
		return cachedTemplate, nil
	}

	templ, err := template.New(text).Funcs(tc.buildFuncMap()).Parse(text)
	if err != nil {
		// do not save our template if it had an error
		return nil, err
	}

	tc.cache[text] = templ

	return templ, err
}

func (tc *TemplateCache) Execute(text string, wr io.Writer, data any) error {
	templ, err := tc.Load(text)
	if err != nil {
		return err
	}

	return templ.Execute(wr, data)
}

func (tc *TemplateCache) ExecuteString(text string, data any) (string, error) {
	sb := strings.Builder{}

	templ, err := tc.Load(text)
	if err != nil {
		return "", err
	}

	err = templ.Execute(&sb, data)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

func (tc *TemplateCache) buildFuncMap() map[string]any {
	funcs := sprig.TxtFuncMap()

	funcs["count"] = tc.Count
	funcs["components"] = tc.Components
	funcs["kinds"] = tc.Kinds
	funcs["bodies"] = tc.Bodies
	funcs["times"] = tc.Times
	funcs["customs"] = tc.Customs

	return funcs
}

// Count will return the number of occurrences of a string in a slice.
// example: yaml
// format: "{{ kinds .Changes | count \"added\" }} kinds"
func (tc *TemplateCache) Count(value string, items []string) (int, error) {
	count := 0

	for _, i := range items {
		if i == value {
			count++
		}
	}

	return count, nil
}

// Components will return all the components from the provided changes.
// example: yaml
// format: "{{components .Changes }} components"
func (tc *TemplateCache) Components(changes []Change) ([]string, error) {
	comps := make([]string, len(changes))

	for i, c := range changes {
		comps[i] = c.Component
	}

	return comps, nil
}

// Kinds will return all the kindsi from the provided changes.
// example: yaml
// format: "{{ kinds .Changes }} kinds"
func (tc *TemplateCache) Kinds(changes []Change) ([]string, error) {
	kinds := make([]string, len(changes))

	for i, c := range changes {
		kinds[i] = c.Kind
	}

	return kinds, nil
}

// Bodies will return all the bodies from the provided changes.
// example: yaml
// format: "{{ bodies .Changes }} bodies"
func (tc *TemplateCache) Bodies(changes []Change) ([]string, error) {
	bodies := make([]string, len(changes))

	for i, c := range changes {
		bodies[i] = c.Body
	}

	return bodies, nil
}

// Times will return all the times from the provided changes.
// example: yaml
// format: "{{ times .Changes }} times"
func (tc *TemplateCache) Times(changes []Change) ([]time.Time, error) {
	times := make([]time.Time, len(changes))

	for i, c := range changes {
		times[i] = c.Time
	}

	return times, nil
}

// Customs will return all the values from the custom map by a key.
// If a key is missing from a change, it will be an empty string.
// example: yaml
// format: "{{ customs .Changes \"Author\" }} authors"
func (tc *TemplateCache) Customs(changes []Change, key string) ([]string, error) {
	values := make([]string, len(changes))

	for i, c := range changes {
		values[i] = c.Custom[key]
	}

	return values, nil
}
