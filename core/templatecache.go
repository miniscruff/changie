package core

import (
	"io"
	"text/template"
)

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

	templ, err := template.New(text).Parse(text)
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
