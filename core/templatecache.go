package core

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

// probably want to move this -miniscruff
type BatchData struct {
	Time            time.Time
	Version         string
	PreviousVersion string
	Changes         []Change
}

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

	templ, err := template.New(text).Funcs(funcMap()).Parse(text)
	tc.cache[text] = templ

	return templ, err
}

func (tc *TemplateCache) Execute(text string, wr io.Writer, data interface{}) error {
	templ, err := tc.Load(text)
	if err != nil {
		return err
	}

	return templ.Execute(wr, data)
}

func funcMap() map[string]interface{} {
	funcs := sprig.TxtFuncMap()

	// custom []Change funcs
	funcs["count"] = countString
	funcs["components"] = changeComponents
	funcs["kinds"] = changeKinds
	funcs["bodies"] = changeBodies
	funcs["times"] = changeTimes
	funcs["customs"] = changeCustoms

	return funcs
}

func countString(item string, items []string) (int, error) {
	count := 0

	for _, i := range items {
		if i == item {
			count++
		}
	}

	return count, nil
}

func changeComponents(changes []Change) ([]string, error) {
	comps := make([]string, len(changes))

	for i, c := range changes {
		comps[i] = c.Component
	}

	return comps, nil
}

func changeKinds(changes []Change) ([]string, error) {
	kinds := make([]string, len(changes))

	for i, c := range changes {
		kinds[i] = c.Kind
	}

	return kinds, nil
}

func changeBodies(changes []Change) ([]string, error) {
	bodies := make([]string, len(changes))

	for i, c := range changes {
		bodies[i] = c.Body
	}

	return bodies, nil
}

func changeTimes(changes []Change) ([]time.Time, error) {
	times := make([]time.Time, len(changes))

	for i, c := range changes {
		times[i] = c.Time
	}

	return times, nil
}

func changeCustoms(changes []Change, key string) ([]string, error) {
	var ok bool

	values := make([]string, len(changes))

	for i, c := range changes {
		values[i], ok = c.Custom[key]
		if !ok {
			return nil, fmt.Errorf("missing custom key: '%v'", key)
		}
	}

	return values, nil
}
