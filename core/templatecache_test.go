package core

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/miniscruff/changie/then"
)

func TestCachesTemplateForReuse(t *testing.T) {
	cache := NewTemplateCache()

	first, err := cache.Load("### {{.Kind}}")
	then.Nil(t, err)

	second, err := cache.Load("### {{.Kind}}")
	then.Nil(t, err)

	// use == to compare pointers
	then.True(t, first == second)
}

func TestCacheExecuteTemplate(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := map[string]string{
		"A": "apple",
	}

	err := cache.Execute("A: {{.A}}", builder, data)
	then.Nil(t, err)
	then.Equals(t, "A: apple", builder.String())
}

func TestCanUseSprigFuncs(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	expected := "HELLO!HELLO!HELLO!HELLO!HELLO!"

	err := cache.Execute("{{ \"hello!\" | upper | repeat 5 }}", builder, nil)
	then.Nil(t, err)
	then.Equals(t, expected, builder.String())
}

func TestCanCountStrings(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Kind: "added", Body: "A"},
			{Kind: "added", Body: "B"},
			{Kind: "removed", Body: "C"},
		},
	}

	err := cache.Execute("{{ kinds .Changes | count \"added\" }} kinds", builder, data)
	then.Nil(t, err)
	then.Equals(t, "2 kinds", builder.String())
}

func TestCanGetKinds(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Kind: "added", Body: "A"},
			{Kind: "added", Body: "B"},
			{Kind: "removed", Body: "C"},
		},
	}

	err := cache.Execute("{{ kinds .Changes }} kinds", builder, data)
	then.Nil(t, err)
	then.Equals(t, "[added added removed] kinds", builder.String())
}

func TestCanGetComponents(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Component: "ui", Body: "A"},
			{Component: "ui", Body: "B"},
			{Component: "backend", Body: "C"},
		},
	}

	err := cache.Execute("{{ components .Changes }} components", builder, data)
	then.Nil(t, err)
	then.Equals(t, "[ui ui backend] components", builder.String())
}

func TestCanGetBodies(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Component: "ui", Body: "A"},
			{Component: "ui", Body: "B"},
			{Component: "backend", Body: "C"},
		},
	}

	err := cache.Execute("{{ bodies .Changes }} bodies", builder, data)
	then.Nil(t, err)
	then.Equals(t, "[A B C] bodies", builder.String())
}

func TestCanGetTimes(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	expected := fmt.Sprintf(
		"[%v %v %v] times",
		"2022-01-01 09:30:00 +0000 UTC",
		"2022-02-01 10:30:00 +0000 UTC",
		"2022-03-01 11:30:00 +0000 UTC",
	)
	data := BatchData{
		Changes: []Change{
			{Time: time.Date(2022, time.January, 1, 9, 30, 0, 0, time.UTC)},
			{Time: time.Date(2022, time.February, 1, 10, 30, 0, 0, time.UTC)},
			{Time: time.Date(2022, time.March, 1, 11, 30, 0, 0, time.UTC)},
		},
	}

	err := cache.Execute("{{ times .Changes }} times", builder, data)
	then.Nil(t, err)
	then.Equals(t, expected, builder.String())
}

func TestCanGetCustoms(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Custom: map[string]string{"Author": "miniscruff"}},
			{Custom: map[string]string{"Author": "rocktavious"}},
			{Custom: map[string]string{"Author": "emmyoop"}},
		},
	}

	err := cache.Execute("{{ customs .Changes \"Author\" }} authors", builder, data)
	then.Nil(t, err)
	then.Equals(t, "[miniscruff rocktavious emmyoop] authors", builder.String())
}

func TestCanGetCustomsMissingKey(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}
	data := BatchData{
		Changes: []Change{
			{Custom: map[string]string{"Author": "miniscruff"}},
			{Custom: map[string]string{"NotAuthor": "rocktavious"}},
			{Custom: map[string]string{"Author": "emmyoop"}},
		},
	}

	err := cache.Execute("{{ customs .Changes \"Author\" | compact }} authors", builder, data)
	then.Nil(t, err)
	then.Equals(t, "[miniscruff emmyoop] authors", builder.String())
}

func TestErrorBadTemplate(t *testing.T) {
	cache := NewTemplateCache()
	_, err := cache.Load("{{...___..__}}")
	then.NotNil(t, err)
}

func TestErrorUsingSameTemplateTwice(t *testing.T) {
	cache := NewTemplateCache()
	_, err := cache.Load("{{...___..__}}")
	then.NotNil(t, err)

	_, err = cache.Load("{{...___..__}}")
	then.NotNil(t, err)
}

func TestErrorBadExecute(t *testing.T) {
	cache := NewTemplateCache()
	builder := &strings.Builder{}

	err := cache.Execute("bad template {{...__.}}", builder, struct{}{})
	then.NotNil(t, err)
	then.Equals(t, builder.Len(), 0)
}
